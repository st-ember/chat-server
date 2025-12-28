package transport

import (
	"github.com/google/uuid"
	dm "github.com/st-ember/chat-server/internal/domain"
)

type Store interface {
	UserStore
	RoomStore
	MessageStore
}

type UserStore interface {
	SaveUser(user *dm.User) error
	GetUserByID(id uuid.UUID) (*dm.User, error)
	GetUserByRemoteAddr(remoteAddr string) (*dm.User, error)
}

type RoomStore interface {
	SaveRoom(room *dm.Room) error
	ListRooms() ([]*dm.Room, error)
}

type MessageStore interface {
	SaveMessage(message *dm.Message) error
	ListMessagesByRoom(roomID uuid.UUID, limit int) ([]*dm.Message, error)
}
