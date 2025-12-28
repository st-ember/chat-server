package domain

import "github.com/google/uuid"

type User struct {
	ID         uuid.UUID
	Nickname   string
	RemoteAddr string
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
	Timestamp int64
}
