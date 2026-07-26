package verifylive

// fake_broker_test.go is the account the tests run against.
//
// It is an in-process broker rather than an httptest server for one reason: the
// properties being tested are about *behaviour under a broker that misbehaves* —
// an idempotency key that is ignored, a modify that leaves the old id readable, a
// conditional that vanishes when the process restarts — and each of those is one
// field on this struct. The HTTP layer is exercised where it belongs, in
// cmd/tossctl's httptest tests and in internal/official's.
//
// Every request is recorded, so a test can assert on what was *not* sent, which
// is the assertion that matters most here.

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/domain"
	"github.com/JungHoonGhae/tossinvest-cli/internal/official"
	"github.com/JungHoonGhae/tossinvest-cli/internal/orderintent"
	"github.com/JungHoonGhae/tossinvest-cli/internal/testenv"
	trading "github.com/JungHoonGhae/tossinvest-cli/internal/trading"
)

// TestMain installs internal/testenv's guard for the whole package.
//
// This is the mechanical half of WORKFLOW's 불변 규칙 for a package whose
// production code exists to place live orders: "실 endpoint 접근은 httptest 대체
// 없이는 금지된다. 문구가 아니라 테스트 인프라가 막는다". A test in this package
// that reaches a Toss hostname — through this package, through internal/official,
// or through an http.Client somebody built by hand — panics instead of trading.
func TestMain(m *testing.M) {
	previous := http.DefaultTransport
	http.DefaultTransport = &testenv.Guard{Base: previous, OnBlock: func(err *testenv.ErrRealHost) {
		panic("verifylive tests: " + err.Error())
	}}
	code := m.Run()
	http.DefaultTransport = previous
	os.Exit(code)
}

// fakeBroker is a scriptable Open API.
type fakeBroker struct {
	mu sync.Mutex

	// --- what the account looks like ---
	accounts   []domain.Account
	holdings   []domain.Position
	sellable   map[string]float64
	quotes     map[string]float64
	limits     map[string]domain.PriceLimits
	orderPages [][]json.RawMessage
	orders     map[string]json.RawMessage
	conds      map[string]domain.ConditionalOrder

	// --- how it misbehaves ---
	// honourIdempotency returns the first order for a repeated key.
	honourIdempotency bool
	// conflictOnDifferentBody returns idempotency-key-conflict for a reused key
	// with a different body.
	conflictOnDifferentBody bool
	// modifyIssuesNewID makes a conditional modify mint a new identifier.
	modifyIssuesNewID bool
	// modifyLeavesOldReadable keeps the pre-modify id readable, which is the
	// failure the report has to catch.
	modifyLeavesOldReadable bool
	// amendIssuesNewID makes an order amend mint a new identifier.
	amendIssuesNewID bool
	// rejectOversell refuses a sell larger than the holding.
	rejectOversell bool
	// dropConditionalsOnRestart deletes conditional orders, so the persistence
	// step can be tested failing as well as passing.
	dropConditionalsOnRestart bool
	// expireKeysOnSleep models the idempotency window closing while the
	// validity-window step waits.
	expireKeysOnSleep bool
	// placeErr, if set, fails every placement.
	placeErr error

	// --- what happened ---
	requests []string
	keys     map[string]keyedOrder
	seq      int
}

type keyedOrder struct {
	orderID string
	body    string
}

func newFakeBroker() *fakeBroker {
	return &fakeBroker{
		accounts:                []domain.Account{{ID: "1", DisplayName: "123-45-678901"}},
		sellable:                map[string]float64{},
		quotes:                  map[string]float64{"005930": 70000},
		limits:                  map[string]domain.PriceLimits{"005930": {UpperLimit: 91000, LowerLimit: 49000}},
		orders:                  map[string]json.RawMessage{},
		conds:                   map[string]domain.ConditionalOrder{},
		keys:                    map[string]keyedOrder{},
		honourIdempotency:       true,
		conflictOnDifferentBody: true,
		modifyIssuesNewID:       true,
		amendIssuesNewID:        true,
		rejectOversell:          true,
	}
}

// withHolding gives the account something to sell.
func (f *fakeBroker) withHolding(symbol string, qty float64) *fakeBroker {
	f.holdings = append(f.holdings, domain.Position{Symbol: symbol, Quantity: qty})
	f.sellable[symbol] = qty
	if _, ok := f.quotes[symbol]; !ok {
		f.quotes[symbol] = 70000
	}
	if _, ok := f.limits[symbol]; !ok {
		f.limits[symbol] = domain.PriceLimits{UpperLimit: 91000, LowerLimit: 49000}
	}
	return f
}

