package application

import (
	"github.com/hung-phan/chat-app/src/infrastructure"
)

var (
	DefaultTCPHub = infrastructure.NewHub()
	DefaultWSHub  = infrastructure.NewHub()
)

func TCPConnectionHandler(tcpClient infrastructure.Client) {
	ch := make(infrastructure.DataChannel)

	go func() {
		for data := range ch {
			tcpClient.Write(data)
		}
	}()

	tcpClient.AddListener(ch)
}

func WSConnectionHandler(wsClient infrastructure.Client) {
	ch := make(infrastructure.DataChannel)

	go func() {
		for data := range ch {
			wsClient.Write(data)
		}
	}()

	wsClient.AddListener(ch)
}
