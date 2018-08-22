package infrastructure

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/gorilla/websocket"
	"io"
	"log"
	"net"
	"sync"
	"time"
)

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

type TCPClient struct {
	*Client
	*ChannelCommunication

	rw   *bufio.ReadWriter
	Conn *net.TCPConn
}

func (tcpClient *TCPClient) cleanUp() {
	tcpClient.CloseAllChannels()
	tcpClient.hub.Del(tcpClient.id)
	tcpClient.Conn.Close()
}

func (tcpClient *TCPClient) readPump() {
	defer tcpClient.once.Do(tcpClient.cleanUp)

	rw := bufio.NewReadWriter(bufio.NewReader(tcpClient.Conn), bufio.NewWriter(tcpClient.Conn))

	// Read from the connection until EOF. Expect a command name as the
	// next input. Call the handler that is registered for this command.
	for {
		cmd, err := rw.ReadBytes('\n')

		switch {
		case err == io.EOF:
			log.Println("reached EOF - close this connection")
			return
		case err != nil:
			log.Println("Error reading command. Got:", err)
			return
		}

		// Trim the request string - ReadString does not strip any newlines.
		tcpClient.GetReceiveChannel() <- cmd[:len(cmd)-1]
	}
}

func (tcpClient *TCPClient) writePump() {
	defer tcpClient.once.Do(tcpClient.cleanUp)

	for data := range tcpClient.GetSendChannel() {
		_, err := tcpClient.rw.Write(data)

		if err != nil {
			log.Println("error:", err)
			return
		}
	}
}

type WSClient struct {
	*Client
	*ChannelCommunication

	Conn *websocket.Conn
}

func (wsClient *WSClient) cleanUp() {
	wsClient.sendCloseSignal()
	wsClient.CloseAllChannels()
	wsClient.hub.Del(wsClient.id)
	wsClient.Conn.Close()
}

func (wsClient *WSClient) sendCloseSignal() error {
	return wsClient.Conn.WriteMessage(websocket.CloseMessage, []byte{})
}

func (wsClient *WSClient) readPump() {
	defer wsClient.once.Do(wsClient.cleanUp)

	for {
		_, data, err := wsClient.Conn.ReadMessage()

		if err != nil {
			if websocket.IsUnexpectedCloseError(
				err,
				websocket.CloseGoingAway,
				websocket.CloseNoStatusReceived,
				websocket.CloseAbnormalClosure,
			) {
				log.Println("error:", err)
			}
			return
		}

		wsClient.GetReceiveChannel() <- data
	}
}

func (wsClient *WSClient) writePump() {
	defer wsClient.once.Do(wsClient.cleanUp)

	for data := range wsClient.GetSendChannel() {
		err := wsClient.Conn.WriteMessage(websocket.TextMessage, data)

		if err != nil {
			log.Println("error:", err)
			return
		}
	}
}

func NewWSClient(hub *Hub, id string, conn *websocket.Conn) *WSClient {
	client := &WSClient{
		Client: &Client{
			hub: hub,
			id:  id,
		},
		ChannelCommunication: &ChannelCommunication{
			sendCh:    make(chan []byte, 256),
			receiveCh: make(chan []byte, 256),
		},
		Conn: conn,
	}

	hub.Set(id, client)

	go client.readPump()
	go client.writePump()

	return client
}

func CreateTCPConnection(address string) (*net.TCPConn, error) {
	var (
		tcpAddr *net.TCPAddr
		conn    *net.TCPConn
		err     error
	)

	tcpAddr, err = net.ResolveTCPAddr("tcp4", address)

	if err != nil {
		return nil, errors.New(fmt.Sprintf("cannot resolve address %s", address))
	}

	conn, err = net.DialTCP("tcp", nil, tcpAddr)

	if err != nil {
		return nil, errors.New(fmt.Sprintf("cannot create TCP connection to address %s", address))
	}

	if err = conn.SetKeepAlive(true); err != nil {
		defer conn.Close()

		return nil, errors.New(fmt.Sprintf("cannot set KeepAlive for TCP connection to address %s", address))
	}

	if err = conn.SetKeepAlivePeriod(30 * time.Minute); err != nil {
		defer conn.Close()

		return nil, errors.New(fmt.Sprintf("cannot set KeepAlivePeriod for TCP connection to address %s", address))
	}

	return conn, nil
}

func NewTCPClient(hub *Hub, address string, conn *net.TCPConn) *TCPClient {
	client := &TCPClient{
		Client: &Client{
			hub: hub,
			id:  address,
		},
		ChannelCommunication: &ChannelCommunication{
			sendCh:    make(chan []byte, 256),
			receiveCh: make(chan []byte, 256),
		},
		Conn: conn,
		rw:   bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn)),
	}

	hub.Set(address, client)

	go client.readPump()
	go client.writePump()

	return client
}
