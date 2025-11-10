package main

import (
	"strings"
	"testing"
)

func TestSimpleDockerfile(t *testing.T) {
	if !strings.Contains(SimpleDockerfile, "FROM golang") {
		t.Error("SimpleDockerfile should contain FROM golang")
	}

	if !strings.Contains(SimpleDockerfile, "WORKDIR") {
		t.Error("SimpleDockerfile should contain WORKDIR")
	}

	if !strings.Contains(SimpleDockerfile, "go build") {
		t.Error("SimpleDockerfile should contain go build")
	}

	if !strings.Contains(SimpleDockerfile, "CMD") {
		t.Error("SimpleDockerfile should contain CMD")
	}

	t.Log("✓ SimpleDockerfile is valid!")
}

func TestMultiStageDockerfile(t *testing.T) {
	if !strings.Contains(MultiStageDockerfile, "FROM golang") {
		t.Error("Should have golang builder stage")
	}

	if !strings.Contains(MultiStageDockerfile, "FROM alpine") {
		t.Error("Should have alpine runtime stage")
	}

	if !strings.Contains(MultiStageDockerfile, "--from=builder") {
		t.Error("Should copy from builder stage")
	}

	if !strings.Contains(MultiStageDockerfile, "go mod download") {
		t.Error("Should cache go modules")
	}

	if !strings.Contains(MultiStageDockerfile, "CGO_ENABLED=0") {
		t.Error("Should disable CGO for static linking")
	}

	if !strings.Contains(MultiStageDockerfile, "GOOS=linux") {
		t.Error("Should specify target OS")
	}

	if !strings.Contains(MultiStageDockerfile, "go test") {
		t.Error("Should run tests in build")
	}

	t.Log("✓ MultiStageDockerfile has best practices!")
}

func TestProductionDockerfile(t *testing.T) {
	if !strings.Contains(ProductionDockerfile, "AS builder") {
		t.Error("Should have builder stage")
	}

	if !strings.Contains(ProductionDockerfile, "AS scanner") {
		t.Error("Should have security scanning stage")
	}

	if !strings.Contains(ProductionDockerfile, "AS tester") {
		t.Error("Should have testing stage")
	}

	if !strings.Contains(ProductionDockerfile, "gosec") {
		t.Error("Should run security checks")
	}

	if !strings.Contains(ProductionDockerfile, "go test") || !strings.Contains(ProductionDockerfile, "-race") {
		t.Error("Should run tests with race detector")
	}

	if !strings.Contains(ProductionDockerfile, "adduser") {
		t.Error("Should create non-root user")
	}

	if !strings.Contains(ProductionDockerfile, "HEALTHCHECK") {
		t.Error("Should have health check")
	}

	if !strings.Contains(ProductionDockerfile, "USER") {
		t.Error("Should switch to non-root user")
	}

	if !strings.Contains(ProductionDockerfile, "read-only") {
		t.Error("Should mention read-only filesystem")
	}

	t.Log("✓ ProductionDockerfile is production-ready!")
}

func TestDockerIgnoreRules(t *testing.T) {
	rules := DockerIgnoreRules()

	expectedPatterns := []string{
		".git",
		"vendor/",
		".vscode/",
		"coverage.out",
		".DS_Store",
	}

	for _, pattern := range expectedPatterns {
		if !strings.Contains(rules, pattern) {
			t.Errorf("DockerIgnoreRules should contain %s", pattern)
		}
	}

	t.Log("✓ DockerIgnoreRules contains expected patterns!")
}

func TestComposeFileContent(t *testing.T) {
	content := ComposeFileContent()

	expectedElements := []string{
		"version: '3.8'",
		"services:",
		"app:",
		"build:",
		"environment:",
		"volumes:",
		"healthcheck:",
		"depends_on:",
		"db:",
		"postgres",
		"networks:",
		"resource",
		"limits:",
		"memory:",
	}

	for _, element := range expectedElements {
		if !strings.Contains(content, element) {
			t.Errorf("ComposeFile should contain %s", element)
		}
	}

	t.Log("✓ ComposeFileContent is complete!")
}

