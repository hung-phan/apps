package infrastructure

import (
	"context"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"net/http"
	"time"
)

func StartHTTPServer(
	address string,
	shutdownSignal chan bool,
	router *mux.Router,
) {
	server := &http.Server{Addr: address, Handler: router}

	go func() {
		Log.Info("start HTTP server", zap.String("address", address))

		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			Log.Fatal("failed to start HTTP server:", zap.Error(err))
		}
	}()

	<-shutdownSignal

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)

	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		Log.Fatal("failed to gracefulShutdown HTTP server:", zap.Error(err))
	}
}
