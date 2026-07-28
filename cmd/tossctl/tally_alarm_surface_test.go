package main

// tally_alarm_surface_test.go is task 4.3 at this surface: the scan output renders
// the identity alarm, and it does not decide anything about it.
//
// The rule lives in internal/candidate. This file asserts two things and no more:
// that a contradictory tally produces an alarm here at all, and that the ordinary
// no-threshold tally does not — because an alarm that fires every day is one an
// operator stops reading, and this system's every day has two vetoes with no
// approved threshold.

import (
	"encoding/json"
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/candidate"
	"github.com/JungHoonGhae/tossinvest-cli/internal/output"
)

// contradictoryTally counts every unmeasured candidate as a pass, which is the
// one-line edit in TallyVetoes' switch that D10 exists to stop.
func contradictoryTally() candidate.VetoTally {
	return candidate.VetoTally{
		Total: 12, Passed: 12,
		Raised: map[candidate.VetoCode]int{},
		NotMeasured: map[candidate.VetoCode]int{
			candidate.VetoNearHigh: 12,
		},
		Reasons: map[candidate.VetoUnmeasured]int{
			candidate.VetoUnmeasured(candidate.LevelNoDayHigh): 12,
		},
	}
}

// ordinaryTally is what this system produces on a normal day: no thresholds for two
// of the three codes, so nothing can pass.
func ordinaryTally() candidate.VetoTally {
	return candidate.VetoTally{
		Total:  12,
		Raised: map[candidate.VetoCode]int{},
		NotMeasured: map[candidate.VetoCode]int{
			candidate.VetoSeenLate: 12, candidate.VetoExtended: 12, candidate.VetoNearHigh: 11,
		},
		Reasons: map[candidate.VetoUnmeasured]int{
			candidate.VetoThresholdAbsent:                      24,
			candidate.VetoUnmeasured(candidate.LevelNoDayHigh): 11,
		},
	}
}

// TestTheScanOutputSaysSoWhenTheTallyContradictsItself.
func TestTheScanOutputSaysSoWhenTheTallyContradictsItself(t *testing.T) {
	res := candidate.CycleResult{
		Market: candidate.MarketKR, At: reportAt, Vetoes: contradictoryTally(),
	}

	table := renderReport(t, output.FormatTable, res)
	if !strings.Contains(table, "ALARM") {
		t.Errorf("a tally whose buckets sum past its own total printed no alarm; the counts "+
			"look ordinary on their own and the arithmetic is the only thing that "+
			"disagrees:\n%s", table)
	}
	// The numbers that produced the finding are on the line, so a reader has
	// something to check rather than a verdict to accept.
	for _, want := range []string{"near_high", "12"} {
		if !strings.Contains(table, want) {
			t.Errorf("the alarm line does not carry %q:\n%s", want, table)
		}
	}

	var report struct {
		Veto struct {
			Alarms []string `json:"alarms"`
		} `json:"veto"`
	}
	if err := json.Unmarshal([]byte(renderReport(t, output.FormatJSON, res)), &report); err != nil {
		t.Fatalf("the JSON report does not parse: %v", err)
	}
	if len(report.Veto.Alarms) == 0 {
		t.Error("the JSON carries no alarm, so a consumer reading the counts sees a tidy " +
			"twelve passes")
	}
}

// TestTheOrdinaryNoThresholdScanRaisesNoAlarm is the negative control, and it is the
// one that decides whether this alarm is worth having. Two of the three vetoes have
// no approved threshold, every candidate is unmeasured on both, and that is not a
// fault — an alarm that fired on it would be noise from the day it shipped.
func TestTheOrdinaryNoThresholdScanRaisesNoAlarm(t *testing.T) {
	table := renderReport(t, output.FormatTable, candidate.CycleResult{
		Market: candidate.MarketKR, At: reportAt, Vetoes: ordinaryTally(),
	})
	if strings.Contains(table, "ALARM") {
		t.Errorf("the state this system is in every day printed an alarm:\n%s", table)
	}
	if !strings.Contains(table, "structurally 0") {
		t.Errorf("and the sentence that explains the zero is gone; the alarm was added "+
			"beside it, not instead of it:\n%s", table)
	}
}

// TestNeitherSurfaceDecidesWhetherTheTallyIsConsistent is task 4.3.
//
// The judgement is candidate.VetoTally.Anomalies. A surface that recomputed the
// inequality would be the same rule implemented twice, and this repository has
// already paid for that arrangement once: the console assembled the discovery
// tallies by hand, the two agreed, and the cost sat in the future as a fourth shadow
// band that appeared on one screen and not the other.
//
// So the check is structural — no arithmetic on tally fields in either renderer.
func TestNeitherSurfaceDecidesWhetherTheTallyIsConsistent(t *testing.T) {
	for _, path := range []string{"candidate.go", "../../internal/console/signals.go"} {
		t.Run(path, func(t *testing.T) {
			file, err := parser.ParseFile(token.NewFileSet(), path, nil, parser.SkipObjectResolution)
			if err != nil {
				t.Fatalf("parsing %s: %v", path, err)
			}
			seen := 0
			ast.Inspect(file, func(n ast.Node) bool {
				sel, ok := n.(*ast.SelectorExpr)
				if !ok {
					return true
				}
				// A renderer reads the counts to print them. What it must not do is
				// compare them with each other, which is what the identity check is.
				switch sel.Sel.Name {
				case "NotMeasured", "Raised", "Passed", "Total":
					seen++
				}
				return true
			})
			if seen == 0 {
				t.Fatalf("%s names none of the tally's counts; this guard is reading "+
					"nothing", path)
			}
			ast.Inspect(file, func(n ast.Node) bool {
				bin, ok := n.(*ast.BinaryExpr)
				if !ok || (bin.Op != token.LEQ && bin.Op != token.GTR &&
					bin.Op != token.LSS && bin.Op != token.GEQ) {
					return true
				}
				if namesATallyCount(bin.X) && namesATallyCount(bin.Y) {
					t.Errorf("%s compares two tally counts with each other (%s); the identity "+
						"is candidate.VetoTally.Anomalies' job, and a rule implemented on two "+
						"screens eventually disagrees with itself", path, bin.Op)
				}
				return true
			})
		})
	}
}

// namesATallyCount reports whether an expression reads one of the tally's counters.
func namesATallyCount(e ast.Expr) bool {
	found := false
	ast.Inspect(e, func(n ast.Node) bool {
		if sel, ok := n.(*ast.SelectorExpr); ok {
			switch sel.Sel.Name {
			case "NotMeasured", "Raised", "Passed", "Total", "Vetoed", "Unmeasured":
				found = true
			}
		}
		return true
	})
	return found
}
