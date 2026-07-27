package console

// orders_test.go covers the orders screen (change console-orders-screen, §4-§6).
//
// The claims worth testing are the ones an operator would be misled by:
//
//	nothing is invented   what the broker did not send renders as — and never as 0,
//	                      because a live order with an average fill price of 0 says
//	                      it filled.
//	nothing is summed     losing the conditional list does not shorten the count.
//	                      A confident zero has to be unreachable.
//	nothing is hidden     a truncated page is "N건 이상", a filtered table shows the
//	                      whole count beside the filtered one, and an unreadable
//	                      ledger says 불명 rather than "somebody placed it by hand".
//	the budget holds      one refresh is one seam call, a second tab and a changed
//	                      filter are free, and a verification suspends refreshing.

import (
	"context"
	"errors"
	"net/http"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/runlock"
)

// --- a reader that counts ----------------------------------------------------------

// countingOrders is the seam the screen is given. It counts, because "one
// refresh is one call" is a rate-budget claim and counting is the only honest way
// to check it.
type countingOrders struct {
	mu     sync.Mutex
	calls  int
	lists  OrdersReading
	err    error
	onCall func()
}

func (o *countingOrders) Orders(context.Context) (OrdersReading, error) {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.calls++
	if o.onCall != nil {
		o.onCall()
	}
	if o.err != nil {
		return OrdersReading{}, o.err
	}
	return o.lists, nil
}

func (o *countingOrders) count() int {
	o.mu.Lock()
	defer o.mu.Unlock()
	return o.calls
}

// livePlainOrder is a market order the broker has not filled: price null, and the
// whole execution object null. It is the normal shape of a live order, not an
// edge case.
func livePlainOrder(id, symbol string) OrderRecord {
	return OrderRecord{
		ID: id, Symbol: symbol, Side: "BUY", Kind: "MARKET", Status: "PENDING",
		Market: "KR", Currency: "KRW",
		Quantity: "10", Price: "", FilledQuantity: "", AverageFilledPrice: "",
		OrderedAt: "2026-07-27T00:30:00Z",
	}
}

// filledPlainOrder is a finished order with every number present.
func filledPlainOrder(id, symbol string) OrderRecord {
	return OrderRecord{
		ID: id, Symbol: symbol, Side: "SELL", Kind: "LIMIT", Status: "FILLED",
		Market: "US", Currency: "USD",
		Quantity: "5", Price: "200", FilledQuantity: "5", AverageFilledPrice: "201.5",
		OrderedAt: "2026-07-27T00:10:00Z",
	}
}

func watchingConditional(id, symbol string) ConditionalRecord {
	return ConditionalRecord{
		ID: id, Symbol: symbol, Market: "KR", Kind: "SINGLE", Status: "WATCHING",
		Quantity: "10", TriggerPrice: "70000", OrderPrice: "",
		ConditionKind: "STOP", CreatedAt: "2026-07-27T00:20:00Z",
	}
}

// ordersHarness wires the screen with a counting seam, a fake clock and a
// journal.
type ordersHarness struct {
	*harness
	reader *countingOrders
	clock  *fakeClock
	// journalPath is the ledger the origin column joins against, empty when the
	// test wants it unwired.
	journalPath string
}

func newOrdersHarness(t *testing.T, reader *countingOrders, tweak ...func(*Options)) *ordersHarness {
	t.Helper()
	clk := newFakeClock()
	oh := &ordersHarness{reader: reader, clock: clk}
	oh.harness = newHarness(t, func(o *Options) {
		o.Orders = reader
		o.Now = clk.Now
		for _, f := range tweak {
			f(o)
		}
		oh.journalPath = o.JournalPath
	})
	return oh
}

// --- §4 the seam and the cache ------------------------------------------------------

