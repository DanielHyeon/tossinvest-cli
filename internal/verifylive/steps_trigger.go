package verifylive

// steps_trigger.go is the one step in this package that wants its order to fill.
//
// Everything else here is built on a single sentence — "the price of an order that
// must not fill" (pricing.go) — and this step deliberately inverts it. The reason
// is that the last unmeasured thing the protection ledger rests on cannot be
// measured any other way: what a conditional order does *after* it fires, how long
// the order it generates takes to become visible, and whether triggeredOrderId
// links the two, are all invisible until one actually fires. Registration, query,
// persistence, modify and cancel have been measured in both markets; this has not.
//
// # What keeps it from being reckless
//
//	opt-in           without --include-trigger the step records exactly the
//	                 unverified observations it recorded when it could not be
//	                 driven at all, and none of this code runs (deferredForm)
//	one share        SINGLE + MARKET + SELL + MinQuantity, all constants
//	its own object   it registers the stop it watches rather than moving somebody
//	                 else's, so the thing it cancels on the way out is the thing it
//	                 created, inside the approval window that created it
//	every ending     is defined, including the one where the cancel and the trigger
//	                 race — see watchTrigger's tail
//
// # Why the timestamps are ours
//
// The broker supplies no fill time. lastExecutedAt is null on every completed
// order in both markets, and a US order's orderedAt carries a date and midnight
// (measurements.md M44); Order.SubmittedAt is a record-mutation time that drifted
// six hours on a real order (M45). So every time this step records is the moment
// *it* observed something, and the polling interval that produced it is written
// down next to it as the error bound. When a 429 backoff lands inside the window
// the backoff is the bound instead, and that is recorded too.
//
// # How the basis question is settled
//
// The broker does not document whether a conditional is evaluated against the last
// trade or against the best bid, and after the fact the two are indistinguishable.
// The trigger goes between them (NearStopTrigger), which makes the *order* of two
// observations the answer rather than a latency:
//
//	trigger seen before any crossing   the bid was already at or below it → bid basis
//	crossing seen first                a trade had to print down to it → last-trade basis
//
// An ordering survives a coarse polling interval; a latency comparison would not.

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/domain"
	"github.com/JungHoonGhae/tossinvest-cli/internal/official"
)

// DefaultTriggerWindow is how long the step waits for the market to reach its
// trigger before giving up and cancelling what it registered.
//
// Ten minutes. It is bounded because the operator is standing over it and because
// the whole time the stop is live it is a real stop on a real holding; it is not
// shorter because "the market has not come to the price yet" is not a failure and
// a window measured in seconds would report one.
const DefaultTriggerWindow = 10 * time.Minute

// TriggerPollIdle is the polling interval while the trigger has not been reached.
//
// Five seconds. Every tick costs a quote read and a conditional read, so ten
// minutes of waiting is about 240 requests — enough to matter against an account
// with the 429 history this one has (M4, M8, M10), which is why the measurement
// asks for the soak and the candidate watch to be stopped first.
const TriggerPollIdle = 5 * time.Second

// TriggerPollActive is the interval once something has happened: the trigger has
// been reached, or the conditional has fired. It bounds the timestamps that are
// the actual measurement, so it is as short as the rate budget bears.
const TriggerPollActive = time.Second

// TriggerLinkWindow is how long the step keeps watching after it first sees the
// conditional fire.
//
// Ninety seconds. If triggeredOrderId has not appeared by then, or the child
// order it names has not filled, that absence IS the result — a protection ledger
// that cannot see its own child order inside a minute and a half is a protection
// ledger that cannot reconcile, and recording "we did not see it" is the honest
// answer rather than waiting until something turns up.
const TriggerLinkWindow = 90 * time.Second

