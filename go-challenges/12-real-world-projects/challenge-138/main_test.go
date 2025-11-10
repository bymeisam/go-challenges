package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

func TestNewRoom(t *testing.T) {
	room := NewRoom("test")

	if room.name != "test" {
		t.Errorf("Expected room name 'test', got '%s'", room.name)
	}

	if room.clients == nil {
		t.Error("Clients map should be initialized")
	}

	if room.history == nil {
		t.Error("History should be initialized")
	}
}

func TestRoomAddRemoveClient(t *testing.T) {
	room := NewRoom("test")
	client := &Client{
		username: "testuser",
		rooms:    make(map[string]bool),
		send:     make(chan Message, 10),
	}

	room.AddClient(client)

	room.mu.RLock()
	if !room.clients[client] {
		t.Error("Client should be in room")
	}
	room.mu.RUnlock()

	room.RemoveClient(client)

	room.mu.RLock()
	if room.clients[client] {
		t.Error("Client should be removed from room")
	}
	room.mu.RUnlock()
}

func TestRoomBroadcast(t *testing.T) {
	room := NewRoom("test")

	client1 := &Client{
		username: "user1",
		rooms:    make(map[string]bool),
		send:     make(chan Message, 10),
	}
	client2 := &Client{
		username: "user2",
		rooms:    make(map[string]bool),
		send:     make(chan Message, 10),
	}

	room.AddClient(client1)
	room.AddClient(client2)

	msg := Message{
		Type:      MessageTypeMessage,
		Username:  "user1",
		Content:   "Hello",
		Room:      "test",
		Timestamp: time.Now(),
	}

	room.Broadcast(msg)

	// Check both clients received the message
	select {
	case received := <-client1.send:
		if received.Content != "Hello" {
			t.Errorf("Expected content 'Hello', got '%s'", received.Content)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Client1 did not receive message")
	}

	select {
	case received := <-client2.send:
		if received.Content != "Hello" {
			t.Errorf("Expected content 'Hello', got '%s'", received.Content)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Client2 did not receive message")
	}
}

func TestRoomHistory(t *testing.T) {
	room := NewRoom("test")

	for i := 0; i < 5; i++ {
		msg := Message{
			Type:      MessageTypeMessage,
			Content:   "Message " + string(rune('0'+i)),
			Timestamp: time.Now(),
		}
		room.addToHistory(msg)
	}

	history := room.GetHistory(3)
	if len(history) != 3 {
		t.Errorf("Expected 3 messages in history, got %d", len(history))
	}

	allHistory := room.GetHistory(10)
	if len(allHistory) != 5 {
		t.Errorf("Expected 5 messages in history, got %d", len(allHistory))
	}
}

func TestRoomHistoryLimit(t *testing.T) {
	room := NewRoom("test")

	// Add more than 100 messages
	for i := 0; i < 150; i++ {
		msg := Message{
			Type:      MessageTypeMessage,
			Content:   "Message",
			Timestamp: time.Now(),
		}
		room.addToHistory(msg)
	}

	history := room.GetHistory(200)
	if len(history) > 100 {
		t.Errorf("History should be limited to 100 messages, got %d", len(history))
	}
}

func TestRoomGetUsers(t *testing.T) {
	room := NewRoom("test")

	client1 := &Client{username: "alice", send: make(chan Message, 10)}
	client2 := &Client{username: "bob", send: make(chan Message, 10)}

	room.AddClient(client1)
	room.AddClient(client2)

	users := room.GetUsers()
	if len(users) != 2 {
		t.Errorf("Expected 2 users, got %d", len(users))
	}

	// Check that both users are in the list
	userMap := make(map[string]bool)
	for _, user := range users {
		userMap[user] = true
	}

	if !userMap["alice"] || !userMap["bob"] {
		t.Error("Not all users are in the list")
	}
}

func TestNewChatServer(t *testing.T) {
	server := NewChatServer()

	if server.rooms == nil {
		t.Error("Rooms map should be initialized")
	}

	if server.clients == nil {
		t.Error("Clients map should be initialized")
	}

	if server.register == nil {
		t.Error("Register channel should be initialized")
	}

	if server.unregister == nil {
		t.Error("Unregister channel should be initialized")
	}

	if server.broadcast == nil {
		t.Error("Broadcast channel should be initialized")
	}
}

func TestGetOrCreateRoom(t *testing.T) {
	server := NewChatServer()

	room1 := server.GetOrCreateRoom("general")
	if room1 == nil {
		t.Fatal("Room should be created")
	}

	if room1.name != "general" {
		t.Errorf("Expected room name 'general', got '%s'", room1.name)
	}

	// Get same room again
	room2 := server.GetOrCreateRoom("general")
	if room1 != room2 {
		t.Error("Should return same room instance")
	}

	// Create different room
	room3 := server.GetOrCreateRoom("random")
	if room1 == room3 {
		t.Error("Should create different room")
	}
}

func TestRegisterUnregisterClient(t *testing.T) {
	server := NewChatServer()
	go server.Run()

	client := &Client{
		username: "testuser",
		rooms:    make(map[string]bool),
		send:     make(chan Message, 10),
		server:   server,
	}

	// Register client
	server.register <- client
	time.Sleep(10 * time.Millisecond)

	server.mu.RLock()
	if !server.clients[client] {
		t.Error("Client should be registered")
	}
	server.mu.RUnlock()

	// Unregister client
	server.unregister <- client
	time.Sleep(10 * time.Millisecond)

	server.mu.RLock()
	if server.clients[client] {
		t.Error("Client should be unregistered")
	}
	server.mu.RUnlock()
}

func TestHandleJoinMessage(t *testing.T) {
	server := NewChatServer()
	go server.Run()

	client := &Client{
		username: "",
		rooms:    make(map[string]bool),
		send:     make(chan Message, 10),
		server:   server,
	}

	server.register <- client
	time.Sleep(10 * time.Millisecond)

	msg := Message{
		Type:     MessageTypeJoin,
		Username: "alice",
		Room:     "general",
	}

	client.handleJoin(msg)

	if client.username != "alice" {
		t.Errorf("Expected username 'alice', got '%s'", client.username)
	}

	if !client.rooms["general"] {
		t.Error("Client should be in 'general' room")
	}

	room := server.rooms["general"]
	if room == nil {
		t.Fatal("Room should be created")
	}

	room.mu.RLock()
	if !room.clients[client] {
		t.Error("Client should be in room's client list")
	}
	room.mu.RUnlock()
}

func TestHandleJoinWithoutUsername(t *testing.T) {
	client := &Client{
		send: make(chan Message, 10),
	}

	msg := Message{
		Type: MessageTypeJoin,
		Room: "general",
	}

	client.handleJoin(msg)

	// Should receive error
	select {
	case errMsg := <-client.send:
		if errMsg.Type != MessageTypeError {
			t.Error("Expected error message")
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Should receive error message")
	}
}

func TestHandleLeaveMessage(t *testing.T) {
	server := NewChatServer()
	go server.Run()

	client := &Client{
		username: "alice",
		rooms:    make(map[string]bool),
		send:     make(chan Message, 10),
		server:   server,
	}

	server.register <- client
	time.Sleep(10 * time.Millisecond)

	// Join room first
	joinMsg := Message{
		Type:     MessageTypeJoin,
		Username: "alice",
		Room:     "general",
	}
	client.handleJoin(joinMsg)

	// Leave room
	leaveMsg := Message{
		Type: MessageTypeLeave,
		Room: "general",
	}
	client.handleLeave(leaveMsg)

	if client.rooms["general"] {
		t.Error("Client should not be in room anymore")
	}
}

func TestHandleChatMessage(t *testing.T) {
	server := NewChatServer()
	go server.Run()

	client1 := &Client{
		username: "alice",
		rooms:    map[string]bool{"general": true},
		send:     make(chan Message, 10),
		server:   server,
	}

	client2 := &Client{
		username: "bob",
		rooms:    map[string]bool{"general": true},
		send:     make(chan Message, 10),
		server:   server,
	}

	server.register <- client1
	server.register <- client2
	time.Sleep(10 * time.Millisecond)

	room := server.GetOrCreateRoom("general")
	room.AddClient(client1)
	room.AddClient(client2)

	msg := Message{
		Type:    MessageTypeMessage,
		Content: "Hello everyone",
		Room:    "general",
	}

	client1.handleChatMessage(msg)
	time.Sleep(10 * time.Millisecond)

	// Both clients should receive the message
	select {
	case received := <-client1.send:
		if received.Content != "Hello everyone" {
			t.Error("Client1 should receive message")
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Client1 did not receive message")
	}

	select {
	case received := <-client2.send:
		if received.Content != "Hello everyone" {
			t.Error("Client2 should receive message")
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Client2 did not receive message")
	}
}

func TestHandlePrivateMessage(t *testing.T) {
	server := NewChatServer()
	go server.Run()

	client1 := &Client{
		username: "alice",
		rooms:    make(map[string]bool),
		send:     make(chan Message, 10),
		server:   server,
	}

	client2 := &Client{
		username: "bob",
		rooms:    make(map[string]bool),
		send:     make(chan Message, 10),
		server:   server,
	}

	server.register <- client1
	server.register <- client2
	time.Sleep(10 * time.Millisecond)

	msg := Message{
		Type:    MessageTypePrivate,
		To:      "bob",
		Content: "Private message",
	}

	client1.handlePrivateMessage(msg)

	// Bob should receive the message
	select {
	case received := <-client2.send:
		if received.Content != "Private message" {
			t.Error("Bob should receive private message")
		}
		if received.Username != "alice" {
			t.Error("Message should be from alice")
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Bob did not receive private message")
	}

	// Alice should receive echo
	select {
	case <-client1.send:
		// OK
	case <-time.After(100 * time.Millisecond):
		t.Error("Alice should receive message echo")
	}
}

func TestHandleUserListMessage(t *testing.T) {
	server := NewChatServer()
	go server.Run()

	client := &Client{
		username: "alice",
		rooms:    make(map[string]bool),
		send:     make(chan Message, 10),
		server:   server,
	}

	server.register <- client
	time.Sleep(10 * time.Millisecond)

	room := server.GetOrCreateRoom("general")
	room.AddClient(client)

	msg := Message{
		Type: MessageTypeUserList,
		Room: "general",
	}

	client.handleUserList(msg)

	select {
	case response := <-client.send:
		if response.Type != MessageTypeUserList {
			t.Error("Should receive user list")
		}
		if len(response.Users) != 1 {
			t.Errorf("Expected 1 user, got %d", len(response.Users))
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Did not receive user list")
	}
}

func TestHandleHistoryMessage(t *testing.T) {
	server := NewChatServer()
	go server.Run()

	client := &Client{
		username: "alice",
		rooms:    make(map[string]bool),
		send:     make(chan Message, 10),
		server:   server,
	}

	room := server.GetOrCreateRoom("general")

	// Add some history
	for i := 0; i < 3; i++ {
		msg := Message{
			Type:    MessageTypeMessage,
			Content: "History message",
		}
		room.addToHistory(msg)
	}

	msg := Message{
		Type: MessageTypeHistory,
		Room: "general",
	}

	client.handleHistory(msg)

	// Should receive 3 history messages
	count := 0
	timeout := time.After(100 * time.Millisecond)

loop:
	for {
		select {
		case <-client.send:
			count++
			if count == 3 {
				break loop
			}
		case <-timeout:
			break loop
		}
	}

	if count != 3 {
		t.Errorf("Expected 3 history messages, got %d", count)
	}
}

func TestWebSocketConnection(t *testing.T) {
	server := NewChatServer()
	go server.Run()

	testServer := httptest.NewServer(http.HandlerFunc(server.HandleWebSocket))
	defer testServer.Close()

	wsURL := "ws" + strings.TrimPrefix(testServer.URL, "http")

	ws, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer ws.Close()

	// Send join message
	joinMsg := Message{
		Type:     MessageTypeJoin,
		Username: "testuser",
		Room:     "general",
	}

	if err := ws.WriteJSON(joinMsg); err != nil {
		t.Fatalf("Failed to send message: %v", err)
	}

	// Receive join confirmation
	var response Message
	ws.SetReadDeadline(time.Now().Add(1 * time.Second))
	if err := ws.ReadJSON(&response); err != nil {
		t.Fatalf("Failed to receive message: %v", err)
	}

	if response.Type != MessageTypeJoin {
		t.Errorf("Expected join message, got %s", response.Type)
	}
}

func TestMessageSerialization(t *testing.T) {
	msg := Message{
		Type:      MessageTypeMessage,
		Username:  "alice",
		Content:   "Hello",
		Room:      "general",
		Timestamp: time.Now(),
	}

	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("Failed to marshal message: %v", err)
	}

	var decoded Message
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal message: %v", err)
	}

	if decoded.Type != msg.Type {
		t.Error("Type mismatch")
	}

	if decoded.Username != msg.Username {
		t.Error("Username mismatch")
	}

	if decoded.Content != msg.Content {
		t.Error("Content mismatch")
	}
}

func TestConcurrentRoomAccess(t *testing.T) {
	room := NewRoom("test")

	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(id int) {
			client := &Client{
				username: "user" + string(rune('0'+id)),
				send:     make(chan Message, 10),
			}
			room.AddClient(client)
			time.Sleep(10 * time.Millisecond)
			room.RemoveClient(client)
			done <- true
		}(i)
	}

	for i := 0; i < 10; i++ {
		<-done
	}

	// No panic = success
}
