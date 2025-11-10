package main

import (
	"crypto/sha1"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
)

const websocketMagicString = "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"

type Message struct {
	Type    string `json:"type"`
	Content string `json:"content"`
	From    string `json:"from,omitempty"`
}

type Client struct {
	conn   io.ReadWriteCloser
	id     string
	hub    *Hub
}

type Hub struct {
	clients    map[*Client]bool
	broadcast  chan Message
	register   chan *Client
	unregister chan *Client
	mu         sync.RWMutex
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan Message),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
			}
			h.mu.Unlock()

		case message := <-h.broadcast:
			h.mu.RLock()
			for client := range h.clients {
				go func(c *Client) {
					if err := c.sendMessage(message); err != nil {
						h.unregister <- c
					}
				}(client)
			}
			h.mu.RUnlock()
		}
	}
}

func (c *Client) sendMessage(msg Message) error {
	data, _ := json.Marshal(msg)
	return c.sendTextFrame(string(data))
}

func (c *Client) sendTextFrame(text string) error {
	payload := []byte(text)
	payloadLen := len(payload)

	var frame []byte
	frame = append(frame, 0x81) // FIN + Text frame

	if payloadLen <= 125 {
		frame = append(frame, byte(payloadLen))
	} else if payloadLen <= 65535 {
		frame = append(frame, 126)
		lenBytes := make([]byte, 2)
		binary.BigEndian.PutUint16(lenBytes, uint16(payloadLen))
		frame = append(frame, lenBytes...)
	} else {
		frame = append(frame, 127)
		lenBytes := make([]byte, 8)
		binary.BigEndian.PutUint64(lenBytes, uint64(payloadLen))
		frame = append(frame, lenBytes...)
	}

	frame = append(frame, payload...)

	_, err := c.conn.Write(frame)
	return err
}

func (c *Client) readMessage() (string, error) {
	// Read first 2 bytes
	header := make([]byte, 2)
	if _, err := io.ReadFull(c.conn, header); err != nil {
		return "", err
	}

	// Check if message is masked (client to server messages must be masked)
	masked := (header[1] & 0x80) != 0
	payloadLen := int(header[1] & 0x7F)

	// Read extended payload length if needed
	if payloadLen == 126 {
		extLen := make([]byte, 2)
		if _, err := io.ReadFull(c.conn, extLen); err != nil {
			return "", err
		}
		payloadLen = int(binary.BigEndian.Uint16(extLen))
	} else if payloadLen == 127 {
		extLen := make([]byte, 8)
		if _, err := io.ReadFull(c.conn, extLen); err != nil {
			return "", err
		}
		payloadLen = int(binary.BigEndian.Uint64(extLen))
	}

	// Read masking key if present
	var maskingKey []byte
	if masked {
		maskingKey = make([]byte, 4)
		if _, err := io.ReadFull(c.conn, maskingKey); err != nil {
			return "", err
		}
	}

	// Read payload
	payload := make([]byte, payloadLen)
	if _, err := io.ReadFull(c.conn, payload); err != nil {
		return "", err
	}

	// Unmask payload if needed
	if masked {
		for i := 0; i < payloadLen; i++ {
			payload[i] ^= maskingKey[i%4]
		}
	}

	return string(payload), nil
}

func (c *Client) handleMessages() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	for {
		text, err := c.readMessage()
		if err != nil {
			break
		}

		var msg Message
		if err := json.Unmarshal([]byte(text), &msg); err == nil {
			msg.From = c.id
			c.hub.broadcast <- msg
		}
	}
}

type WebSocketServer struct {
	mux *http.ServeMux
	hub *Hub
}

func NewWebSocketServer() *WebSocketServer {
	hub := NewHub()
	go hub.Run()

	server := &WebSocketServer{
		mux: http.NewServeMux(),
		hub: hub,
	}
	server.routes()
	return server
}

func (s *WebSocketServer) routes() {
	s.mux.HandleFunc("/ws", s.handleWebSocket())
	s.mux.HandleFunc("/", s.handleHome())
}

func (s *WebSocketServer) handleHome() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprintf(w, `<!DOCTYPE html>
<html>
<head><title>WebSocket Server</title></head>
<body>
<h1>WebSocket Server</h1>
<p>Connect to ws://localhost/ws</p>
</body>
</html>`)
	}
}

func (s *WebSocketServer) handleWebSocket() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Check if upgrade header is present
		if r.Header.Get("Upgrade") != "websocket" {
			http.Error(w, "Not a websocket upgrade", http.StatusBadRequest)
			return
		}

		// Get WebSocket key
		key := r.Header.Get("Sec-WebSocket-Key")
		if key == "" {
			http.Error(w, "Missing Sec-WebSocket-Key", http.StatusBadRequest)
			return
		}

		// Calculate accept key
		acceptKey := s.calculateAcceptKey(key)

		// Hijack the connection
		hijacker, ok := w.(http.Hijacker)
		if !ok {
			http.Error(w, "Hijacking not supported", http.StatusInternalServerError)
			return
		}

		conn, bufrw, err := hijacker.Hijack()
		if err != nil {
			http.Error(w, "Failed to hijack connection", http.StatusInternalServerError)
			return
		}

		// Send upgrade response
		response := fmt.Sprintf(
			"HTTP/1.1 101 Switching Protocols\r\n"+
				"Upgrade: websocket\r\n"+
				"Connection: Upgrade\r\n"+
				"Sec-WebSocket-Accept: %s\r\n\r\n",
			acceptKey,
		)

		if _, err := bufrw.WriteString(response); err != nil {
			conn.Close()
			return
		}
		bufrw.Flush()

		// Create client
		client := &Client{
			conn: conn,
			id:   fmt.Sprintf("client-%d", len(s.hub.clients)+1),
			hub:  s.hub,
		}

		// Register client
		s.hub.register <- client

		// Send welcome message
		welcomeMsg := Message{
			Type:    "welcome",
			Content: fmt.Sprintf("Welcome! Your ID is %s", client.id),
		}
		client.sendMessage(welcomeMsg)

		// Handle incoming messages
		client.handleMessages()
	}
}

func (s *WebSocketServer) calculateAcceptKey(key string) string {
	hash := sha1.New()
	hash.Write([]byte(key + websocketMagicString))
	return base64.StdEncoding.EncodeToString(hash.Sum(nil))
}

func (s *WebSocketServer) BroadcastMessage(msg Message) {
	s.hub.broadcast <- msg
}

func (s *WebSocketServer) GetClientCount() int {
	s.hub.mu.RLock()
	defer s.hub.mu.RUnlock()
	return len(s.hub.clients)
}

func (s *WebSocketServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

func main() {}
