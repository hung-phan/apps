package client_manager

type Client interface {
	GetID() string
	GetHub() *Hub
	GetSendChannel() chan []byte
	GetReceiveChannel() chan []byte
	Flush() error
	Shutdown()
}
