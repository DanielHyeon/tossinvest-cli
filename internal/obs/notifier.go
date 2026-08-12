package obs

// notifier.go grades, delivers and — for critical events — durably remembers
// alerts (harden-execution-base task 4.3, engine-safety "등급화된 알림").
//
// # The two paths
//
//	normal    → publish once, log the failure, carry on
//	critical  → enqueue in the journal outbox, then publish; only a successful
//	            send marks it delivered. Retries are bounded, and an alert still
//	            undelivered after them blocks new entries.
//
// The asymmetry is the whole design. A missed fill notification costs an operator
// some context. A missed IN_DOUBT notification means a live account is in a state
// nobody knows about and the engine keeps trading into it.
//
// # Why the block is not automatically released
//
// Delivery recovering does not mean the alert was read. The gate latch is cleared
// by an explicit operator acknowledgement (Acknowledge), which is also what marks
// the outbox row. "전달 복구 후 수동 확인으로 해제한다" is the spec, and it is the
// right rule: the alert existed to make a human look at something.

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
)

// DefaultCriticalAttempts is how many times one critical alert is published
// before the entry gate is latched.
//
// Three, because the failures worth retrying through are transient (a DNS blip, a
// restarting ntfy container) and the ones that are not — wrong topic, dead token
// — do not improve with a fourth try, while every retry delays the moment the
// operator learns delivery is broken at all.
const DefaultCriticalAttempts = 3

// DefaultRetryDelay is the wait between critical delivery attempts.
const DefaultRetryDelay = 2 * time.Second

// DefaultRemindAfter is how long a critical alert stays quiet after it lands
// before the same condition is worth telling the operator about again.
//
// A persistent condition is re-observed on every cycle — every few seconds for
// the exit loop — so without a bound the reminder is the pager storm it exists
// to replace. An hour is the compromise: it holds a weekend of one stuck
// position to a couple of dozen pushes, and it is short enough that a transport
// that dies after a delivery is found and latched within the hour rather than
// never (a096).
const DefaultRemindAfter = time.Hour

// Event is what a caller reports.
type Event struct {
	Type EventType
	// Key deduplicates. Empty derives one from the type and the context fields,
	// which is right for conditions ("AAPL is in an unknown state") and wrong for
	// occurrences ("a fill happened") — the latter are never critical, so they
	// never reach the outbox.
	Key   string
	Title string
	Body  string
	// Fields is operator context: symbol, attempt id, reason code. It is stored
	// as JSON on the outbox row and appended to the log line.
	Fields map[string]any
}

// Notifier grades and delivers events.
type Notifier struct {
	// Log receives every event, whatever its grade. Optional.
	Log *Logger
	// Publisher sends notifications. Optional: nil means alerts are logged and,
	// for critical ones, still enqueued — an unconfigured transport must not be
	// a silent hole where the durable record should be.
	Publisher Publisher
	// Journal backs the critical outbox. Optional; without it critical events
	// degrade to best-effort and the Notifier says so on every one.
	//
	// It is also what makes the escalation below durable: the same handle owns
	// the operating-mode history.
	Journal *journal.Journal
	// Gate is latched when a critical alert cannot be delivered. Optional.
	Gate *execgw.EntryGate
	// AccountRef scopes the operating-mode tightening that sustained delivery
	// failure triggers (risk-management: critical 알림 outbox 전달 실패 지속 →
	// ENTRY_BLOCKED). Empty leaves the gate latch as the only consequence.
	//
	// The latch and the transition answer different questions. The latch stops
	// this process from opening a position; the transition is what a restart
	// still knows about, and an operator who sees "alerts are not arriving" is
	// exactly the person about to restart something.
	AccountRef string
	// Clock drives the retry waits. Defaults to clock.System().
	Clock clock.Clock
	// Attempts is how many publishes one critical alert gets. Zero uses
	// DefaultCriticalAttempts.
	Attempts int
	// RetryDelay is the wait between them. Zero uses DefaultRetryDelay.
	RetryDelay time.Duration
	// RemindAfter is how long a delivered critical alert stays quiet before the
	// same condition earns another send. Zero uses DefaultRemindAfter.
	RemindAfter time.Duration

	// mu serialises the whole claim-and-send, not just the send. Claiming
	// outside it would let two observations of one condition both read a row
	// that has not been delivered yet and both publish (a096 round 1, blocker 1).
	mu sync.Mutex
}

