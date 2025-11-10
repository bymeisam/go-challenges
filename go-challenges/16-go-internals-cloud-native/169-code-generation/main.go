package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"regexp"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
)

// Challenge 169: Code Generation with AST
// AST Parsing, Code Generation, AST Transformation, Custom Analyzers

// ===== 1. Code Analyzer =====

type CodeAnalyzer struct {
	Functions    map[string]*FunctionInfo
	Types        map[string]*TypeInfo
	Interfaces   map[string]*InterfaceInfo
	Statistics   *CodeStatistics
	mu           sync.RWMutex
}

type FunctionInfo struct {
	Name      string
	Receiver  string
	Params    []*ParamInfo
	Returns   []*ParamInfo
	Lines     int
	Cyclomatic int
	IsPublic  bool
}

type ParamInfo struct {
	Name string
	Type string
}

type TypeInfo struct {
	Name      string
	Kind      string // struct, interface, alias
	Fields    []*ParamInfo
	Methods   []string
	IsPublic  bool
}

type InterfaceInfo struct {
	Name    string
	Methods []*MethodSignature
	IsPublic bool
}

type MethodSignature struct {
	Name    string
	Params  []*ParamInfo
	Returns []*ParamInfo
}

type CodeStatistics struct {
	TotalFunctions   int64
	TotalTypes       int64
	TotalLines       int64
	AverageFunctionSize int64
	ComplexityMetrics map[string]int64
	mu sync.RWMutex
}

func NewCodeAnalyzer() *CodeAnalyzer {
	return &CodeAnalyzer{
		Functions:  make(map[string]*FunctionInfo),
		Types:      make(map[string]*TypeInfo),
		Interfaces: make(map[string]*InterfaceInfo),
		Statistics: &CodeStatistics{
			ComplexityMetrics: make(map[string]int64),
		},
	}
}

func (ca *CodeAnalyzer) AnalyzeCode(sourceCode string) error {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "src.go", sourceCode, parser.ParseComments)
	if err != nil {
		return err
	}

	v := &analyzerVisitor{
		analyzer: ca,
		fset:     fset,
	}

	ast.Walk(v, file)
	ca.updateStatistics()

	return nil
}

type analyzerVisitor struct {
	analyzer *CodeAnalyzer
	fset     *token.FileSet
}

func (v *analyzerVisitor) Visit(node ast.Node) ast.Visitor {
	if node == nil {
		return v
	}

	switch n := node.(type) {
	case *ast.FuncDecl:
		v.analyzeFunctionDecl(n)

	case *ast.TypeSpec:
		v.analyzeTypeSpec(n)

	case *ast.InterfaceType:
		v.analyzeInterfaceType(n)
	}

	return v
}

func (v *analyzerVisitor) analyzeFunctionDecl(fn *ast.FuncDecl) {
	info := &FunctionInfo{
		Name:     fn.Name.Name,
		IsPublic: ast.IsExported(fn.Name.Name),
		Lines:    v.fset.Position(fn.End()).Line - v.fset.Position(fn.Pos()).Line,
	}

	if fn.Recv != nil && len(fn.Recv.List) > 0 {
		info.Receiver = v.extractTypeName(fn.Recv.List[0].Type)
	}

	// Extract parameters
	if fn.Type.Params != nil {
		for _, param := range fn.Type.Params.List {
			typeName := v.extractTypeName(param.Type)
			if len(param.Names) == 0 {
				info.Params = append(info.Params, &ParamInfo{Type: typeName})
			} else {
				for _, name := range param.Names {
					info.Params = append(info.Params, &ParamInfo{Name: name.Name, Type: typeName})
				}
			}
		}
	}

	// Extract returns
	if fn.Type.Results != nil {
		for _, result := range fn.Type.Results.List {
			typeName := v.extractTypeName(result.Type)
			if len(result.Names) == 0 {
				info.Returns = append(info.Returns, &ParamInfo{Type: typeName})
			} else {
				for _, name := range result.Names {
					info.Returns = append(info.Returns, &ParamInfo{Name: name.Name, Type: typeName})
				}
			}
		}
	}

	// Calculate cyclomatic complexity
	info.Cyclomatic = v.calculateCyclomaticComplexity(fn.Body)

	v.analyzer.mu.Lock()
	v.analyzer.Functions[fn.Name.Name] = info
	v.analyzer.mu.Unlock()

	atomic.AddInt64(&v.analyzer.Statistics.TotalFunctions, 1)
	atomic.AddInt64(&v.analyzer.Statistics.TotalLines, int64(info.Lines))
}

