package engine

// alertdelivery.go is a098's answer to "nobody sends what the outbox keeps".
//
// A critical alert reaches the ledger through EnqueueAlert and then waits. The
// synchronous notify path only sends the alert it is holding, and the one thing
// that drains a backlog — Notifier.Flush — has no production caller at all. So a
// row written while the transport was down, or written by a sender that then
// died, stayed PENDING until the same condition happened to be observed again.
// That is the defect. This loop is the subject the sentence was missing.
//
// # Why it does not call Flush
//
// Flush takes the notifier's mutex and holds it across every publish in the
// backlog (notifier.go:734-735, and PendingAlerts(ctx, 0) at :737 asks for all of
// them). The synchronous stop-alert path takes the same mutex. A loop built on
// Flush would therefore make a stop wait for N network round trips with no bound
// on N — the very defect a092 exists to remove, rebuilt here. design D1.1
// rejected that shape; this loop has its own delivery path and takes no mutex at
// all over a publish.
//
// # What separates it from the synchronous sender, then
//
// The ledger claim a099 built. Every row is claimed before it is published and
// settled under that claim's token, so two senders cannot both be sending one
// row. That is why a098 cannot ship without a099 (deploy-pair.txt).
//
// # The settlement rules are a099's, not this file's
//
// They are asymmetric and the asymmetry is the point:
//
//	published + settled     MarkAlertDelivered releases the lease as part of
//	                        settling. Releasing again would be a second write
//	                        saying nothing.
//	published + not settled the lease stays held. Releasing here would let the
//	                        next cycle send a row that already went out — the
//	                        2026-08-08 storm, arriving through the success path.
//	not published           record the attempt, then hand the lease back: this
//	                        row's turn is over and holding it until expiry keeps
//	                        it out of the next cycle for no reason.

import (
	"context"
	"fmt"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/obs"
)

// alertDeliveryInterval is how long the executor waits between cycles.
//
// It is measured rather than borrowed (a098 task 3.2). The measurement gives a
// band, not a number: an empty cycle costs 0.065ms of ledger read, so holding the
// steady-state duty cycle under 1/10,000 puts the floor at ~0.65s; keeping the
// wait from adding more than 5% to the 81s lease a dead sender's row sits behind
// puts the ceiling at ~4.05s. Two seconds sits near the middle of that band.
//
// The floor and the 5% are chosen; 0.065ms and 81s are measured. Anyone moving
// this should move the measurement first — `go test ./internal/app/engine/ -run
// '^$' -bench A098` reproduces it.
const alertDeliveryInterval = 2 * time.Second

// alertDeliveryBatch is how many rows one cycle takes.
//
// Reading is not what bounds it: 100 rows read in 0.449ms while settling a single
// row costs 5.584ms, two orders of magnitude apart. Ten rows is 56ms of ledger
// per cycle, 2.8% of the interval. A hundred would be 28%.
//
// The cost that is not on that scale is the network. With every publish timing
// out, a cycle of ten takes about 100s — which is why the cycle checks for
// cancellation between rows rather than only between cycles.
const alertDeliveryBatch = 10

// alertReleaseTimeout bounds handing a lease back when the cycle's own context is
// already done.
const alertReleaseTimeout = 5 * time.Second

// alertDeliverer drains the critical-alert outbox.
type alertDeliverer struct {
	Journal   *journal.Journal
	Publisher obs.Publisher
	Log       *obs.Logger
	Clock     clock.Clock
	Interval  time.Duration
	Batch     int
	Claimant  string

	// heldReported is which rows this executor has already said are held, keyed
	// by the lease it saw. It is written only from Run's own goroutine (cycle →
	// deliverOne), so it carries no lock.
	//
	// It exists because "held" is worth saying once and worthless said forty
	// times: a stuck row is re-listed every cycle, and at a 2s interval under an
	// 81s lease one abandoned row would write ~40 identical lines before the
	// lease even expires. That is not observation, it is a log storm that buries
	// the line an operator needs (a098 R21).
	heldReported map[int64]time.Time
}

