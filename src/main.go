package main

import (
	"os"
	"os/signal"
	"syscall"
	"fmt"
	"github.com/hung-phan/chat-app/src/infrastructure"
	"github.com/hung-phan/chat-app/src/application"
)

func main() {
	go infrastructure.StartHTTPServer(":3000", application.CreateRouter())
	go infrastructure.StartTCPServer(":3001", application.HandleIncomingMessage)

	gracefulStop := make(chan os.Signal)

	signal.Notify(gracefulStop, syscall.SIGTERM)
	signal.Notify(gracefulStop, syscall.SIGINT)

	sig := <-gracefulStop

	fmt.Printf("caught sig: %+v", sig)

	os.Exit(0)
}
