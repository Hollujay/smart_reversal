// Package dbtest gives integration tests a migrated database.
package dbtest

import (
	"strings"

	"context"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
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

// Tenant creates an institution for one test and removes it, and everything
// that refers to it, when the test ends. It uses the pool's own role, which is
// not subject to row-level security.
func Tenant(t testing.TB, p *db.Pool) uuid.UUID {
	t.Helper()
	ctx := context.Background()
	code := "T-" + strings.ToUpper(uuid.NewString()[:12])
	var id uuid.UUID
	if err := p.QueryRow(ctx, `INSERT INTO institutions (code, name) VALUES ($1, $1) RETURNING id`, code).Scan(&id); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		bg := context.Background()
		// The audit trail refuses deletes, even from the owner. Tests clean up
		// as a superuser with triggers and foreign keys switched off for this
		// one transaction. This only works for a superuser, which test
		// databases use; production roles cannot do it.
		_ = p.WithTx(bg, func(tx pgx.Tx) error {
			if _, err := tx.Exec(bg, `SET LOCAL session_replication_role = replica`); err != nil {
				return err
			}
			for _, q := range []string{
				`DELETE FROM audit_events WHERE tenant_id = $1`,
				`DELETE FROM audit_heads WHERE tenant_id = $1`,
			} {
				if _, err := tx.Exec(bg, q, id); err != nil {
					return err
				}
			}
			return nil
		})
		for _, q := range []string{
			`DELETE FROM account_vault WHERE tenant_id = $1`,
			`DELETE FROM processed_events WHERE tenant_id = $1`,
			`DELETE FROM outbox_events WHERE tenant_id = $1`,
			`DELETE FROM users WHERE tenant_id = $1`,
			`DELETE FROM institutions WHERE id = $1`,
		} {
			_, _ = p.Exec(bg, q, id)
		}
	})
	return id
}
