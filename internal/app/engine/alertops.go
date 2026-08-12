package engine

// alertops.go is the operator's half of a098: reading the critical-alert backlog
// and acknowledging a row in it.
//
// # Why it runs inside the engine
//
// The entry gate is a map behind a mutex in this process (execgw/retry.go:495-510),
// and the one normal path that clears it is obs.Notifier.Acknowledge
// (notifier.go:840). A CLI that wrote to the ledger from its own process would
// leave the operator having acknowledged the backlog while entry stayed shut
// until a restart — and "restart to recover" is not an operator surface (design
// D7.1).
//
// So the surface is a thin projection over the ledger plus a delegation to the
// notifier. Nothing here re-implements the acknowledge rules: the ledger already
// refuses a blank operator (outbox.go:482) and already stores the name (:377), and
// the notifier already decides when the latch comes off. Rebuilding any of that
// would make a second place for it to be wrong.
//
// # What it deliberately does not project
//
// PendingAlert carries no title, body, payload, event key or last error. The
// listing is read wherever operator output is read (안전 불변식 8), and the source
// query loads the body and the payload whether the caller wants them or not
// (outbox.go alertSelect). Returning journal.Alert would therefore publish an
// alert's contents through a command whose job is to count and age rows.
//
// The claim token is absent one level down — journal.Alert has no such field, and
// its doc says why: it is proof of ownership, and anyone who reads it can settle
// somebody else's send.

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/obs"
)

// PendingAlert is one undelivered critical alert, as an operator sees it.
//
// The four columns task 4.4 asks for are id, age, attempt count and the lease —
// enough to tell a row nobody is sending from one a sender is holding, which is
// the whole triage question. Everything that would answer "what did it say" is
// left out on purpose; that answer is in the alert the operator already received,
// or in the ledger for someone with ledger access.
type PendingAlert struct {
	ID int64 `json:"id"`
	// Type is the event type, not its rendered text: "engine.exit_proposal_refused"
	// says which condition is unsent without quoting what was written about it.
	Type string `json:"type"`
	// AgeSeconds is how long the row has been waiting. Seconds rather than a
	// timestamp because this crosses a process boundary and the engine owns the
	// clock — two processes formatting "now" is two chances to disagree.
	AgeSeconds int64 `json:"age_seconds"`
	Attempts   int   `json:"attempts"`
	// ClaimedBy is the sender holding the delivery lease, empty when nobody is.
	ClaimedBy string `json:"claimed_by,omitempty"`
	// ClaimAgeSeconds is how long that lease has been held.
	ClaimAgeSeconds int64 `json:"claim_age_seconds,omitempty"`
	// ClaimExpired says the holder's lease has run out, so the next cycle will
	// take the row over. Without it a stalled sender and a working one look the
	// same: both have a ClaimedBy.
	ClaimExpired bool `json:"claim_expired,omitempty"`
}

// AcknowledgeResult is what the operator learns from acknowledging.
//
// EntryBlocks is the point of it. Acknowledging every undelivered row clears
// exactly one latch — the backlog one — and an operator who is told "done" while
// entry is still shut for a different reason will restart the engine looking for
// a fault that is being reported correctly (task 4.4, R17).
type AcknowledgeResult struct {
	RemainingUndelivered int      `json:"remaining_undelivered"`
	EntryBlocks          []string `json:"entry_blocks,omitempty"`
}

// AlertOperations is the engine-side implementation of both commands.
type AlertOperations struct {
	journal  *journal.Journal
	notifier *obs.Notifier
	gate     *execgw.EntryGate
	clock    clock.Clock
}

// AlertOperations builds the operator surface against this engine's own handles.
//
// It refuses rather than degrades when one is missing: a surface built on a nil
// gate would accept an acknowledgement, write it to the ledger, and leave entry
// blocked — the exact failure D7.1 exists to prevent, arriving through a nil
// check instead of through a second process.
func (c *Context) AlertOperations() (*AlertOperations, error) {
	if c == nil || c.Journal == nil || c.Notifier == nil || c.Entry == nil {
		return nil, fmt.Errorf("%w: the alert operator surface needs the journal, the notifier "+
			"and the entry gate", ErrRuntimeUnavailable)
	}
	return &AlertOperations{
		journal:  c.Journal,
		notifier: c.Notifier,
		gate:     c.Entry,
		clock:    c.clock(),
	}, nil
}

