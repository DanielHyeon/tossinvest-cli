package testenv_test

import (
	"strings"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/testenv"
)

// 금지 목록은 스펙이 부른 네 능력을 모두 이름으로 갖는다 — 하나가 빠지면 모든 폐포 가드가 함께 그 능력을 못 본다(a112 8.2).
func TestTheCapabilityListNamesEveryForbiddenCapability(t *testing.T) {
	want := map[string]bool{
		"internal/official": true, "internal/hybrid": true, "internal/client": true, "internal/trading": true, // broker mutator
		"internal/journal": true,                                    // writable journal
		"internal/execgw":  true, "internal/strategydispatch": true, // Guardian issuer
		"internal/ops": true, "internal/config": true, // activation/toggle writer
	}
	seen := map[string]bool{}
	for _, capability := range testenv.MutationCapabilities() {
		if capability.Why == "" {
			t.Errorf("%s has no reason — a guard failure must say what it nearly allowed", capability.Path)
		}
		seen[capability.Path] = true
	}
	for path := range want {
		if !seen[path] {
			t.Errorf("the capability list lost %s", path)
		}
	}
}

func TestCapabilityMatchesThePackageAndItsChildrenOnly(t *testing.T) {
	for path, want := range map[string]bool{
		testenv.ModulePath + "internal/journal":           true,
		testenv.ModulePath + "internal/journal/sub":       true,
		testenv.ModulePath + "internal/journalview":       false, // 접두어가 같은 다른 패키지
		testenv.ModulePath + "internal/strategyflow":      false,
		"internal/journal":                                false, // 모듈 경로 없는 이름
		testenv.ModulePath + "internal/app/engine":        true,
		testenv.ModulePath + "internal/officialfx":        false, // official 과 접두어만 같다
		testenv.ModulePath + "internal/strategydispatch":  true,
		testenv.ModulePath + "internal/strategyevidencex": false,
	} {
		if _, got := testenv.Capability(path); got != want {
			t.Errorf("Capability(%q)=%v, want %v", path, got, want)
		}
	}
}

func TestForbiddenInReducesTestVariantsAndHonoursTheAllowList(t *testing.T) {
	journal := testenv.ModulePath + "internal/journal"
	official := testenv.ModulePath + "internal/official"
	got := testenv.ForbiddenIn(map[string]bool{
		journal: true, journal + " [x.test]": true, official: true, testenv.ModulePath + "internal/strategyflow": true,
	}, "internal/official")
	if len(got) != 1 || !strings.HasPrefix(got[0], journal+" — ") {
		t.Fatalf("ForbiddenIn = %q, want only %s (once, the test variant folded in; official allowed)", got, journal)
	}
}

func TestReachableStopsAtTheCut(t *testing.T) {
	graph := testenv.ImportGraph{"a": {"b", "c"}, "b": {"d"}, "c": {"e"}, "d": nil, "e": nil}
	got := graph.Reachable([]string{"a"}, map[string]bool{"b": true})
	if !got["a"] || !got["c"] || !got["e"] || got["b"] || got["d"] {
		t.Fatalf("Reachable = %v, want a, c, e (b cut, so d unreachable)", got)
	}
}

// 걸음 넷이 실제로 돈다 — 이 패키지 자신의 걸음에서 시험 변형과 태그 걸음이 모두 무언가를 읽는다(빈 표본 통과 금지).
func TestEveryWalkModeReadsTheTestBinaryOrTheProductionClosure(t *testing.T) {
	self := testenv.ModulePath + "internal/testenv"
	for _, mode := range testenv.WalkModes() {
		graph := testenv.ListDeps(t, ".", mode)
		roots := graph.Roots(self)
		if len(roots) == 0 {
			t.Fatalf("%s: the walk has no root for %s", mode.Name, self)
		}
		if mode.Tests && len(roots) < 2 {
			t.Fatalf("%s: roots %v carry no test variant — -test did not reach the test binary", mode.Name, roots)
		}
	}
}
