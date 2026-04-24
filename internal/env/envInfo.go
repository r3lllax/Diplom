package env

import (
	"os"
	"strconv"
)

const (
	reserveSecret string = "ResberaStedwJ2"
)

func GetJWTSecret() string {
	secret := os.Getenv("JWT_SECRET")
	if len(secret) < 1 {
		return reserveSecret
	}
	return secret
}

func GetTokensLifeTime() (int, int) {
	accessTokenLifetimeInMinutes, err := strconv.Atoi(os.Getenv("JWT_ACCESS_TOKEN_MINUTE_TIME"))
	if err != nil {
		accessTokenLifetimeInMinutes = 15
	}
	refreshTokenLifetimeInHours, err := strconv.Atoi(os.Getenv("JWT_REFRESH_TOKEN_HOURS_TIME"))
	if err != nil {
		refreshTokenLifetimeInHours = 30
	}
	return accessTokenLifetimeInMinutes, refreshTokenLifetimeInHours
}
func GetMaxFileSize() int {
	str := os.Getenv("MAX_FILE_SIZE_MB")
	maxSize, err := strconv.Atoi(str)
	if err != nil {
		maxSize = 32
	}
	return maxSize << 20
}

func GetPasswordHashCost() int {
	str := os.Getenv("PASSWORD_HASH_COST")
	cost, err := strconv.Atoi(str)
	if err != nil {
		cost = 12
	}
	return cost
}
