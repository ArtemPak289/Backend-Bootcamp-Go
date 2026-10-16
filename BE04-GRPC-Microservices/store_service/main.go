package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"store_api/internal/client"
	"store_api/internal/config"
	"store_api/internal/database"
	"store_api/internal/di"
)

// @title Appliance Store REST API with gRPC Authentication
// @version 2.0
// @description RESTful API for Appliance Store integrated with gRPC Authorization Service.
// @termsOfService http://swagger.io/terms/

// @contact.name School 21 Student
// @contact.url https://21-school.ru

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /api/v1
func main() {
	log.Println("[STORE SERVICE] Initializing Appliance Store with IoC Container...")

	container, err := di.BuildContainer()
	if err != nil {
		log.Fatalf("Failed to build DI container: %v", err)
	}

	err = container.Invoke(func(cfg *config.Config, db *sql.DB, authClient client.IAuthClient, srv *http.Server) error {
		defer db.Close()
		defer authClient.Close()

		log.Println("[STORE SERVICE] Running database migrations...")
		if err := database.RunMigrations(db); err != nil {
			log.Fatalf("Database migrations failed: %v", err)
		}
		log.Println("[STORE SERVICE] Database migrations applied successfully.")

		serverErrors := make(chan error, 1)
		go func() {
			log.Printf("[STORE SERVICE] Server listening on http://localhost:%s", cfg.Port)
			log.Printf("[STORE SERVICE] Swagger UI available at http://localhost:%s/swagger/index.html", cfg.Port)
			serverErrors <- srv.ListenAndServe()
		}()

		shutdown := make(chan os.Signal, 1)
		signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

		select {
		case err := <-serverErrors:
			return err
		case sig := <-shutdown:
			log.Printf("[STORE SERVICE] Caught signal %v, shutting down gracefully...", sig)
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			if err := srv.Shutdown(ctx); err != nil {
				_ = srv.Close()
				return err
			}
			log.Println("[STORE SERVICE] Server stopped cleanly")
			return nil
		}
	})

	if err != nil && err != context.Canceled {
		log.Fatalf("Fatal server error: %v", err)
	}
}
