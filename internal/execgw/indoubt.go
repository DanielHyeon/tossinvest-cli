package execgw

// indoubt.go resolves mutations whose outcome the transport could not establish
// (harden-execution-base task 2.7).
//
// # The one rule everything else follows from
//
// Automatic re-submission is forbidden. Unconditionally. The official API has no
// idempotency key, so "try again" and "place a second live order" are the same
// action. This file therefore contains no mutator of any kind — the Resolver has
// no trading service, no broker writer, no submit path — and the only outcomes it
// can produce are: found it, proved it is not there, or could not tell.
//
// # Why absence is expensive to prove
//
// Presence is cheap: the order is in a list. Absence is not, because a list can be
// stale, paginated, or scoped to the wrong status. So "not there" requires all of:
//
//	1. both OPEN and CLOSED walked to the last page, every time,
//	2. N consecutive scans in a row that find nothing (default 3),
//	3. spread over a minimum observation window (default 45s),
//	4. an account cross-check: the buying power and the holding for the symbol
//	   must be consistent with the mutation never having happened.
//
// Fail any of them and the answer is UNRESOLVED_IN_DOUBT — a permanent block on
// the symbol until an operator resolves it — never "probably fine".

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
)

// OrderQuery selects broker orders to scan.
type OrderQuery struct {
	// Status is "OPEN" or "CLOSED". Both are always scanned; the resolver issues
	// one query per status.
	Status string
	Symbol string
	From   string // YYYY-MM-DD, inclusive
	To     string // YYYY-MM-DD, inclusive
	Limit  int
}

// OrderPage is one page of raw broker order JSON.
//
// Raw, not adapted: the derivation needs canceledAt and the matching needs
// orderedAt, and the upstream adapter (internal/official.adaptOrder) drops both
// (issues.md, 2026-07-26). Keeping the payload raw here is what lets
// brokerstate.ParseOfficialOrder see the whole record.
type OrderPage struct {
	Orders     []json.RawMessage
	NextCursor string
	HasNext    bool
}

// OrderPager fetches one page of orders. internal/official.Client satisfies the
// shape via the adapter in orders_source.go.
type OrderPager interface {
	OrdersPageRaw(ctx context.Context, q OrderQuery, cursor string) (OrderPage, error)
}

// OrderReader fetches one order's raw JSON by id.
type OrderReader interface {
	OrderRaw(ctx context.Context, orderID string) (json.RawMessage, error)
}

// AccountReader supplies the delta cross-check inputs.
type AccountReader interface {
	BuyingPower(ctx context.Context, currency string) (float64, error)
	HoldingQuantity(ctx context.Context, symbol string) (float64, error)
}

// Baseline is the account snapshot taken *before* a mutation was dispatched.
//
// It is supplied by the caller rather than read by the gateway on purpose: the
// strategy loop already polls buying power and holdings for its entry decision,
// so reusing that read costs no extra API budget (§0.4). A mutation submitted
// without a baseline can still be journalled and dispatched — but if it ends up
// IN_DOUBT, its absence can never be proven, and it will park as
// UNRESOLVED_IN_DOUBT. That is the documented price of omitting it.
type Baseline struct {
	// BuyingPower is the cash buying power in Currency before dispatch.
	BuyingPower float64 `json:"buying_power"`
	// Holding is the held quantity of the symbol before dispatch.
	Holding float64 `json:"holding"`
	// Currency is the currency BuyingPower is expressed in.
	Currency string `json:"currency"`
}

// baselineNote is how a Baseline travels: JSON inside the immutable intent's
// notes column. The intent row is the one record guaranteed to survive a restart
// alongside the attempt, which is exactly when a baseline is needed.
type baselineNote struct {
	Baseline *Baseline `json:"baseline,omitempty"`
}

// EncodeBaseline renders a baseline for the journal's notes column.
func EncodeBaseline(b *Baseline) string {
	if b == nil {
		return ""
	}
	data, err := json.Marshal(baselineNote{Baseline: b})
	if err != nil {
		return ""
	}
	return string(data)
}

