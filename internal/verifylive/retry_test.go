package verifylive

// retry_test.go states the asymmetry the whole file exists for: a read is sent
// again after a 429, and a mutation never is.

import (
	"errors"
	"testing"
	"time"
)

// seedSettled writes a terminal verdict for every step except the named ones, so
// a test can drive one step of the procedure without the rest of it running.
func (h *harness) seedSettled(t *testing.T, verdict Verdict, except ...StepID) {
	t.Helper()
	skip := map[StepID]bool{}
	for _, id := range except {
		skip[id] = true
	}
	rec, err := OpenRecorder(h.record)
	if err != nil {
		t.Fatalf("OpenRecorder: %v", err)
	}
	defer rec.Close()
	for _, step := range Steps() {
		if skip[step.ID] {
			continue
		}
		if err := rec.Append(Entry{
			Kind: KindStep, StepID: step.ID, Verdict: verdict,
			StartedAt: h.now, FinishedAt: h.now, AccountRef: "****8901",
		}); err != nil {
			t.Fatalf("Append: %v", err)
		}
	}
}

// TestARateLimitedReadIsSentAgain: two 429s and the third attempt answers, which
// is the shape measurements.md M4 recorded and gave up on.
func TestARateLimitedReadIsSentAgain(t *testing.T) {
	broker := newFakeBroker().withThrottle("holdings", 2)
	h := newHarness(t, broker, alwaysConfirm())
	h.seedSettled(t, VerdictPass, StepSellableBaseline)

	if _, err := h.run(Options{}); err != nil {
		t.Fatalf("run: %v", err)
	}
	if got := broker.countRequests("GET /holdings"); got != 3 {
		t.Errorf("GET /holdings was sent %d time(s), want 3 (one attempt plus two retries)", got)
	}
	if got := h.verdict(StepSellableBaseline); got != VerdictPass {
		t.Errorf("verdict = %s, want pass — the third attempt answered", got)
	}
	if want := 45 * time.Second; h.slept != want {
		t.Errorf("the run waited %s between attempts, want %s (15s then 30s)", h.slept, want)
	}
}

// TestARateLimitedReadGivesUpInsideTheBudget. Two extra attempts and no more: the
// broker's actual window is unmeasured (task 1.3) and an unbounded retry would be
// this tool deciding how much of the operator's rate budget to spend.
func TestARateLimitedReadGivesUpInsideTheBudget(t *testing.T) {
	broker := newFakeBroker().withThrottle("holdings", 99)
	h := newHarness(t, broker, alwaysConfirm())
	h.seedSettled(t, VerdictPass, StepSellableBaseline)

	if _, err := h.run(Options{}); err != nil {
		t.Fatalf("run: %v", err)
	}
	if got := broker.countRequests("GET /holdings"); got != 1+readRetryExtraAttempts {
		t.Errorf("GET /holdings was sent %d time(s), want %d", got, 1+readRetryExtraAttempts)
	}
	if got := h.verdict(StepSellableBaseline); got != VerdictFail {
		t.Errorf("verdict = %s, want fail — the rate limit is a measurement result, not a crash", got)
	}
}

// TestOnlyARateLimitIsRetried. Everything else the broker says is an answer, and
// asking again would be this tool disagreeing with it.
func TestOnlyARateLimitIsRetried(t *testing.T) {
	broker := newFakeBroker().withThrottle("holdings", 99)
	broker.throttleErr = errors.New("official: server error")
	h := newHarness(t, broker, alwaysConfirm())
	h.seedSettled(t, VerdictPass, StepSellableBaseline)

	if _, err := h.run(Options{}); err != nil {
		t.Fatalf("run: %v", err)
	}
	if got := broker.countRequests("GET /holdings"); got != 1 {
		t.Errorf("GET /holdings was sent %d time(s), want 1 — only a 429 is retried", got)
	}
	if h.slept != 0 {
		t.Errorf("the run waited %s before giving up on a non-429; it must not back off at all", h.slept)
	}
}

// TestAMutationIsNeverRetried is the safety half, and it is the reason retry.go
// is not a transport-level decorator.
//
// A re-sent POST /api/v1/orders can leave two live orders on somebody's account:
// the broker may have accepted the first and lost the reply, and the idempotency
// key that would distinguish the two is exactly what task 2.7 has not measured.
func TestAMutationIsNeverRetried(t *testing.T) {
	broker := newFakeBroker().withThrottle("place", 99)
	h := newHarness(t, broker, alwaysConfirm())
	h.seedSettled(t, VerdictPass, StepIdempotency)

	if _, err := h.run(Options{}); err != nil {
		t.Fatalf("run: %v", err)
	}
	if got := broker.countRequests("POST /orders "); got != 1 {
		t.Errorf("POST /orders was sent %d time(s), want exactly 1 — a rate-limited placement is never "+
			"re-sent, because the first one may have been accepted", got)
	}
	if h.slept != 0 {
		t.Errorf("the run backed off for %s around a mutation; a mutation gets no retry and no wait", h.slept)
	}
	if got := h.verdict(StepIdempotency); got != VerdictFail {
		t.Errorf("verdict = %s, want fail", got)
	}
}

// TestEveryRateLimitedAttemptIsOnTheRecord. Task 1.3 sizes the retry matrix from
// this file; a record that showed only the attempt that eventually worked would
// hide the measurement the retry was invented for.
func TestEveryRateLimitedAttemptIsOnTheRecord(t *testing.T) {
	broker := newFakeBroker().withThrottle("holdings", 2)
	h := newHarness(t, broker, alwaysConfirm())
	h.seedSettled(t, VerdictPass, StepSellableBaseline)

	if _, err := h.run(Options{}); err != nil {
		t.Fatalf("run: %v", err)
	}
	entry, ok := LastEntry(h.entries(), StepSellableBaseline)
	if !ok {
		t.Fatal("the step wrote no entry")
	}
	attempts, rateLimited := 0, 0
	for _, c := range entry.Calls {
		if c.Endpoint != EndpointReadHoldings {
			continue
		}
		attempts++
		if c.Error != "" {
			rateLimited++
		}
	}
	if attempts != 3 {
		t.Errorf("the record holds %d holdings call(s), want 3 — every attempt is evidence", attempts)
	}
	if rateLimited != 2 {
		t.Errorf("the record holds %d failed holdings call(s), want 2", rateLimited)
	}
}
