package main

import (
	"testing"
	"time"
)

func createTestProduct() *Product {
	return &Product{
		ID:          "prod-123",
		Name:        "Test Product",
		Description: "Test Description",
		Price: &Price{
			Amount:   10000,
			Currency: "USD",
		},
		Category: Electronics,
		Tags:     []string{"tag1", "tag2"},
		Attributes: map[string]string{
			"key1": "value1",
			"key2": "value2",
		},
		Inventory: &Inventory{
			Quantity:    100,
			WarehouseID: "WH-001",
		},
		Reviews: []*Review{
			{
				ID:        "rev-1",
				UserID:    "user-1",
				Rating:    5,
				Comment:   "Great!",
				CreatedAt: time.Now(),
			},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func TestBinarySerializer_MarshalUnmarshal(t *testing.T) {
	product := createTestProduct()
	serializer := NewBinarySerializer()

	// Marshal
	data, err := serializer.MarshalProduct(product)
	if err != nil {
		t.Fatalf("MarshalProduct failed: %v", err)
	}

	if len(data) == 0 {
		t.Error("MarshalProduct returned empty data")
	}

	// Unmarshal
	var decoded Product
	err = serializer.UnmarshalProduct(data, &decoded)
	if err != nil {
		t.Fatalf("UnmarshalProduct failed: %v", err)
	}

	// Verify
	if decoded.ID != product.ID {
		t.Errorf("ID = %v, want %v", decoded.ID, product.ID)
	}
	if decoded.Name != product.Name {
		t.Errorf("Name = %v, want %v", decoded.Name, product.Name)
	}
	if decoded.Category != product.Category {
		t.Errorf("Category = %v, want %v", decoded.Category, product.Category)
	}
}

func TestJSONSerializer_MarshalUnmarshal(t *testing.T) {
	product := createTestProduct()
	serializer := NewJSONSerializer()

	// Marshal
	data, err := serializer.MarshalProduct(product)
	if err != nil {
		t.Fatalf("MarshalProduct failed: %v", err)
	}

	if len(data) == 0 {
		t.Error("MarshalProduct returned empty data")
	}

	// Unmarshal
	var decoded Product
	err = serializer.UnmarshalProduct(data, &decoded)
	if err != nil {
		t.Fatalf("UnmarshalProduct failed: %v", err)
	}

	// Verify
	if decoded.ID != product.ID {
		t.Errorf("ID = %v, want %v", decoded.ID, product.ID)
	}
	if decoded.Name != product.Name {
		t.Errorf("Name = %v, want %v", decoded.Name, product.Name)
	}
	if decoded.Category != product.Category {
		t.Errorf("Category = %v, want %v", decoded.Category, product.Category)
	}
}

func TestRepeatedFields(t *testing.T) {
	product := createTestProduct()
	product.Tags = []string{"tag1", "tag2", "tag3", "tag4", "tag5"}

	serializer := NewBinarySerializer()
	data, err := serializer.MarshalProduct(product)
	if err != nil {
		t.Fatalf("MarshalProduct failed: %v", err)
	}

	var decoded Product
	err = serializer.UnmarshalProduct(data, &decoded)
	if err != nil {
		t.Fatalf("UnmarshalProduct failed: %v", err)
	}

	if len(decoded.Tags) != len(product.Tags) {
		t.Errorf("Tags length = %v, want %v", len(decoded.Tags), len(product.Tags))
	}

	for i, tag := range decoded.Tags {
		if tag != product.Tags[i] {
			t.Errorf("Tag[%d] = %v, want %v", i, tag, product.Tags[i])
		}
	}
}

func TestMapFields(t *testing.T) {
	product := createTestProduct()
	product.Attributes = map[string]string{
		"brand":  "BrandX",
		"model":  "ModelY",
		"color":  "Red",
		"size":   "Large",
		"weight": "5kg",
	}

	serializer := NewBinarySerializer()
	data, err := serializer.MarshalProduct(product)
	if err != nil {
		t.Fatalf("MarshalProduct failed: %v", err)
	}

	var decoded Product
	err = serializer.UnmarshalProduct(data, &decoded)
	if err != nil {
		t.Fatalf("UnmarshalProduct failed: %v", err)
	}

	if len(decoded.Attributes) != len(product.Attributes) {
		t.Errorf("Attributes length = %v, want %v", len(decoded.Attributes), len(product.Attributes))
	}

	for k, v := range product.Attributes {
		if decoded.Attributes[k] != v {
			t.Errorf("Attributes[%s] = %v, want %v", k, decoded.Attributes[k], v)
		}
	}
}

func TestNestedMessages(t *testing.T) {
	product := createTestProduct()
	product.Inventory = &Inventory{
		Quantity:    200,
		WarehouseID: "WH-002",
		Locations: []*StockLocation{
			{
				LocationID: "LOC-1",
				Address:    "Address 1",
				Quantity:   100,
			},
			{
				LocationID: "LOC-2",
				Address:    "Address 2",
				Quantity:   100,
			},
		},
	}

	serializer := NewBinarySerializer()
	data, err := serializer.MarshalProduct(product)
	if err != nil {
		t.Fatalf("MarshalProduct failed: %v", err)
	}

	var decoded Product
	err = serializer.UnmarshalProduct(data, &decoded)
	if err != nil {
		t.Fatalf("UnmarshalProduct failed: %v", err)
	}

	if decoded.Inventory.Quantity != product.Inventory.Quantity {
		t.Errorf("Inventory.Quantity = %v, want %v", decoded.Inventory.Quantity, product.Inventory.Quantity)
	}
	if decoded.Inventory.WarehouseID != product.Inventory.WarehouseID {
		t.Errorf("Inventory.WarehouseID = %v, want %v", decoded.Inventory.WarehouseID, product.Inventory.WarehouseID)
	}
}

func TestValidator_ValidateProduct(t *testing.T) {
	validator := NewValidator()

	tests := []struct {
		name    string
		product *Product
		wantErr bool
	}{
		{
			name:    "valid product",
			product: createTestProduct(),
			wantErr: false,
		},
		{
			name: "missing ID",
			product: &Product{
				Name: "Test",
				Price: &Price{
					Amount:   1000,
					Currency: "USD",
				},
				Category: Electronics,
			},
			wantErr: true,
		},
		{
			name: "missing name",
			product: &Product{
				ID: "123",
				Price: &Price{
					Amount:   1000,
					Currency: "USD",
				},
				Category: Electronics,
			},
			wantErr: true,
		},
		{
			name: "missing price",
			product: &Product{
				ID:       "123",
				Name:     "Test",
				Category: Electronics,
			},
			wantErr: true,
		},
		{
			name: "negative price",
			product: &Product{
				ID:   "123",
				Name: "Test",
				Price: &Price{
					Amount:   -100,
					Currency: "USD",
				},
				Category: Electronics,
			},
			wantErr: true,
		},
		{
			name: "unspecified category",
			product: &Product{
				ID:   "123",
				Name: "Test",
				Price: &Price{
					Amount:   1000,
					Currency: "USD",
				},
				Category: CategoryUnspecified,
			},
			wantErr: true,
		},
		{
			name: "invalid review rating",
			product: &Product{
				ID:   "123",
				Name: "Test",
				Price: &Price{
					Amount:   1000,
					Currency: "USD",
				},
				Category: Electronics,
				Reviews: []*Review{
					{
						ID:     "rev-1",
						Rating: 6,
					},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateProduct(tt.product)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateProduct() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidator_ValidateOrder(t *testing.T) {
	validator := NewValidator()

	tests := []struct {
		name    string
		order   *Order
		wantErr bool
	}{
		{
			name: "valid order",
			order: &Order{
				ID:         "order-1",
				CustomerID: "customer-1",
				Items: []*OrderItem{
					{
						ProductID: "prod-1",
						Quantity:  1,
						UnitPrice: &Price{Amount: 1000, Currency: "USD"},
					},
				},
				Status: Pending,
				ShippingAddress: &Address{
					Street: "123 Main St",
				},
				Total: &Price{Amount: 1000, Currency: "USD"},
			},
			wantErr: false,
		},
		{
			name: "missing ID",
			order: &Order{
				CustomerID: "customer-1",
			},
			wantErr: true,
		},
		{
			name: "missing customer ID",
			order: &Order{
				ID: "order-1",
			},
			wantErr: true,
		},
		{
			name: "no items",
			order: &Order{
				ID:         "order-1",
				CustomerID: "customer-1",
				Items:      []*OrderItem{},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateOrder(tt.order)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateOrder() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCategoryString(t *testing.T) {
	tests := []struct {
		category Category
		want     string
	}{
		{CategoryUnspecified, "UNSPECIFIED"},
		{Electronics, "ELECTRONICS"},
		{Clothing, "CLOTHING"},
		{Books, "BOOKS"},
		{Home, "HOME"},
		{Sports, "SPORTS"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.category.String(); got != tt.want {
				t.Errorf("Category.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSerializationSize(t *testing.T) {
	product := createTestProduct()

	binarySerializer := NewBinarySerializer()
	binaryData, err := binarySerializer.MarshalProduct(product)
	if err != nil {
		t.Fatalf("Binary marshal failed: %v", err)
	}

	jsonSerializer := NewJSONSerializer()
	jsonData, err := jsonSerializer.MarshalProduct(product)
	if err != nil {
		t.Fatalf("JSON marshal failed: %v", err)
	}

	t.Logf("Binary size: %d bytes", len(binaryData))
	t.Logf("JSON size: %d bytes", len(jsonData))
	t.Logf("Compression ratio: %.2f%%", float64(len(binaryData))/float64(len(jsonData))*100)

	// Binary should typically be smaller
	if len(binaryData) > len(jsonData) {
		t.Log("Warning: Binary serialization is larger than JSON (expected for small messages)")
	}
}

func BenchmarkBinaryMarshal(b *testing.B) {
	product := createTestProduct()
	serializer := NewBinarySerializer()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := serializer.MarshalProduct(product)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkBinaryUnmarshal(b *testing.B) {
	product := createTestProduct()
	serializer := NewBinarySerializer()
	data, _ := serializer.MarshalProduct(product)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var p Product
		err := serializer.UnmarshalProduct(data, &p)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkJSONMarshal(b *testing.B) {
	product := createTestProduct()
	serializer := NewJSONSerializer()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := serializer.MarshalProduct(product)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkJSONUnmarshal(b *testing.B) {
	product := createTestProduct()
	serializer := NewJSONSerializer()
	data, _ := serializer.MarshalProduct(product)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var p Product
		err := serializer.UnmarshalProduct(data, &p)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkValidation(b *testing.B) {
	product := createTestProduct()
	validator := NewValidator()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		validator.ValidateProduct(product)
	}
}
