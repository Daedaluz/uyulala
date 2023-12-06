import React, {createContext, useCallback, useMemo} from "react";
import {privateApi} from "../Api/private.ts";
import {adminApi} from "../Api/admin.ts";
import {publicApi} from "../Api/public.ts";
import {useLocalStorage} from "../Hooks/useLocalStorage.ts";


type credentials = {
    clientId: string;
    secret: string;
}

export type ApiContext = {
    privateApi: privateApi;
    publicApi: publicApi;
    adminApi: adminApi;
    credentials: credentials;
    setApiCredentials: (clientId: string, secret: string) => void;
}

const ClientContext = createContext({} as ApiContext);

export const ApiProvider = ({children}: { children: React.ReactNode }) => {

   // const [apiCredentials, setCredentials] = useState<credentials>({clientId: 'demo', secret: 'demo'});
    const [apiCredentials, setCredentials] = useLocalStorage<credentials>('demoUser', {clientId: 'demo', secret: 'demo'});
    const setApiCredentials = useCallback((clientId: string, secret: string) => {
        setCredentials({clientId, secret});
    }, [setCredentials]);

    const privApi = useMemo<privateApi>(() => new privateApi('', apiCredentials.clientId, apiCredentials.secret), [apiCredentials.clientId, apiCredentials.secret]);
    const aApi = useMemo<adminApi>(() => new adminApi('', apiCredentials.clientId, apiCredentials.secret), [apiCredentials.clientId, apiCredentials.secret]);
    const pubApi = useMemo<publicApi>(() => new publicApi(''), []);

    const value = useMemo(() => ({
        privateApi: privApi,
        publicApi: pubApi,
        adminApi: aApi,
        credentials: apiCredentials,
        setApiCredentials
    }), [setApiCredentials, apiCredentials, privApi, pubApi, aApi]);

    return (
        <ClientContext.Provider value={value}>
            {children}
        </ClientContext.Provider>
    );
}

export const useApi = () => {
    const context = React.useContext(ClientContext);
    if (context === undefined) {
        throw new Error('useApi must be used within a ApiProvider');
    }
    return context;
}