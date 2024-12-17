package public

import (
	"log/slog"
	"net/http"
	"net/url"
	api2 "uyulala/internal/api"
	"uyulala/internal/db/challengedb"

	"github.com/gin-gonic/gin"
)

func rejectChallengeHandler(context *gin.Context) {
	challenge, ok := getVerifiedChallenge(context, false)
	if !ok {
		return
	}
	if err := challengedb.SetChallengeStatus(context, challenge.ID, challengedb.StatusRejected); err != nil {
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
				q.Set("code", challenge.ID)
				if oauthContext.Get("state") != "" {
					q.Set("state", oauthContext.Get("state"))
				}
			} else {
				q.Set("challengeId", challenge.ID)
			}
			q.Set("error", "rejected")
			q.Set("error_description", "User rejected the request")
			r.RawQuery = q.Encode()
			redirectURL = r.String()
		}
	}
	api2.RedirectResponse(context, redirectURL)
}
