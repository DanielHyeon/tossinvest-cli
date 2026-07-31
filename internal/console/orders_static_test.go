package console

// orders_static_test.go pins the route exception the orders screen needs and the
// exact size of it (change console-orders-screen, §3).
//
// The route table's account-verb guard fails any path containing "order". That
// ban is right and it stays: /orders/cancel and /order-place are both caught by
// the same loose string test, and looseness is the point. This change opens one
// byte-exact hole in it, and everything here exists so that the hole cannot grow
// without a test failing.

import (
	"net/http"
	"net/url"
	"strings"
	"testing"
)

// TestAPostToTheOrdersScreenIsRefusedByTheReadOnlyWrapper.
//
// This is the runtime half of the exception's condition, and it is the reason
// the `readOnly` wrapper exists at all. The route record carries {Path, Session,
// CSRFGated} and no method, so a static guard asked to confirm "this route is a
// GET" can only look at the CSRF gate — and it would then grant the exception
// BECAUSE the route is unprotected. In that state a POST to /orders is served
// with a session cookie and no CSRF token at all.
func TestAPostToTheOrdersScreenIsRefusedByTheReadOnlyWrapper(t *testing.T) {
	h := newHarness(t)
	h.authenticate(t)

	if resp := h.get(t, "/orders"); resp.StatusCode != http.StatusOK {
		t.Fatalf("GET /orders returned %d, want 200; the screen has to open", resp.StatusCode)
	}

	resp := h.post(t, "/orders", url.Values{})
	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("POST /orders returned %d, want 405; a read route that answers a POST is a route "+
			"whose exception was granted because nothing was protecting it", resp.StatusCode)
	}
	if got := resp.Header.Get("Allow"); !strings.Contains(got, http.MethodGet) {
		t.Errorf("Allow = %q, want it to name GET", got)
	}
}

// TestTheOrdersExceptionAppliesOnlyToTheExactPathThatReadsAndDoesNotAct.
//
// Task 3.5. The exception's condition is three facts and all three are load
// bearing: the byte-exact path, the readOnly wrapper, and the absence of the CSRF
// gate. Break any one and the route goes back to being an ordinary route that
// names an account verb.
func TestTheOrdersExceptionAppliesOnlyToTheExactPathThatReadsAndDoesNotAct(t *testing.T) {
	granted := route{Path: "/orders", Session: true, ReadOnly: true, File: "orders.go"}
	if !routeReadsTheAccountRecord(granted) {
		t.Fatal("the exception is not granted to the route it exists for; the orders screen cannot " +
			"be registered at all")
	}
	for _, tc := range []struct {
		what string
		r    route
	}{
		{"a different path", route{Path: "/positions", Session: true, ReadOnly: true}},
		{"no readOnly wrapper", route{Path: "/orders", Session: true}},
		{"behind the CSRF gate", route{Path: "/orders", Session: true, ReadOnly: true, CSRFGated: true}},
	} {
		if routeReadsTheAccountRecord(tc.r) {
			t.Errorf("the exception was granted with %s; it is granted to a byte-exact path that "+
				"carries the read-only wrapper and is not a state change", tc.what)
		}
	}
}

// TestTheOrdersExceptionDoesNotReachAnyPathBeneathOrBesideIt.
//
// Task 3.6: the exception measures its own size. Each of these is registered as
// a route that reads and is not CSRF-gated — every fact the granted route has
// except the path — and every one of them has to make the guard fail.
//
// It is asserted against the judgement itself rather than against the allowlist's
// contents. "the allowlist does not contain /orders/cancel" measures nothing: it
// would keep passing under a prefix match, under strings.ToLower and under a
// trailing-slash trim, which are the three ways this hole actually grows.
func TestTheOrdersExceptionDoesNotReachAnyPathBeneathOrBesideIt(t *testing.T) {
	for _, path := range []string{
		"/orders/cancel", "/orders/new", "/orders/amend",
		// Case: strings.ToLower on the comparison would let this through, and Go's
		// mux matches paths case-sensitively, so /Orders is a second route.
		"/Orders",
		// Trailing slash: strings.TrimSuffix(p, "/") would let this through, and in
		// Go 1.22+ a trailing-slash pattern is a SUBTREE pattern — /orders/ routes
		// /orders/cancel to that handler.
		"/orders/",
	} {
		r := route{Path: path, Session: true, ReadOnly: true, File: "orders.go"}
		if routeReadsTheAccountRecord(r) {
			t.Errorf("%s was granted the /orders exception; it is one byte-exact path and it is not "+
				"inherited downwards, upwards or across a case change", path)
		}
		if findings := routeFindings(r); len(findings) == 0 {
			t.Errorf("registering %s produced no finding; the guard that bans an order-acting route "+
				"has stopped seeing one", path)
		}
	}
}

// TestTheRouteTableJudgementStillCatchesAnActingRouteThatIsNotOnTheAllowlist.
//
// The positive control for the extraction. A judgement that returns no findings
// for anything would pass every test above.
func TestTheRouteTableJudgementStillCatchesAnActingRouteThatIsNotOnTheAllowlist(t *testing.T) {
	for _, r := range []route{
		{Path: "/order/place", Session: true, File: "x.go"},
		{Path: "/flatten", Session: true, File: "x.go"},
		{Path: "/gate/open", Session: true, File: "x.go"},
		{Path: "/verify/reset", Session: true, File: "x.go"},
		// And a route that carries the readOnly wrapper is not thereby excused:
		// the wrapper is a precondition of the exception, not the exception.
		{Path: "/order/place", Session: true, ReadOnly: true, File: "x.go"},
	} {
		if findings := routeFindings(r); len(findings) == 0 {
			t.Errorf("%s produced no finding", r.Path)
		}
	}
	// And the granted route produces none.
	granted := []route{{Path: "/orders", Session: true, ReadOnly: true, File: "orders.go"}}
	if findings := routeFindings(granted[0]); len(findings) != 0 {
		t.Errorf("the granted /orders route produced findings %v; the screen cannot be registered",
			findings)
	}
}

