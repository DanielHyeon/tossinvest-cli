// Package performance builds a derived, rebuildable lane-performance read model
// from journal lineage and price observations that another component already
// owns. It contains no broker, polling, configuration, lane-control, or LIVE
// approval capability.
package performance

import (
	"errors"
	"fmt"
	"math/big"
	"sort"
	"strings"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/markout"
)

const SemanticsVersion = "lane-performance/v1"

type Status string

const (
	StatusComplete           Status = "complete"
	StatusLinkMissing        Status = "link_missing"
	StatusNotMeasured        Status = "not_measured"
	StatusInsufficientSample Status = "insufficient_sample"
)

type Side string

const (
	SideBuy  Side = "BUY"
	SideSell Side = "SELL"
)

// Lineage is the exact identifier chain used for attribution. CloseID names the
// immutable close record; today the journal's trade_outcomes primary key is the
// position ID, so the two values may deliberately be byte-identical.
type Lineage struct {
	CandidateLifeID    string
	ThresholdVersion   string
	ThresholdSetDigest string
	EvidenceDigest     string
	LaneID             string
	LaneVersion        string
	DecisionID         string
	AttemptID          string
	OrderID            string
	FillID             string
	PositionID         string
	CloseID            string
	PolicyID           string
	PolicyVersion      string
}

func (l Lineage) Status() Status {
	for _, value := range []string{
		l.CandidateLifeID, l.ThresholdVersion, l.ThresholdSetDigest, l.EvidenceDigest,
		l.LaneID, l.LaneVersion, l.DecisionID, l.AttemptID, l.OrderID, l.FillID,
		l.PositionID, l.CloseID, l.PolicyID, l.PolicyVersion,
	} {
		if strings.TrimSpace(value) == "" {
			return StatusLinkMissing
		}
	}
	return StatusComplete
}

type Trade struct {
	ID      string
	Lineage Lineage
	Market  string
	Side    Side

	DecisionAt    time.Time
	DecisionPrice string
	EntryAt       time.Time
	EntryPrice    string
	Quantity      string
	CostTotal     string

	RealizedPnLAfterCosts string
	RealizedR             string
	ClosedAt              time.Time
}

type Observation struct {
	ID            string
	PositionID    string
	At            time.Time
	Price         string
	Source        string
	SourceVersion string
}

type Metric struct {
	Status        Status
	Value         string
	ObservationID string
	ObservedAt    time.Time
	Source        string
	SourceVersion string
}

type MarkoutMetric struct {
	Minutes         int
	Status          Status
	GrossPct        string
	CostAdjustedPct string
	ObservationID   string
	ObservedAt      time.Time
	Source          string
	SourceVersion   string
}

type Snapshot struct {
	CalculatedAt     time.Time
	SemanticsVersion string
	LineageStatus    Status
	Markouts         []MarkoutMetric
	Slippage         Metric
	MFE              Metric
	MAE              Metric
}

func (s Snapshot) Markout(minutes int) MarkoutMetric {
	for _, metric := range s.Markouts {
		if metric.Minutes == minutes {
			return metric
		}
	}
	return MarkoutMetric{Minutes: minutes, Status: StatusNotMeasured}
}

