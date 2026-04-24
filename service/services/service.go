package service

import (
	"GIN/repositories"
	"GIN/valitadion"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type baseService struct {
	Validate     *valitadion.StructValidator
	repositories *repositories.Repositories
}

func NewBaseService(db *pgxpool.Pool, redis *redis.Client) *baseService {
	return &baseService{
		Validate:     valitadion.NewValidator(db),
		repositories: repositories.NewRepositories(db, redis),
	}
}
