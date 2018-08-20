package tcp

import (
	"sync"
)

var DefaultHub = NewHub()

type Hub struct {
	Conns         map[string]*Client
	BroadcastChan chan []byte
	m             sync.RWMutex
}

func (hub *Hub) Set(key string, client *Client) {
	hub.m.Lock()
	defer hub.m.Unlock()

	hub.Conns[key] = client
}

func (hub *Hub) Get(key string) (*Client, bool) {
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
				client.Send <- data
			}()
		}

		hub.m.RUnlock()
	}
}

func (hub *Hub) StopListening() {
	close(hub.BroadcastChan)
}

func NewHub() *Hub {
	hub := &Hub{BroadcastChan: make(chan []byte, 256)}

	go hub.Listen()

	return hub
}
