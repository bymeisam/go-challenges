package main

import (
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"
)

// ========== API Version Types ==========

// APIVersion represents an API version
type APIVersion struct {
	Major      int
	Minor      int
	Deprecated bool
	SunsetDate time.Time
}

// String returns version string (v1.0, v2.1, etc.)
func (v APIVersion) String() string {
	return fmt.Sprintf("v%d.%d", v.Major, v.Minor)
}

// ========== URL Versioning ==========

// URLVersionRouter routes based on URL path version
type URLVersionRouter struct {
	handlers map[string]http.HandlerFunc
	versions map[string]APIVersion
	mu       sync.RWMutex
}

// NewURLVersionRouter creates a URL version router
func NewURLVersionRouter() *URLVersionRouter {
	return &URLVersionRouter{
		handlers: make(map[string]http.HandlerFunc),
		versions: make(map[string]APIVersion),
	}
}

// RegisterVersion registers a versioned handler
func (r *URLVersionRouter) RegisterVersion(version string, handler http.HandlerFunc, meta APIVersion) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.handlers[version]; exists {
		return fmt.Errorf("version %s already registered", version)
	}

	r.handlers[version] = handler
	r.versions[version] = meta
	return nil
}

// RouteRequest routes request to appropriate version
func (r *URLVersionRouter) RouteRequest(version string) (http.HandlerFunc, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	handler, exists := r.handlers[version]
	if !exists {
		return nil, fmt.Errorf("version not found: %s", version)
	}

	versionMeta := r.versions[version]
	if versionMeta.Deprecated {
		fmt.Printf("Warning: Using deprecated API version %s\n", version)
	}

	return handler, nil
}

// GetVersionMetadata returns version information
func (r *URLVersionRouter) GetVersionMetadata(version string) (APIVersion, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	meta, exists := r.versions[version]
	if !exists {
		return APIVersion{}, fmt.Errorf("version not found: %s", version)
	}

	return meta, nil
}

// ========== Header Versioning ==========

// HeaderVersionRouter routes based on Accept header
type HeaderVersionRouter struct {
	handlers map[string]http.HandlerFunc
	versions map[string]APIVersion
	defaultVersion string
	mu       sync.RWMutex
}

// NewHeaderVersionRouter creates a header version router
func NewHeaderVersionRouter(defaultVersion string) *HeaderVersionRouter {
	return &HeaderVersionRouter{
		handlers:       make(map[string]http.HandlerFunc),
		versions:       make(map[string]APIVersion),
		defaultVersion: defaultVersion,
	}
}

// RegisterVersion registers a versioned handler
func (r *HeaderVersionRouter) RegisterVersion(version string, handler http.HandlerFunc, meta APIVersion) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.handlers[version]; exists {
		return fmt.Errorf("version %s already registered", version)
	}

	r.handlers[version] = handler
	r.versions[version] = meta
	return nil
}

// GetVersionFromHeader extracts version from Accept header
func (r *HeaderVersionRouter) GetVersionFromHeader(acceptHeader string) string {
	// Parse Accept: application/vnd.api+json; version=1.0
	parts := strings.Split(acceptHeader, ";")

	for _, part := range parts {
		if strings.Contains(part, "version=") {
			version := strings.TrimSpace(strings.Split(part, "=")[1])
			return version
		}
	}

	return r.defaultVersion
}

// RouteRequest routes based on Accept header
func (r *HeaderVersionRouter) RouteRequest(acceptHeader string) (http.HandlerFunc, string, error) {
	version := r.GetVersionFromHeader(acceptHeader)

	r.mu.RLock()
	handler, exists := r.handlers[version]
	r.mu.RUnlock()

	if !exists {
		// Fall back to default version
		r.mu.RLock()
		handler, exists = r.handlers[r.defaultVersion]
		r.mu.RUnlock()

		if !exists {
			return nil, "", fmt.Errorf("no handler for version %s or default", version)
		}

		version = r.defaultVersion
	}

	return handler, version, nil
}

// ========== Backward Compatibility ==========

// BackwardCompatibilityAdapter adapts old API to new version
type BackwardCompatibilityAdapter struct {
	newHandler http.HandlerFunc
	transforms map[string]func(interface{}) interface{}
	mu         sync.RWMutex
}

// NewBackwardCompatibilityAdapter creates an adapter
func NewBackwardCompatibilityAdapter(newHandler http.HandlerFunc) *BackwardCompatibilityAdapter {
	return &BackwardCompatibilityAdapter{
		newHandler: newHandler,
		transforms: make(map[string]func(interface{}) interface{}),
	}
}

