//go:build tossos_testseams

package strategycoordinator

import (
	"sync"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/breakoutlane"
	"github.com/JungHoonGhae/tossinvest-cli/internal/continuationlane"
	"github.com/JungHoonGhae/tossinvest-cli/internal/reversallane"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyarbiter"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyrouter"
	"github.com/JungHoonGhae/tossinvest-cli/internal/weeklyvaluelane"
)

// spreadScores 는 breakout 이 이기는 점수표다. breakout 은 레인 사전순으로
// 첫째도 막내도 아닌 둘째라, "늘 0번을 고른다"와 "늘 마지막을 고른다"가
// 둘 다 이 시험에서 죽는다.
func spreadScores() map[strategyrouter.Family]uint32 {
	return map[strategyrouter.Family]uint32{
		strategyrouter.FamilyContinuation:   400_000,
		strategyrouter.FamilyReversal:       300_000,
		strategyrouter.FamilyWeeklyValue:    200_000,
		strategyrouter.FamilyBreakoutRetest: 900_000,
	}
}

func TestEachOwnerScopeInOneMarketGetsItsOwnSelection(t *testing.T) {
	f := newFixture(t, spreadScores())
	coordinator := NewMarketCoordinator(strategyrouter.MarketKR, f.now)
	f.submit(t, coordinator, fixtureSymbolFirst, allKRLanes()...)
	f.submit(t, coordinator, fixtureSymbolSecond, allKRLanes()...)

	outcome := coordinator.Arbitrate()
	if outcome.Closed() {
		t.Fatalf("closed: %+v", outcome)
	}
	if len(outcome.Selections) != 2 {
		t.Fatalf("selections=%d, want 2 — 시장 하나에 제안 하나라는 가정이 남아 있다", len(outcome.Selections))
	}
	if outcome.Selections[0].Scope.Symbol != fixtureSymbolFirst || outcome.Selections[1].Scope.Symbol != fixtureSymbolSecond {
		t.Fatalf("scope order=%q,%q, want %q,%q", outcome.Selections[0].Scope.Symbol,
			outcome.Selections[1].Scope.Symbol, fixtureSymbolFirst, fixtureSymbolSecond)
	}
	for _, selection := range outcome.Selections {
		if selection.Family != strategyrouter.FamilyBreakoutRetest || selection.ScorePPM != 900_000 {
			t.Fatalf("%s selected %q/%d, want BREAKOUT_RETEST/900000",
				selection.Scope.Symbol, selection.Family, selection.ScorePPM)
		}
		if selection.LineageIdentity == "" {
			t.Fatalf("%s selected without a lineage identity", selection.Scope.Symbol)
		}
	}
}

func TestSelectionOrderDoesNotDependOnSubmissionOrder(t *testing.T) {
	f := newFixture(t, spreadScores())
	coordinator := NewMarketCoordinator(strategyrouter.MarketKR, f.now)
	// 늦게 오는 종목을 먼저 넣는다. 도착 순서가 순서를 정한다면 여기서 뒤집힌다.
	f.submit(t, coordinator, fixtureSymbolSecond, allKRLanes()...)
	f.submit(t, coordinator, fixtureSymbolFirst, allKRLanes()...)

	outcome := coordinator.Arbitrate()
	if outcome.Closed() || len(outcome.Selections) != 2 {
		t.Fatalf("outcome=%+v", outcome)
	}
	if outcome.Selections[0].Scope.Symbol != fixtureSymbolFirst {
		t.Fatalf("first scope=%q, want %q — 도착 순서가 순서를 정하고 있다",
			outcome.Selections[0].Scope.Symbol, fixtureSymbolFirst)
	}
}

func TestTheSameDedupKeyArrivingAgainCoalescesAndCountsADrop(t *testing.T) {
	f := newFixture(t, spreadScores())
	coordinator := NewMarketCoordinator(strategyrouter.MarketKR, f.now)
	f.submit(t, coordinator, fixtureSymbolFirst, continuationlane.KRContinuationLaneID)
	before := coordinator.Depth()
	f.submit(t, coordinator, fixtureSymbolFirst, continuationlane.KRContinuationLaneID)

	if coordinator.Depth() != before {
		t.Fatalf("depth %d -> %d, want unchanged — 같은 열쇠가 칸을 두 개 차지했다", before, coordinator.Depth())
	}
	if coordinator.Drops() != 1 {
		t.Fatalf("drops=%d, want 1 — 접힌 봉투가 세어지지 않으면 조용한 유실이다", coordinator.Drops())
	}
	if outcome := coordinator.Arbitrate(); outcome.Closed() || len(outcome.Selections) != 1 || outcome.Drops != 1 {
		t.Fatalf("outcome=%+v", outcome)
	}
}

