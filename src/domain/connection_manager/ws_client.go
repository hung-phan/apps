package connection_manager

import (
	"github.com/gorilla/websocket"
	"github.com/segmentio/ksuid"
	"log"
)

type WSClient struct {
	Client
	ChannelCommunication

	Conn *websocket.Conn
}

func (wsClient *WSClient) cleanUp() {
	wsClient.sendCloseSignal()
	wsClient.CloseAllChannels()
	wsClient.hub.Del(wsClient.id)
	wsClient.Conn.Close()
}

func (wsClient *WSClient) sendCloseSignal() error {
	return wsClient.Conn.WriteMessage(websocket.CloseMessage, []byte{})
}

func (wsClient *WSClient) readPump() {
	defer wsClient.once.Do(wsClient.cleanUp)

	for {
		_, data, err := wsClient.Conn.ReadMessage()

		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Fatalln("error:", err)
			}
			return
		}

		wsClient.receiveCh <- data
	}
}

func (wsClient *WSClient) writePump() {
	defer wsClient.once.Do(wsClient.cleanUp)

	for data := range wsClient.sendCh {
		err := wsClient.Conn.WriteMessage(websocket.TextMessage, data)

		if err != nil {
			log.Fatalln("error:", err)
			return
		}
	}
}

func NewWSClient(hub *Hub, id ksuid.KSUID, conn *websocket.Conn) *WSClient {
	client := &WSClient{
		Client: Client{
			hub: hub,
			id:  id.String(),
		},
		ChannelCommunication: ChannelCommunication{
			sendCh:    make(chan []byte, 256),
			receiveCh: make(chan []byte, 256),
		},
		Conn: conn,
	}

	hub.Set(id.String(), client)

	go client.readPump()
	go client.writePump()

	return client
}
