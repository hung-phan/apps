package client_manager

type IChannelCommunication interface {
	GetSendChannel() chan []byte
	GetReceiveChannel() chan []byte
	CloseAllChannels()
}

type ChannelCommunication struct {
	sendCh    chan []byte
	receiveCh chan []byte
}

func (cc *ChannelCommunication) GetSendChannel() chan []byte {
	return cc.sendCh
}

func (cc *ChannelCommunication) GetReceiveChannel() chan []byte {
	return cc.receiveCh
}

func (cc *ChannelCommunication) CloseAllChannels() {
	close(cc.sendCh)
	close(cc.receiveCh)
}
