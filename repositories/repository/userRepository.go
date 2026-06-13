package repository

import (
	errs "GIN/errors"
	"GIN/model"
	"GIN/tdo"
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(dbase *pgxpool.Pool) *UserRepository {
	return &UserRepository{
		db: dbase,
	}
}

func (r *UserRepository) GetUserByEmail(ctx context.Context, email string) (*model.User, error) {
	var findedUser model.User
	err := r.db.QueryRow(ctx, "select * from users u where u.email=$1", email).Scan(&findedUser.Id, &findedUser.Name, &findedUser.Email, &findedUser.Password, &findedUser.Is_private, &findedUser.Photo_file, &findedUser.Registrated_at)
	if err != nil {
		return &model.User{}, errs.New(http.StatusUnauthorized, "Неправильный логин или пароль")
	}

	return &findedUser, nil
}

func (r *UserRepository) GetUser(ctx context.Context, targetUserID, userID int) (string, error) {
	var findedUser string
	var query = `
	WITH user_data AS (
    SELECT 
        u.id,
        u.name,
        u.email,
        u.photo_file,
        u.registrated_at,
        u.is_private,
        (SELECT COALESCE(SUM(sl.listens), 0) FROM songs s LEFT JOIN songs_listens sl ON s.id = sl.song_id WHERE s.user_id = u.id) AS listens,
        (SELECT COALESCE(SUM(usl.listens * s.duration), 0) FROM user_songs_listens usl JOIN songs s ON usl.song_id = s.id WHERE usl.user_id = u.id) AS listenTime,
        (SELECT COALESCE(COUNT(*), 0) FROM liked_songs ls JOIN songs s ON ls.song_id = s.id WHERE s.user_id = u.id) AS likesCount,
        (SELECT COUNT(*) FROM songs WHERE user_id = u.id) AS songsCount
    FROM users u
    WHERE u.id = $1
)
SELECT jsonb_build_object(
    'id', id,
    'name', name,
    'photo_file', photo_file,
    'is_private', is_private
) ||
CASE
    WHEN $2 = $1 THEN jsonb_build_object(
        'email', email,
        'registrated_at', registrated_at,
        'listens', listens,
        'listenTime', listenTime,
        'likesCount', likesCount,
        'songsCount', songsCount
    )
    WHEN $2 != $1 AND is_private = false THEN jsonb_build_object(
        'registrated_at', registrated_at
    )
    ELSE '{}'::jsonb
END AS result
FROM user_data;
	`
	err := r.db.QueryRow(ctx, query, targetUserID, userID).Scan(&findedUser)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", errs.New(http.StatusNotFound, "пользователь не найден, или его профиль приватен")
		}
		log.Println("GET USER ERROR:", err)
		return "", errs.ServerError()
	}
	return findedUser, nil
}

func (r *UserRepository) EditUser(ctx context.Context, userID int, name, photo_file, email string, needDeletePhoto bool) (string, error) {
	var oldPhotoFile string
	err := r.db.QueryRow(ctx, "SELECT photo_file FROM users WHERE id=$1", userID).Scan(&oldPhotoFile)
	if err != nil {
		log.Println("GETTING OLD PHOTO FILE ERROR:", err)
		return "", errs.ServerError()
	}

	setClauses := []string{}
	args := []interface{}{}
	argPos := 1

	if name != "" {
		setClauses = append(setClauses, fmt.Sprintf("name = $%d", argPos))
		args = append(args, name)
		argPos++
	}

	newPhotoFile := photo_file
	if needDeletePhoto {
		newPhotoFile = ""
	}

	if newPhotoFile != "" || needDeletePhoto {
		setClauses = append(setClauses, fmt.Sprintf("photo_file = $%d", argPos))
		args = append(args, newPhotoFile)
		argPos++
	}

	if email != "" {
		setClauses = append(setClauses, fmt.Sprintf("email = $%d", argPos))
		args = append(args, email)
		argPos++
	}

	if len(setClauses) == 0 {
		return "", errs.New(http.StatusOK, "Нечего обновлять")
	}

	args = append(args, userID)
	query := fmt.Sprintf("UPDATE users SET %s WHERE id=$%d", strings.Join(setClauses, ", "), argPos)
	_, err = r.db.Exec(ctx, query, args...)
	if err != nil {
		log.Println("UPDATING USER ERROR:", err)
		return "", errs.ServerError()
	}

	return oldPhotoFile, nil
}