// stepConditionalTrigger observes a conditional order firing (task 2.5).
func (r *Runner) stepConditionalTrigger(ctx context.Context, sr *stepRun) error {
	if r.deferredForm(sr.step) {
		return r.recordTriggerUnverified(sr)
	}

	symbol := r.holdingSymbol
	sellable, err := r.readSellable(ctx, sr, symbol)
	if err != nil {
		return err
	}
	if sellable < MinQuantity {
		// The conditional-register step's stop reserves a share (measured:
		// conditional.reserves_sellable_quantity), so on a one-share holding this
		// is where a full run stops — before anything is sold rather than after.
		sr.skip(fmt.Sprintf("%s의 매도가능수량이 %s주다. 발동 관측은 실제로 팔릴 1주가 필요하고, "+
			"이미 등록된 조건주문이 예약하고 있다면 그것을 먼저 취소해야 한다", symbol, trim(sellable)))
		return nil
	}

	top, err := r.marketTop(ctx, sr, symbol)
	if err != nil {
		return err
	}
	trigger, err := NearStopTrigger(top.last, top.bid, MarketOf(symbol))
	if err != nil {
		sr.skip(truncateError(err))
		return nil
	}
	sr.observe("conditional.trigger.placement", trim(trigger.Price), trigger.Basis)
	sr.observe("conditional.trigger.distinguishes_basis", strconv.FormatBool(trigger.Distinguishes),
		"a trigger sitting on the last trade fires under either reading of the broker's rule and settles "+
			"nothing about which one it uses")
	sr.observe("conditional.trigger.book_at_registration", topOfBook(top),
		"the bid, ask and last trade the trigger price was derived from")

	body := official.ConditionalCreateBody{
		Symbol:        symbol,
		Type:          "SINGLE",
		Quantity:      trim(MinQuantity),
		OrderType:     "MARKET",
		ClientOrderID: newToken("TRIGGER"),
		ExpireDate:    r.expireDate(),
		First:         official.ConditionLegBody{OrderSide: "SELL", TriggerPrice: trim(trigger.Price)},
	}
	// The holding before anything is registered. It is the ground truth the race
	// check falls back on: a cancelled conditional stops reading back entirely
	// (conditional.cancel.gone_after, measured), so "did it fire after all" cannot
	// be answered from the conditional itself — but a share that left the account
	// is not ambiguous.
	held, heldKnown := r.readHolding(ctx, sr, symbol)
	sr.observe("conditional.trigger.holding_before", trim(held),
		"the position this step is measured against; a drop means a share was sold")

	id, err := r.createConditional(ctx, sr, body, trigger.Basis)
	if err != nil {
		return err
	}
	chain := newToken("chain")
	// Joined to the chain, but deliberately NOT held and NOT marked deliberate.
	//
	// markHeld would say two things about this object that are not true. "Held"
	// keeps the cleanup prologue away, which is right for the child order and wrong
	// here: on every planned path this conditional ends terminal inside this step,
	// so the only way it survives is an interrupt — and then it is a stop priced to
	// fire, sitting on a real holding, with nobody watching. "Deliberate" would
	// exempt it from the end-of-run leftover check, which is the one thing that
	// makes the tool refuse to report a clean finish while it is still live.
	sr.joinChain(KindConditional, id, chain,
		"registered to be reached. If this run ends without a verdict for this step, it is a live stop that "+
			"can fire — cancel it with `tossctl verify abort`")

	return r.watchTrigger(ctx, sr, triggerWatch{
		symbol: symbol, conditionalID: id, chain: chain, trigger: trigger,
		heldBefore: held, heldKnown: heldKnown,
	})
}

// readHolding is the position in one symbol, and whether it could be read at all.
//
// A failure is not fatal here and is not guessed around: it costs the race check
// its most reliable signal, and finishTrigger says so rather than treating an
// unreadable account as an unchanged one.
func (r *Runner) readHolding(ctx context.Context, sr *stepRun, symbol string) (float64, bool) {
	positions, err := readRetry(ctx, r, sr, EndpointReadHoldings, map[string]string{"symbol": symbol},
		func(ctx context.Context) ([]domain.Position, error) { return r.broker.Holdings(ctx, symbol) },
		func(p []domain.Position) any { return len(p) })
	if err != nil {
		return 0, false
	}
	for _, p := range positions {
		if strings.EqualFold(p.Symbol, symbol) {
			return p.Quantity, true
		}
	}
	return 0, true
}

// recordTriggerUnverified is the step as it behaves without the opt-in.
//
// Byte for byte what it recorded before it could be driven at all. task 2.6 reads
// these three keys to decide which markets and order types automatic entry stays
// forbidden on, and a run that quietly stopped writing them would let that list
// shrink without anything having been measured.
func (r *Runner) recordTriggerUnverified(sr *stepRun) error {
	sr.observe("conditional.trigger_observed", "false", sr.step.Deferred)
	sr.observe("conditional.triggered_order_id_exposed", "unverified",
		"triggeredOrderId links a conditional to the order it generated; it can only be read after a trigger")
	sr.observe("conditional.triggered_order_latency", "unverified", "")
	sr.deferStep(sr.step.Deferred)
	return nil
}

