package filldetect_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/domain"
	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/filldetect"
	"github.com/JungHoonGhae/tossinvest-cli/internal/official"
)

// Fill-detection tests (harden-execution-base task 3.1).
//
// The properties under test are the fill-detection spec's "폴링이 체결 감지의 권위":
// the open list is walked to its last page, an order that left that list is still
// read by id, the account sweep happens every cycle, the SLO is measured from the
// instant the broker made a fill observable to the instant it was durably
// committed locally, and a sustained violation blocks new entries while a recovery
// releases them again.

var pollStart = time.Date(2026, 3, 30, 1, 30, 0, 0, time.UTC) // 10:30 KST

// --- fakes ------------------------------------------------------------------

// fakePager is an execgw.OrderPager whose pages are scripted per cursor, so a
// test can prove the walk continues past page one.
type fakePager struct {
	mu    sync.Mutex
	pages map[string]execgw.OrderPage
	calls []string
	err   error
}

func newPager(pages ...execgw.OrderPage) *fakePager {
	p := &fakePager{pages: map[string]execgw.OrderPage{}}
	cursor := ""
	for i, page := range pages {
		p.pages[cursor] = page
		if i < len(pages)-1 {
			cursor = page.NextCursor
		}
	}
	return p
}

func (p *fakePager) OrdersPageRaw(_ context.Context, q execgw.OrderQuery, cursor string) (execgw.OrderPage, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.calls = append(p.calls, q.Status+"/"+cursor)
	if p.err != nil {
		return execgw.OrderPage{}, p.err
	}
	page, ok := p.pages[cursor]
	if !ok {
		return execgw.OrderPage{}, nil
	}
	return page, nil
}

func (p *fakePager) callCount() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return len(p.calls)
}

// fakeOrderReader answers single-order reads by id.
type fakeOrderReader struct {
	mu     sync.Mutex
	orders map[string]json.RawMessage
	reads  []string
	err    error
}

func (r *fakeOrderReader) OrderRaw(_ context.Context, orderID string) (json.RawMessage, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.reads = append(r.reads, orderID)
	if r.err != nil {
		return nil, r.err
	}
	raw, ok := r.orders[orderID]
	if !ok {
		return nil, fmt.Errorf("no such order %s", orderID)
	}
	return raw, nil
}

func (r *fakeOrderReader) readIDs() []string {
	r.mu.Lock()
	defer r.mu.Unlock()
	return append([]string(nil), r.reads...)
}

// fakePositions is the account sweep. It doubles as the buying-power reader so a
// test can prove the slower cash cadence.
type fakePositions struct {
	mu        sync.Mutex
	positions []domain.Position
	sweeps    int
	balances  int
	err       error
}

func (p *fakePositions) BuyingPower(_ context.Context, _ string) (float64, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.balances++
	if p.err != nil {
		return 0, p.err
	}
	return 1000, nil
}

func (p *fakePositions) balanceCount() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.balances
}

func (p *fakePositions) Positions(context.Context) ([]domain.Position, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.sweeps++
	if p.err != nil {
		return nil, p.err
	}
	return p.positions, nil
}

func (p *fakePositions) sweepCount() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.sweeps
}

// fakeTracked names the orders that must be followed by id even once they have
// left the open list.
type fakeTracked struct {
	orders []filldetect.TrackedOrder
	err    error
}

func (t fakeTracked) TrackedOrders(context.Context) ([]filldetect.TrackedOrder, error) {
	if t.err != nil {
		return nil, t.err
	}
	return t.orders, nil
}

// fakeLedger is a cumulative-snapshot ledger with the same positive-delta rule as
// the real one, so the detector tests do not depend on the journal.
type fakeLedger struct {
	mu       sync.Mutex
	clk      clock.Clock
	seen     map[string]float64
	applied  []filldetect.Snapshot
	failOnce error
}

func newLedger(clk clock.Clock) *fakeLedger {
	return &fakeLedger{clk: clk, seen: map[string]float64{}}
}

