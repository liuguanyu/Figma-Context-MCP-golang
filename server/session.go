package server

import "time"

type Session struct {
	ID        string
	CreatedAt time.Time
}