// Measure consumes only the observations supplied by the caller. It filters by
// the exact position ID before reusing markout.Measure's inclusive +60 second
// selection rule; it never performs or schedules a quote request.
func Measure(trade Trade, observations []Observation, calculatedAt time.Time) Snapshot {
	snapshot := Snapshot{
		CalculatedAt: calculatedAt.UTC(), SemanticsVersion: SemanticsVersion,
		LineageStatus: trade.Lineage.Status(),
		Slippage:      Metric{Status: StatusNotMeasured},
		MFE:           Metric{Status: StatusNotMeasured}, MAE: Metric{Status: StatusNotMeasured},
	}
	filtered := make([]Observation, 0, len(observations))
	for _, row := range observations {
		if row.PositionID == trade.Lineage.PositionID && !row.At.Before(trade.EntryAt) &&
			(trade.ClosedAt.IsZero() || !row.At.After(trade.ClosedAt)) {
			filtered = append(filtered, row)
		}
	}
	sort.SliceStable(filtered, func(i, j int) bool {
		if filtered[i].At.Equal(filtered[j].At) {
			return filtered[i].ID < filtered[j].ID
		}
		return filtered[i].At.Before(filtered[j].At)
	})
	markoutRows := make([]markout.Observation, 0, len(filtered))
	for _, row := range filtered {
		markoutRows = append(markoutRows, markout.Observation{At: row.At, Price: row.Price})
	}
	report := markout.Measure(trade.EntryAt, trade.EntryPrice, markoutRows)
	costPct, costOK := normalizedCostPct(trade.EntryPrice, trade.Quantity, trade.CostTotal)
	for _, measured := range report.Measurements {
		metric := MarkoutMetric{Minutes: measured.Minutes, Status: StatusNotMeasured}
		if measured.Status == markout.StatusMeasured {
			metric.Status = StatusComplete
			metric.GrossPct = measured.ReturnPct
			if costOK {
				metric.CostAdjustedPct = ratText(new(big.Rat).Sub(rat(measured.ReturnPct), costPct))
			}
			if measured.ObservedAt != nil {
				for _, row := range filtered {
					if row.At.Equal(*measured.ObservedAt) && strings.TrimSpace(row.Price) == measured.ObservedPrice {
						metric.ObservationID = row.ID
						metric.ObservedAt = row.At.UTC()
						metric.Source = row.Source
						metric.SourceVersion = row.SourceVersion
						break
					}
				}
			}
		}
		snapshot.Markouts = append(snapshot.Markouts, metric)
	}
	if slippage, ok := slippagePct(trade.Side, trade.DecisionPrice, trade.EntryPrice); ok {
		snapshot.Slippage = Metric{
			Status: StatusComplete, Value: ratText(slippage), ObservedAt: trade.EntryAt.UTC(),
			Source: "journal-decision-entry", SourceVersion: SemanticsVersion,
		}
	}
	if favorable, adverse, favorableRow, adverseRow, ok := extremaPct(trade.Side, trade.EntryPrice, filtered); ok {
		snapshot.MFE = excursionMetric(favorable, favorableRow, trade.EntryAt)
		snapshot.MAE = excursionMetric(adverse, adverseRow, trade.EntryAt)
	}
	return snapshot
}

func (t Trade) validate() error {
	switch {
	case strings.TrimSpace(t.ID) == "":
		return errors.New("performance: trade id is required")
	case strings.TrimSpace(t.Lineage.PositionID) == "":
		return errors.New("performance: position id is required even when other lineage is missing")
	case strings.TrimSpace(t.Market) == "":
		return errors.New("performance: market is required")
	case t.Side != SideBuy && t.Side != SideSell:
		return fmt.Errorf("performance: unsupported side %q", t.Side)
	case t.EntryAt.IsZero() || t.ClosedAt.IsZero() || t.ClosedAt.Before(t.EntryAt):
		return errors.New("performance: entry/close interval is invalid")
	}
	for name, value := range map[string]string{"pnl after costs": t.RealizedPnLAfterCosts, "realized r": t.RealizedR} {
		if _, ok := decimal(value); !ok {
			return fmt.Errorf("performance: %s %q is not a decimal", name, value)
		}
	}
	for name, value := range map[string]string{"entry price": t.EntryPrice, "quantity": t.Quantity} {
		if parsed, ok := decimal(value); !ok || parsed.Sign() <= 0 {
			return fmt.Errorf("performance: %s %q is not a positive decimal", name, value)
		}
	}
	if cost, ok := decimal(t.CostTotal); !ok || cost.Sign() < 0 {
		return fmt.Errorf("performance: cost total %q is not a non-negative decimal", t.CostTotal)
	}
	if strings.TrimSpace(t.DecisionPrice) != "" {
		if decision, ok := decimal(t.DecisionPrice); !ok || decision.Sign() <= 0 {
			return fmt.Errorf("performance: decision price %q is not a positive decimal", t.DecisionPrice)
		}
	}
	return nil
}

