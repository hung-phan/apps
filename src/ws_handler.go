package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/segmentio/ksuid"
	"net/http"
)

var wsConnections = make(map[ksuid.KSUID]*websocket.Conn)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func SendMessage(conn *websocket.Conn, messageType int, data []byte) error {
	return conn.WriteMessage(messageType, data)
}

func SendMessageViaId(id ksuid.KSUID, messageType int, data []byte) error {
	conn, prs := wsConnections[id]

	if prs == false {
		return errors.New(fmt.Sprintf("connection with ksuid (%s) doesn't exist", id))
	}

	return SendMessage(conn, messageType, data)
}

func HandleMessage(conn *websocket.Conn, messageType int, p []byte) error {
	var data map[string]interface{}

	if err := json.Unmarshal(p, &data); err != nil {
		return errors.New("invalid JSON data")
	}

	switch data["__type"] {
	case "chat.msg":
		id := data["from"].(ksuid.KSUID)
		msg := data["msg"].([]byte)

		if err := SendMessageViaId(id, websocket.TextMessage, msg); err != nil {
			// based on current id, get current config from db and establish tcp connection
			// if the cluster is in same availability zone. Else, send to the message queue
			// in that region and wait for it to process
		}
	}

	return nil
}

func WebSocketHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)

	if err != nil {
		return
	}

	// save generated ksuid to db for authenticated connection
	id := ksuid.New()
	wsConnections[id] = conn

	go func() {
		defer conn.Close()
		defer delete(wsConnections, id)

		for {
			messageType, p, err := conn.ReadMessage()

			if err != nil {
				return
			}

			HandleMessage(conn, messageType, p)
		}
	}()
}
