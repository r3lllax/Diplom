package service

import (
	errs "GIN/errors"
	"GIN/internal/env"
	"GIN/model"
	"GIN/tdo"
	"GIN/tdo/request"
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type UserService struct {
	baseService
}

func NewUserService(BService *baseService) *UserService {
	return &UserService{
		baseService: *BService,
	}
}
func (s *UserService) GetUser(ctx context.Context, targetUserID, userID int) (*model.User, error) {
	user, err := s.repositories.UserRepository.GetUser(ctx, targetUserID, userID)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *UserService) EditUser(ctx *gin.Context, request *request.EditUserRequest, userID, targetUserID int) error {
	isUser := userID == targetUserID
	if !isUser {
		return errs.New(http.StatusForbidden, "нет доступа")
	}
	newPhotoDist, err := uploadRequestFile(request.Photo, ctx, s.repositories.StorageRepository)
	if err != nil {
		return err
	}

	oldPhoto, err := s.repositories.UserRepository.EditUser(ctx.Request.Context(), userID, request.Name, newPhotoDist, request.Email, request.DeletePhoto)
	if err != nil {
		return err
	}

	if (newPhotoDist != "" && oldPhoto != "") || (request.DeletePhoto && oldPhoto != "") {
		err := s.repositories.StorageRepository.DeleteFile("." + oldPhoto)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *UserService) DeleteUser(ctx context.Context, userID int) error {
	err := s.repositories.AuthRepository.IncreaseTokenVersion(ctx, userID)
	if err != nil {
		return err
	}

	_, lifeTimeHours := env.GetTokensLifeTime()

	err = s.repositories.SessionRepository.IncreaseTokenVersion(ctx, userID, time.Duration(lifeTimeHours)*time.Hour)
	if err != nil {
		return err
	}

	err = s.repositories.UserRepository.DeleteUser(ctx, userID)
	if err != nil {
		return err
	}

	return nil

}

func (s *UserService) EditPrivateStatus(ctx context.Context, status bool, userID int) error {
	err := s.repositories.UserRepository.EditPrivateStatus(ctx, status, userID)
	if err != nil {
		return err
	}
	return nil
}

func (s *UserService) GetLikedSongs(ctx context.Context, userID, targetUserID, start, count int) ([]model.SongInLikes, error) {

	likedSongs, err := s.repositories.UserRepository.GetLikedSongs(ctx, targetUserID, userID, start, count)
	if err != nil {
		return []model.SongInLikes{}, err
	}

	return likedSongs, nil

}

func (s *UserService) GetUserSongs(ctx context.Context, userID, start, count int) ([]tdo.UserSong, error) {
	songs, err := s.repositories.UserRepository.GetUserSongs(ctx, userID, start, count)
	if err != nil {
		return []tdo.UserSong{}, err
	}

	var tdoSongs []tdo.UserSong

	for _, song := range songs {
		tdoSong := tdo.NewUserSong(song.Id, song.Duration, song.Author, song.Name, song.File_path, song.Volume_path, song.Uploaded_at, song.Is_available)
		tdoSongs = append(tdoSongs, *tdoSong)
	}

	return tdoSongs, nil
}

func (s *UserService) GetUserListenStatistics(ctx context.Context, userID, start, count int, countSort bool) ([]tdo.UserSongListenStatistic, error) {
	statistics, err := s.repositories.UserRepository.GetUserListenStatistics(ctx, userID, start, count, countSort)
	if err != nil {
		return []tdo.UserSongListenStatistic{}, err
	}
	return statistics, nil
}
func (s *UserService) UserRecentlyListenStatistics(ctx context.Context, userID int) ([]tdo.UserSongListenStatistic, error) {
	statistics, err := s.repositories.UserRepository.GetUserLastListenTracks(ctx, userID)
	if err != nil {
		return []tdo.UserSongListenStatistic{}, err
	}
	return statistics, nil
}

func (s *UserService) GetUserGeneralListenStats(ctx context.Context, userID int) (int, int, int, int, error) {
	songsCount, timeCount, favoritesCount, likesCount, err := s.repositories.UserRepository.GetUserGeneralListenStats(ctx, userID)
	if err != nil {
		return 0, 0, 0, 0, err
	}
	return songsCount, timeCount, favoritesCount, likesCount, nil
}
func (s *UserService) GetUserLikesDuration(ctx context.Context, userID int) (int, error) {
	time, err := s.repositories.UserRepository.GetUserLikesDuration(ctx, userID)
	if err != nil {
		return 0, err
	}
	return time, nil
}
func (s *UserService) GetUserLikesTotalRows(ctx context.Context, targetUserID, userID int) (int, error) {
	time, err := s.repositories.UserRepository.GetUserLikesTotalRows(ctx, targetUserID, userID)
	if err != nil {
		return 0, err
	}
	return time, nil
}

func (s *UserService) GetUserLikesCount(ctx context.Context, userID int) (int, error) {
	likesCount, err := s.repositories.UserRepository.GetUserLikesCount(ctx, userID)
	if err != nil {
		return 0, err
	}
	return likesCount, nil
}

func (s *UserService) GetUserTracksListenStatistics(ctx context.Context, userID, start, count int) ([]tdo.UserTrackListenStatistic, error) {
	list, err := s.repositories.UserRepository.GetUserTracksListenStatistics(ctx, userID, start, count)
	if err != nil {
		return []tdo.UserTrackListenStatistic{}, err
	}

	return list, nil
}

func (s *UserService) GetUserTracksLikesCount(ctx context.Context, userID int) (int, error) {
	likesCount, err := s.repositories.UserRepository.GetUserTracksLikesCount(ctx, userID)
	if err != nil {
		return 0, err
	}
	return likesCount, nil
}

func (s *UserService) GetUserPlaylists(ctx context.Context, userID, targetUserID, start, count int) ([]tdo.UserPlaylist, int, error) {

	playlists, err := s.repositories.UserRepository.GetUserPlaylists(ctx, userID, targetUserID, start, count)
	if err != nil {
		return []tdo.UserPlaylist{}, 0, err
	}
	totalRows, err := s.repositories.PlaylistRepository.GetUserPlaylistsTotalCount(ctx, userID)
	if err != nil {
		return []tdo.UserPlaylist{}, 0, err
	}

	return playlists, totalRows, nil

}
