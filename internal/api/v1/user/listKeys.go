package user

import (
	"database/sql"
	"errors"
	"github.com/gin-gonic/gin"
	"net/http"
	"uyulala/internal/api"
	"uyulala/internal/api/application"
	"uyulala/internal/db/userdb"
)

func listKeys(c *gin.Context) {
	jwt := application.GetCurrentJWT(c)
	subj := jwt.Subject()
	userKeys, err := userdb.GetUserWithKeys(c, subj)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			api.AbortError(c, http.StatusNotFound, "user_not_found", "user not found", nil)
			return
		}
		api.AbortError(c, http.StatusInternalServerError, "internal_error", "internal error", err)
		return
	}
	api.JSONResponse(c, userKeys)
}
