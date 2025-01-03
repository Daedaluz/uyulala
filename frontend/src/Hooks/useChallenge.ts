import {useEffect, useState} from "react";
import {ApiError} from "../Api/common.ts";
import {App, ICredentialRequestOptions, SignData} from "../Api/public.ts";
import {useApi} from "../Context/Api.tsx";


export const useChallenge = (token: string) => {
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<ApiError | undefined>();
    const {publicApi: api} = useApi();
    const [createOptions, setCreateOptions] = useState<CredentialCreationOptions | null>(null);
    const [assertOptions, setAssertOptions] = useState<CredentialRequestOptions | null>(null);
    const [signData, setSignData] = useState<SignData | undefined>(undefined);
    const [app, setApp] = useState<App>({
        admin: false,
        description: "",
        icon: "",
        name: "",
        publicKey: ""
    });

    useEffect(() => {
        if (token && token !== "") {
            setLoading(true);
            setError(undefined);
            setAssertOptions(null);
            setCreateOptions(null);
            api.getChallenge(token).then(challenge => {
                setApp(challenge.app);
                setSignData(challenge.signData);
                if (challenge instanceof ICredentialRequestOptions) {
                    setAssertOptions(challenge);
                } else {
                    setCreateOptions(challenge);
                }
            }).catch(e => {
                setError(e);
            }).finally(() => {
                setLoading(false);
            });
        } else {
            setLoading(false);
            setError({msg: "No token provided", error: "no_token", technicalMsg: "No token provided", code: 0});
        }
    }, [token, api]);

    return {assertOptions, createOptions, app, signData, loading, error}
}
