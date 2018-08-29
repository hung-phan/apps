package application

import "github.com/hung-phan/chat-app/src/infrastructure/client_manager"

func TCPConnectionHandler(_ string, client client_manager.IClient) {
	receiveCh, sendCh := client.GetReceiveChannel(), client.GetSendChannel()

	for data := range receiveCh {
		sendCh <- data
	}
}

func WSConnectionHandler(_ string, client client_manager.IClient) {
	receiveCh, sendCh := client.GetReceiveChannel(), client.GetSendChannel()

	for data := range receiveCh {
		sendCh <- data
	}
}
