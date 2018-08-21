package connection_manager

import (
	"github.com/gorilla/websocket"
	"github.com/segmentio/ksuid"
	"log"
	"sync"
)

type WSClient struct {
	Conn *websocket.Conn

	once      sync.Once
	hub       *Hub
	id        string
	sendCh    chan []byte
	receiveCh chan []byte
}

func (client *WSClient) GetChannels() (chan []byte, chan []byte) {
	return client.sendCh, client.receiveCh
}

func (client *WSClient) CleanUp() {
	close(client.receiveCh)
	close(client.sendCh)

	client.hub.Del(client.id)
	client.Conn.Close()
}

func (client *WSClient) readPump() {
	defer client.once.Do(client.CleanUp)

	for {
		_, data, err := client.Conn.ReadMessage()

		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Fatalln("error:", err)
			}
			break
		}

		client.receiveCh <- data
	}
}

func (client *WSClient) writePump() {
	defer client.once.Do(client.CleanUp)

	for data := range client.sendCh {
		err := client.Conn.WriteMessage(websocket.TextMessage, data)

		if err != nil {
			log.Fatalln("error:", err)
		}
	}

	// write close signal
	client.Conn.WriteMessage(websocket.CloseMessage, []byte{})
}

func NewWSClient(hub *Hub, id ksuid.KSUID, conn *websocket.Conn) *WSClient {
	client := &WSClient{
		Conn:      conn,
		hub:       hub,
		id:        id.String(),
		sendCh:    make(chan []byte, 256),
		receiveCh: make(chan []byte, 256),
	}

	hub.Set(id.String(), client)

	go client.readPump()
	go client.writePump()

	return client
}
