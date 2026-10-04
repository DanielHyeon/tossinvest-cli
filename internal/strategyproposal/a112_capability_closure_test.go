package strategyproposal

// a112 8.2(Manager 판정 2026-10-04): 제안 적재기는 아무것도 바꾸지 않는다 — 그 약속을 이 패키지에서 기계로 지킨다.
//
// 이 패키지는 4 능력 중 하나(internal/journal)를 **직접** 들여온다: 주간 가치 레인의 첫 레그 예약을 읽어야 하기 때문이다(production.go
// buildLaneInput). 그래서 「journal 이 폐포에 없다」 는 참일 수 없고, 대신 둘을 묶는다:
//   ① 폐포 걸음 넷(생산 · 시험 이진 × 무태그 · 태그)에 journal 말고 다른 능력 패키지가 없다.
//   ② journal 사용이 **읽기 전용 문 하나**뿐이다 — 타입 검사기가 해소한 객체로 센다(철자 아님): 패키지 수준 함수는 OpenReadOnly 하나,
//      메서드는 수신자가 ReadOnly(SELECT 전용 핸들 — journal/readonly.go) 인 것뿐, 타입 이름은 읽기 전용 · 값 타입 넷뿐. 쓰기 핸들
//      (`journal.Open` · `*journal.Journal` 의 메서드)이 한 번이라도 나타나면 실패한다.
// 둘 다 빈 표본에서 통과하지 않는다 — OpenReadOnly 와 ReadOnly 메서드가 각각 한 번 이상 보여야 한다.

import (
	"go/types"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/testenv"
)

const a112ProposalPath = testenv.ModulePath + "internal/strategyproposal"

func TestTheProposalClosureReachesNoMutationCapabilityButTheReadOnlyJournal(t *testing.T) {
	for _, mode := range testenv.WalkModes() {
		t.Run(mode.Name, func(t *testing.T) {
			graph := testenv.ListDeps(t, ".", mode)
			roots := graph.Roots(a112ProposalPath)
			if len(roots) == 0 {
				t.Fatalf("the walk has no root for %s", a112ProposalPath)
			}
			reached := graph.Reachable(roots, nil)
			// journal 은 ② 가 쓰임새를 못 박으므로 여기서 허용한다 — 다른 능력은 하나도 안 된다.
			if found := testenv.ForbiddenIn(reached, "internal/journal"); len(found) != 0 {
				t.Fatalf("the proposal closure (%s) reaches a mutation capability:\n%v", mode.Name, found)
			}
			// 양성 대조: 걸음이 깊었다(직접 import 가 아닌 candidate · domain 이 보인다) · journal 이 실제로 폐포에 있다(② 가 재는 대상).
			for _, required := range []string{"internal/strategyflow", "internal/candidate", "internal/domain", "internal/journal"} {
				if !reached[testenv.ModulePath+required] {
					t.Fatalf("%s: the walk did not reach %s — too shallow to prove anything", mode.Name, required)
				}
			}
		})
	}
}

func TestTheProposalUsesTheJournalOnlyThroughItsReadOnlyDoor(t *testing.T) {
	const journalPath = testenv.ModulePath + "internal/journal"
	// 태그 뒤 생산 파일(시험 seam)까지 센다 — 생산 빌드에 없는 파일이라도 쓰기 핸들을 쥐면 그 seam 을 쓰는 시험이 실원장을 쓸 수 있다.
	checked := testenv.TypeCheckProduction(t, ".", a112ProposalPath, "tossos_testseams")
	allowedTypes := map[string]bool{
		"ReadOnly":                         true, // SELECT 전용 핸들
		"ReadOnlyOptions":                  true, // 경로 하나
		"WeeklyFirstLegReservationBinding": true, // 값 구조체(제안 권한이 싣는 예약 결속)
	}
	openReadOnly, readOnlyMethods := 0, 0
	for _, use := range checked.UsesFrom(journalPath) {
		switch object := use.Object.(type) {
		case *types.Func:
			if receiver := testenv.ReceiverNamed(object); receiver != "" {
				if receiver != "ReadOnly" {
					t.Errorf("%s: uses %s — only methods of the read-only handle are allowed", use.Position, testenv.Describe(object))
				}
				readOnlyMethods++
				continue
			}
			if object.Name() != "OpenReadOnly" {
				t.Errorf("%s: uses %s — the only package-level journal function allowed is OpenReadOnly", use.Position, testenv.Describe(object))
			}
			openReadOnly++
		case *types.TypeName:
			if !allowedTypes[object.Name()] {
				t.Errorf("%s: names %s — not one of the read-only or value types this package may hold", use.Position, testenv.Describe(object))
			}
		case *types.Var:
			if !object.IsField() {
				t.Errorf("%s: uses %s — package-level journal state is not a read", use.Position, testenv.Describe(object))
			}
			// 필드 읽기 · 쓰기는 값을 옮길 뿐 능력이 아니다(쓰기 핸들은 위 타입 · 함수 갈래가 막는다).
		default:
			t.Errorf("%s: uses %s — unclassified journal symbol", use.Position, testenv.Describe(object))
		}
	}
	if openReadOnly == 0 || readOnlyMethods == 0 {
		t.Fatalf("OpenReadOnly uses=%d, ReadOnly method uses=%d — the census saw no journal use, so it proves nothing", openReadOnly, readOnlyMethods)
	}
}
