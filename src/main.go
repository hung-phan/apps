package main

import (
	"github.com/hung-phan/chat-app/src/application"
	"github.com/hung-phan/chat-app/src/infrastructure"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	var (
		stopSignal     = make(chan os.Signal)
		tcpStopSignal  = make(chan bool)
		httpStopSignal = make(chan bool)
	)

	signal.Notify(stopSignal, syscall.SIGTERM)
	signal.Notify(stopSignal, syscall.SIGINT)

	go infrastructure.StartHTTPServer(":3000", application.CreateRouter())
	go infrastructure.StartTCPServer(":3001", tcpStopSignal, application.HandleConnection)

	<-stopSignal

	tcpStopSignal <- true
	httpStopSignal <- true

	os.Exit(0)
}
