package execgw

// a124 tasks 2.13 — 진입 게이트의 사유별 **해제 세대** 계약 (design D7, M1 = B′).
//
// 배달 실행자는 판정 근거를 얻은 뒤 잠금 없이 여러 일을 하고 나서 차단을 적용한다. 그 사이
// 운영자의 해제가 끼면, 해제가 근거 **뒤**였는지를 실행자가 알아야 원칙 E(늦은 적용 = 제때
// 적용)를 지킬 수 있다. 그 순서의 증거가 이 세대다. 그래서 계약이 셋이다:
//
//   - 해제 **요청**마다 정확히 +1 (래치가 있었는지와 무관 — 반례 ㉣, Manager 승인)
//   - 그 사유의 해제 요청 없이는 **불변** (다른 사유의 해제 · 잠금 · 종목 단위 · 모드 투영 · 대사 재구성)
//   - 조건부 잠금은 세대 비교와 삽입을 **한 잠금 안에서** 한다
//
// 기존 `revision`(전략 진입 봉인)의 의미는 그대로다 — 실제로 상태가 바뀐 때만 오른다.

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
)

func a124Gate() *EntryGate {
	return NewEntryGate(nil, map[RequiredQuery]time.Duration{})
}

func TestEveryClearRequestAdvancesTheReasonsEpoch(t *testing.T) {
	g := a124Gate()
	if got := g.ClearEpoch(ReasonAlertUndelivered); got != 0 {
		t.Fatalf("fresh epoch = %d, want 0", got)
	}
	rev := g.revision

	// 래치가 없는 게이트에 해제 요청 — 세대는 오르고 revision 은 그대로(실제 변화 없음).
	g.Clear(ReasonAlertUndelivered)
	if got := g.ClearEpoch(ReasonAlertUndelivered); got != 1 {
		t.Fatalf("epoch after a clear with no latch = %d, want 1 (a request is the operator's mark, ㉣)", got)
	}
	if g.revision != rev {
		t.Fatalf("revision moved on a clear that removed nothing: %d → %d", rev, g.revision)
	}

	// 래치가 있는 해제 — 세대 +1, revision +1(기존 의미).
	g.Block(ReasonAlertUndelivered, "x")
	rev = g.revision
	g.Clear(ReasonAlertUndelivered)
	if got := g.ClearEpoch(ReasonAlertUndelivered); got != 2 {
		t.Fatalf("epoch = %d, want 2", got)
	}
	if g.revision != rev+1 {
		t.Fatalf("revision after a real clear = %d, want %d", g.revision, rev+1)
	}
}

func TestNothingButThatReasonsClearMovesItsEpoch(t *testing.T) {
	g := a124Gate()
	g.Clear(ReasonAlertUndelivered)
	before := g.ClearEpoch(ReasonAlertUndelivered)

	g.Block(ReasonAlertUndelivered, "latched")
	g.Block(ReasonBrokerAuthRejected, "auth")
	g.Clear(ReasonBrokerAuthRejected)
	g.Clear(ReasonAlertSenderDown)
	g.BlockSymbol("KR", "005930", ReasonAlertUndelivered, "symbol")
	g.ClearSymbol("KR", "005930", ReasonAlertUndelivered)
	g.ClearSymbolReason(ReasonAlertUndelivered)
	g.ProjectOperatingMode(journal.OperatingModeRecord{Mode: journal.ModeEntryBlocked})
	g.ProjectOperatingMode(journal.OperatingModeRecord{Mode: journal.ModeNormal})
	g.RebuildReconcileProjection(nil)
	g.RecordSuccess(QueryPrice)
	_, _ = g.BlockUnlessClearedSince(ReasonAlertUndelivered, before, "conditional")

	if got := g.ClearEpoch(ReasonAlertUndelivered); got != before {
		t.Fatalf("epoch moved without a clear request of its reason: %d → %d", before, got)
	}
}

