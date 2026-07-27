package console

// orders_page.go is the orders screen's route and handler (change
// console-orders-screen, §5).
//
// There is no form on this page and no POST route beside it. Cancelling an order
// is `tossctl`'s job and the boundary is the operator-console spec's, not a
// convenience this screen is missing. There is likewise no confirmation friction
// of any kind — no typed string, no second click, no extra approval — because
// there is nothing here to confirm (user instruction, 2026-07-27).

import (
	"net/http"
	"time"
)

// registerOrders puts the orders screen on the mux.
//
// session0 first, then `readOnly`: the session gate answers before the method
// gate, so an unauthenticated POST is refused as unauthenticated rather than
// told which methods this path accepts.
//
// It registers from this file rather than from console.go's routes() for the
// reason console-operator-overview established — the route table's static guards
// have to see a route wherever it is declared — and this is the route those
// guards grant their one byte-exact exception to.
func (c *Console) registerOrders(mux *http.ServeMux) {
	mux.HandleFunc("/orders", c.session0(c.readOnly(c.handleOrders)))
}

// handleOrders renders the screen.
func (c *Console) handleOrders(w http.ResponseWriter, r *http.Request) {
	c.render(w, "orders", ordersPage{
		Nav:  "orders",
		Snap: c.orders(r.Context(), filterChoiceFrom(r)),
	})
}

type ordersPage struct {
	Nav  string
	Snap ordersView
}

// Refresh and RefreshSeconds: the screen reloads itself at the orders cache TTL,
// derived from the constant rather than written again (positionsPage's
// precedent). A period shorter than the TTL would be a reload that costs broker
// calls faster than the budget allows; a period equal to it holds one open tab to
// the two calls per TTL the spec fixes for this screen.
func (ordersPage) Refresh() bool       { return true }
func (ordersPage) RefreshSeconds() int { return int(ordersTTL / time.Second) }
