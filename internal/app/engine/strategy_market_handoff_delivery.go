package engine

import (
	"context"
	"errors"

	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyhandoff"
)

// strategyCampaignCASReader 는 소유자 범위의 포지션 캠페인 CAS 를 읽는 좁은 문(생산 = `*journal.Journal`).
type strategyCampaignCASReader interface {
	CurrentPositionCampaignCAS(ctx context.Context, accountRef, market, symbol string) (journal.PositionCampaignCASRead, error)
}

// strategyHandoffDispatcher 는 봉투 하나를 공유 dispatch 에 넘기는 좁은 문(생산 = `*strategyDispatchCycle`).
type strategyHandoffDispatcher interface {
	dispatch(ctx context.Context, delivered strategyhandoff.Delivered) (execgw.Outcome, error)
}

// dispatchStrategyMarketHandoffs 는 한 시장 주기의 handoff 들을 공유 dispatch 에 건네는 **유일한** 생산 자리임.
//
// 2026-09-30 리뷰(codex P1-1 · 보이스 B #2) 수리로 `runProductionStrategyMarketCycle` 의 closure 를 **의미 무변경으로**
// 옮겨 온 함수임. 옮긴 이유는 하나: 원래 자리는 권한 새로 고침 전체가 필요해 어떤 시험도 몸통을 돌지 못했고, 그 자리를 지키던
// 구조 못은 이름 모양만 봐서 「둘째 범위부터 조용히 버림」(몸통 closure 안 카운터) · 이름만 같은 메서드 · 지역 섀도 셋에
// 뚫렸음. 이제 몸통을 스파이로 직접 돌리고(a112_owner_scope_delivery_test.go), 부르는 자리와 몸통의 모양은
// 식별자 해소(go/types)로 못 박음(a112_market_delivery_structure_test.go).
//
// 옮기기 전후 몸통이 같은 코드라는 영수증은 `analysis/measurements/lot-5.6.2-5.2.2/extract-receipt-5.2.2.1-fix.txt`
// (gofmt 정규형 비교 — 수신자 둘의 이름만 다름: `c.Journal` → campaigns, `fresh.dispatch` → dispatcher).
func dispatchStrategyMarketHandoffs(ctx context.Context, campaigns strategyCampaignCASReader, dispatcher strategyHandoffDispatcher,
	handoffs []strategyhandoff.Handoff,
) error {
	return deliverEachStrategyHandoff(handoffs, func(delivered strategyhandoff.Delivered) error {
		lineage := delivered.Result().Lineage
		cas, err := campaigns.CurrentPositionCampaignCAS(ctx, lineage.AccountRef, string(lineage.Market), lineage.Symbol)
		if err != nil {
			return err
		}
		if cas.Claimed || cas.State != "FLAT" && cas.State != "CLOSED" {
			return nil
		}
		_, err = dispatcher.dispatch(ctx, delivered)
		if errors.Is(err, journal.ErrStrategyDispatchLeaseConsumed) {
			return nil
		}
		return err
	})
}
