package migrations_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/smartreversal/smartreversal/apps/api/internal/platform/migrate"
	"github.com/smartreversal/smartreversal/apps/api/migrations"
)

func TestEmbeddedMigrationsLoad(t *testing.T) {
	ms, err := migrate.Load(migrations.FS, ".")
	if err != nil {
		t.Fatal(err)
	}
	if len(ms) == 0 {
		t.Fatal("no migrations embedded")
	}
	for _, m := range ms {
		if m.Down == "" {
			t.Errorf("migration %04d_%s has no down script", m.Version, m.Name)
		}
	}
}

// TestEmbeddedMigrationsApplyAndRevert runs the real migrations against a
// disposable database: up, down to empty, up again.
func TestEmbeddedMigrationsApplyAndRevert(t *testing.T) {
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TEST_DATABASE_URL not set; skipping database integration test")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, url)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()

	ms, err := migrate.Load(migrations.FS, ".")
	if err != nil {
		t.Fatal(err)
	}
	r := migrate.New(pool, ms)

	// Start from an empty schema so the test is repeatable.
	if _, err := r.Down(ctx, len(ms)); err != nil {
		t.Logf("pre-clean down: %v", err)
	}
	if done, err := r.Up(ctx); err != nil || len(done) != len(ms) {
		t.Fatalf("up: %v %v", done, err)
	}
	if done, err := r.Down(ctx, len(ms)); err != nil || len(done) != len(ms) {
		t.Fatalf("down: %v %v", done, err)
	}
	if done, err := r.Up(ctx); err != nil || len(done) != len(ms) {
		t.Fatalf("re-up: %v %v", done, err)
	}
}
