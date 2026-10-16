package server

import (
	"context"
	"crypto/ed25519"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/smartreversal/smartreversal/apps/api/internal/alertparse"
	"github.com/smartreversal/smartreversal/apps/api/internal/audit"
	"github.com/smartreversal/smartreversal/apps/api/internal/authz"
	"github.com/smartreversal/smartreversal/apps/api/internal/cases"
	"github.com/smartreversal/smartreversal/apps/api/internal/configsvc"
	"github.com/smartreversal/smartreversal/apps/api/internal/creport"
	"github.com/smartreversal/smartreversal/apps/api/internal/customer"
	"github.com/smartreversal/smartreversal/apps/api/internal/deadline"
	"github.com/smartreversal/smartreversal/apps/api/internal/identity"
	"github.com/smartreversal/smartreversal/apps/api/internal/platform/dbtest"
	"github.com/smartreversal/smartreversal/apps/api/internal/txn"
)

type reportRig struct {
	*custRig
	ing          *txn.Service
	alice, bob   customer.Customer
	aliceT, bobT string
}

func newReportRig(t *testing.T) *reportRig {
	t.Helper()
	pool := dbtest.Pool(t)
	policy, err := authz.Load(context.Background(), pool)
	if err != nil {
		t.Fatal(err)
	}
	priv, _, _ := identity.GenerateKey()
	signer, _ := identity.NewSigner("k1", priv, 10*time.Minute)
	ks, _ := identity.ParseKeySet("", map[string]ed25519.PublicKey{"k1": signer.PublicKey()})
	app := &audit.Appender{}
	cfg := &configsvc.Service{Audit: app}
	logger, _ := quietLogger()
	cs := &cases.Service{Audit: app, Config: cfg}
	svc := &Services{Pool: pool, Policy: policy, Audit: app, Config: cfg, Txn: &txn.Service{}, Cases: cs,
		CustomerLogin: &customer.Login{Provider: &customer.DevProvider{Pool: pool, Audit: app}, Signer: signer},
		Reports:       &creport.Service{Pool: pool, Config: cfg, Cases: cs, Deadlines: &deadline.Engine{Config: cfg}, Parser: &alertparse.RuleParser{OwnNames: []string{"Wema"}}}}
	cr := &custRig{store: &customer.Store{Pool: pool}, auditRig: &auditRig{pool: pool, signer: signer, tenant: dbtest.Tenant(t, pool), appender: app,
		handler: NewRouter(Options{Logger: logger, Verifier: identity.NewVerifier(ks, nil), Services: svc})}}
	if err := pool.QueryRow(context.Background(), `SELECT code FROM institutions WHERE id = $1`, cr.tenant).Scan(&cr.code); err != nil {
		t.Fatal(err)
	}
	r := &reportRig{custRig: cr, ing: &txn.Service{}}
	r.alice, _ = cr.store.Create(context.Background(), cr.tenant, "Alice", customerPass)
	r.bob, _ = cr.store.Create(context.Background(), cr.tenant, "Bob", customerPass)
	_ = cr.store.LinkAccount(context.Background(), cr.tenant, r.alice.ID, tokA, "6789", "bank onboarding")
	_ = cr.store.LinkAccount(context.Background(), cr.tenant, r.bob.ID, tokB, "4321", "bank onboarding")
	_, r.aliceT = cr.login(t, r.alice.Ref, customerPass)
	_, r.bobT = cr.login(t, r.bob.Ref, customerPass)
	return r
}

func (r *reportRig) post(path, body, tok, key string) *recorder {
	req := newReq(http.MethodPost, path, body, tok)
	if key != "" {
		req.Header.Set("Idempotency-Key", key)
	}
	return serve(r.handler, req)
}

