package connection_manager

import "sync"

type IChannelCommunication interface {
	GetSendChannel() chan []byte
	GetReceiveChannel() chan []byte
	CloseAllChannels()
}

type IClient interface {
	IChannelCommunication

	CleanUp()
}

type ChannelCommunication struct {
	SendCh    chan []byte
	ReceiveCh chan []byte
}

func (cc *ChannelCommunication) GetSendChannel() chan []byte {
	return cc.SendCh
}

func (cc *ChannelCommunication) GetReceiveChannel() chan []byte {
	return cc.SendCh
}

func (cc *ChannelCommunication) CloseAllChannels() {
	close(cc.SendCh)
	close(cc.ReceiveCh)
}

type Client struct {
	once sync.Once
	hub  *Hub
	id   string
}
