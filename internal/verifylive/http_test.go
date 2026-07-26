package verifylive

// http_test.go runs the procedure through the real Open API client against an
// httptest server.
//
// The rest of the tests drive a fake Broker, which is right for asking "what does
// the tool do when the broker misbehaves". This one asks a different question:
// does the wiring actually work over HTTP — the right verbs, the right paths, the
// account header, the envelope, the idempotency key on the wire — and it is the
// only place that can answer it without a live account.
//
// It is also the runtime proof of the mutation discipline: the server records
// every request, so the test can assert that no POST reaches an order path unless
// a confirmation was answered, and that every order created was followed by a
// cancel.

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/official"
)

// apiServer is a stand-in for the Open API that records every request.
type apiServer struct {
	*httptest.Server

	mu       sync.Mutex
	requests []string
	orders   map[string]bool // id -> still working
	conds    map[string]bool
	keys     map[string]string
	seq      int
}

func newAPIServer(t *testing.T) *apiServer {
	t.Helper()
	s := &apiServer{
		orders: map[string]bool{},
		conds:  map[string]bool{},
		keys:   map[string]string{},
	}
	s.Server = httptest.NewServer(http.HandlerFunc(s.serve))
	t.Cleanup(s.Server.Close)
	return s
}

func (s *apiServer) serve(w http.ResponseWriter, r *http.Request) {
	body, _ := readAll(r)
	s.mu.Lock()
	s.requests = append(s.requests, r.Method+" "+r.URL.Path)
	s.mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	path := r.URL.Path

	switch {
	case path == "/oauth2/token":
		fmt.Fprint(w, `{"access_token":"verify-token","expires_in":3600,"token_type":"Bearer"}`)

	case path == "/api/v1/accounts":
		fmt.Fprint(w, `{"result":[{"accountNo":"123-45-678901","accountSeq":7,"accountType":"BROKERAGE"}]}`)

	case path == "/api/v1/holdings":
		fmt.Fprint(w, `{"result":{"items":[{"symbol":"005930","quantity":"2","lastPrice":"70000"}]}}`)

	case path == "/api/v1/sellable-quantity":
		fmt.Fprint(w, `{"result":{"sellableQuantity":"2"}}`)

	case path == "/api/v1/prices":
		fmt.Fprint(w, `{"result":[{"symbol":"005930","lastPrice":"70000","currency":"KRW"}]}`)

	case path == "/api/v1/price-limits":
		fmt.Fprint(w, `{"result":{"upperLimitPrice":"91000","lowerLimitPrice":"49000","currency":"KRW"}}`)

	case path == "/api/v1/orders" && r.Method == http.MethodGet:
		s.writeOrderList(w, r.URL.Query().Get("status"))

	case path == "/api/v1/orders" && r.Method == http.MethodPost:
		s.createOrder(w, body)

	case strings.HasSuffix(path, "/cancel"):
		id := strings.TrimSuffix(strings.TrimPrefix(path, "/api/v1/orders/"), "/cancel")
		s.mu.Lock()
		s.orders[id] = false
		s.mu.Unlock()
		fmt.Fprintf(w, `{"result":{"orderId":%q}}`, id)

	case strings.HasSuffix(path, "/modify") && strings.HasPrefix(path, "/api/v1/orders/"):
		old := strings.TrimSuffix(strings.TrimPrefix(path, "/api/v1/orders/"), "/modify")
		s.mu.Lock()
		s.orders[old] = false
		s.seq++
		id := fmt.Sprintf("ord-%d", s.seq)
		s.orders[id] = true
		s.mu.Unlock()
		fmt.Fprintf(w, `{"result":{"orderId":%q}}`, id)

	case strings.HasPrefix(path, "/api/v1/orders/") && r.Method == http.MethodGet:
		id := strings.TrimPrefix(path, "/api/v1/orders/")
		s.mu.Lock()
		working := s.orders[id]
		s.mu.Unlock()
		status, canceled := "CANCELED", `"2026-07-26T10:00:00+09:00"`
		if working {
			status, canceled = "PENDING", "null"
		}
		fmt.Fprintf(w, `{"result":{"orderId":%q,"symbol":"005930","side":"BUY","orderType":"LIMIT",`+
			`"timeInForce":"DAY","status":%q,"quantity":"1","price":"56000","currency":"KRW",`+
			`"orderedAt":"2026-07-26T09:00:00+09:00","canceledAt":%s,`+
			`"execution":{"filledQuantity":"0","averageFilledPrice":null,"filledAmount":null,`+
			`"commission":null,"tax":null,"filledAt":null,"settlementDate":null}}}`, id, status, canceled)

	case path == "/api/v1/conditional-orders" && r.Method == http.MethodPost:
		s.createConditional(w, body)

	case path == "/api/v1/conditional-orders" && r.Method == http.MethodGet:
		s.writeConditionalList(w)

	case strings.HasSuffix(path, "/modify") && strings.HasPrefix(path, "/api/v1/conditional-orders/"):
		old := strings.TrimSuffix(strings.TrimPrefix(path, "/api/v1/conditional-orders/"), "/modify")
		s.mu.Lock()
		delete(s.conds, old)
		s.seq++
		id := fmt.Sprintf("co-%d", s.seq)
		s.conds[id] = true
		s.mu.Unlock()
		fmt.Fprintf(w, `{"result":{"conditionalOrderId":%q}}`, id)

	case strings.HasPrefix(path, "/api/v1/conditional-orders/") && r.Method == http.MethodDelete:
		id := strings.TrimPrefix(path, "/api/v1/conditional-orders/")
		s.mu.Lock()
		delete(s.conds, id)
		s.mu.Unlock()
		fmt.Fprint(w, `{"result":null}`)

	case strings.HasPrefix(path, "/api/v1/conditional-orders/") && r.Method == http.MethodGet:
		id := strings.TrimPrefix(path, "/api/v1/conditional-orders/")
		s.mu.Lock()
		live := s.conds[id]
		s.mu.Unlock()
		if !live {
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprint(w, `{"error":{"code":"conditional-order-not-found"}}`)
			return
		}
		fmt.Fprintf(w, `{"result":{"conditionalOrderId":%q,"type":"SINGLE","status":"WATCHING",`+
			`"symbol":"005930","market":"KR","quantity":"1","orderType":"MARKET","expireDate":"2026-08-02",`+
			`"first":{"type":"STOP","status":"WATCHING","triggerPrice":"56000","orderPrice":null,`+
			`"targetProfitRate":null,"triggeredOrderId":null},"second":null,`+
			`"createdAt":"2026-07-26T09:00:00+09:00"}}`, id)

	default:
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `{"error":{"code":"not-found"}}`)
	}
}

