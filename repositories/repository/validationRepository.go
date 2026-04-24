package repository

import "github.com/jackc/pgx/v5/pgxpool"

type ValidationRepository struct {
	db *pgxpool.Pool
}

func NewValidationRepository(dbase *pgxpool.Pool) *ValidationRepository {
	return &ValidationRepository{
		db: dbase,
	}
}
