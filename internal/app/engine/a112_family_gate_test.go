//go:build tossos_testseams

package engine

import (
	"context"
	"errors"
	"go/ast"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/continuationlane"
	"github.com/JungHoonGhae/tossinvest-cli/internal/reversallane"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyarbiter"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategycoordinator"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyproposal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyrouter"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyworker"
)

// 이 파일은 태스크 5.1.2.2 의 관문을 값으로 잰다.
//
// 세 상태를 **함께** 세운다. 하나만 세우면 그 시험은 "관문이 있다"만 말하고
// "관문이 무엇을 가르는가"는 말하지 않는다.
//
//	활성화 없음   → 관문이 서지 않는다 → 오늘과 같은 경로 (토글 OFF = upstream)
//	활성화 + 열림 → 그 가족의 제안이 조정자에 닿는다
//	활성화 + 잠김 → 그 가족만 멈추고 이웃 셋은 계속한다

func familyGateFixture(t *testing.T) (*strategyLaneRuntime, strategyrouter.FamilyActivation) {
	t.Helper()
	runtime := newStrategyLaneRuntime(clock.NewFake(time.Date(2026, 9, 3, 1, 0, 0, 0, time.UTC)), nil, "")
	if runtime == nil {
		t.Fatal("생산 레인 런타임이 서지 않았다")
	}
	activation := strategyrouter.FamilyActivationForTest(strategyrouter.MarketKR, 1,
		strategyrouter.AllFourFamiliesForTest(strategyrouter.MarketKR))
	if !activation.Verified() {
		t.Fatal("시험용 활성화가 검증되지 않았다 — 아래 시험은 아무것도 재지 않는다")
	}
	return runtime, activation
}

// collectUnderGate 는 한 시장의 제안 수집을 그 관문 아래에서 돌린다.
//
// 조정자를 직접 부르지 않고 `loader.collect` 로 도는 이유: 관문이 조정 **앞**에
// 선다는 것이 이 태스크의 요점이고, 그 순서는 collectMarket 안에 있다. 조정자를
// 직접 부르면 그 순서가 시험 밖으로 나간다.
func collectUnderGate(t *testing.T, activation strategyrouter.FamilyActivation,
	lanes *strategyLaneRuntime, symbol string, laneIDs ...string,
) strategyProposalMarketAuthority {
	t.Helper()
	now := time.Date(2026, 9, 3, 1, 2, 3, 0, time.UTC)
	loader := testStrategyProposalLoader(t)
	loader.load = func(_ context.Context, config strategyproposal.ProductionConfig,
		targets []strategyproposal.ProductionTarget, _ interfaceOfficialFX,
	) (strategyproposal.ProductionBatchAuthority, error) {
		return arbitrationBatch(t, config, targets, now, symbol, laneIDs), nil
	}
	loader.loadActivation = func(context.Context, StrategyMarket, strategyScheduleMarketAuthority,
		strategyRouteMarketAuthority, time.Time,
	) (strategyrouter.FamilyActivation, error) {
		return activation, nil
	}
	pair := loader.withStrategyLanes(lanes).collect(context.Background(), routeReadySchedulePair(now),
		arbitrationRoutePair(t, now, familyScoresForTest(strategyrouter.MarketKR), symbol, laneIDs...),
		proposalFXPair(now))
	return pair.forMarket(StrategyMarketKR)
}

// selectedFamily 는 이 시장이 여러 가족이 겨룬 그 종목에서 고른 가족이다.
//
// 종목을 이름으로 고르는 이유: KR fixture 는 종목 둘을 경로에 올린다
// (`005930` 은 여러 레인, `000660` 은 지속형 하나). 소유자 범위마다 하나가
// 선택되므로 목록은 둘이고, 관문이 가르는 것은 여러 가족이 겨룬 쪽이다.
// 목록 길이로 판정하면 이 시험은 관문이 아니라 fixture 를 재게 된다.
func selectedFamily(t *testing.T, authority strategyProposalMarketAuthority, symbol string) strategyrouter.Family {
	t.Helper()
	if !authority.snapshot.Ready {
		t.Fatalf("시장이 닫혔다: reason=%s detail=%q", authority.snapshot.Reason, authority.snapshot.ArbitrationDetail)
	}
	for _, entry := range authority.entries {
		if entry.route.approved.Symbol() != symbol {
			continue
		}
		return strategyarbiter.ProposalFamily(strategyarbiter.Proposal{
			Result: entry.authority.Proposal(), Authority: entry.route.route})
	}
	t.Fatalf("%s 가 선택 목록에 없다 (선택=%d 건)", symbol, len(authority.entries))
	return ""
}

