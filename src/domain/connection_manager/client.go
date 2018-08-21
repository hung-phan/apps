package connection_manager

type IClient interface {
	GetChannels() (chan []byte, chan []byte)
	CleanUp()

	readPump()
	writePump()
}
