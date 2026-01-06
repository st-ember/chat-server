package tui

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"

	"github.com/google/uuid"
	"github.com/st-ember/chat-server/client/internal/infra"
	ptc "github.com/st-ember/chat-server/shared/protocol"
)

type ClientTUI struct {
	rooms []ptc.Room
}

func NewClientTUI() *ClientTUI {
	return &ClientTUI{}
}

func (ct *ClientTUI) Start() {
	conn, err := net.Dial("tcp", "localhost:8081")
	if err != nil {
		fmt.Printf("Failed to connect to server: %v", err)
		return
	}

	defer conn.Close()
	fmt.Println("Connected to server")

	token, err := infra.Token()
	if err != nil {
		fmt.Printf("Failed to read from token file: %v", err)
		return
	}

	// send identify message with or without token
	msg := &ptc.Message{
		Header: ptc.Header{
			Type:   ptc.Identify,
			Length: uint32(len(token)),
		},
		Content: []byte(token),
	}
	encodedMsg, err := ptc.EncodeMsg(msg)
	if err != nil {
		fmt.Printf("Error encoding message: %v", err)
		return
	}

	_, err = conn.Write(encodedMsg)
	if err != nil {
		fmt.Printf("Error sending identify message: %v", err)
		return
	}

	// receive messages
	go func() {
		for {
			msg, err := ptc.DecodeMsg(conn)
			if err != nil {
				if err == io.EOF {
					log.Fatalf("Disconnected from server")
					return
				}
				fmt.Printf("Error decoding message: %v\n", err)
				return
			}
			switch msg.Header.Type {
			case ptc.Identify:
				newToken, err := uuid.Parse(string(msg.Content))
				if err != nil {
					fmt.Printf("Invalid UUID format from server: %s\n", string(msg.Content))
					return
				}

				err = infra.StoreToken(newToken)
				if err != nil {
					fmt.Printf("Failed to store token: %v\n", err)
					return
				}
			case ptc.ListRooms:
				var rooms []ptc.Room
				err := ptc.DecodeCustom(msg.Content, &rooms)
				if err != nil {
					fmt.Printf("Failed to decode rooms: %v \n", err)
					return
				}
				ct.rememberRooms(rooms)
				// Show room names
				var roomNames string
				for _, room := range rooms {
					roomNames += room.Name + "\n"
				}
				fmt.Printf("Available rooms: %v", roomNames)
			default:
				fmt.Println(string(msg.Content))
			}
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
		msg := &ptc.Message{
			Header: ptc.Header{},
		}

		if len(args) > 1 {
			msg.Content = []byte(args[1])
		}

		switch cmd {
		case "/create-room":
			msg.Header.Type = ptc.CreateRoom
		case "/rooms":
			msg.Header.Type = ptc.ListRooms
		case "/join-room":
			msg.Header.Type = ptc.JoinRoom
			var roomName = args[1]
			var roomID uuid.UUID
			for _, room := range ct.rooms {
				if room.Name == roomName {
					roomID = room.ID
				}
			}
			if roomID == uuid.Nil {
				fmt.Printf("No rooms match name [%s]", roomName)
				return
			}

			msg.Content = roomID[:]
		case "/leave":
			msg.Header.Type = ptc.Leave
		case "/quit":
			fmt.Println("Quitting...")
			return
		default:
			msg.Header.Type = ptc.Chat
			msg.Content = []byte(text)
		}

		msg.Header.Length = uint32(len(msg.Content))

		encodedMsg, err := ptc.EncodeMsg(msg)
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

func (ct *ClientTUI) rememberRooms(rooms []ptc.Room) {
	ct.rooms = rooms
}