// 활성화가 없으면 관문이 서지 않는다. 그것이 오늘 생산의 상태다.
//
// 생산에는 서명된 4-가족 매니페스트가 배포돼 있지 않다(측정: `~/.config/tossctl`
// 에 strategy-* 매니페스트 0 건). 이 시험은 그 상태에서 조정 결과가 관문 이전과
// **같은 값**임을 확인한다 — 세 가족이 한 종목에 제안하면 점수가 가장 높은
// 하나가 그대로 선택된다.
func TestTheProductionActivationLoaderRunsAndFindsNoManifest(t *testing.T) {
	runtime, _ := familyGateFixture(t)
	now := time.Date(2026, 9, 3, 1, 2, 3, 0, time.UTC)
	// **결속을 전부 유효하게 채운다.** 하나라도 비워 두면 이 시험은 "파일이
	// 없다" 가 아니라 "설정이 어긋났다" 를 재게 되고, 그러면 파일 부재라는
	// 축은 한 번도 안 재진다 — 기억에 "반증 실험은 축을 잘못 고르면 아무것도
	// 반증 못 한다" 로 남은 모양이다.
	//
	// 결속은 이제 digest 핀 하나다 [결정 61] — 열쇠 둘이 사라졌으므로 채울 값도
	// 줄었다. 그래도 이 핀은 **유효한** 값이어야 한다: 비우면 위와 같은 이유로
	// 축이 바뀐다.
	env := map[string]string{
		strategyFamilyActivationKRManifestDigestEnv: "sha256:" + strings.Repeat("c", 64),
	}
	dir := t.TempDir()
	loader := newStrategyProposalAuthorityLoader(dir, filepath.Join(dir, "evidence.db"),
		filepath.Join(dir, "journal.db"), "acct", func(name string) string { return env[name] }).
		withStrategyLanes(runtime)
	// loadActivation 을 **일부러 nil 로 둔다.** 생산 구현이 도는 실행이 하나도
	// 없으면 "seam 을 건너뛴다" 변이가 모든 시험을 통과한다.
	if loader.loadActivation != nil {
		t.Fatal("이 시험은 생산 활성화 로더가 도는 것을 재는 것이다")
	}
	routes := arbitrationRoutePair(t, now, familyScoresForTest(strategyrouter.MarketKR), "005930",
		continuationlane.KRContinuationLaneID).forMarket(StrategyMarketKR)
	routes.snapshot.ManifestDigest = "sha256:" + strings.Repeat("d", 64)
	// 파일이 실제로 없다는 것을 먼저 확인한다.
	path := filepath.Join(dir, strategyrouter.ProductionFamilyActivationFileName(strategyrouter.MarketKR))
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("이 시험은 파일이 없는 상태를 재는 것이다: %v", err)
	}
	activation, err := loader.loadFamilyActivation(context.Background(), StrategyMarketKR,
		routeReadySchedulePair(now).forMarket(StrategyMarketKR), routes, now)
	if !errors.Is(err, strategyrouter.ErrProductionFamilyActivationUnavailable) {
		t.Fatalf("want unavailable, got %v", err)
	}
	if activation.Verified() {
		t.Fatal("파일 없이 검증된 활성화가 나왔다")
	}
	gate := loader.familyGateFor(context.Background(), StrategyMarketKR,
		routeReadySchedulePair(now).forMarket(StrategyMarketKR), routes, now)
	if gate.installed() {
		t.Fatal("서명된 매니페스트가 없는데 관문이 섰다 — 오늘 생산에는 그 파일이 없다")
	}
}

// 활성화가 없으면 관문이 서지 않고 조정 결과는 관문 이전과 같은 값이다.
//
// 생산에는 서명된 4-가족 매니페스트가 배포돼 있지 않다(측정: `~/.config/tossctl`
// 에 strategy-* 매니페스트 0 건). 세 가족이 한 종목에 제안하면 점수가 가장 높은
// 하나가 그대로 선택되어야 한다.
func TestWithoutAVerifiedActivationCoordinationIsUnchanged(t *testing.T) {
	runtime, _ := familyGateFixture(t)
	authority := collectUnderGate(t, strategyrouter.FamilyActivation{}, runtime, "005930",
		continuationlane.KRContinuationLaneID, reversallane.KRReversalLaneID)
	// 점수는 REVERSAL 900_000 > CONTINUATION 400_000 이다. 관문이 서지 않았으니
	// 그 순위가 그대로여야 한다.
	if got := selectedFamily(t, authority, "005930"); got != strategyrouter.FamilyReversal {
		t.Fatalf("선택된 가족=%s, want REVERSAL — 관문이 없는데 순위가 바뀌었다", got)
	}
}

