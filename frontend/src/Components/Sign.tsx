import {useMemo} from "react";
import {App, SignData} from "../Api/public.ts";
import {useApi} from "../Context/Api.tsx";
import {useAlert} from "../Context/Alert.tsx";
import {Button, Paper, Typography} from "@mui/material";
import Markdown from "react-markdown";
import remarkGfm from "remark-gfm";
import remarkRehype from "remark-rehype";

export type SignProps = {
    id: string
    challenge: CredentialRequestOptions
    app: App
    signData?: SignData
}
export const Sign = ({id, challenge, app, signData}: SignProps) => {
    const {publicApi: api} = useApi();
    const {showAlert} = useAlert();
    const signHandler = () => {
        navigator.credentials.get(challenge).then((credential) => {
            if (credential) {
                api.sign(id, credential).then((response) => {
                    if (response.redirect === '') {
                        window.close();
                    } else {
                        window.location.href = response.redirect;
                    }
                });
            }
        }).catch((error) => {
            showAlert('error', 'Error', error.message, 5000);
        });
    }
    const defaultText = `# Sign in to ${app.name}`;
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
        <div className={"displayLayout"}>
            <div className={'signapp'}>
                <Typography variant={'h2'} component={"h2"}>{app.name}</Typography>
            </div>
            <Paper className={'signtext'}>
                <Markdown
                    remarkPlugins={[[remarkGfm, {singleTilde: true}], [remarkRehype, {}]]}>{signData?.text ?? defaultText}</Markdown>
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
    )
}
