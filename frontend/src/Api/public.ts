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

    getChallengePost(token: string) {
        return fetchJSON<JSONChallenge>(`${this.url}/api/v1/challenge`, {
            method: "POST",
            headers: {
                'Content-Type': 'application/x-www-form-urlencoded'
            },
            body: new URLSearchParams([["token", token]]).toString()
        }).then(challenge => {
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

    sign(id: string, data: Credential) {
        const body = authnEncode(data)
        return fetchJSON<RedirectResponse>(`${this.url}/api/v1/challenge`, {
            method: "PUT",
            body: new URLSearchParams([["token", id], ["response", JSON.stringify(body)]]).toString(),
        });
    }

    reject(challenge: string) {
        return fetchJSON<RedirectResponse>(`${this.url}/api/v1/challenge`, {
            headers: {
                'Content-Type': 'application/x-www-form-urlencoded'
            },
            method: "DELETE",
            body: new URLSearchParams([["token", challenge]]).toString()
        });
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
}