// TestOneOrdersRefreshIsOneSeamCallAndTheTTLIsTheSpecFloorOrMore.
//
// Behind the one seam call the adapter makes two broker calls — the plain page
// and the conditional page — which is the rate-budget contract for this screen.
// The console's own half of that contract is that a refresh happens once per TTL,
// lazily, with no server-side poller: a second tab inside the TTL is free.
func TestOneOrdersRefreshIsOneSeamCallAndTheTTLIsTheSpecFloorOrMore(t *testing.T) {
	if ordersTTL < 15*time.Second {
		t.Fatalf("ordersTTL is %s; the operator-console spec fixes a floor of 15 seconds", ordersTTL)
	}

	reader := &countingOrders{lists: OrdersReading{Plain: []OrderRecord{livePlainOrder("o-1", "005930")}}}
	h := newOrdersHarness(t, reader)
	h.authenticate(t)

	h.get(t, "/orders")
	if reader.count() != 1 {
		t.Fatalf("the first render made %d seam call(s), want 1", reader.count())
	}
	h.get(t, "/orders")
	h.get(t, "/orders")
	if reader.count() != 1 {
		t.Errorf("a re-render inside the TTL made %d seam call(s); refresh-mashing and a second tab "+
			"are free or the budget is not bounded by the TTL", reader.count())
	}

	h.clock.advance(ordersTTL)
	h.get(t, "/orders")
	if reader.count() != 2 {
		t.Errorf("after the TTL the screen made %d call(s) in total, want 2", reader.count())
	}

	// And nothing refreshes on its own: the count does not move without a request.
	before := reader.count()
	h.clock.advance(10 * ordersTTL)
	time.Sleep(20 * time.Millisecond)
	if reader.count() != before {
		t.Errorf("the seam was called %d time(s) with no request; there is no server-side poller",
			reader.count()-before)
	}
}

// TestChangingAFilterCostsNoBrokerCall.
//
// D6. Passing the filters to the broker would make /orders?state=live and
// ?state=closed separate cache keys, and the two calls a refresh costs would
// become four out of the same budget. The filters are a view of the reading.
func TestChangingAFilterCostsNoBrokerCall(t *testing.T) {
	reader := &countingOrders{lists: OrdersReading{
		Plain:       []OrderRecord{livePlainOrder("o-1", "005930"), filledPlainOrder("o-2", "AAPL")},
		Conditional: []ConditionalRecord{watchingConditional("co-1", "005930")},
	}}
	h := newOrdersHarness(t, reader)
	h.authenticate(t)

	for _, path := range []string{
		"/orders", "/orders?state=live", "/orders?state=closed",
		"/orders?market=KR", "/orders?market=US&side=SELL",
	} {
		h.get(t, path)
	}
	if reader.count() != 1 {
		t.Errorf("five filter combinations cost %d seam call(s), want 1; a filter passed to the "+
			"broker splits the cache and multiplies the budget", reader.count())
	}
}

// TestTheOrdersRefreshYieldsToALiveVerification.
//
// The same judgement the holdings cache uses, and the same reason: a step lost to
// a 429 costs a person another supervised run, and a screen is worth nothing in
// particular. It is asserted through the run-lock marker because that is the half
// a test can drive without starting a verification.
func TestTheOrdersRefreshYieldsToALiveVerification(t *testing.T) {
	reader := &countingOrders{lists: OrdersReading{Plain: []OrderRecord{livePlainOrder("o-1", "005930")}}}
	lock := filepath.Join(t.TempDir(), "run.lock")
	h := newOrdersHarness(t, reader, func(o *Options) { o.RunLockPath = lock })
	h.authenticate(t)

	holdRunLock(t, lock, h.clock.Now())
	page := body(t, h.get(t, "/orders"))
	if reader.count() != 0 {
		t.Errorf("the screen made %d seam call(s) while a verification was running", reader.count())
	}
	if !strings.Contains(page, "검증 중") {
		t.Error("the page does not say the refresh is suspended; a cold list with no reason reads " +
			"as an empty account")
	}
	if !strings.Contains(page, string(reasonVerifySuspended)) {
		t.Errorf("the unmeasured count does not carry %s; the operator cannot tell waiting from "+
			"fixing", reasonVerifySuspended)
	}

	// Stale marker: the yielding stops rather than wedging the screen forever.
	holdRunLock(t, lock, h.clock.Now().Add(-2*runlock.StaleAfter))
	h.get(t, "/orders")
	if reader.count() != 1 {
		t.Errorf("a stale run marker still suppressed the refresh (%d calls)", reader.count())
	}
}

