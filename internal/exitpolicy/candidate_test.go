package exitpolicy_test

// candidate_test.go covers the composition on its own. The ratchet exercises it
// through the trigger table, but the R4 invariant and the tie-break are the
// original's own contract (protected_stop_candidate.py:104-121) and belong to
// the function that owns them — exit-policy names the composition explicitly, so
// a change to it must fail here and not only in a level's arithmetic.

import (
	"errors"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/exitpolicy"
)

func TestTheCompositionTakesTheMaximum(t *testing.T) {
	t.Parallel()

	out, err := exitpolicy.ComputeProtectedStop(
		exitpolicy.BuildCandidates("9800", "10030", "10160"), "9800")
	if err != nil {
		t.Fatalf("ComputeProtectedStop: %v", err)
	}
	if out.Price != "10160" {
		t.Errorf("price = %s, want the largest candidate", out.Price)
	}
	if out.WinningReason != exitpolicy.CandidateBaselineRatchet {
		t.Errorf("winning reason = %s, want %s", out.WinningReason, exitpolicy.CandidateBaselineRatchet)
	}
	if len(out.Rejected) != 2 {
		t.Errorf("rejected = %+v, want the two that lost", out.Rejected)
	}
}

// TestTheR4FloorHolds is the invariant the module exists to enforce in one
// place: the result is never below the anchor.
func TestTheR4FloorHolds(t *testing.T) {
	t.Parallel()

	out, err := exitpolicy.ComputeProtectedStop(
		exitpolicy.BuildCandidates("10500", "10030", "10160"), "10500")
	if err != nil {
		t.Fatalf("ComputeProtectedStop: %v", err)
	}
	if out.Price != "10500" {
		t.Errorf("price = %s, want the anchor to hold at 10500", out.Price)
	}
	if out.WinningReason != exitpolicy.CandidatePrevious {
		t.Errorf("winning reason = %s, want %s", out.WinningReason, exitpolicy.CandidatePrevious)
	}
}

// TestALowerCandidateIsReportedRejectedRatherThanDropped is why the floor is
// applied after the max and not folded into it: an operator reading "previous
// won" needs to see what proposed something lower.
func TestALowerCandidateIsReportedRejectedRatherThanDropped(t *testing.T) {
	t.Parallel()

	out, err := exitpolicy.ComputeProtectedStop(
		[]exitpolicy.StopCandidate{{Name: exitpolicy.CandidateBaselineRatchet, Price: "9900"}}, "10500")
	if err != nil {
		t.Fatalf("ComputeProtectedStop: %v", err)
	}
	if len(out.Rejected) != 1 || out.Rejected[0].Price != "9900" {
		t.Errorf("rejected = %+v, want the candidate that lost to the floor", out.Rejected)
	}
}

// TestTheTieBreakIsCandidateOrder is protected_stop_candidate.py:75-78: when two
// candidates propose the same price the earlier one wins. It is what attributes
// a BREAKEVEN promotion to `real_breakeven` rather than to `baseline_ratchet`.
func TestTheTieBreakIsCandidateOrder(t *testing.T) {
	t.Parallel()

	out, err := exitpolicy.ComputeProtectedStop(
		exitpolicy.BuildCandidates("9800", "10030", "10030"), "9800")
	if err != nil {
		t.Fatalf("ComputeProtectedStop: %v", err)
	}
	if out.WinningReason != exitpolicy.CandidateRealBreakeven {
		t.Errorf("winning reason = %s, want %s", out.WinningReason, exitpolicy.CandidateRealBreakeven)
	}
}

func TestAnAnchorAloneIsReturnedUnchanged(t *testing.T) {
	t.Parallel()

	out, err := exitpolicy.ComputeProtectedStop(
		exitpolicy.BuildCandidates("9800", "", ""), "9800")
	if err != nil {
		t.Fatalf("ComputeProtectedStop: %v", err)
	}
	if out.Price != "9800" || out.WinningReason != exitpolicy.CandidatePrevious {
		t.Errorf("out = %+v, want the anchor unchanged", out)
	}
}

// TestTheAnchorIsRequired is the one behavioural difference from the original,
// which accepts previous_stop=None and raises only when nothing at all is valid.
// TossOS has no state in which a position is open without a baseline.
func TestTheAnchorIsRequired(t *testing.T) {
	t.Parallel()

	_, err := exitpolicy.ComputeProtectedStop(
		exitpolicy.BuildCandidates("", "", "10160"), "")
	if !errors.Is(err, exitpolicy.ErrRefused) {
		t.Fatalf("err = %v, want a refusal — the baseline column is NOT NULL", err)
	}
}

func TestAnUnreadableCandidateIsRefusedRatherThanSkipped(t *testing.T) {
	t.Parallel()

	_, err := exitpolicy.ComputeProtectedStop(
		exitpolicy.BuildCandidates("9800", "not a price", ""), "9800")
	if !errors.Is(err, exitpolicy.ErrRefused) {
		t.Fatalf("err = %v, want a refusal; skipping it would silently drop a term of the max", err)
	}
}

// TestOnlyTheThreePortedCandidatesExist pins the exclusion. Six of the
// original's nine are signal-derived and this change admits zero signal inputs;
// a tenth appearing here without the exclusion being revisited is what this
// catches.
func TestOnlyTheThreePortedCandidatesExist(t *testing.T) {
	t.Parallel()

	built := exitpolicy.BuildCandidates("1", "2", "3")
	want := []string{
		exitpolicy.CandidatePrevious,
		exitpolicy.CandidateRealBreakeven,
		exitpolicy.CandidateBaselineRatchet,
	}
	if len(built) != len(want) {
		t.Fatalf("built %d candidates, want %d", len(built), len(want))
	}
	for i, name := range want {
		if built[i].Name != name {
			t.Errorf("candidate %d = %s, want %s (the original's relative order)", i, built[i].Name, name)
		}
	}
}
