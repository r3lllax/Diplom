package repository

import (
	errs "GIN/errors"
	"GIN/model"
	"GIN/tdo/response"
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SongRepository struct {
	db  *pgxpool.Pool
	mtx *sync.RWMutex
}

func NewSongRepository(dbase *pgxpool.Pool) *SongRepository {
	return &SongRepository{
		db:  dbase,
		mtx: &sync.RWMutex{},
	}
}

func (r *SongRepository) CreateSong(ctx context.Context, song *model.Song) (int, error) {
	var songID int
	err := r.db.QueryRow(ctx,
		"insert into songs(user_id,author,name,duration,file_path,volume_path,uploaded_at,is_available) values ($1,$2,$3,$4,$5,$6,$7,$8) returning id",
		song.User_id, song.Author, song.Name, song.Duration, song.File_path, song.Volume_path, song.Uploaded_at, song.Is_available).Scan(&songID)
	if err != nil {
		log.Println("CREATE SONG ERROR:", err)
		return 0, errs.ServerError()
	}

	return songID, nil
}

func (r *SongRepository) CreateSongListensRow(ctx context.Context, songID int) error {
	_, err := r.db.Exec(ctx, "insert into songs_listens (song_id,listens) values ($1,$2)", songID, 0)
	if err != nil {
		log.Println("CREATE SONG LISTENS ROW ERROR:", err)
		return errs.ServerError()
	}
	return nil
}

func (r *SongRepository) DeleteSong(ctx context.Context, songID int) (string, string, error) {
	var filePath string
	var volumePath string

	r.mtx.RLock()
	defer r.mtx.RUnlock()

	err := r.db.QueryRow(ctx, "delete from songs where id = $1 returning file_path,volume_path", songID).Scan(&filePath, &volumePath)
	if err != nil {
		log.Println("DELETE DB SONG ERROR:", err)
		return "", "", errs.ServerError()
	}
	return filePath, volumePath, nil

}

func (r *SongRepository) EditSong(ctx context.Context, author, name, newVolume, newFile string, songID, duration int) (string, string, error) {
	setClauses := []string{}
	args := []interface{}{}
	argPos := 1

	if author != "" {
		setClauses = append(setClauses, fmt.Sprintf("author = $%d", argPos))
		args = append(args, author)
		argPos++
	}

	if name != "" {
		setClauses = append(setClauses, fmt.Sprintf("name = $%d", argPos))
		args = append(args, name)
		argPos++
	}

	if newVolume != "" {
		setClauses = append(setClauses, fmt.Sprintf("volume_path = $%d", argPos))
		args = append(args, newVolume)
		argPos++
	}

	if newFile != "" {
		setClauses = append(setClauses, fmt.Sprintf("file_path = $%d", argPos))
		args = append(args, newFile)
		argPos++
	}
	if duration != 0 {
		setClauses = append(setClauses, fmt.Sprintf("duration = $%d", argPos))
		args = append(args, duration)
		argPos++
	}
	args = append(args, songID)
	if len(setClauses) == 0 {
		return "", "", errs.New(http.StatusOK, "Нечего обновлять")
	}

	query := fmt.Sprintf("update songs set %s where id=$%v returning OLD.file_path,OLD.volume_path", strings.Join(setClauses, ", "), len(args))

	var oldFilePath string
	var oldVolumePath string

	err := r.db.QueryRow(ctx, query, args...).Scan(&oldFilePath, &oldVolumePath)
	if err != nil {
		log.Println("UPDATING SONG ERROR:", err)
		return "", "", errs.ServerError()
	}
	return oldFilePath, oldVolumePath, nil
}

func (r *SongRepository) UserOwnSong(ctx context.Context, userID, songID int) (bool, error) {
	var own bool
	err := r.db.QueryRow(ctx, "select exists(select 1 from songs s where s.id = $1 and s.user_id = $2)", songID, userID).Scan(&own)
	if err != nil {
		log.Println("SCAN USER OWN RESULT ERROR:", err)
		return false, errs.ServerError()
	}
	return own, nil
}

func (r *SongRepository) GetSong(ctx context.Context, userID, songID int) (*response.GetSongResponse, error) {
	query := `
	select 
		s.id,
		u.id as "user_id",
		u.name as "username",
		u.photo_file as "photo_file",
		s.author,
		s.name,
		s.duration,
		s.file_path,
		s.volume_path,
		s.is_available,
		sl.listens,
		(select count(1) from liked_songs where id = s.id and s.user_id <> $2) as "likes",
		EXISTS((select 1 from liked_songs ls where ls.song_id = s.id and ls.user_id = $2)) as "is_liked"
from songs s 
join users u on u.id = s.user_id
join songs_listens sl on sl.song_id = s.id
where s.id = $1 and (s.is_available = true or s.user_id = $2)`

	var findedSong response.GetSongResponse
	err := r.db.QueryRow(ctx, query, songID, userID).Scan(
		&findedSong.Id,
		&findedSong.UserInfo.Id,
		&findedSong.UserInfo.Name,
		&findedSong.UserInfo.PhotoFile,
		&findedSong.Author,
		&findedSong.Name,
		&findedSong.Duration,
		&findedSong.FilePath,
		&findedSong.VolumePath,
		&findedSong.IsAvailable,
		&findedSong.Listens,
		&findedSong.Likes,
		&findedSong.IsLiked,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errs.New(http.StatusNotFound, "Трек не найден, либо ограничен в доступе")
		}
		log.Println("GET SONG ERROR:", err)
		return nil, errs.ServerError()
	}
	return &findedSong, nil
}