// 승격된 레인은 자기 가족의 제안을 조정자로 보내고, 승격되지 않은 레인은 멈춘다.
//
// **두 상태를 한 시험에서 함께 세운다.** 이것이 5.1.2.2 가 요구한
// "TestEveryLaneStaysDormantOnAProposalItActuallyOwns 를 두 상태를 보여 주는
// 시험으로 **교체**" 다. 잠듦만 재면 관문이 통과 도장이어도 초록이다.
func TestAPromotedLaneAdmitsItsFamilyWhileAnUnpromotedOneStopsIt(t *testing.T) {
	runtime, all := familyGateFixture(t)
	// REVERSAL 만 켠 활성화. 나머지 셋은 서명이 OFF 라고 말한 상태다.
	onlyContinuation := strategyrouter.FamilyActivationForTest(strategyrouter.MarketKR, 1,
		map[string]bool{continuationlane.KRContinuationLaneID: true})
	for name, entry := range map[string]struct {
		activation strategyrouter.FamilyActivation
		want       strategyrouter.Family
	}{
		// 넷이 다 켜지면 점수 1위(REVERSAL)가 이긴다.
		"all four promoted": {all, strategyrouter.FamilyReversal},
		// CONTINUATION 만 켜지면 점수 1위의 레인이 승격되지 않았으므로 그
		// 제안은 조정자에 닿지 못하고 2위가 이긴다. **관문이 순위를 바꾼다** —
		// 이것이 관문이 통과 도장이 아니라는 증거다.
		"only continuation promoted": {onlyContinuation, strategyrouter.FamilyContinuation},
	} {
		t.Run(name, func(t *testing.T) {
			if !entry.activation.Verified() {
				t.Fatal("활성화가 검증되지 않았다")
			}
			authority := collectUnderGate(t, entry.activation, runtime, "005930",
				continuationlane.KRContinuationLaneID, reversallane.KRReversalLaneID)
			if got := selectedFamily(t, authority, "005930"); got != entry.want {
				t.Fatalf("선택된 가족=%s, want %s", got, entry.want)
			}
			// 관문이 쓴 활성화가 권위와 함께 나와야 한다. 나오지 않으면 같은
			// 주기의 레인 관측이 관문과 다른 승격을 본다.
			if !authority.familyActivation().Verified() {
				t.Fatal("권위가 관문의 활성화를 싣지 않았다")
			}
		})
	}
}

// 잠긴 레인은 자기 가족만 멈추고 이웃은 계속한다. 이 change 가 사려던 것이다.
//
// 여기서 잠그는 것은 REVERSAL — 점수가 가장 높아 관문이 없으면 **이기는**
// 가족이다. 잠긴 뒤에는 CONTINUATION 이 이겨야 한다. 이웃이 계속한다는 것을
// "시장이 안 닫혔다"로만 재면, 승자를 잠갔을 때 아무도 못 이기는 구현도
// 통과한다.
func TestALatchedLaneStopsItsFamilyAndItsPeersKeepTrading(t *testing.T) {
	runtime, activation := familyGateFixture(t)
	latched := 0
	for _, lane := range runtime.lanesFor(StrategyMarketKR) {
		if lane.Key().Family != strategyrouter.FamilyReversal {
			continue
		}
		if _, locked := lane.Fail("a measured fault", true); !locked {
			t.Fatal("비정상 실패 하나로 레인이 잠기지 않았다")
		}
		latched++
	}
	if latched != 1 {
		t.Fatalf("잠근 레인=%d 개, want 1", latched)
	}
	authority := collectUnderGate(t, activation, runtime, "005930",
		continuationlane.KRContinuationLaneID, reversallane.KRReversalLaneID)
	if got := selectedFamily(t, authority, "005930"); got != strategyrouter.FamilyContinuation {
		t.Fatalf("선택된 가족=%s, want CONTINUATION — 점수 1위가 잠겼으면 2위가 이겨야 한다."+
			" 시장이 통째로 닫히면 이 change 가 사려던 가족 단위 고장 격리가 없는 것이다", got)
	}
}

