package execgw_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/domain"
	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/official"
	"github.com/JungHoonGhae/tossinvest-cli/internal/orderintent"
	"github.com/JungHoonGhae/tossinvest-cli/internal/trading"
)

func acceptedResult(orderID string) domain.MutationResult {
	return domain.MutationResult{Kind: "place", Status: "accepted", OrderID: orderID}
}

func cancelIntentFixture() orderintent.CancelIntent {
	return orderintent.CancelIntent{OrderID: "O-1", Symbol: "005930"}
}

// Retry-matrix tests (harden-execution-base task 2.6). The table these pin lives
// in openspec/changes/harden-execution-base/retry-matrix.md; the two files are
// meant to be read together.

// --- classification ---------------------------------------------------------

func TestClassifyQueryError(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want execgw.ErrorClass
	}{
		{"nil", nil, execgw.ClassOK},
		{"transport", official.ErrTransport, execgw.ClassTransient},
		{"server 5xx", official.ErrServer, execgw.ClassTransient},
		{"rate limited", official.ErrRateLimited, execgw.ClassRateLimited},
		{"auth", official.ErrAuth, execgw.ClassAuthFatal},
		{"ip not allowed", official.ErrIPNotAllowed, execgw.ClassAuthFatal},
		// Wrapped since a082 put the status code on the message. The classifier
		// latches the entry gate off this verdict, so the wrapped shape has to be
		// in the table rather than assumed to behave like the bare one.
		{"auth carrying its status code", fmt.Errorf("%w (HTTP 401)", official.ErrAuth), execgw.ClassAuthFatal},
		{"ip not allowed carrying its status code", fmt.Errorf("%w (HTTP 403)", official.ErrIPNotAllowed), execgw.ClassAuthFatal},
		{"bad request", &official.APIError{Code: 400, Body: "nope"}, execgw.ClassPermanent},
		{"not found", &official.APIError{Code: 404}, execgw.ClassPermanent},
		{"context canceled", context.Canceled, execgw.ClassCanceled},
		{"deadline", context.DeadlineExceeded, execgw.ClassCanceled},
		{"unknown", errors.New("surprise"), execgw.ClassTransient},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := execgw.ClassifyQueryError(tc.err); got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
	}
}

// --- query retries ----------------------------------------------------------

// newRetrier builds a retrier on a fake clock with deterministic jitter, so the
// waits asserted below are exact rather than approximate.
func newRetrier(t *testing.T, gate *execgw.EntryGate, src execgw.RetryAfterSource) (*execgw.Retrier, *clock.Fake) {
	t.Helper()
	clk := clock.NewFake(fixedNow)
	policy := execgw.DefaultRetryPolicy()
	policy.Rand = func() float64 { return 0.5 } // 0.5 → no jitter offset
	return &execgw.Retrier{
		Policy:     policy,
		Clock:      clk,
		Gate:       gate,
		RetryAfter: src,
	}, clk
}

// runAsync drives fn on a goroutine and advances the fake clock by exactly the
// backoff the policy asks for, so retries happen without wall-clock time and
// without accidentally spending the total-time budget.
func runAsync(t *testing.T, clk *clock.Fake, waits []time.Duration, fn func() error) error {
	t.Helper()
	done := make(chan error, 1)
	go func() { done <- fn() }()
	for i, wait := range waits {
		if !clk.WaitForSleepers(1, 2*time.Second) {
			t.Fatalf("retrier never slept (iteration %d)", i)
		}
		clk.Advance(wait)
	}
	select {
	case err := <-done:
		return err
	case <-time.After(3 * time.Second):
		t.Fatal("retrier did not finish")
		return nil
	}
}

func TestQueryRetriesTransientThenSucceeds(t *testing.T) {
	r, clk := newRetrier(t, nil, nil)
	calls := 0
	err := runAsync(t, clk, []time.Duration{400 * time.Millisecond}, func() error {
		return r.Query(context.Background(), execgw.QueryBuyingPower, func(context.Context) error {
			calls++
			if calls == 1 {
				return official.ErrTransport
			}
			return nil
		})
	})
	if err != nil {
		t.Fatalf("Query: %v", err)
	}
	if calls != 2 {
		t.Errorf("calls: got %d, want 2 (one retry)", calls)
	}
}

func TestQueryStopsAtTheAttemptBudget(t *testing.T) {
	r, clk := newRetrier(t, nil, nil)
	calls := 0
	err := runAsync(t, clk, []time.Duration{400 * time.Millisecond, 800 * time.Millisecond}, func() error {
		return r.Query(context.Background(), execgw.QueryBuyingPower, func(context.Context) error {
			calls++
			return official.ErrServer
		})
	})
	if err == nil {
		t.Fatal("an exhausted budget must surface the error")
	}
	if calls != execgw.DefaultRetryPolicy().MaxAttempts {
		t.Errorf("calls: got %d, want %d", calls, execgw.DefaultRetryPolicy().MaxAttempts)
	}
}

