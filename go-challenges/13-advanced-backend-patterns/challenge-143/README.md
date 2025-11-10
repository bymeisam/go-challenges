# Challenge 143: Protobuf Serialization

**Difficulty:** ⭐⭐⭐⭐ Hard
**Time Estimate:** 50 minutes

## Description

Implement advanced Protocol Buffers (protobuf) serialization and deserialization with nested messages, repeated fields, enums, maps, and oneof fields. This project demonstrates efficient binary serialization, schema evolution, and interoperability.

## Features

- **Message Definition**: Complex nested message structures
- **Repeated Fields**: Arrays and lists in protobuf
- **Maps**: Key-value pairs in messages
- **Enums**: Strongly typed enumerations
- **Oneof Fields**: Union types for optional fields
- **Nested Messages**: Composition and embedding
- **Binary Serialization**: Efficient encoding/decoding
- **JSON Interoperability**: Convert between protobuf and JSON
- **Schema Evolution**: Backward/forward compatibility
- **Field Options**: Custom options and validation
- **Well-Known Types**: Timestamp, Duration, Any, etc.
- **Performance**: Benchmarking serialization speed

## Protobuf Schema

```protobuf
syntax = "proto3";

package ecommerce;

option go_package = "github.com/example/ecommerce/proto";

import "google/protobuf/timestamp.proto";
import "google/protobuf/any.proto";

// Product catalog message
message Product {
  string id = 1;
  string name = 2;
  string description = 3;
  Price price = 4;
  Category category = 5;
  repeated string tags = 6;
  map<string, string> attributes = 7;
  Inventory inventory = 8;
  repeated Review reviews = 9;
  google.protobuf.Timestamp created_at = 10;
  google.protobuf.Timestamp updated_at = 11;

  oneof discount {
    PercentageDiscount percentage_discount = 12;
    FixedDiscount fixed_discount = 13;
  }
}

message Price {
  int64 amount = 1;  // in cents
  string currency = 2;
}

enum Category {
  CATEGORY_UNSPECIFIED = 0;
  ELECTRONICS = 1;
  CLOTHING = 2;
  BOOKS = 3;
  HOME = 4;
  SPORTS = 5;
}

message Inventory {
  int32 quantity = 1;
  string warehouse_id = 2;
  repeated StockLocation locations = 3;
}

message StockLocation {
  string location_id = 1;
  string address = 2;
  int32 quantity = 3;
}

message Review {
  string id = 1;
  string user_id = 2;
  int32 rating = 3;  // 1-5
  string comment = 4;
  google.protobuf.Timestamp created_at = 5;
}

message PercentageDiscount {
  double percentage = 1;  // 0-100
}

message FixedDiscount {
  int64 amount = 1;  // in cents
}

// Order message
message Order {
  string id = 1;
  string customer_id = 2;
  repeated OrderItem items = 3;
  OrderStatus status = 4;
  Address shipping_address = 5;
  Address billing_address = 6;
  Price total = 7;
  google.protobuf.Timestamp created_at = 8;
  map<string, string> metadata = 9;
}

message OrderItem {
  string product_id = 1;
  int32 quantity = 2;
  Price unit_price = 3;
}

enum OrderStatus {
  ORDER_STATUS_UNSPECIFIED = 0;
  PENDING = 1;
  CONFIRMED = 2;
  SHIPPED = 3;
  DELIVERED = 4;
  CANCELLED = 5;
}

message Address {
  string street = 1;
  string city = 2;
  string state = 3;
  string zip_code = 4;
  string country = 5;
}

// Batch operations
message ProductBatch {
  repeated Product products = 1;
}

message OrderBatch {
  repeated Order orders = 1;
}
```

## Requirements

1. Implement message structs matching the protobuf schema
2. Implement binary serialization (Marshal/Unmarshal)
3. Implement JSON serialization for interoperability
4. Handle nested messages correctly
5. Support repeated fields (slices)
6. Implement map fields
7. Handle oneof fields (union types)
8. Support well-known types (Timestamp)
9. Implement schema validation
10. Test backward compatibility
11. Benchmark serialization performance

## Example Usage

```go
// Create a product
product := &Product{
    ID:          "prod-123",
    Name:        "Laptop",
    Description: "High-performance laptop",
    Price: &Price{
        Amount:   99900,  // $999.00
        Currency: "USD",
    },
    Category: Category_ELECTRONICS,
    Tags:     []string{"computer", "portable", "work"},
    Attributes: map[string]string{
        "brand":  "TechCorp",
        "model":  "X1000",
        "color":  "Silver",
    },
    Inventory: &Inventory{
        Quantity:    50,
        WarehouseID: "WH-001",
    },
}

// Add discount (oneof field)
product.Discount = &Product_PercentageDiscount{
    PercentageDiscount: &PercentageDiscount{
        Percentage: 15.0,
    },
}

// Serialize to binary
data, err := proto.Marshal(product)

// Deserialize from binary
var decoded Product
err = proto.Unmarshal(data, &decoded)

// Convert to JSON
jsonData, err := protojson.Marshal(product)

// Convert from JSON
err = protojson.Unmarshal(jsonData, &product)
```

## Learning Objectives

- Protocol Buffers binary format
- Efficient serialization techniques
- Schema design best practices
- Nested message composition
- Repeated fields and maps
- Oneof fields for union types
- Well-known types usage
- Binary vs JSON trade-offs
- Schema evolution strategies
- Backward/forward compatibility
- Performance optimization
- Cross-language interoperability

## Testing Focus

- Test message serialization/deserialization
- Test nested messages
- Test repeated fields
- Test map fields
- Test oneof fields
- Test JSON interoperability
- Test schema evolution
- Test large message handling
- Test validation
- Benchmark serialization performance
- Compare with JSON performance
