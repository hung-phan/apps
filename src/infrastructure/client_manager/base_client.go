package client_manager

import "sync"

type BaseClient struct {
	ID  string
	Hub *Hub

	mutex     sync.Mutex
	once      sync.Once
	sendCh    chan []byte
	receiveCh chan []byte
}

func (client *BaseClient) GetID() string {
	return client.ID
}

func (client *BaseClient) GetHub() *Hub {
	return client.Hub
}

func (client *BaseClient) GetSendChannel() chan []byte {
	return client.sendCh
}

func (client *BaseClient) GetReceiveChannel() chan []byte {
	return client.receiveCh
}

func (client *BaseClient) shutdownChannels() {
	close(client.sendCh)
	close(client.receiveCh)
}
