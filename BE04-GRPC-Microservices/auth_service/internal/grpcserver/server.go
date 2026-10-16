package grpcserver

import (
	"context"
	"errors"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"auth_service/internal/repository"
	"auth_service/internal/service"
	authpb "auth_service/proto"
)

type Server struct {
	authpb.UnimplementedAuthServiceServer
	authService service.IAuthService
}

func NewServer(authService service.IAuthService) *Server {
	return &Server{
		authService: authService,
	}
}

func (s *Server) Register(ctx context.Context, req *authpb.RegisterRequest) (*authpb.AuthResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is nil")
	}

	user, token, err := s.authService.Register(
		ctx,
		req.GetEmail(),
		req.GetFirstName(),
		req.GetLastName(),
		req.GetPhoneNumber(),
		req.GetPassword(),
	)
	if err != nil {
		if errors.Is(err, repository.ErrUserAlreadyExists) {
			return nil, status.Error(codes.AlreadyExists, "user with this email already exists")
		}
		if errors.Is(err, service.ErrValidation) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		return nil, status.Errorf(codes.Internal, "registration failed: %v", err)
	}

	return &authpb.AuthResponse{
		Token:       token,
		UserId:      user.ID,
		Email:       user.Email,
		FirstName:   user.FirstName,
		LastName:    user.LastName,
		PhoneNumber: user.PhoneNumber,
	}, nil
}

func (s *Server) Login(ctx context.Context, req *authpb.LoginRequest) (*authpb.AuthResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is nil")
	}

	user, token, err := s.authService.Login(ctx, req.GetEmail(), req.GetPassword())
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			return nil, status.Error(codes.Unauthenticated, "invalid email or password")
		}
		return nil, status.Errorf(codes.Internal, "login failed: %v", err)
	}

	return &authpb.AuthResponse{
		Token:       token,
		UserId:      user.ID,
		Email:       user.Email,
		FirstName:   user.FirstName,
		LastName:    user.LastName,
		PhoneNumber: user.PhoneNumber,
	}, nil
}

func (s *Server) ChangePassword(ctx context.Context, req *authpb.ChangePasswordRequest) (*authpb.ChangePasswordResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is nil")
	}

	err := s.authService.ChangePassword(ctx, req.GetEmail(), req.GetOldPassword(), req.GetNewPassword())
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, status.Error(codes.NotFound, "user not found")
		}
		if errors.Is(err, service.ErrOldPasswordMismatch) {
			return nil, status.Error(codes.InvalidArgument, "old password does not match")
		}
		if errors.Is(err, service.ErrValidation) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		return nil, status.Errorf(codes.Internal, "change password failed: %v", err)
	}

	return &authpb.ChangePasswordResponse{
		Success: true,
		Message: "Password updated successfully",
	}, nil
}

func (s *Server) ResetPassword(ctx context.Context, req *authpb.ResetPasswordRequest) (*authpb.ResetPasswordResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is nil")
	}

	err := s.authService.ResetPassword(ctx, req.GetEmail())
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, status.Error(codes.NotFound, "user not found")
		}
		if errors.Is(err, service.ErrValidation) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		return nil, status.Errorf(codes.Internal, "reset password failed: %v", err)
	}

	return &authpb.ResetPasswordResponse{
		Success: true,
		Message: "Password reset instructions and new temporary password generated",
	}, nil
}

func (s *Server) ValidateToken(ctx context.Context, req *authpb.ValidateTokenRequest) (*authpb.ValidateTokenResponse, error) {
	if req == nil || req.GetToken() == "" {
		return &authpb.ValidateTokenResponse{
			Valid:        false,
			ErrorMessage: "token is required",
		}, nil
	}

	claims, err := s.authService.ValidateToken(ctx, req.GetToken())
	if err != nil {
		return &authpb.ValidateTokenResponse{
			Valid:        false,
			ErrorMessage: err.Error(),
		}, nil
	}

	return &authpb.ValidateTokenResponse{
		Valid:        true,
		UserId:       claims.UserID,
		Email:        claims.Email,
		FirstName:    claims.FirstName,
		LastName:     claims.LastName,
		PhoneNumber:  claims.PhoneNumber,
		ErrorMessage: "",
	}, nil
}

func (s *Server) GetUser(ctx context.Context, req *authpb.GetUserRequest) (*authpb.UserResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is nil")
	}

	user, err := s.authService.GetUser(ctx, req.GetUserId(), req.GetEmail())
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, status.Error(codes.NotFound, "user not found")
		}
		return nil, status.Errorf(codes.Internal, "get user failed: %v", err)
	}

	return &authpb.UserResponse{
		UserId:      user.ID,
		Email:       user.Email,
		FirstName:   user.FirstName,
		LastName:    user.LastName,
		PhoneNumber: user.PhoneNumber,
		CreatedAt:   user.CreatedAt.Format(time.RFC3339),
	}, nil
}
