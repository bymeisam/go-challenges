package main

import (
	"testing"
)

func setupTestDB(t *testing.T) *Database {
	db, err := NewDatabase(":memory:")
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	return db
}

func TestCreateAuthor(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	author := &Author{
		Name:  "Test Author",
		Email: "author@example.com",
	}

	err := db.CreateAuthor(author)
	if err != nil {
		t.Fatalf("Failed to create author: %v", err)
	}

	if author.ID == 0 {
		t.Error("Expected ID to be set")
	}
}

func TestCreateBook(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	author := &Author{Name: "Author", Email: "author@example.com"}
	db.CreateAuthor(author)

	book := &Book{
		Title:    "Test Book",
		ISBN:     "123-456",
		AuthorID: author.ID,
	}

	err := db.CreateBook(book)
	if err != nil {
		t.Fatalf("Failed to create book: %v", err)
	}

	if book.ID == 0 {
		t.Error("Expected ID to be set")
	}
}

func TestGetAuthorWithBooks(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	author := &Author{Name: "John Doe", Email: "john@example.com"}
	db.CreateAuthor(author)

	db.CreateBook(&Book{Title: "Book 1", ISBN: "111", AuthorID: author.ID})
	db.CreateBook(&Book{Title: "Book 2", ISBN: "222", AuthorID: author.ID})
	db.CreateBook(&Book{Title: "Book 3", ISBN: "333", AuthorID: author.ID})

	retrieved, err := db.GetAuthorWithBooks(author.ID)
	if err != nil {
		t.Fatalf("Failed to get author with books: %v", err)
	}

	if retrieved.Name != "John Doe" {
		t.Errorf("Expected author name 'John Doe', got %s", retrieved.Name)
	}

	if len(retrieved.Books) != 3 {
		t.Errorf("Expected 3 books, got %d", len(retrieved.Books))
	}
}

func TestGetBookWithAuthor(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	author := &Author{Name: "Jane Smith", Email: "jane@example.com"}
	db.CreateAuthor(author)

	book := &Book{Title: "Test Book", ISBN: "456", AuthorID: author.ID}
	db.CreateBook(book)

	retrieved, err := db.GetBookWithAuthor(book.ID)
	if err != nil {
		t.Fatalf("Failed to get book with author: %v", err)
	}

	if retrieved.Title != "Test Book" {
		t.Errorf("Expected book title 'Test Book', got %s", retrieved.Title)
	}

	if retrieved.Author.Name != "Jane Smith" {
		t.Errorf("Expected author name 'Jane Smith', got %s", retrieved.Author.Name)
	}
}

func TestCreateTag(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	tag := &Tag{Name: "Programming"}

	err := db.CreateTag(tag)
	if err != nil {
		t.Fatalf("Failed to create tag: %v", err)
	}

	if tag.ID == 0 {
		t.Error("Expected ID to be set")
	}
}

func TestAddTagToBook(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	author := &Author{Name: "Author", Email: "author@example.com"}
	db.CreateAuthor(author)

	book := &Book{Title: "Book", ISBN: "789", AuthorID: author.ID}
	db.CreateBook(book)

	tag := &Tag{Name: "Fiction"}
	db.CreateTag(tag)

	err := db.AddTagToBook(book.ID, tag.ID)
	if err != nil {
		t.Fatalf("Failed to add tag to book: %v", err)
	}

	// Verify the association
	retrieved, _ := db.GetBookWithTags(book.ID)
	if len(retrieved.Tags) != 1 {
		t.Errorf("Expected 1 tag, got %d", len(retrieved.Tags))
	}

	if retrieved.Tags[0].Name != "Fiction" {
		t.Errorf("Expected tag 'Fiction', got %s", retrieved.Tags[0].Name)
	}
}

func TestGetBookWithTags(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	author := &Author{Name: "Author", Email: "author@example.com"}
	db.CreateAuthor(author)

	book := &Book{Title: "Multi-tag Book", ISBN: "999", AuthorID: author.ID}
	db.CreateBook(book)

	tag1 := &Tag{Name: "Go"}
	tag2 := &Tag{Name: "Programming"}
	tag3 := &Tag{Name: "Tutorial"}

	db.CreateTag(tag1)
	db.CreateTag(tag2)
	db.CreateTag(tag3)

	db.AddTagToBook(book.ID, tag1.ID)
	db.AddTagToBook(book.ID, tag2.ID)
	db.AddTagToBook(book.ID, tag3.ID)

	retrieved, err := db.GetBookWithTags(book.ID)
	if err != nil {
		t.Fatalf("Failed to get book with tags: %v", err)
	}

	if len(retrieved.Tags) != 3 {
		t.Errorf("Expected 3 tags, got %d", len(retrieved.Tags))
	}
}

