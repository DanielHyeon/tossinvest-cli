// Package candidatesrc wires real market-data clients to internal/candidate's
// Source interface.
//
// # Why this is a separate package
//
// internal/candidate must not see an order path — that is a spec requirement, not
// a preference, and section 6 enforces it by inspecting that package's import
// graph. But the clients that actually serve rankings and prices (internal/hybrid,
// internal/official, internal/client) also carry PlaceOrder, CancelOrder and the
// conditional-order mutations, because they are one client per backend.
//
// So the adapters live here. This package imports both sides; internal/candidate
// imports neither. The isolation is a property of that package's dependency
// graph, and this file is where the two worlds are allowed to meet — read-only,
// one method at a time.
//
// # The market is part of the panel, not a parameter to ignore
//
// Sources do not all serve both markets. WTS's popularity ranking is a Korean
// market product; the official ranking takes a market country. A source handed a
// market it cannot serve must be absent from that market's panel rather than
// present-and-empty, because "responded with no rows" is evidence a symbol left
// the list and "cannot see this market" is not. Discovery cools candidates on the
// first and must not on the second, so Panel decides membership per market.
package candidatesrc

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/candidate"
	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/domain"
	"github.com/JungHoonGhae/tossinvest-cli/internal/official"
)

// previousReadingTTL is how old the remembered reading may be and still be "the
// reading one tick ago".
//
// # Why the memory needs an expiry at all
//
// Existence was the whole condition when this was written, and existence survives
// an outage. A read that fails returns before the swap, so a source that spent an
// hour inside the 429 ladder comes back holding the set it held an hour ago — and
// every symbol still on the list is then answered `no`, which is the answer that
// qualifies a first sighting. Meanwhile cooling plus staleness is forty minutes, so
// that same outage has expired the entire watchlist and the recovering scan
// re-promotes the whole panel at once. The two meet exactly: the reading that
// re-stamps the panel is the one whose memory is least entitled to speak, and the
// verdict comes out measured. That is the path design D3 exists to cut, still open.
//
// # Where the number comes from
//
// Not from here. It is candidate.DefaultStalenessTTL, and it is that value for the
// two reasons that value has:
//
//	The gap it tolerates is the one the design guarantees. DefaultStalenessTTL and
//	MaxRankPriorAge are both twice the longest rung of candidate.BackoffLadder
//	(300s), on the argument written at the ladder: a retreat that long is normal
//	operation and must not expire what it happens during. A source inside its
//	backoff is not out of service, and its memory should survive the whole ladder.
//
//	Past it, the same span is already the one the rest of the system uses for this
//	shape of claim. candidate.nearFirstSighting is DefaultStalenessTTL too, and it
//	bounds how far a reading may sit from first_seen_at and still be the reading
//	that life was first seen in — so for any candidate whose life began before the
//	memory went stale, both NoteFirstRank and MeasureFirstSighting refuse the
//	position regardless of what this memory would have said about it.
//
//	The one case that leaves open is the life that begins on the recovering reading
//	itself. Its first_seen_at is that instant, so that window is wide open, and a
//	set recorded before the outage answers `no` about a symbol it last saw an hour
//	ago — the qualifying answer, arriving at the moment cooling plus staleness has
//	expired the whole watchlist and the scan is re-promoting it. That case is why
//	the bound is here rather than left to nearFirstSighting; the two together close
//	both ends. Beyond this bound the answer is `unknown`, which is the honest one:
//	nobody knows whether a symbol joined the list during an hour we were not
//	reading it.
//
//	Corrected 2026-07-28. This paragraph used to say the candidate the memory would
//	qualify is "gone anyway" past the bound, on the implicit cooling in
//	Store.stateAt. Cooled is not gone: stateAt cools at staleness and expires only
//	at staleness + DefaultCoolingTTL — forty minutes — and a cooled candidate keeps
//	first_seen_at, first_price and first_rank.
//
// Neither the scan interval nor the engine yield needs a term of its own: the
// official ranking's interval is 15s and the WTS list's 5s (cmd/tossctl's
// candidateIntervals), the yield doubles them, and every one of those is far inside
// the ladder's last rung.
const previousReadingTTL = candidate.DefaultStalenessTTL

// previousReading is one source's memory of one market's list.
//
// The instant is half the fact. A set with no instant cannot say whether it is the
// reading one tick ago or the reading before an outage, and those two produce the
// same `no` for a symbol that never left the list.
type previousReading struct {
	symbols map[string]bool
	at      time.Time
}

