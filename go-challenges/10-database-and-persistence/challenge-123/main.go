package main

import (
	"fmt"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Author has many Books (one-to-many)
type Author struct {
	ID    uint   `gorm:"primaryKey"`
	Name  string `gorm:"type:varchar(100);not null"`
	Email string `gorm:"type:varchar(100);uniqueIndex"`
	Books []Book `gorm:"foreignKey:AuthorID"`
}

// Book belongs to Author and has many Tags (many-to-many)
type Book struct {
	ID       uint   `gorm:"primaryKey"`
	Title    string `gorm:"type:varchar(200);not null"`
	ISBN     string `gorm:"type:varchar(20);uniqueIndex"`
	AuthorID uint   `gorm:"not null"`
	Author   Author `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Tags     []Tag  `gorm:"many2many:book_tags;"`
}

// Tag can belong to many Books (many-to-many)
type Tag struct {
	ID    uint   `gorm:"primaryKey"`
	Name  string `gorm:"type:varchar(50);uniqueIndex;not null"`
	Books []Book `gorm:"many2many:book_tags;"`
}

// Company has many Users (one-to-many)
type Company struct {
	ID      uint   `gorm:"primaryKey"`
	Name    string `gorm:"type:varchar(100);not null"`
	Address string `gorm:"type:varchar(200)"`
	Users   []User `gorm:"foreignKey:CompanyID"`
}

// User belongs to Company and has many Projects (many-to-many)
type User struct {
	ID        uint      `gorm:"primaryKey"`
	Name      string    `gorm:"type:varchar(100);not null"`
	Email     string    `gorm:"type:varchar(100);uniqueIndex;not null"`
	CompanyID *uint     `gorm:"index"`
	Company   *Company  `gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	Projects  []Project `gorm:"many2many:user_projects;"`
}

// Project can have many Users (many-to-many)
type Project struct {
	ID          uint   `gorm:"primaryKey"`
	Name        string `gorm:"type:varchar(100);not null"`
	Description string `gorm:"type:text"`
	Users       []User `gorm:"many2many:user_projects;"`
}

type Database struct {
	db *gorm.DB
}

func NewDatabase(dsn string) (*Database, error) {
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, err
	}

	// Auto migrate all models
	if err := db.AutoMigrate(&Author{}, &Book{}, &Tag{}, &Company{}, &User{}, &Project{}); err != nil {
		return nil, err
	}

	return &Database{db: db}, nil
}

