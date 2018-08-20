package tcp

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

type Client struct {
	once    sync.Once
	rw      *bufio.ReadWriter
	Conn    *net.TCPConn
	Send    chan []byte
	Receive chan []byte
}

func (client *Client) cleanUp() {
	client.Conn.Close()

	close(client.Send)
	close(client.Receive)
}

func (client *Client) readPump() {
	defer client.once.Do(client.cleanUp)

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
			log.Fatalln("Error reading command. Got: ", err)
			return
		}

		// Trim the request string - ReadString does not strip any newlines.
		client.Receive <- cmd[:len(cmd)-1]
	}
}

func (client *Client) writePump() {
	defer client.once.Do(client.cleanUp)

	for data := range client.Send {
		_, err := client.rw.Write(data)

		if err != nil {
			log.Fatalf("error: %v\n", err)
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

func NewTCPClient(conn *net.TCPConn) *Client {
	client := &Client{
		rw:      bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn)),
		Conn:    conn,
		Send:    make(chan []byte, 256),
		Receive: make(chan []byte, 256),
	}

	go client.readPump()
	go client.writePump()

	return client
}