// usableAt reports that this memory may answer a question asked at `at`.
//
// Two ways to be unusable and both are ordinary. Older than the TTL is the outage
// above. Later than the reading asking — a clock stepping backwards — is refused
// rather than treated as age zero, because "the previous reading is in the future"
// is not a statement about the market either.
func (p previousReading) usableAt(at time.Time) bool {
	if p.symbols == nil || p.at.IsZero() || at.IsZero() {
		return false
	}
	age := at.Sub(p.at)
	return age >= 0 && age < previousReadingTTL
}

// RankingReader is the one official method the discovery adapters need.
//
// Declared narrowly here for the same reason the operator console declares
// HoldingsReader: what an adapter is handed decides what it can do, and a
// discovery adapter handed the whole official client could place an order in a
// future edit that would still compile.
type RankingReader interface {
	Rankings(ctx context.Context, typ, marketCountry, duration string,
		excludeCaution bool, count int) (domain.Ranking, error)
}

// BudgetReader is the rate-allowance accessor added to the official client for
// this change (official/ratebudget.go). Optional: a client that does not
// implement it simply reports nothing, which is recorded as unreported rather
// than as zero.
type BudgetReader interface {
	RateBudget(path string) official.RateBudget
}

// Official ranking types this package names. TOP_LOSERS is deliberately absent:
// discovery is looking for the start of a move up, and a losers list would raise
// candidates that the veto would then have to reject one at a time.
//
// Naming a type here is not the same as reading it. RankingTopGainers is declared,
// mapped to a source id below, and on no panel — see Panel for why it came off and
// what would put it back. Deleting it would take the argument with it.
const (
	RankingTradingAmount = "MARKET_TRADING_AMOUNT"
	RankingTradingVolume = "MARKET_TRADING_VOLUME"
	RankingTopGainers    = "TOP_GAINERS"
)

// rankingsPath is the request path the official client uses, and therefore the
// key its rate budget is filed under.
//
// Taken from internal/official rather than spelled again: a copy that drifted
// would make RateBudget() miss, and a miss returns an unreported budget — which
// is indistinguishable from "the server sent no headers". D13 decision 2's whole
// measurement would go dark and nothing would fail.
const rankingsPath = official.PathRankings

// rankingSourceID maps a ranking type to its own source id.
//
// One id per type, not one for the endpoint. The scan may only cool a candidate
// when every source that raised it answered, and that check is keyed by source
// id — so three rankings sharing an id let the two that answered mask the one
// that was rate limited, and a candidate only ever raised by the missing list got
// cooled by a scan that never looked at it. From there the cooling clock expires
// it and first_seen_at is gone.
var rankingSourceID = map[string]candidate.SourceID{
	RankingTradingAmount: candidate.SourceOfficialTradingValue,
	RankingTradingVolume: candidate.SourceOfficialTradingVolume,
	RankingTopGainers:    candidate.SourceOfficialGainers,
}

// officialRanking adapts one official ranking type to a candidate.Source.
type officialRanking struct {
	reader RankingReader
	budget BudgetReader
	typ    string
	id     candidate.SourceID
	count  int
	// seen is the previous reading's symbol set, per market. It is what makes the
	// new-entrant fact measurable at all: nothing else in the system holds "what
	// this list looked like last time", and the store cannot answer it because the
	// question is about one source's own consecutive readings rather than about the
	// symbol.
	//
	// Per market, and that is not tidiness. One adapter instance serves whichever
	// market it is handed, so a single set would let a KR reading decide whether a
	// US symbol had just joined the US list — an answer about a different list,
	// arriving as a measurement.
	//
	// A market with no entry has no previous reading, so its first reading reports
	// unknown for every row. That is the state after every process start, and it is
	// the one design D3 refuses to spell as "yes".
	//
	// Two conditions decide whether an entry may answer, and neither of them is
	// "it exists". It has to have arrived whole — which is decided against the set
	// that is actually kept, see rememberRead — and it has to be no older than
	// previousReadingTTL.
	seen map[string]previousReading
	mu   sync.Mutex
	// now is the time source the memory's age is measured against. Injected for
	// clock.System's usual reason: a bound nothing can drive to an instant is a
	// bound nothing tests.
	now clock.Clock
}

