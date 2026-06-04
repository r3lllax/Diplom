package model

import "time"

type Song struct {
	Id           int       `json:"id" validate:"unique=songs_id"`
	User_id      int       `json:"user_id" validate:"exists=users_id"`
	Author       string    `json:"author" validate:"min=1,max=255"`
	Name         string    `json:"name" validate:"min=1,max=255"`
	Duration     int       `json:"duration" validate:"min=10,max=900"`
	File_path    string    `json:"file_path"`
	Volume_path  string    `json:"volume_path"`
	Uploaded_at  time.Time `json:"uploaded_at"`
	Is_available bool      `json:"is_available"`
}

func NewSong(id, userID, duration int, author, name, filePath, volumePath string, uploaded_at time.Time, isAvailable bool) *Song {
	return &Song{
		Id:           id,
		User_id:      userID,
		Author:       author,
		Name:         name,
		Duration:     duration,
		File_path:    filePath,
		Volume_path:  volumePath,
		Uploaded_at:  uploaded_at,
		Is_available: isAvailable,
	}
}

// Variation of song model
type SongInLikes struct {
	Id         int           `json:"id" form:"id" validate:""`
	UserInfo   ShortUserInfo `json:"userInfo" form:"userInfo" validate:""`
	Author     string        `json:"author" form:"author" validate:""`
	Name       string        `json:"name" form:"name" validate:""`
	FilePath   string        `json:"file_path" form:"file_path" validate:""`
	VolumePath string        `json:"volume_path" form:"volume_path" validate:""`
	LikedAt    time.Time     `json:"liked_at" form:"liked_at" validate:""`
	IsLiked    bool          `json:"is_liked" form:"is_liked" validate:""`
}

type SongInGlobalSearch struct {
	Id         int           `json:"id" form:"id" validate:""`
	UserInfo   ShortUserInfo `json:"userInfo" form:"userInfo" validate:""`
	Name       string        `json:"name" form:"name" validate:""`
	Author     string        `json:"author" form:"author" validate:""`
	Duration   int           `json:"duration" form:"duration" validate:""`
	FilePath   string        `json:"file_path" form:"file_path" validate:""`
	VolumePath string        `json:"volume_path" form:"volume_path" validate:""`
	Listens    int           `json:"listens" form:"listens" validate:""`
	Likes      int           `json:"likes" form:"likes" validate:""`
	IsLiked    bool          `json:"is_liked" form:"is_liked" validate:""`
}
