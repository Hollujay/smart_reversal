package identity

import (
	"context"
	"errors"
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

// Authenticator checks operator credentials and issues operator tokens.
type Authenticator struct {
	Store  UserStore
	Signer *Signer
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

	u, err := a.Store.FindByEmail(ctx, tenant, email)
	if errors.Is(err, ErrNotFound) {
		a.burn(password)
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
		return nil, ErrInvalidCredentials
	}

	ok, rehash, verr := VerifyPassword(password, u.PasswordHash)
	if verr != nil || !ok {
		if err := a.Store.RecordFailure(ctx, tenant, u.ID, now, maxAttempts, lockFor); err != nil {
			return nil, err
		}
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

	tok, exp, err := a.Signer.Issue(Principal{UserID: u.ID, TenantID: tenant, Roles: u.Roles}, AudienceOps, now)
	if err != nil {
		return nil, err
	}
	return &LoginResult{Token: tok, ExpiresAt: exp, UserID: u.ID, TenantID: tenant, Roles: u.Roles}, nil
}

// burn performs a password hash against a throwaway hash so that failing early
// takes about as long as failing late.
func (a *Authenticator) burn(password string) {
	a.dummyOnce.Do(func() { a.dummyHash, _ = HashPassword("not-a-real-password-0000") })
	_, _, _ = VerifyPassword(password, a.dummyHash)
}