// triggerWatch is what the observation loop is watching.
type triggerWatch struct {
	symbol        string
	conditionalID string
	chain         string
	trigger       SafePrice
	// heldBefore is the position before the stop was registered, and heldKnown
	// whether it could be read. The race check's last resort.
	heldBefore float64
	heldKnown  bool
}

// book is the top of the order book plus the last trade.
//
// One level a side and no more. US answers /orderbook with a single level and KR
// with ten (measurements.md M49), so anything that read depth here would be
// reading an endpoint limitation in one market and a real book in the other.
type book struct {
	last, bid, ask float64
}

func topOfBook(b book) string {
	return fmt.Sprintf("bid %s / ask %s / last %s", trim(b.bid), trim(b.ask), trim(b.last))
}

// watchTrigger is the two-speed observation.
//
// Both the quote and the conditional are polled from the start. Polling only the
// quote until a crossing is seen — the shape this was first drafted as — cannot
// observe the case it exists to distinguish: under a bid basis the conditional
// fires without any crossing ever appearing in the last trade.
//
// The interval tightens the moment anything happens, because from then on it is
// the error bound on the numbers that are the measurement.
func (r *Runner) watchTrigger(ctx context.Context, sr *stepRun, w triggerWatch) error {
	var (
		obs      = &triggerObservation{interval: TriggerPollIdle}
		started  = r.now()
		deadline = started.Add(r.triggerWindow)
		backoff0 = r.readBackoffs
	)
	fmt.Fprintf(r.out, "    조건주문 %s 발동 대기 — 발동가 %s, 최대 %s\n",
		w.conditionalID, trim(w.trigger.Price), r.triggerWindow)

	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		done, err := r.pollTrigger(ctx, sr, w, obs)
		if err != nil {
			return err
		}
		if done {
			return r.finishTrigger(sr, w, obs, backoff0)
		}

		now := r.now()
		switch {
		case !obs.triggeredAt.IsZero() && now.After(obs.triggeredAt.Add(TriggerLinkWindow)):
			// The fire was seen and the rest of the chain did not arrive. That
			// absence is the result, not a reason to keep waiting.
			return r.finishTrigger(sr, w, obs, backoff0)
		case !obs.crossedAt.IsZero() && now.After(obs.crossedAt.Add(TriggerLinkWindow)):
			// The price reached the trigger and the broker did not fire. This is
			// the single most important negative this step can produce — a
			// protective stop whose condition was met and which did nothing — so it
			// gets its own ending rather than being folded into "the market never
			// came". The step still cancels what it registered on the way out.
			obs.crossedWithoutFiring = true
			return r.concludeUnreached(ctx, sr, w, obs, backoff0)
		case obs.triggeredAt.IsZero() && now.After(deadline):
			return r.concludeUnreached(ctx, sr, w, obs, backoff0)
		}
		if err := r.sleep(ctx, obs.interval); err != nil {
			return err
		}
	}
}

// triggerObservation accumulates the four timestamps, each with the polling
// interval that bounds it.
type triggerObservation struct {
	interval time.Duration

	crossedAt       time.Time
	crossedInterval time.Duration

	triggeredAt       time.Time
	triggeredInterval time.Duration
	triggeredEvidence string
	bookAtTrigger     book

	childID         string
	childSeenAt     time.Time
	childInterval   time.Duration
	childFilledAt   time.Time
	childFilledIntv time.Duration
	childStatus     string

	// cancelled records that this step removed its own conditional on the way out.
	cancelled bool
	// raceUnknown records that after that cancel neither the conditional nor the
	// position could be read, so whether a share was sold is not known. It is the
	// one ending that must never be filed as a clean inconclusive.
	raceUnknown bool
	// crossedWithoutFiring records the negative this step exists to be able to
	// find: the price reached the trigger and the broker did nothing.
	crossedWithoutFiring bool
}

