package infrastructure

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestTCPConnection(t *testing.T) {
	assertInstance := assert.New(t)

	t.Run("should be able to send and receive message from TCP connection", func(t *testing.T) {
		const ClientMessage = "Client message"

		done := make(chan bool)

		go StartTCPServer(":3001", func(connectionType string, client IClient) {
			receiveCh := client.GetReceiveChannel()

			assertInstance.Equal(connectionType, TcpConnection)

			for data := range receiveCh {
				assert.Equal(t, string(data), ClientMessage)

				go client.ScheduleForCleanUp()
			}

			done <- true
		})

		tcpConn, err := CreateTCPConnection(":3001")

		if err != nil {
			assertInstance.Error(err)
		}

		_, err = tcpConn.Write([]byte(ClientMessage + "\n"))

		if err != nil {
			assertInstance.Error(err)
		}

		<-done
		// shutdown tcp server
	})
}