func (f *fakeBroker) log(what string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.requests = append(f.requests, what)
}

func (f *fakeBroker) seen() []string {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]string(nil), f.requests...)
}

func (f *fakeBroker) countRequests(prefix string) int {
	n := 0
	for _, r := range f.seen() {
		if strings.HasPrefix(r, prefix) {
			n++
		}
	}
	return n
}

func (f *fakeBroker) nextID(prefix string) string {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.seq++
	return fmt.Sprintf("%s-%d", prefix, f.seq)
}

// --- reads --------------------------------------------------------------------

func (f *fakeBroker) Accounts(context.Context) ([]domain.Account, error) {
	f.log("GET /accounts")
	return f.accounts, nil
}

func (f *fakeBroker) Holdings(_ context.Context, _ string) ([]domain.Position, error) {
	f.log("GET /holdings")
	return f.holdings, nil
}

func (f *fakeBroker) SellableQuantity(_ context.Context, symbol string) (domain.SellableQuantity, error) {
	f.log("GET /sellable-quantity " + symbol)
	return domain.SellableQuantity{Symbol: symbol, Quantity: f.sellable[symbol]}, nil
}

func (f *fakeBroker) Prices(_ context.Context, symbols []string) ([]domain.Quote, error) {
	f.log("GET /prices")
	out := make([]domain.Quote, 0, len(symbols))
	for _, s := range symbols {
		out = append(out, domain.Quote{Symbol: s, Last: f.quotes[s]})
	}
	return out, nil
}

func (f *fakeBroker) PriceLimits(_ context.Context, symbol string) (domain.PriceLimits, error) {
	f.log("GET /price-limits " + symbol)
	l := f.limits[symbol]
	l.Symbol = symbol
	return l, nil
}

func (f *fakeBroker) OrdersPageRaw(_ context.Context, filter official.OrdersFilter, cursor string) (official.RawOrderPage, error) {
	f.log("GET /orders status=" + filter.Status)
	if filter.Status == "OPEN" {
		f.mu.Lock()
		defer f.mu.Unlock()
		var open []json.RawMessage
		for _, raw := range f.orders {
			var v rawOrderView
			if err := json.Unmarshal(raw, &v); err == nil && v.Status == "PENDING" {
				open = append(open, raw)
			}
		}
		return official.RawOrderPage{Orders: open}, nil
	}
	if len(f.orderPages) == 0 {
		return official.RawOrderPage{}, nil
	}
	idx := 0
	if cursor != "" {
		fmt.Sscanf(cursor, "p%d", &idx)
	}
	if idx >= len(f.orderPages) {
		return official.RawOrderPage{}, nil
	}
	page := official.RawOrderPage{Orders: f.orderPages[idx]}
	if idx+1 < len(f.orderPages) {
		page.HasNext = true
		page.NextCursor = fmt.Sprintf("p%d", idx+1)
	}
	return page, nil
}

func (f *fakeBroker) OrderRawByID(_ context.Context, orderID string) (json.RawMessage, error) {
	f.log("GET /orders/" + orderID)
	f.mu.Lock()
	defer f.mu.Unlock()
	raw, ok := f.orders[orderID]
	if !ok {
		return nil, &official.APIError{Code: 404, Body: `{"error":{"code":"order-not-found"}}`}
	}
	return raw, nil
}

func (f *fakeBroker) ConditionalOrders(_ context.Context, status, symbol, _ string, _ int) (domain.ConditionalOrderList, error) {
	f.log("GET /conditional-orders")
	f.mu.Lock()
	defer f.mu.Unlock()
	var out []domain.ConditionalOrder
	for _, c := range f.conds {
		if status != "" && c.Status != status {
			continue
		}
		if symbol != "" && c.Symbol != symbol {
			continue
		}
		out = append(out, c)
	}
	return domain.ConditionalOrderList{Orders: out}, nil
}

func (f *fakeBroker) ConditionalOrder(_ context.Context, id string) (domain.ConditionalOrder, error) {
	f.log("GET /conditional-orders/" + id)
	f.mu.Lock()
	defer f.mu.Unlock()
	c, ok := f.conds[id]
	if !ok {
		return domain.ConditionalOrder{}, &official.APIError{Code: 404, Body: `{"error":{"code":"conditional-order-not-found"}}`}
	}
	return c, nil
}

// --- writes -------------------------------------------------------------------

