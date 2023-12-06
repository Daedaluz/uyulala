import {ChallengeResponse, authnEncode, fetchJSON, authnDecode, RedirectResponse} from "./common.ts";

export type SignData = {
    text: string;
    data: ArrayBuffer;
}

export type App = {
    admin: boolean;
    description: string;
    icon: string;
    name: string;
    publicKey: string;
}

type JSONChallenge = {
    type: string;
    publicKey: any;
    app: App;
    signData?: SignData;
}

export class ICredentialCreationOptions implements CredentialCreationOptions {
    public publicKey: PublicKeyCredentialCreationOptions;
    public app: App;
    public signData?: SignData;

    constructor(publicKey: PublicKeyCredentialCreationOptions, app: App, signData?: SignData) {
        this.publicKey = publicKey;
        this.app = app;
        this.signData = signData;
    }
}



export type MetadataStatement = {
    description: string
    protocolFamily: string
    authenticatorVersion: number
    icon: string
}

export type StatusReport = {
    status: string
    effectiveDate: string
    authenticatorVersion: number
    certificate: string
    url: string
    certificationDescriptor: string
    certificationNumber: string
    certificationPolicyVersion: string
    certificationRequirementsVersion: string
}

export type Metadata = {
    aaguid: string
    metadataStatement: MetadataStatement
    statusReports: StatusReport[]
}


export class ICredentialRequestOptions implements CredentialRequestOptions {
    public publicKey: PublicKeyCredentialRequestOptions;
    public app: App;
    public signData?: SignData;

    constructor(publicKey: PublicKeyCredentialRequestOptions, app: App, signData?: SignData) {
        this.publicKey = publicKey;
        this.app = app;
        this.signData = signData;
    }
}

export class publicApi {
    private readonly url: string;

    constructor(url: string) {
        this.url = url;
    }

    getChallenge(id: string) {
        return fetchJSON<JSONChallenge>(`${this.url}/api/v1/challenge/${id}`).then(challenge => {
            const pubKey = authnDecode(challenge);
            switch (challenge.type) {
                case "webauthn.create":
                    return new ICredentialCreationOptions(pubKey.publicKey, challenge.app, challenge.signData);
                case "webauthn.get":
                    return new ICredentialRequestOptions(pubKey.publicKey, challenge.app, challenge.signData);
                default:
                    throw new Error("Invalid challenge type");
            }
        })
    }

    createOAuth2Challenge(urlParameters: URLSearchParams) {
        return fetchJSON<ChallengeResponse>(`${this.url}/api/v1/oauth2`, {
            method: "POST",
            headers: {
                'Content-Type': 'application/x-www-form-urlencoded'
            },
            body: urlParameters.toString(),
        })
    }

    sign(id: string, data: Credential) {
        const body = authnEncode(data)
        return fetchJSON<RedirectResponse>(`${this.url}/api/v1/challenge/${id}`, {
            method: "POST",
            body: JSON.stringify(body)
        });
    }

    reject(id: string) {
        return fetchJSON<RedirectResponse>(`${this.url}/api/v1/challenge/${id}`, {
            method: "DELETE"
        });
    }

    getAuthenticatorDescriptor(aaguid: string) {
        return fetchJSON<Metadata>(`${this.url}/api/v1/aaguid/${aaguid}`)
    }
}