// pollTrigger is one tick: read the book, read the conditional, read the child if
// there is one. It reports that the observation is complete.
func (r *Runner) pollTrigger(ctx context.Context, sr *stepRun, w triggerWatch, obs *triggerObservation) (bool, error) {
	// The book is read only while the answer can still change. Once the trigger has
	// been observed the crossing question is settled and the top of book at that
	// moment is already recorded, so continuing to poll /prices and /orderbook
	// would spend the rate budget on nothing — and this window is exactly where a
	// 429 costs the whole measurement (issues.md J5).
	var top book
	if obs.triggeredAt.IsZero() {
		var err error
		top, err = r.marketTop(ctx, sr, w.symbol)
		if err == nil && obs.crossedAt.IsZero() && top.last > 0 && top.last <= w.trigger.Price {
			obs.crossedAt, obs.crossedInterval = r.now(), obs.interval
			obs.interval = TriggerPollActive
			fmt.Fprintf(r.out, "    임계 도달 관측 — 최종체결가 %s <= 발동가 %s\n",
				trim(top.last), trim(w.trigger.Price))
		}
	}

	co, err := r.readConditional(ctx, sr, w.conditionalID)
	if err != nil {
		// A read that failed is not evidence of anything. It is logged as a Call by
		// readConditional, and the loop tries again on the next tick.
		return false, nil
	}
	if evidence, fired := firedEvidence(co); fired && obs.triggeredAt.IsZero() {
		obs.triggeredAt, obs.triggeredInterval = r.now(), obs.interval
		obs.triggeredEvidence, obs.bookAtTrigger = evidence, top
		obs.interval = TriggerPollActive
		fmt.Fprintf(r.out, "    발동 관측 — %s\n", evidence)
	}
	if id := strings.TrimSpace(co.First.TriggeredOrderID); id != "" && obs.childID == "" {
		obs.childID, obs.childSeenAt, obs.childInterval = id, r.now(), obs.interval
		sr.created(KindOrder, id, w.symbol, r.now(),
			"created by conditional "+w.conditionalID+" firing; it is meant to fill")
		sr.markHeld(KindOrder, id, StepConditionalTrigger, w.chain,
			"the child order a trigger produced. Letting it fill IS the measurement — it is not a leak, "+
				"and no cleanup may cancel it")
		fmt.Fprintf(r.out, "    child 주문 노출 %s\n", id)
	}

	if obs.childID == "" {
		return false, nil
	}
	view, err := r.readOrder(ctx, sr, obs.childID)
	if err != nil {
		return false, nil
	}
	obs.childStatus = view.Status
	if parseDecimal(view.Execution.FilledQuantity) <= 0 && !strings.EqualFold(view.Status, "FILLED") {
		return false, nil
	}
	obs.childFilledAt, obs.childFilledIntv = r.now(), obs.interval
	fmt.Fprintf(r.out, "    child 주문 체결 확인 %s (%s)\n", obs.childID, orDash(view.Status))
	return true, nil
}

// concludeUnreached is the ending where the market never came to the trigger.
//
// The step cancels what it registered, inside the approval window that registered
// it — which is what the other eleven mutating steps do with their own objects,
// and is a different thing from the clock-based lease the cleanup prologue was
// deliberately not given (verify-holds-what-it-awaits design.md D3).
//
// Then it reads once more. A cancel and a trigger can race, and the worst possible
// ending is the broker having already sold a share while this tool records that
// nothing happened.
func (r *Runner) concludeUnreached(ctx context.Context, sr *stepRun, w triggerWatch,
	obs *triggerObservation, backoff0 int) error {
	if err := r.cancelConditional(ctx, sr, w.conditionalID, w.symbol,
		"관측 창 안에 임계에 도달하지 않았다 — 이 단계가 등록한 것을 이 단계가 거둔다"); err != nil {
		return err
	}
	obs.cancelled = true

	evidence, fired := r.raceEvidence(ctx, sr, w, obs)
	if !fired {
		return r.finishTrigger(sr, w, obs, backoff0)
	}

	// It fired anyway. Undo none of it: pick the observation back up where it left
	// off, because a child order may already exist and it has to be found.
	sr.observe("conditional.trigger.cancel_race_recheck", "fired",
		"the cancel and the trigger raced and the trigger won: "+evidence)
	obs.triggeredAt, obs.triggeredInterval = r.now(), obs.interval
	obs.triggeredEvidence = evidence + " (observed on the re-read after this step's own cancel)"
	obs.interval = TriggerPollActive
	fmt.Fprintf(r.out, "    취소 직후 재확인에서 발동 흔적 — 관측을 계속한다\n")

	deadline := r.now().Add(TriggerLinkWindow)
	for r.now().Before(deadline) {
		if err := ctx.Err(); err != nil {
			return err
		}
		done, err := r.pollTrigger(ctx, sr, w, obs)
		if err != nil {
			return err
		}
		if done {
			break
		}
		if err := r.sleep(ctx, obs.interval); err != nil {
			return err
		}
	}
	return r.finishTrigger(sr, w, obs, backoff0)
}

