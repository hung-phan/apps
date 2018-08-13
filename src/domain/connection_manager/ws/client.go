package ws

import (
	"github.com/gorilla/websocket"
	"github.com/segmentio/ksuid"
	"log"
)

type Client struct {
	hub     *Hub
	id      ksuid.KSUID
	Conn    *websocket.Conn
	Send    chan []byte
	Receive chan []byte
}

func (client *Client) readPump() {
	defer func() {
		close(client.Receive)
		client.Conn.Close()
		client.hub.Del(client.id)
	}()

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
	defer func() {
		close(client.Send)
		client.Conn.Close()
		client.hub.Del(client.id)
	}()

	for data := range client.Send {
		err := client.Conn.WriteMessage(websocket.TextMessage, data)

		if err != nil {
			log.Printf("error: %v", err)
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
