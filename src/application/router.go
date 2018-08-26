package application

import (
	"github.com/gorilla/mux"
	"github.com/hung-phan/chat-app/src/infrastructure"
)

func CreateRouter() *mux.Router {
	router := mux.NewRouter()

	router.HandleFunc(
		"/ws",
		infrastructure.CreateWebSocketHandler(HandleConnection),
	)

	return router
}

func HandleConnection(_ string, client infrastructure.IClient) {
	receiveCh, sendCh := client.GetReceiveChannel(), client.GetSendChannel()

	for data := range receiveCh {
		sendCh <- data
	}
}
