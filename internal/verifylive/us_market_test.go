package verifylive

// us_market_test.go covers verify-us-market: the run carries a market, and the
// market decides which symbols it will send orders for.
//
// The property that matters is symmetric. A run must place orders only in its own
// market — a KR run must still refuse a US symbol exactly as it did before this
// change, and a US run must refuse a KR one. The old code said "KR or nothing",
// which happened to satisfy half of that.

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/domain"
)

// usBroker is an account holding a real US position: MWG, 115 shares at $1.27,
// and no daily price band — GET /api/v1/price-limits returns null for US.
func usBroker() *fakeBroker {
	b := newFakeBroker()
	b.quotes["MWG"] = 1.27
	b.limits["MWG"] = domain.PriceLimits{}
	return b.withHolding("MWG", 115)
}

func planFor(t *testing.T, h *harness, opts Options) Plan {
	t.Helper()
	return h.runner(t, opts).Plan(context.Background())
}

// TestAUSRunPlansOrdersForTheUSSymbol.
func TestAUSRunPlansOrdersForTheUSSymbol(t *testing.T) {
	h := newHarness(t, usBroker(), alwaysConfirm())
	plan := planFor(t, h, Options{Market: MarketUS, Symbol: "MWG", HoldingSymbol: "MWG"})

	if len(plan.Mutations) == 0 {
		t.Fatalf("a US run plans nothing:\n%s", plan.Excluded)
	}
	for _, m := range plan.Mutations {
		if !strings.EqualFold(m.Symbol, "MWG") {
			t.Errorf("the plan carries %s, which is not this run's symbol", m.Symbol)
		}
		if !strings.EqualFold(MarketOf(m.Symbol), MarketUS) {
			t.Errorf("%s is not a US symbol", m.Symbol)
		}
	}
	// The band is absent, so the offset — not a clamp — decided every price, and
	// the line a person approves has to say that rather than describe a KR rule.
	for _, m := range plan.Mutations {
		if m.Pricing == "" {
			continue
		}
		if strings.Contains(m.Pricing, "clamped inside the day's price band") {
			t.Errorf("a US line claims a clamp that cannot have happened: %s", m.Pricing)
		}
		// Only the lines that price off the last trade describe a band at all;
		// "one tick further" and "identical body" have nothing to say about one.
		if strings.Contains(m.Pricing, "tick grid") {
			if !strings.Contains(m.Pricing, "no daily price band") {
				t.Errorf("a US line does not say the band is absent: %s", m.Pricing)
			}
			if !strings.Contains(m.PricingKO, "가격제한폭이 없어") {
				t.Errorf("the Korean line does not say the band is absent: %s", m.PricingKO)
			}
		}
	}
}

// TestADefaultRunStillRefusesAUSSymbol — the regression that matters. An
// unspecified market is KR, and a KR run has never sent a US order.
func TestADefaultRunStillRefusesAUSSymbol(t *testing.T) {
	h := newHarness(t, usBroker(), alwaysConfirm())
	plan := planFor(t, h, Options{Symbol: "MWG", HoldingSymbol: "MWG"})

	if len(plan.Mutations) != 0 {
		t.Fatalf("a run with no market planned %d US order(s)", len(plan.Mutations))
	}
	if len(plan.Excluded) == 0 {
		t.Fatal("nothing was excluded, so nothing explained why the plan is empty")
	}
}

// TestAUSRunRefusesAKRSymbol — the other direction, which the old rule could not
// express.
func TestAUSRunRefusesAKRSymbol(t *testing.T) {
	broker := usBroker().withHolding("005930", 3)
	h := newHarness(t, broker, alwaysConfirm())
	plan := planFor(t, h, Options{Market: MarketUS, Symbol: "MWG", HoldingSymbol: "005930"})

	for _, m := range plan.Mutations {
		if strings.EqualFold(m.Symbol, "005930") {
			t.Errorf("a US run planned an order for a KR symbol: %+v", m)
		}
	}
	var found bool
	for _, e := range plan.Excluded {
		if strings.Contains(e.Reason, "005930") {
			found = true
		}
	}
	if !found {
		t.Errorf("no exclusion explains why the KR holding was left alone: %+v", plan.Excluded)
	}
}

