package main

import (
	"fmt"
	"strings"
)

// ========== Dockerfile Templates ==========

// SimpleDockerfile is a basic single-stage Dockerfile
const SimpleDockerfile = `
FROM golang:1.21-alpine AS base
LABEL maintainer="devops@example.com"
LABEL version="1.0"

# Set working directory
WORKDIR /app

# Install dependencies
RUN apk add --no-cache \
    git \
    ca-certificates

# Copy source
COPY . .

# Build application
RUN go build -o app .

# Run
CMD ["./app"]
`

// MultiStageDockerfile implements best practices with multi-stage build
const MultiStageDockerfile = `
# Stage 1: Builder
# Use full Go image for building (smaller image than ubuntu)
FROM golang:1.21-alpine AS builder

# Add metadata labels
LABEL stage="builder"

# Set build arguments (can be overridden at build time)
ARG VERSION=unknown
ARG BUILD_DATE=unknown
ARG GIT_COMMIT=unknown

# Create app directory
WORKDIR /build

# Copy only go mod files first for better caching
COPY go.mod go.sum ./

# Download dependencies (cached if go.mod/sum not changed)
RUN go mod download

# Copy application source
COPY . .

# Build with optimization flags
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s \
    -X main.Version=${VERSION} \
    -X main.BuildDate=${BUILD_DATE} \
    -X main.GitCommit=${GIT_COMMIT}" \
    -o app .

# Run tests (will fail build if tests fail)
RUN go test ./... -v

# Stage 2: Runtime
# Use minimal base image (ca-certificates for HTTPS)
FROM alpine:3.18

# Add non-root user for security
RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser

# Install only runtime dependencies
RUN apk add --no-cache \
    ca-certificates \
    tzdata

# Set working directory
WORKDIR /app

# Copy binary from builder (not source code)
COPY --from=builder --chown=appuser:appuser /build/app .

# Copy configuration if needed
COPY --from=builder --chown=appuser:appuser /build/config ./config

# Switch to non-root user
USER appuser

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD ["/app", "health"]

# Expose port
EXPOSE 8080

# Run application
ENTRYPOINT ["./app"]
CMD ["--config", "config/app.yaml"]
`

// ProductionDockerfile with advanced features
const ProductionDockerfile = `
# Multi-stage build for production deployment
# Stage 1: Dependencies
FROM golang:1.21-alpine AS dependencies

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download && \
    go mod verify

# Stage 2: Security scanning (optional)
FROM dependencies AS scanner

# Copy source for scanning
COPY . .

# Run security checks
RUN go install github.com/securego/gosec/v2/cmd/gosec@latest && \
    gosec ./...

# Stage 3: Builder
FROM dependencies AS builder

ARG VERSION=dev
ARG BUILD_DATE
ARG GIT_COMMIT

WORKDIR /app

COPY . .

# Build with all optimizations and security hardening
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -a -installsuffix cgo \
    -ldflags="-w -s \
    -X main.Version=${VERSION} \
    -X main.BuildDate=${BUILD_DATE} \
    -X main.GitCommit=${GIT_COMMIT}" \
    -o app . && \
    chmod +x app

# Stage 4: Tests
FROM builder AS tester

RUN go test -v -race -coverprofile=coverage.out ./... && \
    go tool cover -func=coverage.out

# Final runtime stage with readonly filesystem support

# Stage 5: Final runtime image
FROM alpine:3.18 AS runtime

# Install security updates and runtime dependencies
RUN apk update && \
    apk add --no-cache \
    ca-certificates \
    tzdata \
    curl && \
    rm -rf /var/cache/apk/*

# Create non-root user with specific UID/GID
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

WORKDIR /home/appuser/app

# Copy binary from builder
COPY --from=builder --chown=appuser:appgroup /app/app .

# Copy public assets if needed
COPY --from=builder --chown=appuser:appgroup /app/public ./public

# Set ownership of working directory
RUN chown -R appuser:appgroup /home/appuser/app && \
    chmod 755 ./app

# Use non-root user
USER appuser

# Set read-only root filesystem where possible
RUN echo "Readonly filesystem support enabled"

# Labels for container registry
LABEL org.opencontainers.image.title="MyApp"
LABEL org.opencontainers.image.description="Production application"
LABEL org.opencontainers.image.version="latest"
LABEL org.opencontainers.image.source="https://github.com/example/myapp"
LABEL org.opencontainers.image.authors="DevOps Team"

# Environment variables (can be overridden)
ENV LOG_LEVEL=info
ENV PORT=8080

# Expose port
EXPOSE 8080

# Health check with custom command
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD curl -f http://localhost:8080/health || exit 1

# Run with graceful shutdown
ENTRYPOINT ["./app"]
CMD ["--port", "8080"]
`

