import {RPRWS} from "./RPRWS.ts";
import {RPRIndexedDB} from "./RPRIndexedDB.ts";
import {SMap} from "../utils/SMap.ts";
import {IndexedDBAble, newLogger, waitAMilli} from "../utils/Utils.ts";
import {
    RPRChangeResponse,
    RPRChangeResponseType,
    RPRCreateChangeType,
    RPRDeleteResponse,
    RPRFullSyncResponse,
    RPRFullSyncResponseType,
    RPRItem,
    RPRPartialSyncResponse,
    RPRPartialSyncResponseType,
    RPRRemoveChangeType,
    RPRRemoveCollectionResponseType,
    RPRRequest,
    RPRRequestType,
    RPRResponse,
    RPRResponseTypes,
} from "./RPRProtocol.ts";

export interface IRPRCoordinator {
    init: () => Promise<void>;
    close: () => Promise<void>;

    connectionChange: () => Promise<void>;

    fetchData: <T extends IndexedDBAble>(name: string) => Promise<SMap<T>>;

    subscribeToChangeEvents: <T extends IndexedDBAble>(collectionName: string, id: string, listener: LiveDataChangeListener<T>) => Promise<void>;
    unsubscribeToChangeEvents: (collectionName: string, id: string) => Promise<void>;

    subscribeConnectionStatus: (id: string, listener: (status: string) => void) => Promise<void>;
    unsubscribeConnectionStatus: (id: string) => Promise<void>;
}

export interface LiveDataChangeEvent<T> {
    type: string;

    id: string;

    version: number;

    before?: T;
    after?: T;
}

export type LiveDataChangeListener<T extends IndexedDBAble> = (events: LiveDataChangeEvent<T>[]) => void;

export type StatusChangeListener = (status: string) => void;

const logger = newLogger("RPRCoordinator");

export class RPRCoordinator implements IRPRCoordinator {

    // mapping - collection -> id -> value
    private cachedData: SMap<SMap<IndexedDBAble>> = new SMap<SMap<IndexedDBAble>>();

    // mapping - collection -> id -> version
    private deletedCachedData: SMap<SMap<number>> = new SMap<SMap<number>>();

    private readonly db: RPRIndexedDB = new RPRIndexedDB();

    // mapping - collection -> id -> listener
    private readonly changeListeners: SMap<SMap<LiveDataChangeListener<IndexedDBAble>>> = new SMap<SMap<LiveDataChangeListener<IndexedDBAble>>>();

    // mapping - collection -> listener
    private readonly statusListeners: SMap<StatusChangeListener> = new SMap<StatusChangeListener>();

    private initStarted: boolean = false;
    private started: boolean = false;

    private onTransportConnect = async () => await this.requestSync();

    private onTransportMessage = async (message: string) => {
        const parsed = JSON.parse(message) as RPRResponse;
        if (parsed == null || parsed.type == null || !RPRResponseTypes.includes(parsed.type)) {
            logger.info("Receive message is not valid or the type of response is not valid", message);
            return;
        }

        if (parsed.type === RPRRemoveCollectionResponseType) {
            const message = parsed as unknown as RPRDeleteResponse;
            await this.db.removeCollection(message.collection_name);
            delete this.cachedData[message.collection_name];
            return;
        }

        if (parsed.type === RPRFullSyncResponseType) {
            const message = parsed as unknown as RPRFullSyncResponse;
            await this.db.setItems(message.collection_name, message.values, message.removed_ids, message.version);
            this.appendCacheData(message.collection_name, message.values, message.removed_ids);

            if (message.values != null) {
                const events: LiveDataChangeEvent<IndexedDBAble>[] = [];
                for (const [, item] of message.values.entries()) {
                    events.push({
                        type: RPRCreateChangeType,
                        id: item.id,
                        version: item.updated_at,
                        after: item,
                    })
                }
                this.sendEvents(message.collection_name, events);
            }
            return;
        }

        if (parsed.type === RPRPartialSyncResponseType) {
            const message = parsed as unknown as RPRPartialSyncResponse;
            await this.db.setItems(message.collection_name, message.values, null, null);
            this.appendCacheData(message.collection_name, message.values, null);

            if (message.values != null) {
                const events: LiveDataChangeEvent<IndexedDBAble>[] = [];
                for (const [, item] of message.values.entries()) {
                    events.push({
                        type: RPRCreateChangeType,
                        id: item.id,
                        version: item.updated_at,
                        after: item,
                    })
                }
                this.sendEvents(message.collection_name, events);
            }

            return;
        }

        if (parsed.type === RPRChangeResponseType) {
            const message = parsed as unknown as RPRChangeResponse;

            if (message.change_type === RPRRemoveChangeType) {
                await this.db.deleteItem(message.collection_name, message.id);
                this.removeCacheData(message.collection_name, message);
            } else {
                await this.db.setItem(message.collection_name, message.after);
                this.appendCacheData(message.collection_name, [message.after], null);
            }

            this.sendEvents(message.collection_name, [{
                type: message.change_type,
                id: message.id,
                version: message.updated_at,
                before: message.before,
                after: message.after,
            }])
            return;
        }

        logger.error("transport type is not handled", parsed.type, parsed);
    };

