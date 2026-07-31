package console

// orders.go is the account's order record, read (change console-orders-screen).
//
// # Why this screen exists
//
// /positions answers "what does the account hold" and /history answers "what did
// it finish". Neither answers the question between them — is an order I sent
// still alive — and that is the question an operator asks before deciding whether
// the next verification can send anything at all.
//
// # Three lists, never summed across a failure
//
// A conditional order is on a different endpoint from a plain one, and in this
// product it is the durable artefact rather than the corner case: M18 measured
// that one survives the process that registered it, and internal/verifylive's
// cleanup counts both kinds as leftovers filling the live-exposure cap. A screen
// that called only the plain endpoint would render "미체결 0건" as a measured
// value while a conditional leftover was blocking the next verification — which
// is the exact failure this screen exists to prevent.
//
// The plain endpoint is itself two calls, and the reason is the same one. Its
// `status` parameter is documented `required: true` and it selects between two
// differently shaped answers: OPEN returns every pending order and ignores limit
// and cursor, CLOSED paginates and spans the whole account history when no dates
// are given. Asking for neither — which the first implementation did — is either
// a refused request or one page over all time, and in the second case a live
// order past row 100 is missing from the table AND from the count, rendered as
// "0건 이상". Asking for OPEN by name is the only shape of this call that
// structurally cannot miss a leftover. The finished half is a call of its own
// because a cancelled or rejected order never becomes a trade, so /history never
// shows it either.
//
// So one refresh is three calls, and if only some of them answer the screen says
// so. A confident zero has to be unreachable: the counts are added only when the
// lists behind them were measured.
//
// # It is a reading, and the route table can check that
//
// The route table's account-verb guard fails any path containing "order",
// deliberately and correctly — /orders/cancel and /order-place are caught by the
// same loose string test. /orders is not what that ban aims at, so it is granted
// a byte-exact exception of one path, and only while the route carries the
// `readOnly` wrapper. That wrapper is what makes "it is a read" a checkable fact
// rather than an inference from "it is not CSRF-gated", which would be an
// exception granted BECAUSE the route is unprotected.

import (
	"context"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/brokerstate"
	marketclock "github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/operatorview"
)

// --- the seam --------------------------------------------------------------------

// OrdersReader is the console's entire view of the account's order record: one
// read, and it cannot change anything.
//
// It declares ONE method although a refresh is three broker calls. Those calls
// are the adapter's business (cmd/tossctl), and keeping them there is what stops
// this package from being able to spend the rate budget three times over, or to
// spend it on one endpoint and report another as zero: the outcome of each call
// arrives inside the value, so a partial answer is a partial answer rather than a
// short number.
//
// The filter types are this package's own. official.OrdersFilter cannot be named
// here — internal/official is a banned import, and it is banned because it is the
// client that can place, amend and cancel — and the filters are applied to the
// cached reading anyway (D6), never passed to the broker.
type OrdersReader interface {
	// Orders returns one reading of both order lists.
	Orders(ctx context.Context) (OrdersReading, error)
}

// OrdersReading is one reading of the account's order surface.
//
// Every number in it is the broker's own string. A value that has been through
// float64 and back cannot say "the broker sent nothing here", and the API sends
// price as null for a market order and the whole execution object as null while
// an order is alive — so on this screen an absent value is the normal case, not
// the edge one, and rendering it as 0 says the order filled.
//
// # Three lists, and the first one cannot be short
//
// The plain endpoint's two groups are separate readings because they behave
// differently in the only way that matters here. `status=OPEN` returns every
// pending order — "limit, cursor 는 무시되며" — so the live list is structurally
// whole and its count is a number. `status=CLOSED` paginates, so that list is a
// page and its count can be a floor. Merging them would put the closed page's
// truncation on the live count and turn an exact answer into "N건 이상".
//
// Each list carries its own failure rather than one shared error. Losing the
// conditional list while the plain ones answered is the case the whole screen is
// built around, and it cannot be expressed by returning an error for the set.
type OrdersReading struct {
	// AccountRef is the exact account resolved for all three broker reads. It is
	// required to scope journal evidence; an empty value makes origin unknown.
	AccountRef string
	// Open is the broker's OPEN group of /api/v1/orders: every pending order,
	// whole. This is the list a leftover cannot hide from.
	Open []OrderRecord
	// OpenError is why there is no open list, empty when there is one.
	OpenError string
	// OpenTruncated reports that the broker said there is another page of
	// pending orders. Per the documented contract it is always false — OPEN
	// ignores limit and cursor and returns the lot. It exists so that a broker
	// contradicting its own documentation is reported rather than assumed away:
	// the screen would rather say "N건 이상" than assert an exactness it was just
	// told it does not have.
	OpenTruncated bool

	// Closed is one page of the CLOSED group. It is a separate call and not a
	// luxury: a cancelled or rejected order never becomes a trade, so /history
	// never shows it, and this is the only place that fact exists.
	Closed []OrderRecord
	// ClosedError is why there is no closed list.
	ClosedError string
	// ClosedTruncated reports that the broker said there is another page. The
	// count is then "N건 이상": a number taken off a truncated page is a
	// confidently short one.
	ClosedTruncated bool

	// Conditional is the broker's OPEN group of /api/v1/conditional-orders —
	// the live ones, which is exactly the set that fills the exposure cap.
	Conditional []ConditionalRecord
	// ConditionalError is why there is no conditional list.
	ConditionalError string
	// ConditionalTruncated reports another page of them.
	ConditionalTruncated bool
}

