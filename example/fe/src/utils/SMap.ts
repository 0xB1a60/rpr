// "map" that can be serialized to and from JSON
export class SMap<T> {
    [key: string]: T;

    public static add<T>(map: SMap<T>, key: string, value: T): SMap<T> {
        map[key] = value;
        return map;
    }

    public static remove<T>(map: SMap<T>, key: string): SMap<T> {
        delete map[key];
        return map;
    }

    public static isEmpty<T>(input: SMap<T>): boolean {
        return input == null ? true : Object.keys(input).length === 0;
    }

    public static foreach<T>(input: SMap<T>): [string, T][] {
        return input == null ? null : Object.entries(input);
    }
}
