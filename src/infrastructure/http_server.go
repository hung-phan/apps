package infrastructure

import (
	"context"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"net/http"
	"sync"
	"time"
)

func StartHTTPServer(
	address string,
	stopSignal chan bool,
	wg *sync.WaitGroup,
	router *mux.Router,
) {
	wg.Add(1)

	server := &http.Server{Addr: address, Handler: router}

	go func() {
		Log.Info("start HTTP server", zap.String("address", address))

		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			Log.Fatal("failed to start HTTP server:", zap.Error(err))
		}
	}()

	<-stopSignal

	DefaultWSHub.ExecuteAll(func(client Client) {
		Log.Info("Shutdown client", zap.String("id", client.GetID()))
		client.Shutdown()
	})

	Log.Info("Shutdown HTTP server")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)

	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		Log.Fatal("failed to gracefulShutdown HTTP server:", zap.Error(err))
	}

	wg.Done()
}
