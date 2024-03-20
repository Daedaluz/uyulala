import {useSearchParams} from "react-router-dom";
import {useEffect, useState} from "react";
import {useNavigate} from 'react-router-dom';
import {ApiError} from "./Api/common.ts";
import {useApi} from "./Context/Api.tsx";

export const Authorize = () => {
    const [params] = useSearchParams();
    const {publicApi: api} = useApi();
    const navigate = useNavigate();
    const [error, setError] = useState<ApiError>();

    useEffect(() => {
        console.log("params", params);
        api.createOAuth2Challenge(params).then((response) => {
            navigate(`/authenticator?id=${response.challenge_id}`, {replace: true});
        }).catch((error) => {
            setError(error);
        });
    }, []);

    return (
        <>
            <h1>Authorize</h1>
            {error && <p>{JSON.stringify(error)}</p>}
        </>
    )
}