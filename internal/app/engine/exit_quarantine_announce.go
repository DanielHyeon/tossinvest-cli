package engine

// exit_quarantine_announce.go raises the alert for the moment a stored exit
// snapshot generation is quarantined (change a074, exit-policy "판정 격리의
// 생성은 그 순간에 관측된다").
//
// # Why the refusal alert was not enough
//
// A quarantine already produced an alert: EventExitJudgementRefused, on the
// cycle after the one that created it. Reading the operational ledger on
// 2026-08-04 showed what that costs.
//
//	quarantined_at        alert_outbox.created_at   gap
//	2026-08-03T09:03:40Z  09:03:45Z                 5s
//	2026-08-03T13:45:25Z  13:45:30Z                 5s
//	2026-08-03T14:53:11Z  14:53:16Z                 5s
//
// Five seconds is one observation interval. The quarantine is written inside the
// judgement transaction, the transaction returns ErrExitSnapshotQuarantined, and
// nothing recognised that sentinel — so the creating cycle produced no record at
// all and the *next* working set reported the consequence.
//
// The gap is the smaller half of the problem. The refusal alert is latched per
// position (o.refused), so a position already refused for some other reason
// could be quarantined without producing a single line, ever. And it carries
// none of the quarantine's identity, so nothing in it tells an operator that
// this one needs a human to lift it (a079's release screen) rather than waiting
// for the next quote.
//
// # Why the latch is keyed on the version
//
// The outbox deduplicates on event_key, so a repeated announcement makes no
// second row. What it does do is call deliver() again — which, with no publisher
// wired, means a gate block and an error line every five seconds forever. The
// in-memory latch is the same device o.refused and o.delayAlerted already are.
// Keying it on position|generation|version means a release followed by a
// re-quarantine (a new version) announces again, which is right: that is a new
// fact, and it is the one that says the operator's repair did not hold.

import (
	"context"
	"strconv"

	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/obs"
)

// quarantineAnnouncementKey identifies one quarantine row.
func quarantineAnnouncementKey(positionID string, generation, version int64) string {
	return positionID + "|" + strconv.FormatInt(generation, 10) + "|" + strconv.FormatInt(version, 10)
}

// announceQuarantine reports a quarantine this cycle created.
//
// It is called only from the three creation sites. The read path — a working set
// that finds an already-active quarantine — deliberately does not call it: that
// is an observation of an existing fact, it repeats every cycle, and it is
// already reported by the judgement-refused alert.
func (o *ExitObserver) announceQuarantine(ctx context.Context, p journal.Position,
	q journal.ExitSnapshotQuarantine) {
	if o.quarantineAnnounced == nil {
		// Lazily, so NewExitObserver does not have to be edited to add a map whose
		// zero value already reads correctly.
		o.quarantineAnnounced = map[string]bool{}
	}
	key := quarantineAnnouncementKey(p.ID, q.PositionGeneration, q.Version)
	if o.quarantineAnnounced[key] {
		return
	}
	o.quarantineAnnounced[key] = true
	o.alert(ctx, obs.Event{
		Type:  obs.EventExitSnapshotQuarantined,
		Key:   string(obs.EventExitSnapshotQuarantined) + "|" + key,
		Title: p.Symbol + " is no longer being judged: its exit snapshot generation was quarantined",
		Body: "the stored protection state could not be resolved into one verified candidate, so this " +
			"generation is refused outright — the position is not evaluated at all, its stop included, " +
			"until an operator lifts the quarantine. The stored baseline is preserved and lifting it " +
			"does not change any of it: " + q.Reason + ": " + q.Evidence,
		Fields: map[string]any{
			obs.FieldSymbol:       p.Symbol,
			"market":              p.Market,
			"position_id":         p.ID,
			"position_generation": q.PositionGeneration,
			"quarantine_version":  q.Version,
			"reason":              q.Reason,
			"evidence":            q.Evidence,
			"quarantined_at":      q.QuarantinedAt,
		},
	})
}

// announceQuarantineFromLedger reports a quarantine the *journal* created inside
// a judgement transaction, which the caller only learns about as a sentinel
// error.
//
// The row is read back rather than carried out through the error, so
// RecordExitJudgementResult and quarantineExitSnapshotTx stay untouched (design
// D3). The read is correct because the journal commits the quarantine before
// returning the sentinel (exit_state.go), and because the generation the journal
// quarantines is the one snapshotContext filled from p.InstanceSeq — the same
// number this read asks for.
//
// A failed read is silent. It must not change what the caller does with the
// error it already has: this is observation, and observation that alters an
// outcome is not observation.
func (o *ExitObserver) announceQuarantineFromLedger(ctx context.Context, p journal.Position) {
	q, active, err := o.opts.Journal.ActiveExitSnapshotQuarantine(ctx, p.ID, p.InstanceSeq)
	if err != nil || !active {
		return
	}
	o.announceQuarantine(ctx, p, q)
}
