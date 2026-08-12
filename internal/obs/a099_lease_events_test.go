package obs_test

// a099, the sender's half of the lease: giving up, losing it, and saying which
// of those happened.
//
//	R7  — a sender that spends its budget lets the row go.
//	R15 — a sender that loses the lease stops sending at once.
//	R18 — contention, loss and expiry-takeover are three events, not one.
//	R20 — the default lease outlasts the delivery bound, and the bound is
//	      derived from live configuration rather than copied from it.

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/obs"
)

// a099LoggedSender is one notifier with its log captured, over a journal the
// test also holds so it can inspect the lease directly.
func a099LoggedSender(t *testing.T, pub obs.Publisher, attempts int) (*obs.Notifier, *journal.Journal, *bytes.Buffer, *clock.Fake) {
	t.Helper()
	clk := clock.NewFake(obsNow)
	j := openJournal(t, clk)
	buf := &bytes.Buffer{}
	n := &obs.Notifier{
		Log:         obs.NewLogger(obs.LogOptions{Writer: buf, JSON: true, Clock: clk}),
		Publisher:   pub,
		Journal:     j,
		Gate:        execgw.NewEntryGate(clk, map[execgw.RequiredQuery]time.Duration{}),
		Clock:       clock.System(),
		Attempts:    attempts,
		RetryDelay:  time.Millisecond,
		RemindAfter: a099Remind,
	}
	return n, j, buf, clk
}

// alwaysFails is a transport that is down.
type alwaysFails struct{ calls int }

func (p *alwaysFails) Publish(context.Context, obs.Notification) error {
	p.calls++
	return errors.New("transport is down")
}

// TestASenderThatSpendsItsBudgetReleasesTheRow is R7.
//
// RED before §4.5b: without a release, the row a sender gave up on stays leased
// until it expires, so the flush that exists to pick it up when the transport
// comes back cannot claim it. The alert survives — that was never in doubt — but
// nobody can send it for a lease's length, and these are the alerts that report
// a failed stop.
func TestASenderThatSpendsItsBudgetReleasesTheRow(t *testing.T) {
	pub := &alwaysFails{}
	n, j, _, _ := a099LoggedSender(t, pub, 2)
	ctx := context.Background()

	if err := n.Notify(ctx, a099Event()); err != nil {
		t.Fatalf("Notify: %v", err)
	}
	if pub.calls != 2 {
		t.Fatalf("publish attempts = %d, want 2 — the budget was not spent", pub.calls)
	}

	pending, err := j.PendingAlerts(ctx, 0)
	if err != nil {
		t.Fatalf("PendingAlerts: %v", err)
	}
	if len(pending) != 1 {
		t.Fatalf("pending rows = %d, want 1 — a failed critical alert is preserved", len(pending))
	}
	if pending[0].ClaimedBy != "" || pending[0].ClaimedAt != nil {
		t.Errorf("the abandoned row is still leased by %q since %v — nothing can send it "+
			"until the lease expires", pending[0].ClaimedBy, pending[0].ClaimedAt)
	}

	// And the proof that "not leased" means what it says: another sender takes it.
	claim, err := j.ClaimAlertByID(ctx, pending[0].ID, "flush")
	if err != nil {
		t.Fatalf("ClaimAlertByID: %v", err)
	}
	if claim.Disposition != journal.ClaimAcquired {
		t.Errorf("disposition = %v, want acquired", claim.Disposition)
	}
}

// stealsAfterFirst fails to publish, and between the first attempt and the
// second it lets the sender's lease expire and hands the row to somebody else.
//
// The stall is arranged rather than waited for. A sender that is paused past its
// own lease — a long GC, a suspended VM, a stopped container — is exactly the
// case C7 says the ledger has to survive, and the fake clock is how it is made
// to happen in a test rather than in production.
type stealsAfterFirst struct {
	calls int
	j     *journal.Journal
	clk   *clock.Fake
	id    int64
	t     *testing.T
}

