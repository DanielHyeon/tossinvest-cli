//go:build tossos_testseams

package strategyproposal

import (
	"context"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/breakoutlane"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyrouter"
)

// 사라진 제안은 없던 제안과 달라야 한다.
//
// `LanesFor` 가 빈 목록을 돌려주는 일은 지금 두 가지를 한 이름으로 뭉친다 —
// "이 종목은 원래 제안이 없다"와 "이 종목의 제안이 고장으로 사라졌다".
// 뭉치면 고장 하나가 시장의 목록을 줄이고, 줄어든 목록이 아래 파이프라인의
// `len(entries) != 1` 관문을 오히려 **만족시켜** 상관없는 다른 종목을 풀어 준다.
// 고장 하나가 시스템을 더 관대하게 만드는 것이다.
//
// 여기서는 매니페스트와 자격 집합이 **둘 다 받아들인** 스코프가 봉인된 증거를
// 되살리지 못하게 만든다. 스냅샷 다이제스트는 그대로 두고 ID 만 없는 것으로
// 바꾼다 — 다이제스트를 바꾸면 매니페스트 검증이 먼저 거절해서 정작 보려던
// 안쪽 경로에 닿지 못한다.
func TestAProposalLostAfterAdmissionIsRecordedAsATypedFault(t *testing.T) {
	now := time.Date(2026, 8, 4, 1, 0, 3, 0, time.UTC)
	config, target, fx := productionFixtureWith(t, strategyrouter.MarketKR, now, func(scope *productionScope) {
		scope.SnapshotID = "이-스냅샷은-증거 저장소에 없다"
	})
	batch, err := LoadProductionAuthorityBatch(context.Background(), config, []ProductionTarget{target}, fx)
	if err != nil {
		t.Fatalf("배치 자체가 실패했다: %v — 이 시험은 배치가 성립한 뒤의 조용한 유실을 본다", err)
	}
	if batch.Len() != 0 {
		t.Fatalf("제안이 %d 개 나왔다 — 되살리지 못한 증거로 제안이 만들어졌다", batch.Len())
	}
	absence, faulted := batch.Fault()
	if !faulted {
		t.Fatal("증거를 되살리지 못했는데 배치가 아무 고장도 기록하지 않았다 — 이 종목은 조용히 사라진다")
	}
	if absence.Reason != ProductionAbsenceEvidenceReplay {
		t.Fatalf("reason=%q, want %q", absence.Reason, ProductionAbsenceEvidenceReplay)
	}
	if absence.Symbol != target.Approved.Symbol() {
		t.Fatalf("symbol=%q, want %q — 어느 종목이 사라졌는지 못 적으면 운영자가 찾을 수 없다",
			absence.Symbol, target.Approved.Symbol())
	}
	if absence.LaneID == "" {
		t.Fatal("lane 이 비었다 — 한 종목이 네 가족을 낼 수 있으므로 종목만으로는 어느 것이 사라졌는지 모른다")
	}
}

// 정상 부재는 고장이 아니다.
//
// 매니페스트가 지금 후보와 다른 후보 생애를 가리키면 그 스코프는 원래 제안을
// 만들 수 없다. 고장이 아니므로 시장을 닫지 않는다 — 닫으면 매니페스트가 조금만
// 묵어도 그 시장이 통째로 멈춘다. 이 시험이 없으면 위 시험을 만족시키는 가장
// 쉬운 구현("빈 결과면 무조건 고장")이 통과해 버린다.
func TestAScopeTheCurrentCandidateDoesNotMatchIsAbsenceNotFault(t *testing.T) {
	now := time.Date(2026, 8, 4, 1, 0, 3, 0, time.UTC)
	config, target, fx := productionFixtureWith(t, strategyrouter.MarketKR, now, func(scope *productionScope) {
		scope.CandidateID = "다른-후보-생애"
	})
	batch, err := LoadProductionAuthorityBatch(context.Background(), config, []ProductionTarget{target}, fx)
	if err != nil {
		t.Fatalf("배치가 실패했다: %v", err)
	}
	if batch.Len() != 0 {
		t.Fatalf("제안이 %d 개 나왔다 — 후보 생애가 다른 스코프가 통과했다", batch.Len())
	}
	if absence, faulted := batch.Fault(); faulted {
		t.Fatalf("정상 부재를 고장으로 적었다: %+v — 이러면 묵은 매니페스트 하나가 시장을 멈춘다", absence)
	}
}

// 아무 일 없는 배치는 고장을 들고 있지 않다.
func TestAHealthyBatchCarriesNoFault(t *testing.T) {
	now := time.Date(2026, 8, 4, 1, 0, 3, 0, time.UTC)
	config, target, fx := productionFixture(t, strategyrouter.MarketKR, now)
	batch, err := LoadProductionAuthorityBatch(context.Background(), config, []ProductionTarget{target}, fx)
	if err != nil || batch.Len() != 1 {
		t.Fatalf("정상 배치를 못 만들었다: err=%v len=%d", err, batch.Len())
	}
	if absence, faulted := batch.Fault(); faulted {
		t.Fatalf("정상 배치가 고장을 들고 있다: %+v", absence)
	}
}

