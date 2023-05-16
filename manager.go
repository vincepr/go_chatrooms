/*		Manager controlls all Client that connect trough Websockets
*		We store our client list here that we savely async read write to etc.
 */

package main

import (
	"context"
	"errors"
	"sync"
	"time"
)

var (
	ErrEventNotSupported = errors.New("this event type is not supported")
)

// holds reference to all Clienets currently connected (mutex for async)
type Manager struct {
	clients      ClientList              // stores all connections
	sync.RWMutex                         // needed for async safety
	handlers     map[string]EventHandler // stores all supported EventHandlers for different types.
	otps         RetentionMap            // holds all valid OTPs (One-Time-Passwords)
}

// holds all the different Clients currently connected to that manager/websocket.
type ClientList map[*Client]bool

func NewManager(ctx context.Context) *Manager {
	m := &Manager{
		clients:  make(ClientList),
		handlers: make(map[string]EventHandler),
		otps:     NewRetentionMap(ctx, 5*time.Second),
	}
	m.SetupEventHandlers()
	return m
}

// configures and adds all the handlers
func (m *Manager) SetupEventHandlers() {
	m.handlers[EventSendMessage] = SendMessageHandler
	m.handlers[EventChangeRoom] = ChatRoomHandler
}

func (m *Manager) routeEvent(event Event, c *Client) error {
	if handler, ok := m.handlers[event.Type]; ok {
		// handler is present in the map -> execute that EventHandler
		if err := handler(event, c); err != nil {
			return err
		}
		return nil
	} else {
		return ErrEventNotSupported
	}
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
		client.conn.Close()       // gracefully close the connection
		delete(m.clients, client) // and delete the reference to the connection from current list
	}
}
