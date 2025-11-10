package main

import (
	"fmt"
	"sort"
	"sync"
	"time"
)

// ========== Migration Types ==========

// MigrationDirection represents the direction of migration
type MigrationDirection string

const (
	DirectionUp   MigrationDirection = "up"
	DirectionDown MigrationDirection = "down"
)

// Migration represents a database migration
type Migration struct {
	Version   int       // Version number (e.g., 001, 002)
	Name      string    // Human-readable name
	UpSQL     string    // SQL to apply (up)
	DownSQL   string    // SQL to rollback (down)
	Checksum  string    // Hash for integrity checking
	Timestamp time.Time // When migration was created
}

// MigrationStatus represents the status of a migration
type MigrationStatus struct {
	Version    int
	Name       string
	AppliedAt  time.Time
	ExecutedBy string
	Duration   time.Duration
	Success    bool
	Error      string
}

// MigrationHistory tracks applied migrations
type MigrationHistory struct {
	Applied map[int]*MigrationStatus
	mu      sync.RWMutex
}

// NewMigrationHistory creates a new history tracker
func NewMigrationHistory() *MigrationHistory {
	return &MigrationHistory{
		Applied: make(map[int]*MigrationStatus),
	}
}

// AddApplied records an applied migration
func (h *MigrationHistory) AddApplied(status *MigrationStatus) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.Applied[status.Version] = status
}

// GetApplied retrieves applied migrations
func (h *MigrationHistory) GetApplied() []*MigrationStatus {
	h.mu.RLock()
	defer h.mu.RUnlock()

	statuses := make([]*MigrationStatus, 0, len(h.Applied))
	for _, status := range h.Applied {
		statuses = append(statuses, status)
	}

	// Sort by version
	sort.Slice(statuses, func(i, j int) bool {
		return statuses[i].Version < statuses[j].Version
	})

	return statuses
}

// IsApplied checks if a migration is applied
func (h *MigrationHistory) IsApplied(version int) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	_, exists := h.Applied[version]
	return exists
}

// ========== Migration Manager ==========

// MigrationManager manages database migrations
type MigrationManager struct {
	migrations map[int]*Migration
	history    *MigrationHistory
	mu         sync.RWMutex
}

// NewMigrationManager creates a migration manager
func NewMigrationManager() *MigrationManager {
	return &MigrationManager{
		migrations: make(map[int]*Migration),
		history:    NewMigrationHistory(),
	}
}

// RegisterMigration registers a migration
func (m *MigrationManager) RegisterMigration(migration *Migration) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.migrations[migration.Version]; exists {
		return fmt.Errorf("migration %d already registered", migration.Version)
	}

	m.migrations[migration.Version] = migration
	return nil
}

// GetPending returns pending (not yet applied) migrations
func (m *MigrationManager) GetPending() []*Migration {
	m.mu.RLock()
	defer m.mu.RUnlock()

	pending := []*Migration{}

	// Get all versions in order
	versions := make([]int, 0, len(m.migrations))
	for v := range m.migrations {
		versions = append(versions, v)
	}
	sort.Ints(versions)

	// Find pending migrations
	for _, v := range versions {
		if !m.history.IsApplied(v) {
			pending = append(pending, m.migrations[v])
		}
	}

	return pending
}

// ApplyUp applies pending migrations up
func (m *MigrationManager) ApplyUp() ([]int, error) {
	pending := m.GetPending()
	applied := []int{}

	for _, migration := range pending {
		status := &MigrationStatus{
			Version:    migration.Version,
			Name:       migration.Name,
			AppliedAt:  time.Now(),
			ExecutedBy: "migration-system",
			Success:    true,
		}

		// Simulate execution
		start := time.Now()
		// In real implementation, execute migration.UpSQL
		time.Sleep(10 * time.Millisecond)
		status.Duration = time.Since(start)

		m.history.AddApplied(status)
		applied = append(applied, migration.Version)
	}

	return applied, nil
}

// ApplyDown rolls back migrations down to target version
func (m *MigrationManager) ApplyDown(targetVersion int) ([]int, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	rolled := []int{}
	applied := m.history.GetApplied()

	// Rollback in reverse order
	for i := len(applied) - 1; i >= 0; i-- {
		status := applied[i]
		if status.Version <= targetVersion {
			break
		}

		migration, exists := m.migrations[status.Version]
		if !exists {
			return rolled, fmt.Errorf("migration %d not found", status.Version)
		}

		// Simulate rollback
		start := time.Now()
		// In real implementation, execute migration.DownSQL
		time.Sleep(10 * time.Millisecond)
		duration := time.Since(start)

		// Remove from history
		delete(m.history.Applied, status.Version)

		rolled = append(rolled, migration.Version)

		fmt.Printf("Rolled back migration %d (%s) - %v\n",
			migration.Version, migration.Name, duration)
	}

	return rolled, nil
}

// Status returns current migration status
func (m *MigrationManager) Status() string {
	applied := m.history.GetApplied()
	pending := m.GetPending()

	status := "=== Migration Status ===\n"
	status += fmt.Sprintf("Applied: %d\n", len(applied))
	status += fmt.Sprintf("Pending: %d\n\n", len(pending))

	status += "Applied Migrations:\n"
	for _, mig := range applied {
		status += fmt.Sprintf("  [%d] %s - %v ago\n",
			mig.Version, mig.Name,
			time.Since(mig.AppliedAt).Round(time.Second))
	}

	if len(pending) > 0 {
		status += "\nPending Migrations:\n"
		for _, mig := range pending {
			status += fmt.Sprintf("  [%d] %s\n", mig.Version, mig.Name)
		}
	}

	return status
}