func TestGetBookWithAll(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	author := &Author{Name: "Complete Author", Email: "complete@example.com"}
	db.CreateAuthor(author)

	book := &Book{Title: "Complete Book", ISBN: "000", AuthorID: author.ID}
	db.CreateBook(book)

	tag := &Tag{Name: "Complete"}
	db.CreateTag(tag)
	db.AddTagToBook(book.ID, tag.ID)

	retrieved, err := db.GetBookWithAll(book.ID)
	if err != nil {
		t.Fatalf("Failed to get book with all associations: %v", err)
	}

	if retrieved.Title != "Complete Book" {
		t.Error("Book not loaded correctly")
	}

	if retrieved.Author.Name != "Complete Author" {
		t.Error("Author not loaded correctly")
	}

	if len(retrieved.Tags) != 1 {
		t.Error("Tags not loaded correctly")
	}
}

func TestGetBooksByAuthor(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	author1 := &Author{Name: "Author 1", Email: "author1@example.com"}
	author2 := &Author{Name: "Author 2", Email: "author2@example.com"}
	db.CreateAuthor(author1)
	db.CreateAuthor(author2)

	db.CreateBook(&Book{Title: "Book A1", ISBN: "A1", AuthorID: author1.ID})
	db.CreateBook(&Book{Title: "Book A2", ISBN: "A2", AuthorID: author1.ID})
	db.CreateBook(&Book{Title: "Book B1", ISBN: "B1", AuthorID: author2.ID})

	books, err := db.GetBooksByAuthor(author1.ID)
	if err != nil {
		t.Fatalf("Failed to get books by author: %v", err)
	}

	if len(books) != 2 {
		t.Errorf("Expected 2 books for author 1, got %d", len(books))
	}
}

func TestGetBooksByTag(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	author := &Author{Name: "Author", Email: "author@example.com"}
	db.CreateAuthor(author)

	book1 := &Book{Title: "Book 1", ISBN: "B1", AuthorID: author.ID}
	book2 := &Book{Title: "Book 2", ISBN: "B2", AuthorID: author.ID}
	book3 := &Book{Title: "Book 3", ISBN: "B3", AuthorID: author.ID}

	db.CreateBook(book1)
	db.CreateBook(book2)
	db.CreateBook(book3)

	tag := &Tag{Name: "Popular"}
	db.CreateTag(tag)

	db.AddTagToBook(book1.ID, tag.ID)
	db.AddTagToBook(book2.ID, tag.ID)

	books, err := db.GetBooksByTag(tag.ID)
	if err != nil {
		t.Fatalf("Failed to get books by tag: %v", err)
	}

	if len(books) != 2 {
		t.Errorf("Expected 2 books with tag, got %d", len(books))
	}
}

func TestRemoveTagFromBook(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	author := &Author{Name: "Author", Email: "author@example.com"}
	db.CreateAuthor(author)

	book := &Book{Title: "Book", ISBN: "REM", AuthorID: author.ID}
	db.CreateBook(book)

	tag := &Tag{Name: "ToRemove"}
	db.CreateTag(tag)

	db.AddTagToBook(book.ID, tag.ID)

	// Verify tag was added
	retrieved, _ := db.GetBookWithTags(book.ID)
	if len(retrieved.Tags) != 1 {
		t.Fatalf("Expected 1 tag before removal")
	}

	// Remove tag
	err := db.RemoveTagFromBook(book.ID, tag.ID)
	if err != nil {
		t.Fatalf("Failed to remove tag from book: %v", err)
	}

	// Verify tag was removed
	retrieved, _ = db.GetBookWithTags(book.ID)
	if len(retrieved.Tags) != 0 {
		t.Errorf("Expected 0 tags after removal, got %d", len(retrieved.Tags))
	}
}

func TestCreateCompany(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	company := &Company{
		Name:    "Test Company",
		Address: "123 Main St",
	}

	err := db.CreateCompany(company)
	if err != nil {
		t.Fatalf("Failed to create company: %v", err)
	}

	if company.ID == 0 {
		t.Error("Expected ID to be set")
	}
}

func TestCreateUser(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	company := &Company{Name: "Company"}
	db.CreateCompany(company)

	user := &User{
		Name:      "Test User",
		Email:     "user@example.com",
		CompanyID: &company.ID,
	}

	err := db.CreateUser(user)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	if user.ID == 0 {
		t.Error("Expected ID to be set")
	}
}