func TestTheSameLaneWithADifferentSnapshotClosesTheScope(t *testing.T) {
	f := newFixture(t, spreadScores())
	coordinator := NewMarketCoordinator(strategyrouter.MarketKR, f.now)
	authority := f.authority(t, fixtureSymbolFirst, continuationlane.KRContinuationLaneID)
	key := f.key(t, fixtureSymbolFirst)
	lane := continuationlane.KRContinuationLaneID
	first := coordinator.Submit(Envelope{Scope: key, SnapshotDigest: snapshotDigest(fixtureSymbolFirst, lane),
		Proposal: strategyarbiter.Proposal{Result: f.result(t, fixtureSymbolFirst, lane), Authority: authority}})
	if !first.Admitted {
		t.Fatalf("first submission refused: %+v", first)
	}
	// 같은 레인, 다른 스냅샷. 어느 쪽이 새것인지 정할 봉인된 근거가 없으므로
	// 조용히 덮어쓰지 않고 닫는다.
	second := coordinator.Submit(Envelope{Scope: key, SnapshotDigest: "sha256:snapshot-newer",
		Proposal: strategyarbiter.Proposal{
			Result:    f.resultValidUntil(t, fixtureSymbolFirst, lane, f.now.Add(2*time.Minute)),
			Authority: authority}})
	if second.Admitted {
		t.Fatal("두 번째 스냅샷이 조용히 들어갔다")
	}
	if second.Refusal != strategyarbiter.RefusalSealMismatch || second.Detail != DetailSnapshotConflict {
		t.Fatalf("refusal=%q detail=%q, want %q/%q", second.Refusal, second.Detail,
			strategyarbiter.RefusalSealMismatch, DetailSnapshotConflict)
	}
	outcome := coordinator.Arbitrate()
	if !outcome.Closed() || outcome.Refusal != strategyarbiter.RefusalSealMismatch {
		t.Fatalf("outcome=%+v, want a closed market — Submit 반환값을 버린 호출자가 그냥 지나갔다", outcome)
	}
}

func TestOverflowClosesTheMarketAndEvictsNothing(t *testing.T) {
	f := newFixture(t, spreadScores())
	coordinator := newMarketCoordinator(strategyrouter.MarketKR, f.now, 2)
	f.submit(t, coordinator, fixtureSymbolFirst, continuationlane.KRContinuationLaneID, reversallane.KRReversalLaneID)
	if coordinator.Depth() != 2 {
		t.Fatalf("depth=%d, want 2", coordinator.Depth())
	}
	admissions := f.submit(t, coordinator, fixtureSymbolSecond, breakoutlane.KRLaneID)
	if admissions[0].Admitted || !admissions[0].Overflow {
		t.Fatalf("admission=%+v, want an overflow refusal", admissions[0])
	}
	if coordinator.Depth() != 2 {
		t.Fatalf("depth=%d, want 2 — 자리를 만들려고 무언가를 쫓아냈다", coordinator.Depth())
	}
	outcome := coordinator.Arbitrate()
	if !outcome.Overflow || len(outcome.Selections) != 0 {
		t.Fatalf("outcome=%+v, want an overflow-closed market", outcome)
	}
	if outcome.Refusal != strategyarbiter.RefusalNone {
		t.Fatalf("overflow borrowed the arbitration code %q — 큐 문제를 봉인 문제로 보고하면 안 된다", outcome.Refusal)
	}
}

func TestTheProductionCoordinatorUsesTheServerOwnedCapacity(t *testing.T) {
	f := newFixture(t, spreadScores())
	if got := NewMarketCoordinator(strategyrouter.MarketKR, f.now).Capacity(); got != Capacity {
		t.Fatalf("capacity=%d, want %d", got, Capacity)
	}
	// 골든의 queue.capacity 는 "server-owned positive finite" 다. 그 문장이
	// 요구하는 것은 여기서 보고, 그 수가 어느 영수증에서 왔는지는
	// receipt_contract_test.go 가 본다.
	if Capacity <= 0 {
		t.Fatalf("capacity=%d must be positive and finite", Capacity)
	}
}

