package observability

import (
	"bytes"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestTraceMiddlewareGeneratesAndEchoes(t *testing.T) {
	var inCtx string
	h := TraceMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		inCtx = TraceID(r.Context())
	}))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	got := rec.Header().Get(TraceHeader)
	if got == "" || got != inCtx {
		t.Fatalf("header %q and context %q should match and be set", got, inCtx)
	}
}

func TestTraceMiddlewareKeepsSafeIncomingID(t *testing.T) {
	h := TraceMiddleware(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(TraceHeader, "abc-12345.def")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if got := rec.Header().Get(TraceHeader); got != "abc-12345.def" {
		t.Fatalf("trace id = %q", got)
	}
}

func TestTraceMiddlewareReplacesUnsafeIncomingID(t *testing.T) {
	for _, bad := range []string{"short", "has space in it 123", "line\nbreak12345", strings.Repeat("a", 65), "semi;colon12345"} {
		h := TraceMiddleware(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header["X-Trace-Id"] = []string{bad}
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		if got := rec.Header().Get(TraceHeader); got == bad || got == "" {
			t.Errorf("unsafe id %q was kept or empty result: %q", bad, got)
		}
	}
}

func TestRequestLoggerRecordsStatusAndOmitsQuery(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, nil))

	h := TraceMiddleware(RequestLogger(logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTeapot)
	})))
	req := httptest.NewRequest(http.MethodGet, "/v1/things?account=0123456789&token=abc", nil)
	req.Header.Set("Authorization", "Bearer secret-token-value")
	h.ServeHTTP(httptest.NewRecorder(), req)

	out := buf.String()
	for _, want := range []string{`"status":418`, `"path":"/v1/things"`, `"trace_id"`} {
		if !strings.Contains(out, want) {
			t.Errorf("log missing %s: %s", want, out)
		}
	}
	for _, leak := range []string{"0123456789", "secret-token-value", "token=abc"} {
		if strings.Contains(out, leak) {
			t.Errorf("log leaked %s: %s", leak, out)
		}
	}
}

func TestParseLevel(t *testing.T) {
	if parseLevel("debug") != slog.LevelDebug || parseLevel("WARN") != slog.LevelWarn ||
		parseLevel("error") != slog.LevelError || parseLevel("") != slog.LevelInfo {
		t.Fatal("unexpected level mapping")
	}
}