// 시장이 닫혀도 관문의 활성화는 권위와 함께 나온다.
//
// 반증이 가르쳐 준 자리다. 닫힘 갈래에서 활성화를 빼는 변이(E6)가 첫 배터리에서
// **살아남았다** — 준비된 시장만 재고 있었기 때문이다. 그 상태의 대가는 관측이다:
// 관문은 섰는데 그 주기의 레인 관측은 승격이 빈 값으로 돌아 DORMANT 를 적는다.
// 즉 시장이 고장 난 바로 그 주기에 운영자의 레인 화면이 어두워진다.
func TestAClosedMarketStillCarriesTheGatesActivation(t *testing.T) {
	runtime, activation := familyGateFixture(t)
	now := time.Date(2026, 9, 3, 1, 2, 3, 0, time.UTC)
	// 최고점 동률을 만들어 중재가 시장을 닫게 한다.
	tied := familyScoresForTest(strategyrouter.MarketKR)
	for index := range tied {
		if tied[index].Family == strategyrouter.FamilyReversal {
			tied[index].ScorePPM = 400_000 // 지속형과 같은 점수
		}
	}
	loader := testStrategyProposalLoader(t)
	loader.load = func(_ context.Context, config strategyproposal.ProductionConfig,
		targets []strategyproposal.ProductionTarget, _ interfaceOfficialFX,
	) (strategyproposal.ProductionBatchAuthority, error) {
		return arbitrationBatch(t, config, targets, now, "005930",
			[]string{continuationlane.KRContinuationLaneID, reversallane.KRReversalLaneID}), nil
	}
	loader.loadActivation = func(context.Context, StrategyMarket, strategyScheduleMarketAuthority,
		strategyRouteMarketAuthority, time.Time,
	) (strategyrouter.FamilyActivation, error) {
		return activation, nil
	}
	pair := loader.withStrategyLanes(runtime).collect(context.Background(), routeReadySchedulePair(now),
		arbitrationRoutePair(t, now, tied, "005930",
			continuationlane.KRContinuationLaneID, reversallane.KRReversalLaneID), proposalFXPair(now))
	authority := pair.forMarket(StrategyMarketKR)
	if authority.snapshot.Ready {
		t.Fatalf("이 시험은 닫힌 시장을 재는 것이다: reason=%s", authority.snapshot.Reason)
	}
	if !authority.familyActivation().Verified() {
		t.Fatal("닫힌 시장이 관문의 활성화를 버렸다 — 그 주기의 레인 관측이 승격을 못 본다")
	}
	if authority.familyActivation().Generation() != activation.Generation() {
		t.Fatalf("세대=%d, want %d", authority.familyActivation().Generation(), activation.Generation())
	}
}

// 관문이 만든 봉투와 엔진이 만든 봉투는 **같은 값**이다.
//
// 두 사본이 있는 것은 사실이고, 그것이 갈릴 수 있다는 것도 사실이다. 그래서
// 구조를 단언하지 않고 값을 견준다 — 필드가 하나 늘 때 구조 단언은 통과하고
// 이 등식은 실패한다.
func TestTheFamilyGateAndTheLegacyPathBuildTheSameEnvelope(t *testing.T) {
	runtime, activation := familyGateFixture(t)
	inputs := laneOwnedInputs(t)
	gate := strategyFamilyGate{activation: activation, lanes: runtime.lanesFor(StrategyMarketKR)}
	compared := 0
	for _, input := range inputs {
		legacy := strategycoordinator.Envelope{Scope: input.Scope,
			SnapshotDigest: input.SnapshotDigest, Proposal: input.Proposal}
		admitted, outcome, ok := gate.admit(input)
		if !ok || outcome != strategyworker.OutcomeEmitted {
			t.Fatalf("승격된 레인이 %s 를 내보내지 않았다 (%s)", input.Proposal.Result.Lineage.LaneID, outcome)
		}
		// Envelope 는 슬라이스를 담은 제안을 들고 있어 `==` 로 못 견준다.
		// 필드를 하나씩 견주고, 제안은 봉인된 계보 신원으로 견준다 —
		// 그 신원이 실행 조건 전부를 해시한 값이다.
		if admitted.Scope != legacy.Scope || admitted.SnapshotDigest != legacy.SnapshotDigest ||
			admitted.Proposal.Result.Lineage.Identity != legacy.Proposal.Result.Lineage.Identity ||
			strategyarbiter.ProposalFamily(admitted.Proposal) != strategyarbiter.ProposalFamily(legacy.Proposal) {
			t.Fatalf("관문의 봉투와 기존 경로의 봉투가 다르다:\n gate: %+v\nlegacy: %+v", admitted, legacy)
		}
		compared++
	}
	if compared == 0 {
		t.Fatal("견준 봉투가 0 개다 — 이 시험은 아무것도 증명하지 않는다")
	}
}

