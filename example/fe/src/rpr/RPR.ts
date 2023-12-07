import {wrap} from "comlink";
import {IRPRCoordinator, RPRCoordinator} from "./RPRCoordinator.ts";
import {newLogger} from "../utils/Utils.ts";

const logger = newLogger("RPR");

let WORKER_INSTANCE: IRPRCoordinator = null;
export const initRPR = () => {
    const worker = new Worker(new URL("./RPRWorker.ts", import.meta.url), {
        type: 'module',
    });

    worker.onerror = (e: ErrorEvent) => {
        logger.error("worker error", e.error);
        WORKER_INSTANCE = null;
    }

    worker.onmessageerror = (e) => {
        logger.error("worker message error", e.data);
        WORKER_INSTANCE = null;
    }

    // @ts-ignore
    WORKER_INSTANCE = wrap<RPRCoordinator>(worker);
    WORKER_INSTANCE.init();

    window.ononline = async () => await WORKER_INSTANCE.connectionChange();
    window.onunload = async () => await WORKER_INSTANCE.close();
}

const FALLBACK_INSTANCE: IRPRCoordinator = new RPRCoordinator();
export const getLiveData = (): IRPRCoordinator => {
    if (WORKER_INSTANCE != null) {
        return WORKER_INSTANCE;
    }

    FALLBACK_INSTANCE.init();
    window.onunload = async () => await FALLBACK_INSTANCE.close();
    window.ononline = async () => await FALLBACK_INSTANCE.connectionChange();

    return FALLBACK_INSTANCE;
}