// TestAConditionalReadThatFailedIsNeverAddedIntoTheTotal.
//
// This is the failure the whole change exists to prevent, and the one it could
// most easily commit itself. verifylive's cleanup counts a leftover conditional
// as filling the exposure cap exactly like a plain order, and M18 measured that
// one survives the process that made it. A screen that answered "미체결 0건" while
// a conditional leftover blocked the next verification would be the thing it was
// built to catch.
func TestAConditionalReadThatFailedIsNeverAddedIntoTheTotal(t *testing.T) {
	reader := &countingOrders{lists: OrdersReading{
		Plain:            []OrderRecord{livePlainOrder("o-1", "005930")},
		Conditional:      nil,
		ConditionalError: "429 Too Many Requests",
	}}
	h := newOrdersHarness(t, reader)
	h.authenticate(t)

	page := body(t, h.get(t, "/orders"))
	if !strings.Contains(page, "1건") {
		t.Error("the plain count is missing; the half that was measured is still measured")
	}
	if !strings.Contains(page, "429 Too Many Requests") {
		t.Error("the conditional read's failure is not on the page")
	}
	if !strings.Contains(page, string(reasonBrokerReadFailed)) {
		t.Errorf("the conditional cell does not carry %s", reasonBrokerReadFailed)
	}

	// The combined value must be unmeasured. Reading it off the view rather than
	// off the HTML, because "the string 1건 appears twice" would pass for a page
	// that had quietly used the plain count as the total.
	view := h.Console.orders(context.Background(), orderFilterChoice{})
	if view.Live.Known() {
		t.Errorf("the combined live count is measured (%q) while the conditional list failed; a "+
			"total that means \"the plain orders only\" is the confidently short number that hides "+
			"a leftover", view.Live.Value())
	}
	if !strings.Contains(view.Live.Why(), "합치지 않는다") {
		t.Errorf("the combined value's reason does not say the two are not added: %q", view.Live.Why())
	}
	if !view.PlainLive.Known() || view.PlainLive.Value() != "1건" {
		t.Errorf("the plain count is %q; one list failing must not take the other down",
			view.PlainLive.Value())
	}
}

// TestAPlainReadThatFailedIsNeverAddedIntoTheTotalEither.
//
// The mirror. It is asserted separately because "the conditional half is the
// special one" is exactly the assumption that produces a total which is right in
// one direction only.
func TestAPlainReadThatFailedIsNeverAddedIntoTheTotalEither(t *testing.T) {
	reader := &countingOrders{lists: OrdersReading{
		PlainError:  "500 Internal Server Error",
		Conditional: []ConditionalRecord{watchingConditional("co-1", "005930")},
	}}
	h := newOrdersHarness(t, reader)
	h.authenticate(t)

	view := h.Console.orders(context.Background(), orderFilterChoice{})
	if view.Live.Known() {
		t.Errorf("the combined live count is measured (%q) while the plain list failed", view.Live.Value())
	}
	if !view.ConditionalLive.Known() || view.ConditionalLive.Value() != "1건" {
		t.Errorf("the conditional count is %q, want 1건", view.ConditionalLive.Value())
	}
}

// TestAWholeSeamFailureIsUnmeasuredAndNotAnEmptyAccount.
func TestAWholeSeamFailureIsUnmeasuredAndNotAnEmptyAccount(t *testing.T) {
	reader := &countingOrders{err: errors.New("no Open API credentials")}
	h := newOrdersHarness(t, reader)
	h.authenticate(t)

	page := body(t, h.get(t, "/orders"))
	if strings.Contains(page, "주문이 없다") {
		t.Error("a failed read renders as \"no orders\"; that is the console asserting something " +
			"about an account it could not look at")
	}
	if !strings.Contains(page, "no Open API credentials") {
		t.Error("the failure is not on the page")
	}
	if !strings.Contains(page, "필터를 쓸 수 없다") {
		t.Error("the filters are still offered on an unmeasured list; \"0/—건\" reads as " +
			"\"0 matched\"")
	}
}

