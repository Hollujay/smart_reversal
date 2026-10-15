package worker

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/smartreversal/smartreversal/apps/api/internal/outbox"
	"github.com/smartreversal/smartreversal/apps/api/internal/platform/db"
	"github.com/smartreversal/smartreversal/apps/api/internal/platform/dbtest"
	"github.com/smartreversal/smartreversal/apps/api/internal/platform/redisx"
	"github.com/smartreversal/smartreversal/apps/api/internal/server"
)

func deps(t *testing.T) Deps {
	t.Helper()
	url, addr := os.Getenv("TEST_DATABASE_URL"), os.Getenv("TEST_REDIS_ADDR")
	if url == "" || addr == "" {
		t.Skip("TEST_DATABASE_URL and TEST_REDIS_ADDR are required; skipping integration test")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := db.Open(ctx, url, db.Options{MaxConns: 4})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	rc, err := redisx.Open(ctx, redisx.Options{Addr: addr})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = rc.Close() })

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	return Deps{
		Pool: pool, Redis: rc,
		Logger:         slog.New(slog.NewTextHandler(io.Discard, nil)),
		HealthListener: ln, Version: "test",
		OutboxInterval: 20 * time.Millisecond, OutboxBatchSize: 10,
		Timeouts: server.Timeouts{Read: time.Second, Write: time.Second, Idle: time.Second, Shutdown: 2 * time.Second},
	}
}

func TestWorkerPublishesOutboxAndStopsOnCancel(t *testing.T) {
	d := deps(t)
	ctx := context.Background()
	stream := fmt.Sprintf("test.worker.%d", time.Now().UnixNano())
	t.Cleanup(func() { d.Redis.Del(context.Background(), stream) })
	if _, err := d.Pool.Exec(ctx, `DELETE FROM outbox_events`); err != nil {
		t.Fatalf("outbox_events missing (run migrate up): %v", err)
	}

	runCtx, cancel := context.WithCancel(ctx)
	done := make(chan error, 1)
	go func() { done <- Run(runCtx, d) }()

	base := "http://" + d.HealthListener.Addr().String()
	waitFor(t, func() bool {
		resp, err := http.Get(base + "/v1/ready")
		if err != nil {
			return false
		}
		defer resp.Body.Close()
		return resp.StatusCode == http.StatusOK
	}, "worker ready")

	tid := dbtest.Tenant(t, d.Pool)
	err := d.Pool.WithTenant(ctx, tid, func(tx pgx.Tx) error {
		_, err := outbox.Enqueue(ctx, tx, outbox.Event{
			TenantID: tid, Stream: stream, Type: "test.happened", AggregateID: "A-1", Payload: map[string]int{"n": 1},
		})
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
	waitFor(t, func() bool {
		n, _ := d.Redis.XLen(ctx, stream).Result()
		return n == 1
	}, "event published to the stream")

	cancel()
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("Run returned %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("worker did not stop after cancel")
	}
}

func TestWorkerStopsWhenARelayFails(t *testing.T) {
	d := deps(t)
	d.OutboxInterval = -1 // the relay rejects this and returns an error at once

	done := make(chan error, 1)
	go func() { done <- Run(context.Background(), d) }()
	select {
	case err := <-done:
		if err == nil {
			t.Fatal("expected the relay error to be returned")
		}
	case <-time.After(5 * time.Second):
		t.Fatal("worker kept running after a task failed")
	}
}

func waitFor(t *testing.T, cond func() bool, what string) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if cond() {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatalf("timed out waiting for: %s", what)
}
