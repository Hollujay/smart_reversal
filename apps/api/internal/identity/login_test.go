package identity

import (
	"context"
	"crypto/ed25519"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/smartreversal/smartreversal/apps/api/internal/platform/db"
	"github.com/smartreversal/smartreversal/apps/api/internal/platform/dbtest"
)

const goodPassword = "a long enough passphrase"

type loginRig struct {
	pool   *db.Pool
	store  *PGUserStore
	auth   *Authenticator
	tenant uuid.UUID
	code   string
	now    time.Time
	verify *Verifier
}

func newLoginRig(t *testing.T) *loginRig {
	t.Helper()
	pool := dbtest.Pool(t)
	r := &loginRig{pool: pool, store: &PGUserStore{Pool: pool}, tenant: dbtest.Tenant(t, pool), now: t0}
	if err := pool.QueryRow(context.Background(), `SELECT code FROM institutions WHERE id = $1`, r.tenant).Scan(&r.code); err != nil {
		t.Fatal(err)
	}
	priv, _, _ := GenerateKey()
	signer, err := NewSigner("k1", priv, 10*time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	ks, _ := ParseKeySet("", map[string]ed25519.PublicKey{"k1": signer.PublicKey()})
	r.verify = NewVerifier(ks, func() time.Time { return r.now })
	r.auth = &Authenticator{Store: r.store, Signer: signer, Now: func() time.Time { return r.now }, MaxAttempts: 3, LockFor: 10 * time.Minute}
	return r
}

func (r *loginRig) addUser(t *testing.T, email string, roles ...string) uuid.UUID {
	t.Helper()
	id, err := r.store.Create(context.Background(), r.tenant, email, "Test Person", goodPassword, roles)
	if err != nil {
		t.Fatal(err)
	}
	return id
}

func (r *loginRig) state(t *testing.T, id uuid.UUID) (attempts int, locked *time.Time) {
	t.Helper()
	if err := r.pool.QueryRow(context.Background(), `SELECT failed_attempts, locked_until FROM users WHERE id = $1`, id).Scan(&attempts, &locked); err != nil {
		t.Fatal(err)
	}
	return
}

func TestLoginSuccessIssuesAVerifiableOperatorToken(t *testing.T) {
	r := newLoginRig(t)
	id := r.addUser(t, "Officer@Example.Test", "REVERSAL_OFFICER", "AUDITOR")

	res, err := r.auth.Login(context.Background(), r.code, "  officer@EXAMPLE.test ", goodPassword)
	if err != nil {
		t.Fatal(err)
	}
	p, err := r.verify.Verify(res.Token, AudienceOps)
	if err != nil {
		t.Fatalf("issued token does not verify: %v", err)
	}
	if p.UserID != id || p.TenantID != r.tenant || !p.HasRole("REVERSAL_OFFICER") || !p.HasRole("AUDITOR") || p.MFA {
		t.Fatalf("unexpected principal: %+v", p)
	}
	if _, err := r.verify.Verify(res.Token, AudienceCustomer); err == nil {
		t.Fatal("operator login token accepted as a customer token")
	}
}

func TestEveryFailureLooksTheSame(t *testing.T) {
	r := newLoginRig(t)
	r.addUser(t, "real@example.test", "AUDITOR")
	disabled := r.addUser(t, "off@example.test", "AUDITOR")
	if _, err := r.pool.Exec(context.Background(), `UPDATE users SET status = 'disabled' WHERE id = $1`, disabled); err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()

	cases := map[string]struct{ code, email, pw string }{
		"wrong password":      {r.code, "real@example.test", "wrong password here"},
		"unknown user":        {r.code, "nobody@example.test", goodPassword},
		"unknown institution": {"NO-SUCH-BANK", "real@example.test", goodPassword},
		"disabled user":       {r.code, "off@example.test", goodPassword},
		"empty everything":    {"", "", ""},
		"short password":      {r.code, "real@example.test", "x"},
	}
	for name, c := range cases {
		res, err := r.auth.Login(ctx, c.code, c.email, c.pw)
		if res != nil || err != ErrInvalidCredentials {
			t.Errorf("%s: res=%v err=%v; want ErrInvalidCredentials", name, res, err)
		}
	}
}

func TestLockoutAfterRepeatedFailuresAndRecovery(t *testing.T) {
	r := newLoginRig(t)
	id := r.addUser(t, "lock@example.test", "AUDITOR")
	ctx := context.Background()

	for i := 0; i < 3; i++ {
		if _, err := r.auth.Login(ctx, r.code, "lock@example.test", "wrong password here"); err != ErrInvalidCredentials {
			t.Fatal(err)
		}
	}
	attempts, locked := r.state(t, id)
	if attempts != 3 || locked == nil || !locked.After(r.now) {
		t.Fatalf("attempts=%d locked=%v; want 3 and a future lock", attempts, locked)
	}

	// While locked even the right password is refused, and the lock is not extended.
	r.now = r.now.Add(5 * time.Minute)
	if _, err := r.auth.Login(ctx, r.code, "lock@example.test", goodPassword); err != ErrInvalidCredentials {
		t.Fatalf("locked account accepted the right password: %v", err)
	}
	if _, err := r.auth.Login(ctx, r.code, "lock@example.test", "wrong password here"); err != ErrInvalidCredentials {
		t.Fatal(err)
	}
	attempts2, locked2 := r.state(t, id)
	if attempts2 != 3 || !locked2.Equal(*locked) {
		t.Fatalf("attempts during a lock changed state: attempts=%d lock=%v (was %v)", attempts2, locked2, locked)
	}

	// After the lock expires the right password works and clears the counters.
	r.now = r.now.Add(10 * time.Minute)
	if _, err := r.auth.Login(ctx, r.code, "lock@example.test", goodPassword); err != nil {
		t.Fatalf("login after lock expiry: %v", err)
	}
	attempts3, locked3 := r.state(t, id)
	if attempts3 != 0 || locked3 != nil {
		t.Fatalf("success did not reset: attempts=%d lock=%v", attempts3, locked3)
	}
}

func TestSuccessResetsPartialFailures(t *testing.T) {
	r := newLoginRig(t)
	id := r.addUser(t, "reset@example.test", "AUDITOR")
	ctx := context.Background()
	for i := 0; i < 2; i++ {
		_, _ = r.auth.Login(ctx, r.code, "reset@example.test", "wrong password here")
	}
	if a, _ := r.state(t, id); a != 2 {
		t.Fatalf("attempts=%d", a)
	}
	if _, err := r.auth.Login(ctx, r.code, "reset@example.test", goodPassword); err != nil {
		t.Fatal(err)
	}
	if a, l := r.state(t, id); a != 0 || l != nil {
		t.Fatalf("attempts=%d lock=%v", a, l)
	}
}

func TestConcurrentFailuresAreAllCounted(t *testing.T) {
	r := newLoginRig(t)
	r.auth.MaxAttempts = 1000
	id := r.addUser(t, "race@example.test", "AUDITOR")
	var wg sync.WaitGroup
	for i := 0; i < 12; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _ = r.auth.Login(context.Background(), r.code, "race@example.test", "wrong password here")
		}()
	}
	wg.Wait()
	if a, _ := r.state(t, id); a != 12 {
		t.Fatalf("attempts=%d, want 12", a)
	}
}

