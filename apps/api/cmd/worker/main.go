// Command worker runs the SMARTREVERSAL background processes.
package main

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/smartreversal/smartreversal/apps/api/internal/audit"
	"github.com/smartreversal/smartreversal/apps/api/internal/cases"
	"github.com/smartreversal/smartreversal/apps/api/internal/config"
	"github.com/smartreversal/smartreversal/apps/api/internal/configsvc"
	"github.com/smartreversal/smartreversal/apps/api/internal/detect"
	"github.com/smartreversal/smartreversal/apps/api/internal/observability"
	"github.com/smartreversal/smartreversal/apps/api/internal/platform/db"
	"github.com/smartreversal/smartreversal/apps/api/internal/platform/redisx"
	"github.com/smartreversal/smartreversal/apps/api/internal/server"
	"github.com/smartreversal/smartreversal/apps/api/internal/worker"
)

// version is set at build time with -ldflags "-X main.version=...".
var version = "dev"

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "worker:", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("configuration: %w", err)
	}
	logger := observability.NewLogger(cfg.Environment, cfg.LogLevel)
	logger.Info("worker_starting", slog.String("version", version), slog.Any("config", cfg))

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	pool, err := db.Open(ctx, cfg.DatabaseURL(), db.Options{
		MaxConns:        int32(cfg.DBMaxConns),
		MaxConnLifetime: cfg.DBConnMaxLifetime,
	})
	if err != nil {
		return err
	}
	defer pool.Close()

	rc, err := redisx.Open(ctx, redisx.Options{Addr: cfg.RedisAddr(), Password: cfg.RedisPassword, DB: cfg.RedisDB})
	if err != nil {
		return err
	}
	defer rc.Close()

	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.WorkerHealthPort))
	if err != nil {
		return fmt.Errorf("listen: %w", err)
	}

	app := &audit.Appender{}
	cfgSvc := &configsvc.Service{Audit: app}
	detector := &detect.Detector{Cases: &cases.Service{Audit: app, Config: cfgSvc}, Config: cfgSvc}

	return worker.Run(ctx, worker.Deps{
		Detector:        detector,
		Pool:            pool,
		Redis:           rc,
		Logger:          logger,
		HealthListener:  ln,
		Version:         version,
		OutboxInterval:  cfg.OutboxPollInterval,
		OutboxBatchSize: cfg.OutboxBatchSize,
		Timeouts: server.Timeouts{
			Read:     cfg.ReadTimeout,
			Write:    cfg.WriteTimeout,
			Idle:     cfg.IdleTimeout,
			Shutdown: cfg.ShutdownTimeout,
		},
	})
}
