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

// deliverEachStrategyHandoff 는 handoff 들을 조정자 순서대로 하나씩 몸통에 건넴(태스크 5.2.2.1 · 5.2.2.2).
//
// **범위 거절만 건너뜀(a112 5.2.2.2 — Manager 판정 J4).** 몸통의 오류가 `*strategyScopeRefusal`(그 소유자 범위의 위험 · 계좌 권한 부재 —
// 타입으로 분류, 문구로 가르지 않음)이면 그 범위만 거절하고 다음 범위로 간다 — 한 범위의 결손이 다른 범위를 굶기지 않게. 건너뛴 거절은
// 각각 모아 **끝에 함께 돌려준다**(기록 · 가시 — 조용한 건너뛰기 금지; 감독자가 사이클 오류로 센다). 그 밖의 오류(원장 · Gateway · 중앙
// 무결성 · identity 불일치)는 **첫 오류에서 멈춤** — 예상 밖 오류 뒤에 같은 주기에서 주문을 더 내지 않는 보수 방향. 굶음 결정: 범위 거절이
// 아닌 오류로 앞 범위가 계속 실패하면 뒤 범위는 여전히 굶는다 — 그 오류는 범위의 것이 아니라 주기의 것이므로 의도적으로 남긴다.
// handoff 가 하나이고 범위 거절이 아닌 오류면 그 값을 감싸지 않고 돌려줌(오류 분류 보존).
//
// 거절된 handoff 는 Deliver 가 몸통을 부르지 않고 nil 을 돌려주므로 다음 범위로 넘어감(거절은 오류가 아님, Deliver 머리말).
func deliverEachStrategyHandoff(handoffs []strategyhandoff.Handoff, body func(strategyhandoff.Delivered) error) error {
	var skipped []error
	for _, handoff := range handoffs {
		err := handoff.Deliver(body)
		if err == nil {
			continue
		}
		var scope *strategyScopeRefusal
		if errors.As(err, &scope) {
			skipped = append(skipped, err)
			continue
		}
		// 멈추는 오류 앞에서 건너뛴 범위 거절도 잃지 않음(J4 ③ — 각각 기록). 건너뛴 것이 없으면 값을 감싸지 않음(분류 보존).
		if len(skipped) == 0 {
			return err
		}
		return errors.Join(append(skipped, err)...)
	}
	return errors.Join(skipped...)
}