// DecodeBaseline reads a baseline back out of an intent's notes.
func DecodeBaseline(notes string) (Baseline, bool) {
	if strings.TrimSpace(notes) == "" {
		return Baseline{}, false
	}
	var note baselineNote
	if err := json.Unmarshal([]byte(notes), &note); err != nil || note.Baseline == nil {
		return Baseline{}, false
	}
	return *note.Baseline, true
}

// ResolveConfig tunes the resolution procedure. The defaults are conservative and
// provisional, like the retry matrix's: verify-execution-capability measures the
// real numbers.
type ResolveConfig struct {
	// StableObservations is how many consecutive empty scans absence needs.
	StableObservations int
	// MinObservation is the minimum wall time those scans must span, so a burst
	// of three fast polls against a lagging list cannot pass for stability.
	MinObservation time.Duration
	// PollInterval is the wait between scans.
	PollInterval time.Duration
	// MaxDuration bounds the whole procedure. Exceeding it is UNRESOLVED.
	MaxDuration time.Duration
	// MaxPages bounds one list walk (cursor-loop defence).
	MaxPages int
	// SubmitWindow is how far a broker order's orderedAt may sit from the
	// attempt's dispatch time and still be considered the same order.
	SubmitWindow time.Duration
}

// DefaultResolveConfig is the provisional default.
func DefaultResolveConfig() ResolveConfig {
	return ResolveConfig{
		StableObservations: 3,
		MinObservation:     45 * time.Second,
		PollInterval:       5 * time.Second,
		MaxDuration:        5 * time.Minute,
		MaxPages:           50,
		SubmitWindow:       5 * time.Minute,
	}
}

func (c ResolveConfig) withDefaults() ResolveConfig {
	d := DefaultResolveConfig()
	if c.StableObservations <= 0 {
		c.StableObservations = d.StableObservations
	}
	if c.MinObservation <= 0 {
		c.MinObservation = d.MinObservation
	}
	if c.PollInterval <= 0 {
		c.PollInterval = d.PollInterval
	}
	if c.MaxDuration <= 0 {
		c.MaxDuration = d.MaxDuration
	}
	if c.MaxPages <= 0 {
		c.MaxPages = d.MaxPages
	}
	if c.SubmitWindow <= 0 {
		c.SubmitWindow = d.SubmitWindow
	}
	return c
}

// Resolver settles IN_DOUBT attempts by observation alone.
type Resolver struct {
	// Journal is the record of what was attempted. Required.
	Journal *journal.Journal
	// Orders walks the broker's order lists. Required.
	Orders OrderPager
	// Order reads a single order by id. Required for CANCEL/AMEND resolution
	// (task 2.8); PLACE resolution does not use it.
	Order OrderReader
	// Account supplies the absence cross-check. Without it, absence is
	// unprovable and every unfound attempt parks as UNRESOLVED.
	Account AccountReader
	// Clock drives the stabilisation intervals.
	Clock clock.Clock
	// Gate is latched when an attempt ends UNRESOLVED. Optional.
	Gate   *EntryGate
	Config ResolveConfig
}

// Resolution is the outcome of one resolution pass.
type Resolution struct {
	AttemptID     string
	IntentID      string
	Symbol        string
	State         journal.AttemptState
	BrokerOrderID string
	Reason        ReasonCode
	Detail        string
	// Observations is how many full OPEN+CLOSED scans were performed.
	Observations int
}

// ErrNotResolvable means the attempt is not in a state the resolver acts on.
var ErrNotResolvable = errors.New("execgw: attempt is not awaiting resolution")

