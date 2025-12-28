package main

import (
	"fmt"

	"github.com/st-ember/chat-server/internal/infra/storage"
	"github.com/st-ember/chat-server/internal/transport"
)

func main() {
	// db init
	postgres, err := storage.NewDB("postgres://postgres:Br3akD3na@localhost:5432/chat?sslmode=disable")
	if err != nil {
		fmt.Printf("error connecting to postgres database: %v", err)
		return
	}

	// server init
	server := transport.NewServer(":8081", postgres)
	if err := server.Start(); err != nil {
		fmt.Printf("error starting server: %v", err)
		return
	}

	defer server.Stop()
}
