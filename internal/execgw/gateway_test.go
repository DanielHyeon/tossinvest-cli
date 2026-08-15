package execgw_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/config"
	"github.com/JungHoonGhae/tossinvest-cli/internal/domain"
	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/official"
	"github.com/JungHoonGhae/tossinvest-cli/internal/orderintent"
	"github.com/JungHoonGhae/tossinvest-cli/internal/trading"
)

// Gateway tests (harden-execution-base task 2.5).
//
// The properties under test are the ones the engine-safety spec calls
// "ExecutionGateway 봉인":
//
//   - nothing reaches the broker before the journal commit (journal 선기록),
//   - nothing reaches the broker without a valid, unexpired, unused
//     GuardianDecision bound to the exact intent,
//   - a mutation is dispatched exactly once, whatever the transport does.

var fixedNow = time.Date(2026, 3, 30, 1, 30, 0, 0, time.UTC) // 10:30 KST, in session

// --- fakes ------------------------------------------------------------------

// fakeBroker is a trading.Broker that records what it was asked to do. Its call
// counters are the evidence for "exactly once" and "never called".
type fakeBroker struct {
	mu      sync.Mutex
	places  int
	cancels int
	amends  int
	result  domain.MutationResult
	err     error
}

// preWriteFailureTransport refuses the request before a connection or write can
// begin. Unlike a released TCP port, it has no scheduler, socket, or port-reuse
// dependency, so a NOT_DISPATCHED assertion is deterministic in CI.
type preWriteFailureTransport struct {
	calls  int
	method string
	path   string
}

func (t *preWriteFailureTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	t.calls++
	t.method = req.Method
	t.path = req.URL.Path
	return nil, errors.New("test transport refused before write")
}

func (b *fakeBroker) PlacePendingOrder(context.Context, orderintent.PlaceIntent) (domain.MutationResult, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.places++
	return b.result, b.err
}

func (b *fakeBroker) CancelPendingOrder(context.Context, orderintent.CancelIntent) (domain.MutationResult, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.cancels++
	return b.result, b.err
}

func (b *fakeBroker) AmendPendingOrder(context.Context, orderintent.AmendIntent) (domain.MutationResult, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.amends++
	return b.result, b.err
}

func (b *fakeBroker) GetOrderAvailableActions(context.Context, string) (map[string]any, error) {
	return map[string]any{}, nil
}

func (b *fakeBroker) totals() (int, int, int) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.places, b.cancels, b.amends
}

// --- helpers ----------------------------------------------------------------

func openJournal(t *testing.T, clk clock.Clock) *journal.Journal {
	t.Helper()
	j, err := journal.Open(context.Background(), journal.Options{
		Path:  filepath.Join(t.TempDir(), "journal.db"),
		Clock: clk,
		// TMPDIR is not necessarily on the allowlist (this repo lives on ntfs),
		// so the guard is satisfied with a fixed probe. The guard itself is
		// covered by internal/journal's own tests.
		FSProber: journal.FixedFSProber(journal.FSInfo{Name: "ext4", Magic: journal.MagicExt}),
	})
	if err != nil {
		t.Fatalf("journal.Open: %v", err)
	}
	t.Cleanup(func() { _ = j.Close() })
	return j
}

func openPolicy() config.Trading {
	return config.Trading{
		Place: true, Sell: true, Fractional: true, Cancel: true, Amend: true,
		Conditional: true, AllowLiveOrderActions: true,
	}
}

func placeIntent() orderintent.PlaceIntent {
	return orderintent.PlaceIntent{
		Symbol: "005930", Market: "kr", Side: "buy", OrderType: "limit",
		Quantity: 2, Price: 70000, CurrencyMode: "KRW",
	}
}

// newGateway wires a gateway over a fake broker.
func newGateway(t *testing.T, broker trading.Broker) (*execgw.Gateway, *journal.Journal, *clock.Fake) {
	t.Helper()
	clk := clock.NewFake(fixedNow)
	j := openJournal(t, clk)
	gw, err := execgw.New(execgw.Options{
		Journal:    j,
		Trading:    trading.NewService(openPolicy(), broker),
		Clock:      clk,
		AccountRef: "acct-7",
		Source:     "test",
	})
	if err != nil {
		t.Fatalf("execgw.New: %v", err)
	}
	return gw, j, clk
}

// The tests act as the decision issuer. That is not a shortcut around the
// gateway: a decision is only an authorisation once it is *in the journal the
// gateway reads*, so every helper below writes one and hands back the reference,
// exactly as a Guardian would.

