package strategyrouter

// ScorePPMMax 는 중재 점수의 상한이다.
//
// 숫자를 여기에 다시 적지 않고 매니페스트 검증이 쓰는 상수를 그대로 유도한다.
// 같은 값을 두 곳에 적어 두면 언젠가 한쪽만 고쳐지고, 그때 두 곳은 조용히
// 서로 다른 계약을 말하게 된다.
const ScorePPMMax = productionRouteScorePPMMax

// Known 은 이 값이 정확히 네 전략군 중 하나인지 알려준다.
// 열거에 없는 이름은 "모르는 가족"이고, 모르는 가족의 점수는 견줄 수 없다.
func (family Family) Known() bool {
	switch family {
	case FamilyContinuation, FamilyReversal, FamilyWeeklyValue, FamilyBreakoutRetest:
		return true
	}
	return false
}