func (f *fakeBroker) PlaceOrder(_ context.Context, intent orderintent.PlaceIntent) (trading.MutationResult, error) {
	f.log(fmt.Sprintf("POST /orders %s %s x%v @%v key=%s",
		intent.Side, intent.Symbol, intent.Quantity, intent.Price, intent.ClientOrderID))
	if f.placeErr != nil {
		return trading.MutationResult{}, f.placeErr
	}
	if strings.EqualFold(intent.Side, "sell") && f.rejectOversell && intent.Quantity > f.sellable[intent.Symbol] {
		return trading.MutationResult{}, &official.APIError{
			Code: 422, Body: `{"error":{"code":"insufficient-sellable-quantity","message":"보유수량을 초과합니다"}}`}
	}

	body := fmt.Sprintf("%s|%s|%v|%v", intent.Side, intent.Symbol, intent.Quantity, intent.Price)
	if intent.ClientOrderID != "" {
		f.mu.Lock()
		prior, seen := f.keys[intent.ClientOrderID]
		f.mu.Unlock()
		if seen {
			switch {
			case prior.body == body && f.honourIdempotency:
				return trading.MutationResult{Kind: "place", OrderID: prior.orderID}, nil
			case prior.body != body && f.conflictOnDifferentBody:
				return trading.MutationResult{}, &official.APIError{
					Code: 409,
					Body: `{"error":{"code":"idempotency-key-conflict","message":"동일 키로 다른 본문"}}`,
				}
			}
		}
	}

	id := f.nextID("ord")
	f.mu.Lock()
	f.orders[id] = mustOrderJSON(id, intent.Symbol, strings.ToUpper(intent.Side), "PENDING", intent.Quantity, intent.Price, "")
	if intent.ClientOrderID != "" {
		f.keys[intent.ClientOrderID] = keyedOrder{orderID: id, body: body}
	}
	f.mu.Unlock()
	return trading.MutationResult{Kind: "place", OrderID: id}, nil
}

func (f *fakeBroker) CancelOrder(_ context.Context, orderID string) (trading.MutationResult, error) {
	f.log("POST /orders/" + orderID + "/cancel")
	f.mu.Lock()
	defer f.mu.Unlock()
	raw, ok := f.orders[orderID]
	if !ok {
		return trading.MutationResult{}, &official.APIError{Code: 404, Body: `{"error":{"code":"order-not-found"}}`}
	}
	var v rawOrderView
	_ = json.Unmarshal(raw, &v)
	f.orders[orderID] = mustOrderJSON(orderID, v.Symbol, v.Side, "CANCELED", parseDecimal(v.Quantity),
		parseDecimal(v.Price), "2026-07-26T10:00:00+09:00")
	return trading.MutationResult{Kind: "cancel", OrderID: orderID, CurrentOrderID: orderID}, nil
}

func (f *fakeBroker) ModifyOrder(_ context.Context, intent orderintent.AmendIntent) (trading.MutationResult, error) {
	f.log("POST /orders/" + intent.OrderID + "/modify")
	f.mu.Lock()
	raw, ok := f.orders[intent.OrderID]
	f.mu.Unlock()
	if !ok {
		return trading.MutationResult{}, &official.APIError{Code: 404, Body: `{"error":{"code":"order-not-found"}}`}
	}
	var v rawOrderView
	_ = json.Unmarshal(raw, &v)

	price, qty := parseDecimal(v.Price), parseDecimal(v.Quantity)
	if intent.Price != nil {
		price = *intent.Price
	}
	if intent.Quantity != nil {
		qty = *intent.Quantity
	}

	current := intent.OrderID
	f.mu.Lock()
	if f.amendIssuesNewID {
		current = f.nextIDLocked("ord")
		f.orders[intent.OrderID] = mustOrderJSON(intent.OrderID, v.Symbol, v.Side, "REPLACED", qty, price, "")
	}
	f.orders[current] = mustOrderJSON(current, v.Symbol, v.Side, "PENDING", qty, price, "")
	f.mu.Unlock()
	return trading.MutationResult{Kind: "amend", OriginalOrderID: intent.OrderID, CurrentOrderID: current}, nil
}

