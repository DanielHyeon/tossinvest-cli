package engine_test

// runtime_wiring_test.go covers the production assembly the runtime consumes
// (openspec change add-engine-runtime, tasks 1.1 and 1.4):
//
//	the enumerated refusal   what an operator is shown when the gate will not open
//	the snapshot collector   one account read, in the shape both consumers declare
//	the restart recovery     the landed sequence, actually reached from the engine
//
// It reuses interlock_test.go's harness — the same httptest broker, the same
// isolated config directory, the same test-only clause-6 seam — because the
// question here is about wiring rather than about the interlock, and a second
// harness would be a second definition of "a started engine".

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/app/engine"
	"github.com/JungHoonGhae/tossinvest-cli/internal/attest"
	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/config"
	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/reconcile"
)

// --- the enumerated refusal ---------------------------------------------------------

// TestTheProtectionClauseIsNotEnumerated is the inverse of what this test used
// to assert, and the inversion is the point.
//
// It used to check that clause 6 was named well, because clause 6 was the one
// every correctly configured machine reached. Since interlock-gates-entry-not-exit
// it refuses no start, so naming it in the *startup refusal* enumeration would
// send an operator to fix something that did not stop them — and, worse, to look
// for a setting that does not exist.
//
// The list is "what stopped the engine". Broker-resident protection is no longer
// on it; where it shows up instead is the operating-mode record's
// `entry_permitted` field and the sentence beside it (design D6), and the
// gateway's refusal of any mutation that would raise exposure.
func TestTheProtectionClauseIsNotEnumerated(t *testing.T) {
	dir := isolate(t)
	writeGateConfig(t, dir, fullGate())
	writeCredentials(t, dir, "test-api-key-000000", "test-secret")
	writeAttestation(t, dir, nil)
	srv, _ := interlockServer(t, "123-45")

	_, err := openGateEngine(t, dir, srv, matchedGuardian())
	if err != nil {
		t.Fatalf("a complete configuration must start: %v", err)
	}
	if clauses := engine.UnmetInterlockClauses(err); len(clauses) != 0 {
		t.Errorf("a successful start enumerated unmet clauses: %v", clauses)
	}

	// And the enumeration itself must not carry a line nobody can produce: a
	// clause that never refuses is a message an operator would chase and never
	// clear.
	refusal := fmt.Errorf("%w: %w", engine.ErrAutomationGateRefused, errors.New("some other cause"))
	for _, line := range engine.UnmetInterlockClauses(refusal) {
		if strings.Contains(line, "ProtectionReady") {
			t.Errorf("the enumeration still offers a protection line: %q", line)
		}
	}
}

// TestEachInterlockClauseHasALine walks the clauses an operator can actually
// cause and checks each one is named rather than falling through to the
// catch-all. A clause added to the interlock with no line here would be a
// refusal that says "미상 조항".
func TestEachInterlockClauseHasALine(t *testing.T) {
	cases := map[string]struct {
		gate       config.AutomationGate
		attest     func(*attest.Attestation)
		policy     *config.Trading
		noGuardian bool
		wantHit    string
	}{
		"no Guardian":         {gate: fullGate(), noGuardian: true, wantHit: "Guardian 주입"},
		"an incomplete limit": {gate: partialGate(), wantHit: "위험 한도 전 항목"},
		"a policy that cannot exit": {
			gate:    fullGate(),
			policy:  &config.Trading{Place: true, Cancel: true},
			wantHit: "거래 정책",
		},
		"an expired attestation": {
			gate:    fullGate(),
			attest:  func(a *attest.Attestation) { a.ExpiresAt = interlockNow.Add(-time.Hour) },
			wantHit: "만료되었다",
		},
		"an attestation for another account": {
			gate:    fullGate(),
			attest:  func(a *attest.Attestation) { a.AccountRef = "999-99" },
			wantHit: "다른 계좌",
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			dir := isolate(t)
			if tc.policy != nil {
				writeGateConfigWith(t, dir, tc.gate, *tc.policy)
			} else {
				writeGateConfig(t, dir, tc.gate)
			}
			writeCredentials(t, dir, "test-api-key-000000", "test-secret")
			writeAttestation(t, dir, tc.attest)
			srv, _ := interlockServer(t, "123-45")

			var guardian execgw.Guardian
			if !tc.noGuardian {
				guardian = matchedGuardian()
			}
			var err error
			if tc.noGuardian {
				_, err = openGateWithoutProductionGuardian(t, dir, srv)
			} else {
				_, err = openGateEngine(t, dir, srv, guardian)
			}
			if err == nil {
				t.Fatal("the gate opened")
			}
			joined := strings.Join(engine.UnmetInterlockClauses(err), "\n")
			if !strings.Contains(joined, tc.wantHit) {
				t.Errorf("the enumeration for %s does not name it:\n%s", name, joined)
			}
		})
	}
}