func TestAUserCannotLogInThroughAnotherInstitution(t *testing.T) {
	r := newLoginRig(t)
	r.addUser(t, "only-in-a@example.test", "AUDITOR")
	other := dbtest.Tenant(t, r.pool)
	var otherCode string
	if err := r.pool.QueryRow(context.Background(), `SELECT code FROM institutions WHERE id = $1`, other).Scan(&otherCode); err != nil {
		t.Fatal(err)
	}
	if _, err := r.auth.Login(context.Background(), otherCode, "only-in-a@example.test", goodPassword); err != ErrInvalidCredentials {
		t.Fatalf("cross-institution login: %v", err)
	}
}

func TestOldPasswordHashIsUpgradedOnLogin(t *testing.T) {
	r := newLoginRig(t)
	id := r.addUser(t, "old@example.test", "AUDITOR")
	weak := encodeHash(8*1024, 1, 1, []byte("0123456789abcdef"), argon2Key(goodPassword, []byte("0123456789abcdef"), 1, 8*1024, 1))
	if _, err := r.pool.Exec(context.Background(), `UPDATE users SET password_hash = $2 WHERE id = $1`, id, weak); err != nil {
		t.Fatal(err)
	}
	if _, err := r.auth.Login(context.Background(), r.code, "old@example.test", goodPassword); err != nil {
		t.Fatal(err)
	}
	var h string
	_ = r.pool.QueryRow(context.Background(), `SELECT password_hash FROM users WHERE id = $1`, id).Scan(&h)
	if h == weak {
		t.Fatal("weak hash was not replaced")
	}
	if ok, rehash, err := VerifyPassword(goodPassword, h); !ok || rehash || err != nil {
		t.Fatalf("new hash: ok=%v rehash=%v err=%v", ok, rehash, err)
	}
}