// OrderRecord is one plain order, spelled as the broker spelled it.
//
// Every field is a string, and an empty one means the payload carried no value —
// which is a different fact from zero and is rendered differently.
type OrderRecord struct {
	ID       string
	Symbol   string
	Side     string
	Kind     string
	Status   string
	Market   string
	Currency string

	Quantity           string
	Price              string
	FilledQuantity     string
	AverageFilledPrice string

	OrderedAt  string
	CanceledAt string
}

// ConditionalRecord is one conditional order, spelled as the broker spelled it.
//
// It carries the first leg's numbers only. One row is one conditional order, and
// an OCO's second leg needs a shape of its own; an aggregate of two legs would be
// a number nobody sent.
type ConditionalRecord struct {
	ID     string
	Symbol string
	Market string
	Kind   string
	Status string

	Quantity      string
	TriggerPrice  string
	OrderPrice    string
	ConditionKind string
	// Triggered names the plain order this conditional became, empty while it is
	// still watching.
	Triggered  string
	ExpireDate string
	CreatedAt  string
}

// --- the cache -------------------------------------------------------------------

// ordersTTL is how long one reading is served from memory.
//
// The floor the operator-console spec fixes is 15 seconds and this is exactly
// that, which is shorter than holdingsTTL on purpose: a position's size does not
// move between page loads and a live order's remaining quantity does, and this is
// the screen somebody watches while deciding whether a leftover is gone yet.
const ordersTTL = 15 * time.Second

// ordersTimeout bounds one refresh, for holdings.go's reason: a broker that has
// stopped answering must not hold an HTTP handler open.
const ordersTimeout = 10 * time.Second

// ordersSnapshot is one reading of the order record, as the page sees it.
type ordersSnapshot struct {
	// Wired reports that a reader was supplied at all.
	Wired bool
	// Present reports that a reading exists, possibly an old one.
	Present bool
	// Lists is the reading itself.
	Lists OrdersReading
	// At is when it was taken and Age how old it is now.
	At  time.Time
	Age time.Duration
	// Error is the last whole-refresh failure — the seam itself returning an
	// error, as opposed to one of the two lists failing inside a reading that
	// otherwise arrived.
	Error string
	// Held reports that the refresh was withheld, HeldReason by what.
	Held       bool
	HeldReason string
}

// TakenAt renders the reading's timestamp.
func (s ordersSnapshot) TakenAt() string {
	if !s.Present {
		return "—"
	}
	return s.At.UTC().Format("2006-01-02 15:04:05Z")
}

// AgeSeconds is the reading's age, rounded to a whole second.
func (s ordersSnapshot) AgeSeconds() int { return int(s.Age.Round(time.Second) / time.Second) }

// TTLSeconds is the age at which a reading stops being a measurement, so the
// screens can print the bound they are judging against rather than a bare verdict.
func (ordersSnapshot) TTLSeconds() int { return int(ordersTTL / time.Second) }

// Stale reports a reading older than the TTL.
//
// Past the TTL a reading is not a measurement (design D9). On this screen that is
// rare — the refresh is lazy but it happens on every request outside a hold — and
// on the overview it is the normal case, because that screen refreshes nothing by
// contract and the count therefore freezes at whatever the last /orders visit
// produced. A three-hour-old "0건" rendered as a measured number is a screen
// telling an operator the account is quiet while an order sits on it.
func (s ordersSnapshot) Stale() bool { return s.Present && s.Age > ordersTTL }

// ordersCache is the lazy, single-flight cache in front of one OrdersReader.
//
// It is holdingsCache's arrangement, and for the same reasons: the mutex is held
// across the fetch so two tabs cost one refresh rather than two, and the TTL
// bounds attempts rather than successes so a broker answering 429 is not asked
// again on every page load.
type ordersCache struct {
	reader OrdersReader
	ttl    time.Duration

	mu      sync.Mutex
	lists   OrdersReading
	at      time.Time
	present bool
	lastErr string
	// tried is when the last refresh was attempted, successful or not.
	tried     time.Time
	attempted bool
}

func newOrdersCache(reader OrdersReader, ttl time.Duration) *ordersCache {
	if ttl <= 0 {
		ttl = ordersTTL
	}
	return &ordersCache{reader: reader, ttl: ttl}
}

