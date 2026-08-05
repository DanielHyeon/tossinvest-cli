package reconcile_test

// a083 개정 2 · D7·D8·D9 — credit의 *적용 범위* (issues.md B1).
//
// 개정 1은 credit의 수명만 규정했다: 엄격히 나중의 비교만 credit을 쓸 수 있다.
// 규정하지 않은 것은 credit이 *어느 차단*을 풀 수 있는가였고, 독립 리뷰가 그 구멍으로
// 두 가지를 통과시켰다.
//
//   D7  credit이 계산되기 *전에* 존재하지 않던 차단을 그 credit이 푼다.
//   D8  불일치가 QuantityMismatch에서 ExternalPos로 재분류되면 비교가 "clean"으로
//       보이고 차단이 풀린다 — 불일치는 사라지지 않았고 방향만 뒤집혔다.
//   D9  답할 차단이 없는 credit이 무기한 남아 미래의 차단을 기다린다.
//
// 세 규칙 모두 해제 조건을 좁히는 방향이므로 §0.9를 만족한다.

import (
	"context"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/reconcile"
)

// authorityStore lets a test say what the journal reports as active, which is
// how a block entered by another producer — an oversell projection refusal, an
// orphan fill — reaches the tracker.
type authorityStore struct {
	reconcile.ReconcileStore
	states []journal.ReconcileState
}

func (s *authorityStore) ActiveReconcileStates(context.Context) ([]journal.ReconcileState, error) {
	return s.states, nil
}

func trackerWithStore(clk clock.Clock, gate *execgw.EntryGate, store reconcile.ReconcileStore) *reconcile.Tracker {
	return &reconcile.Tracker{
		Clock: clk, Gate: gate, Journal: store,
		MinInterval: 30 * time.Second, MaxFailures: 3, AccountRef: "acct-7",
	}
}

// externalPosDiffAt: 계좌가 들고 있고 엔진은 모르는 포지션. BlocksEntry()는 이것을
// 세지 않으므로 비교 전체는 "막지 않는다"로 보인다.
func externalPosDiffAt(clk clock.Clock, symbol, quantity string) reconcile.Diff {
	return reconcile.Diff{
		AsOf:       asOfAt(clk),
		AccountRef: "acct-7",
		ExternalPos: []reconcile.ExternalPosition{{
			Holding:    reconcile.Holding{Symbol: symbol, Quantity: quantity, Market: "us"},
			Provenance: reconcile.ProvenanceExternal,
		}},
	}
}

// TestAReclassifiedDisagreementDoesNotReleaseTheBlock (D8).
//
// 재현: 브로커 holdings가 한 주기 동안 심볼을 누락한다 → QuantityMismatch(local 10,
// broker 0) → 차단, 그리고 수렴이 계좌의 0을 투영에 쓴다. 다음 주기에 보유가 돌아온다
// → local 0, broker 10 → ExternalPos. 진입을 막지 않는 분류이므로 옛 규칙에서는
// 차단이 풀리고 신규 진입이 재개된다. **엔진 투영은 0이고 계좌는 10이다.**
func TestAReclassifiedDisagreementDoesNotReleaseTheBlock(t *testing.T) {
	clk := clock.NewFake(asOf)
	gate := execgw.NewEntryGate(clk, map[execgw.RequiredQuery]time.Duration{})
	tracker := newTracker(clk, gate)

	comparison := mismatchDiffAt(clk, "AAPL", "10", "0")
	observe(t, tracker, comparison)
	tracker.AdjustmentApplied(comparison.AsOf, "AAPL")

	clk.Advance(60 * time.Second)
	out := observe(t, tracker, externalPosDiffAt(clk, "AAPL", "10"))

	if len(out.Cleared) != 0 {
		t.Fatalf("cleared = %+v, want nothing: the account holds 10 and the projection holds 0. "+
			"The disagreement was reclassified, not resolved, and releasing here reopens "+
			"entries on a position the engine cannot account for", out.Cleared)
	}
	if gate.CheckEntryFor("us", "AAPL") == nil {
		t.Fatal("entries were reopened while the engine and the account still disagree")
	}
}

