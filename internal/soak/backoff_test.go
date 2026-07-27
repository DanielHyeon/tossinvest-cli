package soak_test

// backoff_test.go covers the one thing the survey retries: a read the broker
// answered 429 (task 1.11, measurements.md M8).
//
// The survey and the live verification share one account and one rate budget.
// On 2026-07-27 a clean cycle — fifteen minutes after the previous one, every
// other endpoint OK — still lost GET /api/v1/orders to a 429, because the CLOSED
// walk paged the account's whole history at the API's default page size and the
// order-by-id read that follows landed inside the same penalty window. A cycle
// that gives up there records a failed endpoint for a condition that clears in
// seconds, and an endpoint that never succeeds is an attestation that can never
// be written.
//
// So: two more attempts, fifteen then thirty seconds apart, on 429 and on
// nothing else. Every attempt is counted as a request, because it cost the
// account's rate budget whether it worked or not.
//
// Nothing here sleeps for real. The runner waits on soak.Options.Clock, the tests
// drive a clock.Fake, and a test that forgot to advance it would hang rather than
// pass on wall time.

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/soak"
)

// errThrottled stands in for official.ErrRateLimited. internal/soak cannot name
// that sentinel — the package is kept out of internal/official's import graph by
// static_test.go — so it learns what a 429 is the same way the real wiring tells
// it: through Options.Classify (cmd/tossctl.classifySoakError).
var errThrottled = errors.New("official: 429 rate limited")

// errRefused is any other failure. It must not be retried: a 400 answered twice
// is a 400, and spending the rate budget to prove it is the opposite of the
// point.
var errRefused = errors.New(`official: 400 invalid-request {"field":"status"}`)

func classifyThrottled(err error) soak.Class {
	switch {
	case err == nil:
		return soak.ClassOK
	case errors.Is(err, errThrottled):
		return soak.ClassRateLimited
	default:
		return soak.ClassOther
	}
}

// throttlingReads answers the first n calls to one read with a 429 and then
// hands over to the stub underneath.
type throttlingReads struct {
	soak.Reads

	mu sync.Mutex

	// pageFailures and orderFailures are how many more times OrdersPage and
	// Order answer with err instead of reading. They count down.
	pageFailures  int
	orderFailures int
	// accountFailures does the same for the credential probe, which must NOT
	// gain a retry: the retry is scoped to the order reads M8 measured.
	accountFailures int

	// err is what a throttled call returns.
	err error

	pageCalls    int
	orderCalls   int
	accountCalls int
}

func throttling(r soak.Reads, err error) *throttlingReads {
	return &throttlingReads{Reads: r, err: err}
}

// take reports whether this call is one of the ones that fails, decrementing the
// budget when it is.
func (s *throttlingReads) take(remaining *int, calls *int) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	*calls++
	if *remaining <= 0 {
		return false
	}
	*remaining--
	return true
}

func (s *throttlingReads) OrdersPage(ctx context.Context, status, cursor string) (soak.OrderPage, error) {
	if s.take(&s.pageFailures, &s.pageCalls) {
		return soak.OrderPage{}, s.err
	}
	return s.Reads.OrdersPage(ctx, status, cursor)
}

func (s *throttlingReads) Order(ctx context.Context, id string) error {
	if s.take(&s.orderFailures, &s.orderCalls) {
		return s.err
	}
	return s.Reads.Order(ctx, id)
}

func (s *throttlingReads) Accounts(ctx context.Context) ([]string, error) {
	if s.take(&s.accountFailures, &s.accountCalls) {
		return nil, s.err
	}
	return s.Reads.Accounts(ctx)
}

func (s *throttlingReads) counts() (pages, orders, accounts int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.pageCalls, s.orderCalls, s.accountCalls
}

// runCycleAdvancingBackoffs runs one cycle and moves the fake clock past every
// wait the runner parks on, so a backoff costs the test microseconds instead of
// forty-five seconds.
//
// The poll is on the fake's sleeper count, which is the same device
// clock.Fake.WaitForSleepers uses: advancing before the code under test has
// reached its sleep is the classic flake, and advancing after it has is
// deterministic.
func runCycleAdvancingBackoffs(t *testing.T, r *soak.Runner, fake *clock.Fake) (soak.Cycle, error) {
	t.Helper()

	type outcome struct {
		cycle soak.Cycle
		err   error
	}
	done := make(chan outcome, 1)
	go func() {
		c, err := r.RunCycle(context.Background())
		done <- outcome{c, err}
	}()

	deadline := time.Now().Add(10 * time.Second)
	for {
		select {
		case o := <-done:
			return o.cycle, o.err
		case <-time.After(time.Millisecond):
		}
		if fake.Sleepers() > 0 {
			// Longer than the longest backoff, so one advance releases any of them.
			fake.Advance(time.Minute)
		}
		if time.Now().After(deadline) {
			t.Fatal("the cycle never finished; the runner is parked on a wait the test cannot see")
		}
	}
}