// Resolve settles one IN_DOUBT attempt.
//
// It is idempotent: an already-terminal attempt is reported back unchanged, which
// is what lets a restart run it over every pending attempt without thinking.
func (r *Resolver) Resolve(ctx context.Context, attemptID string) (Resolution, error) {
	cfg := r.Config.withDefaults()
	clk := r.Clock
	if clk == nil {
		clk = clock.System()
	}
	if r.Journal == nil {
		return Resolution{}, errors.New("execgw: resolution needs a journal")
	}

	rec, err := r.Journal.LookupAttempt(ctx, attemptID)
	if err != nil {
		return Resolution{}, err
	}
	intent, err := r.Journal.LookupIntent(ctx, rec.IntentID)
	if err != nil {
		return Resolution{}, err
	}
	res := Resolution{
		AttemptID: rec.ID,
		IntentID:  rec.IntentID,
		Symbol:    intent.Symbol,
		State:     rec.State,
	}
	if rec.State.IsTerminal() {
		res.BrokerOrderID = rec.BrokerOrderID
		res.Reason = ReasonCode(rec.ReasonCode)
		res.Detail = rec.Detail
		return res, nil
	}
	if rec.State != journal.StateInDoubt {
		return res, fmt.Errorf("%w: attempt %s is %s", ErrNotResolvable, rec.ID, rec.State)
	}
	if r.Orders == nil {
		return res, errors.New("execgw: resolution needs an order pager")
	}

	attempt, err := r.Journal.Resume(ctx, attemptID)
	if err != nil {
		return res, err
	}

	// A cancel and an amend cannot be resolved by looking for an order that
	// matches the request — neither creates one. They have their own procedures
	// in amend_indoubt.go.
	switch rec.Kind {
	case journal.KindCancel:
		return r.resolveCancel(ctx, attempt, rec, intent, res, cfg, clk)
	case journal.KindAmend:
		return r.resolveAmend(ctx, attempt, rec, intent, res, cfg, clk)
	}

	matcher, err := newMatcher(intent, rec, cfg)
	if err != nil {
		return res, err
	}

	started := clk.Now()
	deadline := started.Add(cfg.MaxDuration)
	absent := 0
	firstAbsentAt := time.Time{}

	for {
		res.Observations++
		found, scanErr := r.scanBoth(ctx, matcher, cfg)
		switch {
		case scanErr != nil:
			// A scan that failed is not evidence of absence. Keep trying inside
			// the budget; run out and the attempt parks.
			if !r.waitOrExpire(ctx, clk, cfg, deadline) {
				return r.park(ctx, attempt, res,
					fmt.Sprintf("the order lists could not be walked: %v", scanErr))
			}
			continue

		case len(found) == 1:
			order := found[0]
			res.State = journal.StateConfirmed
			res.BrokerOrderID = order.OrderID
			res.Reason = ReasonCode(journal.ReasonResolvedFound)
			res.Detail = fmt.Sprintf("found order %s in the %s list after %d scan(s)",
				order.OrderID, order.listStatus, res.Observations)
			if err := attempt.ResolveConfirmed(ctx, order.OrderID,
				journal.ReasonResolvedFound, res.Detail); err != nil {
				return res, err
			}
			return res, nil

		case len(found) > 1:
			// The engine allows one in-flight mutation per symbol precisely so
			// this cannot happen. That it did means our model of the account is
			// wrong, which is a stop condition, not a coin flip.
			ids := make([]string, 0, len(found))
			for _, o := range found {
				ids = append(ids, o.OrderID)
			}
			return r.park(ctx, attempt, res,
				"the fingerprint matched more than one broker order ("+strings.Join(ids, ", ")+
					"); picking one would be a guess")

		default: // absent this time round
			absent++
			if firstAbsentAt.IsZero() {
				firstAbsentAt = clk.Now()
			}
			stable := absent >= cfg.StableObservations
			longEnough := clk.Now().Sub(started) >= cfg.MinObservation
			if stable && longEnough {
				proven, why := r.absenceCorroborated(ctx, rec, intent)
				if proven {
					res.State = journal.StateFailedConfirmed
					res.Reason = ReasonCode(journal.ReasonResolvedAbsent)
					res.Detail = fmt.Sprintf("%d consecutive full scans over %s found nothing, and %s",
						absent, clk.Now().Sub(started).Round(time.Second), why)
					if err := attempt.ResolveFailed(ctx, journal.ReasonResolvedAbsent, res.Detail); err != nil {
						return res, err
					}
					return res, nil
				}
				return r.park(ctx, attempt, res, why)
			}
			if !r.waitOrExpire(ctx, clk, cfg, deadline) {
				return r.park(ctx, attempt, res, fmt.Sprintf(
					"neither presence nor absence was established within %s (%d scan(s))",
					cfg.MaxDuration, res.Observations))
			}
		}
	}
}