func TestAConditionalBlockAfterAClearChangesNothing(t *testing.T) {
	g := a124Gate()
	epoch := g.ClearEpoch(ReasonAlertUndelivered)
	g.Clear(ReasonAlertUndelivered)
	rev := g.revision
	if applied, _ := g.BlockUnlessClearedSince(ReasonAlertUndelivered, epoch, "late"); applied {
		t.Fatal("a conditional block on a stale epoch reported that it latched")
	}
	if _, latched := g.Blocks()[ReasonAlertUndelivered]; latched || g.revision != rev {
		t.Fatalf("a stale conditional block changed the gate: latched=%v revision %d→%d", latched, rev, g.revision)
	}
}

func TestAConditionalBlockOnTheCurrentEpochFollowsTheBlockRule(t *testing.T) {
	g := a124Gate()
	epoch := g.ClearEpoch(ReasonAlertUndelivered)
	rev := g.revision
	if applied, inserted := g.BlockUnlessClearedSince(ReasonAlertUndelivered, epoch, "first"); !applied || !inserted {
		t.Fatalf("a conditional block on the current epoch: applied=%v inserted=%v, want both", applied, inserted)
	}
	if g.Blocks()[ReasonAlertUndelivered] != "first" || g.revision != rev+1 {
		t.Fatalf("block rule not followed: detail=%q revision %d→%d", g.Blocks()[ReasonAlertUndelivered], rev, g.revision)
	}
	// 이미 있으면 삽입하지 않는다 — 처음 설명 유지, revision 불변(F9).
	if applied, inserted := g.BlockUnlessClearedSince(ReasonAlertUndelivered, epoch, "second"); !applied || inserted {
		t.Fatalf("a repeated conditional block: applied=%v inserted=%v, want applied and not inserted", applied, inserted)
	}
	if g.Blocks()[ReasonAlertUndelivered] != "first" || g.revision != rev+1 {
		t.Fatalf("a repeated block rewrote the gate: detail=%q revision=%d", g.Blocks()[ReasonAlertUndelivered], g.revision)
	}
}

// TestTheEpochComparisonAndTheLatchAreOneStep — 해제와 조건부 잠금이 겹친다. 경합 검출기 없이도 split-lock 변이를
// 잡는다(변이 M29, 원장 기록). execgw 는 `make test-race` 의 패키지 목록에 없으므로 `-race` 아래 도는 것은 로트의 수동
// 실행뿐이다 — 게이트가 경합 검출기로 이 시험을 돈다고 주장하지 않는다(gstack /review 시험 전문가).
// 불변식: 조건부 잠금이 true 를 돌려준 뒤 **그 세대로는** 해제가 없었다 → 래치가 서 있거나,
// 그 뒤 해제가 있었다면 세대가 올라 있다. 「비교는 통과했는데 삽입 전에 해제가 끼어
// 해제 뒤에 래치가 남았는데 세대는 그대로」 는 나올 수 없다.
func TestTheEpochComparisonAndTheLatchAreOneStep(t *testing.T) {
	g := a124Gate()
	var wg sync.WaitGroup
	stop := make(chan struct{})
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-stop:
				return
			default:
				g.Clear(ReasonAlertUndelivered)
			}
		}
	}()
	for i := 0; i < 2000; i++ {
		e := g.ClearEpoch(ReasonAlertUndelivered)
		if applied, _ := g.BlockUnlessClearedSince(ReasonAlertUndelivered, e, "race"); applied {
			g.mu.Lock()
			_, latched := g.latches[ReasonAlertUndelivered]
			now := g.clearEpochs[ReasonAlertUndelivered]
			g.mu.Unlock()
			if !latched && now == e {
				t.Fatalf("a successful conditional block left no latch and no newer clear (epoch %d)", e)
			}
			// 이 goroutine 만 삽입한다. 그러니 「래치가 있는데 세대가 e 보다 크다」는 삽입이 그 해제 **뒤**에
			// 일어났다는 뜻 — 비교와 삽입이 한 잠금이 아닐 때만 나오는 모양이다(codex 구현 1회차 I4).
			if latched && now != e {
				t.Fatalf("a latch survived a clear that came after the epoch comparison (compared %d, now %d)", e, now)
			}
		}
	}
	close(stop)
	wg.Wait()
}

