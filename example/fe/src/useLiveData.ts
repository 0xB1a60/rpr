import {useCallback, useEffect, useState} from "react";
import {getLiveData} from "./rpr/RPR.ts";
import {SMap} from "./utils/SMap.ts";
import {IndexedDBAble} from "./utils/Utils.ts";
import {RPRCreateChangeType, RPRUpdateChangeType} from "./rpr/RPRProtocol.ts";
import {LiveDataChangeEvent} from "./rpr/RPRCoordinator.ts";
import {nanoid} from "nanoid";
import {proxy} from "comlink";
import {copyStrict} from "fast-copy";

export const LiveDataLoading = "loading";
export const LiveDataError = "error";
export const LiveDataReady = "ready";

export type LiveDataState = "loading" | "ready" | "error";

export interface Result<T extends IndexedDBAble> {
    liveDataState: LiveDataState;
    liveData: SMap<T>;
}

export const useLiveData = <T extends IndexedDBAble>(name: string): Result<T> => {
    const [loadState, setLoadState] = useState<LiveDataState>("loading");
    const [liveData, setLiveData] = useState<SMap<T>>();

    const [deletedItems, setDeletedItems] = useState<SMap<number>>(new SMap<number>());

    const [reFetch, setReFetch] = useState(false);

    const fetchAndSet = useCallback(() => {
        getLiveData().fetchData<T>(name).then((res) => {
            setLoadState("ready");
            setLiveData(res ?? new SMap<T>());
            if (reFetch) {
                setReFetch(false);
                fetchAndSet();
            }
        });
    }, [name, reFetch]);

    useEffect(() => {
        fetchAndSet();
    }, [fetchAndSet]);

    const onDataChange = useCallback((events: LiveDataChangeEvent<T>[]) => {
        if (loadState === "loading") {
            setReFetch(true);
            return;
        }

        setLiveData((prevState: SMap<T>) => {
            const state = copyStrict(prevState);

            for (const [, e] of events.entries()) {
                const stateVersion = state[e.id]?.updated_at;
                if (stateVersion != null && stateVersion > e.version) {
                    continue
                }

                const deletedVersion = deletedItems[e.id];
                if (deletedVersion != null && deletedVersion > e.version) {
                    continue
                }

                if (e.type === RPRCreateChangeType || e.type === RPRUpdateChangeType) {
                    state[e.id] = e.after;
                    setDeletedItems((prevDeleted) => SMap.remove(prevDeleted, e.id));
                    continue
                }
                delete state[e.id];
                setDeletedItems((prevDeleted) => SMap.add(prevDeleted, e.id, e.version));
            }

            return state;
        });
    }, [loadState, deletedItems]);

    useEffect(() => {
        const id = nanoid();
        getLiveData().subscribeToChangeEvents(name, id, proxy(onDataChange));
        return () => {
            getLiveData().unsubscribeToChangeEvents(name, id);
        }
    }, [name, onDataChange]);

    return {
        liveDataState: loadState,
        liveData: liveData,
    };
}