func (s *apiServer) createOrder(w http.ResponseWriter, body []byte) {
	var req struct {
		Symbol        string `json:"symbol"`
		Side          string `json:"side"`
		Quantity      string `json:"quantity"`
		Price         string `json:"price"`
		ClientOrderID string `json:"clientOrderId"`
	}
	_ = json.Unmarshal(body, &req)
	canonical := req.Side + "|" + req.Symbol + "|" + req.Quantity + "|" + req.Price

	s.mu.Lock()
	defer s.mu.Unlock()
	if req.ClientOrderID != "" {
		if prior, seen := s.keys[req.ClientOrderID]; seen {
			id, priorBody, _ := strings.Cut(prior, "\x00")
			if priorBody == canonical {
				fmt.Fprintf(w, `{"result":{"orderId":%q,"clientOrderId":%q}}`, id, req.ClientOrderID)
				return
			}
			w.WriteHeader(http.StatusConflict)
			fmt.Fprint(w, `{"error":{"code":"idempotency-key-conflict","message":"동일 키로 다른 본문"}}`)
			return
		}
	}
	s.seq++
	id := fmt.Sprintf("ord-%d", s.seq)
	s.orders[id] = true
	if req.ClientOrderID != "" {
		s.keys[req.ClientOrderID] = id + "\x00" + canonical
	}
	if req.ClientOrderID == "" {
		fmt.Fprintf(w, `{"result":{"orderId":%q,"clientOrderId":null}}`, id)
		return
	}
	fmt.Fprintf(w, `{"result":{"orderId":%q,"clientOrderId":%q}}`, id, req.ClientOrderID)
}

