package strategyengine

import (
	"strings"
	"testing"
	"time"
)

func TestRRDiagnosticDecimalBoundary(t *testing.T) {
	record := (ParkerConservativeLane{}).Evaluate(passingLaneInput(t)).Decision.Record()
	for _, bad := range []string{"1p5", "1P5", "0x10", "1_0", "1e2", "1/2", strings.Repeat("1", 65)} {
		t.Run(bad, func(t *testing.T) {
			r := record
			r.EntryPrice = bad
			if got := ParkerRRFromDecision(r); got.Space.State != RRInvalid || got.Target.State != RRInvalid {
				t.Errorf("planned: %+v", got)
			}
			r = record
			r.LVNSpacePct = bad
			if got := ParkerRRFromDecision(r); got.Space.State != RRInvalid {
				t.Errorf("space: %+v", got)
			}
			r = record
			r.TargetPrice = bad
			if got := ParkerRRFromDecision(r); got.Target.State != RRInvalid {
				t.Errorf("target: %+v", got)
			}
			at := time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC)
			base := RRFill{AverageEntry: "100", InitialStop: "99", Quantity: "2", PriceCurrency: "KRW", PriceBasis: "unadjusted", StopCurrency: "KRW", StopBasis: "unadjusted", AccountCurrency: "KRW", FXRate: "1", FXFrom: "KRW", FXTo: "KRW", FXSource: "identity", EntryAt: at, FXAt: at, GrossPnL: "3", Costs: "1", PnLCurrency: "KRW", CostsCurrency: "KRW", FillSource: "fill", ExitSource: "exit", CostSource: "cost", EntryFinal: true, ExitFinal: true, CostsFinal: true}
			for i := 0; i < 6; i++ {
				f := base
				switch i {
				case 0:
					f.AverageEntry = bad
				case 1:
					f.InitialStop = bad
				case 2:
					f.Quantity = bad
				case 3:
					f.FXRate = bad
				case 4:
					f.GrossPnL = bad
				case 5:
					f.Costs = bad
				}
				if got := ParkerRealizedRR(f); got.State != RRInvalid {
					t.Errorf("field %d: %+v", i, got)
				}
			}
		})
	}
}

func TestRRDiagnosticSignedPnLAndMissingRemainDistinct(t *testing.T) {
	at := time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC)
	f := RRFill{AverageEntry: "100", InitialStop: "99", Quantity: "2", PriceCurrency: "KRW", PriceBasis: "unadjusted", StopCurrency: "KRW", StopBasis: "unadjusted", AccountCurrency: "KRW", FXRate: "1", FXFrom: "KRW", FXTo: "KRW", FXSource: "identity", EntryAt: at, FXAt: at, GrossPnL: "-3", Costs: "1", PnLCurrency: "KRW", CostsCurrency: "KRW", FillSource: "fill", ExitSource: "exit", CostSource: "cost", EntryFinal: true, ExitFinal: true, CostsFinal: true}
	if got := ParkerRealizedRR(f); got.State != RRKnown || got.Ratio != "-2" {
		t.Fatalf("negative pnl: %+v", got)
	}
	f.GrossPnL = ""
	if got := ParkerRealizedRR(f); got.State != RRUnknown {
		t.Fatalf("missing pnl: %+v", got)
	}
}
