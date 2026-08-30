//go:build tossos_testseams

package strategycoordinator

import (
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/continuationlane"
	"github.com/JungHoonGhae/tossinvest-cli/internal/reversallane"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyarbiter"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyrouter"
)

// 같은 계보가 **스냅샷만 달리해서** 다시 오면 조용히 덮어쓰지 않고 닫는다.
//
// 이 시험이 따로 필요한 이유는 축 하나만 흔들기 위해서다.
// TestTheSameLaneWithADifferentSnapshotClosesTheScope 는 두 번째 봉투에서
// 스냅샷 다이제스트와 후보 유효기한을 **둘 다** 바꾼다. 유효기한은 봉인된
// 계보 신원에 들어가므로 그 시험은 `existing.identity != lineage.Identity`
// 하나만으로도 통과한다 — 즉 `existing.key != key` 쪽 절반은 한 번도
// 시험되지 않는다. 실제로 그 절반을 지우면 그 시험은 그대로 통과했다.
//
// 여기서는 Result 를 **글자 그대로 같은 값**으로 두고 Envelope.SnapshotDigest
// 하나만 바꾼다. lineageIdentity 는 스냅샷 다이제스트를 재료로 쓰지 않으므로
// 두 봉투의 신원은 같고, 다른 것은 열쇠뿐이다. 이 입력은 지어낸 것이 아니라
// 같은 제안을 다른 증거 스냅샷에 재생하면 그대로 나오는 모양이다.
func TestTheSameLineageWithOnlyADifferentSnapshotDigestClosesTheScope(t *testing.T) {
	f := newFixture(t, spreadScores())
	coordinator := NewMarketCoordinator(strategyrouter.MarketKR, f.now)
	lane := continuationlane.KRContinuationLaneID
	// 두 봉투가 같은 Result 를 공유한다. 계보 신원은 반드시 같다.
	result := f.result(t, fixtureSymbolFirst, lane)
	proposal := strategyarbiter.Proposal{Result: result, Authority: f.authority(t, fixtureSymbolFirst, lane)}

	first := coordinator.Submit(Envelope{Scope: f.key(t, fixtureSymbolFirst),
		SnapshotDigest: snapshotDigest(fixtureSymbolFirst, lane), Proposal: proposal})
	if !first.Admitted {
		t.Fatalf("첫 봉투가 안 들어갔다: %+v", first)
	}
	second := coordinator.Submit(Envelope{Scope: f.key(t, fixtureSymbolFirst),
		SnapshotDigest: snapshotDigest(fixtureSymbolFirst, lane) + "-replayed", Proposal: proposal})

	if second.Admitted {
		t.Fatal("스냅샷만 다른 봉투가 같은 칸에 조용히 덮어썼다 — 하나가 소리 없이 사라진다")
	}
	if second.Refusal != strategyarbiter.RefusalSealMismatch || second.Detail != DetailSnapshotConflict {
		t.Fatalf("admission=%+v, want %q/%q", second, strategyarbiter.RefusalSealMismatch, DetailSnapshotConflict)
	}
	if outcome := coordinator.Arbitrate(); !outcome.Closed() {
		t.Fatalf("outcome=%+v — 칸을 두고 다툰 뒤에도 시장이 열려 있다", outcome)
	}
}

// 들어간 봉투의 열쇠에는 가족이 실제로 들어 있다.
//
// 골든이 family 를 dedup_key_fields 여덟 중 하나로 못 박고, ProposalFamily 를
// arbiter 밖으로 내보낸 이유도 그것뿐이다. 그런데 지금까지 그 필드가 채워지는지
// 보는 시험이 하나도 없었다 — TestEveryDedupFieldChangesTheKey 는 Key 를 손으로
// 지어서 쓰므로 Submit 을 거치지 않고, 다른 시험들은 자격 밖 레인이 빈 값을
// 남기는 **음의** 경우만 본다. 그래서 ProposalFamily 가 언제나 "" 를 돌려주게
// 만들어도 두 스위트가 다 통과했다.
func TestAnAdmittedEnvelopeCarriesItsFamilyInTheDedupKey(t *testing.T) {
	f := newFixture(t, spreadScores())
	coordinator := NewMarketCoordinator(strategyrouter.MarketKR, f.now)
	admissions := f.submit(t, coordinator, fixtureSymbolFirst, allKRLanes()...)
	want := map[string]strategyrouter.Family{
		continuationlane.KRContinuationLaneID: strategyrouter.FamilyContinuation,
	}
	for index, laneID := range allKRLanes() {
		if !admissions[index].Admitted {
			t.Fatalf("lane=%s 가 안 들어갔다: %+v", laneID, admissions[index])
		}
		if admissions[index].Key.Family == "" {
			t.Fatalf("lane=%s 의 열쇠에 가족이 비어 있다 — 여덟 필드 중 하나가 늘 같은 값이면 접힘이 잘못된 것을 묶는다", laneID)
		}
		if expected, ok := want[laneID]; ok && admissions[index].Key.Family != expected {
			t.Fatalf("lane=%s family=%q, want %q", laneID, admissions[index].Key.Family, expected)
		}
	}
}

