package main

import (
	"github.com/hung-phan/chat-app/src/application"
	"net/http"
)

func main() {
	http.ListenAndServe("localhost:3000", application.CreateRouter())
}
