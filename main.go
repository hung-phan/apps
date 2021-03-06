package main

import (
	"github.com/hung-phan/apps/application"
	"github.com/hung-phan/apps/infrastructure"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

func main() {
	var (
		wg             = &sync.WaitGroup{}
		stopSignal     = make(chan os.Signal)
		tcpStopSignal  = make(chan bool)
		httpStopSignal = make(chan bool)
	)

	go infrastructure.StartHTTPServer(
		":3000",
		httpStopSignal,
		wg,
		application.CreateRouter(),
	)

	go infrastructure.StartTCPServer(
		":3001",
		tcpStopSignal,
		wg,
		application.ConnectionHandler,
	)

	signal.Notify(stopSignal, syscall.SIGTERM)
	signal.Notify(stopSignal, syscall.SIGINT)

	<-stopSignal

	tcpStopSignal <- true
	httpStopSignal <- true

	wg.Wait()
}