// TestAnUnwiredOrdersSeamSaysSoAndLeavesEveryOtherScreenWorking.
//
// Task 4.7. A build with no order reader is not a broken console: it is a console
// that cannot answer one question, and it has to say which one.
func TestAnUnwiredOrdersSeamSaysSoAndLeavesEveryOtherScreenWorking(t *testing.T) {
	h := newHarness(t)
	h.authenticate(t)

	page := body(t, h.get(t, "/orders"))
	if !strings.Contains(page, string(reasonSeamUnwired)) {
		t.Errorf("an unwired orders screen does not say %s", reasonSeamUnwired)
	}
	if strings.Contains(page, "0건</strong>") {
		t.Error("an unwired seam produced a count of zero; 0 and \"nobody asked\" are different " +
			"facts about somebody's money")
	}
	for _, path := range []string{"/", "/positions", "/history", "/dashboard"} {
		if resp := h.get(t, path); resp.StatusCode != http.StatusOK {
			t.Errorf("GET %s returned %d with the orders seam unwired; one missing seam must not "+
				"take another screen down", path, resp.StatusCode)
		}
	}
}

// --- §5 the screen ------------------------------------------------------------------

// TestALiveOrderRendersItsMissingFillAsADashAndNeverAsZero.
//
// The API sends price as null for a market order and the whole execution object
// as null while an order is pending, so every live order arrives with no filled
// quantity and no average fill price. Rendered as 0 those rows say the order
// filled at nothing, which is the opposite of what is true.
func TestALiveOrderRendersItsMissingFillAsADashAndNeverAsZero(t *testing.T) {
	reader := &countingOrders{lists: OrdersReading{
		Plain: []OrderRecord{livePlainOrder("o-1", "005930"), filledPlainOrder("o-2", "AAPL")},
	}}
	h := newOrdersHarness(t, reader)
	h.authenticate(t)
	h.get(t, "/orders")

	view := h.Console.orders(context.Background(), orderFilterChoice{})
	var live, closed orderRow
	for _, r := range view.Rows {
		switch r.ID {
		case "o-1":
			live = r
		case "o-2":
			closed = r
		}
	}
	for _, tc := range []struct{ field, got, want string }{
		{"live average fill price", live.AvgPrice, "—"},
		{"live filled quantity", live.Filled, "—"},
		{"live price", live.Price, "—"},
		{"live quantity", live.Quantity, "10"},
		{"closed average fill price", closed.AvgPrice, "201.5"},
		{"closed filled quantity", closed.Filled, "5"},
	} {
		if tc.got != tc.want {
			t.Errorf("%s = %q, want %q", tc.field, tc.got, tc.want)
		}
	}
	// And the whole set of columns is on the page.
	page := body(t, h.get(t, "/orders"))
	for _, header := range []string{
		"시각(UTC)", "심볼", "시장", "방향", "상태", "주문수량", "체결수량", "주문가", "평균체결가",
		"주문번호", "발주 주체",
	} {
		if !strings.Contains(page, header) {
			t.Errorf("the table has no %q column", header)
		}
	}
	if strings.Contains(page, "종목명") {
		t.Error("the table has a 종목명 column. The value is in neither the adapted order nor the " +
			"raw payload, so every row would be — and a column that is always — is noise")
	}
}

