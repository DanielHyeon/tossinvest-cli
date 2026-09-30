package engine

// UnobservedRecordsForTest 는 관측자가 들고 있는 a090 미관측 기록 수를 돌려줌. TESTS ONLY — 보유가 끝난 포지션의 기록이
// 정리되는지(메모리가 보유 집합을 따라가는지)는 원장·로그로 드러나지 않으므로 이 좁은 창으로만 잼.
func (o *ExitObserver) UnobservedRecordsForTest() int { return len(o.unobserved) }

// PendingModeNoticesForTest 는 적재 못 한 a090 강화 공지 대기열 길이임. TESTS ONLY — 창 0 기록은 같은 key 를 옛 행에 흡수하므로
// 「적재된 공지가 대기열에서 빠졌는가」 는 행 수로 드러나지 않음.
func (o *ExitObserver) PendingModeNoticesForTest() int { return len(o.pendingModeNotices) }
