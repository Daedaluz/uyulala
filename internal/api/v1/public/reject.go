package public

import (
	"database/sql"
	"errors"
	"log/slog"
	"net/http"
	"net/url"
	api2 "uyulala/internal/api"
	"uyulala/internal/db/challengedb"

	"github.com/gin-gonic/gin"
)

func rejectChallengeHandler(context *gin.Context) {
	challengeID := context.Param("id")
	challenge, err := challengedb.GetChallenge(context, challengeID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			api2.AbortError(context, http.StatusNotFound, "not_found", "Challenge not found", err)
			return
		}
		slog.Error("signChallengeHandler GetChallenge", "error", err)
		api2.AbortError(context, http.StatusInternalServerError, "invalid_challenge", "Unexpected error", err)
		return
	}
	if !challenge.Validate(context) {
		return
	}
	if err := challengedb.SetChallengeStatus(context, challengeID, challengedb.StatusRejected); err != nil {
		slog.Error("signChallengeHandler SetChallengeStatus", "error", err)
		api2.AbortError(context, http.StatusInternalServerError, "invalid_challenge", "Unexpected error", err)
		return
	}
	redirectURL := ""
	if challenge.RedirectURL != "" {
		redirectURL = challenge.RedirectURL
		if r, err := url.Parse(challenge.RedirectURL); err == nil {
			q := r.Query()
			if oauthContext, err := url.ParseQuery(challenge.OAuth2Context); err == nil && len(oauthContext) > 0 {
				q.Set("code", challengeID)
				if oauthContext.Get("state") != "" {
					q.Set("state", oauthContext.Get("state"))
				}
			} else {
				q.Set("challengeId", challengeID)
			}
			q.Set("error", "rejected")
			q.Set("error_description", "User rejected the request")
			r.RawQuery = q.Encode()
			redirectURL = r.String()
		}
	}
	api2.RedirectResponse(context, redirectURL)
}
