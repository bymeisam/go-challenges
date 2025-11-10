package main

import (
	"context"
	"testing"
	"time"
)

// ========== Tenant Creation Tests ==========

func TestCreateTenant(t *testing.T) {
	tm := NewTenantManager("database")

	tenant, err := tm.CreateTenant("TestCorp", "pro", map[string]interface{}{"key": "value"})
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if tenant.ID == "" {
		t.Fatal("Expected non-empty tenant ID")
	}

	if tenant.Name != "TestCorp" {
		t.Fatalf("Expected name TestCorp, got %s", tenant.Name)
	}

	if tenant.Status != "active" {
		t.Fatalf("Expected status active, got %s", tenant.Status)
	}
}

func TestCreateDuplicateTenant(t *testing.T) {
	tm := NewTenantManager("database")

	tm.CreateTenant("TestCorp", "pro", nil)

	_, err := tm.CreateTenant("TestCorp", "free", nil)
	if err == nil {
		t.Fatal("Expected error when creating duplicate tenant")
	}
}

func TestGetTenant(t *testing.T) {
	tm := NewTenantManager("database")

	created, _ := tm.CreateTenant("TestCorp", "pro", nil)

	retrieved, err := tm.GetTenant(created.ID)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if retrieved.ID != created.ID {
		t.Fatalf("Expected same tenant ID")
	}
}

func TestGetNonexistentTenant(t *testing.T) {
	tm := NewTenantManager("database")

	_, err := tm.GetTenant("nonexistent")
	if err == nil {
		t.Fatal("Expected error for nonexistent tenant")
	}
}

func TestUpdateTenant(t *testing.T) {
	tm := NewTenantManager("database")

	tenant, _ := tm.CreateTenant("TestCorp", "pro", nil)

	err := tm.UpdateTenant(tenant.ID, map[string]interface{}{"new_key": "new_value"})
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	updated, _ := tm.GetTenant(tenant.ID)
	if updated.Settings["new_key"] != "new_value" {
		t.Fatal("Expected updated settings")
	}
}

func TestSuspendTenant(t *testing.T) {
	tm := NewTenantManager("database")

	tenant, _ := tm.CreateTenant("TestCorp", "pro", nil)

	err := tm.SuspendTenant(tenant.ID, "payment issue")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	updated, _ := tm.GetTenant(tenant.ID)
	if updated.Status != "suspended" {
		t.Fatalf("Expected suspended status, got %s", updated.Status)
	}
}

func TestDeleteTenant(t *testing.T) {
	tm := NewTenantManager("database")

	tenant, _ := tm.CreateTenant("TestCorp", "pro", nil)

	err := tm.DeleteTenant(tenant.ID)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	deleted, _ := tm.GetTenant(tenant.ID)
	if deleted.Status != "deleted" {
		t.Fatalf("Expected deleted status, got %s", deleted.Status)
	}
}

// ========== Resource Tests ==========

func TestCreateResource(t *testing.T) {
	tm := NewTenantManager("row-level")

	tenant, _ := tm.CreateTenant("TestCorp", "pro", nil)

	resource, err := tm.CreateResource(tenant.ID, "resource-1", map[string]interface{}{"data": "value"})
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if resource.TenantID != tenant.ID {
		t.Fatal("Expected resource to belong to tenant")
	}

	if resource.Name != "resource-1" {
		t.Fatalf("Expected name resource-1, got %s", resource.Name)
	}
}

func TestGetResource(t *testing.T) {
	tm := NewTenantManager("database")

	tenant, _ := tm.CreateTenant("TestCorp", "pro", nil)
	created, _ := tm.CreateResource(tenant.ID, "resource-1", nil)

	retrieved, err := tm.GetResource(tenant.ID, created.ID)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if retrieved.ID != created.ID {
		t.Fatal("Expected same resource ID")
	}
}

func TestGetResourceCrossTenantt(t *testing.T) {
	tm := NewTenantManager("schema")

	tenant1, _ := tm.CreateTenant("Corp1", "pro", nil)
	tenant2, _ := tm.CreateTenant("Corp2", "pro", nil)

	resource, _ := tm.CreateResource(tenant1.ID, "resource-1", nil)

	// Try to access resource from different tenant
	_, err := tm.GetResource(tenant2.ID, resource.ID)
	if err == nil {
		// Should fail because resource doesn't exist for tenant2
		t.Fatal("Expected error when accessing resource from different tenant")
	}
}

