package main

import (
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestAPIVersionString(t *testing.T) {
	v := APIVersion{Major: 1, Minor: 0}
	if v.String() != "v1.0" {
		t.Errorf("Expected v1.0, got %s", v.String())
	}

	v = APIVersion{Major: 2, Minor: 3}
	if v.String() != "v2.3" {
		t.Errorf("Expected v2.3, got %s", v.String())
	}

	t.Log("✓ API Version string works!")
}

func TestURLVersionRouter(t *testing.T) {
	router := NewURLVersionRouter()

	handler := func(w http.ResponseWriter, r *http.Request) {}

	v1 := APIVersion{Major: 1, Minor: 0}
	v2 := APIVersion{Major: 2, Minor: 0}

	router.RegisterVersion("v1", handler, v1)
	router.RegisterVersion("v2", handler, v2)

	// Get v1 handler
	h, err := router.RouteRequest("v1")
	if err != nil {
		t.Fatalf("Should find v1: %v", err)
	}

	if h == nil {
		t.Error("Handler should not be nil")
	}

	// Try missing version
	_, err = router.RouteRequest("v99")
	if err == nil {
		t.Error("Should error on missing version")
	}

	t.Log("✓ URL Version Router works!")
}

func TestURLVersionDuplicate(t *testing.T) {
	router := NewURLVersionRouter()

	handler := func(w http.ResponseWriter, r *http.Request) {}
	v := APIVersion{Major: 1, Minor: 0}

	err := router.RegisterVersion("v1", handler, v)
	if err != nil {
		t.Fatalf("First registration should succeed: %v", err)
	}

	err = router.RegisterVersion("v1", handler, v)
	if err == nil {
		t.Error("Duplicate registration should fail")
	}

	t.Log("✓ URL Version Router prevents duplicates!")
}

func TestHeaderVersionRouter(t *testing.T) {
	router := NewHeaderVersionRouter("v1")

	handler := func(w http.ResponseWriter, r *http.Request) {}

	v1 := APIVersion{Major: 1, Minor: 0}
	v2 := APIVersion{Major: 2, Minor: 0}

	router.RegisterVersion("v1", handler, v1)
	router.RegisterVersion("v2", handler, v2)

	// Get handler with Accept header
	h, version, err := router.RouteRequest("application/json; version=v2")
	if err != nil {
		t.Fatalf("Should find handler: %v", err)
	}

	if version != "v2" {
		t.Errorf("Expected v2, got %s", version)
	}

	if h == nil {
		t.Error("Handler should not be nil")
	}

	t.Log("✓ Header Version Router works!")
}

func TestHeaderVersionDefault(t *testing.T) {
	router := NewHeaderVersionRouter("v1")

	handler := func(w http.ResponseWriter, r *http.Request) {}
	v1 := APIVersion{Major: 1, Minor: 0}

	router.RegisterVersion("v1", handler, v1)

	// No header, should use default
	_, version, err := router.RouteRequest("")
	if err != nil {
		t.Fatalf("Should use default version: %v", err)
	}

	if version != "v1" {
		t.Errorf("Expected v1, got %s", version)
	}

	t.Log("✓ Header Version Router defaults work!")
}

func TestGetVersionFromHeader(t *testing.T) {
	router := NewHeaderVersionRouter("v1")

	tests := []struct {
		header   string
		expected string
	}{
		{"application/json; version=v2.0", "v2.0"},
		{"application/vnd.api+json; version=1.5", "1.5"},
		{"application/json", "v1"}, // default
		{"", "v1"}, // default
	}

	for _, tt := range tests {
		result := router.GetVersionFromHeader(tt.header)
		if result != tt.expected {
			t.Errorf("Header %q: expected %s, got %s", tt.header, tt.expected, result)
		}
	}

	t.Log("✓ Header parsing works!")
}

func TestBackwardCompatibilityAdapter(t *testing.T) {
	newHandler := func(w http.ResponseWriter, r *http.Request) {}
	adapter := NewBackwardCompatibilityAdapter(newHandler)

	// Register transformation
	adapter.RegisterTransform("timestamp", func(v interface{}) interface{} {
		// Convert to milliseconds
		return v
	})

	// Data should be transformable
	oldData := map[string]interface{}{
		"timestamp": 1704067200,
	}

	transformed := adapter.Transform(oldData)
	if transformed == nil {
		t.Error("Should transform data")
	}

	t.Log("✓ Backward Compatibility Adapter works!")
}

func TestVersioningMiddleware(t *testing.T) {
	v := APIVersion{Major: 2, Minor: 1}
	vm := NewVersioningMiddleware(v)

	handler := func(w http.ResponseWriter, r *http.Request) {}
	wrapped := vm.Middleware(handler)

	if wrapped == nil {
		t.Error("Middleware should return wrapped handler")
	}

	t.Log("✓ Versioning Middleware works!")
}

func TestDeprecationHandler(t *testing.T) {
	dh := NewDeprecationHandler()

	notice := &DeprecationNotice{
		Version:     "v1.0",
		Message:     "Use v2.0 instead",
		ReplaceWith: "v2.0",
		SunsetDate:  time.Now().AddDate(0, 6, 0),
	}

	dh.RegisterDeprecation("/api/v1/users", notice)

	// Get deprecation notice
	retrieved, exists := dh.GetDeprecationNotice("/api/v1/users")
	if !exists {
		t.Error("Should find deprecation notice")
	}

	if retrieved.ReplaceWith != "v2.0" {
		t.Error("Should have correct replacement version")
	}

	t.Log("✓ Deprecation Handler works!")
}

func TestDeprecationWarningHeader(t *testing.T) {
	dh := NewDeprecationHandler()

	sunsetDate := time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC)
	notice := &DeprecationNotice{
		Version:     "v1.0",
		ReplaceWith: "v2.0",
		SunsetDate:  sunsetDate,
	}

	dh.RegisterDeprecation("/api/v1/users", notice)

	header := dh.DeprecationWarningHeader("/api/v1/users")

	if !strings.Contains(header, "Deprecated") {
		t.Errorf("Header should contain Deprecated: %s", header)
	}

	if !strings.Contains(header, "Sunset=") {
		t.Errorf("Header should contain Sunset: %s", header)
	}

	t.Log("✓ Deprecation Warning Header works!")
}

