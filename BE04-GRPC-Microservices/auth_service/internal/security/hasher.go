package security

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	"math/big"
)

// IPasswordHasher defines the contract for salted password hashing.
type IPasswordHasher interface {
	HashPassword(password string) (hash string, salt string, err error)
	VerifyPassword(password, hash, salt string) bool
	GenerateRandomPassword(length int) (string, error)
}

type SHA256SaltHasher struct{}

func NewSHA256SaltHasher() IPasswordHasher {
	return &SHA256SaltHasher{}
}

// HashPassword generates a cryptographically secure random salt and returns the SHA-256 hash.
func (h *SHA256SaltHasher) HashPassword(password string) (string, string, error) {
	saltBytes := make([]byte, 16)
	if _, err := rand.Read(saltBytes); err != nil {
		return "", "", fmt.Errorf("failed to generate random salt: %w", err)
	}
	salt := hex.EncodeToString(saltBytes)

	hasher := sha256.New()
	hasher.Write([]byte(password + ":" + salt))
	hash := hex.EncodeToString(hasher.Sum(nil))

	return hash, salt, nil
}

// VerifyPassword computes the SHA-256 hash using the provided salt and compares in constant time.
func (h *SHA256SaltHasher) VerifyPassword(password, hash, salt string) bool {
	hasher := sha256.New()
	hasher.Write([]byte(password + ":" + salt))
	computedHash := hex.EncodeToString(hasher.Sum(nil))

	return subtle.ConstantTimeCompare([]byte(computedHash), []byte(hash)) == 1
}

// GenerateRandomPassword generates a secure random alphanumeric password.
func (h *SHA256SaltHasher) GenerateRandomPassword(length int) (string, error) {
	if length <= 0 {
		length = 12
	}
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*"
	bytes := make([]byte, length)
	for i := range bytes {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", err
		}
		bytes[i] = charset[num.Int64()]
	}
	return string(bytes), nil
}
