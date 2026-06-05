package handler

import (
	errs "GIN/errors"
	"GIN/handler/handlerHelpers"
	"GIN/keys"
	service "GIN/service/services"
	"GIN/tdo/request"
	"GIN/tdo/response"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type UserHandlers struct {
	service *service.UserService
}

func NewUserHandler(userService *service.UserService) *UserHandlers {
	return &UserHandlers{
		service: userService,
	}
}

func (h *UserHandlers) GetUser(ctx *gin.Context) {
	targetUserID, err := strconv.Atoi(ctx.Param("id"))
	userID := ctx.Keys[keys.UserIDKey].(int)
	if err != nil {
		errs.ThrowError(ctx, http.StatusBadRequest, "некорректный id пользователя")
		return
	}
	var ErrWithCode errs.ErrorWithCode
	user, err := h.service.GetUser(ctx.Request.Context(), targetUserID, userID)
	if err != nil {
		if errors.As(err, &ErrWithCode) {
			errs.ThrowError(ctx, ErrWithCode.Code, ErrWithCode.Message)
			return
		}
	}
	if user == nil {
		errs.ThrowServerError(ctx)
		return
	}
	response := response.GetUser{Id: user.Id, Name: user.Name, Photo_file: user.Photo_file, Registrated_at: user.Registrated_at}
	ctx.JSON(http.StatusOK, gin.H{
		"userInfo": response,
	})

}
func (h *UserHandlers) EditUser(ctx *gin.Context) {
	var request request.EditUserRequest
	if err := ctx.ShouldBind(&request); err != nil {
		errs.ThrowError(ctx, http.StatusBadRequest, "неверное тело запроса")
		return
	}

	userID := ctx.Keys[keys.UserIDKey].(int)

	targetUserID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		errs.ThrowError(ctx, http.StatusBadRequest, "некорректный id трека")

		return
	}
	validationErrors := h.service.Validate.ValidateWithErrorsByte(request)
	if len(validationErrors) > 0 {
		errs.ThrowValidationErrors(ctx, validationErrors)
		return
	}

	var ErrWithCode errs.ErrorWithCode

	err = h.service.EditUser(ctx, &request, userID, targetUserID)
	if err != nil {
		if errors.As(err, &ErrWithCode) {
			errs.ThrowError(ctx, ErrWithCode.Code, ErrWithCode.Message)
			return
		}
	}
	ctx.JSON(http.StatusOK, gin.H{
		"message": "данные обновлены",
	})
}
func (h *UserHandlers) EditUserPrivateStatus(ctx *gin.Context) {
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

	var ErrWithCode errs.ErrorWithCode

	err = h.service.EditPrivateStatus(ctx.Request.Context(), status, userID)
	if err != nil {
		if errors.As(err, &ErrWithCode) {
			errs.ThrowError(ctx, ErrWithCode.Code, ErrWithCode.Message)
			return
		}
	}
	ctx.JSON(http.StatusOK, gin.H{
		"message": "приватность профиля успешно изменена",
	})
}
func (h *UserHandlers) DeleteUser(ctx *gin.Context) {

	userID := ctx.Keys[keys.UserIDKey].(int)

	err := h.service.DeleteUser(ctx.Request.Context(), userID)

	var ErrWithCode errs.ErrorWithCode
	if err != nil {
		if errors.As(err, &ErrWithCode) {
			errs.ThrowError(ctx, ErrWithCode.Code, ErrWithCode.Message)
			return
		}
	}
	ctx.JSON(http.StatusNoContent, gin.H{})
}
func (h *UserHandlers) GetUserPlaylists(ctx *gin.Context) {
	start, count := handlerHelpers.GetPaginationFromRequest(ctx)
	userID := ctx.Keys[keys.UserIDKey].(int)
	targetUserID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		errs.ThrowError(ctx, http.StatusBadRequest, "некорректный id пользователя")
		return
	}
	var errWithCode errs.ErrorWithCode
	playlists, err := h.service.GetUserPlaylists(ctx.Request.Context(), userID, targetUserID, start, count)
	if err != nil {
		if errors.As(err, &errWithCode) {
			errs.ThrowError(ctx, errWithCode.Code, errWithCode.Message)
			return
		}
	}

	if len(playlists) == 0 && start == 0 {
		ctx.JSON(http.StatusNotFound, gin.H{
			"message": "нет плейлистов",
		})
		return
	}

	if len(playlists) == 0 && start > 0 {
		ctx.JSON(http.StatusNoContent, gin.H{})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"playlists": playlists,
	})

}

