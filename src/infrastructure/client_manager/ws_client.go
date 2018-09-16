package client_manager

import (
	"github.com/gorilla/websocket"
	"github.com/hung-phan/chat-app/src/infrastructure/logger"
	"go.uber.org/zap"
	"sync"
)

type WSClient struct {
	*BaseClient

	Conn    *websocket.Conn
	once    sync.Once
	ioMutex sync.Mutex
}

func (wsClient *WSClient) Write(data []byte) (int, error) {
	wsClient.ioMutex.Lock()
	defer wsClient.ioMutex.Unlock()

	err := wsClient.Conn.WriteMessage(websocket.TextMessage, data)

	if err != nil {
		return 0, err
	} else {
		return len(data), nil
	}
}

// websocket won't try to do buffering like TCP, so you never need to call Flush() on it
func (wsClient *WSClient) Flush() error {
	return nil
}

func (wsClient *WSClient) GracefulShutdown() {
	wsClient.once.Do(wsClient.shutdown)
}

func (wsClient *WSClient) shutdown() {
	wsClient.sendCloseSignal()
	wsClient.Conn.Close()
	wsClient.Hub.Del(wsClient.ID)
}

func (wsClient *WSClient) sendCloseSignal() error {
	wsClient.ioMutex.Lock()
	defer wsClient.ioMutex.Unlock()

	return wsClient.Conn.WriteMessage(websocket.CloseMessage, []byte{})
}

func (wsClient *WSClient) readPump() {
	defer wsClient.GracefulShutdown()

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

		go wsClient.Broadcast(data)
	}
}

func NewWSClient(hub *Hub, id string, conn *websocket.Conn) *WSClient {
	client := &WSClient{
		BaseClient: &BaseClient{
			ID:  id,
			Hub: hub,
		},
		Conn: conn,
	}

	hub.Set(id, client)

	go client.readPump()

	return client
}
