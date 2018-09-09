package logger

import (
	"github.com/hung-phan/chat-app/src/common/singleflight"
	"go.uber.org/zap"
)

var (
	group  = singleflight.Group{}
	Client = createLogger()
)

func createLogger() *zap.Logger {
	logger, err := group.Do("GetLogger", func() (interface{}, error) {
		return zap.NewProduction()
	})

	if err != nil {
		panic(err)
	}

	return logger.(*zap.Logger)
}
