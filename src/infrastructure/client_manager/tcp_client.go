package client_manager

import (
	"bufio"
	"errors"
	"github.com/hung-phan/chat-app/src/infrastructure/logger"
	"go.uber.org/zap"
	"net"
	"time"
)

var (
	ErrCannotResolveAddress         = errors.New("cannot resolve address")
	ErrCannotCreateTCPConnection    = errors.New("cannot create TCP connection")
	ErrCannotKeepAliveTCPConnection = errors.New("cannot set KeepAlive for TCP connection")
)

type TCPClient struct {
	*BaseClient

	Conn *net.TCPConn

	rw *bufio.ReadWriter
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
	tcpClient.Flush()
	tcpClient.Conn.Close()
	tcpClient.Hub.Del(tcpClient.ID)
	tcpClient.shutdownChannels()
}

func (tcpClient *TCPClient) readPump() {
	defer tcpClient.Shutdown()

	for {
		data, err := tcpClient.rw.ReadBytes('\n')

		if err != nil {
			logger.Client.Debug("tcp read fail", zap.Error(err), zap.String("ID", tcpClient.ID))
			return
		}

		// trim the request string - ReadBytes does not strip any newlines
		tcpClient.receiveCh <- data[:len(data)-1]
	}
}

func (tcpClient *TCPClient) writePump() {
	defer tcpClient.Shutdown()

	for data := range tcpClient.sendCh {
		newData := append(data, '\n')

		tcpClient.mutex.Lock()
		_, err := tcpClient.rw.Write(newData)
		tcpClient.mutex.Unlock()

		if err != nil {
			logger.Client.Debug("tcp write fail", zap.Error(err), zap.String("ID", tcpClient.ID))
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

	if tcpAddr, err = net.ResolveTCPAddr("tcp4", address); err != nil {
		return nil, ErrCannotResolveAddress
	}

	if conn, err = net.DialTCP("tcp", nil, tcpAddr); err != nil {
		return nil, ErrCannotCreateTCPConnection
	}

	if err = conn.SetKeepAlive(true); err != nil {
		defer conn.Close()

		return nil, ErrCannotKeepAliveTCPConnection
	}

	if err = conn.SetKeepAlivePeriod(30 * time.Minute); err != nil {
		defer conn.Close()

		return nil, ErrCannotKeepAliveTCPConnection
	}

	return conn, nil
}

func NewTCPClient(hub *Hub, address string, conn *net.TCPConn) *TCPClient {
	client := &TCPClient{
		BaseClient: &BaseClient{
			ID:  address,
			Hub: hub,

			sendCh:    make(chan []byte, 16),
			receiveCh: make(chan []byte, 16),
		},
		Conn: conn,
		rw:   bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn)),
	}

	hub.Set(address, client)

	go client.readPump()
	go client.writePump()

	return client
}