func (r *UserRepository) Exists(ctx context.Context, userID int) (bool, error) {
	var res bool
	err := r.db.QueryRow(ctx, "select exists(select 1 from users where id = $1)", userID).Scan(&res)
	if err != nil {
		log.Println("USER EXISTS CHECK ERROR:", err)
		return false, errs.ServerError()
	}
	return res, nil
}

func (r *UserRepository) DeleteUser(ctx context.Context, userID int) error {
	_, err := r.db.Exec(ctx, "delete from users where id = $1", userID)
	if err != nil {
		log.Println("DELETE USER ERROR:", err)
		return errs.ServerError()
	}
	return nil
}

func (r *UserRepository) GetLikedSongs(ctx context.Context, targetUserID, userID, start, count int) ([]model.SongInLikes, error) {

	query := `SELECT song_id AS "song_id", user_id, username, user_pfp, author, song_name, file_path, volume_path, liked_at,
    EXISTS(SELECT 1 FROM liked_songs ls WHERE ls.song_id = sub.song_id AND ls.user_id = $2) AS "is_liked"
FROM (
    SELECT ls.song_id, ls.user_id, u.name AS "username", u.photo_file AS "user_pfp",
           s.author, s.name AS "song_name", s.file_path, s.volume_path, s.duration, s.is_available, ls.liked_at
    FROM liked_songs ls
    JOIN songs s ON s.id = ls.song_id
    JOIN users u ON u.id = ls.user_id
    WHERE ls.user_id = $1 AND (u.is_private = false OR u.id = $2)
) AS sub
WHERE is_available = true OR user_id = $2
ORDER BY liked_at DESC
OFFSET $3
LIMIT $4`

	rows, err := r.db.Query(ctx, query, targetUserID, userID, start, count)
	if err != nil {
		log.Println("ERROR WHILE GET USER LIKES:", err)
		return []model.SongInLikes{}, errs.ServerError()
	}
	defer rows.Close()

	var likedSongs []model.SongInLikes

	for rows.Next() {
		var song model.SongInLikes
		err := rows.Scan(&song.Id, &song.UserInfo.Id, &song.UserInfo.Name, &song.UserInfo.Photo_file, &song.Author, &song.Name, &song.FilePath, &song.VolumePath, &song.LikedAt, &song.IsLiked)
		if err != nil {
			log.Println("SCAN SONG IN LIKED SONGS ERROR:", err)
			continue
		}
		likedSongs = append(likedSongs, song)
	}

	return likedSongs, nil

}

func (r *UserRepository) GetUserLikesDuration(ctx context.Context, userID int) (int, error) {
	query := `select coalesce(sum(s.duration),0) as "duration"
from liked_songs ls
join songs s on s.id = ls.song_id
where ls.user_id = $1
	`

	var duration int

	err := r.db.QueryRow(ctx, query, userID).Scan(&duration)
	if err != nil {
		log.Println("ERROR USER LIKES DURATION:", err)
		return 0, errs.ServerError()
	}

	return duration, nil
}
func (r *UserRepository) GetUserLikesTotalRows(ctx context.Context, targetUserID, userID int) (int, error) {
	query := `select count(1)
from liked_songs ls
join songs s on s.id = ls.song_id
where ls.user_id = $1 and (s.is_available = true or s.user_id = $2)`

	var rows int

	err := r.db.QueryRow(ctx, query, targetUserID, userID).Scan(&rows)
	if err != nil {
		log.Println("ERROR GET USER LIKES TOTAL ROWS:", err)
		return 0, errs.ServerError()
	}

	return rows, nil
}
func (r *UserRepository) GetUserSongs(ctx context.Context, userID, start, count int) ([]model.Song, error) {

	query := `select id,author,name,duration,file_path,volume_path,uploaded_at,is_available
from songs 
where user_id = $1
order by uploaded_at desc
offset $2
limit $3
`

	rows, err := r.db.Query(ctx, query, userID, start, count)
	if err != nil {
		log.Println("ERROR WHILE GET USER SONGS:", err)
		return []model.Song{}, errs.ServerError()
	}
	defer rows.Close()

	var userSongs []model.Song

	for rows.Next() {
		var song model.Song
		err := rows.Scan(&song.Id, &song.Author, &song.Name, &song.Duration, &song.File_path, &song.Volume_path, &song.Uploaded_at, &song.Is_available)
		if err != nil {
			log.Println("SCAN SONG IN USER SONGS ERROR:", err)
			continue
		}
		userSongs = append(userSongs, song)
	}

	return userSongs, nil

}

