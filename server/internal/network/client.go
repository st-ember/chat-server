package network

import (
	"fmt"
	"io"
	"net"

	"github.com/google/uuid"
	"github.com/st-ember/chat-server/shared/protocol"
)

type Client struct {
	id       uuid.UUID
	conn     net.Conn
	server   *Server
	outgoing chan *protocol.Message
	room     string
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

		// handle message
		switch msg.Header.Type {
		case protocol.Chat:
			// Broadcast chat message to all clients in the same room
			c.server.mu.RLock()
			for _, client := range c.server.clients {
				// locate clients in the same room
				if client.room == c.room && c.room != "" && client.id != c.id {
					client.outgoing <- msg
				}
			}
			c.server.mu.RUnlock()
		case protocol.Rooms:
			// Send list of rooms to client
			c.server.mu.RLock()
			var roomList []byte
			for roomName := range c.server.rooms {
				roomList = append(roomList, []byte(roomName+"\n")...)
			}
			c.server.mu.RUnlock()

			response := &protocol.Message{
				Header: protocol.Header{
					Type:   protocol.Rooms,
					Length: uint32(len(roomList)),
				},
				Content: roomList,
			}
			c.outgoing <- response
		case protocol.Room:
			c.server.mu.Lock()
			if c.server.rooms[string(msg.Content)] == nil {
				// Make new room
				c.server.rooms[string(msg.Content)] = make(map[uuid.UUID]*Client)
				fmt.Printf("Client %v created room %s\n", c.id, string(msg.Content))
			} else {
				// Join room
				c.server.rooms[string(msg.Content)][c.id] = c
				fmt.Printf("Client %v joined room %s\n", c.id, string(msg.Content))
			}
			if c.room != "" {
				// Leave previous room
				if clientsInRoom, ok := c.server.rooms[c.room]; ok {
					delete(clientsInRoom, c.id)
					fmt.Printf("Client %v left room %s\n", c.id, c.room)
				}
			}
			// Set current room
			c.room = string(msg.Content)
			c.server.mu.Unlock()
		case protocol.Leave:
			c.server.mu.Lock()
			if clientsInRoom, ok := c.server.rooms[string(msg.Content)]; ok {
				delete(clientsInRoom, c.id)
				fmt.Printf("Client %v left room %s\n", c.id, string(msg.Content))
			}
			c.server.mu.Unlock()
		default:
			fmt.Printf("Unhandled message type: %s\n", msg.Header.Type.String())
		}

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
