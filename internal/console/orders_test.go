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
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
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
	if reader.lists.AccountRef == "" {
		reader.lists.AccountRef = "123-45-678901"
	}
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
// Behind the one seam call the adapter makes three broker calls — the pending
// group, the finished page and the conditional page — which is the rate-budget
// contract for this screen.
// The console's own half of that contract is that a refresh happens once per TTL,
// lazily, with no server-side poller: a second tab inside the TTL is free.
func TestOneOrdersRefreshIsOneSeamCallAndTheTTLIsTheSpecFloorOrMore(t *testing.T) {
	if ordersTTL < 15*time.Second {
		t.Fatalf("ordersTTL is %s; the operator-console spec fixes a floor of 15 seconds", ordersTTL)
	}

	reader := &countingOrders{lists: OrdersReading{Open: []OrderRecord{livePlainOrder("o-1", "005930")}}}
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
// ?state=closed separate cache keys, and the three calls a refresh costs would
// become six out of the same budget. The filters are a view of the reading.
func TestChangingAFilterCostsNoBrokerCall(t *testing.T) {
	reader := &countingOrders{lists: OrdersReading{
		Open:        []OrderRecord{livePlainOrder("o-1", "005930")},
		Closed:      []OrderRecord{filledPlainOrder("o-2", "AAPL")},
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
	reader := &countingOrders{lists: OrdersReading{Open: []OrderRecord{livePlainOrder("o-1", "005930")}}}
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
		Open:             []OrderRecord{livePlainOrder("o-1", "005930")},
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
	if !view.OpenLive.Known() || view.OpenLive.Value() != "1건" {
		t.Errorf("the plain count is %q; one list failing must not take the other down",
			view.OpenLive.Value())
	}
}

// TestAPlainReadThatFailedIsNeverAddedIntoTheTotalEither.
//
// The mirror. It is asserted separately because "the conditional half is the
// special one" is exactly the assumption that produces a total which is right in
// one direction only.
func TestAPlainReadThatFailedIsNeverAddedIntoTheTotalEither(t *testing.T) {
	reader := &countingOrders{lists: OrdersReading{
		OpenError:   "500 Internal Server Error",
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
		Open:   []OrderRecord{livePlainOrder("o-1", "005930")},
		Closed: []OrderRecord{filledPlainOrder("o-2", "AAPL")},
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
		Closed:      []OrderRecord{filledPlainOrder("o-2", "AAPL")},
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
		Closed:          []OrderRecord{filledPlainOrder("o-2", "AAPL")},
		ClosedTruncated: true,
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

	// The other direction: OPEN is documented to return every pending order, so
	// this cannot happen — but if a broker says it did, the screen reports what it
	// was told rather than the exactness it assumed. Believing the documentation
	// over the answer is how a leftover goes missing behind a confident number.
	contradicting := &countingOrders{lists: OrdersReading{
		Open:          []OrderRecord{livePlainOrder("o-1", "005930")},
		OpenTruncated: true,
	}}
	h2 := newOrdersHarness(t, contradicting)
	h2.authenticate(t)
	h2.get(t, "/orders")

	view := h2.Console.orders(context.Background(), orderFilterChoice{})
	if !view.Live.Known() || view.Live.Value() != "1건 이상" {
		t.Errorf("the live count is %q when the broker reported another page of pending orders; "+
			"the count is exact BECAUSE that answer is whole, so an answer that says otherwise "+
			"makes it a floor again", view.Live.Value())
	}
}

// TestTheLiveCountIsANumberEvenWhenTheClosedPageWasTruncated.
//
// The two plain groups behave differently and the difference is the whole reason
// there are two calls. status=OPEN returns every pending order — "limit, cursor 는
// 무시되며" — so the live count cannot be short. status=CLOSED paginates, so that
// count can be. Sharing one truncation flag between them turns an exact live
// answer into "N건 이상", which is the one thing an operator cannot resolve: they
// cannot tell whether the leftover they are looking for is in the part that was
// not sent.
func TestTheLiveCountIsANumberEvenWhenTheClosedPageWasTruncated(t *testing.T) {
	reader := &countingOrders{lists: OrdersReading{
		Open:            []OrderRecord{livePlainOrder("o-1", "005930")},
		Closed:          []OrderRecord{filledPlainOrder("o-2", "AAPL")},
		ClosedTruncated: true,
	}}
	h := newOrdersHarness(t, reader)
	h.authenticate(t)
	h.get(t, "/orders")

	view := h.Console.orders(context.Background(), orderFilterChoice{})
	if !view.Live.Known() || view.Live.Value() != "1건" {
		t.Errorf("the live count is %q, want exactly 1건; the closed page's truncation says nothing "+
			"about the pending group, which the broker returns whole", view.Live.Value())
	}
	if !view.OpenLive.Known() || view.OpenLive.Value() != "1건" {
		t.Errorf("the open count is %q, want 1건", view.OpenLive.Value())
	}
	if !view.ClosedCount.Known() || view.ClosedCount.Value() != "1건 이상" {
		t.Errorf("the closed count is %q, want 1건 이상 — that list really was truncated",
			view.ClosedCount.Value())
	}
}

// TestAnOrderInBothGroupsIsOneRowAndIsCountedOnce.
//
// openapi puts PARTIAL_FILLED in BOTH group definitions:
//
//	OPEN   ∈ {PENDING, PARTIAL_FILLED, PENDING_CANCEL, PENDING_REPLACE}
//	CLOSED ∈ {FILLED, CANCELED, REJECTED, REPLACED, …, PARTIAL_FILLED}
//
// so asking for both groups can return one order twice. Two rows for one order is
// one order counted twice in the live total and one row an operator tries to
// cancel twice; the pending copy is the live one and it wins.
func TestAnOrderInBothGroupsIsOneRowAndIsCountedOnce(t *testing.T) {
	both := livePlainOrder("o-both", "005930")
	both.Status = "PARTIAL_FILLED"
	both.FilledQuantity = "4"
	both.AverageFilledPrice = "70000"

	reader := &countingOrders{lists: OrdersReading{
		Open:   []OrderRecord{both},
		Closed: []OrderRecord{both},
	}}
	h := newOrdersHarness(t, reader)
	h.authenticate(t)
	h.get(t, "/orders")

	view := h.Console.orders(context.Background(), orderFilterChoice{})
	if view.Total != 1 {
		t.Errorf("an order returned in both groups produced %d rows; PARTIAL_FILLED belongs to both "+
			"by the API's own definition, and one order is one row", view.Total)
	}
	if !view.Live.Known() || view.Live.Value() != "1건" {
		t.Errorf("the live count is %q, want 1건", view.Live.Value())
	}
	if !view.ClosedCount.Known() || view.ClosedCount.Value() != "0건" {
		t.Errorf("the closed count is %q; the pending copy is the live one and it wins",
			view.ClosedCount.Value())
	}
	if len(view.Rows) == 1 && !view.Rows[0].Live {
		t.Error("the surviving row is not live; the broker just returned it in the pending group")
	}
}

// TestTheOriginColumnSaysUnknownWhenTheLedgerCouldNotBeRead.
//
// Reading an unreadable ledger as "somebody placed these by hand" makes a working
// engine look idle, and an idle engine is what an operator restarts.
func TestTheOriginColumnSaysUnknownWhenTheLedgerCouldNotBeRead(t *testing.T) {
	reader := &countingOrders{lists: OrdersReading{
		Open: []OrderRecord{livePlainOrder("o-1", "005930")},
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
		Open:   []OrderRecord{livePlainOrder("engine-order", "005930")},
		Closed: []OrderRecord{filledPlainOrder("hand-order", "AAPL")},
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

func TestInvalidEvidenceIdentityKeepsOriginUnknown(t *testing.T) {
	path := filepath.Join(t.TempDir(), "journal.db")
	seedEngineJournal(t, path, ordersJournalFixture)
	badTime := livePlainOrder("bad-time", "005930")
	badTime.OrderedAt = "not-a-time"
	badMarket := livePlainOrder("bad-market", "005930")
	badMarket.Market = "UNKNOWN"
	reader := &countingOrders{lists: OrdersReading{Open: []OrderRecord{badTime, badMarket}}}
	h := newOrdersHarness(t, reader, func(o *Options) { o.JournalPath = path })
	h.authenticate(t)
	view := h.Console.orders(context.Background(), orderFilterChoice{})
	if !view.Journal.Readable() {
		t.Fatalf("journal unexpectedly unreadable: %+v", view.Journal)
	}
	for _, row := range view.Rows {
		if row.Origin != originUnknown {
			t.Errorf("invalid identity row %s origin=%q, want unknown", row.ID, row.Origin)
		}
	}
}

func TestEvidenceQueryUsesOnlyRowsRemainingAfterFilter(t *testing.T) {
	path := filepath.Join(t.TempDir(), "journal.db")
	seedEngineJournal(t, path, ordersJournalFixture+`
INSERT INTO positions(id,account_ref,market,symbol,instance_seq,entry_decision_id,state,quantity,avg_price,opened_at)
VALUES ('filtered-position','123-45-678901','kr','005930',1,'d-filter','OPEN','1','70000','2026-07-27T00:20:00Z');
INSERT INTO exit_events(position_id,action,proposed_intent_id,created_at)
VALUES ('filtered-position','LEGACY','intent-1','2026-07-27T00:30:00Z');`)
	closed := make([]OrderRecord, journal.MaxBrokerOrderEvidenceScopes+1)
	for i := range closed {
		closed[i] = filledPlainOrder(fmt.Sprintf("hidden-%03d", i), "AAPL")
	}
	reader := &countingOrders{lists: OrdersReading{
		Open: []OrderRecord{livePlainOrder("engine-order", "005930")}, Closed: closed,
	}}
	h := newOrdersHarness(t, reader, func(o *Options) { o.JournalPath = path })
	h.authenticate(t)
	view := h.Console.orders(context.Background(), orderFilterChoice{State: filterStateLive})
	if !view.Journal.Readable() || len(view.Rows) != 1 || view.Rows[0].ID != "engine-order" ||
		view.Rows[0].Origin != originEngine || !view.Rows[0].ExitEvidence {
		t.Fatalf("filtered visible evidence = journal=%+v rows=%+v", view.Journal, view.Rows)
	}
	if view.Total != len(closed)+1 || view.ClosedCount.Value() != fmt.Sprintf("%d건", len(closed)) {
		t.Fatalf("filter changed counts: total=%d closed=%s", view.Total, view.ClosedCount.Value())
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
VALUES ('a-1','intent-1','PLACE','CONFIRMED',1,'engine-order','fp-1','2026-07-27T00:30:00Z');
`

// TestAWatchingConditionalIsNeverLabelledOtherByAJoinThatCannotSucceed.
//
// The first implementation joined a conditional row on rec.Triggered — the plain
// order the conditional turned into. That field is empty for exactly as long as
// the conditional is still watching, and the adapter asks for nothing but the
// watching ones, so the lookup was engineIDs[""] on every row this screen can
// show. BrokerOrderIDs excludes the empty id by construction, so the answer was
// always a miss and every conditional was labelled 그 밖 — a constant wearing a
// determination's clothes.
//
// The constant was also the wrong one. This screen's own words: "an invented
// 'manual' label on an engine order is an operator concluding the engine is idle
// while it is trading."
//
// The fixture is the one the review demonstrated with: a journal whose
// mutation_attempts really does carry the conditional's own id. Nothing in this
// build writes one there (conditional placement goes through
// trading.Service.ConditionalPlace and internal/verifylive, neither of which opens
// a journal attempt), which is exactly why a miss has to read as 불명 — but a hit
// is evidence, and the screen has to be able to report it the day one appears.
func TestAWatchingConditionalIsNeverLabelledOtherByAJoinThatCannotSucceed(t *testing.T) {
	path := filepath.Join(t.TempDir(), "journal.db")
	seedEngineJournal(t, path, ordersJournalFixture+conditionalJournalFixture)

	reader := &countingOrders{lists: OrdersReading{
		Open: []OrderRecord{livePlainOrder("engine-order", "005930")},
		Conditional: []ConditionalRecord{
			watchingConditional("co-engine", "005930"),
			watchingConditional("co-elsewhere", "AAPL"),
		},
	}}
	h := newOrdersHarness(t, reader, func(o *Options) { o.JournalPath = path })
	h.authenticate(t)
	h.get(t, "/orders")

	view := h.Console.orders(context.Background(), orderFilterChoice{})
	got := map[string]orderOrigin{}
	for _, r := range view.Rows {
		got[r.ID] = r.Origin
	}

	if got["co-engine"] != originEngine {
		t.Errorf("the conditional whose id IS in mutation_attempts is labelled %q, want %q — the "+
			"join is on an id that exists rather than on Triggered, which is empty on every row "+
			"this screen can show", got["co-engine"], originEngine)
	}
	if got["co-elsewhere"] != originUnknown {
		t.Errorf("a conditional the ledger does not know is labelled %q, want %q. The ledger was "+
			"never asked about conditionals — mutation_attempts records PLACE/CANCEL/AMEND of plain "+
			"orders — so its silence is not evidence that a person made this one",
			got["co-elsewhere"], originUnknown)
	}
	if got["engine-order"] != originEngine {
		t.Errorf("the plain engine order is labelled %q; the conditional join must not have "+
			"disturbed the plain one", got["engine-order"])
	}

	page := body(t, h.get(t, "/orders"))
	if !strings.Contains(page, "조건주문의 발주 주체") {
		t.Error("the page does not explain why a conditional's origin can be 불명 with a readable " +
			"ledger; an unexplained 불명 beside an explained 엔진 발주 reads as a bug")
	}
}

// conditionalJournalFixture puts a conditional order's own id in the ledger's
// attempt table. Nothing in this build writes one there today; the fixture exists
// so the join is tested against an id that can match rather than against the empty
// string, which never can.
const conditionalJournalFixture = `
INSERT INTO intents (id, created_at, market, trading_day, account_ref, symbol, side, order_type,
                     time_in_force, quantity, price, currency, source, fingerprint, notes)
VALUES ('intent-2','2026-07-27T00:20:00Z','kr','2026-07-27','123-45-678901','005930','SELL','LIMIT',
        'DAY','10','70000','KRW','engine','fp-2','');
INSERT INTO mutation_attempts (id, intent_id, kind, state, attempt_no, broker_order_id,
                               fingerprint, recorded_at)
VALUES ('a-2','intent-2','PLACE','CONFIRMED',1,'co-engine','fp-2','2026-07-27T00:20:00Z');
`

// TestASideFilterNeverHidesAWatchingConditional.
//
// A conditional order's payload carries no side at all. If a side filter excluded
// the rows it cannot judge, one click on 매수 would empty every live conditional
// off a screen whose entire job is to show that they are still there — and the
// screen would then read as "no buy orders" rather than as "this filter has
// nothing to say about two of these rows".
//
// It is pinned in both directions because "not applicable" and "matches BUY" look
// identical from one of them.
func TestASideFilterNeverHidesAWatchingConditional(t *testing.T) {
	reader := &countingOrders{lists: OrdersReading{
		Open:        []OrderRecord{livePlainOrder("o-buy", "005930")},
		Conditional: []ConditionalRecord{watchingConditional("co-1", "005930")},
	}}
	h := newOrdersHarness(t, reader)
	h.authenticate(t)

	for _, side := range []string{"BUY", "SELL"} {
		page := body(t, h.get(t, "/orders?side="+side))
		if !strings.Contains(page, "co-1") {
			t.Errorf("?side=%s hid the watching conditional order. A payload with no direction that "+
				"falls out of a direction filter is a live leftover hiding behind the filter — the "+
				"same failure as counting it as zero, one interaction later", side)
		}
		view := h.Console.orders(context.Background(),
			filterChoiceFromQuery(t, "/orders?side="+side))
		found := false
		for _, r := range view.Rows {
			if r.ID == "co-1" {
				found = true
			}
		}
		if !found {
			t.Errorf("?side=%s excluded the conditional row from the view", side)
		}
	}

	// And the side filter still narrows the rows it CAN judge.
	sell := h.Console.orders(context.Background(), filterChoiceFromQuery(t, "/orders?side=SELL"))
	for _, r := range sell.Rows {
		if r.ID == "o-buy" {
			t.Error("?side=SELL kept a BUY order; the axis has an answer for plain orders and must " +
				"use it")
		}
	}
}

// TestAMarketFilterNeverHidesAnOrderWhoseMarketIsUnknown.
//
// The same shape as the side axis, one axis over. The plain order endpoint carries
// no market field — it is derived from the currency — so an order in a currency
// that is neither KRW nor USD arrives with no market and renders as "—". Excluding
// it from BOTH market filters means a live order that is on the unfiltered screen
// vanishes from every filtered one, and vanishing is the thing this screen exists
// to stop.
func TestAMarketFilterNeverHidesAnOrderWhoseMarketIsUnknown(t *testing.T) {
	odd := livePlainOrder("o-odd-market", "7203")
	odd.Market, odd.Currency = "", "JPY"

	reader := &countingOrders{lists: OrdersReading{
		Open: []OrderRecord{livePlainOrder("o-kr", "005930"), odd},
	}}
	h := newOrdersHarness(t, reader)
	h.authenticate(t)

	for _, market := range []string{"KR", "US"} {
		page := body(t, h.get(t, "/orders?market="+market))
		if !strings.Contains(page, "o-odd-market") {
			t.Errorf("?market=%s hid a live order whose market this build could not name. An "+
				"unnameable market is not-applicable to the filter, exactly as a missing side is on "+
				"the direction axis — not excluded by it", market)
		}
	}

	// A market this build CAN name is still filtered.
	us := h.Console.orders(context.Background(), filterChoiceFromQuery(t, "/orders?market=US"))
	for _, r := range us.Rows {
		if r.ID == "o-kr" {
			t.Error("?market=US kept a KR order; the axis has an answer for that row and must use it")
		}
	}
}

// TestAFilteredTableShowsTheWholeCountBesideTheFilteredOne.
//
// A filtered screen with only its own count on it reads as "these are all the
// orders", and the row that is not on the screen is the leftover.
func TestAFilteredTableShowsTheWholeCountBesideTheFilteredOne(t *testing.T) {
	reader := &countingOrders{lists: OrdersReading{
		Open:   []OrderRecord{livePlainOrder("o-1", "005930")},
		Closed: []OrderRecord{filledPlainOrder("o-2", "AAPL")},
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
		Open:        []OrderRecord{livePlainOrder("o-1", "005930")},
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
	reader := &countingOrders{lists: OrdersReading{Open: []OrderRecord{odd}}}
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
	reader := &countingOrders{lists: OrdersReading{Open: []OrderRecord{odd}}}
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
		Open:        []OrderRecord{livePlainOrder("o-1", "005930")},
		Closed:      []OrderRecord{filledPlainOrder("o-2", "AAPL")},
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

// TestTheOverviewNeverRendersAStaleOrdersReadingAsAMeasuredNumber.
//
// Design D9. The overview makes no broker call by contract (overview D4), so this
// count freezes at whatever the last /orders visit produced — for as long as
// nobody opens that screen again:
//
//	cache filled while the account was empty → the engine places an order
//	→ nobody opens /orders → three hours later /dashboard says 미체결 건수 0건
//	  with no age and no marker, and /orders says 1건.
//
// The holdings panel three sections up has printed its cache instant and age since
// the day it shipped. D4 handled the COLD cache (never_fetched) and left the OLD
// one measured; empty and stale are different failures, and rendered as the same
// "0" the difference is gone.
func TestTheOverviewNeverRendersAStaleOrdersReadingAsAMeasuredNumber(t *testing.T) {
	reader := &countingOrders{lists: OrdersReading{}}
	h := newOrdersHarness(t, reader)
	h.authenticate(t)

	// Fill the cache while the account is empty, exactly as an operator would.
	h.get(t, "/orders")
	fresh := h.Console.overview(context.Background())
	if !fresh.Open.Count.Known() || fresh.Open.Count.Value() != "0건" {
		t.Fatalf("a reading taken this instant is %q, want 0건", fresh.Open.Count.Value())
	}
	if fresh.Open.Stale {
		t.Error("a reading taken this instant is marked stale")
	}

	// Three hours pass and nothing opens /orders. The overview refreshes nothing,
	// so the number below is about an account that stopped existing three hours
	// ago.
	h.clock.advance(3 * time.Hour)
	old := h.Console.overview(context.Background())
	if reader.count() != 1 {
		t.Fatalf("the overview refreshed the cache (%d calls); its contract is zero broker calls",
			reader.count())
	}
	if !old.Open.Stale {
		t.Error("a three-hour-old reading is not marked stale; past the TTL a reading is not a " +
			"measurement of anything")
	}
	if !old.Open.Present || old.Open.AgeSeconds != 3*60*60 {
		t.Errorf("the panel carries age %ds present=%v; the value has to travel with the instant it "+
			"was taken, exactly as the holdings panel already does",
			old.Open.AgeSeconds, old.Open.Present)
	}
	if old.Open.TakenAt != h.clock.Now().Add(-3*time.Hour).UTC().Format("2006-01-02 15:04:05Z") {
		t.Errorf("the panel's instant is %q, want the moment the cache was filled", old.Open.TakenAt)
	}

	page := body(t, h.get(t, "/dashboard"))
	if !strings.Contains(page, old.Open.TakenAt) {
		t.Error("the overview renders the count without the instant it was read at; a bold number " +
			"with its provenance in another paragraph is not read as one fact")
	}
	if !strings.Contains(page, "TTL") {
		t.Error("the overview does not say the reading is past the cache TTL")
	}
}

// TestTheOrdersCountsCarryTheirOwnProvenanceInTheSameBreath.
//
// P1-3, the same rule on the screen that owns the number. The counts section used
// to print a bold 합계 and leave the cache instant, the age, the withheld refresh
// and the last failure to a separate paragraph below it. Two paragraphs are two
// facts; nobody joins them, and the one that gets read is the bold one.
func TestTheOrdersCountsCarryTheirOwnProvenanceInTheSameBreath(t *testing.T) {
	reader := &countingOrders{lists: OrdersReading{
		Open: []OrderRecord{livePlainOrder("o-1", "005930")},
	}}
	h := newOrdersHarness(t, reader)
	h.authenticate(t)

	page := body(t, h.get(t, "/orders"))
	counts := sectionAround(t, page, "미체결")
	view := h.Console.orders(context.Background(), orderFilterChoice{})

	// The instant is beside the number itself, not merely in the same section:
	// the count and when it was true are one statement.
	total := strings.Index(counts, "<strong>"+view.Live.Value()+"</strong>")
	if total < 0 {
		t.Fatalf("the counts section has no 합계 of %q: %s", view.Live.Value(), counts)
	}
	sameCell := counts[total:]
	if end := strings.Index(sameCell, "</dd>"); end >= 0 {
		sameCell = sameCell[:end]
	}
	for _, want := range []string{view.Broker.TakenAt(), "초 전"} {
		if !strings.Contains(sameCell, want) {
			t.Errorf("the 합계 cell does not carry %q. A bold total with its age in a different "+
				"paragraph reads as a current measurement, and the paragraph is what gets skipped",
				want)
		}
	}

	// And the last failure is in the same section as the counts it qualifies. It
	// is a separate assertion because a reading can be present AND stale for a
	// reason: the refresh that would have replaced it did not answer.
	reader.mu.Lock()
	reader.err = errors.New("429 Too Many Requests")
	reader.mu.Unlock()
	h.clock.advance(ordersTTL)

	failed := sectionAround(t, body(t, h.get(t, "/orders")), "미체결")
	if !strings.Contains(failed, "429 Too Many Requests") {
		t.Error("the failed refresh is not in the section that carries the counts; the numbers " +
			"above it are then the ones before the failure, said with no qualification")
	}
}

// sectionAround returns the <section> containing the first occurrence of marker,
// so a test can assert that two facts are in the same one rather than merely on
// the same page.
func sectionAround(t *testing.T, page, marker string) string {
	t.Helper()
	at := strings.Index(page, marker)
	if at < 0 {
		t.Fatalf("the page has no %q", marker)
	}
	start := strings.LastIndex(page[:at], "<section")
	if start < 0 {
		t.Fatalf("%q is not inside a <section>", marker)
	}
	end := strings.Index(page[start:], "</section>")
	if end < 0 {
		t.Fatalf("the section holding %q is not closed", marker)
	}
	return page[start : start+end]
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
