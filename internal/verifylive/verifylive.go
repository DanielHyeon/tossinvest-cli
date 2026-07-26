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
			Proves: "which of the ten documented OrderStatus values this account has actually produced, and what a CANCEL_REJECTED/REPLACE_REJECTED record looks like in the order list",
			Tasks:  []string{"2.1"},
			Procedure: []string{
				"walk GET /api/v1/orders (all statuses) up to the page cap",
				"record every distinct status, its field shape and its counts — identifiers are digested, never stored",
				"record whether CANCEL_REJECTED / REPLACE_REJECTED appear as their own records and whether they link back to an original order",
			},
		},
		{
			ID:     StepSellableBaseline,
			Title:  "Sellable-quantity baseline",
			Proves: "the resting relationship between holdings.quantity and sellableQuantity before this tool has placed anything",
			Tasks:  []string{"2.8"},
			Procedure: []string{
				"GET /api/v1/holdings",
				"GET /api/v1/sellable-quantity per held symbol",
				"record the difference per symbol as the baseline the later reading is compared against",
			},
		},
		{
			ID:      StepIdempotency,
			Title:   "clientOrderId replay, conflict and scope",
			Proves:  "whether a repeated clientOrderId returns the first order instead of creating a second one, and whether a different body under the same key is refused with idempotency-key-conflict",
			Tasks:   []string{"2.7"},
			Mutates: true,
			Procedure: []string{
				"place ONE minimum-quantity LIMIT buy far from the market under key K",
				"re-send the identical body under the same key K — the identity recovery is the claim under test",
				"assert the same orderId came back and the open-order list did not grow",
				"re-send a DIFFERENT body under the same key K and record the error code",
				"cancel every order this step has resting",
			},
			Mutations: []StepMutation{
				{
					Kind: MutatePlaceOrder, Side: "buy", Quantity: QuantityOne, Pricing: PriceFarBuy,
					Ends: "cancelled inside this step",
				},
				{
					Kind: MutateReplayOrder, Side: "buy", Quantity: QuantityOne, Pricing: PriceIdenticalBody,
					Ends: "the documented answer is the first order coming back and nothing new existing; " +
						"anything it does create is cancelled inside this step",
					Note: "this replay is the claim under test, so it is sent as part of the line above rather " +
						"than as a separate decision",
				},
				{
					Kind: MutateConflictProbe, Side: "buy", Quantity: QuantityOne, Pricing: PriceOneTickFurther,
					Ends: "expected to be refused with idempotency-key-conflict; if the broker accepts it " +
						"instead, the order it creates is cancelled inside this step",
					Note: "deliberately re-uses the key above with a DIFFERENT price — that refusal is the " +
						"measurement",
				},
				{
					Kind: MutateCancelOrder,
					Ends: "the account returns to how this step found it — every order the step has resting " +
						"is cancelled before it returns",
				},
			},
		},
		{
			ID:      StepIdempotencyTTLEdge,
			Title:   "Idempotency validity window edge",
			Proves:  "whether the documented 10-minute key lifetime is the real one — the input to any safety margin on replay",
			Tasks:   []string{"2.7"},
			Mutates: true,
			OptIn:   FlagIncludeTTLEdge,
			Procedure: []string{
				"WARNING: this step deliberately creates a SECOND live order. It is the only way to observe the window closing.",
				"place a minimum-quantity LIMIT buy under key K",
				"wait past the documented window, then re-send the identical body under K",
				"a new orderId means the window has closed; the same one means it is still open",
				"cancel BOTH orders",
			},
			Mutations: []StepMutation{
				{
					Kind: MutatePlaceOrder, Side: "buy", Quantity: QuantityOne, Pricing: PriceFarBuy,
					Ends: "cancelled inside this step, together with the second order",
				},
				{
					Kind: MutateReplayOrder, Side: "buy", Quantity: QuantityOne, Pricing: PriceIdenticalBody,
					Ends: "if the window has closed this creates a SECOND live order — the observation this " +
						"opt-in step exists for — and both are cancelled inside this step",
					Note: "this is the only place the tool plans two live orders at once, and it is the only " +
						"step that does not run by default",
				},
				{
					Kind: MutateCancelOrder,
					Ends: "both orders are cancelled before the step returns",
				},
			},
		},
		{
			ID:      StepOrderCancel,
			Title:   "KR place and cancel",
			Proves:  "POST /api/v1/orders and POST /api/v1/orders/{id}/cancel work on this account, and what the order reads back as afterwards",
			Tasks:   []string{"2.2", "2.1"},
			Mutates: true,
			Procedure: []string{
				"place ONE minimum-quantity LIMIT buy far from the market",
				"read the order back and record its status",
				"cancel it",
				"read it back again and record the post-cancel status and canceledAt",
			},
			Mutations: []StepMutation{
				{
					Kind: MutatePlaceOrder, Side: "buy", Quantity: QuantityOne, Pricing: PriceFarBuy,
					Ends: "cancelled inside this step",
				},
				{
					Kind: MutateCancelOrder,
					Ends: "the account returns to how this step found it — the cancel is what this step measures",
				},
			},
		},
		{
			ID:      StepOrderAmend,
			Title:   "KR amend",
			Proves:  "POST /api/v1/orders/{id}/modify on a KR order, whether the amend issues a new identifier, and what the original reads back as",
			Tasks:   []string{"2.2", "2.1"},
			Mutates: true,
			Procedure: []string{
				"place ONE minimum-quantity LIMIT buy far from the market",
				"amend its price (and quantity, which KR requires) one tick further from the market",
				"record whether the response carries a new orderId and what the original id now reads as",
				"cancel whichever id is live",
			},
			Mutations: []StepMutation{
				{
					Kind: MutatePlaceOrder, Side: "buy", Quantity: QuantityOne, Pricing: PriceFarBuy,
					Ends: "amended once, then cancelled inside this step",
				},
				{
					Kind: MutateAmendOrder, Quantity: QuantityOne, Pricing: PriceOneTickFurther,
					Ends: "whichever identifier is live after the amend is cancelled inside this step",
					Note: "KR requires a quantity on a modify, so the amend re-states the same single share",
				},
				{
					Kind: MutateCancelOrder,
					Ends: "the account returns to how this step found it",
				},
			},
		},
		{
			ID:           StepSellBoundary,
			Title:        "Sell boundaries: partial, full, over-holding",
			Proves:       "what the broker does with a sell for part of a holding, all of it, and more than it — the input to the liquidation reservation formula",
			Tasks:        []string{"2.2", "2.8"},
			Mutates:      true,
			NeedsHolding: true,
			Procedure: []string{
				"place ONE minimum-quantity LIMIT sell far ABOVE the market against a held symbol",
				"read sellableQuantity while it rests and record the reservation",
				"place a sell for the WHOLE sellable holding when that is within --max-sell-quantity",
				"submit a sell for MORE than the holding and record the refusal — if the broker accepts it, it is cancelled at once and the step fails loudly",
				"cancel every resting sell this step placed",
			},
			Mutations: []StepMutation{
				{
					Kind: MutatePlaceOrder, Side: "sell", Quantity: QuantityPartial, Pricing: PriceFarSell,
					Ends: "cancelled before the next boundary is probed",
					Note: "the partial-sell boundary: one share out of a larger holding",
				},
				{
					Kind: MutatePlaceOrder, Side: "sell", Quantity: QuantityWholeHolding, Pricing: PriceFarSell,
					Ends: "cancelled before the next boundary is probed",
					Note: "the whole-holding boundary. It is a resting order for the entire position, so it is " +
						"planned only while that position is within --max-sell-quantity",
				},
				{
					Kind: MutatePlaceOrder, Side: "sell", Quantity: QuantityOverHolding, Pricing: PriceFarSell,
					Ends: "expected to be refused; if the broker accepts it, it is cancelled immediately and " +
						"the step fails",
					Note: "the smallest possible oversell — exactly one share more than the holding",
				},
				{
					Kind: MutateCancelOrder,
					Ends: "every sell this step placed is cancelled before it returns",
				},
			},
		},
		{
			ID:           StepConditionalRegister,
			Title:        "Conditional order: register and query",
			Proves:       "a SINGLE stop can be registered against a holding and read back by id and by status filter",
			Tasks:        []string{"2.5"},
			Mutates:      true,
			NeedsHolding: true,
			Procedure: []string{
				"register a SINGLE MARKET SELL stop, minimum quantity, trigger far BELOW the market",
				"re-send the identical body under the same clientOrderId once, to measure conditional idempotency (task 2.7)",
				"GET /api/v1/conditional-orders/{id} and record the status and leg shape",
				"GET /api/v1/conditional-orders?status=WATCHING and record whether the new one is listed",
				"this conditional is left registered ON PURPOSE — the persistence step needs it to outlive this " +
					"process. The conditional-cancel step below cancels it; if you stop before then, " +
					"`tossctl verify status` prints its id",
			},
			Mutations: []StepMutation{
				{
					Kind: MutateRegisterConditional, Side: "sell", Quantity: QuantityOne, Pricing: PriceFarStop,
					Ends: "LEFT REGISTERED ON PURPOSE. This is the one exposure that outlives the step and the " +
						"process: the persistence measurement is whether the broker still holds it after this " +
						"process exits, so the conditional-cancel step of the resumed run is what removes it. " +
						"Until then `tossctl verify status` prints its id",
					Note: "a SINGLE MARKET SELL stop for one share, trigger far below the market so it cannot fire",
				},
				{
					Kind: MutateReplayConditional, Side: "sell", Quantity: QuantityOne, Pricing: PriceIdenticalBody,
					Ends: "the documented answer is the same conditional coming back; a duplicate is cancelled " +
						"immediately inside this step",
				},
				{
					Kind: MutateCancelConditional,
					Ends: "used only if the replay above produced a duplicate; a duplicate is cancelled at once " +
						"rather than kept",
				},
			},
		},
		{
			ID:        StepSellableReserved,
			Title:     "Sellable quantity with a conditional reservation live",
			Proves:    "whether a registered conditional sell reserves shares in sellableQuantity — the missing term in the liquidation reservation formula",
			Tasks:     []string{"2.8", "2.5"},
			DependsOn: []StepID{StepConditionalRegister},
			Procedure: []string{
				"GET /api/v1/sellable-quantity for the symbol the conditional was registered on",
				"compare against the baseline recorded before anything was placed",
			},
		},
		{
			ID:        StepConditionalPersist,
			Title:     "Conditional order survives this process exiting",
			Proves:    "the broker holds the protection, not this process — the premise the whole unattended design rests on",
			Tasks:     []string{"2.5"},
			DependsOn: []StepID{StepConditionalRegister},
			Procedure: []string{
				"this step cannot pass inside the process that registered the conditional",
				"the run stops here and tells you to start a new one with --resume",
				"the new process reads the conditional back; still present is the pass",
			},
		},
		{
			ID:        StepConditionalTrigger,
			Title:     "Trigger observation and triggeredOrderId latency",
			Proves:    "how long after a trigger the generated order becomes visible, and whether triggeredOrderId links the two",
			Tasks:     []string{"2.5"},
			Deferred:  "별도 세션 — 시장 조건 필요: a trigger cannot be produced on demand without an order that is meant to fill, which this tool never places",
			DependsOn: []StepID{StepConditionalRegister},
			Procedure: []string{
				"recorded as unverified so task 2.6 lists it among the markets and types automatic entry is not allowed on",
			},
		},
		{
			ID:        StepConditionalModify,
			Title:     "Conditional modify issues a new id",
			Proves:    "whether modify is atomic or a cancel-and-recreate — openapi says the old id is invalidated and a new one issued, and 2.6 needs that measured, not quoted",
			Tasks:     []string{"2.5"},
			Mutates:   true,
			DependsOn: []StepID{StepConditionalRegister},
			Procedure: []string{
				"modify the conditional's trigger price by one tick",
				"record the identifier the response returns",
				"read the OLD id back and record whether it is gone",
				"read the NEW id back and record its status",
				"whichever id is live afterwards is cancelled by the conditional-cancel step below",
			},
			Mutations: []StepMutation{
				{
					Kind: MutateModifyConditional, Side: "sell", Quantity: QuantityOne,
					Pricing: PriceOneTickFurther,
					Ends: "whichever identifier is live after the modify is cancelled by the conditional-cancel " +
						"step below; the protection is not added to, it is moved",
				},
			},
		},
		{
			ID:        StepConditionalCancel,
			Title:     "Conditional cancel",
			Proves:    "DELETE /api/v1/conditional-orders/{id} removes the protection, and the account is left as it was found",
			Tasks:     []string{"2.5"},
			Mutates:   true,
			DependsOn: []StepID{StepConditionalRegister},
			Procedure: []string{
				"cancel the live conditional",
				"read it back and record that it is gone",
			},
			Mutations: []StepMutation{
				{
					Kind: MutateCancelConditional,
					Ends: "the account is left exactly as this verification found it — this is the step that " +
						"removes the conditional the register step deliberately left alive",
				},
			},
		},
		{
			ID:     StepCosts,
			Title:  "Realised cost collection",
			Proves: "the commission and tax this account is actually charged — or, honestly, that nothing filled and the cost model has no measured input yet",
			Tasks:  []string{"2.9"},
			Procedure: []string{
				"read back every order this verification created",
				"record execution.commission, execution.tax and execution.filledQuantity for each",
				"an order that never filled contributes nothing and is recorded as such — the far-from-market limits are designed not to fill",
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
