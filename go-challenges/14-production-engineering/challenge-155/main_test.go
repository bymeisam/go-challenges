package main

import (
	"fmt"
	"strings"
	"testing"
)

func TestContainerRegister(t *testing.T) {
	container := NewContainer()

	logger := &SimpleLogger{}
	err := container.Register("logger", logger)

	if err != nil {
		t.Fatalf("Failed to register: %v", err)
	}

	// Duplicate should fail
	err = container.Register("logger", logger)
	if err == nil {
		t.Error("Should not allow duplicate registration")
	}

	t.Log("✓ Container registration works!")
}

func TestContainerGet(t *testing.T) {
	container := NewContainer()

	logger := &SimpleLogger{}
	container.Register("logger", logger)

	svc, err := container.Get("logger")
	if err != nil {
		t.Fatalf("Failed to get: %v", err)
	}

	if svc != logger {
		t.Error("Should return registered service")
	}

	t.Log("✓ Container Get works!")
}

func TestContainerGetMissing(t *testing.T) {
	container := NewContainer()

	_, err := container.Get("missing")
	if err == nil {
		t.Error("Should error on missing service")
	}

	t.Log("✓ Container handles missing services!")
}

func TestContainerFactory(t *testing.T) {
	container := NewContainer()

	container.RegisterFactory("logger", func() *SimpleLogger {
		return &SimpleLogger{}
	})

	svc, err := container.Get("logger")
	if err != nil {
		t.Fatalf("Failed to get from factory: %v", err)
	}

	if svc == nil {
		t.Error("Factory should create service")
	}

	t.Log("✓ Container factory works!")
}

func TestGolangciLintConfig(t *testing.T) {
	config := GetGolangciLintConfig()

	expectedElements := []string{
		".golangci.yml",
		"linters:",
		"enable:",
		"deadline:",
		"gofmt",
		"gosec",
		"govet",
	}

	for _, element := range expectedElements {
		if !strings.Contains(config, element) {
			t.Errorf("Config missing %s", element)
		}
	}

	t.Log("✓ Golangci-lint config is valid!")
}

func TestMockGenerator(t *testing.T) {
	gen := NewMockGenerator("testpkg")
	gen.AddInterface("UserService")
	gen.AddInterface("Repository")

	output := gen.Generate()

	if !strings.Contains(output, "testpkg") {
		t.Error("Should contain package name")
	}

	if !strings.Contains(output, "MockUserService") {
		t.Error("Should generate UserService mock")
	}

	if !strings.Contains(output, "MockRepository") {
		t.Error("Should generate Repository mock")
	}

	if !strings.Contains(output, "DO NOT EDIT") {
		t.Error("Should have generated file notice")
	}

	t.Log("✓ Mock generator works!")
}

func TestGoGenerateDirective(t *testing.T) {
	directives := GetGoGenerateDirective()

	expectedKeys := []string{"mockgen", "stringer", "go-bindata", "easyjson", "protoc", "sql-migrate"}

	for _, key := range expectedKeys {
		if _, exists := directives[key]; !exists {
			t.Errorf("Should have %s directive", key)
		}

		if directives[key] == "" {
			t.Errorf("Directive %s should not be empty", key)
		}
	}

	// Check format
	for key, directive := range directives {
		if !strings.Contains(directive, "//go:generate") {
			t.Errorf("Directive %s should contain //go:generate", key)
		}
	}

	t.Log("✓ Go generate directives are valid!")
}

func TestUserRepository(t *testing.T) {
	repo := NewUserRepository()

	// Save
	err := repo.Save("user1", "John Doe")
	if err != nil {
		t.Fatalf("Failed to save: %v", err)
	}

	// Get
	user, err := repo.Get("user1")
	if err != nil {
		t.Fatalf("Failed to get: %v", err)
	}

	if user != "John Doe" {
		t.Error("Should retrieve saved user")
	}

	// Missing user
	_, err = repo.Get("missing")
	if err == nil {
		t.Error("Should error on missing user")
	}

	t.Log("✓ User repository works!")
}

func TestUserService(t *testing.T) {
	logger := &SimpleLogger{}
	repo := NewUserRepository()
	repo.Save("user1", "John Doe")

	service := NewUserService(logger, repo)

	user, err := service.GetUser("user1")
	if err != nil {
		t.Fatalf("Failed to get user: %v", err)
	}

	if user != "John Doe" {
		t.Error("Should return correct user")
	}

	// Save user
	err = service.SaveUser("user2", "Jane Doe")
	if err != nil {
		t.Fatalf("Failed to save: %v", err)
	}

	// Verify saved
	user, _ = service.GetUser("user2")
	if user != "Jane Doe" {
		t.Error("Should save and retrieve user")
	}

	t.Log("✓ User service with DI works!")
}

func TestAnalyzeCode(t *testing.T) {
	code := `
package main

// This is a comment
func Hello() {
	fmt.Println("Hello")
}

// Another comment


func World() {
	fmt.Println("World")
}
`

	metrics := AnalyzeCode(code)

	if metrics.TotalLines == 0 {
		t.Error("Should have total lines")
	}

	if metrics.CodeLines == 0 {
		t.Error("Should have code lines")
	}

	if metrics.CommentLines != 2 {
		t.Errorf("Expected 2 comment lines, got %d", metrics.CommentLines)
	}

	if metrics.BlankLines == 0 {
		t.Error("Should have blank lines")
	}

	t.Log("✓ Code analysis works!")
}

