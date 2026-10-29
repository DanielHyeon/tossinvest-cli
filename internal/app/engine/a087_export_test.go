package engine

import "time"

// SetA087LowerLimitForTest 는 종단 시험에만 하한가 판 항목을 넣음.
// _test.go 전용이므로 생산 admission nil 경로를 바꾸지 않음.
func SetA087LowerLimitForTest(options *ExitObserverOptions, market, symbol, price string, timestamp time.Time) {
	if options == nil {
		return
	}
	board := newExitLowerLimitBoard(options.Clock, nil, nil, nil, options.AccountRef, nil)
	board.publish(lowerLimitSnapshot{records: map[string]lowerLimitRecord{
		lowerLimitKey(market, symbol): {price: price, timestamp: timestamp},
	}})
	options.lowerLimitBoard = board
}

// A087LowerLimitAttemptedForTest 는 Place 뒤 첫 시도 금지 표시를 읽음.
func A087LowerLimitAttemptedForTest(observer *ExitObserver, market, symbol string, now time.Time) bool {
	if observer == nil {
		return false
	}
	board := observer.opts.lowerLimitBoard
	return board != nil && board.attemptedOn(market, symbol, now)
}
