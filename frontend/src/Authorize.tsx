import {useSearchParams} from "react-router-dom";
import {useEffect, useState} from "react";
import {ApiError, ChallengeResponse} from "./Api/common.ts";
import {useApi} from "./Context/Api.tsx";
import {AnimatedQR} from "./Components/AnimatedQR.tsx";

export const Authorize = () => {
    const [params] = useSearchParams();
    const {publicApi: api} = useApi();
    const [error, setError] = useState<ApiError>();

    const [challenge, setChallenge] = useState<ChallengeResponse | null>(null);
    const [startTime, setStartTime] = useState<Date>(new Date());


    useEffect(() => {
        api.createOAuth2Challenge(params).then((response) => {
            setChallenge(() => response);
            setStartTime(() => new Date());
        }).catch((error) => {
            setError(error);
        });
    }, [api, params]);

    return (
        <>
            <h1>Authorize</h1>
            {challenge && <AnimatedQR challenge={challenge} startTime={startTime} popUp={false}/>}
            {error && <p>{JSON.stringify(error)}</p>}
        </>
    )
}
