package application

import (
	"github.com/gorilla/mux"
)

func CreateRouter() *mux.Router {
	router := mux.NewRouter()

	router.HandleFunc("/ws", WSHandler)

	return router
}
