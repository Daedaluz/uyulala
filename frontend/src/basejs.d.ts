declare module 'basejs' {
    export class Base64 {
        static urlDecode(str: string, padding = ''): ArrayBuffer;
        static urlEncode(buffer: ArrayBuffer, padding = ''): string;
        static encode(buffer: ArrayBuffer, padding = '='): string;
        static decode(str: string, padding = ''): ArrayBuffer;
    }
}