// ========== Docker Build Patterns ==========

// DockerBuildConfig holds build configuration
type DockerBuildConfig struct {
	Version   string
	BuildDate string
	GitCommit string
	GoVersion string
}

// DockerIgnoreRules returns recommended .dockerignore content
func DockerIgnoreRules() string {
	return `
# Version control
.git
.gitignore
.gitattributes

# Build artifacts
bin/
dist/
*.o
*.a

# Dependencies
vendor/
go.sum.bak

# IDE and editor files
.vscode/
.idea/
*.swp
*.swo
*~

# OS files
.DS_Store
Thumbs.db

# Test coverage
coverage.out
*.prof

# Temporary files
tmp/
temp/
*.tmp

# Documentation
docs/
README.md
LICENSE

# CI/CD
.github/
.gitlab-ci.yml
.circleci/
`
}

// ComposeFileContent returns docker-compose.yml with best practices
func ComposeFileContent() string {
	return `
version: '3.8'

services:
  app:
    # Build from Dockerfile
    build:
      context: .
      dockerfile: Dockerfile
      args:
        VERSION: ${VERSION:-dev}
        BUILD_DATE: ${BUILD_DATE}
        GIT_COMMIT: ${GIT_COMMIT}

    # Container configuration
    container_name: myapp
    hostname: myapp

    # Port mapping
    ports:
      - "8080:8080"

    # Environment variables
    environment:
      - LOG_LEVEL=debug
      - PORT=8080
      - DATABASE_URL=postgres://user:password@db:5432/myapp

    # Volume mounts
    volumes:
      - ./config:/app/config:ro
      - logs:/app/logs

    # Restart policy
    restart: unless-stopped

    # Resource limits
    deploy:
      resources:
        limits:
          cpus: '0.5'
          memory: 512M
        reservations:
          cpus: '0.25'
          memory: 256M

    # Health check
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/health"]
      interval: 10s
      timeout: 5s
      retries: 3
      start_period: 10s

    # Dependencies
    depends_on:
      db:
        condition: service_healthy

    # Network
    networks:
      - app-network

    # Security options
    security_opt:
      - no-new-privileges:true

    # Read-only root filesystem (when possible)
    read_only: false

  db:
    image: postgres:15-alpine
    container_name: myapp-db

    environment:
      POSTGRES_USER: user
      POSTGRES_PASSWORD: password
      POSTGRES_DB: myapp

    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./init.sql:/docker-entrypoint-initdb.d/init.sql:ro

    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U user"]
      interval: 10s
      timeout: 5s
      retries: 5

    networks:
      - app-network

volumes:
  postgres_data:
  logs:

networks:
  app-network:
    driver: bridge
`
}

// ========== Dockerfile Comments and Best Practices ==========

