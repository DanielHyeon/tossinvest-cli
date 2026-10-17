package engine

import "github.com/JungHoonGhae/tossinvest-cli/internal/execgw"

// a092 배선 시험용 접근자 — _test.go 이므로 빌드된 바이너리에는 없음.

// OptionsForTest 는 조립이 끝난 exit 관측 루프의 옵션을 돌려줌. TESTS ONLY.
func (o *ExitObserver) OptionsForTest() ExitObserverOptions { return o.opts }

// ExitFloorRetrierForTest 는 청산 수량 상한 공급자가 쓰는 Retrier 를 돌려줌(공급자가 reconcileFloor 가 아니면 nil, false). TESTS ONLY.
func ExitFloorRetrierForTest(f FloorSource) (*execgw.Retrier, bool) {
	rf, ok := f.(*reconcileFloor)
	if !ok || rf == nil {
		return nil, false
	}
	return rf.retrier, true
}

// ContextFloorForTest 는 엔진이 조립한 공유 청산 상한 공급자를 돌려줌. TESTS ONLY.
func (c *Context) ContextFloorForTest() FloorSource { return c.exitFloor }

// FloorsShareAllButRetrierForTest 는 두 공급자가 Retrier 만 다르고 나머지 필드가 같은지 봄. TESTS ONLY.
func FloorsShareAllButRetrierForTest(a, b FloorSource) bool {
	fa, ok1 := a.(*reconcileFloor)
	fb, ok2 := b.(*reconcileFloor)
	if !ok1 || !ok2 || fa == nil || fb == nil {
		return false
	}
	ca, cb := *fa, *fb
	ca.retrier, cb.retrier = nil, nil
	return ca == cb
}
