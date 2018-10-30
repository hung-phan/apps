package infrastructure

import (
	"bufio"
	"errors"
	"go.uber.org/zap"
	"net"
	"sync"
	"time"
)

var (
	ErrCannotResolveAddress         = errors.New("cannot resolve address")
	ErrCannotCreateTCPConnection    = errors.New("cannot create TCP connection")
	ErrCannotKeepAliveTCPConnection = errors.New("cannot set KeepAlive for TCP connection")
)

type TCPClient struct {
	*baseClient

	Conn    *net.TCPConn
	once    sync.Once
	ioMutex sync.Mutex
	rw      *bufio.ReadWriter
}

func (tcpClient *TCPClient) Write(data []byte) (int, error) {
	tcpClient.ioMutex.Lock()
	defer tcpClient.ioMutex.Unlock()

	return tcpClient.rw.Write(append(data, '\n'))
}

func (tcpClient *TCPClient) Flush() error {
	tcpClient.ioMutex.Lock()
	defer tcpClient.ioMutex.Unlock()

	return tcpClient.rw.Flush()
}

func (tcpClient *TCPClient) GracefulShutdown() {
	tcpClient.once.Do(tcpClient.shutdown)
}

func (tcpClient *TCPClient) shutdown() {
	var err error = nil

	err = tcpClient.Flush()

	if err != nil {
		Log.Error("fail to flush TCP", zap.Error(err))
	}

	err = tcpClient.Conn.Close()

	if err != nil {
		Log.Error("fail to close TCP connection", zap.Error(err))
	}

	tcpClient.Hub.Del(tcpClient.ID)
}

func (tcpClient *TCPClient) readPump() {
	defer tcpClient.GracefulShutdown()

	for {
		data, err := tcpClient.rw.ReadBytes('\n')

		if err != nil {
			Log.Debug("tcp read fail", zap.Error(err), zap.String("ID", tcpClient.ID))
			return
		}

		// trim the request string - ReadBytes does not strip any newlines
		go tcpClient.emit(data[:len(data)-1])
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

func NewTCPClient(hub *ClientHub, address string, conn *net.TCPConn) *TCPClient {
	client := &TCPClient{
		baseClient: &baseClient{
			ID:  address,
			Hub: hub,
		},
		Conn: conn,
		rw:   bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn)),
	}

	hub.Set(address, client)

	go client.readPump()

	return client
}
