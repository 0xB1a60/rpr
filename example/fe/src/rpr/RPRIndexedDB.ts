import {IDBPDatabase, openDB} from "idb";
import {SMap} from "../utils/SMap.ts";
import {IndexedDBAble, newLogger, randomInt} from "../utils/Utils.ts";
import {toInteger} from "lodash";
import {PREFETCH_COLLECTIONS} from "./RPRConst.ts";
import {RPRItem} from "./RPRProtocol.ts";

const METADATA_DB_NAME = "meta";
const METADATA_VERSIONS_NAME = "versions";

const DEFAULT_STORE_NAME: string = 'main';
const DEFAULT_VERSION: number = 1;

const logger = newLogger("RPRIndexedDB");

export class RPRIndexedDB {

    private dbInstances: SMap<IDBPDatabase> = new SMap<IDBPDatabase>();

    private metadataDB: IDBPDatabase = null;

    public load = async () => {
        try {
            this.metadataDB = await initMetaDBInstance();
        } catch (e) {
            logger.error("IndexedDB is not supported", e)
            return;
        }

        for (const [, collectionName] of PREFETCH_COLLECTIONS.entries()) {
            this.dbInstances[collectionName] = await initDBInstance(collectionName);
        }
    }

    public isSupported = (): boolean => {
        return this.metadataDB != null;
    }

    public fetchCollectionVersions = async (): Promise<SMap<number>> => {
        const result: SMap<number> = new SMap<number>();
        // inject randomly a deleted collection for feature preview
        if (randomInt(0, 5) === 5) {
            result["deleted_test_collection"] = new Date().getTime();
        }

        if (!this.isSupported()) {
            return undefined;
        }

        for (const [, entry] of (await this.metadataDB.getAll(METADATA_VERSIONS_NAME)).entries()) {
            result[entry.id] = toInteger(entry.value);
        }

        return SMap.isEmpty(result) ? undefined : result;
    }

    public removeCollection = async (collectionName: string) => {
        if (!this.isSupported()) {
            return;
        }

        const instance = this.dbInstances[collectionName];
        if (instance == null) {
            return;
        }

        instance.deleteObjectStore(DEFAULT_STORE_NAME);
        delete this.dbInstances[DEFAULT_STORE_NAME];

        await this.metadataDB.delete(METADATA_VERSIONS_NAME, collectionName);
    }

    public setItem = async (collectionName: string, item: RPRItem) => {
        if (!this.isSupported()) {
            return;
        }

        let instance: IDBPDatabase = this.dbInstances[collectionName];
        if (instance == null) {
            instance = await initDBInstance(collectionName);
        }
        await instance.put(DEFAULT_STORE_NAME, item);
    }

    public setItems = async (collectionName: string, items: RPRItem[], removedIds: SMap<number>, version: number) => {
        if (!this.isSupported()) {
            return;
        }

        let instance: IDBPDatabase = this.dbInstances[collectionName];
        if (instance == null) {
            instance = await initDBInstance(collectionName);
        }

        const tx = instance.transaction(instance.objectStoreNames, 'readwrite');
        const store = tx.objectStore(instance.objectStoreNames[0]);

        if (items != null) {
            for (const [, item] of items.entries()) {
                const current = await store.get(item.id) as IndexedDBAble;
                if (current?.updated_at > item.updated_at) {
                    continue
                }

                await store.put(item);
            }
        }

        if (removedIds != null) {
            for (const [removedId, removedVersion] of SMap.foreach(removedIds)) {
                const current = await store.get(removedId) as IndexedDBAble;
                if (current?.updated_at > removedVersion) {
                    continue
                }

                await store.delete(removedId);
            }
        }

        await tx.done;

        if (version != null) {
            await this.metadataDB.put(METADATA_VERSIONS_NAME, {
                id: collectionName,
                value: version,
            });
        }
    }

    public deleteItem = async (collectionName: string, id: string) => {
        if (!this.isSupported()) {
            return;
        }

        const instance = this.dbInstances[collectionName];
        if (instance != null) {
            await instance.delete(DEFAULT_STORE_NAME, id);
        }
    }

    public prefetchCollections = async (): Promise<SMap<SMap<IndexedDBAble>>> => {
        if (!this.isSupported()) {
            return new SMap<SMap<IndexedDBAble>>();
        }

        const result: SMap<SMap<IndexedDBAble>> = new SMap<SMap<IndexedDBAble>>();
        for (const [, collectionName] of PREFETCH_COLLECTIONS.entries()) {
            result[collectionName] = await this.readCollection(collectionName);
        }
        return result;
    }

    public readCollection = async (collectionName: string): Promise<SMap<IndexedDBAble>> => {
        if (this.dbInstances[collectionName] == null) {
            logger.info("collection does not exist", collectionName)
            return new SMap<IndexedDBAble>();
        }

        const values: SMap<IndexedDBAble> = new SMap<IndexedDBAble>();
        for (const [, entry] of (await this.dbInstances[collectionName].getAll(DEFAULT_STORE_NAME)).entries()) {
            values[entry.id] = entry;
        }
        return values;
    }
}

const initDBInstance = async (name: string) => {
    return await openDB(name, DEFAULT_VERSION, {
        upgrade(database: IDBPDatabase) {
            database.createObjectStore(DEFAULT_STORE_NAME, {
                keyPath: "id",
            });
        },
    })
}

const initMetaDBInstance = async () => {
    return await openDB(METADATA_DB_NAME, DEFAULT_VERSION, {
        upgrade(database: IDBPDatabase) {
            database.createObjectStore(DEFAULT_STORE_NAME, {
                keyPath: "id",
            });
            database.createObjectStore(METADATA_VERSIONS_NAME, {
                keyPath: "id",
            });
        },
    })
}
