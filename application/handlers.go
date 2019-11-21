package application

import (
	"github.com/hung-phan/apps/infrastructure"
	"time"
)

func ConnectionHandler(client infrastructure.Client) {
	ch := make(infrastructure.DataChannel)

	client.AddListener(ch)

	go func() {
		defer close(ch)

		for {
			if client.IsShutdown() {
				break
			}

			select {
			case data := <-ch:
				_, _ = client.Write(data)
			case <-time.After(5 * time.Second):
				// do nothing
			}
		}
	}()

	client.Start()
}
