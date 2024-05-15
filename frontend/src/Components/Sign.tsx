import {useEffect, useMemo, useRef, useState} from "react";
import {App, SignData} from "../Api/public.ts";
import {useApi} from "../Context/Api.tsx";
import {useWebsocket} from "../Hooks/useWebsocket.ts";
import {authnDecode, authnEncode} from "../Api/common.ts";
import {useAlert} from "../Context/Alert.tsx";
import {Button, Paper, Typography} from "@mui/material";
import Markdown from "react-markdown";
import remarkGfm from "remark-gfm";
import remarkRehype from "remark-rehype";
import {QR} from "./QR.tsx";

export type SignProps = {
    id: string
    challenge: CredentialRequestOptions
    app: App
    signData?: SignData
}
export const Sign = ({id, challenge, app, signData}: SignProps) => {
    const {publicApi: api} = useApi();
    const {message, state, send} = useWebsocket(`/api/v1/remote/${id}`);
    const [first, setFirst] = useState(false);

    const {showAlert} = useAlert();

    const status = useRef<string>('connecting')

    const [signPressed, setSignPressed] = useState(false);

    useEffect(() => {
        if (message.event == 'waiting') {
            setFirst(true);
        }
        if (message.event == 'ready' && first) {
            showAlert('info', 'Remote Sign', 'QR Code was scanned', 5000);
        }
        status.current = message.event;
        if (message.response && first) {
            const cred = authnDecode(message);
            api.sign(id, cred).then((response) => {
                if (response.redirect === '') {
                    window.close();
                } else {
                    window.location.href = response.redirect;
                }
            });
        }
    }, [message, state, first]);

    useEffect(() => {
        if (signData && signData.text !== '') {
        } else {
            navigator.credentials.get(challenge).then((credential) => {
                if (credential) {
                    if (status.current === 'ready' && !first) {
                        const signature = authnEncode(credential);
                        send(signature);
                    } else {
                        api.sign(id, credential).then((response) => {
                            if (response.redirect === '') {
                                window.close();
                            } else {
                                window.location.href = response.redirect;
                            }
                        });
                    }
                }
            }).catch((error) => {
                showAlert('error', 'Error', error.message, 5000);
            });
        }
    }, []);

    const signHandler = () => {
        setSignPressed(true);
        navigator.credentials.get(challenge).then((credential) => {
            if (credential) {
                if (status.current === 'ready' && !first) {
                    const signature = authnEncode(credential);
                    send(signature);
                } else {
                    api.sign(id, credential).then((response) => {
                        if (response.redirect === '') {
                            window.close();
                        } else {
                            window.location.href = response.redirect;
                        }
                    });
                }
            }
        }).catch((error) => {
            showAlert('error', 'Error', error.message, 5000);
        });
    }

    const cancelHandler = () => {
        api.reject(id).then((response) => {
            if (response.redirect === '') {
                window.close();
            } else {
                window.location.href = response.redirect;
            }
        });
    }

    const [ok, cancel] = useMemo(() => {
        if (challenge.publicKey && challenge.publicKey.userVerification === 'required') {
            return ["Sign", "Reject"]
        } else {
            return ["Ok", "Cancel"]
        }
    }, [challenge.publicKey]);

    return (
        <>
            {signPressed || !signData ?
                <div className={'signLayout'}>
                    <Paper className={'qr'}>
                        <h1 style={{textAlign: 'center'}}>Authenticate</h1>
                        <QR/>
                        <h2 style={{textAlign: 'center'}}>{app.name}</h2>
                    </Paper>
                </div> :
                <div className={"displayLayout"}>
                    <div className={'signapp'}>
                        <Typography variant={'h2'} component={"h2"}>{app.name}</Typography>
                    </div>
                    <Paper className={'signtext'}>
                        <Markdown
                            remarkPlugins={[[remarkGfm, {singleTilde: true}], [remarkRehype, {}]]}>{signData?.text}</Markdown>
                    </Paper>
                    <div className={'sign'}>
                        <Button variant={'contained'} color={'success'} size={'large'}
                                onClick={signHandler}>{ok}</Button>
                    </div>
                    <div className={'reject'}>
                        <Button variant={'contained'} size={'large'} color={'error'}
                                onClick={cancelHandler}>{cancel}</Button>
                    </div>
                </div>
            }
        </>
    )
}
