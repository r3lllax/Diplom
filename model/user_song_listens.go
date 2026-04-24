package model

import "time"

type UserSongListens struct {
	ID             int
	UserID         int
	SongID         int
	Listens        int
	LastListenTime *time.Time
}
