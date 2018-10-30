package infrastructure

import (
	"bufio"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/segmentio/ksuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"math/rand"
	"net/http"
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

	t.Run("test StartHTTPServer", func(t *testing.T) {
		var (
			router            = mux.NewRouter()
			httpStopSignal    = make(chan bool)
			webSocketUpgrader = websocket.Upgrader{
				CheckOrigin: func(r *http.Request) bool {
					return true
				},
			}
			wsConnectionHandler = func(client Client) {
				ch := make(DataChannel)

				go func() {
					for data := range ch {
						client.Write(data)

						client.RemoveListener(ch)

						close(ch)
					}
				}()

				client.AddListener(ch)
			}
		)

		router.HandleFunc(
			"/ws",
			func(w http.ResponseWriter, r *http.Request) {
				conn, err := webSocketUpgrader.Upgrade(w, r, nil)

				if err != nil {
					Log.Fatal("websocket upgrade fail", zap.Error(err))
					return
				}

				go wsConnectionHandler(NewWSClient(
					NewHub(),
					ksuid.New().String(),
					conn,
				))
			},
		)

		go StartHTTPServer(
			"localhost:3000",
			httpStopSignal,
			router,
		)

		// wait for server to start
		jitter()

		httpStopSignal <- true
	})
}
