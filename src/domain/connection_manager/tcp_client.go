package connection_manager

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"time"
)

type TCPClient struct {
	Client
	ChannelCommunication

	rw   *bufio.ReadWriter
	Conn *net.TCPConn
}

func (tcpClient *TCPClient) CleanUp() {
	tcpClient.CloseAllChannels()
	tcpClient.hub.Del(tcpClient.id)
	tcpClient.Conn.Close()
}

func (tcpClient *TCPClient) readPump() {
	defer tcpClient.once.Do(tcpClient.CleanUp)

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
			log.Fatalln("Error reading command. Got:", err)
			return
		}

		// Trim the request string - ReadString does not strip any newlines.
		tcpClient.ReceiveCh <- cmd[:len(cmd)-1]
	}
}

func (tcpClient *TCPClient) writePump() {
	defer tcpClient.once.Do(tcpClient.CleanUp)

	for data := range tcpClient.SendCh {
		_, err := tcpClient.rw.Write(data)

		if err != nil {
			log.Fatalln("error:", err)
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
		Client: Client{
			hub: hub,
			id:  address,
		},
		ChannelCommunication: ChannelCommunication{
			SendCh:    make(chan []byte, 256),
			ReceiveCh: make(chan []byte, 256),
		},
		Conn: conn,
		rw:   bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn)),
	}

	hub.Set(address, client)

	go client.readPump()
	go client.writePump()

	return client
}
