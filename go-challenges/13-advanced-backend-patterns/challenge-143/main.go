package main

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"time"
)

// Enums
type Category int32

const (
	CategoryUnspecified Category = 0
	Electronics         Category = 1
	Clothing            Category = 2
	Books               Category = 3
	Home                Category = 4
	Sports              Category = 5
)

func (c Category) String() string {
	names := []string{"UNSPECIFIED", "ELECTRONICS", "CLOTHING", "BOOKS", "HOME", "SPORTS"}
	if int(c) < len(names) {
		return names[c]
	}
	return "UNKNOWN"
}

type OrderStatus int32

const (
	OrderStatusUnspecified OrderStatus = 0
	Pending                OrderStatus = 1
	Confirmed              OrderStatus = 2
	Shipped                OrderStatus = 3
	Delivered              OrderStatus = 4
	Cancelled              OrderStatus = 5
)

// Messages
type Price struct {
	Amount   int64
	Currency string
}

type StockLocation struct {
	LocationID string
	Address    string
	Quantity   int32
}

type Inventory struct {
	Quantity    int32
	WarehouseID string
	Locations   []*StockLocation
}

type Review struct {
	ID        string
	UserID    string
	Rating    int32
	Comment   string
	CreatedAt time.Time
}

type PercentageDiscount struct {
	Percentage float64
}

type FixedDiscount struct {
	Amount int64
}

// Oneof interface for discount
type isProduct_Discount interface {
	isProduct_Discount()
}

type Product_PercentageDiscount struct {
	PercentageDiscount *PercentageDiscount
}

type Product_FixedDiscount struct {
	FixedDiscount *FixedDiscount
}

func (*Product_PercentageDiscount) isProduct_Discount() {}
func (*Product_FixedDiscount) isProduct_Discount()      {}

type Product struct {
	ID          string
	Name        string
	Description string
	Price       *Price
	Category    Category
	Tags        []string
	Attributes  map[string]string
	Inventory   *Inventory
	Reviews     []*Review
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Discount    isProduct_Discount
}

type Address struct {
	Street  string
	City    string
	State   string
	ZipCode string
	Country string
}

type OrderItem struct {
	ProductID string
	Quantity  int32
	UnitPrice *Price
}

type Order struct {
	ID              string
	CustomerID      string
	Items           []*OrderItem
	Status          OrderStatus
	ShippingAddress *Address
	BillingAddress  *Address
	Total           *Price
	CreatedAt       time.Time
	Metadata        map[string]string
}

type ProductBatch struct {
	Products []*Product
}

type OrderBatch struct {
	Orders []*Order
}

// Binary Serializer
type BinarySerializer struct{}

func NewBinarySerializer() *BinarySerializer {
	return &BinarySerializer{}
}

// Simplified binary encoding (varint + length-delimited)
func (s *BinarySerializer) writeVarint(w io.Writer, v int64) error {
	buf := make([]byte, binary.MaxVarintLen64)
	n := binary.PutVarint(buf, v)
	_, err := w.Write(buf[:n])
	return err
}

func (s *BinarySerializer) readVarint(r io.Reader) (int64, error) {
	var result int64
	var shift uint
	for {
		var b [1]byte
		_, err := r.Read(b[:])
		if err != nil {
			return 0, err
		}
		result |= int64(b[0]&0x7f) << shift
		if b[0]&0x80 == 0 {
			break
		}
		shift += 7
	}
	return result, nil
}

func (s *BinarySerializer) writeString(w io.Writer, str string) error {
	if err := s.writeVarint(w, int64(len(str))); err != nil {
		return err
	}
	_, err := w.Write([]byte(str))
	return err
}

func (s *BinarySerializer) readString(r io.Reader) (string, error) {
	length, err := s.readVarint(r)
	if err != nil {
		return "", err
	}
	buf := make([]byte, length)
	_, err = io.ReadFull(r, buf)
	if err != nil {
		return "", err
	}
	return string(buf), nil
}

// Product serialization
func (s *BinarySerializer) MarshalProduct(p *Product) ([]byte, error) {
	buf := &byteBuffer{}

	// Field 1: ID
	s.writeString(buf, p.ID)

	// Field 2: Name
	s.writeString(buf, p.Name)

	// Field 3: Description
	s.writeString(buf, p.Description)

	// Field 4: Price
	if p.Price != nil {
		s.writeVarint(buf, p.Price.Amount)
		s.writeString(buf, p.Price.Currency)
	}

	// Field 5: Category
	s.writeVarint(buf, int64(p.Category))

	// Field 6: Tags (repeated)
	s.writeVarint(buf, int64(len(p.Tags)))
	for _, tag := range p.Tags {
		s.writeString(buf, tag)
	}

	// Field 7: Attributes (map)
	s.writeVarint(buf, int64(len(p.Attributes)))
	for k, v := range p.Attributes {
		s.writeString(buf, k)
		s.writeString(buf, v)
	}

	// Field 8: Inventory
	if p.Inventory != nil {
		s.writeVarint(buf, int64(p.Inventory.Quantity))
		s.writeString(buf, p.Inventory.WarehouseID)
	}

	// Timestamps
	s.writeVarint(buf, p.CreatedAt.Unix())
	s.writeVarint(buf, p.UpdatedAt.Unix())

	return buf.Bytes(), nil
}

