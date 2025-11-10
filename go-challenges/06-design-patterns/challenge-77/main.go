package main

type UserService struct {
	repo UserRepository
}

func NewUserService(repo UserRepository) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) RegisterUser(name string) (*User, error) {
	user := User{
		ID:   len(s.repo.(*InMemoryUserRepository).users) + 1,
		Name: name,
	}
	err := s.repo.Save(user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *UserService) GetUser(id int) (*User, error) {
	return s.repo.FindByID(id)
}

func main() {}