func (r *SongRepository) GetSongs(ctx context.Context, userID, start, count int, sorted bool) ([]model.SongInGlobalSearch, error) {
	query := `
	SELECT 
		s.id, 
		u.id AS user_id,
		u.name as "username",
		u.photo_file as "photo_file",
		s.name, 
		s.duration, 
		s.author, 
		s.file_path, 
		s.volume_path, 
		sl.listens,
		COUNT(ls.song_id) AS likes,
		EXISTS((select 1 from liked_songs ls where song_id = s.id and ls.user_id = $1)) as "is_liked"
	FROM songs s
	INNER JOIN users u ON u.id = s.user_id
	INNER JOIN songs_listens sl ON sl.song_id = s.id
	LEFT JOIN liked_songs ls ON ls.song_id = s.id
	WHERE s.is_available = TRUE OR s.user_id = $1
	GROUP BY s.id, u.id, sl.listens
	`
	if sorted {
		query += "\n\torder by sl.listens desc"
	}
	query += `
	OFFSET $2
	LIMIT $3
	`
	var songs []model.SongInGlobalSearch
	rows, err := r.db.Query(ctx, query, userID, start, count)
	if err != nil {
		log.Println("ERROR WHILE GET SONGS:", err)
		return []model.SongInGlobalSearch{}, errs.ServerError()
	}
	defer rows.Close()

	for rows.Next() {
		var song model.SongInGlobalSearch
		err := rows.Scan(&song.Id, &song.UserInfo.Id, &song.UserInfo.Name, &song.UserInfo.Photo_file, &song.Name, &song.Duration, &song.Author, &song.FilePath, &song.VolumePath, &song.Listens, &song.Likes, &song.IsLiked)
		if err != nil {
			log.Println("SCAN SONG IN GLOBAL ERROR:", err)
			continue
		}
		songs = append(songs, song)
	}
	return songs, nil

}

func (r *SongRepository) GetSongsCount(ctx context.Context, userID int) (int, error) {
	query := `SELECT count (1)
		FROM songs s
		WHERE s.is_available = TRUE OR s.user_id = $1
	`

	songsCount := 0
	err := r.db.QueryRow(ctx, query, userID).Scan(&songsCount)
	if err != nil {
		log.Println("ERROR WHILE GET SONGS TOTAL COUNT:", err)
		return songsCount, errs.ServerError()
	}

	return songsCount, nil

}

func (r *SongRepository) ChangeStatus(ctx context.Context, songID int, status string) error {

	_, err := r.db.Exec(ctx, "update songs set is_available = $1 where id = $2", status, songID)
	if err != nil {
		log.Println("CHANGE AVAILABLILITY OF SONG ERROR:", err)
		return errs.ServerError()
	}
	return nil
}

func (r *SongRepository) Like(ctx context.Context, userID, songID int, likedAt time.Time) error {
	_, err := r.db.Exec(ctx, "insert into liked_songs (user_id,song_id,liked_at) values ($1,$2,$3)", userID, songID, likedAt)
	if err != nil {
		log.Println("LIKE SONG ERROR:", err)
		return errs.ServerError()
	}
	return nil
}

