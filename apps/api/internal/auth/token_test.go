package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const testSecret = "a-test-secret-that-is-long-enough-0123456789"

func newTestService(t *testing.T) *AuthService {
	t.Helper()
	svc, err := NewAuthService(testSecret)
	if err != nil {
		t.Fatal(err)
	}
	return svc
}

func TestNewAuthServiceRejectsShortSecret(t *testing.T) {
	for _, secret := range []string{"", "short", "123456789012345678901234567890-"[:MinSecretLength-1]} {
		if _, err := NewAuthService(secret); err == nil {
			t.Fatalf("secret of length %d was accepted", len(secret))
		}
	}
	if _, err := NewAuthService(testSecret); err != nil {
		t.Fatalf("valid secret rejected: %v", err)
	}
}

func TestValidateTokenRoundTrip(t *testing.T) {
	svc := newTestService(t)
	tok, err := svc.GenerateToken("u1", "u1@example.com", RoleAuditor, time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	claims, err := svc.ValidateToken(tok)
	if err != nil {
		t.Fatalf("valid token rejected: %v", err)
	}
	if claims.UserID != "u1" || claims.Role != RoleAuditor {
		t.Fatalf("unexpected claims: %+v", claims)
	}
}

func signWith(t *testing.T, method jwt.SigningMethod, key interface{}, claims *Claims) string {
	t.Helper()
	tok, err := jwt.NewWithClaims(method, claims).SignedString(key)
	if err != nil {
		t.Fatal(err)
	}
	return tok
}

func validClaims() *Claims {
	return &Claims{
		UserID: "u1",
		Role:   RoleAnalystMaker,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    tokenIssuer,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute)),
		},
	}
}

func TestValidateTokenRejectsBadTokens(t *testing.T) {
	svc := newTestService(t)

	expired := validClaims()
	expired.ExpiresAt = jwt.NewNumericDate(time.Now().Add(-time.Minute))

	noExp := validClaims()
	noExp.ExpiresAt = nil

	wrongIssuer := validClaims()
	wrongIssuer.Issuer = "someone-else"

	noIssuer := validClaims()
	noIssuer.Issuer = ""

	cases := map[string]string{
		"expired":           signWith(t, jwt.SigningMethodHS256, []byte(testSecret), expired),
		"no expiry":         signWith(t, jwt.SigningMethodHS256, []byte(testSecret), noExp),
		"wrong issuer":      signWith(t, jwt.SigningMethodHS256, []byte(testSecret), wrongIssuer),
		"missing issuer":    signWith(t, jwt.SigningMethodHS256, []byte(testSecret), noIssuer),
		"wrong secret":      signWith(t, jwt.SigningMethodHS256, []byte("another-secret-that-is-long-enough-000000"), validClaims()),
		"other hmac method": signWith(t, jwt.SigningMethodHS512, []byte(testSecret), validClaims()),
		"alg none":          signWith(t, jwt.SigningMethodNone, jwt.UnsafeAllowNoneSignatureType, validClaims()),
		"garbage":           "not.a.jwt",
		"empty":             "",
	}

	for name, tok := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := svc.ValidateToken(tok); err == nil {
				t.Fatal("token was accepted")
			}
		})
	}
}

func TestAuthMiddlewareRejectsMissingOrMalformedHeader(t *testing.T) {
	svc := newTestService(t)
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("protected handler was reached")
	})

	for name, header := range map[string]string{
		"missing":      "",
		"wrong scheme": "Basic abc",
		"bad token":    "Bearer nope",
	} {
		t.Run(name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if header != "" {
				req.Header.Set("Authorization", header)
			}
			rec := httptest.NewRecorder()
			svc.AuthMiddleware(next).ServeHTTP(rec, req)
			if rec.Code != http.StatusUnauthorized {
				t.Fatalf("status = %d, want 401", rec.Code)
			}
		})
	}
}

func TestAuthMiddlewareAttachesClaims(t *testing.T) {
	svc := newTestService(t)
	tok, err := svc.GenerateToken("u1", "u1@example.com", RoleAuditor, time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	var got *Claims
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = GetUserClaims(r.Context())
	})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	svc.AuthMiddleware(next).ServeHTTP(httptest.NewRecorder(), req)
	if got == nil || got.UserID != "u1" {
		t.Fatalf("claims not attached: %+v", got)
	}
}