func (v *analyzerVisitor) analyzeTypeSpec(spec *ast.TypeSpec) {
	info := &TypeInfo{
		Name:      spec.Name.Name,
		IsPublic:  ast.IsExported(spec.Name.Name),
		Methods:   make([]string, 0),
	}

	switch t := spec.Type.(type) {
	case *ast.StructType:
		info.Kind = "struct"
		if t.Fields != nil {
			for _, field := range t.Fields.List {
				typeName := v.extractTypeName(field.Type)
				for _, name := range field.Names {
					info.Fields = append(info.Fields, &ParamInfo{Name: name.Name, Type: typeName})
				}
			}
		}
	case *ast.InterfaceType:
		info.Kind = "interface"
	default:
		info.Kind = "alias"
	}

	v.analyzer.mu.Lock()
	v.analyzer.Types[spec.Name.Name] = info
	v.analyzer.mu.Unlock()

	atomic.AddInt64(&v.analyzer.Statistics.TotalTypes, 1)
}

func (v *analyzerVisitor) analyzeInterfaceType(iface *ast.InterfaceType) {
	// Interface analysis handled during type spec analysis
}

func (v *analyzerVisitor) extractTypeName(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.SelectorExpr:
		return fmt.Sprintf("%s.%s", v.extractTypeName(t.X), t.Sel.Name)
	case *ast.StarExpr:
		return "*" + v.extractTypeName(t.X)
	case *ast.ArrayType:
		return "[]" + v.extractTypeName(t.Elt)
	case *ast.MapType:
		return fmt.Sprintf("map[%s]%s", v.extractTypeName(t.Key), v.extractTypeName(t.Value))
	default:
		return "unknown"
	}
}

func (v *analyzerVisitor) calculateCyclomaticComplexity(body *ast.BlockStmt) int {
	if body == nil {
		return 1
	}

	complexity := 1
	ast.Inspect(body, func(n ast.Node) bool {
		switch n.(type) {
		case *ast.IfStmt, *ast.ForStmt, *ast.RangeStmt, *ast.CaseClause:
			complexity++
		}
		return true
	})

	return complexity
}

func (ca *CodeAnalyzer) updateStatistics() {
	ca.mu.RLock()
	totalFuncs := int64(len(ca.Functions))
	totalTypes := int64(len(ca.Types))
	totalLines := atomic.LoadInt64(&ca.Statistics.TotalLines)
	ca.mu.RUnlock()

	if totalFuncs > 0 {
		avgSize := totalLines / totalFuncs
		atomic.StoreInt64(&ca.Statistics.AverageFunctionSize, avgSize)
	}
}

func (ca *CodeAnalyzer) GetReport() map[string]interface{} {
	ca.mu.RLock()
	defer ca.mu.RUnlock()

	return map[string]interface{}{
		"total_functions":     atomic.LoadInt64(&ca.Statistics.TotalFunctions),
		"total_types":         atomic.LoadInt64(&ca.Statistics.TotalTypes),
		"total_lines":         atomic.LoadInt64(&ca.Statistics.TotalLines),
		"avg_function_size":   atomic.LoadInt64(&ca.Statistics.AverageFunctionSize),
		"functions":           len(ca.Functions),
		"types":               len(ca.Types),
		"interfaces":          len(ca.Interfaces),
	}
}

// ===== 2. Code Generator =====

type CodeGenerator struct {
	Package string
	Imports []string
	code    strings.Builder
	mu      sync.Mutex
}

