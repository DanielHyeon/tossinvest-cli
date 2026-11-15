package official

import "testing"

// a189: M-B0 참조 허용은 a166 런타임 시험 파일 하나의 정확 경로뿐임을 고정함 —
// 같은 디렉터리의 다른 파일, 같은 이름을 다른 자리에 둔 파일은 여전히 위반으로 셈(접두사·접미사로 넓히는 변경 차단).
func TestA189MBUSGuardAdmitsOnlyTheExactA166RuntimeTestFile(t *testing.T) {
	source := `package accountbudgetruntime
import "github.com/JungHoonGhae/tossinvest-cli/internal/official"
func read() { _, _ = official.A112MBUSCandle(nil, nil, nil) }
`
	if got := a112MBUSFindForbiddenRefs("internal/accountbudgetruntime/v2_token_test.go", source); len(got) != 0 {
		t.Fatalf("the admitted a166 runtime test file produced violations: %#v", got)
	}
	for _, path := range []string{
		"internal/accountbudgetruntime/runtime_test.go",
		"internal/accountbudgetruntime/v2_token.go",
		"internal/other/internal/accountbudgetruntime/v2_token_test.go",
		"tools/other/v2_token_test.go",
	} {
		if got := a112MBUSFindForbiddenRefs(path, source); len(got) != 1 || got[0] != "selector official.A112MBUSCandle" {
			t.Errorf("%s: violations = %#v, want exactly the selector violation", path, got)
		}
	}
}
