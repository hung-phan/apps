package application

import (
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/hung-phan/chat-app/src/infrastructure"
	"github.com/segmentio/ksuid"
	"go.uber.org/zap"
	"net/http"
)

var webSocketUpgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func CreateRouter(wsConnectionHandler func(infrastructure.Client)) *mux.Router {
	var router = mux.NewRouter()

	router.HandleFunc(
		"/ws",
		func(w http.ResponseWriter, r *http.Request) {
			conn, err := webSocketUpgrader.Upgrade(w, r, nil)

			if err != nil {
				infrastructure.Log.Debug("websocket upgrade fail", zap.Error(err))
				return
			}

			go wsConnectionHandler(infrastructure.NewWSClient(
				DefaultWSHub,
				ksuid.New().String(),
				conn,
			))
		},
	)

	return router
}
