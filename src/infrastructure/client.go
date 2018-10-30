package infrastructure

type DataChannel chan []byte

type ClientHandler func(Client)

type Client interface {
	GetID() string
	GetHub() *ClientHub
	AddListener(DataChannel)
	RemoveListener(DataChannel)
	Write([]byte) (int, error)
	Flush() error
	GracefulShutdown()
}
