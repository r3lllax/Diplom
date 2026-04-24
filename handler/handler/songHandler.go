package handler

import (
	errs "GIN/errors"
	"GIN/handler/handlerHelpers"
	"GIN/keys"
	service "GIN/service/services"
	"GIN/tdo/request"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type SongHandlers struct {
	service *service.SongService
}

func NewSongHandler(songService *service.SongService) *SongHandlers {
	return &SongHandlers{
		service: songService,
	}
}

func (h *SongHandlers) CreateSong(ctx *gin.Context) {
	var request request.CreateSongRequest
	if err := ctx.ShouldBind(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": "неправильное тело запроса",
		})
		return
	}

	validationErrors := h.service.Validate.ValidateWithErrorsByte(request)
	if len(validationErrors) > 0 {
		errs.ThrowValidationErrors(ctx, validationErrors)
		return
	}

	var ErrWithCode errs.ErrorWithCode
	err := h.service.CreateSong(ctx, request)
	if err != nil {
		if errors.As(err, &ErrWithCode) {
			errs.ThrowError(ctx, ErrWithCode.Code, ErrWithCode.Message)
			return
		}
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"message": "трек успешно загружен!",
	})
}
func (h *SongHandlers) DeleteSong(ctx *gin.Context) {
	userID := ctx.Keys[keys.UserIDKey].(int)
	songID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		errs.ThrowError(ctx, http.StatusBadRequest, "некорректный id трека")
		return
	}
	var ErrWithCode errs.ErrorWithCode
	err = h.service.DeleteSong(ctx.Request.Context(), userID, songID)
	if err != nil {
		if errors.As(err, &ErrWithCode) {
			errs.ThrowError(ctx, ErrWithCode.Code, ErrWithCode.Message)
			return
		}
	}
	ctx.JSON(http.StatusOK, gin.H{
		"message": "трек успешно удален",
	})
}
func (h *SongHandlers) EditSong(ctx *gin.Context) {
	var request request.EditSongRequest
	if err := ctx.ShouldBind(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": "неправильное тело запроса",
		})
		return
	}
	songID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		errs.ThrowError(ctx, http.StatusBadRequest, "некорректный id трека")
		return
	}
	userID := ctx.Keys[keys.UserIDKey].(int)

	validationErrors := h.service.Validate.ValidateWithErrorsByte(request)
	if len(validationErrors) > 0 {
		errs.ThrowValidationErrors(ctx, validationErrors)
		return
	}

	var ErrWithCode errs.ErrorWithCode

	err = h.service.EditSong(ctx, userID, songID, &request)
	if err != nil {
		if errors.As(err, &ErrWithCode) {
			errs.ThrowError(ctx, ErrWithCode.Code, ErrWithCode.Message)
			return
		}
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "трек успешно изменен",
	})
}
func (h *SongHandlers) GetSong(ctx *gin.Context) {
	songID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		errs.ThrowError(ctx, http.StatusBadRequest, "некорректный id трека")
		return
	}
	userID := ctx.Keys[keys.UserIDKey].(int)

	var ErrWithCode errs.ErrorWithCode
	findedSong, err := h.service.GetSong(ctx.Request.Context(), userID, songID)
	if err != nil {
		if errors.As(err, &ErrWithCode) {
			errs.ThrowError(ctx, ErrWithCode.Code, ErrWithCode.Message)
			return
		}
	}
	ctx.JSON(http.StatusOK, gin.H{
		"songInfo": findedSong,
	})
}

func (h *SongHandlers) GetSongs(ctx *gin.Context) {
	start, count := handlerHelpers.GetPaginationFromRequest(ctx)
	userID := ctx.Keys[keys.UserIDKey].(int)

	var ErrWithCode errs.ErrorWithCode
	songs, err := h.service.GetSongs(ctx.Request.Context(), userID, start, count)
	if err != nil {
		if errors.As(err, &ErrWithCode) {
			errs.ThrowError(ctx, ErrWithCode.Code, ErrWithCode.Message)
			return
		}
	}
	ctx.JSON(http.StatusOK, gin.H{
		"data": songs.Data,
	})
}

func (h *SongHandlers) ChangeSongStatus(ctx *gin.Context) {
	songID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		errs.ThrowError(ctx, http.StatusBadRequest, "некорректный id трека")
		return
	}
	status, err := strconv.ParseBool(ctx.Param("status"))
	if err != nil {
		errs.ThrowError(ctx, http.StatusBadRequest, "некорректный статус трека")
		return
	}

	userID := ctx.Keys[keys.UserIDKey].(int)

	var ErrWithCode errs.ErrorWithCode
	err = h.service.ChangeStatus(ctx.Request.Context(), userID, songID, status)
	if err != nil {
		if errors.As(err, &ErrWithCode) {
			errs.ThrowError(ctx, ErrWithCode.Code, ErrWithCode.Message)
			return
		}
	}
	ctx.JSON(http.StatusOK, gin.H{
		"message": "статус успешно изменен",
	})
}

func (h *SongHandlers) LikeSong(ctx *gin.Context) {
	songID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		errs.ThrowError(ctx, http.StatusBadRequest, "некорректный id трека")
		return
	}
	userID := ctx.Keys[keys.UserIDKey].(int)

	var ErrWithCode errs.ErrorWithCode

	err = h.service.LikeSong(ctx.Request.Context(), userID, songID)
	if err != nil {
		if errors.As(err, &ErrWithCode) {
			errs.ThrowError(ctx, ErrWithCode.Code, ErrWithCode.Message)
			return
		}
	}
	ctx.JSON(http.StatusOK, gin.H{
		"message": "трек добавлен в любимые",
	})
}
func (h *SongHandlers) UnLikeSong(ctx *gin.Context) {
	songID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		errs.ThrowError(ctx, http.StatusBadRequest, "некорректный id трека")
		return
	}
	userID := ctx.Keys[keys.UserIDKey].(int)

	var ErrWithCode errs.ErrorWithCode

	err = h.service.UnlikeSong(ctx.Request.Context(), userID, songID)
	if err != nil {
		if errors.As(err, &ErrWithCode) {
			errs.ThrowError(ctx, ErrWithCode.Code, ErrWithCode.Message)
			return
		}
	}
	ctx.JSON(http.StatusOK, gin.H{
		"message": "трек удален из лайкнутых",
	})
}

func (h *SongHandlers) AddListen(ctx *gin.Context) {
	songID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		errs.ThrowError(ctx, http.StatusBadRequest, "некорректный id трека")
		return
	}
	userID := ctx.Keys[keys.UserIDKey].(int)

	var ErrWithCode errs.ErrorWithCode

	err = h.service.AddListen(ctx.Request.Context(), userID, songID)

	if err != nil {
		if errors.As(err, &ErrWithCode) {
			errs.ThrowError(ctx, ErrWithCode.Code, ErrWithCode.Message)
			return
		}
	}
	ctx.JSON(http.StatusNoContent, gin.H{})
}
