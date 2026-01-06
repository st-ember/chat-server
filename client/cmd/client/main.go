package main

import (
	"github.com/st-ember/chat-server/client/internal/tui"
)

func main() {
	client := tui.NewClientTUI()
	client.Start()
}
