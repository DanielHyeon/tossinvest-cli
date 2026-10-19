package strategyengine

import (
	"reflect"
	"testing"
	"time"
)

func TestParkerRRDiagnosticsFrozenFixture(t *testing.T) {
	in := passingLaneInput(t)
	before := (ParkerConservativeLane{}).Evaluate(in)
	record := before.Decision.Record()
	got := ParkerRRFromDecision(record)
	if got.Space.State != RRKnown || got.Space.Ratio != "12/7" || got.Target.State != RRKnown || got.Target.Ratio != "3" || got.Realized.State != RRUnknown {
		t.Fatalf("diagnostics: %+v", got)
	}
	if got.Space.NumeratorUnit != record.Currency+"/share" || got.Space.PriceBasis != "unadjusted" || got.Space.PolicyVersion != LaneVersion {
		t.Fatalf("provenance: %+v", got.Space)
	}
	after := (ParkerConservativeLane{}).Evaluate(in)
	if !reflect.DeepEqual(before, after) {
		t.Fatal("frozen decision changed")
	}
}

func TestParkerRRDiagnosticsRealized(t *testing.T) {
	fill := RRFill{AverageEntry: "100", InitialStop: "99.3", Quantity: "10", PriceCurrency: "USD", PriceBasis: "unadjusted", StopCurrency: "USD", StopBasis: "unadjusted", FXFrom: "USD", FXTo: "KRW", AccountCurrency: "KRW", FXRate: "1300", FXSource: "execution", EntryAt: time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC), FXAt: time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC), GrossPnL: "13000", Costs: "0", PnLCurrency: "KRW", CostsCurrency: "KRW", FillSource: "execution", ExitSource: "execution", CostSource: "execution", EntryFinal: true, ExitFinal: true, CostsFinal: true}
	original := fill
	got := ParkerRealizedRR(fill)
	if got.State != RRKnown || got.Ratio != "10/7" || got.Denominator != "9100" || got.DenominatorUnit != "KRW" {
		t.Fatalf("got %+v", got)
	}
	if !reflect.DeepEqual(fill, original) {
		t.Fatal("input mutated")
	}
	fill.GrossPnL = "0"
	if got := ParkerRealizedRR(fill); got.State != RRKnown || got.Ratio != "0" {
		t.Fatalf("zero pnl: %+v", got)
	}
	for _, field := range []string{"fill", "cost", "fx", "exit"} {
		x := original
		switch field {
		case "fill":
			x.FillSource = ""
		case "cost":
			x.Costs = ""
		case "fx":
			x.FXAt = time.Time{}
		case "exit":
			x.ExitSource = ""
		}
		if got := ParkerRealizedRR(x); got.State != RRUnknown || got.Ratio != "" {
			t.Errorf("%s: %+v", field, got)
		}
	}
	for _, field := range []string{"riskzero", "risknegative", "quantity", "fx", "currency", "basis", "cost"} {
		x := original
		switch field {
		case "riskzero":
			x.InitialStop = "100"
		case "risknegative":
			x.InitialStop = "101"
		case "quantity":
			x.Quantity = "0"
		case "fx":
			x.FXRate = "-1"
		case "currency":
			x.PnLCurrency = "USD"
		case "basis":
			x.StopBasis = "adjusted"
		case "cost":
			x.Costs = "-1"
		}
		if got := ParkerRealizedRR(x); got.State != RRInvalid || got.Ratio != "" {
			t.Errorf("%s: %+v", field, got)
		}
	}
}

func TestParkerRRDiagnosticsInvalidOrMissingPlan(t *testing.T) {
	record := (ParkerConservativeLane{}).Evaluate(passingLaneInput(t)).Decision.Record()
	for _, tc := range []struct {
		name   string
		mutate func(*DecisionRecord)
		state  RRState
	}{
		{"zero risk", func(r *DecisionRecord) { r.StopPrice = r.EntryPrice }, RRInvalid},
		{"negative risk", func(r *DecisionRecord) { r.StopPrice = "200" }, RRInvalid},
		{"missing stop", func(r *DecisionRecord) { r.StopPrice = "" }, RRUnknown},
		{"missing currency", func(r *DecisionRecord) { r.Currency = "" }, RRUnknown},
		{"bad target", func(r *DecisionRecord) { r.TargetPrice = "-1" }, RRInvalid},
	} {
		t.Run(tc.name, func(t *testing.T) {
			x := record
			tc.mutate(&x)
			got := ParkerRRFromDecision(x)
			if tc.name == "bad target" {
				if got.Target.State != tc.state {
					t.Fatalf("%+v", got)
				}
				return
			}
			if got.Space.State != tc.state || got.Target.State != tc.state {
				t.Fatalf("%+v", got)
			}
		})
	}
}

