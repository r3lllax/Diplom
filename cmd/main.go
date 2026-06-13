package main

import (
	"GIN/connections/cache"
	"GIN/connections/database"
	"GIN/connections/migrations"
	"GIN/router"
	"fmt"
	"log"
	"os"

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

	// Определяем порт
	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "1337"
	}

	r.Run(fmt.Sprintf(":%s", port))

}

//TODO: считывание метаданных трека с помощью redis очередей