// TestOnlyTheAcknowledgementClearsTheUndeliveredReason — 구조 핀. 해제 세대의 의미(「해제 요청은
// 사람 승인의 표식」)는 이 사유를 푸는 비시험 호출자가 `Notifier.Acknowledge` 의 두 자리뿐이라는
// 사실에 선다(design D7, evidence R8). 새 호출자가 생기면 빨강 — 그 편집은 세대 의미를 다시 증명해야 한다.
func TestOnlyTheAcknowledgementClearsTheUndeliveredReason(t *testing.T) {
	root := filepath.Join("..", "..")
	type site struct{ file, fn string }
	var found []site
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			name := d.Name()
			if name == ".git" || name == "vendor" || name == "node_modules" || name == "openspec" || name == ".sdd" {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		fset := token.NewFileSet()
		file, perr := parser.ParseFile(fset, path, nil, 0)
		if perr != nil {
			return perr // 못 읽는 파일을 건너뛰면 그 안의 호출이 조용히 빠진다(Eng 리뷰 F6)
		}
		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Body == nil {
				continue
			}
			ast.Inspect(fn.Body, func(n ast.Node) bool {
				call, ok := n.(*ast.CallExpr)
				if !ok || len(call.Args) != 1 {
					return true
				}
				sel, ok := call.Fun.(*ast.SelectorExpr)
				if !ok || sel.Sel.Name != "Clear" {
					return true
				}
				arg := call.Args[0]
				name := ""
				switch a := arg.(type) {
				case *ast.SelectorExpr:
					name = a.Sel.Name
				case *ast.Ident:
					name = a.Name
				}
				if name == "ReasonAlertUndelivered" {
					recv := ""
					if fn.Recv != nil && len(fn.Recv.List) == 1 {
						switch rt := fn.Recv.List[0].Type.(type) {
						case *ast.StarExpr:
							if id, ok := rt.X.(*ast.Ident); ok {
								recv = id.Name + "."
							}
						case *ast.Ident:
							recv = rt.Name + "."
						}
					}
					found = append(found, site{filepath.ToSlash(path), recv + fn.Name.Name})
				}
				return true
			})
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walking the tree: %v", err)
	}
	if len(found) != 2 {
		t.Fatalf("non-test Clear(ReasonAlertUndelivered) sites = %v, want exactly the two in Notifier.Acknowledge", found)
	}
	for _, s := range found {
		if s.fn != "Notifier.Acknowledge" || !strings.HasSuffix(s.file, "internal/obs/notifier.go") {
			t.Fatalf("Clear(ReasonAlertUndelivered) called from %s in %s — a new caller must re-prove the clear epoch's meaning", s.fn, s.file)
		}
	}
}

// 게이트 잠금 안에서는 map 연산만 한다 — 해제 세대 두 메서드가 잠금 말고 아무것도 부르지 않음을 구조로 못 박는다.
// 이것이 「배달 실행자가 잡는 잠금은 게이트 잠금뿐이고 그 안에서 밖을 부르지 않는다」(spec, design D7)의 받침이다:
// 실행자 쪽 코드는 g.mu 아래에서 무엇을 돌릴 API 가 없고(B′), 이 두 메서드가 그 안을 채운다.
func TestTheEpochMethodsCallNothingUnderTheLock(t *testing.T) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "retry.go", nil, 0)
	if err != nil {
		t.Fatalf("parse retry.go: %v", err)
	}
	want := map[string]bool{"ClearEpoch": false, "BlockUnlessClearedSince": false}
	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Recv == nil {
			continue
		}
		if _, tracked := want[fn.Name.Name]; !tracked {
			continue
		}
		want[fn.Name.Name] = true
		ast.Inspect(fn.Body, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}
			if sel, ok := call.Fun.(*ast.SelectorExpr); ok && (sel.Sel.Name == "Lock" || sel.Sel.Name == "Unlock") {
				return true
			}
			t.Errorf("%s calls %s under the gate lock — only map operations may run there", fn.Name.Name, fset.Position(call.Pos()))
			return true
		})
	}
	for name, seen := range want {
		if !seen {
			t.Fatalf("%s not found in retry.go — the pin is looking at the wrong file", name)
		}
	}
}
