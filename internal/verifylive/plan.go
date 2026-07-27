package verifylive

// plan.go works out, before anything is sent, everything this run can send.
//
// # Why the plan exists
//
// The approval model is one typed confirmation per run (tasks.md 1.5, 사용자 결정
// 2026-07-26), and that model rests entirely on this file being complete. A summary
// that quietly omitted a mutation would be consent obtained for a list the operator
// never saw, which is worse than no consent at all. So the plan is *derived* from
// the three inputs the runner itself walks — the step catalogue, the record, the
// operator's flags — rather than written out by hand alongside them. There is no
// second list to keep in step with the first.
//
// The catalogue carries the mutations as data (Step.Mutations) for the same reason
// the steps themselves are data: an operator can read them with `verify run --list`
// before deciding anything, and a test can assert that every step which mutates
// declares what it will send.
//
// # A class, not an instance
//
// One planned line is one *class* of request: "cancel whatever this step left
// resting" is a single decision, and counting the cancels would make the summary
// depend on how the broker behaved rather than on what the operator is agreeing to.
// The line therefore carries a ceiling (MaxQuantity) rather than an exact number
// wherever the broker's answers decide how many requests it takes.
//
// # Two failure directions, deliberately different
//
//	a step with no plan lines       is skipped, with the reason on the record. It
//	                                never reaches a mutation call at all.
//	a step that would send          stops the whole run. The operator approved a
//	  something outside its lines   list; adapting silently to a symbol that moved
//	                                or a quantity that grew would turn that list
//	                                into a decoration.

import (
	"context"
	"fmt"
	"io"
	"strings"
	"unicode/utf8"

	"github.com/JungHoonGhae/tossinvest-cli/internal/domain"
)

// MutationKind names one class of live request this tool can send.
type MutationKind string

// The classes. Every mutating Broker call in mutate.go declares one of these, and
// nothing may be sent that the approved plan does not carry a matching line for.
const (
	MutatePlaceOrder          MutationKind = "place-order"
	MutateReplayOrder         MutationKind = "replay-order"
	MutateConflictProbe       MutationKind = "conflict-probe"
	MutateCancelOrder         MutationKind = "cancel-order"
	MutateAmendOrder          MutationKind = "amend-order"
	MutateRegisterConditional MutationKind = "register-conditional"
	MutateReplayConditional   MutationKind = "replay-conditional"
	MutateModifyConditional   MutationKind = "modify-conditional"
	MutateCancelConditional   MutationKind = "cancel-conditional"
)

// Verb is how the class reads in a summary line, and — for the classes that have a
// per-mutation prompt — the action that prompt shows. The two are the same string
// on purpose: an operator who switches to --confirm-each should recognise the
// wording from the batch they read the first time.
func (k MutationKind) Verb() string {
	switch k {
	case MutatePlaceOrder:
		return "place a live LIMIT order"
	case MutateReplayOrder:
		return "re-send that identical order under the same clientOrderId"
	case MutateConflictProbe:
		return "probe the idempotency key with a conflicting body"
	case MutateCancelOrder:
		return "cancel a live order"
	case MutateAmendOrder:
		return "amend a live order"
	case MutateRegisterConditional:
		return "register a live conditional order"
	case MutateReplayConditional:
		return "re-send that identical conditional under the same clientOrderId"
	case MutateModifyConditional:
		return "modify a live conditional order"
	case MutateCancelConditional:
		return "cancel a live conditional order"
	default:
		return string(k)
	}
}

// PricingBasis says how a request's price is derived.
//
// It is a rule rather than a number because the price is re-quoted when the step
// actually runs — the last trade moves between the approval and the request, and an
// order priced off a stale quote would be closer to the market than the operator
// agreed to. The rule is what is approved; re-quoting inside it is not a change.
type PricingBasis string

// The bases.
const (
	PriceNone           PricingBasis = ""
	PriceFarBuy         PricingBasis = "far-buy"
	PriceFarSell        PricingBasis = "far-sell"
	PriceFarStop        PricingBasis = "far-stop"
	PriceOneTickFurther PricingBasis = "one-tick-further"
	PriceIdenticalBody  PricingBasis = "identical-body"
)

