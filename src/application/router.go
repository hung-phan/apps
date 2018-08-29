package application

import (
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/hung-phan/chat-app/src/infrastructure"
	"github.com/segmentio/ksuid"
	"github.com/sirupsen/logrus"
	"net/http"
)

var (
	webSocketUpgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
)

func HandleConnection(_ string, client infrastructure.IClient) {
	receiveCh, sendCh := client.GetReceiveChannel(), client.GetSendChannel()

	for data := range receiveCh {
		sendCh <- data
	}
}

func CreateRouter() *mux.Router {
	var (
		hub    = infrastructure.NewHub()
		router = mux.NewRouter()
	)

	router.HandleFunc(
		"/ws",
		func(w http.ResponseWriter, r *http.Request) {
			conn, err := webSocketUpgrader.Upgrade(w, r, nil)

			if err != nil {
				logrus.Debug("upgrade:", err)
				return
			}

			go HandleConnection(infrastructure.WebSocketConnection, infrastructure.NewWSClient(
				hub,
				ksuid.New().String(),
				conn,
			))
		},
	)

	return router
}