func TestParkerRealizedRRRejectsMixedUnitsAndUnavailableEvidence(t *testing.T) {
	base := RRFill{AverageEntry: "100", InitialStop: "99", Quantity: "2", PriceCurrency: "KRW", PriceBasis: "unadjusted", StopCurrency: "KRW", StopBasis: "unadjusted", AccountCurrency: "KRW", FXRate: "1", FXFrom: "KRW", FXTo: "KRW", FXSource: "identity", EntryAt: time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC), FXAt: time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC), GrossPnL: "3", Costs: "1", PnLCurrency: "KRW", CostsCurrency: "KRW", FillSource: "fill", ExitSource: "exit", CostSource: "cost", EntryFinal: true, ExitFinal: true, CostsFinal: true}
	if got := ParkerRealizedRR(base); got.State != RRKnown || got.Ratio != "1" {
		t.Fatalf("same currency: %+v", got)
	}
	for _, tc := range []struct {
		name   string
		mutate func(*RRFill)
		want   RRState
	}{
		{"cost currency", func(f *RRFill) { f.CostsCurrency = "USD" }, RRInvalid},
		{"later fx", func(f *RRFill) { f.FXAt = f.EntryAt.Add(time.Second) }, RRInvalid},
		{"stop currency", func(f *RRFill) { f.StopCurrency = "USD" }, RRInvalid},
		{"FX direction", func(f *RRFill) { f.FXFrom = "USD" }, RRInvalid},
		{"same currency FX", func(f *RRFill) { f.FXRate = "2" }, RRInvalid},
		{"entry not final", func(f *RRFill) { f.EntryFinal = false }, RRUnknown},
		{"exit not final", func(f *RRFill) { f.ExitFinal = false }, RRUnknown},
		{"costs not final", func(f *RRFill) { f.CostsFinal = false }, RRUnknown},
		{"missing quantity", func(f *RRFill) { f.Quantity = "" }, RRUnknown},
		{"missing gross pnl", func(f *RRFill) { f.GrossPnL = "" }, RRUnknown},
		{"missing cost source", func(f *RRFill) { f.CostSource = "" }, RRUnknown},
		{"malformed entry", func(f *RRFill) { f.AverageEntry = "NaN" }, RRInvalid},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := base
			tc.mutate(&f)
			got := ParkerRealizedRR(f)
			if got.State != tc.want || got.Ratio != "" {
				t.Fatalf("%+v", got)
			}
		})
	}
}

func TestParkerRRSeparatesPlannedRiskFromFilledRisk(t *testing.T) {
	record := (ParkerConservativeLane{}).Evaluate(passingLaneInput(t)).Decision.Record()
	planned := ParkerRRFromDecision(record)
	fill := RRFill{AverageEntry: "100.2", InitialStop: record.StopPrice, Quantity: "10", PriceCurrency: record.Currency, StopCurrency: record.Currency, PriceBasis: "unadjusted", StopBasis: "unadjusted", AccountCurrency: record.Currency, FXRate: "1", FXFrom: record.Currency, FXTo: record.Currency, FXSource: "identity", EntryAt: time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC), FXAt: time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC), GrossPnL: "10", Costs: "0", PnLCurrency: record.Currency, CostsCurrency: record.Currency, FillSource: "entry", ExitSource: "exit", CostSource: "cost", EntryFinal: true, ExitFinal: true, CostsFinal: true}
	realized := ParkerRealizedRR(fill)
	if planned.Target.Denominator != "7007/10000" || realized.State != RRKnown || realized.Denominator != "8007/1000" {
		t.Fatalf("planned=%+v realized=%+v", planned.Target, realized)
	}
}

func TestParkerRRExactExampleAndUnknownMetadata(t *testing.T) {
	record := (ParkerConservativeLane{}).Evaluate(passingLaneInput(t)).Decision.Record()
	record.EntryPrice, record.StopPrice, record.TargetPrice, record.LVNSpacePct = "100", "99.3", "102.1", "1.2"
	got := ParkerRRFromDecision(record)
	if got.Space.Numerator != "6/5" || got.Space.Denominator != "7/10" || got.Space.Ratio != "12/7" || got.Target.Ratio != "3" || got.Realized.State != RRUnknown {
		t.Fatalf("%+v", got)
	}
	record.StopPrice = ""
	got = ParkerRRFromDecision(record)
	if got.Space.State != RRUnknown || got.Space.Reason == "" || got.Space.PolicyVersion != LaneVersion || got.Space.NumeratorUnit != record.Currency+"/share" || got.Space.PriceBasis != "unadjusted" {
		t.Fatalf("%+v", got.Space)
	}
}
