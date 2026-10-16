package security

import (
	"strings"
	"testing"
	"time"

	"auth_service/internal/config"
	"auth_service/internal/models"
)

func TestPasswordHasher(t *testing.T) {
	hasher := NewSHA256SaltHasher()

	password := "TestSecretPassword123!"
	hash, salt, err := hasher.HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}

	if hash == "" || salt == "" {
		t.Fatal("Hash or salt is empty")
	}

	if hash == password {
		t.Fatal("Password was not hashed!")
	}

	if !hasher.VerifyPassword(password, hash, salt) {
		t.Fatal("VerifyPassword failed for correct password")
	}

	if hasher.VerifyPassword("WrongPassword!", hash, salt) {
		t.Fatal("VerifyPassword succeeded for wrong password")
	}

	// Test random password generation
	randPass, err := hasher.GenerateRandomPassword(16)
	if err != nil {
		t.Fatalf("GenerateRandomPassword failed: %v", err)
	}
	if len(randPass) != 16 {
		t.Fatalf("Expected random password length 16, got %d", len(randPass))
	}
}

func TestTokenManager(t *testing.T) {
	cfg := &config.Config{
		JWTSecret:        "my-secret-test-key-321",
		TokenExpiryHours: 2,
	}
	tokenMgr := NewJWTTokenManager(cfg)

	user := &models.User{
		ID:          "11111111-2222-3333-4444-555555555555",
		Email:       "test@example.com",
		FirstName:   "Test",
		LastName:    "User",
		PhoneNumber: "+1234567890",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	tokenStr, err := tokenMgr.GenerateToken(user)
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}

	if len(strings.Split(tokenStr, ".")) != 3 {
		t.Fatalf("Invalid JWT token format: %s", tokenStr)
	}

	claims, err := tokenMgr.ValidateToken(tokenStr)
	if err != nil {
		t.Fatalf("ValidateToken failed: %v", err)
	}

	if claims.UserID != user.ID || claims.Email != user.Email {
		t.Fatalf("Token claims mismatch: got user_id=%s, email=%s", claims.UserID, claims.Email)
	}

	// Tampered token test
	tamperedToken := tokenStr[:len(tokenStr)-4] + "abcd"
	_, err = tokenMgr.ValidateToken(tamperedToken)
	if err == nil {
		t.Fatal("ValidateToken should fail for tampered token")
	}
}
