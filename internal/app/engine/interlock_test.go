package engine_test

// interlock_test.go covers the automation-gate startup interlock (task 4.2).
//
// The default case gets as much attention as the failure cases, because "the gate
// is off and nothing changed" is the state every machine is in today and §0.2
// makes it a hard requirement rather than a nice-to-have.

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
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
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/obs"
	"github.com/JungHoonGhae/tossinvest-cli/internal/official"
)

var interlockNow = time.Date(2026, 7, 26, 12, 0, 0, 0, time.UTC)

// stubGuardian is the injected risk authority.
//
// It declares the limits it would stamp on an EXPOSURE_RAISING decision, because
// since task 7.5 that is what the interlock's single-source clause asks of a
// Guardian: a risk authority that cannot say what it authorises against cannot be
// shown to be the same source the audit trail records.
type stubGuardian struct{ limits execgw.Limits }

func (stubGuardian) Authorize(context.Context, execgw.AuthorizationRequest) (execgw.GuardianDecision, error) {
	return execgw.GuardianDecision{}, errors.New("stub guardian authorises nothing")
}

func (g stubGuardian) ExposureLimits() execgw.Limits { return g.limits }

// silentGuardian authorises nothing and says nothing about its limits. It is the
// shape of a Guardian written before the single-source clause existed.
type silentGuardian struct{}

func (silentGuardian) Authorize(context.Context, execgw.AuthorizationRequest) (execgw.GuardianDecision, error) {
	return execgw.GuardianDecision{}, errors.New("silent guardian authorises nothing")
}

// fullGateLimits is the snapshot fullGate() produces. A Guardian carrying it is
// configured from the same source the interlock audits.
func fullGateLimits() execgw.Limits {
	return execgw.Limits{
		MaxQuantity:        execgw.Bound(10),
		MaxNotional:        execgw.Bound(1_000_000),
		MaxTotalExposure:   execgw.Bound(5_000_000),
		MaxDailyLossAmount: execgw.Bound(200_000),
		MaxDailyLossRatio:  execgw.Bound(0.02),
		Currency:           "KRW",
	}
}

// matchedGuardian is a Guardian configured from the audited limits.
func matchedGuardian() stubGuardian { return stubGuardian{limits: fullGateLimits()} }

