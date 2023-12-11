import {useEffect, useState} from "react";
import {useApi} from "../Context/Api.tsx";
import {Metadata} from "../Api/private.ts";

export const useMDS = (aaguid: string) => {
    const [loading, setLoading] = useState(true)
    const [metadata, setMetadata] = useState<Metadata | null>(null)
    const {privateApi} = useApi();

    useEffect(() => {
        privateApi.getAuthenticatorDescriptor(aaguid).then((metadata) => {
            setMetadata(metadata)
        }).catch((e) => {
            console.error(e)
        }).finally(() => {
            setLoading(false);
        });
    }, [aaguid])
    return {loading, metadata}
}