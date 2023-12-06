import {Base64} from "basejs";

export type ChallengeResponse = {
    challenge_id: string;
}

export type RedirectResponse = {
    redirect: string;
}

export type ApiError = {
    code: number;
    msg: string;
    error: string;
    technicalMsg?: string;
}

export function fetchJSON<T>(url: string, opts?: RequestInit): Promise<T> {
    if (typeof opts?.body === "object") {
        opts.body = JSON.stringify(opts.body);
    }
    return fetch(url, opts).then(async res => {
        if (res.status !== 200) {
            throw await res.json() as ApiError;
        }
        return await res.json() as T;
    });
}

export function authnEncode(cred: any): any {
    // Challenge
    const res = {} as any
    if (cred.publicKey) {
        res.publicKey = {}
        if (cred.publicKey.rpId) {
            res.publicKey.rpId = cred.publicKey.rpId;
        }
        if (cred.publicKey.rp) {
            res.publicKey.rp = {...cred.publicKey.rp};
        }
        if (cred.publicKey.pubKeyCredParams) {
            res.publicKey.pubKeyCredParams = [...cred.publicKey.pubKeyCredParams];
        }
        if (cred.publicKey.authenticatorSelection) {
            res.publicKey.authenticatorSelection = {...cred.publicKey.authenticatorSelection};
        }
        if (cred.publicKey.attestation) {
            res.publicKey.attestation = cred.publicKey.attestation;
        }
        if (cred.publicKey.timeout) {
            res.publicKey.timeout = cred.publicKey.timeout;
        }
        if (cred.publicKey.userVerification) {
            res.publicKey.userVerification = cred.publicKey.userVerification;
        }
        res.publicKey.challenge = Base64.urlEncode(cred.publicKey.challenge);
        if (res.publicKey.user) {
            res.publicKey.user = {...cred.publicKey.user};
            res.publicKey.user.id = Base64.urlEncode(cred.publicKey.user.id);
        }
        if (cred.publicKey.excludeCredentials) {
            res.publicKey.excludeCredentials = cred.publicKey.excludeCredentials ? cred.publicKey.excludeCredentials.map((c: any) => {
                const x = {...c}
                x.id = Base64.urlEncode(c.id);
                return x;
            }) : undefined;
        }
        if (cred.publicKey.allowCredentials) {
            res.publicKey.allowCredentials = cred.publicKey.allowCredentials ? cred.publicKey.allowCredentials.map((c: any) => {
                c.id = Base64.urlEncode(c.id);
                return {...c};
            }) : undefined;
        }
    }

    // Response
    if (cred.response) {
        res.response = {}
        if (cred.rawId) {
            res.rawId = Base64.urlEncode(cred.rawId);
        }
        if (cred.response) {
            if (cred.response.clientDataJSON) {
                res.response.clientDataJSON = Base64.urlEncode(cred.response.clientDataJSON);
            }
            if (cred.response.attestationObject) {
                res.response.attestationObject = Base64.urlEncode(cred.response.attestationObject);
            }
            if (cred.response.authenticatorData) {
                res.response.authenticatorData = Base64.urlEncode(cred.response.authenticatorData);
            }
            if (cred.response.signature) {
                res.response.signature = Base64.urlEncode(cred.response.signature);
            }
            if (cred.response.userHandle) {
                res.response.userHandle = Base64.urlEncode(cred.response.userHandle);
            }
        }
        if (cred.authenticatorAttachment) {
            res.authenticatorAttachment = cred.authenticatorAttachment;
        }
        if (cred.type) {
            res.type = cred.type;
        }
        if (cred.id) {
            res.id = cred.id;
        }
    }
    return res;
}


export function authnDecode(json: any): any {
    // Challenge
    const res = {} as any
    if (json.publicKey) {
        res.publicKey = {}
        res.publicKey.challenge = Base64.urlDecode(json.publicKey.challenge);
        if (json.publicKey.rpId) {
            res.publicKey.rpId = json.publicKey.rpId;
        }
        if (json.publicKey.rp) {
            res.publicKey.rp = {...json.publicKey.rp};
        }
        if (json.publicKey.pubKeyCredParams) {
            res.publicKey.pubKeyCredParams = [...json.publicKey.pubKeyCredParams];
        }
        if (json.publicKey.authenticatorSelection) {
            res.publicKey.authenticatorSelection = {...json.publicKey.authenticatorSelection};
        }
        if (json.publicKey.attestation) {
            res.publicKey.attestation = json.publicKey.attestation;
        }
        if (json.publicKey.timeout) {
            res.publicKey.timeout = json.publicKey.timeout;
        }
        if (json.publicKey.userVerification) {
            res.publicKey.userVerification = json.publicKey.userVerification;
        }
        if (json.publicKey.user) {
            res.publicKey.user = {...json.publicKey.user};
            res.publicKey.user.id = Base64.urlDecode(json.publicKey.user.id);
        }
        if (json.publicKey.excludeCredentials) {
            res.publicKey.excludeCredentials = json.publicKey.excludeCredentials ? json.publicKey.excludeCredentials.map((c: any) => {
                const x = {...c}
                x.id = Base64.urlDecode(c.id);
                return x;
            }) : undefined;
        }
        if (json.publicKey.allowCredentials) {
            res.publicKey.allowCredentials = json.publicKey.allowCredentials ? json.publicKey.allowCredentials.map((c: any) => {
                c.id = Base64.urlDecode(c.id);
                return {...c};
            }) : undefined;
        }
    }

    // Response
    if (json.response) {
        res.response = {}
        if (json.rawId) {
            res.rawId = Base64.urlDecode(json.rawId);
        }
        if (json.response) {
            if (json.response.clientDataJSON) {
                res.response.clientDataJSON = Base64.urlDecode(json.response.clientDataJSON);
            }
            if (json.response.attestationObject) {
                res.response.attestationObject = Base64.urlDecode(json.response.attestationObject);
            }
            if (json.response.authenticatorData) {
                res.response.authenticatorData = Base64.urlDecode(json.response.authenticatorData);
            }
            if (json.response.signature) {
                res.response.signature = Base64.urlDecode(json.response.signature);
            }
            if (json.response.userHandle) {
                res.response.userHandle = Base64.urlDecode(json.response.userHandle);
            }
        }
        if (json.authenticatorAttachment) {
            res.authenticatorAttachment = json.authenticatorAttachment;
        }
        if (json.type) {
            res.type = json.type;
        }
        if (json.id) {
            res.id = json.id;
        }
    }
    return res;
}