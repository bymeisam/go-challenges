package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// Protobuf-like message structures
type Role int32

const (
	RoleUnspecified Role = 0
	RoleUser        Role = 1
	RoleAdmin       Role = 2
	RoleModerator   Role = 3
)

func (r Role) String() string {
	switch r {
	case RoleUser:
		return "USER"
	case RoleAdmin:
		return "ADMIN"
	case RoleModerator:
		return "MODERATOR"
	default:
		return "UNSPECIFIED"
	}
}

type User struct {
	ID        string
	Username  string
	Email     string
	FullName  string
	Role      Role
	CreatedAt int64
	UpdatedAt int64
}

type CreateUserRequest struct {
	Username string
	Email    string
	FullName string
	Role     Role
}

type GetUserRequest struct {
	ID string
}

type ListUsersRequest struct {
	RoleFilter string
	Limit      int32
}

type BatchCreateResponse struct {
	CreatedCount int32
	UserIDs      []string
}

type UserUpdateRequest struct {
	UserID string
	Field  string
	Value  string
}

type UserUpdateResponse struct {
	UserID  string
	Success bool
	Message string
}

type DeleteUserRequest struct {
	ID string
}

type DeleteUserResponse struct {
	Success bool
}

// Storage
type UserStore struct {
	mu         sync.RWMutex
	users      map[string]*User
	nextID     int
	updateChan chan *UserUpdateRequest
}

func NewUserStore() *UserStore {
	return &UserStore{
		users:      make(map[string]*User),
		nextID:     1,
		updateChan: make(chan *UserUpdateRequest, 100),
	}
}

func (s *UserStore) Create(username, email, fullName string, role Role) (*User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if username == "" {
		return nil, errors.New("username is required")
	}
	if email == "" {
		return nil, errors.New("email is required")
	}

	// Check for duplicate username
	for _, u := range s.users {
		if u.Username == username {
			return nil, errors.New("username already exists")
		}
	}

	now := time.Now().Unix()
	user := &User{
		ID:        fmt.Sprintf("%d", s.nextID),
		Username:  username,
		Email:     email,
		FullName:  fullName,
		Role:      role,
		CreatedAt: now,
		UpdatedAt: now,
	}

	s.users[user.ID] = user
	s.nextID++
	return user, nil
}

func (s *UserStore) Get(id string) (*User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	user, exists := s.users[id]
	if !exists {
		return nil, errors.New("user not found")
	}
	return user, nil
}

func (s *UserStore) List(roleFilter string, limit int32) []*User {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var users []*User
	for _, user := range s.users {
		if roleFilter != "" && user.Role.String() != roleFilter {
			continue
		}
		users = append(users, user)
		if limit > 0 && len(users) >= int(limit) {
			break
		}
	}
	return users
}

func (s *UserStore) Update(id, field, value string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	user, exists := s.users[id]
	if !exists {
		return errors.New("user not found")
	}

	switch field {
	case "email":
		user.Email = value
	case "full_name":
		user.FullName = value
	default:
		return errors.New("invalid field")
	}

	user.UpdatedAt = time.Now().Unix()
	return nil
}

func (s *UserStore) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.users[id]; !exists {
		return errors.New("user not found")
	}

	delete(s.users, id)
	return nil
}

// gRPC Service Interface
type UserServiceServer interface {
	CreateUser(context.Context, *CreateUserRequest) (*User, error)
	GetUser(context.Context, *GetUserRequest) (*User, error)
	ListUsers(*ListUsersRequest, UserService_ListUsersServer) error
	BatchCreateUsers(UserService_BatchCreateUsersServer) error
	StreamUserUpdates(UserService_StreamUserUpdatesServer) error
	DeleteUser(context.Context, *DeleteUserRequest) (*DeleteUserResponse, error)
}

// Stream interfaces
type UserService_ListUsersServer interface {
	Send(*User) error
	grpc.ServerStream
}

type UserService_BatchCreateUsersServer interface {
	SendAndClose(*BatchCreateResponse) error
	Recv() (*CreateUserRequest, error)
	grpc.ServerStream
}

type UserService_StreamUserUpdatesServer interface {
	Send(*UserUpdateResponse) error
	Recv() (*UserUpdateRequest, error)
	grpc.ServerStream
}

// Service implementation
type userServiceImpl struct {
	store *UserStore
}

func NewUserService(store *UserStore) *userServiceImpl {
	return &userServiceImpl{store: store}
}