// OfficialRanking returns a Source over one official ranking type.
//
// count is capped at the API's documented maximum of 100 rather than passed
// through: a larger value is silently truncated by the server, and a caller that
// asked for 150 and got 100 would compute rank percentiles against a list length
// that never existed.
//
// The cap is why the requested count reported on every Row is o.count rather than
// the caller's argument: the comparison that decides whether a reading came back
// short has to be against the request that was actually made.
//
// An unrecognised ranking type is refused rather than given a fallback id, since
// a fallback would put it back into a collision with a real source.
//
// A nil clock is the system one. The clock is what dates the remembered reading,
// and a source that cannot date its own memory cannot tell the reading one tick ago
// from the reading before an outage.
func OfficialRanking(reader RankingReader, budget BudgetReader, typ string, count int,
	clk clock.Clock) (candidate.Source, error) {

	id, ok := rankingSourceID[typ]
	if !ok {
		return nil, fmt.Errorf("candidatesrc: unknown ranking type %q", typ)
	}
	if count <= 0 || count > 100 {
		count = 100
	}
	if clk == nil {
		clk = clock.System()
	}
	return &officialRanking{
		reader: reader, budget: budget, typ: typ, id: id, count: count, now: clk,
	}, nil
}

func (o *officialRanking) ID() candidate.SourceID { return o.id }

func (o *officialRanking) Read(ctx context.Context, market string) (candidate.Reading, error) {
	// "realtime" is the only duration that answers the question this change asks.
	// A 1d ranking is a summary of a move that has finished.
	raw, err := o.reader.Rankings(ctx, o.typ, market, "realtime", false, o.count)
	if err != nil {
		if errors.Is(err, official.ErrRateLimited) {
			// Reported as this package's own sentinel so the scan counts it as a
			// rate-limited loss. For the ranking that matters more than usual:
			// hybrid gives this endpoint no WTS fallback, so a 429 removes the
			// source outright rather than degrading it.
			return candidate.Reading{}, fmt.Errorf("%w: %s ranking: %v",
				candidate.ErrRateLimited, o.typ, err)
		}
		return candidate.Reading{}, fmt.Errorf("%s ranking: %w", o.typ, err)
	}

	total := len(raw.Items)
	// The previous reading's set is taken and replaced under one lock, before any
	// row is built, so two concurrent reads of the same market cannot each answer
	// from a set the other has already replaced.
	//
	// What goes in is the request, not the comparison. Whether this reading is whole
	// enough to become the memory is a question about the set that gets remembered,
	// and that set is built inside rememberRead — see the note there.
	previous, hadPrevious := o.rememberRead(market, raw.Items, o.count)
	rows := make([]candidate.Row, 0, total)
	for _, item := range raw.Items {
		symbol := strings.TrimSpace(item.Symbol)
		if symbol == "" {
			continue
		}
		rows = append(rows, candidate.Row{
			Symbol:    symbol,
			Rank:      item.Rank,
			RankTotal: total,
			// What was asked for, beside what arrived. The API caps count at 100 and
			// truncates silently, and this file's own comment on OfficialRanking has
			// warned since it was written that a caller who does not keep both numbers
			// computes percentiles against a list length that never existed. Keeping
			// it is what turns that warning into something a store can check.
			RankRequested: o.count,
			NewlyListed:   newlyListed(previous, hadPrevious, symbol),
			// domain.RankingItem is float64 throughout, so a field the API sent as
			// null already reads as 0 by the time it gets here and the distinction
			// is unrecoverable. These three are ranking-defining figures that a
			// ranking row always carries, so formatting them is honest. The field
			// where absent-versus-zero decides a veto — the intraday high — is not
			// in this response at all, and its absence is the absence of a whole
			// reading rather than of a value.
			TradingValue:  decimal(item.TradingAmount),
			TradingVolume: decimal(item.TradingVolume),
			Price:         decimal(item.LastPrice),
		})
	}
	return candidate.Reading{Rows: rows, Budget: o.rateBudget()}, nil
}

