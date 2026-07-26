package engine_test

// interlock_test.go covers the automation-gate startup interlock (task 4.2).
//
// The default case gets as much attention as the failure cases, because "the gate
// is off and nothing changed" is the state every machine is in today and §0.2
// makes it a hard requirement rather than a nice-to-have.

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/app/engine"
	"github.com/JungHoonGhae/tossinvest-cli/internal/attest"
	"github.com/JungHoonGhae/tossinvest-cli/internal/audit"
	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/config"
	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/official"
)

var interlockNow = time.Date(2026, 7, 26, 12, 0, 0, 0, time.UTC)

// stubGuardian is the injected risk authority. Phase 1 only needs it to exist —
// the interlock checks that somebody is authorising orders, not what they decide.
type stubGuardian struct{}

func (stubGuardian) Authorize(context.Context, execgw.AuthorizationRequest) (execgw.GuardianDecision, error) {
	return execgw.GuardianDecision{}, errors.New("stub guardian authorises nothing")
}

// interlockServer answers the account read and counts it, so a test can prove the
// gate-off path performs no I/O.
func interlockServer(t *testing.T, accountNo string) (*httptest.Server, func() int) {
	t.Helper()
	var mu sync.Mutex
	calls := 0

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/oauth2/token":
			_, _ = w.Write([]byte(`{"access_token":"AT","expires_in":3600,"token_type":"Bearer"}`))
		case "/api/v1/accounts":
			mu.Lock()
			calls++
			mu.Unlock()
			_, _ = w.Write([]byte(`{"result":[{"accountNo":"` + accountNo + `","accountSeq":7,"accountType":"BROKERAGE"}]}`))
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)

	return srv, func() int {
		mu.Lock()
		defer mu.Unlock()
		return calls
	}
}

