package repository

import (
	errs "GIN/errors"
	"GIN/model"
	"GIN/tdo"
	"GIN/tdo/response"
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PlaylistRepository struct {
	db *pgxpool.Pool
}

func NewPlaylistRepository(dbase *pgxpool.Pool) *PlaylistRepository {
	return &PlaylistRepository{
		db: dbase,
	}
}

func (r *PlaylistRepository) CreatePlaylist(ctx context.Context, playlist *model.Playlist) (int, error) {
	query := `insert into playlists (user_id,title,description,volume_path,is_private,is_available)
values ($1,$2,$3,$4,$5,$6) returning id`
	var id int
	err := r.db.QueryRow(ctx, query, playlist.UserID, playlist.Title, playlist.Description, playlist.VolumePath, playlist.IsPrivate, playlist.IsAvailable).Scan(&id)
	if err != nil {
		log.Println("ERROR CREATING PLAYLIST:", err)
		return 0, errs.ServerError()
	}
	return id, nil
}

func (r *PlaylistRepository) LikePlaylist(ctx context.Context, userID, playlistID int) error {
	query := `insert into liked_playlists (user_id,playlist_id,liked_at)
values ($1,$2,$3)`
	_, err := r.db.Exec(ctx, query, userID, playlistID, time.Now())
	if err != nil {
		log.Println("ADD PLAYLIST TO LIKED ERROR:", err)
		return errs.ServerError()
	}
	return nil
}

func (r *PlaylistRepository) UserOwnPlaylist(ctx context.Context, userID, playlistID int) (bool, error) {
	var own bool
	err := r.db.QueryRow(ctx, "select exists(select 1 from playlists p where p.id = $1 and p.user_id = $2)", playlistID, userID).Scan(&own)
	if err != nil {
		log.Println("SCAN USER OWN PLAYLIST RESULT ERROR:", err)
		return false, errs.ServerError()
	}
	return own, nil
}

func (r *PlaylistRepository) EditPlaylist(ctx context.Context, title, newVolume string, description *string, playlistID int) (string, error) {
	var oldVolumePath string
	err := r.db.QueryRow(ctx, "SELECT volume_path FROM playlists WHERE id=$1", playlistID).Scan(&oldVolumePath)
	if err != nil {
		log.Println("GETTING OLD VOLUME PATH ERROR:", err)
		return "", errs.ServerError()
	}

	setClauses := []string{}
	args := []interface{}{}
	argPos := 1

	if title != "" {
		setClauses = append(setClauses, fmt.Sprintf("title = $%d", argPos))
		args = append(args, title)
		argPos++
	}

	if description != nil {
		setClauses = append(setClauses, fmt.Sprintf("description = $%d", argPos))
		args = append(args, description)
		argPos++
	}

	if newVolume != "" {
		setClauses = append(setClauses, fmt.Sprintf("volume_path = $%d", argPos))
		args = append(args, newVolume)
		argPos++
	}

	if len(setClauses) == 0 {
		return "", errs.New(http.StatusOK, "Нечего обновлять")
	}

	args = append(args, playlistID)
	query := fmt.Sprintf("UPDATE playlists SET %s WHERE id=$%d", strings.Join(setClauses, ", "), argPos)

	_, err = r.db.Exec(ctx, query, args...)
	if err != nil {
		log.Println("UPDATING PLAYLIST ERROR:", err)
		return "", errs.ServerError()
	}

	return oldVolumePath, nil
}

func (r *PlaylistRepository) DeletePlaylist(ctx context.Context, playlistID int) (string, error) {
	query := `delete from playlists where id = $1 returning volume_path`
	var oldVolume string
	err := r.db.QueryRow(ctx, query, playlistID).Scan(&oldVolume)
	if err != nil {
		log.Println("DELETE PLAYLIST ERROR:", err)
		return "", errs.ServerError()
	}
	return oldVolume, nil
}

func (r *PlaylistRepository) GetInfo(ctx context.Context, userID, playlistID int) (*tdo.PlaylistInfo, error) {
	query := `select 
  p.id,
  p.user_id,
  a.name as username,
  a.photo_file,
  p.title,
  p.description,
  p.volume_path,
  p.is_private,
  p.is_available,
  count(CASE WHEN s.is_available = true OR s.user_id = $2 THEN ps.id END) as "songs_count",
  coalesce(sum(CASE WHEN s.is_available = true OR s.user_id = $2 THEN s.duration ELSE 0 END), 0) as "playlist_duration",
  (select count(1) from liked_playlists lp where lp.playlist_id = p.id and lp.user_id <> p.user_id) as "likes_count",
  (exists(select 1 from liked_playlists lp where lp.playlist_id = p.id and lp.user_id = $2)) as "is_liked"
from playlists p
join users a on p.user_id = a.id
left join playlists_songs ps on ps.playlist_id = p.id
left join songs s on s.id = ps.song_id
where p.id = $1
  and ((p.is_private = false and p.is_available = true ) or p.user_id = $2)
group by p.id, p.user_id, a.name, a.photo_file, p.title, p.description, p.volume_path, p.is_private, p.is_available`
	var playlist tdo.PlaylistInfo
	err := r.db.QueryRow(ctx, query, playlistID, userID).Scan(
		&playlist.Id,
		&playlist.UserInfo.Id,
		&playlist.UserInfo.Name,
		&playlist.UserInfo.Photo_file,
		&playlist.Title,
		&playlist.Description,
		&playlist.VolumePath,
		&playlist.IsPrivate,
		&playlist.IsAvailable,
		&playlist.SongsCount,
		&playlist.Duration,
		&playlist.LikesCount,
		&playlist.IsLiked,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errs.New(http.StatusNotFound, "плейлист не существует или приватен")
		}
		log.Println("ERROR GET PLAYLIST INFO:", err)
		return nil, errs.ServerError()
	}

	return &playlist, nil
}