// waitOrExpire sleeps one poll interval, reporting false when the budget is spent.
func (r *Resolver) waitOrExpire(ctx context.Context, clk clock.Clock, cfg ResolveConfig, deadline time.Time) bool {
	if !clk.Now().Add(cfg.PollInterval).Before(deadline) {
		return false
	}
	return clk.Sleep(ctx, cfg.PollInterval) == nil
}

// park ends the procedure as UNRESOLVED_IN_DOUBT and latches the entry gate.
//
// This is a deliberate, permanent stop: the spec requires the symbol to stay
// blocked until a human resolves it, because an unknown live order is the one
// state an automated system must never trade around.
func (r *Resolver) park(ctx context.Context, attempt *journal.Attempt, res Resolution, detail string) (Resolution, error) {
	res.State = journal.StateUnresolvedInDoubt
	res.Reason = ReasonUnresolvedInDoubt
	res.Detail = detail
	if err := attempt.ResolveUnresolved(ctx, journal.ReasonResolutionExhausted, detail); err != nil {
		return res, err
	}
	if r.Gate != nil {
		r.Gate.Block(ReasonUnresolvedInDoubt,
			fmt.Sprintf("attempt %s on %s is unresolved: %s", res.AttemptID, res.Symbol, detail))
	}
	return res, nil
}

// scanBoth walks OPEN and CLOSED to the last page and returns every distinct
// match.
//
// Both lists, every time, even after a match: a duplicate sitting in the other
// list is exactly the situation that must not be reported as a clean single hit.
//
// # Why the results are deduplicated by orderId before anything counts them
//
// The `status` query parameter is a *group label*, not the order's own status,
// and the groups overlap: "OPEN: 진행 중 주문 그룹 — orders[].status ∈ {PENDING,
// PARTIAL_FILLED, PENDING_CANCEL, PENDING_REPLACE}" and "CLOSED: 종료된 주문
// 그룹 — … REPLACE_REJECTED, PARTIAL_FILLED" (openapi, GET /api/v1/orders). A
// partially filled order is therefore returned by *both* queries — and a
// partial fill plus a lost response is the single most likely way an attempt
// lands in doubt at all.
//
// Without this dedup that order matches twice, the caller reads two matches as
// "we cannot tell which one is ours", and the attempt parks permanently for an
// operator. The duplication is an artefact of how the question was asked, not a
// fact about the account, so it is removed here rather than reasoned about
// upstream. Identity is byte-exact: `orderId` is an opaque token (openapi
// contracts no shape), so two records are the same order only if their ids are
// the same bytes.
func (r *Resolver) scanBoth(ctx context.Context, m *matcher, cfg ResolveConfig) ([]matchedOrder, error) {
	var (
		matches []matchedOrder
		seen    = map[string]int{} // orderId (verbatim) → index into matches
	)
	for _, status := range []string{"OPEN", "CLOSED"} {
		raws, err := ScanOrders(ctx, r.Orders, OrderQuery{Status: status, Symbol: m.symbol}, cfg.MaxPages)
		if err != nil {
			return nil, fmt.Errorf("scanning the %s list: %w", status, err)
		}
		for _, raw := range raws {
			facts, err := parseOrderFacts(raw)
			if err != nil {
				// An unparseable record is not "not a match": it is an unknown.
				// Treating it as absence could close an attempt that exists.
				return nil, fmt.Errorf("an order in the %s list could not be read: %w", status, err)
			}
			if !m.matches(facts) {
				continue
			}
			if at, already := seen[facts.OrderID]; already {
				// Same order, second group. Record where it was seen — the detail
				// an operator reads should say so — and keep one match.
				matches[at].listStatus += "+" + status
				continue
			}
			seen[facts.OrderID] = len(matches)
			matches = append(matches, matchedOrder{orderFacts: facts, listStatus: status})
		}
	}
	return matches, nil
}