func testLimits() execgw.Limits {
	return execgw.Limits{
		MaxQuantity:        execgw.Bound(10),
		MaxNotional:        execgw.Bound(1_000_000),
		MaxTotalExposure:   execgw.Bound(5_000_000),
		MaxDailyLossAmount: execgw.Bound(200_000),
		MaxDailyLossRatio:  execgw.Bound(0.02),
		Currency:           "KRW",
	}
}

func issuerFor(j *journal.Journal, clk clock.Clock) *execgw.Issuer {
	return &execgw.Issuer{Journal: j, Clock: clk, AccountRef: "acct-7", TTL: 30 * time.Second}
}

// entryDecision persists an EXPOSURE_RAISING decision that authorises exactly
// this place.
func entryDecision(t *testing.T, j *journal.Journal, clk clock.Clock,
	intent orderintent.PlaceIntent, limits execgw.Limits,
) execgw.GuardianDecision {
	t.Helper()
	return raisingDecision(t, j, clk, journal.KindPlace, intent.Market, intent.Symbol,
		intent.Side, intent.Quantity, intent.Price, limits)
}

// raisingDecision persists an EXPOSURE_RAISING decision for any mutation that
// adds exposure — a place, or an amend that raises quantity or price.
func raisingDecision(t *testing.T, j *journal.Journal, clk clock.Clock,
	kind journal.MutationKind, market, symbol, side string,
	quantity, price float64, limits execgw.Limits,
) execgw.GuardianDecision {
	t.Helper()
	d, err := issuerFor(j, clk).IssueEntry(context.Background(), execgw.EntryRequest{
		Kind: kind, Market: market, Symbol: symbol, Side: side,
		Quantity: quantity, EntryPrice: price, StopPrice: price * 0.9,
		PolicyVersion: "test/v1", Limits: limits,
	})
	if err != nil {
		t.Fatalf("issuing the entry decision: %v", err)
	}
	holdExposure(t, j, clk, d, quantity*price, limits)
	return d
}

// holdExposure puts the journal in the state a real issuance leaves it in: the
// entry decision, and the open-exposure headroom it consumes, both recorded
// (task 5.1 — the gateway refuses an EXPOSURE_RAISING decision that holds none).
//
// It is a fixture and not an issuer. The production path takes both rows in one
// transaction (journal.RecordDecisionAndReserve, design D1) and internal/execgw's
// RiskGuardian is what tests that; here the decision already exists and all this
// has to do is make the ledger agree with it. The plain execgw.Issuer stays as it
// was — it is the flatten saga's issuer, and flatten issues no entries.
func holdExposure(t *testing.T, j *journal.Journal, clk clock.Clock,
	decision execgw.GuardianDecision, notional float64, limits execgw.Limits,
) {
	t.Helper()
	currency := limits.Currency
	if currency == "" {
		currency = "KRW"
	}
	ceiling := limits.MaxTotalExposure.Value
	if ceiling <= 0 {
		ceiling = 1e12
	}
	version, err := j.ReservationVersion(context.Background(), "acct-7")
	if err != nil {
		t.Fatalf("ReservationVersion: %v", err)
	}
	if _, err := j.Reserve(context.Background(), journal.ReserveRequest{
		DecisionID:      decision.ID,
		AccountRef:      "acct-7",
		SnapshotAsOf:    clk.Now(),
		ObservedVersion: version,
		SnapshotUsage: []journal.AggregateAmount{
			{Kind: journal.ReservationKindOpenExposure, Amount: "0", Currency: currency},
		},
		Limits: []journal.AggregateAmount{
			{
				Kind:     journal.ReservationKindOpenExposure,
				Amount:   strconv.FormatFloat(ceiling, 'f', -1, 64),
				Currency: currency,
			},
		},
		Reservations: []journal.ReservationRequest{
			{
				ID:       "hold-" + decision.ID,
				Kind:     journal.ReservationKindOpenExposure,
				Amount:   strconv.FormatFloat(notional, 'f', -1, 64),
				Currency: currency,
			},
		},
	}); err != nil {
		t.Fatalf("holding the exposure for decision %s: %v", decision.ID, err)
	}
}

// exitDecision persists a RISK_REDUCING decision for a cancel, an amend or a
// reduce-only sell.
func exitDecision(t *testing.T, j *journal.Journal, clk clock.Clock,
	kind journal.MutationKind, market, symbol, side string, maxQuantity float64,
) execgw.GuardianDecision {
	t.Helper()
	d, err := issuerFor(j, clk).IssueReduction(context.Background(), execgw.ReductionRequest{
		Kind: kind, Market: market, Symbol: symbol, Side: side,
		MaxQuantity: maxQuantity, Reason: "test exit",
	})
	if err != nil {
		t.Fatalf("issuing the exit decision: %v", err)
	}
	return d
}

