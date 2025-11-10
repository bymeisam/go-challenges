package main

import (
	"errors"
	"fmt"
	"log"
	"sync"
	"time"
)

// Command represents a write operation
type Command interface {
	GetCommandName() string
	Validate() error
}

// Query represents a read operation
type Query interface {
	GetQueryName() string
}

// Command types
type CreateUserCommand struct {
	UserID string
	Email  string
	Name   string
}

func (c CreateUserCommand) GetCommandName() string { return "CreateUser" }
func (c CreateUserCommand) Validate() error {
	if c.UserID == "" {
		return errors.New("user ID is required")
	}
	if c.Email == "" {
		return errors.New("email is required")
	}
	return nil
}

type UpdateUserCommand struct {
	UserID string
	Name   string
}

func (c UpdateUserCommand) GetCommandName() string { return "UpdateUser" }
func (c UpdateUserCommand) Validate() error {
	if c.UserID == "" {
		return errors.New("user ID is required")
	}
	return nil
}

type DeleteUserCommand struct {
	UserID string
}

func (c DeleteUserCommand) GetCommandName() string { return "DeleteUser" }
func (c DeleteUserCommand) Validate() error {
	if c.UserID == "" {
		return errors.New("user ID is required")
	}
	return nil
}

// Query types
type GetUserQuery struct {
	UserID string
}

func (q GetUserQuery) GetQueryName() string { return "GetUser" }

type ListUsersQuery struct {
	Limit  int
	Offset int
}

func (q ListUsersQuery) GetQueryName() string { return "ListUsers" }

type SearchUsersQuery struct {
	SearchTerm string
}

func (q SearchUsersQuery) GetQueryName() string { return "SearchUsers" }

// Write model (optimized for commands)
type UserWriteModel struct {
	ID        string
	Email     string
	Name      string
	Version   int
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Read model (optimized for queries)
type UserReadModel struct {
	ID        string
	Email     string
	Name      string
	CreatedAt string
}

type UserListItemReadModel struct {
	ID    string
	Email string
	Name  string
}

// Write store
type WriteStore struct {
	mu    sync.RWMutex
	users map[string]*UserWriteModel
}

func NewWriteStore() *WriteStore {
	return &WriteStore{
		users: make(map[string]*UserWriteModel),
	}
}

func (s *WriteStore) Save(user *UserWriteModel) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	user.UpdatedAt = time.Now()
	user.Version++
	s.users[user.ID] = user
	return nil
}

func (s *WriteStore) Get(id string) (*UserWriteModel, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	user, exists := s.users[id]
	if !exists {
		return nil, errors.New("user not found")
	}
	return user, nil
}

func (s *WriteStore) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.users[id]; !exists {
		return errors.New("user not found")
	}
	delete(s.users, id)
	return nil
}

// Read store
type ReadStore struct {
	mu    sync.RWMutex
	users map[string]*UserReadModel
	cache map[string]interface{}
}

func NewReadStore() *ReadStore {
	return &ReadStore{
		users: make(map[string]*UserReadModel),
		cache: make(map[string]interface{}),
	}
}

func (s *ReadStore) Save(user *UserReadModel) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.users[user.ID] = user
	s.invalidateCache()
	return nil
}

func (s *ReadStore) Get(id string) (*UserReadModel, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	user, exists := s.users[id]
	if !exists {
		return nil, errors.New("user not found")
	}
	return user, nil
}

func (s *ReadStore) List(limit, offset int) []*UserListItemReadModel {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Check cache
	cacheKey := fmt.Sprintf("list:%d:%d", limit, offset)
	if cached, exists := s.cache[cacheKey]; exists {
		return cached.([]*UserListItemReadModel)
	}

	var users []*UserListItemReadModel
	count := 0
	for _, user := range s.users {
		if count >= offset && len(users) < limit {
			users = append(users, &UserListItemReadModel{
				ID:    user.ID,
				Email: user.Email,
				Name:  user.Name,
			})
		}
		count++
	}

	s.cache[cacheKey] = users
	return users
}

