package JWT

import (
	"GIN/internal/env"
	"GIN/keys"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func GenerateToken(UserID, tokenVersion int, UserName string, lifeTime time.Duration) (string, error) {
	claims := jwt.MapClaims{
		string(keys.UserIDKey): UserID,
		string(keys.UserName):  UserName,
		"tokenVersion":         tokenVersion,
		"exp":                  time.Now().Add(lifeTime).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(env.GetJWTSecret()))
}