func NewCodeGenerator(pkgName string) *CodeGenerator {
	return &CodeGenerator{
		Package: pkgName,
		Imports: make([]string, 0),
	}
}

func (cg *CodeGenerator) AddImport(importPath string) {
	cg.mu.Lock()
	defer cg.mu.Unlock()

	for _, imp := range cg.Imports {
		if imp == importPath {
			return
		}
	}
	cg.Imports = append(cg.Imports, importPath)
}

func (cg *CodeGenerator) GenerateStructCode(name string, fields []*ParamInfo) string {
	cg.mu.Lock()
	defer cg.mu.Unlock()

	var buf bytes.Buffer
	buf.WriteString("type ")
	buf.WriteString(name)
	buf.WriteString(" struct {\n")

	for _, field := range fields {
		buf.WriteString("\t")
		buf.WriteString(field.Name)
		buf.WriteString(" ")
		buf.WriteString(field.Type)
		buf.WriteString("\n")
	}

	buf.WriteString("}\n")
	return buf.String()
}

func (cg *CodeGenerator) GenerateFunctionCode(name, receiver string, params, returns []*ParamInfo) string {
	cg.mu.Lock()
	defer cg.mu.Unlock()

	var buf bytes.Buffer
	buf.WriteString("func ")

	if receiver != "" {
		buf.WriteString("(r *")
		buf.WriteString(receiver)
		buf.WriteString(") ")
	}

	buf.WriteString(name)
	buf.WriteString("(")

	for i, param := range params {
		if i > 0 {
			buf.WriteString(", ")
		}
		if param.Name != "" {
			buf.WriteString(param.Name)
			buf.WriteString(" ")
		}
		buf.WriteString(param.Type)
	}

	buf.WriteString(")")

	if len(returns) > 0 {
		if len(returns) == 1 {
			buf.WriteString(" ")
			buf.WriteString(returns[0].Type)
		} else {
			buf.WriteString(" (")
			for i, ret := range returns {
				if i > 0 {
					buf.WriteString(", ")
				}
				buf.WriteString(ret.Type)
			}
			buf.WriteString(")")
		}
	}

	buf.WriteString(" {\n\t// TODO: implement\n}\n")
	return buf.String()
}

func (cg *CodeGenerator) GenerateInterfaceCode(name string, methods []*MethodSignature) string {
	cg.mu.Lock()
	defer cg.mu.Unlock()

	var buf bytes.Buffer
	buf.WriteString("type ")
	buf.WriteString(name)
	buf.WriteString(" interface {\n")

	for _, method := range methods {
		buf.WriteString("\t")
		buf.WriteString(method.Name)
		buf.WriteString("(")

		for i, param := range method.Params {
			if i > 0 {
				buf.WriteString(", ")
			}
			if param.Name != "" {
				buf.WriteString(param.Name)
				buf.WriteString(" ")
			}
			buf.WriteString(param.Type)
		}

		buf.WriteString(")")

		if len(method.Returns) > 0 {
			if len(method.Returns) == 1 {
				buf.WriteString(" ")
				buf.WriteString(method.Returns[0].Type)
			} else {
				buf.WriteString(" (")
				for i, ret := range method.Returns {
					if i > 0 {
						buf.WriteString(", ")
					}
					buf.WriteString(ret.Type)
				}
				buf.WriteString(")")
			}
		}
		buf.WriteString("\n")
	}

	buf.WriteString("}\n")
	return buf.String()
}

func (cg *CodeGenerator) GenerateCompleteFile() (string, error) {
	cg.mu.Lock()
	defer cg.mu.Unlock()

	var buf bytes.Buffer
	buf.WriteString("package ")
	buf.WriteString(cg.Package)
	buf.WriteString("\n\n")

	if len(cg.Imports) > 0 {
		buf.WriteString("import (\n")
		for _, imp := range cg.Imports {
			buf.WriteString("\t\"")
			buf.WriteString(imp)
			buf.WriteString("\"\n")
		}
		buf.WriteString(")\n\n")
	}

	buf.WriteString(cg.code.String())

	return buf.String(), nil
}

