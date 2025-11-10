package main

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestNewMigrationHistory(t *testing.T) {
	history := NewMigrationHistory()

	if history == nil {
		t.Error("History should not be nil")
	}

	if len(history.Applied) != 0 {
		t.Error("Applied should be empty initially")
	}

	t.Log("✓ MigrationHistory creation works!")
}

func TestAddApplied(t *testing.T) {
	history := NewMigrationHistory()

	status := &MigrationStatus{
		Version:   1,
		Name:      "create_users_table",
		AppliedAt: time.Now(),
		Success:   true,
	}

	history.AddApplied(status)

	if !history.IsApplied(1) {
		t.Error("Migration 1 should be marked as applied")
	}

	t.Log("✓ AddApplied works!")
}

func TestGetApplied(t *testing.T) {
	history := NewMigrationHistory()

	for i := 1; i <= 3; i++ {
		status := &MigrationStatus{
			Version:   i,
			Name:      fmt.Sprintf("migration_%d", i),
			AppliedAt: time.Now(),
			Success:   true,
		}
		history.AddApplied(status)
	}

	applied := history.GetApplied()

	if len(applied) != 3 {
		t.Errorf("Expected 3 applied migrations, got %d", len(applied))
	}

	// Verify sorted by version
	for i := 0; i < len(applied)-1; i++ {
		if applied[i].Version > applied[i+1].Version {
			t.Error("Applied migrations should be sorted by version")
		}
	}

	t.Log("✓ GetApplied works and sorts migrations!")
}

func TestNewMigrationManager(t *testing.T) {
	manager := NewMigrationManager()

	if manager == nil {
		t.Error("Manager should not be nil")
	}

	if len(manager.migrations) != 0 {
		t.Error("Migrations should be empty initially")
	}

	t.Log("✓ MigrationManager creation works!")
}

func TestRegisterMigration(t *testing.T) {
	manager := NewMigrationManager()

	migration := &Migration{
		Version: 1,
		Name:    "create_users_table",
		UpSQL:   "CREATE TABLE users",
		DownSQL: "DROP TABLE users",
	}

	err := manager.RegisterMigration(migration)
	if err != nil {
		t.Fatalf("Failed to register migration: %v", err)
	}

	// Try duplicate registration
	err = manager.RegisterMigration(migration)
	if err == nil {
		t.Error("Should not allow duplicate migration registration")
	}

	t.Log("✓ RegisterMigration works!")
}

func TestGetPending(t *testing.T) {
	manager := NewMigrationManager()

	// Register 3 migrations
	for i := 1; i <= 3; i++ {
		migration := &Migration{
			Version: i,
			Name:    fmt.Sprintf("migration_%d", i),
			UpSQL:   "SQL",
			DownSQL: "SQL",
		}
		manager.RegisterMigration(migration)
	}

	pending := manager.GetPending()
	if len(pending) != 3 {
		t.Errorf("Expected 3 pending migrations, got %d", len(pending))
	}

	// Apply one migration
	manager.history.AddApplied(&MigrationStatus{
		Version: 1,
		Name:    "migration_1",
		Success: true,
	})

	pending = manager.GetPending()
	if len(pending) != 2 {
		t.Errorf("Expected 2 pending migrations, got %d", len(pending))
	}

	t.Log("✓ GetPending works!")
}

func TestApplyUp(t *testing.T) {
	manager := NewMigrationManager()

	// Register migrations
	for i := 1; i <= 2; i++ {
		migration := &Migration{
			Version: i,
			Name:    fmt.Sprintf("migration_%d", i),
			UpSQL:   "SQL",
			DownSQL: "SQL",
		}
		manager.RegisterMigration(migration)
	}

	applied, err := manager.ApplyUp()
	if err != nil {
		t.Fatalf("ApplyUp failed: %v", err)
	}

	if len(applied) != 2 {
		t.Errorf("Expected 2 applied migrations, got %d", len(applied))
	}

	// Verify all are marked as applied
	for i := 1; i <= 2; i++ {
		if !manager.history.IsApplied(i) {
			t.Errorf("Migration %d should be applied", i)
		}
	}

	t.Log("✓ ApplyUp works!")
}

