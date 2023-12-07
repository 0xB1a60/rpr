import {Logger} from "tslog";

export interface IndexedDBAble {
    id: string;

    created_at: number;
    updated_at: number;

    [key: string]: unknown;
}

export const waitAMilli = async () =>
    await new Promise((resolve) => setTimeout(resolve, 1));

export const newLogger = (name: string): Logger<unknown> => new Logger({
    type: "pretty",
    name: name,
    hideLogPositionForProduction: true,
    minLevel: 0,
});

export const randomInt = (min: number, max: number): number =>
    Math.floor(min + Math.random() * (max + 1 - min))
