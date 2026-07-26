package execgw_test

// Idempotent replay tests (extend-execution-contract tasks 2.1/2.2).
//
// Two things are under test and they are different in kind:
//
//   - the *guards*, which are the reason a replay is safe at all: it resends
//     only the stored bytes, only for an IN_DOUBT attempt, only while the
//     broker's ten-minute key window has margin left, only with the capability
//     attested, and only within a counted cap.
//   - the *response rules*, which are deliberately not the dispatch
//     classifier's: a replay's 422 must never settle an attempt
//     FAILED_CONFIRMED, and its 409 must not spend the cap.

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/obs"
	"github.com/JungHoonGhae/tossinvest-cli/internal/official"
	"github.com/JungHoonGhae/tossinvest-cli/internal/orderintent"
	"github.com/JungHoonGhae/tossinvest-cli/internal/trading"
)

// --- fixture ----------------------------------------------------------------

type replayFixture struct {
	gw   *execgw.Gateway
	j    *journal.Journal
	clk  *clock.Fake
	gate *execgw.EntryGate
	srv  *httptest.Server

	mu       sync.Mutex
	bodies   []string
	handler  func(w http.ResponseWriter, r *http.Request)
	attested bool
}

// newReplayFixture wires a gateway whose broker always fails (so a place lands
// IN_DOUBT with its wire body and key already stored) and whose replay
// transport points at an httptest server the test drives.
func newReplayFixture(t *testing.T, cfg execgw.ReplayConfig) *replayFixture {
	t.Helper()
	clk := clock.NewFake(fixedNow)
	j := openJournal(t, clk)
	gate := execgw.NewEntryGate(clk, map[execgw.RequiredQuery]time.Duration{})
	f := &replayFixture{j: j, clk: clk, gate: gate, attested: true}

	f.srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body := make([]byte, r.ContentLength)
		if r.ContentLength > 0 {
			if _, err := r.Body.Read(body); err != nil && err.Error() != "EOF" {
				t.Errorf("reading the replayed body: %v", err)
			}
		}
		f.mu.Lock()
		f.bodies = append(f.bodies, string(body))
		handler := f.handler
		f.mu.Unlock()
		if handler == nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		handler(w, r)
	}))
	t.Cleanup(f.srv.Close)

	gw, err := execgw.New(execgw.Options{
		Journal:    j,
		Trading:    trading.NewService(openPolicy(), &fakeBroker{err: official.ErrServer}),
		Clock:      clk,
		AccountRef: "acct-7",
		Source:     "test",
		Entry:      gate,
		Replay: execgw.HTTPReplay{
			BaseURL: f.srv.URL,
			HTTP:    f.srv.Client(),
			Headers: func(context.Context) (map[string]string, error) {
				return map[string]string{"X-Tossinvest-Account": "7"}, nil
			},
		},
		ReplayConfig: cfg,
		Attested: func(context.Context) bool {
			f.mu.Lock()
			defer f.mu.Unlock()
			return f.attested
		},
	})
	if err != nil {
		t.Fatalf("execgw.New: %v", err)
	}
	f.gw = gw
	return f
}

func (f *replayFixture) respond(handler func(w http.ResponseWriter, r *http.Request)) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.handler = handler
}

func (f *replayFixture) sentBodies() []string {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]string(nil), f.bodies...)
}

// inDoubtPlace drives one place into IN_DOUBT and returns its attempt record.
func (f *replayFixture) inDoubtPlace(t *testing.T) journal.AttemptRecord {
	t.Helper()
	out, err := f.gw.Place(context.Background(), placeRequest(t, f.j, f.clk))
	if err == nil {
		t.Fatal("the fixture broker must fail so the attempt lands IN_DOUBT")
	}
	if out.State != journal.StateInDoubt {
		t.Fatalf("state: got %s, want IN_DOUBT (%s)", out.State, out.Detail)
	}
	rec, err := f.j.LookupAttempt(context.Background(), out.AttemptID)
	if err != nil {
		t.Fatalf("LookupAttempt: %v", err)
	}
	if rec.WireBody == "" || rec.ClientOrderID == "" {
		t.Fatalf("the attempt must carry a stored body and key: body=%q key=%q", rec.WireBody, rec.ClientOrderID)
	}
	return rec
}