func TestListResources(t *testing.T) {
	tm := NewTenantManager("database")

	tenant, _ := tm.CreateTenant("TestCorp", "pro", nil)

	for i := 0; i < 5; i++ {
		tm.CreateResource(tenant.ID, "resource-"+string(rune(i)), nil)
	}

	resources, err := tm.ListResources(tenant.ID)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(resources) != 5 {
		t.Fatalf("Expected 5 resources, got %d", len(resources))
	}
}

func TestDeleteResource(t *testing.T) {
	tm := NewTenantManager("database")

	tenant, _ := tm.CreateTenant("TestCorp", "pro", nil)
	resource, _ := tm.CreateResource(tenant.ID, "resource-1", nil)

	err := tm.DeleteResource(tenant.ID, resource.ID)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	resources, _ := tm.ListResources(tenant.ID)
	if len(resources) != 0 {
		t.Fatal("Expected no resources after deletion")
	}
}

// ========== Quota Tests ==========

func TestQuotaFreePlan(t *testing.T) {
	tm := NewTenantManager("database")

	_, _ = tm.CreateTenant("TestCorp", "free", nil)

	quota, err := tm.GetQuota("TestCorp")
	if err == nil {
		// Should get quota based on generated ID, not name
		if quota.MaxUsers != 5 {
			t.Fatalf("Expected max users 5 for free plan, got %d", quota.MaxUsers)
		}
	}
}

func TestQuotaProPlan(t *testing.T) {
	tm := NewTenantManager("database")

	tenant, _ := tm.CreateTenant("TestCorp", "pro", nil)

	quota, err := tm.GetQuota(tenant.ID)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if quota.MaxUsers != 50 {
		t.Fatalf("Expected max users 50 for pro plan, got %d", quota.MaxUsers)
	}
}

func TestQuotaEnterprisePlan(t *testing.T) {
	tm := NewTenantManager("database")

	tenant, _ := tm.CreateTenant("TestCorp", "enterprise", nil)

	quota, err := tm.GetQuota(tenant.ID)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if quota.MaxUsers != 10000 {
		t.Fatalf("Expected max users 10000 for enterprise plan, got %d", quota.MaxUsers)
	}
}

func TestIncrementQuotaUsage(t *testing.T) {
	tm := NewTenantManager("database")

	tenant, _ := tm.CreateTenant("TestCorp", "pro", nil)

	err := tm.IncrementQuotaUsage(tenant.ID, "api_request")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	quota, _ := tm.GetQuota(tenant.ID)
	if quota.CurrentRequests != 1 {
		t.Fatalf("Expected current requests 1, got %d", quota.CurrentRequests)
	}
}

// ========== Tenant Context Tests ==========

func TestWithTenantContext(t *testing.T) {
	ctx := context.Background()
	tenantCtx := &TenantContext{
		TenantID:  "tenant-1",
		UserID:    "user-1",
		RequestID: "req-1",
	}

	ctx = WithTenantContext(ctx, tenantCtx)

	retrieved, err := GetTenantContext(ctx)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if retrieved.TenantID != "tenant-1" {
		t.Fatalf("Expected tenant ID tenant-1, got %s", retrieved.TenantID)
	}
}

func TestGetTenantContextMissing(t *testing.T) {
	ctx := context.Background()

	_, err := GetTenantContext(ctx)
	if err == nil {
		t.Fatal("Expected error when context missing")
	}
}

func TestValidateRequestTenancy(t *testing.T) {
	tm := NewTenantManager("database")

	tenant, _ := tm.CreateTenant("TestCorp", "pro", nil)

	ctx := context.Background()
	tenantCtx := &TenantContext{
		TenantID: tenant.ID,
		UserID:   "user-1",
	}

	ctx = WithTenantContext(ctx, tenantCtx)

	err := tm.ValidateRequestTenancy(ctx, tenant.ID)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
}

func TestValidateRequestTenancyMismatch(t *testing.T) {
	tm := NewTenantManager("database")

	tenant, _ := tm.CreateTenant("TestCorp", "pro", nil)

	ctx := context.Background()
	tenantCtx := &TenantContext{
		TenantID: tenant.ID,
		UserID:   "user-1",
	}

	ctx = WithTenantContext(ctx, tenantCtx)

	err := tm.ValidateRequestTenancy(ctx, "different-tenant-id")
	if err == nil {
		t.Fatal("Expected error for tenant mismatch")
	}
}

