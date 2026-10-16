package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"

	"auth_service/internal/models"
	"auth_service/internal/repository"
	"auth_service/internal/security"
)

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrOldPasswordMismatch = errors.New("old password does not match")
	ErrValidation          = errors.New("validation failed")
)

type IAuthService interface {
	Register(ctx context.Context, email, firstName, lastName, phone, password string) (*models.User, string, error)
	Login(ctx context.Context, email, password string) (*models.User, string, error)
	ChangePassword(ctx context.Context, email, oldPassword, newPassword string) error
	ResetPassword(ctx context.Context, email string) error
	ValidateToken(ctx context.Context, token string) (*security.UserClaims, error)
	GetUser(ctx context.Context, userID, email string) (*models.User, error)
}

type AuthServiceImpl struct {
	userRepo   repository.IUserRepository
	hasher     security.IPasswordHasher
	tokenMgr   security.ITokenManager
}

func NewAuthService(
	userRepo repository.IUserRepository,
	hasher security.IPasswordHasher,
	tokenMgr security.ITokenManager,
) IAuthService {
	return &AuthServiceImpl{
		userRepo: userRepo,
		hasher:   hasher,
		tokenMgr: tokenMgr,
	}
}

func (s *AuthServiceImpl) Register(ctx context.Context, email, firstName, lastName, phone, password string) (*models.User, string, error) {
	email = strings.TrimSpace(email)
	firstName = strings.TrimSpace(firstName)
	lastName = strings.TrimSpace(lastName)
	phone = strings.TrimSpace(phone)

	if email == "" || password == "" || firstName == "" || lastName == "" {
		return nil, "", fmt.Errorf("%w: email, first name, last name and password are required", ErrValidation)
	}

	hash, salt, err := s.hasher.HashPassword(password)
	if err != nil {
		return nil, "", fmt.Errorf("failed to hash password: %w", err)
	}

	user := &models.User{
		Email:        email,
		FirstName:    firstName,
		LastName:     lastName,
		PhoneNumber:  phone,
		PasswordHash: hash,
		Salt:         salt,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, "", err
	}

	token, err := s.tokenMgr.GenerateToken(user)
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate token: %w", err)
	}

	log.Printf("[AUTH SERVICE] User registered successfully: ID=%s, Email=%s", user.ID, user.Email)
	return user, token, nil
}

func (s *AuthServiceImpl) Login(ctx context.Context, email, password string) (*models.User, string, error) {
	email = strings.TrimSpace(email)
	if email == "" || password == "" {
		return nil, "", ErrInvalidCredentials
	}

	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, "", ErrInvalidCredentials
		}
		return nil, "", err
	}

	if !s.hasher.VerifyPassword(password, user.PasswordHash, user.Salt) {
		return nil, "", ErrInvalidCredentials
	}

	token, err := s.tokenMgr.GenerateToken(user)
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate token: %w", err)
	}

	log.Printf("[AUTH SERVICE] User logged in: ID=%s, Email=%s", user.ID, user.Email)
	return user, token, nil
}

func (s *AuthServiceImpl) ChangePassword(ctx context.Context, email, oldPassword, newPassword string) error {
	email = strings.TrimSpace(email)
	if email == "" || oldPassword == "" || newPassword == "" {
		return fmt.Errorf("%w: email, old password and new password are required", ErrValidation)
	}

	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return err
	}

	if !s.hasher.VerifyPassword(oldPassword, user.PasswordHash, user.Salt) {
		return ErrOldPasswordMismatch
	}

	newHash, newSalt, err := s.hasher.HashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("failed to hash new password: %w", err)
	}

	if err := s.userRepo.UpdatePassword(ctx, email, newHash, newSalt); err != nil {
		return err
	}

	log.Printf("[AUTH SERVICE] Password changed successfully for user %s", email)
	return nil
}

func (s *AuthServiceImpl) ResetPassword(ctx context.Context, email string) error {
	email = strings.TrimSpace(email)
	if email == "" {
		return fmt.Errorf("%w: email is required", ErrValidation)
	}

	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return err
	}

	// Generate new temporary password
	tempPassword, err := s.hasher.GenerateRandomPassword(12)
	if err != nil {
		return fmt.Errorf("failed to generate temporary password: %w", err)
	}

	newHash, newSalt, err := s.hasher.HashPassword(tempPassword)
	if err != nil {
		return fmt.Errorf("failed to hash temporary password: %w", err)
	}

	if err := s.userRepo.UpdatePassword(ctx, user.Email, newHash, newSalt); err != nil {
		return err
	}

	// Implement stub for sending to email in the form of password output to authorization service console
	log.Printf("\n" +
		"================================================================================\n" +
		"[EMAIL STUB] >>> PASSWORD RECOVERY NOTIFICATION <<<\n" +
		"To: %s\n" +
		"Subject: Password Reset for Appliance Store\n" +
		"Message: Your new temporary password is: %s\n" +
		"Please use this password to sign in and change it immediately.\n" +
		"================================================================================\n",
		user.Email, tempPassword)

	return nil
}

func (s *AuthServiceImpl) ValidateToken(ctx context.Context, token string) (*security.UserClaims, error) {
	token = strings.TrimSpace(token)
	if token == "" {
		return nil, errors.New("empty token")
	}

	claims, err := s.tokenMgr.ValidateToken(token)
	if err != nil {
		return nil, err
	}

	// Verify user exists in database
	_, err = s.userRepo.GetByID(ctx, claims.UserID)
	if err != nil {
		return nil, fmt.Errorf("token user not found in database: %w", err)
	}

	return claims, nil
}

func (s *AuthServiceImpl) GetUser(ctx context.Context, userID, email string) (*models.User, error) {
	if userID != "" {
		return s.userRepo.GetByID(ctx, userID)
	}
	if email != "" {
		return s.userRepo.GetByEmail(ctx, email)
	}
	return nil, fmt.Errorf("%w: user_id or email must be provided", ErrValidation)
}