// absenceCorroborated performs the account cross-check. It answers (proven, why)
// where why is a human-readable sentence used either as the confirmation detail or
// as the reason the attempt parks.
func (r *Resolver) absenceCorroborated(ctx context.Context, rec journal.AttemptRecord,
	intent journal.Intent,
) (bool, string) {
	if r.Account == nil {
		return false, "absence cannot be corroborated: no account reader is configured"
	}
	// The cross-check compares the account now against a baseline taken before
	// this attempt was dispatched. Anything else that was sent on this symbol
	// inside that window moves the same numbers, so the comparison stops meaning
	// what it has to mean, and absence cannot be inferred from it at all
	// (order-execution: "관측 창 동안 같은 심볼에 다른 mutation이 전송되었다면
	// delta 교차 확인은 무효이며 자동 FAILED_CONFIRMED는 금지된다"). Parking is
	// the only honest answer: we cannot prove a negative from contaminated
	// evidence, and guessing here retires an order that may be live.
	if contaminated, why := r.windowContaminated(ctx, rec, intent); contaminated {
		return false, why
	}
	baseline, ok := DecodeBaseline(intent.Notes)
	if !ok {
		return false, "absence cannot be corroborated: the mutation was submitted without a pre-dispatch account baseline"
	}

	holdingNow, err := r.Account.HoldingQuantity(ctx, intent.Symbol)
	if err != nil {
		return false, fmt.Sprintf("absence cannot be corroborated: reading the holding failed: %v", err)
	}
	currency := baseline.Currency
	if currency == "" {
		currency = intent.Currency
	}
	buyingPowerNow, err := r.Account.BuyingPower(ctx, currency)
	if err != nil {
		return false, fmt.Sprintf("absence cannot be corroborated: reading the buying power failed: %v", err)
	}

	holdingDelta := holdingNow - baseline.Holding
	if !nearZero(holdingDelta, baseline.Holding) {
		return false, fmt.Sprintf(
			"the %s holding moved by %s while the order lists showed nothing — the account and the lists disagree",
			intent.Symbol, decimalString(holdingDelta))
	}

	// Buying power is weaker evidence than the holding: another symbol's activity
	// can move it. So it only has to be inconsistent with *this* order having
	// been accepted — a reservation of roughly this order's notional.
	notional := intentNotional(intent)
	bpDelta := buyingPowerNow - baseline.BuyingPower
	if notional > 0 && bpDelta < 0 && math.Abs(bpDelta) >= notional*0.5 {
		return false, fmt.Sprintf(
			"the buying power dropped by %s, consistent with this order (%s) having been accepted",
			decimalString(math.Abs(bpDelta)), decimalString(notional))
	}

	return true, fmt.Sprintf("the %s holding and the buying power are unchanged from the pre-dispatch baseline",
		intent.Symbol)
}

// windowContaminated reports whether another mutation on this symbol was
// dispatched inside the observation window, and says which one.
//
// The window starts at this attempt's own dispatch, because that is when its
// baseline stopped being current. A mutation that was journalled but provably
// never sent does not contaminate anything and is excluded by the journal query.
//
// A journal that cannot be read is treated as contamination rather than as
// cleanliness: "I could not check" and "I checked and it was clean" must not
// produce the same answer when the answer authorises retiring an order.
func (r *Resolver) windowContaminated(ctx context.Context, rec journal.AttemptRecord,
	intent journal.Intent,
) (bool, string) {
	since := rec.DispatchStartedAt
	if strings.TrimSpace(since) == "" {
		since = rec.RecordedAt
	}
	others, err := r.Journal.MutationsDispatchedSince(ctx, intent.Market, intent.Symbol, since, rec.ID)
	if err != nil {
		return true, fmt.Sprintf(
			"absence cannot be corroborated: the journal could not be checked for other %s mutations: %v",
			intent.Symbol, err)
	}
	if len(others) == 0 {
		return false, ""
	}
	ids := make([]string, 0, len(others))
	for _, other := range others {
		ids = append(ids, fmt.Sprintf("%s (%s, %s)", other.ID, other.Kind, other.State))
	}
	return true, fmt.Sprintf(
		"absence cannot be corroborated: %s was also touched by %s while this attempt was being observed, "+
			"so the buying power and holding deltas no longer describe this order alone",
		intent.Symbol, strings.Join(ids, ", "))
}

// nearZero reports whether a delta is zero within a scale-relative tolerance.
// Quantities arrive as decimal strings and are compared as float64.
func nearZero(delta, scale float64) bool {
	tolerance := 1e-9 * math.Max(1, math.Abs(scale))
	return math.Abs(delta) <= tolerance
}