// get returns the current reading, refreshing it if it is older than the TTL.
//
// hold is the verification-in-progress signal and it is the same judgement the
// holdings cache uses — c.verifyHold, which reads this process's run directly and
// another process's through internal/runlock's marker with its staleness bound.
// A second definition of "a verification is running" would be a second thing to
// keep true.
func (c *ordersCache) get(ctx context.Context, now time.Time, hold bool, holdReason string) ordersSnapshot {
	if c == nil || c.reader == nil {
		return ordersSnapshot{Held: hold, HeldReason: holdReason}
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if !hold && (!c.attempted || now.Sub(c.tried) >= c.ttl) {
		c.refreshLocked(ctx, now)
	}

	snap := c.snapshotLocked(now)
	snap.Held, snap.HeldReason = hold, holdReason
	return snap
}

// peek returns the current reading and refreshes nothing, whatever its age.
//
// The overview screen reads this one. Its contract is zero broker calls per
// render (console-operator-overview D4), so it must not be able to trigger the
// two this screen costs — the longest-lived tab in the console must not be the
// one spending the budget.
func (c *ordersCache) peek(now time.Time) ordersSnapshot {
	if c == nil || c.reader == nil {
		return ordersSnapshot{}
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.snapshotLocked(now)
}

func (c *ordersCache) snapshotLocked(now time.Time) ordersSnapshot {
	snap := ordersSnapshot{
		Wired:   true,
		Present: c.present,
		Lists:   c.lists,
		At:      c.at,
		Error:   c.lastErr,
	}
	if c.present {
		snap.Age = now.Sub(c.at)
		if snap.Age < 0 {
			snap.Age = 0
		}
	}
	return snap
}

// refreshLocked performs the one seam call a refresh is allowed. Behind it the
// adapter makes three broker calls: the broker's pending group, its finished
// group and the conditional endpoint.
//
// A failure keeps the previous reading rather than discarding it, exactly as the
// holdings cache does: an operator looking at a two-minute-old order list with
// the failure printed beside it is better informed than one looking at an empty
// table, and the age is on the page either way.
func (c *ordersCache) refreshLocked(ctx context.Context, now time.Time) {
	c.tried, c.attempted = now, true

	fetchCtx, cancel := context.WithTimeout(ctx, ordersTimeout)
	defer cancel()

	lists, err := c.reader.Orders(fetchCtx)
	if err != nil {
		c.lastErr = err.Error()
		return
	}
	c.lists, c.at, c.present, c.lastErr = lists, now, true, ""
}

// --- the rows ---------------------------------------------------------------------

// orderOrigin is who issued an order, and there are three answers.
type orderOrigin string

const (
	// originEngine: the ledger has a mutation attempt whose broker order id is
	// this one.
	originEngine orderOrigin = "engine"
	// originOther: the ledger was read and does not know this order.
	originOther orderOrigin = "other"
	// originUnknown: the ledger could not be read. It is deliberately not
	// originOther — reading an unreadable ledger as "a person placed all of
	// these" makes a working engine look idle, and an idle engine is what an
	// operator restarts.
	originUnknown orderOrigin = "unknown"
)

// orderRow is one line of the table, plain or conditional.
//
// Every cell is already rendered, because the decision "this value is absent, so
// it is — and not 0" is made here where the broker's own string is still
// available, and not in a template where every empty string looks the same.
type orderRow struct {
	// Conditional reports the row came from the conditional endpoint. The two
	// kinds are shown together in the live section — they fill the same exposure
	// cap — and marked, because cancelling them takes different calls.
	Conditional bool

	At       string
	AtRaw    string
	Symbol   string
	Market   string
	Side     string
	Status   string
	Quantity string
	Filled   string
	Price    string
	AvgPrice string
	ID       string

	// Live reports the order can still change at the broker; Unresolved reports
	// that this build could not tell, which is not the same as closed.
	Live       bool
	Unresolved bool
	// Detail is the derivation's own explanation, shown only when Unresolved.
	Detail string

	Origin orderOrigin

	// ExitEvidence is true only when the journal supplies explicit
	// broker_order_id -> attempt.intent_id -> exit_event.proposed_intent_id
	// lineage. Symbol, price and time are never used to populate these fields.
	ExitEvidence  bool
	ExitLine      operatorview.ExitLineView
	ExitAttemptID string
	ExitIntentID  string

	EvidenceKey      orderEvidenceKey
	EvidenceKeyValid bool
}

// OriginText is the label for the origin column.
func (r orderRow) OriginText() string {
	switch r.Origin {
	case originEngine:
		return "엔진 발주"
	case originOther:
		return "그 밖"
	default:
		return "불명"
	}
}

// StateText is the label for the live/closed distinction.
func (r orderRow) StateText() string {
	switch {
	case r.Unresolved:
		return "상태 불명"
	case r.Live:
		return "미체결"
	default:
		return "종결"
	}
}

// dashUnlessSent renders a broker value, and an em dash when the broker sent
// none.
//
// This is the whole reason the raw read exists: "" is "the payload carried no
// value" and it must not become 0, because an average fill price of 0 on a live
// order says it filled.
//
// A value that is not a decimal is printed verbatim rather than grouped.
// groupDecimalText inserts thousands separators positionally, so it would turn a
// string this build did not expect into a different string — and the one thing
// this whole path exists to preserve is what the broker actually said.
func dashUnlessSent(s string) string {
	trimmed := strings.TrimSpace(s)
	if trimmed == "" {
		return "—"
	}
	if _, err := strconv.ParseFloat(trimmed, 64); err != nil {
		return trimmed
	}
	return groupDecimalText(trimmed)
}

// orderInstant renders an order's submission time.
//
// The broker's orderedAt is the actual instant. A value that will not parse is
// shown verbatim rather than dropped or replaced with a dash: an unparseable
// timestamp is still evidence of when the broker thinks the order was sent, and
// hiding it would make the row look like one with no time at all.
func orderInstant(raw string) (shown, unparsed string) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "—", ""
	}
	if t, err := time.Parse(time.RFC3339, trimmed); err == nil {
		return t.UTC().Format("2006-01-02 15:04:05Z"), ""
	}
	return trimmed, trimmed
}

