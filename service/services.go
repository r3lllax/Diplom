package services

import (
	service "GIN/service/services"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type Services struct {
	UserService     *service.UserService
	AuthService     *service.AuthService
	PlaylistService *service.PlaylistService
	SongService     *service.SongService
	StorageService  *service.StorageService
}

func NewServices(db *pgxpool.Pool, redis *redis.Client) *Services {

	base := service.NewBaseService(db, redis)

	return &Services{
		UserService:     service.NewUserService(base),
		AuthService:     service.NewAuthService(base),
		PlaylistService: service.NewPlaylistService(base),
		SongService:     service.NewSongService(base),
		StorageService:  service.NewStorageService(base),
	}
}
