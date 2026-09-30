package engine

import "github.com/JungHoonGhae/tossinvest-cli/internal/exitpolicy"

// a091 시험용 접근자 — _test.go 이므로 빌드된 바이너리에는 없음.

// IsProtectiveForTest 는 보호/익절 판정 술어를 돌려줌(tasks 3.4 — 주문 액션 전수 표). TESTS ONLY.
func IsProtectiveForTest(a exitpolicy.Action) bool {
	return isProtective(exitpolicy.Proposal{Action: a})
}

// CancellationOnlyForTest 는 종료 취소 억제의 판정 술어를 돌려줌(tasks 3.3b). TESTS ONLY.
func CancellationOnlyForTest(err error) bool { return cancellationOnly(err) }