// RegisterTransform registers a transformation
func (bca *BackwardCompatibilityAdapter) RegisterTransform(field string, transform func(interface{}) interface{}) {
	bca.mu.Lock()
	defer bca.mu.Unlock()

	bca.transforms[field] = transform
}

// Transform applies transformations to data
func (bca *BackwardCompatibilityAdapter) Transform(oldData interface{}) interface{} {
	bca.mu.RLock()
	defer bca.mu.RUnlock()

	// In real implementation, would apply transforms recursively
	return oldData
}

// ========== API Versioning Middleware ==========

// VersioningMiddleware adds versioning to responses
type VersioningMiddleware struct {
	currentVersion APIVersion
}

// NewVersioningMiddleware creates versioning middleware
func NewVersioningMiddleware(version APIVersion) *VersioningMiddleware {
	return &VersioningMiddleware{
		currentVersion: version,
	}
}

// Middleware wraps a handler with version headers
func (vm *VersioningMiddleware) Middleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Add API version headers
		w.Header().Set("X-API-Version", vm.currentVersion.String())
		w.Header().Set("Content-Type", "application/json; charset=utf-8")

		next(w, r)
	}
}

// ========== Deprecation Handler ==========

// DeprecationNotice represents a deprecation notice
type DeprecationNotice struct {
	Version      string
	Message      string
	ReplaceWith  string
	SunsetDate   time.Time
	Alternatives []string
}

// DeprecationHandler manages deprecated endpoints
type DeprecationHandler struct {
	notices map[string]*DeprecationNotice
	mu      sync.RWMutex
}

// NewDeprecationHandler creates a deprecation handler
func NewDeprecationHandler() *DeprecationHandler {
	return &DeprecationHandler{
		notices: make(map[string]*DeprecationNotice),
	}
}

// RegisterDeprecation registers a deprecated endpoint
func (dh *DeprecationHandler) RegisterDeprecation(endpoint string, notice *DeprecationNotice) {
	dh.mu.Lock()
	defer dh.mu.Unlock()

	dh.notices[endpoint] = notice
}

// GetDeprecationNotice retrieves deprecation info
func (dh *DeprecationHandler) GetDeprecationNotice(endpoint string) (*DeprecationNotice, bool) {
	dh.mu.RLock()
	defer dh.mu.RUnlock()

	notice, exists := dh.notices[endpoint]
	return notice, exists
}

// DeprecationWarningHeader returns deprecation header
func (dh *DeprecationHandler) DeprecationWarningHeader(endpoint string) string {
	notice, exists := dh.GetDeprecationNotice(endpoint)
	if !exists {
		return ""
	}

	return fmt.Sprintf("Deprecated, Sunset=%s", notice.SunsetDate.Format(time.RFC3339))
}

// ========== Version Negotiation ==========

// VersionNegotiator handles content negotiation for versions
type VersionNegotiator struct {
	supportedVersions map[string]APIVersion
	mu                sync.RWMutex
}

// NewVersionNegotiator creates a version negotiator
func NewVersionNegotiator() *VersionNegotiator {
	return &VersionNegotiator{
		supportedVersions: make(map[string]APIVersion),
	}
}

// AddSupportedVersion adds a supported version
func (vn *VersionNegotiator) AddSupportedVersion(version string, meta APIVersion) {
	vn.mu.Lock()
	defer vn.mu.Unlock()

	vn.supportedVersions[version] = meta
}

// NegotiateVersion negotiates best version for request
func (vn *VersionNegotiator) NegotiateVersion(acceptHeader string, urlVersion string) (string, error) {
	vn.mu.RLock()
	defer vn.mu.RUnlock()

	// Prefer URL version if provided
	if urlVersion != "" {
		if _, exists := vn.supportedVersions[urlVersion]; exists {
			return urlVersion, nil
		}
	}

	// Parse Accept header (e.g., "application/json; version=2.0")
	if acceptHeader != "" {
		parts := strings.Split(acceptHeader, "version=")
		if len(parts) > 1 {
			headerVersion := strings.TrimSpace(strings.Split(parts[1], ";")[0])
			if _, exists := vn.supportedVersions[headerVersion]; exists {
				return headerVersion, nil
			}
		}
	}

	return "", fmt.Errorf("no supported version found")
}

