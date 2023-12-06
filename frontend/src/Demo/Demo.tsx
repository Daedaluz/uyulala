import {useEffect, useMemo, useState} from "react";
import {useLocation, useNavigate, useSearchParams} from "react-router-dom";
import {Key, UserWithKeys} from "../Api/admin.ts";
import {useApi} from "../Context/Api.tsx";
import {BIDResponse} from "../Api/private.ts";
import {ApiError} from "../Api/common.ts";
import {
    Accordion, AccordionActions, AccordionDetails,
    AccordionSummary,
    Alert,
    AlertTitle,
    Button,
    FormControlLabel,
    FormGroup,
    Switch,
    TextField,
    Typography, Table, TableHead, TableBody, TableRow, TableCell, Badge, Popover,

} from "@mui/material";
import {useMDS} from "../Hooks/useMDS.ts";
import {InfoRounded} from "@mui/icons-material";
import {useLocalStorage} from "../Hooks/useLocalStorage.ts";

interface User {
    id: string;
}

type UserKeyProps = {
    user: User
    userKey: Key
}
const UserKey = ({user, userKey}: UserKeyProps) => {
    const {adminApi} = useApi();
    const onDelete = () => {
        adminApi.deleteUserKey(user.id, userKey.hash).then(() => {
            window.location.reload();
        });
    };
    return (
        <TableRow>
            <TableCell>{userKey.key.Authenticator.SignCount}</TableCell>
            <TableCell>{userKey.hash}</TableCell>
            <TableCell><AAGUID aaguid={userKey.aaguid}/></TableCell>
            <TableCell>{userKey.created}</TableCell>
            <TableCell>{userKey.lastUsed}</TableCell>
            <TableCell><Button onClick={onDelete}>Delete Key</Button></TableCell>
        </TableRow>
    )
}

const AAGUID = ({aaguid}: { aaguid: string }) => {
    const {metadata} = useMDS(aaguid)
    const [show, setShow] = useState<boolean>(false);

    const handleShow = () => {
        setShow(!show);
    }

    const content = metadata !== null ? <>
        <InfoRounded fontSize={'small'} onClick={handleShow}/>
        <img alt="metaImage" src={metadata.metadataStatement.icon} height={32} width={32}/>
        {metadata.metadataStatement.description}
        <Popover open={show} onClose={() => setShow(false)} anchorOrigin={{vertical: 'bottom', horizontal: 'center'}}>
            <pre>
                {JSON.stringify(metadata, null, 4)}
            </pre>
        </Popover>
    </> : <>{aaguid}</>;
    return <div style={{display: 'inline-grid', gap: '5px', gridAutoFlow: 'column', alignItems: 'center'}}>
        {content}
    </div>
}

type UserViewProps = {
    user: UserWithKeys;
    name: string;
    text: string;
    verify: boolean;
}
const UserView = ({name, verify, user, text}: UserViewProps) => {
    const keys = user.keys.map((userKey, i) => <UserKey key={i} user={user} userKey={userKey}/>);
    const {adminApi, privateApi} = useApi();
    const navigate = useNavigate();
    const addKey = () => {
        adminApi.addUserKey(name, user.id).then(res => {
            navigate(`/authenticator?id=${res.challenge_id}`)
        });
    }

    const authenticate = () => {
        privateApi.sign({
            userId: user.id,
            redirect: window.location.toString(),
            timeout: 60000,
            text: text,
            userVerification: verify ? 'required' : 'discouraged'
        }).then(res => {
            navigate(`/authenticator?id=${res.challenge_id}`)
        });
    }

    const deleteUser = () => {
        adminApi.deleteUser(user.id).then(() => {
            window.location.reload();
        });
    }

    return (
        <Accordion>
            <AccordionSummary style={{alignItems: 'center', justifyContent: 'space-between', display: 'flex'}}>
                <Badge badgeContent={user.keys.length} color={'primary'}>
                    <Typography>{user.id}</Typography>
                </Badge>
            </AccordionSummary>
            <AccordionDetails>
                <Table>
                    <TableHead>
                        <TableRow>
                            <TableCell>SigCount</TableCell>
                            <TableCell>Key Hash</TableCell>
                            <TableCell>AAGUID</TableCell>
                            <TableCell>Created</TableCell>
                            <TableCell>Last used</TableCell>
                            <TableCell>Actions</TableCell>
                        </TableRow>
                    </TableHead>
                    <TableBody>
                        {...keys}
                    </TableBody>
                </Table>
            </AccordionDetails>
            <AccordionActions>
                <Button onClick={authenticate}>Authenticate</Button>
                <Button onClick={addKey}>Add key</Button>
                <Button onClick={deleteUser}>Delete user</Button>
            </AccordionActions>
        </Accordion>
    )
}

