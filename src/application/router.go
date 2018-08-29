package application

import (
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/hung-phan/chat-app/src/infrastructure/client_manager"
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

func CreateRouter(wsConnectionHandler func(string, client_manager.IClient)) *mux.Router {
	var (
		hub    = client_manager.NewHub()
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

			go wsConnectionHandler(client_manager.WebSocketConnection, client_manager.NewWSClient(
				hub,
				ksuid.New().String(),
				conn,
			))
		},
	)

	return router
}
