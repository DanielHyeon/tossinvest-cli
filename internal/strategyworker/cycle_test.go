//go:build tossos_testseams

package strategyworker

import (
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/continuationlane"
	"github.com/JungHoonGhae/tossinvest-cli/internal/reversallane"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyarbiter"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategycoordinator"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyrouter"
)

// 생산 목록의 여덟은 전부 잠들어 있고, **완전히 정상인 입력에도** 봉투를 내지
// 않는다.
//
// 입력을 일부러 흠집 내지 않는 것이 요점이다. 흠집 난 입력으로 시험하면 거절이
// 잠들었기 때문인지 흠집 때문인지 구별할 수 없고, 그러면 OFF 기본값은 확인되지
// 않은 채 초록으로 남는다.
func TestEveryProductionWorkerIsBornDormantAndEmitsNothing(t *testing.T) {
	f := newFixture(allKRFamilies()...)
	good := f.input(t, continuationlane.KRContinuationLaneID)

	// 같은 입력이 켜진 worker 에서는 실제로 봉투가 된다. 이 대조가 없으면
	// 아래 여덟 개의 DORMANT 는 "입력이 나빴다"로도 설명된다.
	if emitted := effective(t, strategyrouter.MarketKR, strategyrouter.FamilyContinuation).Run(good); emitted.Outcome != OutcomeEmitted {
		t.Fatalf("the control input is not emittable (%s/%s), so the dormant assertions below prove nothing",
			emitted.Outcome, emitted.Detail)
	}

	workers := ProductionWorkers()
	if len(workers) == 0 {
		t.Fatal("no production worker was scanned")
	}
	for index, worker := range workers {
		cycle := worker.Run(good)
		if cycle.Outcome != OutcomeDormant {
			t.Errorf("worker %d (%v) is not dormant: %s", index, worker.Key().Parts(), cycle.Outcome)
		}
		if cycle.Detail != DetailDormant {
			t.Errorf("worker %d detail: want %q, got %q", index, DetailDormant, cycle.Detail)
		}
		if !emptyEnvelope(cycle.Envelope) {
			t.Errorf("worker %d carried an envelope while dormant", index)
		}
		// 잠든 것은 거절이 아니다. 거절 코드를 함께 내면 운영자가 원인을
		// 중재 실패로 읽는다.
		if cycle.Refusal != strategyarbiter.RefusalNone {
			t.Errorf("worker %d attached a refusal code to a dormant cycle: %q", index, cycle.Refusal)
		}
	}
}

// 켜진 worker 는 받은 값을 그대로 담은 조정자 봉투를 낸다.
//
// 필드 세 개를 각각 확인하고, 그 봉투를 실제 조정자에 넣어 본다. 봉투 모양만
// 보면 "조정자가 받아 준다"는 것은 확인되지 않는다.
func TestAnEffectiveWorkerEmitsAnEnvelopeTheCoordinatorAdmitsForThisLane(t *testing.T) {
	f := newFixture(allKRFamilies()...)
	input := f.input(t, continuationlane.KRContinuationLaneID)
	worker := effective(t, strategyrouter.MarketKR, strategyrouter.FamilyContinuation)

	cycle := worker.Run(input)
	if cycle.Outcome != OutcomeEmitted {
		t.Fatalf("want EMITTED, got %s (%s)", cycle.Outcome, cycle.Detail)
	}
	if cycle.Refusal != strategyarbiter.RefusalNone || cycle.Detail != "" {
		t.Errorf("an emitted cycle carries no refusal: %q / %q", cycle.Refusal, cycle.Detail)
	}
	if cycle.Envelope.Scope != input.Scope {
		t.Errorf("scope drifted: want %v, got %v", input.Scope, cycle.Envelope.Scope)
	}
	if cycle.Envelope.SnapshotDigest != input.SnapshotDigest {
		t.Errorf("snapshot digest drifted: want %q, got %q", input.SnapshotDigest, cycle.Envelope.SnapshotDigest)
	}
	if cycle.Envelope.Proposal.Result.Lineage.Identity != input.Proposal.Result.Lineage.Identity {
		t.Errorf("the emitted proposal is not the one that came in")
	}

	coordinator := strategycoordinator.NewMarketCoordinator(strategyrouter.MarketKR, f.now)
	admission := coordinator.Submit(cycle.Envelope)
	if !admission.Admitted {
		t.Fatalf("the coordinator refused the worker's envelope: %s / %s", admission.Refusal, admission.Detail)
	}
	// 조정자가 만든 열쇠는 이 worker 의 레인을 가리켜야 한다. 봉투가 들어가긴
	// 했는데 다른 레인으로 기록되면 여덟 worker 의 귀속이 무너진다.
	if admission.Key.Family != worker.Key().Family || admission.Key.LaneID != worker.Key().LaneID ||
		admission.Key.LaneVersion != worker.Key().LaneVersion {
		t.Errorf("the coordinator attributed the envelope elsewhere: worker %v, coordinator %v/%v/%v",
			worker.Key().Parts(), admission.Key.Family, admission.Key.LaneID, admission.Key.LaneVersion)
	}
}

