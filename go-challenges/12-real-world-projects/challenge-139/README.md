# Challenge 139: URL Shortener

**Difficulty:** ⭐⭐⭐⭐ Hard
**Time Estimate:** 50 minutes

## Description

Build a URL shortening service with Redis backend, custom short codes, analytics tracking, and expiration support. This demonstrates caching, key generation, and building a production-ready service.

## Features

- **Short URL Generation**: Generate short codes for long URLs
- **Custom Aliases**: Support custom short codes
- **Redis Storage**: Fast lookup with Redis
- **Analytics**: Track clicks, referrers, user agents
- **Expiration**: TTL-based URL expiration
- **URL Validation**: Validate and sanitize URLs
- **Rate Limiting**: Prevent abuse
- **QR Code Generation**: Generate QR codes for short URLs
- **API Key Authentication**: Protect API endpoints

## API Endpoints

```
POST   /api/shorten        - Create short URL
GET    /:code              - Redirect to long URL
GET    /api/stats/:code    - Get analytics
DELETE /api/:code          - Delete short URL
GET    /api/qr/:code       - Get QR code
```

## Requirements

1. Use Redis for URL storage
2. Generate collision-resistant short codes
3. Track click analytics
4. Support custom aliases
5. URL validation and sanitization
6. Handle concurrent requests
7. Implement expiration

## Example Usage

```bash
# Create short URL
curl -X POST http://localhost:8080/api/shorten \
  -H "Content-Type: application/json" \
  -d '{"url":"https://example.com/very/long/url"}'

# Response: {"short_url":"http://localhost:8080/abc123"}

# Use short URL
curl -L http://localhost:8080/abc123
# Redirects to https://example.com/very/long/url

# Get stats
curl http://localhost:8080/api/stats/abc123
```

## Learning Objectives

- Redis integration
- URL validation and normalization
- Short code generation algorithms
- Analytics and tracking
- Caching strategies
- Rate limiting implementation
- QR code generation
- RESTful API design