func TestDockerfileBestPractices(t *testing.T) {
	practices := GetDockerfileBestPractices()

	expectedKeys := []string{
		"multi_stage",
		"base_image",
		"layer_caching",
		"security",
		"optimization",
		"health_checks",
		"environment",
		"networking",
	}

	for _, key := range expectedKeys {
		if _, exists := practices[key]; !exists {
			t.Errorf("Should have practice %s", key)
		}

		if practices[key] == "" {
			t.Errorf("Practice %s should not be empty", key)
		}
	}

	t.Log("✓ All best practices documented!")
}

func TestGenerateDockerfile(t *testing.T) {
	config := DockerBuildConfig{
		Version:   "1.0.0",
		BuildDate: "2024-01-01",
		GitCommit: "abc123",
		GoVersion: "1.21",
	}

	dockerfile := GenerateDockerfile(config)

	if !strings.Contains(dockerfile, "1.0.0") {
		t.Error("Dockerfile should contain version")
	}

	if !strings.Contains(dockerfile, "2024-01-01") {
		t.Error("Dockerfile should contain build date")
	}

	if !strings.Contains(dockerfile, "abc123") {
		t.Error("Dockerfile should contain git commit")
	}

	if !strings.Contains(dockerfile, "1.21") {
		t.Error("Dockerfile should contain Go version")
	}

	if !strings.Contains(dockerfile, "AS builder") {
		t.Error("Dockerfile should have builder stage")
	}

	if !strings.Contains(dockerfile, "FROM alpine") {
		t.Error("Dockerfile should have alpine runtime stage")
	}

	t.Log("✓ GenerateDockerfile works correctly!")
}

func TestContainerRegistry(t *testing.T) {
	registry := ContainerRegistry{
		Host: "docker.io",
		Port: 0,
		Org:  "myorg",
		Repo: "myapp",
		Tag:  "v1.0.0",
	}

	imageName := registry.ImageName()
	expectedName := "docker.io/myorg/myapp:v1.0.0"

	if imageName != expectedName {
		t.Errorf("Expected image name %s, got %s", expectedName, imageName)
	}

	t.Log("✓ ContainerRegistry generates correct image names!")
}

func TestContainerRegistryWithPort(t *testing.T) {
	registry := ContainerRegistry{
		Host: "registry.example.com",
		Port: 5000,
		Org:  "team",
		Repo: "service",
		Tag:  "latest",
	}

	imageName := registry.ImageName()
	expectedName := "registry.example.com:5000/team/service:latest"

	if imageName != expectedName {
		t.Errorf("Expected image name %s, got %s", expectedName, imageName)
	}

	t.Log("✓ ContainerRegistry with port works!")
}

func TestRegistryPushCommand(t *testing.T) {
	registry := ContainerRegistry{
		Host: "docker.io",
		Org:  "myorg",
		Repo: "myapp",
		Tag:  "v1.0.0",
	}

	cmd := registry.RegistryPushCommand()

	if !strings.Contains(cmd, "docker push") {
		t.Error("Push command should contain 'docker push'")
	}

	if !strings.Contains(cmd, "docker.io/myorg/myapp:v1.0.0") {
		t.Error("Push command should contain full image name")
	}

	t.Log("✓ RegistryPushCommand generates correct command!")
}

func TestBuildCommand(t *testing.T) {
	cmd := BuildCommand("1.0.0", "2024-01-01", "abc123")

	expectedElements := []string{
		"docker build",
		"--build-arg VERSION=1.0.0",
		"--build-arg BUILD_DATE=2024-01-01",
		"--build-arg GIT_COMMIT=abc123",
		"--label org.opencontainers.image",
		"--tag myapp:1.0.0",
	}

	for _, element := range expectedElements {
		if !strings.Contains(cmd, element) {
			t.Errorf("Build command should contain %s", element)
		}
	}

	t.Log("✓ BuildCommand generates correct command!")
}

