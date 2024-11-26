package v1

import (
	"encoding/gob"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
)

func init() {
	gob.Register(&protocol.URLEncodedBase64{})

	gob.Register(&webauthn.SessionData{})

	gob.Register(&protocol.CredentialCreation{})
	gob.Register(&protocol.CredentialCreationResponse{})

	gob.Register(&protocol.CredentialAssertion{})
	gob.Register(&protocol.CredentialAssertionResponse{})

	gob.Register(&webauthn.Credential{})
	gob.Register(&protocol.ParsedCredentialAssertionData{})
	gob.Register(&protocol.ParsedCredentialCreationData{})
	gob.Register(protocol.CredentialDescriptor{})
}
