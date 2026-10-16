package di

import (
	"database/sql"

	"go.uber.org/dig"
	"google.golang.org/grpc"

	"auth_service/internal/config"
	"auth_service/internal/database"
	"auth_service/internal/grpcserver"
	"auth_service/internal/repository"
	"auth_service/internal/security"
	"auth_service/internal/service"
	authpb "auth_service/proto"
)

// BuildContainer configures the Uber Dig IoC container with all dependencies and interfaces.
func BuildContainer() (*dig.Container, error) {
	container := dig.New()

	// Configuration
	if err := container.Provide(config.Load); err != nil {
		return nil, err
	}

	// Database Connection (injected into repositories)
	if err := container.Provide(func(cfg *config.Config) (*sql.DB, error) {
		return database.Connect(cfg)
	}); err != nil {
		return nil, err
	}

	// Security components (injected via interfaces)
	if err := container.Provide(func() security.IPasswordHasher {
		return security.NewSHA256SaltHasher()
	}); err != nil {
		return nil, err
	}

	if err := container.Provide(func(cfg *config.Config) security.ITokenManager {
		return security.NewJWTTokenManager(cfg)
	}); err != nil {
		return nil, err
	}

	// Repository Layer (injected via IUserRepository interface)
	if err := container.Provide(func(db *sql.DB) repository.IUserRepository {
		return repository.NewPostgresUserRepository(db)
	}); err != nil {
		return nil, err
	}

	// Business Service Layer (injected via IAuthService interface)
	if err := container.Provide(func(
		userRepo repository.IUserRepository,
		hasher security.IPasswordHasher,
		tokenMgr security.ITokenManager,
	) service.IAuthService {
		return service.NewAuthService(userRepo, hasher, tokenMgr)
	}); err != nil {
		return nil, err
	}

	// gRPC Server Handlers
	if err := container.Provide(grpcserver.NewServer); err != nil {
		return nil, err
	}

	// gRPC Server Instance
	if err := container.Provide(func(srv *grpcserver.Server) *grpc.Server {
		grpcServer := grpc.NewServer()
		authpb.RegisterAuthServiceServer(grpcServer, srv)
		return grpcServer
	}); err != nil {
		return nil, err
	}

	return container, nil
}
