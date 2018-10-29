package infrastructure

import (
	"sync"
)

type baseClient struct {
	ID             string
	Hub            *ClientHub
	listeners      []OnDataListener
	listenersMutex sync.RWMutex
}

func (bc *baseClient) GetID() string {
	return bc.ID
}

func (bc *baseClient) GetHub() *ClientHub {
	return bc.Hub
}

func (bc *baseClient) AddListener(listener OnDataListener) {
	bc.listenersMutex.Lock()
	defer bc.listenersMutex.Unlock()

	for _, item := range bc.listeners {
		if &item == &listener {
			bc.listenersMutex.RUnlock()

			Log.Warn("found duplicated listener")

			return
		}
	}

	bc.listeners = append(bc.listeners, listener)
}

func (bc *baseClient) RemoveListener(listener OnDataListener) {
	bc.listenersMutex.Lock()
	defer bc.listenersMutex.Unlock()

	for index, item := range bc.listeners {
		if &item == &listener {
			bc.listeners = append(bc.listeners[:index], bc.listeners[index+1:]...)

			return
		}
	}
}

func (bc *baseClient) Emit(data []byte) {
	bc.listenersMutex.RLock()
	defer bc.listenersMutex.RUnlock()

	for _, listener := range bc.listeners {
		go listener(data)
	}
}
