package protocol

import "github.com/google/uuid"

type Message struct {
	Header  Header
	Content []byte
}

type Header struct {
	Length uint32
	Type   MessageType
}

type MessageType int8

const (
	JoinRoom MessageType = iota + 1
	CreateRoom
	ListRooms
	Leave
	Chat
	Identify
	Info
	Error
)

type Room struct {
	Name string
	ID   uuid.UUID
}

const MaxPayloadSize = 1024
