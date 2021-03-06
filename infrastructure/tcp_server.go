package infrastructure

import (
	"github.com/segmentio/ksuid"
	"go.uber.org/atomic"
	"go.uber.org/zap"
	"net"
	"sync"
)

func StartTCPServer(
	address string,
	stopSignal chan bool,
	wg *sync.WaitGroup,
	ch ClientHandler,
) {
	wg.Add(1)

	listener, err := net.Listen("tcp", address)

	if err != nil {
		Log.Fatal("failed to start TCP server:", zap.Error(err))
	}

	Log.Info("start TCP server", zap.String("address", address))

	var (
		connCh     = make(chan *net.TCPConn)
		isShutdown = atomic.NewBool(false)
	)

	go func() {
		for {
			if isShutdown.Load() {
				break
			}

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
				ksuid.New().String(),
				conn,
			))

		case <-stopSignal:
			DefaultTCPHub.ExecuteAll(func(client Client) {
				client.Shutdown()
			})

			isShutdown.Swap(true)

			if err := listener.Close(); err != nil {
				Log.Fatal("failed to TCP server:", zap.Error(err))
			}

			Log.Info("Shutdown TCP server")

			wg.Done()

			return
		}
	}
}