// bandClause says what bounds the price besides the offset.
//
// KR has a daily price band and the clamp is part of what the operator approves.
// US has none — GET /api/v1/price-limits returns null — so a line that claimed a
// band would be describing a rule that did not run. The KR wording is byte-for-byte
// what it has always been, because Describe's English is what the plan digest is
// taken over and a moved byte there invalidates a resumed run's approval.
func bandClause(market string) string {
	if SameMarket(market, MarketKR) {
		return " and clamped inside the day's price band"
	}
	return " (this market has no daily price band, so the offset alone decides how far the price sits)"
}

func bandClauseKO(market string) string {
	if SameMarket(market, MarketKR) {
		return "당일 가격제한폭 안으로 clamp"
	}
	return "이 시장에는 당일 가격제한폭이 없어 오프셋만으로 거리가 정해진다"
}

// Describe renders the basis with the offset this run is actually using.
func (b PricingBasis) Describe(offset float64, market string) string {
	switch b {
	case PriceFarBuy:
		return fmt.Sprintf("LIMIT, %.0f%% BELOW the last trade at the moment the step runs, snapped to the %s "+
			"tick grid%s — re-quoted by this same rule and never closer "+
			"to the market than it allows", offset*100, market, bandClause(market))
	case PriceFarSell:
		return fmt.Sprintf("LIMIT, %.0f%% ABOVE the last trade at the moment the step runs, snapped to the %s "+
			"tick grid%s — re-quoted by this same rule and never closer "+
			"to the market than it allows", offset*100, market, bandClause(market))
	case PriceFarStop:
		return fmt.Sprintf("MARKET stop whose trigger sits %.0f%% BELOW the last trade at the moment the step "+
			"runs, snapped to the %s tick grid%s, so it cannot fire",
			offset*100, market, bandClause(market))
	case PriceOneTickFurther:
		return "one tick FURTHER from the market than the price above, on the same grid — the smallest change " +
			"the broker will accept as a change"
	case PriceIdenticalBody:
		return "byte-identical to the request above, under the same clientOrderId — a different body would be a " +
			"different order and is not what this measures"
	default:
		return ""
	}
}

// DescribeKO is Describe in Korean: what the approval summary shows.
//
// The English above is what goes into the plan digest and must not move; this is
// the same rule, said in the language the person reading it uses. The two are kept
// in step by TestTheKoreanSummaryDescribesTheSamePlan.
func (b PricingBasis) DescribeKO(offset float64, market string) string {
	switch b {
	case PriceFarBuy:
		return fmt.Sprintf("LIMIT. 단계가 실행되는 시점의 최종체결가보다 %.0f%% 아래, %s 호가 단위에 스냅하고 "+
			"%s — 같은 규칙으로 재호가하며 규칙이 허용하는 것보다 시장에 가까워지지 않는다",
			offset*100, market, bandClauseKO(market))
	case PriceFarSell:
		return fmt.Sprintf("LIMIT. 단계가 실행되는 시점의 최종체결가보다 %.0f%% 위, %s 호가 단위에 스냅하고 "+
			"%s — 같은 규칙으로 재호가하며 규칙이 허용하는 것보다 시장에 가까워지지 않는다",
			offset*100, market, bandClauseKO(market))
	case PriceFarStop:
		return fmt.Sprintf("MARKET 손절. 발동가가 실행 시점 최종체결가보다 %.0f%% 아래이고 %s 호가 단위에 스냅, "+
			"%s — 발동할 수 없는 위치다", offset*100, market, bandClauseKO(market))
	case PriceOneTickFurther:
		return "위 가격보다 시장에서 한 호가 더 먼 값(같은 호가 단위) — 브로커가 '변경'으로 받아주는 최소 변화"
	case PriceIdenticalBody:
		return "위 요청과 바이트 단위로 동일한 본문, 동일한 clientOrderId — 본문이 다르면 다른 주문이고 " +
			"그것은 이 단계가 측정하는 대상이 아니다"
	default:
		return ""
	}
}

// QuantityRule says how many shares a request carries.
//
// Every rule but one resolves to a single share. The exceptions belong to the
// sell-side boundary step, whose whole subject is what the broker does at the edges
// of a holding, and they are resolved against the account before the operator is
// asked rather than described in the abstract.
type QuantityRule string

