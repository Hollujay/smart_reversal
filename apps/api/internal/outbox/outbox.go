// Package outbox implements the transactional outbox. Callers write an event
// in the same PostgreSQL transaction as the state change it describes, so a
// committed change always has its event and a rolled back change never does.
// A relay (relay.go) delivers the rows to Redis Streams afterwards.
package outbox

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

var streamName = regexp.MustCompile(`^[a-z0-9._-]{1,100}$`)

// Event is one message to publish.
type Event struct {
	// TenantID is the institution the event belongs to. Required.
	TenantID uuid.UUID
	// ID identifies the event for idempotent consumers. A random UUID is
	// generated when it is zero.
	ID uuid.UUID
	// Stream is the Redis stream the event is published to.
	Stream string
	// Type names what happened, for example "transaction.ingested".
	Type string
	// AggregateID is the entity the event is about. Events with the same
	// aggregate are published in insertion order.
	AggregateID string
	// Payload is marshalled to JSON. It must not contain secrets or raw
	// account numbers.
	Payload any
	TraceID string
}

// Validate reports whether the event can be stored.
func (e Event) Validate() error {
	switch {
	case e.TenantID == uuid.Nil:
		return errors.New("outbox: tenant id is required")
	case !streamName.MatchString(e.Stream):
		return errors.New("outbox: stream must match [a-z0-9._-]{1,100}")
	case e.Type == "":
		return errors.New("outbox: event type is required")
	case e.AggregateID == "":
		return errors.New("outbox: aggregate id is required")
	}
	return nil
}

// Enqueue stores e using tx. It returns the event id. The event becomes
// visible to the relay only when tx commits.
func Enqueue(ctx context.Context, tx pgx.Tx, e Event) (uuid.UUID, error) {
	if err := e.Validate(); err != nil {
		return uuid.Nil, err
	}
	if e.ID == uuid.Nil {
		e.ID = uuid.New()
	}
	payload, err := json.Marshal(e.Payload)
	if err != nil {
		return uuid.Nil, fmt.Errorf("outbox: marshal payload: %w", err)
	}
	if string(payload) == "null" {
		return uuid.Nil, errors.New("outbox: payload is required")
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO outbox_events (tenant_id, event_id, stream, event_type, aggregate_id, payload, trace_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		e.TenantID, e.ID, e.Stream, e.Type, e.AggregateID, payload, e.TraceID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("outbox: insert: %w", err)
	}
	return e.ID, nil
}
