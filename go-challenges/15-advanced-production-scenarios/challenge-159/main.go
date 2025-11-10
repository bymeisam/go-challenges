package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"
)

// ========== Tenant Models ==========

// TenantContext represents tenant information in request context
type TenantContext struct {
	TenantID      string
	TenantName    string
	Plan          string // free, pro, enterprise
	UserID        string
	RequestID     string
	Timestamp     time.Time
	Metadata      map[string]interface{}
}

// Tenant represents a tenant in the system
type Tenant struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	Plan          string `json:"plan"`
	Status        string `json:"status"` // active, suspended, deleted
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	Settings      map[string]interface{} `json:"settings"`
	DatabaseURL   string `json:"database_url,omitempty"`
	SchemaName    string `json:"schema_name,omitempty"`
}

// TenantResource represents any resource in a multi-tenant system
type TenantResource struct {
	ID        string    `json:"id"`
	TenantID  string    `json:"tenant_id"`
	Name      string    `json:"name"`
	Data      map[string]interface{} `json:"data"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ========== Resource Quota Models ==========

// ResourceQuota defines resource limits for a tenant
type ResourceQuota struct {
	TenantID          string            `json:"tenant_id"`
	MaxUsers          int               `json:"max_users"`
	MaxAPIRequests    int               `json:"max_api_requests"`
	MaxStorage        int64             `json:"max_storage"`
	MaxDatabases      int               `json:"max_databases"`
	CurrentUsers      int               `json:"current_users"`
	CurrentRequests   int               `json:"current_requests"`
	CurrentStorage    int64             `json:"current_storage"`
	CurrentDatabases  int               `json:"current_databases"`
	ResetTime         time.Time         `json:"reset_time"`
	Mu                sync.RWMutex      `json:"-"`
}

// ========== Tenant Manager ==========

// TenantManager manages multi-tenant operations
type TenantManager struct {
	tenants        map[string]*Tenant
	tenantsMu      sync.RWMutex
	resources      map[string][]*TenantResource
	resourcesMu    sync.RWMutex
	quotas         map[string]*ResourceQuota
	quotasMu       sync.RWMutex
	auditLog       []*AuditLogEntry
	auditLogMu     sync.RWMutex
	tenantRoutes   map[string]string // tenant -> database URL
	routesMu       sync.RWMutex
	isolationMode  string // "database", "schema", "row-level"
}

// AuditLogEntry represents an audit log entry
type AuditLogEntry struct {
	Timestamp   time.Time              `json:"timestamp"`
	TenantID    string                 `json:"tenant_id"`
	UserID      string                 `json:"user_id"`
	Action      string                 `json:"action"`
	ResourceID  string                 `json:"resource_id"`
	Details     map[string]interface{} `json:"details"`
}

// NewTenantManager creates a new tenant manager
func NewTenantManager(isolationMode string) *TenantManager {
	return &TenantManager{
		tenants:       make(map[string]*Tenant),
		resources:     make(map[string][]*TenantResource),
		quotas:        make(map[string]*ResourceQuota),
		auditLog:      []*AuditLogEntry{},
		tenantRoutes:  make(map[string]string),
		isolationMode: isolationMode,
	}
}

// ========== Tenant Operations ==========

// CreateTenant creates a new tenant
func (tm *TenantManager) CreateTenant(name, plan string, settings map[string]interface{}) (*Tenant, error) {
	tm.tenantsMu.Lock()
	defer tm.tenantsMu.Unlock()

	tenantID := generateTenantID()

	// Check for duplicate names
	for _, t := range tm.tenants {
		if t.Name == name {
			return nil, errors.New("tenant name already exists")
		}
	}

	tenant := &Tenant{
		ID:        tenantID,
		Name:      name,
		Plan:      plan,
		Status:    "active",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Settings:  settings,
	}

	// Set database configuration based on isolation mode
	switch tm.isolationMode {
	case "database":
		tenant.DatabaseURL = fmt.Sprintf("postgres://localhost/tenant_%s", tenantID)
	case "schema":
		tenant.SchemaName = fmt.Sprintf("tenant_%s", tenantID)
		tenant.DatabaseURL = "postgres://localhost/shared_db"
	case "row-level":
		tenant.SchemaName = "public"
		tenant.DatabaseURL = "postgres://localhost/shared_db"
	}

	tm.tenants[tenantID] = tenant

	// Create quota
	tm.createQuotaForTenant(tenantID, plan)

	// Log audit
	tm.logAudit(tenantID, "", "CREATE_TENANT", tenantID, map[string]interface{}{"plan": plan})

	return tenant, nil
}

// GetTenant retrieves a tenant
func (tm *TenantManager) GetTenant(tenantID string) (*Tenant, error) {
	tm.tenantsMu.RLock()
	defer tm.tenantsMu.RUnlock()

	tenant, exists := tm.tenants[tenantID]
	if !exists {
		return nil, errors.New("tenant not found")
	}

	return tenant, nil
}

// UpdateTenant updates tenant settings
func (tm *TenantManager) UpdateTenant(tenantID string, settings map[string]interface{}) error {
	tm.tenantsMu.Lock()
	defer tm.tenantsMu.Unlock()

	tenant, exists := tm.tenants[tenantID]
	if !exists {
		return errors.New("tenant not found")
	}

	tenant.Settings = settings
	tenant.UpdatedAt = time.Now()

	tm.logAudit(tenantID, "", "UPDATE_TENANT", tenantID, map[string]interface{}{"settings": settings})

	return nil
}

// SuspendTenant suspends a tenant
func (tm *TenantManager) SuspendTenant(tenantID, reason string) error {
	tm.tenantsMu.Lock()
	defer tm.tenantsMu.Unlock()

	tenant, exists := tm.tenants[tenantID]
	if !exists {
		return errors.New("tenant not found")
	}

	tenant.Status = "suspended"
	tenant.UpdatedAt = time.Now()

	tm.logAudit(tenantID, "", "SUSPEND_TENANT", tenantID, map[string]interface{}{"reason": reason})

	return nil
}

// DeleteTenant marks a tenant as deleted
func (tm *TenantManager) DeleteTenant(tenantID string) error {
	tm.tenantsMu.Lock()
	defer tm.tenantsMu.Unlock()

	tenant, exists := tm.tenants[tenantID]
	if !exists {
		return errors.New("tenant not found")
	}

	tenant.Status = "deleted"
	tenant.UpdatedAt = time.Now()

	tm.logAudit(tenantID, "", "DELETE_TENANT", tenantID, nil)

	return nil
}

// ========== Resource Operations ==========

// CreateResource creates a resource for a tenant
func (tm *TenantManager) CreateResource(tenantID, resourceName string, data map[string]interface{}) (*TenantResource, error) {
	// Validate tenant exists
	if _, err := tm.GetTenant(tenantID); err != nil {
		return nil, err
	}

	// Check quota
	if !tm.canAllocateResource(tenantID) {
		return nil, errors.New("quota exceeded")
	}

	tm.resourcesMu.Lock()
	defer tm.resourcesMu.Unlock()

	resource := &TenantResource{
		ID:        generateResourceID(),
		TenantID:  tenantID,
		Name:      resourceName,
		Data:      data,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	tm.resources[tenantID] = append(tm.resources[tenantID], resource)

	tm.logAudit(tenantID, "", "CREATE_RESOURCE", resource.ID, map[string]interface{}{"name": resourceName})

	return resource, nil
}

// GetResource retrieves a resource with isolation validation
func (tm *TenantManager) GetResource(tenantID, resourceID string) (*TenantResource, error) {
	tm.resourcesMu.RLock()
	defer tm.resourcesMu.RUnlock()

	resources, exists := tm.resources[tenantID]
	if !exists {
		return nil, errors.New("tenant has no resources")
	}

	for _, resource := range resources {
		if resource.ID == resourceID {
			// Validate tenant isolation
			if resource.TenantID != tenantID {
				return nil, errors.New("access denied: resource belongs to different tenant")
			}
			return resource, nil
		}
	}

	return nil, errors.New("resource not found")
}

// ListResources lists all resources for a tenant
func (tm *TenantManager) ListResources(tenantID string) ([]*TenantResource, error) {
	if _, err := tm.GetTenant(tenantID); err != nil {
		return nil, err
	}

	tm.resourcesMu.RLock()
	defer tm.resourcesMu.RUnlock()

	resources := tm.resources[tenantID]
	result := make([]*TenantResource, len(resources))
	copy(result, resources)

	return result, nil
}

// DeleteResource deletes a resource
func (tm *TenantManager) DeleteResource(tenantID, resourceID string) error {
	tm.resourcesMu.Lock()
	defer tm.resourcesMu.Unlock()

	resources, exists := tm.resources[tenantID]
	if !exists {
		return errors.New("tenant has no resources")
	}

	for i, resource := range resources {
		if resource.ID == resourceID {
			if resource.TenantID != tenantID {
				return errors.New("access denied: resource belongs to different tenant")
			}
			tm.resources[tenantID] = append(resources[:i], resources[i+1:]...)
			tm.logAudit(tenantID, "", "DELETE_RESOURCE", resourceID, nil)
			return nil
		}
	}

	return errors.New("resource not found")
}

// ========== Resource Quota Management ==========

func (tm *TenantManager) createQuotaForTenant(tenantID, plan string) {
	quota := &ResourceQuota{
		TenantID:       tenantID,
		ResetTime:      time.Now().Add(24 * time.Hour),
	}

	// Set quotas based on plan
	switch plan {
	case "free":
		quota.MaxUsers = 5
		quota.MaxAPIRequests = 1000
		quota.MaxStorage = 1 * 1024 * 1024 * 1024 // 1GB
		quota.MaxDatabases = 1
	case "pro":
		quota.MaxUsers = 50
		quota.MaxAPIRequests = 100000
		quota.MaxStorage = 100 * 1024 * 1024 * 1024 // 100GB
		quota.MaxDatabases = 10
	case "enterprise":
		quota.MaxUsers = 10000
		quota.MaxAPIRequests = 10000000
		quota.MaxStorage = 1000 * 1024 * 1024 * 1024 // 1TB
		quota.MaxDatabases = 100
	}

	tm.quotasMu.Lock()
	tm.quotas[tenantID] = quota
	tm.quotasMu.Unlock()
}

// canAllocateResource checks if a tenant can allocate more resources
func (tm *TenantManager) canAllocateResource(tenantID string) bool {
	tm.quotasMu.RLock()
	quota, exists := tm.quotas[tenantID]
	tm.quotasMu.RUnlock()

	if !exists {
		return false
	}

	quota.Mu.RLock()
	defer quota.Mu.RUnlock()

	// Simple check - in production, use actual resource tracking
	resourceCount := tm.getResourceCount(tenantID)
	return resourceCount < quota.MaxUsers
}

func (tm *TenantManager) getResourceCount(tenantID string) int {
	tm.resourcesMu.RLock()
	defer tm.resourcesMu.RUnlock()

	return len(tm.resources[tenantID])
}

// GetQuota retrieves quota for a tenant
func (tm *TenantManager) GetQuota(tenantID string) (*ResourceQuota, error) {
	tm.quotasMu.RLock()
	defer tm.quotasMu.RUnlock()

	quota, exists := tm.quotas[tenantID]
	if !exists {
		return nil, errors.New("quota not found")
	}

	return quota, nil
}

// IncrementQuotaUsage increments quota usage
func (tm *TenantManager) IncrementQuotaUsage(tenantID string, usageType string) error {
	tm.quotasMu.RLock()
	quota, exists := tm.quotas[tenantID]
	tm.quotasMu.RUnlock()

	if !exists {
		return errors.New("quota not found")
	}

	quota.Mu.Lock()
	defer quota.Mu.Unlock()

	switch usageType {
	case "api_request":
		quota.CurrentRequests++
		if quota.CurrentRequests > quota.MaxAPIRequests {
			quota.CurrentRequests--
			return errors.New("API request quota exceeded")
		}
	case "user":
		quota.CurrentUsers++
		if quota.CurrentUsers > quota.MaxUsers {
			quota.CurrentUsers--
			return errors.New("user quota exceeded")
		}
	}

	return nil
}

// ========== Tenant Context Management ==========

// ContextKey type for context values
type ContextKey string

const TenantContextKey ContextKey = "tenant_context"

// WithTenantContext adds tenant context to a request context
func WithTenantContext(ctx context.Context, tenantCtx *TenantContext) context.Context {
	return context.WithValue(ctx, TenantContextKey, tenantCtx)
}

// GetTenantContext extracts tenant context from request context
func GetTenantContext(ctx context.Context) (*TenantContext, error) {
	tenantCtx, ok := ctx.Value(TenantContextKey).(*TenantContext)
	if !ok {
		return nil, errors.New("tenant context not found")
	}
	return tenantCtx, nil
}

// ValidateRequestTenancy validates that request is for correct tenant
func (tm *TenantManager) ValidateRequestTenancy(ctx context.Context, requestedTenantID string) error {
	tenantCtx, err := GetTenantContext(ctx)
	if err != nil {
		return errors.New("no tenant context in request")
	}

	if tenantCtx.TenantID != requestedTenantID {
		return fmt.Errorf("tenant mismatch: context=%s, requested=%s", tenantCtx.TenantID, requestedTenantID)
	}

	// Check tenant status
	tenant, err := tm.GetTenant(tenantCtx.TenantID)
	if err != nil {
		return errors.New("tenant not found")
	}

	if tenant.Status != "active" {
		return fmt.Errorf("tenant is %s", tenant.Status)
	}

	return nil
}

// ========== Cross-Tenant Prevention ==========

// ValidateResourceAccess validates that a user can access a resource
func (tm *TenantManager) ValidateResourceAccess(ctx context.Context, resourceID string) error {
	tenantCtx, err := GetTenantContext(ctx)
	if err != nil {
		return errors.New("no tenant context")
	}

	tm.resourcesMu.RLock()
	resources := tm.resources[tenantCtx.TenantID]
	tm.resourcesMu.RUnlock()

	for _, resource := range resources {
		if resource.ID == resourceID {
			if resource.TenantID != tenantCtx.TenantID {
				tm.logAudit(tenantCtx.TenantID, tenantCtx.UserID, "CROSS_TENANT_ACCESS_DENIED", resourceID, nil)
				return errors.New("cross-tenant access denied")
			}
			return nil
		}
	}

	return errors.New("resource not found")
}

// ========== Audit Logging ==========

func (tm *TenantManager) logAudit(tenantID, userID, action, resourceID string, details map[string]interface{}) {
	entry := &AuditLogEntry{
		Timestamp:  time.Now(),
		TenantID:   tenantID,
		UserID:     userID,
		Action:     action,
		ResourceID: resourceID,
		Details:    details,
	}

	tm.auditLogMu.Lock()
	tm.auditLog = append(tm.auditLog, entry)
	tm.auditLogMu.Unlock()
}

// GetAuditLog retrieves audit log for a tenant
func (tm *TenantManager) GetAuditLog(tenantID string) []*AuditLogEntry {
	tm.auditLogMu.RLock()
	defer tm.auditLogMu.RUnlock()

	var entries []*AuditLogEntry
	for _, entry := range tm.auditLog {
		if entry.TenantID == tenantID {
			entries = append(entries, entry)
		}
	}

	return entries
}

// ========== Tenant Routing ==========

// RegisterTenantRoute registers a database URL for a tenant
func (tm *TenantManager) RegisterTenantRoute(tenantID, databaseURL string) error {
	if _, err := tm.GetTenant(tenantID); err != nil {
		return err
	}

	tm.routesMu.Lock()
	tm.tenantRoutes[tenantID] = databaseURL
	tm.routesMu.Unlock()

	return nil
}

// GetTenantRoute retrieves the database URL for a tenant
func (tm *TenantManager) GetTenantRoute(tenantID string) (string, error) {
	tm.routesMu.RLock()
	defer tm.routesMu.RUnlock()

	url, exists := tm.tenantRoutes[tenantID]
	if !exists {
		// Return default or tenant's configured database
		tenant, err := tm.GetTenant(tenantID)
		if err != nil {
			return "", err
		}
		return tenant.DatabaseURL, nil
	}

	return url, nil
}

// ========== Tenant Statistics ==========

// GetTenantStats returns statistics for a tenant
func (tm *TenantManager) GetTenantStats(tenantID string) map[string]interface{} {
	resourceCount := 0
	tm.resourcesMu.RLock()
	if resources, exists := tm.resources[tenantID]; exists {
		resourceCount = len(resources)
	}
	tm.resourcesMu.RUnlock()

	auditCount := 0
	tm.auditLogMu.RLock()
	for _, entry := range tm.auditLog {
		if entry.TenantID == tenantID {
			auditCount++
		}
	}
	tm.auditLogMu.RUnlock()

	quota, _ := tm.GetQuota(tenantID)
	quotaData := make(map[string]interface{})
	if quota != nil {
		quota.Mu.RLock()
		quotaData = map[string]interface{}{
			"users": map[string]int{"current": quota.CurrentUsers, "max": quota.MaxUsers},
			"requests": map[string]int{"current": quota.CurrentRequests, "max": quota.MaxAPIRequests},
		}
		quota.Mu.RUnlock()
	}

	return map[string]interface{}{
		"tenant_id":      tenantID,
		"resource_count": resourceCount,
		"audit_count":    auditCount,
		"quota":          quotaData,
	}
}

// ========== Helper Functions ==========

func generateTenantID() string {
	return fmt.Sprintf("tenant_%d", time.Now().UnixNano())
}

func generateResourceID() string {
	return fmt.Sprintf("resource_%d", time.Now().UnixNano())
}

// SerializeTenant serializes tenant to JSON
func SerializeTenant(tenant *Tenant) (string, error) {
	data, err := json.MarshalIndent(tenant, "", "  ")
	return string(data), err
}

func main() {
	// Example multi-tenancy
	tm := NewTenantManager("database")
	_, _ = tm.CreateTenant("TestCorp", "pro", nil)
}
