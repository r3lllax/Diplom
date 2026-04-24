package JWT

import (
	"time"
)

func GetRemainingTime(exp float64) time.Duration {
	expTime := int64(exp)
	now := time.Now().Unix()
	remainingSeconds := expTime - now
	remainingTime := time.Duration(remainingSeconds) * time.Second
	return remainingTime

}
