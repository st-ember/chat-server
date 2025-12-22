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
	Room MessageType = iota + 1
	Rooms
	Leave
	Chat
)

func (mt MessageType) String() string {
	switch mt {
	case Room:
		return "Room"
	case Rooms:
		return "Rooms"
	case Leave:
		return "Leave"
	case Chat:
		return "Chat"
	default:
		return "Unknown"
	}
}

const MaxPayloadSize = 1024