func (r *SongRepository) Unlike(ctx context.Context, userID, songID int) error {
	_, err := r.db.Exec(ctx, "delete from liked_songs where user_id = $1 and song_id = $2", userID, songID)
	if err != nil {
		log.Println("UNLIKE SONG ERROR:", err)
		return errs.ServerError()
	}
	return nil
}

func (r *SongRepository) GetLikesCount(ctx context.Context, userID, songID int) (int, error) {
	var count int
	err := r.db.QueryRow(ctx, "select count(1) from liked_songs ls join songs s on s.id = ls.song_id where ls.song_id = $1 and (s.is_available = true or s.user_id = $2)", songID, userID).Scan(&count)
	if err != nil {
		log.Println("GET SONG LIKES COUNT ERROR:", err)
		return 0, errs.ServerError()
	}
	return count, nil
}

func (r *SongRepository) CreateUserSongListens(ctx context.Context, userID, songID int) error {
	_, err := r.db.Exec(ctx, "insert into user_songs_listens (user_id,song_id,listens) values ($1,$2,$3)", userID, songID, 0)
	if err != nil {
		log.Println("ERROR CREATING USER SONG LISTENTS:", err)
		return errs.ServerError()
	}
	return nil
}

func (r *SongRepository) AddListenToUserSongListens(ctx context.Context, userID, songID int) error {
	_, err := r.db.Exec(ctx, "update user_songs_listens set listens = listens + 1, last_listen_time = $1 where user_id = $2 and song_id = $3", time.Now().UTC(), userID, songID)
	if err != nil {
		log.Println("ERROR UPDATING USER SONG LISTENS:", err)
		return errs.ServerError()
	}
	return nil
}

func (r *SongRepository) AddListenToSong(ctx context.Context, songID int) error {
	_, err := r.db.Exec(ctx, "update songs_listens set listens = listens + 1 where song_id = $1", songID)
	if err != nil {
		log.Println("ADD LISTEN TO SONG ERROR:", err)
		return errs.ServerError()
	}
	return nil
}

func (r *SongRepository) UserSongListensExists(ctx context.Context, userID, songID int) bool {
	var exists bool
	err := r.db.QueryRow(ctx, "select exists(select * from user_songs_listens where user_id = $1 and song_id = $2)", userID, songID).Scan(&exists)
	if err != nil {
		log.Println("USER SONG LISTENS EXISTS ERROR:", err)
		return false
	}
	return exists
}

func (r *SongRepository) GetUserSongListens(ctx context.Context, userID, songID int) (*model.UserSongListens, error) {
	var userListens model.UserSongListens
	err := r.db.QueryRow(ctx, "select * from user_songs_listens where user_id = $1 and song_id = $2", userID, songID).Scan(
		&userListens.ID, &userListens.UserID, &userListens.SongID, &userListens.Listens, &userListens.LastListenTime)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errs.ServerError()
		}
		log.Println("ERROR GET USER SONG LISTENS:", err)
		return nil, errs.ServerError()
	}

	return &userListens, nil
}

func (r *SongRepository) GetSongInfo(ctx context.Context, songID int) (*model.Song, error) {
	var song model.Song
	err := r.db.QueryRow(ctx, "select * from songs s where s.id = $1", songID).Scan(
		&song.Id, &song.User_id, &song.Author, &song.Name, &song.Duration, &song.File_path, &song.Volume_path, &song.Uploaded_at, &song.Is_available,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errs.New(http.StatusNotFound, "трек не найден")
		}
		log.Println("ERROR GET SONG INFO:", err)
		return nil, errs.ServerError()
	}

	return &song, nil
}

func (r *SongRepository) AccessToSong(ctx context.Context, userID, songID int) bool {
	var access bool
	err := r.db.QueryRow(ctx, "select exists(select 1 from songs where id = $1 and (user_id = $2 or is_available = true))", songID, userID).Scan(&access)
	if err != nil {
		log.Println("ERROR WHILE CHECK ACCESS TO SONG:", err)
		return false
	}
	return access
}

func (r *SongRepository) SongLiked(ctx context.Context, userID, songID int) bool {
	var liked bool
	err := r.db.QueryRow(ctx, "select exists(select * from liked_songs ls where ls.user_id = $1 and ls.song_id = $2)", userID, songID).Scan(&liked)
	if err != nil {
		log.Println("ERROR WHILE CHECK LIKED SONG:", err)
		return false
	}

	return liked
}
