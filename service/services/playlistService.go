package service

import (
	errs "GIN/errors"
	"GIN/internal/files"
	"GIN/keys"
	"GIN/model"
	"GIN/tdo"
	"GIN/tdo/request"
	"GIN/tdo/response"
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type PlaylistService struct {
	baseService
}

func NewPlaylistService(BService *baseService) *PlaylistService {
	return &PlaylistService{
		baseService: *BService,
	}
}

func (s *PlaylistService) CreatePlaylist(ctx *gin.Context, request request.CreatePlaylistRequest) error {

	userID := ctx.Keys[keys.UserIDKey].(int)
	volumeDst := files.GenerateFilePath(request.Volume)

	if len(volumeDst) != 0 {
		err := s.repositories.StorageRepository.UploadFile(request.Volume, volumeDst, ctx)
		if err != nil {
			return err
		}
	}

	if len(volumeDst) != 0 {
		volumeDst = volumeDst[1:]
	}

	playlist := model.NewPlaylist(0, userID, request.Title, request.Description, volumeDst, request.IsPrivate, true)
	id, err := s.repositories.PlaylistRepository.CreatePlaylist(ctx.Request.Context(), playlist)
	if err != nil {
		return err
	}
	err = s.repositories.PlaylistRepository.LikePlaylist(ctx, userID, id)
	if err != nil {
		return err
	}
	return nil
}

func (s *PlaylistService) EditPlaylist(ctx *gin.Context, userID, playlistID int, request *request.EditPlaylistRequest) error {
	isAuthor, err := s.repositories.PlaylistRepository.UserOwnPlaylist(ctx, userID, playlistID)
	if err != nil {
		return err
	}
	if !isAuthor {
		return errs.New(http.StatusForbidden, "нет доступа")
	}

	volumeDst, err := uploadRequestFile(request.Volume, ctx, s.repositories.StorageRepository)
	if err != nil {
		return err
	}

	oldVolume, err := s.repositories.PlaylistRepository.EditPlaylist(ctx.Request.Context(), request.Title, volumeDst, request.Description, playlistID)
	if err != nil {
		return err
	}

	if volumeDst != "" && oldVolume != "" {
		err := s.repositories.StorageRepository.DeleteFile("." + oldVolume)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *PlaylistService) DeletePlaylist(ctx context.Context, userID, playlistID int) error {
	isAuthor, err := s.repositories.PlaylistRepository.UserOwnPlaylist(ctx, userID, playlistID)
	if err != nil {
		return err
	}
	if !isAuthor {
		return errs.New(http.StatusForbidden, "нет доступа")
	}

	volumeFileName, err := s.repositories.PlaylistRepository.DeletePlaylist(ctx, playlistID)
	if err != nil {
		return err
	}

	if len(volumeFileName) != 0 {
		volumeFileName = "." + volumeFileName
		err = s.repositories.StorageRepository.DeleteFile(volumeFileName)
		if err != nil {
			return err
		}
	}

	return nil

}

func (s *PlaylistService) GetPlaylist(ctx context.Context, userID, playlistID int) (*tdo.PlaylistInfo, error) {
	info, err := s.repositories.PlaylistRepository.GetInfo(ctx, userID, playlistID)
	if err != nil {
		return nil, err
	}
	return info, nil
}

func (s *PlaylistService) GetPlaylists(ctx context.Context, userID, from, count int) ([]tdo.PlaylistInfo, int, error) {
	info, err := s.repositories.PlaylistRepository.GetPlaylists(ctx, userID, from, count)
	if err != nil {
		return nil, 0, err
	}
	totalRows, err := s.repositories.PlaylistRepository.GetPlaylistsTotalCount(ctx, userID)
	if err != nil {
		return nil, 0, err
	}
	return info, totalRows, nil
}

func (s *PlaylistService) GetPlaylistSongs(ctx context.Context, userID, playlistID, start, count int) ([]response.GetSongResponse, int, error) {
	songs, err := s.repositories.PlaylistRepository.GetSongs(ctx, userID, playlistID, start, count)
	if err != nil {
		return nil, 0, err
	}
	totalRows, err := s.repositories.PlaylistRepository.GetSongsTotalRows(ctx, userID, playlistID)
	if err != nil {
		return nil, 0, err
	}
	return songs, totalRows, nil
}

func (s *PlaylistService) EditPrivateStatus(ctx context.Context, userID, playlistID int, status bool) error {
	isAuthor, err := s.repositories.PlaylistRepository.UserOwnPlaylist(ctx, userID, playlistID)
	if err != nil {
		return err
	}
	if !isAuthor {
		return errs.New(http.StatusForbidden, "нет доступа")
	}

	formatedStatus := strconv.FormatBool(status)

	err = s.repositories.PlaylistRepository.ChangeStatus(ctx, playlistID, formatedStatus)
	if err != nil {
		return err
	}

	return nil
}

func (s *PlaylistService) Like(ctx context.Context, userID, playlistID int) error {
	alreadyLiked := s.repositories.PlaylistRepository.PlaylistLiked(ctx, userID, playlistID)
	if alreadyLiked {
		return errs.New(http.StatusConflict, "плейлист уже лайкнут")
	}

	accessAproved := s.repositories.PlaylistRepository.AccessToPlaylist(ctx, userID, playlistID)
	if !accessAproved {
		return errs.New(http.StatusForbidden, "плейлист не найден, либо ограничен в доступе")
	}
	err := s.repositories.PlaylistRepository.LikePlaylist(ctx, userID, playlistID)
	if err != nil {
		return err
	}
	return nil

}
func (s *PlaylistService) Unlike(ctx context.Context, userID, playlistID int) error {
	liked := s.repositories.PlaylistRepository.PlaylistLiked(ctx, userID, playlistID)
	if !liked {
		return errs.New(http.StatusConflict, "плейлиста нет в лайкнутых")
	}
	isAuthor, err := s.repositories.PlaylistRepository.UserOwnPlaylist(ctx, userID, playlistID)
	if err != nil {
		return err
	}
	if isAuthor {
		return errs.New(http.StatusForbidden, "нет доступа")
	}
	err = s.repositories.PlaylistRepository.Unlike(ctx, userID, playlistID)
	if err != nil {
		return err
	}
	return nil

}

func (s *PlaylistService) AddSongToPlaylist(ctx context.Context, userID, playlistID, songID int) error {
	haveAccessToSong := s.repositories.SongRepository.AccessToSong(ctx, userID, songID)
	if !haveAccessToSong {
		return errs.New(http.StatusForbidden, "песня которую вы хотите добавить в плейлист не существует, или она приватна")
	}
	haveAccessToPlaylist, err := s.repositories.PlaylistRepository.UserOwnPlaylist(ctx, userID, playlistID)
	if err != nil {
		return err
	}
	if !haveAccessToPlaylist {
		return errs.New(http.StatusForbidden, "плейлист в который вы хотите добавить трек не существует, либо вы не его автор")
	}

	alreadyInPlaylist := s.repositories.PlaylistRepository.SongInPlaylist(ctx, songID, playlistID)
	if alreadyInPlaylist {
		return errs.New(http.StatusConflict, "трек уже есть в плейлисте")
	}

	err = s.repositories.PlaylistRepository.AddSongToPlaylist(ctx, playlistID, songID)
	if err != nil {
		return err
	}

	return nil
}

func (s *PlaylistService) DeleteSongFromPlaylist(ctx context.Context, userID, playlistID, songID int) error {
	haveAccessToSong := s.repositories.SongRepository.AccessToSong(ctx, userID, songID)
	if !haveAccessToSong {
		return errs.New(http.StatusForbidden, "песня которую вы хотите удалить из плейлиста не существует, или она приватна")
	}
	haveAccessToPlaylist, err := s.repositories.PlaylistRepository.UserOwnPlaylist(ctx, userID, playlistID)
	if err != nil {
		return err
	}
	if !haveAccessToPlaylist {
		return errs.New(http.StatusForbidden, "плейлист из которого вы хотите удалить трек не существует, либо вы не его автор")
	}

	inPlaylist := s.repositories.PlaylistRepository.SongInPlaylist(ctx, songID, playlistID)
	if !inPlaylist {
		return errs.New(http.StatusConflict, "трека нет в плейлисте")
	}

	err = s.repositories.PlaylistRepository.DeleteSongFromPlaylist(ctx, songID, playlistID)
	if err != nil {
		return err
	}

	return nil
}
