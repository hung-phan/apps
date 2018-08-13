package application

import (
	"github.com/gorilla/websocket"
	"github.com/hung-phan/chat-app/src/domain/connection_manager/ws"
	"github.com/segmentio/ksuid"
	"log"
	"net/http"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func handleRequest(client *ws.Client) {
	for data := range client.Receive {
		log.Println("receive data", data)
	}
}

func WSHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)

	if err != nil {
		log.Println("upgrade:", err)
		return
	}

	client := ws.NewWSClient(ws.DefaultHub, ksuid.New(), conn)

	go handleRequest(client)
}
