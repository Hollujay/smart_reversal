package audit

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
)

// AuditEvent represents an immutable compliance event recorded in the system.
type AuditEvent struct {
	ID         string                 `json:"id"`
	Timestamp  time.Time              `json:"timestamp"`
	ActorID    string                 `json:"actorId"`
	Action     string                 `json:"action"`
	EntityType string                 `json:"entityType"`
	EntityID   string                 `json:"entityId"`
	Details    map[string]interface{} `json:"details"`
	PrevHash   string                 `json:"prevHash"`
	Hash       string                 `json:"hash"`
}

type Service struct {
	mu       sync.RWMutex
	events   []AuditEvent
	lastHash string
}

func NewService() *Service {
	s := &Service{
		events:   make([]AuditEvent, 0),
		lastHash: "GENESIS_BLOCK_SMARTREVERSAL_LEDGER",
	}

	// Seed an initial genesis event
	s.Record("system-boot", "INITIALIZE_SYSTEM", "PLATFORM", "core", map[string]interface{}{
		"version": "0.1.0",
		"mode":    "maker_checker_enforced",
	})

	return s
}

// Record appends an immutable event to the audit trail with cryptographic hash chaining.
func (s *Service) Record(actorID, action, entityType, entityID string, details map[string]interface{}) AuditEvent {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now().UTC()
	detailsBytes, _ := json.Marshal(details)

	payload := fmt.Sprintf("%s|%s|%s|%s|%s|%s|%s",
		s.lastHash,
		now.Format(time.RFC3339Nano),
		actorID,
		action,
		entityType,
		entityID,
		string(detailsBytes),
	)

	hashBytes := sha256.Sum256([]byte(payload))
	hashStr := hex.EncodeToString(hashBytes[:])

	event := AuditEvent{
		ID:         uuid.NewString(),
		Timestamp:  now,
		ActorID:    actorID,
		Action:     action,
		EntityType: entityType,
		EntityID:   entityID,
		Details:    details,
		PrevHash:   s.lastHash,
		Hash:       hashStr,
	}

	s.lastHash = hashStr
	s.events = append(s.events, event)
	return event
}

// Query retrieves audit events filtered by entityId or actorId.
func (s *Service) Query(entityID, actorID string) []AuditEvent {
	s.mu.RLock()
	defer s.mu.RUnlock()

	results := make([]AuditEvent, 0)
	for i := len(s.events) - 1; i >= 0; i-- {
		ev := s.events[i]
		if entityID != "" && ev.EntityID != entityID {
			continue
		}
		if actorID != "" && ev.ActorID != actorID {
			continue
		}
		results = append(results, ev)
	}
	return results
}

// HandleGetAuditLogs handles GET /audit-logs
func (s *Service) HandleGetAuditLogs(w http.ResponseWriter, r *http.Request) {
	entityID := r.URL.Query().Get("entityId")
	actorID := r.URL.Query().Get("actorId")

	logs := s.Query(entityID, actorID)

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(logs)
}
