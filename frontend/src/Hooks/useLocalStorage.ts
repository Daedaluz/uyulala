import {Dispatch, SetStateAction, useState} from "react";

export function useLocalStorage<T>(key: string): [T | undefined, Dispatch<SetStateAction<T | undefined>>];
export function useLocalStorage<T>(key: string, initialValue?: T | (() => T)): [T, Dispatch<SetStateAction<T>>] {
    const [storedValue, setStoredValue] = useState<T>(() => {
        try {
            const item = window.localStorage.getItem(key);
            if (item === null && initialValue !== undefined) {
                const valueToStore = initialValue instanceof Function ? initialValue() : initialValue;
                window.localStorage.setItem(key, JSON.stringify(valueToStore));
            }
            return item ? JSON.parse(item) : initialValue;
        } catch (error) {
            console.log(error);
            return initialValue;
        }
    });

    const setValue = (value: T | ((prev: T) => T)) => {
        try {
            const valueToStore = value instanceof Function ? value(storedValue) : value;
            if (valueToStore === undefined) {
                window.localStorage.removeItem(key);
            } else {
                setStoredValue(valueToStore);
                window.localStorage.setItem(key, JSON.stringify(valueToStore));
            }
        } catch (error) {
            console.log(error);
        }
    };

    return [storedValue, setValue];
}
