package engine

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/orderintent"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyhandoff"
)

type strategyDispatchGateway interface {
	ObserveStrategyProtection(context.Context, string, uint64) (execgw.StrategyProtectionAuthority, error)
	ObserveStrategyEntryGate(context.Context, string, string) (execgw.StrategyEntryGateAuthority, error)
	PlaceClaimedStrategy(context.Context, execgw.StrategyPlaceRequest) (execgw.Outcome, error)
}

// strategyDispatchCycle is package-private and is not a trigger. It owns the
// only sequence that can turn an accepted sealed result into an official
// Gateway request once a separately authorized production worker invokes it.
type strategyDispatchCycle struct {
	journal  *journal.Journal
	gateway  strategyDispatchGateway
	firstLeg *strategyFirstLegAdmissionBridge
	schedule strategyScheduleAuthorityPair
	fx       strategyFXAuthorityPair
	risk     strategyRiskAuthorityPair
	// proposals 는 이 파도의 제안 권한이다. 여기서 읽는 것은 하나뿐이다:
	// 그 시장의 서명된 4-가족 활성화가 요구하는 **ProtectionReady 하한**
	// (태스크 8.8.2). 제안 목록은 봉투가 이미 들고 오므로 읽지 않는다.
	proposals          strategyProposalAuthorityPair
	revalidateSchedule func(context.Context, StrategyMarket, strategyScheduleMarketAuthority) error
	// now 는 실시계다 (태스크 8.7.2). 4-가족 활성화의 만료를 파도 시각이 아니라 **지금**
	// 으로 잰다 — admission 앞 한 번, 그리고 게이트웨이가 브로커 바이트 전에 부르는 최종
	// 검사에서 한 번. 생산 조립(`NewPairedStrategyEntryProductionAssembly`)이 `clk.Now` 를
	// 넣는다. nil 이면 검증된 가족 활성화를 가진 시장의 주문은 거절이다(fail-closed);
	// 검증 안 된 시장(미선언·기존 경로)은 시각을 보지 않으므로 영향이 없다.
	now func() time.Time

	owner *strategyDispatchOwnerCoordinator
}

type strategyDispatchOwnerCoordinator struct {
	mu    sync.Mutex
	owner journal.StrategyDispatchOwner
}

func newStrategyDispatchCycle(jrn *journal.Journal, gateway strategyDispatchGateway, firstLeg *strategyFirstLegAdmissionBridge,
	schedule strategyScheduleAuthorityPair, fx strategyFXAuthorityPair, risk strategyRiskAuthorityPair,
	proposals strategyProposalAuthorityPair, owner *strategyDispatchOwnerCoordinator,
) *strategyDispatchCycle {
	if owner == nil {
		owner = &strategyDispatchOwnerCoordinator{}
	}
	return &strategyDispatchCycle{journal: jrn, gateway: gateway, firstLeg: firstLeg, schedule: schedule,
		fx: fx, risk: risk, proposals: proposals, owner: owner}
}