func (p *stealsAfterFirst) Publish(context.Context, obs.Notification) error {
	p.calls++
	if p.calls != 1 || p.j == nil {
		return errors.New("transport is down")
	}
	pending, err := p.j.PendingAlerts(context.Background(), 0)
	if err != nil || len(pending) == 0 {
		p.t.Fatalf("arranging the steal: pending=%d err=%v", len(pending), err)
	}
	p.id = pending[0].ID
	p.clk.Advance(journal.DefaultAlertLease + time.Second)
	if claim, err := p.j.ClaimAlertByID(context.Background(), p.id, "the-other-sender"); err != nil {
		p.t.Fatalf("arranging the steal: %v", err)
	} else if claim.Disposition != journal.ClaimAcquired {
		p.t.Fatalf("arranging the steal: disposition = %v", claim.Disposition)
	}
	return errors.New("transport is down")
}

// TestASenderThatLosesTheLeaseStopsAtOnce is R15.
//
// RED before §4.5: with no token in the settlement predicate there is nothing to
// lose and nothing to notice, so a stalled sender spends its whole budget
// publishing a row somebody else already owns — the double delivery, one attempt
// at a time.
func TestASenderThatLosesTheLeaseStopsAtOnce(t *testing.T) {
	pub := &stealsAfterFirst{t: t}
	n, j, buf, clk := a099LoggedSender(t, pub, 4)
	pub.j, pub.clk = j, clk
	ctx := context.Background()

	if err := n.Notify(ctx, a099Event()); err != nil {
		t.Fatalf("Notify: %v", err)
	}
	if pub.calls != 1 {
		t.Errorf("publish attempts = %d, want 1 — the remaining budget belongs to whoever "+
			"holds the row now", pub.calls)
	}
	if !strings.Contains(buf.String(), string(obs.EventAlertClaimLost)) {
		t.Errorf("no %s line after the lease was taken; got:\n%s", obs.EventAlertClaimLost, buf.String())
	}
}

// TestContentionLossAndTakeoverAreThreeEvents is R18.
//
// RED before §4.5c. Every one of these arrived as engine.alert_undelivered
// before a099 — one line meaning "nothing was sent", "somebody else sent it" and
// "the previous sender died", which is a line an operator cannot act on.
//
// The token never appears in any of them. It is the proof of ownership, and a
// log line carrying it hands anyone who reads the log the ability to settle
// somebody else's send.
func TestContentionLossAndTakeoverAreThreeEvents(t *testing.T) {
	ctx := context.Background()

	t.Run("contention", func(t *testing.T) {
		pub := newA099Publisher(true)
		defer pub.releaseAll()
		holder, latecomer, _ := a099Pair(t, pub)
		buf := &bytes.Buffer{}
		latecomer.Log = obs.NewLogger(obs.LogOptions{Writer: buf, JSON: true, Clock: clock.NewFake(obsNow)})

		done := make(chan error, 1)
		go func() { done <- holder.Notify(ctx, a099Event()) }()
		pub.waitInside(t)
		if err := latecomer.Notify(ctx, a099Event()); err != nil {
			t.Fatalf("the second sender's Notify: %v", err)
		}
		pub.releaseAll()
		if err := <-done; err != nil {
			t.Fatalf("the holder's Notify: %v", err)
		}
		if !strings.Contains(buf.String(), string(obs.EventAlertClaimHeld)) {
			t.Errorf("no %s line for a skipped row; got:\n%s", obs.EventAlertClaimHeld, buf.String())
		}
		if strings.Contains(buf.String(), string(obs.EventAlertUndelivered)) {
			t.Errorf("contention was reported as %s — that line means the operator was not "+
				"told, and here another sender is telling them; got:\n%s",
				obs.EventAlertUndelivered, buf.String())
		}
	})

	t.Run("takeover", func(t *testing.T) {
		clk := clock.NewFake(obsNow)
		j := openJournal(t, clk)
		buf := &bytes.Buffer{}
		n := &obs.Notifier{
			Log:         obs.NewLogger(obs.LogOptions{Writer: buf, JSON: true, Clock: clk}),
			Publisher:   &a099Publisher{},
			Journal:     j,
			Gate:        execgw.NewEntryGate(clk, map[execgw.RequiredQuery]time.Duration{}),
			Clock:       clock.System(),
			Attempts:    1,
			RetryDelay:  time.Millisecond,
			RemindAfter: a099Remind,
		}
		// A sender that claimed the row and never came back.
		id, err := j.EnqueueAlert(ctx, journal.Alert{
			EventKey: a099Event().Key, Type: string(a099Event().Type), Severity: "critical"})
		if err != nil {
			t.Fatalf("EnqueueAlert: %v", err)
		}
		if _, err := j.ClaimAlertByID(ctx, id, "the-sender-that-died"); err != nil {
			t.Fatalf("ClaimAlertByID: %v", err)
		}
		clk.Advance(journal.DefaultAlertLease + time.Second)

		if err := n.Notify(ctx, a099Event()); err != nil {
			t.Fatalf("Notify: %v", err)
		}
		if !strings.Contains(buf.String(), string(obs.EventAlertClaimStolen)) {
			t.Errorf("no %s line after taking over an expired lease — a takeover means "+
				"something stopped, and that is never silent; got:\n%s",
				obs.EventAlertClaimStolen, buf.String())
		}
		if !strings.Contains(buf.String(), "the-sender-that-died") {
			t.Errorf("the takeover line does not name the previous holder; got:\n%s", buf.String())
		}
	})
}