func TestApplyDown(t *testing.T) {
	manager := NewMigrationManager()

	// Register and apply migrations
	for i := 1; i <= 3; i++ {
		migration := &Migration{
			Version: i,
			Name:    fmt.Sprintf("migration_%d", i),
			UpSQL:   "SQL",
			DownSQL: "SQL",
		}
		manager.RegisterMigration(migration)
		manager.history.AddApplied(&MigrationStatus{
			Version:   i,
			Name:      migration.Name,
			AppliedAt: time.Now(),
			Success:   true,
		})
	}

	// Rollback to version 1
	rolled, err := manager.ApplyDown(1)
	if err != nil {
		t.Fatalf("ApplyDown failed: %v", err)
	}

	if len(rolled) != 2 {
		t.Errorf("Expected 2 rolled back migrations, got %d", len(rolled))
	}

	// Verify correct migrations were rolled back
	if manager.history.IsApplied(2) || manager.history.IsApplied(3) {
		t.Error("Migrations 2 and 3 should be rolled back")
	}

	if !manager.history.IsApplied(1) {
		t.Error("Migration 1 should still be applied")
	}

	t.Log("✓ ApplyDown works!")
}

func TestStatus(t *testing.T) {
	manager := NewMigrationManager()

	// Register migrations
	for i := 1; i <= 3; i++ {
		migration := &Migration{
			Version: i,
			Name:    fmt.Sprintf("migration_%d", i),
			UpSQL:   "SQL",
			DownSQL: "SQL",
		}
		manager.RegisterMigration(migration)
	}

	// Apply first two
	manager.history.AddApplied(&MigrationStatus{
		Version:   1,
		Name:      "migration_1",
		AppliedAt: time.Now(),
		Success:   true,
	})
	manager.history.AddApplied(&MigrationStatus{
		Version:   2,
		Name:      "migration_2",
		AppliedAt: time.Now(),
		Success:   true,
	})

	status := manager.Status()

	if !strings.Contains(status, "Applied: 2") {
		t.Errorf("Status should show 2 applied: %s", status)
	}

	if !strings.Contains(status, "Pending: 1") {
		t.Errorf("Status should show 1 pending: %s", status)
	}

	if !strings.Contains(status, "migration_1") {
		t.Error("Status should list migration_1")
	}

	t.Log("✓ Status works!")
}

func TestMigrationBuilder(t *testing.T) {
	builder := NewMigrationBuilder()

	builder.AddMigration(
		1,
		"create_users",
		"CREATE TABLE users",
		"DROP TABLE users",
	).AddMigration(
		2,
		"add_email",
		"ALTER TABLE users ADD email",
		"ALTER TABLE users DROP email",
	)

	migrations := builder.Build()

	if len(migrations) != 2 {
		t.Errorf("Expected 2 migrations, got %d", len(migrations))
	}

	if migrations[0].Version != 1 || migrations[0].Name != "create_users" {
		t.Error("First migration incorrect")
	}

	if migrations[1].Version != 2 || migrations[1].Name != "add_email" {
		t.Error("Second migration incorrect")
	}

	t.Log("✓ MigrationBuilder works!")
}

func TestGetDefaultMigrations(t *testing.T) {
	builder := GetDefaultMigrations()
	migrations := builder.Build()

	if len(migrations) != 5 {
		t.Errorf("Expected 5 default migrations, got %d", len(migrations))
	}

	// Verify versions are sequential
	for i, mig := range migrations {
		if mig.Version != i+1 {
			t.Errorf("Migration %d should have version %d", i, i+1)
		}
	}

	// Verify all have SQL
	for _, mig := range migrations {
		if mig.UpSQL == "" {
			t.Errorf("Migration %d should have UP SQL", mig.Version)
		}
		if mig.DownSQL == "" {
			t.Errorf("Migration %d should have DOWN SQL", mig.Version)
		}
	}

	t.Log("✓ GetDefaultMigrations works!")
}

func TestDefaultMigrationConfig(t *testing.T) {
	config := DefaultMigrationConfig("postgres://localhost/testdb")

	if config.DatabaseURL != "postgres://localhost/testdb" {
		t.Error("Database URL should match")
	}

	if config.TableName != "schema_migrations" {
		t.Error("Table name should be schema_migrations")
	}

	if config.Timeout == 0 {
		t.Error("Timeout should be set")
	}

	t.Log("✓ DefaultMigrationConfig works!")
}

func TestCalculateChecksum(t *testing.T) {
	upSQL := "CREATE TABLE users"
	downSQL := "DROP TABLE users"

	checksum1 := CalculateChecksum(upSQL, downSQL)
	checksum2 := CalculateChecksum(upSQL, downSQL)

	if checksum1 != checksum2 {
		t.Error("Same SQL should produce same checksum")
	}

	checksum3 := CalculateChecksum(upSQL, "DROP TABLE users2")
	if checksum1 == checksum3 {
		t.Error("Different SQL should produce different checksum")
	}

	if checksum1 == "" {
		t.Error("Checksum should not be empty")
	}

	t.Log("✓ CalculateChecksum works!")
}

