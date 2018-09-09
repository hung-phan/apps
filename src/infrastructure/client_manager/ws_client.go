package client_manager

import (
	"github.com/gorilla/websocket"
	"github.com/hung-phan/chat-app/src/infrastructure/logger"
	"go.uber.org/zap"
)

type WSClient struct {
	*Client
	*ChannelCommunication

	Conn *websocket.Conn
}

// websocket won't try to do buffering like TCP, so you never need to call Flush() on it
func (wsClient *WSClient) Flush() error {
	return nil
}

func (wsClient *WSClient) Shutdown() {
	wsClient.once.Do(wsClient.scheduleForShutdown)
}

func (wsClient *WSClient) scheduleForShutdown() {
	wsClient.sendCloseSignal()
	wsClient.CloseAllChannels()

	wsClient.Conn.Close()
	wsClient.Hub.Del(wsClient.ID)
}

func (wsClient *WSClient) sendCloseSignal() error {
	return wsClient.Conn.WriteMessage(websocket.CloseMessage, []byte{})
}

func (wsClient *WSClient) readPump() {
	defer wsClient.Shutdown()

	for {
		_, data, err := wsClient.Conn.ReadMessage()

		if err != nil {
			if websocket.IsUnexpectedCloseError(
				err,
				websocket.CloseGoingAway,
				websocket.CloseNoStatusReceived,
				websocket.CloseAbnormalClosure,
			) {
				logger.Client.Debug("ws read fail", zap.Error(err), zap.String("ID", wsClient.ID))
			}
			return
		}

		wsClient.receiveCh <- data
	}
}

func (wsClient *WSClient) writePump() {
	defer wsClient.Shutdown()

	for data := range wsClient.sendCh {
		err := wsClient.Conn.WriteMessage(websocket.TextMessage, data)

		if err != nil {
			logger.Client.Debug("ws write fail", zap.Error(err), zap.String("ID", wsClient.ID))
			return
		}
	}
}

func NewWSClient(hub *Hub, id string, conn *websocket.Conn) *WSClient {
	client := &WSClient{
		Client: &Client{
			Hub: hub,
			ID:  id,
		},
		ChannelCommunication: &ChannelCommunication{
			sendCh:    make(chan []byte, 256),
			receiveCh: make(chan []byte, 256),
		},
		Conn: conn,
	}

	hub.Set(id, client)

	go client.readPump()
	go client.writePump()

	return client
}