// ===== 3. Code Transformer =====

type CodeTransformer struct {
	patterns map[string]*TransformPattern
	mu       sync.RWMutex
}

type TransformPattern struct {
	Name         string
	Pattern      string
	Replacement  string
	Description  string
	regex        *regexp.Regexp
	Transformations int64
}

func NewCodeTransformer() *CodeTransformer {
	return &CodeTransformer{
		patterns: make(map[string]*TransformPattern),
	}
}

func (ct *CodeTransformer) AddPattern(name, pattern, replacement, description string) error {
	ct.mu.Lock()
	defer ct.mu.Unlock()

	regex, err := regexp.Compile(pattern)
	if err != nil {
		return err
	}

	ct.patterns[name] = &TransformPattern{
		Name:        name,
		Pattern:     pattern,
		Replacement: replacement,
		Description: description,
		regex:       regex,
	}

	return nil
}

func (ct *CodeTransformer) Transform(code string) (string, map[string]int64) {
	ct.mu.RLock()
	patterns := ct.patterns
	ct.mu.RUnlock()

	result := code
	transformCounts := make(map[string]int64)

	for name, pattern := range patterns {
		matches := pattern.regex.FindAllString(result, -1)
		count := int64(len(matches))

		if count > 0 {
			result = pattern.regex.ReplaceAllString(result, pattern.Replacement)
			transformCounts[name] = count
			atomic.AddInt64(&pattern.Transformations, count)
		}
	}

	return result, transformCounts
}

func (ct *CodeTransformer) GetStats() map[string]int64 {
	ct.mu.RLock()
	defer ct.mu.RUnlock()

	stats := make(map[string]int64)
	for name, pattern := range ct.patterns {
		stats[name] = atomic.LoadInt64(&pattern.Transformations)
	}
	return stats
}

// ===== 4. Linter =====

type CustomLinter struct {
	rules      map[string]*LintRule
	violations []*LintViolation
	mu         sync.RWMutex
}

type LintRule struct {
	ID          string
	Name        string
	Description string
	Pattern     string
	Severity    string // error, warning, info
	regex       *regexp.Regexp
	Hits        int64
}

type LintViolation struct {
	Rule    string
	Line    int
	Column  int
	Message string
	Code    string
}

func NewCustomLinter() *CustomLinter {
	return &CustomLinter{
		rules:      make(map[string]*LintRule),
		violations: make([]*LintViolation, 0),
	}
}

func (cl *CustomLinter) AddRule(id, name, description, pattern, severity string) error {
	cl.mu.Lock()
	defer cl.mu.Unlock()

	regex, err := regexp.Compile(pattern)
	if err != nil {
		return err
	}

	cl.rules[id] = &LintRule{
		ID:          id,
		Name:        name,
		Description: description,
		Pattern:     pattern,
		Severity:    severity,
		regex:       regex,
	}

	return nil
}

func (cl *CustomLinter) Lint(code string) []*LintViolation {
	cl.mu.RLock()
	rules := cl.rules
	cl.mu.RUnlock()

	violations := make([]*LintViolation, 0)
	lines := strings.Split(code, "\n")

	for ruleID, rule := range rules {
		for lineNum, line := range lines {
			matches := rule.regex.FindAllStringIndex(line, -1)
			for _, match := range matches {
				violations = append(violations, &LintViolation{
					Rule:    ruleID,
					Line:    lineNum + 1,
					Column:  match[0] + 1,
					Message: rule.Description,
					Code:    line,
				})
				atomic.AddInt64(&rule.Hits, 1)
			}
		}
	}

	cl.mu.Lock()
	cl.violations = violations
	cl.mu.Unlock()

	return violations
}

