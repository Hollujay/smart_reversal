package identity

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

// ErrInvalidCredentials is the only login failure callers see. It is returned
// for an unknown institution, an unknown user, a wrong password, a disabled
// account and a locked account alike, so a caller cannot learn which accounts
// exist.
var ErrInvalidCredentials = errors.New("identity: invalid credentials")

const (
	defaultMaxAttempts = 5
	defaultLockFor     = 15 * time.Minute
)

// LoginOutcome describes how a login attempt ended.
type LoginOutcome string

const (
	LoginSucceeded     LoginOutcome = "succeeded"
	LoginWrongPassword LoginOutcome = "wrong_password"
	LoginUnknownUser   LoginOutcome = "unknown_user"
	LoginAccountLocked LoginOutcome = "account_locked"
	LoginAccountOff    LoginOutcome = "account_disabled"
)

// LoginEvent is what an Observer is told about each attempt in a known
// institution. UserID is zero when the user does not exist. Attempts against an
// unknown institution have no tenant to record against and are not reported.
type LoginEvent struct {
	TenantID uuid.UUID
	UserID   uuid.UUID
	// EmailDigest is a short hash of the normalised email, so repeated attempts
	// can be correlated without storing what an attacker typed.
	EmailDigest string
	Outcome     LoginOutcome
}

// LoginObserver records login attempts. A failure on a successful login stops
// the login: an authentication that cannot be audited does not happen.
type LoginObserver interface {
	Observe(ctx context.Context, e LoginEvent) error
}

// Authenticator checks operator credentials and issues operator tokens.
type Authenticator struct {
	Store  UserStore
	Signer *Signer
	// Observer, when set, is told about every attempt in a known institution.
	Observer LoginObserver
	// Now defaults to time.Now.
	Now func() time.Time
	// MaxAttempts failed logins lock an account for LockFor. Defaults 5 and 15m.
	MaxAttempts int
	LockFor     time.Duration

	dummyOnce sync.Once
	dummyHash string
}

// LoginResult is what a successful login returns.
type LoginResult struct {
	Token     string
	ExpiresAt time.Time
	UserID    uuid.UUID
	TenantID  uuid.UUID
	Roles     []string
}

func (a *Authenticator) now() time.Time {
	if a.Now != nil {
		return a.Now()
	}
	return time.Now()
}

// Login verifies the credentials and returns a token. MFA is not issued here,
// so tokens from this path carry mfa=false and cannot perform actions that
// require it.
func (a *Authenticator) Login(ctx context.Context, institutionCode, email, password string) (*LoginResult, error) {
	maxAttempts, lockFor := a.MaxAttempts, a.LockFor
	if maxAttempts <= 0 {
		maxAttempts = defaultMaxAttempts
	}
	if lockFor <= 0 {
		lockFor = defaultLockFor
	}

	tenant, err := a.Store.TenantIDByCode(ctx, institutionCode)
	if errors.Is(err, ErrNotFound) {
		a.burn(password)
		return nil, ErrInvalidCredentials
	}
	if err != nil {
		return nil, err
	}

	digest := emailDigest(email)
	u, err := a.Store.FindByEmail(ctx, tenant, email)
	if errors.Is(err, ErrNotFound) {
		a.burn(password)
		a.note(ctx, LoginEvent{TenantID: tenant, EmailDigest: digest, Outcome: LoginUnknownUser})
		return nil, ErrInvalidCredentials
	}
	if err != nil {
		return nil, err
	}

	now := a.now()
	locked := u.LockedUntil != nil && u.LockedUntil.After(now)
	if u.Status != "active" || locked {
		// Spend the same time as a real check, and do not count the attempt:
		// counting would let an attacker extend someone else's lock.
		a.burn(password)
		outcome := LoginAccountLocked
		if u.Status != "active" {
			outcome = LoginAccountOff
		}
		a.note(ctx, LoginEvent{TenantID: tenant, UserID: u.ID, EmailDigest: digest, Outcome: outcome})
		return nil, ErrInvalidCredentials
	}

	ok, rehash, verr := VerifyPassword(password, u.PasswordHash)
	if verr != nil || !ok {
		if err := a.Store.RecordFailure(ctx, tenant, u.ID, now, maxAttempts, lockFor); err != nil {
			return nil, err
		}
		a.note(ctx, LoginEvent{TenantID: tenant, UserID: u.ID, EmailDigest: digest, Outcome: LoginWrongPassword})
		return nil, ErrInvalidCredentials
	}

	newHash := ""
	if rehash {
		if h, err := HashPassword(password); err == nil {
			newHash = h
		}
	}
	if err := a.Store.RecordSuccess(ctx, tenant, u.ID, newHash); err != nil {
		return nil, err
	}

	if a.Observer != nil {
		if err := a.Observer.Observe(ctx, LoginEvent{TenantID: tenant, UserID: u.ID, EmailDigest: digest, Outcome: LoginSucceeded}); err != nil {
			return nil, fmt.Errorf("identity: login not recorded: %w", err)
		}
	}
	tok, exp, err := a.Signer.Issue(Principal{UserID: u.ID, TenantID: tenant, Roles: u.Roles}, AudienceOps, now)
	if err != nil {
		return nil, err
	}
	return &LoginResult{Token: tok, ExpiresAt: exp, UserID: u.ID, TenantID: tenant, Roles: u.Roles}, nil
}

// note reports a failed attempt. A failure to record it must not turn into a
// different answer for the caller, so the error is dropped here; the observer
// is responsible for logging it.
func (a *Authenticator) note(ctx context.Context, e LoginEvent) {
	if a.Observer != nil {
		_ = a.Observer.Observe(ctx, e)
	}
}

func emailDigest(email string) string {
	sum := sha256.Sum256([]byte(NormalizeEmail(email)))
	return hex.EncodeToString(sum[:6])
}

// burn performs a password hash against a throwaway hash so that failing early
// takes about as long as failing late.
func (a *Authenticator) burn(password string) {
	a.dummyOnce.Do(func() { a.dummyHash, _ = HashPassword("not-a-real-password-0000") })
	_, _, _ = VerifyPassword(password, a.dummyHash)
}