// replayAsync runs the entry point while advancing the fake clock past every
// wait it performs.
func replayAsync(t *testing.T, f *replayFixture, attemptID string, advance time.Duration) execgw.ReplayOutcome {
	t.Helper()
	type result struct {
		out execgw.ReplayOutcome
		err error
	}
	done := make(chan result, 1)
	go func() {
		out, err := f.gw.ReplayInDoubt(context.Background(), attemptID)
		done <- result{out, err}
	}()
	deadline := time.Now().Add(10 * time.Second)
	for {
		select {
		case r := <-done:
			if r.err != nil {
				t.Fatalf("ReplayInDoubt: %v", r.err)
			}
			return r.out
		case <-time.After(2 * time.Millisecond):
			if f.clk.Sleepers() > 0 {
				f.clk.Advance(advance)
			}
			if time.Now().After(deadline) {
				t.Fatal("the replay never finished")
			}
		}
	}
}

func okBody(orderID, clientOrderID string) string {
	payload := map[string]any{"result": map[string]any{"orderId": orderID}}
	if clientOrderID != "" {
		payload["result"].(map[string]any)["clientOrderId"] = clientOrderID
	}
	encoded, _ := json.Marshal(payload)
	return string(encoded)
}

func errorBody(code, message string) string {
	encoded, _ := json.Marshal(map[string]any{
		"error": map[string]any{"code": code, "message": message},
	})
	return string(encoded)
}

// --- 2.1: the guards --------------------------------------------------------

// TestReplaySendsExactlyTheStoredBody is the whole point of storing bytes: the
// replay is byte-identical to the original request, so the broker recognises the
// key and returns the original order instead of creating a second one.
func TestReplaySendsExactlyTheStoredBody(t *testing.T) {
	f := newReplayFixture(t, execgw.ReplayConfig{})
	rec := f.inDoubtPlace(t)
	f.respond(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(okBody("O-recovered", rec.ClientOrderID)))
	})

	out := replayAsync(t, f, rec.ID, time.Second)

	bodies := f.sentBodies()
	if len(bodies) != 1 {
		t.Fatalf("replay requests: got %d, want exactly 1", len(bodies))
	}
	if bodies[0] != rec.WireBody {
		t.Errorf("replayed body:\n got %s\nwant %s", bodies[0], rec.WireBody)
	}
	if out.State != journal.StateConfirmed {
		t.Fatalf("state: got %s, want CONFIRMED (%s)", out.State, out.Detail)
	}
	if out.BrokerOrderID != "O-recovered" {
		t.Errorf("broker order id: got %q, want O-recovered", out.BrokerOrderID)
	}
	if out.QueryFallback {
		t.Error("a recovered identity must not ask for the query fallback")
	}

	stored, err := f.j.LookupAttempt(context.Background(), rec.ID)
	if err != nil {
		t.Fatalf("LookupAttempt: %v", err)
	}
	if stored.State != journal.StateConfirmed || stored.BrokerOrderID != "O-recovered" {
		t.Errorf("journal: state=%s brokerOrderID=%q", stored.State, stored.BrokerOrderID)
	}
	if stored.ReasonCode != journal.ReasonReplayRecovered {
		t.Errorf("reason code: got %q, want %q", stored.ReasonCode, journal.ReasonReplayRecovered)
	}
	if stored.ReplayCount != 1 {
		t.Errorf("replay count: got %d, want 1", stored.ReplayCount)
	}
}