// TestOnlyAnInterlockRefusalIsEnumerated. The command tells "the gate said no"
// from "the journal would not open" by this answer being nil, so a non-interlock
// error that produced a clause list would send an operator looking at their gate
// config for a disk problem.
func TestOnlyAnInterlockRefusalIsEnumerated(t *testing.T) {
	for _, err := range []error{
		nil,
		errors.New("engine: opening the journal: disk full"),
		engine.ErrOfficialCredentialsRequired,
		engine.ErrAccountUnresolved,
	} {
		if got := engine.UnmetInterlockClauses(err); got != nil {
			t.Errorf("UnmetInterlockClauses(%v) = %v, want nil", err, got)
		}
	}
}

// --- the account read ---------------------------------------------------------------

// TestTheSnapshotCollectorReadsTheAccountOnce is the §0.4 half: one holdings
// request per snapshot, whichever consumer asked, and the raw path is the one
// taken so an adoption's cost basis is the broker's own decimal string.
func TestTheSnapshotCollectorReadsTheAccountOnce(t *testing.T) {
	dir := isolate(t)
	writeGateConfig(t, dir, config.AutomationGate{})
	writeCredentials(t, dir, "test-api-key-000000", "test-secret")
	srv, counts := accountServer(t, "123-45")

	eng, err := openGateEngine(t, dir, srv, nil)
	if err != nil {
		t.Fatalf("engine: %v", err)
	}

	before := counts()
	snap, err := eng.SnapshotCollector(nil).Collect(context.Background())
	if err != nil {
		t.Fatalf("Collect: %v", err)
	}
	after := counts()

	if got := after.holdings - before.holdings; got != 1 {
		t.Errorf("the snapshot made %d holdings requests, want 1", got)
	}
	if got := after.balance - before.balance; got != 1 {
		t.Errorf("the snapshot made %d balance requests, want 1 (one currency)", got)
	}
	if len(snap.Holdings) != 1 {
		t.Fatalf("holdings = %+v", snap.Holdings)
	}
	if got := snap.Holdings[0].CostBasisRaw; got != "55000.1234" {
		t.Errorf("cost basis = %q; the raw path preserves the broker's own decimal", got)
	}
	if got := snap.Holdings[0].Market; got != "kr" {
		t.Errorf("market = %q, want kr", got)
	}
}

// TestTheSnapshotCurrencyIsTheGatesLimitCurrency, and a gate that names none
// falls back to KRW rather than reading every currency the account could hold.
func TestTheSnapshotCurrencyIsTheGatesLimitCurrency(t *testing.T) {
	dir := isolate(t)
	gate := fullGate()
	gate.Enabled = false
	gate.LimitCurrency = "usd"
	writeGateConfig(t, dir, gate)
	writeCredentials(t, dir, "test-api-key-000000", "test-secret")
	srv, _ := accountServer(t, "123-45")

	eng, err := openGateEngine(t, dir, srv, nil)
	if err != nil {
		t.Fatalf("engine: %v", err)
	}
	if got := eng.SnapshotCurrencies(); len(got) != 1 || got[0] != "USD" {
		t.Errorf("currencies = %v, want [USD]", got)
	}

	dir2 := isolate(t)
	writeGateConfig(t, dir2, config.AutomationGate{})
	writeCredentials(t, dir2, "test-api-key-000000", "test-secret")
	srv2, _ := accountServer(t, "123-45")
	eng2, err := openGateEngine(t, dir2, srv2, nil)
	if err != nil {
		t.Fatalf("engine: %v", err)
	}
	if got := eng2.SnapshotCurrencies(); len(got) != 1 || got[0] != engine.DefaultLimitCurrency {
		t.Errorf("currencies = %v, want [%s]", got, engine.DefaultLimitCurrency)
	}
}

// --- the restart recovery (task 1.4) --------------------------------------------------

