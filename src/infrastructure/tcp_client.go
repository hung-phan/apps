package infrastructure

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"net"
	"time"
)

type TCPClient struct {
	*Client
	*ChannelCommunication

	rw   *bufio.ReadWriter
	Conn *net.TCPConn
}

func (tcpClient *TCPClient) Flush() error {
	tcpClient.mutex.Lock()
	defer tcpClient.mutex.Unlock()

	return tcpClient.rw.Flush()
}

func (tcpClient *TCPClient) Shutdown() {
	tcpClient.once.Do(tcpClient.scheduleForShutdown)
}

func (tcpClient *TCPClient) scheduleForShutdown() {
	tcpClient.CloseAllChannels()

	tcpClient.Flush()
	tcpClient.Conn.Close()
	tcpClient.Hub.Del(tcpClient.id)
}

func (tcpClient *TCPClient) readPump() {
	defer tcpClient.Shutdown()

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
	defer tcpClient.Shutdown()

	for data := range tcpClient.GetSendChannel() {
		newData := append(data, '\n')

		tcpClient.mutex.Lock()
		_, err := tcpClient.rw.Write(newData)
		tcpClient.mutex.Unlock()

		if err != nil {
			logrus.WithFields(logrus.Fields{"id": tcpClient.id}).Debug(err)
			return
		}
	}
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
