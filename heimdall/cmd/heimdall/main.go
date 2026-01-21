package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/AnubhavMadhav/project-heimdall/internal/adapters/config"
	"github.com/AnubhavMadhav/project-heimdall/internal/adapters/mcp"
	"github.com/AnubhavMadhav/project-heimdall/internal/adapters/postgres"
	"github.com/AnubhavMadhav/project-heimdall/internal/adapters/security"
	"github.com/AnubhavMadhav/project-heimdall/internal/core/services"
	"github.com/AnubhavMadhav/project-heimdall/pkg/logger"
)

func main() {
	// 1. Initialize Logger
	log := logger.New(slog.LevelInfo)
	slog.SetDefault(log)

	// 2. Load Config
	cfg, err := config.Load()
	if err != nil {
		log.Error("Failed to load configuration", "error", err)
		os.Exit(1)
	}

	// 3. Infrastructure (Adapters)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	repo, err := postgres.NewPostgresRepo(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Error("Failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer repo.Close(context.Background())

	validator := security.NewValidator()

	// 4. Domain Layer (Service)
	gatekeeper := services.NewGatekeeperService(repo, validator, log)

	// 5. Interface Layer (MCP Server)
	mcpServer := mcp.NewMCPServer(cfg.ServiceName, cfg.Version, gatekeeper, log)

	// 6. Run Server
	go func() {
		log.Info("Heimdall is listening on Stdio...")
		if err := mcpServer.Start(); err != nil {
			log.Error("MCP Server error", "error", err)
			cancel()
		}
	}()

	// 7. Graceful Shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan
	log.Info("Shutting down Heimdall...")
}
