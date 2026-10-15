package exceptions

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

type Severity string

const (
	SeverityLow      Severity = "LOW"
	SeverityMedium   Severity = "MEDIUM"
	SeverityHigh     Severity = "HIGH"
	SeverityCritical Severity = "CRITICAL"
)

type Type string

const (
	TypeDebitWithoutCredit Type = "DEBIT_WITHOUT_CREDIT"
	TypeDuplicateDebit     Type = "DUPLICATE_DEBIT"
	TypeStatusMismatch     Type = "STATUS_MISMATCH"
	TypeSuspenseBalance    Type = "SUSPENSE_BALANCE"
	TypeTimeoutException   Type = "TIMEOUT_EXCEPTION"
)

type FinancialException struct {
	ID                   string    `json:"id"`
	TransactionReference string    `json:"transactionReference"`
	Type                 Type      `json:"type"`
	Severity             Severity  `json:"severity"`
	DiscrepancyAmount    string    `json:"discrepancyAmount"`
	Status               string    `json:"status"` // DETECTED, UNDER_INVESTIGATION, RESOLVED
	DetectedAt           time.Time `json:"detectedAt"`
	Description          string    `json:"description"`
}

type Store struct {
	mu         sync.RWMutex
	exceptions map[string]FinancialException
	order      []string
}

func NewStore() *Store {
	s := &Store{
		exceptions: make(map[string]FinancialException),
		order:      make([]string, 0),
	}
	s.seedDefaultExceptions()
	return s
}

func (s *Store) seedDefaultExceptions() {
	now := time.Now().UTC()
	defaults := []FinancialException{
		{
			ID:                   "exc-001-nip-9901",
			TransactionReference: "TXN-20261015-NIP-9901",
			Type:                 TypeDebitWithoutCredit,
			Severity:             SeverityHigh,
			DiscrepancyAmount:    "150000.00",
			Status:               "DETECTED",
			DetectedAt:           now.Add(-20 * time.Minute),
			Description:          "Customer account debited; beneficiary switch reported response timeout without ACK.",
		},
		{
			ID:                   "exc-002-pos-4402",
			TransactionReference: "TXN-20261015-POS-4402",
			Type:                 TypeDuplicateDebit,
			Severity:             SeverityCritical,
			DiscrepancyAmount:    "45000.00",
			Status:               "UNDER_INVESTIGATION",
			DetectedAt:           now.Add(-40 * time.Minute),
			Description:          "Duplicate authorization hold posted on terminal retry within 1.4 seconds.",
		},
		{
			ID:                   "exc-003-nip-3320",
			TransactionReference: "TXN-20261015-NIP-3320",
			Type:                 TypeSuspenseBalance,
			Severity:             SeverityHigh,
			DiscrepancyAmount:    "500000.00",
			Status:               "DETECTED",
			DetectedAt:           now.Add(-8 * time.Minute),
			Description:          "Settlement pool ledger variance exceeds threshold in suspense account #99201.",
		},
	}

	for _, e := range defaults {
		s.exceptions[e.ID] = e
		s.order = append(s.order, e.ID)
	}
}

func (s *Store) List(severity, excType string) []FinancialException {
	s.mu.RLock()
	defer s.mu.RUnlock()

	results := make([]FinancialException, 0)
	for i := len(s.order) - 1; i >= 0; i-- {
		e := s.exceptions[s.order[i]]
		if severity != "" && strings.ToUpper(string(e.Severity)) != strings.ToUpper(severity) {
			continue
		}
		if excType != "" && strings.ToUpper(string(e.Type)) != strings.ToUpper(excType) {
			continue
		}
		results = append(results, e)
	}
	return results
}

func (s *Store) GetByID(id string) (FinancialException, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	e, ok := s.exceptions[id]
	return e, ok
}

func (s *Store) UpdateStatus(id, status string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	e, ok := s.exceptions[id]
	if !ok {
		return false
	}
	e.Status = status
	s.exceptions[id] = e
	return true
}

func (s *Store) Add(ref string, excType Type, severity Severity, amount, desc string) FinancialException {
	s.mu.Lock()
	defer s.mu.Unlock()

	id := "exc-" + uuid.NewString()[:8]
	e := FinancialException{
		ID:                   id,
		TransactionReference: ref,
		Type:                 excType,
		Severity:             severity,
		DiscrepancyAmount:    amount,
		Status:               "DETECTED",
		DetectedAt:           time.Now().UTC(),
		Description:          desc,
	}
	s.exceptions[id] = e
	s.order = append(s.order, id)
	return e
}

func (s *Store) HandleListExceptions(w http.ResponseWriter, r *http.Request) {
	severity := r.URL.Query().Get("severity")
	excType := r.URL.Query().Get("type")

	items := s.List(severity, excType)

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(items)
}