// Pending lists the undelivered critical alerts, oldest id first.
//
// It asks for all of them rather than a page. A capped listing would report a
// backlog of "200" for a backlog of thousands, and an operator deciding whether
// the transport is broken reads that number as the answer — a silent cap is worse
// here than a long response on a socket only this host can open.
func (o *AlertOperations) Pending(ctx context.Context) ([]PendingAlert, error) {
	rows, err := o.journal.PendingAlerts(ctx, 0)
	if err != nil {
		return nil, fmt.Errorf("engine: listing the critical alert backlog: %w", err)
	}
	now := o.clock.Now()
	out := make([]PendingAlert, 0, len(rows))
	for _, row := range rows {
		out = append(out, projectPendingAlert(row, now))
	}
	return out, nil
}

// projectPendingAlert is the narrowing, in one place so there is one place to
// audit against 불변식 8.
func projectPendingAlert(row journal.Alert, now time.Time) PendingAlert {
	out := PendingAlert{
		ID:         row.ID,
		Type:       row.Type,
		AgeSeconds: elapsedSeconds(row.CreatedAt, now),
		Attempts:   row.Attempts,
		ClaimedBy:  row.ClaimedBy,
	}
	if row.ClaimedAt != nil {
		out.ClaimAgeSeconds = elapsedSeconds(*row.ClaimedAt, now)
	}
	if row.ClaimExpiresAt != nil {
		out.ClaimExpired = now.After(*row.ClaimExpiresAt)
	}
	return out
}

// elapsedSeconds never reports a negative age. A row stamped in the future is a
// clock that moved, and "-4" in an age column reads as a bug in the listing
// rather than as what it is.
func elapsedSeconds(from, now time.Time) int64 {
	d := now.Sub(from)
	if d < 0 {
		return 0
	}
	return int64(d / time.Second)
}

// Acknowledge records that a person has seen these alerts.
//
// With no ids it acknowledges the whole backlog, which is Notifier.Acknowledge's
// own convention — kept rather than re-decided, because this function is a
// delegation and a delegation that changes the meaning of its arguments is a
// second implementation wearing the first one's name.
//
// The operator's name is the caller's to supply and neither this function nor the
// notifier invents one; the ledger refuses a blank (outbox.go:482) and stores what
// it is given (:377). That refusal is the audit trail's only guard, so a caller
// passing a fixed string defeats it while passing every check here.
func (o *AlertOperations) Acknowledge(ctx context.Context, operator string,
	ids ...int64) (AcknowledgeResult, error) {
	if err := o.notifier.Acknowledge(ctx, operator, ids...); err != nil {
		return AcknowledgeResult{}, fmt.Errorf("engine: acknowledging critical alerts: %w", err)
	}
	remaining, err := o.journal.UndeliveredCount(ctx)
	if err != nil {
		return AcknowledgeResult{}, fmt.Errorf("engine: counting undelivered alerts: %w", err)
	}
	return AcknowledgeResult{
		RemainingUndelivered: remaining,
		EntryBlocks:          entryBlockReasons(o.gate),
	}, nil
}

// entryBlockReasons lists the latches still standing, sorted so two runs of the
// same command say the same thing — Blocks() hands back a map.
func entryBlockReasons(gate *execgw.EntryGate) []string {
	if gate == nil {
		return nil
	}
	blocks := gate.Blocks()
	if len(blocks) == 0 {
		return nil
	}
	// The reason codes travel; the details do not. A latch detail is written by
	// whoever set it and this one is read by whoever runs the command.
	out := make([]string, 0, len(blocks))
	for reason := range blocks {
		out = append(out, string(reason))
	}
	sort.Strings(out)
	return out
}
