package main

import (
	"context"
	"database/sql"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"google.golang.org/grpc"

	"auth_service/internal/config"
	"auth_service/internal/di"
)

func main() {
	log.Println("[AUTH SERVICE] Initializing Authorization Service with IoC Container...")

	container, err := di.BuildContainer()
	if err != nil {
		log.Fatalf("Failed to build DI container: %v", err)
	}

	err = container.Invoke(func(cfg *config.Config, db *sql.DB, grpcServer *grpc.Server) error {
		defer db.Close()

		listenAddr := ":" + cfg.GRPCPort
		lis, err := net.Listen("tcp", listenAddr)
		if err != nil {
			return err
		}

		log.Printf("[AUTH SERVICE] gRPC server listening on %s (DB: %s@%s:%s/%s)",
			listenAddr, cfg.DBUser, cfg.DBHost, cfg.DBPort, cfg.DBName)

		serverErrors := make(chan error, 1)
		go func() {
			serverErrors <- grpcServer.Serve(lis)
		}()

		shutdown := make(chan os.Signal, 1)
		signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

		select {
		case err := <-serverErrors:
			return err
		case sig := <-shutdown:
			log.Printf("[AUTH SERVICE] Caught signal %v, gracefully stopping gRPC server...", sig)
			grpcServer.GracefulStop()
			log.Println("[AUTH SERVICE] Server stopped cleanly")
			return nil
		}
	})

	if err != nil && err != context.Canceled {
		log.Fatalf("Fatal error running auth service: %v", err)
	}
}