func (l *fakeLedger) Apply(_ context.Context, snap filldetect.Snapshot) (filldetect.Applied, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.failOnce != nil {
		err := l.failOnce
		l.failOnce = nil
		return filldetect.Applied{}, err
	}
	l.applied = append(l.applied, snap)
	prev := l.seen[snap.OrderID]
	delta := snap.FilledQuantity - prev
	if delta < 0 {
		return filldetect.Applied{
			OrderID: snap.OrderID, FailClosed: true,
			Reason: execgw.ReasonBrokerStateUnknown,
		}, nil
	}
	l.seen[snap.OrderID] = snap.FilledQuantity
	return filldetect.Applied{
		OrderID:     snap.OrderID,
		Delta:       delta,
		Changed:     delta > 0,
		CommittedAt: l.clk.Now(),
	}, nil
}

func (l *fakeLedger) appliedIDs() []string {
	l.mu.Lock()
	defer l.mu.Unlock()
	out := make([]string, 0, len(l.applied))
	for _, s := range l.applied {
		out = append(out, s.OrderID)
	}
	return out
}

// --- payload helpers --------------------------------------------------------

type rawOrder struct {
	id       string
	symbol   string
	status   string
	quantity string
	filled   string
	avgPrice string
	filledAt string
	canceled string
}

func (o rawOrder) json() json.RawMessage {
	if o.quantity == "" {
		o.quantity = "10"
	}
	if o.status == "" {
		o.status = "OPEN"
	}
	if o.symbol == "" {
		o.symbol = "AAPL"
	}
	fields := []string{
		`"orderId":` + quote(o.id),
		`"symbol":` + quote(o.symbol),
		`"side":"BUY"`,
		`"status":` + quote(o.status),
		`"quantity":` + quote(o.quantity),
		`"currency":"USD"`,
	}
	if o.canceled != "" {
		fields = append(fields, `"canceledAt":`+quote(o.canceled))
	}
	exec := []string{`"filledQuantity":` + quote(orZero(o.filled))}
	if o.avgPrice != "" {
		exec = append(exec, `"averageFilledPrice":`+quote(o.avgPrice))
	}
	if o.filledAt != "" {
		exec = append(exec, `"filledAt":`+quote(o.filledAt))
	}
	fields = append(fields, `"execution":{`+strings.Join(exec, ",")+`}`)
	return json.RawMessage(`{` + strings.Join(fields, ",") + `}`)
}

func quote(s string) string { return `"` + s + `"` }

func orZero(s string) string {
	if s == "" {
		return "0"
	}
	return s
}

func page(next string, orders ...rawOrder) execgw.OrderPage {
	p := execgw.OrderPage{NextCursor: next, HasNext: next != ""}
	for _, o := range orders {
		p.Orders = append(p.Orders, o.json())
	}
	return p
}

// newDetector wires a detector over the supplied fakes.
func newDetector(t *testing.T, pager *fakePager, reader *fakeOrderReader, positions *fakePositions,
	tracked fakeTracked, gate *execgw.EntryGate) (*filldetect.Detector, *clock.Fake, *fakeLedger) {
	t.Helper()
	clk := clock.NewFake(pollStart)
	ledger := newLedger(clk)
	d := &filldetect.Detector{
		Orders:    pager,
		Order:     reader,
		Positions: positions,
		Tracked:   tracked,
		Ledger:    ledger,
		Clock:     clk,
		Gate:      gate,
	}
	return d, clk, ledger
}

// --- tests ------------------------------------------------------------------