func (cl *CustomLinter) GetReport() map[string]interface{} {
	cl.mu.RLock()
	defer cl.mu.RUnlock()

	totalViolations := int64(0)
	violationsByRule := make(map[string]int64)

	for _, rule := range cl.rules {
		hits := atomic.LoadInt64(&rule.Hits)
		totalViolations += hits
		violationsByRule[rule.ID] = hits
	}

	return map[string]interface{}{
		"total_violations":    totalViolations,
		"violations_by_rule":  violationsByRule,
		"violation_count":     len(cl.violations),
	}
}

// ===== Main Demo =====

func main() {
	fmt.Println("=== Code Generation with AST ===\n")

	// 1. Code Analyzer
	fmt.Println("1. Code Analyzer")
	analyzer := NewCodeAnalyzer()

	sampleCode := `
package main

func Add(a, b int) int {
	if a > 0 {
		return a + b
	}
	return 0
}

type User struct {
	Name string
	Age int
}

func (u *User) GetName() string {
	return u.Name
}
`

	err := analyzer.AnalyzeCode(sampleCode)
	if err != nil {
		fmt.Printf("Analysis error: %v\n", err)
	}

	report := analyzer.GetReport()
	fmt.Printf("Analysis Report: %+v\n\n", report)

	// 2. Code Generator
	fmt.Println("2. Code Generator")
	gen := NewCodeGenerator("main")
	gen.AddImport("fmt")

	// Generate struct
	fields := []*ParamInfo{
		{Name: "ID", Type: "int"},
		{Name: "Name", Type: "string"},
	}
	structCode := gen.GenerateStructCode("Person", fields)
	fmt.Printf("Generated Struct:\n%s\n", structCode)

	// Generate function
	params := []*ParamInfo{{Name: "x", Type: "int"}}
	returns := []*ParamInfo{{Type: "int"}}
	funcCode := gen.GenerateFunctionCode("Double", "", params, returns)
	fmt.Printf("Generated Function:\n%s\n", funcCode)

	// Generate interface
	methods := []*MethodSignature{
		{
			Name:    "Read",
			Params:  []*ParamInfo{{Name: "b", Type: "[]byte"}},
			Returns: []*ParamInfo{{Type: "int"}, {Type: "error"}},
		},
	}
	ifaceCode := gen.GenerateInterfaceCode("Reader", methods)
	fmt.Printf("Generated Interface:\n%s\n", ifaceCode)

	// 3. Code Transformer
	fmt.Println("3. Code Transformer")
	transformer := NewCodeTransformer()
	transformer.AddPattern(
		"var_to_const",
		"var (\\w+) =",
		"const $1 =",
		"Transform var to const",
	)

	testCode := "var count = 10"
	transformed, counts := transformer.Transform(testCode)
	fmt.Printf("Original: %s\n", testCode)
	fmt.Printf("Transformed: %s\n", transformed)
	fmt.Printf("Transformation counts: %v\n\n", counts)

	// 4. Custom Linter
	fmt.Println("4. Custom Linter")
	linter := NewCustomLinter()
	linter.AddRule("no_todo", "Missing implementation", "TODO comments should be removed", "TODO", "warning")
	linter.AddRule("long_var", "Long variable names", "Variable names should be concise", "\\b[a-z]{20,}\\b", "info")

	lintCode := "// TODO: fix this\nvar veryLongVariableNameThatShouldBeShorter = 5"
	violations := linter.Lint(lintCode)
	fmt.Printf("Linting Results: %d violations\n", len(violations))
	for _, v := range violations {
		fmt.Printf("  Line %d, Col %d: %s\n", v.Line, v.Column, v.Message)
	}

	lintReport := linter.GetReport()
	fmt.Printf("Linter Report: %+v\n\n", lintReport)

	// 5. Summary
	fmt.Println("5. Features Summary")
	fmt.Println("  - AST-based code analysis")
	fmt.Println("  - Programmatic code generation")
	fmt.Println("  - Pattern-based code transformation")
	fmt.Println("  - Custom linting rules")
	fmt.Println("  - Complexity metrics")
	fmt.Println("  - Type extraction")

	fmt.Println("\n=== Complete ===")
}
