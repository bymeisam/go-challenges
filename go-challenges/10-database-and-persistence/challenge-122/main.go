package main

import (
	"fmt"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type User struct {
	ID        uint           `gorm:"primaryKey"`
	CreatedAt time.Time      `gorm:"autoCreateTime"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime"`
	DeletedAt gorm.DeletedAt `gorm:"index"`
	Name      string         `gorm:"type:varchar(100);not null"`
	Email     string         `gorm:"type:varchar(100);uniqueIndex;not null"`
	Age       int            `gorm:"type:int"`
	Active    bool           `gorm:"default:true"`
}

type Product struct {
	ID          uint           `gorm:"primaryKey"`
	CreatedAt   time.Time      `gorm:"autoCreateTime"`
	UpdatedAt   time.Time      `gorm:"autoUpdateTime"`
	DeletedAt   gorm.DeletedAt `gorm:"index"`
	Name        string         `gorm:"type:varchar(200);not null"`
	Description string         `gorm:"type:text"`
	Price       float64        `gorm:"type:decimal(10,2);not null"`
	Stock       int            `gorm:"type:int;default:0"`
}

type Database struct {
	db *gorm.DB
}

// NewDatabase creates a new GORM database connection
func NewDatabase(dsn string) (*Database, error) {
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, err
	}

	// Auto migrate the schemas
	if err := db.AutoMigrate(&User{}, &Product{}); err != nil {
		return nil, err
	}

	return &Database{db: db}, nil
}

// Close closes the database connection
func (d *Database) Close() error {
	sqlDB, err := d.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// CreateUser creates a new user
func (d *Database) CreateUser(user *User) error {
	return d.db.Create(user).Error
}

// GetUser retrieves a user by ID
func (d *Database) GetUser(id uint) (*User, error) {
	var user User
	result := d.db.First(&user, id)
	if result.Error != nil {
		return nil, result.Error
	}
	return &user, nil
}

// GetUserByEmail retrieves a user by email
func (d *Database) GetUserByEmail(email string) (*User, error) {
	var user User
	result := d.db.Where("email = ?", email).First(&user)
	if result.Error != nil {
		return nil, result.Error
	}
	return &user, nil
}

// GetAllUsers retrieves all users
func (d *Database) GetAllUsers() ([]User, error) {
	var users []User
	result := d.db.Find(&users)
	if result.Error != nil {
		return nil, result.Error
	}
	return users, nil
}

// GetActiveUsers retrieves all active users
func (d *Database) GetActiveUsers() ([]User, error) {
	var users []User
	result := d.db.Where("active = ?", true).Find(&users)
	if result.Error != nil {
		return nil, result.Error
	}
	return users, nil
}

// UpdateUser updates a user
func (d *Database) UpdateUser(user *User) error {
	return d.db.Save(user).Error
}

// UpdateUserFields updates specific fields of a user
func (d *Database) UpdateUserFields(id uint, updates map[string]interface{}) error {
	return d.db.Model(&User{}).Where("id = ?", id).Updates(updates).Error
}

// DeleteUser soft deletes a user
func (d *Database) DeleteUser(id uint) error {
	return d.db.Delete(&User{}, id).Error
}

// HardDeleteUser permanently deletes a user
func (d *Database) HardDeleteUser(id uint) error {
	return d.db.Unscoped().Delete(&User{}, id).Error
}

// CreateProduct creates a new product
func (d *Database) CreateProduct(product *Product) error {
	return d.db.Create(product).Error
}

// GetProduct retrieves a product by ID
func (d *Database) GetProduct(id uint) (*Product, error) {
	var product Product
	result := d.db.First(&product, id)
	if result.Error != nil {
		return nil, result.Error
	}
	return &product, nil
}

// GetProductsByPriceRange retrieves products within a price range
func (d *Database) GetProductsByPriceRange(minPrice, maxPrice float64) ([]Product, error) {
	var products []Product
	result := d.db.Where("price BETWEEN ? AND ?", minPrice, maxPrice).Find(&products)
	if result.Error != nil {
		return nil, result.Error
	}
	return products, nil
}

// SearchProducts searches products by name
func (d *Database) SearchProducts(query string) ([]Product, error) {
	var products []Product
	result := d.db.Where("name LIKE ?", "%"+query+"%").Find(&products)
	if result.Error != nil {
		return nil, result.Error
	}
	return products, nil
}

// UpdateProductStock updates product stock
func (d *Database) UpdateProductStock(id uint, stock int) error {
	return d.db.Model(&Product{}).Where("id = ?", id).Update("stock", stock).Error
}

// DeleteProduct soft deletes a product
func (d *Database) DeleteProduct(id uint) error {
	return d.db.Delete(&Product{}, id).Error
}

// BatchCreateUsers creates multiple users in a batch
func (d *Database) BatchCreateUsers(users []User) error {
	return d.db.Create(&users).Error
}

// CountUsers counts all users
func (d *Database) CountUsers() (int64, error) {
	var count int64
	result := d.db.Model(&User{}).Count(&count)
	return count, result.Error
}

// GetUsersPaginated retrieves users with pagination
func (d *Database) GetUsersPaginated(page, pageSize int) ([]User, error) {
	var users []User
	offset := (page - 1) * pageSize
	result := d.db.Offset(offset).Limit(pageSize).Find(&users)
	if result.Error != nil {
		return nil, result.Error
	}
	return users, nil
}

// GetUsersOrderedByAge retrieves users ordered by age
func (d *Database) GetUsersOrderedByAge(ascending bool) ([]User, error) {
	var users []User
	order := "age DESC"
	if ascending {
		order = "age ASC"
	}
	result := d.db.Order(order).Find(&users)
	if result.Error != nil {
		return nil, result.Error
	}
	return users, nil
}

func main() {
	db, err := NewDatabase(":memory:")
	if err != nil {
		fmt.Printf("Error creating database: %v\n", err)
		return
	}
	defer db.Close()

	// Create a user
	user := &User{
		Name:   "John Doe",
		Email:  "john@example.com",
		Age:    30,
		Active: true,
	}

	err = db.CreateUser(user)
	if err != nil {
		fmt.Printf("Error creating user: %v\n", err)
		return
	}
	fmt.Printf("Created user: %+v\n", user)

	// Get user
	retrieved, _ := db.GetUser(user.ID)
	fmt.Printf("Retrieved user: %+v\n", retrieved)

	// Update user
	user.Age = 31
	db.UpdateUser(user)
	fmt.Println("User updated")

	// Create a product
	product := &Product{
		Name:        "Laptop",
		Description: "High-performance laptop",
		Price:       999.99,
		Stock:       10,
	}

	db.CreateProduct(product)
	fmt.Printf("Created product: %+v\n", product)

	// Search products
	products, _ := db.SearchProducts("Laptop")
	fmt.Printf("Found products: %+v\n", products)
}