// Notify grades an event and delivers it.
//
// It returns an error only when a *critical* event could not be made durable —
// that is, when the outbox write itself failed. A failed send is not an error to
// the caller: it has already been handled here, by latching the gate, and
// bubbling it up would make every call site re-implement that decision.
func (n *Notifier) Notify(ctx context.Context, e Event) error {
	severity := SeverityOf(e.Type)
	n.logEvent(e, severity)

	if severity != SeverityCritical {
		n.publishBestEffort(ctx, e, severity)
		return nil
	}
	return n.notifyCritical(ctx, e)
}

// logEvent writes the structured line for an event.
func (n *Notifier) logEvent(e Event, severity Severity) {
	if n.Log == nil {
		return
	}
	args := make([]any, 0, 2*len(e.Fields)+2)
	for k, v := range e.Fields {
		args = append(args, k, v)
	}
	if strings.TrimSpace(e.Body) != "" {
		args = append(args, FieldDetail, e.Body)
	}
	if severity == SeverityCritical {
		n.Log.Warn(e.Type, args...)
		return
	}
	n.Log.Event(e.Type, args...)
}

// publishBestEffort sends an ordinary alert and forgets about it.
func (n *Notifier) publishBestEffort(ctx context.Context, e Event, severity Severity) {
	if n.Publisher == nil {
		return
	}
	if err := n.Publisher.Publish(ctx, notificationFor(e, severity)); err != nil && n.Log != nil {
		// Logged, not escalated: this grade is best-effort by definition, and
		// treating its failure as an incident would make the grading meaningless.
		n.Log.Warn(EventAlertUndelivered,
			FieldEvent, string(e.Type),
			FieldSeverity, string(severity),
			FieldError, err.Error())
	}
}

// notifyCritical is the durable path.
func (n *Notifier) notifyCritical(ctx context.Context, e Event) error {
	if n.Journal == nil {
		// No durable store. Say so loudly and still try to send: degrading to
		// best-effort silently would leave an operator believing the outbox has
		// their back.
		if n.Log != nil {
			n.Log.Warn(EventAlertUndelivered,
				FieldEvent, string(e.Type),
				FieldDetail, "no journal is wired, so this critical alert is not durable")
		}
		n.publishBestEffort(ctx, e, SeverityCritical)
		return nil
	}

	record := journal.Alert{
		EventKey: n.eventKey(e),
		Type:     string(e.Type),
		Severity: string(SeverityCritical),
		Title:    e.Title,
		Body:     e.Body,
		Payload:  encodeFields(e.Fields),
	}
	// Durable first, exactly like an intent: a record that only exists in memory
	// is a record that does not survive the crash it is warning about.
	sent, owed, err := n.claimAndDeliver(ctx, record, e)
	if err != nil {
		// The entry gate is already latched — claimAndDeliver did that under the
		// mutex. What is missing is the half that outlives this process.
		//
		// a097's first draft stopped at the latch and called that the cautious
		// choice. It was not: EntryGate keeps its latches in a map in memory, and
		// a claim that failed leaves no outbox row either, so a restart erases
		// the block *and* the evidence and reopens entries with nobody told. The
		// comment on escalate names exactly this — the part that is lost is "the
		// part that survives a restart".
		//
		// Escalating may itself fail, because the store that could not take the
		// alert is the store the transition is written to. It is still attempted:
		// escalate logs that failure at error level, and "a restart will reopen
		// entries" on the record beats the same fact nowhere.
		//
		// Outside the mutex, like the escalation below it, and for the same
		// reason: an announcer wired to this Notifier re-enters Notify.
		n.escalate(ctx, e)
		return fmt.Errorf("obs: recording a critical alert: %w", err)
	}

	if owed && !sent {
		// Outside claimAndDeliver, and that is not tidiness: it holds n.mu, the
		// escalation announces through a ModeAnnouncer, and an announcer wired to
		// this Notifier would re-enter Notify and deadlock on a mutex Go does not
		// make reentrant.
		n.escalate(ctx, e)
	}
	return nil
}

