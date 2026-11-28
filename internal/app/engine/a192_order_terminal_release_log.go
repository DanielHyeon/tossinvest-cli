package engine

// a192_order_terminal_release_log.go 는 missing-order 종목 차단이 종결 증명으로 풀린 것을 운영자가 로그에서 볼 수
// 있게 함(a192 design D7, M6 — 로그만, 알림 없음).
//
// 싣는 것은 종목·시장·해제 원인·결속 주문 수뿐임. 계좌번호·주문번호·수량·가격은 싣지 않음 — 그 값들은 장부 행의
// 근거에 있고, 로그는 「무엇이 언제 풀렸나」만 말하면 됨.

import (
	"sort"

	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/obs"
	"github.com/JungHoonGhae/tossinvest-cli/internal/reconcile"
)

// eventReconcileBlockReleased 는 종결 증명 해제 한 건의 로그 이벤트임. a191 선례대로 obs 패키지를 고치지 않고
// 이 자리의 리터럴로 둠(등급표에 없으므로 normal).
const eventReconcileBlockReleased obs.EventType = "reconcile.block_released"

// eventReconcileBlockReleaseHeld 는 결속된 차단의 판정 사유가 바뀔 때 한 줄 — 유지된 이유(M5)를 운영자가 봄.
const eventReconcileBlockReleaseHeld obs.EventType = "reconcile.block_release_held"

// logOrderTerminalReleases 는 이번 관측이 커밋한 해제 중 종결 증명 해제마다 한 줄을 씀.
func (d *ReconcileDriver) logOrderTerminalReleases(outcome reconcile.Outcome) {
	if d.opts.Log == nil {
		return
	}
	for _, block := range outcome.Cleared {
		bound, ok := block.ReleasedByOrderTerminal()
		if !ok {
			continue
		}
		d.opts.Log.Event(eventReconcileBlockReleased,
			obs.FieldSymbol, block.Symbol,
			"market", block.Market,
			"release_cause", journal.ReconcileReleaseOrderTerminalConfirmed,
			"bound_orders", bound)
	}
	if d.opts.Tracker == nil {
		return
	}
	// 유지 사유는 바뀔 때만 씀. 결속이 사라진 종목은 기억에서 지움(다시 생기면 다시 씀).
	reasons := d.opts.Tracker.OrderTerminalHoldReasons()
	if d.terminalHoldLogged == nil {
		d.terminalHoldLogged = map[string]string{}
	}
	logged := d.terminalHoldLogged
	symbols := make([]string, 0, len(reasons))
	for symbol := range reasons {
		symbols = append(symbols, symbol)
	}
	sort.Strings(symbols)
	for _, symbol := range symbols {
		if logged[symbol] == reasons[symbol] {
			continue
		}
		logged[symbol] = reasons[symbol]
		d.opts.Log.Event(eventReconcileBlockReleaseHeld, obs.FieldSymbol, symbol, "reason", reasons[symbol])
	}
	for symbol := range logged {
		if _, ok := reasons[symbol]; !ok {
			delete(logged, symbol)
		}
	}
}
