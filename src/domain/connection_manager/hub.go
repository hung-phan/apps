package connection_manager

import "sync"

var (
	DefaultWSHub  = NewHub()
	DefaultTCPHub = NewHub()
)

type Hub struct {
	m             sync.RWMutex
	Conns         map[string]IClient
	BroadcastChan chan []byte
}

func (hub *Hub) Set(key string, client IClient) {
	hub.m.Lock()
	defer hub.m.Unlock()

	hub.Conns[key] = client
}

func (hub *Hub) Get(key string) (IClient, bool) {
	hub.m.RLock()
	defer hub.m.RUnlock()

	conn, ok := hub.Conns[key]

	if ok {
		return conn, true
	} else {
		return nil, false
	}
}

func (hub *Hub) Del(key string) {
	hub.m.Lock()
	defer hub.m.Unlock()

	delete(hub.Conns, key)
}

func (hub *Hub) Broadcast(data []byte) {
	hub.BroadcastChan <- data
}

func (hub *Hub) Listen() {
	for data := range hub.BroadcastChan {
		hub.m.RLock()

		for _, client := range hub.Conns {
			go func() {
				client.GetSendChannel() <- data
			}()
		}

		hub.m.RUnlock()
	}
}

func (hub *Hub) StopListening() {
	close(hub.BroadcastChan)
}

func NewHub() *Hub {
	hub := &Hub{
		Conns:         make(map[string]IClient),
		BroadcastChan: make(chan []byte, 256),
	}

	go hub.Listen()

	return hub
}
