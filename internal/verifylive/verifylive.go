// Package verifylive is the guided live-account verification the
// verify-execution-capability change needs a human to run once (task 1.5).
//
// # What it is, and what it is not
//
// internal/soak proves the read side of the automation gate and structurally
// cannot prove anything else: it holds no mutation transport at all. The other
// half — that an order can be placed, cancelled and amended, that a conditional
// order survives this process dying, that the broker's documented idempotency key
// behaves the way the document says — cannot be established from a build, from a
// fixture, or from a read-only survey. It can only be established by placing real
// orders on a real account.
//
// So this package places real orders. Everything else about its design follows
// from that one sentence:
//
//	operator tool     it is driven by a person at a terminal, never by CI, never
//	                  by an agent, never by a cron job. Before anything is sent the
//	                  run lists every live request it plans to make and waits for
//	                  one typed, expiring confirmation covering exactly that list
//	                  (plan.go, confirm.go); --confirm-each asks per mutation
//	                  instead. There is no flag anywhere that answers either.
//	minimum exposure  one share, limit-only, priced far enough from the market
//	                  that it cannot fill (pricing.go), cancelled inside the same
//	                  step, and never more than one live order from this tool at a
//	                  time (Runner.ledger).
//	evidence first    every step appends a durable JSONL line carrying request and
//	                  response digests, latencies and a verdict (record.go), and
//	                  the approval itself is one more line. The point of the
//	                  exercise is the record, not the orders.
//
// WORKFLOW §0 forbids unattended LIVE order side effects and the 불변 규칙 forbid
// automated tests that place real orders. Both hold here: this package's own
// tests drive a fake broker or an httptest server, and internal/testenv's guard
// makes a test that reaches a real Toss hostname fail rather than trade.
//
// # The step list is the contract
//
// Steps() is the whole procedure, in order, as data. Each step says what it
// proves, which measurement task it feeds, whether it mutates, what it needs from
// the account, and — in Step.Mutations — every class of live request it will send.
// An operator can read it before running anything (`tossctl verify run --list`),
// and the runner walks exactly that list: there is no hidden step, no step that
// runs without appearing in it, and no request that can be sent without a
// Step.Mutations line to authorise it (plan.go).
package verifylive

import (
	"context"
	"encoding/json"

	"github.com/JungHoonGhae/tossinvest-cli/internal/domain"
	"github.com/JungHoonGhae/tossinvest-cli/internal/official"
	"github.com/JungHoonGhae/tossinvest-cli/internal/orderintent"
	trading "github.com/JungHoonGhae/tossinvest-cli/internal/trading"
)

// StepID names one step of the procedure. It is the record's join key, so the
// values are stable strings rather than an iota: a renamed constant would orphan
// every line already written about it.
type StepID string

// The steps, in the order the runner walks them.
const (
	// StepReadFixtures collects the order status enum shapes from the account's
	// existing history (task 2.1). Read-only.
	StepReadFixtures StepID = "read-fixtures"
	// StepSellableBaseline snapshots holdings against sellableQuantity before
	// anything is placed (task 2.8, the "before" half).
	StepSellableBaseline StepID = "sellable-baseline"
	// StepIdempotency is the clientOrderId replay check (task 2.7).
	StepIdempotency StepID = "idempotency"
	// StepIdempotencyTTLEdge probes the validity window (task 2.7). Opt-in: it
	// deliberately creates a second live order.
	StepIdempotencyTTLEdge StepID = "idempotency-ttl-edge"
	// StepOrderCancel is place-then-cancel on a KR limit order (task 2.2).
	StepOrderCancel StepID = "order-cancel"
	// StepOrderAmend is the KR amend path (task 2.2).
	StepOrderAmend StepID = "order-amend"
	// StepSellBoundary is the sell-side boundary set (task 2.2). Needs a holding.
	StepSellBoundary StepID = "sell-boundary"
	// StepConditionalRegister registers a SINGLE conditional order (task 2.5).
	StepConditionalRegister StepID = "conditional-register"
	// StepSellableReserved re-reads sellableQuantity while the conditional is
	// live (task 2.8, the "after" half).
	StepSellableReserved StepID = "sellable-reserved"
	// StepConditionalPersist proves the conditional survived this process exiting
	// (task 2.5). It is the reason the tool is resumable.
	StepConditionalPersist StepID = "conditional-persist"
	// StepConditionalTrigger is trigger observation and triggeredOrderId latency
	// (task 2.5). Deferred by default: it needs the market to come to the price.
	StepConditionalTrigger StepID = "conditional-trigger"
	// StepConditionalModify checks that a modify issues a new id and invalidates
	// the old one (task 2.5).
	StepConditionalModify StepID = "conditional-modify"
	// StepConditionalCancel cancels the conditional and confirms it is gone.
	StepConditionalCancel StepID = "conditional-cancel"
	// StepCosts collects execution.commission from whatever this run filled
	// (task 2.9's input).
	StepCosts StepID = "costs"
)

