package service

import (
	errs "GIN/errors"
	"GIN/internal/files"
	"GIN/keys"
	"GIN/model"
	"GIN/tdo/request"
	"GIN/tdo/response"
	"context"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type SongService struct {
	baseService
}

func NewSongService(BService *baseService) *SongService {
	return &SongService{
		baseService: *BService,
	}
}

func (s *SongService) CreateSong(ctx *gin.Context, request request.CreateSongRequest) error {
	audioDst := files.GenerateFilePath(request.Audio)
	volumeDst := files.GenerateFilePath(request.Volume)

	err := s.repositories.StorageRepository.UploadFile(request.Audio, audioDst, ctx)
	if err != nil {
		return err
	}
	if len(volumeDst) != 0 {
		err = s.repositories.StorageRepository.UploadFile(request.Volume, volumeDst, ctx)
		if err != nil {
			return err
		}
	}
	duration, err := files.GetAudioDuration(audioDst)
	if err != nil {
		log.Println("ERROR TRYING GET AUDO DURATION:", err)
		return errs.ServerError()
	}
	if len(volumeDst) != 0 {
		volumeDst = volumeDst[1:]
	}

	song := model.NewSong(0, ctx.Keys[keys.UserIDKey].(int), int(duration), request.Author, request.Name, audioDst[1:], volumeDst, time.Now(), true)
	createdID, err := s.repositories.SongRepository.CreateSong(ctx.Request.Context(), song)
	if err != nil {
		return err
	}

	err = s.repositories.SongRepository.CreateSongListensRow(ctx, createdID)
	if err != nil {
		return err
	}

	return nil
}

func (s *SongService) DeleteSong(ctx context.Context, userID, songID int) error {

	isAuthor, err := s.repositories.SongRepository.UserOwnSong(ctx, userID, songID)
	if err != nil {
		return err
	}
	if !isAuthor {
		return errs.New(http.StatusForbidden, "нет доступа")
	}

	fileName, volumeFileName, err := s.repositories.SongRepository.DeleteSong(ctx, songID)
	if err != nil {
		return err
	}

	fileName = "." + fileName

	err = s.repositories.StorageRepository.DeleteFile(fileName)
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

func (s *SongService) EditSong(ctx *gin.Context, userID, songID int, request *request.EditSongRequest) error {
	isAuthor, err := s.repositories.SongRepository.UserOwnSong(ctx, userID, songID)
	if err != nil {
		return err
	}
	if !isAuthor {
		return errs.New(http.StatusForbidden, "нет доступа")
	}
	audioDst, err := uploadRequestFile(request.Audio, ctx, s.repositories.StorageRepository)
	if err != nil {
		return err
	}

	volumeDst, err := uploadRequestFile(request.Volume, ctx, s.repositories.StorageRepository)
	if err != nil {
		return err
	}

	duration := 0
	if len(audioDst) > 0 {
		floatDuration, _ := files.GetAudioDuration("." + audioDst)
		duration = int(floatDuration)
	}

	oldFile, oldVolume, err := s.repositories.SongRepository.EditSong(ctx.Request.Context(), request.Author, request.Name, volumeDst, audioDst, songID, duration)
	if err != nil {
		return err
	}

	if audioDst != "" && oldFile != "" {
		err := s.repositories.StorageRepository.DeleteFile("." + oldFile)
		if err != nil {
			return err
		}
	}

	if volumeDst != "" && oldVolume != "" {
		err := s.repositories.StorageRepository.DeleteFile("." + oldVolume)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *SongService) GetSong(ctx context.Context, userID, songID int) (*response.GetSongResponse, error) {

	findedSong, err := s.repositories.SongRepository.GetSong(ctx, userID, songID)
	if err != nil {
		return nil, err
	}

	return findedSong, nil
}

func (s *SongService) GetSongs(ctx context.Context, userID, start, count int) (*response.GetSongsResponse, error) {
	songs, err := s.repositories.SongRepository.GetSongs(ctx, userID, start, count)
	if err != nil {
		return nil, err
	}

	findedSongs := response.GetSongsResponse{}
	findedSongs.Data = songs

	return &findedSongs, nil
}

func (s *SongService) ChangeStatus(ctx context.Context, userID, songID int, status bool) error {
	isAuthor, err := s.repositories.SongRepository.UserOwnSong(ctx, userID, songID)
	if err != nil {
		return err
	}
	if !isAuthor {
		return errs.New(http.StatusForbidden, "нет доступа")
	}

	formatedStatus := strconv.FormatBool(status)

	err = s.repositories.SongRepository.ChangeStatus(ctx, songID, formatedStatus)
	if err != nil {
		return err
	}

	return nil
}

func (s *SongService) LikeSong(ctx context.Context, userID, songID int) error {
	alreadyLiked := s.repositories.SongRepository.SongLiked(ctx, userID, songID)
	if alreadyLiked {
		return errs.New(http.StatusConflict, "трек уже лайкнут")
	}
	accessAproved := s.repositories.SongRepository.AccessToSong(ctx, userID, songID)
	if !accessAproved {
		return errs.New(http.StatusForbidden, "трек не найден, либо ограничен в доступе")
	}
	err := s.repositories.SongRepository.Like(ctx, userID, songID, time.Now())
	if err != nil {
		return err
	}
	return nil
}

func (s *SongService) UnlikeSong(ctx context.Context, userID, songID int) error {
	liked := s.repositories.SongRepository.SongLiked(ctx, userID, songID)
	if !liked {
		return errs.New(http.StatusNotFound, "трека нет в лайкнутых")
	}
	err := s.repositories.SongRepository.Unlike(ctx, userID, songID)
	if err != nil {
		return err
	}
	return nil
}

func (s *SongService) AddListen(ctx context.Context, userID, songID int) error {
	access := s.repositories.SongRepository.AccessToSong(ctx, userID, songID)
	if !access {
		return errs.New(http.StatusNotFound, "трек не найден, либо ограничен в доступе")
	}
	listenStoryExists := s.repositories.SongRepository.UserSongListensExists(ctx, userID, songID)
	if !listenStoryExists {
		err := s.repositories.SongRepository.CreateUserSongListens(ctx, userID, songID)
		if err != nil {
			return err
		}
	}

	// Проверка на время выполняется на фронте, раскоментировать чтобы прослушивания засчитывались раз в длительность песни
	// song, err := s.repositories.SongRepository.GetSongInfo(ctx, songID)
	// if err != nil {
	// 	return err
	// }
	// listensInfo, err := s.repositories.SongRepository.GetUserSongListens(ctx, userID, songID)
	// if err != nil {
	// 	return err
	// }

	// if listensInfo.LastListenTime != nil {
	// 	timeListenAccessTreshold := listensInfo.LastListenTime.Add(time.Duration(song.Duration) * time.Second)
	// 	canAddListen := time.Now().UTC().After(timeListenAccessTreshold)
	// 	if !canAddListen {
	// 		return errs.New(http.StatusForbidden, "нельзя добавить прослушивание(слишком короткий промежуток)")
	// 	}
	// }

	err := s.repositories.SongRepository.AddListenToUserSongListens(ctx, userID, songID)
	if err != nil {
		return err
	}
	s.repositories.SongRepository.AddListenToSong(ctx, songID)

	return nil
}