// The rules.
const (
	// QuantityNone is a request that carries no quantity at all: a cancel.
	QuantityNone QuantityRule = ""
	// QuantityOne is one share, which is what every probe order is.
	QuantityOne QuantityRule = "one"
	// QuantityPartial is one share out of a larger holding. It is not planned at
	// all on a one-share holding, where a partial sell and a whole-holding sell
	// would be the same order.
	QuantityPartial QuantityRule = "partial"
	// QuantityWholeHolding is the entire sellable holding, and is not planned when
	// that is more than --max-sell-quantity.
	QuantityWholeHolding QuantityRule = "whole-holding"
	// QuantityOverHolding is one share MORE than the holding: the smallest possible
	// oversell, which is what makes it safe to ask the question.
	QuantityOverHolding QuantityRule = "over-holding"
)

// StepMutation is one class of request a catalogue step declares it will send.
type StepMutation struct {
	// Kind is the class.
	Kind MutationKind `json:"kind"`
	// Side is "buy" or "sell", empty for a request that carries neither.
	Side string `json:"side,omitempty"`
	// Quantity is the rule, resolved against the account when the plan is built.
	Quantity QuantityRule `json:"quantity,omitempty"`
	// Pricing is how the price is derived.
	Pricing PricingBasis `json:"pricing,omitempty"`
	// Ends says how the exposure this creates goes away. A declared mutation with
	// no answer here is one this tool should not be making.
	Ends string `json:"ends"`
	// Note is anything else the operator has to know before typing.
	Note string `json:"note,omitempty"`

	// EndsKO and NoteKO are what the operator actually reads (task 1.8 ③).
	//
	// They are carried alongside the English rather than instead of it, and they
	// are json:"-" so they never reach a serialiser. That is not tidiness: the
	// English text above is hashed into approval.plan_digest, the record's proof
	// that a person was shown a list of exactly this shape. Translating the hashed
	// field would give every future run a different digest from the ones already
	// on the operator's disk, and "the same plan was re-approved across a restart"
	// — measured on 2026-07-26 (measurements.md M3) — would stop being checkable.
	EndsKO string `json:"-"`
	NoteKO string `json:"-"`
}

// PlannedMutation is one line of the summary: a StepMutation resolved against this
// run's symbols, quantities and offset.
type PlannedMutation struct {
	// Ordinal is the line's number in the printed summary, from 1.
	Ordinal int `json:"ordinal"`
	// Step is the step that will send it.
	Step StepID `json:"step"`
	// Kind is the class.
	Kind MutationKind `json:"kind"`
	// Symbol, Side and Quantity are the resolved request.
	Symbol   string `json:"symbol,omitempty"`
	Side     string `json:"side,omitempty"`
	Quantity string `json:"quantity,omitempty"`
	// MaxQuantity is the ceiling this line authorises. A request for more than it
	// is outside the approved batch and stops the run.
	MaxQuantity float64 `json:"max_quantity,omitempty"`
	// Pricing is the resolved price basis.
	Pricing string `json:"pricing,omitempty"`
	// Ends and Note come from the catalogue.
	Ends string `json:"ends"`
	Note string `json:"note,omitempty"`

	// The Korean the summary is rendered from. json:"-" for the reason given on
	// StepMutation: everything above this line is hashed into the plan digest, and
	// the digest has to keep meaning the same thing across builds.
	EndsKO    string `json:"-"`
	NoteKO    string `json:"-"`
	PricingKO string `json:"-"`
}

// Headline is the one-line identity of the request, in the original language. It
// is what the per-mutation prompt and the package's own tests read.
func (m PlannedMutation) Headline() string {
	what := m.shape(m.Quantity)
	if len(what) == 0 {
		return m.Kind.Verb()
	}
	return m.Kind.Verb() + " — " + strings.Join(what, " ")
}

// HeadlineKO is Headline in the display language.
func (m PlannedMutation) HeadlineKO() string {
	what := m.shape(m.quantityKO())
	if len(what) == 0 {
		return m.Kind.VerbKO()
	}
	return m.Kind.VerbKO() + " — " + strings.Join(what, " ")
}

// quantityKO renders the share count in the display language.
//
// It is derived from MaxQuantity — the number the plan actually authorises — rather
// than from the English text, so the two headlines cannot come to describe
// different quantities.
func (m PlannedMutation) quantityKO() string {
	if m.Quantity == "" {
		return ""
	}
	if m.MaxQuantity > 0 {
		return trim(m.MaxQuantity) + "주"
	}
	return m.Quantity
}

// shape is the side/quantity/symbol trio both headlines are built from, so the two
// languages can never describe different requests.
func (m PlannedMutation) shape(quantity string) []string {
	var what []string
	if m.Side != "" {
		what = append(what, strings.ToUpper(m.Side))
	}
	if quantity != "" {
		what = append(what, quantity)
	}
	if m.Symbol != "" {
		on := m.Symbol + " (" + MarketOf(m.Symbol) + ")"
		if len(what) > 0 {
			on = "of " + on
		}
		what = append(what, on)
	}
	return what
}

