package main

type User struct {
	ID   int
	Name string
}

type UserRepository interface {
	FindByID(id int) (*User, error)
	Save(user *User) error
}

type UserService struct {
	repo UserRepository
}

func NewUserService(repo UserRepository) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) GetUser(id int) (*User, error) {
	return s.repo.FindByID(id)
}

// Real implementation
type DBUserRepository struct {
	users map[int]*User
}

func NewDBUserRepository() *DBUserRepository {
	return &DBUserRepository{
		users: make(map[int]*User),
	}
}

func (r *DBUserRepository) FindByID(id int) (*User, error) {
	return r.users[id], nil
}

func (r *DBUserRepository) Save(user *User) error {
	r.users[user.ID] = user
	return nil
}

// Mock implementation for testing
type MockUserRepository struct {
	users map[int]*User
}

func NewMockUserRepository() *MockUserRepository {
	return &MockUserRepository{
		users: map[int]*User{
			1: {ID: 1, Name: "Test User"},
		},
	}
}

func (r *MockUserRepository) FindByID(id int) (*User, error) {
	return r.users[id], nil
}

func (r *MockUserRepository) Save(user *User) error {
	r.users[user.ID] = user
	return nil
}

func main() {}
