package reconcile

// compare.go turns a snapshot and the engine's own record into a list of
// disagreements (harden-execution-base task 3.4).
//
// # The comparison key
//
// (account, symbol, lineage-resolved current order id). The lineage step is not
// optional: the official API answers a modify with a *new* order number, so the
// id the engine recorded when it placed an order is not the id the broker will
// show once that order has been amended. Comparing raw ids would report the
// original as missing and the replacement as somebody else's — one order turned
// into two discrepancies, both wrong.
//
// # Two tolerances, and only one of them can stop trading
//
//	quantity      no business tolerance at all. A share is a share.
//	average price a documented epsilon, and it never blocks entries.
//
// The asymmetry is deliberate. A quantity disagreement means the engine does not
// know its own exposure, which is the condition new entries must not be added on
// top of. An average-price disagreement means two systems rounded a weighted mean
// differently; it is worth reporting and worthless as a trading signal, and
// blocking on it would be an outage caused by arithmetic.
//
// The 1e-9 relative bound on quantities is *not* a tolerance in that sense. It is
// the width of a float64 round trip: quantities arrive as decimal strings and are
// summed as float64, so 0.1 + 0.2 comes back as 0.30000000000000004. Rejecting
// that as a discrepancy would block trading on an artefact of binary arithmetic.
// The same bound is used by internal/brokerstate and internal/execgw.

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"

	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/riskcalc"
)

// quantityRoundTripEpsilon is the float64 round-trip width, not a business
// tolerance. See the file comment.
const quantityRoundTripEpsilon = 1e-9

// DefaultPriceEpsilon is the documented average-price tolerance: one part per
// million, relative. A KRW 70,000 average may legitimately differ by 0.07 between
// the broker's rounding and ours; anything larger is worth an operator's
// attention and still never blocks an entry.
const DefaultPriceEpsilon = 1e-6

// LocalOrder is an order the engine believes is live at the broker.
type LocalOrder struct {
	// OrderID is the *current* order number, after lineage resolution.
	OrderID string
	// OriginalOrderID is what the engine originally recorded, when an amendment
	// moved it.
	OriginalOrderID string
	Symbol          string
	Market          string
}

// LocalState is what the engine believes.
type LocalState struct {
	AccountRef string
	// Positions is the held quantity per symbol as a decimal string, taken from
	// the position projection: the sum of the account's non-CLOSED instances of
	// that symbol.
	Positions map[string]string
	// OpenOrders is keyed by the lineage-resolved current order id.
	OpenOrders map[string]LocalOrder
}

// PositionProjection is the projection half of the journal, as this package
// needs it.
//
// It is an interface so the coupling is one method wide and so a caller can be
// tested against a fixed projection. *journal.Journal satisfies it.
type PositionProjection interface {
	Positions(ctx context.Context, accountRef string) ([]journal.Position, error)
}

// HeldBySymbol reduces the position projection to one quantity per symbol.
//
// Non-CLOSED instances only, summed at symbol level — the two halves of the
// delta's reduction rule (SHALL — 비교는 심볼 수준에서 수행하고 투영은 비-CLOSED
// 인스턴스의 합으로 축약한다). CLOSED is final, so a closed instance holds
// nothing and counting it would report exposure on every symbol the engine has
// ever traded; the symbol-level sum is what the holdings snapshot can be
// compared against at all, because whether it carries a market dimension is
// `[미측정]` and until it does two instances of one symbol are one unit.
//
// The arithmetic is riskcalc's exact decimal, not a float sum. A projection is
// the thing a quantity mismatch is judged against with tolerance zero, and a
// binary round trip in the middle of that judgement is a disagreement nobody can
// explain.
func HeldBySymbol(ctx context.Context, p PositionProjection, accountRef string) (map[string]string, error) {
	instances, err := p.Positions(ctx, accountRef)
	if err != nil {
		return nil, err
	}
	held := make(map[string]string, len(instances))
	for _, instance := range instances {
		if instance.State == journal.PositionClosed {
			continue
		}
		symbol := strings.ToUpper(strings.TrimSpace(instance.Symbol))
		running, ok := held[symbol]
		if !ok {
			running = "0"
		}
		quantity := strings.TrimSpace(instance.Quantity)
		if quantity == "" {
			quantity = "0"
		}
		sum, err := riskcalc.AddDecimal(running, quantity)
		if err != nil {
			// A quantity the projection cannot add is not a quantity to guess at:
			// the comparison it feeds decides whether the engine may open a
			// position, and a symbol silently dropped from it reads as flat.
			return nil, fmt.Errorf(
				"reconcile: summing the projected quantity of %s (instance %s holds %q): %w",
				symbol, instance.ID, instance.Quantity, err)
		}
		held[symbol] = sum
	}
	return held, nil
}

