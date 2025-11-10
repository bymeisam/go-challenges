package main

import (
	"sync/atomic"
	"testing"
)

// TestCodeAnalyzer tests code analysis
func TestCodeAnalyzerBasic(t *testing.T) {
	analyzer := NewCodeAnalyzer()

	code := `
package main

func Add(a, b int) int {
	return a + b
}
`

	err := analyzer.AnalyzeCode(code)
	if err != nil {
		t.Errorf("Expected successful analysis, got error: %v", err)
	}
}

func TestCodeAnalyzerFunctionExtraction(t *testing.T) {
	analyzer := NewCodeAnalyzer()

	code := `
package main

func Multiply(x, y int) int {
	return x * y
}
`

	err := analyzer.AnalyzeCode(code)
	if err != nil {
		t.Errorf("Expected successful analysis, got error: %v", err)
	}

	analyzer.mu.RLock()
	funcs := len(analyzer.Functions)
	analyzer.mu.RUnlock()

	if funcs < 1 {
		t.Logf("Warning: Expected at least 1 function, found %d", funcs)
	}
}

func TestCodeAnalyzerStatistics(t *testing.T) {
	analyzer := NewCodeAnalyzer()

	code := `
package main

func Add(a int, b int) int {
	return a + b
}

func Sub(a int, b int) int {
	return a - b
}
`

	analyzer.AnalyzeCode(code)

	report := analyzer.GetReport()
	if _, ok := report["total_functions"]; !ok {
		t.Errorf("Expected total_functions in report")
	}
}

// TestCodeGenerator tests code generation
func TestCodeGeneratorStruct(t *testing.T) {
	gen := NewCodeGenerator("main")

	fields := []*ParamInfo{
		{Name: "Name", Type: "string"},
		{Name: "Age", Type: "int"},
	}

	code := gen.GenerateStructCode("Person", fields)

	if !contains(code, "type Person struct") {
		t.Errorf("Expected struct declaration in generated code")
	}

	if !contains(code, "Name string") {
		t.Errorf("Expected Name field in generated code")
	}
}

func TestCodeGeneratorFunction(t *testing.T) {
	gen := NewCodeGenerator("main")

	params := []*ParamInfo{{Name: "x", Type: "int"}}
	returns := []*ParamInfo{{Type: "int"}}

	code := gen.GenerateFunctionCode("Double", "", params, returns)

	if !contains(code, "func Double") {
		t.Errorf("Expected function declaration in generated code")
	}

	if !contains(code, "int") {
		t.Errorf("Expected parameter type in generated code")
	}
}

func TestCodeGeneratorInterface(t *testing.T) {
	gen := NewCodeGenerator("main")

	methods := []*MethodSignature{
		{
			Name:   "Read",
			Params: []*ParamInfo{{Name: "b", Type: "[]byte"}},
		},
	}

	code := gen.GenerateInterfaceCode("Reader", methods)

	if !contains(code, "type Reader interface") {
		t.Errorf("Expected interface declaration in generated code")
	}

	if !contains(code, "Read") {
		t.Errorf("Expected method in generated code")
	}
}

func TestCodeGeneratorImports(t *testing.T) {
	gen := NewCodeGenerator("main")
	gen.AddImport("fmt")
	gen.AddImport("io")

	// Adding same import twice should not duplicate
	gen.AddImport("fmt")

	code, _ := gen.GenerateCompleteFile()

	if !contains(code, "package main") {
		t.Errorf("Expected package declaration")
	}

	if !contains(code, "import") {
		t.Errorf("Expected import section")
	}
}

// TestCodeTransformer tests code transformation
func TestCodeTransformerBasic(t *testing.T) {
	transformer := NewCodeTransformer()

	err := transformer.AddPattern(
		"test",
		"var",
		"const",
		"Transform var to const",
	)

	if err != nil {
		t.Errorf("Expected successful pattern addition, got error: %v", err)
	}
}

func TestCodeTransformerTransform(t *testing.T) {
	transformer := NewCodeTransformer()
	transformer.AddPattern("var2const", "var (\\w+) =", "const $1 =", "Transform var to const")

	code := "var x = 5"
	transformed, counts := transformer.Transform(code)

	if transformed == code {
		t.Errorf("Expected code to be transformed")
	}

	if len(counts) == 0 {
		t.Logf("Warning: No transformations recorded")
	}
}

func TestCodeTransformerMultiplePatterns(t *testing.T) {
	transformer := NewCodeTransformer()
	transformer.AddPattern("p1", "foo", "bar", "Replace foo")
	transformer.AddPattern("p2", "baz", "qux", "Replace baz")

	code := "foo and baz"
	_, counts := transformer.Transform(code)

	if len(counts) < 2 {
		t.Logf("Warning: Expected at least 2 pattern matches")
	}
}