// 아직 만들지 않은 것은 고장이 아니다.
//
// 돌파-되돌림 레인은 필요한 증거(ATR·RVOL·윗꼬리·거래량 확장)가 저장되지도
// 만들어지지도 않아서 입구에서 닫혀 있다. 그것은 이미 아는 구조적 사유이지
// 무언가 잘못된 것이 아니다. 고장으로 세면 **돌파 스코프가 실린 매니페스트 하나가
// 그 시장을 매 주기 영영 닫는다** — 5.4.2 리뷰가 큐 상한에서 잡아낸 것과 같은 모양의
// 고장이다(정상 입력을 시스템이 거부하게 만드는 것).
//
// 이 시험은 처음에 `errors.Is(ErrBreakoutEvidenceUnavailable)` 라는 **예외 하나**를
// 지켰다. 지금은 그 예외가 없다 — `buildLaneInput` 의 오류 공간 전체가 정상 부재로
// 바뀌었기 때문이다(TestAnEvaluationRefusalIsAbsenceNotFault 가 그 이유를 적고 있다).
// 그래도 이 시험을 남긴다. 예외를 지키는 시험이 아니라 **이 레인이 시장을 닫지
// 않는다**는 성질을 지키는 시험이고, 그 성질이 깨지는 것이 실제 위험이다.
//
// 경로 결정과 매니페스트 스코프를 **함께** 돌파 레인으로 옮긴다. 하나만 옮기면
// 자격 집합이 그 스코프를 안 받아서 buildLaneInput 에 닿기도 전에 정상 부재로 걸러지고,
// 그러면 이 시험은 겨눈 것을 못 본 채 통과한다.
func TestABreakoutLaneWithNoEvidenceYetIsAbsenceNotFault(t *testing.T) {
	now := time.Date(2026, 8, 4, 1, 0, 3, 0, time.UTC)
	config, target, fx := productionFixtureOn(t, strategyrouter.MarketKR, now,
		&productionLaneOverride{laneID: breakoutlane.KRLaneID, version: breakoutlane.LaneVersionV1,
			horizon: strategyrouter.HorizonShort}, nil)
	batch, err := LoadProductionAuthorityBatch(context.Background(), config, []ProductionTarget{target}, fx)
	if err != nil {
		t.Fatalf("배치가 실패했다: %v", err)
	}
	if batch.Len() != 0 {
		t.Fatalf("돌파 레인이 제안을 %d 개 냈다 — 없는 증거로 값을 지어냈다", batch.Len())
	}
	if absence, faulted := batch.Fault(); faulted {
		t.Fatalf("돌파 증거 부재를 고장으로 적었다: %+v — 이러면 그 시장이 매 주기 닫힌다", absence)
	}
}

// 평가가 이유를 들고 거절한 것은 고장이 아니다.
//
// `Propose` 는 거절할 때 Code 를 채운 Result 를 돌려주고, 그 Result 는
// `ValidProposal()` 이 거짓이다. 그래서 "봉인이 안 섰다"와 "시장 상태가 안 맞아서
// 거절했다"가 같은 검사에 걸린다. 후자를 고장으로 세면 **스프레드가 한 번 넓어질
// 때마다 그 시장이 통째로 닫힌다** — 동결 골든이 quote_fx_sizing 으로 이름 붙여 둔,
// 매일 일어나는 정상 거절들이다.
//
// 여기서는 목표가를 진입가 아래로 내려 보호적이지 않은 목표로 만든다.
// validScopes 는 가격 사이의 관계를 보지 않으므로 이 스코프는 매니페스트 검증과
// 자격 집합을 통과해 평가까지 간 뒤 거기서 거절된다.
func TestAnEvaluationRefusalIsAbsenceNotFault(t *testing.T) {
	now := time.Date(2026, 8, 4, 1, 0, 3, 0, time.UTC)
	config, target, fx := productionFixtureWith(t, strategyrouter.MarketKR, now, func(scope *productionScope) {
		scope.TargetPriceMinor = "100"
	})
	batch, err := LoadProductionAuthorityBatch(context.Background(), config, []ProductionTarget{target}, fx)
	if err != nil {
		t.Fatalf("배치가 실패했다: %v", err)
	}
	if batch.Len() != 0 {
		t.Fatalf("제안이 %d 개 나왔다 — 보호적이지 않은 목표가 통과했다. 이 시험은 거절된 평가를 봐야 한다", batch.Len())
	}
	if absence, faulted := batch.Fault(); faulted {
		t.Fatalf("정상 거절을 고장으로 적었다: %+v — 이러면 스프레드 한 번에 시장이 닫힌다", absence)
	}
}

// 건네받은 경로 권한을 쓸 수 없으면 고장이다.
//
// 이것은 "이 레인은 자격이 없다"와 다르다. 자격이 있는지조차 말할 수 없는 상태이고,
// 그 스코프가 제안을 못 낸 이유를 우리는 모른다. 모르면 닫는다.
func TestAnUnusableRouteAuthorityIsAFault(t *testing.T) {
	now := time.Date(2026, 8, 4, 1, 0, 3, 0, time.UTC)
	config, target, fx := productionFixture(t, strategyrouter.MarketKR, now)
	// 소유자 리비전이 0 이면 RouteSet 이 아무 결정도 세우지 못한다.
	target.Router.ExpectedOwnerRevision = 0
	// 전제부터 확인한다. 이것이 참이 아니면 아래 단언은 다른 것을 보고 있다.
	if routed := strategyrouter.RouteSet(target.Router); routed.Code == strategyrouter.RefusalNone && routed.Valid() {
		t.Fatal("경로 권한이 여전히 쓸 만하다 — 이 시험은 못 쓰는 권한을 봐야 한다")
	}
	batch, err := LoadProductionAuthorityBatch(context.Background(), config, []ProductionTarget{target}, fx)
	if err != nil {
		t.Fatalf("배치가 실패했다: %v", err)
	}
	absence, faulted := batch.Fault()
	if !faulted {
		t.Fatal("경로 권한을 못 쓰는데 아무 고장도 기록하지 않았다")
	}
	if absence.Reason != ProductionAbsenceRouteUnusable {
		t.Fatalf("reason=%q, want %q", absence.Reason, ProductionAbsenceRouteUnusable)
	}
}