// 남의 레인 제안은 거절한다.
//
// 여기서 유일하게 다른 것은 레인이다. 계정·종목·세대·스냅샷·권한은 모두 같으므로,
// 거절이 났다면 그 이유는 레인 하나뿐이다.
func TestAWorkerRefusesAProposalFromAnotherLane(t *testing.T) {
	f := newFixture(allKRFamilies()...)
	foreign := f.input(t, reversallane.KRReversalLaneID)
	worker := effective(t, strategyrouter.MarketKR, strategyrouter.FamilyContinuation)

	cycle := worker.Run(foreign)
	if cycle.Outcome != OutcomeRefused {
		t.Fatalf("want REFUSED, got %s", cycle.Outcome)
	}
	if cycle.Refusal != strategyarbiter.RefusalSealMismatch {
		t.Errorf("want %q, got %q", strategyarbiter.RefusalSealMismatch, cycle.Refusal)
	}
	if cycle.Detail != DetailNotThisLane {
		t.Errorf("want %q, got %q", DetailNotThisLane, cycle.Detail)
	}
	if !emptyEnvelope(cycle.Envelope) {
		t.Error("a refused cycle carried an envelope")
	}

	// 그 제안의 주인은 거절하지 않는다 — 거절이 레인 때문임을 이 줄이 못 박는다.
	if owner := effective(t, strategyrouter.MarketKR, strategyrouter.FamilyReversal).Run(foreign); owner.Outcome != OutcomeEmitted {
		t.Fatalf("the reversal worker refused its own lane's proposal: %s / %s", owner.Outcome, owner.Detail)
	}
}

// 다른 시장의 worker 도 거절한다.
//
// 레인 ID 는 시장마다 다르므로 레인 검사만으로도 걸리지만, 시장 비교가 사라져도
// 걸리는지는 이 시험이 따로 본다 — 검사 하나가 다른 검사에 업혀 통과하고 있으면
// 앞의 것을 지워도 초록이다.
func TestAWorkerOfTheOtherMarketRefuses(t *testing.T) {
	f := newFixture(allKRFamilies()...)
	krInput := f.input(t, continuationlane.KRContinuationLaneID)

	cycle := effective(t, strategyrouter.MarketUS, strategyrouter.FamilyContinuation).Run(krInput)
	if cycle.Outcome != OutcomeRefused {
		t.Fatalf("the US continuation worker accepted a KR proposal: %s", cycle.Outcome)
	}
}

// 봉인이 깨진 제안은 거절한다.
//
// 계보는 봉인이 성립할 때만 값이다. 봉인을 안 보고 계보의 레인 이름만 읽으면
// 위조된 계보가 "네 레인이야"라고 말하는 것을 그대로 믿게 된다.
func TestAWorkerRefusesAProposalWhoseSealNoLongerHolds(t *testing.T) {
	f := newFixture(allKRFamilies()...)
	input := f.input(t, continuationlane.KRContinuationLaneID)
	worker := effective(t, strategyrouter.MarketKR, strategyrouter.FamilyContinuation)

	if before := worker.Run(input); before.Outcome != OutcomeEmitted {
		t.Fatalf("the untouched input must be emittable first: %s / %s", before.Outcome, before.Detail)
	}

	// 계보는 그대로 두고 수량만 바꾼다. 레인 이름은 여전히 이 worker 의 것이므로
	// 봉인을 보지 않는 구현은 이것을 통과시킨다.
	tampered := input
	tampered.Proposal.Result.Quantity++
	if tampered.Proposal.Result.Lineage.LaneID != worker.Key().LaneID {
		t.Fatal("the tampering changed the lane, so it does not test the seal")
	}

	cycle := worker.Run(tampered)
	if cycle.Outcome != OutcomeRefused {
		t.Fatalf("want REFUSED for a broken seal, got %s", cycle.Outcome)
	}
	if cycle.Refusal != strategyarbiter.RefusalSealMismatch {
		t.Errorf("want %q, got %q", strategyarbiter.RefusalSealMismatch, cycle.Refusal)
	}
}

