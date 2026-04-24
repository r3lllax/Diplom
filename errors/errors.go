package errs

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type ErrorWithCode struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (e ErrorWithCode) Error() string {
	return e.Message
}

func New(code int, message string) ErrorWithCode {
	return ErrorWithCode{
		Code:    code,
		Message: message,
	}
}

func ThrowError(ctx *gin.Context, code int, messge string) {
	ctx.JSON(code, gin.H{
		"code":    code,
		"message": messge,
	})
}
func ThrowServerError(ctx *gin.Context) {
	ctx.JSON(http.StatusInternalServerError, gin.H{
		"code":    http.StatusInternalServerError,
		"message": "Ошибка сервера",
	})
}

func ServerError() error {
	return New(http.StatusInternalServerError, "Ошибка сервера")
}
func ThrowValidationErrors(ctx *gin.Context, errors []byte) {
	ctx.Writer.Header().Add("Content-Type", "application/json")
	ctx.Writer.WriteHeader(http.StatusUnprocessableEntity)
	ctx.Writer.Write(errors)
}

func ThrowUnauthorizedError(ctx *gin.Context) {
	ctx.JSON(http.StatusUnauthorized, gin.H{
		"code":    http.StatusUnauthorized,
		"message": "Ошибка авторизации",
	})
}
func UnauthorizedError() error {
	return New(http.StatusUnauthorized, "Ошибка авторизации")
}
