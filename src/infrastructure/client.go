package infrastructure

type OnDataListener func([]byte)

type ClientBroadcastResult struct {
	nn  int
	err error
}

type Client interface {
	GetID() string
	GetHub() *ClientHub
	AddListener(OnDataListener)
	RemoveListener(OnDataListener)
	Write([]byte) (int, error)
	Flush() error
	GracefulShutdown()
}
