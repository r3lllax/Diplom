package model

type Playlist struct {
	Id          int    `json:"id" form:"id" validate:""`
	UserID      int    `json:"user_id" form:"user_id" validate:""`
	Title       string `json:"title" form:"title" validate:""`
	Description string `json:"description" form:"description" validate:""`
	VolumePath  string `json:"volume_path" form:"volume_path" validate:""`
	IsPrivate   bool   `json:"is_private" form:"is_private" validate:""`
	IsAvailable bool   `json:"is_available" form:"is_available" validate:""`
}

func NewPlaylist(id, userID int, title, description, volumePath string, isPrivate, isAvailable bool) *Playlist {
	return &Playlist{
		Id:          id,
		UserID:      userID,
		Title:       title,
		Description: description,
		VolumePath:  volumePath,
		IsPrivate:   isPrivate,
		IsAvailable: isAvailable,
	}
}
