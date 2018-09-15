package client_manager

type IClient interface {
	GetID() string
	GetHub() *Hub
	GetSendChannel() chan []byte
	GetReceiveChannel() chan []byte
	Flush() error
	Shutdown()
}
