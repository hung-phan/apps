package application

import (
	"github.com/gorilla/mux"
	"github.com/hung-phan/chat-app/src/infrastructure"
	"log"
)

func HandleIncomingMessage(connectionType string, client infrastructure.IClient) {
	channel := client.GetReceiveChannel()

	log.Println("connection type:", connectionType)

	for data := range channel {
		log.Println("receive data:", data)
	}
}

func CreateRouter() *mux.Router {
	router := mux.NewRouter()

	router.HandleFunc("/ws", infrastructure.CreateWebSocketHandler(HandleIncomingMessage))

	return router
}
