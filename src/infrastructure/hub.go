package infrastructure

import "sync"

type Hub struct {
	m             sync.RWMutex
	Clients       map[string]IClient
	BroadcastChan chan []byte
}

func (hub *Hub) Set(key string, client IClient) {
	hub.m.Lock()
	defer hub.m.Unlock()

	hub.Clients[key] = client
}

func (hub *Hub) Get(key string) (IClient, bool) {
	hub.m.RLock()
	defer hub.m.RUnlock()

	conn, ok := hub.Clients[key]

	if ok {
		return conn, true
	} else {
		return nil, false
	}
}

func (hub *Hub) Del(key string) {
	hub.m.Lock()
	defer hub.m.Unlock()

	delete(hub.Clients, key)
}

func (hub *Hub) Broadcast(data []byte) {
	hub.BroadcastChan <- data
}

func (hub *Hub) Listen() {
	for data := range hub.BroadcastChan {
		hub.m.RLock()

		for _, client := range hub.Clients {
			go func() {
				client.GetSendChannel() <- data
			}()
		}

		hub.m.RUnlock()
	}
}

func (hub *Hub) Shutdown() {
	close(hub.BroadcastChan)

	for _, client := range hub.Clients {
		client.CleanUp()
	}
}

func NewHub() *Hub {
	hub := &Hub{
		Clients:       make(map[string]IClient),
		BroadcastChan: make(chan []byte, 256),
	}

	go hub.Listen()

	return hub
}