func TestASubmissionFromAnotherMarketIsRefused(t *testing.T) {
	f := newFixture(t, spreadScores())
	coordinator := NewMarketCoordinator(strategyrouter.MarketUS, f.now)
	admissions := f.submit(t, coordinator, fixtureSymbolFirst, continuationlane.KRContinuationLaneID)
	if admissions[0].Admitted {
		t.Fatal("KR 제안이 US 조정자에 들어갔다")
	}
	if admissions[0].Refusal != strategyarbiter.RefusalSealMismatch || admissions[0].Detail != DetailMarketScope {
		t.Fatalf("admission=%+v", admissions[0])
	}
}

func TestOneRefusedScopeClosesTheWholeMarket(t *testing.T) {
	// 두 가족이 같은 점수를 낸다. 동점은 fail-closed 다.
	f := newFixture(t, map[strategyrouter.Family]uint32{
		strategyrouter.FamilyContinuation:   500_000,
		strategyrouter.FamilyReversal:       500_000,
		strategyrouter.FamilyWeeklyValue:    100_000,
		strategyrouter.FamilyBreakoutRetest: 100_000,
	})
	coordinator := NewMarketCoordinator(strategyrouter.MarketKR, f.now)
	f.submit(t, coordinator, fixtureSymbolFirst, weeklyvaluelane.KRWeeklyLaneID)
	f.submit(t, coordinator, fixtureSymbolSecond, continuationlane.KRContinuationLaneID, reversallane.KRReversalLaneID)

	outcome := coordinator.Arbitrate()
	if outcome.Refusal != strategyarbiter.RefusalTie {
		t.Fatalf("refusal=%q, want %q", outcome.Refusal, strategyarbiter.RefusalTie)
	}
	if len(outcome.Selections) != 0 {
		t.Fatalf("selections=%d, want 0 — 한 범위가 닫혔는데 다른 범위가 풀려 나갔다", len(outcome.Selections))
	}
	if outcome.Scope.Symbol != fixtureSymbolSecond {
		t.Fatalf("closed scope=%q, want %q", outcome.Scope.Symbol, fixtureSymbolSecond)
	}
}

// 채점표에 없는 레인은 큐에 들어가되 중재자가 거절한다. 거절 이유를 두 곳에서
// 판정하지 않는다 — 판정하는 곳이 둘이면 같은 입력에 두 진단이 나온다.
func TestALaneOutsideTheApprovedScoreTableIsRefusedByTheArbiterNotByIntake(t *testing.T) {
	// weekly 만 채점표에 있다. breakout 은 어느 가족인지 알 수 없다.
	f := newFixture(t, map[strategyrouter.Family]uint32{strategyrouter.FamilyWeeklyValue: 100_000})
	coordinator := NewMarketCoordinator(strategyrouter.MarketKR, f.now)
	admissions := f.submit(t, coordinator, fixtureSymbolFirst, breakoutlane.KRLaneID)
	if !admissions[0].Admitted || admissions[0].Key.Family != "" {
		t.Fatalf("admission=%+v, want an admitted envelope with no family name", admissions[0])
	}
	outcome := coordinator.Arbitrate()
	if outcome.Refusal != strategyarbiter.RefusalUncalibrated || outcome.Detail != strategyarbiter.DetailUnknownFamily {
		t.Fatalf("outcome=%+v, want the arbiter's own uncalibrated code", outcome)
	}
}

func TestABrokenProposalSealIsRefusedBeforeItBecomesAKey(t *testing.T) {
	f := newFixture(t, spreadScores())
	coordinator := NewMarketCoordinator(strategyrouter.MarketKR, f.now)
	result := f.result(t, fixtureSymbolFirst, continuationlane.KRContinuationLaneID)
	result.Quantity++ // 봉인 뒤 한 글자를 바꾼다.
	admission := coordinator.Submit(Envelope{Scope: f.key(t, fixtureSymbolFirst),
		SnapshotDigest: snapshotDigest(fixtureSymbolFirst, continuationlane.KRContinuationLaneID),
		Proposal: strategyarbiter.Proposal{Result: result,
			Authority: f.authority(t, fixtureSymbolFirst, continuationlane.KRContinuationLaneID)}})
	if admission.Admitted {
		t.Fatal("봉인이 깨진 제안으로 열쇠를 만들었다")
	}
	if admission.Refusal != strategyarbiter.RefusalSealMismatch || admission.Detail != strategyarbiter.DetailProposalSeal {
		t.Fatalf("admission=%+v", admission)
	}
}