func TestQueryDoesNotRetryPermanentOrAuth(t *testing.T) {
	for _, tc := range []struct {
		name string
		err  error
	}{
		{"permanent 400", &official.APIError{Code: 400}},
		{"auth 401", official.ErrAuth},
	} {
		t.Run(tc.name, func(t *testing.T) {
			gate := execgw.NewEntryGate(clock.NewFake(fixedNow), nil)
			r, _ := newRetrier(t, gate, nil)
			calls := 0
			err := r.Query(context.Background(), execgw.QueryOpenOrders, func(context.Context) error {
				calls++
				return tc.err
			})
			if err == nil {
				t.Fatal("want the error to surface")
			}
			if calls != 1 {
				t.Errorf("calls: got %d, want exactly 1 (no retry)", calls)
			}
		})
	}
}

// TestAuthFailureLatchesEntryImmediately is the spec's "401/403은 즉시 신규 진입
// 차단". The block must survive a later successful query — only credential
// recovery plus an explicit clear reopens it.
func TestAuthFailureLatchesEntryImmediately(t *testing.T) {
	clk := clock.NewFake(fixedNow)
	gate := execgw.NewEntryGate(clk, map[execgw.RequiredQuery]time.Duration{})
	r, _ := newRetrier(t, gate, nil)

	if err := gate.CheckEntry(); err != nil {
		t.Fatalf("gate with no required queries must start open: %v", err)
	}
	_ = r.Query(context.Background(), execgw.QueryOpenOrders, func(context.Context) error {
		return official.ErrAuth
	})

	blocked := gate.CheckEntry()
	if blocked == nil {
		t.Fatal("a 401 must block new entries immediately")
	}
	if blocked.Reason != execgw.ReasonBrokerAuthRejected {
		t.Errorf("reason: got %q, want %q", blocked.Reason, execgw.ReasonBrokerAuthRejected)
	}

	// A later success must NOT quietly reopen it.
	if err := r.Query(context.Background(), execgw.QueryOpenOrders, func(context.Context) error {
		return nil
	}); err != nil {
		t.Fatalf("recovery query: %v", err)
	}
	if gate.CheckEntry() == nil {
		t.Error("an auth latch must not clear itself on the next successful poll")
	}
	gate.Clear(execgw.ReasonBrokerAuthRejected)
	if err := gate.CheckEntry(); err != nil {
		t.Errorf("an explicitly cleared latch must reopen the gate: %v", err)
	}
}

// TestRetryAfterIsHonouredUpToTheCap covers both halves of the 429 rule: a header
// inside the cap is waited out exactly, and one beyond the cap aborts the query
// instead of parking the engine for an hour.
func TestRetryAfterIsHonouredUpToTheCap(t *testing.T) {
	t.Run("within the cap", func(t *testing.T) {
		src := &stubRetryAfter{d: 2 * time.Second}
		r, clk := newRetrier(t, nil, src)
		calls := 0
		start := clk.Now()

		done := make(chan error, 1)
		go func() {
			done <- r.Query(context.Background(), execgw.QueryPrice, func(context.Context) error {
				calls++
				if calls == 1 {
					return official.ErrRateLimited
				}
				return nil
			})
		}()
		if !clk.WaitForSleepers(1, 2*time.Second) {
			t.Fatal("the retrier never waited for Retry-After")
		}
		clk.Advance(2 * time.Second)
		if err := <-done; err != nil {
			t.Fatalf("Query: %v", err)
		}
		if waited := clk.Now().Sub(start); waited != 2*time.Second {
			t.Errorf("waited %s, want exactly the 2s Retry-After", waited)
		}
		if calls != 2 {
			t.Errorf("calls: got %d, want 2", calls)
		}
	})

	t.Run("beyond the cap", func(t *testing.T) {
		src := &stubRetryAfter{d: time.Hour}
		r, _ := newRetrier(t, nil, src)
		calls := 0
		err := r.Query(context.Background(), execgw.QueryPrice, func(context.Context) error {
			calls++
			return official.ErrRateLimited
		})
		if err == nil {
			t.Fatal("a Retry-After beyond the cap must abort the query")
		}
		if calls != 1 {
			t.Errorf("calls: got %d, want 1 — waiting past the cap is worse than failing", calls)
		}
	})
}

type stubRetryAfter struct {
	mu sync.Mutex
	d  time.Duration
}

func (s *stubRetryAfter) RetryAfter() (time.Duration, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.d <= 0 {
		return 0, false
	}
	return s.d, true
}

