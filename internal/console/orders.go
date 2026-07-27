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
// # Two lists, never summed across a failure
//
// A conditional order is on a different endpoint from a plain one, and in this
// product it is the durable artefact rather than the corner case: M18 measured
// that one survives the process that registered it, and internal/verifylive's
// cleanup counts both kinds as leftovers filling the live-exposure cap. A screen
// that called only the plain endpoint would render "미체결 0건" as a measured
// value while a conditional leftover was blocking the next verification — which
// is the exact failure this screen exists to prevent.
//
// So one refresh is two calls, and if only one of them answers the screen says
// so. A confident zero has to be unreachable: the two counts are added only when
// both were measured.
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
)

// --- the seam --------------------------------------------------------------------

// OrdersReader is the console's entire view of the account's order record: one
// read, and it cannot change anything.
//
// It declares ONE method although a refresh is two broker calls. The two calls
// are the adapter's business (cmd/tossctl), and keeping them there is what stops
// this package from being able to spend the rate budget twice, or to spend it on
// one endpoint and report the other as zero: the outcome of each call arrives
// inside the value, so a partial answer is a partial answer rather than a short
// number.
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
// The two lists carry their own failures rather than one shared error. Losing the
// conditional list while the plain one answered is the case the whole screen is
// built around, and it cannot be expressed by returning an error for the pair.
type OrdersReading struct {
	// Plain is one page of /api/v1/orders, both live and finished.
	Plain []OrderRecord
	// PlainError is why there is no plain list, empty when there is one.
	PlainError string
	// PlainTruncated reports that the broker said there is another page. The
	// count is then "N건 이상": a number taken off a truncated page is a
	// confidently short one, and what goes missing is a leftover.
	PlainTruncated bool

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
// adapter makes two broker calls, one per endpoint.
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
// The live/closed split is internal/brokerstate's answer, not a switch written
// here. That package already models the ten-value OrderStatus enum, the
// contradictions it makes expressible, and the fail-closed UNKNOWN for anything
// it cannot explain; a second definition of "still open" on this screen would
// drift from the one the engine acts on, and the drift would show up as a
// leftover the screen calls finished.
func rowFromOrder(rec OrderRecord, origin orderOrigin) orderRow {
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
	row.Unresolved = derived.FailClosed
	row.Live = !derived.Terminal && !derived.FailClosed
	if derived.FailClosed {
		row.Detail = derived.Detail
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
		return "—"
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

// ledgerOrigins reads the set of broker order ids the engine's own attempts
// recorded.
//
// The second return value is the ledger's verdict, and the caller renders every
// row as 불명 when it is not readable. That is the whole reason this returns a
// verdict rather than an empty set: an empty set and an unread ledger produce the
// same map, and one of them means "the engine placed none of these" while the
// other means "nobody looked".
func (c *Console) ledgerOrigins(ctx context.Context) (map[string]bool, journalView) {
	ro, jv := c.openJournal(ctx)
	if ro == nil {
		return nil, jv
	}
	defer ro.Close()

	ids, err := ro.BrokerOrderIDs(ctx)
	if err != nil {
		jv.State, jv.Detail = journalFailed, err.Error()
		return nil, jv
	}
	out := make(map[string]bool, len(ids))
	for _, id := range ids {
		out[id] = true
	}
	return out, jv
}

// originOf decides one row's origin from the ledger's answer.
func originOf(id string, engineIDs map[string]bool, jv journalView) orderOrigin {
	if !jv.Readable() {
		return originUnknown
	}
	if engineIDs[strings.TrimSpace(id)] {
		return originEngine
	}
	return originOther
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

	// Plain and Conditional are each list's own measurement. They are separate
	// because losing one of the two is the case this screen is built around.
	Plain       reading
	Conditional reading

	// LiveCount and LiveTotal are the live counts of each list, and Live is the
	// two added — which happens only when BOTH were measured. A total that
	// silently drops the conditional half is the confidently short number that
	// hides a leftover.
	PlainLive       reading
	ConditionalLive reading
	Live            reading

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

	engineIDs, jv := c.ledgerOrigins(ctx)
	v.Journal = jv

	v.Plain = listUnmeasured(v.Broker, v.Broker.Lists.PlainError)
	v.Conditional = listUnmeasured(v.Broker, v.Broker.Lists.ConditionalError)

	var all []orderRow
	plainLive := 0
	if v.Plain.Known() {
		for _, rec := range v.Broker.Lists.Plain {
			row := rowFromOrder(rec, originOf(rec.ID, engineIDs, jv))
			if row.Live || row.Unresolved {
				plainLive++
			}
			all = append(all, row)
		}
	}
	conditionalLive := 0
	if v.Conditional.Known() {
		for _, rec := range v.Broker.Lists.Conditional {
			all = append(all, rowFromConditional(rec, originOf(rec.Triggered, engineIDs, jv)))
			conditionalLive++
		}
	}

	v.Truncated = (v.Plain.Known() && v.Broker.Lists.PlainTruncated) ||
		(v.Conditional.Known() && v.Broker.Lists.ConditionalTruncated)

	v.PlainLive = countReading(v.Plain, plainLive, v.Broker.Lists.PlainTruncated)
	v.ConditionalLive = countReading(v.Conditional, conditionalLive, v.Broker.Lists.ConditionalTruncated)
	v.Live = combinedLive(v.Plain, v.Conditional, plainLive, conditionalLive, v.Truncated)

	sortRows(all)
	v.Total = len(all)
	v.Enabled = v.Plain.Known() || v.Conditional.Known()
	v.Filtered = choice.Applied()
	v.Rows = filterRows(all, choice)
	v.Shown = len(v.Rows)
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
func combinedLive(plain, conditional reading, plainN, conditionalN int, truncated bool) reading {
	switch {
	case plain.Known() && conditional.Known():
		return measured(countText(plainN+conditionalN, truncated))
	case plain.Known():
		return unmeasuredWithDetail(reasonFor(conditional),
			"일반 주문 "+countText(plainN, truncated)+"은 읽었고 조건주문은 읽지 못했다. "+
				"둘을 합치지 않는다 — 조건주문도 노출 상한을 채우는 잔여물이다")
	case conditional.Known():
		return unmeasuredWithDetail(reasonFor(plain),
			"조건주문 "+countText(conditionalN, truncated)+"은 읽었고 일반 주문은 읽지 못했다. "+
				"둘을 합치지 않는다")
	default:
		return plain
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
