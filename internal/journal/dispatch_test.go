package journal

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// TestClassifyHTTPMutation is the transport-fault table. A mutation has no
// idempotency key at this broker, so the only safe classification of "we do not
// know" is AMBIGUOUS — never "probably fine" and never "probably not sent".
func TestClassifyHTTPMutation(t *testing.T) {
	transportErr := errors.New("connection reset")

	cases := []struct {
		name   string
		send   SendState
		status int
		err    error
		want   DispatchClass
	}{
		// Failure before the request left the process: provably nothing happened.
		{"dial error", SendNotStarted, 0, transportErr, DispatchNotSent},
		// Failure while writing: bytes may have reached the broker.
		{"error mid-write", SendPartial, 0, transportErr, DispatchAmbiguous},
		// Failure after the request was fully written: the broker may have acted.
		{"error after write", SendComplete, 0, transportErr, DispatchAmbiguous},
		{"read timeout after write", SendComplete, 0, context.DeadlineExceeded, DispatchAmbiguous},

		{"200", SendComplete, 200, nil, DispatchAcked},
		{"201", SendComplete, 201, nil, DispatchAcked},
		{"204", SendComplete, 204, nil, DispatchAcked},

		// Well-formed refusals: the broker decided, and it decided no.
		{"400 bad request", SendComplete, 400, nil, DispatchRejected},
		{"401 unauthorised", SendComplete, 401, nil, DispatchRejected},
		{"403 forbidden", SendComplete, 403, nil, DispatchRejected},
		{"404 not found", SendComplete, 404, nil, DispatchRejected},
		{"405 method not allowed", SendComplete, 405, nil, DispatchRejected},
		{"415 unsupported media", SendComplete, 415, nil, DispatchRejected},
		{"422 unprocessable", SendComplete, 422, nil, DispatchRejected},

		// Everything else non-2xx is unknown, including the tempting ones: a 429
		// or a 500 can arrive after the order reached the matching engine.
		{"408 timeout", SendComplete, 408, nil, DispatchAmbiguous},
		{"409 conflict", SendComplete, 409, nil, DispatchAmbiguous},
		{"429 rate limited", SendComplete, 429, nil, DispatchAmbiguous},
		{"500", SendComplete, 500, nil, DispatchAmbiguous},
		{"502", SendComplete, 502, nil, DispatchAmbiguous},
		{"503", SendComplete, 503, nil, DispatchAmbiguous},
		{"504", SendComplete, 504, nil, DispatchAmbiguous},
		{"302 redirect", SendComplete, 302, nil, DispatchAmbiguous},
		{"no status and no error", SendComplete, 0, nil, DispatchAmbiguous},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := ClassifyHTTPMutation(tc.send, tc.status, tc.err)
			if got.Class != tc.want {
				t.Fatalf("Class = %s, want %s (reason %s)", got.Class, tc.want, got.ReasonCode)
			}
			if got.Class != DispatchAcked && got.ReasonCode == "" {
				t.Error("a non-acked classification needs a reason code")
			}
			if tc.err != nil && !errors.Is(got.Err, tc.err) {
				t.Errorf("Err = %v, want %v", got.Err, tc.err)
			}
		})
	}
}

// --- transport fault injection ---------------------------------------------

// faultKind selects what the fake broker does with a mutation request.
type faultKind int

const (
	faultNone faultKind = iota
	faultRefuseConnection
	faultCloseAfterRequest
	faultHangThenClose
)

// fakeBroker is an httptest stand-in for the official order endpoint. No test in
// this package ever talks to a real Toss host.
type fakeBroker struct {
	url      string
	requests *atomic.Int64
}