func TestScanCommand(t *testing.T) {
	cmd := ScanCommand("myapp:latest")

	if !strings.Contains(cmd, "trivy image") {
		t.Error("Scan command should contain 'trivy image'")
	}

	if !strings.Contains(cmd, "myapp:latest") {
		t.Error("Scan command should contain image name")
	}

	if !strings.Contains(cmd, "CRITICAL") {
		t.Error("Scan command should check for CRITICAL vulnerabilities")
	}

	t.Log("✓ ScanCommand generates correct command!")
}

func TestDockerfileContentValidity(t *testing.T) {
	dockerfiles := []struct {
		name    string
		content string
	}{
		{"Simple", SimpleDockerfile},
		{"MultiStage", MultiStageDockerfile},
		{"Production", ProductionDockerfile},
	}

	for _, df := range dockerfiles {
		// Check basic structure
		if !strings.Contains(df.content, "FROM") {
			t.Errorf("%s Dockerfile should have FROM instruction", df.name)
		}

		// COUNT FROM statements (should be at least 1)
		fromCount := strings.Count(df.content, "FROM")
		if fromCount < 1 {
			t.Errorf("%s Dockerfile should have at least one FROM", df.name)
		}

		// Check it's not empty
		if strings.TrimSpace(df.content) == "" {
			t.Errorf("%s Dockerfile is empty", df.name)
		}
	}

	t.Log("✓ All Dockerfiles have valid structure!")
}

func TestCachingStrategy(t *testing.T) {
	// Verify go.mod copy comes before source copy
	modCopyIndex := strings.Index(MultiStageDockerfile, "COPY go.mod go.sum")
	sourceIndex := strings.Index(MultiStageDockerfile, "COPY . .")

	if modCopyIndex == -1 {
		t.Error("Should copy go.mod/go.sum")
	}

	if sourceIndex == -1 {
		t.Error("Should copy source")
	}

	if modCopyIndex >= sourceIndex {
		t.Error("go.mod/sum should be copied before source for better caching")
	}

	t.Log("✓ Dockerfile uses optimal layer caching!")
}

func TestSecurityFeatures(t *testing.T) {
	securityFeatures := map[string]string{
		"Non-root user":   "adduser",
		"Alpine base":     "alpine",
		"CA certificates": "ca-certificates",
	}

	for feature, keyword := range securityFeatures {
		if !strings.Contains(ProductionDockerfile, keyword) {
			t.Errorf("ProductionDockerfile missing %s (%s)", feature, keyword)
		}
	}

	t.Log("✓ ProductionDockerfile includes security features!")
}

func TestComposeDependencies(t *testing.T) {
	compose := ComposeFileContent()

	// Check service depends_on
	if !strings.Contains(compose, "depends_on:") {
		t.Error("Should specify service dependencies")
	}

	if !strings.Contains(compose, "condition:") {
		t.Error("Should specify service startup condition")
	}

	// Check database service
	if !strings.Contains(compose, "db:") {
		t.Error("Should include database service")
	}

	if !strings.Contains(compose, "postgres") {
		t.Error("Should use PostgreSQL image")
	}

	t.Log("✓ Docker Compose has proper dependencies!")
}

func TestComposeLimits(t *testing.T) {
	compose := ComposeFileContent()

	if !strings.Contains(compose, "limits:") {
		t.Error("Should specify resource limits")
	}

	if !strings.Contains(compose, "cpus:") {
		t.Error("Should specify CPU limits")
	}

	if !strings.Contains(compose, "memory:") {
		t.Error("Should specify memory limits")
	}

	t.Log("✓ Docker Compose has resource constraints!")
}

func BenchmarkGenerateDockerfile(b *testing.B) {
	config := DockerBuildConfig{
		Version:   "1.0.0",
		BuildDate: "2024-01-01",
		GitCommit: "abc123",
		GoVersion: "1.21",
	}

	for i := 0; i < b.N; i++ {
		GenerateDockerfile(config)
	}
}
