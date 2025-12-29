package domain

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID       uuid.UUID
	Nickname string
}

type Room struct {
	ID   uuid.UUID
	Name string
}

type Message struct {
	ID        uuid.UUID
	RoomID    uuid.UUID
	UserID    uuid.UUID
	Content   string
	Timestamp time.Time
}
