package response

import (
	"GIN/model"
	"time"
)

type GetSongResponse struct {
	Id          int                  `json:"id" form:"id" validate:""`
	UserInfo    GetSongResponse_User `json:"userInfo" form:"userInfo" validate:""`
	Author      string               `json:"author" form:"author" validate:""`
	Name        string               `json:"name" form:"name" validate:""`
	Duration    int                  `json:"duration" form:"duration" validate:""`
	FilePath    string               `json:"filePath" form:"filePath" validate:""`
	VolumePath  string               `json:"volumePath" form:"volumePath" validate:""`
	IsAvailable bool                 `json:"isAvailable" form:"isAvailable" validate:""`
	Listens     int                  `json:"listens" form:"listens" validate:""`
	Likes       int                  `json:"likes" form:"likes" validate:""`
	AddedAt     time.Time            `json:"added_at" form:"added_at" validate:""`
}
type GetSongsResponse struct {
	Data []model.SongInGlobalSearch `json:"data"`
}

type GetSongResponse_User struct {
	Id        int    `json:"id" form:"id" validate:""`
	Name      string `json:"name" form:"name" validate:""`
	PhotoFile string `json:"photo_file" form:"photo_file" validate:""`
}

type GetUser struct {
	Id             int       `json:"id" form:"id" validate:""`
	Name           string    `json:"name" form:"name" validate:""`
	Photo_file     string    `json:"photo_file" form:"photo_file" validate:""`
	Registrated_at time.Time `json:"registrated_at" form:"registrated_at" validate:""`
}