// FlagIncludeTTLEdge is the opt-in that unlocks StepIdempotencyTTLEdge. It is
// named here rather than in the command so the step catalogue can print the exact
// flag an operator has to pass.
const FlagIncludeTTLEdge = "--include-ttl-edge"

// Step is one entry of the procedure.
type Step struct {
	// ID is the record's join key.
	ID StepID `json:"id"`
	// Title is the one-line name an operator sees.
	Title string `json:"title"`
	// Proves is what a passing verdict establishes. It is written for the person
	// deciding whether to type a confirmation, not for a changelog.
	Proves string `json:"proves"`
	// Tasks are the verify-execution-capability tasks this step feeds.
	Tasks []string `json:"tasks"`
	// Mutates reports that the step places, amends or cancels something live.
	Mutates bool `json:"mutates"`
	// Mutations declares every class of live request the step will send, as data.
	//
	// It is what the batch approval enumerates and what the plan authorises
	// (plan.go), so it is not a description of the step — it is the step's
	// permission slip. A mutating step with an empty list can send nothing, and
	// verifylive_test.go asserts the two flags agree.
	Mutations []StepMutation `json:"mutations,omitempty"`
	// NeedsHolding reports that the step cannot run on an account with nothing in
	// it. Such a step is skipped with a reason; the tool never buys to create the
	// holding it needs.
	NeedsHolding bool `json:"needs_holding,omitempty"`
	// ActsOnConditional reports that the step's requests are about the conditional
	// order this verification already registered, rather than about the run's probe
	// symbol.
	//
	// It exists because the approval list has to name the object. The plan is built
	// before anything runs and has to state which symbol each line is about; a step
	// that resolves its target from the record at run time (liveConditional) would
	// otherwise be listed against a symbol nobody is going to send anything for —
	// which is exactly what stopped the KR run twice on 2026-07-29. It is
	// deliberately NOT NeedsHolding: a leftover conditional has to stay cancellable
	// on an account whose holding has since gone, which is when a leftover matters
	// most.
	ActsOnConditional bool `json:"acts_on_conditional,omitempty"`
	// OptIn is the flag that unlocks the step, empty when it runs by default.
	OptIn string `json:"opt_in,omitempty"`
	// Deferred marks a step that this tool cannot drive at all, and says so up
	// front rather than pretending it might pass.
	Deferred string `json:"deferred,omitempty"`
	// DependsOn are the steps that must have passed first. A step whose
	// dependency did not pass is skipped, not attempted.
	DependsOn []StepID `json:"depends_on,omitempty"`
	// Procedure is what the step will actually do, in order.
	Procedure []string `json:"procedure"`
}

