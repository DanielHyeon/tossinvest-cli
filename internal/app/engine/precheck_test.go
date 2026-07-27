package engine_test

// precheck_test.go covers the derived cancel/amend pre-check (task 4.1).
//
// Two properties are load-bearing and each has its own test:
//
//  1. the answer comes from the official single-order read, derived through the
//     brokerstate priority table — not from a WTS session;
//  2. the pre-check can never abort a mutation (§0.3), no matter how the read
//     fails.

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/app/engine"
	"github.com/JungHoonGhae/tossinvest-cli/internal/official"
	"github.com/JungHoonGhae/tossinvest-cli/internal/orderintent"
)

// precheckServer answers the token/account handshake plus one single-order read
// whose body the test controls, and records the order paths it was asked for.
func precheckServer(t *testing.T, order func(w http.ResponseWriter, r *http.Request)) (*httptest.Server, func() []string) {
	t.Helper()
	var mu sync.Mutex
	var seen []string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.URL.Path == "/oauth2/token":
			_, _ = w.Write([]byte(`{"access_token":"AT","expires_in":3600,"token_type":"Bearer"}`))
		case r.URL.Path == "/api/v1/accounts":
			_, _ = w.Write([]byte(`{"result":[{"accountNo":"123-45","accountSeq":7,"accountType":"BROKERAGE"}]}`))
		case strings.HasPrefix(r.URL.Path, "/api/v1/orders/"):
			mu.Lock()
			seen = append(seen, r.Method+" "+r.URL.Path)
			mu.Unlock()
			order(w, r)
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)

	return srv, func() []string {
		mu.Lock()
		defer mu.Unlock()
		return append([]string(nil), seen...)
	}
}

func newPrecheckEngine(t *testing.T, srv *httptest.Server) *engine.Context {
	t.Helper()
	dir := isolate(t)
	writeEngineConfig(t, dir)
	writeCredentials(t, dir, "test-api-key-000000", "test-secret")

	eng, err := engine.New(engine.Options{
		ConfigDir: dir,
		OfficialOptions: []official.Option{
			official.WithBaseURL(srv.URL),
			official.WithHTTPClient(srv.Client()),
		},
	})
	if err != nil {
		t.Fatalf("engine.New: %v", err)
	}
	return eng
}

