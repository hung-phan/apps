package infrastructure

import (
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/segmentio/ksuid"
	"go.uber.org/zap"
	"math/rand"
	"net/http"
	"sync"
	"testing"
	"time"
)

func TestStartHTTPServer(t *testing.T) {
	var (
		jitter = func() {
			time.Sleep(time.Duration(rand.Intn(200)) * time.Millisecond)
		}
	)

	t.Run("test StartHTTPServer", func(t *testing.T) {
		var (
			router            = mux.NewRouter()
			wg                = &sync.WaitGroup{}
			httpStopSignal    = make(chan bool)
			webSocketUpgrader = websocket.Upgrader{
				CheckOrigin: func(r *http.Request) bool {
					return true
				},
			}
			wsConnectionHandler = func(client Client) {
				ch := make(DataChannel)

				client.AddListener(ch)

				go func() {
					for data := range ch {
						_, _ = client.Write(data)

						client.RemoveListener(ch)

						close(ch)
					}
				}()

				client.Start()
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
					ksuid.New().String(),
					conn,
				))
			},
		)

		go StartHTTPServer(
			"localhost:3000",
			httpStopSignal,
			wg,
			router,
		)

		// wait for server to start
		jitter()

		httpStopSignal <- true
		wg.Wait()
	})
}
