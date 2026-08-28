//go:build tossos_testseams

package strategyflow

// LaneInputMatchesForTest 는 태그 없는 LaneInput 을 밖에서 확인하기 위한 시험용 이음매다.
// LaneInput 은 비교 불가능한 필드를 담고 있어 제로값 비교가 불가능하기 때문이다.
// 이 파일은 tossos_testseams 태그 없이는 빌드되지 않으므로 생산 코드에 들어가지 않는다.
func LaneInputMatchesForTest(input LaneInput, descriptor Descriptor) bool {
	return input.matches(descriptor)
}