// LocalStateFromJournal reads the engine's belief out of the journal.
//
// The position half is the projection (position-ledger: reconciliation의 로컬
// 상태는 이 투영을 소비한다 SHALL). It used to be a second sum over the fill
// ledger, and two derivations of one quantity is exactly what the spec forbids
// (SHALL NOT — fills-only 파생과 별도의 두 번째 포지션 계산을 두지 않는다): an
// adjustment converges the projection and leaves the fills alone, so the
// fills-only sum would keep reporting a disagreement the account already settled
// and the engine would stay blocked on a difference that no longer exists.
//
// The open-order half is unchanged: lineage-resolved current order ids, because
// the official API answers a modify with a new order number.
func LocalStateFromJournal(ctx context.Context, j *journal.Journal, accountRef string) (LocalState, error) {
	if j == nil {
		return LocalState{}, fmt.Errorf("reconcile: a journal is required to read local state")
	}
	positions, err := HeldBySymbol(ctx, j, accountRef)
	if err != nil {
		return LocalState{}, err
	}
	tracked, err := j.TrackedFillOrders(ctx, accountRef)
	if err != nil {
		return LocalState{}, err
	}

	state := LocalState{
		AccountRef: accountRef,
		Positions:  positions,
		OpenOrders: make(map[string]LocalOrder, len(tracked)),
	}
	for _, t := range tracked {
		current, err := j.ResolveCurrentOrderIDScoped(ctx, t.OrderID, journal.OrderLineageScope{
			AccountRef: t.AccountRef,
			Market:     t.Market,
			TradingDay: t.TradingDay,
			Symbol:     t.Symbol,
			Side:       t.Side,
		})
		if err != nil {
			// A broken lineage chain is not something to reconcile around: the
			// record of what replaced what is itself wrong.
			return LocalState{}, err
		}
		state.OpenOrders[current] = LocalOrder{
			OrderID:         current,
			OriginalOrderID: t.OrderID,
			Symbol:          strings.ToUpper(strings.TrimSpace(t.Symbol)),
			Market:          strings.ToLower(strings.TrimSpace(t.Market)),
		}
	}
	return state, nil
}

// Provenance says where a broker record came from, as far as the engine can tell.
type Provenance string

const (
	// ProvenanceEngine — the engine has a local intent for it.
	ProvenanceEngine Provenance = "engine"
	// ProvenanceExternal — the broker has it and the engine never asked for it:
	// a manual order in the app, another tool, a corporate action.
	ProvenanceExternal Provenance = "external"
)

// QuantityMismatch is a disagreement that blocks new entries.
type QuantityMismatch struct {
	Symbol string
	// Local is what the engine believed, Broker is what the account says.
	Local  string
	Broker string
}

// Authority returns the value that wins. It is always the broker's: the account
// is the authority, and this method exists so that fact is expressed in code
// rather than in a comment somebody can forget to read.
func (m QuantityMismatch) Authority() string { return m.Broker }

// PriceDeviation is an average-price disagreement. Reported, never blocking.
type PriceDeviation struct {
	Symbol    string
	Local     string
	Broker    string
	Relative  float64
	Epsilon   float64
	Blocking  bool // always false; present so the JSON record says so explicitly
	Explained string
}

// ExternalOrder is a broker order with no local intent.
type ExternalOrder struct {
	BrokerOrder
	Provenance Provenance
}

// ExternalPosition is a holding with no local record.
type ExternalPosition struct {
	Holding
	Provenance Provenance
}

// Diff is the outcome of one comparison.
type Diff struct {
	AsOf        string
	AccountRef  string
	Matched     int
	Quantities  []QuantityMismatch
	Prices      []PriceDeviation
	ExternalOrd []ExternalOrder
	ExternalPos []ExternalPosition
	// MissingOrders are orders the engine believes are live that the account
	// does not show. They are treated exactly like a quantity mismatch: the
	// engine's picture of its own exposure is wrong.
	MissingOrders []LocalOrder
}

// BlocksEntry reports whether this diff must stop new exposure.
//
// Price deviations are excluded by construction, per the spec's "평균단가는 …
// 진입 차단 판정에서 제외". External records are excluded too: somebody trading
// their own account by hand is not a malfunction, and blocking on it would make
// the engine unusable next to its owner. Both are still reported and alerted on.
func (d Diff) BlocksEntry() bool {
	return len(d.Quantities) > 0 || len(d.MissingOrders) > 0
}

