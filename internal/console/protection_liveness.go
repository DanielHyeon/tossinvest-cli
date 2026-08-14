package console

// protection_liveness.go decides whether the protection line the console has on
// disk may still be read as the line that is in force (change a077).
//
// A111 makes a flat evaluation a durable heartbeat as well as a state update.
// Every console path consequently applies the same exact 30-second freshness
// rule; stopped engines close it immediately, while unwired and unavailable
// readers fail closed only on the evidence age/integrity rule.

import (
	"strings"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/enginelock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/operatorview"
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
	wired, status := c.readProtectionMarker(now)
	return protectionLivenessAt(wired, status, now)
}

// readProtectionMarker performs the one marker filesystem read for a console
// response. Callers that need a later response-time authority must re-evaluate
// this returned Status rather than reading the filesystem again.
func (c *Console) readProtectionMarker(now time.Time) (bool, enginelock.Status) {
	path := strings.TrimSpace(c.opts.EngineMarker)
	if path == "" {
		return false, enginelock.Status{}
	}
	return true, enginelock.Read(path, now)
}

// protectionLivenessAt re-evaluates a marker status at the response time
// without another filesystem read. Keep the enginelock boundary exact: a
// marker at StaleAfter is still running, while a stopped read is never upgraded
// when the wall clock moves backward.
func protectionLivenessAt(wired bool, status enginelock.Status, now time.Time) protectionLiveness {
	if !wired {
		return protectionLiveness{}
	}
	status.Running = status.Running && !status.RefreshedAt.IsZero() &&
		now.Sub(status.RefreshedAt) <= enginelock.StaleAfter
	return protectionLiveness{Wired: true, Running: status.Running}
}

// exitFreshness decides whether a persisted snapshot may still be read as the
// current protection line.
func exitFreshness(view journal.ExitSnapshotView, asOf time.Time, live protectionLiveness) journal.ExitSnapshotView {
	switch {
	case !live.Wired:
		return operatorview.ApplyExitFreshness(view, asOf, operatorview.ExitLivenessUnwired)
	case !live.Running:
		return operatorview.ApplyExitFreshness(view, asOf, operatorview.ExitLivenessStopped)
	default:
		return operatorview.ApplyExitFreshness(view, asOf, operatorview.ExitLivenessRunning)
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
// label would be a poor trade. An empty cache yields no names — the pre-a077
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