func placeRequest(t *testing.T, j *journal.Journal, clk clock.Clock) execgw.PlaceRequest {
	t.Helper()
	intent := placeIntent()
	return execgw.PlaceRequest{
		Intent:   intent,
		Decision: entryDecision(t, j, clk, intent, testLimits()),
	}
}

// --- tests ------------------------------------------------------------------

// TestPlaceHappyPathConfirms is the baseline: a valid decision produces exactly
// one broker call and a CONFIRMED attempt carrying the broker's order id.
func TestPlaceHappyPathConfirms(t *testing.T) {
	broker := &fakeBroker{result: domain.MutationResult{Kind: "place", Status: "accepted", OrderID: "O-1"}}
	gw, j, clk := newGateway(t, broker)

	out, err := gw.Place(context.Background(), placeRequest(t, j, clk))
	if err != nil {
		t.Fatalf("Place: %v", err)
	}
	if out.State != journal.StateConfirmed {
		t.Errorf("state: got %s, want CONFIRMED (%s)", out.State, out.Detail)
	}
	if out.BrokerOrderID != "O-1" {
		t.Errorf("broker order id: got %q, want O-1", out.BrokerOrderID)
	}
	if places, _, _ := broker.totals(); places != 1 {
		t.Errorf("broker place calls: got %d, want exactly 1", places)
	}

	rec, err := j.LookupAttempt(context.Background(), out.AttemptID)
	if err != nil {
		t.Fatalf("LookupAttempt: %v", err)
	}
	if rec.State != journal.StateConfirmed || rec.BrokerOrderID != "O-1" {
		t.Errorf("journal attempt: got state=%s brokerOrderID=%q", rec.State, rec.BrokerOrderID)
	}
	intent, err := j.LookupIntent(context.Background(), out.IntentID)
	if err != nil {
		t.Fatalf("LookupIntent: %v", err)
	}
	if intent.Symbol != "005930" || intent.Side != "BUY" || intent.Quantity != "2" {
		t.Errorf("journalled intent is not the submitted one: %+v", intent)
	}
	if intent.Fingerprint == "" {
		t.Error("intent must carry a fingerprint for IN_DOUBT matching")
	}
}