// TestReplayStoresTheIdentifierVerbatim: `orderId` is an opaque token (openapi
// contracts no shape), so what the broker sent is what is stored — whitespace,
// case and all.
func TestReplayStoresTheIdentifierVerbatim(t *testing.T) {
	f := newReplayFixture(t, execgw.ReplayConfig{})
	rec := f.inDoubtPlace(t)
	const padded = "  O-Padded-Id  "
	f.respond(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(okBody(padded, rec.ClientOrderID)))
	})

	out := replayAsync(t, f, rec.ID, time.Second)
	if out.BrokerOrderID != padded {
		t.Fatalf("broker order id: got %q, want %q verbatim", out.BrokerOrderID, padded)
	}
	stored, err := f.j.LookupAttempt(context.Background(), rec.ID)
	if err != nil {
		t.Fatalf("LookupAttempt: %v", err)
	}
	if stored.BrokerOrderID != padded {
		t.Errorf("stored id: got %q, want %q verbatim", stored.BrokerOrderID, padded)
	}
}

// TestReplayIsOffWithoutAttestation is the default state of this build: the
// capability has not been measured against the real broker, so nothing is
// resent [미측정 — 2b 전 비활성].
func TestReplayIsOffWithoutAttestation(t *testing.T) {
	f := newReplayFixture(t, execgw.ReplayConfig{})
	rec := f.inDoubtPlace(t)
	f.mu.Lock()
	f.attested = false
	f.mu.Unlock()

	out := replayAsync(t, f, rec.ID, time.Second)
	if len(f.sentBodies()) != 0 {
		t.Fatalf("an unattested replay must send nothing, sent %d", len(f.sentBodies()))
	}
	if out.Reason != execgw.ReasonReplayNotAttested || !out.QueryFallback {
		t.Errorf("outcome: reason=%s fallback=%v, want replay_not_attested + fallback", out.Reason, out.QueryFallback)
	}
	if out.State != journal.StateInDoubt {
		t.Errorf("state: got %s, want the attempt left IN_DOUBT", out.State)
	}
}

// TestReplayIsRefusedWithoutAnAttestationHook pins the *default*: a gateway that
// was never given an attestation check has not been attested.
func TestReplayIsRefusedWithoutAnAttestationHook(t *testing.T) {
	clk := clock.NewFake(fixedNow)
	j := openJournal(t, clk)
	gw, err := execgw.New(execgw.Options{
		Journal: j, Trading: trading.NewService(openPolicy(), &fakeBroker{err: official.ErrServer}),
		Clock: clk, AccountRef: "acct-7", Source: "test",
		Replay: stubReplay{},
	})
	if err != nil {
		t.Fatalf("execgw.New: %v", err)
	}
	out, err := gw.Place(context.Background(), placeRequest(t, j, clk))
	if err == nil || out.State != journal.StateInDoubt {
		t.Fatalf("setup: state=%s err=%v", out.State, err)
	}
	res, err := gw.ReplayInDoubt(context.Background(), out.AttemptID)
	if err != nil {
		t.Fatalf("ReplayInDoubt: %v", err)
	}
	if res.Reason != execgw.ReasonReplayNotAttested {
		t.Errorf("reason: got %s, want replay_not_attested by default", res.Reason)
	}
}

type stubReplay struct{}

func (stubReplay) ReplayPlace(context.Context, execgw.ReplayBody) (execgw.ReplayResponse, error) {
	return execgw.ReplayResponse{}, nil
}

// TestReplayIsRefusedOutsideTheKeyWindow: the key is documented to live ten
// minutes, and the margin keeps the last minute of that unused because the only
// evidence the window is open is a local clock reading.
func TestReplayIsRefusedOutsideTheKeyWindow(t *testing.T) {
	f := newReplayFixture(t, execgw.ReplayConfig{})
	rec := f.inDoubtPlace(t)
	f.clk.Advance(execgw.ReplayKeyTTL - execgw.DefaultReplayMargin)

	out := replayAsync(t, f, rec.ID, time.Second)
	if len(f.sentBodies()) != 0 {
		t.Fatalf("an expired key must not be replayed, sent %d", len(f.sentBodies()))
	}
	if out.Reason != execgw.ReasonReplayExpired || !out.QueryFallback {
		t.Errorf("outcome: reason=%s fallback=%v, want replay_key_window_expired + fallback",
			out.Reason, out.QueryFallback)
	}
}