func newFakeBroker(t *testing.T, fault faultKind, status int, orderID string) fakeBroker {
	t.Helper()
	var requests atomic.Int64

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		_, _ = io.ReadAll(r.Body)

		switch fault {
		case faultCloseAfterRequest, faultHangThenClose:
			// The request was fully received; the connection dies before any
			// response. This is the worst case for the caller: the broker may
			// well have executed the mutation.
			conn, _, err := http.NewResponseController(w).Hijack()
			if err != nil {
				t.Errorf("hijack: %v", err)
				return
			}
			if fault == faultHangThenClose {
				time.Sleep(20 * time.Millisecond)
			}
			conn.Close()
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		if orderID != "" {
			_, _ = w.Write([]byte(`{"result":{"orderId":"` + orderID + `"}}`))
		} else {
			_, _ = w.Write([]byte(`{"result":{}}`))
		}
	})

	srv := httptest.NewServer(handler)
	url := srv.URL + "/api/v1/orders"

	if fault == faultRefuseConnection {
		// Close the server first: the client then fails to connect, which is the
		// "failure before the request" class.
		srv.Close()
		return fakeBroker{url: url, requests: &requests}
	}
	t.Cleanup(srv.Close)
	return fakeBroker{url: url, requests: &requests}
}

// httpDispatch is the dispatch function under test: it posts the mutation and
// classifies the transport result. This is the shape the execution gateway will
// use.
func httpDispatch(broker fakeBroker) DispatchFunc {
	return func(ctx context.Context, a *Attempt) DispatchOutcome {
		ctx, tracker := WithSendTracker(ctx)
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, broker.url,
			strings.NewReader(`{"symbol":"AAPL","quantity":"10"}`))
		if err != nil {
			return DispatchOutcome{Class: DispatchNotSent, ReasonCode: "request_build_failed", Err: err}
		}
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{Timeout: 2 * time.Second}
		resp, err := client.Do(req)
		status := 0
		var body []byte
		if resp != nil {
			status = resp.StatusCode
			body, _ = io.ReadAll(resp.Body)
			resp.Body.Close()
		}

		out := ClassifyHTTPMutation(tracker.State(), status, err)
		if out.Class == DispatchAcked {
			var envelope struct {
				Result struct {
					OrderID string `json:"orderId"`
				} `json:"result"`
			}
			if err := json.Unmarshal(body, &envelope); err == nil {
				out.BrokerOrderID = envelope.Result.OrderID
			}
		}
		return out
	}
}

