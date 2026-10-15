// Package db owns the PostgreSQL connection pool and transaction helper.
// PostgreSQL is the system of record for every financial fact.
package db

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Options tunes the pool.
type Options struct {
	MaxConns        int32
	MaxConnLifetime time.Duration
}

// Pool wraps pgxpool.Pool. It satisfies the server readiness Checker.
type Pool struct {
	*pgxpool.Pool
}

// Open creates a pool and verifies connectivity. Errors never include the
// connection URL, which carries the password.
func Open(ctx context.Context, url string, o Options) (*Pool, error) {
	cfg, err := pgxpool.ParseConfig(url)
	if err != nil {
		return nil, errors.New("db: invalid connection configuration")
	}
	if o.MaxConns > 0 {
		cfg.MaxConns = o.MaxConns
	}
	if o.MaxConnLifetime > 0 {
		cfg.MaxConnLifetime = o.MaxConnLifetime
	}

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, errors.New("db: could not create pool")
	}
	p := &Pool{Pool: pool}
	if err := p.Check(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("db: initial ping failed: %w", errors.Unwrap(err))
	}
	return p, nil
}

// Name identifies the dependency in readiness output.
func (p *Pool) Name() string { return "postgres" }

// Check pings the database.
func (p *Pool) Check(ctx context.Context) error {
	if err := p.Ping(ctx); err != nil {
		return fmt.Errorf("db ping: %w", err)
	}
	return nil
}

// WithTx runs fn inside a transaction. It commits when fn returns nil and rolls
// back when fn returns an error or panics.
func (p *Pool) WithTx(ctx context.Context, fn func(tx pgx.Tx) error) (err error) {
	tx, err := p.Begin(ctx)
	if err != nil {
		return fmt.Errorf("db begin: %w", err)
	}
	defer func() {
		if r := recover(); r != nil {
			_ = tx.Rollback(context.WithoutCancel(ctx))
			panic(r)
		}
	}()

	if err := fn(tx); err != nil {
		if rbErr := tx.Rollback(context.WithoutCancel(ctx)); rbErr != nil {
			return errors.Join(err, fmt.Errorf("db rollback: %w", rbErr))
		}
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("db commit: %w", err)
	}
	return nil
}
