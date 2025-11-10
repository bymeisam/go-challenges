package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// OAuth2 Token Response
type OAuth2TokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
}

// JWT Claims
type Claims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

// Secrets configuration
var (
	JWTSecret        = []byte("your-256-bit-secret-key-change-in-production")
	AESKey           = []byte("32-byte-key-for-aes-256-encry!") // 32 bytes for AES-256
	AccessTokenTTL   = 15 * time.Minute
	RefreshTokenTTL  = 7 * 24 * time.Hour
)

// ========== Password Hashing ==========

// HashPassword hashes a password using bcrypt
func HashPassword(password string) (string, error) {
	// Cost of 12 is a good balance between security and performance
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// VerifyPassword checks if a password matches the hash
func VerifyPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// ========== JWT Token Management ==========

// GenerateAccessToken creates a new JWT access token
func GenerateAccessToken(userID, email string) (string, error) {
	claims := Claims{
		UserID: userID,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(AccessTokenTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "myapp",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(JWTSecret)
}

// GenerateRefreshToken creates a refresh token
func GenerateRefreshToken(userID string) (string, error) {
	claims := jwt.RegisteredClaims{
		Subject:   userID,
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(RefreshTokenTTL)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		Issuer:    "myapp",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(JWTSecret)
}

// ValidateToken validates and parses a JWT token
func ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Verify signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return JWTSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

// ========== AES Encryption ==========

// EncryptAES encrypts data using AES-256-GCM
func EncryptAES(plaintext string, key []byte) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// DecryptAES decrypts AES-256-GCM encrypted data
func DecryptAES(ciphertext string, key []byte) (string, error) {
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", errors.New("ciphertext too short")
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

// ========== Input Validation ==========

var (
	emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	sqlRegex   = regexp.MustCompile(`(?i)(union|select|insert|update|delete|drop|;|--|'|"|\*|xp_)`)
)

// ValidateEmail validates email format
func ValidateEmail(email string) error {
	if !emailRegex.MatchString(email) {
		return errors.New("invalid email format")
	}
	return nil
}

// SanitizeInput removes potentially dangerous characters
func SanitizeInput(input string) string {
	// Remove potential SQL injection patterns
	input = strings.TrimSpace(input)
	
	// HTML escape
	replacer := strings.NewReplacer(
		"<", "&lt;",
		">", "&gt;",
		"&", "&amp;",
		"\"", "&quot;",
		"'", "&#39;",
	)
	
	return replacer.Replace(input)
}

// DetectSQLInjection checks for SQL injection attempts
func DetectSQLInjection(input string) bool {
	return sqlRegex.MatchString(input)
}

// ValidatePassword checks password strength
func ValidatePassword(password string) error {
	if len(password) < 8 {
		return errors.New("password must be at least 8 characters")
	}
	
	hasUpper := false
	hasLower := false
	hasDigit := false
	
	for _, char := range password {
		switch {
		case 'A' <= char && char <= 'Z':
			hasUpper = true
		case 'a' <= char && char <= 'z':
			hasLower = true
		case '0' <= char && char <= '9':
			hasDigit = true
		}
	}
	
	if !hasUpper || !hasLower || !hasDigit {
		return errors.New("password must contain uppercase, lowercase, and digit")
	}
	
	return nil
}

// ========== Secure Comparison ==========

// SecureCompare performs constant-time string comparison
func SecureCompare(a, b string) bool {
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}

// ========== OAuth2 Flow ==========

type AuthorizationRequest struct {
	ClientID     string `json:"client_id"`
	RedirectURI  string `json:"redirect_uri"`
	ResponseType string `json:"response_type"`
	Scope        string `json:"scope"`
	State        string `json:"state"`
}

type TokenRequest struct {
	GrantType    string `json:"grant_type"`
	Code         string `json:"code"`
	RedirectURI  string `json:"redirect_uri"`
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	RefreshToken string `json:"refresh_token"`
}

// OAuth2Server represents a simplified OAuth2 authorization server
type OAuth2Server struct {
	authCodes     map[string]string // code -> userID
	clients       map[string]string // clientID -> clientSecret
}

func NewOAuth2Server() *OAuth2Server {
	return &OAuth2Server{
		authCodes: make(map[string]string),
		clients: map[string]string{
			"test-client-id": "test-client-secret",
		},
	}
}

// GenerateAuthCode creates an authorization code
func (s *OAuth2Server) GenerateAuthCode(userID, clientID string) (string, error) {
	// Generate random code
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	code := base64.URLEncoding.EncodeToString(b)
	
	s.authCodes[code] = userID
	
	// Code expires after 10 minutes (in production, use Redis with TTL)
	go func() {
		time.Sleep(10 * time.Minute)
		delete(s.authCodes, code)
	}()
	
	return code, nil
}

// ExchangeCodeForToken exchanges authorization code for tokens
func (s *OAuth2Server) ExchangeCodeForToken(req TokenRequest) (*OAuth2TokenResponse, error) {
	// Validate client
	secret, ok := s.clients[req.ClientID]
	if !ok || !SecureCompare(secret, req.ClientSecret) {
		return nil, errors.New("invalid client credentials")
	}
	
	// Validate authorization code
	userID, ok := s.authCodes[req.Code]
	if !ok {
		return nil, errors.New("invalid or expired authorization code")
	}
	
	// Delete code (one-time use)
	delete(s.authCodes, req.Code)
	
	// Generate tokens
	accessToken, err := GenerateAccessToken(userID, "user@example.com")
	if err != nil {
		return nil, err
	}
	
	refreshToken, err := GenerateRefreshToken(userID)
	if err != nil {
		return nil, err
	}
	
	return &OAuth2TokenResponse{
		AccessToken:  accessToken,
		TokenType:    "Bearer",
		ExpiresIn:    int(AccessTokenTTL.Seconds()),
		RefreshToken: refreshToken,
	}, nil
}

// RefreshAccessToken generates new access token from refresh token
func (s *OAuth2Server) RefreshAccessToken(refreshToken string) (*OAuth2TokenResponse, error) {
	// Validate refresh token
	token, err := jwt.Parse(refreshToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return JWTSecret, nil
	})
	
	if err != nil || !token.Valid {
		return nil, errors.New("invalid refresh token")
	}
	
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid token claims")
	}
	
	userID, ok := claims["sub"].(string)
	if !ok {
		return nil, errors.New("missing user ID")
	}
	
	// Generate new access token
	accessToken, err := GenerateAccessToken(userID, "user@example.com")
	if err != nil {
		return nil, err
	}
	
	return &OAuth2TokenResponse{
		AccessToken: accessToken,
		TokenType:   "Bearer",
		ExpiresIn:   int(AccessTokenTTL.Seconds()),
	}, nil
}

// ========== HTTP Handlers Example ==========

func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Missing authorization header", http.StatusUnauthorized)
			return
		}
		
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, "Invalid authorization header", http.StatusUnauthorized)
			return
		}
		
		claims, err := ValidateToken(parts[1])
		if err != nil {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}
		
		// Add user info to context (simplified)
		w.Header().Set("X-User-ID", claims.UserID)
		next(w, r)
	}
}

func ProtectedHandler(w http.ResponseWriter, r *http.Request) {
	userID := w.Header().Get("X-User-ID")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Protected resource",
		"user_id": userID,
	})
}

func main() {}
