package client_manager

type Listener func([]byte)

type Client interface {
	GetID() string
	GetHub() *Hub
	AddListener(Listener)
	RemoveListener(Listener)
	Broadcast([]byte)
	Write([]byte) (int, error)
	Flush() error
	GracefulShutdown()
}
