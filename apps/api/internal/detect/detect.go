// Package detect turns transaction state into exception cases (B3).
//
// Every detector is deterministic and explains itself: the case it opens carries
// the rule id and version, the evidence it used and a plain explanation. AI
// scores may later sit beside these verdicts, but never replace them.
package detect

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/smartreversal/smartreversal/apps/api/internal/cases"
	"github.com/smartreversal/smartreversal/apps/api/internal/configsvc"
	"github.com/smartreversal/smartreversal/apps/api/internal/notify"
	"github.com/smartreversal/smartreversal/apps/api/internal/stream"
	"github.com/smartreversal/smartreversal/apps/api/internal/txn"
)

// Rule ids and versions. A change to a rule's behaviour bumps its version.
const (
	RuleFailure   = "DET-FAIL-1"
	RuleUnknown   = "DET-UNK-1"
	RuleDuplicate = "DET-DUP-1"
	RuleStuck     = "DET-STUCK-1"
	RuleVersion   = "1.0"
)

// Finding is one detection result.
type Finding struct {
	Case    cases.Case
	Created bool
}

// Detector evaluates transactions.
type Detector struct {
	Cases  *cases.Service
	Config *configsvc.Service
	Now    func() time.Time
}

func (d *Detector) now() time.Time {
	if d.Now != nil {
		return d.Now().UTC()
	}
	return time.Now().UTC()
}

type txnRow struct {
	ID             uuid.UUID
	Ref            string
	Rail           string
	Amount         int64
	Currency       string
	DebitToken     string
	BeneficiaryTok string
	SessionID      string
	Lifecycle      string
	CreditStatus   string
	DebitStatus    string
	FirstDebitAt   *time.Time
	LastEventAt    time.Time
}

const txnSelect = `SELECT id, instruction_ref, rail, amount_minor, currency, debit_account_token, beneficiary_token, session_id,
	lifecycle_state, credit_status, debit_status, first_debit_at, last_event_at FROM transactions`

func scanTxn(row pgx.Row) (txnRow, error) {
	var t txnRow
	err := row.Scan(&t.ID, &t.Ref, &t.Rail, &t.Amount, &t.Currency, &t.DebitToken, &t.BeneficiaryTok, &t.SessionID,
		&t.Lifecycle, &t.CreditStatus, &t.DebitStatus, &t.FirstDebitAt, &t.LastEventAt)
	return t, err
}

func (d *Detector) evidence(ctx context.Context, tx pgx.Tx, tenant, id uuid.UUID) ([]cases.Evidence, error) {
	rows, err := tx.Query(ctx, `SELECT id, event_type, source_system, source_event_ref, payload_hash, received_at FROM transaction_events
		WHERE tenant_id = $1 AND txn_id = $2 ORDER BY source_ts, id`, tenant, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []cases.Evidence
	for rows.Next() {
		var eid uuid.UUID
		var typ, sys, ref, hash string
		var at time.Time
		if err := rows.Scan(&eid, &typ, &sys, &ref, &hash, &at); err != nil {
			return nil, err
		}
		out = append(out, cases.Evidence{Kind: "transaction_event", Ref: eid.String(), SourceSystem: sys, FetchedAt: at.UTC(), Hash: hash, IsFact: true,
			Detail: map[string]string{"event_type": typ, "source_event_ref": ref}})
	}
	return out, rows.Err()
}

// Evaluate runs every detector against one transaction and returns what it found.
// It is idempotent: a transaction seen again yields the same cases.
func (d *Detector) Evaluate(ctx context.Context, tx pgx.Tx, tenant, txnID uuid.UUID) ([]Finding, error) {
	t, err := scanTxn(tx.QueryRow(ctx, txnSelect+` WHERE tenant_id = $1 AND id = $2`, tenant, txnID))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil // the transaction is gone or not ours: nothing to do
	}
	if err != nil {
		return nil, err
	}
	ev, err := d.evidence(ctx, tx, tenant, txnID)
	if err != nil {
		return nil, err
	}

	var out []Finding
	add := func(f Finding, err error) error {
		if err != nil {
			return err
		}
		if f.Case.ID != uuid.Nil {
			out = append(out, f)
		}
		return nil
	}

	switch t.Lifecycle {
	case txn.StateConfirmedFailure:
		if err := add(d.open(ctx, tx, tenant, t, cases.CodeNoCredit, RuleFailure, "", "beneficiary_side_failure",
			"The debit was posted and the beneficiary credit was reported as failed by the source.", ev)); err != nil {
			return nil, err
		}
	case txn.StateUnknown:
		if err := add(d.open(ctx, tx, tenant, t, cases.CodeUnknown, RuleUnknown, "", "switch_response_failure",
			"The debit was posted and the switch did not answer. The outcome is unknown, which is not the same as failed.", ev)); err != nil {
			return nil, err
		}
	case txn.StateConflicting:
		if err := add(d.open(ctx, tx, tenant, t, cases.CodeUnknown, RuleUnknown, "", "conflicting_evidence",
			"Sources disagree: one reports the credit confirmed and another reports it failed.", ev)); err != nil {
			return nil, err
		}
	case txn.StateConfirmedSuccess:
		if err := d.closeLateSuccess(ctx, tx, tenant, t); err != nil {
			return nil, err
		}
	}

	if t.DebitStatus == txn.DebitPosted && t.FirstDebitAt != nil {
		if err := add(d.duplicate(ctx, tx, tenant, t, ev)); err != nil {
			return nil, err
		}
	}
	return out, nil
}

