// Package dbtest gives integration tests a migrated database.
package dbtest

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/smartreversal/smartreversal/apps/api/internal/platform/db"
	"github.com/smartreversal/smartreversal/apps/api/internal/platform/migrate"
	"github.com/smartreversal/smartreversal/apps/api/migrations"
)

// Pool opens TEST_DATABASE_URL, applies every pending migration and returns the
// pool. The test is skipped when the variable is not set. Use a disposable
// database: tests insert and delete rows.
func Pool(t testing.TB) *db.Pool {
	t.Helper()
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TEST_DATABASE_URL not set; skipping database integration test")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	p, err := db.Open(ctx, url, db.Options{MaxConns: 8})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(p.Close)

	ms, err := migrate.Load(migrations.FS, ".")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := migrate.New(p.Pool, ms).Up(ctx); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return p
}
