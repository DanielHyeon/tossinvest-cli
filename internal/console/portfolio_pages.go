package console

// portfolio_pages.go is the two dashboard handlers (change
// add-operator-dashboard, task 2.3).
//
// Both are GET readings and neither is registered behind the CSRF gate, because
// there is nothing on either page to submit. That is asserted from both
// directions in static_test.go: a state-changing route without the gate fails,
// and a read route behind it fails too — a screen that only answered POSTs would
// be a screen nobody can open.
//
// The positions screen reloads itself at the holdings cache TTL — no faster
// (change refresh-positions-screen; spec "rate budget 보호"). Each reload is an
// ordinary lazy request, so an open tab costs at most one holdings call per TTL
// and a verification in progress still suspends the broker call entirely. The
// history screen renders frozen values and stays manual.

import (
	"net/http"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/operatorview"
	"github.com/JungHoonGhae/tossinvest-cli/internal/verifylive"
)

type positionsPage struct {
	chrome
	Snap positionsView
}

// brokerStateView is the broker-cache banner's data plus this screen's fold
// state.
//
// The "brokerstate" template is invoked with the holdings snapshot as its dot,
// so `$` inside it is that snapshot and not the page — and the fold in it needs
// the page's explain state. Embedding keeps every existing `.Wired`, `.Held`,
// `.TakenAt` reference working and adds the one field.
type brokerStateView struct {
	holdingsSnapshot
	Explain explainState
}

// BrokerState is what the banner template is given.
func (p positionsPage) BrokerState() brokerStateView {
	return brokerStateView{holdingsSnapshot: p.Snap.Holdings, Explain: p.Explain}
}

// RefreshSeconds is the reload period: exactly the holdings cache TTL, derived
// from it so the two cannot drift apart — a period under the TTL would be a
// reload that costs broker calls faster than the budget the spec fixes.
//
// Refresh is the embedded chrome's field, set by the handler below.
func (positionsPage) RefreshSeconds() int { return int(holdingsTTL / time.Second) }

func (c *Console) handlePositions(w http.ResponseWriter, r *http.Request) {
	snap := c.positions(r.Context())
	page := positionsPage{
		chrome: c.chromeFor("positions", verifylive.MarketKR, snap.Holdings.freshness()),
		Snap:   snap,
	}
	page.Refresh = true
	// Same reason as the orders screen: a reloading screen cannot keep a native
	// fold open (change a055 §6, explain.go).
	page.Explain = explainFrom(r)
	attachPositionExitLines(page.Snap.Rows, c.now())
	if c.opts.Settings != nil {
		if block, _, err := c.opts.Settings.Load(); err == nil {
			// One Load stamps both lists: two reads could return two different
			// snapshots and draw a row that is on neither or on both.
			for i := range page.Snap.Rows {
				page.Snap.Rows[i].Designated = block.Included(page.Snap.Rows[i].Symbol)
				page.Snap.Rows[i].Excluded = block.Excludes(page.Snap.Rows[i].Symbol)
			}
		}
	}
	c.render(w, "positions", page)
}

// attachPositionExitLines selects the already-persisted effective snapshot for
// display. Freshness is evaluated at the screen boundary; no exit value is
// recomputed here or in the template.
func attachPositionExitLines(rows []positionRow, asOf time.Time) {
	for i := range rows {
		row := &rows[i]
		if !row.HasExit {
			continue
		}
		snapshot := row.Exit.Snapshot.WithFreshness(asOf, holdingsTTL)
		source := operatorview.Source{
			UnknownReason:     snapshot.UnknownReason,
			StaleReason:       snapshot.StaleReason,
			RemainingQuantity: row.JournalQuantity,
			EffectiveSource:   "persisted effective snapshot",
		}
		if snapshot.Snapshot != nil {
			source.Snapshot = &snapshot.Snapshot.Line
			source.ObservationSource = snapshot.Snapshot.ObservationSource
			source.ObservedAt = snapshot.Snapshot.ObservedAt
		}
		row.ExitLine = operatorview.BuildExitLine(source)
	}
}

type historyPage struct {
	chrome
	Snap historyView
}

func (c *Console) handleHistory(w http.ResponseWriter, r *http.Request) {
	c.render(w, "history", historyPage{
		chrome: c.chromeOnRequest("history"),
		Snap:   c.history(r.Context()),
	})
}