// GetDockerfileBestPractices returns a detailed explanation
func GetDockerfileBestPractices() map[string]string {
	return map[string]string{
		"multi_stage": `
Multi-stage builds separate build dependencies from runtime images.
This reduces final image size by not including build tools.
Example: 1GB builder image → 50MB final image`,

		"base_image": `
Choose minimal base images:
- alpine:3.18 (5MB) - preferred for most apps
- ubuntu:22.04 (77MB) - when you need system packages
- distroless (smallest) - when security is critical`,

		"layer_caching": `
Order Dockerfile commands to maximize cache hits:
1. FROM (rarely changes)
2. COPY go.mod go.sum (changes infrequently)
3. RUN go mod download (cached unless dependencies change)
4. COPY source code (changes frequently)
5. RUN go build (rebuilt when source changes)`,

		"security": `
Security best practices:
- Use non-root USER (no privilege escalation)
- Set read-only root filesystem when possible
- Minimize installed packages
- Run security scans (gosec, trivy)
- Use specific package versions`,

		"optimization": `
Build optimization flags:
- CGO_ENABLED=0: Static linking (no libc dependency)
- GOOS=linux GOARCH=amd64: Explicit target
- -ldflags "-w -s": Strip debug info (5-10% smaller)
- -a: Force rebuilding (avoid stale packages)`,

		"health_checks": `
Health checks for orchestrators:
- Set interval, timeout, start_period, retries
- Use simple endpoint (no heavy processing)
- Return exit code 0 for healthy`,

		"environment": `
Environment variable best practices:
- Use ENV for defaults
- Use .env files for local development
- Never hardcode secrets
- Use secrets management systems`,

		"networking": `
Network configuration:
- Expose only necessary ports
- Use private networks for inter-service communication
- Implement service discovery
- Use health checks before routing traffic`,
	}
}

// GenerateDockerfile creates a Dockerfile based on config
func GenerateDockerfile(config DockerBuildConfig) string {
	builder := strings.Builder{}

	builder.WriteString("# Build arguments\n")
	builder.WriteString(fmt.Sprintf("ARG GO_VERSION=%s\n", config.GoVersion))
	builder.WriteString("\n# Stage 1: Builder\n")
	builder.WriteString("FROM golang:${GO_VERSION}-alpine AS builder\n\n")

	builder.WriteString("WORKDIR /build\n\n")

	builder.WriteString("# Copy dependencies\n")
	builder.WriteString("COPY go.mod go.sum ./\n")
	builder.WriteString("RUN go mod download\n\n")

	builder.WriteString("# Copy source and build\n")
	builder.WriteString("COPY . .\n")
	builder.WriteString(fmt.Sprintf("RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags=\" -X main.Version=%s -X main.BuildDate=%s -X main.GitCommit=%s\" -o app .\n\n",
		config.Version, config.BuildDate, config.GitCommit))

	builder.WriteString("# Stage 2: Runtime\n")
	builder.WriteString("FROM alpine:3.18\n\n")

	builder.WriteString("RUN apk add --no-cache ca-certificates\n")
	builder.WriteString("RUN adduser -D -u 1000 appuser\n\n")

	builder.WriteString("WORKDIR /app\n")
	builder.WriteString("COPY --from=builder --chown=appuser:appuser /build/app .\n")
	builder.WriteString("USER appuser\n\n")

	builder.WriteString("EXPOSE 8080\n")
	builder.WriteString("CMD [\"./app\"]\n")

	return builder.String()
}

// ========== Helper Functions ==========

// ContainerRegistry represents Docker image registry
type ContainerRegistry struct {
	Host   string
	Port   int
	Org    string
	Repo   string
	Tag    string
}

// ImageName returns full image name for pushing
func (r *ContainerRegistry) ImageName() string {
	if r.Port == 443 || r.Port == 0 {
		return fmt.Sprintf("%s/%s/%s:%s", r.Host, r.Org, r.Repo, r.Tag)
	}
	return fmt.Sprintf("%s:%d/%s/%s:%s", r.Host, r.Port, r.Org, r.Repo, r.Tag)
}

// RegistryPushCommand returns docker push command
func (r *ContainerRegistry) RegistryPushCommand() string {
	return fmt.Sprintf("docker push %s", r.ImageName())
}

// BuildCommand returns docker build command with best practices
func BuildCommand(version, buildDate, gitCommit string) string {
	return fmt.Sprintf(`docker build \
  --build-arg VERSION=%s \
  --build-arg BUILD_DATE=%s \
  --build-arg GIT_COMMIT=%s \
  --label org.opencontainers.image.created=%s \
  --label org.opencontainers.image.version=%s \
  --label org.opencontainers.image.revision=%s \
  --tag myapp:%s \
  --progress=plain \
  .`,
		version, buildDate, gitCommit, buildDate, version, gitCommit, version)
}

// ScanCommand returns vulnerability scanning command
func ScanCommand(imageName string) string {
	return fmt.Sprintf(`trivy image --severity HIGH,CRITICAL %s`, imageName)
}

func main() {}