func TestGetCompanyWithUsers(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	company := &Company{Name: "Tech Corp"}
	db.CreateCompany(company)

	db.CreateUser(&User{Name: "User 1", Email: "user1@example.com", CompanyID: &company.ID})
	db.CreateUser(&User{Name: "User 2", Email: "user2@example.com", CompanyID: &company.ID})
	db.CreateUser(&User{Name: "User 3", Email: "user3@example.com", CompanyID: &company.ID})

	retrieved, err := db.GetCompanyWithUsers(company.ID)
	if err != nil {
		t.Fatalf("Failed to get company with users: %v", err)
	}

	if len(retrieved.Users) != 3 {
		t.Errorf("Expected 3 users, got %d", len(retrieved.Users))
	}
}

func TestGetUserWithCompany(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	company := &Company{Name: "Business Inc"}
	db.CreateCompany(company)

	user := &User{Name: "Employee", Email: "employee@example.com", CompanyID: &company.ID}
	db.CreateUser(user)

	retrieved, err := db.GetUserWithCompany(user.ID)
	if err != nil {
		t.Fatalf("Failed to get user with company: %v", err)
	}

	if retrieved.Company == nil {
		t.Fatal("Expected company to be loaded")
	}

	if retrieved.Company.Name != "Business Inc" {
		t.Errorf("Expected company 'Business Inc', got %s", retrieved.Company.Name)
	}
}

func TestCreateProject(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	project := &Project{
		Name:        "Test Project",
		Description: "Project Description",
	}

	err := db.CreateProject(project)
	if err != nil {
		t.Fatalf("Failed to create project: %v", err)
	}

	if project.ID == 0 {
		t.Error("Expected ID to be set")
	}
}

func TestAddUserToProject(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	user := &User{Name: "Developer", Email: "dev@example.com"}
	db.CreateUser(user)

	project := &Project{Name: "Project Alpha"}
	db.CreateProject(project)

	err := db.AddUserToProject(user.ID, project.ID)
	if err != nil {
		t.Fatalf("Failed to add user to project: %v", err)
	}

	// Verify association
	retrieved, _ := db.GetProjectWithUsers(project.ID)
	if len(retrieved.Users) != 1 {
		t.Errorf("Expected 1 user, got %d", len(retrieved.Users))
	}
}

func TestGetProjectWithUsers(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	project := &Project{Name: "Team Project"}
	db.CreateProject(project)

	user1 := &User{Name: "User 1", Email: "u1@example.com"}
	user2 := &User{Name: "User 2", Email: "u2@example.com"}
	user3 := &User{Name: "User 3", Email: "u3@example.com"}

	db.CreateUser(user1)
	db.CreateUser(user2)
	db.CreateUser(user3)

	db.AddUserToProject(user1.ID, project.ID)
	db.AddUserToProject(user2.ID, project.ID)
	db.AddUserToProject(user3.ID, project.ID)

	retrieved, err := db.GetProjectWithUsers(project.ID)
	if err != nil {
		t.Fatalf("Failed to get project with users: %v", err)
	}

	if len(retrieved.Users) != 3 {
		t.Errorf("Expected 3 users, got %d", len(retrieved.Users))
	}
}

func TestGetUserWithProjects(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	user := &User{Name: "Multi-tasker", Email: "multi@example.com"}
	db.CreateUser(user)

	project1 := &Project{Name: "Project 1"}
	project2 := &Project{Name: "Project 2"}

	db.CreateProject(project1)
	db.CreateProject(project2)

	db.AddUserToProject(user.ID, project1.ID)
	db.AddUserToProject(user.ID, project2.ID)

	retrieved, err := db.GetUserWithProjects(user.ID)
	if err != nil {
		t.Fatalf("Failed to get user with projects: %v", err)
	}

	if len(retrieved.Projects) != 2 {
		t.Errorf("Expected 2 projects, got %d", len(retrieved.Projects))
	}
}

func TestCountBooksByAuthor(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	author := &Author{Name: "Prolific Writer", Email: "writer@example.com"}
	db.CreateAuthor(author)

	for i := 0; i < 5; i++ {
		db.CreateBook(&Book{
			Title:    "Book",
			ISBN:     string(rune('A' + i)),
			AuthorID: author.ID,
		})
	}

	count, err := db.CountBooksByAuthor(author.ID)
	if err != nil {
		t.Fatalf("Failed to count books: %v", err)
	}

	if count != 5 {
		t.Errorf("Expected 5 books, got %d", count)
	}
}

func TestCountUsersByCompany(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	company := &Company{Name: "Big Corp"}
	db.CreateCompany(company)

	for i := 0; i < 7; i++ {
		db.CreateUser(&User{
			Name:      "User",
			Email:     string(rune('a'+i)) + "@example.com",
			CompanyID: &company.ID,
		})
	}

	count, err := db.CountUsersByCompany(company.ID)
	if err != nil {
		t.Fatalf("Failed to count users: %v", err)
	}

	if count != 7 {
		t.Errorf("Expected 7 users, got %d", count)
	}
}
