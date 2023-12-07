import {newLogger, waitAMilli} from "../utils/Utils.ts";
import {isString} from "lodash";
import {OFFLINE_STATUS, ONLINE_STATUS, WS_URL} from "./RPRConst.ts";

const RECONNECT_INTERVAL: number = 15_000;

const logger = newLogger("RPRWS");

export class RPRWS {

    private reconnectInterval: number = null;
    private instance: WebSocket = null;

    public constructor(private readonly onConnectFunc: () => Promise<void>,
                       private readonly onMessageFunc: (value: string) => Promise<void>,
                       private readonly onStatusFunc: (status: string) => void) {
    }

    public init = async () => {
        this.instance = this.initWs();
    }

    private initWs = (): WebSocket => {
        const local: WebSocket = new WebSocket(WS_URL);
        local.onclose = (e) => {
            logger.debug("ws close", e.code);

            this.onStatusFunc(OFFLINE_STATUS);

            if (this.reconnectInterval == null) {
                this.reconnectInterval = setInterval(() => {
                    this.instance = this.initWs();
                }, RECONNECT_INTERVAL) as unknown as number;
            }
        }
        local.onopen = async () => {
            logger.debug("ws open");

            this.onStatusFunc(ONLINE_STATUS);

            if (this.reconnectInterval != null) {
                clearInterval(this.reconnectInterval);
                this.reconnectInterval = null;
            }

            await this.onConnectFunc();
        }
        local.onmessage = (e: MessageEvent) => this.onMessage(e.data);
        return local;
    }

    private onMessage = async (message: unknown) => {
        if (!isString(message)) {
            logger.debug("non-string message is not supported")
            return;
        }

        logger.debug("incoming message", JSON.parse(message));
        await this.onMessageFunc(message);
    }

    public tryToConnectIfClosed = async () => {
        if (this.instance == null) {
            return;
        }

        if (this.instance.readyState === WebSocket.CONNECTING || this.instance.readyState === WebSocket.OPEN) {
            return;
        }

        if (this.reconnectInterval != null) {
            clearInterval(this.reconnectInterval);
            this.reconnectInterval = null;
        }

        this.instance = this.initWs();
    };

    public close = async () => {
        logger.info("close");
        if (this.reconnectInterval != null) {
            clearInterval(this.reconnectInterval);
            this.reconnectInterval = null;
        }
        this.instance?.close();
    }

    public send = async (message: unknown) => {
        logger.debug("sending message", message);
        // busy wait until init completes
        while (this.instance == null || this.instance.readyState === WebSocket.CONNECTING) {
            await waitAMilli();
        }

        if (this.instance.readyState === WebSocket.CLOSED || this.instance.readyState === WebSocket.CLOSING) {
            logger.debug("socket is closed");
            return;
        }

        this.instance.send(JSON.stringify(message));
    }
}
