// Package worker runs the background processes: the outbox relay and a small
// health endpoint for orchestrators. Stream consumers are added by the phases
// that own them.
package worker

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"time"

	"github.com/smartreversal/smartreversal/apps/api/internal/outbox"
	"github.com/smartreversal/smartreversal/apps/api/internal/platform/db"
	"github.com/smartreversal/smartreversal/apps/api/internal/platform/redisx"
	"github.com/smartreversal/smartreversal/apps/api/internal/server"
)

// Deps are the worker's collaborators.
type Deps struct {
	Pool   *db.Pool
	Redis  *redisx.Client
	Logger *slog.Logger
	// HealthListener serves /v1/health and /v1/ready.
	HealthListener net.Listener
	Version        string

	OutboxInterval  time.Duration
	OutboxBatchSize int
	Timeouts        server.Timeouts
}

// Run starts every background task and blocks until ctx is cancelled or one
// task fails. A failure of one task stops the others, so an orchestrator can
// restart the whole process rather than leave it half alive.
func Run(ctx context.Context, d Deps) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	relay := &outbox.Relay{
		Pool:      d.Pool,
		Redis:     d.Redis.Client,
		Logger:    d.Logger,
		BatchSize: d.OutboxBatchSize,
	}
	health := server.NewRouter(server.Options{
		Logger:   d.Logger,
		Version:  d.Version,
		Service:  "worker",
		Checkers: []server.Checker{d.Pool, d.Redis},
	})

	errc := make(chan error, 2)
	go func() {
		errc <- server.Serve(ctx, d.HealthListener, health, d.Logger, d.Timeouts)
	}()
	go func() {
		errc <- relay.Run(ctx, d.OutboxInterval)
	}()

	var first error
	for i := 0; i < 2; i++ {
		err := <-errc
		if err != nil && first == nil && !errors.Is(err, context.Canceled) {
			first = err
		}
		cancel() // when one task ends, stop the other
	}
	return first
}
