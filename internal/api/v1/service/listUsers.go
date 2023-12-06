package service

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"uyulala/internal/api"
	"uyulala/internal/db/userdb"
)

func listUsersHandler(ctx *gin.Context) {
	users, err := userdb.ListUsersWithKeys(ctx)
	if err != nil {
		api.AbortError(ctx, http.StatusInternalServerError, "internal_error", "Internal error", err)
		return
	}
	ctx.JSON(http.StatusOK, users)
}
