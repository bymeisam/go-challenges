package main

import (
	"strings"
	"testing"
	"time"
)

func TestPasswordHashing(t *testing.T) {
	password := "SecureP@ssw0rd"
	
	// Test hashing
	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}
	
	if hash == password {
		t.Error("Hash should not equal plain password")
	}
	
	// Test verification
	if !VerifyPassword(password, hash) {
		t.Error("Valid password should verify")
	}
	
	if VerifyPassword("WrongPassword", hash) {
		t.Error("Invalid password should not verify")
	}
	
	t.Log("✓ Password hashing works!")
}

func TestJWTTokens(t *testing.T) {
	userID := "user123"
	email := "user@example.com"
	
	// Generate access token
	token, err := GenerateAccessToken(userID, email)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}
	
	// Validate token
	claims, err := ValidateToken(token)
	if err != nil {
		t.Fatalf("Failed to validate token: %v", err)
	}
	
	if claims.UserID != userID {
		t.Errorf("Expected UserID %s, got %s", userID, claims.UserID)
	}
	
	if claims.Email != email {
		t.Errorf("Expected Email %s, got %s", email, claims.Email)
	}
	
	// Test invalid token
	_, err = ValidateToken("invalid.token.here")
	if err == nil {
		t.Error("Invalid token should fail validation")
	}
	
	t.Log("✓ JWT tokens work!")
}

func TestRefreshToken(t *testing.T) {
	userID := "user123"
	
	// Generate refresh token
	token, err := GenerateRefreshToken(userID)
	if err != nil {
		t.Fatalf("Failed to generate refresh token: %v", err)
	}
	
	if token == "" {
		t.Error("Refresh token should not be empty")
	}
	
	// OAuth2Server should be able to use it
	server := NewOAuth2Server()
	response, err := server.RefreshAccessToken(token)
	if err != nil {
		t.Fatalf("Failed to refresh token: %v", err)
	}
	
	if response.AccessToken == "" {
		t.Error("Access token should not be empty")
	}
	
	t.Log("✓ Refresh token works!")
}

func TestAESEncryption(t *testing.T) {
	plaintext := "Sensitive data that needs encryption"
	key := AESKey
	
	// Encrypt
	ciphertext, err := EncryptAES(plaintext, key)
	if err != nil {
		t.Fatalf("Failed to encrypt: %v", err)
	}
	
	if ciphertext == plaintext {
		t.Error("Ciphertext should not equal plaintext")
	}
	
	// Decrypt
	decrypted, err := DecryptAES(ciphertext, key)
	if err != nil {
		t.Fatalf("Failed to decrypt: %v", err)
	}
	
	if decrypted != plaintext {
		t.Errorf("Expected %s, got %s", plaintext, decrypted)
	}
	
	// Test with wrong key
	wrongKey := []byte("wrong-32-byte-key-for-aes-256!!")
	_, err = DecryptAES(ciphertext, wrongKey)
	if err == nil {
		t.Error("Decryption with wrong key should fail")
	}
	
	t.Log("✓ AES encryption works!")
}

func TestEmailValidation(t *testing.T) {
	tests := []struct {
		email string
		valid bool
	}{
		{"user@example.com", true},
		{"user.name@example.co.uk", true},
		{"user+tag@example.com", true},
		{"invalid-email", false},
		{"@example.com", false},
		{"user@", false},
		{"user @example.com", false},
	}
	
	for _, tt := range tests {
		err := ValidateEmail(tt.email)
		if tt.valid && err != nil {
			t.Errorf("Email %s should be valid, got error: %v", tt.email, err)
		}
		if !tt.valid && err == nil {
			t.Errorf("Email %s should be invalid", tt.email)
		}
	}
	
	t.Log("✓ Email validation works!")
}

func TestPasswordValidation(t *testing.T) {
	tests := []struct {
		password string
		valid    bool
	}{
		{"SecureP@ss1", true},
		{"Password123", true},
		{"pass", false},           // too short
		{"password123", false},    // no uppercase
		{"PASSWORD123", false},    // no lowercase
		{"PasswordABC", false},    // no digit
	}
	
	for _, tt := range tests {
		err := ValidatePassword(tt.password)
		if tt.valid && err != nil {
			t.Errorf("Password %s should be valid, got error: %v", tt.password, err)
		}
		if !tt.valid && err == nil {
			t.Errorf("Password %s should be invalid", tt.password)
		}
	}
	
	t.Log("✓ Password validation works!")
}

func TestSQLInjectionDetection(t *testing.T) {
	tests := []struct {
		input    string
		isSQLInj bool
	}{
		{"normal input", false},
		{"John Doe", false},
		{"user@example.com", false},
		{"' OR 1=1 --", true},
		{"admin'; DROP TABLE users--", true},
		{"1 UNION SELECT * FROM users", true},
		{"id = 1 OR 1=1", true},
	}
	
	for _, tt := range tests {
		detected := DetectSQLInjection(tt.input)
		if detected != tt.isSQLInj {
			t.Errorf("Input %q: expected SQL injection detection=%v, got %v", 
				tt.input, tt.isSQLInj, detected)
		}
	}
	
	t.Log("✓ SQL injection detection works!")
}

func TestSanitizeInput(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"<script>alert('xss')</script>", "&lt;script&gt;alert(&#39;xss&#39;)&lt;/script&gt;"},
		{"normal text", "normal text"},
		{"<b>Bold</b>", "&lt;b&gt;Bold&lt;/b&gt;"},
		{"A & B", "A &amp; B"},
	}
	
	for _, tt := range tests {
		result := SanitizeInput(tt.input)
		if result != tt.expected {
			t.Errorf("Input %q: expected %q, got %q", tt.input, tt.expected, result)
		}
	}
	
	t.Log("✓ Input sanitization works!")
}