// TestPrecheckDerivesStateFromOfficialOrder walks the states the priority table
// distinguishes and asserts the pre-check reports each one, together with the
// advisory cancel/amend booleans.
//
// The asymmetry between the two booleans is the point of the table: cancel stays
// available for everything that is not provably finished, while amend — which can
// raise exposure — is offered only for an order the derivation says is still open.
func TestPrecheckDerivesStateFromOfficialOrder(t *testing.T) {
	cases := []struct {
		name         string
		body         string
		wantState    string
		wantTerminal bool
		wantFailed   bool
		wantCancel   bool
		wantAmend    bool
	}{
		{
			name:       "open and unfilled",
			body:       `{"result":{"orderId":"O-1","status":"OPEN","quantity":"10","execution":{"filledQuantity":"0"}}}`,
			wantState:  "OPEN_UNFILLED",
			wantCancel: true,
			wantAmend:  true,
		},
		{
			name:       "open with a partial fill",
			body:       `{"result":{"orderId":"O-1","status":"OPEN","quantity":"10","execution":{"filledQuantity":"4"}}}`,
			wantState:  "OPEN_PARTIALLY_FILLED",
			wantCancel: true,
			wantAmend:  true,
		},
		{
			name:         "cancelled",
			body:         `{"result":{"orderId":"O-1","status":"CLOSED","quantity":"10","canceledAt":"2026-07-26T01:02:03Z","execution":{"filledQuantity":"0"}}}`,
			wantState:    "CANCELLED",
			wantTerminal: true,
			wantCancel:   false,
			wantAmend:    false,
		},
		{
			name:         "fully filled",
			body:         `{"result":{"orderId":"O-1","status":"CLOSED","quantity":"10","execution":{"filledQuantity":"10"}}}`,
			wantState:    "FILLED",
			wantTerminal: true,
			wantCancel:   false,
			wantAmend:    false,
		},
		{
			// A status this build has never seen must not be guessed at. It is
			// not terminal (we do not know that), so the exit stays open, but it
			// is fail-closed, so no amend is offered.
			name:       "status this build does not understand",
			body:       `{"result":{"orderId":"O-1","status":"SETTLING","quantity":"10","execution":{"filledQuantity":"0"}}}`,
			wantState:  "UNKNOWN_BROKER_STATE",
			wantFailed: true,
			wantCancel: true,
			wantAmend:  false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			srv, paths := precheckServer(t, func(w http.ResponseWriter, _ *http.Request) {
				_, _ = w.Write([]byte(tc.body))
			})
			eng := newPrecheckEngine(t, srv)

			actions, err := eng.BrokerForTest().GetOrderAvailableActions(context.Background(), "O-1")
			if err != nil {
				t.Fatalf("the pre-check must never return an error: %v", err)
			}
			if got := actions[engine.ActionKeyChecked]; got != true {
				t.Errorf("%s = %v, want true", engine.ActionKeyChecked, got)
			}
			if got := actions[engine.ActionKeyState]; got != tc.wantState {
				t.Errorf("%s = %v, want %q", engine.ActionKeyState, got, tc.wantState)
			}
			if got := actions[engine.ActionKeyTerminal]; got != tc.wantTerminal {
				t.Errorf("%s = %v, want %v", engine.ActionKeyTerminal, got, tc.wantTerminal)
			}
			if got := actions[engine.ActionKeyFailClosed]; got != tc.wantFailed {
				t.Errorf("%s = %v, want %v", engine.ActionKeyFailClosed, got, tc.wantFailed)
			}
			if got := actions[engine.ActionKeyCancel]; got != tc.wantCancel {
				t.Errorf("%s = %v, want %v", engine.ActionKeyCancel, got, tc.wantCancel)
			}
			if got := actions[engine.ActionKeyAmend]; got != tc.wantAmend {
				t.Errorf("%s = %v, want %v", engine.ActionKeyAmend, got, tc.wantAmend)
			}

			// It has to be the official single-order endpoint, once.
			if got := paths(); len(got) != 1 || got[0] != "GET /api/v1/orders/O-1" {
				t.Errorf("pre-check reads = %v, want exactly [GET /api/v1/orders/O-1]", got)
			}
		})
	}
}

// TestPrecheckNeverBlocksTheExit is the §0.3 test.
//
// Each arm is a way the read can fail. In every one of them the pre-check must
// return no error — because trading.Service turns an error here into an aborted
// cancel, and an engine that cannot cancel during a broker outage is the failure
// this rule exists to prevent.
func TestPrecheckNeverBlocksTheExit(t *testing.T) {
	cases := []struct {
		name    string
		handler func(w http.ResponseWriter, r *http.Request)
	}{
		{
			name: "server error",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte(`{"error":{"message":"boom"}}`))
			},
		},
		{
			name: "order not found",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusNotFound)
				_, _ = w.Write([]byte(`{"error":{"message":"no such order"}}`))
			},
		},
		{
			name: "rate limited",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Retry-After", "120")
				w.WriteHeader(http.StatusTooManyRequests)
				_, _ = w.Write([]byte(`{"error":{"message":"slow down"}}`))
			},
		},
		{
			name: "credential rejected",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusForbidden)
				_, _ = w.Write([]byte(`{"error":{"message":"forbidden"}}`))
			},
		},
		{
			name: "unparseable body",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				_, _ = w.Write([]byte(`not json at all`))
			},
		},
		{
			name: "broker hangs past the pre-check timeout",
			handler: func(w http.ResponseWriter, r *http.Request) {
				select {
				case <-r.Context().Done():
				case <-time.After(30 * time.Second):
				}
				_, _ = w.Write([]byte(`{}`))
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			srv, _ := precheckServer(t, tc.handler)
			eng := newPrecheckEngine(t, srv)

			actions, err := eng.BrokerForTest().GetOrderAvailableActions(context.Background(), "O-9")
			if err != nil {
				t.Fatalf("a failed pre-check must not become a failed cancel: %v", err)
			}
			if got := actions[engine.ActionKeyChecked]; got != false {
				t.Errorf("%s = %v, want false — an unread order is not an observation",
					engine.ActionKeyChecked, got)
			}
			if got := actions[engine.ActionKeyCancel]; got != true {
				t.Errorf("%s = %v, want true — the exit stays open when we cannot tell",
					engine.ActionKeyCancel, got)
			}
			if got := actions[engine.ActionKeyAmend]; got != false {
				t.Errorf("%s = %v, want false — new exposure needs a positive observation",
					engine.ActionKeyAmend, got)
			}
		})
	}
}