// rememberRead swaps in this reading's symbol set for `market` and returns the one
// it replaced, together with whether that one may be used.
//
// The two returns are the three states at their source. `usable` false is not an
// empty previous reading — an empty one would be evidence that every symbol here is
// new, and no usable previous reading at all is evidence of nothing.
//
// # A short reading does not become the previous reading
//
// The swap used to be unconditional, and that made a degraded reading the yardstick
// for the next whole one. Three rows arriving out of a hundred requested is not a
// list of three; it is a hundred-row list we saw three rows of. Replace the memory
// with those three and the next complete reading finds ninety-seven symbols "absent
// from the previous reading" and reports every one of them as a new entrant —
// measured, stored on the candidate row, and rendered as 신규 진입. This package's own
// newlylisted.go says the same thing one field over ("a truncated reading is not a
// list", design D4), and the comparison that decides it is the one the source
// already makes: requested against arrived.
//
// So a short reading leaves the memory alone. The next whole reading then compares
// against the last whole reading, which is older by however long the degradation
// lasted — and that is what the age bound on the memory is for.
//
// An empty reading is short by the same comparison and is left out for the same
// reason. It used to be swapped in deliberately, on the argument that "it saw
// nothing" is still what it saw; that argument holds only for a source that asked
// for nothing, and neither of these does.
//
// # The comparison is against the set that is kept, not the rows that arrived
//
// Corrected 2026-07-28. `whole` used to be computed by the caller as
// `o.count == len(raw.Items)`, and len(raw.Items) counts rows this function then
// drops: a row whose symbol is blank is absent from `current` and absent from the
// reading's rows, but it was still being counted towards the request. Three items
// of the three requested with two of them blank was therefore a whole reading whose
// memory held one symbol, and the next whole reading reported the other two as new
// entrants — measured, stored in candidates.first_rank_newly_listed, which is
// write-once. All three blank was worse: no rows at all, so the scan filed the
// reading under Missing, while the memory became an empty *non-nil* set that
// usableAt accepts, and every row of the next reading answered `yes`.
//
// So the request comes in and the comparison is made here, where `current` exists.
// Both halves are needed: the request has to match what arrived, and what arrived
// has to be what is kept. RankTotal is left as len(items) — the row count is what
// the endpoint served and it is the right percentile denominator; it is only the
// memory that must not be built out of a set that lost rows.
func (o *officialRanking) rememberRead(market string, items []domain.RankingItem, requested int) (
	map[string]bool, bool) {

	key := strings.ToUpper(strings.TrimSpace(market))
	current := make(map[string]bool, len(items))
	for _, item := range items {
		if symbol := strings.TrimSpace(item.Symbol); symbol != "" {
			current[symbol] = true
		}
	}
	whole := requested == len(items) && len(current) == len(items)
	at := o.now.Now()
	o.mu.Lock()
	defer o.mu.Unlock()
	if o.seen == nil {
		o.seen = map[string]previousReading{}
	}
	previous := o.seen[key]
	if whole {
		o.seen[key] = previousReading{symbols: current, at: at}
	}
	if !previous.usableAt(at) {
		return nil, false
	}
	return previous.symbols, true
}

// newlyListed is the three-state answer for one symbol.
//
// No previous reading means unknown, and that is the whole of design D2: "absent
// from the previous reading" is vacuously true of every row of a first reading, so
// spelling it `yes` would mark the entire panel as new every time the process
// starts. The bool that used to sit in this field could not say it.
func newlyListed(previous map[string]bool, had bool, symbol string) candidate.NewlyListed {
	if !had {
		return candidate.NewlyListedUnknown()
	}
	if previous[symbol] {
		return candidate.NewlyListedNo()
	}
	return candidate.NewlyListedYes()
}

// rateBudget translates the official client's reading into the scan's.
//
// A client without the accessor, or a response without the headers, produces a
// Budget with Reported false. That is the honest answer and it is not zero: a
// scheduler that read a missing header as "no calls left" would retreat from an
// endpoint that never asked it to.
func (o *officialRanking) rateBudget() candidate.Budget {
	if o.budget == nil {
		return candidate.Budget{}
	}
	b := o.budget.RateBudget(rankingsPath)
	if !b.Reported {
		return candidate.Budget{}
	}
	return candidate.Budget{
		Limit: b.Limit, Remaining: b.Remaining, Reset: b.Reset, Reported: true,
	}
}

// PopularityReader is the one WTS method the popularity adapter needs.
type PopularityReader interface {
	GetStockRanking(ctx context.Context, size int) (domain.StockRanking, error)
}

// wtsPopular adapts the WTS realtime popularity ranking.
type wtsPopular struct {
	reader PopularityReader
	size   int
	// seen is the previous reading's symbol set, keyed by market for
	// officialRanking's reason. This source serves only KR, so the map holds one
	// entry in practice — it is keyed anyway because the guard below is on Read and
	// a keyed map cannot be the thing that stops being true if that guard moves.
	//
	// Same two conditions as officialRanking's: whole, and no older than
	// previousReadingTTL.
	seen map[string]previousReading
	mu   sync.Mutex
	now  clock.Clock
}