// TestSecondReplayRechecksTheWindow is the spec's "두 번째 재생의 시간 재검사":
// the elapsed check runs before *each* send, not once per procedure.
func TestSecondReplayRechecksTheWindow(t *testing.T) {
	f := newReplayFixture(t, execgw.ReplayConfig{})
	rec := f.inDoubtPlace(t)
	// An answer that settles nothing, so the procedure wants a second replay.
	f.respond(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
	})

	// Every wait jumps the clock past what is left of the key's life.
	out := replayAsync(t, f, rec.ID, execgw.ReplayKeyTTL)

	if len(f.sentBodies()) != 1 {
		t.Fatalf("replay requests: got %d, want the second one refused by the window check",
			len(f.sentBodies()))
	}
	if out.Reason != execgw.ReasonReplayExpired || !out.QueryFallback {
		t.Errorf("outcome: reason=%s fallback=%v, want replay_key_window_expired + fallback",
			out.Reason, out.QueryFallback)
	}
}

// TestReplayStopsAtTheCap: two counted replays and then the query fallback,
// whatever the broker keeps answering.
func TestReplayStopsAtTheCap(t *testing.T) {
	f := newReplayFixture(t, execgw.ReplayConfig{})
	rec := f.inDoubtPlace(t)
	f.respond(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	out := replayAsync(t, f, rec.ID, execgw.DefaultReplayMinInterval)

	if got := len(f.sentBodies()); got != execgw.DefaultMaxReplays {
		t.Fatalf("replay requests: got %d, want the cap of %d", got, execgw.DefaultMaxReplays)
	}
	if out.Reason != execgw.ReasonReplayExhausted || !out.QueryFallback {
		t.Errorf("outcome: reason=%s fallback=%v, want replay_attempts_exhausted + fallback",
			out.Reason, out.QueryFallback)
	}
	stored, err := f.j.LookupAttempt(context.Background(), rec.ID)
	if err != nil {
		t.Fatalf("LookupAttempt: %v", err)
	}
	if stored.State != journal.StateInDoubt {
		t.Errorf("state: got %s, want the attempt left IN_DOUBT for the fallback", stored.State)
	}
	if stored.ReplayCount != execgw.DefaultMaxReplays {
		t.Errorf("stored replay count: got %d, want %d", stored.ReplayCount, execgw.DefaultMaxReplays)
	}
}

// TestReplayKeepsTheMinimumInterval: the second replay does not follow the first
// immediately.
func TestReplayKeepsTheMinimumInterval(t *testing.T) {
	f := newReplayFixture(t, execgw.ReplayConfig{MinInterval: 30 * time.Second})
	rec := f.inDoubtPlace(t)
	var stamps []time.Time
	f.respond(func(w http.ResponseWriter, _ *http.Request) {
		f.mu.Lock()
		stamps = append(stamps, f.clk.Now())
		f.mu.Unlock()
		w.WriteHeader(http.StatusInternalServerError)
	})

	replayAsync(t, f, rec.ID, 10*time.Second)

	f.mu.Lock()
	defer f.mu.Unlock()
	if len(stamps) != 2 {
		t.Fatalf("replays: got %d, want 2", len(stamps))
	}
	if gap := stamps[1].Sub(stamps[0]); gap < 30*time.Second {
		t.Errorf("gap between replays: got %s, want at least the 30s minimum interval", gap)
	}
}

// TestReplayRefusesASettledAttempt: nothing to recover, and no fallback to ask
// for. This is what lets a recovery sweep run the entry point over everything it
// finds without checking states itself.
func TestReplayRefusesASettledAttempt(t *testing.T) {
	f := newReplayFixture(t, execgw.ReplayConfig{})
	rec := f.inDoubtPlace(t)
	attempt, err := f.j.Resume(context.Background(), rec.ID)
	if err != nil {
		t.Fatalf("Resume: %v", err)
	}
	if err := attempt.ResolveConfirmed(context.Background(), "O-found",
		journal.ReasonResolvedFound, "found by observation"); err != nil {
		t.Fatalf("ResolveConfirmed: %v", err)
	}

	out, err := f.gw.ReplayInDoubt(context.Background(), rec.ID)
	if err != nil {
		t.Fatalf("ReplayInDoubt: %v", err)
	}
	if len(f.sentBodies()) != 0 {
		t.Fatalf("a settled attempt must not be replayed, sent %d", len(f.sentBodies()))
	}
	if out.State != journal.StateConfirmed || out.QueryFallback {
		t.Errorf("outcome: state=%s fallback=%v, want CONFIRMED and no fallback", out.State, out.QueryFallback)
	}
}

// TestReplayRefusesACancel: cancel and modify carry no clientOrderId (openapi
// puts it on OrderCreateRequest and nowhere else), so resending one would be a
// second cancel rather than a replay.
func TestReplayRefusesACancel(t *testing.T) {
	f := newReplayFixture(t, execgw.ReplayConfig{})
	intent := placeIntent()
	decision := exitDecision(t, f.j, f.clk, journal.KindCancel, intent.Market, intent.Symbol, "SELL", 2)
	out, err := f.gw.Cancel(context.Background(), execgw.CancelRequest{
		Intent: orderintent.CancelIntent{OrderID: "O-target", Symbol: intent.Symbol},
		Order: execgw.OrderRef{
			Market: intent.Market, Side: "SELL", Quantity: 2, Price: 70000, Currency: "KRW",
		},
		Decision: decision,
	})
	if err == nil || out.State != journal.StateInDoubt {
		t.Fatalf("setup: state=%s err=%v", out.State, err)
	}

	res, err := f.gw.ReplayInDoubt(context.Background(), out.AttemptID)
	if err != nil {
		t.Fatalf("ReplayInDoubt: %v", err)
	}
	if len(f.sentBodies()) != 0 {
		t.Fatalf("a cancel must not be replayed, sent %d", len(f.sentBodies()))
	}
	if res.Reason != execgw.ReasonReplayIneligible || !res.QueryFallback {
		t.Errorf("outcome: reason=%s fallback=%v, want replay_ineligible + fallback", res.Reason, res.QueryFallback)
	}
}

// TestReplayBodyCannotBeBuiltOutsideThePackage is the structural half of
// "저장된 wire body 외 전송 불가": the entry point takes an attempt id and
// nothing else, and the only value it can hand a transport has no exported way
// to be filled in.
func TestReplayBodyCannotBeBuiltOutsideThePackage(t *testing.T) {
	body := reflect.TypeOf(execgw.ReplayBody{})
	if body.NumField() != 1 {
		t.Fatalf("ReplayBody has %d fields; the type carries exactly the stored bytes", body.NumField())
	}
	if field := body.Field(0); field.PkgPath == "" {
		t.Errorf("ReplayBody.%s is exported, so any caller could compose a body to replay", field.Name)
	}
	if !(execgw.ReplayBody{}).Empty() {
		t.Error("the zero ReplayBody — the only one an outside package can make — must be empty")
	}

	method, ok := reflect.TypeOf(&execgw.Gateway{}).MethodByName("ReplayInDoubt")
	if !ok {
		t.Fatal("Gateway has no ReplayInDoubt method")
	}
	// receiver, context, attempt id — and nothing that could carry a body.
	if got := method.Type.NumIn(); got != 3 {
		t.Fatalf("ReplayInDoubt takes %d inputs; it must take only a context and an attempt id", got)
	}
	if got := method.Type.In(2); got.Kind() != reflect.String {
		t.Errorf("ReplayInDoubt's second input is %s; it must be the attempt id", got)
	}
}

// TestReplayUsesTheStoredBodyEvenAfterASerializerChange is the spec's
// "직렬화 규칙 변경 후 재생": the stored bytes are resent as they are, because
// rebuilding them under new rules is what produces an idempotency-key conflict.
func TestReplayUsesTheStoredBodyEvenAfterASerializerChange(t *testing.T) {
	f := newReplayFixture(t, execgw.ReplayConfig{})
	rec := f.inDoubtPlace(t)

	// Re-record the same decision's attempt with a body written by an older
	// serializer, which is what a binary upgrade leaves behind.
	const legacyBody = `{"symbol":"005930","side":"BUY","orderType":"LIMIT","quantity":"2.0","price":"70000.0"}`
	ref := entryDecision(t, f.j, f.clk, placeIntent(), testLimits())
	decision, err := f.j.LookupDecision(context.Background(), ref.ID)
	if err != nil {
		t.Fatalf("LookupDecision: %v", err)
	}
	intent, err := f.j.LookupIntent(context.Background(), rec.IntentID)
	if err != nil {
		t.Fatalf("LookupIntent: %v", err)
	}
	intent.ID = "legacy-intent"
	legacy, err := f.j.Prepare(context.Background(), journal.PrepareRequest{
		Intent:            intent,
		Kind:              journal.KindPlace,
		AttemptID:         "legacy-attempt",
		AccountRef:        decision.AccountRef,
		DecisionID:        decision.ID,
		SafetyClass:       decision.SafetyClass,
		Generation:        decision.Generation,
		ClientOrderID:     decision.ClientOrderID,
		WireBody:          legacyBody,
		SerializerVersion: "official/order-create/v0",
	})
	if err != nil {
		t.Fatalf("Prepare(legacy): %v", err)
	}
	if err := legacy.MarkDispatchStarted(context.Background()); err != nil {
		t.Fatalf("MarkDispatchStarted: %v", err)
	}
	if err := legacy.MarkInDoubt(context.Background(), "test", "outcome unknown"); err != nil {
		t.Fatalf("MarkInDoubt: %v", err)
	}
	f.respond(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(okBody("O-legacy", decision.ClientOrderID)))
	})

	out := replayAsync(t, f, "legacy-attempt", time.Second)
	bodies := f.sentBodies()
	if len(bodies) != 1 || bodies[0] != legacyBody {
		t.Fatalf("replayed body: got %v, want the stored legacy bytes", bodies)
	}
	if out.State != journal.StateConfirmed {
		t.Errorf("state: got %s, want CONFIRMED (%s)", out.State, out.Detail)
	}
}

