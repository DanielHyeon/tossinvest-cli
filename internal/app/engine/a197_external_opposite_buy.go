package engine

// a197_external_opposite_buy.go 는 손절을 막는 외부 매수 미체결 취소 단계임(골격 — a197 3.1 RED).

import (
	"context"

	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/exitpolicy"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
)

// externalBuyItem 은 주기 끝에 처리할 「열 조건」 하나임.
type externalBuyItem struct {
	positionID  string
	instanceSeq int64
	market      string
	symbol      string
}

func (o *ExitObserver) noteExternalBuyBlock(m managed) {}

func (o *ExitObserver) runExternalBuyStage(ctx context.Context) {}

// externalBuyItemCurrent 는 실행 직전 재검증임(골격).
func externalBuyItemCurrent(item externalBuyItem, positions []journal.Position) bool { return true }

// A197AttemptKeyPrefix · A197UnclearedKeyPrefix 는 단계 알림 key 의 접두임(골격).
const (
	A197AttemptKeyPrefix   = "exit.liquidation_delayed|a197-attempt|"
	A197UnclearedKeyPrefix = "exit.liquidation_delayed|a197-uncleared|"
)

// externalBuyOpens 는 열 조건 판정임(골격).
func externalBuyOpens(p exitpolicy.Proposal, out execgw.Outcome) bool { return false }
