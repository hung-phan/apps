package infrastructure

import (
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/segmentio/ksuid"
	"github.com/sirupsen/logrus"
	"net"
	"net/http"
	"sync"
)

const (
	TcpConnection       = "TCP_CONNECTION"
	WebSocketConnection = "WEB_SOCKET_CONNECTION"
)

var (
	webSocketUpgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
)

type ConnectionHandler func(string, IClient)

func CreateWebSocketHandler(connectionHandler ConnectionHandler) http.HandlerFunc {
	hub := NewHub()

	return func(w http.ResponseWriter, r *http.Request) {
		conn, err := webSocketUpgrader.Upgrade(w, r, nil)

		if err != nil {
			logrus.Debug("upgrade:", err)
			return
		}

		go connectionHandler(WebSocketConnection, NewWSClient(
			hub,
			ksuid.New().String(),
			conn,
		))
	}
}

func StartTCPServer(address string, shutdownSignal chan bool, connectionHandler ConnectionHandler) {
	listener, err := net.Listen("tcp", address)

	if err != nil {
		logrus.Fatalln("failed to start server:", err)
	}

	logrus.WithField("address", address).Info("start TCP server")

	var (
		hub        = NewHub()
		connCh     = make(chan *net.TCPConn)
		rwMutex    = sync.RWMutex{}
		isShutdown = false
	)

	go func() {
		for {
			rwMutex.RLock()
			if isShutdown {
				break
			}
			rwMutex.RUnlock()

			conn, err := listener.Accept()

			if err != nil {
				logrus.Debug("failed to accept new connection request:", err)
				continue
			}

			connCh <- conn.(*net.TCPConn)
		}
	}()

	for {
		select {
		case conn := <-connCh:
			go connectionHandler(TcpConnection, NewTCPClient(
				hub,
				ksuid.New().String(),
				conn,
			))

		case <-shutdownSignal:
			hub.Shutdown()
			listener.Close()

			rwMutex.Lock()
			isShutdown = true
			rwMutex.Unlock()
		}
	}
}

func StartHTTPServer(address string, router *mux.Router) error {
	logrus.Println("start HTTP server at", address)

	return http.ListenAndServe(address, router)
}