func TestConcurrentSubmitsFromEveryLaneAreSafe(t *testing.T) {
	f := newFixture(t, spreadScores())
	coordinator := NewMarketCoordinator(strategyrouter.MarketKR, f.now)
	// 재료는 시험 goroutine 에서 다 만들어 둔다. Submit 만 동시에 부른다.
	jobs := make([]Envelope, 0, 8)
	for _, symbol := range []string{fixtureSymbolFirst, fixtureSymbolSecond} {
		authority := f.authority(t, symbol, allKRLanes()...)
		for _, laneID := range allKRLanes() {
			jobs = append(jobs, Envelope{Scope: f.key(t, symbol), SnapshotDigest: snapshotDigest(symbol, laneID),
				Proposal: strategyarbiter.Proposal{Result: f.result(t, symbol, laneID), Authority: authority}})
		}
	}
	var wait sync.WaitGroup
	for _, job := range jobs {
		wait.Add(1)
		go func(job Envelope) {
			defer wait.Done()
			coordinator.Submit(job)
		}(job)
	}
	wait.Wait()
	outcome := coordinator.Arbitrate()
	if outcome.Closed() || len(outcome.Selections) != 2 {
		t.Fatalf("outcome=%+v", outcome)
	}
}

// Submit 과 Arbitrate 를 실제로 겹쳐 돌린다.
//
// TestConcurrentSubmitsFromEveryLaneAreSafe 는 이름과 달리 이 겹침을 만들지
// 않는다. 그 시험은 wait.Wait() 로 모든 Submit 이 끝난 **뒤에** Arbitrate 를
// 부르므로 둘이 동시에 도는 순간이 아예 없다. 여기서는 동시에 돌린다.
//
// **이 시험이 증명하지 못하는 것을 먼저 적는다.** Arbitrate 가 잠금을 일찍
// 놓는 판본(review.md 의 M-F)을 이 시험은 **죽이지 못했다 — 실제로 돌려 보고
// 살아남는 것을 확인했다.** 이유는 시험이 약해서가 아니라 그 결함이 밖에서
// 관측 불가능하기 때문이다: 올바른 판본에서도 Arbitrate 가 경쟁에서 이기면
// 닫히지 않은 결과가 정당하게 나오므로, "닫히지 않았다"는 관측만으로는 두
// 판본을 가를 수 없다. 가르려면 임계 구역 안에 주입 지점이 있어야 한다.
// 그 확정적 증명은 8 worker/2 coordinator 하네스를 세우는 태스크 5.7 의 몫이다.
//
// 이 시험이 지금 지키는 것은 둘이다. race detector 가 볼 실제 자료 경쟁이
// 없다는 것, 그리고 결과가 언제나 두 선형화 가능한 답(닫힘이면 선택 0 개,
// 아니면 두 범위를 다 고른 온전한 목록) 중 하나이지 "닫히지도 않았는데
// 목록만 짧은" 셋째 답이 아니라는 것이다.
func TestAMarketClosedWhileArbitrateIsRunningStillReportsClosed(t *testing.T) {
	for round := range 200 {
		f := newFixture(t, spreadScores())
		coordinator := NewMarketCoordinator(strategyrouter.MarketKR, f.now)
		for _, symbol := range []string{fixtureSymbolFirst, fixtureSymbolSecond} {
			authority := f.authority(t, symbol, allKRLanes()...)
			for _, laneID := range allKRLanes() {
				coordinator.Submit(Envelope{Scope: f.key(t, symbol), SnapshotDigest: snapshotDigest(symbol, laneID),
					Proposal: strategyarbiter.Proposal{Result: f.result(t, symbol, laneID), Authority: authority}})
			}
		}
		// 다른 시장에서 온 봉투다. Submit 의 첫 검사에서 바로 결함이 되므로
		// 해시도 봉인 검사도 거치지 않고 조정자를 닫는다.
		otherMarket, err := strategyrouter.NewOwnerKey(fixtureAccount, strategyrouter.MarketUS, fixtureSymbolFirst, 1)
		if err != nil {
			t.Fatal(err)
		}
		intruder := Envelope{Scope: otherMarket,
			SnapshotDigest: snapshotDigest(fixtureSymbolFirst, continuationlane.KRContinuationLaneID),
			Proposal: strategyarbiter.Proposal{Result: f.result(t, fixtureSymbolFirst, continuationlane.KRContinuationLaneID),
				Authority: f.authority(t, fixtureSymbolFirst, continuationlane.KRContinuationLaneID)}}

		var wait sync.WaitGroup
		wait.Add(1)
		var outcome Outcome
		go func() {
			defer wait.Done()
			outcome = coordinator.Arbitrate()
		}()
		coordinator.Submit(intruder)
		wait.Wait()

		// 선형화 가능한 답은 둘뿐이다. 닫혔거나, 아니면 두 범위를 다 고른
		// 온전한 목록이거나. "닫히지 않았는데 목록이 짧다"는 셋째 답은
		// 조용한 유실이며 아래 파이프라인의 "시장에 하나" 관문을 오히려 만족시킨다.
		if outcome.Closed() {
			if len(outcome.Selections) != 0 {
				t.Fatalf("round=%d 닫혔는데 선택 %d 개가 새어 나왔다", round, len(outcome.Selections))
			}
			continue
		}
		if len(outcome.Selections) != 2 {
			t.Fatalf("round=%d outcome=%+v — 닫히지도 않고 목록만 짧아졌다", round, outcome)
		}
	}
}

