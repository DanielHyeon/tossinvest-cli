package engine

// FollowedExitOrdersForTest 는 관측자가 체결을 따라가는 청산 주문 수임(a129). TESTS ONLY — 따라가기 항목은 메모리에만 있어
// 「전용 로거가 없으면 항목도 원장 읽기도 없다」 · 「종결되면 항목이 지워진다」 는 로그·원장으로 드러나지 않으므로 이 좁은 창으로 잼.
func (o *ExitObserver) FollowedExitOrdersForTest() int { return len(o.exitFills) }
