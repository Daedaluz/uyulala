package public

import (
	"net/http"
	"uyulala/internal/api"
	"uyulala/internal/db/appdb"
	"uyulala/internal/db/challengedb"

	"github.com/gin-gonic/gin"
	"github.com/go-webauthn/webauthn/protocol"
	"github.com/golang-jwt/jwt/v5"
)

type getChallengeClaims struct {
	QRData     string `json:"challenge_id"`
	Duration   int    `json:"duration"`
	Persistent bool   `json:"persistent"`
}

func (g getChallengeClaims) GetExpirationTime() (*jwt.NumericDate, error) {
	return nil, nil
}

func (g getChallengeClaims) GetIssuedAt() (*jwt.NumericDate, error) {
	return nil, nil
}

func (g getChallengeClaims) GetNotBefore() (*jwt.NumericDate, error) {
	return nil, nil
}

func (g getChallengeClaims) GetIssuer() (string, error) {
	return "", nil
}

func (g getChallengeClaims) GetSubject() (string, error) {
	return "", nil
}

func (g getChallengeClaims) GetAudience() (jwt.ClaimStrings, error) {
	return nil, nil
}

func getChallengeHandlerPost(ctx *gin.Context) {
	var err error
	data, ok := getVerifiedChallenge(ctx, true)
	if !ok {
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
	if err := challengedb.SetChallengeStatus(ctx, data.ID, challengedb.StatusViewed); err != nil {
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
			"nonce": data.Nonce,
			"text":  data.SignatureText,
			"data":  data.SignatureData,
		}
	}
	ctx.JSON(200, res)
}