func (d *Detector) open(ctx context.Context, tx pgx.Tx, tenant uuid.UUID, t txnRow, code, rule, risk, root, why string, ev []cases.Evidence) (Finding, error) {
	amount := t.Amount
	id := t.ID
	c, created, err := d.Cases.Open(ctx, tx, tenant, cases.OpenInput{
		Classification: code, Source: cases.SourceBot, RiskGrade: risk, TxnID: &id, ExposureMinor: &amount, Currency: t.Currency,
		Rail: t.Rail, RootCause: root, RuleID: rule, RuleVersion: RuleVersion, Explanation: why, Evidence: ev, Actor: cases.System,
		At: d.now(),
	})
	if err != nil {
		return Finding{}, err
	}
	if created {
		if _, err := notify.Record(ctx, tx, tenant, c.ID, notify.OpsExceptionOpened(c.Ref, code, t.Ref, t.Amount, t.Currency), c.Ref+":ops:opened"); err != nil {
			return Finding{}, err
		}
		if _, err := notify.Record(ctx, tx, tenant, c.ID, notify.CustomerDetected(t.Amount, t.Currency, t.Ref), c.Ref+":customer:detected"); err != nil {
			return Finding{}, err
		}
	}
	return Finding{Case: c, Created: created}, nil
}

// duplicate looks for another debit with the same account, amount, currency and
// beneficiary inside the configured window. It never matches on account, amount
// and time alone when the earlier transfer failed: a customer who retries after a
// confirmed failure is not a duplicate.
func (d *Detector) duplicate(ctx context.Context, tx pgx.Tx, tenant uuid.UUID, t txnRow, ev []cases.Evidence) (Finding, error) {
	window, err := d.Config.Effective(ctx, tx, tenant, "duplicate.window_seconds", d.now())
	if errors.Is(err, configsvc.ErrNotConfigured) {
		return Finding{}, nil // no approved window, so no duplicate rule runs
	}
	if err != nil {
		return Finding{}, err
	}
	span := time.Duration(window.Value) * time.Second

	rows, err := tx.Query(ctx, txnSelect+` WHERE tenant_id = $1 AND id <> $2 AND debit_account_token = $3 AND amount_minor = $4
		AND currency = $5 AND beneficiary_token = $6 AND debit_status = 'POSTED' AND first_debit_at IS NOT NULL
		AND first_debit_at BETWEEN $7 AND $8 ORDER BY first_debit_at`,
		tenant, t.ID, t.DebitToken, t.Amount, t.Currency, t.BeneficiaryTok, t.FirstDebitAt.Add(-span), t.FirstDebitAt.Add(span))
	if err != nil {
		return Finding{}, err
	}
	var others []txnRow
	for rows.Next() {
		o, err := scanTxn(rows)
		if err != nil {
			rows.Close()
			return Finding{}, err
		}
		others = append(others, o)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return Finding{}, err
	}

	// The case belongs to the later of each pair, so a duplicate is flagged once,
	// whichever of the two is seen first.
	var result Finding
	for _, o := range others {
		later, earlier := t, o
		if o.FirstDebitAt.After(*t.FirstDebitAt) || (o.FirstDebitAt.Equal(*t.FirstDebitAt) && o.ID.String() > t.ID.String()) {
			later, earlier = o, t
		}
		if earlier.CreditStatus == txn.CreditFailed {
			continue // a retry after a confirmed failure
		}
		risk := "MEDIUM"
		if later.SessionID != "" && later.SessionID == earlier.SessionID {
			risk = "HIGH"
		}
		if earlier.CreditStatus == txn.CreditConfirmed && later.CreditStatus == txn.CreditConfirmed {
			risk = "HIGH"
		}
		laterEv := ev
		if later.ID != t.ID {
			if laterEv, err = d.evidence(ctx, tx, tenant, later.ID); err != nil {
				return Finding{}, err
			}
		}
		laterEv = append(laterEv, cases.Evidence{Kind: "related_transaction", Ref: earlier.Ref, IsFact: true,
			Detail: map[string]any{"role": "earlier_debit", "instruction_ref": earlier.Ref, "first_debit_at": earlier.FirstDebitAt.UTC().Format(time.RFC3339),
				"credit_status": earlier.CreditStatus, "window_seconds": window.Value}})
		f, err := d.open(ctx, tx, tenant, later, cases.CodeDuplicate, RuleDuplicate, risk, "duplicate_retry",
			fmt.Sprintf("Another debit with the same account, amount, currency and beneficiary (%s) was posted within %d seconds. Risk %s.", earlier.Ref, window.Value, risk), laterEv)
		if err != nil {
			return Finding{}, err
		}
		if f.Created || result.Case.ID == uuid.Nil {
			result = f
		}
	}
	return result, nil
}