// WTSPopular returns a Source over the WTS popularity ranking.
//
// It is an additive source. Everything downstream must keep working when it is
// gone, because a WTS session expiring is an ordinary event rather than an
// incident — which is why the official ranking, not this, is the one required to
// be sufficient alone.
func WTSPopular(reader PopularityReader, size int, clk clock.Clock) candidate.Source {
	if size <= 0 {
		size = 30
	}
	if clk == nil {
		clk = clock.System()
	}
	return &wtsPopular{reader: reader, size: size, now: clk}
}

func (w *wtsPopular) ID() candidate.SourceID { return candidate.SourceWTSPopular }

func (w *wtsPopular) Read(ctx context.Context, market string) (candidate.Reading, error) {
	// The guard belongs on the source, not only on Panel. A caller that builds a
	// panel by hand, or reuses a KR panel for a US scan, would otherwise file
	// Korean rows under US — and a scan treats a responding source as evidence
	// about the market it was asked for.
	if m := strings.ToUpper(strings.TrimSpace(market)); m != candidate.MarketKR {
		return candidate.Reading{}, fmt.Errorf(
			"candidatesrc: the WTS popularity ranking is a %s product and cannot serve %s",
			candidate.MarketKR, m)
	}
	raw, err := w.reader.GetStockRanking(ctx, w.size)
	if err != nil {
		return candidate.Reading{}, fmt.Errorf("wts popularity ranking: %w", err)
	}

	total := len(raw.Stocks)
	previous, hadPrevious := w.rememberRead(market, raw.Stocks, w.size)
	rows := make([]candidate.Row, 0, total)
	for _, s := range raw.Stocks {
		symbol := wtsSymbol(s)
		if symbol == "" {
			continue
		}
		// Rank and symbol, and nothing else. domain.RankedStock carries no
		// trading value, volume or price, so this source can raise a candidate
		// and can never contribute to its rate or acceleration. Leaving those
		// fields empty is the whole point: empty means the source did not report
		// it, and section 3 computes rates per source precisely so that a silent
		// zero from here cannot dilute a real figure from the official ranking.
		//
		// The two facts about the reading itself are not in that category. They are
		// things this adapter knows and nothing downstream can recover: how many rows
		// it asked for, and what the list looked like last time.
		rows = append(rows, candidate.Row{
			Symbol: symbol, Rank: s.Rank, RankTotal: total,
			RankRequested: w.size,
			NewlyListed:   newlyListed(previous, hadPrevious, symbol),
		})
	}
	// WTS sends no rate headers, so the budget is unreported — not zero.
	return candidate.Reading{Rows: rows}, nil
}

// wtsSymbol is the identity this source reports a row under.
//
// It is a function rather than two lines inside the loop because the previous
// reading's set has to be keyed by the same string the rows are keyed by. A set
// built from `s.Symbol` while the rows fall back to `s.ProductCode` would report
// every fallback row as a new entrant on every single reading.
func wtsSymbol(s domain.RankedStock) string {
	if symbol := strings.TrimSpace(s.Symbol); symbol != "" {
		return symbol
	}
	return strings.TrimSpace(s.ProductCode)
}

// rememberRead is officialRanking.rememberRead for this source's row type, with
// the same two conditions and for the same reasons — including the one about which
// count the wholeness comparison is made against. A stock with neither a Symbol nor
// a ProductCode is dropped from `current` and from the rows, and counting it
// towards the requested size would let a degraded reading become the yardstick for
// the next whole one.
func (w *wtsPopular) rememberRead(market string, stocks []domain.RankedStock, requested int) (
	map[string]bool, bool) {

	key := strings.ToUpper(strings.TrimSpace(market))
	current := make(map[string]bool, len(stocks))
	for _, s := range stocks {
		if symbol := wtsSymbol(s); symbol != "" {
			current[symbol] = true
		}
	}
	whole := requested == len(stocks) && len(current) == len(stocks)
	at := w.now.Now()
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.seen == nil {
		w.seen = map[string]previousReading{}
	}
	previous := w.seen[key]
	if whole {
		w.seen[key] = previousReading{symbols: current, at: at}
	}
	if !previous.usableAt(at) {
		return nil, false
	}
	return previous.symbols, true
}