// TestAReclassifiedDisagreementRefutesTheCredit (D8, 정산 쪽).
//
// 재분류가 해제를 막기만 하고 credit을 남겨 두면, 그 다음 정말로 clean한 비교가
// 옛 credit으로 차단을 푼다. 재분류는 반박이어야 한다.
func TestAReclassifiedDisagreementRefutesTheCredit(t *testing.T) {
	clk := clock.NewFake(asOf)
	gate := execgw.NewEntryGate(clk, map[execgw.RequiredQuery]time.Duration{})
	tracker := newTracker(clk, gate)

	comparison := mismatchDiffAt(clk, "AAPL", "10", "0")
	observe(t, tracker, comparison)
	tracker.AdjustmentApplied(comparison.AsOf, "AAPL")

	clk.Advance(60 * time.Second)
	observe(t, tracker, externalPosDiffAt(clk, "AAPL", "10"))

	clk.Advance(60 * time.Second)
	out := observe(t, tracker, cleanDiffAt(clk))
	if len(out.Cleared) != 0 {
		t.Fatalf("cleared = %+v, want nothing: the comparison in between disagreed about this "+
			"symbol, so the adjustment did not settle it and a new one is needed", out.Cleared)
	}
	if len(out.AwaitingAdjustment) != 1 || out.AwaitingAdjustment[0].Symbol != "AAPL" {
		t.Fatalf("awaiting = %+v, want AAPL still waiting on an adjustment", out.AwaitingAdjustment)
	}
}

// TestACreditIsSpentByTheFirstComparisonThatCouldUseIt (D7, 개정 2 재작성).
//
// 최초 D7은 `Block.Since`를 credit의 비교 스탬프와 비교했다. **비교 대상이 틀렸다** —
// `Since`는 사이클 *끝*에 쓰이는 `reddconcile_states.entered_at`이고 credit은 스냅샷
// 읽기가 *시작된* 시각을 들고 있다. 실계좌에서는 그 사이에 사이클 전체가 들어가므로
// 정상 경로가 영원히 해제되지 않았다 — a083이 고치려던 결함 그 자체.
//
// 대신 credit의 수명을 **해제 기회 한 번**으로 묶는다. 시계 비교가 없고, 6시간 묵은
// credit이 나중에 생긴 남의 차단을 푸는 경로도 같이 사라진다.
func TestACreditIsSpentByTheFirstComparisonThatCouldUseIt(t *testing.T) {
	clk := clock.NewFake(asOf)
	gate := execgw.NewEntryGate(clk, map[execgw.RequiredQuery]time.Duration{})
	store := &authorityStore{ReconcileStore: openJournal(t)}
	tracker := trackerWithStore(clk, gate, store)

	comparison := mismatchDiffAt(clk, "AAPL", "10", "4")
	observe(t, tracker, comparison)
	tracker.AdjustmentApplied(comparison.AsOf, "AAPL")

	// The re-read after the adjustment agrees and releases. That is the credit's
	// one opportunity, and it took it.
	clk.Advance(60 * time.Second)
	if out := observe(t, tracker, cleanDiffAt(clk)); len(out.Cleared) != 1 {
		t.Fatalf("cleared = %+v (awaiting %+v), want the ordinary release: a credit that "+
			"cannot spend itself on the block it converged is the a083 defect",
			out.Cleared, out.AwaitingAdjustment)
	}

	// Six hours later another producer enters its own QUANTITY_MISMATCH block on
	// the same symbol. Nothing reconciled it, and no credit is left to pretend
	// otherwise.
	clk.Advance(6 * time.Hour)
	store.states = []journal.ReconcileState{{
		ID: "rc-oversell", AccountRef: "acct-7", Symbol: "AAPL", ScopeMarket: "US",
		Cause:     journal.ReconcileCauseQuantityMismatch,
		Evidence:  "oversell: the fill stream sold more than the projection holds",
		EnteredAt: clk.Now(),
	}}
	if err := tracker.Refresh(context.Background()); err != nil {
		t.Fatalf("Refresh: %v", err)
	}

	clk.Advance(60 * time.Second)
	out := observe(t, tracker, cleanDiffAt(clk))
	if len(out.Cleared) != 0 {
		t.Fatalf("cleared = %+v, want nothing: the credit was spent six hours ago on the "+
			"block it actually converged. A credit answers the disagreement it settled, "+
			"not every later one that happens to share a symbol", out.Cleared)
	}
	if gate.CheckEntryFor("us", "AAPL") == nil {
		t.Fatal("entries were reopened on an oversell nobody reconciled")
	}
}

