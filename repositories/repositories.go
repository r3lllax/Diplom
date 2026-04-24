package repositories

import (
	errs "GIN/errors"
	"GIN/repositories/repository"
	"context"
	"errors"
	"log"
	"strconv"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type Repositories struct {
	UserRepository       *repository.UserRepository
	SongRepository       *repository.SongRepository
	AuthRepository       *repository.AuthRepository
	ValidationRepository *repository.ValidationRepository
	PlaylistRepository   *repository.PlaylistRepository
	StorageRepository    *repository.StorageRepository
	SessionRepository    *repository.SessionRepository
}

func NewRepositories(db *pgxpool.Pool, redis *redis.Client) *Repositories {
	return &Repositories{
		UserRepository:       repository.NewUserRepository(db),
		SongRepository:       repository.NewSongRepository(db),
		AuthRepository:       repository.NewAuthRepository(db),
		ValidationRepository: repository.NewValidationRepository(db),
		PlaylistRepository:   repository.NewPlaylistRepository(db),
		StorageRepository:    repository.NewStorageRepository(),
		SessionRepository:    repository.NewSessionRepository(redis),
	}
}

func (repos *Repositories) GetTokenVersion(ctx context.Context, userID int) (int, error) {
	var ErrWithCode errs.ErrorWithCode
	exists, err := repos.SessionRepository.TokenVersionExists(ctx, userID)
	if err != nil {
		return 0, errs.ServerError()
	}

	if !exists {
		currentTokenVersion, err := repos.AuthRepository.GetUserTokenVersion(ctx, userID)
		if err != nil {
			if errors.As(err, &ErrWithCode) {
				return 0, err
			}

			return 0, errs.ServerError()
		}
		err = repos.SessionRepository.CreateTokenVersionSession(ctx, userID, currentTokenVersion)
		if err != nil {
			return 0, errs.ServerError()
		}
	}

	value, err := repos.SessionRepository.GetTokenVersionSession(ctx, userID)
	if err != nil {
		return 0, errs.ServerError()
	}

	serverTokenVersion, err := strconv.Atoi(value)
	if err != nil {
		log.Println("CONVERT SERVER TOKEN VERSION ERROR:", err)
		return 0, errs.ServerError()
	}
	return serverTokenVersion, nil
}