func (h *UserHandlers) GetUserLikes(ctx *gin.Context) {
	start, count := handlerHelpers.GetPaginationFromRequest(ctx)
	userID := ctx.Keys[keys.UserIDKey].(int)
	targetUserID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		errs.ThrowError(ctx, http.StatusBadRequest, "некорректный id пользователя")
		return
	}

	var ErrWithCode errs.ErrorWithCode

	likedSongs, err := h.service.GetLikedSongs(ctx.Request.Context(), userID, targetUserID, start, count)
	if err != nil {
		if errors.As(err, &ErrWithCode) {
			errs.ThrowError(ctx, ErrWithCode.Code, ErrWithCode.Message)
			return
		}
	}
	if len(likedSongs) == 0 && start == 0 {
		ctx.JSON(http.StatusOK, gin.H{
			"message": "Нет любимых треков",
		})
		return
	}

	if len(likedSongs) == 0 && start > 0 {
		ctx.JSON(http.StatusNoContent, gin.H{})

		return
	}
	duration, err := h.service.GetUserLikesDuration(ctx.Request.Context(), userID)
	if err != nil {
		if errors.As(err, &ErrWithCode) {
			errs.ThrowError(ctx, ErrWithCode.Code, ErrWithCode.Message)
			return
		}
	}
	totalRows, err := h.service.GetUserLikesTotalRows(ctx.Request.Context(), targetUserID, userID)
	if err != nil {
		if errors.As(err, &ErrWithCode) {
			errs.ThrowError(ctx, ErrWithCode.Code, ErrWithCode.Message)
			return
		}
	}

	ctx.JSON(http.StatusOK, gin.H{
		"totalRows": totalRows,
		"duration":  duration,
		"songs":     likedSongs,
	})

}

func (h *UserHandlers) UserSongs(ctx *gin.Context) {
	start, count := handlerHelpers.GetPaginationFromRequest(ctx)
	userID := ctx.Keys[keys.UserIDKey].(int)

	var ErrWithCode errs.ErrorWithCode
	songs, err := h.service.GetUserSongs(ctx.Request.Context(), userID, start, count)
	if err != nil {
		if errors.As(err, &ErrWithCode) {
			errs.ThrowError(ctx, ErrWithCode.Code, ErrWithCode.Message)
			return
		}
	}
	if len(songs) == 0 && start == 0 {
		ctx.JSON(http.StatusOK, gin.H{
			"message": "У вас нет загруженных треков",
		})
		return
	}
	if len(songs) == 0 && start > 0 {
		ctx.JSON(http.StatusNoContent, gin.H{})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"songs": songs,
	})
}