// TestPollWalksTheOpenListToTheLastPage is the pagination-completion requirement:
// an order sitting on page two that is never read is a live position the engine
// does not know it has.
func TestPollWalksTheOpenListToTheLastPage(t *testing.T) {
	pager := newPager(
		page("cursor-2", rawOrder{id: "o-1", filled: "0"}),
		page("", rawOrder{id: "o-2", filled: "3"}),
	)
	d, _, ledger := newDetector(t, pager, &fakeOrderReader{}, &fakePositions{}, fakeTracked{}, nil)

	cycle, err := d.PollOnce(context.Background())
	if err != nil {
		t.Fatalf("PollOnce: %v", err)
	}
	if pager.callCount() != 2 {
		t.Fatalf("pager calls = %d, want 2 (the walk must continue past page one)", pager.callCount())
	}
	if cycle.OpenOrders != 2 {
		t.Fatalf("OpenOrders = %d, want 2", cycle.OpenOrders)
	}
	if got := ledger.appliedIDs(); len(got) != 2 || got[0] != "o-1" || got[1] != "o-2" {
		t.Fatalf("applied snapshots = %v, want both pages", got)
	}
}

// TestPollReadsTrackedOrdersThatLeftTheOpenList covers the fill that closes an
// order: it disappears from the open list, and only OrderByID can say it filled
// rather than being cancelled.
func TestPollReadsTrackedOrdersThatLeftTheOpenList(t *testing.T) {
	reader := &fakeOrderReader{orders: map[string]json.RawMessage{
		"o-gone": rawOrder{id: "o-gone", status: "CLOSED", quantity: "10", filled: "10"}.json(),
	}}
	tracked := fakeTracked{orders: []filldetect.TrackedOrder{
		{OrderID: "o-gone", Symbol: "AAPL", Market: "us"},
		{OrderID: "o-open", Symbol: "AAPL", Market: "us"},
	}}
	pager := newPager(page("", rawOrder{id: "o-open", filled: "1"}))

	d, _, ledger := newDetector(t, pager, reader, &fakePositions{}, tracked, nil)
	cycle, err := d.PollOnce(context.Background())
	if err != nil {
		t.Fatalf("PollOnce: %v", err)
	}

	if got := reader.readIDs(); len(got) != 1 || got[0] != "o-gone" {
		t.Fatalf("single-order reads = %v, want only the order missing from the open list", got)
	}
	if cycle.TrackedReads != 1 {
		t.Fatalf("TrackedReads = %d, want 1", cycle.TrackedReads)
	}
	ids := ledger.appliedIDs()
	if len(ids) != 2 {
		t.Fatalf("applied = %v, want the open order and the tracked one", ids)
	}
}

// TestPollSweepsTheAccountEveryCycle pins the third leg of the authority triple.
func TestPollSweepsTheAccountEveryCycle(t *testing.T) {
	positions := &fakePositions{positions: []domain.Position{{Symbol: "AAPL", Quantity: 10}}}
	gate := execgw.NewEntryGate(clock.NewFake(pollStart), map[execgw.RequiredQuery]time.Duration{})
	d, _, _ := newDetector(t, newPager(page("")), &fakeOrderReader{}, positions, fakeTracked{}, gate)

	cycle, err := d.PollOnce(context.Background())
	if err != nil {
		t.Fatalf("PollOnce: %v", err)
	}
	if positions.sweepCount() != 1 {
		t.Fatalf("account sweeps = %d, want 1", positions.sweepCount())
	}
	if cycle.Positions != 1 {
		t.Fatalf("Positions = %d, want 1", cycle.Positions)
	}
}

// TestPollStampsQueryFreshness proves a successful cycle is what clears a
// staleness block: the detector is the thing that keeps the required reads fresh.
func TestPollStampsQueryFreshness(t *testing.T) {
	clk := clock.NewFake(pollStart)
	gate := execgw.NewEntryGate(clk, execgw.DefaultStaleness())
	account := &fakePositions{}
	d := &filldetect.Detector{
		Orders:    newPager(page("")),
		Order:     &fakeOrderReader{},
		Positions: account,
		Balance:   account,
		Tracked:   fakeTracked{},
		Ledger:    newLedger(clk),
		Clock:     clk,
		Gate:      gate,
		Config:    filldetect.Config{Currencies: []string{"USD"}},
	}

	if rejected := gate.CheckEntry(); rejected == nil {
		t.Fatal("a gate that has never seen a successful poll must block entries")
	}
	if _, err := d.PollOnce(context.Background()); err != nil {
		t.Fatalf("PollOnce: %v", err)
	}
	// price is not a fill-detection read, so it is still unobserved; the three the
	// detector owns must now be fresh.
	rejected := gate.CheckEntry()
	if rejected == nil || !strings.Contains(rejected.Detail, string(execgw.QueryPrice)) {
		t.Fatalf("after a poll only the price read should still be stale, got %v", rejected)
	}
}

