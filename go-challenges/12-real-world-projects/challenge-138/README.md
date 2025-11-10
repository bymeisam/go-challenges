# Challenge 138: Chat Server

**Difficulty:** ⭐⭐⭐⭐ Hard
**Time Estimate:** 55 minutes

## Description

Build a real-time chat server using WebSockets with support for multiple rooms, private messages, and user presence. This demonstrates WebSocket handling, broadcast patterns, and stateful connections.

## Features

- **WebSocket Connections**: Real-time bidirectional communication
- **Multiple Rooms**: Users can join different chat rooms
- **Private Messages**: Direct messages between users
- **User Presence**: Track online/offline status
- **Message History**: Store recent messages
- **User Authentication**: Simple username-based auth
- **Broadcast Messages**: Send to all users in a room
- **Typing Indicators**: Show when users are typing
- **Message Persistence**: Optional message storage

## Message Types

- **Join**: User joins a room
- **Leave**: User leaves a room
- **Message**: Send message to room
- **Private**: Send private message
- **Typing**: Typing indicator
- **UserList**: Request user list
- **History**: Request message history

## Requirements

1. Use gorilla/websocket for WebSocket handling
2. Implement pub/sub pattern for message distribution
3. Track user connections and rooms
4. Handle connection failures gracefully
5. Support concurrent connections
6. Message validation and sanitization
7. Rate limiting to prevent spam

## Example Usage

```javascript
// Client-side JavaScript
const ws = new WebSocket('ws://localhost:8080/ws');

ws.send(JSON.stringify({
    type: 'join',
    username: 'alice',
    room: 'general'
}));

ws.send(JSON.stringify({
    type: 'message',
    content: 'Hello everyone!',
    room: 'general'
}));
```

## Learning Objectives

- WebSocket protocol and handling
- Concurrent connection management
- Pub/sub messaging patterns
- State management for connections
- Graceful connection handling
- Real-time communication patterns
- Message broadcasting
- WebSocket security considerations
