package infrastructure

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
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

	Flush() error
	CleanUp()
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

type Client struct {
	once sync.Once
	Hub  *Hub
	id   string
}

type TCPClient struct {
	*Client
	*ChannelCommunication

	rw   *bufio.ReadWriter
	Conn *net.TCPConn
}

func (tcpClient *TCPClient) Flush() error {
	return tcpClient.rw.Flush()
}

func (tcpClient *TCPClient) CleanUp() {
	tcpClient.once.Do(tcpClient.scheduleForShutdown)
}

func (tcpClient *TCPClient) scheduleForShutdown() {
	tcpClient.CloseAllChannels()

	tcpClient.Flush()
	tcpClient.Conn.Close()
	tcpClient.Hub.Del(tcpClient.id)
}

func (tcpClient *TCPClient) readPump() {
	defer tcpClient.CleanUp()

	for {
		data, err := tcpClient.rw.ReadBytes('\n')

		if err != nil {
			logrus.WithFields(logrus.Fields{"id": tcpClient.id}).Debug(err)
			return
		}

		// trim the request string - ReadBytes does not strip any newlines
		tcpClient.GetReceiveChannel() <- data[:len(data)-1]
	}
}

func (tcpClient *TCPClient) writePump() {
	defer tcpClient.CleanUp()

	for data := range tcpClient.GetSendChannel() {
		newData := append(data, '\n')

		_, err := tcpClient.rw.Write(newData)

		if err != nil {
			logrus.WithFields(logrus.Fields{"id": tcpClient.id}).Debug(err)
			return
		}
	}
}

type WSClient struct {
	*Client
	*ChannelCommunication

	Conn *websocket.Conn
}

func (wsClient *WSClient) Flush() error {
	return nil
}

func (wsClient *WSClient) CleanUp() {
	wsClient.once.Do(wsClient.scheduleForShutdown)
}

func (wsClient *WSClient) scheduleForShutdown() {
	wsClient.sendCloseSignal()
	wsClient.CloseAllChannels()

	wsClient.Conn.Close()
	wsClient.Hub.Del(wsClient.id)
}

func (wsClient *WSClient) sendCloseSignal() error {
	return wsClient.Conn.WriteMessage(websocket.CloseMessage, []byte{})
}

func (wsClient *WSClient) readPump() {
	defer wsClient.CleanUp()

	for {
		_, data, err := wsClient.Conn.ReadMessage()

		if err != nil {
			if websocket.IsUnexpectedCloseError(
				err,
				websocket.CloseGoingAway,
				websocket.CloseNoStatusReceived,
				websocket.CloseAbnormalClosure,
			) {
				logrus.WithFields(logrus.Fields{"id": wsClient.id}).Debug(err)
			}
			return
		}

		wsClient.GetReceiveChannel() <- data
	}
}

func (wsClient *WSClient) writePump() {
	defer wsClient.CleanUp()

	for data := range wsClient.GetSendChannel() {
		err := wsClient.Conn.WriteMessage(websocket.TextMessage, data)

		if err != nil {
			logrus.WithFields(logrus.Fields{"id": wsClient.id}).Debug(err)
			return
		}
	}
}

func NewWSClient(hub *Hub, id string, conn *websocket.Conn) *WSClient {
	client := &WSClient{
		Client: &Client{
			Hub: hub,
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
			Hub: hub,
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