    private onTransportStatusChange = (value: string) => {
        for (const [, listener] of SMap.foreach(this.statusListeners)) {
            listener(value);
        }
    }

    // ws
    private readonly wsInstance: RPRWS = new RPRWS(this.onTransportConnect, this.onTransportMessage, this.onTransportStatusChange);

    public init = async () => {
        if (this.initStarted) {
            return;
        }

        this.initStarted = true;

        await Promise.all([
            this.db.load(),
            this.wsInstance.init(),
        ]);

        this.cachedData = await this.db.prefetchCollections();

        this.started = true;
    }

    public close = async () => {
        await this.wsInstance?.close();
    }

    public fetchData = async <T extends IndexedDBAble>(name: string): Promise<SMap<T>> => {
        // busy wait until init completes
        while (!this.started) {
            await waitAMilli();
        }

        const data = this.cachedData[name];
        if (data == null) {
            return await this.db.readCollection(name) as unknown as SMap<T>;
        }
        return data as unknown as SMap<T>;
    }

    public subscribeToChangeEvents = async <T extends IndexedDBAble>(collectionName: string, id: string, listener: LiveDataChangeListener<T>) => {
        const listeners = this.changeListeners[collectionName] ?? new SMap<LiveDataChangeListener<T>>();
        listeners[id] = listener;
        // @ts-ignore
        this.changeListeners[collectionName] = listeners;
    }

    public unsubscribeToChangeEvents = async (collectionName: string, id: string) => {
        delete this.changeListeners[collectionName]?.[id];
    }

    public connectionChange = async () => await this.wsInstance?.tryToConnectIfClosed();

    public subscribeConnectionStatus = async (id: string, listener: StatusChangeListener): Promise<void> => {
        this.statusListeners[id] = listener;
    }

    public unsubscribeConnectionStatus = async (id: string): Promise<void> => {
        delete this.statusListeners[id];
    }

    private requestSync = async () => {
        while (!this.started) {
            await waitAMilli();
        }

        await this.wsInstance?.send({
            type: RPRRequestType,
            collection_versions: await this.db.fetchCollectionVersions(),
        } as RPRRequest);
    }

    private isOld = (collectionName: string, id: string, newVersion: number): boolean => {
        const newValues = this.cachedData[collectionName] ?? new SMap<IndexedDBAble>();
        const currentCachedVersion = newValues[id]?.updated_at;
        if (currentCachedVersion != null && currentCachedVersion > newVersion) {
            return true
        }

        const currentDeleted = this.deletedCachedData[collectionName]?.[id]
        return currentDeleted != null && currentDeleted > newVersion;
    }

    private sendEvents = <T extends IndexedDBAble>(collectionName: string, events: LiveDataChangeEvent<T>[]) => {
        const listeners = this.changeListeners[collectionName];
        if (listeners == null) {
            return;
        }

        for (const [_, listener] of SMap.foreach(listeners)) {
            listener(events);
        }
    }

    private appendCacheData = (collectionName: string, values: RPRItem[], removedIds: SMap<number>) => {
        const newValues = this.cachedData[collectionName] ?? new SMap<IndexedDBAble>();
        values?.forEach((value) => {
            if (!this.isOld(collectionName, value.id, value.updated_at)) {
                newValues[value.id] = value;
                delete this.deletedCachedData[collectionName]?.[value.id];
            }
        });

        if (removedIds != null) {
            for (const [id, updatedAt] of SMap.foreach(removedIds)) {
                if (this.isOld(collectionName, id, updatedAt)) {
                    continue
                }
                delete newValues[id];
            }
        }
        this.cachedData[collectionName] = newValues;
    }

    private removeCacheData = (collectionName: string, change: RPRChangeResponse) => {
        if (this.isOld(collectionName, change.id, change.updated_at)) {
            return;
        }

        const newDeleted = this.deletedCachedData[collectionName] ?? new SMap<number>();
        newDeleted[change.id] = change.updated_at;
        this.deletedCachedData[collectionName] = newDeleted;

        const newValues = this.cachedData[collectionName] ?? new SMap<IndexedDBAble>();
        delete newValues[change.id];
        this.cachedData[collectionName] = newValues;
    }
}
