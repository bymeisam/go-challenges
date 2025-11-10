package main

import (
	"context"
	"sync"
	"testing"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestUserStore_Create(t *testing.T) {
	store := NewUserStore()

	tests := []struct {
		name     string
		username string
		email    string
		fullName string
		role     Role
		wantErr  bool
	}{
		{
			name:     "valid user",
			username: "john_doe",
			email:    "john@example.com",
			fullName: "John Doe",
			role:     RoleUser,
			wantErr:  false,
		},
		{
			name:     "empty username",
			username: "",
			email:    "test@example.com",
			fullName: "Test User",
			role:     RoleUser,
			wantErr:  true,
		},
		{
			name:     "empty email",
			username: "test",
			email:    "",
			fullName: "Test User",
			role:     RoleUser,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, err := store.Create(tt.username, tt.email, tt.fullName, tt.role)
			if (err != nil) != tt.wantErr {
				t.Errorf("Create() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if user.Username != tt.username {
					t.Errorf("Create() username = %v, want %v", user.Username, tt.username)
				}
				if user.Email != tt.email {
					t.Errorf("Create() email = %v, want %v", user.Email, tt.email)
				}
				if user.Role != tt.role {
					t.Errorf("Create() role = %v, want %v", user.Role, tt.role)
				}
			}
		})
	}
}

func TestUserStore_DuplicateUsername(t *testing.T) {
	store := NewUserStore()

	_, err := store.Create("duplicate", "user1@example.com", "User 1", RoleUser)
	if err != nil {
		t.Fatalf("First Create() failed: %v", err)
	}

	_, err = store.Create("duplicate", "user2@example.com", "User 2", RoleUser)
	if err == nil {
		t.Error("Create() should fail for duplicate username")
	}
}

func TestUserStore_Get(t *testing.T) {
	store := NewUserStore()
	created, _ := store.Create("test_user", "test@example.com", "Test User", RoleUser)

	tests := []struct {
		name    string
		id      string
		wantErr bool
	}{
		{"existing user", created.ID, false},
		{"non-existing user", "999", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, err := store.Get(tt.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && user.ID != tt.id {
				t.Errorf("Get() id = %v, want %v", user.ID, tt.id)
			}
		})
	}
}

func TestUserStore_List(t *testing.T) {
	store := NewUserStore()

	// Create test users
	store.Create("user1", "user1@example.com", "User 1", RoleUser)
	store.Create("admin1", "admin1@example.com", "Admin 1", RoleAdmin)
	store.Create("user2", "user2@example.com", "User 2", RoleUser)
	store.Create("mod1", "mod1@example.com", "Mod 1", RoleModerator)

	tests := []struct {
		name       string
		roleFilter string
		limit      int32
		wantCount  int
	}{
		{"all users", "", 0, 4},
		{"filter by USER role", "USER", 0, 2},
		{"filter by ADMIN role", "ADMIN", 0, 1},
		{"with limit", "", 2, 2},
		{"filter and limit", "USER", 1, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			users := store.List(tt.roleFilter, tt.limit)
			if len(users) != tt.wantCount {
				t.Errorf("List() count = %v, want %v", len(users), tt.wantCount)
			}
		})
	}
}

