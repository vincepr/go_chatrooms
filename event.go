package main

import "encoding/json"

// Corresponds to Messages sent over the Websockets
// used to differentiate kinds of Messages
type Event struct {
	Type string `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

// Signature to easily extend Different Event-Types later
type EventHandler func(event Event, c *Client) error

const(
	// EventSendMessage is the event name for new chat messages sent
	EventSendMessage = "send_message"
)

type SendMessageEvent struct {
	Message string `json:"message"`
	From	string `json:"from"`
}