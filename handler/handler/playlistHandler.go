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

type PlaylistHandlers struct {
	service *service.PlaylistService
}

func NewPlaylistHandler(playlistService *service.PlaylistService) *PlaylistHandlers {
	return &PlaylistHandlers{
		service: playlistService,
	}
}

func (h *PlaylistHandlers) CreatePlaylist(ctx *gin.Context) {
	var request request.CreatePlaylistRequest
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
	err := h.service.CreatePlaylist(ctx, request)
	if err != nil {
		if errors.As(err, &ErrWithCode) {
			errs.ThrowError(ctx, ErrWithCode.Code, ErrWithCode.Message)
			return
		}
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"message": "плейлист создан",
	})
}
func (h *PlaylistHandlers) EditPlaylist(ctx *gin.Context) {
	var request request.EditPlaylistRequest
	if err := ctx.ShouldBind(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": "неправильное тело запроса",
		})
		return
	}

	playlistID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		errs.ThrowError(ctx, http.StatusBadRequest, "некорректный id плейлиста")
		return
	}
	userID := ctx.Keys[keys.UserIDKey].(int)

	validationErrors := h.service.Validate.ValidateWithErrorsByte(request)
	if len(validationErrors) > 0 {
		errs.ThrowValidationErrors(ctx, validationErrors)
		return
	}

	var ErrWithCode errs.ErrorWithCode

	err = h.service.EditPlaylist(ctx, userID, playlistID, &request)
	if err != nil {
		if errors.As(err, &ErrWithCode) {
			errs.ThrowError(ctx, ErrWithCode.Code, ErrWithCode.Message)
			return
		}
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "плейлист успешно изменен",
	})
}
func (h *PlaylistHandlers) DeletePlaylist(ctx *gin.Context) {
	userID := ctx.Keys[keys.UserIDKey].(int)
	playlistID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		errs.ThrowError(ctx, http.StatusBadRequest, "некорректный id плейлиста")
		return
	}
	var ErrWithCode errs.ErrorWithCode
	err = h.service.DeletePlaylist(ctx.Request.Context(), userID, playlistID)
	if err != nil {
		if errors.As(err, &ErrWithCode) {
			errs.ThrowError(ctx, ErrWithCode.Code, ErrWithCode.Message)
			return
		}
	}
	ctx.JSON(http.StatusOK, gin.H{
		"message": "плейлист успешно удален",
	})
}
func (h *PlaylistHandlers) GetPlaylist(ctx *gin.Context) {
	userID := ctx.Keys[keys.UserIDKey].(int)
	playlistID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		errs.ThrowError(ctx, http.StatusBadRequest, "некорректный id плейлиста")
		return
	}
	var ErrWithCode errs.ErrorWithCode
	info, err := h.service.GetPlaylist(ctx.Request.Context(), userID, playlistID)
	if err != nil {
		if errors.As(err, &ErrWithCode) {
			errs.ThrowError(ctx, ErrWithCode.Code, ErrWithCode.Message)
			return
		}
	}
	ctx.JSON(http.StatusOK, gin.H{
		"playlist_info": info,
	})
}

func (h *PlaylistHandlers) GetPlaylists(ctx *gin.Context) {
	start, count := handlerHelpers.GetPaginationFromRequest(ctx)
	userID := ctx.Keys[keys.UserIDKey].(int)
	var ErrWithCode errs.ErrorWithCode
	info, totalRows, err := h.service.GetPlaylists(ctx.Request.Context(), userID, start, count)
	if err != nil {
		if errors.As(err, &ErrWithCode) {
			errs.ThrowError(ctx, ErrWithCode.Code, ErrWithCode.Message)
			return
		}
	}
	ctx.JSON(http.StatusOK, gin.H{
		"totalRows":     totalRows,
		"playlist_info": info,
	})
}

func (h *PlaylistHandlers) GetPlaylistSongs(ctx *gin.Context) {
	start, count := handlerHelpers.GetPaginationFromRequest(ctx)
	userID := ctx.Keys[keys.UserIDKey].(int)
	playlistID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		errs.ThrowError(ctx, http.StatusBadRequest, "некорректный id плейлиста")
		return
	}
	var ErrWithCode errs.ErrorWithCode
	songs, totalRows, err := h.service.GetPlaylistSongs(ctx.Request.Context(), userID, playlistID, start, count)
	if err != nil {
		if errors.As(err, &ErrWithCode) {
			errs.ThrowError(ctx, ErrWithCode.Code, ErrWithCode.Message)
			return
		}
	}

	if len(songs) == 0 && start == 0 {
		ctx.JSON(http.StatusOK, gin.H{
			"message": "Нет треков",
		})
		return
	}
	if len(songs) == 0 && start > 0 {
		ctx.JSON(http.StatusNoContent, gin.H{})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"totalRows": totalRows,
		"songs":     songs,
	})
}

