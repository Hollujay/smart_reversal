package server

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// reversalSummary is a reversal as shown inside a case.
type reversalSummary struct {
	ID               string `json:"id"`
	AcceptedStatus   string `json:"accepted_status"`
	ProcessingStatus string `json:"processing_status"`
	FinancialStatus  string `json:"financial_status"`
}

// reversalsForCase lists a case's reversals. Filled in with the reversal workflow.
func (s *Services) reversalsForCase(r *http.Request, tx pgx.Tx, tenant, caseID uuid.UUID) ([]reversalSummary, error) {
	return []reversalSummary{}, nil
}
