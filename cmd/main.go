package main

import (
	"GIN/connections/cache"
	"GIN/connections/database"
	"GIN/connections/migrations"
	"GIN/router"
	"fmt"
	"log"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Unable to load env with godotenv:", err)
	}

	pool := database.Connect()
	redis := cache.Connect()
	migrations.Migrate(pool)

	r := router.Init(pool, redis)

	r.Run(fmt.Sprintf(":%d", 1337))

}

//TODO: считывание метаданных трека с помощью redis очередей
