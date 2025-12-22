package main

import (
	"fmt"

	"github.com/st-ember/chat-server/internal/network"
)

func main() {
	server := network.NewServer(":8081")
	if err := server.Start(); err != nil {
		fmt.Printf("error starting server: %v", err)
		return
	}

	defer server.Stop()
}