// --- 2.2: the response rules ------------------------------------------------

// TestReplay409DoesNotConsumeTheCap is the openapi 409 rule: "동일 주문 키에
// 대해 처리 중인 요청이 있습니다" means the original request is still being
// processed, so the replay learned nothing and must not spend the allowance.
func TestReplay409DoesNotConsumeTheCap(t *testing.T) {
	f := newReplayFixture(t, execgw.ReplayConfig{})
	rec := f.inDoubtPlace(t)

	var calls int
	f.respond(func(w http.ResponseWriter, _ *http.Request) {
		f.mu.Lock()
		calls++
		n := calls
		f.mu.Unlock()
		if n == 1 {
			w.WriteHeader(http.StatusConflict)
			_, _ = w.Write([]byte(errorBody("request-in-progress",
				"동일 주문 키에 대해 처리 중인 요청이 있습니다. 잠시 후 다시 시도해 주세요.")))
			return
		}
		_, _ = w.Write([]byte(okBody("O-after-409", rec.ClientOrderID)))
	})

	out := replayAsync(t, f, rec.ID, execgw.DefaultReplayMinInterval)

	if len(f.sentBodies()) != 2 {
		t.Fatalf("replay requests: got %d, want 2 (the 409 then the answer)", len(f.sentBodies()))
	}
	if out.State != journal.StateConfirmed || out.BrokerOrderID != "O-after-409" {
		t.Fatalf("outcome: state=%s id=%q (%s)", out.State, out.BrokerOrderID, out.Detail)
	}
	stored, err := f.j.LookupAttempt(context.Background(), rec.ID)
	if err != nil {
		t.Fatalf("LookupAttempt: %v", err)
	}
	if stored.ReplayCount != 1 {
		t.Errorf("replay count: got %d, want 1 — the 409 was refunded", stored.ReplayCount)
	}
}

