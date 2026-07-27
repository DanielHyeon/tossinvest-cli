package measure

// population.go is the non-synthetic half of design D4 (task 5.5).
//
// # Why the grid is not enough, and what only this can give
//
// The synthetic grid shows where a threshold's boundary falls. It cannot show how
// often real entries land near it, because its density is whatever its author
// chose. That distinction matters most for one number: the minimum stop-width
// constant `k` a later change may want. Deriving `k` from stop widths somebody
// picked for a grid is circular, so the only honest source is a set of entries
// that actually happened.
//
// TossOS has none — no production Guardian entry verdict exists yet. The
// predecessor system does. So this reads real entry geometries from outside,
// pushes them through the same chain, and reports what each candidate would have
// refused.
//
// # Rules on the external data, and why each one
//
//	path is an argument   Never a default, never discovered. Analysis must not find
//	                      a live database by accident.
//	read only             The rows are somebody's trading history; this measures
//	                      them and writes nothing.
//	never committed       The file does not enter the repository. A local database
//	                      of real trades in git is a data leak with a long tail.
//	absence is not failure The measurement stands on the grid alone (SHALL — 외부
//	                      데이터 부재가 측정 부재가 되어서는 안 된다). What absence
//	                      changes is that `k` stays open, and the output says so.
//
// # The fallback, and its honesty requirement
//
// When no database is reachable, a small table transcribed from a post-mortem
// document can stand in. It is ≤ 8 rows and it is not a sample of anything — the
// source and the row count are printed beside every number derived from it, so
// nobody reads eight hand-copied trades as a population.

import (
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/costs"
	"github.com/JungHoonGhae/tossinvest-cli/internal/risk"
)

// RealEntry is one entry that actually happened, in the predecessor's records.
type RealEntry struct {
	// Label identifies the trade in its source. No account or personal data: this
	// package measures geometry, and the geometry is all it takes.
	Label               string
	Market              costs.Market
	Entry, Stop, Target string
}

// PopulationSource describes where the entries came from, for the output header.
type PopulationSource struct {
	// Kind is "database" or "document".
	Kind string
	// Description names it without exposing a path — an output artifact must not
	// carry a filesystem location from someone's machine.
	Description string
	// SampleSize is the row count. Printed beside every derived number so that a
	// ratio over eight rows cannot be read as a rate.
	SampleSize int
}

// PopulationReport is what the real entries showed.
type PopulationReport struct {
	Source     PopulationSource
	Entries    []RealEntryResult
	Thresholds []ThresholdOutcome
	// StopWidths are the observed widths as fractions of entry, which is the one
	// thing the grid genuinely cannot supply.
	StopWidths []string
}

// RealEntryResult is one real entry as the chain sees it today.
type RealEntryResult struct {
	RealEntry
	StopWidth               string
	BreakEven, Gross, Net   string
	Allowed                 bool
	StoppedStep, ReasonCode string
}

// StockOS058Entries is the documented fallback: the eight entries StockOS's 058
// post-mortem tabulates, all of them losses.
//
// It is a transcription, not a sample. Eight trades from one strategy on one day
// support no rate and no distribution — they are here so that the harness has a
// non-synthetic geometry to run at all when no database is reachable, and every
// number derived from them is printed with the row count attached.
//
// The geometry is 058's: entry at the open, a fixed 0.70% stop, target at 1.5×
// the risk. Prices are normalised to 100 because the post-mortem tabulates the
// ratio and not the instrument, and a fabricated instrument price would be a
// number nobody measured.
func StockOS058Entries() ([]RealEntry, PopulationSource) {
	entries := make([]RealEntry, 0, 8)
	for i := 1; i <= 8; i++ {
		entries = append(entries, RealEntry{
			Label:  fmt.Sprintf("058-%d", i),
			Market: costs.MarketKR,
			Entry:  "100",
			Stop:   "99.30",
			Target: "101.05",
		})
	}
	return entries, PopulationSource{
		Kind: "document",
		Description: "StockOS 058 post-mortem, transcribed. Eight entries, zero wins, " +
			"one strategy, one session — a transcription, not a sample. Geometry normalised " +
			"to entry 100 because the document tabulates ratios rather than instruments.",
		SampleSize: 8,
	}
}

