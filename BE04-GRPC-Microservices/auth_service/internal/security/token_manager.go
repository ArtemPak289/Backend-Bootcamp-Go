package security

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"auth_service/internal/config"
	"auth_service/internal/models"
)

type UserClaims struct {
	UserID      string `json:"user_id"`
	Email       string `json:"email"`
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	PhoneNumber string `json:"phone_number"`
	jwt.RegisteredClaims
}

type ITokenManager interface {
	GenerateToken(user *models.User) (string, error)
	ValidateToken(tokenStr string) (*UserClaims, error)
}

type JWTTokenManager struct {
	secretKey   []byte
	expiryHours int
}

func NewJWTTokenManager(cfg *config.Config) ITokenManager {
	return &JWTTokenManager{
		secretKey:   []byte(cfg.JWTSecret),
		expiryHours: cfg.TokenExpiryHours,
	}
}

func (m *JWTTokenManager) GenerateToken(user *models.User) (string, error) {
	now := time.Now()
	claims := UserClaims{
		UserID:      user.ID,
		Email:       user.Email,
		FirstName:   user.FirstName,
		LastName:    user.LastName,
		PhoneNumber: user.PhoneNumber,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   user.ID,
			Issuer:    "auth_service",
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Duration(m.expiryHours) * time.Hour)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(m.secretKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}
	return tokenString, nil
}

func (m *JWTTokenManager) ValidateToken(tokenStr string) (*UserClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &UserClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return m.secretKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("token validation failed: %w", err)
	}

	claims, ok := token.Claims.(*UserClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token claims")
	}

	return claims, nil
}
