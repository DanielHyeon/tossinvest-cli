package journal

import (
	"context"
	"errors"
	"strings"
	"testing"
)

// adoption_test.go covers the adoption record and the set-once reference
// (change adopt-external-positions tasks 1.1, 1.2 and 1.4).
//
// The two things worth being precise about, because both are safety properties
// rather than conveniences:
//
//	set-once   a live position's synthetic t0 is the baseline its stop is
//	           measured from. Repointing it mid-life would move that baseline,
//	           which §0.9 permits in no direction that is not conservative and
//	           this one is not directional at all.
//	raw basis  the broker's averagePurchasePrice is stored as the string it
//	           arrived as. Whether it includes fees is [미측정 — 2b 실측 대상],
//	           and a re-rendered float would destroy the evidence needed to find
//	           out.

// adoptable inserts an external position — no entry decision — and returns it.
func adoptable(t *testing.T, j *Journal, id string) {
	t.Helper()
	insertPosition(t, j, id, nil)
}

func sampleAdoption(positionID string) AdoptionRequest {
	return AdoptionRequest{
		PositionID:    positionID,
		Symbol:        "005930",
		Market:        "kr",
		Quantity:      "10",
		CostBasis:     "55000.0000",
		ObservedPrice: "70000",
		SyntheticStop: "66500",
		ObservedAt:    "2026-03-30T00:30:00Z",
	}
}

func TestAdoptPositionRecordsAndPointsTheProjection(t *testing.T) {
	j := openTestJournal(t)
	ctx := context.Background()
	adoptable(t, j, "p-ext")

	adoption, err := j.AdoptPosition(ctx, sampleAdoption("p-ext"))
	if err != nil {
		t.Fatalf("AdoptPosition: %v", err)
	}
	if !strings.HasPrefix(adoption.ID, "adopt-") {
		t.Errorf("adoption id = %q, want the derived adopt- form", adoption.ID)
	}
	if adoption.PreimageDigest == "" {
		t.Error("an adoption with no digest cannot be proved to be the source of its baseline")
	}
	// The raw string, character for character. "55000.0000" must not become
	// "55000": the trailing zeros are the broker's own rendering and this column
	// is evidence.
	if adoption.CostBasis != "55000.0000" {
		t.Errorf("cost basis = %q, want the broker's raw string", adoption.CostBasis)
	}
	if adoption.CostBasisSource != CostBasisBrokerAvg {
		t.Errorf("cost basis source = %q, want %q", adoption.CostBasisSource, CostBasisBrokerAvg)
	}

	stored, err := j.LookupPosition(ctx, "p-ext")
	if err != nil {
		t.Fatal(err)
	}
	if stored.AdoptionID != adoption.ID {
		t.Errorf("positions.adoption_id = %q, want %q", stored.AdoptionID, adoption.ID)
	}
	if stored.EntryDecisionID != "" {
		t.Errorf("entry_decision_id = %q; adopting must not write that column", stored.EntryDecisionID)
	}
	if !stored.ExitEligible() || !stored.Adopted() {
		t.Errorf("position after adoption: eligible=%v adopted=%v, want both true",
			stored.ExitEligible(), stored.Adopted())
	}
	// The quantity and the average price are the projection's; an adoption
	// records a t0 and moves neither.
	if stored.Quantity != "10" || stored.AvgPrice != "70000" {
		t.Errorf("projection after adoption = %s @ %s, want it untouched",
			stored.Quantity, stored.AvgPrice)
	}

	byPosition, err := j.AdoptionOf(ctx, "p-ext")
	if err != nil {
		t.Fatalf("AdoptionOf: %v", err)
	}
	if byPosition != adoption {
		t.Errorf("AdoptionOf = %+v, want %+v", byPosition, adoption)
	}
}

