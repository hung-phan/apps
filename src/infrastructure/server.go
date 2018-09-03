package infrastructure

import (
	"context"
	"github.com/gorilla/mux"
	"github.com/hung-phan/chat-app/src/infrastructure/client_manager"
	"github.com/segmentio/ksuid"
	"github.com/sirupsen/logrus"
	"net"
	"net/http"
	"sync"
	"time"
)

func StartTCPServer(address string, shutdownSignal chan bool, connectionHandler func(string, client_manager.IClient)) {
	listener, err := net.Listen("tcp", address)

	if err != nil {
		logrus.Fatalln("failed to start server:", err)
	}

	logrus.WithField("address", address).Info("start TCP server")

	var (
		tcpHub     = client_manager.NewHub()
		connCh     = make(chan *net.TCPConn)
		rwMutex    = sync.Mutex{}
		isShutdown = false
	)

	go func() {
		for {
			rwMutex.Lock()
			if isShutdown {
				break
			}
			rwMutex.Unlock()

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
			go connectionHandler(client_manager.TcpConnection, client_manager.NewTCPClient(
				tcpHub,
				ksuid.New().String(),
				conn,
			))

		case <-shutdownSignal:
			tcpHub.Shutdown()
			listener.Close()

			rwMutex.Lock()
			isShutdown = true
			rwMutex.Unlock()
		}
	}
}

func StartHTTPServer(address string, shutdownSignal chan bool, router *mux.Router) {
	server := &http.Server{Addr: address, Handler: router}

	go func() {
		logrus.Println("start HTTP server at", address)

		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			logrus.Fatalln(err)
		}
	}()

	<-shutdownSignal

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)

	defer cancel()

	server.Shutdown(ctx)
}