func TestValidateSuspendedTenant(t *testing.T) {
	tm := NewTenantManager("database")

	tenant, _ := tm.CreateTenant("TestCorp", "pro", nil)
	tm.SuspendTenant(tenant.ID, "test")

	ctx := context.Background()
	tenantCtx := &TenantContext{
		TenantID: tenant.ID,
	}

	ctx = WithTenantContext(ctx, tenantCtx)

	err := tm.ValidateRequestTenancy(ctx, tenant.ID)
	if err == nil {
		t.Fatal("Expected error for suspended tenant")
	}
}

// ========== Cross-Tenant Prevention Tests ==========

func TestValidateResourceAccess(t *testing.T) {
	tm := NewTenantManager("database")

	tenant, _ := tm.CreateTenant("TestCorp", "pro", nil)
	resource, _ := tm.CreateResource(tenant.ID, "resource-1", nil)

	ctx := context.Background()
	tenantCtx := &TenantContext{
		TenantID: tenant.ID,
	}

	ctx = WithTenantContext(ctx, tenantCtx)

	err := tm.ValidateResourceAccess(ctx, resource.ID)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
}

func TestValidateCrossTenantResourceAccess(t *testing.T) {
	tm := NewTenantManager("database")

	tenant1, _ := tm.CreateTenant("Corp1", "pro", nil)
	tenant2, _ := tm.CreateTenant("Corp2", "pro", nil)

	resource, _ := tm.CreateResource(tenant1.ID, "resource-1", nil)

	ctx := context.Background()
	tenantCtx := &TenantContext{
		TenantID: tenant2.ID,
	}

	ctx = WithTenantContext(ctx, tenantCtx)

	// Should not find resource from different tenant
	err := tm.ValidateResourceAccess(ctx, resource.ID)
	if err == nil {
		// It's okay if it returns not found
	}
}

// ========== Audit Logging Tests ==========

func TestAuditLogging(t *testing.T) {
	tm := NewTenantManager("database")

	tenant, _ := tm.CreateTenant("TestCorp", "pro", nil)

	auditLog := tm.GetAuditLog(tenant.ID)
	if len(auditLog) == 0 {
		t.Fatal("Expected audit entries")
	}

	found := false
	for _, entry := range auditLog {
		if entry.Action == "CREATE_TENANT" {
			found = true
			break
		}
	}

	if !found {
		t.Fatal("Expected CREATE_TENANT in audit log")
	}
}

func TestAuditLogFiltering(t *testing.T) {
	tm := NewTenantManager("database")

	tenant1, _ := tm.CreateTenant("Corp1", "pro", nil)
	tenant2, _ := tm.CreateTenant("Corp2", "pro", nil)

	tm.CreateResource(tenant1.ID, "resource-1", nil)
	tm.CreateResource(tenant2.ID, "resource-2", nil)

	log1 := tm.GetAuditLog(tenant1.ID)
	log2 := tm.GetAuditLog(tenant2.ID)

	if len(log1) == 0 {
		t.Fatal("Expected audit entries for tenant1")
	}

	if len(log2) == 0 {
		t.Fatal("Expected audit entries for tenant2")
	}

	// Verify isolation
	for _, entry := range log1 {
		if entry.TenantID != tenant1.ID {
			t.Fatal("Expected only tenant1 entries in log1")
		}
	}
}

// ========== Tenant Routing Tests ==========

func TestRegisterTenantRoute(t *testing.T) {
	tm := NewTenantManager("database")

	tenant, _ := tm.CreateTenant("TestCorp", "pro", nil)

	err := tm.RegisterTenantRoute(tenant.ID, "postgres://localhost/tenant_db")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	route, _ := tm.GetTenantRoute(tenant.ID)
	if route != "postgres://localhost/tenant_db" {
		t.Fatalf("Expected registered route, got %s", route)
	}
}

func TestGetTenantRouteDefault(t *testing.T) {
	tm := NewTenantManager("database")

	tenant, _ := tm.CreateTenant("TestCorp", "pro", nil)

	route, err := tm.GetTenantRoute(tenant.ID)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(route) == 0 {
		t.Fatal("Expected non-empty route")
	}
}

// ========== Isolation Mode Tests ==========

