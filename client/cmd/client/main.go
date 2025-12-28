package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"

	"github.com/st-ember/chat-server/shared/protocol"
)

func main() {
	conn, err := net.Dial("tcp", "localhost:8081")
	if err != nil {
		fmt.Printf("Failed to connect to server: %v", err)

		return
	}

	defer conn.Close()

	fmt.Println("Connected to server")

	// receive messages
	go func() {
		for {
			msg, err := protocol.Decode(conn)
			if err != nil {
				if err == io.EOF {
					log.Fatalf("Disconnected from server")
					return
				}
				fmt.Printf("Error decoding message: %v\n", err)
				return
			}

			fmt.Printf("Received message: %v", string(msg.Content))

		}
	}()

	// send messages
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("> ")
		input, _ := reader.ReadString('\n')
		text := strings.TrimSpace(input)

		if text == "" {
			continue
		}

		args := strings.SplitN(text, " ", 2)
		cmd := args[0]
		msg := &protocol.Message{
			Header: protocol.Header{},
		}

		if len(args) > 1 {
			msg.Content = []byte(args[1])
		}

		switch cmd {
		case "/rooms":
			msg.Header.Type = protocol.Rooms
		case "/room":
			msg.Header.Type = protocol.JoinRoom
		case "/leave":
			msg.Header.Type = protocol.Leave
		case "/quit":
			fmt.Println("Quitting...")
			return
		default:
			msg.Header.Type = protocol.Chat
			msg.Content = []byte(text)
		}

		msg.Header.Length = uint32(len(msg.Content))

		encodedMsg, err := protocol.Encode(msg)
		if err != nil {
			fmt.Printf("Error encoding message: %v", err)
			continue
		}

		_, err = conn.Write(encodedMsg)
		if err != nil {
			fmt.Printf("Error sending message: %v", err)
			return
		}
	}
}