// TestGuardianRefusalsNeverReachBroker is the table of decision defects. Every
// one of them must be refused with the mutation journalled as NOT_DISPATCHED and
// the broker untouched — the journal-first ordering is what makes a refusal
// auditable rather than invisible.
//
// The defects are now expressed where they live: in the *persisted row*. A
// decision that does not authorise this order is one whose preimage says
// something else, and the only way to build one is to write it — which is
// precisely the guarantee under test (engine-safety "결정 영속과 신뢰 경계").
func TestGuardianRefusalsNeverReachBroker(t *testing.T) {
	intent := placeIntent()

	cases := []struct {
		name    string
		decide  func(t *testing.T, j *journal.Journal, clk clock.Clock) execgw.GuardianDecision
		advance time.Duration
		want    execgw.ReasonCode
	}{
		{
			name: "no decision at all",
			decide: func(*testing.T, *journal.Journal, clock.Clock) execgw.GuardianDecision {
				return execgw.GuardianDecision{}
			},
			want: execgw.ReasonGuardianMissing,
		},
		{
			name: "a decision nobody persisted",
			decide: func(*testing.T, *journal.Journal, clock.Clock) execgw.GuardianDecision {
				return execgw.GuardianDecision{ID: "dec-that-was-never-written"}
			},
			want: execgw.ReasonGuardianMissing,
		},
		{
			name: "preimage of a different order",
			decide: func(t *testing.T, j *journal.Journal, clk clock.Clock) execgw.GuardianDecision {
				other := intent
				other.Quantity = 9
				return entryDecision(t, j, clk, other, testLimits())
			},
			want: execgw.ReasonGuardianIntentMismatch,
		},
		{
			name: "preimage with a different entry price",
			decide: func(t *testing.T, j *journal.Journal, clk clock.Clock) execgw.GuardianDecision {
				other := intent
				other.Price = 69000
				return entryDecision(t, j, clk, other, testLimits())
			},
			want: execgw.ReasonGuardianIntentMismatch,
		},
		{
			name: "a buy wearing a RISK_REDUCING class",
			decide: func(t *testing.T, j *journal.Journal, clk clock.Clock) execgw.GuardianDecision {
				return exitDecision(t, j, clk, journal.KindPlace,
					intent.Market, intent.Symbol, intent.Side, intent.Quantity)
			},
			want: execgw.ReasonGuardianClassMismatch,
		},
		{
			name: "expired before submission",
			decide: func(t *testing.T, j *journal.Journal, clk clock.Clock) execgw.GuardianDecision {
				return entryDecision(t, j, clk, intent, testLimits())
			},
			advance: 31 * time.Second,
			want:    execgw.ReasonGuardianExpired,
		},
		{
			name: "quantity over the snapshot limit",
			decide: func(t *testing.T, j *journal.Journal, clk clock.Clock) execgw.GuardianDecision {
				limits := testLimits()
				limits.MaxQuantity = execgw.Bound(1)
				return entryDecision(t, j, clk, intent, limits)
			},
			want: execgw.ReasonGuardianLimitExceeded,
		},
		{
			name: "notional over the snapshot limit",
			decide: func(t *testing.T, j *journal.Journal, clk clock.Clock) execgw.GuardianDecision {
				limits := testLimits()
				limits.MaxNotional = execgw.Bound(1000)
				return entryDecision(t, j, clk, intent, limits)
			},
			want: execgw.ReasonGuardianLimitExceeded,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			broker := &fakeBroker{}
			gw, j, clk := newGateway(t, broker)

			decision := tc.decide(t, j, clk)
			clk.Advance(tc.advance)

			out, err := gw.Place(context.Background(), execgw.PlaceRequest{Intent: intent, Decision: decision})
			var rejected *execgw.RejectedError
			if !errors.As(err, &rejected) {
				t.Fatalf("want a RejectedError, got %v", err)
			}
			if rejected.Reason != tc.want {
				t.Errorf("reason: got %q, want %q", rejected.Reason, tc.want)
			}
			if places, _, _ := broker.totals(); places != 0 {
				t.Errorf("broker was called %d time(s) for a refused mutation", places)
			}
			if out.AttemptID == "" {
				t.Fatal("a refusal must still be journalled")
			}
			rec, err := j.LookupAttempt(context.Background(), out.AttemptID)
			if err != nil {
				t.Fatalf("LookupAttempt: %v", err)
			}
			if rec.State != journal.StateNotDispatched {
				t.Errorf("journal state: got %s, want NOT_DISPATCHED", rec.State)
			}
			if rec.ReasonCode != string(tc.want) {
				t.Errorf("journal reason code: got %q, want %q", rec.ReasonCode, tc.want)
			}
		})
	}
}

// TestNonceIsOneShot covers the spec's "nonce 재사용" scenario: the same decision
// cannot authorise a second submission, even for the identical intent.
func TestNonceIsOneShot(t *testing.T) {
	broker := &fakeBroker{result: domain.MutationResult{OrderID: "O-1"}}
	gw, j, clk := newGateway(t, broker)

	req := placeRequest(t, j, clk)
	if _, err := gw.Place(context.Background(), req); err != nil {
		t.Fatalf("first Place: %v", err)
	}

	_, err := gw.Place(context.Background(), req)
	var rejected *execgw.RejectedError
	if !errors.As(err, &rejected) {
		t.Fatalf("want a RejectedError on reuse, got %v", err)
	}
	if rejected.Reason != execgw.ReasonGuardianNonceReused {
		t.Errorf("reason: got %q, want %q", rejected.Reason, execgw.ReasonGuardianNonceReused)
	}
	if places, _, _ := broker.totals(); places != 1 {
		t.Errorf("broker place calls: got %d, want 1 (the reuse must not reach it)", places)
	}
}

// TestNonceIsOneShotUnderRace runs the same decision from many goroutines. Exactly
// one may win; a lost race that let two through would be a duplicated live order.
func TestNonceIsOneShotUnderRace(t *testing.T) {
	broker := &fakeBroker{result: domain.MutationResult{OrderID: "O-1"}}
	gw, j, clk := newGateway(t, broker)
	req := placeRequest(t, j, clk)

	const workers = 8
	var (
		wg       sync.WaitGroup
		mu       sync.Mutex
		accepted int
	)
	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			if _, err := gw.Place(context.Background(), req); err == nil {
				mu.Lock()
				accepted++
				mu.Unlock()
			}
		}()
	}
	wg.Wait()

	if accepted != 1 {
		t.Errorf("accepted submissions: got %d, want exactly 1", accepted)
	}
	if places, _, _ := broker.totals(); places != 1 {
		t.Errorf("broker place calls: got %d, want exactly 1", places)
	}
}