// PlanExclusion is a mutating step this run will NOT touch, and why.
//
// It is printed with the plan because "what is not going to happen" is half of what
// makes a list of live orders readable, and because a step excluded for a reason
// the operator can fix (no holding, a symbol that is not KR) is one they may want
// to fix before approving anything.
type PlanExclusion struct {
	Step   StepID `json:"step"`
	Reason string `json:"reason"`
}

// Plan is everything this run can send.
type Plan struct {
	RunID   string `json:"run_id"`
	Account string `json:"account"`
	// Mutations are the approved-or-not lines, in the order the steps run.
	Mutations []PlannedMutation `json:"mutations"`
	// Excluded are the mutating steps this run will skip.
	Excluded []PlanExclusion `json:"excluded,omitempty"`
	// Notes are plan-wide caveats, printed under the list.
	Notes []string `json:"notes,omitempty"`
}

// Digest fingerprints the list, so the record can say which list was approved
// without carrying order bodies into the evidence file.
func (p Plan) Digest() string { return Digest(p.Mutations) }

// Covers reports whether the plan authorises a step to mutate at all.
func (p Plan) Covers(step StepID) bool {
	for _, m := range p.Mutations {
		if m.Step == step {
			return true
		}
	}
	return false
}

// MutatingSteps lists the steps the plan authorises, in order and without repeats.
func (p Plan) MutatingSteps() []StepID {
	var out []StepID
	seen := map[StepID]bool{}
	for _, m := range p.Mutations {
		if !seen[m.Step] {
			seen[m.Step] = true
			out = append(out, m.Step)
		}
	}
	return out
}

// ExclusionReason returns why a mutating step is not in the plan.
func (p Plan) ExclusionReason(step StepID) string {
	for _, e := range p.Excluded {
		if e.Step == step {
			return e.Reason
		}
	}
	return "it was not listed in the batch the operator approved"
}

// Authorises is the whole enforcement surface: mutate.go sends nothing this
// returns false for.
//
// The comparison is deliberately exact on the things a person read off the summary
// — the step, the class of request, the symbol and the side — and a ceiling on the
// one thing the broker's own answers can move, the quantity. A symbol substitution
// therefore cannot pass, which is the point: a run that has to send something else
// asks again rather than adapting.
func (p Plan) Authorises(step StepID, kind MutationKind, symbol, side string, quantity float64) bool {
	for _, m := range p.Mutations {
		if m.Step != step || m.Kind != kind {
			continue
		}
		if m.Symbol != "" && !strings.EqualFold(m.Symbol, strings.TrimSpace(symbol)) {
			continue
		}
		if m.Side != "" && !strings.EqualFold(m.Side, strings.TrimSpace(side)) {
			continue
		}
		if m.MaxQuantity <= 0 {
			// A class that carries no quantity — a cancel. A request that does
			// carry one is not this class.
			if quantity > 0 {
				continue
			}
			return true
		}
		if quantity <= m.MaxQuantity+quantityTolerance {
			return true
		}
	}
	return false
}

// quantityTolerance absorbs float noise in a comparison of share counts. It is far
// below one share, so it can never authorise an extra share.
const quantityTolerance = 1e-9

// WriteLines renders the plan as the numbered list the operator reads.
//
// It renders the Korean (task 1.8 ③). The English the KO fields shadow is still
// on every line — it is what approval.plan_digest is computed over — but it is not
// what anybody reads, and orFallback keeps a line that has not been translated yet
// visible rather than blank.
func (p Plan) WriteLines(w io.Writer) {
	for _, m := range p.Mutations {
		writeWrapped(w, fmt.Sprintf("  %2d. %-22s ", m.Ordinal, m.Step), m.HeadlineKO())
		if m.Pricing != "" {
			writeWrapped(w, "        가격    ", orFallback(m.PricingKO, m.Pricing))
		}
		writeWrapped(w, "        종료    ", orFallback(m.EndsKO, m.Ends))
		if m.Note != "" {
			writeWrapped(w, "        비고    ", orFallback(m.NoteKO, m.Note))
		}
	}
	if len(p.Notes) > 0 {
		fmt.Fprintln(w)
		for _, n := range p.Notes {
			writeWrapped(w, "      · ", n)
		}
	}
	if len(p.Excluded) > 0 {
		fmt.Fprintf(w, "\n  이 실행이 건드리지 않는 mutating 단계:\n")
		for _, e := range p.Excluded {
			writeWrapped(w, fmt.Sprintf("    %-22s ", e.Step), e.Reason)
		}
	}
}

