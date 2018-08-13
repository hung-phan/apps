package tcp

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net"
	"time"
)

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

func CreateChannelsForTCPConnection(address string) error {
	conn, err := CreateTCPConnection(address)

	if err != nil {
		return err
	}

	var (
		errorChan   = make(chan error)
		incomingMsg = make(chan []byte)
		outgoingMsg = make(chan []byte)
	)

	go func() {
		for {
			data, err := bufio.NewReader(conn).ReadBytes('\n')

			if err != nil {
				errorChan <- err

				if io.EOF == err {
					return
				}
			}

			incomingMsg <- data
		}
	}()

	go func() {
		for data := range outgoingMsg {
			fmt.Fprintln(conn, data)
		}
	}()

	return nil
}