// TestJournalFailureBlocksSubmission is the spec's "커밋 실패 시 제출 차단": a
// journal that cannot record the intent must stop the mutation before the broker
// is touched.
func TestJournalFailureBlocksSubmission(t *testing.T) {
	broker := &fakeBroker{}
	gw, j, clk := newGateway(t, broker)
	// The decision is issued while the journal still works: what is under test is
	// the intent write failing, not the issuer's.
	req := placeRequest(t, j, clk)
	if err := j.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	if _, err := gw.Place(context.Background(), req); err == nil {
		t.Fatal("a closed journal must refuse the mutation")
	}
	if places, _, _ := broker.totals(); places != 0 {
		t.Errorf("broker was called %d time(s) although the journal never committed", places)
	}
}

// TestPolicyRefusalIsNotDispatched: the user's config still governs. A disabled
// action is refused by trading.Service, and the gateway must record that as
// "never left the process" rather than as an unknown outcome.
func TestPolicyRefusalIsNotDispatched(t *testing.T) {
	broker := &fakeBroker{}
	clk := clock.NewFake(fixedNow)
	j := openJournal(t, clk)
	closed := config.Trading{} // every toggle off
	gw, err := execgw.New(execgw.Options{
		Journal: j, Trading: trading.NewService(closed, broker), Clock: clk,
		AccountRef: "acct-7", Source: "test",
	})
	if err != nil {
		t.Fatalf("execgw.New: %v", err)
	}

	out, err := gw.Place(context.Background(), placeRequest(t, j, clk))
	if err == nil {
		t.Fatal("a disabled trading policy must refuse the mutation")
	}
	if out.State != journal.StateNotDispatched {
		t.Errorf("state: got %s, want NOT_DISPATCHED", out.State)
	}
	if places, _, _ := broker.totals(); places != 0 {
		t.Errorf("broker calls: got %d, want 0", places)
	}
}

// TestCancelAndAmendGoThroughTheSameGate keeps the two other mutation verbs on
// the identical path — a cancel that skipped the journal would be exactly the
// hole the gateway exists to close.
func TestCancelAndAmendGoThroughTheSameGate(t *testing.T) {
	broker := &fakeBroker{result: domain.MutationResult{OrderID: "O-9", CurrentOrderID: "O-9"}}
	gw, j, clk := newGateway(t, broker)
	ctx := context.Background()

	cancelIntent := orderintent.CancelIntent{OrderID: "O-1", Symbol: "005930"}
	cancelReq := execgw.CancelRequest{
		Intent:   cancelIntent,
		Order:    execgw.OrderRef{Market: "kr", Side: "BUY", Quantity: 2, Price: 70000, Currency: "KRW"},
		Decision: exitDecision(t, j, clk, journal.KindCancel, "kr", "005930", "BUY", 2),
	}
	out, err := gw.Cancel(ctx, cancelReq)
	if err != nil {
		t.Fatalf("Cancel: %v", err)
	}
	if out.State != journal.StateConfirmed {
		t.Errorf("cancel state: got %s (%s)", out.State, out.Detail)
	}
	rec, err := j.LookupAttempt(ctx, out.AttemptID)
	if err != nil {
		t.Fatalf("LookupAttempt: %v", err)
	}
	if rec.Kind != journal.KindCancel || rec.TargetOrderID != "O-1" {
		t.Errorf("cancel attempt: kind=%s target=%q", rec.Kind, rec.TargetOrderID)
	}

	newPrice := 70500.0
	amendIntent := orderintent.AmendIntent{OrderID: "O-1", Price: &newPrice}
	amendReq := execgw.AmendRequest{
		Intent: amendIntent,
		Symbol: "005930",
		Order:  execgw.OrderRef{Market: "kr", Side: "BUY", Quantity: 2, Price: 70000, Currency: "KRW"},
		// The amend lowers nothing and raises nothing (only the price moves up by
		// 500 on a resting BUY, which does raise exposure), so it is authorised as
		// an exposure-raising amend.
		Decision: raisingDecision(t, j, clk, journal.KindAmend, "kr", "005930", "BUY",
			2, newPrice, testLimits()),
	}
	amendOut, err := gw.Amend(ctx, amendReq)
	if err != nil {
		t.Fatalf("Amend: %v", err)
	}
	if amendOut.State != journal.StateConfirmed {
		t.Errorf("amend state: got %s (%s)", amendOut.State, amendOut.Detail)
	}
	if _, cancels, amends := broker.totals(); cancels != 1 || amends != 1 {
		t.Errorf("broker calls: cancels=%d amends=%d, want 1/1", cancels, amends)
	}
}