// Clean reports a diff with nothing in it at all.
func (d Diff) Clean() bool {
	return !d.BlocksEntry() && len(d.Prices) == 0 &&
		len(d.ExternalOrd) == 0 && len(d.ExternalPos) == 0
}

// Summary is a one-line description for logs and alerts.
func (d Diff) Summary() string {
	if d.Clean() {
		return fmt.Sprintf("account and engine agree on %d position(s)", d.Matched)
	}
	parts := make([]string, 0, 4)
	if n := len(d.Quantities); n > 0 {
		parts = append(parts, fmt.Sprintf("%d quantity mismatch(es)", n))
	}
	if n := len(d.MissingOrders); n > 0 {
		parts = append(parts, fmt.Sprintf("%d order(s) the account does not show", n))
	}
	if n := len(d.ExternalOrd); n > 0 {
		parts = append(parts, fmt.Sprintf("%d external order(s)", n))
	}
	if n := len(d.ExternalPos); n > 0 {
		parts = append(parts, fmt.Sprintf("%d external position(s)", n))
	}
	if n := len(d.Prices); n > 0 {
		parts = append(parts, fmt.Sprintf("%d average-price deviation(s) (not blocking)", n))
	}
	return strings.Join(parts, ", ")
}

// Comparer holds the tolerances.
type Comparer struct {
	// PriceEpsilon is the relative average-price tolerance. Zero takes
	// DefaultPriceEpsilon.
	PriceEpsilon float64
}

// Compare judges a snapshot against the engine's belief.
//
// Quantities and order identity only. Average prices are a separate call
// (ComparePositionPrices) because they are a separate decision: they can never
// block an entry, and keeping them out of this function means no future edit can
// accidentally make them do so.
func (c Comparer) Compare(snap Snapshot, local LocalState) Diff {
	diff := Diff{AccountRef: snap.AccountRef}
	if !snap.AsOf.IsZero() {
		diff.AsOf = snap.AsOf.UTC().Format("2006-01-02T15:04:05Z07:00")
	}

	// --- positions ----------------------------------------------------------
	brokerBySymbol := make(map[string]Holding, len(snap.Holdings))
	for _, h := range snap.Holdings {
		brokerBySymbol[strings.ToUpper(strings.TrimSpace(h.Symbol))] = h
	}

	symbols := make([]string, 0, len(brokerBySymbol)+len(local.Positions))
	seen := map[string]bool{}
	for symbol := range brokerBySymbol {
		if !seen[symbol] {
			seen[symbol] = true
			symbols = append(symbols, symbol)
		}
	}
	for symbol := range local.Positions {
		key := strings.ToUpper(strings.TrimSpace(symbol))
		if !seen[key] {
			seen[key] = true
			symbols = append(symbols, key)
		}
	}
	sort.Strings(symbols)

	for _, symbol := range symbols {
		holding, heldAtBroker := brokerBySymbol[symbol]
		localQty, knownLocally := local.Positions[symbol]

		brokerQty := "0"
		if heldAtBroker {
			brokerQty = canonicalDecimal(holding.Quantity)
		}
		if !knownLocally {
			localQty = "0"
		}
		localQty = canonicalDecimal(localQty)

		bothZero := isZeroDecimal(localQty) && isZeroDecimal(brokerQty)
		switch {
		case bothZero:
			// Nothing on either side. Not a match worth counting, not a problem.
		case !knownLocally || isZeroDecimal(localQty):
			// The account holds something the engine never bought. That is a
			// fact about the owner, not a malfunction.
			diff.ExternalPos = append(diff.ExternalPos, ExternalPosition{
				Holding: holding, Provenance: ProvenanceExternal,
			})
		case !quantitiesAgree(localQty, brokerQty):
			diff.Quantities = append(diff.Quantities, QuantityMismatch{
				Symbol: symbol, Local: localQty, Broker: brokerQty,
			})
		default:
			diff.Matched++
		}

	}

	// --- open orders --------------------------------------------------------
	brokerOrders := make(map[string]BrokerOrder, len(snap.OpenOrders))
	for _, o := range snap.OpenOrders {
		brokerOrders[strings.TrimSpace(o.OrderID)] = o
		if _, known := local.OpenOrders[strings.TrimSpace(o.OrderID)]; !known {
			diff.ExternalOrd = append(diff.ExternalOrd, ExternalOrder{
				BrokerOrder: o, Provenance: ProvenanceExternal,
			})
		}
	}

	missing := make([]LocalOrder, 0)
	for id, order := range local.OpenOrders {
		if _, present := brokerOrders[id]; !present {
			missing = append(missing, order)
		}
	}
	sort.Slice(missing, func(i, j int) bool { return missing[i].OrderID < missing[j].OrderID })
	if len(missing) > 0 {
		diff.MissingOrders = missing
	}
	sort.Slice(diff.ExternalOrd, func(i, j int) bool {
		return diff.ExternalOrd[i].OrderID < diff.ExternalOrd[j].OrderID
	})
	return diff
}