func (r *PlaylistRepository) GetPlaylists(ctx context.Context, userID, from, count int) ([]tdo.PlaylistInfo, error) {
	query := `select 
  p.id,
  p.user_id,
  a.name as username,
  a.photo_file,
  p.title,
  p.description,
  p.volume_path,
  p.is_private,
  p.is_available,
  count(CASE WHEN s.is_available = true OR s.user_id = $1 THEN ps.id END) as "songs_count",
  coalesce(sum(CASE WHEN s.is_available = true OR s.user_id = $1 THEN s.duration ELSE 0 END), 0) as "playlist_duration",
  (select count(1) from liked_playlists lp where lp.playlist_id = p.id and lp.user_id <> p.user_id) as "likes_count",
  (exists(select 1 from liked_playlists lp where lp.playlist_id = p.id and lp.user_id = $1)) as "is_liked"
from playlists p
join users a on p.user_id = a.id
left join playlists_songs ps on ps.playlist_id = p.id
left join songs s on s.id = ps.song_id
where ((p.is_private = false and p.is_available = true) or p.user_id = $1)
group by p.id, p.user_id, a.name, a.photo_file, p.title, p.description, p.volume_path, p.is_private, p.is_available
offset $2
limit $3
`
	var playlists []tdo.PlaylistInfo
	rows, err := r.db.Query(ctx, query, userID, from, count)
	if err != nil {
		log.Println("ERROR GET PLAYLISTS:", err)
		return nil, errs.ServerError()
	}
	defer rows.Close()
	for rows.Next() {
		var playlist tdo.PlaylistInfo
		err := rows.Scan(
			&playlist.Id,
			&playlist.UserInfo.Id,
			&playlist.UserInfo.Name,
			&playlist.UserInfo.Photo_file,
			&playlist.Title,
			&playlist.Description,
			&playlist.VolumePath,
			&playlist.IsPrivate,
			&playlist.IsAvailable,
			&playlist.SongsCount,
			&playlist.Duration,
			&playlist.LikesCount,
			&playlist.IsLiked,
		)
		if err != nil {
			log.Println("ERROR SCAN PLAYLIST IN PLAYLISTS:", err)
			continue
		}
		playlists = append(playlists, playlist)
	}
	return playlists, nil
}

