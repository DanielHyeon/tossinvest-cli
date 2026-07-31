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
)

type positionsPage struct {
	Nav  string
	Snap positionsView
}

// Refresh is what the head template reads: the positions screen asks the
// browser to reload, at the period RefreshSeconds names.
func (positionsPage) Refresh() bool { return true }

// RefreshSeconds is the reload period: exactly the holdings cache TTL, derived
// from it so the two cannot drift apart — a period under the TTL would be a
// reload that costs broker calls faster than the budget the spec fixes.
func (positionsPage) RefreshSeconds() int { return int(holdingsTTL / time.Second) }

func (c *Console) handlePositions(w http.ResponseWriter, r *http.Request) {
	page := positionsPage{Nav: "positions", Snap: c.positions(r.Context())}
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
	Nav  string
	Snap historyView
}

func (historyPage) Refresh() bool { return false }

func (c *Console) handleHistory(w http.ResponseWriter, r *http.Request) {
	c.render(w, "history", historyPage{Nav: "history", Snap: c.history(r.Context())})
}