func (o Observation) validate() error {
	if strings.TrimSpace(o.ID) == "" || strings.TrimSpace(o.PositionID) == "" || o.At.IsZero() ||
		strings.TrimSpace(o.Source) == "" || strings.TrimSpace(o.SourceVersion) == "" {
		return errors.New("performance: observation identity, instant, source, and source version are required")
	}
	value, ok := decimal(o.Price)
	if !ok || value.Sign() <= 0 {
		return fmt.Errorf("performance: observation price %q is not positive decimal", o.Price)
	}
	return nil
}

func slippagePct(side Side, decisionPrice, entryPrice string) (*big.Rat, bool) {
	decision, ok := positiveDecimal(decisionPrice)
	if !ok {
		return nil, false
	}
	entry, ok := positiveDecimal(entryPrice)
	if !ok {
		return nil, false
	}
	delta := new(big.Rat)
	if side == SideBuy {
		delta.Sub(entry, decision)
	} else if side == SideSell {
		delta.Sub(decision, entry)
	} else {
		return nil, false
	}
	return delta.Quo(delta, decision).Mul(delta, big.NewRat(100, 1)), true
}

func extremaPct(side Side, entryPrice string, observations []Observation) (*big.Rat, *big.Rat, *Observation, *Observation, bool) {
	entry, ok := positiveDecimal(entryPrice)
	if !ok || len(observations) == 0 {
		return nil, nil, nil, nil, false
	}
	favorable, adverse := new(big.Rat), new(big.Rat)
	var favorableRow, adverseRow *Observation
	measured := false
	for _, row := range observations {
		price, valid := positiveDecimal(row.Price)
		if !valid {
			continue
		}
		measured = true
		movement := new(big.Rat).Sub(price, entry)
		if side == SideSell {
			movement.Neg(movement)
		} else if side != SideBuy {
			return nil, nil, nil, nil, false
		}
		movement.Quo(movement, entry).Mul(movement, big.NewRat(100, 1))
		if movement.Cmp(favorable) > 0 {
			favorable = new(big.Rat).Set(movement)
			copy := row
			favorableRow = &copy
		}
		if movement.Cmp(adverse) < 0 {
			adverse = new(big.Rat).Set(movement)
			copy := row
			adverseRow = &copy
		}
	}
	return favorable, adverse, favorableRow, adverseRow, measured
}

func excursionMetric(value *big.Rat, row *Observation, entryAt time.Time) Metric {
	metric := Metric{Status: StatusComplete, Value: ratText(value)}
	if row == nil {
		metric.ObservedAt = entryAt.UTC()
		metric.Source = "journal-entry"
		metric.SourceVersion = SemanticsVersion
		return metric
	}
	metric.ObservationID = row.ID
	metric.ObservedAt = row.At.UTC()
	metric.Source = row.Source
	metric.SourceVersion = row.SourceVersion
	return metric
}

func normalizedCostPct(entryPrice, quantity, cost string) (*big.Rat, bool) {
	entry, ok := positiveDecimal(entryPrice)
	if !ok {
		return nil, false
	}
	qty, ok := positiveDecimal(quantity)
	if !ok {
		return nil, false
	}
	costValue, ok := decimal(cost)
	if !ok || costValue.Sign() < 0 {
		return nil, false
	}
	base := new(big.Rat).Mul(entry, qty)
	costPct := new(big.Rat).Quo(costValue, base)
	return costPct.Mul(costPct, big.NewRat(100, 1)), true
}

func decimal(value string) (*big.Rat, bool) {
	text := strings.TrimSpace(value)
	if text == "" || strings.ContainsAny(text, "eE") {
		return nil, false
	}
	valueRat, ok := new(big.Rat).SetString(text)
	return valueRat, ok
}

func positiveDecimal(value string) (*big.Rat, bool) {
	parsed, ok := decimal(value)
	return parsed, ok && parsed.Sign() > 0
}

func rat(value string) *big.Rat {
	parsed, _ := decimal(value)
	if parsed == nil {
		return new(big.Rat)
	}
	return parsed
}

func ratText(value *big.Rat) string {
	if value == nil {
		return ""
	}
	if value.IsInt() {
		return value.Num().String()
	}
	text := strings.TrimRight(value.FloatString(12), "0")
	text = strings.TrimRight(text, ".")
	if text == "-0" || text == "" {
		return "0"
	}
	return text
}
