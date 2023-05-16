/*		Here we define all Types our Socket-Api Supports
*		(ex. Messages, Broadcasts, ChangeRoom)
 */

package main

import (
	"encoding/json"
	"fmt"
	"time"
)

// Corresponds to Messages sent over the Websockets
// used to differentiate kinds of Messages. Acts as a Wrapper for the Event Types.
type Event struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

// Signature to easily extend Different Event-Types later
type EventHandler func(event Event, c *Client) error

const (
	// all different types of messages our system supports get an identifier:

	EventSendMessage = "send_message"
	//EventNewMessage = "new_message"		// server should never receive this, he only sends it!
	EventChangeRoom = "change_room"
)

type SendMessageEvent struct {
	Message string `json:"message"`
	From    string `json:"from"`
}

// Logic for the distribution and forwarding of messages:
type NewMessageEvent struct {
	SendMessageEvent
	Sent time.Time `json:"sent"` // we add a timestamp on the server
}

type ChangeRoomEvent struct {
	Name string `json:"name"`
}

// Logic for the distribution and forwarding of messages (forwards to all sockets in the same chat-ROOM)
func SendMessageHandler(event Event, c *Client) error {
	// marshal payload into target format
	var chatevent SendMessageEvent
	if err := json.Unmarshal(event.Payload, &chatevent); err != nil {
		return fmt.Errorf("bad payload in request: %v", err)
	}

	// Build Outgoing Message to other parcipiants in the same room:
	var broadcastMsg NewMessageEvent
	broadcastMsg.Sent = time.Now()
	broadcastMsg.Message = chatevent.Message
	broadcastMsg.From = chatevent.From
	data, err := json.Marshal(broadcastMsg)
	if err != nil {
		return fmt.Errorf("failed to marshal broadcast: %v", err)
	}

	// Build payload:
	var eventInQueue Event
	eventInQueue.Payload = data
	eventInQueue.Type = EventSendMessage

	// Broadcast payload to other Clients (check if in same chatroom)
	for client := range c.manager.clients {
		if client.chatroom == c.chatroom {
			client.egress <- eventInQueue
		}
	}
	return nil
}

func ChatRoomHandler(event Event, c *Client) error {
	// marshal payload into target format
	var roomEvent ChangeRoomEvent
	if err := json.Unmarshal(event.Payload, &roomEvent); err != nil {
		return fmt.Errorf("bad payload in roomChange request: %v", err)
	}
	c.chatroom = roomEvent.Name
	return nil
}
