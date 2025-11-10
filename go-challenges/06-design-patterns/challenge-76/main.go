package main

type User struct {
	ID   int
	Name string
}

type UserRepository interface {
	Save(user User) error
	FindByID(id int) (*User, error)
	FindAll() ([]User, error)
}

type InMemoryUserRepository struct {
	users map[int]User
}

func NewInMemoryUserRepository() *InMemoryUserRepository {
	return &InMemoryUserRepository{
		users: make(map[int]User),
	}
}

func (r *InMemoryUserRepository) Save(user User) error {
	r.users[user.ID] = user
	return nil
}

func (r *InMemoryUserRepository) FindByID(id int) (*User, error) {
	user, exists := r.users[id]
	if !exists {
		return nil, nil
	}
	return &user, nil
}

func (r *InMemoryUserRepository) FindAll() ([]User, error) {
	users := make([]User, 0, len(r.users))
	for _, user := range r.users {
		users = append(users, user)
	}
	return users, nil
}

func main() {}