// TestTradingViewsAreRegisteredReadOnlyAndNothingElseIs.
//
// The exception is worth exactly one route. This is what fails if a second one is
// quietly wrapped in `reading` to borrow it, and what fails if /orders loses the
// wrapper and starts serving POSTs.
func TestTheOrdersRouteIsRegisteredReadOnlyAndNothingElseIs(t *testing.T) {
	found := map[string]bool{"/orders": false, "/positions": false}
	for _, r := range registeredRoutes(t) {
		if _, tradingView := found[r.Path]; tradingView {
			found[r.Path] = true
			if !r.ReadOnly {
				t.Errorf("%s is registered without the readOnly wrapper", r.Path)
			}
			if r.CSRFGated {
				t.Error("/orders is behind the CSRF gate; it is a reading and nobody could open it")
			}
			if !r.Session {
				t.Error("/orders is registered without session0")
			}
			continue
		}
		if r.ReadOnly {
			t.Errorf("%s carries the readOnly wrapper. It is the orders exception's precondition and "+
				"nothing else needs it: every other read route on this console is already covered by "+
				"not naming an account verb", r.Path)
		}
	}
	for path, registered := range found {
		if !registered {
			t.Errorf("%s is not registered", path)
		}
	}
}

// --- the seam allowlist (task 3.7) ------------------------------------------------

// TestTheOrdersSeamIsTheOnlyCapabilityWithVerbExemptionsAndTheyAreEnumerated.
//
// The verb list is shared by the route table and the injection surface, and
// "order" and "conditional" are both on it. The seam this change adds is spelled
// with both, and the tempting repair is to delete them from the list — which
// would let a future /order/place and a future CreateConditionalOrder through at
// the same moment, on the strength of a screen that only reads.
//
// So the list is untouched and the exemption is per field, per spelling, with the
// argument written beside it. This is what fails when the hatch grows: an extra
// name, an exemption on an unenumerated seam, or a reason nobody wrote.
//
// It used to say "exactly one seam", which was a fact about the code stated as a
// rule. console-owns-the-operating-toggles produced the second: a seam whose one
// method writes `engine.automation_gate.enabled`, which the operator-console spec
// requires to be *named* rather than merely tolerated — an audit line that cannot
// say which route turned the engine loose is not an audit line. The guarantee
// that mattered was never the count; it was that every exemption is enumerated
// here with its argument, and that is what this now checks.
func TestTheSeamsWithVerbExemptionsAreEnumeratedAndArgued(t *testing.T) {
	want := map[string]map[string]bool{
		"Orders": {
			"Orders": true, "OrdersReader": true, "OrdersReading": true,
			"OrderRecord": true, "ConditionalRecord": true,
		},
		"Gate":                      {"Gate": true, "GateSwitch": true},
		"AcquireUpdateEngineLock":   {"AcquireUpdateEngineLock": true},
		"CheckUpdateVerifyActivity": {"CheckUpdateVerifyActivity": true},
		"SystemUpdater":             {"SystemUpdater": true, "Install": true},
	}

	for field, cap := range consoleCapabilities {
		if len(cap.VerbExemptions) == 0 {
			continue
		}
		names, enumerated := want[field]
		if !enumerated {
			t.Errorf("console.Options.%s carries verb exemptions and is not enumerated here. A seam "+
				"that needs a forbidden word needs an argument for it in this file, beside the "+
				"others, where the set can be read at a glance", field)
			continue
		}
		for name, why := range cap.VerbExemptions {
			if !names[name] {
				t.Errorf("the %s seam exempts %q, which is not one of the names enumerated for it. "+
					"An exemption is the size of the argument written beside it", field, name)
			}
			if strings.TrimSpace(why) == "" {
				t.Errorf("the exemption for %q has no reason; an unargued exemption is the verb list "+
					"quietly getting shorter", name)
			}
		}
		for name := range names {
			if _, ok := cap.VerbExemptions[name]; !ok {
				t.Errorf("the %s seam no longer exempts %q; either the seam was renamed and this "+
					"list is stale, or the guard is passing for a reason nobody checked", field, name)
			}
		}
	}
	for field := range want {
		if len(consoleCapabilities[field].VerbExemptions) == 0 {
			t.Errorf("%s is enumerated as needing verb exemptions but carries none; a stale "+
				"entry here is an exemption nobody would notice being re-added", field)
		}
	}

	// The shared verb lists themselves are untouched. This is the repair the
	// exemption exists to make unnecessary, and it is worth failing on directly:
	// with "order" gone, a route called /order/place and a seam method called
	// PlaceOrder both stop being caught, and neither failure names this change.
	for _, verb := range []string{"order", "conditional"} {
		found := false
		for _, v := range mutationVerbs {
			if v == verb {
				found = true
			}
		}
		if !found {
			t.Errorf("%q is no longer a mutation verb. The orders seam is handled by an enumerated "+
				"exemption precisely so this list does not have to shrink", verb)
		}
	}
	found := false
	for _, v := range accountVerbs {
		if v == "order" {
			found = true
		}
	}
	if !found {
		t.Error("\"order\" is no longer an account verb; /order/place would be registered in silence")
	}
}