// TestRunTransportFaults is the end-to-end lifecycle under injected transport
// faults: before, during and after the request. Each row asserts the terminal (or
// blocking) state the journal ends in, and that the mutation was sent at most once
// — a mutation is never retried automatically, whatever the failure.
func TestRunTransportFaults(t *testing.T) {
	cases := []struct {
		name         string
		fault        faultKind
		status       int
		orderID      string
		wantClass    DispatchClass
		wantState    AttemptState
		wantReason   string
		wantRequests int64
		wantBlocking bool
	}{
		{
			name: "failure before the request leaves", fault: faultRefuseConnection,
			wantClass: DispatchNotSent, wantState: StateNotDispatched,
			wantReason: ReasonDispatchNotSent, wantRequests: 0,
		},
		{
			name: "connection dies after the request was received", fault: faultCloseAfterRequest,
			wantClass: DispatchAmbiguous, wantState: StateInDoubt,
			wantReason: ReasonDispatchAmbiguous, wantRequests: 1, wantBlocking: true,
		},
		{
			name: "response never arrives", fault: faultHangThenClose,
			wantClass: DispatchAmbiguous, wantState: StateInDoubt,
			wantReason: ReasonDispatchAmbiguous, wantRequests: 1, wantBlocking: true,
		},
		{
			name: "broker accepts and names the order", status: 200, orderID: "order-42",
			wantClass: DispatchAcked, wantState: StateConfirmed,
			wantReason: ReasonBrokerAcknowledged, wantRequests: 1,
		},
		{
			name: "broker accepts without naming an order", status: 200,
			wantClass: DispatchAcked, wantState: StateInDoubt,
			wantReason: ReasonAckWithoutOrderID, wantRequests: 1, wantBlocking: true,
		},
		{
			name: "broker rejects the request", status: 400,
			wantClass: DispatchRejected, wantState: StateFailedConfirmed,
			wantReason: ReasonDispatchRejected, wantRequests: 1,
		},
		{
			name: "broker returns a server error", status: 500,
			wantClass: DispatchAmbiguous, wantState: StateInDoubt,
			wantReason: ReasonDispatchAmbiguous, wantRequests: 1, wantBlocking: true,
		},
		{
			name: "broker rate limits the mutation", status: 429,
			wantClass: DispatchAmbiguous, wantState: StateInDoubt,
			wantReason: ReasonDispatchAmbiguous, wantRequests: 1, wantBlocking: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			j := openTestJournal(t)
			ctx := context.Background()
			broker := newFakeBroker(t, tc.fault, tc.status, tc.orderID)

			res, err := j.Run(ctx, testRequest(), httpDispatch(broker))
			if err != nil {
				t.Fatalf("Run: %v", err)
			}
			if res.Class != tc.wantClass {
				t.Errorf("Class = %s, want %s", res.Class, tc.wantClass)
			}
			if res.Final != tc.wantState {
				t.Fatalf("Final = %s, want %s (reason %q)", res.Final, tc.wantState, res.ReasonCode)
			}
			if res.ReasonCode != tc.wantReason {
				t.Errorf("ReasonCode = %q, want %q", res.ReasonCode, tc.wantReason)
			}
			if got := broker.requests.Load(); got != tc.wantRequests {
				t.Errorf("broker saw %d requests, want %d (mutations are never retried)", got, tc.wantRequests)
			}

			rec, err := j.LookupAttempt(ctx, "attempt-1")
			if err != nil {
				t.Fatal(err)
			}
			if rec.State != tc.wantState {
				t.Errorf("stored state = %s, want %s", rec.State, tc.wantState)
			}
			if rec.DispatchStartedAt == "" {
				t.Error("dispatch_started_at must be recorded before the request goes out")
			}
			if tc.wantState == StateConfirmed && rec.BrokerOrderID != tc.orderID {
				t.Errorf("broker_order_id = %q, want %q", rec.BrokerOrderID, tc.orderID)
			}

			// Blocking states must show up in the recovery report; settled ones
			// must not.
			report, err := j.RecoverPending(ctx)
			if err != nil {
				t.Fatal(err)
			}
			if got := len(report.Blocked) > 0; got != tc.wantBlocking {
				t.Errorf("blocking = %v, want %v (report %+v)", got, tc.wantBlocking, report)
			}
		})
	}
}

// TestRunSkipsDispatchWhenTheJournalWriteFails is the ordering guarantee at the
// second write too: if DISPATCH_STARTED cannot be recorded, nothing is sent. A
// mutation the journal does not know about is exactly what the crash-recovery
// procedure cannot reason about.
func TestRunSkipsDispatchWhenTheJournalWriteFails(t *testing.T) {
	j := openTestJournal(t)
	ctx := context.Background()

	a, err := j.Prepare(ctx, testRequest())
	if err != nil {
		t.Fatal(err)
	}

	var dispatched atomic.Bool
	// Break the journal after the intent was recorded.
	if err := j.db.Close(); err != nil {
		t.Fatal(err)
	}

	_, err = a.Dispatch(ctx, func(context.Context, *Attempt) DispatchOutcome {
		dispatched.Store(true)
		return DispatchOutcome{Class: DispatchAcked, BrokerOrderID: "order-1"}
	})
	if err == nil {
		t.Fatal("Dispatch must fail when DISPATCH_STARTED cannot be recorded")
	}
	if dispatched.Load() {
		t.Fatal("the broker must not be called when the journal write failed")
	}
}