// TestTheUSAmendSendsNoQuantity.
//
// OrderModifyRequest.quantity: "KR 주식: 필수 … US 주식: 전달 불가. 제공 시
// `400 us-modify-quantity-not-supported`". The measurement is the same one — does a
// modify mint a new identifier — but the request the broker accepts is different.
func TestTheUSAmendSendsNoQuantity(t *testing.T) {
	broker := usBroker()
	h := newHarness(t, broker, alwaysConfirm())
	if _, err := h.run(Options{Market: MarketUS, Symbol: "MWG", HoldingSymbol: "MWG"}); err != nil {
		t.Fatalf("run: %v", err)
	}

	amends := broker.amendIntents()
	if len(amends) == 0 {
		t.Fatal("the US run never reached the amend step")
	}
	for _, a := range amends {
		if a.Quantity != nil {
			t.Errorf("a US amend carried a quantity (%v); the broker rejects that with us-modify-quantity-not-supported",
				*a.Quantity)
		}
		if a.Price == nil {
			t.Error("a US amend carried no price; a price-only modify is the whole request")
		}
	}
}

// TestTheKRAmendStillSendsTheQuantity — the KR contract is unchanged.
func TestTheKRAmendStillSendsTheQuantity(t *testing.T) {
	broker := newFakeBroker().withHolding("005930", 3)
	h := newHarness(t, broker, alwaysConfirm())
	if _, err := h.run(Options{HoldingSymbol: "005930"}); err != nil {
		t.Fatalf("run: %v", err)
	}

	amends := broker.amendIntents()
	if len(amends) == 0 {
		t.Fatal("the KR run never reached the amend step")
	}
	for _, a := range amends {
		if a.Quantity == nil {
			t.Error("a KR amend carried no quantity; the broker requires it")
		}
	}
}

// --- the record is per market -------------------------------------------------

// TestTheRecordFileIsPerMarket. A verdict is about an account *and a market*; one
// file for both would let a US pass settle a KR step, which would record a
// capability nobody measured.
func TestTheRecordFileIsPerMarket(t *testing.T) {
	kr := RecordFileName(MarketKR)
	us := RecordFileName(MarketUS)
	if kr != FileName {
		t.Errorf("the KR record moved to %q; the existing evidence file is %q", kr, FileName)
	}
	if kr == us {
		t.Fatal("both markets write the same record file")
	}
	if RecordFileName("") != FileName {
		t.Error("an unspecified market does not read the existing KR record")
	}
}

// --- the session advisory ------------------------------------------------------

func atKST(t *testing.T, y int, m time.Month, d, hh, mm int) time.Time {
	t.Helper()
	return time.Date(y, m, d, hh, mm, 0, 0, time.FixedZone("KST", 9*60*60))
}

func atET(t *testing.T, y int, m time.Month, d, hh, mm int) time.Time {
	t.Helper()
	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		t.Skipf("no tzdata on this host: %v", err)
	}
	return time.Date(y, m, d, hh, mm, 0, 0, loc)
}

// TestTheUSAdvisoryReadsTheUSSession — through internal/clock's calendar, so the
// DST rule is the one the rest of the product already uses.
func TestTheUSAdvisoryReadsTheUSSession(t *testing.T) {
	inside := SessionAdvisoryFor(MarketUS, atET(t, 2026, 7, 27, 10, 0))
	if inside.Outside {
		t.Errorf("10:00 ET on a Monday reads as outside the US session: %+v", inside)
	}
	outside := SessionAdvisoryFor(MarketUS, atET(t, 2026, 7, 27, 8, 0))
	if !outside.Outside {
		t.Errorf("08:00 ET reads as inside the US session: %+v", outside)
	}
	// Winter: 09:30 ET is 23:30 KST rather than 22:30, and the advisory has to
	// follow the zone rather than a fixed offset.
	winter := SessionAdvisoryFor(MarketUS, atET(t, 2026, 12, 14, 10, 0))
	if winter.Outside {
		t.Errorf("10:00 ET in December reads as outside the US session: %+v", winter)
	}
}

