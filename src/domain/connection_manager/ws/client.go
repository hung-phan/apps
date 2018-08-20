package ws

import (
	"github.com/gorilla/websocket"
	"github.com/segmentio/ksuid"
	"log"
	"sync"
)

type Client struct {
	once    sync.Once
	hub     *Hub
	id      ksuid.KSUID
	Conn    *websocket.Conn
	Send    chan []byte
	Receive chan []byte
}

func (client *Client) cleanUp() {
	client.Conn.Close()
	client.hub.Del(client.id)

	close(client.Send)
	close(client.Receive)
}

func (client *Client) readPump() {
	defer client.once.Do(client.cleanUp)

	for {
		_, data, err := client.Conn.ReadMessage()

		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Println("error:", err)
			}
			break
		}

		client.Receive <- data
	}
}

func (client *Client) writePump() {
	defer client.once.Do(client.cleanUp)

	for data := range client.Send {
		err := client.Conn.WriteMessage(websocket.TextMessage, data)

		if err != nil {
			log.Fatalf("error: %v\n", err)
		}
	}

	// write close signal
	client.Conn.WriteMessage(websocket.CloseMessage, []byte{})
}

func NewWSClient(hub *Hub, id ksuid.KSUID, conn *websocket.Conn) *Client {
	client := &Client{
		hub:     hub,
		id:      id,
		Conn:    conn,
		Send:    make(chan []byte, 256),
		Receive: make(chan []byte, 256),
	}

	hub.Set(id, client)

	go client.readPump()
	go client.writePump()

	return client
}
