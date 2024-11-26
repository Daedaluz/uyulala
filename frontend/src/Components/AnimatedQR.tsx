import {ChallengeResponse} from "../Api/common.ts";
import {useEffect, useMemo, useState} from "react";

import * as jose from "jose/jwt/sign";
import {QR} from "./QR.tsx";
import {Button, Stack} from "@mui/material";

interface AnimatedQRProps {
    challenge: ChallengeResponse
    startTime: Date
}

export const AnimatedQR = ({startTime, challenge}: AnimatedQRProps) => {
    const [duration, setDuration] = useState(() => {
        // Get time since start in seconds
        return Math.floor((Date.now() - startTime.getTime()) / 1000);
    })

    useEffect(() => {
        const updater = window.setInterval(() => {
            setDuration(Math.floor((Date.now() - startTime.getTime()) / 1000));
        }, 1000)
        return () => {
            window.clearInterval(updater);
        }
    }, [startTime, challenge]);

    const [tokenData, setTokenData] = useState<string>("");

    useEffect(() => {
        // Calculate the current QR code to display.
        // JWT with {challenge: challenge, time: duration} as claims signed with HMAC-SHA256
        const payload = {challenge_id: challenge.challenge_id, duration: duration};
        const secret = new TextEncoder().encode(challenge.secret);
        new jose.SignJWT(payload)
            .setProtectedHeader({alg: 'HS256'})
            .sign(secret)
            .then(token => {
                setTokenData(token);
            })
    }, []);

    useEffect(() => {
        const payload = {challenge_id: challenge.challenge_id, duration: duration};
        const secret = new TextEncoder().encode(challenge.secret);
        new jose.SignJWT(payload)
            .setProtectedHeader({alg: 'HS256'})
            .sign(secret)
            .then(token => {
                setTokenData(token);
            })
    }, [duration, challenge]);

    const qrData = useMemo(() => {
        // Generate a URL with the token as a query parameter
        return `${window.location.origin}/authenticator?&token=${tokenData}`;
    }, [tokenData])


    return <Stack spacing={2} width={325}>
        <QR value={qrData}/>
        <Button fullWidth={false} variant={'contained'} onClick={() => window.open(qrData, 'createUser', 'resizeable,height=800,width=600')} color={'primary'}>Use this computer</Button>
    </Stack>
}
