import {useSearchParams} from "react-router-dom";
import {useEffect, useState} from "react";
import {useNavigate} from 'react-router-dom';
import {ApiError} from "./Api/common.ts";
import {useApi} from "./Context/Api.tsx";
import * as jose from "jose/jwt/sign";

export const Authorize = () => {
    const [params] = useSearchParams();
    const {publicApi: api} = useApi();
    const navigate = useNavigate();
    const [error, setError] = useState<ApiError>();

    useEffect(() => {
        api.createOAuth2Challenge(params).then((response) => {
            const payload = {challenge_id: response.challenge_id, persistent: true};
            const secret = new TextEncoder().encode(response.secret);
            new jose.SignJWT(payload)
                .setProtectedHeader({alg: 'HS256'})
                .sign(secret)
                .then(token => {
                    navigate(`/authorize?token=${token}`);
                })
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
