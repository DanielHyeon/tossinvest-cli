package strategyhandoff

// a112 8.2(Manager 판정 2026-10-04) — 태그 뒤 시험 파일까지 걷는다. 이 패키지의 기존 폐포 가드(dependency_closure_test.go)는
// `go list` 를 태그 없이 불러 `//go:build tossos_testseams` 시험 파일과 그것들이 들여오는 것을 보지 못했다(8.2 census). 여기서는
// 공용 걸음(testenv.WalkModes — 생산 · 시험 이진 × 무태그 · 태그)과 공용 금지 목록(testenv.MutationCapabilities — 기존 목록에 없던
// internal/strategydispatch 포함)으로 같은 약속을 다시 잰다. 기존 가드는 그대로 둔다(그 목록은 이 패키지에 적힌 판단 기록이다).

import (
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/testenv"
)

func TestTheTaggedTestClosureAlsoReachesNoMutationCapability(t *testing.T) {
	self := testenv.ModulePath + "internal/strategyhandoff"
	for _, mode := range testenv.WalkModes() {
		t.Run(mode.Name, func(t *testing.T) {
			graph := testenv.ListDeps(t, ".", mode)
			roots := graph.Roots(self)
			if len(roots) == 0 {
				t.Fatalf("the walk has no root for %s", self)
			}
			reached := graph.Reachable(roots, nil)
			if found := testenv.ForbiddenIn(reached); len(found) != 0 {
				t.Fatalf("%s: the closure reaches a mutation capability:\n%v", mode.Name, found)
			}
			// 양성 대조: 걸음이 두 단계 넘게 내려갔다(candidate · domain 은 strategyflow 를 거쳐야 나온다).
			for _, required := range []string{"internal/strategyflow", "internal/candidate", "internal/domain"} {
				if !reached[testenv.ModulePath+required] {
					t.Fatalf("%s: the walk did not reach %s — too shallow to prove anything", mode.Name, required)
				}
			}
			if mode.Tests && len(roots) < 2 {
				t.Fatalf("%s: roots %v carry no test variant — -test did not reach the test binary", mode.Name, roots)
			}
		})
	}
}