// TestTheRestartRecoveryIsTheLandedOne is the consumption test: an attempt the
// previous process recorded and never dispatched is closed by the sequence the
// engine assembles, using journal.RecoverPending's own rules. No recovery code is
// added by this change and this proves the existing one is reached.
func TestTheRestartRecoveryIsTheLandedOne(t *testing.T) {
	dir := isolate(t)
	writeGateConfig(t, dir, config.AutomationGate{})
	writeCredentials(t, dir, "test-api-key-000000", "test-secret")
	srv, _ := accountServer(t, "123-45")

	eng, err := openGateEngine(t, dir, srv, nil)
	if err != nil {
		t.Fatalf("engine: %v", err)
	}
	ctx := context.Background()

	// A crash between "the intent was recorded" and "the request was sent" is the
	// RECORDED state, and the journal's rule is that it is provably unsent.
	attempt := recordPendingAttempt(t, eng.Journal, eng.AccountRef)

	// A real clock: the sequence's stabilisation genuinely waits, and the engine's
	// own clock in these tests is a fake nobody advances.
	recovery, err := eng.Recovery(reconcile.Options{
		Clock:     clock.System(),
		Stabilise: reconcile.Stabilisation{Interval: 10 * time.Millisecond},
	})
	if err != nil {
		t.Fatalf("Recovery: %v", err)
	}
	// The gate is latched by the constructor, before anything runs.
	if _, blocked := eng.Entry.Blocks()[execgw.ReasonRecoveryIncomplete]; !blocked {
		t.Error("the entry gate is open before recovery completed")
	}

	report, err := recovery.Run(ctx)
	if err != nil {
		t.Fatalf("recovery.Run: %v", err)
	}
	if !report.Complete {
		t.Fatal("the recovery did not complete")
	}
	if len(report.NotDispatched) != 1 || report.NotDispatched[0] != attempt {
		t.Errorf("NotDispatched = %v, want [%s] — journal.RecoverPending's own rule", report.NotDispatched, attempt)
	}

	rec, err := eng.Journal.LookupAttempt(ctx, attempt)
	if err != nil {
		t.Fatalf("LookupAttempt: %v", err)
	}
	if rec.State != journal.StateNotDispatched {
		t.Errorf("attempt state = %s, want NOT_DISPATCHED", rec.State)
	}
}

// TestARecoveryThatIsNeverRunLeavesTheGateShut. reconcile.New arms in its
// constructor precisely so this is true; the runtime relies on it by running the
// sequence before any loop starts.
func TestARecoveryThatIsNeverRunLeavesTheGateShut(t *testing.T) {
	dir := isolate(t)
	writeGateConfig(t, dir, config.AutomationGate{})
	writeCredentials(t, dir, "test-api-key-000000", "test-secret")
	srv, _ := accountServer(t, "123-45")

	eng, err := openGateEngine(t, dir, srv, nil)
	if err != nil {
		t.Fatalf("engine: %v", err)
	}
	if _, err := eng.Recovery(reconcile.Options{Clock: clock.System()}); err != nil {
		t.Fatalf("Recovery: %v", err)
	}
	if _, blocked := eng.Entry.Blocks()[execgw.ReasonRecoveryIncomplete]; !blocked {
		t.Error("an unrun recovery left the entry gate open")
	}
}

// --- helpers ---------------------------------------------------------------------------

type accountCounts struct{ holdings, balance, orders int }

// accountServer answers the reads a snapshot makes, and counts them.
func accountServer(t *testing.T, accountNo string) (*httptest.Server, func() accountCounts) {
	t.Helper()
	var mu sync.Mutex
	var counts accountCounts

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		mu.Lock()
		switch r.URL.Path {
		case "/api/v1/holdings":
			counts.holdings++
		case "/api/v1/buying-power":
			counts.balance++
		case "/api/v1/orders":
			counts.orders++
		}
		mu.Unlock()

		switch r.URL.Path {
		case "/oauth2/token":
			_, _ = w.Write([]byte(`{"access_token":"AT","expires_in":3600,"token_type":"Bearer"}`))
		case "/api/v1/accounts":
			_, _ = w.Write([]byte(`{"result":[{"accountNo":"` + accountNo +
				`","accountSeq":7,"accountType":"BROKERAGE"}]}`))
		case "/api/v1/holdings":
			_, _ = w.Write([]byte(`{"result":{"items":[{"symbol":"005930","marketCountry":"KR",` +
				`"currency":"KRW","quantity":"10","averagePurchasePrice":"55000.1234","lastPrice":"70000"}]}}`))
		case "/api/v1/buying-power":
			_, _ = w.Write([]byte(`{"result":{"cashBuyingPower":"1000000","currency":"KRW"}}`))
		case "/api/v1/orders":
			_, _ = w.Write([]byte(`{"result":{"items":[]}}`))
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)

	return srv, func() accountCounts {
		mu.Lock()
		defer mu.Unlock()
		return counts
	}
}

// recordPendingAttempt writes the intent-and-attempt pair a crash before dispatch
// leaves behind, and returns the attempt id.
func recordPendingAttempt(t *testing.T, j *journal.Journal, accountRef string) string {
	t.Helper()
	attempt, err := j.Prepare(context.Background(), journal.PrepareRequest{
		Intent: journal.Intent{
			ID: "intent-recovery-1", Market: "kr", TradingDay: "2026-03-30",
			AccountRef: accountRef, Symbol: "005930", Side: "BUY", OrderType: "LIMIT",
			Quantity: "1", Price: "70000", Currency: "KRW", Source: "engine",
			Fingerprint: "fp-recovery-1",
		},
		Kind: journal.KindPlace, AttemptID: "attempt-recovery-1",
	})
	if err != nil {
		t.Fatalf("Prepare: %v", err)
	}
	return attempt.ID()
}

// partialGate is a gate with one ceiling missing — the "부분적으로 무제한인
// 게이트" clause 2 refuses.
func partialGate() config.AutomationGate {
	gate := fullGate()
	gate.MaxTotalExposure = 0
	return gate
}
