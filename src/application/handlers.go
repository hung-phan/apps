package application

import (
	"github.com/hung-phan/chat-app/src/infrastructure/client_manager"
)

var (
	DefaultTCPHub = client_manager.NewHub()
	DefaultWSHub  = client_manager.NewHub()
)

func TCPConnectionHandler(tcpClient client_manager.Client) {
	receiveCh, sendCh := tcpClient.GetReceiveChannel(), tcpClient.GetSendChannel()

	for data := range receiveCh {
		sendCh <- data
	}
}

func WSConnectionHandler(wsClient client_manager.Client) {
	receiveCh, sendCh := wsClient.GetReceiveChannel(), wsClient.GetSendChannel()

	for data := range receiveCh {
		sendCh <- data
	}
}
