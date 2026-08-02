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

	"github.com/JungHoonGhae/tossinvest-cli/internal/verifylive"
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
	snap := c.orders(r.Context(), filterChoiceFrom(r))
	page := ordersPage{
		chrome: c.chromeFor("orders", verifylive.MarketKR, snap.Broker.freshness()),
		Snap:   snap,
	}
	page.Refresh = true
	// This screen reloads itself, so a native <details> would close on the next
	// tick. The folds are URL-driven instead (change a055 §6, explain.go).
	page.Explain = explainFrom(r)
	c.render(w, "orders", page)
}

type ordersPage struct {
	chrome
	Snap ordersView
}

// ordersScreen is the orders view plus this screen's fold state.
//
// The "ordercounts" and "ordertable" sub-templates are invoked with the view as
// their dot, so `$` inside them is the view and not the page — and both hold a
// disclosure that has to be URL-driven on a screen that reloads itself. Embedding
// keeps every existing `.Rows`, `.OpenLive`, `.Broker` reference working and adds
// the one field (change a055 §6).
type ordersScreen struct {
	ordersView
	Explain explainState
}

// Screen is what those two sub-templates are given.
func (p ordersPage) Screen() ordersScreen {
	return ordersScreen{ordersView: p.Snap, Explain: p.Explain}
}

// RefreshSeconds is the reload period: the orders cache TTL, derived from the
// constant rather than written again (positionsPage's precedent). A period
// shorter than the TTL would be a reload that costs broker calls faster than the
// budget allows; a period equal to it holds one open tab to the three calls per
// TTL the spec fixes for this screen.
//
// Refresh is the embedded chrome's field, set by the handler above.
func (ordersPage) RefreshSeconds() int { return int(ordersTTL / time.Second) }