func TestTestHelper(t *testing.T) {
	helper := &TestHelper{}

	// AssertEqual
	if !helper.AssertEqual("hello", "hello") {
		t.Error("AssertEqual should work for equal values")
	}

	if helper.AssertEqual("hello", "world") {
		t.Error("AssertEqual should fail for different values")
	}

	// AssertNil
	var nilVal interface{} = nil
	if !helper.AssertNil(nilVal) {
		t.Error("AssertNil should detect nil")
	}

	// AssertNotNil
	notNilVal := "not nil"
	if !helper.AssertNotNil(notNilVal) {
		t.Error("AssertNotNil should work")
	}

	t.Log("✓ Test helper works!")
}

func TestGetLintRules(t *testing.T) {
	rules := GetLintRules()

	if len(rules) == 0 {
		t.Error("Should have lint rules")
	}

	expectedRules := []string{
		"unused-variables",
		"error-checking",
		"cyclomatic-complexity",
		"naming-conventions",
		"security-issues",
		"code-duplication",
	}

	ruleNames := make([]string, 0)
	for _, rule := range rules {
		ruleNames = append(ruleNames, rule.Name)
	}

	for _, expected := range expectedRules {
		found := false
		for _, name := range ruleNames {
			if name == expected {
				found = true
				break
			}
		}

		if !found {
			t.Errorf("Should have rule %s", expected)
		}
	}

	// Check all rules have required fields
	for _, rule := range rules {
		if rule.Name == "" {
			t.Error("Rule should have name")
		}

		if rule.Description == "" {
			t.Error("Rule should have description")
		}

		if rule.Severity == "" {
			t.Error("Rule should have severity")
		}
	}

	t.Log("✓ Lint rules are complete!")
}

func TestDependencyInjectionWithMultipleServices(t *testing.T) {
	container := NewContainer()

	logger := &SimpleLogger{}
	repo := NewUserRepository()

	container.Register("logger", logger)
	container.Register("repository", repo)

	// Create service manually with DI
	service := NewUserService(logger, repo)

	err := service.SaveUser("test", "Test User")
	if err != nil {
		t.Fatalf("Failed: %v", err)
	}

	user, err := service.GetUser("test")
	if err != nil {
		t.Fatalf("Failed: %v", err)
	}

	if user != "Test User" {
		t.Error("Should retrieve saved user")
	}

	t.Log("✓ Dependency injection with multiple services works!")
}

func TestLinterIntegration(t *testing.T) {
	config := GetGolangciLintConfig()
	rules := GetLintRules()

	// Verify config covers important rules
	importantRules := []string{"errcheck", "gosimple", "govet", "unused"}

	for _, rule := range importantRules {
		if !strings.Contains(config, rule) {
			t.Logf("Note: Config might not include %s", rule)
		}
	}

	if len(rules) < 5 {
		t.Error("Should have multiple lint rules")
	}

	t.Log("✓ Linter integration is valid!")
}

func TestCodeQualityMetrics(t *testing.T) {
	code := `package main

// Comment line 1
func Test() {
	x := 1
	// Comment line 2
	if x > 0 {
		fmt.Println("hello")
	}

	// Comment line 3
}`

	metrics := AnalyzeCode(code)

	if metrics.CommentLines < 3 {
		t.Errorf("Expected at least 3 comment lines, got %d", metrics.CommentLines)
	}

	if metrics.CodeLines < 5 {
		t.Errorf("Expected at least 5 code lines, got %d", metrics.CodeLines)
	}

	if metrics.BlankLines < 1 {
		t.Errorf("Expected at least 1 blank line, got %d", metrics.BlankLines)
	}

	t.Logf("Metrics: Total=%d, Code=%d, Comments=%d, Blank=%d",
		metrics.TotalLines, metrics.CodeLines, metrics.CommentLines, metrics.BlankLines)

	t.Log("✓ Code quality metrics are accurate!")
}

func TestMockGeneratorWithMultipleInterfaces(t *testing.T) {
	gen := NewMockGenerator("api")

	interfaces := []string{"Handler", "Middleware", "Router", "Validator"}
	for _, iface := range interfaces {
		gen.AddInterface(iface)
	}

	output := gen.Generate()

	for _, iface := range interfaces {
		mockName := fmt.Sprintf("Mock%s", iface)
		if !strings.Contains(output, mockName) {
			t.Errorf("Should generate mock for %s", iface)
		}
	}

	t.Log("✓ Mock generator handles multiple interfaces!")
}

func TestRepositoryInterface(t *testing.T) {
	repo := NewUserRepository()

	// Test interface compliance
	var _ Repository = repo

	err := repo.Save("1", map[string]string{"name": "John"})
	if err != nil {
		t.Fatalf("Failed to save: %v", err)
	}

	data, err := repo.Get("1")
	if err != nil {
		t.Fatalf("Failed to get: %v", err)
	}

	if data == nil {
		t.Error("Should retrieve data")
	}

	t.Log("✓ Repository interface works!")
}

func BenchmarkContainerGet(b *testing.B) {
	container := NewContainer()
	container.Register("logger", &SimpleLogger{})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		container.Get("logger")
	}
}

func BenchmarkAnalyzeCode(b *testing.B) {
	code := `
package main

// Comment
func Test() {
	x := 1
	fmt.Println(x)
}
`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		AnalyzeCode(code)
	}
}

func BenchmarkUserServiceGetUser(b *testing.B) {
	logger := &SimpleLogger{}
	repo := NewUserRepository()
	repo.Save("user1", "John Doe")

	service := NewUserService(logger, repo)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		service.GetUser("user1")
	}
}