// TestRetryAfterTransportCapturesTheHeader proves the header actually reaches the
// policy through a real round trip, in both the seconds and the HTTP-date form.
func TestRetryAfterTransportCapturesTheHeader(t *testing.T) {
	clk := clock.NewFake(fixedNow)
	for _, tc := range []struct {
		name   string
		header string
		want   time.Duration
	}{
		{"seconds", "7", 7 * time.Second},
		{"http date", fixedNow.Add(12 * time.Second).UTC().Format(http.TimeFormat), 12 * time.Second},
		{"garbage", "soon", 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Retry-After", tc.header)
				w.WriteHeader(http.StatusTooManyRequests)
			}))
			t.Cleanup(srv.Close)

			rt := execgw.NewRetryAfterTransport(srv.Client().Transport, clk)
			hc := &http.Client{Transport: rt}
			resp, err := hc.Get(srv.URL)
			if err != nil {
				t.Fatalf("GET: %v", err)
			}
			resp.Body.Close()

			got, ok := rt.RetryAfter()
			if tc.want == 0 {
				if ok {
					t.Errorf("an unparseable header must not produce a delay, got %s", got)
				}
				return
			}
			if !ok || got != tc.want {
				t.Errorf("RetryAfter: got %s (ok=%v), want %s", got, ok, tc.want)
			}
			// The value is consumed: a stale Retry-After must not govern a later call.
			if _, again := rt.RetryAfter(); again {
				t.Error("Retry-After must be consumed on read")
			}
		})
	}
}

// --- staleness --------------------------------------------------------------

// TestStalenessBlocksAndAutoClears is the spec's "필수 조회 장기 실패 → 신규 진입
// 차단, 조회 복구 후 자동 해제", plus the fail-closed opening position: a query
// that never succeeded is infinitely stale.
func TestStalenessBlocksAndAutoClears(t *testing.T) {
	clk := clock.NewFake(fixedNow)
	gate := execgw.NewEntryGate(clk, map[execgw.RequiredQuery]time.Duration{
		execgw.QueryBuyingPower: 45 * time.Second,
	})

	blocked := gate.CheckEntry()
	if blocked == nil || blocked.Reason != execgw.ReasonQueryStale {
		t.Fatalf("a never-observed required query must block entry, got %v", blocked)
	}

	gate.RecordSuccess(execgw.QueryBuyingPower)
	if err := gate.CheckEntry(); err != nil {
		t.Fatalf("a fresh query must open the gate: %v", err)
	}

	clk.Advance(44 * time.Second)
	if err := gate.CheckEntry(); err != nil {
		t.Fatalf("inside the threshold the gate stays open: %v", err)
	}
	clk.Advance(2 * time.Second)
	if blocked := gate.CheckEntry(); blocked == nil || blocked.Reason != execgw.ReasonQueryStale {
		t.Fatalf("past the threshold entry must be blocked, got %v", blocked)
	}

	gate.RecordSuccess(execgw.QueryBuyingPower)
	if err := gate.CheckEntry(); err != nil {
		t.Errorf("recovery must auto-clear the staleness block: %v", err)
	}
}

// TestGatewayBlocksNewEntryButNotExits: the block is on new exposure only. A
// cancel — and a sell — must still get through, or a rate-limit blip would trap
// the engine in a position it cannot leave (§0.3).
func TestGatewayBlocksNewEntryButNotExits(t *testing.T) {
	broker := &fakeBroker{result: acceptedResult("O-1")}
	clk := clock.NewFake(fixedNow)
	j := openJournal(t, clk)
	gate := execgw.NewEntryGate(clk, map[execgw.RequiredQuery]time.Duration{
		execgw.QueryBuyingPower: 45 * time.Second,
	})
	gw, err := execgw.New(execgw.Options{
		Journal: j, Trading: trading.NewService(openPolicy(), broker), Clock: clk,
		AccountRef: "acct-7", Source: "test", Entry: gate,
	})
	if err != nil {
		t.Fatalf("execgw.New: %v", err)
	}
	ctx := context.Background()

	// buying power was never observed → new entries are blocked.
	out, err := gw.Place(ctx, placeRequest(t, j, clk))
	var rejected *execgw.RejectedError
	if !errors.As(err, &rejected) || rejected.Reason != execgw.ReasonQueryStale {
		t.Fatalf("want a stale-query refusal, got %v", err)
	}
	if out.State != journal.StateNotDispatched {
		t.Errorf("a blocked entry must still be journalled as NOT_DISPATCHED, got %s", out.State)
	}
	if places, _, _ := broker.totals(); places != 0 {
		t.Errorf("broker place calls: got %d, want 0", places)
	}

	// the exit path stays open.
	cancelIntent := cancelIntentFixture()
	if _, err := gw.Cancel(ctx, execgw.CancelRequest{
		Intent:   cancelIntent,
		Order:    execgw.OrderRef{Market: "kr", Side: "BUY", Quantity: 2, Price: 70000, Currency: "KRW"},
		Decision: exitDecision(t, j, clk, journal.KindCancel, "kr", "005930", "BUY", 2),
	}); err != nil {
		t.Errorf("a blocked entry gate must not block a cancel: %v", err)
	}

	gate.RecordSuccess(execgw.QueryBuyingPower)
	if _, err := gw.Place(ctx, placeRequest(t, j, clk)); err != nil {
		t.Errorf("entry must resume once the required query is fresh: %v", err)
	}
}