func (r *PlaylistRepository) GetPlaylistsTotalCount(ctx context.Context, userID int) (int, error) {
	query := `
	select count(distinct p.id)
from playlists p
join users a on p.user_id = a.id
left join playlists_songs ps on ps.playlist_id = p.id
left join songs s on s.id = ps.song_id
where ((p.is_private = false and p.is_available = true ) or p.user_id = $1)
`
	var count int
	err := r.db.QueryRow(ctx, query, userID).Scan(&count)
	if err != nil {
		log.Println("ERROR GET PLAYLIST TOTAL COUNT:", err)
		return count, errs.ServerError()
	}

	return count, nil
}
func (r *PlaylistRepository) GetUserPlaylistsTotalCount(ctx context.Context, userID int) (int, error) {
	query := `
	select count(1)
from liked_playlists lp 
join playlists p on p.id = lp.playlist_id
join users u on u.id = p.user_id
where lp.user_id = $1
 	and (p.is_private = false or p.user_id = $1) 
  	and (p.is_available = true or p.user_id = $1)`
	var count int
	err := r.db.QueryRow(ctx, query, userID).Scan(&count)
	if err != nil {
		log.Println("ERROR GET PLAYLIST TOTAL COUNT:", err)
		return count, errs.ServerError()
	}

	return count, nil
}

func (r *PlaylistRepository) GetSongs(ctx context.Context, userID, playlistID, start, count int) ([]response.GetSongResponse, error) {
	query := `select 
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
	EXISTS(
    SELECT 1 
    FROM liked_songs ls
    WHERE ls.song_id = s.id AND ls.user_id = $2
) AS is_liked,
	ps.added_at
from playlists_songs ps
join songs s on s.id = ps.song_id
join users u on u.id = s.user_id
join playlists p on p.id = ps.playlist_id
join songs_listens sl on sl.song_id = s.id
where ps.playlist_id = $1 and (s.is_available = true or s.user_id = $2) and (p.is_private = false or p.user_id = $2)
offset $3
limit $4
`
	rows, err := r.db.Query(ctx, query, playlistID, userID, start, count)
	if err != nil {
		log.Println("ERROR GET PLAYLIST SONGS:", err)
		return []response.GetSongResponse{}, errs.ServerError()
	}
	defer rows.Close()

	var songs []response.GetSongResponse

	for rows.Next() {
		var song response.GetSongResponse
		err := rows.Scan(&song.Id, &song.UserInfo.Id, &song.UserInfo.Name, &song.UserInfo.PhotoFile, &song.Author, &song.Name, &song.Duration, &song.FilePath, &song.VolumePath, &song.IsAvailable, &song.Listens, &song.Likes, &song.IsLiked, &song.AddedAt)
		if err != nil {
			log.Println("ERROR SCAN PLAYLIST SONG:", err)
			continue
		}
		songs = append(songs, song)
	}
	return songs, nil
}
func (r *PlaylistRepository) UserPlaylistsWithSongContext(ctx context.Context, userID, songID int) ([]tdo.ShortPlaylistWithSongContext, error) {
	query := `
	SELECT 
    p.id, 
    p.title,
    EXISTS (
        SELECT 1 
        FROM playlists_songs ps 
        WHERE ps.playlist_id = p.id AND ps.song_id = $2
    ) AS has_song
FROM liked_playlists lp
JOIN playlists p ON p.id = lp.playlist_id
WHERE lp.user_id = $1 and p.user_id = $1;

`
	rows, err := r.db.Query(ctx, query, userID, songID)
	if err != nil {
		log.Println("ERROR GET PLAYLIST SONGS:", err)
		return []tdo.ShortPlaylistWithSongContext{}, errs.ServerError()
	}
	defer rows.Close()

	var playlists []tdo.ShortPlaylistWithSongContext

	for rows.Next() {
		var playlist tdo.ShortPlaylistWithSongContext
		err := rows.Scan(
			&playlist.Id,
			&playlist.Title,
			&playlist.HasSong,
		)
		if err != nil {
			log.Println("ERROR SCAN PLAYLISTS WITH SONG CONTEXT:", err)
			continue
		}
		playlists = append(playlists, playlist)
	}
	return playlists, nil
}
func (r *PlaylistRepository) GetSongsTotalRows(ctx context.Context, userID, playlistID int) (int, error) {
	query := `select count(1)
from playlists_songs ps
join songs s on s.id = ps.song_id
join users u on u.id = s.user_id
join playlists p on p.id = ps.playlist_id
join songs_listens sl on sl.song_id = s.id
where ps.playlist_id = $1 and (s.is_available = true or s.user_id = $2) and (p.is_private = false or p.user_id = $2)
`
	var count int
	err := r.db.QueryRow(ctx, query, playlistID, userID).Scan(&count)
	if err != nil {
		log.Println("ERROR GET PLAYLIST SONGS TOTAL ROWS COUNT:", err)
		return 0, errs.ServerError()
	}
	return count, nil
}