// Run cycles until ctx ends.
//
// It cycles first and waits afterwards, so an engine starting with a backlog
// begins draining it immediately rather than one interval later.
//
// A cancelled context is a clean stop and is reported *as a cancellation*:
// ctx.Err(), never nil.
//
// ⛔ The first draft returned nil here, reasoning that "a clean stop is not a
// failure". Under the runtime that reads this, the reasoning inverts: the
// runtime judges a stop with Runtime.gracefulStop, which requires the error to
// be the context's (runtime.go). A nil return has no context in it, so every
// ordinary shutdown would be classified as the sender dying — and that is the
// one thing that latches the entry gate. The judgement is deliberately the same
// predicate the supervised loops get (design D8.3), so the convention is the
// return's to match, not the predicate's.
//
// A cycle that fails does not stop the loop. The backlog is exactly what should
// outlive a transient ledger or transport fault, and a loop that exits on the
// first error would leave nobody sending again.
func (d *alertDeliverer) Run(ctx context.Context) error {
	for {
		if err := d.cycle(ctx); err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			d.logf(obs.EventAlertUndelivered, err, "a delivery cycle failed")
		}
		// Both clocks return ctx.Err() when the wait is cancelled (clock.go,
		// fake.go), so this is the cancellation travelling out unchanged rather
		// than a second judgement about what happened.
		if err := d.Clock.Sleep(ctx, d.interval()); err != nil {
			return err
		}
	}
}

// cycle claims, publishes and settles one batch.
//
// It returns an error only for a failure that made the whole cycle impossible —
// reading the backlog. One row failing is that row's problem, not the backlog's:
// returning early there is how a flush once abandoned every remaining critical
// alert because the first row had settled underneath it.
func (d *alertDeliverer) cycle(ctx context.Context) error {
	// An episode nobody can still be in is not worth remembering. Dropping the
	// lapsed ones here is what keeps the map bounded by *currently contended*
	// rows rather than by every row this engine has ever seen contended.
	d.forgetLapsedHeld(d.Clock.Now())
	pending, err := d.Journal.PendingAlerts(ctx, d.batch())
	if err != nil {
		return fmt.Errorf("listing the critical alert backlog: %w", err)
	}
	for _, alert := range pending {
		// Between rows, not only between cycles: with a dead transport this loop
		// spends a publish timeout per row, and a batch of them is far longer
		// than the interval. Run's wait being cancellable is not enough on its own.
		if ctx.Err() != nil {
			return nil
		}
		d.deliverOne(ctx, alert)
	}
	return nil
}

// deliverOne is one row's whole life in this cycle.
func (d *alertDeliverer) deliverOne(ctx context.Context, alert journal.Alert) {
	claim, err := d.Journal.ClaimAlertByID(ctx, alert.ID, d.claimant())
	if err != nil {
		d.logf(obs.EventAlertUndelivered, err, "claiming an alert failed", "alert_id", alert.ID)
		return
	}
	switch claim.Disposition {
	case journal.ClaimAcquired:
		// The row is ours now, so whatever episode of contention it was in is
		// over. Forgetting here is what makes the next one reportable.
		d.forgetHeld(alert.ID)
	case journal.ClaimHeldElsewhere:
		// Another sender holds a live lease. Skipping is right — but not
		// silently: a row that is held every cycle is not contention, it is a
		// sender that stopped without settling, and the ledger only reopens it
		// when the lease expires.
		//
		// Once per lease, though. Saying it every cycle would report the same
		// fact ~40 times before that lease expires, and forty copies of one line
		// is how the *next* line stops being read (R21).
		if d.reportHeld(alert.ID, claim.ExpiresAt) {
			d.logf(obs.EventAlertClaimHeld, nil, "an alert is held by another sender",
				"alert_id", alert.ID, "claimed_by", claim.ClaimedBy,
				"claim_expires_at", claim.ExpiresAt.UTC().Format(time.RFC3339))
		}
		return
	default:
		// Settled between the listing and here. Normal.
		d.forgetHeld(alert.ID)
		return
	}
	if claim.Stole {
		// A steal is a designed recovery and never silent: somebody died or
		// stalled while holding this row.
		d.logf(obs.EventAlertClaimHeld, nil, "an expired alert lease was taken over",
			"alert_id", alert.ID, "stole_from", claim.StoleBy)
	}

	if d.Publisher == nil {
		// No transport at all. The row stays PENDING and the lease goes back;
		// saying so once per cycle is what keeps a silently unconfigured engine
		// from looking like a working one.
		d.logf(obs.EventAlertUndelivered, nil, "no publisher is configured, so nothing can be sent",
			"alert_id", alert.ID)
		d.release(ctx, alert.ID, claim.Token)
		return
	}

	perr := d.Publisher.Publish(ctx, obs.Notification{
		Type:     obs.EventType(alert.Type),
		Severity: obs.SeverityCritical,
		Title:    alert.Title,
		Body:     alert.Body,
	})
	if perr != nil {
		if _, err := d.Journal.MarkAlertAttemptFailed(ctx, alert.ID, claim.Token, perr.Error()); err != nil {
			d.logf(obs.EventAlertUndelivered, err, "recording a failed attempt failed", "alert_id", alert.ID)
		}
		d.release(ctx, alert.ID, claim.Token)
		return
	}

	settled, err := d.Journal.MarkAlertDelivered(ctx, alert.ID, claim.Token)
	if err != nil {
		// Published but not settled, and the write failed rather than being
		// refused. The lease stays: see the file comment.
		d.logf(obs.EventAlertUndelivered, err, "settling a delivered alert failed", "alert_id", alert.ID)
		return
	}
	if settled.Outcome != journal.SettleApplied {
		// Published, but somebody else's row to settle. It is out; the holder
		// records it. Still no release.
		d.logf(obs.EventAlertUndelivered, nil, "a delivered alert was settled by somebody else",
			"alert_id", alert.ID, "outcome", settled.Outcome.String())
	}
}

