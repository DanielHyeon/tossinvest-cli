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
)

type positionsPage struct {
	Nav  string
	Snap positionsView
	// CSRF rides along for the per-row designation forms (settings.go); the
	// page itself stays a GET reading.
	CSRF string
	// CanDesignate reports the settings seam is wired and readable, so the
	// designation buttons render at all.
	CanDesignate bool
}

// Refresh is what the head template reads: the positions screen asks the
// browser to reload, at the period RefreshSeconds names.
func (positionsPage) Refresh() bool { return true }

// RefreshSeconds is the reload period: exactly the holdings cache TTL, derived
// from it so the two cannot drift apart — a period under the TTL would be a
// reload that costs broker calls faster than the budget the spec fixes.
func (positionsPage) RefreshSeconds() int { return int(holdingsTTL / time.Second) }

func (c *Console) handlePositions(w http.ResponseWriter, r *http.Request) {
	page := positionsPage{Nav: "positions", Snap: c.positions(r.Context()), CSRF: c.csrf}
	if c.opts.Settings != nil {
		if block, _, err := c.opts.Settings.Load(); err == nil {
			page.CanDesignate = true
			for i := range page.Snap.Rows {
				page.Snap.Rows[i].Designated = block.Included(page.Snap.Rows[i].Symbol)
			}
		}
	}
	c.render(w, "positions", page)
}

type historyPage struct {
	Nav  string
	Snap historyView
}

func (historyPage) Refresh() bool { return false }

func (c *Console) handleHistory(w http.ResponseWriter, r *http.Request) {
	c.render(w, "history", historyPage{Nav: "history", Snap: c.history(r.Context())})
}
