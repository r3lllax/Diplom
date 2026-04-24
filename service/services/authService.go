package service

import (
	"GIN/JWT"
	errs "GIN/errors"
	"GIN/internal/env"
	"GIN/internal/files"
	"GIN/model"
	"GIN/tdo/request"
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	baseService
}

func NewAuthService(BService *baseService) *AuthService {
	return &AuthService{
		baseService: *BService,
	}
}

func (s *AuthService) Login(ctx context.Context, request request.Login) (access, refresh string, err error) {

	targetUser, err := s.repositories.UserRepository.GetUserByEmail(ctx, request.Email)
	if err != nil {
		return "", "", err
	}

	err = bcrypt.CompareHashAndPassword([]byte(targetUser.Password), []byte(request.Password))
	if err != nil {
		return "", "", errs.New(http.StatusUnauthorized, "неправильный логин или пароль")
	}

	exists, err := s.repositories.SessionRepository.TokenVersionExists(ctx, targetUser.Id)
	if err != nil {
		return "", "", err
	}
	var currentTokenVersion int

	if !exists {
		currentTokenVersion, err = s.repositories.AuthRepository.GetUserTokenVersion(ctx, targetUser.Id)
		if err != nil {
			return "", "", err
		}
		err = s.repositories.SessionRepository.CreateTokenVersionSession(ctx, targetUser.Id, currentTokenVersion)
		if err != nil {
			return "", "", err
		}
	}
	strVersion, err := s.repositories.SessionRepository.GetTokenVersionSession(ctx, targetUser.Id)
	if err != nil {
		return "", "", err
	}

	intVersion, err := strconv.Atoi(strVersion)
	if err != nil {
		return "", "", errs.ServerError()
	}
	currentTokenVersion = intVersion

	accessLifeTimeMinutes, refreshLifeTimeHours := env.GetTokensLifeTime()

	accessToken, err := JWT.GenerateToken(targetUser.Id, currentTokenVersion, targetUser.Name, time.Duration(accessLifeTimeMinutes)*time.Minute)
	if err != nil {
		log.Println("Generate Access token error:", err)
		return "", "", errs.ServerError()
	}

	refreshToken, err := JWT.GenerateToken(targetUser.Id, currentTokenVersion, targetUser.Name, time.Duration(refreshLifeTimeHours)*time.Hour)
	if err != nil {
		log.Println("Generate Refresh token error:", err)
		return "", "", errs.ServerError()
	}
	err = s.repositories.SessionRepository.CreateSession(ctx, refreshToken)
	if err != nil {
		log.Println("Create session error:", err)
		return "", "", errs.ServerError()
	}
	return accessToken, refreshToken, nil

}
func (s *AuthService) Registrate(ctx *gin.Context, request request.Registration) error {
	dst := ""
	if request.Photo != nil {
		dst = files.GenerateFilePath(request.Photo)
		err := s.baseService.repositories.StorageRepository.UploadFile(request.Photo, dst, ctx)
		if err != nil {
			return err
		}
		dst = dst[1:]
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(request.Password), env.GetPasswordHashCost())
	if err != nil {
		return errs.ServerError()
	}

	user := model.NewUser(0, request.Name, request.Email, string(hashedPassword), false, dst, time.Now())
	id, err := s.repositories.AuthRepository.Registrate(ctx.Request.Context(), *user)
	if err != nil {
		return err
	}
	err = s.repositories.AuthRepository.CreateTokenVersionRow(ctx, id)
	if err != nil {
		return err
	}

	return nil
}
func (s *AuthService) Refresh(ctx context.Context, refreshToken string, userID int) (string, string, error) {

	refreshTokenJWT, err := JWT.ParseToken(refreshToken)
	if err != nil {
		return "", "", errs.UnauthorizedError()
	}
	claims, err := JWT.ConvertToken(refreshTokenJWT)
	if err != nil {
		return "", "", errs.UnauthorizedError()
	}
	refreshTokenVersion := int(claims["tokenVersion"].(float64))

	serverTokenVersion, err := s.repositories.GetTokenVersion(ctx, userID)
	if err != nil {
		return "", "", err
	}

	if refreshTokenVersion < serverTokenVersion {
		return "", "", errs.UnauthorizedError()
	}
	exists, err := s.repositories.SessionRepository.Exists(ctx, fmt.Sprintf("refresh:%s", refreshToken))
	if err != nil || !exists {
		log.Println("EXISTS SESSION CHECK ERROR:", err)
		return "", "", errs.UnauthorizedError()
	}
	value, err := s.repositories.SessionRepository.GetValue(ctx, fmt.Sprintf("refresh:%s", refreshToken))
	if err != nil {
		return "", "", errs.ServerError()
	}
	parts := strings.Split(value, ",")

	userName := strings.Split(parts[1], ":")[1]

	accessLifeTimeMinutes, refreshLifeTimeHours := env.GetTokensLifeTime()

	tokenVersion, err := s.repositories.SessionRepository.GetTokenVersionSession(ctx, userID)
	if err != nil {
		return "", "", err
	}
	tokenInt, err := strconv.Atoi(tokenVersion)
	if err != nil {
		return "", "", errs.ServerError()
	}

	NewAccessToken, err := JWT.GenerateToken(userID, tokenInt, userName, time.Duration(accessLifeTimeMinutes)*time.Minute)
	if err != nil {
		log.Println("Generate Access token error:", err)
		return "", "", errs.ServerError()
	}

	newRefreshToken, err := JWT.GenerateToken(userID, tokenInt, userName, time.Duration(refreshLifeTimeHours)*time.Hour)
	if err != nil {
		log.Println("Generate Refresh token error:", err)
		return "", "", errs.ServerError()
	}

	err = s.repositories.SessionRepository.UpdateSession(ctx, fmt.Sprintf("refresh:%s", refreshToken), fmt.Sprintf("refresh:%s", newRefreshToken), time.Duration(refreshLifeTimeHours)*time.Hour)
	if err != nil {
		log.Println("UPDATE SESSION ERROR:", err)
		return "", "", errs.ServerError()
	}

	return NewAccessToken, newRefreshToken, nil

}