func TestCreateValidates(t *testing.T) {
	r := newLoginRig(t)
	ctx := context.Background()
	if _, err := r.store.Create(ctx, r.tenant, "a@example.test", "A", "short", []string{"AUDITOR"}); !errors.Is(err, ErrPasswordPolicy) {
		t.Errorf("short password: %v", err)
	}
	if _, err := r.store.Create(ctx, r.tenant, "a@example.test", "A", goodPassword, []string{"bad role"}); err == nil {
		t.Error("bad role accepted")
	}
	r.addUser(t, "dup@example.test", "AUDITOR")
	if _, err := r.store.Create(ctx, r.tenant, "DUP@example.test", "A", goodPassword, nil); err == nil {
		t.Error("duplicate email (different case) accepted")
	}
}

func TestStoreNotFound(t *testing.T) {
	r := newLoginRig(t)
	if _, err := r.store.TenantIDByCode(context.Background(), "NOPE-NOPE"); !errors.Is(err, ErrNotFound) {
		t.Errorf("institution: %v", err)
	}
	if _, err := r.store.FindByEmail(context.Background(), r.tenant, "x@example.test"); !errors.Is(err, ErrNotFound) {
		t.Errorf("user: %v", err)
	}
}

func TestBootstrapCreatesInstitutionAndOperatorThatCanLogIn(t *testing.T) {
	r := newLoginRig(t)
	ctx := context.Background()
	code := "BOOT-" + strings.ToUpper(uuid.NewString()[:8])
	t.Cleanup(func() {
		_, _ = r.pool.Exec(context.Background(), `DELETE FROM users WHERE tenant_id IN (SELECT id FROM institutions WHERE code = $1)`, code)
		_, _ = r.pool.Exec(context.Background(), `DELETE FROM institutions WHERE code = $1`, code)
	})

	tenant, user, err := Bootstrap(ctx, r.pool, BootstrapSpec{
		InstitutionCode: code, Email: "first@example.test", Password: goodPassword, Roles: []string{"ADMIN"},
	})
	if err != nil || tenant == uuid.Nil || user == uuid.Nil {
		t.Fatalf("bootstrap: %v %v %v", tenant, user, err)
	}
	res, err := r.auth.Login(ctx, code, "first@example.test", goodPassword)
	if err != nil || !contains(res.Roles, "ADMIN") {
		t.Fatalf("bootstrapped operator cannot log in: %v %+v", err, res)
	}

	// A second operator in the same institution reuses it.
	tenant2, _, err := Bootstrap(ctx, r.pool, BootstrapSpec{
		InstitutionCode: code, Email: "second@example.test", Password: goodPassword, Roles: []string{"AUDITOR"},
	})
	if err != nil || tenant2 != tenant {
		t.Fatalf("second bootstrap: tenant %v (want %v) err %v", tenant2, tenant, err)
	}
	if _, _, err := Bootstrap(ctx, r.pool, BootstrapSpec{InstitutionCode: code, Email: "first@example.test", Password: goodPassword}); err == nil {
		t.Error("duplicate email accepted")
	}
	for name, spec := range map[string]BootstrapSpec{
		"no code":     {Email: "a@example.test", Password: goodPassword},
		"no email":    {InstitutionCode: code, Password: goodPassword},
		"no password": {InstitutionCode: code, Email: "a@example.test"},
		"bad code":    {InstitutionCode: "lower case", Email: "a@example.test", Password: goodPassword},
	} {
		if _, _, err := Bootstrap(ctx, r.pool, spec); err == nil {
			t.Errorf("%s accepted", name)
		}
	}
}

func contains(list []string, want string) bool {
	for _, s := range list {
		if s == want {
			return true
		}
	}
	return false
}
