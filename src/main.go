package main

import (
	"github.com/hung-phan/chat-app/src/application"
	"github.com/hung-phan/chat-app/src/infrastructure"
	"github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	go infrastructure.StartHTTPServer(":3000", application.CreateRouter())
	go infrastructure.StartTCPServer(":3001", application.HandleIncomingMessage)

	gracefulStop := make(chan os.Signal)

	signal.Notify(gracefulStop, syscall.SIGTERM)
	signal.Notify(gracefulStop, syscall.SIGINT)

	sig := <-gracefulStop

	logrus.Printf("caught sig: %+v", sig)

	os.Exit(0)
}
