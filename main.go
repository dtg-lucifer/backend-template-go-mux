package main

import (
	"context"
	"flag"
	"log/slog"
	"os"

	"github.com/joho/godotenv"
	"github.com/your-username/go-mux-backend-template/config"
	"github.com/your-username/go-mux-backend-template/internal/server"
	"github.com/your-username/go-mux-backend-template/pkg"
)

func main() {
	configPath := flag.String("config", "config.yaml", "path to config file")
	flag.Parse()

	// Load .env — non-fatal if missing (production uses real env vars)
	if err := godotenv.Load(); err != nil {
		slog.Warn("No .env file found, relying on environment variables")
	}

	logger := pkg.NewLogger()
	if logger == nil {
		slog.Error("Failed to initialise logger, aborting")
		os.Exit(1)
	}
	defer logger.Close()

	cfg, err := config.NewConfig(*configPath)
	if err != nil {
		logger.Error("Failed to load config", "error", err)
		os.Exit(1)
	}

	srv := server.New(cfg, logger)

	if err := srv.Setup(context.Background()); err != nil {
		logger.Error("Failed to set up server", "error", err)
		os.Exit(1)
	}

	// Start blocks until SIGINT/SIGTERM, then calls srv.Shutdown() internally
	srv.Start()
}