// TestTheDefaultLeaseOutlastsTheDeliveryBound is R20.
//
// A lease shorter than one sender's cycle loses the row while that sender is
// still working through a budget this code told it to spend, and then two
// senders publish with nothing having gone wrong. The inequality is the whole of
// C7's first guarantee.
//
// Asserting the inequality alone is not enough: a DefaultAlertDeliveryBound that
// returned zero would satisfy it and mean nothing. So each of the four inputs is
// moved and the result has to move with it — that is the difference between a
// derivation and a number.
func TestTheDefaultLeaseOutlastsTheDeliveryBound(t *testing.T) {
	bound := obs.DefaultAlertDeliveryBound()
	if bound <= 0 {
		t.Fatalf("bound = %v — a bound of zero would make every comparison below vacuous", bound)
	}
	if journal.DefaultAlertLease <= bound {
		t.Errorf("DefaultAlertLease = %v is not longer than the delivery bound %v — a sender "+
			"would lose its row midway through its own retry budget",
			journal.DefaultAlertLease, bound)
	}

	base := obs.AlertDeliveryBound(obs.DefaultCriticalAttempts, obs.DefaultPublishTimeout,
		obs.DefaultRetryDelay, journal.DefaultBusyTimeout)
	if base != bound {
		t.Fatalf("DefaultAlertDeliveryBound() = %v but the same inputs give %v", bound, base)
	}

	for _, c := range []struct {
		name string
		got  time.Duration
	}{
		{"attempts", obs.AlertDeliveryBound(obs.DefaultCriticalAttempts+1, obs.DefaultPublishTimeout,
			obs.DefaultRetryDelay, journal.DefaultBusyTimeout)},
		{"publish timeout", obs.AlertDeliveryBound(obs.DefaultCriticalAttempts, obs.DefaultPublishTimeout+time.Second,
			obs.DefaultRetryDelay, journal.DefaultBusyTimeout)},
		{"retry delay", obs.AlertDeliveryBound(obs.DefaultCriticalAttempts, obs.DefaultPublishTimeout,
			obs.DefaultRetryDelay+time.Second, journal.DefaultBusyTimeout)},
		{"write wait", obs.AlertDeliveryBound(obs.DefaultCriticalAttempts, obs.DefaultPublishTimeout,
			obs.DefaultRetryDelay, journal.DefaultBusyTimeout+time.Second)},
	} {
		if c.got <= base {
			t.Errorf("raising the %s left the bound at %v (base %v) — the bound is not derived "+
				"from that input, so a configuration change would silently shorten the margin",
				c.name, c.got, base)
		}
	}
}
