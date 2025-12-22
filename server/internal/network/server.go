package network

import (
	"fmt"
	"net"
	"sync"

	"github.com/google/uuid"
	"github.com/st-ember/chat-server/shared/protocol"
)

type Server struct {
	listenAddr string
	listener   net.Listener
	clients    map[uuid.UUID]*Client
	rooms      map[string]map[uuid.UUID]*Client
	mu         sync.RWMutex
	quit       chan struct{}
}

func NewServer(listenAddr string) *Server {
	return &Server{
		listenAddr: listenAddr,
		clients:    make(map[uuid.UUID]*Client),
		quit:       make(chan struct{}),
		rooms:      make(map[string]map[uuid.UUID]*Client),
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

	// Block until quit signal is received
	<-s.quit

	// Graceful shutdown
	fmt.Println("Shutting down server...")
	s.mu.Lock()
	defer s.mu.Unlock()

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

func (s *Server) handleConnection(conn net.Conn) {
	client := &Client{
		id:       uuid.New(),
		conn:     conn,
		server:   s,
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
