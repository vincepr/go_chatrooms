package main

import (
	"encoding/json"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

var (
	// pongWait is how long we wait, before assuming the client is dead and cleanup
	// pingInterval is when we send the next ping, that our client has to answer
	pongWait     = 10 * time.Second
	pingInterval = (pongWait * 9) / 10
	// in bytes. Maximum size of one Message.
	msgMaxSize int64 = 512
)

// each Client gets a Client struct once upgraded to a Websocket Connection (from HTTP)
type Client struct {
	conn    *websocket.Conn // the websocket connection
	manager *Manager        // reference to the manager that handles it
	egress  chan Event      // to avoid concurrent writes on the Websocket we use this since it blocks
}

func NewClient(conn *websocket.Conn, manager *Manager) *Client {
	return &Client{
		conn:    conn,
		manager: manager,
		egress:  make(chan Event),
	}
}

// Goroutine (1 for each client) that handles reading Messages coming in from Clients
func (c *Client) readMessages() {
	// cleanup when finished/disconected; remove itself from client-pool:
	defer func() {
		c.manager.removeClient(c)
	}()

	// Limit connection Size of a single message:
	c.conn.SetReadLimit(msgMaxSize)

	// Configure max-wait-time for Pong response. use current_time+pongWait
	if err := c.conn.SetReadDeadline(time.Now().Add(pongWait)); err != nil {
		log.Println("could not set timeout:", err)
		return
	}
	// handle the Pong received:
	c.conn.SetPongHandler(c.pongHandler)

	for {
		// read the next message in the queue for the connection:
		_, payload, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				// Any Errors (that are not simple disconects) we log out:
				log.Println("Connection Closed Unexpected: ", err)
			}
			break
		}
		// Marshal incoming JSON data into the Event struct
		var request Event
		if err := json.Unmarshal(payload, &request); err != nil {
			log.Printf("Error json.Unmarshal: %v", err)
			break // TODO: maybe handle this gracefully? request sending again etc...
		}
		// Route the Event
		if err := c.manager.routeEvent(request, c); err != nil {
			log.Println("Error routeEvent:", err)
		}
	}
}

// Goroutine (1 for each client) that listens for new messages and sends them to the client
func (c *Client) writeMessages() {
	// create a ticker that triggers the ping signal to check if connection is alive
	ticker := time.NewTicker(pingInterval)

	// gracefully close and cleanup
	defer func() {
		ticker.Stop()
		c.manager.removeClient(c)
	}()

	for {
		log.Println("12")
		select {
		// handle next message in queue(channel egress)
		case msg, ok := <-c.egress:
			if !ok {
				if err := c.conn.WriteMessage(websocket.CloseMessage, nil); err != nil {
					log.Println("Connection closed because of:", err)
				}
				return // close this goroutine because client sent close-signal
			}
			data, err := json.Marshal(msg)
			if err != nil {
				log.Println("Couldnt json.Marshal:", err)
			}

			// write regular text msg to connection
			if err := c.conn.WriteMessage(websocket.TextMessage, data); err != nil {
				log.Println("Failed Writing to Channel:", err)
			}
			log.Println("dbg: sent message sucessfully")
		// time send next ping, checking if connection is still alive
		case <-ticker.C:
			log.Println("sending ping")
			if err := c.conn.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
				log.Println("no pong recieved in time: ", err)
				return // got no pong back in time -> we close
			}
		}
	}
}

// handle the received Pong Message Type from the client
func (c *Client) pongHandler(pongMsg string) error {
	log.Println("pong")
	// setup next Ping:
	return c.conn.SetReadDeadline(time.Now().Add(pongWait))
}
