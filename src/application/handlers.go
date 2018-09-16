package application

import (
	"github.com/hung-phan/chat-app/src/infrastructure/client_manager"
)

var (
	DefaultTCPHub = client_manager.NewHub()
	DefaultWSHub  = client_manager.NewHub()
)

func TCPConnectionHandler(tcpClient client_manager.Client) {
	tcpClient.AddListener(func(data []byte) {
		tcpClient.Write(data)
	})
}

func WSConnectionHandler(wsClient client_manager.Client) {
	wsClient.AddListener(func(data []byte) {
		wsClient.Write(data)
	})
}