// claimAndDeliver asks the outbox whether this send is owed and, if it is,
// performs it — both under the delivery mutex.
//
// Both under one lock because they are one decision. Claiming outside it lets
// two observations of the same condition read the same not-yet-delivered row,
// each conclude the send is owed, and each publish; the second then fails to
// mark a row that is already DELIVERED, which is exactly the `no such alert`
// line the 2026-08-08 storm left behind (a096 round 1, blocker 1).
//
// The mutex stays, and what it is for has changed. It excludes senders *inside
// this Notifier* and never did anything else — a second Notifier, or a flush on
// another one, walked straight past it. Exclusion between senders is now the
// ledger's, held as a lease on the row itself (journal/alert_claim.go), and this
// lock is the local half of it.
//
// It reports whether the send happened and whether it was owed at all. A send
// that was never owed is not a delivery failure and must not escalate — and
// neither is a send another holder took over.
func (n *Notifier) claimAndDeliver(
	ctx context.Context, record journal.Alert, e Event,
) (sent bool, owed bool, err error) {
	n.mu.Lock()
	defer n.mu.Unlock()

	claim, err := n.Journal.ClaimAlertForDelivery(ctx, record, n.remindAfter(), n.claimant())
	if err != nil {
		// Nothing was written and nothing was sent, so this is strictly worse
		// than the failures below it: a spent retry budget at least leaves a
		// PENDING row for a later flush to find. Here there is no row.
		//
		// Returning the error was the whole response until a097, and one caller
		// discards it (internal/flatten/flatten.go:694). Latching here instead
		// makes the outcome independent of whether a caller checks — which it
		// has to be, because callers keep being added.
		//
		// Entries only. Exits are untouched: no alert failure may slow a stop.
		detail := fmt.Sprintf(
			"a critical %s alert could not be recorded in the outbox: %v", e.Type, err)
		if n.Log != nil {
			n.Log.Error(EventAlertUndelivered, err, FieldEvent, string(e.Type))
		}
		if n.Gate != nil {
			n.Gate.Block(execgw.ReasonAlertUndelivered, detail)
		}
		return false, false, err
	}
	switch claim.Disposition {
	case journal.ClaimSettled:
		// The operator already has this one and the reminder window has not
		// elapsed. Sending again would tell them nothing new and would spend the
		// channel they need for the next distinct condition.
		//
		// The line for this observation was still written: logEvent runs in
		// Notify ahead of the grading branch, so the record of how long the
		// condition persisted does not depend on whether it was sent.
		return false, false, nil
	case journal.ClaimHeldElsewhere:
		// Another sender is publishing this row right now. Skipping is the
		// exclusion doing its job, so this is not a failure and nothing is owed
		// of *this* observation.
		//
		// The entry gate is not touched — not blocked, not cleared. Blocking
		// would latch a reason that has no automatic release on an event that is
		// entirely normal, leaving a successful delivery behind a lock nobody
		// can open; clearing would let a routine race unlock a block a real
		// failure put there. Contention is logged and nothing else.
		if n.Log != nil {
			n.Log.Event(EventAlertClaimHeld,
				FieldEvent, string(e.Type),
				"alert_id", claim.ID,
				"claimed_by", claim.ClaimedBy,
				"claim_age_ms", claimAgeMS(claim.ClaimedAt, n.now()))
		}
		return false, false, nil
	}

	if claim.Stole && n.Log != nil {
		// Acquired, but from somebody who never came back for it.
		n.Log.Warn(EventAlertClaimStolen,
			FieldEvent, string(e.Type),
			"alert_id", claim.ID,
			"claimed_by", claim.StoleBy,
			"claim_age_ms", claimAgeMS(claim.StoleAt, n.now()))
	}
	sent, lost := n.deliver(ctx, claim.ID, claim.Token, e)
	if lost {
		// The row moved on without us. Nothing is owed of this observation, so
		// the caller does not escalate: the operating mode tightens on sustained
		// *delivery* failure, and losing a race to another sender is not that.
		return sent, false, nil
	}
	return sent, true, nil
}