func TestUserStore_Update(t *testing.T) {
	store := NewUserStore()
	user, _ := store.Create("test_user", "old@example.com", "Old Name", RoleUser)

	tests := []struct {
		name    string
		id      string
		field   string
		value   string
		wantErr bool
	}{
		{"update email", user.ID, "email", "new@example.com", false},
		{"update full_name", user.ID, "full_name", "New Name", false},
		{"invalid field", user.ID, "invalid", "value", true},
		{"non-existing user", "999", "email", "test@example.com", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := store.Update(tt.id, tt.field, tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("Update() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}

	// Verify updates
	updated, _ := store.Get(user.ID)
	if updated.Email != "new@example.com" {
		t.Errorf("Email not updated, got %v", updated.Email)
	}
	if updated.FullName != "New Name" {
		t.Errorf("FullName not updated, got %v", updated.FullName)
	}
}

func TestUserStore_Delete(t *testing.T) {
	store := NewUserStore()
	user, _ := store.Create("test_user", "test@example.com", "Test User", RoleUser)

	// Delete existing user
	err := store.Delete(user.ID)
	if err != nil {
		t.Errorf("Delete() error = %v", err)
	}

	// Verify deletion
	_, err = store.Get(user.ID)
	if err == nil {
		t.Error("Get() should fail after deletion")
	}

	// Delete non-existing user
	err = store.Delete("999")
	if err == nil {
		t.Error("Delete() should fail for non-existing user")
	}
}

func TestUserService_CreateUser(t *testing.T) {
	store := NewUserStore()
	service := NewUserService(store)
	ctx := context.Background()

	tests := []struct {
		name      string
		req       *CreateUserRequest
		wantErr   bool
		wantCode  codes.Code
	}{
		{
			name: "valid request",
			req: &CreateUserRequest{
				Username: "john_doe",
				Email:    "john@example.com",
				FullName: "John Doe",
				Role:     RoleUser,
			},
			wantErr: false,
		},
		{
			name: "empty username",
			req: &CreateUserRequest{
				Username: "",
				Email:    "test@example.com",
				FullName: "Test",
				Role:     RoleUser,
			},
			wantErr:  true,
			wantCode: codes.InvalidArgument,
		},
		{
			name: "empty email",
			req: &CreateUserRequest{
				Username: "test",
				Email:    "",
				FullName: "Test",
				Role:     RoleUser,
			},
			wantErr:  true,
			wantCode: codes.InvalidArgument,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, err := service.CreateUser(ctx, tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateUser() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				st, ok := status.FromError(err)
				if !ok {
					t.Error("Error is not a gRPC status error")
					return
				}
				if st.Code() != tt.wantCode {
					t.Errorf("CreateUser() code = %v, want %v", st.Code(), tt.wantCode)
				}
			} else {
				if user.Username != tt.req.Username {
					t.Errorf("CreateUser() username = %v, want %v", user.Username, tt.req.Username)
				}
			}
		})
	}
}

func TestUserService_GetUser(t *testing.T) {
	store := NewUserStore()
	service := NewUserService(store)
	ctx := context.Background()

	// Create a test user
	created, _ := service.CreateUser(ctx, &CreateUserRequest{
		Username: "test_user",
		Email:    "test@example.com",
		FullName: "Test User",
		Role:     RoleUser,
	})

	tests := []struct {
		name     string
		id       string
		wantErr  bool
		wantCode codes.Code
	}{
		{"existing user", created.ID, false, codes.OK},
		{"non-existing user", "999", true, codes.NotFound},
		{"empty id", "", true, codes.InvalidArgument},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, err := service.GetUser(ctx, &GetUserRequest{ID: tt.id})
			if (err != nil) != tt.wantErr {
				t.Errorf("GetUser() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				st, ok := status.FromError(err)
				if !ok {
					t.Error("Error is not a gRPC status error")
					return
				}
				if st.Code() != tt.wantCode {
					t.Errorf("GetUser() code = %v, want %v", st.Code(), tt.wantCode)
				}
			} else {
				if user.ID != tt.id {
					t.Errorf("GetUser() id = %v, want %v", user.ID, tt.id)
				}
			}
		})
	}
}

func TestUserService_DeleteUser(t *testing.T) {
	store := NewUserStore()
	service := NewUserService(store)
	ctx := context.Background()

	// Create a test user
	created, _ := service.CreateUser(ctx, &CreateUserRequest{
		Username: "test_user",
		Email:    "test@example.com",
		FullName: "Test User",
		Role:     RoleUser,
	})

	// Delete user
	resp, err := service.DeleteUser(ctx, &DeleteUserRequest{ID: created.ID})
	if err != nil {
		t.Errorf("DeleteUser() error = %v", err)
	}
	if !resp.Success {
		t.Error("DeleteUser() success should be true")
	}

	// Verify deletion
	_, err = service.GetUser(ctx, &GetUserRequest{ID: created.ID})
	if err == nil {
		t.Error("GetUser() should fail after deletion")
	}

	// Delete non-existing user
	_, err = service.DeleteUser(ctx, &DeleteUserRequest{ID: "999"})
	if err == nil {
		t.Error("DeleteUser() should fail for non-existing user")
	}
	st, _ := status.FromError(err)
	if st.Code() != codes.NotFound {
		t.Errorf("DeleteUser() code = %v, want %v", st.Code(), codes.NotFound)
	}
}

func TestUserService_ListUsers_Mock(t *testing.T) {
	store := NewUserStore()
	service := NewUserService(store)

	// Create test users
	store.Create("user1", "user1@example.com", "User 1", RoleUser)
	store.Create("admin1", "admin1@example.com", "Admin 1", RoleAdmin)
	store.Create("user2", "user2@example.com", "User 2", RoleUser)

	// Mock stream
	type mockStream struct {
		users []*User
	}

	stream := &mockStream{users: []*User{}}

	// Get users from store directly (simulating stream)
	users := store.List("USER", 0)

	if len(users) != 2 {
		t.Errorf("Expected 2 USER role users, got %d", len(users))
	}
}

func TestRoleString(t *testing.T) {
	tests := []struct {
		role Role
		want string
	}{
		{RoleUser, "USER"},
		{RoleAdmin, "ADMIN"},
		{RoleModerator, "MODERATOR"},
		{RoleUnspecified, "UNSPECIFIED"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.role.String(); got != tt.want {
				t.Errorf("Role.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConcurrentAccess(t *testing.T) {
	store := NewUserStore()
	service := NewUserService(store)
	ctx := context.Background()

	const goroutines = 50
	var wg sync.WaitGroup
	wg.Add(goroutines)

	// Concurrent creates
	for i := 0; i < goroutines; i++ {
		go func(id int) {
			defer wg.Done()
			service.CreateUser(ctx, &CreateUserRequest{
				Username: "user" + string(rune(id+'0')),
				Email:    "user" + string(rune(id+'0')) + "@example.com",
				FullName: "User",
				Role:     RoleUser,
			})
		}(i)
	}

	wg.Wait()

	// Verify some users were created (due to unique constraint, not all will succeed)
	users := store.List("", 0)
	if len(users) == 0 {
		t.Error("No users created")
	}
}

func TestConcurrentReadWrite(t *testing.T) {
	store := NewUserStore()
	service := NewUserService(store)
	ctx := context.Background()

	// Create initial users
	for i := 0; i < 10; i++ {
		service.CreateUser(ctx, &CreateUserRequest{
			Username: "user" + string(rune(i+'0')),
			Email:    "user" + string(rune(i+'0')) + "@example.com",
			FullName: "User",
			Role:     RoleUser,
		})
	}

	const goroutines = 20
	var wg sync.WaitGroup
	wg.Add(goroutines)

	// Mix of reads and writes
	for i := 0; i < goroutines; i++ {
		go func(id int) {
			defer wg.Done()

			if id%2 == 0 {
				// Read
				service.GetUser(ctx, &GetUserRequest{ID: "1"})
			} else {
				// Write
				store.Update("1", "email", "updated@example.com")
			}
		}(i)
	}

	wg.Wait()
}

func TestUserStoreTimestamps(t *testing.T) {
	store := NewUserStore()

	user, err := store.Create("test_user", "test@example.com", "Test", RoleUser)
	if err != nil {
		t.Fatalf("Create() failed: %v", err)
	}

	if user.CreatedAt == 0 {
		t.Error("CreatedAt should be set")
	}
	if user.UpdatedAt == 0 {
		t.Error("UpdatedAt should be set")
	}
	if user.CreatedAt != user.UpdatedAt {
		t.Error("CreatedAt and UpdatedAt should be equal on creation")
	}

	// Wait a bit and update
	time.Sleep(10 * time.Millisecond)
	store.Update(user.ID, "email", "new@example.com")

	updated, _ := store.Get(user.ID)
	if updated.UpdatedAt <= user.UpdatedAt {
		t.Error("UpdatedAt should be updated after modification")
	}
	if updated.CreatedAt != user.CreatedAt {
		t.Error("CreatedAt should not change")
	}
}

func BenchmarkCreateUser(b *testing.B) {
	store := NewUserStore()
	service := NewUserService(store)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		service.CreateUser(ctx, &CreateUserRequest{
			Username: "user" + string(rune(i+'0')),
			Email:    "user@example.com",
			FullName: "User",
			Role:     RoleUser,
		})
	}
}

func BenchmarkGetUser(b *testing.B) {
	store := NewUserStore()
	service := NewUserService(store)
	ctx := context.Background()

	// Create a user
	user, _ := service.CreateUser(ctx, &CreateUserRequest{
		Username: "test",
		Email:    "test@example.com",
		FullName: "Test",
		Role:     RoleUser,
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		service.GetUser(ctx, &GetUserRequest{ID: user.ID})
	}
}

func BenchmarkListUsers(b *testing.B) {
	store := NewUserStore()

	// Create 100 users
	for i := 0; i < 100; i++ {
		store.Create("user"+string(rune(i+'0')), "email@example.com", "User", RoleUser)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		store.List("", 0)
	}
}

func BenchmarkConcurrentReads(b *testing.B) {
	store := NewUserStore()
	service := NewUserService(store)
	ctx := context.Background()

	// Create users
	for i := 0; i < 100; i++ {
		service.CreateUser(ctx, &CreateUserRequest{
			Username: "user" + string(rune(i+'0')),
			Email:    "email@example.com",
			FullName: "User",
			Role:     RoleUser,
		})
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			service.GetUser(ctx, &GetUserRequest{ID: "1"})
		}
	})
}