func (s *userServiceImpl) CreateUser(ctx context.Context, req *CreateUserRequest) (*User, error) {
	// Extract metadata
	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		if requestID := md.Get("request-id"); len(requestID) > 0 {
			log.Printf("Processing CreateUser request: %s", requestID[0])
		}
	}

	// Validate request
	if req.Username == "" {
		return nil, status.Error(codes.InvalidArgument, "username is required")
	}
	if req.Email == "" {
		return nil, status.Error(codes.InvalidArgument, "email is required")
	}

	user, err := s.store.Create(req.Username, req.Email, req.FullName, req.Role)
	if err != nil {
		if err.Error() == "username already exists" {
			return nil, status.Error(codes.AlreadyExists, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return user, nil
}

func (s *userServiceImpl) GetUser(ctx context.Context, req *GetUserRequest) (*User, error) {
	if req.ID == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	user, err := s.store.Get(req.ID)
	if err != nil {
		return nil, status.Error(codes.NotFound, "user not found")
	}

	return user, nil
}

func (s *userServiceImpl) ListUsers(req *ListUsersRequest, stream UserService_ListUsersServer) error {
	users := s.store.List(req.RoleFilter, req.Limit)

	for _, user := range users {
		if err := stream.Send(user); err != nil {
			return status.Error(codes.Internal, err.Error())
		}
		// Simulate streaming delay
		time.Sleep(10 * time.Millisecond)
	}

	return nil
}

func (s *userServiceImpl) BatchCreateUsers(stream UserService_BatchCreateUsersServer) error {
	var createdCount int32
	var userIDs []string

	for {
		req, err := stream.Recv()
		if err == io.EOF {
			// Client finished sending
			return stream.SendAndClose(&BatchCreateResponse{
				CreatedCount: createdCount,
				UserIDs:      userIDs,
			})
		}
		if err != nil {
			return status.Error(codes.Internal, err.Error())
		}

		user, err := s.store.Create(req.Username, req.Email, req.FullName, req.Role)
		if err != nil {
			log.Printf("Failed to create user %s: %v", req.Username, err)
			continue
		}

		createdCount++
		userIDs = append(userIDs, user.ID)
	}
}

func (s *userServiceImpl) StreamUserUpdates(stream UserService_StreamUserUpdatesServer) error {
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return status.Error(codes.Internal, err.Error())
		}

		// Process update
		err = s.store.Update(req.UserID, req.Field, req.Value)

		resp := &UserUpdateResponse{
			UserID:  req.UserID,
			Success: err == nil,
		}

		if err != nil {
			resp.Message = err.Error()
		} else {
			resp.Message = "Updated successfully"
		}

		if err := stream.Send(resp); err != nil {
			return status.Error(codes.Internal, err.Error())
		}
	}
}

func (s *userServiceImpl) DeleteUser(ctx context.Context, req *DeleteUserRequest) (*DeleteUserResponse, error) {
	if req.ID == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	err := s.store.Delete(req.ID)
	if err != nil {
		return nil, status.Error(codes.NotFound, "user not found")
	}

	return &DeleteUserResponse{Success: true}, nil
}

// Interceptors
func loggingInterceptor(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {
	start := time.Now()

	// Call handler
	resp, err := handler(ctx, req)

	// Log
	duration := time.Since(start)
	status := "OK"
	if err != nil {
		status = "ERROR"
	}

	log.Printf("[%s] %s - %v", status, info.FullMethod, duration)

	return resp, err
}

func authInterceptor(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {
	// Extract metadata
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "missing metadata")
	}

	// Check for auth token (simplified)
	tokens := md.Get("authorization")
	if len(tokens) == 0 {
		// For demo, we'll allow requests without auth
		log.Println("Warning: No auth token provided")
	} else {
		log.Printf("Auth token: %s", tokens[0])
	}

	return handler(ctx, req)
}

// Server wrapper
type Server struct {
	grpcServer *grpc.Server
	store      *UserStore
	service    *userServiceImpl
}

func NewServer() *Server {
	store := NewUserStore()
	service := NewUserService(store)

	// Create gRPC server with interceptors
	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			loggingInterceptor,
			authInterceptor,
		),
	)

	return &Server{
		grpcServer: grpcServer,
		store:      store,
		service:    service,
	}
}

func (s *Server) Start(address string) error {
	lis, err := net.Listen("tcp", address)
	if err != nil {
		return fmt.Errorf("failed to listen: %v", err)
	}

	log.Printf("gRPC server listening on %s", address)
	return s.grpcServer.Serve(lis)
}

func (s *Server) Stop() {
	log.Println("Stopping gRPC server...")
	s.grpcServer.GracefulStop()
}

func (s *Server) GetService() *userServiceImpl {
	return s.service
}

// Client wrapper
type Client struct {
	conn *grpc.ClientConn
}

func NewClient(address string) (*Client, error) {
	conn, err := grpc.Dial(address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
		grpc.WithTimeout(5*time.Second),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect: %v", err)
	}

	return &Client{conn: conn}, nil
}

func (c *Client) Close() error {
	return c.conn.Close()
}

func (c *Client) CreateUser(ctx context.Context, req *CreateUserRequest) (*User, error) {
	// Add metadata
	md := metadata.New(map[string]string{
		"request-id":    fmt.Sprintf("req-%d", time.Now().UnixNano()),
		"authorization": "Bearer demo-token",
	})
	ctx = metadata.NewOutgoingContext(ctx, md)

	var user User
	err := c.conn.Invoke(ctx, "/user.UserService/CreateUser", req, &user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (c *Client) GetUser(ctx context.Context, req *GetUserRequest) (*User, error) {
	var user User
	err := c.conn.Invoke(ctx, "/user.UserService/GetUser", req, &user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (c *Client) DeleteUser(ctx context.Context, req *DeleteUserRequest) (*DeleteUserResponse, error) {
	var resp DeleteUserResponse
	err := c.conn.Invoke(ctx, "/user.UserService/DeleteUser", req, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// For demonstration purposes
func demonstrateUsage() {
	// This would be used in a real scenario
	// Here we show the structure

	store := NewUserStore()
	service := NewUserService(store)

	// Create user
	ctx := context.Background()
	user, err := service.CreateUser(ctx, &CreateUserRequest{
		Username: "john_doe",
		Email:    "john@example.com",
		FullName: "John Doe",
		Role:     RoleUser,
	})

	if err != nil {
		log.Printf("Error creating user: %v", err)
		return
	}

	log.Printf("Created user: %+v", user)

	// Get user
	fetchedUser, err := service.GetUser(ctx, &GetUserRequest{ID: user.ID})
	if err != nil {
		log.Printf("Error fetching user: %v", err)
		return
	}

	log.Printf("Fetched user: %+v", fetchedUser)
}

func main() {
	// Run demonstration
	demonstrateUsage()

	// To run actual server:
	// server := NewServer()
	// if err := server.Start(":50051"); err != nil {
	//     log.Fatalf("Failed to start server: %v", err)
	// }
}