func (d *Database) Close() error {
	sqlDB, err := d.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// CreateAuthor creates an author
func (d *Database) CreateAuthor(author *Author) error {
	return d.db.Create(author).Error
}

// CreateBook creates a book with author relationship
func (d *Database) CreateBook(book *Book) error {
	return d.db.Create(book).Error
}

// CreateTag creates a tag
func (d *Database) CreateTag(tag *Tag) error {
	return d.db.Create(tag).Error
}

// AddTagToBook adds a tag to a book (many-to-many)
func (d *Database) AddTagToBook(bookID, tagID uint) error {
	var book Book
	if err := d.db.First(&book, bookID).Error; err != nil {
		return err
	}

	var tag Tag
	if err := d.db.First(&tag, tagID).Error; err != nil {
		return err
	}

	return d.db.Model(&book).Association("Tags").Append(&tag)
}

// GetAuthorWithBooks retrieves an author with all their books
func (d *Database) GetAuthorWithBooks(id uint) (*Author, error) {
	var author Author
	err := d.db.Preload("Books").First(&author, id).Error
	if err != nil {
		return nil, err
	}
	return &author, nil
}

// GetBookWithAuthor retrieves a book with its author
func (d *Database) GetBookWithAuthor(id uint) (*Book, error) {
	var book Book
	err := d.db.Preload("Author").First(&book, id).Error
	if err != nil {
		return nil, err
	}
	return &book, nil
}

// GetBookWithTags retrieves a book with all its tags
func (d *Database) GetBookWithTags(id uint) (*Book, error) {
	var book Book
	err := d.db.Preload("Tags").First(&book, id).Error
	if err != nil {
		return nil, err
	}
	return &book, nil
}

// GetBookWithAll retrieves a book with author and tags
func (d *Database) GetBookWithAll(id uint) (*Book, error) {
	var book Book
	err := d.db.Preload("Author").Preload("Tags").First(&book, id).Error
	if err != nil {
		return nil, err
	}
	return &book, nil
}

// GetBooksByAuthor retrieves all books by a specific author
func (d *Database) GetBooksByAuthor(authorID uint) ([]Book, error) {
	var books []Book
	err := d.db.Where("author_id = ?", authorID).Find(&books).Error
	return books, err
}

// GetBooksByTag retrieves all books with a specific tag
func (d *Database) GetBooksByTag(tagID uint) ([]Book, error) {
	var tag Tag
	err := d.db.Preload("Books").First(&tag, tagID).Error
	if err != nil {
		return nil, err
	}
	return tag.Books, nil
}

// RemoveTagFromBook removes a tag from a book
func (d *Database) RemoveTagFromBook(bookID, tagID uint) error {
	var book Book
	if err := d.db.First(&book, bookID).Error; err != nil {
		return err
	}

	var tag Tag
	if err := d.db.First(&tag, tagID).Error; err != nil {
		return err
	}

	return d.db.Model(&book).Association("Tags").Delete(&tag)
}

// CreateCompany creates a company
func (d *Database) CreateCompany(company *Company) error {
	return d.db.Create(company).Error
}

// CreateUser creates a user with company relationship
func (d *Database) CreateUser(user *User) error {
	return d.db.Create(user).Error
}

// CreateProject creates a project
func (d *Database) CreateProject(project *Project) error {
	return d.db.Create(project).Error
}

// AddUserToProject adds a user to a project (many-to-many)
func (d *Database) AddUserToProject(userID, projectID uint) error {
	var project Project
	if err := d.db.First(&project, projectID).Error; err != nil {
		return err
	}

	var user User
	if err := d.db.First(&user, userID).Error; err != nil {
		return err
	}

	return d.db.Model(&project).Association("Users").Append(&user)
}

// GetCompanyWithUsers retrieves a company with all its users
func (d *Database) GetCompanyWithUsers(id uint) (*Company, error) {
	var company Company
	err := d.db.Preload("Users").First(&company, id).Error
	if err != nil {
		return nil, err
	}
	return &company, nil
}

// GetUserWithCompany retrieves a user with their company
func (d *Database) GetUserWithCompany(id uint) (*User, error) {
	var user User
	err := d.db.Preload("Company").First(&user, id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetProjectWithUsers retrieves a project with all assigned users
func (d *Database) GetProjectWithUsers(id uint) (*Project, error) {
	var project Project
	err := d.db.Preload("Users").First(&project, id).Error
	if err != nil {
		return nil, err
	}
	return &project, nil
}

// GetUserWithProjects retrieves a user with all their projects
func (d *Database) GetUserWithProjects(id uint) (*User, error) {
	var user User
	err := d.db.Preload("Projects").First(&user, id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// CountBooksByAuthor counts books for a specific author
func (d *Database) CountBooksByAuthor(authorID uint) (int64, error) {
	var count int64
	err := d.db.Model(&Book{}).Where("author_id = ?", authorID).Count(&count).Error
	return count, err
}

// CountUsersByCompany counts users for a specific company
func (d *Database) CountUsersByCompany(companyID uint) (int64, error) {
	var count int64
	err := d.db.Model(&User{}).Where("company_id = ?", companyID).Count(&count).Error
	return count, err
}

func main() {
	db, err := NewDatabase(":memory:")
	if err != nil {
		fmt.Printf("Error creating database: %v\n", err)
		return
	}
	defer db.Close()

	// Create author
	author := &Author{
		Name:  "John Smith",
		Email: "john@example.com",
	}
	db.CreateAuthor(author)

	// Create books for the author
	book1 := &Book{
		Title:    "Go Programming",
		ISBN:     "123-456",
		AuthorID: author.ID,
	}
	db.CreateBook(book1)

	// Create tags
	tag1 := &Tag{Name: "Programming"}
	tag2 := &Tag{Name: "Go"}
	db.CreateTag(tag1)
	db.CreateTag(tag2)

	// Add tags to book
	db.AddTagToBook(book1.ID, tag1.ID)
	db.AddTagToBook(book1.ID, tag2.ID)

	// Retrieve with associations
	authorWithBooks, _ := db.GetAuthorWithBooks(author.ID)
	fmt.Printf("Author: %s, Books: %d\n", authorWithBooks.Name, len(authorWithBooks.Books))

	bookWithAll, _ := db.GetBookWithAll(book1.ID)
	fmt.Printf("Book: %s, Author: %s, Tags: %d\n",
		bookWithAll.Title, bookWithAll.Author.Name, len(bookWithAll.Tags))
}