// TestBuyingPowerUsesItsOwnSlowerCadence keeps the cash read inside the rate-limit
// budget the retry matrix accounts for (§0.4): cash moves on fills and transfers,
// not on every 3-second look at the order list.
func TestBuyingPowerUsesItsOwnSlowerCadence(t *testing.T) {
	clk := clock.NewFake(pollStart)
	account := &fakePositions{}
	d := &filldetect.Detector{
		Orders: newPager(page("")), Order: &fakeOrderReader{},
		Positions: account, Balance: account, Tracked: fakeTracked{},
		Ledger: newLedger(clk), Clock: clk,
		Config: filldetect.Config{
			PollInterval: 3 * time.Second, BalanceInterval: 15 * time.Second,
			Currencies: []string{"USD"},
		},
	}

	for i := 0; i < 4; i++ {
		cycle, err := d.PollOnce(context.Background())
		if err != nil {
			t.Fatalf("poll %d: %v", i, err)
		}
		if want := i == 0; cycle.ReadBuyingPower != want {
			t.Fatalf("poll %d ReadBuyingPower = %v, want %v", i, cycle.ReadBuyingPower, want)
		}
		clk.Advance(3 * time.Second)
	}
	if account.balanceCount() != 1 {
		t.Fatalf("buying-power reads = %d, want 1 inside 12s of polling", account.balanceCount())
	}

	clk.Advance(4 * time.Second) // 16s since the last read
	cycle, err := d.PollOnce(context.Background())
	if err != nil {
		t.Fatalf("poll after the balance interval: %v", err)
	}
	if !cycle.ReadBuyingPower {
		t.Fatal("the balance interval elapsed; buying power should have been refreshed")
	}
}

// TestSLOMeasuresBrokerVisibleToLocalCommit is the measurement-point definition:
// the clock starts when the broker made the fill observable (execution.filledAt),
// not when the poller happened to look.
func TestSLOMeasuresBrokerVisibleToLocalCommit(t *testing.T) {
	filledAt := pollStart.Add(-4 * time.Second)
	pager := newPager(page("", rawOrder{
		id: "o-1", quantity: "10", filled: "4",
		filledAt: filledAt.Format(time.RFC3339),
	}))
	d, _, _ := newDetector(t, pager, &fakeOrderReader{}, &fakePositions{}, fakeTracked{}, nil)

	cycle, err := d.PollOnce(context.Background())
	if err != nil {
		t.Fatalf("PollOnce: %v", err)
	}
	if len(cycle.Latencies) != 1 {
		t.Fatalf("latency samples = %v, want one per positive delta", cycle.Latencies)
	}
	if want := 4 * time.Second; cycle.Latencies[0] != want {
		t.Fatalf("latency = %s, want %s (filledAt → durable commit)", cycle.Latencies[0], want)
	}
}

// TestSLOFallsBackToThePreviousPollWhenTheBrokerGivesNoFillTime keeps the
// measurement conservative: with no broker timestamp the fill could have been
// visible any time since the previous look, so the oldest possible instant is
// used and the measured latency is an upper bound.
func TestSLOFallsBackToThePreviousPollWhenTheBrokerGivesNoFillTime(t *testing.T) {
	pager := newPager(page("", rawOrder{id: "o-1", quantity: "10", filled: "0"}))
	d, clk, _ := newDetector(t, pager, &fakeOrderReader{}, &fakePositions{}, fakeTracked{}, nil)

	if _, err := d.PollOnce(context.Background()); err != nil {
		t.Fatalf("first PollOnce: %v", err)
	}
	clk.Advance(6 * time.Second)
	pager.mu.Lock()
	pager.pages[""] = page("", rawOrder{id: "o-1", quantity: "10", filled: "4"})
	pager.mu.Unlock()

	cycle, err := d.PollOnce(context.Background())
	if err != nil {
		t.Fatalf("second PollOnce: %v", err)
	}
	if len(cycle.Latencies) != 1 {
		t.Fatalf("latency samples = %v, want 1", cycle.Latencies)
	}
	if want := 6 * time.Second; cycle.Latencies[0] != want {
		t.Fatalf("latency = %s, want %s (previous poll → this commit)", cycle.Latencies[0], want)
	}
}

