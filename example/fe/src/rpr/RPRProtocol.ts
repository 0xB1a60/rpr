import {SMap} from "../utils/SMap.ts";

export const RPRRequestType = "sync";

export const RPRRemoveCollectionResponseType = "remove_collection"
export const RPRFullSyncResponseType = "full_sync"
export const RPRPartialSyncResponseType = "partial_sync"
export const RPRChangeResponseType = "change"

export const RPRResponseTypes = [
    RPRRemoveCollectionResponseType,
    RPRFullSyncResponseType,
    RPRPartialSyncResponseType,
    RPRChangeResponseType,
];

export const RPRCreateChangeType = "create"
export const RPRUpdateChangeType = "update"
export const RPRRemoveChangeType = "remove"

export interface RPRRequest {
    type: string;
    collection_versions?: SMap<number>;
}

export interface RPRResponse {
    type: string;
}

export interface RPRDeleteResponse {
    collection_name: string;
}

export interface RPRItem {
    id: string;

    created_at: number;
    updated_at: number;

    [key: string]: unknown;
}

export interface RPRChangeResponse {
    collection_name: string;

    change_type: string;

    id: string;

    updated_at: number;

    before?: RPRItem;
    after?: RPRItem;
}

export interface RPRPartialSyncResponse {
    collection_name: string;

    values: RPRItem[];
}

export interface RPRFullSyncResponse {
    collection_name: string;

    values: RPRItem[];

    removed_ids: SMap<number>;

    version: number;
}