// decimalFor parses one of the broker's decimal strings for the state
// derivation.
//
// Only internal/brokerstate reads the result. The rendered value is always the
// original string, so an unparseable number degrades into an unresolved STATE
// rather than into a fabricated zero on the screen.
func decimalFor(s string) float64 {
	v, err := strconv.ParseFloat(strings.TrimSpace(s), 64)
	if err != nil {
		return 0
	}
	return v
}

// rowFromOrder turns one plain order into a line.
//
// # pending is the broker's answer, not a derivation
//
// The live/closed split comes from WHICH GROUP the row was fetched in.
// `status=OPEN` is documented as "모든 대기 중 주문을 전량 반환" — the broker's own
// statement of what is still alive, and complete by construction — so it is the
// authority for the count. Deriving aliveness from the per-order status string
// instead would put a second definition of "still open" on this screen, and the
// two would disagree exactly when the status is one this build has not seen: the
// leftover would be filed under 종결 by a screen built to find it.
//
// # internal/brokerstate still judges the status column
//
// That package models the ten-value OrderStatus enum, the contradictions it makes
// expressible, and the fail-closed UNKNOWN for anything it cannot explain. Its
// verdict is what the 상태 불명 label reports.
//
// When the two disagree in the dangerous direction — the broker returned the row
// in the pending group and the status string reads as terminal — the row is marked
// unresolved and stays in the live count. Fail-closed here means "it can still
// change". The opposite disagreement is not flagged: PARTIAL_FILLED is documented
// as belonging to BOTH groups, so a closed-group row whose status reads as open is
// the API's own overlap rather than a contradiction.
func rowFromOrder(rec OrderRecord, origin orderOrigin, pending bool) orderRow {
	shown, unparsed := orderInstant(rec.OrderedAt)
	row := orderRow{
		At:       shown,
		AtRaw:    unparsed,
		Symbol:   rec.Symbol,
		Market:   marketLabel(rec.Market),
		Side:     sideLabel(rec.Side),
		Status:   dashIfEmpty(rec.Status),
		Quantity: dashUnlessSent(rec.Quantity),
		Filled:   dashUnlessSent(rec.FilledQuantity),
		Price:    dashUnlessSent(rec.Price),
		AvgPrice: dashUnlessSent(rec.AverageFilledPrice),
		ID:       dashIfEmpty(rec.ID),
		Origin:   origin,
	}

	view := brokerstate.OrderView{
		OrderID:        rec.ID,
		RawStatus:      rec.Status,
		Canceled:       strings.TrimSpace(rec.CanceledAt) != "",
		Quantity:       decimalFor(rec.Quantity),
		FilledQuantity: decimalFor(rec.FilledQuantity),
	}
	derived := brokerstate.Derive(view)
	row.Live = pending
	row.Unresolved = derived.FailClosed
	switch {
	case derived.FailClosed:
		row.Detail = derived.Detail
	case pending && derived.Terminal:
		row.Unresolved = true
		row.Detail = "브로커는 이 주문을 미체결(OPEN) 그룹으로 돌려줬는데 상태 문자열 " +
			strings.TrimSpace(rec.Status) + " 은(는) 종결로 읽힌다. 둘 중 하나가 틀렸고, " +
			"종결로 읽지 않는다 — 아직 바뀔 수 있는 주문을 끝난 것으로 세면 잔여물이 사라진다."
	}
	return row
}

// rowFromConditional turns one conditional order into a line.
//
// It is always live. The list is the broker's OPEN group — the request asked for
// exactly the conditionals that are still watching — so the console is not
// deriving a state here, it is reporting which question it asked. Its own status
// string is in the status column, unmodified, because that vocabulary (WATCHING)
// is not the plain order enum and this build must not pretend to map it.
//
// Most of a plain order's columns have no counterpart, and they render as —
// rather than as zero: a conditional order has no fill and no average fill price
// because it has not become an order yet.
func rowFromConditional(rec ConditionalRecord, origin orderOrigin) orderRow {
	shown, unparsed := orderInstant(rec.CreatedAt)
	return orderRow{
		Conditional: true,
		At:          shown,
		AtRaw:       unparsed,
		Symbol:      rec.Symbol,
		Market:      marketLabel(rec.Market),
		Side:        "—",
		Status:      dashIfEmpty(rec.Status),
		Quantity:    dashUnlessSent(rec.Quantity),
		Filled:      "—",
		Price:       dashUnlessSent(firstNonBlank(rec.OrderPrice, rec.TriggerPrice)),
		AvgPrice:    "—",
		ID:          dashIfEmpty(rec.ID),
		Live:        true,
		Origin:      origin,
	}
}

