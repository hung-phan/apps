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

	conn         *net.TCPConn
	startOnce    sync.Once
	shutdownOnce sync.Once
	m            sync.Mutex
	rw           *bufio.ReadWriter
}

func (tcpClient *TCPClient) Start() {
	tcpClient.m.Lock()
	defer tcpClient.m.Unlock()

	tcpClient.isStarted = true

	go tcpClient.startOnce.Do(tcpClient.readPump)
}

func (tcpClient *TCPClient) Shutdown() {
	tcpClient.m.Lock()
	defer tcpClient.m.Unlock()

	tcpClient.isShutdown = true

	go tcpClient.shutdownOnce.Do(tcpClient.gracefulShutdown)
}

func (tcpClient *TCPClient) Write(data []byte) (int, error) {
	tcpClient.m.Lock()
	defer tcpClient.m.Unlock()

	return tcpClient.rw.Write(append(data, '\n'))
}

func (tcpClient *TCPClient) Flush() error {
	tcpClient.m.Lock()
	defer tcpClient.m.Unlock()

	return tcpClient.rw.Flush()
}

func (tcpClient *TCPClient) readPump() {
	defer tcpClient.Shutdown()

	for {
		data, err := tcpClient.rw.ReadBytes('\n')

		if err != nil {
			Log.Debug("tcp read fail", zap.Error(err), zap.String("id", tcpClient.GetID()))
			return
		}

		// trim the request string - ReadBytes does not strip any newlines
		go tcpClient.broadcastToChannel(data[:len(data)-1])
	}
}

func (tcpClient *TCPClient) gracefulShutdown() {
	err := tcpClient.Flush()

	if err != nil {
		Log.Error("fail to flush TCP", zap.Error(err))
	}

	err = tcpClient.conn.Close()

	if err != nil {
		Log.Error("fail to close TCP connection", zap.Error(err))
	}

	tcpClient.GetHub().Del(tcpClient.GetID())
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
		defer func() {
			if err := conn.Close(); err != nil {
				Log.Error("fail to close TCP connection", zap.Error(err))
			}
		}()

		return nil, ErrCannotKeepAliveTCPConnection
	}

	if err = conn.SetKeepAlivePeriod(30 * time.Minute); err != nil {
		defer func() {
			if err := conn.Close(); err != nil {
				Log.Error("fail to close TCP connection", zap.Error(err))
			}
		}()

		return nil, ErrCannotKeepAliveTCPConnection
	}

	return conn, nil
}

func NewTCPClient(address string, conn *net.TCPConn) *TCPClient {
	client := &TCPClient{
		baseClient: &baseClient{
			id:  address,
			hub: DefaultTCPHub,
		},
		conn: conn,
		rw:   bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn)),
	}

	DefaultTCPHub.Set(address, client)

	return client
}
