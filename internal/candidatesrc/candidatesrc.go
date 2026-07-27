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

	"github.com/JungHoonGhae/tossinvest-cli/internal/candidate"
	"github.com/JungHoonGhae/tossinvest-cli/internal/domain"
	"github.com/JungHoonGhae/tossinvest-cli/internal/official"
)

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

// Official ranking types this change reads. TOP_LOSERS is deliberately absent:
// discovery is looking for the start of a move up, and a losers list would raise
// candidates that the veto would then have to reject one at a time.
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
}

// OfficialRanking returns a Source over one official ranking type.
//
// count is capped at the API's documented maximum of 100 rather than passed
// through: a larger value is silently truncated by the server, and a caller that
// asked for 150 and got 100 would compute rank percentiles against a list length
// that never existed.
//
// An unrecognised ranking type is refused rather than given a fallback id, since
// a fallback would put it back into a collision with a real source.
func OfficialRanking(reader RankingReader, budget BudgetReader, typ string, count int) (candidate.Source, error) {
	id, ok := rankingSourceID[typ]
	if !ok {
		return nil, fmt.Errorf("candidatesrc: unknown ranking type %q", typ)
	}
	if count <= 0 || count > 100 {
		count = 100
	}
	return &officialRanking{reader: reader, budget: budget, typ: typ, id: id, count: count}, nil
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
}

// WTSPopular returns a Source over the WTS popularity ranking.
//
// It is an additive source. Everything downstream must keep working when it is
// gone, because a WTS session expiring is an ordinary event rather than an
// incident — which is why the official ranking, not this, is the one required to
// be sufficient alone.
func WTSPopular(reader PopularityReader, size int) candidate.Source {
	if size <= 0 {
		size = 30
	}
	return &wtsPopular{reader: reader, size: size}
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
	rows := make([]candidate.Row, 0, total)
	for _, s := range raw.Stocks {
		symbol := strings.TrimSpace(s.Symbol)
		if symbol == "" {
			symbol = strings.TrimSpace(s.ProductCode)
		}
		if symbol == "" {
			continue
		}
		// Rank and symbol, and nothing else. domain.RankedStock carries no
		// trading value, volume or price, so this source can raise a candidate
		// and can never contribute to its rate or acceleration. Leaving those
		// fields empty is the whole point: empty means the source did not report
		// it, and section 3 computes rates per source precisely so that a silent
		// zero from here cannot dilute a real figure from the official ranking.
		rows = append(rows, candidate.Row{Symbol: symbol, Rank: s.Rank, RankTotal: total})
	}
	// WTS sends no rate headers, so the budget is unreported — not zero.
	return candidate.Reading{Rows: rows}, nil
}

// Panel builds the source list for one market.
//
// Membership is per market rather than per installation. A source that cannot see
// a market must be absent from its panel, because a scan treats "this source
// answered and did not list the symbol" as evidence the symbol left — and a
// source that structurally cannot see the market provides no such evidence. A
// present-but-always-empty source would therefore cool every candidate in that
// market on every scan, and the cooling clock would then expire them.
func Panel(market string, official RankingReader, budget BudgetReader, wts PopularityReader) []candidate.Source {
	market = strings.ToUpper(strings.TrimSpace(market))
	var sources []candidate.Source

	if official != nil {
		// Trading value first: it is the ranking that answers "where is money
		// going now" rather than "what has already moved", which is the
		// difference between early discovery and a gainers list.
		//
		// The errors are discarded here because the three types are compile-time
		// constants of this package with entries in rankingSourceID — a failure
		// would be a defect in this file, and TestEveryPanelSourceHasItsOwnID
		// fails if one ever slips.
		for _, typ := range []string{RankingTradingAmount, RankingTradingVolume, RankingTopGainers} {
			if src, err := OfficialRanking(official, budget, typ, 100); err == nil {
				sources = append(sources, src)
			}
		}
	}
	// The WTS popularity ranking is a Korean-market product. Handing it a US
	// market would produce Korean rows filed under US — worse than absent.
	if wts != nil && market == candidate.MarketKR {
		sources = append(sources, WTSPopular(wts, 30))
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
