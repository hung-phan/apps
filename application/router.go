package application

import (
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/hung-phan/apps/infrastructure"
	"github.com/segmentio/ksuid"
	"go.uber.org/zap"
	"net/http"
)

var webSocketUpgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func CreateRouter() *mux.Router {
	var router = mux.NewRouter()

	router.HandleFunc(
		"/ws",
		func(w http.ResponseWriter, r *http.Request) {
			conn, err := webSocketUpgrader.Upgrade(w, r, nil)

			if err != nil {
				infrastructure.Log.Debug("websocket upgrade fail", zap.Error(err))
				return
			}

			go WSConnectionHandler(infrastructure.NewWSClient(
				ksuid.New().String(),
				conn,
			))
		},
	)

	return router
}
