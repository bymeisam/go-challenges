package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// Message types
const (
	MessageTypeJoin     = "join"
	MessageTypeLeave    = "leave"
	MessageTypeMessage  = "message"
	MessageTypePrivate  = "private"
	MessageTypeTyping   = "typing"
	MessageTypeUserList = "userlist"
	MessageTypeHistory  = "history"
	MessageTypeError    = "error"
)

// Message represents a chat message
type Message struct {
	Type      string    `json:"type"`
	Username  string    `json:"username"`
	Content   string    `json:"content"`
	Room      string    `json:"room"`
	To        string    `json:"to,omitempty"`
	Timestamp time.Time `json:"timestamp"`
	Users     []string  `json:"users,omitempty"`
}

// Client represents a connected user
type Client struct {
	conn     *websocket.Conn
	username string
	rooms    map[string]bool
	send     chan Message
	server   *ChatServer
}

// Room represents a chat room
type Room struct {
	name     string
	clients  map[*Client]bool
	history  []Message
	mu       sync.RWMutex
}

// NewRoom creates a new room
func NewRoom(name string) *Room {
	return &Room{
		name:    name,
		clients: make(map[*Client]bool),
		history: make([]Message, 0, 100),
	}
}

// AddClient adds a client to the room
func (r *Room) AddClient(client *Client) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.clients[client] = true
}

// RemoveClient removes a client from the room
func (r *Room) RemoveClient(client *Client) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.clients, client)
}

// Broadcast sends a message to all clients in the room
func (r *Room) Broadcast(message Message) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Add to history
	if message.Type == MessageTypeMessage {
		r.addToHistory(message)
	}

	for client := range r.clients {
		select {
		case client.send <- message:
		default:
			// Client's send channel is full, skip
		}
	}
}

// addToHistory adds a message to room history
func (r *Room) addToHistory(message Message) {
	if len(r.history) >= 100 {
		r.history = r.history[1:]
	}
	r.history = append(r.history, message)
}

// GetHistory returns recent messages
func (r *Room) GetHistory(limit int) []Message {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if limit <= 0 || limit > len(r.history) {
		limit = len(r.history)
	}

	start := len(r.history) - limit
	return r.history[start:]
}

// GetUsers returns list of usernames in the room
func (r *Room) GetUsers() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	users := make([]string, 0, len(r.clients))
	for client := range r.clients {
		users = append(users, client.username)
	}
	return users
}

// ChatServer manages the chat system
type ChatServer struct {
	rooms      map[string]*Room
	clients    map[*Client]bool
	register   chan *Client
	unregister chan *Client
	broadcast  chan Message
	mu         sync.RWMutex
	upgrader   websocket.Upgrader
}

// NewChatServer creates a new chat server
func NewChatServer() *ChatServer {
	return &ChatServer{
		rooms:      make(map[string]*Room),
		clients:    make(map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan Message, 256),
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				return true // Allow all origins (configure properly in production)
			},
		},
	}
}

// Run starts the chat server
func (s *ChatServer) Run() {
	for {
		select {
		case client := <-s.register:
			s.registerClient(client)
		case client := <-s.unregister:
			s.unregisterClient(client)
		case message := <-s.broadcast:
			s.handleBroadcast(message)
		}
	}
}

// registerClient registers a new client
func (s *ChatServer) registerClient(client *Client) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.clients[client] = true
}

// unregisterClient unregisters a client
func (s *ChatServer) unregisterClient(client *Client) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.clients[client]; ok {
		delete(s.clients, client)
		close(client.send)

		// Remove from all rooms
		for roomName := range client.rooms {
			if room, exists := s.rooms[roomName]; exists {
				room.RemoveClient(client)

				// Notify room about user leaving
				leaveMsg := Message{
					Type:      MessageTypeLeave,
					Username:  client.username,
					Room:      roomName,
					Timestamp: time.Now(),
				}
				room.Broadcast(leaveMsg)
			}
		}
	}
}

// handleBroadcast handles message broadcasting
func (s *ChatServer) handleBroadcast(message Message) {
	s.mu.RLock()
	room, exists := s.rooms[message.Room]
	s.mu.RUnlock()

	if exists {
		room.Broadcast(message)
	}
}

// GetOrCreateRoom gets or creates a room
func (s *ChatServer) GetOrCreateRoom(name string) *Room {
	s.mu.Lock()
	defer s.mu.Unlock()

	room, exists := s.rooms[name]
	if !exists {
		room = NewRoom(name)
		s.rooms[name] = room
	}
	return room
}

// HandleWebSocket handles WebSocket connections
func (s *ChatServer) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}

	client := &Client{
		conn:   conn,
		rooms:  make(map[string]bool),
		send:   make(chan Message, 256),
		server: s,
	}

	s.register <- client

	go client.writePump()
	go client.readPump()
}

// readPump reads messages from the WebSocket
func (c *Client) readPump() {
	defer func() {
		c.server.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		var msg Message
		err := c.conn.ReadJSON(&msg)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		msg.Timestamp = time.Now()
		c.handleMessage(msg)
	}
}

// writePump writes messages to the WebSocket
func (c *Client) writePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.conn.WriteJSON(message); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// handleMessage handles incoming messages
func (c *Client) handleMessage(msg Message) {
	switch msg.Type {
	case MessageTypeJoin:
		c.handleJoin(msg)
	case MessageTypeLeave:
		c.handleLeave(msg)
	case MessageTypeMessage:
		c.handleChatMessage(msg)
	case MessageTypePrivate:
		c.handlePrivateMessage(msg)
	case MessageTypeTyping:
		c.handleTyping(msg)
	case MessageTypeUserList:
		c.handleUserList(msg)
	case MessageTypeHistory:
		c.handleHistory(msg)
	default:
		c.sendError("Unknown message type")
	}
}