// Steps returns the procedure.
//
// The order matters twice over: sellable-reserved has to run while the
// conditional is live, and conditional-persist has to run in a *different*
// process from conditional-register, which is what splits a full run into two
// invocations.
func Steps() []Step {
	return []Step{
		{
			ID:     StepReadFixtures,
			Title:  "Order status enum fixtures from existing history",
			Proves: "문서화된 열 가지 OrderStatus 값 중 이 계좌가 실제로 만들어낸 것이 무엇이고, CANCEL_REJECTED/REPLACE_REJECTED 레코드가 주문 목록에서 어떤 모양인지",
			Tasks:  []string{"2.1"},
			Procedure: []string{
				"GET /api/v1/orders를 CLOSED(커서 페이지네이션)와 OPEN(전량 반환) 두 그룹 모두 " +
					"페이지 상한까지 훑는다 — status는 필수이고 '전체'에 해당하는 값이 없다",
				"서로 다른 status·필드 모양·건수를 기록한다 — 식별자는 digest만 남기고 값 자체는 저장하지 않는다",
				"CANCEL_REJECTED / REPLACE_REJECTED가 독립 레코드로 나타나는지, 원주문으로 되짚는 필드를 갖는지 기록한다",
			},
		},
		{
			ID:     StepSellableBaseline,
			Title:  "Sellable-quantity baseline",
			Proves: "이 도구가 아무것도 접수하기 전의 holdings.quantity와 sellableQuantity 사이의 관계",
			Tasks:  []string{"2.8"},
			Procedure: []string{
				"GET /api/v1/holdings",
				"보유 종목마다 GET /api/v1/sellable-quantity",
				"종목별 차이를 뒤 단계의 비교 기준선으로 기록한다",
			},
		},
		{
			ID:      StepIdempotency,
			Title:   "clientOrderId replay, conflict and scope",
			Proves:  "같은 clientOrderId를 다시 보내면 두 번째 주문을 만드는 대신 첫 주문을 돌려주는지, 같은 키에 본문이 다르면 idempotency-key-conflict로 거부되는지",
			Tasks:   []string{"2.7"},
			Mutates: true,
			Procedure: []string{
				"키 K로 최소 수량 LIMIT 매수 1건을 시장에서 멀리 접수한다",
				"같은 키 K로 동일한 본문을 재전송한다 — 동일성 복구가 이 단계의 검증 대상이다",
				"같은 orderId가 돌아오고 미체결 목록이 늘지 않았음을 확인한다",
				"같은 키 K로 본문이 다른 요청을 보내고 에러 코드를 기록한다",
				"이 단계가 남긴 주문을 전부 취소한다",
			},
			Mutations: []StepMutation{
				{
					Kind: MutatePlaceOrder, Side: "buy", Quantity: QuantityOne, Pricing: PriceFarBuy,
					Ends:   "cancelled inside this step",
					EndsKO: "이 단계 안에서 취소된다",
				},
				{
					Kind: MutateReplayOrder, Side: "buy", Quantity: QuantityOne, Pricing: PriceIdenticalBody,
					Ends: "the documented answer is the first order coming back and nothing new existing; " +
						"anything it does create is cancelled inside this step",
					EndsKO: "문서상 정답은 첫 주문이 그대로 돌아오고 새로 생기는 것이 없다는 것이다. " +
						"그래도 새로 생기면 이 단계 안에서 취소된다",
					Note: "this replay is the claim under test, so it is sent as part of the line above rather " +
						"than as a separate decision",
					NoteKO: "이 재전송 자체가 검증 대상이므로 별도 결정이 아니라 위 줄의 일부로 나간다",
				},
				{
					Kind: MutateConflictProbe, Side: "buy", Quantity: QuantityOne, Pricing: PriceOneTickFurther,
					Ends: "expected to be refused with idempotency-key-conflict; if the broker accepts it " +
						"instead, the order it creates is cancelled inside this step",
					EndsKO: "idempotency-key-conflict로 거부될 것으로 예상한다. 브로커가 받아들이면 " +
						"그 주문은 이 단계 안에서 취소된다",
					Note: "deliberately re-uses the key above with a DIFFERENT price — that refusal is the " +
						"measurement",
					NoteKO: "위 키를 의도적으로 재사용하되 가격을 다르게 한다 — 그 거부가 곧 측정값이다",
				},
				{
					Kind: MutateCancelOrder,
					Ends: "the account returns to how this step found it — every order the step has resting " +
						"is cancelled before it returns",
					EndsKO: "계좌는 이 단계가 발견한 상태로 되돌아간다 — 남아 있는 주문은 반환 전에 전부 취소된다",
				},
			},
		},
		{
			ID:      StepIdempotencyTTLEdge,
			Title:   "Idempotency validity window edge",
			Proves:  "문서상 10분인 키 수명이 실제 수명인지 — 재생 안전 마진의 입력값",
			Tasks:   []string{"2.7"},
			Mutates: true,
			OptIn:   FlagIncludeTTLEdge,
			Procedure: []string{
				"경고: 이 단계는 의도적으로 두 번째 라이브 주문을 만든다. 창이 닫히는 것을 관측할 유일한 방법이다.",
				"키 K로 최소 수량 LIMIT 매수를 접수한다",
				"문서상 창을 넘겨 기다린 뒤 같은 키 K로 동일한 본문을 재전송한다",
				"새 orderId가 오면 창이 닫힌 것이고, 같은 것이 오면 아직 열려 있는 것이다",
				"주문 두 건을 모두 취소한다",
			},
			Mutations: []StepMutation{
				{
					Kind: MutatePlaceOrder, Side: "buy", Quantity: QuantityOne, Pricing: PriceFarBuy,
					Ends:   "cancelled inside this step, together with the second order",
					EndsKO: "두 번째 주문과 함께 이 단계 안에서 취소된다",
				},
				{
					Kind: MutateReplayOrder, Side: "buy", Quantity: QuantityOne, Pricing: PriceIdenticalBody,
					Ends: "if the window has closed this creates a SECOND live order — the observation this " +
						"opt-in step exists for — and both are cancelled inside this step",
					EndsKO: "창이 닫혀 있으면 이것이 두 번째 라이브 주문을 만든다 — 이 옵트인 단계가 존재하는 " +
						"이유인 관측이다 — 그리고 둘 다 이 단계 안에서 취소된다",
					Note: "this is the only place the tool plans two live orders at once, and it is the only " +
						"step that does not run by default",
					NoteKO: "도구가 라이브 주문 두 건을 동시에 계획하는 유일한 지점이고, 기본 실행되지 않는 " +
						"유일한 단계다",
				},
				{
					Kind:   MutateCancelOrder,
					Ends:   "both orders are cancelled before the step returns",
					EndsKO: "단계가 반환되기 전에 두 주문 모두 취소된다",
				},
			},
		},
		{
			ID:      StepOrderCancel,
			Title:   "KR place and cancel",
			Proves:  "이 계좌에서 POST /api/v1/orders와 POST /api/v1/orders/{id}/cancel이 동작하는지, 취소 후 주문이 무엇으로 읽히는지",
			Tasks:   []string{"2.2", "2.1"},
			Mutates: true,
			Procedure: []string{
				"최소 수량 LIMIT 매수 1건을 시장에서 멀리 접수한다",
				"주문을 다시 읽어 status를 기록한다",
				"취소한다",
				"다시 읽어 취소 후 status와 canceledAt을 기록한다",
			},
			Mutations: []StepMutation{
				{
					Kind: MutatePlaceOrder, Side: "buy", Quantity: QuantityOne, Pricing: PriceFarBuy,
					Ends:   "cancelled inside this step",
					EndsKO: "이 단계 안에서 취소된다",
				},
				{
					Kind:   MutateCancelOrder,
					Ends:   "the account returns to how this step found it — the cancel is what this step measures",
					EndsKO: "계좌는 이 단계가 발견한 상태로 되돌아간다 — 이 취소 자체가 측정 대상이다",
				},
			},
		},
		{
			ID:      StepOrderAmend,
			Title:   "KR amend",
			Proves:  "KR 주문에 대한 POST /api/v1/orders/{id}/modify, 정정이 새 식별자를 발급하는지, 원래 ID가 무엇으로 읽히는지",
			Tasks:   []string{"2.2", "2.1"},
			Mutates: true,
			Procedure: []string{
				"최소 수량 LIMIT 매수 1건을 시장에서 멀리 접수한다",
				"가격(그리고 KR이 요구하는 수량)을 시장에서 한 호가 더 멀리 정정한다",
				"응답에 새 orderId가 실리는지, 원래 ID가 이제 무엇으로 읽히는지 기록한다",
				"살아 있는 쪽 ID를 취소한다",
			},
			Mutations: []StepMutation{
				{
					Kind: MutatePlaceOrder, Side: "buy", Quantity: QuantityOne, Pricing: PriceFarBuy,
					Ends:   "amended once, then cancelled inside this step",
					EndsKO: "한 번 정정한 뒤 이 단계 안에서 취소된다",
				},
				{
					Kind: MutateAmendOrder, Quantity: QuantityOne, Pricing: PriceOneTickFurther,
					Ends:   "whichever identifier is live after the amend is cancelled inside this step",
					EndsKO: "정정 후 살아 있는 쪽 식별자가 이 단계 안에서 취소된다",
					Note:   "KR requires a quantity on a modify, so the amend re-states the same single share",
					NoteKO: "KR은 정정에 수량을 요구하므로 같은 1주를 그대로 다시 적는다",
				},
				{
					Kind:   MutateCancelOrder,
					Ends:   "the account returns to how this step found it",
					EndsKO: "계좌는 이 단계가 발견한 상태로 되돌아간다",
				},
			},
		},
		{
			ID:           StepSellBoundary,
			Title:        "Sell boundaries: partial, full, over-holding",
			Proves:       "보유의 일부·전부·초과를 매도할 때 브로커가 어떻게 반응하는지 — 청산 예약 공식의 입력",
			Tasks:        []string{"2.2", "2.8"},
			Mutates:      true,
			NeedsHolding: true,
			Procedure: []string{
				"보유 종목에 대해 최소 수량 LIMIT 매도 1건을 시장보다 한참 위에 접수한다",
				"그 주문이 남아 있는 동안 sellableQuantity를 읽어 예약분을 기록한다",
				"--max-sell-quantity 안이라면 매도가능 전량에 대한 매도를 접수한다",
				"보유보다 많은 수량의 매도를 제출하고 거부를 기록한다 — 브로커가 받아들이면 즉시 취소하고 단계는 크게 실패한다",
				"이 단계가 접수한 미체결 매도를 전부 취소한다",
			},
			Mutations: []StepMutation{
				{
					Kind: MutatePlaceOrder, Side: "sell", Quantity: QuantityPartial, Pricing: PriceFarSell,
					Ends:   "cancelled before the next boundary is probed",
					EndsKO: "다음 경계를 시험하기 전에 취소된다",
					Note:   "the partial-sell boundary: one share out of a larger holding",
					NoteKO: "부분 매도 경계: 더 큰 보유 중 1주",
				},
				{
					Kind: MutatePlaceOrder, Side: "sell", Quantity: QuantityWholeHolding, Pricing: PriceFarSell,
					Ends:   "cancelled before the next boundary is probed",
					EndsKO: "다음 경계를 시험하기 전에 취소된다",
					Note: "the whole-holding boundary. It is a resting order for the entire position, so it is " +
						"planned only while that position is within --max-sell-quantity",
					NoteKO: "전량 경계. 포지션 전체에 대한 미체결 주문이므로 그 포지션이 --max-sell-quantity " +
						"안에 있을 때만 계획된다",
				},
				{
					Kind: MutatePlaceOrder, Side: "sell", Quantity: QuantityOverHolding, Pricing: PriceFarSell,
					Ends: "expected to be refused; if the broker accepts it, it is cancelled immediately and " +
						"the step fails",
					EndsKO: "거부될 것으로 예상한다. 브로커가 받아들이면 즉시 취소하고 단계는 실패한다",
					Note:   "the smallest possible oversell — exactly one share more than the holding",
					NoteKO: "가능한 가장 작은 초과 매도 — 보유보다 정확히 1주 많다",
				},
				{
					Kind:   MutateCancelOrder,
					Ends:   "every sell this step placed is cancelled before it returns",
					EndsKO: "이 단계가 접수한 매도는 반환 전에 전부 취소된다",
				},
			},
		},
		{
			ID:           StepConditionalRegister,
			Title:        "Conditional order: register and query",
			Proves:       "보유 종목에 SINGLE 손절을 등록하고 id와 status 필터로 다시 읽을 수 있는지",
			Tasks:        []string{"2.5"},
			Mutates:      true,
			NeedsHolding: true,
			Procedure: []string{
				"SINGLE MARKET SELL 손절을 최소 수량으로, 발동가는 시장보다 한참 아래에 등록한다",
				"조건주문 멱등성 측정을 위해 같은 clientOrderId로 동일한 본문을 한 번 재전송한다 (task 2.7)",
				"GET /api/v1/conditional-orders/{id}로 status와 leg 모양을 기록한다",
				"GET /api/v1/conditional-orders?status=WATCHING으로 새 조건주문이 목록에 뜨는지 기록한다",
				"이 조건주문은 의도적으로 등록된 채 남긴다 — 존속 단계가 이 프로세스보다 오래 살아남은 것을 " +
					"읽어야 하기 때문이다. 아래 conditional-cancel 단계가 취소한다. 그 전에 멈추면 " +
					"`tossctl verify status`가 id를 출력한다",
			},
			Mutations: []StepMutation{
				{
					Kind: MutateRegisterConditional, Side: "sell", Quantity: QuantityOne, Pricing: PriceFarStop,
					Ends: "LEFT REGISTERED ON PURPOSE. This is the one exposure that outlives the step and the " +
						"process: the persistence measurement is whether the broker still holds it after this " +
						"process exits, so the conditional-cancel step of the resumed run is what removes it. " +
						"Until then `tossctl verify status` prints its id",
					EndsKO: "의도적으로 등록된 채 남긴다. 단계와 프로세스보다 오래 사는 유일한 노출이다 — " +
						"이 프로세스가 종료된 뒤에도 브로커가 들고 있는지가 존속 측정이므로, " +
						"이어하기 실행의 conditional-cancel 단계가 이것을 제거한다. " +
						"그 전까지는 `tossctl verify status`가 id를 출력한다",
					Note:   "a SINGLE MARKET SELL stop for one share, trigger far below the market so it cannot fire",
					NoteKO: "1주짜리 SINGLE MARKET SELL 손절. 발동가가 시장보다 한참 아래라 발동할 수 없다",
				},
				{
					Kind: MutateReplayConditional, Side: "sell", Quantity: QuantityOne, Pricing: PriceIdenticalBody,
					Ends: "the documented answer is the same conditional coming back; a duplicate is cancelled " +
						"immediately inside this step",
					EndsKO: "문서상 정답은 같은 조건주문이 돌아오는 것이다. 중복이 생기면 이 단계 안에서 즉시 취소된다",
				},
				{
					Kind: MutateCancelConditional,
					Ends: "used only if the replay above produced a duplicate; a duplicate is cancelled at once " +
						"rather than kept",
					EndsKO: "위 재전송이 중복을 만들었을 때만 쓴다. 중복은 보관하지 않고 즉시 취소한다",
				},
			},
		},
		{
			ID:        StepSellableReserved,
			Title:     "Sellable quantity with a conditional reservation live",
			Proves:    "등록된 조건부 매도가 sellableQuantity에서 주식을 예약하는지 — 청산 예약 공식에서 빠져 있는 항",
			Tasks:     []string{"2.8", "2.5"},
			DependsOn: []StepID{StepConditionalRegister},
			Procedure: []string{
				"조건주문을 등록한 종목에 대해 GET /api/v1/sellable-quantity",
				"아무것도 접수하기 전에 기록한 기준선과 비교한다",
			},
		},
		{
			ID:        StepConditionalPersist,
			Title:     "Conditional order survives this process exiting",
			Proves:    "보호를 들고 있는 것이 이 프로세스가 아니라 브로커라는 것 — 무인 설계 전체가 얹혀 있는 전제",
			Tasks:     []string{"2.5"},
			DependsOn: []StepID{StepConditionalRegister},
			Procedure: []string{
				"이 단계는 조건주문을 등록한 프로세스 안에서는 통과할 수 없다",
				"실행은 여기서 멈추고 새 프로세스로 --resume 하라고 알린다 (콘솔에서는 [콘솔 재시작] 버튼)",
				"새 프로세스가 조건주문을 다시 읽는다. 여전히 있으면 통과다",
			},
		},
		{
			ID:        StepConditionalTrigger,
			Title:     "Trigger observation and triggeredOrderId latency",
			Proves:    "발동 후 생성된 주문이 얼마 만에 보이는지, triggeredOrderId가 둘을 잇는지",
			Tasks:     []string{"2.5"},
			Deferred:  "별도 세션 — 시장 조건 필요: 체결될 의도의 주문 없이는 발동을 임의로 만들 수 없고, 이 도구는 그런 주문을 내지 않는다",
			DependsOn: []StepID{StepConditionalRegister},
			Procedure: []string{
				"미검증으로 기록해 task 2.6이 자동 진입 금지 시장·유형 목록에 넣도록 한다",
			},
		},
		{
			ID:                StepConditionalModify,
			Title:             "Conditional modify issues a new id",
			Proves:            "정정이 원자적인지 아니면 취소 후 재생성인지 — openapi는 기존 id가 무효화되고 새 id가 발급된다고 적고 있고, 2.6은 그것을 인용이 아니라 실측으로 요구한다",
			Tasks:             []string{"2.5"},
			Mutates:           true,
			ActsOnConditional: true,
			DependsOn:         []StepID{StepConditionalRegister},
			Procedure: []string{
				"조건주문의 발동가를 한 호가 정정한다",
				"응답이 돌려주는 식별자를 기록한다",
				"기존 id를 다시 읽어 사라졌는지 기록한다",
				"새 id를 다시 읽어 status를 기록한다",
				"정정 후 살아 있는 쪽은 아래 conditional-cancel 단계가 취소한다",
			},
			Mutations: []StepMutation{
				{
					Kind: MutateModifyConditional, Side: "sell", Quantity: QuantityOne,
					Pricing: PriceOneTickFurther,
					Ends: "whichever identifier is live after the modify is cancelled by the conditional-cancel " +
						"step below; the protection is not added to, it is moved",
					EndsKO: "정정 후 살아 있는 쪽 식별자는 아래 conditional-cancel 단계가 취소한다. " +
						"보호가 늘어나는 것이 아니라 옮겨진다",
				},
			},
		},
		{
			ID:                StepConditionalCancel,
			Title:             "Conditional cancel",
			Proves:            "DELETE /api/v1/conditional-orders/{id}가 보호를 제거하고, 계좌가 발견 당시 상태로 남는지",
			Tasks:             []string{"2.5"},
			Mutates:           true,
			ActsOnConditional: true,
			DependsOn:         []StepID{StepConditionalRegister},
			Procedure: []string{
				"살아 있는 조건주문을 취소한다",
				"다시 읽어 사라졌음을 기록한다",
			},
			Mutations: []StepMutation{
				{
					Kind: MutateCancelConditional,
					Ends: "the account is left exactly as this verification found it — this is the step that " +
						"removes the conditional the register step deliberately left alive",
					EndsKO: "계좌는 이 검증이 발견한 그대로 남는다 — 등록 단계가 의도적으로 살려 둔 조건주문을 " +
						"제거하는 것이 이 단계다",
				},
			},
		},
		{
			ID:     StepCosts,
			Title:  "Realised cost collection",
			Proves: "이 계좌에 실제로 부과되는 수수료와 세금 — 또는 정직하게 말해, 아무것도 체결되지 않아 비용 모델에 실측 입력이 아직 없다는 사실",
			Tasks:  []string{"2.9"},
			Procedure: []string{
				"이 검증이 만든 모든 주문을 다시 읽는다",
				"각각의 execution.commission, execution.tax, execution.filledQuantity를 기록한다",
				"체결되지 않은 주문은 기여분이 없고 그대로 기록된다 — 시장에서 먼 LIMIT은 체결되지 않도록 설계되어 있다",
			},
		},
	}
}