func (s *apiServer) createConditional(w http.ResponseWriter, body []byte) {
	var req struct {
		Symbol        string `json:"symbol"`
		Quantity      string `json:"quantity"`
		ClientOrderID string `json:"clientOrderId"`
		First         struct {
			TriggerPrice string `json:"triggerPrice"`
		} `json:"first"`
	}
	_ = json.Unmarshal(body, &req)
	canonical := "cond|" + req.Symbol + "|" + req.Quantity + "|" + req.First.TriggerPrice

	s.mu.Lock()
	defer s.mu.Unlock()
	if prior, seen := s.keys[req.ClientOrderID]; seen && req.ClientOrderID != "" {
		id, priorBody, _ := strings.Cut(prior, "\x00")
		if priorBody == canonical {
			fmt.Fprintf(w, `{"result":{"conditionalOrderId":%q,"clientOrderId":%q}}`, id, req.ClientOrderID)
			return
		}
	}
	s.seq++
	id := fmt.Sprintf("co-%d", s.seq)
	s.conds[id] = true
	if req.ClientOrderID != "" {
		s.keys[req.ClientOrderID] = id + "\x00" + canonical
	}
	fmt.Fprintf(w, `{"result":{"conditionalOrderId":%q,"clientOrderId":%q}}`, id, req.ClientOrderID)
}

func (s *apiServer) writeOrderList(w http.ResponseWriter, status string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	var parts []string
	for id, working := range s.orders {
		if status == "OPEN" && !working {
			continue
		}
		state := "CANCELED"
		if working {
			state = "PENDING"
		}
		parts = append(parts, fmt.Sprintf(`{"orderId":%q,"symbol":"005930","status":%q,"quantity":"1",`+
			`"price":"56000","execution":{"filledQuantity":"0"}}`, id, state))
	}
	fmt.Fprintf(w, `{"result":{"orders":[%s],"nextCursor":"","hasNext":false}}`, strings.Join(parts, ","))
}

func (s *apiServer) writeConditionalList(w http.ResponseWriter) {
	s.mu.Lock()
	defer s.mu.Unlock()
	var parts []string
	for id := range s.conds {
		parts = append(parts, fmt.Sprintf(`{"conditionalOrderId":%q,"type":"SINGLE","status":"WATCHING",`+
			`"symbol":"005930","market":"KR","quantity":"1","orderType":"MARKET","expireDate":"2026-08-02",`+
			`"first":{"type":"STOP","status":"WATCHING","triggerPrice":"56000"},"second":null,`+
			`"createdAt":"2026-07-26T09:00:00+09:00"}`, id))
	}
	fmt.Fprintf(w, `{"result":{"conditionalOrders":[%s],"nextCursor":"","hasNext":false}}`,
		strings.Join(parts, ","))
}

func (s *apiServer) seen() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]string(nil), s.requests...)
}

func (s *apiServer) client(t *testing.T) *official.Client {
	t.Helper()
	return official.New(
		official.Credentials{APIKey: "k", SecretKey: "s"},
		filepath.Join(t.TempDir(), "token.json"),
		official.WithBaseURL(s.URL),
		official.WithHTTPClient(s.Client()),
		official.WithAccountSeq(7),
	)
}

func readAll(r *http.Request) ([]byte, error) {
	if r.Body == nil {
		return nil, nil
	}
	buf := make([]byte, 0, 512)
	tmp := make([]byte, 512)
	for {
		n, err := r.Body.Read(tmp)
		buf = append(buf, tmp[:n]...)
		if err != nil {
			return buf, nil
		}
	}
}

// --- the tests -----------------------------------------------------------------

// TestProcedureOverHTTPUsesTheDocumentedEndpoints.
func TestProcedureOverHTTPUsesTheDocumentedEndpoints(t *testing.T) {
	srv := newAPIServer(t)
	client := srv.client(t)
	op := alwaysConfirm()

	record := filepath.Join(t.TempDir(), FileName)
	first := runOverHTTP(t, client, op, record, 1)
	if !first.Halted {
		t.Fatalf("the first run did not stop for the persistence check: %+v", first.Outcomes)
	}
	second := runOverHTTP(t, client, op, record, 2)
	if len(second.Outstanding) != 0 {
		t.Fatalf("the account is left holding %+v", second.Outstanding)
	}

	requests := srv.seen()
	for _, want := range []string{
		"POST /api/v1/orders",
		"GET /api/v1/orders",
		"GET /api/v1/prices",
		"GET /api/v1/price-limits",
		"GET /api/v1/sellable-quantity",
		"POST /api/v1/conditional-orders",
		"GET /api/v1/conditional-orders",
	} {
		if !containsRequest(requests, want) {
			t.Errorf("the procedure never issued %s over HTTP", want)
		}
	}
	if !containsPrefixSuffix(requests, "POST /api/v1/orders/", "/cancel") {
		t.Error("no order cancel reached the wire")
	}
	if !containsPrefixSuffix(requests, "POST /api/v1/orders/", "/modify") {
		t.Error("no order amend reached the wire")
	}
	if !containsPrefixSuffix(requests, "POST /api/v1/conditional-orders/", "/modify") {
		t.Error("no conditional modify reached the wire")
	}
	if !containsPrefix(requests, "DELETE /api/v1/conditional-orders/") {
		t.Error("no conditional cancel reached the wire")
	}

	// Nothing must be left working on the server either.
	srv.mu.Lock()
	defer srv.mu.Unlock()
	for id, working := range srv.orders {
		if working {
			t.Errorf("order %s is still working on the server", id)
		}
	}
	if len(srv.conds) != 0 {
		t.Errorf("conditional orders are still registered: %v", srv.conds)
	}
}

