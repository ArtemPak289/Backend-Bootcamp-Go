package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"store_api/internal/config"
	"store_api/internal/database"
	"store_api/internal/handlers"
	"store_api/internal/mappers"
	"store_api/internal/repository"
	"store_api/internal/router"
)

// @title Appliance Store REST API
// @version 1.0
// @description RESTful API for Appliance Store with PostgreSQL database.
// @termsOfService http://swagger.io/terms/

// @contact.name School 21 Student
// @contact.url https://21-school.ru

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /api/v1
func main() {
	cfg := config.Load()

	log.Printf("Connecting to PostgreSQL at %s:%s (db=%s, user=%s)...",
		cfg.DBHost, cfg.DBPort, cfg.DBName, cfg.DBUser)

	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatalf("Database connection failed: %v", err)
	}
	defer db.Close()

	if !cfg.ReadOnly {
		log.Println("Running database migrations...")
		if err := database.RunMigrations(db); err != nil {
			log.Fatalf("Database migrations failed: %v", err)
		}
		log.Println("Database migrations applied successfully.")
	} else {
		log.Println("Instance running in READ-ONLY mode: skipping database migrations.")
	}

	// Initialize repositories (DAL)
	addrRepo := repository.NewAddressRepository(db)
	clientRepo := repository.NewClientRepository(db)
	productRepo := repository.NewProductRepository(db)
	supplierRepo := repository.NewSupplierRepository(db)
	imageRepo := repository.NewImageRepository(db)

	// Initialize mappers
	clientMapper := mappers.NewClientMapper()
	productMapper := mappers.NewProductMapper()
	supplierMapper := mappers.NewSupplierMapper()
	imageMapper := mappers.NewImageMapper()

	// Initialize handlers (Controllers)
	clientHandler := handlers.NewClientHandler(clientRepo, addrRepo, clientMapper)
	productHandler := handlers.NewProductHandler(productRepo, supplierRepo, imageRepo, productMapper)
	supplierHandler := handlers.NewSupplierHandler(supplierRepo, addrRepo, supplierMapper)
	imageHandler := handlers.NewImageHandler(imageRepo, productRepo, imageMapper)

	// Initialize router
	r := router.NewRouter(cfg, clientHandler, productHandler, supplierHandler, imageHandler)

	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Server run context
	serverErrors := make(chan error, 1)
	go func() {
		log.Printf("Server listening on http://localhost:%s", cfg.Port)
		log.Printf("Swagger UI available at http://localhost:%s/swagger/index.html", cfg.Port)
		serverErrors <- srv.ListenAndServe()
	}()

	// Graceful shutdown
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-serverErrors:
		log.Fatalf("Error starting server: %v", err)
	case sig := <-shutdown:
		log.Printf("Caught signal %v, shutting down server...", sig)
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := srv.Shutdown(ctx); err != nil {
			_ = srv.Close()
			log.Fatalf("Could not stop server gracefully: %v", err)
		}
	}
}