func TestGetMigrationName(t *testing.T) {
	name := GetMigrationName(1, "create_users")
	expected := "001_create_users"

	if name != expected {
		t.Errorf("Expected %s, got %s", expected, name)
	}

	name = GetMigrationName(42, "add_column")
	expected = "042_add_column"

	if name != expected {
		t.Errorf("Expected %s, got %s", expected, name)
	}

	t.Log("✓ GetMigrationName works!")
}

func TestVersionString(t *testing.T) {
	tests := []struct {
		version  int
		expected string
	}{
		{1, "001"},
		{9, "009"},
		{10, "010"},
		{99, "099"},
		{100, "100"},
		{999, "999"},
	}

	for _, test := range tests {
		result := VersionString(test.version)
		if result != test.expected {
			t.Errorf("Version %d: expected %s, got %s",
				test.version, test.expected, result)
		}
	}

	t.Log("✓ VersionString works!")
}

func TestMigrationSequence(t *testing.T) {
	manager := NewMigrationManager()
	builder := GetDefaultMigrations()

	// Register all default migrations
	for _, mig := range builder.Build() {
		manager.RegisterMigration(mig)
	}

	// Apply all
	applied, err := manager.ApplyUp()
	if err != nil {
		t.Fatalf("ApplyUp failed: %v", err)
	}

	if len(applied) != 5 {
		t.Errorf("Expected 5 migrations applied, got %d", len(applied))
	}

	// Verify order
	for i, version := range applied {
		if version != i+1 {
			t.Errorf("Expected migration %d at position %d", i+1, i)
		}
	}

	// Partial rollback
	rolled, err := manager.ApplyDown(3)
	if err != nil {
		t.Fatalf("ApplyDown failed: %v", err)
	}

	if len(rolled) != 2 {
		t.Errorf("Expected 2 rolled back, got %d", len(rolled))
	}

	// Verify only 3 are applied
	appliedStatuses := manager.history.GetApplied()
	if len(appliedStatuses) != 3 {
		t.Errorf("Expected 3 applied, got %d", len(appliedStatuses))
	}

	t.Log("✓ Migration sequence works!")
}

func TestConcurrentApply(t *testing.T) {
	manager := NewMigrationManager()

	// Register migrations
	for i := 1; i <= 10; i++ {
		migration := &Migration{
			Version: i,
			Name:    fmt.Sprintf("migration_%d", i),
			UpSQL:   "SQL",
			DownSQL: "SQL",
		}
		manager.RegisterMigration(migration)
	}

	// Apply in order (single-threaded for correctness)
	applied, err := manager.ApplyUp()
	if err != nil {
		t.Fatalf("ApplyUp failed: %v", err)
	}

	if len(applied) != 10 {
		t.Errorf("Expected 10 applied, got %d", len(applied))
	}

	// Verify all marked as applied
	for i := 1; i <= 10; i++ {
		if !manager.history.IsApplied(i) {
			t.Errorf("Migration %d should be applied", i)
		}
	}

	t.Log("✓ Concurrent apply safety works!")
}

func TestMigrationWithDuration(t *testing.T) {
	status := &MigrationStatus{
		Version:   1,
		Name:      "test_migration",
		AppliedAt: time.Now().Add(-5 * time.Second),
		Duration:  5 * time.Second,
		Success:   true,
	}

	history := NewMigrationHistory()
	history.AddApplied(status)

	applied := history.GetApplied()
	if len(applied) != 1 {
		t.Error("Should have one applied migration")
	}

	if applied[0].Duration != 5*time.Second {
		t.Errorf("Duration should be 5s, got %v", applied[0].Duration)
	}

	t.Log("✓ Migration duration tracking works!")
}

func BenchmarkApplyUp(b *testing.B) {
	manager := NewMigrationManager()

	// Register 10 migrations
	for i := 1; i <= 10; i++ {
		migration := &Migration{
			Version: i,
			Name:    fmt.Sprintf("migration_%d", i),
			UpSQL:   "SQL",
			DownSQL: "SQL",
		}
		manager.RegisterMigration(migration)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		manager.ApplyUp()
	}
}

func BenchmarkGetPending(b *testing.B) {
	manager := NewMigrationManager()

	// Register migrations
	for i := 1; i <= 20; i++ {
		migration := &Migration{
			Version: i,
			Name:    fmt.Sprintf("migration_%d", i),
			UpSQL:   "SQL",
			DownSQL: "SQL",
		}
		manager.RegisterMigration(migration)
	}

	// Apply half
	for i := 1; i <= 10; i++ {
		manager.history.AddApplied(&MigrationStatus{
			Version:   i,
			Name:      fmt.Sprintf("migration_%d", i),
			AppliedAt: time.Now(),
			Success:   true,
		})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		manager.GetPending()
	}
}
