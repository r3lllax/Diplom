package repository

import (
	errs "GIN/errors"
	"GIN/model"
	"context"
	"errors"
	"log"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AuthRepository struct {
	db *pgxpool.Pool
}

func NewAuthRepository(dbase *pgxpool.Pool) *AuthRepository {
	return &AuthRepository{
		db: dbase,
	}
}
func (r *AuthRepository) Registrate(ctx context.Context, user model.User) (int, error) {
	var id int
	err := r.db.QueryRow(ctx, "insert into users (name,email,password,is_private,photo_file,registrated_at) values($1,$2,$3,$4,$5,$6) returning id",
		user.Name, user.Email, user.Password, user.Is_private, user.Photo_file, user.Registrated_at).Scan(&id)
	if err != nil {
		log.Println("CREATE USER ERROR:", err)
		return 0, errs.ServerError()
	}
	return id, nil
}
func (r *AuthRepository) TokenVersionRowExists(ctx context.Context, userID int) bool {
	var res bool
	err := r.db.QueryRow(ctx, "select exists(select 1 from users_tokens_versions where user_id = $1)", userID).Scan(&res)
	if err != nil {
		log.Println("GET TOKEN VERSION ROW EXISTS ERROR:", err)
		return false
	}
	return res
}

func (r *AuthRepository) CreateTokenVersionRow(ctx context.Context, userID int) error {
	_, err := r.db.Exec(ctx, "insert into users_tokens_versions (user_id,token_version) values ($1,$2)", userID, 1)
	if err != nil {
		log.Println("CREATE TOKEN VERSION ROW ERROR:", err)
		return errs.ServerError()
	}
	return nil
}

func (r *AuthRepository) IncreaseTokenVersion(ctx context.Context, userID int) error {
	_, err := r.db.Exec(ctx, "update users_tokens_versions set token_version = token_version + 1 where id = $1", userID)
	if err != nil {
		log.Println("INCREASE TOKEN VERSION ERROR:", err)
		return errs.ServerError()
	}
	return nil
}

func (r *AuthRepository) GetUserTokenVersion(ctx context.Context, userID int) (int, error) {
	var version int
	err := r.db.QueryRow(ctx, "select token_version from users_tokens_versions where user_id = $1", userID).Scan(&version)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, errs.UnauthorizedError()
		}
		log.Println("GET USER TOKEN VERSION FROM DB ERROR:", err)
		return 0, errs.ServerError()
	}
	return version, nil
}