// TestAConditionalOrderIsCountedAsLiveAndMarkedAsWhatItIs.
func TestAConditionalOrderIsCountedAsLiveAndMarkedAsWhatItIs(t *testing.T) {
	reader := &countingOrders{lists: OrdersReading{
		Plain:       []OrderRecord{filledPlainOrder("o-2", "AAPL")},
		Conditional: []ConditionalRecord{watchingConditional("co-1", "005930")},
	}}
	h := newOrdersHarness(t, reader)
	h.authenticate(t)
	h.get(t, "/orders")

	view := h.Console.orders(context.Background(), orderFilterChoice{})
	if !view.Live.Known() || view.Live.Value() != "1건" {
		t.Errorf("the live count is %q; the only live thing on this account is a conditional order "+
			"and it fills the exposure cap like any other", view.Live.Value())
	}

	page := body(t, h.get(t, "/orders"))
	if !strings.Contains(page, "조건주문") {
		t.Error("the conditional row is not marked; cancelling one takes a different call")
	}
	if !strings.Contains(page, "co-1") {
		t.Error("the conditional order's id is not on the page")
	}
	// It is live, and its unfilled columns are dashes rather than zeroes.
	for _, r := range view.Rows {
		if !r.Conditional {
			continue
		}
		if !r.Live {
			t.Error("the conditional row is not live; the list is the broker's OPEN group")
		}
		if r.Filled != "—" || r.AvgPrice != "—" {
			t.Errorf("a conditional order reports a fill (%q/%q); it has not become an order yet",
				r.Filled, r.AvgPrice)
		}
	}
}

// TestATruncatedPageIsACountWithAFloorAndNotANumber.
func TestATruncatedPageIsACountWithAFloorAndNotANumber(t *testing.T) {
	reader := &countingOrders{lists: OrdersReading{
		Plain:          []OrderRecord{livePlainOrder("o-1", "005930")},
		PlainTruncated: true,
	}}
	h := newOrdersHarness(t, reader)
	h.authenticate(t)

	page := body(t, h.get(t, "/orders"))
	if !strings.Contains(page, "1건 이상") {
		t.Error("a truncated page reports a settled count; the rows the broker did not send may " +
			"include the leftover the operator is looking for")
	}
	if !strings.Contains(page, "페이지가 잘렸다") {
		t.Error("the page does not say it was truncated")
	}
}

// TestTheOriginColumnSaysUnknownWhenTheLedgerCouldNotBeRead.
//
// Reading an unreadable ledger as "somebody placed these by hand" makes a working
// engine look idle, and an idle engine is what an operator restarts.
func TestTheOriginColumnSaysUnknownWhenTheLedgerCouldNotBeRead(t *testing.T) {
	reader := &countingOrders{lists: OrdersReading{
		Plain: []OrderRecord{livePlainOrder("o-1", "005930")},
	}}
	h := newOrdersHarness(t, reader) // no JournalPath: unwired
	h.authenticate(t)
	h.get(t, "/orders")

	view := h.Console.orders(context.Background(), orderFilterChoice{})
	for _, r := range view.Rows {
		if r.Origin != originUnknown {
			t.Errorf("row %s has origin %q with no ledger; it must be 불명", r.ID, r.Origin)
		}
		if r.OriginText() != "불명" {
			t.Errorf("row %s renders its origin as %q", r.ID, r.OriginText())
		}
	}

	page := body(t, h.get(t, "/orders"))
	// The cell, not the page: the notice below the table explains why "그 밖" is
	// deliberately NOT used, and a substring search would fail on the explanation.
	if strings.Contains(page, "<td>그 밖</td>") {
		t.Error("an unreadable ledger produced a \"그 밖\" cell")
	}
	// The notice is a page-level sentence, said once, not repeated per row.
	if n := strings.Count(page, `발주 주체는 전부 "불명"이다`); n != 1 {
		t.Errorf("the ledger notice appears %d times; it is one sentence about the page", n)
	}
}

