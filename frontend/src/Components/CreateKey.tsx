import {useApi} from "../Context/Api.tsx";
import {useAlert} from "../Context/Alert.tsx";
import {Button, Paper} from "@mui/material";

export type CreateKeyProps = {
    id: string
    challenge: CredentialCreationOptions
}

export const CreateKey = ({id, challenge}: CreateKeyProps) => {
    const {publicApi: api} = useApi();
    const {showAlert} = useAlert();

    const create = () => {
        navigator.credentials.create(challenge).then((credential) => {
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

    return (
        <>
            <div className={'signLayout'}>
                <Paper className={'qr'} style={{marginBottom: 20}}>
                    <h1 style={{textAlign: 'center'}}>Register</h1>
                    <h2 style={{textAlign: 'center'}}>{challenge.publicKey?.user.displayName}</h2>
                    <Button onClick={() => create()} variant={'contained'} color={'primary'}>Register</Button>
                    <Button onClick={() => window.close()} variant={'contained'} color={'error'}>Cancel</Button>
                </Paper>
            </div>
        </>
    )
}