// TestAThrottledOrderWalkIsRetriedAndSucceeds is the RED case from M8: the
// broker answers the order list with a 429 and then, seconds later, answers it
// normally. Before task 1.11 the walk gave up on the first one and recorded a
// failed endpoint, which is how a clean 41-cycle run ended with GET
// /api/v1/orders never once successful.
func TestAThrottledOrderWalkIsRetriedAndSucceeds(t *testing.T) {
	reads := throttling(newStubReads(), errThrottled)
	reads.pageFailures = 2 // both extra attempts are needed

	r, fake := newRunner(t, reads, func(o *soak.Options) {
		o.Classify = classifyThrottled
	})

	c, err := runCycleAdvancingBackoffs(t, r, fake)
	if err != nil {
		t.Fatalf("RunCycle: %v", err)
	}

	orders := endpointResult(t, c, soak.EndpointOrders)
	if !orders.OK {
		t.Errorf("%s: OK = false (%s / %s) — a 429 that clears on retry must not fail the endpoint",
			soak.EndpointOrders, orders.Class, orders.Error)
	}
	if orders.Class != soak.ClassOK {
		t.Errorf("%s: Class = %q, want %q — rate_limited is for a walk that never got through",
			soak.EndpointOrders, orders.Class, soak.ClassOK)
	}
	// Three attempts on the CLOSED first page, one on the OPEN page.
	if want := 4; orders.Requests != want {
		t.Errorf("%s: Requests = %d, want %d — every attempt spent the account's rate budget "+
			"and has to be counted", soak.EndpointOrders, orders.Requests, want)
	}
	if pages, _, _ := reads.counts(); pages != 4 {
		t.Errorf("the walk made %d page call(s), want 4", pages)
	}
	if !c.Completeness.Evaluated || !c.Completeness.OK {
		t.Errorf("Completeness = %+v, want an evaluated pass once the walk got through", c.Completeness)
	}
}

// TestAThrottledOrderByIDReadIsRetried. The order-by-id read is the second half
// of M8: it goes out immediately after the walk, so the same penalty window
// catches it.
func TestAThrottledOrderByIDReadIsRetried(t *testing.T) {
	reads := throttling(newStubReads(), errThrottled)
	reads.orderFailures = 1

	r, fake := newRunner(t, reads, func(o *soak.Options) {
		o.Classify = classifyThrottled
	})

	c, err := runCycleAdvancingBackoffs(t, r, fake)
	if err != nil {
		t.Fatalf("RunCycle: %v", err)
	}

	byID := endpointResult(t, c, soak.EndpointOrderByID)
	if !byID.OK {
		t.Errorf("%s: OK = false (%s / %s)", soak.EndpointOrderByID, byID.Class, byID.Error)
	}
	if byID.Class != soak.ClassOK {
		t.Errorf("%s: Class = %q, want %q", soak.EndpointOrderByID, byID.Class, soak.ClassOK)
	}
	if want := 2; byID.Requests != want {
		t.Errorf("%s: Requests = %d, want %d", soak.EndpointOrderByID, byID.Requests, want)
	}
	if _, orders, _ := reads.counts(); orders != 2 {
		t.Errorf("the order-by-id probe made %d call(s), want 2", orders)
	}
}

// TestAWalkThrottledOnEveryAttemptIsRecordedAsRateLimited. The retry is finite,
// and when it runs out the record has to say plainly that the broker refused —
// `soak status` reports the throttling as the observed rate ceiling, and a
// swallowed 429 would take that measurement away.
func TestAWalkThrottledOnEveryAttemptIsRecordedAsRateLimited(t *testing.T) {
	reads := throttling(newStubReads(), errThrottled)
	reads.pageFailures = 3 // one more than the retry can absorb

	var progress progressBuffer
	r, fake := newRunner(t, reads, func(o *soak.Options) {
		o.Classify = classifyThrottled
		o.Progress = &progress
	})

	c, err := runCycleAdvancingBackoffs(t, r, fake)
	if err != nil {
		t.Fatalf("RunCycle: %v", err)
	}

	orders := endpointResult(t, c, soak.EndpointOrders)
	if orders.OK {
		t.Errorf("%s: OK = true although every attempt was refused", soak.EndpointOrders)
	}
	if orders.Class != soak.ClassRateLimited {
		t.Errorf("%s: Class = %q, want %q", soak.EndpointOrders, orders.Class, soak.ClassRateLimited)
	}
	// Three attempts on the CLOSED page, then one on the OPEN page, which the
	// stub answers because the failure budget is spent.
	if want := 4; orders.Requests != want {
		t.Errorf("%s: Requests = %d, want %d", soak.EndpointOrders, orders.Requests, want)
	}
	if !strings.Contains(progress.String(), "429") {
		t.Errorf("the survey retried without telling the operator why:\n%s", progress.String())
	}
}

