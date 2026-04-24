package middleware

import (
	"GIN/JWT"
	errs "GIN/errors"
	"GIN/keys"
	"GIN/repositories"
	"errors"
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func Authorize(repos *repositories.Repositories) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		tokenStr := ctx.Request.Header.Get("Authorization")
		tokenInfo := strings.Split(tokenStr, " ")
		if len(tokenInfo) < 2 || tokenInfo[0] != "Bearer" {
			errs.ThrowUnauthorizedError(ctx)
			ctx.Abort()
			return
		}

		if exists, _ := repos.SessionRepository.Exists(ctx.Request.Context(), fmt.Sprintf("blacklist:%s", tokenInfo[1])); exists {
			errs.ThrowUnauthorizedError(ctx)
			ctx.Abort()
			return
		}

		token, err := JWT.ParseToken(tokenInfo[1])
		if err != nil {
			if errors.Is(err, jwt.ErrTokenExpired) {
				errs.ThrowUnauthorizedError(ctx)
				ctx.Abort()
				return
			}

			errs.ThrowUnauthorizedError(ctx)
			ctx.Abort()
			return

		}
		info, err := JWT.ConvertToken(token)
		if err != nil {
			errs.ThrowUnauthorizedError(ctx)
			ctx.Abort()
			return
		}

		tokenVersion := int(info["tokenVersion"].(float64))
		userID := int(info[string(keys.UserIDKey)].(float64))
		userName := info[string(keys.UserName)]

		var ErrWithCode errs.ErrorWithCode

		serverTokenVersion, err := repos.GetTokenVersion(ctx, userID)
		if err != nil {
			if errors.As(err, &ErrWithCode) {
				errs.ThrowError(ctx, ErrWithCode.Code, ErrWithCode.Message)
				ctx.Abort()
				return
			}
		}

		if tokenVersion < serverTokenVersion {
			errs.ThrowUnauthorizedError(ctx)
			ctx.Abort()
			return
		}
		ctx.Set(keys.UserIDKey, userID)
		ctx.Set(keys.UserName, userName)
		ctx.Next()
	}
}