// Parts 는 DedupKeyFields 와 **순서까지** 같다.
//
// 앞선 시험은 길이만 봤다. 길이만 보면 값 순서를 아무렇게나 섞어도 통과하고,
// 그러면 이 열쇠를 이름과 짝지어 기록하는 쪽이 값을 엉뚱한 이름에 붙인다.
func TestPartsMatchesTheGoldenFieldOrderPositionally(t *testing.T) {
	key := Key{
		Scope: strategyrouter.OwnerKey{AccountRef: "acct-값", Market: strategyrouter.MarketKR,
			Symbol: "심볼-값", PositionGeneration: 7},
		Family: "가족-값", LaneID: "레인-값", LaneVersion: "판본-값", SnapshotDigest: "스냅샷-값",
	}
	want := []any{"acct-값", strategyrouter.MarketKR, "심볼-값", uint64(7),
		strategyrouter.Family("가족-값"), "레인-값", "판본-값", "스냅샷-값"}
	parts := key.Parts()
	if len(parts) != len(want) {
		t.Fatalf("parts=%d, want %d", len(parts), len(want))
	}
	for index, name := range DedupKeyFields() {
		if parts[index] != want[index] {
			t.Fatalf("%d 번째(%s) = %v, want %v — 값이 이름과 어긋났다", index, name, parts[index], want[index])
		}
	}
}

// scopeOrderLess 는 앞 세 값이 같으면 세대로 가른다.
//
// 지금 배선에서는 한 시장 안에서 종목이 유일하므로 이 갈래에 닿지 않는다.
// 그래도 이 함수는 OwnerKey 전체의 전순서라고 선언되어 있고, 선언한 것은
// 시험한다 — 마지막 비교를 false 로 바꿔도 아무 시험이 안 깨지던 자리다.
func TestScopeOrderFallsThroughToPositionGeneration(t *testing.T) {
	older := strategyrouter.OwnerKey{AccountRef: "acct", Market: strategyrouter.MarketKR, Symbol: "000660", PositionGeneration: 1}
	newer := strategyrouter.OwnerKey{AccountRef: "acct", Market: strategyrouter.MarketKR, Symbol: "000660", PositionGeneration: 2}
	if !scopeOrderLess(older, newer) {
		t.Fatal("세대 1 이 세대 2 보다 앞서지 않는다")
	}
	if scopeOrderLess(newer, older) {
		t.Fatal("세대 2 가 세대 1 보다 앞선다 — 전순서가 아니다")
	}
	if scopeOrderLess(older, older) {
		t.Fatal("같은 열쇠가 자기보다 앞선다")
	}
}

// 한 범위 안에 결함 있는 제안이 둘이면, 운영자가 보는 진단은 언제나 같아야 한다.
//
// 중재자는 제안 목록을 걸으며 **처음 만난** 결함에서 돌아선다(arbiter.go 5단계).
// 그래서 범위 안 순서가 흔들리면 같은 입력이 주기마다 다른 이유를 보고한다.
// 어느 것이 이겼는지는 안 바뀌므로(점수 비교는 순서와 무관하고 동점은 닫힌다)
// 이 순서가 지키는 것은 선택이 아니라 **진단의 재현성**이고, 골든의
// queue.ordering 이 "deterministic" 이라고 적은 것이 바로 그것이다.
//
// 정렬을 지우면 grouped 는 Go 의 무작위 맵 순회 순서를 그대로 물려받는다.
// 그래서 한 판이 아니라 여러 판을 돌린다 — 한 판만 보면 무작위가 우연히
// 같은 답을 내고 지나간다.
func TestTwoFaultyLanesInOneScopeAlwaysReportTheSameRefusal(t *testing.T) {
	f := newFixture(t, spreadScores())
	// 지속형 레인만 자격 집합에 넣되 증거 다이제스트를 어긋나게 만든다.
	// 그러면 지속형은 "증거 결속" 으로, 자격 집합에 없는 반전형은
	// "자격 없는 레인" 으로 각각 다른 이유에 걸린다.
	corrupted := f.candidate(t, fixtureSymbolFirst, continuationlane.KRContinuationLaneID)
	corrupted.EvidenceDigest = "sha256:이-증거로-만든-제안이-아니다"
	request, err := strategyrouter.MultiCandidateRouteFixture(f.key(t, fixtureSymbolFirst), f.now, corrupted)
	if err != nil {
		t.Fatal(err)
	}
	authority := strategyrouter.ProductionRouteAuthorityFromRequestForTest(request, fixtureCalibration, f.scores)

	const rounds = 64
	for round := range rounds {
		coordinator := NewMarketCoordinator(strategyrouter.MarketKR, f.now)
		for _, laneID := range []string{continuationlane.KRContinuationLaneID, reversallane.KRReversalLaneID} {
			admission := coordinator.Submit(Envelope{Scope: f.key(t, fixtureSymbolFirst),
				SnapshotDigest: snapshotDigest(fixtureSymbolFirst, laneID),
				Proposal:       strategyarbiter.Proposal{Result: f.result(t, fixtureSymbolFirst, laneID), Authority: authority}})
			if !admission.Admitted {
				t.Fatalf("round=%d lane=%s 가 큐에 못 들어갔다: %+v — 이 시험은 둘 다 중재까지 가야 성립한다",
					round, laneID, admission)
			}
		}
		outcome := coordinator.Arbitrate()
		// CONTINUATION < REVERSAL 이므로 가족순 정렬은 지속형을 먼저 걷는다.
		// 그래서 보고되는 이유는 지속형의 것이다.
		if outcome.Refusal != strategyarbiter.RefusalSealMismatch || outcome.Detail != strategyarbiter.DetailEvidenceBinding {
			t.Fatalf("round=%d outcome=%+v, want %q/%q — 범위 안 순서가 새어 나와 진단이 흔들린다",
				round, outcome, strategyarbiter.RefusalSealMismatch, strategyarbiter.DetailEvidenceBinding)
		}
	}
}
