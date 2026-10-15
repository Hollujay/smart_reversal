package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func serveWithRole(t *testing.T, claims *Claims, allowed ...Role) int {
	t.Helper()
	called := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	if claims != nil {
		req = req.WithContext(context.WithValue(req.Context(), userContextKey{}, claims))
	}
	rec := httptest.NewRecorder()
	RequireRoles(allowed...)(next).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK && called {
		t.Fatalf("handler ran although status was %d", rec.Code)
	}
	return rec.Code
}

func TestRequireRolesAllowsListedRole(t *testing.T) {
	got := serveWithRole(t, &Claims{Role: RoleSupervisorChecker}, RoleSupervisorChecker)
	if got != http.StatusOK {
		t.Fatalf("status = %d, want 200", got)
	}
}

func TestRequireRolesRejectsUnlistedRole(t *testing.T) {
	got := serveWithRole(t, &Claims{Role: RoleAnalystMaker}, RoleSupervisorChecker)
	if got != http.StatusForbidden {
		t.Fatalf("status = %d, want 403", got)
	}
}

func TestRequireRolesAdminHasNoImplicitBypass(t *testing.T) {
	got := serveWithRole(t, &Claims{Role: RoleAdmin}, RoleSupervisorChecker)
	if got != http.StatusForbidden {
		t.Fatalf("admin passed a check it was not listed for: status = %d, want 403", got)
	}
}

func TestRequireRolesAdminAllowedWhenListed(t *testing.T) {
	got := serveWithRole(t, &Claims{Role: RoleAdmin}, RoleAdmin)
	if got != http.StatusOK {
		t.Fatalf("status = %d, want 200", got)
	}
}

func TestRequireRolesWithoutClaimsIsUnauthorized(t *testing.T) {
	got := serveWithRole(t, nil, RoleAdmin)
	if got != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", got)
	}
}