// claimant names this sender in the lease it takes. It is for the operator and
// the log; exclusion rests on the token, never on the name.
func (n *Notifier) claimant() string {
	if name := strings.TrimSpace(n.AccountRef); name != "" {
		return "notifier:" + name
	}
	return "notifier"
}

// now is the notifier's clock, which the tests replace.
func (n *Notifier) now() time.Time {
	if n.Clock != nil {
		return n.Clock.Now()
	}
	return clock.System().Now()
}

// claimAgeMS is how long the lease being reported has existed. Zero when the
// stamp is missing, which is what an unparsable or absent claimed_at means.
func claimAgeMS(since, now time.Time) int64 {
	if since.IsZero() {
		return 0
	}
	age := now.Sub(since)
	if age < 0 {
		return 0
	}
	return age.Milliseconds()
}

// remindAfter is the configured reminder window, or the default.
func (n *Notifier) remindAfter() time.Duration {
	if n.RemindAfter > 0 {
		return n.RemindAfter
	}
	return DefaultRemindAfter
}

// escalate persists the automatic tightening that sustained critical-alert
// delivery failure triggers.
//
// # No announcer
//
// The transition is announced to nobody on purpose. The transport that would
// carry the announcement is the one that just failed its whole retry budget, so
// announcing would enqueue a second undeliverable alert and spend a second
// budget on it. The transition is durable in the journal and in the structured
// log, the operator's original alert is still PENDING in the outbox waiting for
// the transport to come back, and the account is blocked either way. It also
// removes the only path by which this could re-enter deliver.
//
// # No error return
//
// Notify's contract is that a failed *send* is not the caller's problem — it has
// already been handled, by latching the gate — and only a failed outbox *write*
// is. A failed escalation is the same shape as a failed send: the in-process
// block stands, and what was lost is the part that survives a restart. It is
// logged at error level rather than bubbled, because every call site would
// otherwise have to re-implement this same judgement.
func (n *Notifier) escalate(ctx context.Context, e Event) {
	if n.Journal == nil || strings.TrimSpace(n.AccountRef) == "" {
		return
	}
	_, changed, err := n.Journal.EscalateOperatingMode(ctx, n.AccountRef,
		journal.ModeTriggerCriticalAlertUndelivered, nil)
	switch {
	case err != nil && n.Log != nil:
		n.Log.Error(EventOperatingMode, err,
			FieldAccount, n.AccountRef,
			FieldEvent, string(e.Type),
			FieldDetail, "the undelivered critical alert did not reach the operating mode, "+
				"so a restart would lift the block")
	case changed && n.Log != nil:
		// Not a Notify: see above. The line is the record.
		n.Log.Warn(EventOperatingMode,
			FieldAccount, n.AccountRef,
			FieldToState, journal.ModeEntryBlocked,
			FieldReason, journal.ModeTriggerCriticalAlertUndelivered,
			FieldDetail, "new entries are blocked until an operator acknowledges the alert backlog")
	}
}