func intentNotional(intent journal.Intent) float64 {
	qty, err := strconv.ParseFloat(strings.TrimSpace(intent.Quantity), 64)
	if err != nil {
		return 0
	}
	price, err := strconv.ParseFloat(strings.TrimSpace(intent.Price), 64)
	if err != nil {
		return 0
	}
	return qty * price
}

// --- matching ---------------------------------------------------------------

// orderFacts is the subset of a broker order the matcher reads. brokerstate owns
// the *state* derivation; this is only about identity.
type orderFacts struct {
	OrderID   string
	Symbol    string
	Side      string
	Status    string
	Quantity  float64
	Price     float64
	OrderedAt time.Time
	// HasOrderedAt is false when the payload carried no usable timestamp.
	HasOrderedAt bool
}

type matchedOrder struct {
	orderFacts
	listStatus string
}

type officialOrderFacts struct {
	OrderID   string `json:"orderId"`
	Symbol    string `json:"symbol"`
	Side      string `json:"side"`
	Status    string `json:"status"`
	Quantity  string `json:"quantity"`
	Price     string `json:"price"`
	OrderedAt string `json:"orderedAt"`
}

// ErrMalformedOrder means a broker order payload could not be read.
var ErrMalformedOrder = errors.New("execgw: malformed broker order payload")

func parseOrderFacts(raw json.RawMessage) (orderFacts, error) {
	var payload officialOrderFacts
	if err := json.Unmarshal(raw, &payload); err != nil {
		return orderFacts{}, fmt.Errorf("%w: %v", ErrMalformedOrder, err)
	}
	if strings.TrimSpace(payload.OrderID) == "" {
		return orderFacts{}, fmt.Errorf("%w: no orderId", ErrMalformedOrder)
	}
	facts := orderFacts{
		// Verbatim: `orderId` is opaque (openapi contracts no shape), so it is
		// stored and compared exactly as received. Symbol, side and status are
		// not identifiers — they are enum-ish fields the matcher compares
		// case-insensitively — so normalising those is safe and stays.
		OrderID: payload.OrderID,
		Symbol:  strings.ToUpper(strings.TrimSpace(payload.Symbol)),
		Side:    strings.ToUpper(strings.TrimSpace(payload.Side)),
		Status:  strings.ToUpper(strings.TrimSpace(payload.Status)),
	}
	var err error
	if facts.Quantity, err = parseOptionalDecimal(payload.Quantity); err != nil {
		return orderFacts{}, fmt.Errorf("%w: quantity: %v", ErrMalformedOrder, err)
	}
	if facts.Price, err = parseOptionalDecimal(payload.Price); err != nil {
		return orderFacts{}, fmt.Errorf("%w: price: %v", ErrMalformedOrder, err)
	}
	if ts := strings.TrimSpace(payload.OrderedAt); ts != "" {
		if parsed, perr := time.Parse(time.RFC3339, ts); perr == nil {
			facts.OrderedAt = parsed.UTC()
			facts.HasOrderedAt = true
		}
	}
	return facts, nil
}

func parseOptionalDecimal(s string) (float64, error) {
	trimmed := strings.TrimSpace(s)
	if trimmed == "" || trimmed == "null" {
		return 0, nil
	}
	v, err := strconv.ParseFloat(trimmed, 64)
	if err != nil {
		return 0, fmt.Errorf("%q is not a decimal", trimmed)
	}
	return v, nil
}

// matcher decides whether a broker order is the one an attempt tried to create.
type matcher struct {
	symbol   string
	side     string
	quantity float64
	price    float64
	// window is the interval around dispatch time an order's orderedAt may fall
	// in. Zero-value bounds mean "no usable dispatch timestamp", in which case
	// time is not used as a discriminator at all.
	windowStart, windowEnd time.Time
	useWindow              bool
	// targetOrderID is set for CANCEL/AMEND resolution (task 2.8).
	targetOrderID string
}

