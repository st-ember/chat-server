package protocol

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
	Rooms
	Leave
	Chat
)

const MaxPayloadSize = 1024
