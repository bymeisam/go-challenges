package main

import (
	"testing"
	"time"
)

func TestCreateUserCommand_Validate(t *testing.T) {
	tests := []struct {
		name    string
		cmd     CreateUserCommand
		wantErr bool
	}{
		{
			name: "valid command",
			cmd: CreateUserCommand{
				UserID: "user-1",
				Email:  "test@example.com",
				Name:   "Test User",
			},
			wantErr: false,
		},
		{
			name: "missing user ID",
			cmd: CreateUserCommand{
				Email: "test@example.com",
				Name:  "Test User",
			},
			wantErr: true,
		},
		{
			name: "missing email",
			cmd: CreateUserCommand{
				UserID: "user-1",
				Name:   "Test User",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cmd.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestWriteStore_Save(t *testing.T) {
	store := NewWriteStore()

	user := &UserWriteModel{
		ID:        "user-1",
		Email:     "test@example.com",
		Name:      "Test User",
		CreatedAt: time.Now(),
	}

	err := store.Save(user)
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	if user.Version != 1 {
		t.Errorf("Version = %d, want 1", user.Version)
	}

	loaded, err := store.Get("user-1")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if loaded.Email != user.Email {
		t.Errorf("Email = %s, want %s", loaded.Email, user.Email)
	}
}

func TestWriteStore_Delete(t *testing.T) {
	store := NewWriteStore()

	user := &UserWriteModel{
		ID:        "user-1",
		Email:     "test@example.com",
		Name:      "Test User",
		CreatedAt: time.Now(),
	}

	store.Save(user)

	err := store.Delete("user-1")
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	_, err = store.Get("user-1")
	if err == nil {
		t.Error("Get should fail after delete")
	}
}

func TestReadStore_Save(t *testing.T) {
	store := NewReadStore()

	user := &UserReadModel{
		ID:        "user-1",
		Email:     "test@example.com",
		Name:      "Test User",
		CreatedAt: time.Now().Format(time.RFC3339),
	}

	err := store.Save(user)
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	loaded, err := store.Get("user-1")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if loaded.Email != user.Email {
		t.Errorf("Email = %s, want %s", loaded.Email, user.Email)
	}
}

func TestReadStore_List(t *testing.T) {
	store := NewReadStore()

	// Add users
	for i := 1; i <= 10; i++ {
		user := &UserReadModel{
			ID:        "user-" + string(rune(i+'0')),
			Email:     "user@example.com",
			Name:      "User",
			CreatedAt: time.Now().Format(time.RFC3339),
		}
		store.Save(user)
	}

	users := store.List(5, 0)
	if len(users) != 5 {
		t.Errorf("List returned %d users, want 5", len(users))
	}
}

func TestReadStore_Search(t *testing.T) {
	store := NewReadStore()

	users := []*UserReadModel{
		{ID: "1", Email: "alice@example.com", Name: "Alice", CreatedAt: ""},
		{ID: "2", Email: "bob@example.com", Name: "Bob", CreatedAt: ""},
		{ID: "3", Email: "charlie@example.com", Name: "Charlie", CreatedAt: ""},
	}

	for _, user := range users {
		store.Save(user)
	}

	results := store.Search("bob")
	if len(results) != 1 {
		t.Errorf("Search returned %d results, want 1", len(results))
	}

	if results[0].Name != "Bob" {
		t.Errorf("Found user name = %s, want Bob", results[0].Name)
	}
}

func TestReadStore_Cache(t *testing.T) {
	store := NewReadStore()

	// Add users
	for i := 1; i <= 5; i++ {
		user := &UserReadModel{
			ID:    "user-" + string(rune(i+'0')),
			Email: "user@example.com",
			Name:  "User",
		}
		store.Save(user)
	}

	// First list (cache miss)
	users1 := store.List(5, 0)

	// Second list (cache hit)
	users2 := store.List(5, 0)

	if len(users1) != len(users2) {
		t.Error("Cached results should match")
	}
}

func TestCommandBus_Execute(t *testing.T) {
	bus := NewCommandBus()

	executed := false
	handler := func(cmd Command) error {
		executed = true
		return nil
	}

	bus.RegisterHandler("CreateUser", handler)

	cmd := CreateUserCommand{
		UserID: "user-1",
		Email:  "test@example.com",
		Name:   "Test User",
	}

	err := bus.Execute(cmd)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if !executed {
		t.Error("Handler should have been executed")
	}
}

func TestCommandBus_ValidationFailure(t *testing.T) {
	bus := NewCommandBus()

	handler := func(cmd Command) error {
		return nil
	}

	bus.RegisterHandler("CreateUser", handler)

	// Invalid command (missing user ID)
	cmd := CreateUserCommand{
		Email: "test@example.com",
		Name:  "Test User",
	}

	err := bus.Execute(cmd)
	if err == nil {
		t.Error("Execute should fail validation")
	}
}

func TestCommandBus_Events(t *testing.T) {
	bus := NewCommandBus()

	handler := func(cmd Command) error {
		return nil
	}

	bus.RegisterHandler("CreateUser", handler)

	cmd := CreateUserCommand{
		UserID: "user-1",
		Email:  "test@example.com",
		Name:   "Test User",
	}

	bus.Execute(cmd)

	// Check event was published
	select {
	case event := <-bus.Events():
		if event.Type != "CreateUser" {
			t.Errorf("Event type = %s, want CreateUser", event.Type)
		}
	case <-time.After(1 * time.Second):
		t.Error("Timeout waiting for event")
	}
}

func TestQueryBus_Execute(t *testing.T) {
	bus := NewQueryBus()

	handler := func(query Query) (interface{}, error) {
		return &UserReadModel{
			ID:    "user-1",
			Email: "test@example.com",
			Name:  "Test User",
		}, nil
	}

	bus.RegisterHandler("GetUser", handler)

	query := GetUserQuery{UserID: "user-1"}

	result, err := bus.Execute(query)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	user := result.(*UserReadModel)
	if user.ID != "user-1" {
		t.Errorf("User ID = %s, want user-1", user.ID)
	}
}

func TestCQRSApp_CreateUser(t *testing.T) {
	app := NewCQRSApp()

	cmd := CreateUserCommand{
		UserID: "user-1",
		Email:  "test@example.com",
		Name:   "Test User",
	}

	err := app.commandBus.Execute(cmd)
	if err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}

	// Wait for eventual consistency
	time.Sleep(100 * time.Millisecond)

	// Query read model
	query := GetUserQuery{UserID: "user-1"}
	result, err := app.queryBus.Execute(query)
	if err != nil {
		t.Fatalf("GetUser failed: %v", err)
	}

	user := result.(*UserReadModel)
	if user.Email != cmd.Email {
		t.Errorf("Email = %s, want %s", user.Email, cmd.Email)
	}
}

func TestCQRSApp_UpdateUser(t *testing.T) {
	app := NewCQRSApp()

	// Create user
	app.commandBus.Execute(CreateUserCommand{
		UserID: "user-1",
		Email:  "test@example.com",
		Name:   "Old Name",
	})

	time.Sleep(50 * time.Millisecond)

	// Update user
	app.commandBus.Execute(UpdateUserCommand{
		UserID: "user-1",
		Name:   "New Name",
	})

	time.Sleep(50 * time.Millisecond)

	// Query
	result, _ := app.queryBus.Execute(GetUserQuery{UserID: "user-1"})
	user := result.(*UserReadModel)

	if user.Name != "New Name" {
		t.Errorf("Name = %s, want New Name", user.Name)
	}
}

func TestCQRSApp_DeleteUser(t *testing.T) {
	app := NewCQRSApp()

	// Create user
	app.commandBus.Execute(CreateUserCommand{
		UserID: "user-1",
		Email:  "test@example.com",
		Name:   "Test User",
	})

	time.Sleep(50 * time.Millisecond)

	// Delete user
	app.commandBus.Execute(DeleteUserCommand{
		UserID: "user-1",
	})

	time.Sleep(50 * time.Millisecond)

	// Query should fail
	_, err := app.queryBus.Execute(GetUserQuery{UserID: "user-1"})
	if err == nil {
		t.Error("GetUser should fail after deletion")
	}
}

func TestCQRSApp_ListUsers(t *testing.T) {
	app := NewCQRSApp()

	// Create users
	for i := 1; i <= 5; i++ {
		app.commandBus.Execute(CreateUserCommand{
			UserID: "user-" + string(rune(i+'0')),
			Email:  "user@example.com",
			Name:   "User",
		})
	}

	time.Sleep(100 * time.Millisecond)

	// List users
	result, err := app.queryBus.Execute(ListUsersQuery{
		Limit:  10,
		Offset: 0,
	})

	if err != nil {
		t.Fatalf("ListUsers failed: %v", err)
	}

	users := result.([]*UserListItemReadModel)
	if len(users) != 5 {
		t.Errorf("List returned %d users, want 5", len(users))
	}
}

func TestCQRSApp_SearchUsers(t *testing.T) {
	app := NewCQRSApp()

	// Create users
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

	time.Sleep(100 * time.Millisecond)

	// Search
	result, _ := app.queryBus.Execute(SearchUsersQuery{
		SearchTerm: "alice",
	})

	users := result.([]*UserReadModel)
	if len(users) != 1 {
		t.Errorf("Search returned %d results, want 1", len(users))
	}
}

func TestEventualConsistency(t *testing.T) {
	app := NewCQRSApp()

	// Create user
	app.commandBus.Execute(CreateUserCommand{
		UserID: "user-1",
		Email:  "test@example.com",
		Name:   "Test User",
	})

	// Immediately query (might not be synced yet)
	_, err := app.queryBus.Execute(GetUserQuery{UserID: "user-1"})

	// Wait for sync
	time.Sleep(100 * time.Millisecond)

	// Should be available now
	_, err = app.queryBus.Execute(GetUserQuery{UserID: "user-1"})
	if err != nil {
		t.Error("User should be available after sync")
	}
}

func BenchmarkCommandExecution(b *testing.B) {
	app := NewCQRSApp()

	cmd := CreateUserCommand{
		UserID: "user-1",
		Email:  "test@example.com",
		Name:   "Test User",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		app.commandBus.Execute(cmd)
	}
}

func BenchmarkQueryExecution(b *testing.B) {
	app := NewCQRSApp()

	// Create user
	app.commandBus.Execute(CreateUserCommand{
		UserID: "user-1",
		Email:  "test@example.com",
		Name:   "Test User",
	})

	time.Sleep(100 * time.Millisecond)

	query := GetUserQuery{UserID: "user-1"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		app.queryBus.Execute(query)
	}
}

func BenchmarkListQueryWithCache(b *testing.B) {
	app := NewCQRSApp()

	// Create users
	for i := 0; i < 100; i++ {
		app.commandBus.Execute(CreateUserCommand{
			UserID: "user-" + string(rune(i+'0')),
			Email:  "user@example.com",
			Name:   "User",
		})
	}

	time.Sleep(200 * time.Millisecond)

	query := ListUsersQuery{Limit: 10, Offset: 0}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		app.queryBus.Execute(query)
	}
}