func (s *ReadStore) Search(term string) []*UserReadModel {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var results []*UserReadModel
	for _, user := range s.users {
		if contains(user.Name, term) || contains(user.Email, term) {
			results = append(results, user)
		}
	}
	return results
}

func (s *ReadStore) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.users, id)
	s.invalidateCache()
	return nil
}

func (s *ReadStore) invalidateCache() {
	s.cache = make(map[string]interface{})
}

// Event for synchronization
type DomainEvent struct {
	Type      string
	Data      interface{}
	Timestamp time.Time
}

// Command bus
type CommandBus struct {
	handlers map[string]CommandHandler
	events   chan DomainEvent
}

type CommandHandler func(Command) error

func NewCommandBus() *CommandBus {
	return &CommandBus{
		handlers: make(map[string]CommandHandler),
		events:   make(chan DomainEvent, 100),
	}
}

func (bus *CommandBus) RegisterHandler(commandName string, handler CommandHandler) {
	bus.handlers[commandName] = handler
}

func (bus *CommandBus) Execute(cmd Command) error {
	// Validate command
	if err := cmd.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Execute handler
	handler, exists := bus.handlers[cmd.GetCommandName()]
	if !exists {
		return fmt.Errorf("no handler for command: %s", cmd.GetCommandName())
	}

	if err := handler(cmd); err != nil {
		return err
	}

	// Publish event
	event := DomainEvent{
		Type:      cmd.GetCommandName(),
		Data:      cmd,
		Timestamp: time.Now(),
	}

	select {
	case bus.events <- event:
	default:
		log.Println("Event channel full")
	}

	return nil
}

func (bus *CommandBus) Events() <-chan DomainEvent {
	return bus.events
}

// Query bus
type QueryBus struct {
	handlers map[string]QueryHandler
}

type QueryHandler func(Query) (interface{}, error)

func NewQueryBus() *QueryBus {
	return &QueryBus{
		handlers: make(map[string]QueryHandler),
	}
}

func (bus *QueryBus) RegisterHandler(queryName string, handler QueryHandler) {
	bus.handlers[queryName] = handler
}

func (bus *QueryBus) Execute(query Query) (interface{}, error) {
	handler, exists := bus.handlers[query.GetQueryName()]
	if !exists {
		return nil, fmt.Errorf("no handler for query: %s", query.GetQueryName())
	}

	return handler(query)
}

// CQRS Application
type CQRSApp struct {
	commandBus *CommandBus
	queryBus   *QueryBus
	writeStore *WriteStore
	readStore  *ReadStore
}

func NewCQRSApp() *CQRSApp {
	app := &CQRSApp{
		commandBus: NewCommandBus(),
		queryBus:   NewQueryBus(),
		writeStore: NewWriteStore(),
		readStore:  NewReadStore(),
	}

	app.setupHandlers()
	app.startEventProcessor()

	return app
}