func TestAnEnvelopeWithoutASnapshotDigestIsRefused(t *testing.T) {
	f := newFixture(t, spreadScores())
	coordinator := NewMarketCoordinator(strategyrouter.MarketKR, f.now)
	lane := continuationlane.KRContinuationLaneID
	admission := coordinator.Submit(Envelope{Scope: f.key(t, fixtureSymbolFirst),
		Proposal: strategyarbiter.Proposal{Result: f.result(t, fixtureSymbolFirst, lane),
			Authority: f.authority(t, fixtureSymbolFirst, lane)}})
	if admission.Admitted {
		t.Fatal("스냅샷 다이제스트 없이 열쇠가 만들어졌다 — 모든 스냅샷이 한 칸으로 뭉친다")
	}
	if admission.Refusal != strategyarbiter.RefusalSealMismatch || admission.Detail != DetailNoSnapshot {
		t.Fatalf("admission=%+v", admission)
	}
}

// 다이제스트만 열쇠에 넣으면, 잘못 적힌 다이제스트 하나가 서로 다른 두 제안을
// 한 칸에 겹쳐 놓고 하나를 조용히 지운다. 봉인된 계보 신원까지 봐야 한다.
func TestTwoDifferentProposalsSharingOneSnapshotDigestFailClosed(t *testing.T) {
	f := newFixture(t, spreadScores())
	coordinator := NewMarketCoordinator(strategyrouter.MarketKR, f.now)
	lane := continuationlane.KRContinuationLaneID
	key := f.key(t, fixtureSymbolFirst)
	authority := f.authority(t, fixtureSymbolFirst, lane)
	same := "sha256:snapshot-one-value-for-both"
	first := coordinator.Submit(Envelope{Scope: key, SnapshotDigest: same,
		Proposal: strategyarbiter.Proposal{Result: f.result(t, fixtureSymbolFirst, lane), Authority: authority}})
	if !first.Admitted {
		t.Fatalf("first=%+v", first)
	}
	second := coordinator.Submit(Envelope{Scope: key, SnapshotDigest: same,
		Proposal: strategyarbiter.Proposal{
			Result:    f.resultValidUntil(t, fixtureSymbolFirst, lane, f.now.Add(3*time.Minute)),
			Authority: authority}})
	if second.Admitted {
		t.Fatal("다른 제안이 같은 다이제스트를 달고 조용히 덮어썼다")
	}
	if second.Detail != DetailSnapshotConflict {
		t.Fatalf("second=%+v", second)
	}
	if coordinator.Depth() != 1 {
		t.Fatalf("depth=%d, want 1", coordinator.Depth())
	}
}