// TestTheOriginColumnTellsAnEngineOrderFromAnyOther.
func TestTheOriginColumnTellsAnEngineOrderFromAnyOther(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "journal.db")
	seedEngineJournal(t, path, ordersJournalFixture)

	reader := &countingOrders{lists: OrdersReading{
		Plain: []OrderRecord{livePlainOrder("engine-order", "005930"), filledPlainOrder("hand-order", "AAPL")},
	}}
	h := newOrdersHarness(t, reader, func(o *Options) { o.JournalPath = path })
	h.authenticate(t)
	h.get(t, "/orders")

	view := h.Console.orders(context.Background(), orderFilterChoice{})
	want := map[string]orderOrigin{"engine-order": originEngine, "hand-order": originOther}
	for _, r := range view.Rows {
		if w, ok := want[r.ID]; ok && r.Origin != w {
			t.Errorf("row %s has origin %q, want %q", r.ID, r.Origin, w)
		}
	}
	page := body(t, h.get(t, "/orders"))
	for _, label := range []string{"엔진 발주", "그 밖"} {
		if !strings.Contains(page, label) {
			t.Errorf("the page does not distinguish %q", label)
		}
	}
}

// ordersJournalFixture records one attempt the broker acked with the order id the
// screen will see.
const ordersJournalFixture = `
INSERT INTO intents (id, created_at, market, trading_day, account_ref, symbol, side, order_type,
                     time_in_force, quantity, price, currency, source, fingerprint, notes)
VALUES ('intent-1','2026-07-27T00:30:00Z','kr','2026-07-27','123-45-678901','005930','BUY','MARKET',
        'DAY','10',NULL,'KRW','engine','fp-1','');
INSERT INTO mutation_attempts (id, intent_id, kind, state, attempt_no, broker_order_id,
                               fingerprint, recorded_at)
VALUES ('a-1','intent-1','PLACE','RECORDED',1,'engine-order','fp-1','2026-07-27T00:30:00Z');
`

// TestAFilteredTableShowsTheWholeCountBesideTheFilteredOne.
//
// A filtered screen with only its own count on it reads as "these are all the
// orders", and the row that is not on the screen is the leftover.
func TestAFilteredTableShowsTheWholeCountBesideTheFilteredOne(t *testing.T) {
	reader := &countingOrders{lists: OrdersReading{
		Plain: []OrderRecord{livePlainOrder("o-1", "005930"), filledPlainOrder("o-2", "AAPL")},
	}}
	h := newOrdersHarness(t, reader)
	h.authenticate(t)

	page := body(t, h.get(t, "/orders?market=KR"))
	if !strings.Contains(page, ">1</strong>/2건") {
		t.Error("the filtered screen does not show N/M; a filter that hides the count hides the rows")
	}
	if strings.Contains(page, "AAPL") {
		t.Error("the market filter did not apply")
	}

	view := h.Console.orders(context.Background(), orderFilterChoice{Market: "KR"})
	if view.Shown != 1 || view.Total != 2 {
		t.Errorf("Shown/Total = %d/%d, want 1/2", view.Shown, view.Total)
	}

	// An unrecognised value is dropped rather than matched against nothing: a
	// hand-typed ?market=KRW emptying the table would read as a measurement.
	all := h.Console.orders(context.Background(), filterChoiceFromQuery(t, "/orders?market=KRW"))
	if all.Shown != 2 {
		t.Errorf("an unrecognised filter value narrowed the table to %d row(s); it means nothing "+
			"and an empty table would read as a measurement", all.Shown)
	}
}

// filterChoiceFromQuery parses a choice out of a URL, for the tests that want to
// drive the query-string reader directly.
func filterChoiceFromQuery(t *testing.T, target string) orderFilterChoice {
	t.Helper()
	req, err := http.NewRequest(http.MethodGet, target, nil)
	if err != nil {
		t.Fatalf("building %s: %v", target, err)
	}
	return filterChoiceFrom(req)
}

// TestTheOrdersScreenReloadsItselfAtTheCacheTTLAndNoFaster.
func TestTheOrdersScreenReloadsItselfAtTheCacheTTLAndNoFaster(t *testing.T) {
	if got := (ordersPage{}).RefreshSeconds(); got != int(ordersTTL/time.Second) {
		t.Errorf("the reload period is %ds and the cache TTL is %s; a period under the TTL is a "+
			"reload that spends broker calls faster than the budget allows", got, ordersTTL)
	}
	reader := &countingOrders{}
	h := newOrdersHarness(t, reader)
	h.authenticate(t)
	page := body(t, h.get(t, "/orders"))
	if !strings.Contains(page, `http-equiv="refresh" content="15"`) {
		t.Error("the page carries no meta refresh at the cache TTL")
	}
}