func (r *reportRig) transfer(t *testing.T, token, ref string, amount int64, at time.Time, types ...string) {
	t.Helper()
	for i, typ := range types {
		err := r.pool.WithTenant(context.Background(), r.tenant, func(tx pgx.Tx) error {
			_, err := r.ing.Ingest(context.Background(), tx, r.tenant, txn.Envelope{SourceSystem: "SWITCH_A", SourceEventRef: fmt.Sprintf("EV-%s-%d", ref, i), EventType: typ,
				SourceTS: at.Add(time.Duration(i) * time.Second), Txn: txn.Core{InstructionRef: ref, Channel: "MOBILE", Rail: "NIP", AmountMinor: txn.Amount(amount), Currency: "NGN", DebitAccountToken: token}})
			return err
		})
		if err != nil {
			t.Fatal(err)
		}
	}
}

const reportAlert = "Wema Alert: Acct 0123456789 debited NGN 18,500.00 on 2026-10-14 09:12. Desc: NIP TRANSFER | Ref: NIP202610140912"

func TestACustomerCreatesAReportAndConfirmsItOverHTTPAndNobodyElseCanSeeIt(t *testing.T) {
	r := newReportRig(t)
	at := time.Date(2026, 10, 14, 8, 12, 0, 0, time.UTC)
	r.transfer(t, tokA, "NIP202610140912", 1850000, at, "DEBIT_POSTED", "TRANSFER_SENT", "CREDIT_CONFIRMED")

	body, _ := json.Marshal(map[string]any{"text": reportAlert, "consent": true, "consent_version": "v1"})
	if rec := r.post("/v1/customer/reports", string(body), "", "rep-http-0001"); rec.Code != http.StatusUnauthorized {
		t.Fatalf("anonymous: %d", rec.Code)
	}
	if rec := r.post("/v1/customer/reports", string(body), r.token(t, r.tenant, uuid.New(), "SUPERVISOR"), "rep-http-0001"); rec.Code != http.StatusUnauthorized {
		t.Fatalf("an operator token: %d", rec.Code)
	}
	if rec := r.post("/v1/customer/reports", string(body), r.aliceT, ""); rec.Code != http.StatusBadRequest {
		t.Fatalf("no idempotency key: %d", rec.Code)
	}
	noConsent, _ := json.Marshal(map[string]any{"text": reportAlert, "consent": false, "consent_version": "v1"})
	if rec := r.post("/v1/customer/reports", string(noConsent), r.aliceT, "rep-http-0002"); rec.Code != http.StatusUnprocessableEntity || !strings.Contains(rec.Body.String(), "CONSENT_REQUIRED") {
		t.Fatalf("no consent: %d %s", rec.Code, rec.Body.String())
	}

	rec := r.post("/v1/customer/reports", string(body), r.aliceT, "rep-http-0001")
	var rep creport.Report
	_ = json.Unmarshal(rec.Body.Bytes(), &rep)
	if rec.Code != http.StatusCreated || rep.Parsed.Fields["amount_minor"].Value != "1850000" || rep.Parsed.Fields["account_tail"].Value != "6789" {
		t.Fatalf("%d %s", rec.Code, rec.Body.String())
	}
	if strings.Contains(rec.Body.String(), "0123456789") {
		t.Fatal("the raw account number came back")
	}
	if again := r.post("/v1/customer/reports", string(body), r.aliceT, "rep-http-0001"); again.Code != http.StatusOK {
		t.Fatalf("repeat: %d", again.Code)
	}

	confirm := `{"amount_minor":1850000,"occurred_at":"2026-10-14T09:12:00+01:00","account_tail":"6789","reference":"NIP202610140912"}`
	path := "/v1/customer/reports/" + rep.ID.String()
	rec = r.post(path+"/confirm", confirm, r.aliceT, "")
	var out creport.Report
	_ = json.Unmarshal(rec.Body.Bytes(), &out)
	if rec.Code != http.StatusOK || out.Outcome == nil || out.Outcome.MatchStatus != creport.MatchVerifiedSuccess || out.Outcome.Closure == nil {
		t.Fatalf("%d %s", rec.Code, rec.Body.String())
	}
	if rec := r.post(path+"/confirm", `{"amount_minor":"1850.00","occurred_at":"2026-10-14T09:12:00+01:00"}`, r.aliceT, ""); rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("a fractional amount: %d", rec.Code)
	}
	if rec := r.post(path+"/confirm", `{"amount_minor":1850000}`, r.aliceT, ""); rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("missing time: %d", rec.Code)
	}

	// Bob cannot see, list, read or confirm Alice's report.
	for _, c := range []struct{ method, p, body string }{{"GET", path, ""}, {"POST", path + "/confirm", confirm}} {
		var rec *recorder
		if c.method == "GET" {
			rec = serve(r.handler, newReq(c.method, c.p, "", r.bobT))
		} else {
			rec = r.post(c.p, c.body, r.bobT, "")
		}
		if rec.Code != http.StatusNotFound {
			t.Errorf("%s %s as another customer: %d", c.method, c.p, rec.Code)
		}
	}
	list := serve(r.handler, newReq("GET", "/v1/customer/reports", "", r.bobT))
	if list.Code != http.StatusOK || strings.Contains(list.Body.String(), rep.ID.String()) {
		t.Fatalf("another customer's report was listed: %s", list.Body.String())
	}
	if rec := serve(r.handler, newReq("GET", "/v1/customer/reports/not-an-id", "", r.aliceT)); rec.Code != http.StatusNotFound {
		t.Fatalf("%d", rec.Code)
	}
}

