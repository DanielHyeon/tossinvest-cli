package strategyworker

// a112 8.2(Manager 판정 2026-10-04) — 레인 · 증거 · worker 패키지 전부를 공용 걸음 넷(생산 · 시험 이진 × 무태그 · 태그)과 공용 금지
// 목록으로 한 번에 잰다. 패키지마다 있는 기존 가드는 직접 import 만 보거나 · 금지 목록이 좁거나 · 태그 뒤 시험을 안 걸었다(8.2 census
// `analysis/measurements/gate-8.1-8.3-2026-10-04/guard-census-8.2.md`). 이 표는 그 셋을 한 규칙으로 닫는다.
//
// 예외는 이름과 이유로만 연다(조용한 건너뛰기 금지): 시험 이진에 한해 journal 을 들여오는 패키지 둘 — 외부 시험 패키지가 임시 디렉터리의
// 실원장 픽스처(journal.Open)를 써서 적재기가 진짜 스키마를 읽는지 잰다(strategyrouter a127_real_journal_test.go ·
// strategycoordinator receipt_contract_test.go, 둘 다 태그 뒤). 생산 걸음에는 예외가 없다. officialbars · strategyproposal ·
// strategyprojection 은 자기 패키지에 전용 가드가 있다(official · 읽기 전용 journal 의 쓰임새를 심볼로 재야 해서).

import (
	"sync"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/testenv"
)

var a112LaneEvidenceWorkerPackages = []string{
	"breakoutlane", "continuationlane", "reversallane", "weeklyvaluelane",
	"strategyflow", "strategyevidence", "strategyworker",
	"strategycoordinator", "strategyarbiter", "strategyhandoff", "strategyrouter",
}

// a112TestOnlyJournalFixture 는 시험 이진(태그 걸음)에서만 journal 을 허용하는 패키지다 — 이유는 파일 머리.
var a112TestOnlyJournalFixture = map[string]bool{"strategyrouter": true, "strategycoordinator": true}

func TestEveryLaneEvidenceAndWorkerPackageReachesNoMutationCapability(t *testing.T) {
	var mu sync.Mutex
	usedException := map[string]bool{}
	// 걸음 44 개를 병렬로 — 묶음(t.Run "walks")이 자식이 다 끝날 때까지 기다리므로 아래 예외 사용 검사는 전부 끝난 뒤에 돈다.
	t.Run("walks", func(t *testing.T) {
		for _, name := range a112LaneEvidenceWorkerPackages {
			for _, mode := range testenv.WalkModes() {
				t.Run(name+"/"+mode.Name, func(t *testing.T) {
					t.Parallel()
					self := testenv.ModulePath + "internal/" + name
					graph := testenv.ListDeps(t, "../"+name, mode)
					roots := graph.Roots(self)
					if len(roots) == 0 {
						t.Fatalf("the walk has no root for %s", self)
					}
					var allowed []string
					if mode.Tests && a112TestOnlyJournalFixture[name] {
						allowed = []string{"internal/journal"}
					}
					reached := graph.Reachable(roots, nil)
					if found := testenv.ForbiddenIn(reached, allowed...); len(found) != 0 {
						t.Fatalf("%s (%s) reaches a mutation capability:\n%v", name, mode.Name, found)
					}
					// journal 은 시험 이진에서 `journal [p.test]` 변형 이름으로 나올 수 있다 — 변형을 접어 세는 ForbiddenIn 으로 잰다.
					if len(allowed) != 0 && len(testenv.ForbiddenIn(reached)) != 0 {
						mu.Lock()
						usedException[name] = true
						mu.Unlock()
					}
				})
			}
		}
	})
	// 예외가 아직 필요한지 잰다 — 픽스처가 사라졌는데 예외만 남으면 그 자리가 조용한 구멍이 된다.
	for name := range a112TestOnlyJournalFixture {
		if !usedException[name] {
			t.Errorf("the test-only journal exception for %s is no longer exercised by any walk — remove it", name)
		}
	}
}
