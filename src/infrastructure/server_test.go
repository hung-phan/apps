package infrastructure

import (
	"bufio"
	"github.com/stretchr/testify/assert"
	"math/rand"
	"testing"
	"time"
)

func TestServer(t *testing.T) {
	assertInstance := assert.New(t)

	t.Run("Test StartTCPServer", func(t *testing.T) {
		var (
			tcpStopSignal = make(chan bool)
			msg           = "Message"
			jitter = func() {
				time.Sleep(time.Duration(rand.Intn(200)) * time.Millisecond)
			}
			connectionHandler = func(connectionType string, client IClient) {
				assertInstance.Equal(connectionType, TcpConnection)

				for data := range client.GetReceiveChannel() {
					client.GetSendChannel() <- data

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

		go StartTCPServer("localhost:3001", tcpStopSignal, connectionHandler)

		// wait for server to start
		jitter()

		tcpConn, err := CreateTCPConnection("localhost:3001")
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
}
