package infrastructure

import (
	"context"
	"github.com/gorilla/mux"
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
	hub *ClientHub,
	ch ClientHandler,
) {
	listener, err := net.Listen("tcp", address)

	if err != nil {
		Log.Fatal("failed to start TCP server:", zap.Error(err))
	}

	Log.Info("start TCP server", zap.String("address", address))

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
				Log.Debug("failed to accept new connection request:", zap.Error(err))
				continue
			}

			connCh <- conn.(*net.TCPConn)
		}
	}()

	for {
		select {
		case conn := <-connCh:
			go ch(NewTCPClient(
				hub,
				ksuid.New().String(),
				conn,
			))

		case <-shutdownSignal:
			hub.ExecuteAll(func(client Client) {
				client.GracefulShutdown()
			})
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
		Log.Info("start HTTP server", zap.String("address", address))

		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			Log.Fatal("failed to start HTTP server:", zap.Error(err))
		}
	}()

	<-shutdownSignal

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)

	defer cancel()

	server.Shutdown(ctx)
}
