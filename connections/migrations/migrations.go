package migrations

import (
	"fmt"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

func Migrate(pool *pgxpool.Pool) {
	db := stdlib.OpenDBFromPool(pool)
	defer db.Close()

	if err := goose.SetDialect(string(goose.DialectPostgres)); err != nil {
		panic(err)
	}

	if err := goose.Up(db, "./migrations"); err != nil {
		log.Printf("MIGRATIONS ERROR: %v\n", err)
	}

	fmt.Println("migrations successfully executed.")
}