func (app *CQRSApp) setupHandlers() {
	// Command handlers
	app.commandBus.RegisterHandler("CreateUser", func(cmd Command) error {
		c := cmd.(CreateUserCommand)

		user := &UserWriteModel{
			ID:        c.UserID,
			Email:     c.Email,
			Name:      c.Name,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		return app.writeStore.Save(user)
	})

	app.commandBus.RegisterHandler("UpdateUser", func(cmd Command) error {
		c := cmd.(UpdateUserCommand)

		user, err := app.writeStore.Get(c.UserID)
		if err != nil {
			return err
		}

		user.Name = c.Name
		return app.writeStore.Save(user)
	})

	app.commandBus.RegisterHandler("DeleteUser", func(cmd Command) error {
		c := cmd.(DeleteUserCommand)
		return app.writeStore.Delete(c.UserID)
	})

	// Query handlers
	app.queryBus.RegisterHandler("GetUser", func(query Query) (interface{}, error) {
		q := query.(GetUserQuery)
		return app.readStore.Get(q.UserID)
	})

	app.queryBus.RegisterHandler("ListUsers", func(query Query) (interface{}, error) {
		q := query.(ListUsersQuery)
		return app.readStore.List(q.Limit, q.Offset), nil
	})

	app.queryBus.RegisterHandler("SearchUsers", func(query Query) (interface{}, error) {
		q := query.(SearchUsersQuery)
		return app.readStore.Search(q.SearchTerm), nil
	})
}

func (app *CQRSApp) startEventProcessor() {
	go func() {
		for event := range app.commandBus.Events() {
			app.processEvent(event)
		}
	}()
}

func (app *CQRSApp) processEvent(event DomainEvent) {
	// Synchronize read model with write model
	switch event.Type {
	case "CreateUser":
		cmd := event.Data.(CreateUserCommand)
		readModel := &UserReadModel{
			ID:        cmd.UserID,
			Email:     cmd.Email,
			Name:      cmd.Name,
			CreatedAt: event.Timestamp.Format(time.RFC3339),
		}
		app.readStore.Save(readModel)

	case "UpdateUser":
		cmd := event.Data.(UpdateUserCommand)
		readModel, err := app.readStore.Get(cmd.UserID)
		if err == nil {
			readModel.Name = cmd.Name
			app.readStore.Save(readModel)
		}

	case "DeleteUser":
		cmd := event.Data.(DeleteUserCommand)
		app.readStore.Delete(cmd.UserID)
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || (len(s) > 0 && (s[0:len(substr)] == substr || contains(s[1:], substr))))
}

func main() {
	// Create CQRS application
	app := NewCQRSApp()

	// Execute commands
	fmt.Println("Executing commands...")

	app.commandBus.Execute(CreateUserCommand{
		UserID: "user-1",
		Email:  "alice@example.com",
		Name:   "Alice",
	})

	app.commandBus.Execute(CreateUserCommand{
		UserID: "user-2",
		Email:  "bob@example.com",
		Name:   "Bob",
	})

	app.commandBus.Execute(CreateUserCommand{
		UserID: "user-3",
		Email:  "charlie@example.com",
		Name:   "Charlie",
	})

	app.commandBus.Execute(UpdateUserCommand{
		UserID: "user-1",
		Name:   "Alice Updated",
	})

	// Wait for eventual consistency
	time.Sleep(100 * time.Millisecond)

	// Execute queries
	fmt.Println("\nExecuting queries...")

	// Get user
	result, err := app.queryBus.Execute(GetUserQuery{UserID: "user-1"})
	if err != nil {
		log.Printf("Query failed: %v", err)
	} else {
		user := result.(*UserReadModel)
		fmt.Printf("User: %s - %s (%s)\n", user.ID, user.Name, user.Email)
	}

	// List users
	result, _ = app.queryBus.Execute(ListUsersQuery{Limit: 10, Offset: 0})
	users := result.([]*UserListItemReadModel)
	fmt.Println("\nUser list:")
	for _, user := range users {
		fmt.Printf("  - %s: %s (%s)\n", user.ID, user.Name, user.Email)
	}

	// Search users
	result, _ = app.queryBus.Execute(SearchUsersQuery{SearchTerm: "bob"})
	searchResults := result.([]*UserReadModel)
	fmt.Println("\nSearch results for 'bob':")
	for _, user := range searchResults {
		fmt.Printf("  - %s: %s (%s)\n", user.ID, user.Name, user.Email)
	}

	// Delete user
	app.commandBus.Execute(DeleteUserCommand{UserID: "user-2"})
	time.Sleep(50 * time.Millisecond)

	// List again
	result, _ = app.queryBus.Execute(ListUsersQuery{Limit: 10, Offset: 0})
	users = result.([]*UserListItemReadModel)
	fmt.Println("\nUser list after deletion:")
	for _, user := range users {
		fmt.Printf("  - %s: %s (%s)\n", user.ID, user.Name, user.Email)
	}

	fmt.Println("\nDemo completed!")
}