func newMatcher(intent journal.Intent, rec journal.AttemptRecord, cfg ResolveConfig) (*matcher, error) {
	quantity, err := strconv.ParseFloat(strings.TrimSpace(intent.Quantity), 64)
	if err != nil {
		return nil, fmt.Errorf("execgw: intent %s has an unreadable quantity %q", intent.ID, intent.Quantity)
	}
	price, _ := parseOptionalDecimal(intent.Price)

	m := &matcher{
		symbol:        strings.ToUpper(strings.TrimSpace(intent.Symbol)),
		side:          strings.ToUpper(strings.TrimSpace(intent.Side)),
		quantity:      quantity,
		price:         price,
		targetOrderID: rec.TargetOrderID,
	}
	// The submission window comes from the attempt's own dispatch timestamp,
	// which is why the fingerprint hash does not carry it: it is knowable at
	// resolution time and not before.
	stamp := rec.DispatchStartedAt
	if stamp == "" {
		stamp = rec.RecordedAt
	}
	if dispatched, err := time.Parse(time.RFC3339, stamp); err == nil {
		m.windowStart = dispatched.Add(-cfg.SubmitWindow)
		m.windowEnd = dispatched.Add(cfg.SubmitWindow)
		m.useWindow = true
	}
	return m, nil
}

// matches applies the fingerprint fields.
//
// Every "unknown" resolves towards *match*, not away from it: a missing
// orderedAt, or a price we cannot compare, keeps the order as a candidate. A
// false non-match would let the resolver declare absence for an order that
// exists, which is the one error that leaves a live position untracked.
func (m *matcher) matches(o orderFacts) bool {
	if m.symbol != "" && o.Symbol != "" && o.Symbol != m.symbol {
		return false
	}
	if m.side != "" && o.Side != "" && o.Side != m.side {
		return false
	}
	if !qtyClose(o.Quantity, m.quantity) {
		return false
	}
	if m.price > 0 && o.Price > 0 && !qtyClose(o.Price, m.price) {
		return false
	}
	if m.useWindow && o.HasOrderedAt {
		if o.OrderedAt.Before(m.windowStart) || o.OrderedAt.After(m.windowEnd) {
			return false
		}
	}
	return true
}

func qtyClose(a, b float64) bool {
	scale := math.Max(1, math.Max(math.Abs(a), math.Abs(b)))
	return math.Abs(a-b) <= 1e-9*scale
}

// --- pagination completion --------------------------------------------------

var (
	// ErrTooManyPages means a list walk hit the page ceiling. Fail closed: a
	// partial list is not a list.
	ErrTooManyPages = errors.New("execgw: order list exceeded the page limit")
	// ErrCursorLoop means the broker handed back a cursor it already gave us.
	ErrCursorLoop = errors.New("execgw: order list pagination repeated a cursor")
)

// ScanOrders walks a list to its last page and returns every raw order.
//
// It refuses to return a partial list. Upstream's Orders() keeps only the first
// page (design D4), and a resolver reasoning about "the whole list" from one page
// would report absence for anything further down — which is why the completion
// loop, its cursor-repeat check and its page ceiling all live on the safety side
// of the seam rather than in the client.
func ScanOrders(ctx context.Context, pager OrderPager, q OrderQuery, maxPages int) ([]json.RawMessage, error) {
	if pager == nil {
		return nil, errors.New("execgw: no order pager configured")
	}
	if maxPages <= 0 {
		maxPages = DefaultResolveConfig().MaxPages
	}

	var (
		out    []json.RawMessage
		cursor string
		seen   = map[string]bool{}
	)
	for page := 1; ; page++ {
		if page > maxPages {
			return nil, fmt.Errorf("%w: stopped after %d pages of the %s list",
				ErrTooManyPages, maxPages, q.Status)
		}
		result, err := pager.OrdersPageRaw(ctx, q, cursor)
		if err != nil {
			return nil, err
		}
		out = append(out, result.Orders...)

		next := strings.TrimSpace(result.NextCursor)
		if !result.HasNext || next == "" {
			return out, nil
		}
		if seen[next] {
			return nil, fmt.Errorf("%w: cursor %q was handed back twice while walking the %s list",
				ErrCursorLoop, next, q.Status)
		}
		seen[next] = true
		cursor = next
	}
}