func TestAdoptPositionIsSetOnce(t *testing.T) {
	j := openTestJournal(t)
	ctx := context.Background()
	adoptable(t, j, "p-ext")

	first, err := j.AdoptPosition(ctx, sampleAdoption("p-ext"))
	if err != nil {
		t.Fatal(err)
	}

	// A second adoption of the same instance at a different price: the exact
	// case set-once exists for, because it would move a live baseline.
	second := sampleAdoption("p-ext")
	second.ObservedPrice = "80000"
	second.SyntheticStop = "76000"
	second.ObservedAt = "2026-03-30T01:00:00Z"
	if _, err := j.AdoptPosition(ctx, second); !errors.Is(err, ErrPositionAlreadyAdopted) {
		t.Fatalf("re-adopting: %v, want ErrPositionAlreadyAdopted", err)
	}

	stored, err := j.LookupPosition(ctx, "p-ext")
	if err != nil {
		t.Fatal(err)
	}
	if stored.AdoptionID != first.ID {
		t.Errorf("adoption_id = %q after a refused re-adoption, want the original %q",
			stored.AdoptionID, first.ID)
	}
	// The refused adoption must not have left its record behind either: a row in
	// position_adoptions nothing points at is an adoption that never happened.
	var rows int
	if err := j.db.QueryRowContext(ctx, "SELECT count(*) FROM position_adoptions").Scan(&rows); err != nil {
		t.Fatal(err)
	}
	if rows != 1 {
		t.Errorf("position_adoptions rows = %d, want 1", rows)
	}
}

// TestAdoptPositionReplayIsIdempotent is the crash-recovery half: the id is
// derived from what the adoption is, so re-running the identical call after a
// crash between the commit and whatever came next returns the record on disk
// rather than refusing or writing a second one.
func TestAdoptPositionReplayIsIdempotent(t *testing.T) {
	j := openTestJournal(t)
	ctx := context.Background()
	adoptable(t, j, "p-ext")

	first, err := j.AdoptPosition(ctx, sampleAdoption("p-ext"))
	if err != nil {
		t.Fatal(err)
	}
	again, err := j.AdoptPosition(ctx, sampleAdoption("p-ext"))
	if err != nil {
		t.Fatalf("replaying the identical adoption: %v", err)
	}
	if again != first {
		t.Errorf("replay = %+v, want the stored %+v", again, first)
	}
	var rows int
	if err := j.db.QueryRowContext(ctx, "SELECT count(*) FROM position_adoptions").Scan(&rows); err != nil {
		t.Fatal(err)
	}
	if rows != 1 {
		t.Errorf("position_adoptions rows after a replay = %d, want 1", rows)
	}
}

func TestAdoptPositionRefusesAnEngineEnteredPosition(t *testing.T) {
	j := openTestJournal(t)
	ctx := context.Background()
	insertDecision(t, j, "decision-1", "nonce-adopt-1")
	insertPosition(t, j, "p-engine", "decision-1")

	if _, err := j.AdoptPosition(ctx, sampleAdoption("p-engine")); !errors.Is(err, ErrPositionNotAdoptable) {
		t.Fatalf("adopting an engine-entered position: %v, want ErrPositionNotAdoptable", err)
	}
}

func TestAdoptPositionRefusesAnUnusableT0(t *testing.T) {
	ctx := context.Background()

	cases := map[string]func(r *AdoptionRequest){
		"stop at the observation": func(r *AdoptionRequest) { r.SyntheticStop = r.ObservedPrice },
		"stop above it":           func(r *AdoptionRequest) { r.SyntheticStop = "70001" },
		"no observation":          func(r *AdoptionRequest) { r.ObservedPrice = "0" },
		"no quantity":             func(r *AdoptionRequest) { r.Quantity = "0" },
		"no observation instant":  func(r *AdoptionRequest) { r.ObservedAt = "" },
	}
	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			j := openTestJournal(t)
			id := "p-" + strings.ReplaceAll(name, " ", "-")
			adoptable(t, j, id)
			req := sampleAdoption(id)
			mutate(&req)
			if _, err := j.AdoptPosition(ctx, req); !errors.Is(err, ErrInvalidRequest) {
				t.Fatalf("adopting with %s: %v, want ErrInvalidRequest", name, err)
			}
			stored, err := j.LookupPosition(ctx, id)
			if err != nil {
				t.Fatal(err)
			}
			if stored.AdoptionID != "" {
				t.Errorf("a refused adoption left adoption_id = %q", stored.AdoptionID)
			}
		})
	}
}

