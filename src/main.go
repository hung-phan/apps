package main

import (
	"github.com/hung-phan/chat-app/src/application"
	"github.com/hung-phan/chat-app/src/infrastructure"
)

func main() {
	go infrastructure.StartTCPServer(":3001", application.HandleIncomingMessage)
	infrastructure.StartHTTPServer(":3000", application.CreateRouter())
}
