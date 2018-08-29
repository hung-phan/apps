package client_manager

import "sync"

type Hub struct {
	rwMutex     sync.RWMutex
	clients     map[string]IClient
	broadcastCh chan []byte
}

func (hub *Hub) Get(key string) IClient {
	hub.rwMutex.RLock()
	defer hub.rwMutex.RUnlock()

	if conn, ok := hub.clients[key]; ok {
		return conn
	}

	return nil
}

func (hub *Hub) Set(key string, client IClient) {
	hub.rwMutex.Lock()
	defer hub.rwMutex.Unlock()

	hub.clients[key] = client
}

func (hub *Hub) Del(key string) {
	hub.rwMutex.Lock()
	defer hub.rwMutex.Unlock()

	delete(hub.clients, key)
}

func (hub *Hub) Broadcast(data []byte) {
	hub.broadcastCh <- data
}

func (hub *Hub) listen() {
	for data := range hub.broadcastCh {
		hub.rwMutex.RLock()

		for _, client := range hub.clients {
			go func(client IClient) {
				client.GetSendChannel() <- data
			}(client)
		}

		hub.rwMutex.RUnlock()
	}
}

func (hub *Hub) Shutdown() {
	close(hub.broadcastCh)

	for _, client := range hub.clients {
		client.Shutdown()
	}
}

func NewHub() *Hub {
	hub := &Hub{
		clients:     make(map[string]IClient),
		broadcastCh: make(chan []byte, 256),
	}

	go hub.listen()

	return hub
}