func TestAdoptPositionRecordsAnAbsentCostBasis(t *testing.T) {
	j := openTestJournal(t)
	ctx := context.Background()
	adoptable(t, j, "p-ext")

	req := sampleAdoption("p-ext")
	req.CostBasis = ""
	adoption, err := j.AdoptPosition(ctx, req)
	if err != nil {
		t.Fatal(err)
	}
	if adoption.CostBasisSource != CostBasisAbsent {
		t.Errorf("cost basis source = %q, want %q", adoption.CostBasisSource, CostBasisAbsent)
	}
	// NULL, not "": the column has to be able to say "the broker reported none".
	var basis any
	if err := j.db.QueryRowContext(ctx,
		"SELECT cost_basis FROM position_adoptions WHERE id = ?", adoption.ID).Scan(&basis); err != nil {
		t.Fatal(err)
	}
	if basis != nil {
		t.Errorf("cost_basis = %v, want NULL when the broker reported none", basis)
	}
}

// TestOpenAdoptedExitStateSeedsFromTheAdoptionRecord is task 1.4's contract: the
// adopted arm of the opening path takes its t0 from the adoption record and
// looks no decision up, and the watermark seeds itself from the entry price
// exactly as the engine-entered arm's does.
func TestOpenAdoptedExitStateSeedsFromTheAdoptionRecord(t *testing.T) {
	j := openTestJournal(t)
	ctx := context.Background()
	adoptable(t, j, "p-ext")

	adoption, err := j.AdoptPosition(ctx, sampleAdoption("p-ext"))
	if err != nil {
		t.Fatal(err)
	}
	state, err := j.OpenAdoptedExitState(ctx, "p-ext")
	if err != nil {
		t.Fatalf("OpenAdoptedExitState: %v", err)
	}
	if state.EntryPrice != adoption.ObservedPrice {
		t.Errorf("entry price = %q, want the observation %q", state.EntryPrice, adoption.ObservedPrice)
	}
	if state.InitialStop != adoption.SyntheticStop {
		t.Errorf("initial stop = %q, want the synthetic stop %q", state.InitialStop, adoption.SyntheticStop)
	}
	if state.Baseline != adoption.SyntheticStop {
		t.Errorf("baseline = %q, want the synthetic stop at t0", state.Baseline)
	}
	// R = 0 at t0: the watermark is the entry, so nothing has been reached.
	if state.HighWater != adoption.ObservedPrice {
		t.Errorf("high water = %q, want the entry price %q", state.HighWater, adoption.ObservedPrice)
	}
	if state.RatchetLevel != RatchetNone {
		t.Errorf("ratchet level = %q, want %q", state.RatchetLevel, RatchetNone)
	}
	if state.InitialRisk != "3500" {
		t.Errorf("initial risk = %q, want 70000 − 66500", state.InitialRisk)
	}
}

// TestOpenExitStateRefusesAPositionWithNeitherRecord keeps the negative half of
// the single predicate: widening eligibility must not have widened it to
// everything.
func TestOpenExitStateRefusesAPositionWithNeitherRecord(t *testing.T) {
	j := openTestJournal(t)
	ctx := context.Background()
	adoptable(t, j, "p-bare")

	_, err := j.OpenExitState(ctx, ExitStateSeed{
		PositionID: "p-bare", EntryPrice: "70000", InitialStop: "66500",
	})
	if !errors.Is(err, ErrPositionNotExitEligible) {
		t.Fatalf("opening an exit state on an unjustified position: %v, want ErrPositionNotExitEligible", err)
	}
}
