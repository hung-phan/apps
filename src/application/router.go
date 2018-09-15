package application

import (
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/hung-phan/chat-app/src/infrastructure/client_manager"
	"github.com/hung-phan/chat-app/src/infrastructure/logger"
	"github.com/segmentio/ksuid"
	"go.uber.org/zap"
	"net/http"
)

var (
	webSocketUpgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
)

func CreateRouter(wsConnectionHandler func(client_manager.Client)) *mux.Router {
	var (
		router = mux.NewRouter()
	)

	router.HandleFunc(
		"/ws",
		func(w http.ResponseWriter, r *http.Request) {
			conn, err := webSocketUpgrader.Upgrade(w, r, nil)

			if err != nil {
				logger.Client.Debug("websocket upgrade fail", zap.Error(err))
				return
			}

			go wsConnectionHandler(client_manager.NewWSClient(
				DefaultWSHub,
				ksuid.New().String(),
				conn,
			))
		},
	)

	return router
}
