package application

import (
	"github.com/hung-phan/chat-app/src/infrastructure"
)

var (
	DefaultTCPHub = infrastructure.NewHub()
	DefaultWSHub  = infrastructure.NewHub()
)

func TCPConnectionHandler(tcpClient infrastructure.Client) {
	tcpClient.AddListener(func(data []byte) {
		tcpClient.Write(data)
	})
}

func WSConnectionHandler(wsClient infrastructure.Client) {
	wsClient.AddListener(func(data []byte) {
		wsClient.Write(data)
	})
}
