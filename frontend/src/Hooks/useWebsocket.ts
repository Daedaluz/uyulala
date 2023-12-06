import {useCallback, useEffect, useRef, useState} from "react";


function pathToURL(path: string) {
    const protocol = window.location.protocol === "https:" ? "wss" : "ws";
    return `${protocol}://${window.location.host}${path}`;
}

type HelperState = {
    event: string;
    response?: any;
}

export const useWebsocket = (path: string) => {
    const socket = useRef<WebSocket | null>(null);

    const [state, setState] = useState<string>("connecting");
    const [message, setMessage] = useState<HelperState>({event: 'connecting'});

    const send = useCallback((msg: any) => {
        const payload = JSON.stringify(msg);
        if (socket.current) {
            socket.current.send(payload);
        }
    }, []);

    const close = useCallback(() => {
        if (socket.current) {
            socket.current.close();
            socket.current = null;
        }
    }, []);

    const onOpen = useCallback(() => {
        setState("open");
    }, []);

    const onClose = useCallback(() => {
        setState("closed");
    }, []);

    const onError = useCallback(() => {
        setState("error");
    }, []);

    const onMessage = useCallback((msg: MessageEvent) => {
        setMessage(() => JSON.parse(msg.data))
    }, []);

    useEffect(() => {
        socket.current = new WebSocket(pathToURL(path));
        socket.current.onerror = onError;
        socket.current.onclose = onClose;
        socket.current.onopen = onOpen;
        return () => {
            if (socket.current) {
                socket.current.close();
            }
        }
    }, [path]);

    useEffect(() => {
        if(socket.current) {
            socket.current.onmessage = (msg) => {
                onMessage(msg);
            }
        }
    }, [onMessage]);

    return {
        send,
        close,
        state,
        message,
    }
}