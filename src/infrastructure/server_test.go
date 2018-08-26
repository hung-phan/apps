package infrastructure

import (
	"bufio"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestTCPConnection(t *testing.T) {
	assertInstance := assert.New(t)

	t.Run("should be able to send and receive message from TCP connection", func(t *testing.T) {
		var (
			tcpStopSignal     = make(chan bool)
			clientMessage     = "Message from client"
			serverMessage     = "Message from server"
			connectionHandler = func(connectionType string, client IClient) {
				assertInstance.Equal(connectionType, TcpConnection)

				for data := range client.GetReceiveChannel() {
					assert.Equal(t, string(data), clientMessage)

					client.GetSendChannel() <- []byte(serverMessage)

					time.Sleep(1 * time.Second)

					client.Flush()
				}
			}
			assertError = func(err error) {
				if err != nil {
					assertInstance.Fail(err.Error())
				}
			}
		)

		go StartTCPServer(":3001", tcpStopSignal, connectionHandler)

		tcpConn, err := CreateTCPConnection(":3001")
		assertError(err)

		rw := bufio.NewReadWriter(bufio.NewReader(tcpConn), bufio.NewWriter(tcpConn))

		_, err = rw.Write([]byte(clientMessage + "\n"))
		assertError(err)

		err = rw.Flush()
		assertError(err)

		data, err := rw.ReadBytes('\n')
		assertError(err)

		assert.Equal(t, string(data[:len(data)-1]), serverMessage)

		tcpStopSignal <- true
	})
}
