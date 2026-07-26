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
// There is no meta refresh on either. The broker reading is cached for
// holdingsTTL and a page that reloaded itself every two seconds would spend the
// operator's attention on a number that has not moved; the cache's age is printed
// instead, so refreshing is a decision rather than a default.

import "net/http"

type positionsPage struct {
	Nav  string
	Snap positionsView
}

// Refresh is what the head template reads. The dashboard's two-second reload
// belongs to a verification in progress, not to a holdings table.
func (positionsPage) Refresh() bool { return false }

func (c *Console) handlePositions(w http.ResponseWriter, r *http.Request) {
	c.render(w, "positions", positionsPage{Nav: "positions", Snap: c.positions(r.Context())})
}

type historyPage struct {
	Nav  string
	Snap historyView
}

func (historyPage) Refresh() bool { return false }

func (c *Console) handleHistory(w http.ResponseWriter, r *http.Request) {
	c.render(w, "history", historyPage{Nav: "history", Snap: c.history(r.Context())})
}