// TestTheOrdersScreenOffersNoWayToActOnAnOrder.
//
// D8, and the 2026-07-27 user instruction about confirmation friction. There is
// nothing on this page to submit, so there is nothing to confirm — and the way
// that stays true is that a form appearing here fails a test.
func TestTheOrdersScreenOffersNoWayToActOnAnOrder(t *testing.T) {
	reader := &countingOrders{lists: OrdersReading{
		Plain:       []OrderRecord{livePlainOrder("o-1", "005930")},
		Conditional: []ConditionalRecord{watchingConditional("co-1", "005930")},
	}}
	h := newOrdersHarness(t, reader)
	h.authenticate(t)
	page := body(t, h.get(t, "/orders"))

	for _, banned := range []string{
		"<form", "<button", "<input", "method=\"post\"", "method=post",
		// A cancel/amend link would be an act one GET away.
		"/orders/cancel", "/orders/amend", "/verify/order",
		// Confirmation friction of any kind, in the spellings this repo has used.
		"확인 문구", "확인 문자열", "타이핑", "입력하여 확인", "다시 한 번",
	} {
		if strings.Contains(page, banned) {
			t.Errorf("the orders screen contains %q; it is a 관측창 — there is nothing to submit "+
				"and therefore nothing to confirm", banned)
		}
	}
	if !strings.Contains(page, "취소가 필요하면") {
		t.Error("the page does not say where cancelling actually happens")
	}
}

// TestAnUnresolvableBrokerStateIsNotReportedAsClosed.
//
// internal/brokerstate is the judge, and its whole discipline is that a
// combination the documentation cannot explain fails closed. A screen that filed
// those rows under 종결 would drop a live order out of the count.
func TestAnUnresolvableBrokerStateIsNotReportedAsClosed(t *testing.T) {
	odd := livePlainOrder("o-odd", "005930")
	odd.Status = "SOMETHING_NEW"
	reader := &countingOrders{lists: OrdersReading{Plain: []OrderRecord{odd}}}
	h := newOrdersHarness(t, reader)
	h.authenticate(t)
	h.get(t, "/orders")

	view := h.Console.orders(context.Background(), orderFilterChoice{})
	if len(view.Rows) != 1 {
		t.Fatalf("want 1 row, got %d", len(view.Rows))
	}
	row := view.Rows[0]
	if !row.Unresolved {
		t.Error("a status this build does not know was resolved anyway")
	}
	if row.StateText() != "상태 불명" {
		t.Errorf("the row says %q, want 상태 불명", row.StateText())
	}
	if !view.Live.Known() || view.Live.Value() != "1건" {
		t.Errorf("the live count is %q; an unresolved order is not a finished one", view.Live.Value())
	}
	if !strings.Contains(body(t, h.get(t, "/orders")), "SOMETHING_NEW") {
		t.Error("the broker's own status string is not shown")
	}
}

// TestAnUnparseableOrderTimeIsShownVerbatimRatherThanDropped.
func TestAnUnparseableOrderTimeIsShownVerbatimRatherThanDropped(t *testing.T) {
	odd := livePlainOrder("o-1", "005930")
	odd.OrderedAt = "27/07/2026 09:30"
	reader := &countingOrders{lists: OrdersReading{Plain: []OrderRecord{odd}}}
	h := newOrdersHarness(t, reader)
	h.authenticate(t)

	page := body(t, h.get(t, "/orders"))
	if !strings.Contains(page, "27/07/2026 09:30") {
		t.Error("an unparseable timestamp was dropped; the row then looks like an order with no time")
	}
	if !strings.Contains(page, "RFC3339로 해석되지 않은") {
		t.Error("the page does not explain the verbatim timestamp")
	}
}