// StepByID returns one step of the catalogue.
func StepByID(id StepID) (Step, bool) {
	for _, s := range Steps() {
		if s.ID == id {
			return s, true
		}
	}
	return Step{}, false
}

// Broker is everything the procedure needs from the Open API.
//
// *official.Client satisfies it as it stands. The interface exists so the tests
// can drive the whole procedure without HTTP at all, and so the mutation surface
// is enumerable in one place: six methods can change the account, and
// static_test.go asserts that no file outside mutate.go calls one of them.
type Broker interface {
	Accounts(ctx context.Context) ([]domain.Account, error)
	Holdings(ctx context.Context, symbol string) ([]domain.Position, error)
	SellableQuantity(ctx context.Context, symbol string) (domain.SellableQuantity, error)
	Prices(ctx context.Context, symbols []string) ([]domain.Quote, error)
	PriceLimits(ctx context.Context, symbol string) (domain.PriceLimits, error)
	OrdersPageRaw(ctx context.Context, filter official.OrdersFilter, cursor string) (official.RawOrderPage, error)
	OrderRawByID(ctx context.Context, orderID string) (json.RawMessage, error)
	ConditionalOrders(ctx context.Context, status, symbol, cursor string, limit int) (domain.ConditionalOrderList, error)
	ConditionalOrder(ctx context.Context, id string) (domain.ConditionalOrder, error)

	// --- the mutation surface ---

	PlaceOrder(ctx context.Context, intent orderintent.PlaceIntent) (trading.MutationResult, error)
	CancelOrder(ctx context.Context, orderID string) (trading.MutationResult, error)
	ModifyOrder(ctx context.Context, intent orderintent.AmendIntent) (trading.MutationResult, error)
	CreateConditionalOrder(ctx context.Context, body official.ConditionalCreateBody) (domain.ConditionalOrderRef, error)
	ModifyConditionalOrderRef(ctx context.Context, id string, body official.ConditionalModifyBody) (domain.ConditionalOrderRef, error)
	CancelConditionalOrder(ctx context.Context, id string) error
}

// MutationMethods are the Broker methods that can change a live account.
//
// The list is data because static_test.go greps for it: a seventh mutation added
// to the interface without being added here fails that test, which is the point.
func MutationMethods() []string {
	return []string{
		"PlaceOrder",
		"CancelOrder",
		"ModifyOrder",
		"CreateConditionalOrder",
		"ModifyConditionalOrderRef",
		"CancelConditionalOrder",
	}
}
