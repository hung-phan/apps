package infrastructure

import "sync"

type Hub struct {
	rwMutex       sync.RWMutex
	Clients       map[string]IClient
	BroadcastChan chan []byte
}

func (hub *Hub) Get(key string) (IClient, bool) {
	hub.rwMutex.RLock()
	defer hub.rwMutex.RUnlock()

	conn, ok := hub.Clients[key]

	if ok {
		return conn, true
	} else {
		return nil, false
	}
}

func (hub *Hub) Set(key string, client IClient) {
	hub.rwMutex.Lock()
	defer hub.rwMutex.Unlock()

	hub.Clients[key] = client
}

func (hub *Hub) Del(key string) {
	hub.rwMutex.Lock()
	defer hub.rwMutex.Unlock()

	delete(hub.Clients, key)
}

func (hub *Hub) Broadcast(data []byte) {
	hub.BroadcastChan <- data
}

func (hub *Hub) listen() {
	for data := range hub.BroadcastChan {
		hub.rwMutex.RLock()

		for _, client := range hub.Clients {
			go func(client IClient) {
				client.GetSendChannel() <- data
			}(client)
		}

		hub.rwMutex.RUnlock()
	}
}

func (hub *Hub) Shutdown() {
	close(hub.BroadcastChan)

	for _, client := range hub.Clients {
		client.Shutdown()
	}
}

func NewHub() *Hub {
	hub := &Hub{
		Clients:       make(map[string]IClient),
		BroadcastChan: make(chan []byte, 256),
	}

	go hub.listen()

	return hub
}
