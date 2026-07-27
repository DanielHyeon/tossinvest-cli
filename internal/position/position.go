package position

import (
	"errors"
	"fmt"

	"github.com/JungHoonGhae/tossinvest-cli/internal/riskcalc"
)

// ErrInvalidEvent means an event could not be read at all: a side that is not a
// side, a quantity that is not a decimal, a negative delta. It is a programming
// error in the caller, not a broker disagreement, and it is deliberately not a
// refusal — a refusal is a judgement about a real event, and this is the absence
// of one.
var ErrInvalidEvent = errors.New("position: the event cannot be read")

// State is one of design D7's six position states. The schema's CHECK
// constraint pins the same six (internal/journal/core_domain.go), so no other
// value can reach this package from storage.
type State string

const (
	// Flat is "no instance". The projection never writes a FLAT row: a position
	// row is created by the fill that first gives it a quantity, and the absence
	// of a row is what FLAT means to a reader.
	Flat    State = "FLAT"
	Opening State = "OPENING"
	Open    State = "OPEN"
	Scaling State = "SCALING"
	Closing State = "CLOSING"
	Closed  State = "CLOSED"
)

// ValidState reports whether s is one of the six.
func ValidState(s State) bool {
	switch s {
	case Flat, Opening, Open, Scaling, Closing, Closed:
		return true
	default:
		return false
	}
}

// Role is the order's part in the position, re-derived from the intent's side.
// Long-only: a BUY raises exposure and a SELL reduces it.
type Role string

const (
	Entry Role = "ENTRY"
	Exit  Role = "EXIT"
)

// RoleForSide maps an intent side to a role. Anything that is not a side this
// build issues is refused rather than defaulted: a fill projected in the wrong
// direction is worse than a fill not projected at all.
func RoleForSide(side string) (Role, error) {
	switch side {
	case "BUY", "buy", "Buy":
		return Entry, nil
	case "SELL", "sell", "Sell":
		return Exit, nil
	default:
		return "", fmt.Errorf("%w: %q is not an intent side this build issues", ErrInvalidEvent, side)
	}
}

// Movement is what the observation's delta does to the instance's quantity. See
// the package comment for why the classification is against the quantity rather
// than against the delta alone.
type Movement string

const (
	MoveNone       Movement = "NONE"
	MoveAdds       Movement = "ADDS"
	MoveReduces    Movement = "REDUCES"
	MoveFlattens   Movement = "FLATTENS"
	MoveOvershoots Movement = "OVERSHOOTS"
)

// Disposition is what the filling order can still do. It is the dimension
// lineage enters through.
type Disposition string

const (
	Working   Disposition = "WORKING"
	Done      Disposition = "DONE"
	Succeeded Disposition = "SUCCEEDED"
)

// Refusal names a transition the table does not allow. The caller turns a
// refusal into the durable RECONCILE state; the projection itself changes
// nothing, because the alternative is inventing a quantity nobody observed.
type Refusal string

const (
	// RefusalNone is the absence of a refusal.
	RefusalNone Refusal = ""
	// RefusalUnattributedSell — a sell filled against a symbol the projection
	// holds nothing of, and no instance has ever been opened or the last one is
	// long gone. 매도 체결 귀속 실패.
	RefusalUnattributedSell Refusal = "UNATTRIBUTED_SELL"
	// RefusalSellOnClosed — a sell filled against an instance that is already
	// CLOSED. CLOSED 종결성: a closed instance is not reopened by a late sell.
	RefusalSellOnClosed Refusal = "SELL_ON_CLOSED"
	// RefusalOversell — the sell is larger than what is held.
	RefusalOversell Refusal = "OVERSELL"
	// RefusalEntryWhileClosing — a buy filled while an exit order is working.
	RefusalEntryWhileClosing Refusal = "ENTRY_WHILE_CLOSING"
)

// Instance is the projected position as it stands before the event.
//
// The zero value is a FLAT instance with no quantity and no cost basis, which is
// what a symbol nothing has been traded on looks like.
type Instance struct {
	State State
	// Quantity is a decimal string; "" reads as "0", because "no row" and "zero
	// quantity" are the same position.
	Quantity string
	// AvgPrice is the average acquisition cost as a decimal string, or "" when
	// it is not known. See the package comment: "" is never 0.
	AvgPrice string
}

