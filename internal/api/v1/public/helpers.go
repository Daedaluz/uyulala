package public

import (
	"database/sql"
	"errors"
	"log/slog"
	"net/http"
	"net/url"
	"time"
	"uyulala/internal/api"
	"uyulala/internal/db/challengedb"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/spf13/viper"
)

func parseRedirectURI(vars url.Values) (*url.URL, error) {
	uri := vars.Get("redirect_uri")
	if uri == "" {
		return nil, errors.New("missing redirect_uri")
	}
	parsed, err := url.Parse(uri)
	if err != nil {
		return nil, err
	}
	if parsed.Fragment != "" {
		return nil, errors.New("invalid redirect_uri (Must not contain fragment)")
	}
	return parsed, nil
}

func getVerifiedChallenge(ctx *gin.Context, timeSensitive bool) (*challengedb.Data, bool) {
	tokenString, ok := ctx.GetPostForm("token")
	if !ok {
		api.AbortError(ctx, http.StatusBadRequest, "invalid_request", "Missing 'token' post parameter", nil)
		return nil, false
	}
	claims := &getChallengeClaims{}
	var challenge *challengedb.Data
	_, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (any, error) {
		var err error
		challengeID := claims.QRData
		challenge, err = challengedb.GetChallenge(ctx, challengeID)
		if err != nil {
			return nil, err
		}
		return []byte(challenge.Secret), nil
	}, jwt.WithoutClaimsValidation())
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			api.AbortError(ctx, http.StatusNotFound, "not_found", "Challenge not found", err)
			return nil, false
		}
		slog.Error("signChallengeHandler GetChallenge", "error", err)
		api.AbortError(ctx, http.StatusInternalServerError, "invalid_challenge", "Unexpected error", err)
		return nil, false
	}
	if !claims.Persistent && timeSensitive {
		backendDuration := time.Since(challenge.Created)
		frontendDuration := time.Second * time.Duration(claims.Duration)
		timeDiff := (backendDuration - frontendDuration).Abs()
		if timeDiff > viper.GetDuration("challenge.max_time_diff") {
			api.AbortError(ctx, http.StatusBadRequest, "invalid_request", "Token too old", nil)
			return nil, false
		}
	}

	if !challenge.Validate(ctx) {
		return nil, false
	}
	return challenge, true
}