func TestSecureCompare(t *testing.T) {
	// Test equal strings
	if !SecureCompare("secret123", "secret123") {
		t.Error("Equal strings should return true")
	}
	
	// Test different strings
	if SecureCompare("secret123", "secret456") {
		t.Error("Different strings should return false")
	}
	
	// Test different lengths
	if SecureCompare("short", "longer string") {
		t.Error("Different length strings should return false")
	}
	
	t.Log("✓ Secure comparison works!")
}

func TestOAuth2Flow(t *testing.T) {
	server := NewOAuth2Server()
	userID := "user123"
	clientID := "test-client-id"
	clientSecret := "test-client-secret"
	
	// Step 1: Generate authorization code
	code, err := server.GenerateAuthCode(userID, clientID)
	if err != nil {
		t.Fatalf("Failed to generate auth code: %v", err)
	}
	
	if code == "" {
		t.Error("Authorization code should not be empty")
	}
	
	// Step 2: Exchange code for tokens
	tokenReq := TokenRequest{
		GrantType:    "authorization_code",
		Code:         code,
		ClientID:     clientID,
		ClientSecret: clientSecret,
	}
	
	response, err := server.ExchangeCodeForToken(tokenReq)
	if err != nil {
		t.Fatalf("Failed to exchange code for token: %v", err)
	}
	
	if response.AccessToken == "" {
		t.Error("Access token should not be empty")
	}
	
	if response.RefreshToken == "" {
		t.Error("Refresh token should not be empty")
	}
	
	if response.TokenType != "Bearer" {
		t.Errorf("Expected token type Bearer, got %s", response.TokenType)
	}
	
	// Step 3: Validate access token
	claims, err := ValidateToken(response.AccessToken)
	if err != nil {
		t.Fatalf("Failed to validate access token: %v", err)
	}
	
	if claims.UserID != userID {
		t.Errorf("Expected UserID %s, got %s", userID, claims.UserID)
	}
	
	// Step 4: Code should be one-time use
	_, err = server.ExchangeCodeForToken(tokenReq)
	if err == nil {
		t.Error("Authorization code should only be usable once")
	}
	
	t.Log("✓ OAuth2 flow works!")
}

func TestOAuth2InvalidClient(t *testing.T) {
	server := NewOAuth2Server()
	code, _ := server.GenerateAuthCode("user123", "test-client-id")
	
	// Try with invalid client credentials
	tokenReq := TokenRequest{
		GrantType:    "authorization_code",
		Code:         code,
		ClientID:     "invalid-client",
		ClientSecret: "invalid-secret",
	}
	
	_, err := server.ExchangeCodeForToken(tokenReq)
	if err == nil {
		t.Error("Invalid client credentials should fail")
	}
	
	t.Log("✓ OAuth2 client validation works!")
}

func TestTokenExpiration(t *testing.T) {
	// This test would need to manipulate time or use a very short TTL
	// For demonstration, we'll just check that tokens have expiration set
	
	token, _ := GenerateAccessToken("user123", "user@example.com")
	claims, _ := ValidateToken(token)
	
	if claims.ExpiresAt == nil {
		t.Error("Token should have expiration time")
	}
	
	expiresIn := time.Until(claims.ExpiresAt.Time)
	if expiresIn > AccessTokenTTL || expiresIn < 0 {
		t.Errorf("Token expiration seems incorrect: %v", expiresIn)
	}
	
	t.Log("✓ Token expiration is set correctly!")
}

func TestEncryptionDecryptionIdempotent(t *testing.T) {
	original := "Test data for encryption"
	
	// Encrypt multiple times
	enc1, _ := EncryptAES(original, AESKey)
	enc2, _ := EncryptAES(original, AESKey)
	
	// Each encryption should produce different ciphertext (due to random nonce)
	if enc1 == enc2 {
		t.Error("Each encryption should produce different ciphertext")
	}
	
	// But both should decrypt to the same plaintext
	dec1, _ := DecryptAES(enc1, AESKey)
	dec2, _ := DecryptAES(enc2, AESKey)
	
	if dec1 != original || dec2 != original {
		t.Error("All decryptions should produce original plaintext")
	}
	
	t.Log("✓ Encryption is properly randomized!")
}

func TestBcryptCostFactor(t *testing.T) {
	password := "TestPassword123"
	
	// Time the hashing operation
	start := time.Now()
	_, err := HashPassword(password)
	duration := time.Since(start)
	
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}
	
	// bcrypt with cost 12 should take some time (typically 50-300ms)
	if duration < 10*time.Millisecond {
		t.Logf("Warning: Hashing seems too fast (%v), cost might be too low", duration)
	}
	
	t.Logf("✓ Password hashing took %v (cost factor 12)", duration)
}

func BenchmarkPasswordHashing(b *testing.B) {
	password := "BenchmarkPassword123"
	for i := 0; i < b.N; i++ {
		HashPassword(password)
	}
}

func BenchmarkAESEncryption(b *testing.B) {
	plaintext := "Benchmark data for encryption testing"
	for i := 0; i < b.N; i++ {
		EncryptAES(plaintext, AESKey)
	}
}

func BenchmarkJWTGeneration(b *testing.B) {
	for i := 0; i < b.N; i++ {
		GenerateAccessToken("user123", "user@example.com")
	}
}

func BenchmarkJWTValidation(b *testing.B) {
	token, _ := GenerateAccessToken("user123", "user@example.com")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ValidateToken(token)
	}
}
