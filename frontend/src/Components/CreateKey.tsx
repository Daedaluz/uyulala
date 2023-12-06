import {useEffect, useRef, useState} from "react";
import {useApi} from "../Context/Api.tsx";
import {App} from "../Api/public.ts";
import {useAlert} from "../Context/Alert.tsx";
import {QR} from "./QR.tsx";
import {useWebsocket} from "../Hooks/useWebsocket.ts";
import {authnDecode, authnEncode} from "../Api/common.ts";

export type CreateKeyProps = {
    id: string
    challenge: CredentialCreationOptions
    app: App
}

export const CreateKey = ({id, app, challenge}: CreateKeyProps) => {
    const {publicApi: api} = useApi();
    const {showAlert} = useAlert();
    const {message, state, send} = useWebsocket(`/api/v1/remote/${id}`);
    const status = useRef<string>('connecting')

    const [first, setFirst] = useState(false);

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
        navigator.credentials.create(challenge).then((credential) => {
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
    }, []);

    return (
        <>
            <div className={'signLayout'}>
                <div className={'qr'}>
                    <h1 style={{textAlign: 'center', color: '#FFFFFF'}}>Create key</h1>
                    <h2 style={{textAlign: 'center', color: '#FFFFFF'}}>{challenge.publicKey?.user.displayName}</h2>
                    <QR app={app}/>
                </div>
            </div>
        </>
    )
}