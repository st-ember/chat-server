package transport

import (
	"fmt"
	"net"
	"sync"

	"github.com/google/uuid"
	dm "github.com/st-ember/chat-server/internal/domain"
	"github.com/st-ember/chat-server/shared/protocol"
)

type Server struct {
	listenAddr string
	listener   net.Listener
	clients    map[uuid.UUID]*Client
	rooms      map[uuid.UUID]*Room
	store      Store
	mu         sync.RWMutex
	cmdCh      chan command
	quit       chan struct{}
}

func NewServer(
	listenAddr string,
	store Store,
) *Server {
	return &Server{
		listenAddr: listenAddr,
		store:      store,
		clients:    make(map[uuid.UUID]*Client),
		rooms:      make(map[uuid.UUID]*Room),
		cmdCh:      make(chan command, 10),
		quit:       make(chan struct{}),
	}
}

func (s *Server) Start() error {
	ln, err := net.Listen("tcp", s.listenAddr)
	if err != nil {
		return fmt.Errorf("failed to listen to address %s: %w", s.listenAddr, err)
	}
	defer ln.Close()
	s.listener = ln
	fmt.Printf("Server listening on %s\n", s.listenAddr)

	go s.acceptLoop()
	go s.cmdLoop()

	err = s.populateRooms()
	if err != nil {
		return fmt.Errorf("failed to populate rooms from store: %w", err)
	}

	// Block until quit signal is received
	<-s.quit

	// Graceful shutdown
	fmt.Println("Shutting down server...")

	// Close all client connections
	for _, client := range s.clients {
		client.conn.Close()
	}

	return nil
}

func (s *Server) Stop() {
	close(s.quit)
}

func (s *Server) acceptLoop() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			// Check if the server is shutting down
			select {
			case <-s.quit:
				return
			default:
				fmt.Printf("Failed to accept connection: %v\n", err)
				continue
			}
		}

		fmt.Println("new connection from:", conn.RemoteAddr())
		go s.handleConnection(conn)
	}
}

func (s *Server) cmdLoop() {
	for {
		cmd := <-s.cmdCh
		switch cmd.cmdType {
		case createRoomCmd:
			room := NewRoom(string(cmd.content))
			roomID := uuid.New()
			s.rooms[roomID] = room
			s.store.SaveRoom(&dm.Room{
				ID:   roomID,
				Name: string(cmd.content),
			})
		case joinRoomCmd:
			room := s.rooms[cmd.roomID]
			client, exists := s.clients[cmd.clientID]
			if exists {
				client.roomID = cmd.roomID
				room.AddClient(client)
			}
		case leaveRoomCmd:
			room := s.rooms[cmd.roomID]
			if room != nil {
				client, exists := s.clients[cmd.clientID]
				if exists {
					room.RemoveClient(client)
					client.roomID = uuid.Nil
				}
			}
		case chatToRoomCmd:
			client, exists := s.clients[cmd.clientID]
			if exists {
				room := s.rooms[cmd.roomID]
				if room != nil {
					msg := &protocol.Message{
						Header: protocol.Header{
							Type:   protocol.Chat,
							Length: uint32(len(cmd.content)),
						},
						Content: cmd.content,
					}
					for _, c := range room.clients {
						if c.id != client.id {
							c.outgoing <- msg
						}
					}
				}
			}
		case listRooms:
			client, exists := s.clients[cmd.clientID]
			if exists {
				roomList := ""
				for _, room := range s.rooms {
					roomList += room.name + "\n"
				}

				response := &protocol.Message{
					Header: protocol.Header{
						Type:   protocol.Rooms,
						Length: uint32(len(roomList)),
					},
					Content: []byte(roomList),
				}
				client.outgoing <- response
			}
		default:
			fmt.Printf("Unhandled command type: %s\n", cmd.cmdType)
		}
	}
}

func (s *Server) populateRooms() error {
	rooms, err := s.store.ListRooms()
	if err != nil {
		return fmt.Errorf("error listing rooms from store: %w", err)
	}

	for _, roomData := range rooms {
		room := NewRoom(roomData.Name)
		s.rooms[roomData.ID] = room
	}

	return nil
}

func (s *Server) handleConnection(conn net.Conn) {
	user, err := s.getOrCreateUser(conn.RemoteAddr().String())
	if err != nil {
		fmt.Printf("error getting or creating user: %v\n", err)
		conn.Close()
		return
	}
	fmt.Printf("User connected: %s (Nickname: %s)\n", user.ID, user.Nickname)

	client := &Client{
		id:       user.ID,
		nick:     user.Nickname,
		conn:     conn,
		outgoing: make(chan *protocol.Message, 10),
	}

	s.mu.Lock()
	s.clients[client.id] = client
	s.mu.Unlock()

	go client.writeLoop()
	// Block until readLoop exits
	client.readLoop()

	// Cleanup on disconnect
	s.mu.Lock()
	delete(s.clients, client.id)
	s.mu.Unlock()
	client.conn.Close()
	fmt.Println("connection closed from:", conn.RemoteAddr())
}

func (s *Server) getOrCreateUser(remoteAddr string) (*dm.User, error) {
	user, err := s.store.GetUserByRemoteAddr(remoteAddr)
	if err != nil {
		return nil, fmt.Errorf("error retrieving user by remote address: %w", err)
	}

	if user != nil {
		return user, nil
	}

	// new user
	newUser := &dm.User{
		ID:         uuid.New(),
		Nickname:   "guest",
		RemoteAddr: remoteAddr,
	}
	if err := s.store.SaveUser(newUser); err != nil {
		return nil, fmt.Errorf("error saving new user: %w", err)
	}

	return newUser, nil
}

type command struct {
	clientID uuid.UUID
	roomID   uuid.UUID
	cmdType  commandType
	content  []byte
}

type commandType string

const (
	createRoomCmd commandType = "create_room"
	joinRoomCmd   commandType = "join_room"
	leaveRoomCmd  commandType = "leave_room"
	chatToRoomCmd commandType = "chat_to_room"
	listRooms     commandType = "list_rooms"
)