func (s *BinarySerializer) UnmarshalProduct(data []byte, p *Product) error {
	buf := &byteBuffer{data: data}

	// Field 1: ID
	id, err := s.readString(buf)
	if err != nil {
		return err
	}
	p.ID = id

	// Field 2: Name
	name, err := s.readString(buf)
	if err != nil {
		return err
	}
	p.Name = name

	// Field 3: Description
	desc, err := s.readString(buf)
	if err != nil {
		return err
	}
	p.Description = desc

	// Field 4: Price
	amount, err := s.readVarint(buf)
	if err != nil {
		return err
	}
	currency, err := s.readString(buf)
	if err != nil {
		return err
	}
	p.Price = &Price{Amount: amount, Currency: currency}

	// Field 5: Category
	cat, err := s.readVarint(buf)
	if err != nil {
		return err
	}
	p.Category = Category(cat)

	// Field 6: Tags
	tagCount, err := s.readVarint(buf)
	if err != nil {
		return err
	}
	p.Tags = make([]string, tagCount)
	for i := int64(0); i < tagCount; i++ {
		tag, err := s.readString(buf)
		if err != nil {
			return err
		}
		p.Tags[i] = tag
	}

	// Field 7: Attributes
	attrCount, err := s.readVarint(buf)
	if err != nil {
		return err
	}
	p.Attributes = make(map[string]string)
	for i := int64(0); i < attrCount; i++ {
		k, err := s.readString(buf)
		if err != nil {
			return err
		}
		v, err := s.readString(buf)
		if err != nil {
			return err
		}
		p.Attributes[k] = v
	}

	// Field 8: Inventory
	qty, err := s.readVarint(buf)
	if err != nil {
		return err
	}
	warehouseID, err := s.readString(buf)
	if err != nil {
		return err
	}
	p.Inventory = &Inventory{Quantity: int32(qty), WarehouseID: warehouseID}

	// Timestamps
	createdAt, err := s.readVarint(buf)
	if err != nil {
		return err
	}
	p.CreatedAt = time.Unix(createdAt, 0)

	updatedAt, err := s.readVarint(buf)
	if err != nil {
		return err
	}
	p.UpdatedAt = time.Unix(updatedAt, 0)

	return nil
}

// JSON Interoperability
type JSONSerializer struct{}

func NewJSONSerializer() *JSONSerializer {
	return &JSONSerializer{}
}

func (s *JSONSerializer) MarshalProduct(p *Product) ([]byte, error) {
	// Convert to JSON-friendly struct
	type jsonProduct struct {
		ID          string            `json:"id"`
		Name        string            `json:"name"`
		Description string            `json:"description"`
		Price       *Price            `json:"price,omitempty"`
		Category    string            `json:"category"`
		Tags        []string          `json:"tags,omitempty"`
		Attributes  map[string]string `json:"attributes,omitempty"`
		Inventory   *Inventory        `json:"inventory,omitempty"`
		Reviews     []*Review         `json:"reviews,omitempty"`
		CreatedAt   string            `json:"created_at"`
		UpdatedAt   string            `json:"updated_at"`
	}

	jp := jsonProduct{
		ID:          p.ID,
		Name:        p.Name,
		Description: p.Description,
		Price:       p.Price,
		Category:    p.Category.String(),
		Tags:        p.Tags,
		Attributes:  p.Attributes,
		Inventory:   p.Inventory,
		Reviews:     p.Reviews,
		CreatedAt:   p.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   p.UpdatedAt.Format(time.RFC3339),
	}

	return json.Marshal(jp)
}

