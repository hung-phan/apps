package client_manager

import (
	"sync"
)

type IClient interface {
	IChannelCommunication

	GetID() string
	GetHub() *Hub

	Flush() error
	Shutdown()
}

type Client struct {
	mutex sync.Mutex
	once  sync.Once
	Hub   *Hub
	ID    string
}

func (client *Client) GetID() string {
	return client.ID
}

func (client *Client) GetHub() *Hub {
	return client.Hub
}
