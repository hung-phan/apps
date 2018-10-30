package application

import (
	"github.com/hung-phan/chat-app/src/infrastructure"
	"time"
)

var (
	DefaultTCPHub = infrastructure.NewHub()
	DefaultWSHub  = infrastructure.NewHub()
)

func TCPConnectionHandler(tcpClient infrastructure.Client) {
	ch := make(infrastructure.DataChannel)

	go func() {
		defer close(ch)

		for {
			if tcpClient.IsShutdown() {
				break
			}

			select {
			case data := <- ch:
				tcpClient.Write(data)
			case <- time.After(5 * time.Second):
				// do nothing
			}
		}
	}()

	tcpClient.AddListener(ch)
}

func WSConnectionHandler(wsClient infrastructure.Client) {
	ch := make(infrastructure.DataChannel)

	go func() {
		defer close(ch)

		for {
			if wsClient.IsShutdown() {
				break
			}

			select {
			case data := <- ch:
				wsClient.Write(data)
			case <- time.After(5 * time.Second):
				// do nothing
			}
		}
	}()

	wsClient.AddListener(ch)
}