// TestSustainedSLOViolationBlocksEntriesAndRecoveryReleasesThem is the spec's
// "SLO 위반 지속" scenario, end to end on a fake clock.
func TestSustainedSLOViolationBlocksEntriesAndRecoveryReleasesThem(t *testing.T) {
	clk := clock.NewFake(pollStart)
	gate := execgw.NewEntryGate(clk, map[execgw.RequiredQuery]time.Duration{})
	pager := newPager(page(""))
	d := &filldetect.Detector{
		Orders: pager, Order: &fakeOrderReader{}, Positions: &fakePositions{},
		Tracked: fakeTracked{}, Ledger: newLedger(clk), Clock: clk, Gate: gate,
		Config: filldetect.Config{SLO: filldetect.SLO{
			Target: 5 * time.Second, Percentile: 0.95, Window: time.Minute,
			MinSamples: 2, Grace: 30 * time.Second,
		}},
	}

	// Two slow fills: the p95 is over target, but the violation is not yet
	// sustained, so entries stay open.
	for i, id := range []string{"o-1", "o-2"} {
		pager.mu.Lock()
		pager.pages[""] = page("", rawOrder{
			id: id, quantity: "10", filled: "5",
			filledAt: clk.Now().Add(-20 * time.Second).Format(time.RFC3339),
		})
		pager.mu.Unlock()
		if _, err := d.PollOnce(context.Background()); err != nil {
			t.Fatalf("poll %d: %v", i, err)
		}
		clk.Advance(time.Second)
	}
	if rejected := gate.CheckEntry(); rejected != nil {
		t.Fatalf("a violation shorter than the grace period must not block yet: %v", rejected)
	}
	if !d.Health().SLO.Violated {
		t.Fatal("the SLO should already read as violated before the grace period expires")
	}

	// Past the grace period the block latches.
	clk.Advance(31 * time.Second)
	pager.mu.Lock()
	pager.pages[""] = page("", rawOrder{
		id: "o-3", quantity: "10", filled: "5",
		filledAt: clk.Now().Add(-20 * time.Second).Format(time.RFC3339),
	})
	pager.mu.Unlock()
	if _, err := d.PollOnce(context.Background()); err != nil {
		t.Fatalf("poll after grace: %v", err)
	}
	rejected := gate.CheckEntry()
	if rejected == nil || rejected.Reason != execgw.ReasonFillDetectionSLO {
		t.Fatalf("sustained violation must block new entries, got %v", rejected)
	}
	if !d.Health().EntryBlocked {
		t.Fatal("Health must report the block")
	}

	// Recovery: the window ages out and fast fills take over. The block clears
	// itself — no operator action, unlike an auth latch.
	clk.Advance(2 * time.Minute)
	for i, id := range []string{"o-4", "o-5"} {
		pager.mu.Lock()
		pager.pages[""] = page("", rawOrder{
			id: id, quantity: "10", filled: "5",
			filledAt: clk.Now().Format(time.RFC3339),
		})
		pager.mu.Unlock()
		if _, err := d.PollOnce(context.Background()); err != nil {
			t.Fatalf("recovery poll %d: %v", i, err)
		}
		clk.Advance(time.Second)
	}
	if rejected := gate.CheckEntry(); rejected != nil {
		t.Fatalf("recovery must release the SLO block automatically, got %v", rejected)
	}
	if d.Health().EntryBlocked {
		t.Fatal("Health should no longer report a block after recovery")
	}
}

