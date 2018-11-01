package infrastructure

import (
	"github.com/segmentio/ksuid"
	"go.uber.org/zap"
	"net"
	"sync"
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