func firstNonBlank(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}

// marketUnknownLabel is what a row whose market this build could not name renders
// as. It is a value rather than a literal because the market filter has to be able
// to recognise it: an unknown market is not-applicable to a filter, never
// excluded by one.
const marketUnknownLabel = "—"

// marketLabel normalises the broker's market to the two this console names, and
// says so when it is neither.
//
// The plain order endpoint has no market field at all — it is derived from the
// currency — so an unrecognised currency arrives here as empty, and empty is
// rendered as unknown rather than silently filed under one of the two.
func marketLabel(market string) string {
	switch strings.ToUpper(strings.TrimSpace(market)) {
	case "KR":
		return "KR"
	case "US":
		return "US"
	case "":
		return marketUnknownLabel
	default:
		return strings.ToUpper(strings.TrimSpace(market))
	}
}

// sideLabel renders BUY/SELL in the words the rest of the console uses.
func sideLabel(side string) string {
	switch strings.ToUpper(strings.TrimSpace(side)) {
	case "BUY":
		return "매수"
	case "SELL":
		return "매도"
	case "":
		return "—"
	default:
		return strings.ToUpper(strings.TrimSpace(side))
	}
}

// --- the origin join --------------------------------------------------------------

// ledgerOrderEvidence reads origin ids and exit lineage from the same read-only
// journal handle. A partial journal answer is not treated as trustworthy: both
// the origin and the evidence labels fall back to unknown when either read
// fails.
type orderEvidenceKey struct {
	ID, AccountRef, Market, TradingDay string
}

func evidenceKey(id, accountRef, market, rawAt string) (orderEvidenceKey, bool) {
	accountRef = strings.TrimSpace(accountRef)
	if len(id) == 0 || accountRef == "" {
		return orderEvidenceKey{}, false
	}
	at, err := time.Parse(time.RFC3339, strings.TrimSpace(rawAt))
	if err != nil {
		return orderEvidenceKey{}, false
	}
	m, err := marketclock.ParseMarket(market)
	if err != nil {
		return orderEvidenceKey{}, false
	}
	day, err := m.TradingDay(at)
	if err != nil {
		return orderEvidenceKey{}, false
	}
	return orderEvidenceKey{ID: id, AccountRef: accountRef, Market: string(m), TradingDay: day}, true
}

// orderDedupeKey is the broker's exact order identity used only to collapse the
// documented OPEN/CLOSED overlap. A valid identity uses the same account,
// canonical market and market-local trading day as evidence lookup. If those
// fields cannot be derived, the raw market/time remain in a tagged fallback so
// an invalid row cannot hide a different row that merely reuses its opaque id.
type orderDedupeKey struct {
	ID, AccountRef, Market, TradingDay string
	RawAt                              string
	Fallback                           bool
}

func dedupeKey(id, accountRef, market, rawAt string) (orderDedupeKey, bool) {
	if len(id) == 0 {
		return orderDedupeKey{}, false
	}
	if evidence, valid := evidenceKey(id, accountRef, market, rawAt); valid {
		return orderDedupeKey{ID: evidence.ID, AccountRef: evidence.AccountRef,
			Market: evidence.Market, TradingDay: evidence.TradingDay}, true
	}
	canonicalMarket := market
	if parsed, err := marketclock.ParseMarket(market); err == nil {
		canonicalMarket = string(parsed)
	}
	return orderDedupeKey{ID: id, AccountRef: strings.TrimSpace(accountRef),
		Market: canonicalMarket, RawAt: rawAt, Fallback: true}, true
}

func visibleOrderEvidenceScopes(rows []orderRow) []journal.BrokerOrderScope {
	seen := map[orderEvidenceKey]bool{}
	var scopes []journal.BrokerOrderScope
	for _, row := range rows {
		key := row.EvidenceKey
		if !row.EvidenceKeyValid || seen[key] {
			continue
		}
		seen[key] = true
		scopes = append(scopes, journal.BrokerOrderScope{BrokerOrderID: key.ID,
			AccountRef: key.AccountRef, Market: key.Market, TradingDay: key.TradingDay})
	}
	return scopes
}

