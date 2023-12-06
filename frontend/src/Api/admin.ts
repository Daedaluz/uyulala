import {ChallengeResponse, fetchJSON} from "./common.ts";
import {Base64} from "basejs";

export type Key = {
    hash: string
    created: string
    lastUsed: string
    aaguid: string
    key: {
        ID: string
        PublicKey: string
        AttestationType: string
        Authenticator: {
            AAGUID: string
            SignCount: number
            CloneWarning: boolean
            Attachment: string
        }
        Flags: {
            UserPresent: boolean
            UserVerified: boolean
            BackupEligible: boolean
            BackupState: boolean
        }
    }
}

export type UserWithKeys = {
    id: string;
    created: string;
    keys: Key[];
}

type statusResponse = {
    status: string;
}

export class adminApi {
    private readonly url: string;
    private readonly authHeader: string;

    constructor(url: string, clientId: string, secret: string) {
        this.url = url;
        this.authHeader = `Basic ${Base64.encode(new TextEncoder().encode(`${clientId}:${secret}`))}`
    }

    listUsers() {
        return fetchJSON<UserWithKeys[]>(`${this.url}/api/v1/service/list/users`, {
            method: 'GET',
            headers: {
                Authorization: this.authHeader,
                ['Content-Type']: 'application/json'
            },
        })
    }

    registerUser(name: string) {
        return fetchJSON<ChallengeResponse>(`${this.url}/api/v1/service/create/user`, {
            method: 'POST',
            headers: {
                Authorization: this.authHeader,
                ['Content-Type']: 'application/json'
            },
            body: JSON.stringify({suggestedName: name, timeout: 380, redirect: location.toString()})
        })
    }

    addUserKey(name: string, uid: string) {
        return fetchJSON<ChallengeResponse>(`${this.url}/api/v1/service/create/key`, {
            method: 'POST',
            headers: {
                Authorization: this.authHeader,
                ['Content-Type']: 'application/json'
            },
            body: JSON.stringify({suggestedName: name, userId: uid, timeout: 380, redirect: location.toString()})
        })
    }

    deleteUserKey(uid: string, keyHash: string) {
        return fetchJSON<statusResponse>(`${this.url}/api/v1/service/delete/key`, {
            method: 'POST',
            headers: {
                Authorization: this.authHeader,
                ['Content-Type']: 'application/json',
            },
            body: JSON.stringify({userId: uid, keyHash: keyHash})
        })
    }

    deleteUser(uid: string) {
        return fetchJSON<statusResponse>(`${this.url}/api/v1/service/delete/user`, {
            method: 'POST',
            headers: {
                Authorization: this.authHeader,
                ['Content-Type']: 'application/json',
            },
            body: JSON.stringify({userId: uid})
        })
    }
}
