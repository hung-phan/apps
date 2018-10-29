package infrastructure

import (
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
	"sync"
)

type WSClient struct {
	*baseClient

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

// websocket won't try to do buffering like TCP, so WebSocket Flush is a no op
func (wsClient *WSClient) Flush() error {
	return nil
}

func (wsClient *WSClient) GracefulShutdown() {
	wsClient.once.Do(wsClient.shutdown)
}

func (wsClient *WSClient) shutdown() {
	var err error = nil

	err = wsClient.sendCloseSignal()

	if err != nil {
		Log.Error("fail to send close signal", zap.Error(err))
	}

	err = wsClient.Conn.Close()

	if err != nil {
		Log.Error("fail to close WebSocket connection", zap.Error(err))
	}

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
				Log.Debug("ws read fail", zap.Error(err), zap.String("ID", wsClient.ID))
			}
			return
		}

		go wsClient.Emit(data)
	}
}

func NewWSClient(hub *ClientHub, id string, conn *websocket.Conn) *WSClient {
	client := &WSClient{
		baseClient: &baseClient{
			ID:  id,
			Hub: hub,
		},
		Conn: conn,
	}

	hub.Set(id, client)

	go client.readPump()

	return client
}
