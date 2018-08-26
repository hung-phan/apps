package infrastructure

import (
	"sync"
)

type IClient interface {
	IChannelCommunication

	Flush() error
	Shutdown()
}

type Client struct {
	once sync.Once
	Hub  *Hub
	id   string
}
