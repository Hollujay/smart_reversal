// Command api runs the SMARTREVERSAL HTTP API.
package main

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/smartreversal/smartreversal/apps/api/internal/config"
	"github.com/smartreversal/smartreversal/apps/api/internal/observability"
	"github.com/smartreversal/smartreversal/apps/api/internal/server"
)

// version is set at build time with -ldflags "-X main.version=...".
var version = "dev"

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "api:", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("configuration: %w", err)
	}

	logger := observability.NewLogger(cfg.Environment, cfg.LogLevel)
	logger.Info("api_starting", slog.String("version", version), slog.Any("config", cfg))

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.APIPort))
	if err != nil {
		return fmt.Errorf("listen: %w", err)
	}

	handler := server.NewRouter(server.Options{Logger: logger, Version: version})
	return server.Serve(ctx, ln, handler, logger, server.Timeouts{
		Read:     cfg.ReadTimeout,
		Write:    cfg.WriteTimeout,
		Idle:     cfg.IdleTimeout,
		Shutdown: cfg.ShutdownTimeout,
	})
}