// TestRateLimitedPollIsClassifiedAsOutage separates "the broker is throttling us"
// from "our pipeline is slow": only the second is an SLO violation, and mixing
// them would make the SLO fire for something no amount of local work can fix.
func TestRateLimitedPollIsClassifiedAsOutage(t *testing.T) {
	clk := clock.NewFake(pollStart)
	pager := newPager(page(""))
	pager.err = official.ErrRateLimited
	d := &filldetect.Detector{
		Orders: pager, Order: &fakeOrderReader{}, Positions: &fakePositions{},
		Tracked: fakeTracked{}, Ledger: newLedger(clk), Clock: clk,
	}

	if _, err := d.PollOnce(context.Background()); err == nil {
		t.Fatal("a throttled poll must report an error")
	}
	health := d.Health()
	if !health.Outage.Active {
		t.Fatal("a throttled poll must be classified as an outage")
	}
	if health.Outage.Class != execgw.ClassRateLimited {
		t.Fatalf("outage class = %s, want %s", health.Outage.Class, execgw.ClassRateLimited)
	}
	if health.Outage.Since != pollStart {
		t.Fatalf("outage start = %s, want %s", health.Outage.Since, pollStart)
	}
	if health.SLO.Violated {
		t.Fatal("an outage is not an SLO violation")
	}

	// Recovery clears the outage.
	clk.Advance(10 * time.Second)
	pager.mu.Lock()
	pager.err = nil
	pager.mu.Unlock()
	if _, err := d.PollOnce(context.Background()); err != nil {
		t.Fatalf("recovered poll: %v", err)
	}
	if d.Health().Outage.Active {
		t.Fatal("a successful poll must clear the outage")
	}
}

// TestTransportOutageIsClassifiedSeparately keeps the two failure families
// distinguishable in the health report an operator reads.
func TestTransportOutageIsClassifiedSeparately(t *testing.T) {
	clk := clock.NewFake(pollStart)
	pager := newPager(page(""))
	pager.err = fmt.Errorf("dialing: %w", official.ErrTransport)
	d := &filldetect.Detector{
		Orders: pager, Order: &fakeOrderReader{}, Positions: &fakePositions{},
		Tracked: fakeTracked{}, Ledger: newLedger(clk), Clock: clk,
	}
	if _, err := d.PollOnce(context.Background()); err == nil {
		t.Fatal("a broken transport must report an error")
	}
	health := d.Health()
	if health.Outage.Class != execgw.ClassTransient {
		t.Fatalf("outage class = %s, want %s", health.Outage.Class, execgw.ClassTransient)
	}
	if health.Outage.Consecutive != 1 {
		t.Fatalf("consecutive failures = %d, want 1", health.Outage.Consecutive)
	}
}

// TestOutageDoesNotStampFreshness is the fail-closed half of outage handling: a
// failed poll must leave the required reads stale so entries block on their own.
func TestOutageDoesNotStampFreshness(t *testing.T) {
	clk := clock.NewFake(pollStart)
	gate := execgw.NewEntryGate(clk, map[execgw.RequiredQuery]time.Duration{
		execgw.QueryOpenOrders: 20 * time.Second,
	})
	pager := newPager(page(""))
	d := &filldetect.Detector{
		Orders: pager, Order: &fakeOrderReader{}, Positions: &fakePositions{},
		Tracked: fakeTracked{}, Ledger: newLedger(clk), Clock: clk, Gate: gate,
	}
	if _, err := d.PollOnce(context.Background()); err != nil {
		t.Fatalf("first poll: %v", err)
	}
	if rejected := gate.CheckEntry(); rejected != nil {
		t.Fatalf("a fresh read must not block: %v", rejected)
	}

	pager.mu.Lock()
	pager.err = official.ErrRateLimited
	pager.mu.Unlock()
	clk.Advance(25 * time.Second)
	if _, err := d.PollOnce(context.Background()); err == nil {
		t.Fatal("expected the throttled poll to fail")
	}
	if rejected := gate.CheckEntry(); rejected == nil || rejected.Reason != execgw.ReasonQueryStale {
		t.Fatalf("an outage must leave the read stale, got %v", rejected)
	}
}

