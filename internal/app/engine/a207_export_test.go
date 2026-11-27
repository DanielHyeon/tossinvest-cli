package engine

// RefusalStreakForTest 는 포지션 하나의 연속 해제 셈임(a207 D5). TESTS ONLY — _test.go 라 빌드된 바이너리에는 없음.
func (o *ExitObserver) RefusalStreakForTest(positionID string) int {
	if s := o.refusalStreaks[positionID]; s != nil {
		return s.count
	}
	return 0
}
