package positioncampaign

import "testing"

func TestOrderWatermarksReplacementAndLateTerminalFill(t *testing.T) {
	ledger, err := NewLegLedger("10")
	if err != nil {
		t.Fatal(err)
	}
	if err := ledger.LinkOrder("old", "", "10"); err != nil {
		t.Fatal(err)
	}
	first, err := ledger.Observe(OrderObservation{OrderID: "old", Cumulative: "4"})
	if err != nil || first.Delta != "4" || ledger.Filled != "4" || ledger.Residual != "6" {
		t.Fatalf("first=%+v ledger=%+v err=%v", first, ledger, err)
	}
	if err := ledger.MarkTerminal("old"); err != nil {
		t.Fatal(err)
	}
	if err := ledger.LinkOrder("new", "old", "6"); err != nil {
		t.Fatal(err)
	}
	if got := ledger.Orders["new"].CarryBaseline; got != "4" {
		t.Fatalf("carry baseline=%q, want 4", got)
	}
	if got := ledger.Orders["new"].Remaining; got != "6" {
		t.Fatalf("successor remaining=%q, want 6", got)
	}
	if _, err := ledger.Observe(OrderObservation{OrderID: "new", Cumulative: "2"}); err != nil {
		t.Fatal(err)
	}
	late, err := ledger.Observe(OrderObservation{OrderID: "old", Cumulative: "5"})
	if err != nil {
		t.Fatal(err)
	}
	if late.Delta != "1" || !late.LateTerminal || !late.Reconcile {
		t.Fatalf("late=%+v, want preserved delta and reconcile", late)
	}
	if ledger.Filled != "7" || ledger.Residual != "3" || ledger.Orders["new"].Remaining != "3" {
		t.Fatalf("ledger after late fill=%+v", ledger)
	}
	retry, err := ledger.Observe(OrderObservation{OrderID: "old", Cumulative: "5"})
	if err != nil || retry.Delta != "0" || ledger.Filled != "7" {
		t.Fatalf("retry=%+v ledger=%+v err=%v", retry, ledger, err)
	}
	lower, err := ledger.Observe(OrderObservation{OrderID: "old", Cumulative: "3"})
	if err != nil || lower.Delta != "0" || ledger.Orders["old"].Cumulative != "5" {
		t.Fatalf("lower=%+v ledger=%+v err=%v", lower, ledger, err)
	}
}

func TestOrderWatermarkCapExcessAndAmbiguousLineagePreserveDelta(t *testing.T) {
	ledger, _ := NewLegLedger("3")
	_ = ledger.LinkOrder("order", "", "3")
	out, err := ledger.Observe(OrderObservation{OrderID: "order", Cumulative: "4", LineageAmbiguous: true})
	if err != nil {
		t.Fatal(err)
	}
	if out.Delta != "4" || !out.CapExceeded || !out.Reconcile || ledger.Filled != "4" || ledger.Residual != "0" {
		t.Fatalf("out=%+v ledger=%+v", out, ledger)
	}
}
