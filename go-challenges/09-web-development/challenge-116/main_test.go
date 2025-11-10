package main

import (
	"bufio"
	"crypto/sha1"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestWebSocket(t *testing.T) {
	server := NewWebSocketServer()

	t.Run("HomeRoute", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		w := httptest.NewRecorder()

		server.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		body := w.Body.String()
		if !strings.Contains(body, "WebSocket Server") {
			t.Error("Expected WebSocket Server title")
		}
	})

	t.Run("WebSocketUpgradeWithoutHeader", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/ws", nil)
		w := httptest.NewRecorder()

		server.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", w.Code)
		}
	})

	t.Run("WebSocketUpgradeWithoutKey", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/ws", nil)
		req.Header.Set("Upgrade", "websocket")
		w := httptest.NewRecorder()

		server.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", w.Code)
		}
	})

	t.Run("CalculateAcceptKey", func(t *testing.T) {
		// Test with known values from RFC 6455
		key := "dGhlIHNhbXBsZSBub25jZQ=="
		expectedAccept := "s3pPLMBiTxaQ9kYGzzhZRbK+xOo="

		acceptKey := server.calculateAcceptKey(key)

		if acceptKey != expectedAccept {
			t.Errorf("Expected accept key '%s', got '%s'", expectedAccept, acceptKey)
		}
	})

	t.Log("✓ WebSocket server works!")
}

// Integration test with actual WebSocket connection
func TestWebSocketConnection(t *testing.T) {
	server := NewWebSocketServer()

	// Start test server
	testServer := httptest.NewServer(server)
	defer testServer.Close()

	// Replace http:// with ws://
	wsURL := "ws" + strings.TrimPrefix(testServer.URL, "http") + "/ws"

	t.Run("WebSocketHandshake", func(t *testing.T) {
		// Connect to server
		conn, err := net.Dial("tcp", strings.TrimPrefix(testServer.URL, "http://"))
		if err != nil {
			t.Fatalf("Failed to connect: %v", err)
		}
		defer conn.Close()

		// Generate WebSocket key
		key := generateWebSocketKey()

		// Send upgrade request
		request := fmt.Sprintf(
			"GET /ws HTTP/1.1\r\n"+
				"Host: %s\r\n"+
				"Upgrade: websocket\r\n"+
				"Connection: Upgrade\r\n"+
				"Sec-WebSocket-Key: %s\r\n"+
				"Sec-WebSocket-Version: 13\r\n\r\n",
			testServer.URL,
			key,
		)

		if _, err := conn.Write([]byte(request)); err != nil {
			t.Fatalf("Failed to send upgrade request: %v", err)
		}

		// Read response
		reader := bufio.NewReader(conn)
		response, err := reader.ReadString('\n')
		if err != nil {
			t.Fatalf("Failed to read response: %v", err)
		}

		if !strings.Contains(response, "101 Switching Protocols") {
			t.Errorf("Expected 101 Switching Protocols, got: %s", response)
		}

		// Read headers
		for {
			line, _ := reader.ReadString('\n')
			if line == "\r\n" {
				break
			}
			if strings.HasPrefix(line, "Sec-WebSocket-Accept:") {
				acceptKey := strings.TrimSpace(strings.TrimPrefix(line, "Sec-WebSocket-Accept:"))
				expectedAccept := calculateExpectedAccept(key)
				if acceptKey != expectedAccept {
					t.Errorf("Expected accept key '%s', got '%s'", expectedAccept, acceptKey)
				}
			}
		}
	})

	t.Run("ClientCount", func(t *testing.T) {
		initialCount := server.GetClientCount()
		if initialCount < 0 {
			t.Error("Client count should be non-negative")
		}
	})

	t.Run("BroadcastMessage", func(t *testing.T) {
		msg := Message{
			Type:    "test",
			Content: "Test broadcast",
		}

		// This should not panic
		server.BroadcastMessage(msg)
	})

	t.Log("✓ WebSocket connection works!")
}

func TestHub(t *testing.T) {
	t.Run("HubCreation", func(t *testing.T) {
		hub := NewHub()
		if hub == nil {
			t.Fatal("Hub should not be nil")
		}

		if hub.clients == nil {
			t.Error("Hub clients map should be initialized")
		}
	})

	t.Run("HubChannels", func(t *testing.T) {
		hub := NewHub()

		if hub.broadcast == nil {
			t.Error("Broadcast channel should be initialized")
		}
		if hub.register == nil {
			t.Error("Register channel should be initialized")
		}
		if hub.unregister == nil {
			t.Error("Unregister channel should be initialized")
		}
	})

	t.Log("✓ Hub works!")
}

func TestMessage(t *testing.T) {
	t.Run("MessageSerialization", func(t *testing.T) {
		msg := Message{
			Type:    "chat",
			Content: "Hello, World!",
			From:    "user123",
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
			t.Errorf("Expected type '%s', got '%s'", msg.Type, decoded.Type)
		}
		if decoded.Content != msg.Content {
			t.Errorf("Expected content '%s', got '%s'", msg.Content, decoded.Content)
		}
		if decoded.From != msg.From {
			t.Errorf("Expected from '%s', got '%s'", msg.From, decoded.From)
		}
	})

	t.Log("✓ Message serialization works!")
}

// Helper functions for testing

func generateWebSocketKey() string {
	return base64.StdEncoding.EncodeToString([]byte("test-websocket-key12"))
}

func calculateExpectedAccept(key string) string {
	hash := sha1.New()
	hash.Write([]byte(key + websocketMagicString))
	return base64.StdEncoding.EncodeToString(hash.Sum(nil))
}

func createWebSocketFrame(text string, masked bool) []byte {
	payload := []byte(text)
	payloadLen := len(payload)

	var frame []byte
	frame = append(frame, 0x81) // FIN + Text frame

	maskBit := byte(0)
	if masked {
		maskBit = 0x80
	}

	if payloadLen <= 125 {
		frame = append(frame, byte(payloadLen)|maskBit)
	} else if payloadLen <= 65535 {
		frame = append(frame, 126|maskBit)
		lenBytes := make([]byte, 2)
		binary.BigEndian.PutUint16(lenBytes, uint16(payloadLen))
		frame = append(frame, lenBytes...)
	}

	if masked {
		maskingKey := []byte{0x12, 0x34, 0x56, 0x78}
		frame = append(frame, maskingKey...)

		maskedPayload := make([]byte, payloadLen)
		for i := 0; i < payloadLen; i++ {
			maskedPayload[i] = payload[i] ^ maskingKey[i%4]
		}
		frame = append(frame, maskedPayload...)
	} else {
		frame = append(frame, payload...)
	}

	return frame
}
