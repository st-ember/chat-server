package transport

import (
	"fmt"
	"io"
	"net"

	"github.com/google/uuid"
	"github.com/st-ember/chat-server/shared/protocol"
)

type Client struct {
	id       uuid.UUID
	nick     string
	conn     net.Conn
	roomID   uuid.UUID
	cmdCh    chan command
	outgoing chan *protocol.Message
}

func (c *Client) readLoop() {
	defer close(c.outgoing)

	for {
		msg, err := protocol.Decode(c.conn)
		if err != nil {
			// make exception for EOF errors (graceful disconnects)
			if err != io.EOF {
				fmt.Printf("Failed to decode message from client %v: %v\n", c.id, err)
			}
			return
		}

		// show message in console
		fmt.Println(string(msg.Content))

		cmd := command{
			clientID: c.id,
		}

		// handle message
		switch msg.Header.Type {
		case protocol.Chat:
			cmd.cmdType = chatToRoomCmd
			cmd.roomID = c.roomID
			cmd.content = msg.Content
		case protocol.Rooms:
			cmd.cmdType = listRooms
		case protocol.JoinRoom:
			cmd.cmdType = joinRoomCmd
			cmd.content = msg.Content
		case protocol.CreateRoom:
			cmd.cmdType = createRoomCmd
			cmd.content = msg.Content
		case protocol.Leave:
			cmd.cmdType = leaveRoomCmd
			cmd.content = msg.Content
		default:
			fmt.Printf("Unhandled message type: %v\n", msg.Header.Type)
		}

		c.cmdCh <- cmd
	}
}

func (c *Client) writeLoop() {
	for msg := range c.outgoing {
		encodedMesg, err := protocol.Encode(msg)
		if err != nil {
			fmt.Printf("error encoding message: %v", err)
			continue
		}

		if _, err := c.conn.Write(encodedMesg); err != nil {
			fmt.Printf("write error to %s: %v", c.conn.RemoteAddr(), err)
			return
		}
	}
}