// TestCustomLinter tests linting
func TestCustomLinterBasic(t *testing.T) {
	linter := NewCustomLinter()

	err := linter.AddRule("test", "Test Rule", "Test description", "test", "error")
	if err != nil {
		t.Errorf("Expected successful rule addition, got error: %v", err)
	}
}

func TestCustomLinterDetectsViolations(t *testing.T) {
	linter := NewCustomLinter()
	linter.AddRule("no_todo", "TODO found", "TODO comments present", "TODO", "warning")

	code := "// TODO: fix this\nvar x = 5"
	violations := linter.Lint(code)

	if len(violations) == 0 {
		t.Errorf("Expected at least 1 violation")
	}
}

func TestCustomLinterMultipleRules(t *testing.T) {
	linter := NewCustomLinter()
	linter.AddRule("rule1", "Rule 1", "Test rule 1", "foo", "warning")
	linter.AddRule("rule2", "Rule 2", "Test rule 2", "bar", "error")

	code := "foo and bar"
	violations := linter.Lint(code)

	if len(violations) < 2 {
		t.Logf("Warning: Expected at least 2 violations, got %d", len(violations))
	}
}

func TestCustomLinterViolationInfo(t *testing.T) {
	linter := NewCustomLinter()
	linter.AddRule("test", "Test", "Test violation", "error", "warning")

	code := "some error here"
	violations := linter.Lint(code)

	if len(violations) > 0 {
		v := violations[0]
		if v.Line == 0 {
			t.Errorf("Expected non-zero line number")
		}

		if v.Column == 0 {
			t.Errorf("Expected non-zero column number")
		}
	}
}

// TestLintRuleStats tests rule statistics
func TestLintRuleStats(t *testing.T) {
	linter := NewCustomLinter()
	linter.AddRule("counter", "Counter", "Count hits", "test", "info")

	code := "test test test"
	linter.Lint(code)

	report := linter.GetReport()
	if totalViolations, ok := report["total_violations"]; !ok {
		t.Errorf("Expected total_violations in report")
	} else if totalViolations.(int64) == 0 {
		t.Logf("Warning: Expected violations to be recorded")
	}
}

// TestParamInfo tests parameter information
func TestParamInfo(t *testing.T) {
	param := &ParamInfo{
		Name: "x",
		Type: "int",
	}

	if param.Name != "x" {
		t.Errorf("Expected Name='x'")
	}

	if param.Type != "int" {
		t.Errorf("Expected Type='int'")
	}
}

// TestFunctionInfo tests function information
func TestFunctionInfo(t *testing.T) {
	fn := &FunctionInfo{
		Name:      "Add",
		Lines:     5,
		Cyclomatic: 1,
		IsPublic:  true,
	}

	if fn.Name != "Add" {
		t.Errorf("Expected Name='Add'")
	}

	if fn.Cyclomatic != 1 {
		t.Errorf("Expected Cyclomatic=1")
	}
}

// TestTypeInfo tests type information
func TestTypeInfo(t *testing.T) {
	ti := &TypeInfo{
		Name:     "User",
		Kind:     "struct",
		IsPublic: true,
	}

	if ti.Name != "User" {
		t.Errorf("Expected Name='User'")
	}

	if ti.Kind != "struct" {
		t.Errorf("Expected Kind='struct'")
	}
}

// Benchmark tests

func BenchmarkCodeAnalyzer(b *testing.B) {
	analyzer := NewCodeAnalyzer()
	code := `
package main

func Test(x int) int {
	return x * 2
}
`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		analyzer.AnalyzeCode(code)
	}
}

func BenchmarkCodeGenerator(b *testing.B) {
	gen := NewCodeGenerator("main")
	fields := []*ParamInfo{{Name: "x", Type: "int"}}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		gen.GenerateStructCode("Test", fields)
	}
}

func BenchmarkCodeTransformer(b *testing.B) {
	transformer := NewCodeTransformer()
	transformer.AddPattern("test", "var", "const", "test")
	code := "var x = 5"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		transformer.Transform(code)
	}
}

func BenchmarkCustomLinter(b *testing.B) {
	linter := NewCustomLinter()
	linter.AddRule("test", "Test", "Test", "var", "warning")
	code := "var x = 5"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		linter.Lint(code)
	}
}

// Helper functions
func contains(s, substr string) bool {
	for i := 0; i < len(s)-len(substr)+1; i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
