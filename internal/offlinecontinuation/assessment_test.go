package offlinecontinuation

import (
	"math"
	"math/big"
	"strings"
	"testing"

	evidence "github.com/JungHoonGhae/tossinvest-cli/internal/campaignstateevidence"
)

func TestCalculateBoundsAndNoBudget(t *testing.T) {
	base := Result{InspectionOnly: true, LedgerCurrentness: "unverified", RiskBoundSource: "caller_assumption"}
	snapshot := evidence.Evidence{Plan: evidence.Plan{Budget: "1000"}, Risk: evidence.Risk{Filled: "200", Held: "300"}}
	got := calculate(base, snapshot, "100", 8)
	if got.Status != statusAssessed || got.QOffline == nil || *got.QOffline != 5 || got.PlannedRisk == nil || *got.PlannedRisk != "500" {
		t.Fatalf("got %+v", got)
	}
	zero := calculate(base, evidence.Evidence{Plan: evidence.Plan{Budget: "1000"}, Risk: evidence.Risk{Filled: "1000", Held: "0"}}, "100", 8)
	if zero.Status != statusNoBudget || zero.QCandidate != nil || zero.QOffline != nil || zero.PlannedRisk != nil {
		t.Fatalf("exhausted budget=%+v", zero)
	}
}

func TestCalculateRejectsOverdrawnAndSumOverflow(t *testing.T) {
	base := Result{}
	overdraft := calculate(base, evidence.Evidence{Plan: evidence.Plan{Budget: "10"}, Risk: evidence.Risk{Filled: "9", Held: "2"}}, "1", 1)
	if overdraft.Code != "RISK_OVERDRAWN" || overdraft.QOffline != nil {
		t.Fatalf("overdraft=%+v", overdraft)
	}
	max := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 256), big.NewInt(1)).String()
	overflow := calculate(base, evidence.Evidence{Plan: evidence.Plan{Budget: max}, Risk: evidence.Risk{Filled: max, Held: "1"}}, "1", 1)
	if overflow.Code != "RISK_SUM_OVERFLOW" || overflow.QOffline != nil {
		t.Fatalf("overflow=%+v", overflow)
	}
}

func TestCalculateLimitsBeforeUint64Conversion(t *testing.T) {
	max256 := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 256), big.NewInt(1)).String()
	got := calculate(Result{}, evidence.Evidence{Plan: evidence.Plan{Budget: max256}, Risk: evidence.Risk{Filled: "0", Held: "0"}}, "1", math.MaxUint64)
	if got.Status != statusAssessed || got.QOffline == nil || *got.QOffline != math.MaxUint64 {
		t.Fatalf("candidate bound=%+v", got)
	}
	two255 := new(big.Int).Lsh(big.NewInt(1), 255).String()
	got = calculate(Result{}, evidence.Evidence{Plan: evidence.Plan{Budget: max256}, Risk: evidence.Risk{Filled: "0", Held: "0"}}, two255, math.MaxUint64)
	if got.Status != statusAssessed || got.QOffline == nil || *got.QOffline != 1 || got.PlannedRisk == nil || *got.PlannedRisk != two255 {
		t.Fatalf("large unit=%+v", got)
	}
}

func TestBoundedDecimalLimits(t *testing.T) {
	max256 := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 256), big.NewInt(1)).String()
	if len(max256) != 78 {
		t.Fatalf("unexpected max256 decimal length %d", len(max256))
	}
	if _, err := boundedDecimal(max256, true); err != nil {
		t.Fatalf("256-bit bound rejected: %v", err)
	}
	for _, raw := range []string{strings.Repeat("9", 79), new(big.Int).Lsh(big.NewInt(1), 256).String(), "01", "-1", ""} {
		if _, err := boundedDecimal(raw, false); err == nil {
			t.Fatalf("accepted %q", raw)
		}
	}
}