// ComparePositionPrices compares an engine-side average price against the
// account's, applying the documented epsilon.
//
// It is exported separately from Compare because the engine's average price is
// derived from the fill ledger rather than carried on LocalState: a caller that
// has it can ask for the comparison, and a caller that does not simply gets no
// price deviations rather than a spurious one against zero.
func (c Comparer) ComparePositionPrices(snap Snapshot, localAverages map[string]string) []PriceDeviation {
	epsilon := c.PriceEpsilon
	if epsilon <= 0 {
		epsilon = DefaultPriceEpsilon
	}
	var out []PriceDeviation
	for _, h := range snap.Holdings {
		symbol := strings.ToUpper(strings.TrimSpace(h.Symbol))
		localPrice, ok := localAverages[symbol]
		if !ok {
			continue
		}
		if dev, differs := comparePrices(symbol, localPrice, h.AveragePrice, epsilon); differs {
			out = append(out, dev)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Symbol < out[j].Symbol })
	return out
}

// comparePrices reports a deviation when the decimal strings differ by more than
// the relative epsilon. Equal strings short-circuit: the spec asks for a decimal
// string comparison first, and two identical strings cannot differ numerically.
func comparePrices(symbol, local, broker string, epsilon float64) (PriceDeviation, bool) {
	localCanon := canonicalDecimal(local)
	brokerCanon := canonicalDecimal(broker)
	if localCanon == brokerCanon {
		return PriceDeviation{}, false
	}
	lv, lerr := strconv.ParseFloat(localCanon, 64)
	bv, berr := strconv.ParseFloat(brokerCanon, 64)
	if lerr != nil || berr != nil {
		return PriceDeviation{
			Symbol: symbol, Local: localCanon, Broker: brokerCanon,
			Epsilon:   epsilon,
			Explained: "one of the average prices is not a decimal",
		}, true
	}
	scale := math.Max(1, math.Max(math.Abs(lv), math.Abs(bv)))
	relative := math.Abs(lv-bv) / scale
	if relative <= epsilon {
		return PriceDeviation{}, false
	}
	return PriceDeviation{
		Symbol: symbol, Local: localCanon, Broker: brokerCanon,
		Relative: relative, Epsilon: epsilon,
		Explained: "average prices differ by more than the documented epsilon; " +
			"reported, and excluded from the entry-block decision",
	}, true
}

// quantitiesAgree applies the "허용 오차 0" rule: equal decimal strings, or a
// difference no wider than a float64 round trip.
func quantitiesAgree(a, b string) bool {
	if a == b {
		return true
	}
	av, aerr := strconv.ParseFloat(a, 64)
	bv, berr := strconv.ParseFloat(b, 64)
	if aerr != nil || berr != nil {
		return false
	}
	scale := math.Max(1, math.Max(math.Abs(av), math.Abs(bv)))
	return math.Abs(av-bv) <= quantityRoundTripEpsilon*scale
}

func isZeroDecimal(s string) bool {
	v, err := strconv.ParseFloat(strings.TrimSpace(s), 64)
	if err != nil {
		return false
	}
	return math.Abs(v) <= quantityRoundTripEpsilon
}

// parseBrokerOrder reads one raw broker order into the snapshot's shape.
func parseBrokerOrder(raw json.RawMessage) (BrokerOrder, error) {
	var payload officialOrder
	if err := json.Unmarshal(raw, &payload); err != nil {
		return BrokerOrder{}, fmt.Errorf("an open order could not be read: %v", err)
	}
	if strings.TrimSpace(payload.OrderID) == "" {
		return BrokerOrder{}, fmt.Errorf("an open order has no orderId")
	}
	order := BrokerOrder{
		OrderID: strings.TrimSpace(payload.OrderID),
		Symbol:  strings.ToUpper(strings.TrimSpace(payload.Symbol)),
		Side:    strings.ToUpper(strings.TrimSpace(payload.Side)),
		Status:  strings.ToUpper(strings.TrimSpace(payload.Status)),
	}
	order.Quantity = derefDecimal(payload.Quantity)
	order.Price = derefDecimal(payload.Price)
	if payload.Execution != nil {
		order.FilledQuantity = derefDecimal(payload.Execution.FilledQuantity)
	}
	return order, nil
}

func derefDecimal(s *string) string {
	if s == nil {
		return "0"
	}
	return canonicalDecimal(*s)
}
