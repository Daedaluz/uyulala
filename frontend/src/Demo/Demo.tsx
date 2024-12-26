import {useEffect, useMemo, useState} from "react";
import {useLocation, useNavigate, useSearchParams} from "react-router-dom";
import {Key, UserWithKeys} from "../Api/admin.ts";
import {useApi} from "../Context/Api.tsx";
import {BIDResponse} from "../Api/private.ts";
import {ApiError, ChallengeResponse} from "../Api/common.ts";
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
import {AnimatedQR} from "../Components/AnimatedQR.tsx";

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
            <TableCell>{userKey.key.authenticator.signCount}</TableCell>
            <TableCell>{userKey.hash}</TableCell>
            <TableCell><AAGUID aaguid={userKey.aaguid}/></TableCell>
            <TableCell>{userKey.created}</TableCell>
            <TableCell>{userKey.lastUsed}</TableCell>
            <TableCell><Button variant="contained" onClick={onDelete}>Delete Key</Button></TableCell>
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
        <img alt="metaImage" src={metadata.metadataStatement.icon ?? `data:${metadata.metadataStatement.Icon.Opaque}`}
             height={32} width={32}/>
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
    setChallenge: (challenge: ChallengeResponse) => void;
    setStartTime: (startTime: Date) => void;
    setResult: (response: BIDResponse | ApiError | null) => void;
}
const UserView = ({name, verify, user, text, setStartTime, setChallenge, setResult}: UserViewProps) => {
    const keys = user.keys.map((userKey, i) => <UserKey key={i} user={user} userKey={userKey}/>);
    const {adminApi, privateApi} = useApi();
    const addKey = () => {
        setResult(null);
        adminApi.addUserKey(name, user.id).then(res => {
            setChallenge(res);
            setStartTime(new Date);
        });
    }

    const authenticate = () => {
        setResult(null);
        privateApi.sign({
            userId: user.id,
            timeout: 120,
            text: text,
            userVerification: verify ? 'required' : 'discouraged'
        }).then(res => {
            setChallenge(res);
            setStartTime(new Date);
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
                <Button variant="contained" onClick={authenticate}>Authenticate</Button>
                <Button variant="contained" onClick={addKey}>Add key</Button>
                <Button variant="contained" onClick={deleteUser}>Delete user</Button>
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


    const [challenge, setChallenge] = useState<ChallengeResponse | null>(null);
    const [startTime, setStartTime] = useState<Date>(new Date());

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
        setResult(null);
        privateApi.sign({
            timeout: 120,
            text: signText,
            userVerification: requireUserVerification ? 'required' : 'discouraged'
        }).then(res => {
            setStartTime(() => new Date);
            setChallenge(() => res);
        });
    }

    const registerUser = () => {
        setResult(null);
        adminApi.registerUser(name).then((res) => {
            setChallenge(() => res);
            setStartTime(() => new Date);
        })
    }

    useEffect(() => {
        if (challenge !== null) {
            const interval = setInterval(() => {
                privateApi.collect(challenge.challenge_id).then((res) => {
                    setResult(res);
                    window.clearInterval(interval);
                    setChallenge(null);
                    adminApi.listUsers().then((res) => {
                        setUsers(res);
                    });
                }).catch(e => {
                    setResult(e);
                    if(e.status !== 'pending' && e.status !== 'viewed') {
                        window.clearInterval(interval);
                        setChallenge(null);
                    }
                });
            }, 500);
            return () => window.clearInterval(interval);
        }
    }, [challenge, privateApi]);

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
                    <Button variant="contained" onClick={registerUser} style={{marginRight: '10px'}}>
                        Create user
                    </Button>
                    <Button variant="contained" onClick={authenticate}>Authenticate any user</Button>
                </FormGroup>
            </FormGroup>
            {challenge !== null &&
                <AnimatedQR challenge={challenge!} startTime={startTime} popUp={true}/>
            }
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
                                                   setChallenge={setChallenge}
                                                   setStartTime={setStartTime}
                                                   setResult={setResult}
                />)}
            </div>
        </div>
    )
}
