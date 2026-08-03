package console

// protection_liveness.go decides whether the protection line the console has on
// disk may still be read as the line that is in force (change a061).
//
// # The bug this file exists to correct
//
// The engine writes exit_states.last_observed_at from exactly one place —
// journal's record() — and the observation loop returns before reaching it
// whenever the judgement moved nothing:
//
//	snapshot = snapshot.ChangedFromState(…)
//	if !snapshot.Changed {
//	        return nil          // internal/app/engine/exitloop.go
//	}
//	return o.record(…)
//
// So that column is the last time the line *changed*, not the last time it was
// looked at. The screens read it as an observation heartbeat and bounded it by
// holdingsTTL — the broker cache's TTL, thirty seconds, a number with no
// relationship to a protection line's lifetime. The result was that a position
// trading below its recorded high-water, judged correctly every five seconds by
// a healthy engine, went blank thirty seconds after its last change and stayed
// blank all day. On 2026-08-03 that was every managed position in the account.
//
// # What replaced it
//
// The five values are a state, not a measurement. The last judgement fixed them;
// the next judgement is what moves them. Time passing does not make them wrong.
// Three things make them untrustworthy and this file asks about the first two —
// the third (lifecycle generation, integrity) was already being asked:
//
//	nobody is maintaining them   no engine is running against this journal
//	this position is excluded    an active exit-snapshot quarantine, which makes
//	                             the exit loop refuse the position outright
//
// An unwired marker is neither. The operator-console spec forbids drawing it as
// a stopped engine, so a console that cannot see the engine keeps the judgement
// it made before a061 rather than being told something it cannot know.

import (
	"strings"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/enginelock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
)

// protectionLiveness answers a narrower question than engineView: is anything
// maintaining the protection lines this console draws?
type protectionLiveness struct {
	// Wired reports that a marker path was supplied at all.
	Wired bool
	// Running reports a marker touched inside enginelock.StaleAfter. A crashed
	// engine therefore reads as running for at most that long, which is the trade
	// engineproc.go's header already accepted for the status line — and the strip
	// at the top of the same page shows the operator the very same verdict.
	Running bool
}

func (c *Console) protectionLiveness(now time.Time) protectionLiveness {
	path := strings.TrimSpace(c.opts.EngineMarker)
	if path == "" {
		return protectionLiveness{}
	}
	return protectionLiveness{Wired: true, Running: enginelock.Read(path, now).Running}
}

// exitFreshness decides whether a persisted snapshot may still be read as the
// current protection line.
func exitFreshness(view journal.ExitSnapshotView, asOf time.Time, live protectionLiveness) journal.ExitSnapshotView {
	switch {
	case !live.Wired:
		// This console cannot see the engine, and an unwired marker must not be
		// drawn as a stopped one. Keep the judgement this screen made before
		// a061: conservative, and wrong for a reason that is at least unchanged.
		return view.WithFreshness(asOf, holdingsTTL)
	case !live.Running:
		view.Stale, view.StaleReason = true, "engine_not_running"
		return view
	default:
		// An engine is running against this journal. The snapshot's age is
		// evidence of nothing; its integrity still is.
		return view.WithIntegrity(asOf)
	}
}

// Quarantined reports that the engine has taken this position out of judgement.
func (r positionRow) Quarantined() bool { return r.Quarantine != nil }

// holdingNames indexes the stock names the console has already read, by the same
// (market, symbol) key the positions join uses.
//
// The name is a property of the instrument and not of the account, so a holding
// read for one account names the same symbol for a ledger row belonging to
// another. The market has to stay in the key for that to hold: "005930" on KR
// and a US listing of the same ticker are different instruments.
//
// It peeks rather than gets. The caller (/position-management) makes no broker
// call today, and a screen that started spending the §0.4 rate budget for a
// label would be a poor trade. An empty cache yields no names — the pre-a061
// rendering — and never an invented one.
func (c *Console) holdingNames(now time.Time) map[string]string {
	rows := c.holdings.peek(now).Rows
	out := make(map[string]string, len(rows))
	for _, h := range rows {
		if name := strings.TrimSpace(h.Name); name != "" {
			out[symbolKey(h.MarketType, h.Symbol)] = name
		}
	}
	return out
}