// handleJoin handles join room requests
func (c *Client) handleJoin(msg Message) {
	if msg.Username == "" {
		c.sendError("Username required")
		return
	}

	if msg.Room == "" {
		c.sendError("Room name required")
		return
	}

	c.username = msg.Username
	c.rooms[msg.Room] = true

	room := c.server.GetOrCreateRoom(msg.Room)
	room.AddClient(c)

	// Send join confirmation
	joinMsg := Message{
		Type:      MessageTypeJoin,
		Username:  c.username,
		Room:      msg.Room,
		Timestamp: time.Now(),
	}
	room.Broadcast(joinMsg)

	// Send recent history
	history := room.GetHistory(20)
	for _, histMsg := range history {
		c.send <- histMsg
	}
}

// handleLeave handles leave room requests
func (c *Client) handleLeave(msg Message) {
	if room, exists := c.server.rooms[msg.Room]; exists {
		room.RemoveClient(c)
		delete(c.rooms, msg.Room)

		leaveMsg := Message{
			Type:      MessageTypeLeave,
			Username:  c.username,
			Room:      msg.Room,
			Timestamp: time.Now(),
		}
		room.Broadcast(leaveMsg)
	}
}

// handleChatMessage handles chat messages
func (c *Client) handleChatMessage(msg Message) {
	if msg.Content == "" {
		return
	}

	if !c.rooms[msg.Room] {
		c.sendError("Not in room: " + msg.Room)
		return
	}

	msg.Username = c.username
	msg.Type = MessageTypeMessage
	msg.Timestamp = time.Now()

	c.server.broadcast <- msg
}

// handlePrivateMessage handles private messages
func (c *Client) handlePrivateMessage(msg Message) {
	if msg.To == "" || msg.Content == "" {
		c.sendError("Recipient and content required")
		return
	}

	msg.Username = c.username
	msg.Timestamp = time.Now()

	// Find recipient
	c.server.mu.RLock()
	defer c.server.mu.RUnlock()

	for client := range c.server.clients {
		if client.username == msg.To {
			client.send <- msg
			c.send <- msg // Echo back to sender
			return
		}
	}

	c.sendError("User not found: " + msg.To)
}

// handleTyping handles typing indicators
func (c *Client) handleTyping(msg Message) {
	msg.Username = c.username
	msg.Timestamp = time.Now()
	c.server.broadcast <- msg
}

// handleUserList handles user list requests
func (c *Client) handleUserList(msg Message) {
	if room, exists := c.server.rooms[msg.Room]; exists {
		users := room.GetUsers()
		response := Message{
			Type:      MessageTypeUserList,
			Room:      msg.Room,
			Users:     users,
			Timestamp: time.Now(),
		}
		c.send <- response
	}
}

// handleHistory handles history requests
func (c *Client) handleHistory(msg Message) {
	if room, exists := c.server.rooms[msg.Room]; exists {
		history := room.GetHistory(50)
		for _, histMsg := range history {
			c.send <- histMsg
		}
	}
}

// sendError sends an error message to the client
func (c *Client) sendError(errMsg string) {
	msg := Message{
		Type:      MessageTypeError,
		Content:   errMsg,
		Timestamp: time.Now(),
	}
	c.send <- msg
}

func main() {
	server := NewChatServer()
	go server.Run()

	http.HandleFunc("/ws", server.HandleWebSocket)

	// Serve a simple HTML client
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		html := `
<!DOCTYPE html>
<html>
<head>
    <title>Chat Server</title>
    <style>
        body { font-family: Arial, sans-serif; max-width: 800px; margin: 50px auto; }
        #messages { border: 1px solid #ccc; height: 400px; overflow-y: scroll; padding: 10px; margin-bottom: 10px; }
        .message { margin: 5px 0; }
        .join { color: green; }
        .leave { color: red; }
        input, button { padding: 10px; margin: 5px; }
    </style>
</head>
<body>
    <h1>WebSocket Chat</h1>
    <div id="messages"></div>
    <input id="username" placeholder="Username" />
    <input id="room" placeholder="Room" value="general" />
    <button onclick="join()">Join</button>
    <br>
    <input id="message" placeholder="Message" />
    <button onclick="send()">Send</button>

    <script>
        let ws = null;

        function connect() {
            ws = new WebSocket('ws://' + window.location.host + '/ws');
            ws.onmessage = function(event) {
                const msg = JSON.parse(event.data);
                displayMessage(msg);
            };
        }

        function join() {
            const username = document.getElementById('username').value;
            const room = document.getElementById('room').value;
            if (!ws) connect();
            ws.send(JSON.stringify({type: 'join', username: username, room: room}));
        }

        function send() {
            const content = document.getElementById('message').value;
            const room = document.getElementById('room').value;
            ws.send(JSON.stringify({type: 'message', content: content, room: room}));
            document.getElementById('message').value = '';
        }

        function displayMessage(msg) {
            const div = document.createElement('div');
            div.className = 'message ' + msg.type;
            div.textContent = '[' + msg.type + '] ' + (msg.username || '') + ': ' + (msg.content || '');
            document.getElementById('messages').appendChild(div);
            document.getElementById('messages').scrollTop = document.getElementById('messages').scrollHeight;
        }

        connect();
    </script>
</body>
</html>
`
		fmt.Fprint(w, html)
	})

	log.Println("Chat server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