// TestMalformedOrderPayloadFailsTheCycle refuses to treat an unreadable record as
// "nothing there": a payload we cannot parse is an unknown, and an unknown must
// not be reported as an empty open list.
func TestMalformedOrderPayloadFailsTheCycle(t *testing.T) {
	clk := clock.NewFake(pollStart)
	pager := newPager(execgw.OrderPage{Orders: []json.RawMessage{json.RawMessage(`{"orderId":`)}})
	d := &filldetect.Detector{
		Orders: pager, Order: &fakeOrderReader{}, Positions: &fakePositions{},
		Tracked: fakeTracked{}, Ledger: newLedger(clk), Clock: clk,
	}
	if _, err := d.PollOnce(context.Background()); err == nil {
		t.Fatal("an unparseable order must fail the cycle rather than be skipped")
	}
}

// TestDetectionNeedsNoWTSSession is the spec's "WTS 세션 만료 중 체결 발생"
// scenario expressed structurally: every source the detector declares is an
// official-API read, so nothing about it can depend on a browser session.
func TestDetectionNeedsNoWTSSession(t *testing.T) {
	clk := clock.NewFake(pollStart)
	filled := rawOrder{id: "o-1", status: "CLOSED", quantity: "4", filled: "4",
		filledAt: pollStart.Add(-2 * time.Second).Format(time.RFC3339)}
	d := &filldetect.Detector{
		Orders:    newPager(page("")),
		Order:     &fakeOrderReader{orders: map[string]json.RawMessage{"o-1": filled.json()}},
		Positions: &fakePositions{},
		Tracked:   fakeTracked{orders: []filldetect.TrackedOrder{{OrderID: "o-1", Symbol: "AAPL"}}},
		Ledger:    newLedger(clk),
		Clock:     clk,
	}
	cycle, err := d.PollOnce(context.Background())
	if err != nil {
		t.Fatalf("PollOnce: %v", err)
	}
	if cycle.Fills != 1 {
		t.Fatalf("fills = %d, want 1 detected with no session in sight", cycle.Fills)
	}
}

// TestRunStopsOnContextCancellation and keeps polling across a transient failure:
// an outage is a reason to keep trying, not a reason to give up on the account.
func TestRunKeepsPollingThroughAnOutage(t *testing.T) {
	clk := clock.NewFake(pollStart)
	pager := newPager(page(""))
	pager.err = official.ErrRateLimited
	d := &filldetect.Detector{
		Orders: pager, Order: &fakeOrderReader{}, Positions: &fakePositions{},
		Tracked: fakeTracked{}, Ledger: newLedger(clk), Clock: clk,
		Config: filldetect.Config{PollInterval: 3 * time.Second},
	}

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- d.Run(ctx) }()

	if !clk.WaitForSleepers(1, 2*time.Second) {
		t.Fatal("Run should be waiting out its poll interval")
	}
	pager.mu.Lock()
	pager.err = nil
	pager.mu.Unlock()
	clk.Advance(3 * time.Second)

	deadline := time.Now().Add(2 * time.Second)
	for d.Health().Outage.Active {
		if time.Now().After(deadline) {
			t.Fatal("Run did not poll again after the outage cleared")
		}
		time.Sleep(time.Millisecond)
	}

	cancel()
	clk.Advance(3 * time.Second)
	if err := <-done; !errors.Is(err, context.Canceled) {
		t.Fatalf("Run returned %v, want context.Canceled", err)
	}
}

// TestDetectorRefusesIncompleteWiring: a detector missing a source would silently
// report an empty account, which is the most dangerous wrong answer available.
func TestDetectorRefusesIncompleteWiring(t *testing.T) {
	d := &filldetect.Detector{Clock: clock.NewFake(pollStart)}
	if _, err := d.PollOnce(context.Background()); err == nil {
		t.Fatal("a detector with no sources must refuse to run")
	}
}
