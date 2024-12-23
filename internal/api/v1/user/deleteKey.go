package user

import (
	"database/sql"
	"errors"
	"net/http"
	"uyulala/internal/api"
	"uyulala/internal/api/application"
	"uyulala/internal/db/userdb"

	"github.com/gin-gonic/gin"
)

func deleteKey(c *gin.Context) {
	jwt := application.GetCurrentJWT(c)
	subj := jwt.Subject()
	keyToDelete := c.Param("key")
	userKeys, err := userdb.GetUserWithKeys(c, subj)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			api.AbortError(c, http.StatusNotFound, "user_not_found", "user not found", nil)
			return
		}
		api.AbortError(c, http.StatusInternalServerError, "internal_error", "internal error", err)
		return
	}
	if len(userKeys.Credentials) == 1 {
		api.AbortError(c, http.StatusBadRequest, "cannot_delete_last_key", "cannot delete last key", nil)
		return
	}
	var key *userdb.UserKey
	for i := range userKeys.Credentials {
		if userKeys.Credentials[i].Hash == keyToDelete {
			key = &userKeys.Credentials[i]
			break
		}
	}
	if key == nil {
		api.AbortError(c, http.StatusNotFound, "key_not_found", "key not found", nil)
		return
	}
	if err := userdb.DeleteUserKey(c, subj, key.Hash); err != nil {
		api.AbortError(c, http.StatusInternalServerError, "internal_error", "internal error", err)
		return
	}
	api.DeletedResponse(c)
}
