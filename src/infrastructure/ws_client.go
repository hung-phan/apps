package infrastructure

import (
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
	"sync"
)

type WSClient struct {
	*baseClient

	conn         *websocket.Conn
	startOnce    sync.Once
	shutdownOnce sync.Once
	m            sync.Mutex
}

func (wsClient *WSClient) Start() {
	wsClient.m.Lock()
	defer wsClient.m.Unlock()

	wsClient.isStarted = true

	go wsClient.startOnce.Do(wsClient.readPump)
}

func (wsClient *WSClient) Shutdown() {
	wsClient.m.Lock()
	defer wsClient.m.Unlock()

	wsClient.isShutdown = true

	go wsClient.shutdownOnce.Do(wsClient.gracefulShutdown)
}

func (wsClient *WSClient) Write(data []byte) (int, error) {
	wsClient.m.Lock()
	defer wsClient.m.Unlock()

	if err := wsClient.conn.WriteMessage(websocket.TextMessage, data); err != nil {
		return 0, err
	}

	return len(data), nil
}

// websocket won't try to do buffering like TCP, so WebSocket Flush is a no op
func (wsClient *WSClient) Flush() error {
	return nil
}

func (wsClient *WSClient) readPump() {
	defer wsClient.Shutdown()

	for {
		_, data, err := wsClient.conn.ReadMessage()

		if err != nil {
			if websocket.IsUnexpectedCloseError(
				err,
				websocket.CloseGoingAway,
				websocket.CloseNoStatusReceived,
				websocket.CloseAbnormalClosure,
			) {
				Log.Debug("ws read fail", zap.Error(err), zap.String("id", wsClient.GetID()))
			}
			return
		}

		go wsClient.broadcastToChannel(data)
	}
}

func (wsClient *WSClient) gracefulShutdown() {
	err := wsClient.sendCloseSignal()

	if err != nil {
		Log.Error("fail to send close signal", zap.Error(err))
	}

	err = wsClient.conn.Close()

	if err != nil {
		Log.Error("fail to close WebSocket connection", zap.Error(err))
	}

	wsClient.GetHub().Del(wsClient.GetID())
}

func (wsClient *WSClient) sendCloseSignal() error {
	wsClient.m.Lock()
	defer wsClient.m.Unlock()

	return wsClient.conn.WriteMessage(websocket.CloseMessage, []byte{})
}

func NewWSClient(hub *ClientHub, id string, conn *websocket.Conn) *WSClient {
	client := &WSClient{
		baseClient: &baseClient{
			id:  id,
			hub: hub,
		},
		conn: conn,
	}

	hub.Set(id, client)

	return client
}
