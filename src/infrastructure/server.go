package infrastructure

import (
	"context"
	"github.com/gorilla/mux"
	"github.com/hung-phan/chat-app/src/infrastructure/client_manager"
	"github.com/hung-phan/chat-app/src/infrastructure/logger"
	"github.com/segmentio/ksuid"
	"go.uber.org/zap"
	"net"
	"net/http"
	"sync"
	"time"
)

func StartTCPServer(
	address string,
	shutdownSignal chan bool,
	hub *client_manager.Hub,
	connectionHandler func(client_manager.Client),
) {
	listener, err := net.Listen("tcp", address)

	if err != nil {
		logger.Client.Fatal("failed to start TCP server:", zap.Error(err))
	}

	logger.Client.Info("start TCP server", zap.String("address", address))

	var (
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
				logger.Client.Debug("failed to accept new connection request:", zap.Error(err))
				continue
			}

			connCh <- conn.(*net.TCPConn)
		}
	}()

	for {
		select {
		case conn := <-connCh:
			go connectionHandler(client_manager.NewTCPClient(
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

func StartHTTPServer(
	address string,
	shutdownSignal chan bool,
	router *mux.Router,
) {
	server := &http.Server{Addr: address, Handler: router}

	go func() {
		logger.Client.Info("start HTTP server", zap.String("address", address))

		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			logger.Client.Fatal("failed to start HTTP server:", zap.Error(err))
		}
	}()

	<-shutdownSignal

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)

	defer cancel()

	server.Shutdown(ctx)
}
