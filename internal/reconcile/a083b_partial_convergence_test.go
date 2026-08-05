package reconcile_test

// a083 개정 2 · D10 — 커밋된 수렴은 반환 전에 credit한다 (issues.md B2).
//
// ConvergeQuantities는 심볼별 오류로 루프 *안에서* 반환하는데 crediter 호출은 루프
// *밖*에 있었다. 그래서 이미 커밋된 심볼이 credit 없이 남는다. 그리고 투영이
// 수렴했으므로 다음 비교는 그 심볼에 동의하고 — mismatch.go의 해제 규칙이 적은 대로 —
// 다시 credit받을 길이 없다. 영구 차단이며, a083이 고치려던 결함 그 자체다.
//
// 함수의 doc comment는 이미 "What committed before it stays in the report"라고
// 적고 있었다. 여기서 고치는 것은 credit이 그 약속을 지키게 하는 것이다.

import (
	"context"
	"errors"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/reconcile"
)

// partialMismatch: 첫 심볼은 투영과 맞아 커밋되고, 둘째는 ExpectedPrevQuantity가
// 어긋나 ErrAdjustmentStale로 패스를 멈춘다.
func partialMismatch(t *testing.T) reconcile.Diff {
	t.Helper()
	first := stampedMismatch("AAPL", "10", "7")
	second := stampedMismatch("MSFT", "99", "5")
	first.Quantities = append(first.Quantities, second.Quantities...)
	return first
}

// TestACommittedConvergenceIsCreditedEvenWhenThePassStops: credit을 잃으면 그
// 심볼은 영구 차단이다. 다음 비교가 동의하므로 diff.Quantities에 다시 들어오지
// 않고, 일치하는 심볼에는 아무도 조정을 쓰지 않는다.
func TestACommittedConvergenceIsCreditedEvenWhenThePassStops(t *testing.T) {
	ctx := context.Background()
	j := openJournal(t)
	heldPosition(t, j, "AAPL", "10")
	heldPosition(t, j, "MSFT", "10")
	credit := &recordingCrediter{}

	c := &reconcile.Converger{Journal: j, Credit: credit, AccountRef: "acct-7"}
	report, err := c.ConvergeQuantities(ctx, partialMismatch(t))
	if !errors.Is(err, journal.ErrAdjustmentStale) {
		t.Fatalf("err = %v, want the pass to stop on a stale adjustment", err)
	}

	if len(report.Converged) != 1 || report.Converged[0].Symbol != "AAPL" {
		t.Fatalf("converged = %+v, want AAPL committed before the pass stopped", report.Converged)
	}
	if len(credit.credited) != 1 || len(credit.credited[0]) != 1 || credit.credited[0][0] != "AAPL" {
		t.Fatalf("crediter saw %+v, want one call carrying AAPL: its projection was converged "+
			"and committed, so the next comparison agrees about it and nothing will ever "+
			"write another adjustment for it — an uncredited commit is a block no "+
			"automatic path can lift",
			credit.credited)
	}
	if len(report.Credited) != 1 || report.Credited[0] != "AAPL" {
		t.Errorf("report.Credited = %+v, want AAPL; the field's own doc comment says it is "+
			"the symbols handed to the crediter", report.Credited)
	}
}

// TestTheCreditOnAStoppedPassCarriesTheSameComparison: 부분 credit이라도 스탬프는
// 같은 비교여야 한다. 그것이 없으면 D2b의 순서 규칙이 이 경로에서만 꺼진다.
func TestTheCreditOnAStoppedPassCarriesTheSameComparison(t *testing.T) {
	ctx := context.Background()
	j := openJournal(t)
	heldPosition(t, j, "AAPL", "10")
	heldPosition(t, j, "MSFT", "10")
	credit := &recordingCrediter{}

	diff := partialMismatch(t)
	c := &reconcile.Converger{Journal: j, Credit: credit, AccountRef: "acct-7"}
	if _, err := c.ConvergeQuantities(ctx, diff); !errors.Is(err, journal.ErrAdjustmentStale) {
		t.Fatalf("err = %v, want the pass to stop on a stale adjustment", err)
	}

	if len(credit.comparisons) != 1 {
		t.Fatalf("comparisons = %+v, want exactly one credit", credit.comparisons)
	}
	if credit.comparisons[0] != diff.AsOf {
		t.Errorf("credited comparison = %q, want the diff's as-of %q", credit.comparisons[0], diff.AsOf)
	}
}

// TestNothingIsCreditedWhenNothingCommitted: 첫 심볼부터 멈추면 credit도 없다.
// 부분 credit은 "커밋됐다"는 사실 진술이지 해제 허가가 아니다.
func TestNothingIsCreditedWhenNothingCommitted(t *testing.T) {
	ctx := context.Background()
	j := openJournal(t)
	heldPosition(t, j, "AAPL", "10")
	credit := &recordingCrediter{}

	c := &reconcile.Converger{Journal: j, Credit: credit, AccountRef: "acct-7"}
	if _, err := c.ConvergeQuantities(ctx, stampedMismatch("AAPL", "99", "5")); !errors.Is(err, journal.ErrAdjustmentStale) {
		t.Fatalf("err = %v, want the pass to stop on a stale adjustment", err)
	}
	if len(credit.credited) != 0 {
		t.Fatalf("crediter saw %+v, want nothing: no adjustment committed", credit.credited)
	}
}