// orFallback prefers the translation and falls back to the original.
//
// A missing Korean line is a gap in the display, not a reason to show an operator
// nothing where a live request's reversal should be.
func orFallback(translated, original string) string {
	if strings.TrimSpace(translated) != "" {
		return translated
	}
	return original
}

// writeWrapped prints a label and its text wrapped under a hanging indent, because
// a safety summary nobody can read across an eighty-column terminal is a safety
// summary nobody reads.
func writeWrapped(w io.Writer, label, text string) {
	const width = 96
	// Runes, not bytes: a label with a bullet in it would otherwise indent its
	// continuation lines by the wrong amount.
	labelWidth := utf8.RuneCountInString(label)
	indent := strings.Repeat(" ", labelWidth)
	for i, line := range wrapText(text, width-labelWidth) {
		if i == 0 {
			fmt.Fprintf(w, "%s%s\n", label, line)
			continue
		}
		fmt.Fprintf(w, "%s%s\n", indent, line)
	}
}

func wrapText(text string, width int) []string {
	if width < 20 {
		width = 20
	}
	words := strings.Fields(text)
	if len(words) == 0 {
		return []string{""}
	}
	var (
		lines []string
		cur   string
	)
	for _, word := range words {
		switch {
		case cur == "":
			cur = word
		case len(cur)+1+len(word) <= width:
			cur += " " + word
		default:
			lines = append(lines, cur)
			cur = word
		}
	}
	return append(lines, cur)
}

// --- building the plan ----------------------------------------------------------

// planStatus is what happened to one declared mutation when it was resolved.
type planStatus int

const (
	// planOK: the line is on the list.
	planOK planStatus = iota
	// planOmitted: the step will deliberately not send this one, and the plan says
	// so. The step stays on the list for its other lines.
	planOmitted
	// planUnknown: the account could not be read well enough to say what would be
	// sent. The whole step comes off the list, because an unreadable quantity is
	// not something to ask a person to approve.
	planUnknown
)

// Plan works out what this run can send. It makes read-only calls and no others.
func (r *Runner) Plan(ctx context.Context) Plan {
	plan := Plan{RunID: r.runID, Account: maskedAccount(r.accountRef)}
	sellable, sellableKnown := r.planSellable(ctx)

	willRun := map[StepID]bool{}
	passed := func(id StepID) bool { return willRun[id] || Passed(r.prior, id) }

	ordinal := 0
	for _, step := range Steps() {
		if settled, verdict := r.settled(step.ID); settled {
			if step.Mutates {
				plan.Excluded = append(plan.Excluded, PlanExclusion{
					Step: step.ID,
					Reason: fmt.Sprintf("기록에 이미 %s로 판정되어 있다. 다시 실행하면 이미 끝난 측정을 위해 "+
						"두 번째 실주문을 내게 된다 (--redo가 이를 무시한다)", verdict),
				})
			}
			continue
		}
		if reason, skip := r.preflightStatic(step, passed); skip {
			if step.Mutates {
				plan.Excluded = append(plan.Excluded, PlanExclusion{Step: step.ID, Reason: reason})
			}
			continue
		}
		willRun[step.ID] = true
		if !step.Mutates {
			continue
		}

		symbol := r.mutationSymbol(step)
		lines, notes, ok := r.planStep(step, symbol, sellable, sellableKnown)
		if !ok {
			willRun[step.ID] = false
			plan.Excluded = append(plan.Excluded, PlanExclusion{
				Step: step.ID,
				Reason: fmt.Sprintf("이 단계가 %s에 대해 무엇을 보낼지 말할 수 있을 만큼 계좌를 읽지 못했다. "+
					"모르는 채로 승인받는 대신 배치에서 뺀다", symbol),
			})
			continue
		}
		for _, line := range lines {
			ordinal++
			line.Ordinal = ordinal
			plan.Mutations = append(plan.Mutations, line)
		}
		plan.Notes = append(plan.Notes, notes...)
	}
	return plan
}