// TestTheUSAdvisorySaysItsClosedResponseIsUnmeasured.
//
// The KR advisory quotes a code this account actually returned (M1). Nothing
// equivalent has been observed for US, and borrowing KR's would be the tool
// asserting a fact about a market it has never had refused.
func TestTheUSAdvisorySaysItsClosedResponseIsUnmeasured(t *testing.T) {
	us := SessionAdvisoryFor(MarketUS, atET(t, 2026, 7, 27, 8, 0))
	if strings.Contains(us.Detail, "order-hours-closed") {
		t.Errorf("the US advisory quotes the KR measurement as if it were its own: %s", us.Detail)
	}
	if !strings.Contains(us.Detail, "미측정") {
		t.Errorf("the US advisory does not say the closed-market response is unmeasured: %s", us.Detail)
	}
	kr := SessionAdvisoryFor(MarketKR, atKST(t, 2026, 7, 26, 21, 31))
	if !strings.Contains(kr.Detail, "order-hours-closed") {
		t.Errorf("the KR advisory lost its measured code: %s", kr.Detail)
	}
}

// --- the record says which market it measured ------------------------------------

// observationDetail returns a step's recorded detail for a key.
func observationDetail(t *testing.T, entries []Entry, step StepID, key string) string {
	t.Helper()
	e, ok := LastEntry(entries, step)
	if !ok {
		t.Fatalf("no record entry for %s", step)
	}
	for _, o := range e.Observations {
		if o.Key == key {
			return o.Detail
		}
	}
	t.Fatalf("%s recorded no %s", step, key)
	return ""
}

// TestTheRecordDoesNotCallAUSRequestAKROne.
//
// The request already follows the market (TestTheUSAmendSendsNoQuantity), but the
// sentence written beside it did not: both details were fixed strings naming KR.
// A US run therefore recorded that the broker "accepted a KR price+quantity
// amend" on a request that carried no quantity at all.
//
// That is not cosmetic. The record is the evidence a later change reads to decide
// what the broker does, and a false sentence in it is worse than a missing one —
// change 2c reads exactly these lines to write the attribution rules.
func TestTheRecordDoesNotCallAUSRequestAKROne(t *testing.T) {
	h := newHarness(t, usBroker(), alwaysConfirm())
	if _, err := h.run(Options{Market: MarketUS, Symbol: "MWG", HoldingSymbol: "MWG"}); err != nil {
		t.Fatalf("run: %v", err)
	}
	entries := h.entries()

	place := observationDetail(t, entries, StepOrderCancel, "order.place.ok")
	if strings.Contains(place, MarketKR) {
		t.Errorf("a US placement was recorded as a KR one: %s", place)
	}
	if !strings.Contains(place, MarketUS) {
		t.Errorf("the placement detail does not say which market it measured: %s", place)
	}

	amend := observationDetail(t, entries, StepOrderAmend, "order.amend.ok")
	if strings.Contains(amend, MarketKR) {
		t.Errorf("a US amend was recorded as a KR one: %s", amend)
	}
	if !strings.Contains(amend, MarketUS) {
		t.Errorf("the amend detail does not say which market it measured: %s", amend)
	}
	// The US request omits the quantity, so a detail claiming one describes a
	// request the broker would have rejected with us-modify-quantity-not-supported.
	if strings.Contains(amend, "quantity") {
		t.Errorf("a US amend detail claims a quantity the request never carried: %s", amend)
	}
}

// TestTheKRRecordStillSaysKR — the other half. Naming the market must not turn
// into naming nothing.
func TestTheKRRecordStillSaysKR(t *testing.T) {
	h := newHarness(t, newFakeBroker().withHolding("005930", 3), alwaysConfirm())
	if _, err := h.run(Options{HoldingSymbol: "005930"}); err != nil {
		t.Fatalf("run: %v", err)
	}
	entries := h.entries()

	if d := observationDetail(t, entries, StepOrderCancel, "order.place.ok"); !strings.Contains(d, MarketKR) {
		t.Errorf("the KR placement detail no longer says KR: %s", d)
	}
	amend := observationDetail(t, entries, StepOrderAmend, "order.amend.ok")
	if !strings.Contains(amend, MarketKR) {
		t.Errorf("the KR amend detail no longer says KR: %s", amend)
	}
	if !strings.Contains(amend, "quantity") {
		t.Errorf("the KR amend detail drops the quantity the broker requires: %s", amend)
	}
}

