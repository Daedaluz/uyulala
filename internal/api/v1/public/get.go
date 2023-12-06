package public

import (
	"github.com/gin-gonic/gin"
	"github.com/go-webauthn/webauthn/protocol"
	"net/http"
	"uyulala/internal/api"
	"uyulala/internal/db/appdb"
	"uyulala/internal/db/challengedb"
)

func getChallengeHandler(ctx *gin.Context) {
	challengeID := ctx.Param("id")
	data, err := challengedb.GetChallenge(ctx, challengeID)
	if err != nil {
		api.AbortError(ctx, http.StatusNotFound, "no_challenge", "Challenge wasn't found", err)
		return
	}
	if !data.Validate(ctx) {
		return
	}

	var challengeRes any
	switch data.Type {
	case "webauthn.create":
		challenge := &protocol.CredentialCreation{}
		err = data.Expand(challenge, nil)
		challengeRes = challenge.Response
	case "webauthn.get":
		challenge := &protocol.CredentialAssertion{}
		err = data.Expand(challenge, nil)
		challengeRes = challenge.Response
	}
	if err != nil {
		api.AbortError(ctx, http.StatusInternalServerError, "internal_error", "Unexpected error", err)
		return
	}
	if err := challengedb.SetChallengeStatus(ctx, challengeID, challengedb.StatusViewed); err != nil {
		api.AbortError(ctx, http.StatusInternalServerError, "internal_error", "Unexpected error", err)
		return
	}
	res := gin.H{"type": data.Type, "publicKey": challengeRes, "expire": data.Expire.Unix()}

	app, _ := appdb.GetApplication(ctx, data.AppID)
	if app != nil {
		res["app"] = app
	}

	if data.SignatureText != "" {
		res["signData"] = gin.H{
			"text": data.SignatureText,
			"data": data.SignatureData,
		}
	}
	ctx.JSON(200, res)
}
