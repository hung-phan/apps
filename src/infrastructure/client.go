package infrastructure

type OnDataListener func([]byte)

type ClientHandler func(Client)

type Client interface {
	GetID() string
	GetHub() *ClientHub
	AddListener(OnDataListener)
	RemoveListener(OnDataListener)
	Write([]byte) (int, error)
	Flush() error
	GracefulShutdown()
}
