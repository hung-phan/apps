package main

import (
	"github.com/gorilla/mux"
	"net/http"
)

func startWebserver() {
	r := mux.NewRouter()

	r.HandleFunc("/ws", WebSocketHandler)

	http.ListenAndServe("localhost:3000", r)
}

func main() {
	startWebserver()
}