func TestVersionNegotiator(t *testing.T) {
	vn := NewVersionNegotiator()

	v1 := APIVersion{Major: 1, Minor: 0}
	v2 := APIVersion{Major: 2, Minor: 0}

	vn.AddSupportedVersion("v1", v1)
	vn.AddSupportedVersion("v2", v2)

	// Prefer URL version
	version, err := vn.NegotiateVersion("", "v2")
	if err != nil {
		t.Fatalf("Should negotiate: %v", err)
	}

	if version != "v2" {
		t.Errorf("Expected v2, got %s", version)
	}

	// Fallback to header
	version, err = vn.NegotiateVersion("application/json; version=v1", "")
	if err != nil {
		t.Fatalf("Should negotiate from header: %v", err)
	}

	if version != "v1" {
		t.Errorf("Expected v1, got %s", version)
	}

	t.Log("✓ Version Negotiator works!")
}

func TestGetSupportedVersions(t *testing.T) {
	vn := NewVersionNegotiator()

	versions := []string{"v1", "v2", "v3"}
	for _, v := range versions {
		vn.AddSupportedVersion(v, APIVersion{})
	}

	supported := vn.GetSupportedVersions()

	if len(supported) != 3 {
		t.Errorf("Expected 3 versions, got %d", len(supported))
	}

	// Check all are present
	for _, v := range versions {
		found := false
		for _, s := range supported {
			if s == v {
				found = true
				break
			}
		}

		if !found {
			t.Errorf("Missing version %s", v)
		}
	}

	t.Log("✓ Get Supported Versions works!")
}

func TestAPIVersionManager(t *testing.T) {
	defaultV := APIVersion{Major: 2, Minor: 0}
	manager := NewAPIVersionManager(defaultV)

	handler := func(w http.ResponseWriter, r *http.Request) {}

	v1 := APIVersion{Major: 1, Minor: 0}
	v2 := APIVersion{Major: 2, Minor: 0}

	manager.RegisterVersion("v1", handler, v1)
	manager.RegisterVersion("v2", handler, v2)

	supported := manager.negotiator.GetSupportedVersions()
	if len(supported) < 2 {
		t.Error("Should have registered versions")
	}

	t.Log("✓ API Version Manager works!")
}

func TestMarkDeprecated(t *testing.T) {
	defaultV := APIVersion{Major: 2, Minor: 0}
	manager := NewAPIVersionManager(defaultV)

	sunsetDate := time.Now().AddDate(1, 0, 0)
	manager.MarkDeprecated("v1", "v2", sunsetDate)

	notice, exists := manager.deprecations.GetDeprecationNotice("v1")
	if !exists {
		t.Error("Should find deprecation notice")
	}

	if notice.ReplaceWith != "v2" {
		t.Error("Should have correct replacement")
	}

	t.Log("✓ Mark Deprecated works!")
}

func TestGetMigrationGuides(t *testing.T) {
	guides := GetMigrationGuides()

	if len(guides) == 0 {
		t.Error("Should have migration guides")
	}

	for _, guide := range guides {
		if guide.FromVersion == "" {
			t.Error("Guide should have FromVersion")
		}

		if guide.ToVersion == "" {
			t.Error("Guide should have ToVersion")
		}

		if len(guide.Changes) == 0 {
			t.Error("Guide should have Changes")
		}
	}

	t.Log("✓ Migration Guides are complete!")
}