// TestAmendRaisingQuantityIsLimitChecked: an amend that increases exposure is a
// risk-increasing mutation and must be measured against the limit snapshot, the
// same as a place.
func TestAmendRaisingQuantityIsLimitChecked(t *testing.T) {
	broker := &fakeBroker{result: domain.MutationResult{CurrentOrderID: "O-9"}}
	gw, j, clk := newGateway(t, broker)

	qty := 50.0
	amendIntent := orderintent.AmendIntent{OrderID: "O-1", Quantity: &qty}
	_, err := gw.Amend(context.Background(), execgw.AmendRequest{
		Intent: amendIntent,
		Symbol: "005930",
		Order:  execgw.OrderRef{Market: "kr", Side: "BUY", Quantity: 2, Price: 70000, Currency: "KRW"},
		Decision: raisingDecision(t, j, clk, journal.KindAmend, "kr", "005930", "BUY",
			qty, 70000, testLimits()),
	})
	var rejected *execgw.RejectedError
	if !errors.As(err, &rejected) || rejected.Reason != execgw.ReasonGuardianLimitExceeded {
		t.Fatalf("want guardian_limit_exceeded, got %v", err)
	}
	if _, _, amends := broker.totals(); amends != 0 {
		t.Errorf("broker amend calls: got %d, want 0", amends)
	}
}

// --- transport classification over a real HTTP round trip -------------------