func TestTooManyReportsGetAFriendlyRetryAfter(t *testing.T) {
	r := newReportRig(t)
	body, _ := json.Marshal(map[string]any{"text": reportAlert, "consent": true, "consent_version": "v1"})
	for i := 0; i < creport.DefaultReportsPerHour; i++ {
		if rec := r.post("/v1/customer/reports", string(body), r.aliceT, fmt.Sprintf("rep-rate-%04d", i)); rec.Code != http.StatusCreated {
			t.Fatalf("%d: %d %s", i, rec.Code, rec.Body.String())
		}
	}
	rec := r.post("/v1/customer/reports", string(body), r.aliceT, "rep-rate-9999")
	if rec.Code != http.StatusTooManyRequests || rec.Header().Get("Retry-After") == "" || !strings.Contains(rec.Body.String(), "try again") {
		t.Fatalf("%d %s %v", rec.Code, rec.Body.String(), rec.Header())
	}
	if rec := r.post("/v1/customer/reports", string(body), r.bobT, "rep-rate-0000"); rec.Code != http.StatusCreated {
		t.Fatalf("one customer's limit applied to another: %d", rec.Code)
	}
}

// Scope acceptance 5: a pasted sample alert is parsed, classified, given the right
// countdown, and queued as CUS001 on the same queue operators work from.
func TestAPastedAlertBecomesACustomerCaseWithTheRightCountdownOnTheSharedQueue(t *testing.T) {
	r := newReportRig(t)
	// Synthetic policy, approved the normal way: a 24 hour clock and a 12 hour internal target.
	setCfg := func(key string, v any, verify string) {
		raw, _ := json.Marshal(v)
		err := r.pool.WithTenant(context.Background(), r.tenant, func(tx pgx.Tx) error {
			svc := &configsvc.Service{Audit: r.appender}
			ver, err := svc.Propose(context.Background(), tx, r.tenant, uuid.New(), configsvc.ProposeInput{Key: key, Value: raw, SourceReference: "SYNTHETIC DEMO POLICY - not approved for production",
				EffectiveFrom: time.Now(), VerifyStatus: verify, Reason: "synthetic value for a test"})
			if err != nil {
				return err
			}
			_, err = svc.Decide(context.Background(), tx, r.tenant, uuid.New(), key, ver.Version, true, "")
			return err
		})
		if err != nil {
			t.Fatal(err)
		}
	}
	setCfg(configsvc.ClockKey("FAILED_INSTANT_TRANSFER"), configsvc.Clock{Situation: "Failed instant transfer", Duration: 24, Unit: "hours", StartEvent: "dispute_reported", Calendar: "WAT", Owner: "Compliance"}, "reported")
	setCfg("internal_target_hours.FAILED_INSTANT_TRANSFER", 12, "unverified")
	time.Sleep(10 * time.Millisecond)

	at := time.Now().Add(-2 * time.Hour).UTC().Truncate(time.Minute)
	r.transfer(t, tokA, "NIP202610140912", 1850000, at, "DEBIT_POSTED", "TRANSFER_SENT", "CREDIT_FAILED")
	alert := "Wema Alert: Acct 0123456789 debited NGN 18,500.00 on " + at.In(time.FixedZone("WAT", 3600)).Format("2006-01-02 15:04") + ". Desc: NIP TRANSFER | Ref: NIP202610140912"
	body, _ := json.Marshal(map[string]any{"text": alert, "consent": true, "consent_version": "v1"})
	rec := r.post("/v1/customer/reports", string(body), r.aliceT, "rep-accept-0001")
	var rep creport.Report
	_ = json.Unmarshal(rec.Body.Bytes(), &rep)
	if rec.Code != http.StatusCreated || rep.Parsed.Fields["amount_minor"].Value != "1850000" || rep.GuidedForm {
		t.Fatalf("parsed: %d %s", rec.Code, rec.Body.String())
	}
	path := "/v1/customer/reports/" + rep.ID.String()
	confirm := `{"amount_minor":1850000,"occurred_at":"` + at.Format(time.RFC3339) + `","account_tail":"6789","reference":"NIP202610140912"}`
	if rec = r.post(path+"/confirm", confirm, r.aliceT, ""); rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"matched_transaction"`) {
		t.Fatalf("confirm: %d %s", rec.Code, rec.Body.String())
	}

	rec = serve(r.handler, newReq("GET", path+"/deadline", "", r.aliceT))
	var dl deadline.Result
	_ = json.Unmarshal(rec.Body.Bytes(), &dl)
	if rec.Code != http.StatusOK || dl.Dispute != deadline.FailedTransfer || dl.Regulatory == nil || dl.Regulatory.Status != "reported" || dl.Internal == nil || dl.Binding == nil {
		t.Fatalf("deadline: %d %s", rec.Code, rec.Body.String())
	}
	if !dl.Internal.Due.Equal(at.Add(2 * time.Second).Add(12 * time.Hour)) {
		t.Fatalf("internal target %v", dl.Internal.Due)
	}
	if dl.State != deadline.OnTrack || dl.RemainingSec <= 0 {
		t.Fatalf("%+v", dl)
	}

	if rec = serve(r.handler, newReq("GET", path+"/draft", "", r.aliceT)); rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), "NGN 18,500.00") {
		t.Fatalf("draft: %d %s", rec.Code, rec.Body.String())
	}
	if rec = r.post(path+"/complaint", `{}`, r.bobT, ""); rec.Code != http.StatusNotFound {
		t.Fatalf("another customer submitted: %d", rec.Code)
	}
	rec = r.post(path+"/complaint", `{}`, r.aliceT, "")
	var sub creport.Submitted
	_ = json.Unmarshal(rec.Body.Bytes(), &sub)
	if rec.Code != http.StatusCreated || !sub.Created || sub.CaseRef == "" || sub.Dispute != deadline.FailedTransfer {
		t.Fatalf("submit: %d %s", rec.Code, rec.Body.String())
	}
	if again := r.post(path+"/complaint", `{}`, r.aliceT, ""); again.Code != http.StatusOK {
		t.Fatalf("repeat: %d", again.Code)
	}

	// The same queue operators work from now shows it as CUS001, from a customer, with both deadlines apart.
	ops := r.token(t, r.tenant, uuid.New(), "TXN_ANALYST")
	rec = serve(r.handler, newReq("GET", "/v1/exceptions", "", ops))
	var list struct{ Items []caseJSON }
	_ = json.Unmarshal(rec.Body.Bytes(), &list)
	var found *caseJSON
	for i := range list.Items {
		if list.Items[i].CaseRef == sub.CaseRef {
			found = &list.Items[i]
		}
	}
	if rec.Code != http.StatusOK || found == nil || found.Classification != "CUS001" || found.Source != "customer" || found.SLADueAt == nil || found.InternalTarget == nil || found.RegulatoryClock != "FAILED_INSTANT_TRANSFER" {
		t.Fatalf("%d %s", rec.Code, rec.Body.String())
	}
	if !found.InternalTarget.Before(*found.SLADueAt) {
		t.Fatalf("the bank's own target should come first: %v %v", found.InternalTarget, found.SLADueAt)
	}
}