// TestPrecheckTimeoutDoesNotInheritCallerDeadline pins that the bound is the
// pre-check's own: a caller with a generous context still gets a prompt answer,
// so a hanging broker delays an exit by precheckTimeout and not by minutes.
func TestPrecheckTimeoutDoesNotInheritCallerDeadline(t *testing.T) {
	release := make(chan struct{})
	t.Cleanup(func() { close(release) })

	srv, _ := precheckServer(t, func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-r.Context().Done():
		case <-release:
		}
		_, _ = w.Write([]byte(`{}`))
	})
	eng := newPrecheckEngine(t, srv)

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	start := time.Now()
	if _, err := eng.BrokerForTest().GetOrderAvailableActions(ctx, "O-1"); err != nil {
		t.Fatalf("pre-check returned an error: %v", err)
	}
	if elapsed := time.Since(start); elapsed > 20*time.Second {
		t.Errorf("pre-check took %s; it must bound itself rather than inherit the caller's deadline", elapsed)
	}
}

// TestEngineCancelAndAmendWithExpiredWTSSession is the engine-safety scenario
// "WTS 세션 만료 중 취소", end to end.
//
// An expired web session sits in the config directory the engine was pointed at,
// and every non-official host is refused at the transport. The cancel and the
// amend both have to complete anyway, through the official path alone.
func TestEngineCancelAndAmendWithExpiredWTSSession(t *testing.T) {
	dir := isolate(t)
	writeEngineConfig(t, dir)
	writeCredentials(t, dir, "test-api-key-000000", "test-secret")
	writeExpiredSession(t, dir)

	srv, officialCalls := officialStub(t)
	spy := &spyTransport{officialHost: strings.TrimPrefix(srv.URL, "http://"), inner: srv.Client().Transport}

	eng, err := engine.New(engine.Options{
		ConfigDir: dir,
		OfficialOptions: []official.Option{
			official.WithBaseURL(srv.URL),
			official.WithHTTPClient(&http.Client{Transport: spy}),
		},
	})
	if err != nil {
		t.Fatalf("engine.New: %v", err)
	}

	ctx := context.Background()
	svc := eng.TradingServiceForTest()

	cancel := orderintent.CancelIntent{OrderID: "O-1", Symbol: "005930"}
	if _, err := svc.Cancel(ctx, cancel, executeOpts(svc.PreviewCancel(cancel).ConfirmToken)); err != nil {
		t.Fatalf("cancel with an expired web session: %v", err)
	}

	newPrice := 70500.0
	amend := orderintent.AmendIntent{OrderID: "O-1", Price: &newPrice}
	if _, err := svc.Amend(ctx, amend, executeOpts(svc.PreviewAmend(amend).ConfirmToken)); err != nil {
		t.Fatalf("amend with an expired web session: %v", err)
	}

	if foreign := spy.calls(); len(foreign) != 0 {
		t.Errorf("the engine reached %d non-official endpoint(s): %v", len(foreign), foreign)
	}
	calls := officialCalls()
	// The pre-check now costs one derived read per mutation, on the official
	// single-order endpoint — never on WTS.
	if calls["GET /api/v1/orders/O-1"] != 2 {
		t.Errorf("want 2 pre-check reads (one per mutation), got %d; calls=%v",
			calls["GET /api/v1/orders/O-1"], calls)
	}
	for _, want := range []string{"POST /api/v1/orders/O-1/cancel", "POST /api/v1/orders/O-1/modify"} {
		if calls[want] == 0 {
			t.Errorf("official endpoint %q was never called; got %v", want, calls)
		}
	}
}

// writeExpiredSession drops a long-expired web session into the config dir. The
// engine must not read it; the point of the fixture is that its presence and its
// expiry are both irrelevant to the official path.
func writeExpiredSession(t *testing.T, dir string) {
	t.Helper()
	expired := time.Now().Add(-72 * time.Hour).UTC()
	body, err := json.Marshal(map[string]any{
		"provider":   "test",
		"cookies":    map[string]string{"session": "stale"},
		"expires_at": expired.Format(time.RFC3339),
	})
	if err != nil {
		t.Fatalf("marshal session: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "session.json"), body, 0o600); err != nil {
		t.Fatalf("write session: %v", err)
	}
}
