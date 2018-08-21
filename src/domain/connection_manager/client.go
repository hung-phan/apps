package connection_manager

import "sync"

type IChannelCommunication interface {
	GetSendChannel() chan []byte
	GetReceiveChannel() chan []byte
	CloseAllChannels()
}

type IClient interface {
	IChannelCommunication

	cleanUp()
}

type ChannelCommunication struct {
	sendCh    chan []byte
	receiveCh chan []byte
}

func (cc *ChannelCommunication) GetSendChannel() chan []byte {
	return cc.sendCh
}

func (cc *ChannelCommunication) GetReceiveChannel() chan []byte {
	return cc.sendCh
}

func (cc *ChannelCommunication) CloseAllChannels() {
	close(cc.sendCh)
	close(cc.receiveCh)
}

type Client struct {
	once sync.Once
	hub  *Hub
	id   string
}