// gateLimitsWith is fullGateLimits with one bound changed, for the cases that
// need the Guardian to agree with a config that is itself wrong.
func gateLimitsWith(mutate func(*execgw.Limits)) execgw.Limits {
	l := fullGateLimits()
	mutate(&l)
	return l
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

// openTradingPolicy is a policy that satisfies the interlock: it can enter, and
// — the part clause 3 is about — it can exit.
func openTradingPolicy() config.Trading {
	return config.Trading{Place: true, Sell: true, Cancel: true, Amend: true, AllowLiveOrderActions: true}
}

// writeGateConfig writes a config with the automation gate in the requested state.
func writeGateConfig(t *testing.T, dir string, gate config.AutomationGate) {
	t.Helper()
	writeGateConfigWith(t, dir, gate, openTradingPolicy())
}

// writeGateConfigWith is writeGateConfig with the trading policy under test.
func writeGateConfigWith(t *testing.T, dir string, gate config.AutomationGate, policy config.Trading) {
	t.Helper()
	cfg := config.DefaultFile()
	cfg.Trading = policy
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
	return openGateEngineLogging(t, dir, srv, guardian, nil)
}

// openGateEngineLogging is openGateEngine with the structured log captured.
func openGateEngineLogging(t *testing.T, dir string, srv *httptest.Server,
	guardian execgw.Guardian, logs io.Writer,
) (*engine.Context, error) {
	t.Helper()
	return openGate(t, dir, srv, guardian, logs, false)
}

// openProtectedGateEngine is openGateEngine with interlock clause 6 satisfied.
//
// Clause 6 (broker-resident protective execution) is an unmet constant in this
// build by design, so nothing this repository can configure produces a verified
// gate. The tests that are about the *other* clauses — the acceptance audit
// record, the verified log line, the toggle trail — would otherwise stop
// exercising an accepted start at all, which would leave that path untested
// until the protective-order change lands. They say so out loud by calling this
// helper, which reaches the test-only seam in export_test.go.
func openProtectedGateEngine(t *testing.T, dir string, srv *httptest.Server,
	guardian execgw.Guardian,
) (*engine.Context, error) {
	t.Helper()
	return openGate(t, dir, srv, guardian, nil, true)
}

// openProtectedGateEngineLogging is openProtectedGateEngine with the log captured.
func openProtectedGateEngineLogging(t *testing.T, dir string, srv *httptest.Server,
	guardian execgw.Guardian, logs io.Writer,
) (*engine.Context, error) {
	t.Helper()
	return openGate(t, dir, srv, guardian, logs, true)
}

func openGate(t *testing.T, dir string, srv *httptest.Server,
	guardian execgw.Guardian, logs io.Writer, protectionReady bool,
) (*engine.Context, error) {
	t.Helper()
	opts := engine.Options{
		ConfigDir: dir,
		Clock:     clock.NewFake(interlockNow),
		Guardian:  guardian,
		AuditFile: filepath.Join(dir, "audit.log"),
		Operator:  "test-operator",
		OfficialOptions: []official.Option{
			official.WithBaseURL(srv.URL),
			official.WithHTTPClient(srv.Client()),
		},
	}
	if logs != nil {
		opts.Logger = obs.NewLogger(obs.LogOptions{
			Writer: logs, JSON: true, Clock: clock.NewFake(interlockNow),
		})
	}
	opts.SetJournalProberForTest(journal.FixedFSProber(journal.FSInfo{
		Name: "ext4", Magic: journal.MagicExt,
	}))
	if protectionReady {
		opts.SetProtectionReadyForTest()
	}
	eng, err := engine.New(opts)
	if eng != nil {
		t.Cleanup(func() { _ = eng.Close() })
	}
	return eng, err
}

// fullGate is a gate with every limit the interlock requires (task 7.5). Before
// 7.5 only the two per-order ceilings existed and a gate carrying them was
// complete; a gate that stops there is now refused, which is the point.
func fullGate() config.AutomationGate {
	return config.AutomationGate{
		Enabled:            true,
		MaxOrderQuantity:   10,
		MaxOrderNotional:   1_000_000,
		MaxTotalExposure:   5_000_000,
		MaxDailyLossAmount: 200_000,
		MaxDailyLossRatio:  0.02,
		LimitCurrency:      "KRW",
	}
}

// TestGateOffStartsAndDoesNoGateWork is the §0.2 test, narrowed by task 7.1.
//
// The original assertion was "no broker call at all". That is no longer the
// contract and the change is deliberate: the account read moved out of the gate's
// verification and into construction, because the journal, the gateway and the
// reconcile projection are all scoped by the account (design D8 step 1). What
// remains true — and is what §0.2 is actually about — is that the gate itself
// does nothing: no attestation is read, no Guardian is published, no refusal is
// possible. So the account read is asserted at exactly one, and the gate's own
// outputs are asserted absent.
func TestGateOffStartsAndDoesNoGateWork(t *testing.T) {
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
	if !eng.Automation.AttestationExpiresAt.IsZero() {
		t.Error("the gate-off path must not read an attestation")
	}
	if got := accountCalls(); got != 1 {
		t.Errorf("account reads = %d, want exactly 1 — the construction-path read (task 7.1), "+
			"and nothing the gate adds to it", got)
	}
	if eng.AccountRef != "123-45" {
		t.Errorf("AccountRef = %q; the gate-off engine must still know its account", eng.AccountRef)
	}
}

// TestGateOnWithEverythingInPlaceStarts is the positive case.
func TestGateOnWithEverythingInPlaceStarts(t *testing.T) {
	dir := isolate(t)
	writeGateConfig(t, dir, fullGate())
	writeCredentials(t, dir, "test-api-key-000000", "test-secret")
	writeAttestation(t, dir, nil)
	srv, accountCalls := interlockServer(t, "123-45")

	eng, err := openProtectedGateEngine(t, dir, srv, matchedGuardian())
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
	// partialGate returns fullGate with one limit removed, which is the
	// combination clause 1 is written against: "부분적으로 무제한인 게이트는
	// 허가된 게이트가 아니다".
	partialGate := func(drop func(*config.AutomationGate)) config.AutomationGate {
		gate := fullGate()
		drop(&gate)
		return gate
	}

	cases := []struct {
		name        string
		gate        config.AutomationGate
		trading     *config.Trading
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
			guardian:  matchedGuardian(),
			accountNo: "123-45",
			wantErr:   engine.ErrLimitsRequired,
		},
		{
			name: "only the per-order ceilings are set",
			gate: partialGate(func(g *config.AutomationGate) {
				g.MaxTotalExposure, g.MaxDailyLossAmount, g.MaxDailyLossRatio = 0, 0, 0
			}),
			writeAtt:    true,
			guardian:    matchedGuardian(),
			accountNo:   "123-45",
			wantErr:     engine.ErrLimitsRequired,
			wantMessage: "total open exposure",
		},
		{
			name:        "the total exposure limit is missing",
			gate:        partialGate(func(g *config.AutomationGate) { g.MaxTotalExposure = 0 }),
			writeAtt:    true,
			guardian:    matchedGuardian(),
			accountNo:   "123-45",
			wantErr:     engine.ErrLimitsRequired,
			wantMessage: "total open exposure",
		},
		{
			name:        "the daily loss amount is missing",
			gate:        partialGate(func(g *config.AutomationGate) { g.MaxDailyLossAmount = 0 }),
			writeAtt:    true,
			guardian:    matchedGuardian(),
			accountNo:   "123-45",
			wantErr:     engine.ErrLimitsRequired,
			wantMessage: "daily loss amount",
		},
		{
			name:        "the daily loss ratio is missing",
			gate:        partialGate(func(g *config.AutomationGate) { g.MaxDailyLossRatio = 0 }),
			writeAtt:    true,
			guardian:    matchedGuardian(),
			accountNo:   "123-45",
			wantErr:     engine.ErrLimitsRequired,
			wantMessage: "daily loss ratio",
		},
		{
			name:        "a limit is set but not a usable number",
			gate:        partialGate(func(g *config.AutomationGate) { g.MaxDailyLossRatio = 1.5 }),
			writeAtt:    true,
			guardian:    stubGuardian{limits: gateLimitsWith(func(l *execgw.Limits) { l.MaxDailyLossRatio = execgw.Bound(1.5) })},
			accountNo:   "123-45",
			wantErr:     engine.ErrLimitsRequired,
			wantMessage: "bounds nothing",
		},
		{
			name:        "no currency on the money bounds",
			gate:        partialGate(func(g *config.AutomationGate) { g.LimitCurrency = "" }),
			writeAtt:    true,
			guardian:    stubGuardian{limits: gateLimitsWith(func(l *execgw.Limits) { l.Currency = "" })},
			accountNo:   "123-45",
			wantErr:     engine.ErrLimitsRequired,
			wantMessage: "currency",
		},
		{
			name:      "selling is disabled",
			gate:      fullGate(),
			trading:   &config.Trading{Place: true, Cancel: true, Amend: true, AllowLiveOrderActions: true},
			writeAtt:  true,
			guardian:  matchedGuardian(),
			accountNo: "123-45",
			// engine-safety: "매수는 가능한데 청산이 불가능한 조합으로는 기동할
			// 수 없다".
			wantErr:     engine.ErrTradingPolicyRefused,
			wantMessage: "trading.sell",
		},
		{
			name:        "live order actions are disabled",
			gate:        fullGate(),
			trading:     &config.Trading{Place: true, Sell: true, Cancel: true, Amend: true},
			writeAtt:    true,
			guardian:    matchedGuardian(),
			accountNo:   "123-45",
			wantErr:     engine.ErrTradingPolicyRefused,
			wantMessage: "allow_live_order_actions",
		},
		{
			// Placing is disabled. The clause used not to look, and the shape it
			// let through is the one its own sentence forbids: a sell IS a place
			// (internal/trading refuses on policy.Place regardless of side), so
			// this engine starts and then refuses its first stop.
			name:        "placing is disabled",
			gate:        fullGate(),
			trading:     &config.Trading{Sell: true, Cancel: true, Amend: true, AllowLiveOrderActions: true},
			writeAtt:    true,
			guardian:    matchedGuardian(),
			accountNo:   "123-45",
			wantErr:     engine.ErrTradingPolicyRefused,
			wantMessage: "trading.place",
		},
		{
			// Cancelling is disabled. The exit observer cancels its own armed
			// proposal, so an exit path without it stalls rather than exits.
			name:        "cancelling is disabled",
			gate:        fullGate(),
			trading:     &config.Trading{Place: true, Sell: true, Amend: true, AllowLiveOrderActions: true},
			writeAtt:    true,
			guardian:    matchedGuardian(),
			accountNo:   "123-45",
			wantErr:     engine.ErrTradingPolicyRefused,
			wantMessage: "trading.cancel",
		},
		{
			name:     "the Guardian carries different limits",
			gate:     fullGate(),
			writeAtt: true,
			guardian: stubGuardian{limits: gateLimitsWith(func(l *execgw.Limits) {
				l.MaxTotalExposure = execgw.Bound(50_000_000)
			})},
			accountNo:   "123-45",
			wantErr:     engine.ErrGuardianLimitsMismatch,
			wantMessage: "do not match the configured limits",
		},
		{
			name:        "the Guardian cannot state its limits",
			gate:        fullGate(),
			writeAtt:    true,
			guardian:    silentGuardian{},
			accountNo:   "123-45",
			wantErr:     engine.ErrGuardianLimitsMismatch,
			wantMessage: "does not report its exposure limits",
		},
		{
			name:      "no attestation at all",
			gate:      fullGate(),
			writeAtt:  false,
			guardian:  matchedGuardian(),
			accountNo: "123-45",
			wantErr:   attest.ErrMissing,
		},
		{
			name:        "attestation expired",
			gate:        fullGate(),
			writeAtt:    true,
			attestation: func(a *attest.Attestation) { a.ExpiresAt = interlockNow.Add(-time.Hour) },
			guardian:    matchedGuardian(),
			accountNo:   "123-45",
			wantErr:     attest.ErrExpired,
			// engine-safety's scenario: "기동이 거부되고 재검증 안내가 출력된다".
			wantMessage: "verify-execution-capability",
		},
		{
			name:      "attestation is for another account",
			gate:      fullGate(),
			writeAtt:  true,
			guardian:  matchedGuardian(),
			accountNo: "999-99",
			wantErr:   attest.ErrAccountMismatch,
		},
		{
			name:        "attestation misses an endpoint the engine calls",
			gate:        fullGate(),
			writeAtt:    true,
			attestation: func(a *attest.Attestation) { a.Endpoints = []string{"GET /api/v1/accounts"} },
			guardian:    matchedGuardian(),
			accountNo:   "123-45",
			wantErr:     attest.ErrEndpointNotAttested,
		},
		{
			// The list grew in task 5.2 (the exit policy observes the latest
			// price), and an attestation written before that covers everything
			// *except* the new call. That is the shape the drift guard exists for,
			// so it gets its own case rather than being covered by the
			// one-endpoint attestation above.
			name:        "attestation predates the price read",
			gate:        fullGate(),
			writeAtt:    true,
			attestation: func(a *attest.Attestation) { a.Endpoints = withoutEndpoint("GET /api/v1/prices") },
			guardian:    matchedGuardian(),
			accountNo:   "123-45",
			wantErr:     attest.ErrEndpointNotAttested,
			wantMessage: "prices",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			dir := isolate(t)
			policy := openTradingPolicy()
			if tc.trading != nil {
				policy = *tc.trading
			}
			writeGateConfigWith(t, dir, tc.gate, policy)
			writeCredentials(t, dir, "test-api-key-000000", "test-secret")
			if tc.writeAtt {
				writeAttestation(t, dir, tc.attestation)
			}
			srv, _ := interlockServer(t, tc.accountNo)

			var eng *engine.Context
			var err error
			if tc.name == "no Guardian injected" {
				eng, err = openGateWithoutProductionGuardian(t, dir, srv)
			} else {
				eng, err = openGateEngine(t, dir, srv, tc.guardian)
			}
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
	if _, err := openGateEngine(t, dir, srv, matchedGuardian()); err != nil {
		t.Fatalf("first start: %v", err)
	}

	// Operator turns the gate on and raises a limit.
	writeGateConfig(t, dir, fullGate())
	if _, err := openProtectedGateEngine(t, dir, srv, matchedGuardian()); err != nil {
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

	if _, err := openGateEngine(t, dir, srv, matchedGuardian()); err == nil {
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

// --- task 7.5: what the accepted start leaves behind ------------------------

// TestAcceptedStartAuditsEveryLimit.
//
// The acceptance record used to name two of the five limits, which was fine when
// two was all there were. An operator reading the trail a month later is asking
// "what was this engine authorised to do that day", and a record that answers
// two-fifths of that question answers none of it.
func TestAcceptedStartAuditsEveryLimit(t *testing.T) {
	dir := isolate(t)
	writeGateConfig(t, dir, fullGate())
	writeCredentials(t, dir, "test-api-key-000000", "test-secret")
	writeAttestation(t, dir, nil)
	srv, _ := interlockServer(t, "123-45")

	if _, err := openProtectedGateEngine(t, dir, srv, matchedGuardian()); err != nil {
		t.Fatalf("a fully attested gate must start: %v", err)
	}

	entries := readAudit(t, filepath.Join(dir, "audit.log"))
	accepted := lastEntryFor(entries, "engine.automation_gate.enabled", audit.ActionGateAccepted)
	if accepted == nil {
		t.Fatalf("no acceptance recorded; entries = %+v", entries)
	}
	for _, want := range []string{
		"max_order_quantity=10",
		"max_order_notional=1000000",
		"max_total_exposure=5000000",
		"max_daily_loss_amount=200000",
		"max_daily_loss_ratio=0.02",
		"currency=KRW",
	} {
		if !strings.Contains(accepted.Detail, want) {
			t.Errorf("acceptance detail %q is missing %q", accepted.Detail, want)
		}
	}
	if strings.Contains(accepted.Detail, "123-45") {
		t.Errorf("the account must be masked in the audit detail: %q", accepted.Detail)
	}

	// The three limits added in 7.5 are settings, so a change to one of them is
	// auditable in its own right (§0.5).
	for _, setting := range []string{
		"engine.automation_gate.max_total_exposure",
		"engine.automation_gate.max_daily_loss_amount",
		"engine.automation_gate.max_daily_loss_ratio",
	} {
		if lastEntryFor(entries, setting, audit.ActionLimitChange) == nil {
			t.Errorf("no audit entry for %s; a limit nobody can diff is a limit nobody controls", setting)
		}
	}
}

// TestGateDecisionIsLoggedStructurally is the other half of §0.5's requirement:
// the audit trail is for the operator's history, the structured log is what a
// running system counts.
func TestGateDecisionIsLoggedStructurally(t *testing.T) {
	t.Run("verified", func(t *testing.T) {
		dir := isolate(t)
		writeGateConfig(t, dir, fullGate())
		writeCredentials(t, dir, "test-api-key-000000", "test-secret")
		writeAttestation(t, dir, nil)
		srv, _ := interlockServer(t, "123-45")

		var logs bytes.Buffer
		if _, err := openProtectedGateEngineLogging(t, dir, srv, matchedGuardian(), &logs); err != nil {
			t.Fatalf("start: %v", err)
		}
		line := decodeLastLog(t, logs.String())

		if line["event"] != string(obs.EventOperatingMode) {
			t.Errorf("event = %v, want %s", line["event"], obs.EventOperatingMode)
		}
		if line["gate_enabled"] != true || line["gate_verified"] != true {
			t.Errorf("gate fields = %v/%v, want true/true", line["gate_enabled"], line["gate_verified"])
		}
		if line["max_total_exposure"] != "5000000" {
			t.Errorf("max_total_exposure = %v", line["max_total_exposure"])
		}
		if _, present := line["reason"]; present {
			t.Errorf("a verified start must carry no refusal reason: %v", line["reason"])
		}
	})

	t.Run("refused", func(t *testing.T) {
		dir := isolate(t)
		writeGateConfig(t, dir, fullGate())
		writeCredentials(t, dir, "test-api-key-000000", "test-secret")
		writeAttestation(t, dir, func(a *attest.Attestation) {
			a.ExpiresAt = interlockNow.Add(-time.Hour)
		})
		srv, _ := interlockServer(t, "123-45")

		var logs bytes.Buffer
		if _, err := openGateEngineLogging(t, dir, srv, matchedGuardian(), &logs); err == nil {
			t.Fatal("startup must be refused")
		}
		line := decodeLastLog(t, logs.String())

		if line["level"] != "WARN" {
			t.Errorf("level = %v, want WARN — a refused gate is not routine", line["level"])
		}
		if line["gate_verified"] != false {
			t.Errorf("gate_verified = %v, want false", line["gate_verified"])
		}
		reason, _ := line["reason"].(string)
		if !strings.Contains(reason, "expired") {
			t.Errorf("reason = %q, want the cause in it", reason)
		}
	})

	t.Run("gate off", func(t *testing.T) {
		dir := isolate(t)
		writeGateConfig(t, dir, config.AutomationGate{})
		writeCredentials(t, dir, "test-api-key-000000", "test-secret")
		srv, _ := interlockServer(t, "123-45")

		var logs bytes.Buffer
		if _, err := openGateEngineLogging(t, dir, srv, nil, &logs); err != nil {
			t.Fatalf("start: %v", err)
		}
		line := decodeLastLog(t, logs.String())
		if line["gate_enabled"] != false || line["gate_verified"] != false {
			t.Errorf("gate fields = %v/%v, want false/false", line["gate_enabled"], line["gate_verified"])
		}
	})
}

// decodeLastLog parses the last JSON line written to the log.
func decodeLastLog(t *testing.T, out string) map[string]any {
	t.Helper()
	lines := strings.Split(strings.TrimSpace(out), "\n")
	if len(lines) == 0 || lines[0] == "" {
		t.Fatal("the interlock wrote no structured log line")
	}
	var line map[string]any
	if err := json.Unmarshal([]byte(lines[len(lines)-1]), &line); err != nil {
		t.Fatalf("decoding %q: %v", lines[len(lines)-1], err)
	}
	return line
}

// --- task 5.2: clause 6, broker-resident protection --------------------------

// withoutEndpoint is the engine's endpoint list with one call removed: the shape
// of an attestation written before the list grew.
func withoutEndpoint(drop string) []string {
	var out []string
	for _, e := range engine.RequiredEndpoints() {
		if e != drop {
			out = append(out, e)
		}
	}
	return out
}

// TestNothingElseRefusesTheOperatorConfiguration is what this test became.
//
// It used to be TestTheGateRefusesWithoutBrokerSideProtection, and its premise —
// that a fully attested operator configuration still cannot start — is exactly
// what interlock-gates-entry-not-exit inverts. The behaviour it asserted now
// lives in two places: the runtime coming up is interlock_entry_test.go, and the
// refusal of a raising mutation is internal/execgw's protection_test.go.
//
// What does not live anywhere else is the assertion this keeps: that on a
// configuration an operator can actually produce, *no clause objects*. Clause 6
// used to prove that by being reached last; with clause 6 gone from the list, the
// only way to say it is to say it — enumerate every refusal and require the
// absence of all of them, and require the audit to record an acceptance rather
// than a refusal.
//
// It matters because the failure it catches is silent. A start that refuses for
// an earlier reason would still look like "the engine did not come up", and the
// operator would go on believing it is the protective-order change they are
// waiting for.
func TestNothingElseRefusesTheOperatorConfiguration(t *testing.T) {
	dir := isolate(t)
	writeGateConfig(t, dir, fullGate())
	writeCredentials(t, dir, "test-api-key-000000", "test-secret")
	writeAttestation(t, dir, nil)
	srv, _ := interlockServer(t, "123-45")

	eng, err := openGateEngine(t, dir, srv, matchedGuardian())
	for _, clause := range []error{
		engine.ErrAutomationGateRefused,
		engine.ErrGuardianRequired,
		engine.ErrLimitsRequired,
		engine.ErrTradingPolicyRefused,
		engine.ErrGuardianLimitsMismatch,
		engine.ErrGatewayRequired,
		engine.ErrKeylessTransport,
		attest.ErrExpired,
		attest.ErrAccountMismatch,
		attest.ErrEndpointNotAttested,
	} {
		if errors.Is(err, clause) {
			t.Errorf("a complete operator configuration was refused by %v", clause)
		}
	}
	if err != nil {
		t.Fatalf("start: %v", err)
	}
	if eng == nil {
		t.Fatal("a start with no refusal must return an engine")
	}

	// The acceptance is on the record, and no refusal is. An audit trail that
	// shows a refusal here would describe a gate that did not come up (§0.5).
	entries := readAudit(t, filepath.Join(dir, "audit.log"))
	if refused := lastEntryFor(entries, "engine.automation_gate.enabled",
		audit.ActionGateRefused); refused != nil {
		t.Errorf("a successful start recorded a refusal: %+v", refused)
	}
	if accepted := lastEntryFor(entries, "engine.automation_gate.enabled",
		audit.ActionGateAccepted); accepted == nil {
		t.Fatalf("the acceptance was not recorded; entries = %+v", entries)
	}
}

// TestTheProfileReportsItsProtectionReadiness: an operator asking "why will the
// gate not come up" can see the answer without turning it on.
func TestTheProfileReportsItsProtectionReadiness(t *testing.T) {
	dir := isolate(t)
	writeGateConfig(t, dir, config.AutomationGate{})
	writeCredentials(t, dir, "test-api-key-000000", "test-secret")
	srv, _ := interlockServer(t, "123-45")

	eng, err := openGateEngine(t, dir, srv, nil)
	if err != nil {
		t.Fatalf("an engine with the gate off must start: %v", err)
	}
	if eng.Automation.Protection != engine.ProtectionUnwired {
		t.Errorf("Protection = %q, want %q — this build wires no protective execution",
			eng.Automation.Protection, engine.ProtectionUnwired)
	}
}

// TestTheEngineRequiresThePriceRead. The exit policy's observation loop calls it
// on a timer, so an attestation that does not cover it does not describe the
// engine (engine-safety clause 2's drift rule). The soak proves the call and the
// retry matrix bounds it; this list is what makes the attestation demand it.
func TestTheEngineRequiresThePriceRead(t *testing.T) {
	var found bool
	for _, e := range engine.RequiredEndpoints() {
		if e == "GET /api/v1/prices" {
			found = true
		}
	}
	if !found {
		t.Errorf("RequiredEndpoints() = %v, want the price read the exit policy observes",
			engine.RequiredEndpoints())
	}
}
