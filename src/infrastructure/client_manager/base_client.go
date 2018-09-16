package client_manager

import (
	"github.com/hung-phan/chat-app/src/infrastructure/logger"
	"sync"
)

type BaseClient struct {
	ID         string
	Hub        *Hub
	listeners  []Listener
	stateMutex sync.RWMutex
}

func (baseClient *BaseClient) GetID() string {
	return baseClient.ID
}

func (baseClient *BaseClient) GetHub() *Hub {
	return baseClient.Hub
}

func (baseClient *BaseClient) AddListener(listener Listener) {
	baseClient.stateMutex.Lock()
	defer baseClient.stateMutex.Unlock()

	for _, item := range baseClient.listeners {
		if &item == &listener {
			baseClient.stateMutex.RUnlock()

			logger.Client.Fatal("found duplicated listener")

			return
		}
	}

	baseClient.listeners = append(baseClient.listeners, listener)
}

func (baseClient *BaseClient) RemoveListener(listener Listener) {
	baseClient.stateMutex.Lock()
	defer baseClient.stateMutex.Unlock()

	for index, item := range baseClient.listeners {
		if &item == &listener {
			baseClient.listeners = append(baseClient.listeners[:index], baseClient.listeners[index+1:]...)

			return
		}
	}
}

func (baseClient *BaseClient) Broadcast(data []byte) {
	baseClient.stateMutex.RLock()
	defer baseClient.stateMutex.RUnlock()

	for _, item := range baseClient.listeners {
		go item(data)
	}
}