func (f *fakeBroker) CreateConditionalOrder(_ context.Context, body official.ConditionalCreateBody) (domain.ConditionalOrderRef, error) {
	f.log("POST /conditional-orders " + body.Symbol + " key=" + body.ClientOrderID)
	canonical := fmt.Sprintf("cond|%s|%s|%s|%s", body.Symbol, body.Type, body.Quantity, body.First.TriggerPrice)
	if body.ClientOrderID != "" {
		f.mu.Lock()
		prior, seen := f.keys[body.ClientOrderID]
		f.mu.Unlock()
		if seen && prior.body == canonical && f.honourIdempotency {
			return domain.ConditionalOrderRef{ID: prior.orderID, ClientOrderID: body.ClientOrderID}, nil
		}
	}
	id := f.nextID("co")
	f.mu.Lock()
	f.conds[id] = domain.ConditionalOrder{
		ID: id, Type: body.Type, Status: "WATCHING", Symbol: body.Symbol, Market: "KR",
		Quantity: parseDecimal(body.Quantity), OrderType: body.OrderType,
		First: domain.ConditionalOrderCondition{
			Type: "STOP", Status: "WATCHING", TriggerPrice: parseDecimal(body.First.TriggerPrice),
		},
	}
	if body.ClientOrderID != "" {
		f.keys[body.ClientOrderID] = keyedOrder{orderID: id, body: canonical}
	}
	// A registered conditional reserves the shares, which is the property
	// StepSellableReserved is looking for.
	f.sellable[body.Symbol] -= parseDecimal(body.Quantity)
	f.mu.Unlock()
	return domain.ConditionalOrderRef{ID: id, ClientOrderID: body.ClientOrderID}, nil
}

func (f *fakeBroker) ModifyConditionalOrderRef(_ context.Context, id string, body official.ConditionalModifyBody) (domain.ConditionalOrderRef, error) {
	f.log("POST /conditional-orders/" + id + "/modify")
	f.mu.Lock()
	defer f.mu.Unlock()
	prior, ok := f.conds[id]
	if !ok {
		return domain.ConditionalOrderRef{}, &official.APIError{Code: 404, Body: `{"error":{"code":"conditional-order-not-found"}}`}
	}
	prior.First.TriggerPrice = parseDecimal(body.First.TriggerPrice)
	if !f.modifyIssuesNewID {
		f.conds[id] = prior
		return domain.ConditionalOrderRef{ID: id}, nil
	}
	newID := f.nextIDLocked("co")
	prior.ID = newID
	f.conds[newID] = prior
	if !f.modifyLeavesOldReadable {
		delete(f.conds, id)
	}
	return domain.ConditionalOrderRef{ID: newID}, nil
}

func (f *fakeBroker) CancelConditionalOrder(_ context.Context, id string) error {
	f.log("DELETE /conditional-orders/" + id)
	f.mu.Lock()
	defer f.mu.Unlock()
	c, ok := f.conds[id]
	if !ok {
		return &official.APIError{Code: 404, Body: `{"error":{"code":"conditional-order-not-found"}}`}
	}
	f.sellable[c.Symbol] += c.Quantity
	delete(f.conds, id)
	return nil
}

func (f *fakeBroker) nextIDLocked(prefix string) string {
	f.seq++
	return fmt.Sprintf("%s-%d", prefix, f.seq)
}

// expireKeys models the documented ten-minute idempotency window closing.
func (f *fakeBroker) expireKeys() {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.keys = map[string]keyedOrder{}
}

// restart models the process ending: the broker keeps (or, when the test asks,
// loses) its conditional orders.
func (f *fakeBroker) restart() {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.dropConditionalsOnRestart {
		f.conds = map[string]domain.ConditionalOrder{}
	}
}

func mustOrderJSON(id, symbol, side, status string, qty, price float64, canceledAt string) json.RawMessage {
	m := map[string]any{
		"orderId":     id,
		"symbol":      symbol,
		"side":        side,
		"orderType":   "LIMIT",
		"timeInForce": "DAY",
		"status":      status,
		"quantity":    trim(qty),
		"price":       trim(price),
		"currency":    "KRW",
		"orderedAt":   "2026-07-26T09:00:00+09:00",
		"canceledAt":  canceledAt,
		"execution": map[string]any{
			"filledQuantity":     "0",
			"averageFilledPrice": nil,
			"filledAmount":       nil,
			"commission":         nil,
			"tax":                nil,
			"filledAt":           nil,
			"settlementDate":     nil,
		},
	}
	b, err := json.Marshal(m)
	if err != nil {
		panic(err)
	}
	return b
}

// --- a scripted operator ------------------------------------------------------

// operator answers both gates from a script: the run-wide batch approval and, when
// the test asks for --confirm-each, each mutation's own prompt.
type operator struct {
	// answer decides each per-mutation prompt. Returning nil is an approval.
	answer func(m Mutation) error
	// answerBatch decides the run-wide approval.
	answerBatch func(b Batch) error

	seen    []Mutation
	batches []Batch
}

func alwaysConfirm() *operator {
	return &operator{
		answer:      func(Mutation) error { return nil },
		answerBatch: func(Batch) error { return nil },
	}
}