func (r *UserRepository) EditPrivateStatus(ctx context.Context, status bool, userID int) error {
	_, err := r.db.Exec(ctx, "update users set is_private = $1 where id = $2", status, userID)
	if err != nil {
		log.Println("EDIT USER PRIVATE STATUS ERROR:", err)
		return errs.ServerError()
	}
	return nil
}
func (r *UserRepository) GetUserLastListenTracks(ctx context.Context, userID int) ([]tdo.UserSongListenStatistic, error) {

	query := `select ul.song_id,s.Author,s.name,s.volume_path,ul.last_listen_time,s.is_available,ul.listens,(s.duration * ul.listens) as "listen_time"
from user_songs_listens ul
join songs s on s.id = ul.song_id
where ul.user_id = $1
order by ul.last_listen_time desc
limit 5
`

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		log.Println("GET USER LAST LISTENS ERROR:", err)
		return []tdo.UserSongListenStatistic{}, errs.ServerError()
	}
	defer rows.Close()

	var list []tdo.UserSongListenStatistic

	for rows.Next() {
		var song tdo.UserSongListenStatistic
		err := rows.Scan(&song.Id, &song.Author, &song.Name, &song.VolumePath, &song.LastListenTime, &song.IsAvailable, &song.Listens, &song.ListenTime)
		if err != nil {
			log.Println("SCAN SONG FROM USER LAST LISTENS ERROR:", err)
			continue
		}
		list = append(list, song)

	}

	return list, nil
}

func (r *UserRepository) GetUserListenStatistics(ctx context.Context, userID, start, count int, countSort bool) ([]tdo.UserSongListenStatistic, error) {

	query := `select ul.song_id,s.Author,s.name,s.volume_path,ul.last_listen_time,s.is_available,ul.listens,(s.duration * ul.listens) as "listen_time"
from user_songs_listens ul
join songs s on s.id = ul.song_id 
where ul.user_id = $1
`
	if countSort {
		query += `order by listen_time desc `
	} else {
		query += `order by ul.listens desc `

	}

	query += `
	offset $2
	limit $3
	`

	rows, err := r.db.Query(ctx, query, userID, start, count)
	if err != nil {
		log.Println("GET USER LISTEN STATISTICS ERROR:", err)
		return []tdo.UserSongListenStatistic{}, errs.ServerError()
	}
	defer rows.Close()

	var list []tdo.UserSongListenStatistic

	for rows.Next() {
		var song tdo.UserSongListenStatistic
		err := rows.Scan(&song.Id, &song.Author, &song.Name, &song.VolumePath, &song.LastListenTime, &song.IsAvailable, &song.Listens, &song.ListenTime)
		if err != nil {
			log.Println("SCAN SONG FROM USER LISTEN STATISTIC ERROR:", err)
			continue
		}
		list = append(list, song)

	}

	return list, nil
}

func (r *UserRepository) GetUserGeneralListenStats(ctx context.Context, userID int) (int, int, int, int, error) {
	query := `
        SELECT 
            COALESCE(SUM(ul.listens), 0) AS listenSongsCount,
            COALESCE(SUM(s.duration * ul.listens), 0) AS listen_time,
            COALESCE((
                SELECT COUNT(1)
                FROM liked_songs ls
                WHERE ls.user_id = $1
            ), 0) AS likesCount,
            COALESCE((
                SELECT COUNT(1)
                FROM liked_songs ls
                JOIN songs s ON s.id = ls.song_id
                WHERE s.user_id = $1 AND ls.user_id <> $1
            ), 0) AS userSongsLikes
        FROM user_songs_listens ul
        JOIN songs s ON s.id = ul.song_id
        WHERE ul.user_id = $1
    `

	var listenSongsCount, listenTime, likesCount, userSongsLikes int

	err := r.db.QueryRow(ctx, query, userID).Scan(&listenSongsCount, &listenTime, &likesCount, &userSongsLikes)
	if err != nil {
		log.Println("ERROR GETTING USER GENERAL STATISTICS:", err)
		return 0, 0, 0, 0, errs.ServerError()
	}

	return listenSongsCount, listenTime, likesCount, userSongsLikes, nil
}

func (r *UserRepository) GetUserLikesCount(ctx context.Context, userID int) (int, error) {
	query := `select count(1) from liked_songs where user_id = $1`

	var count int

	err := r.db.QueryRow(ctx, query, userID).Scan(&count)
	if err != nil {
		log.Println("ERROR GETTING USER LIKES COUNT:", err)
		return 0, errs.ServerError()
	}

	return count, nil
}

