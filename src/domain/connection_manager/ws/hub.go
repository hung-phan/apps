package ws

import (
	"github.com/segmentio/ksuid"
	"sync"
)

var DefaultHub = NewHub()

type Hub struct {
	conns         sync.Map
	BroadcastChan chan []byte
}

func (hub *Hub) Set(key ksuid.KSUID, client *Client) {
	hub.conns.Store(key, client)
}

func (hub *Hub) Get(key ksuid.KSUID) (*Client, bool) {
	conn, ok := hub.conns.Load(key)

	if ok {
		return conn.(*Client), true
	} else {
		return nil, false
	}
}

func (hub *Hub) Del(key ksuid.KSUID) {
	hub.conns.Delete(key)
}

func (hub *Hub) Broadcast(data []byte) {
	hub.BroadcastChan <- data
}

func (hub *Hub) Listen() {
	for data := range hub.BroadcastChan {
		hub.conns.Range(func(key, value interface{}) bool {
			client := value.(*Client)

			go func() {
				client.Send <- data
			}()

			return true
		})
	}
}

func NewHub() *Hub {
	hub := &Hub{BroadcastChan: make(chan []byte, 256)}

	go hub.Listen()

	return hub
}