// 서명된 활성화가 요구한 ProtectionReady 하한을 **주문 경로**가 지킨다
// (태스크 8.8.2).
//
// **앞 판본이 왜 아무것도 막지 못했는지 먼저 적는다.** 그것은 이 결속을
// `buildProductionStrategyMarketWorker` 안에 두고 `Effective` 를 껐다. 그런데
// 주문은 refresh worker 의 사이클이 `dispatchHandoff().Deliver` 로 내보내고 그
// 경로는 그 서술자를 읽지 않는다 — 화면만 어두워지고 주문은 계속 나갔다.
// 게다가 결속 대상이 per-cycle 스냅샷 봉인이라 사람이 서명한 상수가 어떤 정상
// 입력으로도 같아질 수 없었고, 시험은 **살아 있는 값을 읽어 매니페스트에 넣어**
// 그 사실을 가렸다. 그래서 이 시험은 값을 시스템에서 읽지 않는다: 하한을 숫자로
// 적고, 배선의 보호 세대(9)와 견준다.
//
// 세 경우를 함께 세운다. 하나만 세우면 "항상 거절" 판본도 통과한다.
func TestTheOrderPathRefusesAProtectionPostureOlderThanTheSignedFloor(t *testing.T) {
	const wiredProtectionGeneration = 9
	for name, entry := range map[string]struct {
		activation func() strategyrouter.FamilyActivation
		wantPlaced int
	}{
		"활성화가 없으면 하한을 아무도 요구하지 않는다 — 오늘의 값": {
			activation: func() strategyrouter.FamilyActivation { return strategyrouter.FamilyActivation{} },
			wantPlaced: 1},
		"배선의 보호 세대가 하한과 같으면 주문이 나간다": {
			activation: func() strategyrouter.FamilyActivation {
				return strategyrouter.FamilyActivationWithBindingsForTest(strategyrouter.MarketKR, 1,
					strategyrouter.AllFourFamiliesForTest(strategyrouter.MarketKR), wiredProtectionGeneration)
			}, wantPlaced: 1},
		"배선의 보호 세대가 하한보다 낮으면 주문이 거절된다": {
			activation: func() strategyrouter.FamilyActivation {
				return strategyrouter.FamilyActivationWithBindingsForTest(strategyrouter.MarketKR, 1,
					strategyrouter.AllFourFamiliesForTest(strategyrouter.MarketKR), wiredProtectionGeneration+1)
			}, wantPlaced: 0},
	} {
		t.Run(name, func(t *testing.T) {
			cycle, proposals, _, spy := pairedStrategyDispatchCycleFixture(t)
			kr := proposals.forMarket(StrategyMarketKR)
			kr.activation = entry.activation()
			proposals.kr = kr
			cycle.proposals = proposals
			result, handedOff := kr.dispatchHandoff().Single()
			if !handedOff {
				t.Fatal("배선이 건네줄 제안을 만들지 못했다 — 이 시험은 아무것도 재지 않는다")
			}
			_, err := cycle.dispatch(context.Background(), deliverForTest(t, result))
			spy.mu.Lock()
			placed := len(spy.calls)
			spy.mu.Unlock()
			if placed != entry.wantPlaced {
				t.Fatalf("브로커 주문=%d 건, want %d (err=%v)", placed, entry.wantPlaced, err)
			}
			if entry.wantPlaced == 0 && err == nil {
				t.Fatal("주문은 안 나갔는데 오류도 없다 — 조용한 거절은 운영자가 못 본다")
			}
		})
	}
}

