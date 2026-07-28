package candidate

// negative_request_test.go covers the two boundary refusals this change added for
// the requested row count, and it exists because producing the Branch Test Map for
// section 3 found that neither of them was reached by anything.
//
// # Why an unreached refusal is worth a test rather than a note
//
// The count is the one integer in this store where 0 is not a smaller quantity: it
// means "nobody recorded a request", and truncationOf answers TruncationUnknown for
// it deliberately. A negative would be swallowed by that same arm — truncationOf
// tests `requested <= 0` — so a row carrying -1 would be stored, read back, and
// reported as a reading nobody measured, which is the one spelling in this package
// that is supposed to mean the fact was never taken. The two guards below are what
// stop a number nothing could have produced from arriving dressed as an absence.
//
// truncation_test.go already names the case: its table has "a negative request,
// which validate refuses at the boundary". That sentence was the only thing holding
// the claim up. This is the test behind it.

import (
	"context"
	"strings"
	"testing"
)

// TestANegativeRequestedCountIsRefusedByTheObservationBoundary.
func TestANegativeRequestedCountIsRefusedByTheObservationBoundary(t *testing.T) {
	// The arithmetic that makes the refusal load-bearing: truncationOf folds a
	// negative into the same answer it gives an unrecorded request, so nothing
	// downstream can tell them apart once a row like this is stored.
	if got := truncationOf(-1, 100); got != TruncationUnknown() {
		t.Fatalf("truncationOf(-1, 100) = %s, want unknown — if a negative is now "+
			"distinguishable downstream then this refusal is about something else", got)
	}

	s, _ := openStore(t)
	ctx := context.Background()
	err := s.RecordObservations(ctx, []Observation{{
		Market: MarketKR, Symbol: "005930", Source: SourceOfficialTradingValue,
		ObservedAt: t0,
		Reported:   Reported{Rank: 1, RankTotal: 100, RankRequested: -1},
	}})
	if err == nil {
		t.Fatal("an observation asking for -1 rows was recorded. A negative is not a " +
			"smaller absence: truncationOf reads it as unknown, so the row would read " +
			"back as a reading nobody measured rather than as the impossible number it is")
	}
	if !strings.Contains(err.Error(), "absent is 0, not a negative") {
		t.Errorf("the refusal reads %q; it has to name what is wrong with the number, "+
			"because the caller's next move is to send 0", err)
	}

	// And nothing was written — the validation pass runs over the whole batch before
	// the transaction opens, so a rejected row cannot leave a partial one behind.
	rows, err := s.Observations(ctx, MarketKR, "005930", t0)
	if err != nil {
		t.Fatalf("Observations: %v", err)
	}
	if len(rows) != 0 {
		t.Errorf("%d observation(s) survived a refused batch", len(rows))
	}
}

// TestANegativeRequestedCountIsRefusedByTheFirstRankWrite is the same guard on the
// column that outlives the observations table.
//
// It matters more here than one field over: first_rank is written once, so a
// position stored with a nonsense qualifier answers for the rest of the candidate's
// life and nothing can correct it — the facts belong to a reading that is gone.
func TestANegativeRequestedCountIsRefusedByTheFirstRankWrite(t *testing.T) {
	s, _ := openStore(t)
	ctx := context.Background()
	if _, err := s.Promote(ctx, MarketKR, "005930", t0); err != nil {
		t.Fatalf("Promote: %v", err)
	}

	bad := storedFirstRank(4, 100, t0, SourceOfficialTradingValue)
	bad.Requested = -1
	if _, err := s.NoteFirstRank(ctx, MarketKR, "005930", bad); err == nil {
		t.Fatal("NoteFirstRank accepted a position whose reading asked for -1 rows")
	} else if !strings.Contains(err.Error(), "absent is 0, not a negative") {
		t.Errorf("the refusal reads %q, and it has to say which number is wrong", err)
	}

	// Refused before anything is written, so the write-once column is still free for
	// the next reading that can qualify.
	stored, _, err := s.FirstRank(ctx, MarketKR, "005930")
	if err != nil {
		t.Fatalf("FirstRank: %v", err)
	}
	if stored.Recorded() {
		t.Errorf("the refused write stored %d of %d anyway; first_rank is written once, so "+
			"this position would answer for the rest of the candidate's life",
			stored.Rank, stored.Total)
	}

	// The same position with the count absent rather than negative is accepted, which
	// is what makes the refusal about the sign rather than about the qualifier being
	// missing. Absent is an ordinary state — every row written before schema 4 has it.
	ok := storedFirstRank(4, 100, t0, SourceOfficialTradingValue)
	ok.Requested = 0
	if _, err := s.NoteFirstRank(ctx, MarketKR, "005930", ok); err != nil {
		t.Fatalf("a position with no recorded request was refused: %v — absent is the "+
			"state of every row an older build wrote", err)
	}
}