// TestAReadRefusedForAnyOtherReasonIsNotRetried. The backoff is scoped to the
// one failure that clears by waiting. Retrying a 400 would spend the rate budget
// to be told the same thing three times — and it is the rate budget that is
// scarce.
func TestAReadRefusedForAnyOtherReasonIsNotRetried(t *testing.T) {
	reads := throttling(newStubReads(), errRefused)
	reads.pageFailures = 1

	r, fake := newRunner(t, reads, func(o *soak.Options) {
		o.Classify = classifyThrottled
	})

	c, err := runCycleAdvancingBackoffs(t, r, fake)
	if err != nil {
		t.Fatalf("RunCycle: %v", err)
	}

	orders := endpointResult(t, c, soak.EndpointOrders)
	if orders.OK {
		t.Errorf("%s: OK = true although the first page was refused", soak.EndpointOrders)
	}
	if orders.Class != soak.ClassOther {
		t.Errorf("%s: Class = %q, want %q", soak.EndpointOrders, orders.Class, soak.ClassOther)
	}
	// One refused CLOSED page, one OPEN page. No retry.
	if want := 2; orders.Requests != want {
		t.Errorf("%s: Requests = %d, want %d — a non-429 failure must be recorded, not retried",
			soak.EndpointOrders, orders.Requests, want)
	}
	if pages, _, _ := reads.counts(); pages != 2 {
		t.Errorf("the walk made %d page call(s), want 2", pages)
	}
}

// TestOnlyTheOrderReadsRetry. The credential probe is the survey's whole point
// and its failure is the measurement; it is not one of the two reads M8 found
// bursting, and widening the retry to it would delay the one signal the soak
// exists to record.
func TestOnlyTheOrderReadsRetry(t *testing.T) {
	reads := throttling(newStubReads(), errThrottled)
	reads.accountFailures = 1

	r, fake := newRunner(t, reads, func(o *soak.Options) {
		o.Classify = classifyThrottled
	})

	c, err := runCycleAdvancingBackoffs(t, r, fake)
	if err != nil {
		t.Fatalf("RunCycle: %v", err)
	}

	accounts := endpointResult(t, c, soak.EndpointAccounts)
	if accounts.OK {
		t.Errorf("%s: OK = true although the read was refused", soak.EndpointAccounts)
	}
	if accounts.Requests != 1 {
		t.Errorf("%s: Requests = %d, want 1 — the credential probe is not retried",
			soak.EndpointAccounts, accounts.Requests)
	}
	if _, _, calls := reads.counts(); calls != 1 {
		t.Errorf("the credential probe made %d call(s), want 1", calls)
	}
}

// TestAnInterruptedBackoffStopsWaiting. A soak is ended with Ctrl-C. Parking for
// thirty seconds must not be thirty seconds the operator has to wait through.
func TestAnInterruptedBackoffStopsWaiting(t *testing.T) {
	reads := throttling(newStubReads(), errThrottled)
	reads.pageFailures = 2

	r, fake := newRunner(t, reads, func(o *soak.Options) {
		o.Classify = classifyThrottled
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done := make(chan soak.Cycle, 1)
	go func() {
		c, _ := r.RunCycle(ctx)
		done <- c
	}()

	if !fake.WaitForSleepers(1, 5*time.Second) {
		t.Fatal("the walk never parked on a backoff after the broker answered 429")
	}
	cancel()

	select {
	case c := <-done:
		orders := endpointResult(t, c, soak.EndpointOrders)
		if orders.Class != soak.ClassRateLimited {
			t.Errorf("%s: Class = %q, want %q — the 429 is still why the read has no answer",
				soak.EndpointOrders, orders.Class, soak.ClassRateLimited)
		}
		// One attempt per status group and no retry after either: a cancelled
		// wait returns rather than sleeping, so neither walk spends its extra
		// attempts. (The cycle still surveys both groups — a failed walk has
		// never aborted the cycle, and this change does not make it start.)
		if pages, _, _ := reads.counts(); pages != 2 {
			t.Errorf("the walk made %d page call(s) after the cancellation, want 2 — "+
				"a cancelled backoff must not go on to spend its retries", pages)
		}
		if orders.Requests != 2 {
			t.Errorf("%s: Requests = %d, want 2", soak.EndpointOrders, orders.Requests)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("the backoff outlived the cancellation; the wait does not honour the context")
	}
}

// TestTheBackoffMatchesTheVerificationTool. internal/soak cannot import
// internal/verifylive — that package imports internal/official, which
// static_test.go keeps out of this one's import graph — so the policy exists
// twice. This is what keeps the two copies in step: the numbers are asserted,
// not just cross-referenced in a comment.
func TestTheBackoffMatchesTheVerificationTool(t *testing.T) {
	if soak.ReadRetryExtraAttempts != 2 {
		t.Errorf("ReadRetryExtraAttempts = %d, want 2 (verifylive.ReadRetryExtraAttempts)",
			soak.ReadRetryExtraAttempts)
	}
	for extra, want := range []time.Duration{15 * time.Second, 30 * time.Second} {
		if got := soak.ReadRetryBackoff(extra); got != want {
			t.Errorf("ReadRetryBackoff(%d) = %s, want %s (verifylive.ReadRetryBackoff)", extra, got, want)
		}
	}
}