func TestGetVersioningBestPractices(t *testing.T) {
	practices := GetVersioningBestPractices()

	expectedKeys := []string{
		"url_versioning",
		"header_versioning",
		"backward_compatibility",
		"deprecation_strategy",
		"version_numbers",
		"migration_support",
	}

	for _, key := range expectedKeys {
		if _, exists := practices[key]; !exists {
			t.Errorf("Missing practice: %s", key)
		}

		if practices[key] == "" {
			t.Errorf("Practice %s should not be empty", key)
		}
	}

	t.Log("✓ Best practices documented!")
}

func TestVersionMigration(t *testing.T) {
	defaultV := APIVersion{Major: 1, Minor: 0}
	manager := NewAPIVersionManager(defaultV)

	v1Handler := func(w http.ResponseWriter, r *http.Request) {}
	v2Handler := func(w http.ResponseWriter, r *http.Request) {}

	v1 := APIVersion{Major: 1, Minor: 0}
	v2 := APIVersion{Major: 2, Minor: 0}

	manager.RegisterVersion("v1", v1Handler, v1)
	manager.RegisterVersion("v2", v2Handler, v2)

	// Mark v1 as deprecated
	manager.MarkDeprecated("v1", "v2", time.Now().AddDate(1, 0, 0))

	// Should still work but with warning
	notice, exists := manager.deprecations.GetDeprecationNotice("v1")
	if !exists {
		t.Error("Should have deprecation notice for v1")
	}

	if notice.ReplaceWith != "v2" {
		t.Error("Should suggest v2 as replacement")
	}

	t.Log("✓ Version Migration pattern works!")
}

func TestMultipleVersionsCoexist(t *testing.T) {
	defaultV := APIVersion{Major: 3, Minor: 0}
	manager := NewAPIVersionManager(defaultV)

	handler := func(w http.ResponseWriter, r *http.Request) {}

	versions := map[string]APIVersion{
		"v1": {Major: 1, Minor: 0},
		"v2": {Major: 2, Minor: 0},
		"v3": {Major: 3, Minor: 0},
	}

	for versionStr, versionMeta := range versions {
		manager.RegisterVersion(versionStr, handler, versionMeta)
	}

	supported := manager.negotiator.GetSupportedVersions()
	if len(supported) != 3 {
		t.Errorf("Expected 3 versions, got %d", len(supported))
	}

	t.Log("✓ Multiple versions coexist!")
}

func TestVersionNegotiationPriority(t *testing.T) {
	vn := NewVersionNegotiator()

	vn.AddSupportedVersion("v1", APIVersion{Major: 1})
	vn.AddSupportedVersion("v2", APIVersion{Major: 2})

	// URL should take priority over header
	version, err := vn.NegotiateVersion("application/json; version=v1", "v2")
	if err != nil {
		t.Fatalf("Should negotiate: %v", err)
	}

	if version != "v2" {
		t.Errorf("URL version should have priority, got %s", version)
	}

	t.Log("✓ Version negotiation priority works!")
}

func TestBackwardCompatibilityChain(t *testing.T) {
	// Simulate chaining adapters for multiple versions
	newHandler := func(w http.ResponseWriter, r *http.Request) {}
	adapter := NewBackwardCompatibilityAdapter(newHandler)

	// Register multiple transformations
	adapter.RegisterTransform("id", func(v interface{}) interface{} { return v })
	adapter.RegisterTransform("timestamp", func(v interface{}) interface{} { return v })
	adapter.RegisterTransform("status", func(v interface{}) interface{} { return v })

	data := map[string]interface{}{
		"id":        1,
		"timestamp": 1704067200,
		"status":    "active",
	}

	result := adapter.Transform(data)
	if result == nil {
		t.Error("Should transform data")
	}

	t.Log("✓ Backward compatibility chain works!")
}

func BenchmarkVersionNegotiation(b *testing.B) {
	vn := NewVersionNegotiator()
	vn.AddSupportedVersion("v1", APIVersion{Major: 1})
	vn.AddSupportedVersion("v2", APIVersion{Major: 2})
	vn.AddSupportedVersion("v3", APIVersion{Major: 3})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		vn.NegotiateVersion("", "v2")
	}
}

func BenchmarkAPIVersionManagerRegister(b *testing.B) {
	defaultV := APIVersion{Major: 1, Minor: 0}
	manager := NewAPIVersionManager(defaultV)

	handler := func(w http.ResponseWriter, r *http.Request) {}
	v := APIVersion{Major: 1, Minor: 0}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		manager.RegisterVersion("v1", handler, v)
	}
}
