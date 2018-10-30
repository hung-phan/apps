package infrastructure

import "sync"

type ClientHub struct {
	rwMutex sync.RWMutex
	clients map[string]Client
}

func (ch *ClientHub) Get(key string) Client {
	ch.rwMutex.RLock()
	defer ch.rwMutex.RUnlock()

	if conn, ok := ch.clients[key]; ok {
		return conn
	}

	return nil
}

func (ch *ClientHub) Set(key string, client Client) {
	ch.rwMutex.Lock()
	defer ch.rwMutex.Unlock()

	ch.clients[key] = client
}

func (ch *ClientHub) Del(key string) {
	ch.rwMutex.Lock()
	defer ch.rwMutex.Unlock()

	delete(ch.clients, key)
}

func (ch *ClientHub) ExecuteAll(fn ClientHandler) {
	ch.rwMutex.RLock()
	defer ch.rwMutex.RUnlock()

	for _, client := range ch.clients {
		go fn(client)
	}
}

func NewHub() *ClientHub {
	return &ClientHub{
		clients: make(map[string]Client),
	}
}
