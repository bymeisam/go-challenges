# Challenge 149: Security Best Practices

**Difficulty:** ⭐⭐⭐⭐ Hard | **Time:** 75 min

Implement comprehensive security practices for production Go applications.

## Learning Objectives
- OAuth2 authorization flow
- JWT with refresh tokens
- Password hashing with bcrypt
- AES encryption/decryption
- TLS configuration
- Secrets management
- Input validation and sanitization
- SQL injection prevention

## Security Topics Covered
1. **Authentication**: OAuth2, JWT, refresh tokens
2. **Cryptography**: bcrypt, AES, signing
3. **Transport Security**: TLS/HTTPS
4. **Input Validation**: XSS, SQL injection prevention
5. **Secrets Management**: Environment variables, vault patterns

## Tasks
1. Implement OAuth2 authorization code flow
2. Create secure JWT with refresh token rotation
3. Hash passwords with bcrypt
4. Encrypt/decrypt sensitive data with AES
5. Validate and sanitize user input
6. Prevent SQL injection attacks

```bash
go test -v
```

## Production Tips
- Never store passwords in plain text
- Use HTTPS in production
- Rotate secrets regularly
- Implement rate limiting for auth endpoints
- Use secure session management
- Enable CORS properly
- Validate all user input
- Use parameterized queries