// 어느 레인도 자기 것이라 하지 않는 제안은 조정자에 닿지 않는다.
//
// 활성화된 런타임에서 주인 없는 제안이 통과하면 이 관문이 있는 이유가 사라진다.
// 레인 목록을 비워 그 상태를 만든다 — 그러면 `installed` 가 거짓이 되어 통과할
// 수도 있는데, 그래서 레인을 **다른 시장의 것으로** 채운다.
func TestAProposalNoLaneOwnsIsStoppedRatherThanPassedThrough(t *testing.T) {
	runtime, _ := familyGateFixture(t)
	usActivation := strategyrouter.FamilyActivationForTest(strategyrouter.MarketUS, 1,
		strategyrouter.AllFourFamiliesForTest(strategyrouter.MarketUS))
	gate := strategyFamilyGate{activation: usActivation, lanes: runtime.lanesFor(StrategyMarketUS)}
	if !gate.installed() {
		t.Fatal("관문이 서지 않았다 — 이 시험은 아무것도 재지 않는다")
	}
	for _, input := range laneOwnedInputs(t) {
		envelope, outcome, ok := gate.admit(input)
		if ok {
			t.Fatalf("US 레인이 KR 제안을 통과시켰다: %+v", envelope)
		}
		if outcome != strategyworker.OutcomeRefused {
			t.Fatalf("멈춘 이유=%s, want REFUSED", outcome)
		}
	}
}

// 레인은 제안 수집 **앞**에서 세워진다.
//
// 순서가 안전이다. 관문은 레인의 잠금을 읽어 판정하고 그 잠금은 원장에서
// 태어난다(5.3.3). 뒤에 세우면 재시작 뒤 **첫 주기**에 durably 잠긴 레인이 열린
// 것으로 읽히고 그 가족의 제안이 조정자에 닿는다. 그 창은 한 주기뿐이라 어떤
// 행동 시험도 우연히 잡지 못하므로 순서를 구조로 못 박는다.
func TestTheLanesAreBuiltBeforeTheProposalsAreCollected(t *testing.T) {
	file := parseEngineFile(t, filepath.Join(".", "strategy_entry_supervisor.go"))
	decl := engineFuncDecl(t, file, "NewPairedStrategyEntryProductionAssembly")
	laneLine, proposalLine := 0, 0
	ast.Inspect(decl, func(node ast.Node) bool {
		call, ok := node.(*ast.CallExpr)
		if !ok {
			return true
		}
		selector, isSelector := call.Fun.(*ast.SelectorExpr)
		if !isSelector {
			return true
		}
		switch selector.Sel.Name {
		case "productionStrategyLanes":
			laneLine = int(selector.Sel.Pos())
		case "withStrategyLanes":
			proposalLine = int(selector.Sel.Pos())
		}
		return true
	})
	if laneLine == 0 || proposalLine == 0 {
		t.Fatalf("읽을 자리가 사라졌다: productionStrategyLanes=%d withStrategyLanes=%d",
			laneLine, proposalLine)
	}
	if laneLine >= proposalLine {
		t.Fatal("레인이 제안 수집보다 뒤에 세워진다 — 재시작 뒤 첫 주기에 잠긴 레인이 열린다")
	}
}