// deliver publishes one outbox row under the retry budget, latching the gate if
// it cannot. It reports whether the delivery is *settled* — published and
// recorded as published — which is not the same as whether it went out. A send
// this system cannot account for counts as unsettled (a096 round 2).
//
// PRECONDITION: the caller holds n.mu, and holds it across the claim that
// produced id as well as this send. One delivery loop at a time — two
// goroutines publishing the same backlog would double-send and race on the
// attempt counter — and one claim-and-send at a time, or two observations of
// the same condition each conclude the send is owed and each publish.
// The lease travels with id: every settlement below presents the token this
// send was claimed under, and a settlement the ledger refuses because the token
// no longer matches ends the send then and there. A sender that stalled past its
// own lease has already lost the row to somebody else, and the worst thing it
// can do at that point is keep publishing.
//
// The second return says the lease left this sender — through contention or an
// operator's acknowledgement. The caller uses it to keep quiet: neither outcome
// is this observation's delivery failure, so neither escalates.
func (n *Notifier) deliver(ctx context.Context, id int64, token string, e Event) (sent bool, lost bool) {
	attempts := n.Attempts
	if attempts <= 0 {
		attempts = DefaultCriticalAttempts
	}
	msg := notificationFor(e, SeverityCritical)

	var lastErr error
	for attempt := 1; attempt <= attempts; attempt++ {
		if n.Publisher == nil {
			lastErr = errors.New("no notification publisher is configured")
			break
		}
		err := n.Publisher.Publish(ctx, msg)
		if err == nil {
			settled, markErr := n.Journal.MarkAlertDelivered(ctx, id, token)
			if markErr == nil {
				switch settled.Outcome {
				case journal.SettleApplied:
					return true, false
				case journal.SettleLeaseLost, journal.SettleAlreadySettled:
					// The push went out and the row was not ours to settle. It
					// is out either way, so the send happened; what did not
					// happen is our record of it, and the holder that owns the
					// row now writes that.
					//
					// Not a gate event. Losing a lease is the exclusion working
					// and an operator acknowledging mid-flight is a person doing
					// their job; latching new entries on either would punish the
					// normal case.
					n.logLeaseLost(settled, id, e)
					return true, true
				case journal.SettleNotFound:
					markErr = fmt.Errorf("the outbox row disappeared while it was being delivered")
				}
			}
			if markErr == nil {
				return true, false
			}
			// The push landed and the record of it did not. Reporting success
			// here — which this did until a096 round 2 — leaves the row PENDING,
			// so the next observation finds the send owed and sends again, and
			// the one after that. That is the 2026-08-08 storm reached through
			// the success path, and it is silent, because a caller told the send
			// succeeded neither latches the gate nor escalates.
			//
			// "The push went out" is not "the operator is covered" when this
			// system's own record says it is not. So an unaccountable send is an
			// unsettled one: stop new entries, let exits through.
			detail := fmt.Sprintf(
				"a critical %s alert was published but could not be recorded as delivered: %v",
				e.Type, markErr)
			if n.Log != nil {
				n.Log.Error(EventAlertUndelivered, markErr,
					FieldEvent, string(e.Type),
					"alert_id", id)
			}
			if n.Gate != nil {
				n.Gate.Block(execgw.ReasonAlertUndelivered, detail)
			}
			// The lease is deliberately kept. It is now the only thing
			// suppressing a re-send of an alert the operator already has, and
			// releasing it here would hand the row straight back to the next
			// observation — the storm, through the success path. Expiry lets go
			// eventually; nothing else does.
			return false, false
		}
		lastErr = err
		failed, markErr := n.Journal.MarkAlertAttemptFailed(ctx, id, token, err.Error())
		if markErr != nil {
			if n.Log != nil {
				n.Log.Error(EventAlertUndelivered, markErr)
			}
		} else if failed.Outcome != journal.SettleApplied {
			// The row stopped being ours between attempts. Stop now: the retries
			// left in the budget belong to whoever holds it, and spending them
			// is the double delivery again, one attempt at a time.
			n.logLeaseLost(failed, id, e)
			return false, true
		}
		if attempt < attempts {
			if !n.wait(ctx) {
				break
			}
		}
	}

	// Out of attempts. The row stays PENDING — it is preserved, not abandoned —
	// and new entries stop.
	//
	// The lease does not stay. This sender is finished with the row and is about
	// to tell the gate so; holding the claim would leave the row locked against
	// the flush that is supposed to pick it up when the transport comes back.
	// This is the one place a sender lets go without settling, which is why
	// releasing is its own verb.
	if released, relErr := n.Journal.ReleaseAlertClaim(ctx, id, token); relErr != nil {
		if n.Log != nil {
			n.Log.Error(EventAlertUndelivered, relErr, FieldEvent, string(e.Type), "alert_id", id)
		}
	} else if released.Outcome == journal.SettleLeaseLost {
		n.logLeaseLost(released, id, e)
	}
	detail := fmt.Sprintf("a critical %s alert could not be delivered after %d attempts: %v",
		e.Type, attempts, lastErr)
	if n.Log != nil {
		n.Log.Error(EventAlertUndelivered, lastErr,
			FieldEvent, string(e.Type),
			"alert_id", id)
	}
	if n.Gate != nil {
		n.Gate.Block(execgw.ReasonAlertUndelivered, detail)
	}
	return false, false
}