func (h *PlaylistHandlers) ChangePlaylistStatus(ctx *gin.Context) {
	statusStr := ctx.Query("status")
	if len(statusStr) == 0 {
		statusStr = "false"
	}
	status, err := strconv.ParseBool(statusStr)
	if err != nil {
		errs.ThrowError(ctx, http.StatusBadRequest, "некорректное значение статуса")
		return
	}
	userID := ctx.Keys[keys.UserIDKey].(int)
	playlistID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		errs.ThrowError(ctx, http.StatusBadRequest, "некорректный id плейлиста")
		return
	}

	var ErrWithCode errs.ErrorWithCode

	err = h.service.EditPrivateStatus(ctx.Request.Context(), userID, playlistID, status)
	if err != nil {
		if errors.As(err, &ErrWithCode) {
			errs.ThrowError(ctx, ErrWithCode.Code, ErrWithCode.Message)
			return
		}
	}
	ctx.JSON(http.StatusOK, gin.H{
		"message": "приватность плейлиста успешно изменена",
	})
}

func (h *PlaylistHandlers) DeleteSongFromPlaylist(ctx *gin.Context) {
	userID := ctx.Keys[keys.UserIDKey].(int)
	playlistID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		errs.ThrowError(ctx, http.StatusBadRequest, "некорректный id плейлиста")
		return
	}
	songID, err := strconv.Atoi(ctx.Param("songID"))
	if err != nil {
		errs.ThrowError(ctx, http.StatusBadRequest, "некорректный id трека")
		return
	}

	var errWithCode errs.ErrorWithCode

	err = h.service.DeleteSongFromPlaylist(ctx.Request.Context(), userID, playlistID, songID)
	if err != nil {
		if errors.As(err, &errWithCode) {
			errs.ThrowError(ctx, errWithCode.Code, errWithCode.Message)
			return
		}
	}
	ctx.JSON(http.StatusOK, gin.H{
		"message": "трек удален из плейлиста",
	})
}
func (h *PlaylistHandlers) AddSongToPlaylist(ctx *gin.Context) {
	userID := ctx.Keys[keys.UserIDKey].(int)
	playlistID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		errs.ThrowError(ctx, http.StatusBadRequest, "некорректный id плейлиста")
		return
	}
	songID, err := strconv.Atoi(ctx.Param("songID"))
	if err != nil {
		errs.ThrowError(ctx, http.StatusBadRequest, "некорректный id трека")
		return
	}

	var errWithCode errs.ErrorWithCode

	err = h.service.AddSongToPlaylist(ctx.Request.Context(), userID, playlistID, songID)
	if err != nil {
		if errors.As(err, &errWithCode) {
			errs.ThrowError(ctx, errWithCode.Code, errWithCode.Message)
			return
		}
	}
	ctx.JSON(http.StatusOK, gin.H{
		"message": "трек добавлен в плейлист",
	})
}

func (h *PlaylistHandlers) LikePlaylist(ctx *gin.Context) {
	userID := ctx.Keys[keys.UserIDKey].(int)
	playlistID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		errs.ThrowError(ctx, http.StatusBadRequest, "некорректный id плейлиста")
		return
	}
	var errWithCode errs.ErrorWithCode
	err = h.service.Like(ctx.Request.Context(), userID, playlistID)
	if err != nil {
		if errors.As(err, &errWithCode) {
			errs.ThrowError(ctx, errWithCode.Code, errWithCode.Message)
			return
		}
	}

	ctx.JSON(http.StatusNoContent, gin.H{})
}
func (h *PlaylistHandlers) UnLikePlaylist(ctx *gin.Context) {
	userID := ctx.Keys[keys.UserIDKey].(int)
	playlistID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		errs.ThrowError(ctx, http.StatusBadRequest, "некорректный id плейлиста")
		return
	}
	var errWithCode errs.ErrorWithCode
	err = h.service.Unlike(ctx.Request.Context(), userID, playlistID)
	if err != nil {
		if errors.As(err, &errWithCode) {
			errs.ThrowError(ctx, errWithCode.Code, errWithCode.Message)
			return
		}
	}

	ctx.JSON(http.StatusNoContent, gin.H{})
}