// raceEvidence answers "did it fire after all" once the step has cancelled its own
// conditional.
//
// The obvious check — read the conditional again — is the one that does not work.
// A cancelled conditional stops reading back on this broker (measured:
// conditional.cancel.gone_after), so a 404 says "cancelled" and "fired and gone"
// with the same words. It is still asked first, because a broker that does keep it
// readable answers unambiguously and cheaply.
//
// The fallback is the account itself. A share that left it is not ambiguous, it is
// what "the broker sold something" actually means, and it is what a person would
// go and look at. When even that cannot be read the step says so and fails rather
// than reporting a clean nothing-happened — an unreadable account after a cancel
// that may have raced a sale is precisely the state that must not be filed as
// inconclusive.
func (r *Runner) raceEvidence(ctx context.Context, sr *stepRun, w triggerWatch,
	obs *triggerObservation) (string, bool) {
	if co, err := r.readConditional(ctx, sr, w.conditionalID); err == nil {
		if evidence, fired := firedEvidence(co); fired {
			sr.observe("conditional.trigger.cancel_race_recheck", "fired",
				"the cancel and the trigger raced and the trigger won: "+evidence)
			return evidence, true
		}
		sr.observe("conditional.trigger.cancel_race_recheck", "clean",
			"the conditional still reads back after the cancel and shows no sign of having fired")
		return "", false
	}

	held, ok := r.readHolding(ctx, sr, w.symbol)
	sr.observe("conditional.trigger.holding_after", trim(held),
		"read after this step cancelled its own conditional, because a cancelled conditional no longer "+
			"reads back and the position is the only thing left that can say whether a share was sold")
	switch {
	case !ok || !w.heldKnown:
		obs.raceUnknown = true
		sr.observe("conditional.trigger.cancel_race_recheck", "unreadable",
			"neither the conditional nor the position could be read after the cancel, so whether a share "+
				"was sold is unknown")
		return "", false
	case held < w.heldBefore:
		evidence := fmt.Sprintf("the position fell from %s to %s after the cancel",
			trim(w.heldBefore), trim(held))
		sr.observe("conditional.trigger.cancel_race_recheck", "fired", evidence)
		return evidence, true
	default:
		sr.observe("conditional.trigger.cancel_race_recheck", "clean",
			"the conditional is gone and the position is unchanged at "+trim(held))
		return "", false
	}
}

