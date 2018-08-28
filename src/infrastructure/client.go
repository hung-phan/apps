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
	mutex sync.Mutex
	once  sync.Once
	Hub   *Hub
	id    string
}
