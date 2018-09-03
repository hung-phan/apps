package application

import "github.com/hung-phan/chat-app/src/infrastructure/client_manager"

func TCPConnectionHandler(tcpClient client_manager.IClient) {
	receiveCh, sendCh := tcpClient.GetReceiveChannel(), tcpClient.GetSendChannel()

	for data := range receiveCh {
		sendCh <- data
	}
}

func WSConnectionHandler(wsClient client_manager.IClient) {
	receiveCh, sendCh := wsClient.GetReceiveChannel(), wsClient.GetSendChannel()

	for data := range receiveCh {
		sendCh <- data
	}
}