func (h *UserHandlers) UserListenStatistics(ctx *gin.Context) {
	start, count := handlerHelpers.GetPaginationFromRequest(ctx)
	userID := ctx.Keys[keys.UserIDKey].(int)

	countSort := ctx.Query("countSort") == "true"

	var ErrWithCode errs.ErrorWithCode
	statistic, err := h.service.GetUserListenStatistics(ctx, userID, start, count, countSort)
	if err != nil {
		if errors.As(err, &ErrWithCode) {
			errs.ThrowError(ctx, ErrWithCode.Code, ErrWithCode.Message)
			return
		}
	}

	if len(statistic) == 0 && start == 0 {
		ctx.JSON(http.StatusOK, gin.H{
			"message": "у вас пока нет статистики, прослушайте треки, чтобы она появилась",
		})
		return
	}
	if len(statistic) == 0 && start > 0 {
		ctx.JSON(http.StatusNoContent, gin.H{})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"statistics": statistic,
	})
}
func (h *UserHandlers) UserRecentlyListenStatistics(ctx *gin.Context) {
	userID := ctx.Keys[keys.UserIDKey].(int)

	var ErrWithCode errs.ErrorWithCode
	statistic, err := h.service.UserRecentlyListenStatistics(ctx, userID)
	if err != nil {
		if errors.As(err, &ErrWithCode) {
			errs.ThrowError(ctx, ErrWithCode.Code, ErrWithCode.Message)
			return
		}
	}

	if len(statistic) == 0 {
		ctx.JSON(http.StatusOK, gin.H{
			"message": "у вас пока нет статистики, прослушайте треки, чтобы она появилась",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"statistics": statistic,
	})
}

func (h *UserHandlers) UserGeneralListenStats(ctx *gin.Context) {
	userID := ctx.Keys[keys.UserIDKey].(int)
	var ErrWithCode errs.ErrorWithCode
	listenSongsCount, listenTime, likesCount, userSongsLikes, err := h.service.GetUserGeneralListenStats(ctx, userID)
	if err != nil {
		if errors.As(err, &ErrWithCode) {
			errs.ThrowError(ctx, ErrWithCode.Code, ErrWithCode.Message)
			return
		}
	}

	ctx.JSON(http.StatusOK, gin.H{
		"listenSongsCount": listenSongsCount,
		"totalListenTime":  listenTime,
		"likesCount":       likesCount,
		"userSongsLikes":   userSongsLikes,
	})
}

func (h *UserHandlers) UserLikesCount(ctx *gin.Context) {
	userID := ctx.Keys[keys.UserIDKey].(int)
	var ErrWithCode errs.ErrorWithCode
	count, err := h.service.GetUserLikesCount(ctx, userID)
	if err != nil {
		if errors.As(err, &ErrWithCode) {
			errs.ThrowError(ctx, ErrWithCode.Code, ErrWithCode.Message)
			return
		}
	}
	ctx.JSON(http.StatusOK, gin.H{
		"likesCount": count,
	})
}

func (h *UserHandlers) UserTracksListenStatistics(ctx *gin.Context) {
	start, count := handlerHelpers.GetPaginationFromRequest(ctx)
	userID := ctx.Keys[keys.UserIDKey].(int)
	var ErrWithCode errs.ErrorWithCode
	statistics, err := h.service.GetUserTracksListenStatistics(ctx, userID, start, count)
	if err != nil {
		if errors.As(err, &ErrWithCode) {
			errs.ThrowError(ctx, ErrWithCode.Code, ErrWithCode.Message)
			return
		}
	}

	if len(statistics) == 0 && start == 0 {
		ctx.JSON(http.StatusOK, gin.H{
			"message": "у вас пока нет загруженных треков",
		})
		return
	}
	if len(statistics) == 0 && start > 0 {
		ctx.JSON(http.StatusNoContent, gin.H{})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"statistics": statistics,
	})
}

func (h *UserHandlers) UserTracksLikesCount(ctx *gin.Context) {
	userID := ctx.Keys[keys.UserIDKey].(int)
	var ErrWithCode errs.ErrorWithCode
	count, err := h.service.GetUserTracksLikesCount(ctx, userID)
	if err != nil {
		if errors.As(err, &ErrWithCode) {
			errs.ThrowError(ctx, ErrWithCode.Code, ErrWithCode.Message)
			return
		}
	}
	ctx.JSON(http.StatusOK, gin.H{
		"likesCount": count,
	})
}
