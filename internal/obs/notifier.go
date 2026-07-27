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
	id, err := n.Journal.EnqueueAlert(ctx, record)
	if err != nil {
		return fmt.Errorf("obs: recording a critical alert: %w", err)
	}

	if !n.deliver(ctx, id, e) {
		// Outside deliver, and that is not tidiness: deliver holds n.mu, the
		// escalation announces through a ModeAnnouncer, and an announcer wired to
		// this Notifier would re-enter Notify and deadlock on a mutex Go does not
		// make reentrant.
		n.escalate(ctx, e)
	}
	return nil
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
// it cannot. It reports whether the alert went out.
func (n *Notifier) deliver(ctx context.Context, id int64, e Event) bool {
	// One delivery loop at a time: two goroutines publishing the same backlog
	// would double-send and race on the attempt counter.
	n.mu.Lock()
	defer n.mu.Unlock()

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
			if markErr := n.Journal.MarkAlertDelivered(ctx, id); markErr != nil && n.Log != nil {
				n.Log.Error(EventAlertUndelivered, markErr)
			}
			return true
		}
		lastErr = err
		if markErr := n.Journal.MarkAlertAttemptFailed(ctx, id, err.Error()); markErr != nil && n.Log != nil {
			n.Log.Error(EventAlertUndelivered, markErr)
		}
		if attempt < attempts {
			if !n.wait(ctx) {
				break
			}
		}
	}

	// Out of attempts. The row stays PENDING — it is preserved, not abandoned —
	// and new entries stop.
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
	return false
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
	pending, err := n.Journal.PendingAlerts(ctx, 0)
	if err != nil {
		return 0, 0, err
	}
	for _, alert := range pending {
		if n.Publisher == nil {
			break
		}
		msg := Notification{
			Type:     EventType(alert.Type),
			Severity: SeverityCritical,
			Title:    alert.Title,
			Body:     alert.Body,
		}
		if perr := n.Publisher.Publish(ctx, msg); perr != nil {
			_ = n.Journal.MarkAlertAttemptFailed(ctx, alert.ID, perr.Error())
			continue
		}
		if merr := n.Journal.MarkAlertDelivered(ctx, alert.ID); merr != nil {
			return delivered, 0, merr
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