// finishTrigger writes the measurement and decides the verdict.
func (r *Runner) finishTrigger(sr *stepRun, w triggerWatch, obs *triggerObservation, backoff0 int) error {
	backoffs := r.readBackoffs - backoff0
	sr.observe("conditional.trigger.backoffs_in_window", strconv.Itoa(backoffs),
		"429 backoffs inside the observation window. Where one of these landed, the error bound on the "+
			"timestamps around it is the backoff and not the polling interval")

	r.observeStamp(sr, "condition_crossed_at", obs.crossedAt, obs.crossedInterval,
		"the first poll at which the last trade was at or below the trigger")
	r.observeStamp(sr, "trigger_first_observed_at", obs.triggeredAt, obs.triggeredInterval,
		obs.triggeredEvidence)
	r.observeStamp(sr, "triggered_order_id_first_seen_at", obs.childSeenAt, obs.childInterval,
		"child order "+orNone(obs.childID))
	r.observeStamp(sr, "child_order_filled_at", obs.childFilledAt, obs.childFilledIntv,
		"status "+orDash(obs.childStatus))

	// The three keys the report has always read, now answered.
	sr.observe("conditional.trigger_observed", strconv.FormatBool(!obs.triggeredAt.IsZero()),
		obs.triggeredEvidence)
	sr.observe("conditional.triggered_order_id_exposed", strconv.FormatBool(obs.childID != ""),
		orNone(obs.childID))
	if !obs.triggeredAt.IsZero() && !obs.childSeenAt.IsZero() {
		sr.observe("conditional.triggered_order_latency",
			strconv.FormatInt(obs.childSeenAt.Sub(obs.triggeredAt).Milliseconds(), 10),
			"milliseconds between observing the trigger and observing the identifier it produced, bounded "+
				"by a "+obs.childInterval.String()+" polling interval")
	} else {
		sr.observe("conditional.triggered_order_latency", "unverified", "")
	}

	sr.observe("conditional.trigger.basis", triggerBasis(obs), triggerBasisDetail(obs))
	if !obs.triggeredAt.IsZero() {
		sr.observe("conditional.trigger.book_at_trigger", topOfBook(obs.bookAtTrigger),
			"the top of book at the poll that first saw the trigger — the only evidence that narrows what "+
				"the broker evaluated")
	}

	switch {
	case !obs.childFilledAt.IsZero():
		// The whole point. The child is gone from the account because it filled,
		// and the conditional is gone because it fired; both are terminal, and the
		// record says which ending each one had.
		sr.filled(KindOrder, obs.childID, w.symbol, obs.childFilledAt, w.chain,
			"filled — this is the measurement, and it is why the order was held rather than cancelled")
		sr.filled(KindConditional, w.conditionalID, w.symbol, obs.triggeredAt, w.chain,
			"fired and produced order "+obs.childID+", so it no longer exists as a conditional")
		sr.pass()
		return nil
	case !obs.triggeredAt.IsZero():
		// The conditional is NOT recorded as terminal here, and that is deliberate.
		// It looks like it fired, but the step did not see the fill that would make
		// that certain, and the two ways of being wrong are not symmetric: calling a
		// live stop gone leaves a fire-capable order nothing is tracking, while
		// calling a gone one live costs a 404 on a cancel nobody needed. The
		// end-of-run check will report it and `verify abort` clears it.
		sr.observe("conditional.trigger.conditional_presumed_fired", "true",
			"the conditional shows a trigger but the step did not observe what it produced, so it is left "+
				"on the record as live rather than assumed gone")
		sr.fail("조건주문 %s의 발동은 관측했지만 %s 안에 %s. 이 조건주문과 child 주문은 살아 있는 것으로 "+
			"기록에 남는다 — `tossctl verify status`가 출력하고 `tossctl verify abort`가 끝낸다",
			w.conditionalID, TriggerLinkWindow, missingLink(obs))
		return nil
	case obs.crossedWithoutFiring:
		// The one negative that matters most to 2c: the price met the stop's
		// condition and the broker did nothing with it.
		sr.observe("conditional.fires_when_its_condition_is_met", "false",
			fmt.Sprintf("the last trade reached %s and no trigger was observed in the %s that followed",
				trim(w.trigger.Price), TriggerLinkWindow))
		sr.fail("최종체결가가 발동가 %s에 닿았는데 그 뒤 %s 동안 발동이 관측되지 않았다. 조건주문 %s는 "+
			"취소했다 — 보호 주문이 조건을 만족하고도 발동하지 않는다면 2c의 보호 원장은 이 브로커의 "+
			"조건주문에 기댈 수 없다", trim(w.trigger.Price), TriggerLinkWindow, w.conditionalID)
		return nil
	case obs.cancelled && obs.raceUnknown:
		// Not a skip. The step cancelled something that may have sold a share and
		// then could not read the account, and filing that as "nothing happened"
		// is the one outcome this design refuses to produce.
		sr.fail("조건주문 %s를 취소했지만 그 직후 조건주문도 보유 수량도 읽지 못했다. 취소와 발동이 "+
			"경합했는지 알 수 없다 — `tossctl holdings`로 %s 수량을 직접 확인하라",
			w.conditionalID, w.symbol)
		return nil
	case obs.cancelled:
		sr.skip(fmt.Sprintf("INCONCLUSIVE — %s 안에 시장이 발동가 %s에 닿지 않았다. 이 단계가 등록한 "+
			"조건주문을 취소했고 계좌에 남은 것은 없다. 측정 실패가 아니라 시장 조건이 오지 않은 것이다",
			r.triggerWindow, trim(w.trigger.Price)))
		return nil
	default:
		sr.fail("발동 관측이 결말에 도달하지 못했다 — 조건주문 %s가 아직 계좌에 있을 수 있다", w.conditionalID)
		return nil
	}
}