// 가족을 유도할 수 없으면 거절한다.
//
// 가족은 제안이 스스로 신고하는 값이 아니라 봉인된 권한의 가족 점수 행에서
// 유도하는 값이다. 그 행을 뺀 권한을 주면, 계보의 레인은 그대로인데 가족만
// 세울 수 없는 상태가 된다 — 신고를 믿는 구현은 이것도 통과시킨다.
func TestAWorkerRefusesWhenTheFamilyCannotBeDerivedFromTheSealedAuthority(t *testing.T) {
	worker := effective(t, strategyrouter.MarketKR, strategyrouter.FamilyContinuation)

	// 대조: 가족 행이 있으면 같은 레인이 봉투가 된다.
	withRow := newFixture(allKRFamilies()...).input(t, continuationlane.KRContinuationLaneID)
	if cycle := worker.Run(withRow); cycle.Outcome != OutcomeEmitted {
		t.Fatalf("the control input is not emittable: %s / %s", cycle.Outcome, cycle.Detail)
	}

	// continuation 만 빼고 나머지 셋만 점수 행에 담는다.
	withoutRow := newFixture(strategyrouter.FamilyReversal, strategyrouter.FamilyWeeklyValue,
		strategyrouter.FamilyBreakoutRetest).input(t, continuationlane.KRContinuationLaneID)
	if withoutRow.Proposal.Result.Lineage.LaneID != worker.Key().LaneID {
		t.Fatal("the fixture changed the lane, so it does not test the family derivation")
	}
	if family := strategyarbiter.ProposalFamily(withoutRow.Proposal); family != "" {
		t.Fatalf("the fixture still derives a family (%q), so it does not test the derivation", family)
	}

	cycle := worker.Run(withoutRow)
	if cycle.Outcome != OutcomeRefused {
		t.Fatalf("want REFUSED when the family cannot be derived, got %s", cycle.Outcome)
	}
	if cycle.Detail != DetailNotThisLane {
		t.Errorf("want %q, got %q", DetailNotThisLane, cycle.Detail)
	}
}

// emptyEnvelope 는 봉투가 비었는지 본다.
//
// `Envelope` 은 제안을 담고 있어 비교 가능한 타입이 아니다. 그래서 채워졌다면
// 반드시 값이 있어야 하는 세 자리를 각각 본다 — 하나만 보면 나머지 둘이 채워진
// 봉투가 "비었다"로 통과한다.
func emptyEnvelope(envelope strategycoordinator.Envelope) bool {
	return envelope.Scope == (strategyrouter.OwnerKey{}) && envelope.SnapshotDigest == "" &&
		envelope.Proposal.Result.Lineage.Identity == ""
}

// 잠긴 레인은 **켜져 있어도** 봉투를 내지 않는다.
//
// lane_test.go 의 같은 이름 시험은 잠들어 있는 레인을 쓴다. 그것만으로는
// 잠금이 실제로 배출을 막는지 알 수 없다 — 잠든 worker 는 어차피 아무것도
// 안 내기 때문이다. 여기서는 같은 입력에 봉투를 내던 레인을 잠그고, 그 뒤로
// 아무것도 안 나오는지 본다.
func TestALatchedLaneEmitsNothingEvenWhenItIsEffective(t *testing.T) {
	f := newFixture(allKRFamilies()...)
	good := f.input(t, continuationlane.KRContinuationLaneID)
	lane := newLane(effective(t, strategyrouter.MarketKR, strategyrouter.FamilyContinuation),
		ProductionRuntimePolicy(), clock.NewFake(f.now))

	if before := lane.Run(good); before.Outcome != OutcomeEmitted {
		t.Fatalf("the lane must emit before it is latched, got %s / %s", before.Outcome, before.Detail)
	}
	if _, latched := lane.Fail("evidence refresh failed", false); !latched {
		t.Fatal("the lane did not latch")
	}

	cycle := lane.Run(good)
	if cycle.Outcome != OutcomeLatched {
		t.Fatalf("a latched effective lane reported %s", cycle.Outcome)
	}
	if !emptyEnvelope(cycle.Envelope) {
		t.Error("a latched effective lane still handed out an envelope")
	}
}