// refuseWhere refuses selected per-mutation prompts. It approves the batch, since
// the property it exists to test is what one refused *step* costs — which is only
// observable under --confirm-each.
func refuseWhere(pred func(m Mutation) bool) *operator {
	op := alwaysConfirm()
	op.answer = func(m Mutation) error {
		if pred(m) {
			return ErrRefused
		}
		return nil
	}
	return op
}

// refuseBatch declines the run-wide approval.
func refuseBatch(err error) *operator {
	op := alwaysConfirm()
	op.answerBatch = func(Batch) error { return err }
	return op
}

func (o *operator) confirmer() Confirmer {
	return func(m Mutation) error {
		o.seen = append(o.seen, m)
		return o.answer(m)
	}
}

func (o *operator) batchConfirmer() BatchConfirmer {
	return func(b Batch) error {
		o.batches = append(o.batches, b)
		return o.answerBatch(b)
	}
}

// lastBatch is the plan the operator was most recently shown.
func (o *operator) lastBatch(t *testing.T) Batch {
	t.Helper()
	if len(o.batches) == 0 {
		t.Fatal("the operator was never shown a batch to approve")
	}
	return o.batches[len(o.batches)-1]
}

func (o *operator) actions() []string {
	out := make([]string, 0, len(o.seen))
	for _, m := range o.seen {
		out = append(out, string(m.Step)+": "+m.Action)
	}
	return out
}

// --- harness ------------------------------------------------------------------

type harness struct {
	t        *testing.T
	broker   *fakeBroker
	op       *operator
	record   string
	now      time.Time
	slept    time.Duration
	instance int
}

func newHarness(t *testing.T, broker *fakeBroker, op *operator) *harness {
	t.Helper()
	return &harness{
		t:      t,
		broker: broker,
		op:     op,
		record: t.TempDir() + "/verify.jsonl",
		now:    time.Date(2026, 7, 26, 1, 0, 0, 0, time.UTC),
	}
}

// run performs one invocation of the tool: a fresh process identity, the record
// re-read from disk, the runner built and walked.
func (h *harness) run(opts Options) (Summary, error) {
	h.t.Helper()
	runner, closeRecord, err := h.build(opts)
	if err != nil {
		return Summary{}, err
	}
	defer closeRecord()
	return runner.Run(context.Background())
}

// runner builds one invocation without walking it, for the tests that inspect or
// adjust the plan before the run starts.
func (h *harness) runner(t *testing.T, opts Options) *Runner {
	t.Helper()
	runner, closeRecord, err := h.build(opts)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	t.Cleanup(closeRecord)
	return runner
}

func (h *harness) build(opts Options) (*Runner, func(), error) {
	h.t.Helper()
	prior, err := LoadEntries(h.record)
	if err != nil {
		h.t.Fatalf("LoadEntries: %v", err)
	}
	rec, err := OpenRecorder(h.record)
	if err != nil {
		h.t.Fatalf("OpenRecorder: %v", err)
	}
	closeRecord := func() { _ = rec.Close() }

	h.instance++
	opts.Broker = h.broker
	opts.Recorder = rec
	opts.Confirm = h.op.confirmer()
	opts.ConfirmBatch = h.op.batchConfirmer()
	opts.Prior = prior
	if opts.AccountRef == "" {
		opts.AccountRef = "123-45-678901"
	}
	if opts.Symbol == "" {
		opts.Symbol = "005930"
	}
	opts.Now = func() time.Time {
		h.now = h.now.Add(50 * time.Millisecond)
		return h.now
	}
	opts.Sleep = func(_ context.Context, d time.Duration) error {
		h.slept += d
		h.now = h.now.Add(d)
		if h.broker.expireKeysOnSleep {
			h.broker.expireKeys()
		}
		return nil
	}
	opts.Process = Process{
		PID:        1000 + h.instance,
		InstanceID: fmt.Sprintf("proc-%d", h.instance),
		StartedAt:  h.now,
	}

	runner, err := New(opts)
	if err != nil {
		closeRecord()
		return nil, func() {}, err
	}
	return runner, closeRecord, nil
}

func (h *harness) entries() []Entry {
	h.t.Helper()
	entries, err := LoadEntries(h.record)
	if err != nil {
		h.t.Fatalf("LoadEntries: %v", err)
	}
	return entries
}

func (h *harness) verdict(id StepID) Verdict {
	h.t.Helper()
	e, ok := LastEntry(h.entries(), id)
	if !ok {
		return ""
	}
	return e.Verdict
}

func (h *harness) observation(id StepID, key string) (string, bool) {
	h.t.Helper()
	for _, e := range h.entries() {
		if e.StepID != id {
			continue
		}
		for _, o := range e.Observations {
			if o.Key == key {
				return o.Value, true
			}
		}
	}
	return "", false
}
