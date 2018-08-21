package application

import (
	"github.com/gorilla/websocket"
	"github.com/hung-phan/chat-app/src/domain/connection_manager"
	"github.com/segmentio/ksuid"
	"log"
	"net/http"
)

var (
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
)

func handleRequest(client *connection_manager.WSClient) {
	for data := range client.GetReceiveChannel() {
		log.Println("receive data", data)
	}
}

func WSHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)

	if err != nil {
		log.Fatalln("upgrade:", err)
		return
	}

	client := connection_manager.NewWSClient(
		connection_manager.DefaultWSHub,
		ksuid.New(),
		conn,
	)

	go handleRequest(client)
}
