package logger

import (
	"go.uber.org/zap"
)

var (
	Client = createLogger()
)

func createLogger() *zap.Logger {
	logger, err := zap.NewProduction()

	if err != nil {
		panic(err)
	}

	return logger
}