// dispatch 는 봉투에 담겨 온 값만 받는다.
//
// 인자가 strategyflow.Result 가 아니라 strategyhandoff.Delivered 인 것이
// 요점이다. 경계를 지나지 않은 Result 를 여기 넘기는 편집은 **컴파일되지
// 않는다** — 앞선 세 라운드의 적대 리뷰가 뚫은 세 철자(rawSelection, relay,
// rawTailProposal)가 전부 그 자리에서 멈춘다. 이름을 보는 검사가 아니라
// 서명이 막으므로 새 철자를 만들 여지가 없다.
//
// 밖에서 만든 영값 봉투는 아래 첫 줄에서 걸린다 — 그 성질은 가정이 아니라
// TestAForgedEnvelopeIsRefusedBeforeAnyGatewayCall 이 값으로 확인한다.
func (cycle *strategyDispatchCycle) dispatch(ctx context.Context, delivered strategyhandoff.Delivered) (execgw.Outcome, error) {
	result := delivered.Result()
	accepted, refusal := validateStrategyFirstLegResult(result)
	if refusal.Code != "" {
		return execgw.Outcome{}, errors.New(refusal.Detail)
	}
	if cycle == nil || cycle.journal == nil || cycle.gateway == nil || cycle.firstLeg == nil || ctx == nil {
		return execgw.Outcome{}, errors.New("engine: strategy dispatch cycle unavailable")
	}
	market := StrategyMarket(accepted.market)
	schedule, fx := cycle.schedule.forMarket(market), cycle.fx.forMarket(market)
	if !schedule.snapshot.Ready || schedule.restore.Activation == nil || schedule.desired.Revision == 0 ||
		schedule.calendar.Version == "" || !fx.snapshot.Ready || !fx.read.valid {
		return execgw.Outcome{}, errors.New("engine: current market schedule or FX authority unavailable")
	}
	// Complete every fallible read-only boundary before q_final admission commits
	// the campaign, aggregate reservation and five bucket holds. The Gateway will
	// still re-check protection and entry authority twice around SUBMITTING.
	if _, err := strategyFirstLegPlaceIntent(accepted, result.Quantity); err != nil {
		return execgw.Outcome{}, err
	}
	protection, err := cycle.gateway.ObserveStrategyProtection(ctx, strings.ToLower(string(market)), result.Quantity)
	if err != nil {
		return execgw.Outcome{}, err
	}
	// 서명된 4-가족 활성화가 요구한 ProtectionReady 하한 (태스크 8.8.2).
	//
	// **이 자리인 이유.** 보호 세대는 주문을 내려는 순간에만 존재하는 사실이고,
	// 이 함수가 그 사실을 들고 있으면서 주문을 거절할 수 있는 유일한 자리다.
	// 앞 판본은 이 결속을 worker 서술자를 만드는 함수에 두었는데, 그 서술자는
	// 화면과 승격만 움직이고 이 경로는 그것을 읽지 않는다 — 즉 어떤 주문도
	// 막지 못했다(8.5 적대 리뷰가 값으로 보였다).
	//
	// **등식이 아니라 하한인 이유.** 보호 준비 상태는 살아 있고 스냅샷마다
	// 세대가 오른다. 등식을 걸면 사람이 서명한 상수가 어떤 정상 입력으로도
	// 참이 될 수 없다. 하한은 "내가 승인한 것보다 오래된 보호 자세로는 내지
	// 마라"를 그대로 말하고, 세대가 단조 증가하므로 안전 방향으로만 어긋난다.
	//
	// 활성화가 없는 시장은 아무것도 요구하지 않는다 — 그것이 오늘의 동작이다.
	if activation := cycle.proposals.forMarket(market).familyActivation(); activation.Verified() {
		if protection.Generation() < activation.ProtectionReadyMinGeneration() {
			return execgw.Outcome{}, fmt.Errorf(
				"engine: protection readiness generation %d is below the signed four-family floor %d",
				protection.Generation(), activation.ProtectionReadyMinGeneration())
		}
	}
	// 4-가족 활성화의 만료를 **실시계로** 잰다 (태스크 8.7.2).
	//
	// 활성화는 파도 시각에 검증되고 주문은 그 뒤에 나간다. 두 자리에서 잰다: 여기
	// (q_final admission **앞** — 위 머리 주석의 "admission 이 커밋하기 전에 실패할 수 있는
	// 읽기 전용 경계를 전부 끝낸다"), 그리고 아래 `FinalAuthorityCheck` 안(게이트웨이가
	// 브로커 바이트 **전**에 부른다). 규칙(검증 안 된 시장은 상한 그대로 · 남은 수명 0
	// 이하면 만료 · 결과는 상한을 넘지 않음)은 strategyrouter 한 곳에 있고 여기는 배관이다.
	//
	// lease TTL 은 상대 시간이고 원장이 자기 시각에 더하므로, lease 행의 명목 만료는
	// 여기서 발급까지 걸린 δ 만큼 활성화 만료를 넘을 수 있다. 그래도 SUBMITTING 은 만료
	// 뒤 최종 검사를 통과할 수 없다 — 그것이 이 불변식의 정확한 형태다.
	//
	// 실시계가 없는데 가족 활성화가 검증돼 있으면 거절한다(fail-closed). 생산 조립은
	// 언제나 시계를 넣고, 검증 안 된 시장(미선언·기존 경로)은 시각을 보지 않는다.
	family := cycle.proposals.forMarket(market).familyActivation()
	if family.Verified() && cycle.now == nil {
		return execgw.Outcome{}, errors.New("engine: four-family activation lifetime cannot be measured without a clock")
	}
	leaseCeiling, err := family.LeaseCeiling(cycle.clockNow(), 30*time.Second)
	if err != nil {
		return execgw.Outcome{}, err
	}
	reconciliation, err := cycle.gateway.ObserveStrategyEntryGate(ctx, strings.ToLower(string(market)), accepted.result.Lineage.Symbol)
	if err != nil {
		return execgw.Outcome{}, err
	}
	owner, err := cycle.dispatchOwner(ctx)
	if err != nil {
		return execgw.Outcome{}, err
	}
	admitted := cycle.firstLeg.admit(ctx, result)
	if admitted.Code != StrategyFirstLegAdmitted {
		// 수집 오류는 사슬째 올라간다(a112 5.2.2.2 J4 · 리뷰 수리) — 문구는 같다(Detail 이 곧 그 오류의 문구). 범위 거절 타입과 결함 원인의
		// 신원이 모두 보존됨.
		if admitted.cause != nil {
			return execgw.Outcome{}, fmt.Errorf("engine: first-leg admission %s: %w", admitted.Code, admitted.cause)
		}
		return execgw.Outcome{}, fmt.Errorf("engine: first-leg admission %s: %s", admitted.Code, admitted.Detail)
	}
	decision, err := cycle.journal.LookupDecision(ctx, admitted.Receipt.DecisionID)
	if err != nil || decision.Generation < 0 {
		return execgw.Outcome{}, errors.New("engine: authoritative Guardian decision generation unavailable")
	}
	// Journal generations are zero-based. Dispatch authority generations are
	// one-based so zero remains the fail-closed "unavailable" sentinel.
	guardianGeneration := uint64(decision.Generation) + 1
	// 위험 정책 세대는 **그 범위의** 번들에서 읽는다(a112 5.2.2.2 — 범위마다 번들이 따로다; 범위 하나면 오늘과 같은 번들).
	riskGeneration := uint64(0)
	if key, keyed := strategyOwnerKeyOf(result.Lineage); keyed {
		if bundle, scoped := cycle.risk.forMarket(market).forScope(key); scoped {
			riskGeneration = bundle.Generation()
		}
	}
	if riskGeneration == 0 {
		return execgw.Outcome{}, errors.New("engine: signed risk policy generation unavailable")
	}
	activationGeneration := schedule.restore.Activation.Generation()
	activationExpiresAt := schedule.restore.Activation.ExpiresAt()
	now := cycle.schedule.observedAt
	if activationGeneration == 0 || activationExpiresAt.IsZero() || now.IsZero() || !now.Before(activationExpiresAt) || cycle.revalidateSchedule == nil {
		return execgw.Outcome{}, errors.New("engine: signed activation lifetime or final revalidator unavailable")
	}
	evidence := journal.StrategyDispatchVerifiedEvidence{Market: journal.StrategyDispatchMarket(market),
		ActivationGeneration: activationGeneration, ActivationDigest: schedule.snapshot.ActivationManifestDigest,
		CalendarGeneration: schedule.desired.Revision, CalendarDigest: schedule.calendar.Version,
		ProtectionGeneration: protection.Generation(), ProtectionSerial: strconv.FormatUint(protection.Generation(), 10), ProtectionDigest: protection.Digest(),
		ReconciliationGeneration: reconciliation.Generation(), ReconciliationDigest: reconciliation.Digest(),
		RiskPolicyGeneration: riskGeneration, GuardianGeneration: guardianGeneration, BuildDigest: strategyRuntimeBuildDigest()}
	ttl := min(leaseCeiling, activationExpiresAt.Sub(now))
	lease, err := cycle.journal.IssueVerifiedFirstLegStrategyDispatchLease(ctx, journal.VerifiedFirstLegStrategyDispatchLeaseRequest{
		Receipt: admitted.Receipt, Owner: owner, Evidence: evidence, TTL: ttl,
	})
	if err != nil {
		return execgw.Outcome{}, err
	}
	claimed, err := cycle.journal.ClaimStrategyDispatchLease(ctx, journal.StrategyDispatchLeaseCAS{
		LeaseID: lease.LeaseID, ExpectedRevision: lease.Revision, OwnerEpoch: owner.Epoch, FencingToken: owner.FencingToken,
	})
	if err != nil {
		return execgw.Outcome{}, err
	}
	intent, err := strategyFirstLegPlaceIntent(accepted, admitted.Receipt.QFinal)
	if err != nil {
		return execgw.Outcome{}, err
	}
	return cycle.gateway.PlaceClaimedStrategy(ctx, execgw.StrategyPlaceRequest{Intent: intent,
		Decision: execgw.GuardianDecision{ID: admitted.Receipt.DecisionID},
		Lease: journal.StrategyDispatchLeaseCAS{LeaseID: claimed.LeaseID, ExpectedRevision: claimed.Revision,
			OwnerEpoch: owner.Epoch, FencingToken: owner.FencingToken},
		FXAuthority: fx.read.evidence, IntentID: admitted.Receipt.AttemptID,
		EntryGateAuthority: reconciliation,
		FinalAuthorityCheck: func(checkCtx context.Context) error {
			// 스케줄 재검증이 **먼저**다. 그 재검증은 달력을 다시 읽는 I/O 이므로, 가족 만료를
			// 앞에서 보면 그 I/O 동안 만료가 지나가도 이 클로저가 통과를 돌려준다(Codex 재리뷰
			// P1). 가족 만료는 이 클로저가 돌려주기 **직전**, 브로커 바이트 전에 본다 (태스크 8.7.2).
			if err := cycle.revalidateSchedule(checkCtx, market, schedule); err != nil {
				return err
			}
			_, err := family.LeaseCeiling(cycle.clockNow(), 30*time.Second)
			return err
		}})
}

