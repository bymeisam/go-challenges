# Challenge 142: gRPC Service

**Difficulty:** ⭐⭐⭐⭐ Hard
**Time Estimate:** 60 minutes

## Description

Build a complete gRPC service for a user management system with protobuf definitions, server implementation, and client. This project demonstrates gRPC service design, unary and streaming RPCs, interceptors, error handling, and metadata usage.

## Features

- **Protocol Buffers**: Define service and messages using proto3
- **Unary RPCs**: Simple request-response operations
- **Server Streaming**: Stream multiple responses to client
- **Client Streaming**: Stream multiple requests from client
- **Bidirectional Streaming**: Full-duplex communication
- **Interceptors**: Server-side middleware for logging and auth
- **Error Handling**: Proper gRPC status codes and error details
- **Metadata**: Request/response metadata handling
- **TLS Support**: Secure communication (optional)
- **Health Checking**: gRPC health check protocol
- **Reflection**: gRPC server reflection for debugging

## Protobuf Definition

```protobuf
syntax = "proto3";

package user;

option go_package = "github.com/example/userservice/proto";

// User service definition
service UserService {
  // Unary RPC: Create a new user
  rpc CreateUser(CreateUserRequest) returns (User);

  // Unary RPC: Get a user by ID
  rpc GetUser(GetUserRequest) returns (User);

  // Server streaming: List users with filters
  rpc ListUsers(ListUsersRequest) returns (stream User);

  // Client streaming: Batch create users
  rpc BatchCreateUsers(stream CreateUserRequest) returns (BatchCreateResponse);

  // Bidirectional streaming: Chat/real-time updates
  rpc StreamUserUpdates(stream UserUpdateRequest) returns (stream UserUpdateResponse);

  // Delete user
  rpc DeleteUser(DeleteUserRequest) returns (DeleteUserResponse);
}

message User {
  string id = 1;
  string username = 2;
  string email = 3;
  string full_name = 4;
  Role role = 5;
  int64 created_at = 6;
  int64 updated_at = 7;
}

enum Role {
  ROLE_UNSPECIFIED = 0;
  ROLE_USER = 1;
  ROLE_ADMIN = 2;
  ROLE_MODERATOR = 3;
}

message CreateUserRequest {
  string username = 1;
  string email = 2;
  string full_name = 3;
  Role role = 4;
}

message GetUserRequest {
  string id = 1;
}

message ListUsersRequest {
  string role_filter = 1;
  int32 limit = 2;
}

message BatchCreateResponse {
  int32 created_count = 1;
  repeated string user_ids = 2;
}

message UserUpdateRequest {
  string user_id = 1;
  string field = 2;
  string value = 3;
}

message UserUpdateResponse {
  string user_id = 1;
  bool success = 2;
  string message = 3;
}

message DeleteUserRequest {
  string id = 1;
}

message DeleteUserResponse {
  bool success = 1;
}
```

## Requirements

1. Define protobuf messages and service
2. Implement all RPC methods (unary, streaming)
3. Create gRPC server with proper configuration
4. Implement client with all RPC calls
5. Add server interceptors for logging and validation
6. Use proper gRPC status codes for errors
7. Implement metadata handling (auth tokens, request IDs)
8. Add graceful shutdown
9. Use in-memory storage with concurrent access control
10. Write comprehensive tests for all RPC methods

## Implementation Notes

Since we're implementing without code generation, we'll create:

1. Manual protobuf message structs
2. Server interface implementation
3. Client wrapper for easy usage
4. Stream handling with channels
5. Custom marshaling/unmarshaling
6. Interceptor chain implementation

## Example Usage

```go
// Server
server := NewUserServiceServer()
server.Start(":50051")

// Client
client := NewUserServiceClient(":50051")

// Unary RPC
user, err := client.CreateUser(ctx, &CreateUserRequest{
    Username: "john_doe",
    Email: "john@example.com",
    FullName: "John Doe",
    Role: Role_USER,
})

// Server streaming
stream, err := client.ListUsers(ctx, &ListUsersRequest{
    Limit: 10,
})
for {
    user, err := stream.Recv()
    if err == io.EOF {
        break
    }
    // Process user
}

// Client streaming
stream, err := client.BatchCreateUsers(ctx)
for _, req := range requests {
    stream.Send(req)
}
resp, err := stream.CloseAndRecv()

// Bidirectional streaming
stream, err := client.StreamUserUpdates(ctx)
go func() {
    for {
        resp, err := stream.Recv()
        // Handle response
    }
}()
stream.Send(&UserUpdateRequest{...})
```

## Learning Objectives

- gRPC service design principles
- Protocol Buffers (protobuf) usage
- Unary vs streaming RPC patterns
- Stream multiplexing and flow control
- gRPC error handling and status codes
- Interceptor pattern implementation
- Metadata propagation
- Context usage in gRPC
- Performance optimization techniques
- Testing gRPC services
- HTTP/2 fundamentals
- Service discovery concepts

## Testing Focus

- Test all unary RPC methods
- Test server streaming with multiple items
- Test client streaming aggregation
- Test bidirectional streaming
- Test error handling and status codes
- Test interceptors execution
- Test metadata transmission
- Test concurrent client connections
- Test graceful shutdown
- Benchmark RPC performance