// Event is one observation offered to the projection.
//
// It carries the order's own numbers rather than the position's, because the
// disposition and the cost contribution are both properties of the order. The
// previous pair is what makes the cost exact: the position's share of an order
// is `filled × average`, and the contribution of this observation is the change
// in that product.
type Event struct {
	Role Role
	// Delta is the newly filled quantity, unsigned, as a decimal string. "0"
	// for a correction or for a terminal observation that filled nothing — both
	// reach the apply hook (issues.md, task 0.3).
	Delta string
	// OrderQuantity is the order's ordered quantity — 원주문 수량. "" means the
	// ordered quantity is unknown, in which case completion cannot be judged and
	// the order counts as still working until the broker calls it terminal.
	OrderQuantity string
	// OrderFilled is the order's cumulative filled quantity after this
	// observation, PrevOrderFilled the same before it.
	OrderFilled     string
	PrevOrderFilled string
	// OrderAvgPrice is the order's average filled price after this observation,
	// PrevOrderAvgPrice the same before it. "" is "the broker reported none".
	OrderAvgPrice     string
	PrevOrderAvgPrice string
	// Terminal is the caller's derived verdict that the order can no longer
	// change at the broker (internal/brokerstate). Never a guess made here.
	Terminal bool
	// HasSuccessor reports a `replaces` edge leaving this order: an amendment
	// created a child that carries the remainder (internal/journal/lineage.go).
	HasSuccessor bool
}

// Outcome is what one event does to one instance.
type Outcome struct {
	// Row is the transition table row that decided this, e.g. "E23". It is
	// carried so an operator reading a projection change can find the rule that
	// produced it, and so the tests can assert the table was consulted rather
	// than re-derived.
	Row         string
	Movement    Movement
	Disposition Disposition
	// Next is the state after the event. On a refusal it is the state before —
	// a refused transition transitions nothing.
	Next State
	// Quantity and AvgPrice are the instance after the event, unchanged on a
	// refusal.
	Quantity string
	AvgPrice string
	// NewInstance reports that this event opens a fresh position instance: the
	// caller mints the next instance_seq and a new entry_decision_id rather than
	// writing into the row that is already there.
	NewInstance bool
	// Refusal is empty for an allowed transition.
	Refusal Refusal
	// Reason is the human sentence for the refusal, empty otherwise.
	Reason string
}

// Reconcile reports whether the outcome is a refusal that must raise the
// durable RECONCILE state.
func (o Outcome) Reconcile() bool { return o.Refusal != RefusalNone }

// Closed reports whether the instance reached its terminal state.
func (o Outcome) Closed() bool { return o.Next == Closed && !o.Reconcile() }

// Apply runs one event against one instance and returns what the transition
// table says happens.
//
// It never panics and it never corrects arithmetic. A transition the table does
// not allow comes back as a refusal with the instance untouched, because the
// account — not this package — is the authority on a quantity the projection
// cannot derive (position-ledger: 허용되지 않은 전이는 오류이며 RECONCILE로
// 전이한다, 산식 보정 금지).
func Apply(inst Instance, ev Event) (Outcome, error) {
	state := inst.State
	if state == "" {
		state = Flat
	}
	if !ValidState(state) {
		return Outcome{}, fmt.Errorf("%w: %q is not a position state", ErrInvalidEvent, inst.State)
	}
	if ev.Role != Entry && ev.Role != Exit {
		return Outcome{}, fmt.Errorf("%w: %q is not an order role", ErrInvalidEvent, ev.Role)
	}

	quantity, err := nonNegative("instance quantity", orZero(inst.Quantity))
	if err != nil {
		return Outcome{}, err
	}
	delta, err := nonNegative("fill delta", orZero(ev.Delta))
	if err != nil {
		return Outcome{}, err
	}

	movement, err := classifyMovement(ev.Role, delta, quantity)
	if err != nil {
		return Outcome{}, err
	}
	disposition, err := classifyDisposition(ev)
	if err != nil {
		return Outcome{}, err
	}

	row, found := Lookup(state, ev.Role, movement, disposition)
	if !found {
		// Structurally unreachable: the table has a row for every combination
		// classifyMovement and classifyDisposition can produce, and
		// TestTheTableCoversEveryReachableCombination proves it. Reported rather
		// than assumed away, because an unhandled state must not become a silent
		// no-op on a live account.
		return Outcome{}, fmt.Errorf(
			"%w: no transition row for (%s, %s, %s, %s)",
			ErrInvalidEvent, state, ev.Role, movement, disposition)
	}

	out := Outcome{
		Row:         row.ID,
		Movement:    movement,
		Disposition: disposition,
		Next:        state,
		Quantity:    quantity,
		AvgPrice:    inst.AvgPrice,
	}
	if row.Refusal != RefusalNone {
		out.Refusal = row.Refusal
		out.Reason = refusalReason(row.Refusal, ev.Role, delta, quantity, state)
		return out, nil
	}

	base := Instance{State: state, Quantity: quantity, AvgPrice: inst.AvgPrice}
	if row.NewInstance {
		// A re-entry starts its own cost basis. Carrying the closed instance's
		// average forward would price the new position with the old one's trades.
		base = Instance{State: Flat, Quantity: "0", AvgPrice: Unknown}
	}

	out.Next = row.Next
	out.NewInstance = row.NewInstance

	switch ev.Role {
	case Entry:
		next, err := riskcalc.AddDecimal(base.Quantity, delta)
		if err != nil {
			return Outcome{}, fmt.Errorf("%w: adding the fill delta: %v", ErrInvalidEvent, err)
		}
		avg, err := reaverage(base, next, ev)
		if err != nil {
			return Outcome{}, err
		}
		out.Quantity, out.AvgPrice = next, avg
	case Exit:
		next, err := riskcalc.SubDecimal(base.Quantity, delta)
		if err != nil {
			return Outcome{}, fmt.Errorf("%w: subtracting the fill delta: %v", ErrInvalidEvent, err)
		}
		// A sell realises P&L; it does not move the unit cost of what is left.
		out.Quantity, out.AvgPrice = next, base.AvgPrice
	}
	return out, nil
}