// TestRunRefusesUnknownDispatchClass keeps an unclassified outcome from becoming a
// silent success.
func TestRunRefusesUnknownDispatchClass(t *testing.T) {
	j := openTestJournal(t)
	ctx := context.Background()

	res, err := j.Run(ctx, testRequest(), func(context.Context, *Attempt) DispatchOutcome {
		return DispatchOutcome{Class: DispatchClass("SOMETHING_NEW")}
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if res.Final != StateInDoubt {
		t.Fatalf("Final = %s, want IN_DOUBT", res.Final)
	}
	if res.ReasonCode != ReasonUnknownDispatchClass {
		t.Errorf("ReasonCode = %q, want %q", res.ReasonCode, ReasonUnknownDispatchClass)
	}
}

// TestRunPropagatesPrepareFailure keeps a rejected request from reaching the broker.
func TestRunPropagatesPrepareFailure(t *testing.T) {
	j := openTestJournal(t)
	ctx := context.Background()

	var dispatched atomic.Bool
	req := testRequest()
	req.Intent.Symbol = "" // invalid

	if _, err := j.Run(ctx, req, func(context.Context, *Attempt) DispatchOutcome {
		dispatched.Store(true)
		return DispatchOutcome{Class: DispatchAcked}
	}); !errors.Is(err, ErrInvalidRequest) {
		t.Fatalf("want ErrInvalidRequest, got %v", err)
	}
	if dispatched.Load() {
		t.Fatal("an invalid request must never reach the broker")
	}
}

// TestRunNilDispatch guards the programming error rather than sending nothing and
// reporting success.
func TestRunNilDispatch(t *testing.T) {
	j := openTestJournal(t)
	if _, err := j.Run(context.Background(), testRequest(), nil); !errors.Is(err, ErrInvalidRequest) {
		t.Fatalf("want ErrInvalidRequest, got %v", err)
	}
}

// TestConcurrentRunsAreSerialised is the -race guard on the journal write path: the
// engine runs its resolution and submission work in goroutines, and the journal is
// a single-writer store.
func TestConcurrentRunsAreSerialised(t *testing.T) {
	j := openTestJournal(t)
	broker := newFakeBroker(t, faultNone, 200, "order-shared")
	ctx := context.Background()

	const workers = 6
	var wg sync.WaitGroup
	errs := make([]error, workers)
	states := make([]AttemptState, workers)

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			req := testRequest()
			req.Intent.ID = "intent-" + strconv.Itoa(i)
			req.AttemptID = "attempt-" + strconv.Itoa(i)
			res, err := j.Run(ctx, req, httpDispatch(broker))
			errs[i] = err
			states[i] = res.Final
		}(i)
	}
	wg.Wait()

	for i := 0; i < workers; i++ {
		if errs[i] != nil {
			t.Errorf("worker %d: %v", i, errs[i])
			continue
		}
		if states[i] != StateConfirmed {
			t.Errorf("worker %d final = %s, want CONFIRMED", i, states[i])
		}
	}
	if got := broker.requests.Load(); got != workers {
		t.Errorf("broker saw %d requests, want %d", got, workers)
	}
}

// TestSendTrackerReportsCompletion exercises the httptrace wiring the "before /
// during / after" classification depends on.
func TestSendTrackerReportsCompletion(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.ReadAll(r.Body)
		w.WriteHeader(200)
	}))
	defer srv.Close()

	ctx, tracker := WithSendTracker(context.Background())
	if tracker.State() != SendNotStarted {
		t.Fatalf("initial state = %v, want SendNotStarted", tracker.State())
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, srv.URL, strings.NewReader("{}"))
	if err != nil {
		t.Fatal(err)
	}
	resp, err := srv.Client().Do(req)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if tracker.State() != SendComplete {
		t.Fatalf("state after a successful round trip = %v, want SendComplete", tracker.State())
	}

	// A connection that never opens must leave the tracker untouched.
	dead, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	addr := dead.Addr().String()
	dead.Close()

	ctx2, tracker2 := WithSendTracker(context.Background())
	req2, err := http.NewRequestWithContext(ctx2, http.MethodPost, "http://"+addr, strings.NewReader("{}"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := (&http.Client{Timeout: time.Second}).Do(req2); err == nil {
		t.Fatal("expected a dial failure")
	}
	if tracker2.State() != SendNotStarted {
		t.Fatalf("state after a dial failure = %v, want SendNotStarted", tracker2.State())
	}
}
