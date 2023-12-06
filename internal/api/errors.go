package api

import (
	"github.com/gin-gonic/gin"
)

type Error struct {
	Code             int    `json:"code"`
	Message          string `json:"msg"`
	Err              string `json:"error"`
	TechnicalMessage string `json:"technicalMsg"`
}

func (e Error) Error() string {
	return e.Message
}

func AbortError(ctx *gin.Context, code int, errorType, msg string, err error) {
	e := &Error{
		Code:    code,
		Err:     errorType,
		Message: msg,
	}
	if err != nil {
		e.TechnicalMessage = err.Error()
	}
	ctx.AbortWithStatusJSON(code, e)
}