// ========== Migration Collection Builder ==========

// MigrationBuilder helps build migrations
type MigrationBuilder struct {
	migrations []*Migration
}

// NewMigrationBuilder creates a builder
func NewMigrationBuilder() *MigrationBuilder {
	return &MigrationBuilder{
		migrations: []*Migration{},
	}
}

// AddMigration adds a migration to the builder
func (mb *MigrationBuilder) AddMigration(version int, name, upSQL, downSQL string) *MigrationBuilder {
	migration := &Migration{
		Version:   version,
		Name:      name,
		UpSQL:     upSQL,
		DownSQL:   downSQL,
		Timestamp: time.Now(),
	}
	mb.migrations = append(mb.migrations, migration)
	return mb
}

// Build returns migrations
func (mb *MigrationBuilder) Build() []*Migration {
	return mb.migrations
}

// ========== Sample Migrations ==========

// GetDefaultMigrations returns common database migration examples
func GetDefaultMigrations() *MigrationBuilder {
	builder := NewMigrationBuilder()

	// Migration 001: Create users table
	builder.AddMigration(
		1,
		"create_users_table",
		`
CREATE TABLE IF NOT EXISTS users (
  id SERIAL PRIMARY KEY,
  email VARCHAR(255) UNIQUE NOT NULL,
  username VARCHAR(100) UNIQUE NOT NULL,
  password_hash VARCHAR(255) NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  is_active BOOLEAN DEFAULT true
);

CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_username ON users(username);
`,
		`
DROP INDEX IF EXISTS idx_users_username;
DROP INDEX IF EXISTS idx_users_email;
DROP TABLE IF EXISTS users;
`,
	)

	// Migration 002: Create posts table
	builder.AddMigration(
		2,
		"create_posts_table",
		`
CREATE TABLE IF NOT EXISTS posts (
  id SERIAL PRIMARY KEY,
  user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  title VARCHAR(255) NOT NULL,
  content TEXT NOT NULL,
  published_at TIMESTAMP,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_posts_user_id ON posts(user_id);
CREATE INDEX idx_posts_published_at ON posts(published_at);
`,
		`
DROP INDEX IF EXISTS idx_posts_published_at;
DROP INDEX IF EXISTS idx_posts_user_id;
DROP TABLE IF EXISTS posts;
`,
	)

	// Migration 003: Add email verification
	builder.AddMigration(
		3,
		"add_email_verification",
		`
ALTER TABLE users
ADD COLUMN email_verified BOOLEAN DEFAULT false,
ADD COLUMN email_verified_at TIMESTAMP;

CREATE TABLE IF NOT EXISTS email_verifications (
  id SERIAL PRIMARY KEY,
  user_id INTEGER NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
  token VARCHAR(255) NOT NULL UNIQUE,
  expires_at TIMESTAMP NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_email_verifications_token ON email_verifications(token);
`,
		`
DROP INDEX IF EXISTS idx_email_verifications_token;
DROP TABLE IF EXISTS email_verifications;
ALTER TABLE users
DROP COLUMN IF EXISTS email_verified,
DROP COLUMN IF EXISTS email_verified_at;
`,
	)

	// Migration 004: Add comments table
	builder.AddMigration(
		4,
		"create_comments_table",
		`
CREATE TABLE IF NOT EXISTS comments (
  id SERIAL PRIMARY KEY,
  post_id INTEGER NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
  user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  content TEXT NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_comments_post_id ON comments(post_id);
CREATE INDEX idx_comments_user_id ON comments(user_id);
`,
		`
DROP INDEX IF EXISTS idx_comments_user_id;
DROP INDEX IF EXISTS idx_comments_post_id;
DROP TABLE IF EXISTS comments;
`,
	)

	// Migration 005: Add user profiles
	builder.AddMigration(
		5,
		"create_user_profiles_table",
		`
CREATE TABLE IF NOT EXISTS user_profiles (
  id SERIAL PRIMARY KEY,
  user_id INTEGER NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
  bio TEXT,
  avatar_url VARCHAR(255),
  website VARCHAR(255),
  location VARCHAR(100),
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_user_profiles_user_id ON user_profiles(user_id);
`,
		`
DROP INDEX IF EXISTS idx_user_profiles_user_id;
DROP TABLE IF EXISTS user_profiles;
`,
	)

	return builder
}

// ========== Migration Configuration ==========

// MigrationConfig holds migration configuration
type MigrationConfig struct {
	DatabaseURL      string
	TableName        string // Table to store migration history
	Timeout          time.Duration
	SkipChecksumValidation bool
}

// DefaultMigrationConfig returns default configuration
func DefaultMigrationConfig(dbURL string) *MigrationConfig {
	return &MigrationConfig{
		DatabaseURL:     dbURL,
		TableName:       "schema_migrations",
		Timeout:         30 * time.Second,
		SkipChecksumValidation: false,
	}
}

// ========== Migration Utilities ==========

// CalculateChecksum calculates migration checksum
func CalculateChecksum(upSQL, downSQL string) string {
	// Simple checksum (in production, use crypto/sha256)
	sum := 0
	for _, c := range upSQL + downSQL {
		sum += int(c)
	}
	return fmt.Sprintf("%x", sum)
}

// GetMigrationName formats migration name
func GetMigrationName(version int, name string) string {
	return fmt.Sprintf("%03d_%s", version, name)
}

// VersionString formats version for display
func VersionString(version int) string {
	return fmt.Sprintf("%03d", version)
}

func main() {}
