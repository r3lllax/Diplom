package repository

import (
	"GIN/JWT"
	errs "GIN/errors"
	"GIN/internal/env"
	"GIN/keys"
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

type SessionRepository struct {
	client *redis.Client
}

func NewSessionRepository(c *redis.Client) *SessionRepository {
	return &SessionRepository{
		client: c,
	}
}

func (s *SessionRepository) AddAccessTokenToBlackList(ctx context.Context, token string, lifetime time.Duration) error {
	parsedToken, err := JWT.ParseToken(token)
	if err != nil {
		return err
	}
	data, err := JWT.ConvertToken(parsedToken)
	if err != nil {
		return err
	}
	val := fmt.Sprintf("userID:%v,name:%v", data[string(keys.UserIDKey)], data[string(keys.UserName)])

	err = s.client.Set(ctx, fmt.Sprintf("blacklist:%v", token), val, lifetime).Err()
	if err != nil {
		return err
	}
	return nil

}

func (s *SessionRepository) CreateSession(ctx context.Context, token string) error {
	parsedToken, err := JWT.ParseToken(token)
	if err != nil {
		return err
	}
	data, err := JWT.ConvertToken(parsedToken)
	if err != nil {
		return err
	}
	tokenVersion := int(data["tokenVersion"].(float64))

	val := fmt.Sprintf("%s:%v,name:%v,version:%v", keys.UserIDKey, data[string(keys.UserIDKey)], data[string(keys.UserName)], tokenVersion)

	err = s.client.Set(ctx, fmt.Sprintf("refresh:%s", token), val, 30*24*time.Hour).Err()

	if err != nil {
		log.Println("Create session error:", err)
		return err
	}
	return nil

}

func (s *SessionRepository) GetValue(ctx context.Context, token string) (string, error) {
	value, err := s.client.Get(ctx, token).Result()
	if err != nil {
		log.Println("Get session error:", err)
		return "", err
	}
	return value, nil

}

func (s *SessionRepository) Exists(ctx context.Context, token string) (bool, error) {
	exists, err := s.client.Exists(ctx, token).Result()
	if err != nil {
		return false, err
	} else if exists == 0 {
		return false, nil
	} else {
		return true, nil
	}
}

// Чтобы обновить рефреш токен, передавать не просто токен, а "refresh:<your_token>"
// Чтобы поставить новый рефреш токен, передавать новый аргумент также, "refresh:<your_token>"
func (s *SessionRepository) UpdateSession(ctx context.Context, oldToken string, newToken string, expTime time.Duration) error {
	value, err := s.GetValue(ctx, oldToken)
	if err != nil {
		return err
	}

	err = s.client.Set(ctx, newToken, value, expTime).Err()
	if err != nil {
		return err
	}

	err = s.DeleteSession(ctx, oldToken)
	if err != nil {
		return err
	}
	return nil
}

func (s *SessionRepository) DeleteSession(ctx context.Context, token string) error {
	result, err := s.client.Del(ctx, token).Result()
	if err != nil {
		log.Println("Delete session error:", err)
		return err
	}
	if result == 0 {
		log.Println("Delete session error:", err)
		return err
	}

	return nil
}

func (s *SessionRepository) CreateTokenVersionSession(ctx context.Context, userID, tokenVersion int) error {
	_, lifeTimeHours := env.GetTokensLifeTime()
	err := s.client.Set(ctx, fmt.Sprintf("tokenVersions:token_%v", userID), tokenVersion, time.Duration(lifeTimeHours)*24*time.Hour).Err()
	if err != nil {
		log.Println("Create token version session error:", err)
		return err
	}
	return nil

}

func (s *SessionRepository) GetTokenVersionSession(ctx context.Context, userID int) (string, error) {
	value, err := s.client.Get(ctx, fmt.Sprintf("tokenVersions:token_%v", userID)).Result()
	if err != nil {
		if err == redis.Nil {
			return "", nil
		}
		log.Println("Get token version session error:", err)
		return "", err
	}

	return value, nil

}

func (s *SessionRepository) TokenVersionExists(ctx context.Context, userID int) (bool, error) {
	exists, err := s.client.Exists(ctx, fmt.Sprintf("tokenVersions:token_%v", userID)).Result()
	if err != nil {
		log.Println("EXISTS TOKEN VERSION ERROR:", err)
		if err == redis.Nil {
			return false, nil
		}
		return false, errs.ServerError()
	} else if exists == 0 {
		return false, nil
	} else {
		return true, nil
	}
}

func (s *SessionRepository) IncreaseTokenVersion(ctx context.Context, userID int, expTime time.Duration) error {
	value, err := s.GetTokenVersionSession(ctx, userID)
	if err != nil {
		return err
	}
	version, err := strconv.Atoi(value)
	if err != nil {
		log.Println("CONVERTE TOKEN VERSION VALUE ERROR:", err)
		return errs.ServerError()
	}
	err = s.DeleteTokenVersion(ctx, userID)
	if err != nil {
		return err
	}

	version += 1

	err = s.CreateTokenVersionSession(ctx, userID, version)
	if err != nil {
		return err
	}

	return nil
}

func (s *SessionRepository) DeleteTokenVersion(ctx context.Context, userID int) error {
	err := s.DeleteSession(ctx, fmt.Sprintf("tokenVersions:token_%v", userID))
	if err != nil {
		if err == redis.Nil {
			return nil
		}
		log.Println("DELETE TOKEN VERSION FROM REDIS ERROR:", err)
		return errs.ServerError()
	}
	return nil
}

//TODO:написать функцию продления EXPIRE записи о версии токена, и, если запись есть, обновлять при логине и рефреше
