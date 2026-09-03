package engine

import (
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyarbiter"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategycoordinator"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyproposal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyrouter"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyworker"
)

// strategyMarketArbitration 은 한 시장을 조정한 결과다.
//
// outcome 은 어느 범위가 무엇을 골랐는지(또는 왜 닫혔는지)를 들고 있고,
// byIdentity 는 그 선택을 다시 이 프로세스의 레인 권한으로 되돌리는 색인이다.
// 조정자는 봉인된 계보 신원만 건네주므로 되돌리는 일은 여기서 한다 —
// 조정자가 엔진 자료구조를 들고 다니면 그 순간 순수하지 않게 된다.
type strategyMarketArbitration struct {
	outcome    strategycoordinator.Outcome
	byIdentity map[string]strategyProposalEntryAuthority
	collision  bool
	// gated 는 4-가족 관문이 조정자 앞에서 멈춘 제안들의 결과 종류다.
	//
	// 수만 세지 않고 종류를 담는 이유: DORMANT(안 켰다) · LATCHED(고장으로
	// 닫혔다) · REFUSED(주인 없다)는 운영자가 할 조치가 전부 다르다. 하나로
	// 뭉치면 복구 증거가 필요한 상태가 "아직 안 켰다" 로 보인다.
	gated []strategyworker.Outcome
}

// coordinateMarketProposals 는 한 시장의 모든 가족 제안을 조정자에 넣고 중재까지 마친다.
//
// 아무것도 여기서 고르지 않는다. 고르는 규칙은 strategyarbiter 한 곳에,
// 받고 접고 줄 세우는 규칙은 strategycoordinator 한 곳에 있다. 이 함수는
// 그 둘에 필요한 값을 모아 주고 결과를 되돌릴 색인을 만드는 일만 한다.
//
// 기대 범위의 종목은 경로 권한이 아니라 승인된 후보에서 읽는다. 권한이 스스로
// 말한 종목을 그 권한을 검사하는 데 다시 쓰면, 어긋남을 잡아내려던 검사가
// 언제나 참이 되어 아무것도 잡지 못한다.
func coordinateMarketProposals(accountRef string, market StrategyMarket, routes []strategyRouteEntryAuthority,
	batch strategyproposal.ProductionBatchAuthority, observedAt time.Time, gate strategyFamilyGate,
) (strategyMarketArbitration, int) {
	routerMarket := strategyRouterMarket(market)
	coordinator := strategycoordinator.NewMarketCoordinator(routerMarket, observedAt)
	arbitration := strategyMarketArbitration{byIdentity: make(map[string]strategyProposalEntryAuthority, batch.Len())}
	refused := 0
	for _, route := range routes {
		lanes := batch.LanesFor(route.approved.Symbol())
		if len(lanes) == 0 {
			refused++
			continue
		}
		scope := strategyrouter.OwnerKey{AccountRef: accountRef, Market: routerMarket,
			Symbol: route.approved.Symbol(), PositionGeneration: route.route.Request().Key.PositionGeneration}
		for _, lane := range lanes {
			result := lane.Proposal()
			// 스냅샷 다이제스트는 레인 권한이 들고 있는 값을 그대로 옮긴다.
			// 이 값은 봉인된 계보 안에 없으므로 여기서 건네야 한다.
			envelope := strategycoordinator.Envelope{Scope: scope,
				SnapshotDigest: lane.SnapshotDigest(),
				Proposal:       strategyarbiter.Proposal{Result: result, Authority: route.route}}
			// 이 가족의 레인이 이 제안을 조정자로 보내는가 (태스크 5.1.2.2).
			//
			// 관문이 서지 않은 시장에서는 아무것도 묻지 않고 위 봉투가 그대로
			// 들어간다 — 그것이 오늘의 동작이고, 서명된 4-가족 활성화가 없는
			// 시장의 값이다. 선 시장에서는 레인이 만든 봉투가 들어가고,
			// 잠긴 레인의 가족은 여기서 멈춘다.
			// `else if` 를 쓰지 않는다. Go AST 는 `} else if` 를 **같은 좌표의
			// else 와 if 두 노드**로 내므로, 좌표로 분기를 세는 이 change 의
			// 열거표에서 두 줄이 구별되지 않는다.
			admitted, outcome, ok := gate.admit(strategyworker.Input{Scope: scope,
				SnapshotDigest: lane.SnapshotDigest(), Proposal: envelope.Proposal})
			if !ok {
				arbitration.gated = append(arbitration.gated, outcome)
				continue
			}
			if outcome == strategyworker.OutcomeEmitted {
				envelope = admitted
			}
			admission := coordinator.Submit(envelope)
			if !admission.Admitted {
				// 거절은 조정자가 이미 붙들고 있다. 여기서 되돌릴 색인만 만들지 않는다.
				continue
			}
			if _, exists := arbitration.byIdentity[result.Lineage.Identity]; exists {
				// 같은 계보 신원이 두 레인에 있으면 선택을 되돌릴 곳이 하나로
				// 정해지지 않는다. 아무거나 고르지 않고 닫는다.
				//
				// 중재까지 가지 않고 여기서 돌아가므로 Outcome 은 비어 있다.
				// 그래도 접힌 봉투 수는 조정자가 이미 세어 두었으니 옮긴다 —
				// 안 옮기면 이 경로만 "유실 0" 이라고 거짓으로 보고한다.
				arbitration.collision = true
				arbitration.outcome.Drops = coordinator.Drops()
				return arbitration, refused
			}
			arbitration.byIdentity[result.Lineage.Identity] = strategyProposalEntryAuthority{route: route, authority: lane}
		}
	}
	arbitration.outcome = coordinator.Arbitrate()
	return arbitration, refused
}

// entries 는 조정자가 고른 순서 그대로 레인 권한 목록을 만든다.
// 신원을 되돌리지 못하면 false 다 — 못 찾은 자리를 건너뛰면 그 종목이
// 조용히 사라진 채 목록만 짧아진다.
func (arbitration strategyMarketArbitration) entries() ([]strategyProposalEntryAuthority, bool) {
	values := make([]strategyProposalEntryAuthority, 0, len(arbitration.outcome.Selections))
	for _, selection := range arbitration.outcome.Selections {
		entry, ok := arbitration.byIdentity[selection.LineageIdentity]
		if !ok {
			return nil, false
		}
		values = append(values, entry)
	}
	return values, true
}
