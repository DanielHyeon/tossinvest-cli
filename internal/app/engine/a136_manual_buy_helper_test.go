package engine_test

import (
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
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

// a136ZeroManualBuyBoard 는 시계의 현재 순간에 「수동 매수 0」 을 게시한 보드.
func a136ZeroManualBuyBoard(clk interface{ Now() time.Time }) *execgw.ManualBuyBoard {
	board := execgw.NewManualBuyBoard()
	total := a136ZeroManualBuy(clk.Now())
	board.Publish(execgw.ManualBuySnapshot{AsOf: total.AsOf, Day: total.Day, ByCurrency: map[string]string{}})
	return board
}