// classifyMovement turns the delta into what it does to the held quantity.
func classifyMovement(role Role, delta, quantity string) (Movement, error) {
	zero, err := riskcalc.CompareDecimal(delta, "0")
	if err != nil {
		return "", fmt.Errorf("%w: comparing the delta with zero: %v", ErrInvalidEvent, err)
	}
	if zero == 0 {
		return MoveNone, nil
	}
	if role == Entry {
		return MoveAdds, nil
	}
	cmp, err := riskcalc.CompareDecimal(delta, quantity)
	if err != nil {
		return "", fmt.Errorf("%w: comparing the delta with the held quantity: %v", ErrInvalidEvent, err)
	}
	switch {
	case cmp < 0:
		return MoveReduces, nil
	case cmp == 0:
		return MoveFlattens, nil
	default:
		return MoveOvershoots, nil
	}
}

// classifyDisposition answers "what can this order still do", which is the only
// question lineage is asked.
//
// The order of the branches is the order of the evidence's strength. Reaching
// 원주문 수량 is arithmetic and cannot be revised. A replace edge is a recorded
// fact about where the remainder went, so it outranks a terminal state that says
// only that *this* order is over. Terminal without an edge means nothing carries
// the remainder, so the job is over too.
func classifyDisposition(ev Event) (Disposition, error) {
	if q := orEmpty(ev.OrderQuantity); q != "" {
		cmp, err := riskcalc.CompareDecimal(orZero(ev.OrderFilled), q)
		if err != nil {
			return "", fmt.Errorf("%w: comparing the order's fill with its quantity: %v", ErrInvalidEvent, err)
		}
		if cmp >= 0 {
			return Done, nil
		}
	}
	switch {
	case ev.HasSuccessor:
		return Succeeded, nil
	case ev.Terminal:
		return Done, nil
	default:
		return Working, nil
	}
}

func refusalReason(r Refusal, role Role, delta, quantity string, state State) string {
	switch r {
	case RefusalUnattributedSell:
		return fmt.Sprintf(
			"a sell of %s filled against a symbol the projection holds none of; "+
				"the account is the authority on what is held", delta)
	case RefusalSellOnClosed:
		return fmt.Sprintf(
			"a sell of %s filled against an instance that is already CLOSED; "+
				"a closed instance is final and is not reopened by a late sell", delta)
	case RefusalOversell:
		return fmt.Sprintf(
			"a sell of %s filled against a held quantity of %s; the projection will not "+
				"invent the difference", delta, quantity)
	case RefusalEntryWhileClosing:
		return fmt.Sprintf(
			"a buy of %s filled while an exit order was working (%s); "+
				"the position cannot be growing and shrinking on one instruction set", delta, state)
	default:
		return ""
	}
}

func nonNegative(field, value string) (string, error) {
	canonical, err := riskcalc.CanonicalDecimal(value)
	if err != nil {
		return "", fmt.Errorf("%w: %s %q: %v", ErrInvalidEvent, field, value, err)
	}
	negative, err := riskcalc.IsNegativeDecimal(canonical)
	if err != nil {
		return "", fmt.Errorf("%w: %s %q: %v", ErrInvalidEvent, field, value, err)
	}
	if negative {
		return "", fmt.Errorf("%w: %s is %s; quantities in a projection are unsigned and the "+
			"direction comes from the intent", ErrInvalidEvent, field, canonical)
	}
	return canonical, nil
}

// orZero reads an absent quantity as zero. An absent *price* is not zero and is
// handled by Unknown in decimal.go; the asymmetry is the point.
func orZero(s string) string {
	if orEmpty(s) == "" {
		return "0"
	}
	return s
}

func orEmpty(s string) string {
	for _, r := range s {
		if r != ' ' && r != '\t' && r != '\n' && r != '\r' {
			return s
		}
	}
	return ""
}
