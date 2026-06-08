package model

import "time"

type User struct {
	Id             int       `json:"id" validate:"unique=users_id"`
	Name           string    `json:"name" validate:"min=1,max=30,alphanum"`
	Email          string    `json:"email" validate:"email,unique=users_email"`
	Password       string    `json:"password" validate:"min=6,max=255"`
	Is_private     bool      `json:"is_private"`
	Photo_file     string    `json:"photo_file"`
	Registrated_at time.Time `json:"registrated_at"`
}

func NewUser(id int, name, email, password string, is_private bool, photo_file string, registrated_at time.Time) *User {
	return &User{
		Id:             id,
		Name:           name,
		Email:          email,
		Password:       password,
		Is_private:     is_private,
		Photo_file:     photo_file,
		Registrated_at: registrated_at,
	}
}

type ShortUserInfo struct {
	Id         int    `json:"id" form:"id" validate:""`
	Name       string `json:"name" form:"name" validate:""`
	Photo_file string `json:"photo_file" form:"photo_file" validate:""`
}
type UserProfile struct {
	Id         int    `json:"id" form:"id" validate:""`
	Name       string `json:"name" form:"name" validate:""`
	Email      string `json:"email" validate:"email,unique=users_email"`
	Photo_file string `json:"photo_file" form:"photo_file" validate:""`
	Is_private bool   `json:"is_private"`
	Listens    int    `json:"listens"`
	ListenTime int    `json:"listenTime"`
	LikesCount int    `json:"likesCount"`
	SongsCount int    `json:"songsCount"`
}