// release hands a lease back for a row this cycle is done with.
//
// The context is detached because a cycle cancelled partway must not leave the
// rows it claimed locked behind it: the ledger only reopens them when the lease
// expires, and until then nobody sends them.
func (d *alertDeliverer) release(ctx context.Context, id int64, token string) {
	relCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), alertReleaseTimeout)
	defer cancel()
	if _, err := d.Journal.ReleaseAlertClaim(relCtx, id, token); err != nil {
		d.logf(obs.EventAlertUndelivered, err, "releasing an alert lease failed", "alert_id", id)
	}
}

// reportHeld says whether this held row is worth a line, and remembers that it
// answered yes.
//
// The episode is keyed by the *lease*, not by the row. A second sender taking
// the row after the first one's lease lapsed is a new fact about a different
// sender, and collapsing the two would hide the handover.
func (d *alertDeliverer) reportHeld(id int64, expires time.Time) bool {
	if d.heldReported == nil {
		d.heldReported = make(map[int64]time.Time)
	}
	if seen, ok := d.heldReported[id]; ok && seen.Equal(expires) {
		return false
	}
	d.heldReported[id] = expires
	return true
}

func (d *alertDeliverer) forgetHeld(id int64) {
	delete(d.heldReported, id)
}

// forgetLapsedHeld drops episodes whose lease has already run out.
//
// It bounds the map by what is contended *now*. Without it a long-running engine
// keeps one entry per row it ever found held, including rows that were settled
// by somebody else and will never be listed again.
func (d *alertDeliverer) forgetLapsedHeld(now time.Time) {
	for id, expires := range d.heldReported {
		if !expires.After(now) {
			delete(d.heldReported, id)
		}
	}
}

func (d *alertDeliverer) interval() time.Duration {
	if d.Interval <= 0 {
		return alertDeliveryInterval
	}
	return d.Interval
}

func (d *alertDeliverer) batch() int {
	if d.Batch <= 0 {
		return alertDeliveryBatch
	}
	return d.Batch
}

func (d *alertDeliverer) claimant() string {
	if d.Claimant == "" {
		return "engine.alert_delivery"
	}
	return d.Claimant
}

// logf never carries the claim token: whoever reads it can settle somebody
// else's send (a099, ClaimResult.Token).
func (d *alertDeliverer) logf(event obs.EventType, err error, detail string, args ...any) {
	if d.Log == nil {
		return
	}
	args = append([]any{obs.FieldDetail, detail}, args...)
	if err != nil {
		d.Log.Error(event, err, args...)
		return
	}
	d.Log.Warn(event, args...)
}