// TestRefusingEveryConfirmationIssuesNoMutatingRequest is the runtime half of the
// safety claim, watched at the HTTP layer: with every confirmation declined, the
// server must see nothing but GETs and the OAuth exchange.
func TestRefusingEveryConfirmationIssuesNoMutatingRequest(t *testing.T) {
	srv := newAPIServer(t)
	client := srv.client(t)
	op := refuseWhere(func(Mutation) bool { return true })

	record := filepath.Join(t.TempDir(), FileName)
	runOverHTTP(t, client, op, record, 1)

	for _, req := range srv.seen() {
		method, path, _ := strings.Cut(req, " ")
		if method == http.MethodPost && path == "/oauth2/token" {
			continue
		}
		if method != http.MethodGet {
			t.Errorf("a refused verification issued %s — nothing may reach the account without a typed "+
				"confirmation", req)
		}
	}
	if len(op.seen) == 0 {
		t.Error("no confirmation was ever asked for")
	}
}

// TestIdempotencyKeyIsSentOnTheWire. The key is the whole subject of task 2.7; a
// tool that measured its behaviour without actually sending it would measure
// nothing.
func TestIdempotencyKeyIsSentOnTheWire(t *testing.T) {
	srv := newAPIServer(t)
	client := srv.client(t)
	record := filepath.Join(t.TempDir(), FileName)
	runOverHTTP(t, client, alwaysConfirm(), record, 1)

	srv.mu.Lock()
	keys := len(srv.keys)
	srv.mu.Unlock()
	if keys == 0 {
		t.Fatal("no request carried a clientOrderId")
	}

	entries, err := LoadEntries(record)
	if err != nil {
		t.Fatalf("LoadEntries: %v", err)
	}
	rep := BuildReport(record, entries, timeNowUTC())
	if !rep.ReplayEnabled {
		t.Errorf("a server that honours the key still left replay disabled; unverified: %v", rep.Unverified)
	}
	conflict := findAttribute(t, rep, "idempotency.conflict_error_code")
	if conflict.Value != "idempotency-key-conflict" {
		t.Errorf("conflict_error_code = %q, want the broker's own code decoded out of the error envelope",
			conflict.Value)
	}
}

func runOverHTTP(t *testing.T, client *official.Client, op *operator, record string, instance int) Summary {
	t.Helper()
	prior, err := LoadEntries(record)
	if err != nil {
		t.Fatalf("LoadEntries: %v", err)
	}
	rec, err := OpenRecorder(record)
	if err != nil {
		t.Fatalf("OpenRecorder: %v", err)
	}
	defer rec.Close()

	runner, err := New(Options{
		Broker:        client,
		Recorder:      rec,
		Confirm:       op.confirmer(),
		AccountRef:    "123-45-678901",
		Symbol:        "005930",
		HoldingSymbol: "005930",
		Prior:         prior,
		Process:       Process{PID: 4000 + instance, InstanceID: fmt.Sprintf("http-%d", instance)},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	summary, err := runner.Run(context.Background())
	if err != nil {
		t.Logf("run %d reported: %v", instance, err)
	}
	return summary
}

func containsRequest(requests []string, want string) bool {
	for _, r := range requests {
		if r == want {
			return true
		}
	}
	return false
}

func containsPrefix(requests []string, prefix string) bool {
	for _, r := range requests {
		if strings.HasPrefix(r, prefix) {
			return true
		}
	}
	return false
}

func containsPrefixSuffix(requests []string, prefix, suffix string) bool {
	for _, r := range requests {
		if strings.HasPrefix(r, prefix) && strings.HasSuffix(r, suffix) {
			return true
		}
	}
	return false
}

// timeNowUTC keeps the report's clock out of the test bodies.
func timeNowUTC() time.Time { return time.Now().UTC() }