// writeGateConfig writes a config with the automation gate in the requested state.
func writeGateConfig(t *testing.T, dir string, gate config.AutomationGate) {
	t.Helper()
	cfg := config.DefaultFile()
	cfg.Trading = config.Trading{Place: true, Cancel: true, Amend: true, AllowLiveOrderActions: true}
	cfg.Engine.AutomationGate = gate
	data, err := json.Marshal(cfg)
	if err != nil {
		t.Fatalf("marshal config: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "config.json"), data, 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}
}

func writeAttestation(t *testing.T, dir string, mutate func(*attest.Attestation)) string {
	t.Helper()
	a := attest.Attestation{
		FormatVersion: attest.FormatVersion,
		AccountRef:    "123-45",
		IssuedAt:      interlockNow.Add(-24 * time.Hour),
		ExpiresAt:     interlockNow.Add(30 * 24 * time.Hour),
		SoakDays:      3,
		Endpoints:     engine.RequiredEndpoints(),
		VerifiedBy:    "operator",
	}
	if mutate != nil {
		mutate(&a)
	}
	path := filepath.Join(dir, attest.FileName)
	if err := attest.Save(path, a); err != nil {
		t.Fatalf("Save attestation: %v", err)
	}
	return path
}

func openGateEngine(t *testing.T, dir string, srv *httptest.Server, guardian execgw.Guardian) (*engine.Context, error) {
	t.Helper()
	return engine.New(engine.Options{
		ConfigDir: dir,
		Clock:     clock.NewFake(interlockNow),
		Guardian:  guardian,
		AuditFile: filepath.Join(dir, "audit.log"),
		Operator:  "test-operator",
		OfficialOptions: []official.Option{
			official.WithBaseURL(srv.URL),
			official.WithHTTPClient(srv.Client()),
		},
	})
}

func fullGate() config.AutomationGate {
	return config.AutomationGate{
		Enabled:          true,
		MaxOrderQuantity: 10,
		MaxOrderNotional: 1_000_000,
		LimitCurrency:    "KRW",
	}
}

// TestGateOffStartsAndTouchesNothing is the §0.2 test. With the gate off — which
// is every config in existence — startup must behave exactly as it did before the
// interlock: no attestation read, no broker call, no refusal.
func TestGateOffStartsAndTouchesNothing(t *testing.T) {
	dir := isolate(t)
	writeGateConfig(t, dir, config.AutomationGate{})
	writeCredentials(t, dir, "test-api-key-000000", "test-secret")
	srv, accountCalls := interlockServer(t, "123-45")

	eng, err := openGateEngine(t, dir, srv, nil)
	if err != nil {
		t.Fatalf("an engine with the gate off must start: %v", err)
	}
	if eng.Automation.Enabled {
		t.Error("Automation.Enabled = true with the gate off")
	}
	if eng.Automation.Verified {
		t.Error("an off gate is not a verified gate")
	}
	if eng.Guardian != nil {
		t.Error("no Guardian may be published when the gate is off")
	}
	if got := accountCalls(); got != 0 {
		t.Errorf("the gate-off path made %d broker call(s); it must make none", got)
	}
}

// TestGateOnWithEverythingInPlaceStarts is the positive case.
func TestGateOnWithEverythingInPlaceStarts(t *testing.T) {
	dir := isolate(t)
	writeGateConfig(t, dir, fullGate())
	writeCredentials(t, dir, "test-api-key-000000", "test-secret")
	writeAttestation(t, dir, nil)
	srv, accountCalls := interlockServer(t, "123-45")

	eng, err := openGateEngine(t, dir, srv, stubGuardian{})
	if err != nil {
		t.Fatalf("a fully attested gate must start: %v", err)
	}
	if !eng.Automation.Verified {
		t.Error("Verified = false after a successful interlock")
	}
	if eng.Guardian == nil {
		t.Error("the injected Guardian must be published once verified")
	}
	if eng.Automation.Limits.MaxQuantity.Value != 10 || eng.Automation.Limits.MaxNotional.Value != 1_000_000 {
		t.Errorf("limits = %+v, want the configured ceilings", eng.Automation.Limits)
	}
	if eng.Automation.Limits.Currency != "KRW" {
		t.Errorf("limit currency = %q", eng.Automation.Limits.Currency)
	}
	if got := accountCalls(); got != 1 {
		t.Errorf("account reads = %d, want exactly 1", got)
	}
	if masked := eng.Automation.MaskedAccount(); strings.Contains(masked, "123") {
		t.Errorf("the account must be masked in reportable output, got %q", masked)
	}
}

// TestGateOnRefusals walks every precondition. Each one must refuse startup and
// return no engine — a partially interlocked engine is one that can still trade.
func TestGateOnRefusals(t *testing.T) {
	cases := []struct {
		name        string
		gate        config.AutomationGate
		attestation func(*attest.Attestation)
		writeAtt    bool
		guardian    execgw.Guardian
		accountNo   string
		wantErr     error
		wantMessage string
	}{
		{
			name:      "no Guardian injected",
			gate:      fullGate(),
			writeAtt:  true,
			guardian:  nil,
			accountNo: "123-45",
			wantErr:   engine.ErrGuardianRequired,
		},
		{
			name: "limits are zero",
			gate: config.AutomationGate{Enabled: true, LimitCurrency: "KRW"},
			// The attestation is fine; the limits are not, and "no limit" is not
			// an authorisation.
			writeAtt:  true,
			guardian:  stubGuardian{},
			accountNo: "123-45",
			wantErr:   engine.ErrLimitsRequired,
		},
		{
			name:      "no attestation at all",
			gate:      fullGate(),
			writeAtt:  false,
			guardian:  stubGuardian{},
			accountNo: "123-45",
			wantErr:   attest.ErrMissing,
		},
		{
			name:        "attestation expired",
			gate:        fullGate(),
			writeAtt:    true,
			attestation: func(a *attest.Attestation) { a.ExpiresAt = interlockNow.Add(-time.Hour) },
			guardian:    stubGuardian{},
			accountNo:   "123-45",
			wantErr:     attest.ErrExpired,
			// engine-safety's scenario: "기동이 거부되고 재검증 안내가 출력된다".
			wantMessage: "verify-execution-capability",
		},
		{
			name:      "attestation is for another account",
			gate:      fullGate(),
			writeAtt:  true,
			guardian:  stubGuardian{},
			accountNo: "999-99",
			wantErr:   attest.ErrAccountMismatch,
		},
		{
			name:        "attestation misses an endpoint the engine calls",
			gate:        fullGate(),
			writeAtt:    true,
			attestation: func(a *attest.Attestation) { a.Endpoints = []string{"GET /api/v1/accounts"} },
			guardian:    stubGuardian{},
			accountNo:   "123-45",
			wantErr:     attest.ErrEndpointNotAttested,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			dir := isolate(t)
			writeGateConfig(t, dir, tc.gate)
			writeCredentials(t, dir, "test-api-key-000000", "test-secret")
			if tc.writeAtt {
				writeAttestation(t, dir, tc.attestation)
			}
			srv, _ := interlockServer(t, tc.accountNo)

			eng, err := openGateEngine(t, dir, srv, tc.guardian)
			if err == nil {
				t.Fatal("startup must be refused")
			}
			if eng != nil {
				t.Error("a refused startup must return no engine at all")
			}
			if !errors.Is(err, engine.ErrAutomationGateRefused) {
				t.Errorf("err = %v, want it to wrap ErrAutomationGateRefused", err)
			}
			if !errors.Is(err, tc.wantErr) {
				t.Errorf("err = %v, want it to wrap %v", err, tc.wantErr)
			}
			if tc.wantMessage != "" && !strings.Contains(err.Error(), tc.wantMessage) {
				t.Errorf("message %q does not tell the operator about %q", err, tc.wantMessage)
			}
		})
	}
}

// TestGateToggleIsAudited is engine-safety's "게이트 토글 변경" scenario: the audit
// log carries the previous value, the new value and the time.
func TestGateToggleIsAudited(t *testing.T) {
	dir := isolate(t)
	writeCredentials(t, dir, "test-api-key-000000", "test-secret")
	writeAttestation(t, dir, nil)
	srv, _ := interlockServer(t, "123-45")

	// First start: gate off. This establishes the baseline the later change is
	// measured against.
	writeGateConfig(t, dir, config.AutomationGate{})
	if _, err := openGateEngine(t, dir, srv, stubGuardian{}); err != nil {
		t.Fatalf("first start: %v", err)
	}

	// Operator turns the gate on and raises a limit.
	writeGateConfig(t, dir, fullGate())
	if _, err := openGateEngine(t, dir, srv, stubGuardian{}); err != nil {
		t.Fatalf("second start: %v", err)
	}

	entries := readAudit(t, filepath.Join(dir, "audit.log"))
	toggle := lastEntryFor(entries, "engine.automation_gate.enabled", audit.ActionGateToggle)
	if toggle == nil {
		t.Fatalf("no gate toggle recorded; entries = %+v", entries)
	}
	if toggle.Old != "false" || toggle.New != "true" {
		t.Errorf("toggle old/new = %q/%q, want false/true", toggle.Old, toggle.New)
	}
	if toggle.At.IsZero() {
		t.Error("the toggle entry has no timestamp")
	}
	if toggle.Subject != "test-operator" {
		t.Errorf("subject = %q, want the operator identity", toggle.Subject)
	}

	quantity := lastEntryFor(entries, "engine.automation_gate.max_order_quantity", audit.ActionLimitChange)
	if quantity == nil {
		t.Fatalf("no limit change recorded; entries = %+v", entries)
	}
	if quantity.Old != "0" || quantity.New != "10" {
		t.Errorf("limit old/new = %q/%q, want 0/10", quantity.Old, quantity.New)
	}

	if accepted := lastEntryFor(entries, "engine.automation_gate.enabled", audit.ActionGateAccepted); accepted == nil {
		t.Error("a verified start must be recorded")
	}
}

// TestRefusedStartIsAudited: a refused start produces no engine, so the audit log
// is the only trace that somebody tried to enable an unverified gate.
func TestRefusedStartIsAudited(t *testing.T) {
	dir := isolate(t)
	writeGateConfig(t, dir, fullGate())
	writeCredentials(t, dir, "test-api-key-000000", "test-secret")
	writeAttestation(t, dir, func(a *attest.Attestation) { a.ExpiresAt = interlockNow.Add(-time.Hour) })
	srv, _ := interlockServer(t, "123-45")

	if _, err := openGateEngine(t, dir, srv, stubGuardian{}); err == nil {
		t.Fatal("startup must be refused")
	}

	entries := readAudit(t, filepath.Join(dir, "audit.log"))
	refused := lastEntryFor(entries, "engine.automation_gate.enabled", audit.ActionGateRefused)
	if refused == nil {
		t.Fatalf("the refusal was not recorded; entries = %+v", entries)
	}
	if !strings.Contains(refused.Detail, "expired") {
		t.Errorf("refusal detail = %q, want the cause in it", refused.Detail)
	}
}

// TestUnchangedSettingsDoNotGrowTheAuditLog: an engine that restarts every day
// must not write four lines a day about settings nobody touched.
func TestUnchangedSettingsDoNotGrowTheAuditLog(t *testing.T) {
	dir := isolate(t)
	writeGateConfig(t, dir, config.AutomationGate{})
	writeCredentials(t, dir, "test-api-key-000000", "test-secret")
	srv, _ := interlockServer(t, "123-45")

	if _, err := openGateEngine(t, dir, srv, nil); err != nil {
		t.Fatalf("first start: %v", err)
	}
	first := len(readAudit(t, filepath.Join(dir, "audit.log")))

	for i := 0; i < 3; i++ {
		if _, err := openGateEngine(t, dir, srv, nil); err != nil {
			t.Fatalf("restart %d: %v", i, err)
		}
	}
	if got := len(readAudit(t, filepath.Join(dir, "audit.log"))); got != first {
		t.Errorf("audit entries grew from %d to %d across restarts that changed nothing", first, got)
	}
}

func readAudit(t *testing.T, path string) []audit.Entry {
	t.Helper()
	log, err := audit.Open(audit.Options{Path: path})
	if err != nil {
		t.Fatalf("open audit log: %v", err)
	}
	entries, err := log.Entries()
	if err != nil {
		t.Fatalf("read audit log: %v", err)
	}
	return entries
}

func lastEntryFor(entries []audit.Entry, setting, action string) *audit.Entry {
	for i := len(entries) - 1; i >= 0; i-- {
		if entries[i].Setting == setting && entries[i].Action == action {
			return &entries[i]
		}
	}
	return nil
}
