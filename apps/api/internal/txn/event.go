// Package txn ingests transaction events and keeps each transaction's state as a
// projection of its events.
//
// Events are evidence: they are stored once, never changed, and deduplicated by
// the source's own reference plus a hash of the payload. The transaction row is
// recomputed from all of its events every time one arrives, using only which
// events exist and not the order they arrived in, so late, repeated and
// out-of-order events converge on one answer.
package txn

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"time"
)

// Event types. The names are illustrative and sources will differ in practice;
// real adapters map their own codes onto these (VERIFY with each rail's spec).
const (
	EvDebitPosted         = "DEBIT_POSTED"
	EvTransferSent        = "TRANSFER_SENT"
	EvCreditConfirmed     = "CREDIT_CONFIRMED"
	EvCreditFailed        = "CREDIT_FAILED"
	EvTransferTimeout     = "TRANSFER_RESPONSE_TIMEOUT"
	EvSettlementCompleted = "SETTLEMENT_COMPLETED"
	EvReversalPosted      = "REVERSAL_POSTED"
)

var eventTypes = map[string]bool{
	EvDebitPosted: true, EvTransferSent: true, EvCreditConfirmed: true, EvCreditFailed: true,
	EvTransferTimeout: true, EvSettlementCompleted: true, EvReversalPosted: true,
}

// KnownEventType reports whether t is a canonical event type.
func KnownEventType(t string) bool { return eventTypes[t] }

var channels = map[string]bool{
	"MOBILE": true, "INTERNET": true, "USSD": true, "POS": true, "ATM": true,
	"CARD": true, "GATEWAY": true, "MERCHANT": true, "API": true,
}

var (
	sourceSystemRe = regexp.MustCompile(`^[A-Z0-9_-]{2,40}$`)
	sourceRefRe    = regexp.MustCompile(`^[A-Za-z0-9._:-]{4,100}$`)
	instructionRe  = regexp.MustCompile(`^[A-Za-z0-9._:-]{4,64}$`)
	railRe         = regexp.MustCompile(`^[A-Z0-9_]{2,30}$`)
	tokenRe        = regexp.MustCompile(`^acct_[0-9a-f]{32}$`)
	bankRe         = regexp.MustCompile(`^[A-Za-z0-9 ._-]{0,60}$`)
	sessionRe      = regexp.MustCompile(`^[A-Za-z0-9._:-]{0,64}$`)
	currencyRe     = regexp.MustCompile(`^[A-Z]{3}$`)
)

// MaxAmountMinor bounds an amount well inside int64, so sums stay safe.
const MaxAmountMinor int64 = 9_000_000_000_000_000

// Amount is an integer number of minor units. In JSON it is a decimal string
// (D14), because many JSON readers lose precision on large integers.
type Amount int64

// UnmarshalJSON accepts only a quoted string of digits with no sign, no leading
// zeros and no fraction.
func (a *Amount) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return errors.New("amount_minor must be a string of digits")
	}
	n, err := ParseAmount(s)
	if err != nil {
		return err
	}
	*a = Amount(n)
	return nil
}

// String returns the amount as decimal digits.
func (a Amount) String() string { return strconv.FormatInt(int64(a), 10) }

// MarshalJSON writes the amount as a decimal string.
func (a Amount) MarshalJSON() ([]byte, error) { return json.Marshal(strconv.FormatInt(int64(a), 10)) }

// ParseAmount parses a decimal string of minor units.
func ParseAmount(s string) (int64, error) {
	if s == "" || len(s) > 16 {
		return 0, errors.New("amount_minor must be a string of 1 to 16 digits")
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return 0, errors.New("amount_minor must contain only digits")
		}
	}
	if len(s) > 1 && s[0] == '0' {
		return 0, errors.New("amount_minor must not have leading zeros")
	}
	n, err := strconv.ParseInt(s, 10, 64)
	if err != nil || n < 1 || n > MaxAmountMinor {
		return 0, fmt.Errorf("amount_minor must be between 1 and %d", MaxAmountMinor)
	}
	return n, nil
}

