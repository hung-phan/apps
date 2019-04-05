package application

import (
	"github.com/hung-phan/apps/infrastructure"
	"time"
)

func TCPConnectionHandler(tcpClient infrastructure.Client) {
	ch := make(infrastructure.DataChannel)

	tcpClient.AddListener(ch)

	go func() {
		defer close(ch)

		for {
			if tcpClient.IsShutdown() {
				break
			}

			select {
			case data := <-ch:
				_, _ = tcpClient.Write(data)
			case <-time.After(5 * time.Second):
				// do nothing
			}
		}
	}()

	tcpClient.Start()
}

func WSConnectionHandler(wsClient infrastructure.Client) {
	ch := make(infrastructure.DataChannel)

	wsClient.AddListener(ch)

	go func() {
		defer close(ch)

		for {
			if wsClient.IsShutdown() {
				break
			}

			select {
			case data := <-ch:
				_, _ = wsClient.Write(data)
			case <-time.After(5 * time.Second):
				// do nothing
			}
		}
	}()

	wsClient.Start()
}
