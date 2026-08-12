package main

// soak_protection_test.go is the HTTP-level evidence for the three
// resident-protection reads (a100 tasks 0.10 (a)).
//
// internal/soak's own tests drive the probes through a stub, so they prove what
// RunCycle does with an answer and nothing about how the answer is obtained.
// Everything between soak.Reads and the broker — the path each read is sent to,
// the query it carries, and the identifier the list hands to the by-id read —
// only exists in cmd/tossctl's adapter, and only a real request can show it.
//
// The server here is a second fixture rather than a case added to
// newSoakServer's switch. Adding cases there would have put a diff hunk inside
// an existing function, which the Function Logic Map gate reads as an edit to
// that function's logic; a test fixture is not evidence about production logic,
// and the gate should not be made to say it is.

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"strings"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/soak"
	"github.com/JungHoonGhae/tossinvest-cli/internal/testenv"
)

// newProtectionSoakServer answers everything newSoakServer answers, plus the
// three conditional reads. The conditional it returns is the one the by-id read
// then has to ask for.
func newProtectionSoakServer(t *testing.T) *soakServer {
	t.Helper()
	s := &soakServer{}
	s.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.mu.Lock()
		s.requests = append(s.requests, r.Method+" "+r.URL.Path)
		if r.URL.Path == "/api/v1/conditional-orders" || r.URL.Path == "/api/v1/sellable-quantity" {
			s.orderListQueries = append(s.orderListQueries, r.URL.Query())
		}
		s.mu.Unlock()

		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.URL.Path == "/oauth2/token":
			fmt.Fprint(w, `{"access_token":"soak-token","expires_in":3600,"token_type":"Bearer"}`)
		case r.URL.Path == "/api/v1/accounts":
			fmt.Fprint(w, `{"result":[{"accountNo":"123-45-678901","accountSeq":7,"accountType":"BROKERAGE"}]}`)
		case r.URL.Path == "/api/v1/buying-power":
			fmt.Fprint(w, `{"result":{"cashBuyingPower":"1000000","currency":"KRW"}}`)
		case r.URL.Path == "/api/v1/holdings":
			fmt.Fprint(w, `{"result":{"items":[{"symbol":"005930","quantity":"10","lastPrice":"70000"}]}}`)
		case r.URL.Path == "/api/v1/orders":
			fmt.Fprint(w, `{"result":{"orders":[{"orderId":"ord-1","symbol":"005930","status":"OPEN",`+
				`"quantity":"1","price":"70000"}],"nextCursor":"","hasNext":false}}`)
		case r.URL.Path == "/api/v1/conditional-orders":
			// Only the OPEN group carries one, so a walk that read a single group
			// and stopped would be visible as a missing request rather than as a
			// silently equal answer.
			if r.URL.Query().Get("status") != "OPEN" {
				fmt.Fprint(w, `{"result":{"conditionalOrders":[],"nextCursor":"","hasNext":false}}`)
				return
			}
			fmt.Fprint(w, `{"result":{"conditionalOrders":[{"conditionalOrderId":"cond-1",`+
				`"clientOrderId":"cli-1","type":"SINGLE","status":"WATCHING","symbol":"005930",`+
				`"market":"KRX","quantity":"1","orderType":"MARKET",`+
				`"first":{"orderSide":"SELL","triggerPrice":"69000"}}],"nextCursor":"","hasNext":false}}`)
		case strings.HasPrefix(r.URL.Path, "/api/v1/conditional-orders/"):
			fmt.Fprint(w, `{"result":{"conditionalOrderId":"cond-1","clientOrderId":"cli-1",`+
				`"type":"SINGLE","status":"WATCHING","symbol":"005930","market":"KRX","quantity":"1",`+
				`"orderType":"MARKET","first":{"orderSide":"SELL","triggerPrice":"69000"}}}`)
		case r.URL.Path == "/api/v1/sellable-quantity":
			fmt.Fprint(w, `{"result":{"sellableQuantity":"10"}}`)
		case strings.HasPrefix(r.URL.Path, "/api/v1/orders/"):
			fmt.Fprint(w, `{"result":{"orderId":"ord-1","symbol":"005930","status":"OPEN","quantity":"1"}}`)
		case r.URL.Path == "/api/v1/prices":
			fmt.Fprint(w, `{"result":[{"symbol":"005930","lastPrice":"70000","currency":"KRW"}]}`)
		default:
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprint(w, `{"message":"not found"}`)
		}
	}))
	t.Cleanup(s.Server.Close)
	return s
}

