package transactions

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

type Channel string

const (
	ChannelNIP         Channel = "NIP"
	ChannelPOS         Channel = "POS"
	ChannelCard        Channel = "CARD"
	ChannelWebhook     Channel = "WEBHOOK"
	ChannelDirectDebit Channel = "DIRECT_DEBIT"
)

type Status string

const (
	StatusSuccess   Status = "SUCCESS"
	StatusPending   Status = "PENDING"
	StatusFailed    Status = "FAILED"
	StatusUncertain Status = "UNCERTAIN"
	StatusReversed  Status = "REVERSED"
)

// Transaction represents an ingested financial transaction record.
type Transaction struct {
	ID                 string    `json:"id"`
	Reference          string    `json:"reference"`
	Amount             string    `json:"amount"` // fixed-precision NUMERIC string
	Currency           string    `json:"currency"`
	Channel            Channel   `json:"channel"`
	Status             Status    `json:"status"`
	SenderAccount      string    `json:"senderAccount"`
	BeneficiaryAccount string    `json:"beneficiaryAccount"`
	CreatedAt          time.Time `json:"createdAt"`
	Narrative          string    `json:"narrative,omitempty"`
}

type Store struct {
	mu           sync.RWMutex
	transactions map[string]Transaction
	order        []string
}

func NewStore() *Store {
	s := &Store{
		transactions: make(map[string]Transaction),
		order:        make([]string, 0),
	}
	s.seedDefaultData()
	return s
}

func (s *Store) seedDefaultData() {
	now := time.Now().UTC()
	defaults := []Transaction{
		{
			ID:                 uuid.NewString(),
			Reference:          "TXN-20261015-NIP-9901",
			Amount:             "150000.00",
			Currency:           "NGN",
			Channel:            ChannelNIP,
			Status:             StatusUncertain,
			SenderAccount:      "0123456789",
			BeneficiaryAccount: "9876543210",
			CreatedAt:          now.Add(-25 * time.Minute),
			Narrative:          "Interbank instant transfer timeout at switch gateway",
		},
		{
			ID:                 uuid.NewString(),
			Reference:          "TXN-20261015-POS-4402",
			Amount:             "45000.00",
			Currency:           "NGN",
			Channel:            ChannelPOS,
			Status:             StatusFailed,
			SenderAccount:      "1122334455",
			BeneficiaryAccount: "5544332211",
			CreatedAt:          now.Add(-48 * time.Minute),
			Narrative:          "Terminal debit occurred without switch settlement confirmation",
		},
		{
			ID:                 uuid.NewString(),
			Reference:          "TXN-20261015-CRD-1109",
			Amount:             "285000.50",
			Currency:           "NGN",
			Channel:            ChannelCard,
			Status:             StatusSuccess,
			SenderAccount:      "2233445566",
			BeneficiaryAccount: "6655443322",
			CreatedAt:          now.Add(-1 * time.Hour),
			Narrative:          "E-commerce settlement confirmed",
		},
		{
			ID:                 uuid.NewString(),
			Reference:          "TXN-20261015-NIP-3320",
			Amount:             "500000.00",
			Currency:           "NGN",
			Channel:            ChannelNIP,
			Status:             StatusPending,
			SenderAccount:      "3344556677",
			BeneficiaryAccount: "7766554433",
			CreatedAt:          now.Add(-10 * time.Minute),
			Narrative:          "High-value corporate outward transfer in suspense queue",
		},
		{
			ID:                 uuid.NewString(),
			Reference:          "TXN-20261015-WBH-8821",
			Amount:             "82000.00",
			Currency:           "NGN",
			Channel:            ChannelWebhook,
			Status:             StatusReversed,
			SenderAccount:      "4455667788",
			BeneficiaryAccount: "8877665544",
			CreatedAt:          now.Add(-2 * time.Hour),
			Narrative:          "Duplicate charge reversed after dual authorization",
		},
	}

	for _, t := range defaults {
		s.transactions[t.ID] = t
		s.order = append(s.order, t.ID)
	}
}

func (s *Store) List(ref, channel, status string, limit int) []Transaction {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if limit <= 0 || limit > 100 {
		limit = 50
	}

	results := make([]Transaction, 0)
	for i := len(s.order) - 1; i >= 0; i-- {
		t := s.transactions[s.order[i]]
		if ref != "" && !strings.Contains(strings.ToUpper(t.Reference), strings.ToUpper(ref)) {
			continue
		}
		if channel != "" && strings.ToUpper(string(t.Channel)) != strings.ToUpper(channel) {
			continue
		}
		if status != "" && strings.ToUpper(string(t.Status)) != strings.ToUpper(status) {
			continue
		}
		results = append(results, t)
		if len(results) >= limit {
			break
		}
	}
	return results
}

func (s *Store) GetByID(id string) (Transaction, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	t, ok := s.transactions[id]
	return t, ok
}

func (s *Store) UpdateStatus(id string, status Status) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	t, ok := s.transactions[id]
	if !ok {
		return false
	}
	t.Status = status
	s.transactions[id] = t
	return true
}

func (s *Store) HandleQueryTransactions(w http.ResponseWriter, r *http.Request) {
	ref := r.URL.Query().Get("reference")
	channel := r.URL.Query().Get("channel")
	status := r.URL.Query().Get("status")
	limitStr := r.URL.Query().Get("limit")

	limit := 50
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	results := s.List(ref, channel, status, limit)

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(results)
}
