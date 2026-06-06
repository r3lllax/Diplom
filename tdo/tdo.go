package tdo

import (
	"GIN/model"
	"time"
)

type UserSong struct {
	Id           int       `json:"id" validate:"unique=songs_id"`
	Author       string    `json:"author" validate:"min=1,max=255"`
	Name         string    `json:"name" validate:"min=1,max=255"`
	Duration     int       `json:"duration" validate:"min=10,max=900"`
	File_path    string    `json:"file_path"`
	Volume_path  string    `json:"volume_path"`
	Uploaded_at  time.Time `json:"uploaded_at"`
	Is_available bool      `json:"is_available"`
}

func NewUserSong(id, duration int, author, name, file_path, volume_path string, uploaded_at time.Time, available bool) *UserSong {
	return &UserSong{
		Id:           id,
		Author:       author,
		Name:         name,
		Duration:     duration,
		File_path:    file_path,
		Volume_path:  volume_path,
		Uploaded_at:  uploaded_at,
		Is_available: available,
	}
}

type UserSongListenStatistic struct {
	Id             int       `json:"id"`
	Author         string    `json:"author"`
	Name           string    `json:"name"`
	VolumePath     string    `json:"volume_path"`
	LastListenTime time.Time `json:"last_listen_time"`
	IsAvailable    bool      `json:"is_available"`
	Listens        int       `json:"listens"`
	ListenTime     int       `json:"listen_time"`
}

type UserTrackListenStatistic struct {
	Id          int    `json:"id"`
	Author      string `json:"author"`
	Name        string `json:"name"`
	VolumePath  string `json:"volume_path"`
	IsAvailable bool   `json:"is_available"`
	Listens     int    `json:"listens"`
}

type UserPlaylist struct {
	Id               int                 `json:"id" form:"" validate:""`
	Title            string              `json:"title" form:"" validate:""`
	Description      string              `json:"description" form:"" validate:""`
	VolumePath       string              `json:"volume_path" form:"" validate:""`
	IsPrivate        bool                `json:"is_private" form:"" validate:""`
	IsAvailable      bool                `json:"is_available" form:"" validate:""`
	AuthorInfo       model.ShortUserInfo `json:"user_info" form:"" validate:""`
	SongsCount       *int                `json:"songs_count" form:"" validate:""`
	PlaylistDuration *int                `json:"playlist_duration" form:"" validate:""`
	LikesCount       *int                `json:"likes_count" form:"" validate:""`
	IsLiked          bool                `json:"is_liked" form:"is_liked" validate:""`
}

type PlaylistInfo struct {
	Id          int                 `json:"id" form:"id" validate:""`
	UserInfo    model.ShortUserInfo `json:"userInfo" form:"userInfo" validate:""`
	Title       string              `json:"title" form:"title" validate:""`
	Description string              `json:"description" form:"description" validate:""`
	VolumePath  string              `json:"volume_path" form:"volume_path" validate:""`
	IsPrivate   bool                `json:"is_private" form:"is_private" validate:""`
	IsAvailable bool                `json:"is_available" form:"is_available" validate:""`
	SongsCount  *int                `json:"songs_count" form:"songs_count" validate:""`
	Duration    *int                `json:"duration" form:"duration" validate:""`
	LikesCount  *int                `json:"likes_count" form:"likes_count" validate:""`
	IsLiked     bool                `json:"is_liked" form:"is_liked" validate:""`
}