// logLeaseLost records that a settlement was refused, and by whom it is held
// now. It is the event the operator needs to tell contention from breakage; the
// token itself never appears, because whoever can read it can settle somebody
// else's send.
func (n *Notifier) logLeaseLost(res journal.SettleResult, id int64, e Event) {
	if n.Log == nil {
		return
	}
	if res.Outcome == journal.SettleAlreadySettled {
		n.Log.Event(EventAlertClaimLost,
			FieldEvent, string(e.Type),
			"alert_id", id,
			FieldDetail, "the row was settled by an operator or an earlier delivery")
		return
	}
	n.Log.Warn(EventAlertClaimLost,
		FieldEvent, string(e.Type),
		"alert_id", id,
		"claimed_by", res.ClaimedBy,
		"claim_age_ms", claimAgeMS(res.ClaimedAt, n.now()))
}

// wait sleeps between attempts, reporting false when the context ended.
func (n *Notifier) wait(ctx context.Context) bool {
	delay := n.RetryDelay
	if delay <= 0 {
		delay = DefaultRetryDelay
	}
	clk := n.Clock
	if clk == nil {
		clk = clock.System()
	}
	return clk.Sleep(ctx, delay) == nil
}

// Flush retries every pending outbox row.
//
// It is what a supervising loop calls periodically and what an operator triggers
// after fixing the transport. A run that empties the backlog does *not* clear the
// gate: see Acknowledge.
func (n *Notifier) Flush(ctx context.Context) (delivered int, remaining int, err error) {
	if n.Journal == nil {
		return 0, 0, nil
	}
	// The same mutex the notify path holds. Without it a flush and an
	// observation can publish the same row at the same moment, which is the
	// double-send this package exists to prevent (a096 round 1, blocker 1).
	n.mu.Lock()
	defer n.mu.Unlock()

	pending, err := n.Journal.PendingAlerts(ctx, 0)
	if err != nil {
		return 0, 0, err
	}
	for _, alert := range pending {
		if n.Publisher == nil {
			break
		}
		// Claim per row. A flush that published straight from the listing was
		// the one send path with no lease around it, and a lease with a bypass
		// beside it is not a lease: every row this loop sent was a row another
		// sender could be sending at the same moment.
		//
		// The list was read before any of this, so a row can settle or be taken
		// between the read and here. That is what the claim is for.
		claim, cerr := n.Journal.ClaimAlertByID(ctx, alert.ID, n.claimant())
		if cerr != nil {
			return delivered, 0, cerr
		}
		if claim.Disposition != journal.ClaimAcquired {
			// Held by another sender, or no longer pending. Neither is this
			// loop's failure and neither touches the gate.
			if claim.Disposition == journal.ClaimHeldElsewhere && n.Log != nil {
				n.Log.Event(EventAlertClaimHeld,
					FieldEvent, alert.Type,
					"alert_id", claim.ID,
					"claimed_by", claim.ClaimedBy,
					"claim_age_ms", claimAgeMS(claim.ClaimedAt, n.now()))
			}
			continue
		}
		if claim.Stole && n.Log != nil {
			n.Log.Warn(EventAlertClaimStolen,
				FieldEvent, alert.Type,
				"alert_id", claim.ID,
				"claimed_by", claim.StoleBy,
				"claim_age_ms", claimAgeMS(claim.StoleAt, n.now()))
		}
		msg := Notification{
			Type:     EventType(alert.Type),
			Severity: SeverityCritical,
			Title:    alert.Title,
			Body:     alert.Body,
		}
		if perr := n.Publisher.Publish(ctx, msg); perr != nil {
			_, _ = n.Journal.MarkAlertAttemptFailed(ctx, alert.ID, claim.Token, perr.Error())
			// This row's turn is over whether or not the attempt was recorded,
			// so the lease goes back. Holding it until expiry would keep the row
			// out of the next flush for no reason: unlike deliver, nothing here
			// is mid-way through a retry budget.
			_, _ = n.Journal.ReleaseAlertClaim(ctx, alert.ID, claim.Token)
			continue
		}
		settled, merr := n.Journal.MarkAlertDelivered(ctx, alert.ID, claim.Token)
		if merr != nil {
			return delivered, 0, merr
		}
		if settled.Outcome != journal.SettleApplied {
			// Published, but somebody else's row to settle. It is out; the
			// holder records it.
			n.logLeaseLost(settled, alert.ID, Event{Type: EventType(alert.Type)})
			continue
		}
		delivered++
	}
	remaining, err = n.Journal.UndeliveredCount(ctx)
	return delivered, remaining, err
}

