package cache

import (
	"fmt"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
)

func Connect() *redis.Client {
	REDIS_HOST := os.Getenv("REDIS_HOST")
	REDIS_PORT := os.Getenv("REDIS_PORT")
	connstr := fmt.Sprintf("%s:%s", REDIS_HOST, REDIS_PORT)
	client := redis.NewClient(&redis.Options{
		Addr:            connstr,
		Password:        "",
		DB:              0,
		MinIdleConns:    20,
		PoolSize:        100,
		ConnMaxLifetime: time.Minute * 30,
		PoolTimeout:     2 * time.Second,
	})
	return client
}
