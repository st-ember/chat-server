package transport

import "github.com/google/uuid"

type Room struct {
	name    string
	clients map[uuid.UUID]*Client
}

func NewRoom(name string) *Room {
	return &Room{
		name:    name,
		clients: make(map[uuid.UUID]*Client),
	}
}

func (r *Room) AddClient(c *Client) {
	r.clients[c.id] = c
}

func (r *Room) RemoveClient(c *Client) {
	delete(r.clients, c.id)
}
