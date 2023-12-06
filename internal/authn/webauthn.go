package authn

import (
	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/spf13/viper"
)

func CreateWebauthnConfig() *webauthn.WebAuthn {
	var trueValue = true
	res, err := webauthn.New(&webauthn.Config{
		RPID:                  viper.GetString("webauthn.id"),
		RPDisplayName:         viper.GetString("webauthn.display_name"),
		RPOrigins:             viper.GetStringSlice("webauthn.origins"),
		AttestationPreference: protocol.ConveyancePreference(viper.GetString("webauthn.attestation")),
		AuthenticatorSelection: protocol.AuthenticatorSelection{
			RequireResidentKey: &trueValue,
			ResidentKey:        "required",
			UserVerification:   "required",
		},
		EncodeUserIDAsString: false,
		Debug:                false,
		Timeouts: webauthn.TimeoutsConfig{
			Login: webauthn.TimeoutConfig{
				Enforce:    false,
				Timeout:    0,
				TimeoutUVD: 0,
			},
			Registration: webauthn.TimeoutConfig{
				Enforce:    false,
				Timeout:    0,
				TimeoutUVD: 0,
			},
		},
	})
	if err != nil {
		panic(err)
	}
	return res
}