func (c *Console) ledgerOrderEvidence(ctx context.Context,
	scopes []journal.BrokerOrderScope,
) (map[orderEvidenceKey]bool,
	map[orderEvidenceKey]journal.BrokerOrderExitLink, journalView,
) {
	ro, jv := c.openJournal(ctx)
	if ro == nil {
		return nil, nil, jv
	}
	defer ro.Close()

	if len(scopes) == 0 {
		return map[orderEvidenceKey]bool{}, map[orderEvidenceKey]journal.BrokerOrderExitLink{}, jv
	}
	links, err := ro.BrokerOrderExitLinks(ctx, scopes)
	if err != nil {
		jv.State, jv.Detail = journalFailed, err.Error()
		return nil, nil, jv
	}
	engineIDs := make(map[orderEvidenceKey]bool, len(links))
	byOrder := make(map[orderEvidenceKey]journal.BrokerOrderExitLink, len(links))
	for _, link := range links {
		key := orderEvidenceKey{ID: link.BrokerOrderID, AccountRef: link.AccountRef,
			Market: link.Market, TradingDay: link.TradingDay}
		engineIDs[key] = link.Engine
		if link.Event.ID != 0 || link.Ambiguous {
			byOrder[key] = link
		}
	}
	return engineIDs, byOrder, jv
}

// attachOrderExitEvidence maps an immutable, already-selected event snapshot.
// It performs no lookup and no policy evaluation; absence is rendered as an
// explicit unlinked state by the template.
func attachOrderExitEvidence(row *orderRow, link journal.BrokerOrderExitLink, linked bool) {
	if !linked {
		return
	}
	row.ExitEvidence = true
	row.ExitAttemptID = link.AttemptID
	row.ExitIntentID = link.IntentID
	if link.Ambiguous {
		row.ExitLine = operatorview.BuildExitLine(operatorview.Source{
			UnknownReason: link.UnknownReason,
		})
		return
	}
	effective := link.Event.Evaluation.Effective
	source := operatorview.Source{
		UnknownReason:   effective.UnknownReason,
		EffectiveSource: link.Event.Evaluation.EffectiveSource,
	}
	if effective.Snapshot != nil {
		source.Snapshot = &effective.Snapshot.Line
		source.ObservationSource = effective.Snapshot.ObservationSource
		source.ObservedAt = effective.Snapshot.ObservedAt
	}
	row.ExitLine = operatorview.BuildExitLine(source)
}

// originOf decides one plain order's origin from the ledger's answer.
func originOf(key orderEvidenceKey, valid bool, engineIDs map[orderEvidenceKey]bool, jv journalView) orderOrigin {
	if !jv.Readable() {
		return originUnknown
	}
	if !valid {
		return originUnknown
	}
	if valid && engineIDs[key] {
		return originEngine
	}
	return originOther
}

// conditionalOriginOf decides one conditional order's origin, and it is a
// different function because the ledger answers a different amount of the
// question.
//
// # The ledger was never asked about conditionals
//
// mutation_attempts is written by internal/execgw, whose three kinds are PLACE,
// CANCEL and AMEND of a PLAIN order; its broker_order_id is what PlaceOrder,
// CancelOrder and ModifyOrder handed back. Registering a conditional order goes
// through trading.Service.ConditionalPlace and internal/verifylive, neither of
// which opens a journal attempt at all. So in this build a conditional order's id
// is never written to that column, and its ABSENCE from the set is not evidence of
// anything: "그 밖" would be a verdict on a question nobody put to the ledger.
//
// # And the id it used to be asked about was always empty
//
// The first implementation joined on rec.Triggered — the plain order a conditional
// turned into — which is empty for exactly as long as the conditional is still
// watching, and the adapter requests only the OPEN group. The lookup was therefore
// engineIDs[""] on every row the screen can show, and BrokerOrderIDs excludes the
// empty id by design. The label was a constant wearing a determination's clothes,
// and the constant was the wrong one: an invented "manual" label on an engine
// order is an operator concluding the engine is idle while it is trading.
//
// A positive hit on the conditional's own composite identity is still honoured.
// The triggered plain id is deliberately not tried here: its actual submission
// instant is not in this payload, and borrowing CreatedAt would guess a trading
// day. If it is visible as a plain row, that row is scoped from its own timestamp.
// A miss is 불명.
func conditionalOriginOf(key orderEvidenceKey, valid bool,
	engineIDs map[orderEvidenceKey]bool, jv journalView,
) orderOrigin {
	if !jv.Readable() {
		return originUnknown
	}
	if valid && engineIDs[key] {
		return originEngine
	}
	return originUnknown
}

// --- the counts -------------------------------------------------------------------

// countText renders a row count, with "이상" when the page was truncated.
func countText(n int, truncated bool) string {
	if truncated {
		return strconv.Itoa(n) + "건 이상"
	}
	return strconv.Itoa(n) + "건"
}

