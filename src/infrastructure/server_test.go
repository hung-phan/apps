package infrastructure

import (
	"bufio"
	"github.com/hung-phan/chat-app/src/application"
	"github.com/hung-phan/chat-app/src/infrastructure/client_manager"
	"github.com/stretchr/testify/assert"
	"math/rand"
	"testing"
	"time"
)

func TestInfrastructure(t *testing.T) {
	var (
		assertInstance = assert.New(t)
		jitter         = func() {
			time.Sleep(time.Duration(rand.Intn(200)) * time.Millisecond)
		}
	)

	t.Run("test StartTCPServer", func(t *testing.T) {
		var (
			tcpStopSignal        = make(chan bool)
			msg                  = "Message"
			tcpConnectionHandler = func(connectionType string, client client_manager.IClient) {
				sameClient := client.GetHub().Get(client.GetID())

				assertInstance.Equal(connectionType, client_manager.TcpConnection)
				assertInstance.Equal(client, sameClient)

				for data := range client.GetReceiveChannel() {
					client.GetHub().Broadcast(data)

					// enough time for client to send the message so we can force
					// the connection to flush it later
					jitter()

					client.Flush()
				}
			}
			assertError = func(err error) {
				if err != nil {
					assertInstance.Fail(err.Error())
				}
			}
		)

		go StartTCPServer(
			"localhost:3001",
			tcpStopSignal,
			tcpConnectionHandler,
		)

		// wait for server to start
		jitter()

		tcpConn, err := client_manager.CreateTCPConnection("localhost:3001")
		assertError(err)

		rw := bufio.NewReadWriter(bufio.NewReader(tcpConn), bufio.NewWriter(tcpConn))

		_, err = rw.Write([]byte(msg + "\n"))
		assertError(err)

		err = rw.Flush()
		assertError(err)

		data, err := rw.ReadBytes('\n')
		assertError(err)

		assert.Equal(t, string(data[:len(data)-1]), msg)

		tcpStopSignal <- true
	})

	t.Run("test StartHTTPServer", func(t *testing.T) {
		var (
			httpStopSignal      = make(chan bool)
			wsConnectionHandler = func(connectionType string, client client_manager.IClient) {
				sameClient := client.GetHub().Get(client.GetID())

				assertInstance.Equal(connectionType, client_manager.WebSocketConnection)
				assertInstance.Equal(client, sameClient)

				for data := range client.GetReceiveChannel() {
					client.GetHub().Broadcast(data)
				}
			}
		)

		go StartHTTPServer(
			"localhost:3000",
			httpStopSignal,
			application.CreateRouter(wsConnectionHandler),
		)

		// wait for server to start
		jitter()

		httpStopSignal <- true
	})
}
