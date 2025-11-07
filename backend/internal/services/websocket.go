package services

import (
	"encoding/json"
	"log"
	"sync"

	"github.com/gofiber/websocket/v2"
)

type WebSocketHub struct {
	clients    map[*websocket.Conn]bool
	broadcast  chan []byte
	register   chan *websocket.Conn
	unregister chan *websocket.Conn
	mutex      sync.RWMutex
}

type WSMessage struct {
	Type     string      `json:"type"`
	PolicyID string      `json:"policy_id,omitempty"`
	Data     interface{} `json:"data,omitempty"`
}

func NewWebSocketHub() *WebSocketHub {
	return &WebSocketHub{
		clients:    make(map[*websocket.Conn]bool),
		broadcast:  make(chan []byte),
		register:   make(chan *websocket.Conn),
		unregister: make(chan *websocket.Conn),
	}
}

func (h *WebSocketHub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mutex.Lock()
			h.clients[client] = true
			h.mutex.Unlock()

		case client := <-h.unregister:
			h.mutex.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				client.Close()
			}
			h.mutex.Unlock()

		case message := <-h.broadcast:
			h.mutex.RLock()
			for client := range h.clients {
				err := client.WriteMessage(websocket.TextMessage, message)
				if err != nil {
					client.Close()
					delete(h.clients, client)
				}
			}
			h.mutex.RUnlock()
		}
	}
}

func (h *WebSocketHub) HandleConnection(c *websocket.Conn) {
	h.register <- c
	defer func() {
		h.unregister <- c
	}()

	for {
		_, _, err := c.ReadMessage()
		if err != nil {
			break
		}
	}
}

func (h *WebSocketHub) BroadcastVoteUpdate(policyID string, upvotes, downvotes int) {
	msg := WSMessage{
		Type:     "vote_update",
		PolicyID: policyID,
		Data: map[string]interface{}{
			"upvotes":   upvotes,
			"downvotes": downvotes,
		},
	}

	data, err := json.Marshal(msg)
	if err != nil {
		log.Println("Failed to marshal websocket message:", err)
		return
	}

	h.broadcast <- data
}

func (h *WebSocketHub) BroadcastPolicyUpdate(policyID, status string) {
	msg := WSMessage{
		Type:     "policy_update",
		PolicyID: policyID,
		Data: map[string]interface{}{
			"status": status,
		},
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return
	}

	h.broadcast <- data
}