// listUnmeasured maps one list's failure onto the console's reason vocabulary.
//
// A wired seam that failed to read is broker_read_failed; an unwired one is
// seam_unwired; a cache nothing has filled yet is never_fetched; and a refresh
// withheld for a running verification is verify_suspended. All four are already
// in the enumeration because all four ask the operator for a different thing.
func listUnmeasured(snap ordersSnapshot, listErr string) reading {
	switch {
	case !snap.Wired:
		return unmeasuredFor(reasonSeamUnwired)
	case !snap.Present && snap.Held:
		return unmeasuredFor(reasonVerifySuspended)
	case !snap.Present && snap.Error != "":
		return unmeasuredWithDetail(reasonBrokerReadFailed, snap.Error)
	case !snap.Present:
		return unmeasuredFor(reasonNeverFetched)
	case listErr != "":
		return unmeasuredWithDetail(reasonBrokerReadFailed, listErr)
	default:
		return measured("")
	}
}

// --- assembly ---------------------------------------------------------------------

// ordersView is the whole screen.
type ordersView struct {
	Now time.Time
	// Broker is the cache reading behind the lists, for its timestamp and its
	// withheld-refresh notice.
	Broker ordersSnapshot
	// Journal is the ledger's one verdict for this render. The origin column
	// reads it and the page prints the notice once, never per row.
	Journal journalView

	// Open, Closed and Conditional are each list's own measurement. They are
	// separate because losing one of the three is the case this screen is built
	// around.
	Open        reading
	Closed      reading
	Conditional reading

	// OpenLive and ConditionalLive are the live counts of the two live lists, and
	// Live is the two added — which happens only when BOTH were measured. A total
	// that silently drops the conditional half is the confidently short number
	// that hides a leftover.
	//
	// OpenLive is a number and not a floor: it is the size of the broker's OPEN
	// group, which the API returns whole.
	OpenLive        reading
	ConditionalLive reading
	Live            reading
	// ClosedCount is how many finished orders the closed page carried. It can be
	// a floor — that list paginates.
	ClosedCount reading

	// Rows is what the table shows after the filter, live first.
	Rows []orderRow
	// Shown and Total are the filtered count and the whole count, so a filtered
	// screen cannot read as "these are all the orders".
	Shown int
	Total int
	// Filtered reports a filter is applied at all.
	Filtered bool
	// Truncated reports that at least one list had another page, so Total is a
	// floor rather than a number.
	Truncated bool

	// Filters are the link sets the page renders. Enabled is false when neither
	// list was measured: "0/—건" reads as "0 matched", so a filter that cannot
	// mean anything is not offered.
	Filters []orderFilterGroup
	Enabled bool

	// Selected is the filter in effect, for the page's own summary.
	Selected orderFilterChoice
}

// NowText renders the instant the screen was built.
func (v ordersView) NowText() string { return v.Now.UTC().Format("2006-01-02 15:04:05Z") }

// AnyUnresolved reports at least one row whose state this build could not
// derive, so the page can explain the label once instead of per row.
func (v ordersView) AnyUnresolved() bool {
	for _, r := range v.Rows {
		if r.Unresolved {
			return true
		}
	}
	return false
}

// AnyConditionalOrigin reports at least one conditional row on a page whose ledger
// WAS readable, so the screen can explain once why those rows can still say 불명.
//
// Without the sentence an unexplained 불명 sitting beside an explained 엔진 발주
// reads as a bug in the join, and the next person deletes the honesty to make the
// column look consistent.
func (v ordersView) AnyConditionalOrigin() bool {
	if !v.Journal.Readable() {
		return false
	}
	for _, r := range v.Rows {
		if r.Conditional {
			return true
		}
	}
	return false
}

// AnyUnparsedTime reports at least one row whose timestamp would not parse.
func (v ordersView) AnyUnparsedTime() bool {
	for _, r := range v.Rows {
		if r.AtRaw != "" {
			return true
		}
	}
	return false
}

