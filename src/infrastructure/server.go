package infrastructure

import (
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/segmentio/ksuid"
	"github.com/sirupsen/logrus"
	"net"
	"net/http"
)

const (
	TcpConnection       = "TCP_CONNECTION"
	WebSocketConnection = "WEBSOCKET_CONNECTION"
)

var (
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
)

func CreateWebSocketHandler(handler func(string, IClient)) func(http.ResponseWriter, *http.Request) {
	hub := NewHub()

	return func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)

		if err != nil {
			logrus.Debug("upgrade:", err)
			return
		}

		client := NewWSClient(
			hub,
			ksuid.New().String(),
			conn,
		)

		go handler(WebSocketConnection, client)
	}
}

func StartTCPServer(address string, shutdownSignal chan bool, connectionHandler func(string, IClient)) {
	listener, err := net.Listen("tcp", address)

	if err != nil {
		logrus.Fatalln("failed to start server:", err)
	}

	logrus.WithField("address", address).Info("start TCP server")

	var (
		hub            = NewHub()
		isShuttingDown = false
	)

	go func() {
		<-shutdownSignal

		isShuttingDown = true

		hub.Shutdown()
		listener.Close()
	}()

	for {
		if isShuttingDown {
			break
		}

		conn, err := listener.Accept()

		if err != nil {
			logrus.Debug("failed to accept new connection request:", err)
			continue
		}

		go connectionHandler(TcpConnection, NewTCPClient(
			hub,
			ksuid.New().String(),
			conn.(*net.TCPConn),
		))
	}
}

func StartHTTPServer(address string, router *mux.Router) error {
	logrus.Println("start HTTP server at", address)

	return http.ListenAndServe(address, router)
}
