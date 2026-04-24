package request

import "mime/multipart"

type Registration struct {
	Name            string                `form:"name" json:"name" validate:"required,min=1,max=30,alphanum"`
	Email           string                `form:"email" json:"email" validate:"required,email,unique=users_email"`
	Password        string                `form:"password" json:"password" validate:"required,min=6,max=25"`
	PasswordConfirm string                `form:"passwordConfirm" json:"passwordConfirm" validate:"required,min=6,max=25,eqfield=Password"`
	Photo           *multipart.FileHeader `form:"photo_file" validate:"omitempty,photo,filesize"`
}

type Login struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type CreateSongRequest struct {
	Author string                `form:"author" json:"author" validate:"required,min=1,max=255"`
	Name   string                `form:"name" json:"name" validate:"required,min=1,max=255"`
	Volume *multipart.FileHeader `form:"volume" json:"volume" validate:"omitempty,photo,filesize"`
	Audio  *multipart.FileHeader `form:"audio" json:"audio" validate:"required,audio,filesize"`
}

type EditSongRequest struct {
	Author string                `form:"author" json:"author" validate:"omitempty,min=1,max=255"`
	Name   string                `form:"name" json:"name" validate:"omitempty,min=1,max=255"`
	Volume *multipart.FileHeader `form:"volume" json:"volume" validate:"omitempty,photo,filesize"`
	Audio  *multipart.FileHeader `form:"audio" json:"audio" validate:"omitempty,audio,filesize"`
}

type EditUserRequest struct {
	Name        string                `json:"name" form:"name" validate:"omitempty,min=1,max=30,alphanum"`
	Email       string                `json:"email" form:"email" validate:"omitempty,email,unique=users_email"`
	Photo       *multipart.FileHeader `json:"photo" form:"photo" validate:"omitempty,photo,filesize"`
	DeletePhoto bool                  `json:"deletePhoto" form:"deletePhoto" validate:"omitempty,boolean"`
}

type CreatePlaylistRequest struct {
	Title       string                `json:"title" form:"title" validate:"required,min=1,max=50,rueng"`
	Description string                `json:"description" form:"description" validate:"omitempty,max=1000,rueng"`
	Volume      *multipart.FileHeader `json:"volume" form:"volume" validate:"omitempty,filesize,photo"`
	IsPrivate   bool                  `json:"is_private" form:"is_private" validate:""`
}
type EditPlaylistRequest struct {
	Title       string                `json:"title" form:"title" validate:"omitempty,min=1,max=50,rueng"`
	Description *string               `json:"description" form:"description" validate:"omitempty,max=1000,rueng"`
	Volume      *multipart.FileHeader `json:"volume" form:"volume" validate:"omitempty,filesize,photo"`
}