// orders builds the screen.
//
// Every panel is a state the page has words for; nothing here returns an error.
func (c *Console) orders(ctx context.Context, choice orderFilterChoice) ordersView {
	now := c.now()
	hold, holdReason := c.verifyHold(now)

	v := ordersView{Now: now, Selected: choice}
	v.Broker = c.ordersCache.get(ctx, now, hold, holdReason)

	lists := v.Broker.Lists
	v.Open = listUnmeasured(v.Broker, lists.OpenError)
	v.Closed = listUnmeasured(v.Broker, lists.ClosedError)
	v.Conditional = listUnmeasured(v.Broker, lists.ConditionalError)

	var all []orderRow
	openLive := 0
	// pending is the set of ids the broker just called pending. The closed page
	// is filtered against it because PARTIAL_FILLED is documented as belonging to
	// BOTH groups, so one order can arrive in both answers — and two rows for one
	// order would be one order counted twice and one row an operator cancels
	// twice.
	pending := make(map[orderDedupeKey]bool)
	if v.Open.Known() {
		for _, rec := range lists.Open {
			key, valid := evidenceKey(rec.ID, lists.AccountRef, rec.Market, rec.OrderedAt)
			row := rowFromOrder(rec, originUnknown, true)
			row.EvidenceKey, row.EvidenceKeyValid = key, valid
			all = append(all, row)
			if identity, ok := dedupeKey(rec.ID, lists.AccountRef, rec.Market, rec.OrderedAt); ok {
				pending[identity] = true
			}
			openLive++
		}
	}
	closedCount := 0
	if v.Closed.Known() {
		for _, rec := range lists.Closed {
			identity, dedupe := dedupeKey(rec.ID, lists.AccountRef, rec.Market, rec.OrderedAt)
			if dedupe && pending[identity] {
				continue
			}
			key, valid := evidenceKey(rec.ID, lists.AccountRef, rec.Market, rec.OrderedAt)
			row := rowFromOrder(rec, originUnknown, false)
			row.EvidenceKey, row.EvidenceKeyValid = key, valid
			all = append(all, row)
			closedCount++
		}
	}
	conditionalLive := 0
	if v.Conditional.Known() {
		for _, rec := range lists.Conditional {
			key, valid := evidenceKey(rec.ID, lists.AccountRef, rec.Market, rec.CreatedAt)
			row := rowFromConditional(rec, originUnknown)
			row.EvidenceKey, row.EvidenceKeyValid = key, valid
			all = append(all, row)
			conditionalLive++
		}
	}

	v.Truncated = (v.Open.Known() && lists.OpenTruncated) ||
		(v.Closed.Known() && lists.ClosedTruncated) ||
		(v.Conditional.Known() && lists.ConditionalTruncated)

	v.OpenLive = countReading(v.Open, openLive, lists.OpenTruncated)
	v.ClosedCount = countReading(v.Closed, closedCount, lists.ClosedTruncated)
	v.ConditionalLive = countReading(v.Conditional, conditionalLive, lists.ConditionalTruncated)
	v.Live = combinedLive(v.Open, v.Conditional, openLive, conditionalLive,
		lists.OpenTruncated || lists.ConditionalTruncated)

	sortRows(all)
	v.Total = len(all)
	v.Enabled = v.Open.Known() || v.Closed.Known() || v.Conditional.Known()
	v.Filtered = choice.Applied()
	v.Rows = filterRows(all, choice)
	v.Shown = len(v.Rows)
	scopes := visibleOrderEvidenceScopes(v.Rows)
	engineIDs, exitLinks, jv := c.ledgerOrderEvidence(ctx, scopes)
	v.Journal = jv
	for i := range v.Rows {
		row := &v.Rows[i]
		if row.Conditional {
			row.Origin = conditionalOriginOf(row.EvidenceKey, row.EvidenceKeyValid, engineIDs, jv)
		} else {
			row.Origin = originOf(row.EvidenceKey, row.EvidenceKeyValid, engineIDs, jv)
		}
		link, linked := exitLinks[row.EvidenceKey]
		attachOrderExitEvidence(row, link, linked)
	}
	v.Filters = filterGroups(choice)
	return v
}

// countReading renders one list's live count, or carries its unmeasured reason.
func countReading(measurement reading, n int, truncated bool) reading {
	if !measurement.Known() {
		return measurement
	}
	return measured(countText(n, truncated))
}

// combinedLive adds the two live counts, and refuses to when either list is
// unmeasured.
//
// This is the spec's "부분 실패는 합산하지 않는다" and it is the point of the
// screen. A total that quietly means "the plain orders only" is a measured-looking
// zero while a conditional leftover holds the exposure cap — the failure this
// change exists to prevent, committed by this change.
func combinedLive(open, conditional reading, openN, conditionalN int, truncated bool) reading {
	switch {
	case open.Known() && conditional.Known():
		return measured(countText(openN+conditionalN, truncated))
	case open.Known():
		return unmeasuredWithDetail(reasonFor(conditional),
			"일반 주문 "+countText(openN, truncated)+"은 읽었고 조건주문은 읽지 못했다. "+
				"둘을 합치지 않는다 — 조건주문도 노출 상한을 채우는 잔여물이다")
	case conditional.Known():
		return unmeasuredWithDetail(reasonFor(open),
			"조건주문 "+countText(conditionalN, truncated)+"은 읽었고 일반 주문은 읽지 못했다. "+
				"둘을 합치지 않는다")
	default:
		return open
	}
}

// reasonFor recovers an unmeasured reading's code so a combined value can carry
// the same one rather than inventing a new sentence.
func reasonFor(r reading) unmeasuredReason {
	return unmeasuredReason(r.Code())
}

// sortRows puts the live orders first and the newest first within each group.
//
// Live first because the question the screen answers is "is anything still
// alive"; an operator who has to scroll past this morning's fills to find out is
// reading the wrong screen. The order is total — id breaks a tie — so two renders
// of the same reading cannot disagree.
func sortRows(rows []orderRow) {
	sort.SliceStable(rows, func(i, j int) bool {
		a, b := rows[i], rows[j]
		if a.liveRank() != b.liveRank() {
			return a.liveRank() < b.liveRank()
		}
		if a.At != b.At {
			return a.At > b.At
		}
		return a.ID < b.ID
	})
}

func (r orderRow) liveRank() int {
	switch {
	case r.Unresolved:
		return 0
	case r.Live:
		return 1
	default:
		return 2
	}
}
