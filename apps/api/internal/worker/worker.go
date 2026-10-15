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

	"github.com/smartreversal/smartreversal/apps/api/internal/detect"
	"github.com/smartreversal/smartreversal/apps/api/internal/outbox"
	"github.com/smartreversal/smartreversal/apps/api/internal/platform/db"
	"github.com/smartreversal/smartreversal/apps/api/internal/platform/redisx"
	"github.com/smartreversal/smartreversal/apps/api/internal/server"
	"github.com/smartreversal/smartreversal/apps/api/internal/stream"
	"github.com/smartreversal/smartreversal/apps/api/internal/txn"
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

	// Detector, when set, consumes the transactions stream and runs the
	// stuck-transaction sweep every SweepInterval (default 30 seconds).
	Detector      *detect.Detector
	SweepInterval time.Duration
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

	tasks := 2
	errc := make(chan error, 4)
	if d.Detector != nil {
		tasks += 2
		consumer := &stream.Consumer{Pool: d.Pool, Redis: d.Redis.Client, Logger: d.Logger,
			Stream: txn.Stream, Group: "detector", Name: "worker-1", Handler: d.Detector.Handle}
		go func() { errc <- consumer.Run(ctx) }()
		go func() { errc <- sweepLoop(ctx, d) }()
	}
	go func() {
		errc <- server.Serve(ctx, d.HealthListener, health, d.Logger, d.Timeouts)
	}()
	go func() {
		errc <- relay.Run(ctx, d.OutboxInterval)
	}()

	var first error
	for i := 0; i < tasks; i++ {
		err := <-errc
		if err != nil && first == nil && !errors.Is(err, context.Canceled) {
			first = err
		}
		cancel() // when one task ends, stop the other
	}
	return first
}

// sweepLoop runs the stuck-transaction sweep until ctx is cancelled. A failed
// sweep is logged and tried again on the next tick.
func sweepLoop(ctx context.Context, d Deps) error {
	every := d.SweepInterval
	if every <= 0 {
		every = 30 * time.Second
	}
	t := time.NewTicker(every)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-t.C:
			if n, err := d.Detector.SweepAll(ctx, d.Pool); err != nil && ctx.Err() == nil {
				d.Logger.Error("sweep_failed", "error", err)
			} else if n > 0 {
				d.Logger.Info("sweep_opened_cases", "count", n)
			}
		}
	}
}