func missingLink(obs *triggerObservation) string {
	if obs.childID == "" {
		return "triggeredOrderId가 노출되지 않았다"
	}
	return "child 주문 " + obs.childID + "의 체결을 확인하지 못했다"
}

// observeStamp writes one of the four timestamps together with what bounds it.
//
// A stamp with no interval beside it cannot be read afterwards: the broker gives
// no time of its own (M44), so the only honest form of "it happened at T" here is
// "this tool saw it at T, having last looked I ago".
func (r *Runner) observeStamp(sr *stepRun, name string, at time.Time, interval time.Duration, detail string) {
	key := "conditional.trigger." + name
	if at.IsZero() {
		sr.observe(key, "unobserved", detail)
		return
	}
	sr.observe(key, at.UTC().Format(time.RFC3339Nano),
		fmt.Sprintf("observed by this tool, ±%s (the polling interval). %s", interval, detail))
}

// triggerBasis answers the question the placement was chosen to answer.
func triggerBasis(obs *triggerObservation) string {
	switch {
	case obs.triggeredAt.IsZero():
		return "unobserved"
	case obs.crossedAt.IsZero():
		// It fired and the last trade was never seen at or below the trigger, which
		// it cannot have been: the trigger sits above the bid.
		return "bid"
	case obs.crossedAt.After(obs.triggeredAt):
		return "bid"
	default:
		return "last-trade"
	}
}

func triggerBasisDetail(obs *triggerObservation) string {
	switch triggerBasis(obs) {
	case "bid":
		return "the conditional fired before any trade printed at or below the trigger, so the broker was " +
			"evaluating the bid rather than the last trade"
	case "last-trade":
		return "a trade printed at or below the trigger before the conditional fired, which is consistent " +
			"with a last-trade basis and does not rule out a bid basis reached at the same moment"
	default:
		return "no trigger was observed, so the basis is still unmeasured"
	}
}

// firedEvidence reports what says the conditional fired.
//
// The status vocabulary of a fired conditional is itself unmeasured — this step is
// the thing that measures it — so the match is deliberately loose and the raw
// value is recorded either way. triggeredOrderId is the unambiguous signal and is
// checked first; a status is corroboration.
//
// A cancelled conditional must not read as fired, which is why the match is on
// positive words rather than on "not WATCHING": concludeUnreached asks this
// question immediately after its own cancel.
func firedEvidence(co conditionalView) (string, bool) {
	if id := strings.TrimSpace(co.First.TriggeredOrderID); id != "" {
		return "triggeredOrderId " + id, true
	}
	for _, s := range []string{co.First.Status, co.Status} {
		up := strings.ToUpper(strings.TrimSpace(s))
		for _, hint := range []string{"TRIGGER", "EXECUT", "ORDERED", "FILLED"} {
			if strings.Contains(up, hint) {
				return "status " + up, true
			}
		}
	}
	return "", false
}

// marketTop reads the last trade and the best bid and ask.
//
// Two calls, because domain.Quote carries no bid: the last trade comes from
// /prices, which every other step already uses, and the top of book from
// /orderbook. A book that cannot be read is an error rather than a zero — a
// trigger placed without a bid is a trigger placed by guesswork.
func (r *Runner) marketTop(ctx context.Context, sr *stepRun, symbol string) (book, error) {
	quotes, err := readRetry(ctx, r, sr, EndpointReadPrices, map[string]string{"symbol": symbol},
		func(ctx context.Context) ([]domain.Quote, error) { return r.broker.Prices(ctx, []string{symbol}) },
		func(q []domain.Quote) any { return len(q) })
	if err != nil {
		return book{}, err
	}
	var out book
	for _, q := range quotes {
		if strings.EqualFold(q.Symbol, symbol) && q.Last > 0 {
			out.last = q.Last
			break
		}
	}
	if out.last <= 0 && len(quotes) > 0 {
		out.last = quotes[0].Last
	}

	ob, err := readRetry(ctx, r, sr, EndpointReadOrderbook, map[string]string{"symbol": symbol},
		func(ctx context.Context) (domain.OrderBook, error) { return r.broker.Orderbook(ctx, symbol) },
		func(b domain.OrderBook) any { return len(b.Bids) })
	if err != nil {
		return book{}, err
	}
	if len(ob.Bids) > 0 {
		out.bid = ob.Bids[0].Price
	}
	if len(ob.Offers) > 0 {
		out.ask = ob.Offers[0].Price
	}
	return out, nil
}
