package infrastructure

import (
	"bufio"
	"github.com/stretchr/testify/assert"
	"math/rand"
	"testing"
	"time"
)

func TestStartTCPServer(t *testing.T) {
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
			tcpConnectionHandler = func(client Client) {
				ch := make(DataChannel)

				go func() {
					for data := range ch {
						client.Write(data)

						// enough time for client to send the message so we can force
						// the connection to flush it later
						jitter()

						client.Flush()
						client.RemoveListener(ch)

						close(ch)
					}
				}()

				client.AddListener(ch)
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
			NewHub(),
			tcpConnectionHandler,
		)

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