// TestSoakRunSurveysTheProtectionReads: the wiring reaches all three and records
// them as successes. Without this the adapter could be pointed at the wrong path
// and every unit test would still pass, because they answer through a stub.
func TestSoakRunSurveysTheProtectionReads(t *testing.T) {
	configDir := testenv.Isolate(t)
	srv := newProtectionSoakServer(t)
	pointSoakAt(t, srv, filepath.Join(configDir, "token.json"))

	if _, _, err := runCLI(t, "--config-dir", configDir, "soak", "run", "--cycles", "1", "--interval", "0"); err != nil {
		t.Fatalf("soak run: %v", err)
	}

	cycles, err := soak.LoadCycles(filepath.Join(configDir, soak.FileName))
	if err != nil {
		t.Fatalf("LoadCycles: %v", err)
	}
	if len(cycles) != 1 {
		t.Fatalf("recorded %d cycle(s), want 1", len(cycles))
	}

	for _, want := range []string{
		soak.EndpointConditionalOrders,
		soak.EndpointConditionalByID,
		soak.EndpointSellableQuantity,
	} {
		found := false
		for _, e := range cycles[0].Endpoints {
			if e.Endpoint != want {
				continue
			}
			found = true
			if !e.OK {
				t.Errorf("%s: OK = false (%s / %s)", want, e.Class, e.Error)
			}
		}
		if !found {
			t.Errorf("the cycle recorded nothing for %s", want)
		}
	}
}

// TestSoakWalksBothConditionalGroups. The gateway's List and its post-cancel
// confirmation each read OPEN and then CLOSED (protectionofficial/gateway.go:113,
// :189). A survey that proved one group would attest an endpoint the engine uses
// in a way nobody measured.
func TestSoakWalksBothConditionalGroups(t *testing.T) {
	configDir := testenv.Isolate(t)
	srv := newProtectionSoakServer(t)
	pointSoakAt(t, srv, filepath.Join(configDir, "token.json"))

	if _, _, err := runCLI(t, "--config-dir", configDir, "soak", "run", "--cycles", "1", "--interval", "0"); err != nil {
		t.Fatalf("soak run: %v", err)
	}

	groups := map[string]bool{}
	sellableSymbol := ""
	for _, q := range srv.orderLists() {
		if status := q.Get("status"); status != "" {
			groups[status] = true
		}
		if symbol := q.Get("symbol"); symbol != "" {
			sellableSymbol = symbol
		}
	}
	for _, want := range []string{"OPEN", "CLOSED"} {
		if !groups[want] {
			t.Errorf("the conditional-order list was never read with status=%s", want)
		}
	}
	if sellableSymbol != "005930" {
		t.Errorf("sellable-quantity asked for symbol %q, want the surveyed symbol", sellableSymbol)
	}
}

// TestSoakReadsTheConditionalTheListReturned. The by-id read has to use an
// identifier the list actually handed back; a hard-coded or empty one would
// attest an endpoint against an order that does not exist.
func TestSoakReadsTheConditionalTheListReturned(t *testing.T) {
	configDir := testenv.Isolate(t)
	srv := newProtectionSoakServer(t)
	pointSoakAt(t, srv, filepath.Join(configDir, "token.json"))

	if _, _, err := runCLI(t, "--config-dir", configDir, "soak", "run", "--cycles", "1", "--interval", "0"); err != nil {
		t.Fatalf("soak run: %v", err)
	}

	want := "GET /api/v1/conditional-orders/" + url.PathEscape("cond-1")
	for _, req := range srv.seen() {
		if req == want {
			return
		}
	}
	t.Errorf("no request to %s; the by-id read did not use the identifier the list returned (saw %v)",
		want, srv.seen())
}

// TestSoakStillIssuesNoMutatingRequestWithTheProtectionReads. The conditional
// endpoints share their paths with create, modify and cancel, so the read-only
// property has to be re-established against a server that answers them.
func TestSoakStillIssuesNoMutatingRequestWithTheProtectionReads(t *testing.T) {
	configDir := testenv.Isolate(t)
	srv := newProtectionSoakServer(t)
	pointSoakAt(t, srv, filepath.Join(configDir, "token.json"))

	if _, _, err := runCLI(t, "--config-dir", configDir, "soak", "run", "--cycles", "2", "--interval", "0"); err != nil {
		t.Fatalf("soak run: %v", err)
	}

	sawConditional := false
	for _, req := range srv.seen() {
		method, path, _ := strings.Cut(req, " ")
		if strings.Contains(path, "/conditional-orders") {
			sawConditional = true
		}
		if method == http.MethodPost && path == "/oauth2/token" {
			continue
		}
		if method != http.MethodGet {
			t.Errorf("the soak issued %s — it is read-only and must issue nothing but GETs "+
				"(and the OAuth token exchange)", req)
		}
		if strings.Contains(path, "/cancel") || strings.Contains(path, "/modify") {
			t.Errorf("the soak reached %s; that is a mutation path", req)
		}
	}
	if !sawConditional {
		t.Error("no conditional-order request happened; this test proves nothing without one")
	}
}