func (r *UserRepository) GetUserTracksListenStatistics(ctx context.Context, userID, start, count int) ([]tdo.UserTrackListenStatistic, error) {
	query := `select s.id,s.author,s.name,s.volume_path,sl.listens,s.is_available
from songs_listens sl
join songs s on s.id = sl.song_id
where s.user_id = $1
order by sl.listens desc
offset $2
limit $3`

	rows, err := r.db.Query(ctx, query, userID, start, count)
	if err != nil {
		log.Println("GET USER TRACKS LISTEN STATISTICS ERROR:", err)
		return []tdo.UserTrackListenStatistic{}, errs.ServerError()
	}
	defer rows.Close()

	var list []tdo.UserTrackListenStatistic

	for rows.Next() {
		var song tdo.UserTrackListenStatistic
		err := rows.Scan(&song.Id, &song.Author, &song.Name, &song.VolumePath, &song.Listens, &song.IsAvailable)
		if err != nil {
			log.Println("SCAN SONG FROM USER TRACKS LISTEN STATISTIC ERROR:", err)
			continue
		}
		list = append(list, song)

	}

	return list, nil
}

func (r *UserRepository) GetUserTracksGeneralInfo(ctx context.Context, userID int) (int, int, int, int, error) {
	query := `
	SELECT
    (SELECT COUNT(1) FROM songs s WHERE s.user_id = $1) AS songsCount,
    (SELECT COUNT(1) 
     FROM liked_songs ls 
     JOIN songs s ON s.id = ls.song_id 
     WHERE s.user_id = $1) AS tracksLikes,
    (SELECT SUM(sl.listens) 
     FROM songs_listens sl 
     JOIN songs s ON s.id = sl.song_id 
     WHERE s.user_id = $1) AS tracksListensCount,
    (SELECT COUNT(1) 
     FROM songs s 
     WHERE s.user_id = $1 AND s.is_available = true) AS publicTracks;
	`

	var songsCount int
	var tracksLikes int
	var tracksListensCount int
	var publicTracksCount int

	err := r.db.QueryRow(ctx, query, userID).Scan(&songsCount, &tracksLikes, &tracksListensCount, &publicTracksCount)
	if err != nil {
		log.Println("ERROR GETTING USER TRACKS GENERAL INFO:", err)
		return 0, 0, 0, 0, errs.ServerError()
	}

	return songsCount, tracksLikes, tracksListensCount, publicTracksCount, nil
}

func (r *UserRepository) GetUserPlaylists(ctx context.Context, userID, targetUserID, start, count int) ([]tdo.UserPlaylist, error) {

	query := `select 
	p.id, 
	p.title, 
	p.description, 
	p.volume_path, 
	p.is_private, 
	p.is_available, 
	p.user_id as author_id,
	a.name as author_name,
	a.photo_file as author_photo_file,
	count(ps.id) as "songs_count",
	coalesce(sum(s.duration),0) as "playlist_duration",
	(select count(1) from liked_playlists lp where lp.playlist_id = p.id and lp.user_id <> $1) as likes_count,
  	(exists(select 1 from liked_playlists lp where lp.playlist_id = p.id and lp.user_id = $1)) as "is_liked"
		from liked_playlists lp
join playlists p on p.id = lp.playlist_id
join users u on u.id = lp.user_id
left join users a on p.user_id = a.id
left join playlists_songs ps on ps.playlist_id = p.id
left join songs s on ps.song_id = s.id
where lp.user_id = $1 
  and (u.is_private = false or u.id = $2) 
  and (p.is_private = false or p.user_id = $2) 
  and (p.is_available = true or p.user_id = $2)
group by p.id, p.title, p.description, p.volume_path, p.is_private, p.is_available, p.user_id, a.name, a.photo_file
offset $3
limit $4
`
	rows, err := r.db.Query(ctx, query, targetUserID, userID, start, count)
	if err != nil {
		log.Println("GET USER PLAYLISTS ERROR:", err)
		return []tdo.UserPlaylist{}, errs.ServerError()
	}
	defer rows.Close()

	var playlists []tdo.UserPlaylist

	for rows.Next() {
		var playlist tdo.UserPlaylist
		err := rows.Scan(
			&playlist.Id,
			&playlist.Title,
			&playlist.Description,
			&playlist.VolumePath,
			&playlist.IsPrivate,
			&playlist.IsAvailable,
			&playlist.AuthorInfo.Id,
			&playlist.AuthorInfo.Name,
			&playlist.AuthorInfo.Photo_file,
			&playlist.SongsCount,
			&playlist.PlaylistDuration,
			&playlist.LikesCount,
			&playlist.IsLiked,
		)
		if err != nil {
			log.Println("SCAN PLAYLIST FROM USER PLAYLISTS ERROR:", err)
			continue
		}
		playlists = append(playlists, playlist)
	}

	return playlists, nil
}
