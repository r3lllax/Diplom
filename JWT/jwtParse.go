package JWT

import (
	errs "GIN/errors"
	"GIN/internal/env"
	"net/http"

	"github.com/golang-jwt/jwt/v5"
)

func ParseToken(tokenString string) (*jwt.Token, error) {
	return jwt.Parse(tokenString, func(t *jwt.Token) (any, error) {
		return []byte(env.GetJWTSecret()), nil
	})
}

func ConvertToken(token *jwt.Token) (jwt.MapClaims, error) {
	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		return claims, nil
	}
	return jwt.MapClaims{}, errs.New(http.StatusUnauthorized, "Не удалось излвечь информацию из токена")
}