// Core holds the facts about a transaction that never change once the first
// event has set them.
type Core struct {
	InstructionRef    string `json:"instruction_ref"`
	Channel           string `json:"channel"`
	Rail              string `json:"rail"`
	AmountMinor       Amount `json:"amount_minor"`
	Currency          string `json:"currency"`
	DebitAccountToken string `json:"debit_account_token"`
	BeneficiaryBank   string `json:"beneficiary_bank,omitempty"`
	BeneficiaryToken  string `json:"beneficiary_token,omitempty"`
	SessionID         string `json:"session_id,omitempty"`
}

// Envelope is one event as a source system sends it.
type Envelope struct {
	SourceSystem   string    `json:"source_system"`
	SourceEventRef string    `json:"source_event_ref"`
	EventType      string    `json:"event_type"`
	SourceTS       time.Time `json:"source_ts"`
	Txn            Core      `json:"txn"`
	// Refs are external references that let evidence from other sources be matched
	// to this transaction. Keys are RefTypes.
	Refs map[string]string `json:"refs,omitempty"`
}

// RefTypes are the external reference kinds accepted.
var RefTypes = map[string]bool{"RRN": true, "STAN": true, "SESSION_ID": true, "LEDGER_ID": true, "GATEWAY_REF": true}

var refValueRe = regexp.MustCompile(`^[A-Za-z0-9._:-]{1,64}$`)

// ErrInvalid wraps every validation failure.
var ErrInvalid = errors.New("txn: invalid event")

func bad(format string, a ...any) error {
	return fmt.Errorf("%w: "+format, append([]any{ErrInvalid}, a...)...)
}

// Validate checks the envelope. now is used to refuse timestamps from the future.
func (e *Envelope) Validate(now time.Time) error {
	switch {
	case !sourceSystemRe.MatchString(e.SourceSystem):
		return bad("source_system must be 2 to 40 capitals, digits, dashes or underscores")
	case !sourceRefRe.MatchString(e.SourceEventRef):
		return bad("source_event_ref must be 4 to 100 characters from A-Z a-z 0-9 . _ : -")
	case !eventTypes[e.EventType]:
		return bad("event_type %q is not one of the known event types", e.EventType)
	case e.SourceTS.IsZero():
		return bad("source_ts is required")
	case e.SourceTS.After(now.Add(5 * time.Minute)):
		return bad("source_ts is in the future")
	case e.SourceTS.Before(now.AddDate(-2, 0, 0)):
		return bad("source_ts is more than two years old")
	case !instructionRe.MatchString(e.Txn.InstructionRef):
		return bad("txn.instruction_ref must be 4 to 64 characters from A-Z a-z 0-9 . _ : -")
	case !channels[e.Txn.Channel]:
		return bad("txn.channel %q is not a known channel", e.Txn.Channel)
	case !railRe.MatchString(e.Txn.Rail):
		return bad("txn.rail must be 2 to 30 capitals, digits or underscores")
	case e.Txn.AmountMinor < 1 || int64(e.Txn.AmountMinor) > MaxAmountMinor:
		return bad("txn.amount_minor is out of range")
	case !currencyRe.MatchString(e.Txn.Currency):
		return bad("txn.currency must be a three letter code")
	case !tokenRe.MatchString(e.Txn.DebitAccountToken):
		return bad("txn.debit_account_token must be an account token; raw account numbers are not accepted")
	case e.Txn.BeneficiaryToken != "" && !tokenRe.MatchString(e.Txn.BeneficiaryToken):
		return bad("txn.beneficiary_token must be an account token")
	case !bankRe.MatchString(e.Txn.BeneficiaryBank):
		return bad("txn.beneficiary_bank contains unsupported characters")
	case !sessionRe.MatchString(e.Txn.SessionID):
		return bad("txn.session_id contains unsupported characters")
	case len(e.Refs) > len(RefTypes):
		return bad("too many refs")
	}
	for k, v := range e.Refs {
		if !RefTypes[k] {
			return bad("refs.%s is not a known reference type", k)
		}
		if !refValueRe.MatchString(v) {
			return bad("refs.%s contains unsupported characters", k)
		}
	}
	return nil
}

// Hash returns a stable hash of the validated envelope. The same event sent
// twice has the same hash. The same reference with different content does not.
func (e *Envelope) Hash() string {
	c := *e
	c.SourceTS = c.SourceTS.UTC().Truncate(time.Microsecond)
	b, _ := json.Marshal(c)
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}
