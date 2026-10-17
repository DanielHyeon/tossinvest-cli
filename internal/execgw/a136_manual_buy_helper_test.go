package execgw_test

import (
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
)

// a136ZeroManualBuy 는 「엔진 밖 수동 매수 0」 스냅숏(그 순간 신선) — 수동 매수를 재지 않는 시험의 필수 칸.
func a136ZeroManualBuy(at time.Time) *journal.ManualBuyTotal {
	day, err := clock.AccountTradingDay(at)
	if err != nil {
		panic(err)
	}
	return &journal.ManualBuyTotal{Amount: "0", AsOf: at, Day: day}
}

// a136ZeroManualBuyBoard 는 시계의 현재 순간에 「수동 매수 0」 을 게시한 보드. 시계를 2분 넘게 돌리는 시험은 다시 게시해야 함.
func a136ZeroManualBuyBoard(clk interface{ Now() time.Time }) *execgw.ManualBuyBoard {
	board := execgw.NewManualBuyBoard()
	now := clk.Now()
	day, err := clock.AccountTradingDay(now)
	if err != nil {
		panic(err)
	}
	board.Publish(execgw.ManualBuySnapshot{AsOf: now, Day: day, ByCurrency: map[string]string{}})
	return board
}