// clockNow 는 실시계의 지금이다. 시계가 없으면 영값을 준다 — 그 값은 검증 안 된 가족
// 활성화(시각을 보지 않는다)에만 닿는다. 검증된 활성화가 시계 없이 여기 오는 길은
// dispatch 의 첫 가족 검사가 막는다.
func (cycle *strategyDispatchCycle) clockNow() time.Time {
	if cycle.now == nil {
		return time.Time{}
	}
	return cycle.now()
}

func (cycle *strategyDispatchCycle) dispatchOwner(ctx context.Context) (journal.StrategyDispatchOwner, error) {
	if cycle.owner == nil {
		return journal.StrategyDispatchOwner{}, errors.New("engine: strategy dispatch owner coordinator unavailable")
	}
	cycle.owner.mu.Lock()
	defer cycle.owner.mu.Unlock()
	if cycle.owner.owner.Epoch != 0 {
		return cycle.owner.owner, nil
	}
	owner, err := cycle.journal.AcquireStrategyDispatchOwner(ctx, "paired-kr-us-strategy-runtime")
	if err != nil {
		return journal.StrategyDispatchOwner{}, err
	}
	cycle.owner.owner = owner
	return owner, nil
}

func strategyFirstLegPlaceIntent(accepted strategyFirstLegAccepted, quantity uint64) (orderintent.PlaceIntent, error) {
	if quantity == 0 || quantity > 1<<53 {
		return orderintent.PlaceIntent{}, errors.New("engine: strategy quantity is outside exact Gateway boundary")
	}
	priceText, ok := accepted.result.ExecutionTerms.Entry().MajorDecimal()
	if !ok {
		return orderintent.PlaceIntent{}, errors.New("engine: strategy entry price is invalid")
	}
	price, err := strconv.ParseFloat(priceText, 64)
	if err != nil || price <= 0 {
		return orderintent.PlaceIntent{}, errors.New("engine: strategy entry price cannot reach Gateway")
	}
	roundTrip, normalized := journal.NormalizeDecimal(strconv.FormatFloat(price, 'f', -1, 64))
	if !normalized || roundTrip != priceText {
		return orderintent.PlaceIntent{}, errors.New("engine: strategy entry price loses precision at Gateway boundary")
	}
	return orderintent.NormalizePlace(orderintent.PlaceInput{Symbol: accepted.result.Lineage.Symbol,
		Market: strings.ToLower(string(accepted.market)), Side: "buy", OrderType: "limit",
		Quantity: float64(quantity), Price: price, CurrencyMode: accepted.currency})
}

var _ strategyDispatchGateway = (*execgw.Gateway)(nil)