// TestReplay422KeyConflictParksAndNeverFails is the rule the dispatch classifier
// would get wrong: it treats 422 as a definitive rejection, but a replay's
// `422 idempotency-key-conflict` says nothing at all about the original order.
func TestReplay422KeyConflictParksAndNeverFails(t *testing.T) {
	f := newReplayFixture(t, execgw.ReplayConfig{})
	rec := f.inDoubtPlace(t)
	f.respond(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnprocessableEntity)
		_, _ = w.Write([]byte(errorBody("idempotency-key-conflict",
			"동일한 clientOrderId 로 다른 내용의 주문을 요청할 수 없습니다.")))
	})

	out := replayAsync(t, f, rec.ID, time.Second)

	if out.State == journal.StateFailedConfirmed {
		t.Fatal("a replay's 422 must never settle the attempt as FAILED_CONFIRMED")
	}
	if out.State != journal.StateUnresolvedInDoubt {
		t.Fatalf("state: got %s, want UNRESOLVED_IN_DOUBT (%s)", out.State, out.Detail)
	}
	if out.Reason != execgw.ReasonReplayKeyConflict {
		t.Errorf("reason: got %s, want replay_idempotency_key_conflict", out.Reason)
	}
	stored, err := f.j.LookupAttempt(context.Background(), rec.ID)
	if err != nil {
		t.Fatalf("LookupAttempt: %v", err)
	}
	if stored.State != journal.StateUnresolvedInDoubt ||
		stored.ReasonCode != journal.ReasonReplayKeyConflict {
		t.Errorf("journal: state=%s reason=%s", stored.State, stored.ReasonCode)
	}
	if f.gate.CheckEntry() == nil {
		t.Error("a parked attempt must latch the entry gate")
	}
}