// --- the cancel retry (measurements.md M16) --------------------------------------

// TestACancelRetriesWhileTheBrokerSaysAlreadyProcessing.
//
// Measured on 2026-07-27: a cancel came back 409 already-processing with
// data.retryAfterSeconds, the tool did not retry, and the order it could not
// cancel filled the live-exposure cap and blocked the two steps after it.
func TestACancelRetriesWhileTheBrokerSaysAlreadyProcessing(t *testing.T) {
	broker := newFakeBroker().withHolding("005930", 3)
	broker.cancelAlreadyProcessing = 1
	h := newHarness(t, broker, alwaysConfirm())

	if _, err := h.run(Options{HoldingSymbol: "005930"}); err != nil {
		t.Fatalf("run: %v", err)
	}

	// The refusal is consumed by whichever step cancels first; what matters is
	// that no step was left with an order it could not cancel.
	if anyFailureMentioning(h.entries(), "already-processing") {
		t.Error("a transient already-processing still ended a step")
	}
	for _, a := range Outstanding(h.entries()) {
		// The conditional left for the persistence step is deliberate; an order
		// nobody chose to leave is the failure this retry exists to prevent.
		if !a.Deliberate {
			t.Errorf("the run left %s %s live after a retryable refusal", a.Kind, a.ID)
		}
	}
	// The record says a retry happened rather than presenting it as a clean first try.
	if !anyObservation(h.entries(), "order.cancel.retries") {
		t.Error("the record does not say the cancel had to be retried")
	}
}

// TestACancelStopsRetryingAtTheCap — a broker that never lets go still ends the
// step, and the order is reported as still live.
func TestACancelStopsRetryingAtTheCap(t *testing.T) {
	broker := newFakeBroker().withHolding("005930", 3)
	broker.cancelAlreadyProcessing = 99
	h := newHarness(t, broker, alwaysConfirm())

	if _, err := h.run(Options{HoldingSymbol: "005930"}); err != nil {
		t.Fatalf("run: %v", err)
	}
	if !anyFailureMentioning(h.entries(), "already-processing") {
		t.Error("a cancel that is refused every time did not end its step")
	}
	if n := broker.countRequests("POST /orders/"); n > 12 {
		t.Errorf("%d order-scoped POST(s); the retry looks uncapped", n)
	}
	if len(Outstanding(h.entries())) == 0 {
		t.Error("an order that could not be cancelled is not reported as still live")
	}
}

// TestAPlacementIsNeverRetried — the exception is cancels and only cancels.
//
// order-cancel is the first step that places anything, so a transient refusal on
// placement has to end that step rather than be retried into a pass.
func TestAPlacementIsNeverRetried(t *testing.T) {
	broker := newFakeBroker().withHolding("005930", 3)
	broker.placeAlreadyProcessing = 1
	h := newHarness(t, broker, alwaysConfirm())

	if _, err := h.run(Options{HoldingSymbol: "005930"}); err != nil {
		t.Fatalf("run: %v", err)
	}
	if !anyFailureMentioning(h.entries(), "already-processing") {
		t.Error("a refused placement was retried away instead of ending its step")
	}
}

// --- the conditional list filter (measurements.md M17) ---------------------------

// TestTheConditionalListUsesAnAllowedStatusFilter.
//
// Measured: the endpoint allows OPEN and CLOSED only, and answers 400
// invalid-request with allowedValues for anything else. The tool was sending
// WATCHING, which is a conditional's own status rather than a filter value.
func TestTheConditionalListUsesAnAllowedStatusFilter(t *testing.T) {
	broker := newFakeBroker().withHolding("005930", 3)
	broker.rejectBadConditionalStatus = true
	h := newHarness(t, broker, alwaysConfirm())

	if _, err := h.run(Options{HoldingSymbol: "005930"}); err != nil {
		t.Fatalf("run: %v", err)
	}
	if !observationEquals(t, h.entries(), StepConditionalRegister, "conditional.list_by_status.ok", "true") {
		t.Error("the conditional list still uses a status filter the broker refuses")
	}
}