// officialGateway wires the gateway over the official client pointed at an
// httptest server, so the transport-fault classification is exercised end to end
// rather than against a hand-made error.
func officialGateway(t *testing.T, h http.HandlerFunc) (*execgw.Gateway, *journal.Journal, *clock.Fake, *int) {
	t.Helper()
	var posts int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost && strings.HasPrefix(r.URL.Path, "/api/v1/orders") {
			posts++
		}
		if r.URL.Path == "/oauth2/token" {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"access_token":"AT","expires_in":3600,"token_type":"Bearer"}`))
			return
		}
		h(w, r)
	}))
	t.Cleanup(srv.Close)

	off := official.New(official.Credentials{APIKey: "k", SecretKey: "s"},
		filepath.Join(t.TempDir(), "token.json"),
		official.WithBaseURL(srv.URL), official.WithHTTPClient(srv.Client()),
		official.WithAccountSeq(7))

	clk := clock.NewFake(fixedNow)
	j := openJournal(t, clk)
	gw, err := execgw.New(execgw.Options{
		Journal:    j,
		Trading:    trading.NewService(openPolicy(), &officialTestBroker{off: off}),
		Clock:      clk,
		AccountRef: "acct-7",
		Source:     "test",
	})
	if err != nil {
		t.Fatalf("execgw.New: %v", err)
	}
	return gw, j, clk, &posts
}

// officialTestBroker mirrors internal/app/engine's officialBroker: it is what the
// engine profile passes to trading.Service.
type officialTestBroker struct{ off *official.Client }

func (b *officialTestBroker) PlacePendingOrder(ctx context.Context, i orderintent.PlaceIntent) (domain.MutationResult, error) {
	return b.off.PlaceOrder(ctx, i)
}

func (b *officialTestBroker) CancelPendingOrder(ctx context.Context, i orderintent.CancelIntent) (domain.MutationResult, error) {
	return b.off.CancelOrder(ctx, i.OrderID)
}

func (b *officialTestBroker) AmendPendingOrder(ctx context.Context, i orderintent.AmendIntent) (domain.MutationResult, error) {
	return b.off.ModifyOrder(ctx, i)
}

func (b *officialTestBroker) GetOrderAvailableActions(context.Context, string) (map[string]any, error) {
	return map[string]any{}, nil
}

// TestTransportOutcomeTable pins how a broker response becomes an attempt state.
// The pessimistic entries are the important ones: 429 and 5xx are IN_DOUBT, never
// "failed", because a rate limiter can answer after the order reached the book.
func TestTransportOutcomeTable(t *testing.T) {
	cases := []struct {
		name   string
		status int
		body   string
		want   journal.AttemptState
	}{
		{"accepted", http.StatusOK, `{"result":{"orderId":"O-1"}}`, journal.StateConfirmed},
		{"bad request", http.StatusBadRequest, `{"message":"bad symbol"}`, journal.StateFailedConfirmed},
		{"unauthorized", http.StatusUnauthorized, `{"message":"nope"}`, journal.StateFailedConfirmed},
		{"rate limited", http.StatusTooManyRequests, `{"message":"slow down"}`, journal.StateInDoubt},
		{"server error", http.StatusInternalServerError, `{"message":"boom"}`, journal.StateInDoubt},
		{"gateway timeout", http.StatusGatewayTimeout, `{"message":"timeout"}`, journal.StateInDoubt},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			gw, j, clk, posts := officialGateway(t, func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tc.status)
				_, _ = w.Write([]byte(tc.body))
			})

			out, _ := gw.Place(context.Background(), placeRequest(t, j, clk))
			if out.State != tc.want {
				t.Errorf("state: got %s, want %s (%s)", out.State, tc.want, out.Detail)
			}
			// 401 is retried once by the official client with a refreshed token;
			// every other status must produce exactly one mutation request. Either
			// way the gateway itself must never retry.
			if tc.status != http.StatusUnauthorized && *posts != 1 {
				t.Errorf("mutation POSTs: got %d, want exactly 1 (mutations are never retried)", *posts)
			}
		})
	}
}

// TestUnreachableBrokerIsNotDispatched: a connection that never carried a byte is
// the one case we can safely close without a resolution procedure.
func TestUnreachableBrokerIsNotDispatched(t *testing.T) {
	transport := &preWriteFailureTransport{}
	client := &http.Client{Transport: transport}

	off := official.New(official.Credentials{APIKey: "k", SecretKey: "s"},
		filepath.Join(t.TempDir(), "token.json"),
		official.WithBaseURL("https://broker.test"), official.WithHTTPClient(client), official.WithAccountSeq(7))

	clk := clock.NewFake(fixedNow)
	j := openJournal(t, clk)
	gw, err := execgw.New(execgw.Options{
		Journal: j, Trading: trading.NewService(openPolicy(), &officialTestBroker{off: off}),
		Clock: clk, AccountRef: "acct-7", Source: "test",
	})
	if err != nil {
		t.Fatalf("execgw.New: %v", err)
	}

	out, _ := gw.Place(context.Background(), placeRequest(t, j, clk))
	if out.State != journal.StateNotDispatched {
		t.Errorf("state: got %s, want NOT_DISPATCHED (%s)", out.State, out.Detail)
	}
	// The single call is token acquisition. The transport refuses before any
	// socket write, so the order POST cannot be built or dispatched.
	if transport.calls != 1 {
		t.Errorf("transport calls: got %d, want exactly 1 token request", transport.calls)
	}
	if transport.method != http.MethodPost || transport.path != "/oauth2/token" {
		t.Errorf("sole transport request: got %s %s, want POST /oauth2/token", transport.method, transport.path)
	}
}

// --- structural seal --------------------------------------------------------

// TestGatewayExposesNoRawMutator is the compile-surface half of "raw mutator
// 미노출": no exported field or method of the gateway hands back something that
// can mutate orders on its own. If this test has to change, the seal is being
// opened.
func TestGatewayExposesNoRawMutator(t *testing.T) {
	brokerIface := reflect.TypeOf((*trading.Broker)(nil)).Elem()
	condIface := reflect.TypeOf((*trading.ConditionalBroker)(nil)).Elem()
	serviceType := reflect.TypeOf((*trading.Service)(nil))

	gwType := reflect.TypeOf((*execgw.Gateway)(nil))
	for i := 0; i < gwType.NumMethod(); i++ {
		m := gwType.Method(i)
		for out := 0; out < m.Type.NumOut(); out++ {
			ret := m.Type.Out(out)
			switch {
			case ret == serviceType:
				t.Errorf("method %s returns *trading.Service: the wrapped mutator must stay private", m.Name)
			case ret.Kind() == reflect.Interface && (ret.Implements(brokerIface) || ret.Implements(condIface)):
				t.Errorf("method %s returns a mutator interface", m.Name)
			case ret.Kind() != reflect.Interface && (ret.Implements(brokerIface) || ret.Implements(condIface)):
				t.Errorf("method %s returns a mutator implementation", m.Name)
			}
		}
	}

	gwStruct := reflect.TypeOf(execgw.Gateway{})
	for i := 0; i < gwStruct.NumField(); i++ {
		if f := gwStruct.Field(i); f.IsExported() {
			t.Errorf("Gateway.%s is exported: gateway state must not be reachable from outside", f.Name)
		}
	}
}

// --- pre-minted intent ids (task 7.4) ---------------------------------------

// TestAPreMintedIntentIDIsTheOneRecorded is the seam the exit observation loop's
// crash contract needs: it arms `exit_states.pending_intent_id` before anything
// is sent, so the id it armed has to be the id the intent is recorded under.
func TestAPreMintedIntentIDIsTheOneRecorded(t *testing.T) {
	broker := &fakeBroker{result: domain.MutationResult{Kind: "place", Status: "accepted", OrderID: "O-9"}}
	gw, j, clk := newGateway(t, broker)

	intent := orderintent.PlaceIntent{
		Symbol: "005930", Market: "kr", Side: "sell", OrderType: "limit",
		Quantity: 4, Price: 71000, CurrencyMode: "KRW",
	}
	out, err := gw.Place(context.Background(), execgw.PlaceRequest{
		Intent:   intent,
		Decision: exitDecision(t, j, clk, journal.KindPlace, "kr", "005930", "SELL", 4),
		IntentID: "exit-intent-1",
	})
	if err != nil {
		t.Fatalf("Place: %v", err)
	}
	if out.IntentID != "exit-intent-1" {
		t.Errorf("outcome intent id = %q, want the pre-minted one", out.IntentID)
	}
	if out.State != journal.StateConfirmed {
		t.Fatalf("state = %s, want CONFIRMED", out.State)
	}
	rec, err := j.LookupIntent(context.Background(), "exit-intent-1")
	if err != nil {
		t.Fatalf("LookupIntent: %v", err)
	}
	if rec.Symbol != "005930" || rec.Side != "SELL" {
		t.Errorf("recorded intent = %+v, want the submitted sell", rec)
	}
}

// TestAnOmittedIntentIDStillMintsOne keeps the default path a default: every
// other caller passes nothing and must be unaffected.
func TestAnOmittedIntentIDStillMintsOne(t *testing.T) {
	broker := &fakeBroker{result: domain.MutationResult{Kind: "place", Status: "accepted", OrderID: "O-10"}}
	gw, j, clk := newGateway(t, broker)

	out, err := gw.Place(context.Background(), placeRequest(t, j, clk))
	if err != nil {
		t.Fatalf("Place: %v", err)
	}
	if out.IntentID == "" {
		t.Fatal("the gateway must mint an intent id when the caller names none")
	}
}

// TestAReusedIntentIDIsARecoveryAndNotASecondIntent states the property the
// crash contract rests on, and the one it does *not* provide.
//
// Reusing the id records no second intent — Prepare recognises it — so a
// re-submission after a crash that never sent anything is an identity recovery
// under the id the proposal already armed. What Prepare does not do is refuse a
// second *attempt*, so "has this proposal already been submitted" is a question
// the observation loop has to ask before it re-submits, and IntentAttempted is
// what it asks with.
func TestAReusedIntentIDIsARecoveryAndNotASecondIntent(t *testing.T) {
	broker := &fakeBroker{result: domain.MutationResult{Kind: "place", Status: "accepted", OrderID: "O-11"}}
	gw, j, clk := newGateway(t, broker)
	ctx := context.Background()

	intent := orderintent.PlaceIntent{
		Symbol: "005930", Market: "kr", Side: "sell", OrderType: "limit",
		Quantity: 1, Price: 71000, CurrencyMode: "KRW",
	}
	attempted, err := j.IntentAttempted(ctx, "exit-intent-dup")
	if err != nil || attempted {
		t.Fatalf("IntentAttempted before any place = %v, %v; want false", attempted, err)
	}

	first := execgw.PlaceRequest{
		Intent:   intent,
		Decision: exitDecision(t, j, clk, journal.KindPlace, "kr", "005930", "SELL", 1),
		IntentID: "exit-intent-dup",
	}
	if _, err := gw.Place(ctx, first); err != nil {
		t.Fatalf("first Place: %v", err)
	}
	if attempted, err := j.IntentAttempted(ctx, "exit-intent-dup"); err != nil || !attempted {
		t.Fatalf("IntentAttempted after the place = %v, %v; want true", attempted, err)
	}

	// A second place under the same id with *different* terms is refused: the id
	// names one order, and letting its quantity move would make the armed
	// proposal describe something other than what was sent.
	mutated := first
	mutated.Intent.Quantity = 2
	mutated.Decision = exitDecision(t, j, clk, journal.KindPlace, "kr", "005930", "SELL", 2)
	if _, err := gw.Place(ctx, mutated); !errors.Is(err, journal.ErrIntentMutated) {
		t.Fatalf("re-placing a mutated intent: %v, want ErrIntentMutated", err)
	}
	places, _, _ := broker.totals()
	if places != 1 {
		t.Errorf("broker places = %d, want 1 — the mutated re-place must not reach the broker", places)
	}
}
