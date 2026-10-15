package reversal

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/smartreversal/smartreversal/apps/api/internal/cases"
)

// AutoResult reports what automation did with one case.
type AutoResult struct {
	CaseRef  string
	Outcome  string // "reversed", "awaiting_approval", "held", "in_progress", "existing", "failed"
	Reversal *Reversal
	Gate     Gate
}

// AutoReverse is the automatic path for one case. It asks for a reversal as the
// system, which passes through exactly the same eight checks as a person's
// request. If policy lets automation proceed it executes and verifies; if not,
// it leaves the reversal waiting for a person's approval; if the gate fails it
// does nothing at all.
func (s *Service) AutoReverse(ctx context.Context, tenant, caseID uuid.UUID) (AutoResult, error) {
	var ref string
	err := s.Pool.WithTenant(ctx, tenant, func(tx pgx.Tx) error {
		c, err := s.Cases.Get(ctx, tx, tenant, caseID)
		ref = c.Ref
		return err
	})
	if err != nil {
		return AutoResult{}, err
	}
	res := AutoResult{CaseRef: ref}

	out, err := s.Request(ctx, tenant, RequestInput{CaseID: caseID, IdempotencyKey: "auto-" + ref + "-v1", ReasonCode: "AUTOMATIC_ASSESSMENT", Who: Requester{}})
	if err != nil {
		return res, err
	}
	res.Gate = out.Gate
	if out.Blocked {
		res.Outcome = "held"
		return res, nil
	}
	rev := out.Reversal
	res.Reversal = &rev
	if out.Existing {
		res.Outcome = "existing"
	}
	if rev.ApprovalStatus == "PENDING" {
		res.Outcome = "awaiting_approval"
		return res, nil
	}
	if rev.ApprovalStatus == "REJECTED" {
		res.Outcome = "held"
		return res, nil
	}

	if rev, err = s.Execute(ctx, tenant, rev.ID); err != nil {
		return res, err
	}
	if rev, err = s.Advance(ctx, tenant, rev.ID); err != nil {
		return res, err
	}
	res.Reversal = &rev
	switch {
	case rev.FinancialStatus == "VERIFIED":
		res.Outcome = "reversed"
	case rev.FinancialStatus == "FAILED":
		res.Outcome = "failed"
	default:
		res.Outcome = "in_progress"
	}
	return res, nil
}

// Sweep runs AutoReverse for every open detected case of a reversible type that
// has no reversal yet, and advances reversals that are still in flight. It is the
// scheduled job behind the automation (FR-38), and goes through the same gate as
// any other request.
func (s *Service) Sweep(ctx context.Context, tenant uuid.UUID) ([]AutoResult, error) {
	var candidates, inflight []uuid.UUID
	err := s.Pool.WithTenant(ctx, tenant, func(tx pgx.Tx) error {
		rows, err := tx.Query(ctx, `SELECT c.id FROM cases c
			WHERE c.tenant_id = $1 AND c.source = 'bot' AND c.classification IN ('FAIL001', 'DUP001') AND c.status IN ('DETECTED', 'UNDER_INVESTIGATION')
			  AND NOT EXISTS (SELECT 1 FROM reversals r WHERE r.tenant_id = c.tenant_id AND r.case_id = c.id)
			ORDER BY c.opened_at`, tenant)
		if err != nil {
			return err
		}
		for rows.Next() {
			var id uuid.UUID
			if err := rows.Scan(&id); err != nil {
				rows.Close()
				return err
			}
			candidates = append(candidates, id)
		}
		rows.Close()
		if err := rows.Err(); err != nil {
			return err
		}
		rows, err = tx.Query(ctx, `SELECT id FROM reversals WHERE tenant_id = $1 AND financial_status = 'UNVERIFIED' AND accepted_status IN ('ACCEPTED', 'UNKNOWN')
			AND approval_status IN ('NOT_REQUIRED', 'APPROVED') ORDER BY created_at`, tenant)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var id uuid.UUID
			if err := rows.Scan(&id); err != nil {
				return err
			}
			inflight = append(inflight, id)
		}
		return rows.Err()
	})
	if err != nil {
		return nil, err
	}

	var out []AutoResult
	for _, id := range candidates {
		r, err := s.AutoReverse(ctx, tenant, id)
		if err != nil {
			return out, fmt.Errorf("reversal: automatic path for case %s: %w", id, err)
		}
		out = append(out, r)
	}
	for _, id := range inflight {
		if _, err := s.Advance(ctx, tenant, id); err != nil {
			return out, fmt.Errorf("reversal: advance %s: %w", id, err)
		}
	}
	return out, nil
}

var _ = cases.CodeNoCredit
