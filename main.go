package main

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var (
	// upgrades the incoming HTTP(S) request to a Websocket (or at least tries to)
	websocketUpgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		// to avoid Cross-Site Requests we can set a allowed origin:
		CheckOrigin: checkOrigin,
	}
)

func main() {
	initAPI()
	log.Fatal(http.ListenAndServe(":5555", nil))
}

func initAPI() {
	manager := NewManager()

	http.Handle("/", http.FileServer(http.Dir("./public")))
	http.HandleFunc("/ws", manager.serveWS)
}

// checkOrigin will check origin and return true if its allowed
func checkOrigin(r *http.Request) bool {
	// Grab the request origin
	origin := r.Header.Get("Origin")

	switch origin {
	case "http://localhost:8080":
		return true
	case "http://vprobst.de:5555":
		return true
	default:
		return false
	}
}