func (s *AuthService) Logout(ctx context.Context, refreshToken, accessToken string) {
	parsedToken, err := JWT.ParseToken(accessToken)
	if err != nil {
		return
	}

	data, _ := JWT.ConvertToken(parsedToken)
	temp, _ := env.GetTokensLifeTime()
	tokenLifetime := time.Duration(temp)
	if data != nil {
		exp, _ := strconv.ParseFloat(fmt.Sprintf("%v", data["exp"]), 64)
		tokenLifetime = JWT.GetRemainingTime(exp)
	}
	err = s.repositories.SessionRepository.AddAccessTokenToBlackList(ctx, fmt.Sprintf("%v", accessToken), tokenLifetime)
	if err != nil {
		log.Println("ADD TOKEN TO BLACKLIST ERROR:", err)
	}

	if exists, _ := s.repositories.SessionRepository.Exists(ctx, fmt.Sprintf("refresh:%s", refreshToken)); !exists {
		return
	}
	if len(refreshToken) > 0 {
		err := s.repositories.SessionRepository.DeleteSession(ctx, fmt.Sprintf("refresh:%s", refreshToken))
		if err != nil {
			log.Println("DELETE REDIS SESSION WHILE LOGOUT ERROR:", err)
		}
	}

}

func (s *AuthService) LogoutAll(ctx context.Context, userID int, refreshToken, accessToken string) error {

	err := s.repositories.AuthRepository.IncreaseTokenVersion(ctx, userID)
	if err != nil {
		return err
	}

	_, lifeTimeHours := env.GetTokensLifeTime()

	err = s.repositories.SessionRepository.IncreaseTokenVersion(ctx, userID, time.Duration(lifeTimeHours)*time.Hour)
	if err != nil {
		return err
	}

	if exists, _ := s.repositories.SessionRepository.Exists(ctx, fmt.Sprintf("refresh:%s", refreshToken)); !exists {
		return errs.UnauthorizedError()
	}
	if len(refreshToken) > 0 {
		err := s.repositories.SessionRepository.DeleteSession(ctx, fmt.Sprintf("refresh:%s", refreshToken))
		if err != nil {
			log.Println("DELETE REDIS SESSION WHILE LOGOUT ERROR:", err)
		}
	}
	return nil

}
