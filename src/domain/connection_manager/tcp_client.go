package connection_manager

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"sync"
	"time"
)

type TCPClient struct {
	Conn *net.TCPConn

	once      sync.Once
	hub       *Hub
	id        string
	rw        *bufio.ReadWriter
	sendCh    chan []byte
	receiveCh chan []byte
}

func (client *TCPClient) GetChannels() (chan []byte, chan []byte) {
	return client.sendCh, client.receiveCh
}

func (client *TCPClient) CleanUp() {

	close(client.sendCh)
	close(client.receiveCh)

	client.hub.Del(client.id)
	client.Conn.Close()
}

func (client *TCPClient) readPump() {
	defer client.once.Do(client.CleanUp)

	rw := bufio.NewReadWriter(bufio.NewReader(client.Conn), bufio.NewWriter(client.Conn))

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
		client.receiveCh <- cmd[:len(cmd)-1]
	}
}

func (client *TCPClient) writePump() {
	defer client.once.Do(client.CleanUp)

	for data := range client.sendCh {
		_, err := client.rw.Write(data)

		if err != nil {
			log.Fatalln("error:", err)
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

func NewTCPClient(hub *Hub, id string, conn *net.TCPConn) *TCPClient {
	client := &TCPClient{
		Conn:      conn,
		hub:       hub,
		id:        id,
		rw:        bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn)),
		sendCh:    make(chan []byte, 256),
		receiveCh: make(chan []byte, 256),
	}

	go client.readPump()
	go client.writePump()

	return client
}
