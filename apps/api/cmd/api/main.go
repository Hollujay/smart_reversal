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

	"github.com/smartreversal/smartreversal/apps/api/internal/audit"
	"github.com/smartreversal/smartreversal/apps/api/internal/authz"
	"github.com/smartreversal/smartreversal/apps/api/internal/config"
	"github.com/smartreversal/smartreversal/apps/api/internal/datashield"
	"github.com/smartreversal/smartreversal/apps/api/internal/identity"
	"github.com/smartreversal/smartreversal/apps/api/internal/observability"
	"github.com/smartreversal/smartreversal/apps/api/internal/platform/db"
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

	pool, err := db.Open(ctx, cfg.DatabaseURL(), db.Options{
		MaxConns:        int32(cfg.DBMaxConns),
		MaxConnLifetime: cfg.DBConnMaxLifetime,
	})
	if err != nil {
		return err
	}
	defer pool.Close()

	signer, verifier, err := identity.Load(cfg.JWTKeyID, cfg.JWTPrivateKey, cfg.JWTPublicKeys, cfg.JWTTokenTTL)
	if err != nil {
		return err
	}

	policy, err := authz.Load(ctx, pool)
	if err != nil {
		return err
	}
	appender := &audit.Appender{}

	var vault *datashield.Vault
	if cfg.AccountTokenKey != "" && cfg.AccountEncKey != "" {
		tokenizer, err := datashield.NewTokenizer(cfg.AccountTokenKey)
		if err != nil {
			return err
		}
		sealer, err := datashield.NewSealer(cfg.AccountEncKeyID, cfg.AccountEncKey)
		if err != nil {
			return err
		}
		vault = &datashield.Vault{Tokenizer: tokenizer, Sealer: sealer}
	} else {
		logger.Warn("account_vault_disabled", "reason", "ACCOUNT_TOKEN_KEY and ACCOUNT_ENC_KEY are not both set; the unmask route is not mounted")
	}

	handler := server.NewRouter(server.Options{
		Logger:   logger,
		Version:  version,
		Checkers: []server.Checker{pool},
		Verifier: verifier,
		Auth: &identity.Authenticator{
			Store:    &identity.PGUserStore{Pool: pool},
			Signer:   signer,
			Observer: &identity.AuditLoginObserver{Pool: pool, Appender: appender, Logger: logger},
		},
		Services: &server.Services{Pool: pool, Policy: policy, Audit: appender, Vault: vault},
	})
	return server.Serve(ctx, ln, handler, logger, server.Timeouts{
		Read:     cfg.ReadTimeout,
		Write:    cfg.WriteTimeout,
		Idle:     cfg.IdleTimeout,
		Shutdown: cfg.ShutdownTimeout,
	})
}