// TestReplayKeyConflictEnqueuesTheCriticalAlert pins the durable half of the
// alert path, and pins the event string against internal/obs's own constant —
// obs imports execgw, so the constant cannot be imported the other way.
func TestReplayKeyConflictEnqueuesTheCriticalAlert(t *testing.T) {
	f := newReplayFixture(t, execgw.ReplayConfig{})
	rec := f.inDoubtPlace(t)
	f.respond(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnprocessableEntity)
		_, _ = w.Write([]byte(errorBody("idempotency-key-conflict", "conflict")))
	})

	replayAsync(t, f, rec.ID, time.Second)

	pending, err := f.j.PendingAlerts(context.Background(), 0)
	if err != nil {
		t.Fatalf("PendingAlerts: %v", err)
	}
	if len(pending) != 1 {
		t.Fatalf("outbox rows: got %d, want exactly 1", len(pending))
	}
	alert := pending[0]
	if alert.Type != string(obs.EventOrderUnresolved) {
		t.Errorf("event type: got %q, want %q", alert.Type, obs.EventOrderUnresolved)
	}
	if obs.SeverityOf(obs.EventType(alert.Type)) != obs.SeverityCritical {
		t.Errorf("event %q is not graded critical", alert.Type)
	}
	if alert.Severity != string(obs.SeverityCritical) {
		t.Errorf("severity: got %q, want %q", alert.Severity, obs.SeverityCritical)
	}
	if alert.Payload == "" {
		t.Error("the alert carries no operator context")
	}
}

// TestReplayEchoingADifferentKeyParks: the echo is documented to come back
// "요청 시 전달한 값 그대로", so a different key describes somebody else's order.
func TestReplayEchoingADifferentKeyParks(t *testing.T) {
	f := newReplayFixture(t, execgw.ReplayConfig{})
	rec := f.inDoubtPlace(t)
	f.respond(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(okBody("O-someone-else", "a-different-key")))
	})

	out := replayAsync(t, f, rec.ID, time.Second)
	if out.State != journal.StateUnresolvedInDoubt {
		t.Fatalf("state: got %s, want UNRESOLVED_IN_DOUBT (%s)", out.State, out.Detail)
	}
	stored, err := f.j.LookupAttempt(context.Background(), rec.ID)
	if err != nil {
		t.Fatalf("LookupAttempt: %v", err)
	}
	if stored.BrokerOrderID == "O-someone-else" {
		t.Error("an order id from a response about a different key must not be attached to this attempt")
	}
}

