import {ChallengeResponse, fetchJSON} from "./common.ts";
import {Base64} from "basejs";
import {SignData} from "./public.ts";


type SignRequest = {
    userId?: string;
    userVerification?: "required" | "preferred" | "discouraged";
    text?: string;
    data?: ArrayBuffer | string;
    timeout?: number;
    redirect?: string;
};

type OAuth2Response = {
    access_token: string;
    token_type: string;
    id_token: string;
};

export type Assertion = {}

export type BIDResponse = {
    challengeId: string;
    userId: string;
    signatureData: SignData;
    assertion: any;
    signature: any;
    credential: any;
    signed: string;
    status: string;
}

export type MetadataStatementIcon = {
    Scheme: string
    Opaque: string
}

export type MetadataStatement = {
    description: string
    protocolFamily: string
    authenticatorVersion: number
    icon?: string
    Icon: MetadataStatementIcon
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


export class privateApi {
    private readonly url: string;
    private readonly clientId: string;
    private readonly secret: string;
    private readonly authHeader: string;

    constructor(url: string, clientId: string, secret: string) {
        this.url = url;
        this.clientId = clientId;
        this.secret = secret;
        this.authHeader = `Basic ${Base64.encode(new TextEncoder().encode(`${clientId}:${secret}`))}`
    }

    sign(req?: SignRequest) {
        if (req === undefined) {
            req = {};
        }
        if (req?.data) {
            if (typeof req.data === "string") {
                req.data = new TextEncoder().encode(req.data);
            }
            req.data = Base64.urlEncode(req.data);
        }
        return fetchJSON<ChallengeResponse>(`${this.url}/api/v1/sign`, {
            method: 'POST',
            headers: {
                Authorization: this.authHeader,
                ['Content-Type']: 'application/json'
            },
            body: JSON.stringify(req)
        })
    }

    collect(challengeId: string) {
        const body = {
            challengeId: challengeId
        }
        return fetchJSON<BIDResponse>(`${this.url}/api/v1/collect`, {
            method: 'POST',
            headers: {
                Authorization: this.authHeader,
                ['Content-Type']: 'application/json'
            },
            body: JSON.stringify(body)
        })
    }

    OAuth2Exchange(challengeId: string, pkce?: string) {
        const hdr = {
            ['Content-Type']: 'application/json'
        } as any;
        const form = new URLSearchParams();
        form.set('client_id', this.clientId);
        form.set('code', challengeId);
        form.set('grant_type', 'authorization_code');
        if (pkce) {
            form.set('code_verifier', pkce);
        } else {
            form.set('client_secret', this.secret);
        }
        return fetchJSON<OAuth2Response>(`${this.url}/api/v1/collect`, {
            method: 'GET',
            headers: hdr,
            body: form.toString()
        })
    }

    getAuthenticatorDescriptor(aaguid: string) {
        return fetchJSON<Metadata>(`${this.url}/api/v1/mds/${aaguid}`, {
            headers: {
                Authorization: this.authHeader,
                ['Content-Type']: 'application/json'
            },
        })
    }
}