// Population runs real entries through the chain.
func Population(
	model costs.Model, entries []RealEntry, source PopulationSource, now time.Time,
) (PopulationReport, error) {
	report := PopulationReport{Source: source}
	points := make([]GridPoint, 0, len(entries))

	for _, e := range entries {
		policy, _, err := gridPolicy(e.Market)
		if err != nil {
			return PopulationReport{}, err
		}
		ratios := risk.MeasureEntry(model, e.Market, e.Entry, e.Stop, e.Target)
		width, err := stopWidthOf(e.Entry, e.Stop)
		if err != nil {
			return PopulationReport{}, fmt.Errorf("measure: entry %s: %w", e.Label, err)
		}

		point := GridPoint{
			Market: e.Market, Entry: e.Entry, Stop: e.Stop, Target: e.Target,
			StopWidth: width,
			BreakEven: ratios.BreakEvenPrice, Gross: ratios.GrossRewardRisk, Net: ratios.NetRewardRisk,
		}
		verdict := risk.Evaluate(risk.Input{
			Now:     now,
			Intent:  gridIntent(e.Market, point),
			Account: unconstrainedAccount(e.Market),
			Policy:  policy,
			Costs:   model,
		})
		point.Allowed = verdict.Allowed
		point.StoppedStep = verdict.Step
		point.ReasonCode = string(verdict.Reason)
		points = append(points, point)

		report.Entries = append(report.Entries, RealEntryResult{
			RealEntry: e, StopWidth: width,
			BreakEven: ratios.BreakEvenPrice, Gross: ratios.GrossRewardRisk, Net: ratios.NetRewardRisk,
			Allowed: verdict.Allowed, StoppedStep: verdict.Step, ReasonCode: string(verdict.Reason),
		})
		report.StopWidths = append(report.StopWidths, width)
	}
	report.Thresholds = thresholdOutcomes(points)
	return report, nil
}

// stopWidthOf returns (entry − stop) / entry, which is the axis the grid cannot
// supply honestly.
func stopWidthOf(entry, stop string) (string, error) {
	e, ok := new(big.Rat).SetString(entry)
	if !ok {
		return "", fmt.Errorf("entry price %q is not a decimal", entry)
	}
	s, ok := new(big.Rat).SetString(stop)
	if !ok {
		return "", fmt.Errorf("stop price %q is not a decimal", stop)
	}
	if e.Sign() <= 0 {
		return "", fmt.Errorf("entry price %q is not positive", entry)
	}
	width := new(big.Rat).Sub(e, s)
	return risk.RatioText(new(big.Rat).Quo(width, e)), nil
}

func (p PopulationReport) render() string {
	var b strings.Builder
	b.WriteString("## Real-trade population\n\n")
	fmt.Fprintf(&b, "- Source: %s (`%s`)\n", p.Source.Description, p.Source.Kind)
	fmt.Fprintf(&b, "- Rows: **%d**. Every number below is over %d rows and is not a rate.\n\n",
		p.Source.SampleSize, p.Source.SampleSize)

	b.WriteString("### Observed stop widths\n\n")
	b.WriteString("This is the one axis the synthetic grid cannot supply — the grid's widths " +
		"are chosen, these were traded.\n\n")
	fmt.Fprintf(&b, "- widths: `%s`\n\n", strings.Join(p.StopWidths, "`, `"))

	b.WriteString("### Entries through today's chain\n\n")
	b.WriteString("| label | market | entry | stop | target | width | gross | net | verdict |\n")
	b.WriteString("|---|---|---|---|---|---|---|---|---|\n")
	for _, e := range p.Entries {
		verdict := "ALLOW"
		if !e.Allowed {
			verdict = e.StoppedStep + "/" + e.ReasonCode
		}
		fmt.Fprintf(&b, "| %s | %s | %s | %s | %s | %s | %s | %s | %s |\n",
			e.Label, e.Market, e.Entry, e.Stop, e.Target, e.StopWidth,
			dash(e.Gross), dash(e.Net), verdict)
	}

	b.WriteString("\n### Candidate net thresholds against these entries\n\n")
	b.WriteString("| candidate | would refuse | would keep | largest refused | smallest kept |\n")
	b.WriteString("|---|---|---|---|---|\n")
	for _, o := range p.Thresholds {
		fmt.Fprintf(&b, "| %s | %d | %d | %s | %s |\n",
			o.Threshold, o.BelowThreshold, o.AtOrAbove, dash(o.MaxRefused), dash(o.MinKept))
	}
	fmt.Fprintf(&b, "\n%s\n", StopWidthCircularityNote)
	return b.String()
}
