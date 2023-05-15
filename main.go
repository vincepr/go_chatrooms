package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
)

func main() {
	// root ctx and CancelFunc that can cancel RetentionMap goroutines:
	rootCtx := context.Background()
	ctx, cancel := context.WithCancel(rootCtx)
	defer cancel()

	initAPI(ctx)
	log.Fatal(http.ListenAndServe(":5555", nil))
}

func initAPI(ctx context.Context) {
	manager := NewManager(ctx)

	http.Handle("/", http.FileServer(http.Dir("./public")))
	http.HandleFunc("/login", manager.loginHandler)
	http.HandleFunc("/ws", manager.serveWS)

	http.HandleFunc("/debug", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, len(manager.clients))
	})
}

// used to filter out Cross-Site trafic if needed.
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
