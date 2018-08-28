package infrastructure

import (
	"sync"
)

type IClient interface {
	IChannelCommunication

	Flush() error
	Shutdown()
	GetHub() *Hub
}

type Client struct {
	mutex sync.Mutex
	once  sync.Once
	Hub   *Hub
	ID    string
}

func (client *Client) GetHub() *Hub {
	return client.Hub
}