// planStep resolves one step's declared mutations against the account.
func (r *Runner) planStep(step Step, symbol string, sellable float64, sellableKnown bool) ([]PlannedMutation, []string, bool) {
	var (
		lines []PlannedMutation
		notes []string
	)
	for _, sm := range step.Mutations {
		text, max, status := r.planQuantity(sm.Quantity, sellable, sellableKnown)
		switch status {
		case planUnknown:
			return nil, nil, false
		case planOmitted:
			notes = append(notes, r.omissionNote(step, sm, symbol, sellable))
			continue
		}
		lines = append(lines, PlannedMutation{
			Step:        step.ID,
			Kind:        sm.Kind,
			Symbol:      symbol,
			Side:        sm.Side,
			Quantity:    text,
			MaxQuantity: max,
			Pricing:     sm.Pricing.Describe(r.offset, MarketOf(symbol)),
			Ends:        sm.Ends,
			Note:        sm.Note,
			// Display only. None of these three reaches the digest.
			PricingKO: sm.Pricing.DescribeKO(r.offset, MarketOf(symbol)),
			EndsKO:    sm.EndsKO,
			NoteKO:    sm.NoteKO,
		})
	}
	if len(lines) == 0 {
		// Every line was omitted, so the step sends nothing. Leaving it "authorised
		// but with no lines" would be a step that could mutate without a line, so
		// it comes off the list.
		return nil, nil, false
	}
	return lines, notes, true
}

// planQuantity resolves a quantity rule against the account.
func (r *Runner) planQuantity(rule QuantityRule, sellable float64, known bool) (string, float64, planStatus) {
	switch rule {
	case QuantityNone:
		return "", 0, planOK
	case QuantityOne:
		return trim(MinQuantity) + " share", MinQuantity, planOK
	case QuantityPartial:
		if !known {
			return "", 0, planUnknown
		}
		if sellable <= MinQuantity {
			return "", 0, planOmitted
		}
		return trim(MinQuantity) + " share", MinQuantity, planOK
	case QuantityWholeHolding:
		if !known {
			return "", 0, planUnknown
		}
		if sellable < MinQuantity || sellable > r.maxSellQuantity {
			return "", 0, planOmitted
		}
		return trim(sellable) + " share(s)", sellable, planOK
	case QuantityOverHolding:
		if !known {
			return "", 0, planUnknown
		}
		if sellable < MinQuantity {
			return "", 0, planOmitted
		}
		return trim(sellable+MinQuantity) + " share(s)", sellable + MinQuantity, planOK
	default:
		return "", 0, planUnknown
	}
}

// omissionNote explains a line the plan deliberately left out.
func (r *Runner) omissionNote(step Step, sm StepMutation, symbol string, sellable float64) string {
	switch sm.Quantity {
	case QuantityPartial:
		return fmt.Sprintf("%s: %s holds %s share(s), so a partial sell and a whole-holding sell would be the "+
			"same order — no separate partial sell is planned", step.ID, symbol, trim(sellable))
	case QuantityWholeHolding:
		return fmt.Sprintf("%s: the whole-holding sell would be %s share(s), above the --max-sell-quantity cap "+
			"of %s — it is NOT placed and the boundary is recorded as unverified",
			step.ID, trim(sellable), trim(r.maxSellQuantity))
	case QuantityOverHolding:
		return fmt.Sprintf("%s: %s reports nothing sellable, so the over-holding probe is not planned",
			step.ID, symbol)
	default:
		return fmt.Sprintf("%s: one %s request is not planned for this account", step.ID, sm.Kind)
	}
}

// planSellable reads the one number the plan cannot state without the account: how
// much of the held symbol the broker says can be sold.
//
// A failure is not fatal and is not guessed around either — the sell-side step
// simply does not go on the list, and the plan says why.
//
// It goes through readRetry with no step to log against: this read runs while the
// plan is still being built, and a 429 here silently costs the operator the whole
// sell-side batch (measurements.md M4 is what that looks like). Retrying a read
// before anything has been approved cannot send anything.
func (r *Runner) planSellable(ctx context.Context) (float64, bool) {
	if r.holdingSymbol == "" {
		return 0, false
	}
	sq, err := readRetry(ctx, r, nil, EndpointReadSellable, nil,
		func(ctx context.Context) (domain.SellableQuantity, error) {
			return r.broker.SellableQuantity(ctx, r.holdingSymbol)
		}, nil)
	if err != nil {
		return 0, false
	}
	return sq.Quantity, true
}