// 관문이 멈춘 제안이 시장의 소유자 범위 수를 줄여서는 안 된다 (태스크 8.8.1).
//
// **8.5 적대 리뷰가 실측으로 연 자리다.** 이 저장소는 같은 위험을 이미 두 번
// 고쳤고(5.4.2 큐 넘침, 5.4.3 잃어버린 제안) 그 근거가
// `strategy_proposal_authority.go` 와 `strategycoordinator/coordinator.go` 의
// 주석에 적혀 있다: 닫힌 범위만 목록에서 빼면 목록이 둘에서 하나로 줄고, 짧아진
// 목록이 아래 파이프라인의 `strategyhandoff.Capacity = 1` 관문을 **오히려
// 만족시켜** 막으려던 것과 상관없는 다른 종목이 대신 풀린다. 5.1.2.2 의 관문이
// 세 번째 사례를 만들었다.
//
// KR fixture 는 종목 둘을 올린다: `005930` 은 여러 가족, `000660` 은 지속형
// 하나뿐이다. 그래서 **지속형을 잠그면** `000660` 범위는 봉투를 하나도 못 낸다.
// 그 상태에서 시장이 열려 있으면 `005930` 하나가 남아 경계를 통과하고, 고장
// 하나가 시스템을 더 관대하게 만든 것이다.
//
// 잠그는 가족이 `TestALatchedLaneStopsItsFamilyAndItsPeersKeepTrading` 과 다른
// 것이 요점이다. 거기서는 역전형을 잠가 두 범위가 **둘 다** 봉투를 내므로 시장이
// 열려 있어야 하고, 여기서는 한 범위가 통째로 비므로 닫혀야 한다. 두 시험이
// 함께 있어야 "가족 격리"와 "범위 소멸"이 구별된다.
func TestAGatedFamilyMustNotShrinkTheMarketIntoTheExactlyOneValve(t *testing.T) {
	runtime, activation := familyGateFixture(t)
	latched := 0
	for _, lane := range runtime.lanesFor(StrategyMarketKR) {
		if lane.Key().Family != strategyrouter.FamilyContinuation {
			continue
		}
		if _, locked := lane.Fail("a measured fault", true); !locked {
			t.Fatal("비정상 실패 하나로 레인이 잠기지 않았다")
		}
		latched++
	}
	if latched != 1 {
		t.Fatalf("잠근 레인=%d 개, want 1 — fixture 가 바뀌면 이 시험은 다른 것을 잰다", latched)
	}
	authority := collectUnderGate(t, activation, runtime, "005930",
		continuationlane.KRContinuationLaneID, reversallane.KRReversalLaneID)
	if authority.snapshot.Ready {
		t.Fatalf("시장이 열려 있다 (선택=%d 건) — 잠긴 가족이 한 범위를 통째로 지웠는데도"+
			" 남은 하나가 `strategyhandoff.Capacity = 1` 관문을 만족시킨다."+
			" 고장 하나가 시스템을 더 관대하게 만들었다", len(authority.entries))
	}
	if authority.snapshot.Reason != StrategyProposalFamilyGateClosed {
		t.Fatalf("reason=%s, want %s", authority.snapshot.Reason, StrategyProposalFamilyGateClosed)
	}
	if authority.snapshot.RefusedCount != authority.snapshot.RoutedCount {
		t.Fatalf("거절=%d, 경로=%d — 시장이 닫히면 경로에 오른 종목 전부가 제안을 못 낸 것이다",
			authority.snapshot.RefusedCount, authority.snapshot.RoutedCount)
	}
	if _, handedOff := authority.dispatchHandoff().Single(); handedOff {
		t.Fatal("닫힌 시장이 공유 dispatch 경계에 값을 건넸다")
	}
}

// 관문이 무엇을 왜 멈췄는지가 운영자에게 보여야 한다 (태스크 8.8.1).
//
// 앞 판본의 `arbitration.gated` 는 **쓰이기만 하고 아무도 안 읽었다**. 리뷰어
// 다섯이 각각 짚었다. 그 필드의 주석은 DORMANT(안 켰다)·LATCHED(고장으로
// 닫혔다)·REFUSED(주인 없다)를 운영자가 가른다고 약속하는데, 코드는 그것을
// 함수 끝에서 버렸다. 약속을 지키거나 약속을 지우거나 둘 중 하나여야 한다.
func TestTheGateSaysWhichKindOfStopItMade(t *testing.T) {
	runtime, activation := familyGateFixture(t)
	for _, lane := range runtime.lanesFor(StrategyMarketKR) {
		if lane.Key().Family != strategyrouter.FamilyReversal {
			continue
		}
		if _, locked := lane.Fail("a measured fault", true); !locked {
			t.Fatal("비정상 실패 하나로 레인이 잠기지 않았다")
		}
	}
	authority := collectUnderGate(t, activation, runtime, "005930",
		continuationlane.KRContinuationLaneID, reversallane.KRReversalLaneID)
	if !authority.snapshot.Ready {
		t.Fatalf("시장이 닫혔다: reason=%s — 두 범위가 다 봉투를 냈으므로 열려 있어야 한다",
			authority.snapshot.Reason)
	}
	if authority.snapshot.GatedCount != 1 {
		t.Fatalf("멈춘 제안=%d 건, want 1", authority.snapshot.GatedCount)
	}
	want := string(strategyworker.OutcomeLatched)
	if len(authority.snapshot.GatedOutcomes) != 1 || authority.snapshot.GatedOutcomes[0] != want {
		t.Fatalf("멈춘 종류=%v, want [%s] — 복구 증거가 필요한 상태가 '아직 안 켰다'로 보이면 안 된다",
			authority.snapshot.GatedOutcomes, want)
	}
}
