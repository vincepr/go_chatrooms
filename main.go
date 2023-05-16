/*
*		Main entry point for the server.
 */

package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
)

func main() {
	// root context and CancelFunc can cancel RetentionMap goroutines:
	rootCtx := context.Background()
	ctx, cancel := context.WithCancel(rootCtx)
	defer cancel()

	initAPI(ctx)
	log.Fatal(http.ListenAndServe(":5555", nil))
}

// setup the Routes and our Manager-struct controlling the Websockets
func initAPI(ctx context.Context) {
	manager := NewManager(ctx)

	http.Handle("/", http.FileServer(http.Dir("./public")))
	http.HandleFunc("/login", manager.loginHandler)
	http.HandleFunc("/ws", manager.serveWS)

	http.HandleFunc("/debug", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, len(manager.clients))
	})
}
