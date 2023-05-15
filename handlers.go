package main

import (
	"encoding/json"
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

// the WebSocket HandleFunc - websockets traffic at "/ws"
func (m *Manager) serveWS(w http.ResponseWriter, r *http.Request) {
	// Get the OTP from the request and try to verify it
	otp := r.URL.Query().Get("otp")
	if otp == "" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	if !m.otps.VertifyOTP(otp) {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// Upgrade the HTTP request to a continous Websocket
	log.Println("New connection upgrade request")
	conn, err := websocketUpgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Failed WF Upgrade:", err)
		return
	}

	// Add the connection the the pool and setup the Client goroutines/workers
	client := NewClient(conn, m)
	m.addClient(client)
	go client.readMessages()
	go client.writeMessages()
}

// request to auth sent from client user
type userLoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// response of successful authentification sent to user
type response struct {
	OTP string `json:"otp"`
}

// the Login HandleFunc - verify a login request at "/login"
func (m *Manager) loginHandler(w http.ResponseWriter, r *http.Request) {
	// parse incoming JSON:
	var req userLoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	// check access token
	if req.Username == "bob" && req.Password == "123" {
		// generate new OTP for this user:
		otp := m.otps.NewOTP()
		resp := response{
			OTP: otp.Key,
		}
		data, err := json.Marshal(resp)
		if err != nil {
			log.Println(err)
			return
		}
		// return response to successful authenticated user:
		w.WriteHeader(http.StatusOK)
		w.Write(data)
		return
	}
	// Failure to auth
	w.WriteHeader(http.StatusUnauthorized)
}