// TestMutationsAreNeverRetried is §0.1 of the matrix, end to end: a 429 on a
// place produces exactly one POST and an IN_DOUBT attempt. A retry here would be
// a duplicated live order.
func TestMutationsAreNeverRetried(t *testing.T) {
	var posts int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/oauth2/token" {
			_, _ = w.Write([]byte(`{"access_token":"AT","expires_in":3600,"token_type":"Bearer"}`))
			return
		}
		posts++
		w.Header().Set("Retry-After", "1")
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = w.Write([]byte(`{"message":"slow down"}`))
	}))
	t.Cleanup(srv.Close)

	clk := clock.NewFake(fixedNow)
	off := official.New(official.Credentials{APIKey: "k", SecretKey: "s"},
		filepath.Join(t.TempDir(), "token.json"),
		official.WithBaseURL(srv.URL),
		official.WithHTTPClient(&http.Client{Transport: execgw.NewRetryAfterTransport(srv.Client().Transport, clk)}),
		official.WithAccountSeq(7))

	j := openJournal(t, clk)
	gw, err := execgw.New(execgw.Options{
		Journal: j, Trading: trading.NewService(openPolicy(), &officialTestBroker{off: off}),
		Clock: clk, AccountRef: "acct-7", Source: "test",
	})
	if err != nil {
		t.Fatalf("execgw.New: %v", err)
	}

	out, _ := gw.Place(context.Background(), placeRequest(t, j, clk))
	if out.State != journal.StateInDoubt {
		t.Errorf("state: got %s, want IN_DOUBT", out.State)
	}
	if posts != 1 {
		t.Errorf("mutation POSTs: got %d, want exactly 1", posts)
	}
}

// TestDefaultsMatchThePublishedMatrix keeps the numbers in retry-matrix.md and the
// numbers in the code from drifting apart. If this fails, change both.
func TestDefaultsMatchThePublishedMatrix(t *testing.T) {
	p := execgw.DefaultRetryPolicy()
	if p.MaxAttempts != 3 {
		t.Errorf("MaxAttempts: got %d, want 3", p.MaxAttempts)
	}
	if p.Budget != 8*time.Second {
		t.Errorf("Budget: got %s, want 8s", p.Budget)
	}
	if p.BaseBackoff != 400*time.Millisecond {
		t.Errorf("BaseBackoff: got %s, want 400ms", p.BaseBackoff)
	}
	if p.MaxBackoff != 3*time.Second {
		t.Errorf("MaxBackoff: got %s, want 3s", p.MaxBackoff)
	}
	if p.MaxRetryAfter != 30*time.Second {
		t.Errorf("MaxRetryAfter: got %s, want 30s", p.MaxRetryAfter)
	}
	if p.JitterFraction != 0.25 {
		t.Errorf("JitterFraction: got %v, want 0.25", p.JitterFraction)
	}

	want := map[execgw.RequiredQuery]time.Duration{
		execgw.QueryOpenOrders:  20 * time.Second,
		execgw.QueryBuyingPower: 45 * time.Second,
		execgw.QueryHoldings:    60 * time.Second,
		execgw.QueryPrice:       15 * time.Second,
	}
	got := execgw.DefaultStaleness()
	if len(got) != len(want) {
		t.Fatalf("staleness thresholds: got %v, want %v", got, want)
	}
	for k, v := range want {
		if got[k] != v {
			t.Errorf("staleness[%s]: got %s, want %s", k, got[k], v)
		}
	}
}

// TestJitterStaysBounded: jitter exists to desynchronise retries, not to invent
// unbounded waits.
func TestJitterStaysBounded(t *testing.T) {
	p := execgw.DefaultRetryPolicy()
	base := time.Second
	for _, r := range []float64{0, 0.25, 0.5, 0.75, 1} {
		p.Rand = func() float64 { return r }
		d := p.BackoffFor(1, base)
		lo := time.Duration(float64(base) * (1 - p.JitterFraction))
		hi := time.Duration(float64(base) * (1 + p.JitterFraction))
		if d < lo || d > hi {
			t.Errorf("rand=%v produced %s, outside [%s, %s]", r, d, lo, hi)
		}
	}
}
