package service

import (
	"net/http"
	"uyulala/internal/api"
	"uyulala/internal/db/userdb"

	"github.com/gin-gonic/gin"
)

func listUsersHandler(ctx *gin.Context) {
	users, err := userdb.ListUsersWithKeys(ctx)
	if err != nil {
		api.AbortError(ctx, http.StatusInternalServerError, "internal_error", "Internal error", err)
		return
	}
	ctx.JSON(http.StatusOK, users)
}
