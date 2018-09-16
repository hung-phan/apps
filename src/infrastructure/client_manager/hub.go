package client_manager

import "sync"

type Hub struct {
	rwMutex sync.RWMutex
	clients map[string]Client
}

func (hub *Hub) Get(key string) Client {
	hub.rwMutex.RLock()
	defer hub.rwMutex.RUnlock()

	if conn, ok := hub.clients[key]; ok {
		return conn
	}

	return nil
}

func (hub *Hub) Set(key string, client Client) {
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
	hub.rwMutex.RLock()
	defer hub.rwMutex.RUnlock()

	for _, client := range hub.clients {
		go client.Write(data)
	}
}

func (hub *Hub) Shutdown() {
	for _, client := range hub.clients {
		client.GracefulShutdown()
	}
}

func NewHub() *Hub {
	return &Hub{
		clients: make(map[string]Client),
	}
}