export const DemoPage = () => {
    const {
        adminApi, privateApi,
        setApiCredentials, credentials
    } = useApi();
    const [name, setName] = useLocalStorage('username', 'Kalle Anka');
    const [users, setUsers] = useState<UserWithKeys[]>([]);
    const navigate = useNavigate();
    const location = useLocation();
    const [params] = useSearchParams();

    const [requireUserVerification, setRequireUserVerification] = useLocalStorage<boolean>('requireUV', false);

    const [signText, setSignText] = useState<string>('');

    useEffect(() => {
        if (params.get('challengeId') !== undefined) {
            navigate('', {replace: true, state: {challengeId: params.get('challengeId')}})
        }
    }, [params]);

    const [result, setResult] = useState<BIDResponse | ApiError | null>(null);
    useEffect(() => {
        if (location?.state?.challengeId) {
            privateApi.collect(location.state.challengeId).then((res) => {
                setResult(res);
            }).catch(e => {
                setResult(e);
            });
        }
    }, [location.state]);

    useEffect(() => {
        adminApi.listUsers().then((res) => {
            setUsers(res);
        }).catch((_) => {
            setUsers([]);
        });
    }, [adminApi]);

    const authenticate = () => {
        privateApi.sign({
            redirect: window.location.toString(),
            timeout: 60000,
            text: signText,
            userVerification: requireUserVerification ? 'required' : 'discouraged'
        }).then(res => {
            navigate(`/authenticator?id=${res.challenge_id}`)
        });
    }

    const registerUser = () => {
        adminApi.registerUser(name).then((res) => {
            navigate(`/authenticator?id=${res.challenge_id}`)
        })
    }

    const resp = result as BIDResponse;
    const alertSeverity = useMemo(() => {
        if (resp === null) {
            return
        }
        switch (resp.status) {
            case 'rejected':
                return 'error';
            case 'signed':
                return 'success';
        }
    }, [resp])

    return (
        <div style={{
            paddingTop: '1em',
            display: 'grid',
            gridGap: '1em',
            paddingLeft: '1em',
            paddingRight: '1em',
            color: '#FFFFFF',
            height: 'auto',
            width: 'auto',
            overflowX: 'hidden',
            minHeight: '100vh',
        }}>
            <FormGroup style={{display: 'grid', gridGap: '1em'}}>
                <TextField value={credentials.clientId}
                           onChange={(e) => setApiCredentials(e.target.value, credentials.secret)}
                           label={'Client ID'}
                />
                <TextField value={credentials.secret}
                           label={'Client Secret'}
                           onChange={(e) => setApiCredentials(credentials.clientId, e.target.value)}
                />
                <TextField value={signText} multiline={true} label={'Text to sign'}
                           onChange={(e) => setSignText(e.target.value)}/>
                <FormControlLabel label={'Require user verification'}
                                  control={<Switch checked={requireUserVerification}
                                                   onChange={(e) => setRequireUserVerification(e.currentTarget.checked)}/>
                                  }/>
                <TextField label={'User name'}
                           onChange={(e) => setName(e.target.value)}
                           value={name}
                />
                <FormGroup style={{display: 'flex', flexDirection: 'row'}}>
                    <Button onClick={registerUser}>Create user</Button>
                    <Button onClick={authenticate}>Authenticate any user</Button>
                </FormGroup>
            </FormGroup>
            {result &&
                <Alert severity={alertSeverity} style={{width: '90vw'}}>
                    <AlertTitle>Result</AlertTitle>
                    <pre>{JSON.stringify(result, null, 4)}</pre>
                </Alert>
            }
            <div>
                {users?.map((user, i) => <UserView key={i}
                                                   user={user}
                                                   text={signText}
                                                   name={name}
                                                   verify={requireUserVerification}
                />)}
            </div>
        </div>
    )
}
