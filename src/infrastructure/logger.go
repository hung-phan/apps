package infrastructure

import (
	"go.uber.org/zap"
)

var (
	Log = createLogger()
)

func createLogger() *zap.Logger {
	logger, err := zap.NewProduction()

	if err != nil {
		panic(err)
	}

	return logger
}