// Panel builds the source list for one market.
//
// Membership is per market rather than per installation. A source that cannot see
// a market must be absent from its panel, because a scan treats "this source
// answered and did not list the symbol" as evidence the symbol left — and a
// source that structurally cannot see the market provides no such evidence. A
// present-but-always-empty source would therefore cool every candidate in that
// market on every scan, and the cooling clock would then expire them.
func Panel(market string, official RankingReader, budget BudgetReader, wts PopularityReader,
	clk clock.Clock) []candidate.Source {

	market = strings.ToUpper(strings.TrimSpace(market))
	var sources []candidate.Source

	if official != nil {
		// Trading value first: it is the ranking that answers "where is money
		// going now" rather than "what has already moved", which is the
		// difference between early discovery and a gainers list.
		//
		// The errors are discarded here because every type in this literal is a
		// compile-time constant of this package with an entry in rankingSourceID —
		// a failure would be a defect in this file rather than a condition to handle.
		//
		// Discarding it is only defensible while something fails when it happens, and
		// until this change nothing did. This comment used to name
		// TestEveryPanelSourceHasItsOwnID, and that test does not catch it: it walks
		// the panel it is handed and checks it for duplicate ids and emptiness, while a
		// type with no rankingSourceID entry yields a panel that is one element shorter,
		// still unique and still not empty. So the suite stayed green for a panel that
		// read two lists while its author believed it read three — the failure shape
		// this change exists to remove, sitting under the sentence that claimed to
		// prevent it. The guard that does catch it is
		// TestEveryRankingTypeThePanelNamesBecomesASourceItBuilds in
		// snapshot_drift_test.go, which compares the types this literal names against
		// the sources Panel returns when it is called.
		//
		// # RankingTopGainers is deliberately absent, and its constant is not
		//
		// It was in this literal and it never answered. Read sends `realtime`, and
		// the API refuses that duration for TOP_GAINERS and TOP_LOSERS with a 400
		// `unsupported-ranking-duration` — a fact docs/migration/openapi.latest.json
		// carried for forty-six days before the wiring arrived, and
		// snapshot_drift_test.go is what now reads it. Nothing failed while this ran:
		// the source was filed as a degradation on every single scan, which turned
		// the degradation flag into a light that is always on.
		//
		// The repair is not a different duration. `realtime` is the duration that
		// answers the question this package asks (see Read); a ranking type that
		// cannot be asked in that window is a ranking type that does not answer it,
		// so the source comes off the panel and the constants stay for the day it is
		// reconsidered (design D2).
		//
		// Putting it back needs all four of these, and the fourth is the hard one:
		//
		//	a measured, approved threshold for the late-entry veto — it is what
		//	would filter a list of moves that have already happened;
		//	shadow bands that resolve the range this list lives in, so the filter
		//	can discriminate inside it;
		//	an intraday observation of whether the 1d ranking updates during the
		//	session or is a daily batch;
		//	a design for ranks measured over different windows sharing one
		//	percentile, because a 1d position and a realtime position are not the
		//	same measurement and the row carries no window.
		for _, typ := range []string{RankingTradingAmount, RankingTradingVolume} {
			if src, err := OfficialRanking(official, budget, typ, 100, clk); err == nil {
				sources = append(sources, src)
			}
		}
	}
	// The WTS popularity ranking is a Korean-market product. Handing it a US
	// market would produce Korean rows filed under US — worse than absent.
	if wts != nil && market == candidate.MarketKR {
		sources = append(sources, WTSPopular(wts, 30, clk))
	}
	return sources
}

// decimal formats a float the way this package's Row wants it: a plain decimal
// string, with a genuine zero rendered as "0" rather than dropped.
//
// NaN and infinity come back empty, which means absent. They are reachable:
// internal/official's parseDecimal returns whatever strconv.ParseFloat produced
// whenever err is nil, and ParseFloat accepts "NaN", "Inf" and "Infinity" without
// complaint. FormatFloat would render those as "NaN" and "+Inf", which parse
// straight back — and an infinite trading value clears every threshold it is
// compared against, which is the same manufactured-maximum-signal failure the
// observation validator already refuses for ranks.
//
// It does not otherwise reconstruct absence. By the time a domain.RankingItem
// exists the API's null has already become 0.0, and inventing an empty string for
// that would be a guess dressed as a measurement.
func decimal(v float64) string {
	if math.IsNaN(v) || math.IsInf(v, 0) {
		return ""
	}
	return strconv.FormatFloat(v, 'f', -1, 64)
}