// TestReplayOther422IsNotAFailure: only the key conflict is named by the spec.
// Every other 422 describes the *replay*, so it settles nothing either — and in
// particular it does not become FAILED_CONFIRMED the way the dispatch classifier
// would make it.
func TestReplayOther422IsNotAFailure(t *testing.T) {
	f := newReplayFixture(t, execgw.ReplayConfig{})
	rec := f.inDoubtPlace(t)
	f.respond(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnprocessableEntity)
		_, _ = w.Write([]byte(errorBody("insufficient-buying-power", "주문 가능 금액이 부족합니다.")))
	})

	out := replayAsync(t, f, rec.ID, execgw.DefaultReplayMinInterval)
	if out.State == journal.StateFailedConfirmed || out.State == journal.StateUnresolvedInDoubt {
		t.Fatalf("state: got %s, want the attempt left IN_DOUBT for the fallback", out.State)
	}
	if !out.QueryFallback || out.Reason != execgw.ReasonReplayExhausted {
		t.Errorf("outcome: reason=%s fallback=%v, want the cap spent and the fallback asked for",
			out.Reason, out.QueryFallback)
	}
}

// TestReplayWithNoAnswerCountsAndFallsBack: a lost response is exactly what put
// the attempt in doubt in the first place. It is counted, and when the cap is
// spent the query fallback takes over.
func TestReplayWithNoAnswerCountsAndFallsBack(t *testing.T) {
	f := newReplayFixture(t, execgw.ReplayConfig{})
	rec := f.inDoubtPlace(t)
	f.srv.Close() // nothing is listening: the replay gets no answer at all

	out := replayAsync(t, f, rec.ID, execgw.DefaultReplayMinInterval)
	if out.Sent != execgw.DefaultMaxReplays {
		t.Errorf("replays sent: got %d, want %d", out.Sent, execgw.DefaultMaxReplays)
	}
	if !out.QueryFallback || out.Reason != execgw.ReasonReplayExhausted {
		t.Errorf("outcome: reason=%s fallback=%v", out.Reason, out.QueryFallback)
	}
	stored, err := f.j.LookupAttempt(context.Background(), rec.ID)
	if err != nil {
		t.Fatalf("LookupAttempt: %v", err)
	}
	if stored.State != journal.StateInDoubt || stored.ReplayCount != execgw.DefaultMaxReplays {
		t.Errorf("journal: state=%s replayCount=%d", stored.State, stored.ReplayCount)
	}
}

// TestReplayStopsWaitingOnAPermanent409: 409 does not consume the cap, so the
// wait bound is what stops a broker that says "in progress" forever.
func TestReplayStopsWaitingOnAPermanent409(t *testing.T) {
	f := newReplayFixture(t, execgw.ReplayConfig{MaxWaits: 2})
	rec := f.inDoubtPlace(t)
	f.respond(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusConflict)
		_, _ = w.Write([]byte(errorBody("request-in-progress", "처리 중")))
	})

	out := replayAsync(t, f, rec.ID, execgw.DefaultReplayMinInterval)
	if !out.QueryFallback || out.Reason != execgw.ReasonReplayExhausted {
		t.Errorf("outcome: reason=%s fallback=%v, want the wait bound to hand over to the fallback",
			out.Reason, out.QueryFallback)
	}
	stored, err := f.j.LookupAttempt(context.Background(), rec.ID)
	if err != nil {
		t.Fatalf("LookupAttempt: %v", err)
	}
	if stored.ReplayCount != 0 {
		t.Errorf("replay count: got %d, want 0 — every 409 was refunded", stored.ReplayCount)
	}
	if stored.State != journal.StateInDoubt {
		t.Errorf("state: got %s, want IN_DOUBT", stored.State)
	}
}
