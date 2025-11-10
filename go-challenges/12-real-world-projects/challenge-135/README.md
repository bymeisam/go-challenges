# Challenge 135: REST API Service

**Difficulty:** ⭐⭐⭐⭐ Hard
**Time Estimate:** 60 minutes

## Description

Build a full-featured REST API service for a blog platform with database integration, middleware, validation, and proper error handling. This project demonstrates modern API development practices in Go.

## Features

- **RESTful Endpoints**: CRUD operations for articles
- **Database Integration**: SQLite with proper connection pooling
- **Middleware**: Logging, authentication, CORS, rate limiting
- **Request Validation**: Input validation and sanitization
- **Error Handling**: Consistent error responses
- **Pagination**: List endpoints with pagination support
- **Authentication**: JWT-based authentication
- **Health Check**: Status endpoint for monitoring
- **Graceful Shutdown**: Proper cleanup on exit

## API Endpoints

```
POST   /api/auth/register      - Register new user
POST   /api/auth/login         - Login and get JWT token
GET    /api/articles           - List articles (paginated)
GET    /api/articles/:id       - Get article by ID
POST   /api/articles           - Create article (authenticated)
PUT    /api/articles/:id       - Update article (authenticated)
DELETE /api/articles/:id       - Delete article (authenticated)
GET    /api/health             - Health check
```

## Requirements

1. Use Chi router for HTTP routing
2. SQLite database for storage
3. JWT for authentication
4. Middleware for logging and auth
5. Proper HTTP status codes
6. JSON request/response format
7. Input validation
8. Error handling middleware

## Example Usage

```bash
# Register user
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{"username":"john","password":"secret123"}'

# Login
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"john","password":"secret123"}'

# Create article (with token)
curl -X POST http://localhost:8080/api/articles \
  -H "Authorization: Bearer TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"title":"My Article","content":"Article content","author":"john"}'

# List articles
curl http://localhost:8080/api/articles?page=1&limit=10
```

## Learning Objectives

- RESTful API design principles
- HTTP routing and middleware
- Database operations with SQL
- JWT authentication
- Request validation
- Error handling patterns
- API documentation
- Testing HTTP handlers
