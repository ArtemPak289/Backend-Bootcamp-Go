package di

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"go.uber.org/dig"

	"store_api/internal/auth"
	"store_api/internal/client"
	"store_api/internal/config"
	"store_api/internal/database"
	"store_api/internal/handlers"
	"store_api/internal/mappers"
	"store_api/internal/repository"
	"store_api/internal/router"
)

// BuildContainer sets up the Uber Dig IoC container with interface implementations.
func BuildContainer() (*dig.Container, error) {
	c := dig.New()

	// Configuration
	if err := c.Provide(config.Load); err != nil {
		return nil, err
	}

	// Database Connection
	if err := c.Provide(func(cfg *config.Config) (*sql.DB, error) {
		return database.Connect(cfg)
	}); err != nil {
		return nil, err
	}

	// Repositories (injected via interfaces through IoC container)
	if err := c.Provide(func(db *sql.DB) repository.AddressRepository {
		return repository.NewAddressRepository(db)
	}); err != nil {
		return nil, err
	}

	if err := c.Provide(func(db *sql.DB) repository.ClientRepository {
		return repository.NewClientRepository(db)
	}); err != nil {
		return nil, err
	}

	if err := c.Provide(func(db *sql.DB) repository.ProductRepository {
		return repository.NewProductRepository(db)
	}); err != nil {
		return nil, err
	}

	if err := c.Provide(func(db *sql.DB) repository.SupplierRepository {
		return repository.NewSupplierRepository(db)
	}); err != nil {
		return nil, err
	}

	if err := c.Provide(func(db *sql.DB) repository.ImageRepository {
		return repository.NewImageRepository(db)
	}); err != nil {
		return nil, err
	}

	// Mappers
	if err := c.Provide(mappers.NewClientMapper); err != nil {
		return nil, err
	}
	if err := c.Provide(mappers.NewProductMapper); err != nil {
		return nil, err
	}
	if err := c.Provide(mappers.NewSupplierMapper); err != nil {
		return nil, err
	}
	if err := c.Provide(mappers.NewImageMapper); err != nil {
		return nil, err
	}

	// gRPC Auth Client (injected via IAuthClient interface)
	if err := c.Provide(func(cfg *config.Config) (client.IAuthClient, error) {
		return client.NewAuthClient(cfg)
	}); err != nil {
		return nil, err
	}

	// Middleware
	if err := c.Provide(auth.NewAuthMiddleware); err != nil {
		return nil, err
	}

	// Controllers / Handlers
	if err := c.Provide(handlers.NewClientHandler); err != nil {
		return nil, err
	}
	if err := c.Provide(handlers.NewProductHandler); err != nil {
		return nil, err
	}
	if err := c.Provide(handlers.NewSupplierHandler); err != nil {
		return nil, err
	}
	if err := c.Provide(handlers.NewImageHandler); err != nil {
		return nil, err
	}
	if err := c.Provide(handlers.NewAuthHandler); err != nil {
		return nil, err
	}

	// Router
	if err := c.Provide(router.NewRouter); err != nil {
		return nil, err
	}

	// HTTP Server
	if err := c.Provide(func(cfg *config.Config, r *mux.Router) *http.Server {
		return &http.Server{
			Addr:         ":" + cfg.Port,
			Handler:      r,
			ReadTimeout:  15 * time.Second,
			WriteTimeout: 15 * time.Second,
			IdleTimeout:  60 * time.Second,
		}
	}); err != nil {
		return nil, err
	}

	return c, nil
}
