package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"sync"
)

var (
	ErrEventNotSupported = errors.New("this event type is not supported")
)

// holds reference to all Clienets currently connected (mutex for async)
type Manager struct {
	clients ClientList					// stores all connections
	sync.RWMutex						// needed for async safety
	handlers map[string]EventHandler	// stores all supported EventHandlers for different types.
}

// holds all the different Clients currently connected to that manager/websocket.
type ClientList map[*Client] bool

func NewManager() *Manager {
	m := &Manager{
		clients: make(ClientList),
		handlers: make(map[string]EventHandler),
	}
	m.SetupEventHandlers()
	return m
}

// configures and adds all the handlers
func (m *Manager) SetupEventHandlers(){
	m.handlers[EventSendMessage] = func(e Event, c *Client) error{
		fmt.Println(e)
		return nil
	}
}

func (m *Manager) routeEvent(event Event, c *Client) error {
	if handler, ok := m.handlers[event.Type]; ok{
		// handler is present in the map -> execute that EventHandler
		if err := handler(event, c); err != nil {
			return err
		}
		return nil
	} else {
		return ErrEventNotSupported
	}
}


// the WebSocket HandleFunc
func (m *Manager) serveWS(w http.ResponseWriter, r *http.Request) {

	log.Println("New connection upgrade request")
	conn, err := websocketUpgrader.Upgrade(w, r , nil)
	if err != nil {
		log.Println("Failed WF Upgrade:", err)
		return
	}

	client := NewClient(conn, m)
	m.addClient(client)
	go client.readMessages()
	go client.writeMessages()
}

// add the newly connected client to our List of all current clients
func (m *Manager) addClient(client *Client) {
	m.Lock()
	defer m.Unlock()
	m.clients[client] = true
}

// remove client and cleanup (example after they disconect/timeout)
func (m *Manager) removeClient(client *Client) {
	m.Lock()
	defer m.Unlock()
	if _, ok := m.clients[client]; ok {
		client.conn.Close()				// gracefully close the connection
		delete(m.clients, client)		// and delete the reference to the connection from current list
	}
}

