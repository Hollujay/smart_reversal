package auth

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHandleLoginIsDisabled(t *testing.T) {
	svc := NewAuthService("a-test-secret-that-is-long-enough-0123456789")

	cases := []struct {
		name string
		body string
	}{
		{"known demo user", `{"email":"analyst@smartreversal.io","password":"wrong"}`},
		{"unknown user", `{"email":"nobody@example.com","password":""}`},
		{"empty body", ``},
		{"malformed body", `{`},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/v1/auth/login", strings.NewReader(tc.body))
			rec := httptest.NewRecorder()

			svc.HandleLogin(rec, req)

			if rec.Code != http.StatusNotImplemented {
				t.Fatalf("status = %d, want %d", rec.Code, http.StatusNotImplemented)
			}
			body := rec.Body.String()
			if strings.Contains(body, `"token"`) {
				t.Fatalf("response must not contain a token field: %s", body)
			}
			if !strings.Contains(body, "LOGIN_DISABLED") {
				t.Fatalf("response should state login is disabled: %s", body)
			}
		})
	}
}
