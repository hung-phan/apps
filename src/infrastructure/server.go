package infrastructure

import (
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/segmentio/ksuid"
	"log"
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
	TCPHub       = NewHub()
	WebSocketHub = NewHub()
)

func CreateWebSocketHandler(handler func(string, IClient)) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)

		if err != nil {
			log.Println("upgrade:", err)
			return
		}

		client := NewWSClient(
			WebSocketHub,
			ksuid.New().String(),
			conn,
		)

		go handler(WebSocketConnection, client)
	}
}

func StartTCPServer(address string, handler func(string, IClient)) {
	listener, err := net.Listen("tcp", address)

	if err != nil {
		log.Fatalln("unable to establish TCP server on", address)
	}

	log.Println(">> start TCP server at", address)
	for {
		conn, err := listener.Accept()

		if err != nil {
			log.Println("failed accepting a connection request:", err)
			continue
		}

		client := NewTCPClient(
			TCPHub,
			ksuid.New().String(),
			conn.(*net.TCPConn),
		)

		go handler(TcpConnection, client)
	}
}

func StartHTTPServer(address string, router *mux.Router) error {
	log.Println(">> start HTTP server at", address)

	return http.ListenAndServe(address, router)
}