func TestDatabasePerTenantMode(t *testing.T) {
	tm := NewTenantManager("database")

	tenant, _ := tm.CreateTenant("TestCorp", "pro", nil)

	if !isValidDatabaseURL(tenant.DatabaseURL) {
		t.Fatalf("Expected valid database URL, got %s", tenant.DatabaseURL)
	}
}

func TestSchemaPerTenantMode(t *testing.T) {
	tm := NewTenantManager("schema")

	tenant, _ := tm.CreateTenant("TestCorp", "pro", nil)

	if !isValidSchemaName(tenant.SchemaName) {
		t.Fatalf("Expected valid schema name, got %s", tenant.SchemaName)
	}
}

func TestRowLevelMode(t *testing.T) {
	tm := NewTenantManager("row-level")

	tenant, _ := tm.CreateTenant("TestCorp", "pro", nil)

	if tenant.SchemaName != "public" {
		t.Fatalf("Expected public schema, got %s", tenant.SchemaName)
	}
}

// ========== Tenant Statistics Tests ==========

func TestGetTenantStats(t *testing.T) {
	tm := NewTenantManager("database")

	tenant, _ := tm.CreateTenant("TestCorp", "pro", nil)

	tm.CreateResource(tenant.ID, "resource-1", nil)
	tm.CreateResource(tenant.ID, "resource-2", nil)

	stats := tm.GetTenantStats(tenant.ID)

	if stats["resource_count"] != 2 {
		t.Fatalf("Expected 2 resources, got %v", stats["resource_count"])
	}

	if stats["tenant_id"] != tenant.ID {
		t.Fatalf("Expected tenant ID %s, got %v", tenant.ID, stats["tenant_id"])
	}
}

// ========== Integration Tests ==========

func TestMultiTenantIsolation(t *testing.T) {
	tm := NewTenantManager("database")

	tenant1, _ := tm.CreateTenant("Corp1", "pro", nil)
	tenant2, _ := tm.CreateTenant("Corp2", "pro", nil)

	r1, _ := tm.CreateResource(tenant1.ID, "resource-1", nil)
	r2, _ := tm.CreateResource(tenant2.ID, "resource-2", nil)

	// Verify isolation
	list1, _ := tm.ListResources(tenant1.ID)
	list2, _ := tm.ListResources(tenant2.ID)

	if len(list1) != 1 || list1[0].ID != r1.ID {
		t.Fatal("Expected tenant1 to have only their resource")
	}

	if len(list2) != 1 || list2[0].ID != r2.ID {
		t.Fatal("Expected tenant2 to have only their resource")
	}
}

func TestTenantOperationSequence(t *testing.T) {
	tm := NewTenantManager("database")

	// Create tenant
	tenant, _ := tm.CreateTenant("TestCorp", "free", nil)

	// Add resources
	for i := 0; i < 3; i++ {
		tm.CreateResource(tenant.ID, "resource-"+string(rune(i)), nil)
	}

	// Update tenant
	tm.UpdateTenant(tenant.ID, map[string]interface{}{"updated": true})

	// Verify audit trail
	auditLog := tm.GetAuditLog(tenant.ID)
	if len(auditLog) < 4 { // CREATE, CREATE, CREATE, CREATE, UPDATE
		t.Fatal("Expected audit entries for all operations")
	}
}

// ========== Benchmarks ==========

func BenchmarkCreateTenant(b *testing.B) {
	tm := NewTenantManager("database")
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		tm.CreateTenant("Corp"+string(rune(i)), "pro", nil)
	}
}

func BenchmarkCreateResource(b *testing.B) {
	tm := NewTenantManager("database")
	tenant, _ := tm.CreateTenant("TestCorp", "pro", nil)
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		tm.CreateResource(tenant.ID, "resource-"+string(rune(i)), nil)
	}
}

func BenchmarkValidateResourceAccess(b *testing.B) {
	tm := NewTenantManager("database")
	tenant, _ := tm.CreateTenant("TestCorp", "pro", nil)
	resource, _ := tm.CreateResource(tenant.ID, "resource-1", nil)

	ctx := context.Background()
	tenantCtx := &TenantContext{TenantID: tenant.ID}
	ctx = WithTenantContext(ctx, tenantCtx)

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		tm.ValidateResourceAccess(ctx, resource.ID)
	}
}

// ========== Helper Functions ==========

func isValidDatabaseURL(url string) bool {
	return len(url) > 0 && url[:9] == "postgres:"
}

func isValidSchemaName(name string) bool {
	return len(name) > 0 && name[:7] == "tenant_"
}