// TestACreditWithNothingLeftToAnswerIsDiscarded (D9).
//
// TTL 대신 참조 무결성으로 만료를 만든다. 쓸 수 있게 된 credit에 그것이 풀 수 있는
// 차단이 하나도 없으면 버린다 — 남겨 두면 미래의 차단만 기다리게 되고 D7이 이미
// 그것을 금지한다. 그래서 버리는 것이 항상 안전하고 t.adjusted가 차단된 심볼 수로
// 묶인다.
func TestACreditWithNothingLeftToAnswerIsDiscarded(t *testing.T) {
	clk := clock.NewFake(asOf)
	gate := execgw.NewEntryGate(clk, map[execgw.RequiredQuery]time.Duration{})
	tracker := newTracker(clk, gate)

	// A credit for a symbol that carries no block at all.
	tracker.AdjustmentApplied(asOfAt(clk), "AAPL")
	clk.Advance(60 * time.Second)
	observe(t, tracker, cleanDiffAt(clk))

	// Later the symbol does earn a block, and no one adjusts it.
	clk.Advance(60 * time.Second)
	observe(t, tracker, mismatchDiffAt(clk, "AAPL", "10", "4"))

	clk.Advance(60 * time.Second)
	out := observe(t, tracker, cleanDiffAt(clk))
	if len(out.Cleared) != 0 {
		t.Fatalf("cleared = %+v, want nothing: the credit that released it was written before "+
			"this block existed and had already answered nothing", out.Cleared)
	}
	if len(out.AwaitingAdjustment) != 1 {
		t.Fatalf("awaiting = %+v, want the block still waiting on an adjustment of its own",
			out.AwaitingAdjustment)
	}
}

// TestAnUnrelatedMismatchStillDoesNotDiscardAnAnsweredCredit: 개정 2의 세 규칙이
// 개정 1의 D2b를 되돌리지 않는지 고정한다. 다른 심볼이 diff를 계속 dirty하게 두는
// 동안, 일치하는 심볼의 credit은 살아 있어야 한다 — 일치하는 심볼에는 아무도 조정을
// 쓰지 않으므로 그것을 버리면 영구 차단이다.
func TestAnUnrelatedMismatchStillDoesNotDiscardAnAnsweredCredit(t *testing.T) {
	clk := clock.NewFake(asOf)
	gate := execgw.NewEntryGate(clk, map[execgw.RequiredQuery]time.Duration{})
	tracker := newTracker(clk, gate)

	both := mismatchDiffAt(clk, "AAPL", "10", "4")
	both.Quantities = append(both.Quantities, mismatchDiffAt(clk, "MSFT", "8", "3").Quantities...)
	observe(t, tracker, both)
	tracker.AdjustmentApplied(both.AsOf, "AAPL", "MSFT")

	// AAPL agrees now; MSFT does not, so the whole comparison still blocks and the
	// release branch never runs. AAPL's credit must survive that.
	clk.Advance(60 * time.Second)
	observe(t, tracker, mismatchDiffAt(clk, "MSFT", "8", "3"))

	clk.Advance(60 * time.Second)
	out := observe(t, tracker, cleanDiffAt(clk))
	if len(out.Cleared) != 1 || out.Cleared[0].Symbol != "AAPL" {
		t.Fatalf("cleared = %+v (awaiting %+v), want AAPL: its adjustment was never refuted, "+
			"and an agreeing symbol can never earn another credit", out.Cleared, out.AwaitingAdjustment)
	}
	if len(out.AwaitingAdjustment) != 1 || out.AwaitingAdjustment[0].Symbol != "MSFT" {
		t.Fatalf("awaiting = %+v, want MSFT: a later comparison disagreed about it, so its "+
			"credit was refuted and it needs a new adjustment", out.AwaitingAdjustment)
	}
}