// --- §5.10 and §6 --------------------------------------------------------------------

// TestTheOverviewOpenOrdersPanelHasARealValueOnceTheSeamIsWired.
//
// console-operator-overview left that panel unmeasured with the reason
// seam_unwired, which is a request. This is the answer to it — and the overview
// still makes no broker call of its own, so the value comes from the cache the
// orders screen filled.
func TestTheOverviewOpenOrdersPanelHasARealValueOnceTheSeamIsWired(t *testing.T) {
	reader := &countingOrders{lists: OrdersReading{
		Plain:       []OrderRecord{livePlainOrder("o-1", "005930"), filledPlainOrder("o-2", "AAPL")},
		Conditional: []ConditionalRecord{watchingConditional("co-1", "005930")},
	}}
	h := newOrdersHarness(t, reader)
	h.authenticate(t)

	// Cold: the overview says so and points at the screen that fills it, rather
	// than spending the budget itself.
	cold := h.Console.overview(context.Background())
	if cold.Open.Count.Known() {
		t.Errorf("the overview reported %q before anything filled the cache", cold.Open.Count.Value())
	}
	if cold.Open.Count.Code() != string(reasonNeverFetched) {
		t.Errorf("the cold panel says %q, want %s", cold.Open.Count.Code(), reasonNeverFetched)
	}
	if reader.count() != 0 {
		t.Errorf("the overview made %d seam call(s); its contract is zero broker calls per render",
			reader.count())
	}
	if cold.Open.Where != "/orders" {
		t.Errorf("the cold panel points at %q, want /orders", cold.Open.Where)
	}

	h.get(t, "/orders")
	warm := h.Console.overview(context.Background())
	if !warm.Open.Count.Known() || warm.Open.Count.Value() != "2건" {
		t.Errorf("the overview reports %q, want 2건 (one live plain order and one conditional)",
			warm.Open.Count.Value())
	}
	if reader.count() != 1 {
		t.Errorf("the overview refreshed the cache (%d calls); it peeks", reader.count())
	}
	if warm.Open.Count.Code() == string(reasonSeamUnwired) {
		t.Error("the overview's open-orders panel still says seam 미배선 with the seam wired")
	}
	if !strings.Contains(body(t, h.get(t, "/dashboard")), "2건") {
		t.Error("the overview page does not render the live count it now has")
	}
}

// TestTheOrdersScreenIsReachableFromEveryOtherScreen.
func TestTheOrdersScreenIsReachableFromEveryOtherScreen(t *testing.T) {
	h := newOrdersHarness(t, &countingOrders{})
	h.authenticate(t)
	for _, path := range []string{"/", "/dashboard", "/positions", "/history", "/orders"} {
		page := body(t, h.get(t, path))
		if !strings.Contains(page, `href="/orders"`) {
			t.Errorf("%s has no link to the orders screen", path)
		}
	}
	// And the refusal names it, so a mistyped path lists the screens that exist.
	if !strings.Contains(body(t, h.get(t, "/nope")), "주문") {
		t.Error("the 404 body does not list the orders screen")
	}
}

// TestABrokerValueThatIsNotADecimalIsPrintedVerbatim.
//
// The thousands grouping is positional, so a string this build did not expect
// would come out as a different string — and preserving what the broker actually
// said is the one thing the raw read exists for. An unexpected value has to reach
// the operator unchanged so they can compare it with the payload.
func TestABrokerValueThatIsNotADecimalIsPrintedVerbatim(t *testing.T) {
	for _, tc := range []struct{ in, want string }{
		{"", "—"},
		{"0", "0"},
		{"1234567", "1,234,567"},
		{"201.5", "201.5"},
		{"-2000", "-2,000"},
		{"NOT_A_NUMBER", "NOT_A_NUMBER"},
		{"1.2.3", "1.2.3"},
	} {
		if got := dashUnlessSent(tc.in); got != tc.want {
			t.Errorf("dashUnlessSent(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}
