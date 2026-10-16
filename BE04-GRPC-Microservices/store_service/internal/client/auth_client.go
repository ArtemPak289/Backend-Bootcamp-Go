package client

import (
	"context"
	"fmt"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"store_api/internal/config"
	authpb "store_api/proto"
)

type IAuthClient interface {
	Register(ctx context.Context, email, firstName, lastName, phone, password string) (*authpb.AuthResponse, error)
	Login(ctx context.Context, email, password string) (*authpb.AuthResponse, error)
	ChangePassword(ctx context.Context, email, oldPassword, newPassword string) (*authpb.ChangePasswordResponse, error)
	ResetPassword(ctx context.Context, email string) (*authpb.ResetPasswordResponse, error)
	ValidateToken(ctx context.Context, token string) (*authpb.ValidateTokenResponse, error)
	GetUser(ctx context.Context, userID, email string) (*authpb.UserResponse, error)
	Close() error
}

type GRPCAuthClient struct {
	conn   *grpc.ClientConn
	client authpb.AuthServiceClient
}

func NewAuthClient(cfg *config.Config) (IAuthClient, error) {
	log.Printf("[STORE SERVICE] Dialing Auth gRPC service at %s (HTTP/2)...", cfg.AuthServiceGRPCAddr)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(
		ctx,
		cfg.AuthServiceGRPCAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		log.Printf("[STORE SERVICE] Warning: could not connect to Auth Service immediately (%v), will retry on calls", err)
		// Non-blocking connection fallback
		conn, err = grpc.Dial(cfg.AuthServiceGRPCAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			return nil, fmt.Errorf("failed to dial auth service: %w", err)
		}
	}

	return &GRPCAuthClient{
		conn:   conn,
		client: authpb.NewAuthServiceClient(conn),
	}, nil
}

func (c *GRPCAuthClient) Register(ctx context.Context, email, firstName, lastName, phone, password string) (*authpb.AuthResponse, error) {
	return c.client.Register(ctx, &authpb.RegisterRequest{
		Email:       email,
		FirstName:   firstName,
		LastName:    lastName,
		PhoneNumber: phone,
		Password:    password,
	})
}

func (c *GRPCAuthClient) Login(ctx context.Context, email, password string) (*authpb.AuthResponse, error) {
	return c.client.Login(ctx, &authpb.LoginRequest{
		Email:    email,
		Password: password,
	})
}

func (c *GRPCAuthClient) ChangePassword(ctx context.Context, email, oldPassword, newPassword string) (*authpb.ChangePasswordResponse, error) {
	return c.client.ChangePassword(ctx, &authpb.ChangePasswordRequest{
		Email:       email,
		OldPassword: oldPassword,
		NewPassword: newPassword,
	})
}

func (c *GRPCAuthClient) ResetPassword(ctx context.Context, email string) (*authpb.ResetPasswordResponse, error) {
	return c.client.ResetPassword(ctx, &authpb.ResetPasswordRequest{
		Email: email,
	})
}

func (c *GRPCAuthClient) ValidateToken(ctx context.Context, token string) (*authpb.ValidateTokenResponse, error) {
	return c.client.ValidateToken(ctx, &authpb.ValidateTokenRequest{
		Token: token,
	})
}

func (c *GRPCAuthClient) GetUser(ctx context.Context, userID, email string) (*authpb.UserResponse, error) {
	return c.client.GetUser(ctx, &authpb.GetUserRequest{
		UserId: userID,
		Email:  email,
	})
}

func (c *GRPCAuthClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}