// closeLateSuccess closes an open "unknown" case once the transaction is later
// confirmed successful, so nobody reverses a transaction that actually succeeded.
func (d *Detector) closeLateSuccess(ctx context.Context, tx pgx.Tx, tenant uuid.UUID, t txnRow) error {
	id := t.ID
	list, err := d.Cases.List(ctx, tx, tenant, cases.Filter{Classification: cases.CodeUnknown, TxnID: &id}, nil, 50)
	if err != nil {
		return err
	}
	for _, c := range list {
		if cases.IsTerminal(c.Status) || !cases.CanTransition(c.Status, cases.StClosedNoException) {
			continue
		}
		if _, err := d.Cases.Transition(ctx, tx, tenant, c.ID, cases.System, cases.StClosedNoException, 0, "LATE_SUCCESS_CONFIRMED"); err != nil {
			return err
		}
	}
	return nil
}

// Handle is the stream consumer for the "transactions" stream.
func (d *Detector) Handle(ctx context.Context, tx pgx.Tx, m stream.Message) error {
	var p struct {
		TxnID uuid.UUID `json:"txn_id"`
	}
	if err := json.Unmarshal(m.Payload, &p); err != nil || p.TxnID == uuid.Nil {
		return fmt.Errorf("detect: bad payload for event %s", m.EventID)
	}
	_, err := d.Evaluate(ctx, tx, m.TenantID, p.TxnID)
	return err
}

// SweepStuck opens an "unknown" case for transactions that are still pending
// after the rail's approved waiting threshold. A rail with no approved threshold
// is skipped: nothing is assumed.
func (d *Detector) SweepStuck(ctx context.Context, tx pgx.Tx, tenant uuid.UUID) ([]Finding, error) {
	rows, err := tx.Query(ctx, txnSelect+` WHERE tenant_id = $1 AND lifecycle_state IN ('PENDING', 'DEBITED') AND debit_status = 'POSTED'`, tenant)
	if err != nil {
		return nil, err
	}
	var stuck []txnRow
	for rows.Next() {
		t, err := scanTxn(rows)
		if err != nil {
			rows.Close()
			return nil, err
		}
		stuck = append(stuck, t)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, err
	}

	var out []Finding
	for _, t := range stuck {
		v, err := d.Config.Effective(ctx, tx, tenant, "waiting_threshold_seconds."+t.Rail, d.now())
		if errors.Is(err, configsvc.ErrNotConfigured) || errors.Is(err, configsvc.ErrUnknownKey) {
			continue
		}
		if err != nil {
			return nil, err
		}
		if d.now().Sub(t.LastEventAt) < time.Duration(v.Value)*time.Second {
			continue
		}
		ev, err := d.evidence(ctx, tx, tenant, t.ID)
		if err != nil {
			return nil, err
		}
		f, err := d.open(ctx, tx, tenant, t, cases.CodeUnknown, RuleStuck, "", "switch_response_failure",
			fmt.Sprintf("No final answer %d seconds after the last event, past the rail's waiting threshold of %d seconds.", int(d.now().Sub(t.LastEventAt).Seconds()), v.Value), ev)
		if err != nil {
			return nil, err
		}
		out = append(out, f)
	}
	return out, nil
}