func (s *JSONSerializer) UnmarshalProduct(data []byte, p *Product) error {
	type jsonProduct struct {
		ID          string            `json:"id"`
		Name        string            `json:"name"`
		Description string            `json:"description"`
		Price       *Price            `json:"price,omitempty"`
		Category    string            `json:"category"`
		Tags        []string          `json:"tags,omitempty"`
		Attributes  map[string]string `json:"attributes,omitempty"`
		Inventory   *Inventory        `json:"inventory,omitempty"`
		Reviews     []*Review         `json:"reviews,omitempty"`
		CreatedAt   string            `json:"created_at"`
		UpdatedAt   string            `json:"updated_at"`
	}

	var jp jsonProduct
	if err := json.Unmarshal(data, &jp); err != nil {
		return err
	}

	p.ID = jp.ID
	p.Name = jp.Name
	p.Description = jp.Description
	p.Price = jp.Price
	p.Tags = jp.Tags
	p.Attributes = jp.Attributes
	p.Inventory = jp.Inventory
	p.Reviews = jp.Reviews

	// Parse category
	categories := map[string]Category{
		"ELECTRONICS": Electronics,
		"CLOTHING":    Clothing,
		"BOOKS":       Books,
		"HOME":        Home,
		"SPORTS":      Sports,
	}
	if cat, ok := categories[jp.Category]; ok {
		p.Category = cat
	}

	// Parse timestamps
	createdAt, err := time.Parse(time.RFC3339, jp.CreatedAt)
	if err != nil {
		return err
	}
	p.CreatedAt = createdAt

	updatedAt, err := time.Parse(time.RFC3339, jp.UpdatedAt)
	if err != nil {
		return err
	}
	p.UpdatedAt = updatedAt

	return nil
}

// Validator
type Validator struct{}

func NewValidator() *Validator {
	return &Validator{}
}

func (v *Validator) ValidateProduct(p *Product) error {
	if p.ID == "" {
		return errors.New("product ID is required")
	}
	if p.Name == "" {
		return errors.New("product name is required")
	}
	if p.Price == nil {
		return errors.New("product price is required")
	}
	if p.Price.Amount < 0 {
		return errors.New("product price must be non-negative")
	}
	if p.Price.Currency == "" {
		return errors.New("product currency is required")
	}
	if p.Category == CategoryUnspecified {
		return errors.New("product category is required")
	}
	for _, review := range p.Reviews {
		if review.Rating < 1 || review.Rating > 5 {
			return errors.New("review rating must be between 1 and 5")
		}
	}
	return nil
}

func (v *Validator) ValidateOrder(o *Order) error {
	if o.ID == "" {
		return errors.New("order ID is required")
	}
	if o.CustomerID == "" {
		return errors.New("customer ID is required")
	}
	if len(o.Items) == 0 {
		return errors.New("order must have at least one item")
	}
	if o.Total == nil {
		return errors.New("order total is required")
	}
	if o.ShippingAddress == nil {
		return errors.New("shipping address is required")
	}
	return nil
}

// Helper: simple byte buffer
type byteBuffer struct {
	data []byte
	pos  int
}

func (b *byteBuffer) Write(p []byte) (n int, err error) {
	b.data = append(b.data, p...)
	return len(p), nil
}

func (b *byteBuffer) Read(p []byte) (n int, err error) {
	if b.pos >= len(b.data) {
		return 0, io.EOF
	}
	n = copy(p, b.data[b.pos:])
	b.pos += n
	return n, nil
}

func (b *byteBuffer) Bytes() []byte {
	return b.data
}

func main() {
	// Create a sample product
	product := &Product{
		ID:          "prod-123",
		Name:        "Laptop",
		Description: "High-performance laptop",
		Price: &Price{
			Amount:   99900,
			Currency: "USD",
		},
		Category: Electronics,
		Tags:     []string{"computer", "portable", "work"},
		Attributes: map[string]string{
			"brand": "TechCorp",
			"model": "X1000",
			"color": "Silver",
		},
		Inventory: &Inventory{
			Quantity:    50,
			WarehouseID: "WH-001",
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Validate
	validator := NewValidator()
	if err := validator.ValidateProduct(product); err != nil {
		log.Fatalf("Validation failed: %v", err)
	}

	// Binary serialization
	binarySerializer := NewBinarySerializer()
	binaryData, err := binarySerializer.MarshalProduct(product)
	if err != nil {
		log.Fatalf("Binary marshal failed: %v", err)
	}
	fmt.Printf("Binary size: %d bytes\n", len(binaryData))

	var decodedProduct Product
	if err := binarySerializer.UnmarshalProduct(binaryData, &decodedProduct); err != nil {
		log.Fatalf("Binary unmarshal failed: %v", err)
	}
	fmt.Printf("Decoded product: %s\n", decodedProduct.Name)

	// JSON serialization
	jsonSerializer := NewJSONSerializer()
	jsonData, err := jsonSerializer.MarshalProduct(product)
	if err != nil {
		log.Fatalf("JSON marshal failed: %v", err)
	}
	fmt.Printf("JSON size: %d bytes\n", len(jsonData))
	fmt.Printf("JSON data: %s\n", string(jsonData))

	var jsonProduct Product
	if err := jsonSerializer.UnmarshalProduct(jsonData, &jsonProduct); err != nil {
		log.Fatalf("JSON unmarshal failed: %v", err)
	}
	fmt.Printf("JSON decoded product: %s\n", jsonProduct.Name)

	fmt.Printf("\nCompression ratio: %.2f%%\n", float64(len(binaryData))/float64(len(jsonData))*100)
}