// Acknowledge is the operator's release.
//
// It marks the outbox rows and, only when none is left pending, clears the entry
// gate. Delivery recovering is not enough on its own: the alert existed to make a
// human look at something, and "the network came back" is not that human.
//
// It holds the delivery mutex for the same reason the notify path does. The
// release is a read-then-decide: count what is still pending, and clear the gate
// only if that count is zero. A send running alongside it can turn a settled row
// back into a pending one — a096 re-arms a row whose reminder window elapsed —
// and if that lands between the count and the clear, the gate opens while an
// undelivered critical alert exists. Which is the one thing the gate is for.
func (n *Notifier) Acknowledge(ctx context.Context, operator string, ids ...int64) error {
	if strings.TrimSpace(operator) == "" {
		return errors.New("obs: acknowledging an alert requires the operator's identity")
	}
	if n.Journal == nil {
		if n.Gate != nil {
			n.Gate.Clear(execgw.ReasonAlertUndelivered)
		}
		return nil
	}

	n.mu.Lock()
	defer n.mu.Unlock()

	if len(ids) == 0 {
		pending, err := n.Journal.PendingAlerts(ctx, 0)
		if err != nil {
			return err
		}
		for _, alert := range pending {
			ids = append(ids, alert.ID)
		}
	}
	for _, id := range ids {
		if err := n.Journal.AcknowledgeAlert(ctx, id, operator); err != nil &&
			!errors.Is(err, journal.ErrAlertNotFound) {
			return err
		}
	}

	remaining, err := n.Journal.UndeliveredCount(ctx)
	if err != nil {
		return err
	}
	if remaining == 0 && n.Gate != nil {
		n.Gate.Clear(execgw.ReasonAlertUndelivered)
	}
	return nil
}

// eventKey builds the dedupe key for an event.
//
// The default is the event type plus its symbol and attempt id when present:
// those are what make "the same condition" the same. A caller with a better key
// supplies one.
func (n *Notifier) eventKey(e Event) string {
	if key := strings.TrimSpace(e.Key); key != "" {
		return key
	}
	parts := []string{string(e.Type)}
	for _, field := range []string{FieldAttemptID, FieldOrderID, FieldSymbol} {
		if v, ok := e.Fields[field]; ok {
			parts = append(parts, fmt.Sprint(v))
		}
	}
	return strings.Join(parts, "|")
}

func notificationFor(e Event, severity Severity) Notification {
	title := strings.TrimSpace(e.Title)
	if title == "" {
		title = string(e.Type)
	}
	body := e.Body
	if len(e.Fields) > 0 {
		body = strings.TrimSpace(body + "\n" + renderFields(e.Fields))
	}
	return Notification{Type: e.Type, Severity: severity, Title: title, Body: body}
}

func encodeFields(fields map[string]any) string {
	if len(fields) == 0 {
		return ""
	}
	data, err := json.Marshal(fields)
	if err != nil {
		return ""
	}
	return string(data)
}

// renderFields writes the context into the notification body in a stable order,
// because an operator comparing two alerts should not have to diff shuffled lines.
func renderFields(fields map[string]any) string {
	keys := make([]string, 0, len(fields))
	for k := range fields {
		keys = append(keys, k)
	}
	sortStrings(keys)

	var b strings.Builder
	for _, k := range keys {
		fmt.Fprintf(&b, "%s: %v\n", k, fields[k])
	}
	return strings.TrimRight(b.String(), "\n")
}

func sortStrings(s []string) {
	for i := 1; i < len(s); i++ {
		for j := i; j > 0 && s[j] < s[j-1]; j-- {
			s[j], s[j-1] = s[j-1], s[j]
		}
	}
}