// GetSupportedVersions returns list of supported versions
func (vn *VersionNegotiator) GetSupportedVersions() []string {
	vn.mu.RLock()
	defer vn.mu.RUnlock()

	versions := make([]string, 0, len(vn.supportedVersions))
	for v := range vn.supportedVersions {
		versions = append(versions, v)
	}

	return versions
}

// ========== API Version Manager ==========

// APIVersionManager manages all versioning aspects
type APIVersionManager struct {
	urlRouter     *URLVersionRouter
	headerRouter  *HeaderVersionRouter
	deprecations  *DeprecationHandler
	negotiator    *VersionNegotiator
	middleware    *VersioningMiddleware
}

// NewAPIVersionManager creates version manager
func NewAPIVersionManager(defaultVersion APIVersion) *APIVersionManager {
	headerRouter := NewHeaderVersionRouter(defaultVersion.String())

	return &APIVersionManager{
		urlRouter:    NewURLVersionRouter(),
		headerRouter: headerRouter,
		deprecations: NewDeprecationHandler(),
		negotiator:   NewVersionNegotiator(),
		middleware:   NewVersioningMiddleware(defaultVersion),
	}
}

// RegisterVersion registers a versioned API
func (avm *APIVersionManager) RegisterVersion(version string, handler http.HandlerFunc, meta APIVersion) error {
	if err := avm.urlRouter.RegisterVersion(version, handler, meta); err != nil {
		return err
	}

	if err := avm.headerRouter.RegisterVersion(version, handler, meta); err != nil {
		return err
	}

	avm.negotiator.AddSupportedVersion(version, meta)
	return nil
}

// MarkDeprecated marks a version as deprecated
func (avm *APIVersionManager) MarkDeprecated(version string, replaceWith string, sunsetDate time.Time) {
	notice := &DeprecationNotice{
		Version:     version,
		ReplaceWith: replaceWith,
		SunsetDate:  sunsetDate,
		Message:     fmt.Sprintf("API version %s is deprecated. Use %s instead.", version, replaceWith),
	}

	avm.deprecations.RegisterDeprecation(version, notice)
}

// ========== Version Migration Helper ==========

// MigrationGuide provides migration information
type MigrationGuide struct {
	FromVersion string
	ToVersion   string
	Changes     []string
	Examples    map[string]string
}

// GetMigrationGuides returns migration guides
func GetMigrationGuides() []MigrationGuide {
	return []MigrationGuide{
		{
			FromVersion: "v1.0",
			ToVersion:   "v2.0",
			Changes: []string{
				"User.id changed from string to int64",
				"Response wrapped in 'data' field",
				"Timestamps now in milliseconds",
				"Removed deprecated endpoints",
			},
			Examples: map[string]string{
				"response_format": `
// v1.0: {"id": "123", "name": "John"}
// v2.0: {"data": {"id": 123, "name": "John"}}
`,
				"timestamp_format": `
// v1.0: {"created": "2024-01-01T00:00:00Z"}
// v2.0: {"created": 1704067200000}
`,
			},
		},
		{
			FromVersion: "v2.0",
			ToVersion:   "v3.0",
			Changes: []string{
				"Authentication now requires Bearer token",
				"Rate limits per user instead of per IP",
				"New webhook system",
				"GraphQL endpoint added",
			},
		},
	}
}

// ========== Versioning Best Practices ==========

// GetVersioningBestPractices returns documentation
func GetVersioningBestPractices() map[string]string {
	return map[string]string{
		"url_versioning": `
URL path versioning:
- GET /api/v1/users
- GET /api/v2/users
- Clear and simple
- Visible in logs`,

		"header_versioning": `
Accept header versioning:
- Accept: application/json; version=1.0
- Cleaner URLs
- Content negotiation
- Can support multiple versions`,

		"backward_compatibility": `
Maintain backward compatibility:
- Add new fields (don't remove old ones)
- Keep old endpoints running
- Use adapters for format changes
- Document migrations`,

		"deprecation_strategy": `
Deprecation process:
1. Announce deprecation (header + docs)
2. Set sunset date (6-12 months out)
3. Monitor usage
4. Remove after sunset date
5. Provide migration guides`,

		"version_numbers": `
Semantic versioning:
- MAJOR: Breaking changes (v1.0 -> v2.0)
- MINOR: New features (v2.0 -> v2.1)
- PATCH: Bug fixes (v2.1 -> v2.1.1)`,

		"migration_support": `
Support version migration:
- Provide transformation adapters
- Include examples in documentation
- Offer migration tools
- Communicate timeline clearly`,
	}
}

func main() {}