func (r *PlaylistRepository) ChangeStatus(ctx context.Context, playlistID int, status string) error {
	_, err := r.db.Exec(ctx, "update playlists set is_private = $1 where id = $2", status, playlistID)
	if err != nil {
		log.Println("EDIT PLAYLIST PRIVATE STATUS ERROR:", err)
		return errs.ServerError()
	}
	return nil
}

func (r *PlaylistRepository) PlaylistLiked(ctx context.Context, userID, playlistID int) bool {
	query := `select exists(select 1 from liked_playlists where user_id = $1 and playlist_id = $2)`
	var alreadyLiked bool
	err := r.db.QueryRow(ctx, query, userID, playlistID).Scan(&alreadyLiked)
	if err != nil {
		log.Println("CHECK LIKED PLAYLIST ERROR:", err)
		return false
	}
	return alreadyLiked
}

func (r *PlaylistRepository) AccessToPlaylist(ctx context.Context, userID, playlistID int) bool {
	var access bool
	err := r.db.QueryRow(ctx, "select exists(select 1 from playlists where id = $1 and (user_id = $2 or is_private = false))", playlistID, userID).Scan(&access)
	if err != nil {
		log.Println("ERROR WHILE CHECK ACCESS TO PLAYLIST:", err)
		return false
	}
	return access
}

func (r PlaylistRepository) Unlike(ctx context.Context, userID, playlistID int) error {
	query := `delete from liked_playlists lp where user_id = $1 and playlist_id = $2`
	_, err := r.db.Exec(ctx, query, userID, playlistID)
	if err != nil {
		log.Println("ERROR UNLIKE PLAYLIST:", err)
		return errs.ServerError()
	}
	return nil
}

func (r PlaylistRepository) AddSongToPlaylist(ctx context.Context, playlistID, songID int) error {
	query := `insert into playlists_songs (playlist_id,song_id,added_at) values ($1,$2,$3)`
	_, err := r.db.Exec(ctx, query, playlistID, songID, time.Now())
	if err != nil {
		log.Println("ERROR ADD SONG TO PLAYLIST:", err)
		return errs.ServerError()
	}
	return nil
}

func (r *PlaylistRepository) DeleteSongFromPlaylist(ctx context.Context, songID, playlistID int) error {
	query := `delete from playlists_songs where song_id = $1 and playlist_id = $2`
	_, err := r.db.Exec(ctx, query, songID, playlistID)
	if err != nil {
		log.Println("DELETE SONG FROM PLAYLIST ERROR:", err)
		return errs.ServerError()
	}
	return nil
}

func (r *PlaylistRepository) SongInPlaylist(ctx context.Context, songID, playlistID int) bool {
	query := `select exists(select 1 from playlists_songs where song_id = $1 and playlist_id = $2)`
	var alreadyInPlaylist bool
	err := r.db.QueryRow(ctx, query, songID, playlistID).Scan(&alreadyInPlaylist)
	if err != nil {
		log.Println("CHECK SONG IN PLAYLIST ERROR:", err)
		return false
	}
	return alreadyInPlaylist
}
