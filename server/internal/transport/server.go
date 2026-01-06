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
			client, exists := s.clients[cmd.clientID]
			if exists {
				room := NewRoom(string(cmd.content))
				roomID := uuid.New()
				s.rooms[roomID] = room
				s.store.SaveRoom(&dm.Room{
					ID:   roomID,
					Name: string(cmd.content),
				})
				room.AddClient(client)
				content := fmt.Sprintf("room [%s] created", string(cmd.content))
				msg := &protocol.Message{
					Header: protocol.Header{
						Type:   protocol.Info,
						Length: uint32(len(content)),
					},
					Content: []byte(content),
				}
				client.outgoing <- msg
				fmt.Printf("client %s joined room %s\n", client.nick, room.name)
			}
		case joinRoomCmd:
			client, exists := s.clients[cmd.clientID]
			if exists {
				// Find specified room.
				room := s.rooms[uuid.UUID(cmd.content)]
				if room == nil {
					// Notify client
					content := fmt.Sprintf("room with ID %v does not exist", cmd.content)
					fmt.Println(content)
					msg := &protocol.Message{
						Header: protocol.Header{
							Type:   protocol.Error,
							Length: uint32(len(content)),
						},
						Content: []byte(content),
					}
					client.outgoing <- msg
					return
				}

				client.roomID = cmd.roomID
				room.AddClient(client)
				// Query chat history
				messages, err := s.store.ListMessagesByRoom(cmd.roomID, 20)
				if err != nil {
					fmt.Printf("error retrieving messages for room %v: %v\n", cmd.roomID, err)
					continue
				}
				// Send back chat history
				for _, msgData := range messages {
					msg := &protocol.Message{
						Header: protocol.Header{
							Type:   protocol.Chat,
							Length: uint32(len(msgData.Content)),
						},
						Content: []byte(msgData.Content),
					}
					client.outgoing <- msg
				}
				fmt.Printf("client %s joined room %s\n", client.nick, room.name)
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
				var rooms []protocol.Room
				for id, room := range s.rooms {
					rooms = append(rooms, protocol.Room{Name: room.name, ID: id})
				}

				content, err := protocol.EncodeCustom(rooms)
				if err != nil {
					fmt.Printf("Failed to encode rooms: %v\n", err)
				}

				response := &protocol.Message{
					Header: protocol.Header{
						Type:   protocol.ListRooms,
						Length: uint32(len(content)),
					},
					Content: content,
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

	var roomNameList string
	for _, roomData := range rooms {
		room := NewRoom(roomData.Name)
		s.rooms[roomData.ID] = room
		roomNameList += roomData.Name + "\r\n"
	}

	fmt.Printf("Rooms from db include: %s\n", roomNameList)

	return nil
}

func (s *Server) handleConnection(conn net.Conn) {
	// wait for identify message
	msg, err := protocol.DecodeMsg(conn)
	if err != nil {
		fmt.Printf("Failed to decode identify message: %v\n", err)
		conn.Close()
		return
	}

	// first message must be identify
	if msg.Header.Type != protocol.Identify {
		fmt.Printf("First message from client is not Identify\n")
		conn.Close()
		return
	}

	var client *Client
	// new client, create user
	if len(msg.Content) == 0 {
		client = &Client{
			id:       uuid.New(),
			conn:     conn,
			cmdCh:    s.cmdCh,
			outgoing: make(chan *protocol.Message, 10),
		}

		user := &dm.User{
			ID:       client.id,
			Nickname: "Guest_" + client.id.String()[:8],
		}

		err := s.store.SaveUser(user)
		if err != nil {
			fmt.Printf("Error saving new user to store: %v\n", err)
			conn.Close()
			return
		}

		// send back identify response with new uuid
		identifyResp := &protocol.Message{
			Header: protocol.Header{
				Type:   protocol.Identify,
				Length: uint32(len(client.id.String())),
			},
			Content: []byte(client.id.String()),
		}

		encodedResp, err := protocol.EncodeMsg(identifyResp)
		if err != nil {
			fmt.Printf("Error encoding identify response: %v\n", err)
			conn.Close()
			return
		}

		_, err = conn.Write(encodedResp)
		if err != nil {
			fmt.Printf("Error sending identify response: %v\n", err)
			conn.Close()
			return
		}
	} else {
		// identify client with token (uuid)
		uuid, err := uuid.Parse(string(msg.Content))
		if err != nil {
			fmt.Printf("Invalid UUID format from client: %s\n", string(msg.Content))
			conn.Close()
			return
		}

		existingUser, err := s.store.GetUserByID(uuid)
		if err != nil {
			fmt.Printf("Error retrieving user from store: %v\n", err)
			conn.Close()
			return
		}

		client = &Client{
			id:       existingUser.ID,
			nick:     existingUser.Nickname,
			conn:     conn,
			cmdCh:    s.cmdCh,
			outgoing: make(chan *protocol.Message, 10),
		}
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
