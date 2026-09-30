package engine

// a124 tasks 2.14 — 집행 경계 핀 (design D10, freeze 12·13회차 AC1 · AD1 · AD2).
//
// 이 change 가 쓰는 운영 모드 승격은 **원장 행**이다. a124 착지 때는 그 행을 진입 게이트로 투영하는 투영기
// (`Journal.SetModeProjector`)와 기동 복원(`Journal.RestoreOperatingModeProjection`)의 생산 호출자가 0 이라
// 모드 행이 아무것도 막지 않았다. a092 단위 ④가 그 배선을 착지시켰다(`mode_projection_wiring.go` — a124 핀 (b) ·
// (c)가 예고한 정상 경로, a092 tasks 21.7(d)). 이 파일은 바뀐 사실을 **문서와 같게** 못 박는다:
//
//	(a) 이 change(a124)의 시험은 픽스처에서 투영기를 직접 묶지 않는다 — 묶는 것은 생산 조립 하나다
//	(b) 두 함수의 비시험 호출자 = 생산 조립의 투영 배선 한 곳(`bindOperatingModeProjection`) — 다른 곳이 묶으면
//	    재묶기 거절 · 이중 투영이 된다
//	(c) 생산 조립(buildGateway)의 게이트에서, 승인은 알림 사유만 풀고 **모드 사유는 남는다**(원장의 ENTRY_BLOCKED
//	    가 투영됨). 미전달 0 재시작 뒤에도 기동 복원이 모드 사유를 세운다. 모드 사유를 푸는 것은 사람의 완화
//	    (`tossctl engine mode-release`, a092)다. 허용·강제 재잠금 두 변형은 알림 사유가 **있다**.

import (
	"context"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/obs"
)

var a124ProjectionBinders = map[string]bool{"SetModeProjector": true, "RestoreOperatingModeProjection": true}

// a124ProjectionCalls 는 파일들에서 두 함수를 **부르는** 자리를 센다(문자열 · 주석은 세지 않는다).
func a124ProjectionCalls(t *testing.T, root string, keep func(path string) bool) []string {
	t.Helper()
	var sites []string
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			switch d.Name() {
			case ".git", "vendor", "node_modules", "openspec", ".sdd":
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") || !keep(path) {
			return nil
		}
		fset := token.NewFileSet()
		f, perr := parser.ParseFile(fset, path, nil, 0)
		if perr != nil {
			// 못 읽는 파일을 건너뛰면 그 안의 호출이 조용히 빠진다 — 멈춘다(Eng 리뷰 F6).
			return perr
		}
		ast.Inspect(f, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}
			if sel, ok := call.Fun.(*ast.SelectorExpr); ok && a124ProjectionBinders[sel.Sel.Name] {
				sites = append(sites, fset.Position(call.Pos()).String())
			}
			return true
		})
		return nil
	})
	if err != nil {
		t.Fatalf("walking %s: %v", root, err)
	}
	return sites
}

// (a)
func TestA124FixturesNeverBindTheModeProjector(t *testing.T) {
	root := filepath.Join("..", "..", "..")
	sites := a124ProjectionCalls(t, root, func(p string) bool {
		base := filepath.Base(p)
		return strings.HasPrefix(base, "a124_") && strings.HasSuffix(base, "_test.go")
	})
	if len(sites) != 0 {
		t.Fatalf("an a124 test binds or restores the mode projector %v — it would test enforcement production does not have", sites)
	}
}

// (b)
func TestTheModeProjectorIsBoundOnlyByTheEngineAssembly(t *testing.T) {
	root := filepath.Join("..", "..", "..")
	// 양성 대조: 같은 걸음이 시험 파일의 호출자는 찾아야 한다 — 못 찾으면 계측기가 눈먼 것이다.
	if tests := a124ProjectionCalls(t, root, func(p string) bool { return strings.HasSuffix(p, "_test.go") }); len(tests) == 0 {
		t.Fatal("the walk found no test caller either — it is not looking at the tree")
	}
	sites := a124ProjectionCalls(t, root, func(p string) bool { return !strings.HasSuffix(p, "_test.go") })
	if len(sites) != 2 {
		t.Fatalf("production binders of the mode projection = %v, want exactly SetModeProjector and "+
			"RestoreOperatingModeProjection in the engine assembly (a092 unit 4)", sites)
	}
	for _, site := range sites {
		if !strings.Contains(filepath.ToSlash(site), "internal/app/engine/mode_projection_wiring.go:") {
			t.Errorf("the mode projection is bound outside the engine assembly at %s — a second binder is a second projection", site)
		}
	}
}

// a124ProductionFixture 는 생산 조립의 게이트 위에 실행자를 세운다.
func a124ProductionFixture(t *testing.T) (*journal.Journal, engineWiring, *alertDeliverer, *a124Ledger) {
	t.Helper()
	ctx := context.Background()
	j := openTestJournal(t)
	if err := bindApplyHooks(j); err != nil {
		t.Fatalf("bindApplyHooks: %v", err)
	}
	wiring := a098BuildGateway(t, ctx, j)
	led := &a124Ledger{Journal: j}
	d := &alertDeliverer{
		Journal: j, Publisher: a124Failing(), Clock: clock.System(), Claimant: "a124-boundary",
		Gate: wiring.entry, AccountRef: "123-45", ledger: led,
	}
	if _, err := j.EnqueueAlert(ctx, journal.Alert{EventKey: "a124-boundary", Type: "x", Severity: "critical", Title: "t"}); err != nil {
		t.Fatalf("EnqueueAlert: %v", err)
	}
	return j, wiring, d, led
}

func a124Reasons(gate *execgw.EntryGate) (alert, mode bool) {
	blocks := gate.Blocks()
	_, alert = blocks[execgw.ReasonAlertUndelivered]
	_, mode = blocks[execgw.ReasonOperatingModeBlocked]
	return alert, mode
}

func a124ModeRow(t *testing.T, j *journal.Journal) string {
	t.Helper()
	snap, err := j.CurrentOperatingMode(context.Background(), "123-45")
	if err != nil {
		t.Fatalf("CurrentOperatingMode: %v", err)
	}
	return snap.Mode
}

// a124ObserveEverything 는 생산 임계값을 그대로 둔 채 필수 조회에 정당한 관측을 준다.
func a124ObserveEverything(gate *execgw.EntryGate) {
	for kind := range execgw.DefaultStaleness() {
		gate.RecordSuccess(kind)
	}
}

// (c) 승인 뒤 · 미전달 0 재시작 뒤 — a092 배선 뒤에는 모드 사유가 남는다.
func TestTheLedgerModeRowEnforcesEntryOnceProjectionIsWired(t *testing.T) {
	ctx := context.Background()
	j, wiring, d, _ := a124ProductionFixture(t)
	for i := 0; i < alertAttemptLimit; i++ {
		_ = d.cycle(ctx)
	}
	if alert, _ := a124Reasons(wiring.entry); !alert || a124ModeRow(t, j) != journal.ModeEntryBlocked {
		t.Fatalf("arrangement: the executor did not latch and escalate (alert=%v mode=%s)", alert, a124ModeRow(t, j))
	}

	if err := (&obs.Notifier{Journal: j, Gate: wiring.entry}).Acknowledge(ctx, "operator"); err != nil {
		t.Fatalf("Acknowledge: %v", err)
	}
	// ① 사유 단언: 승인은 알림 사유만 풂. 승격이 남긴 모드 행은 산 게이트에 투영돼 있음.
	if alert, mode := a124Reasons(wiring.entry); alert || !mode {
		t.Fatalf("after acknowledgement: alert reason=%v mode reason=%v — want the alert reason gone and the mode reason kept", alert, mode)
	}
	// ② 원장 행
	if got := a124ModeRow(t, j); got != journal.ModeEntryBlocked {
		t.Fatalf("mode row = %s, want %s kept in the ledger", got, journal.ModeEntryBlocked)
	}
	// ③ 통제된 경우: 정당한 관측을 준 뒤에도 진입은 모드 사유로 거절됨.
	a124ObserveEverything(wiring.entry)
	if rejected := wiring.entry.CheckEntryFor("KR", "005930"); rejected == nil || rejected.Reason != execgw.ReasonOperatingModeBlocked {
		t.Fatalf("with every required query observed, entry check = %v — want %s", rejected, execgw.ReasonOperatingModeBlocked)
	}

	// 미전달 0 재시작 — 새 프로세스는 원장을 새로 연다(한 핸들에 투영기는 한 번만 묶임). 기동 복원이 모드 사유를 다시 세움.
	path := j.Path()
	if err := j.Close(); err != nil {
		t.Fatalf("closing the journal for the restart: %v", err)
	}
	reopened, err := journal.Open(ctx, journal.Options{
		Path: path, FSProber: journal.FixedFSProber(journal.FSInfo{Name: "ext4", Magic: journal.MagicExt}),
	})
	if err != nil {
		t.Fatalf("reopening the journal: %v", err)
	}
	t.Cleanup(func() { _ = reopened.Close() })
	if err := bindApplyHooks(reopened); err != nil {
		t.Fatalf("bindApplyHooks: %v", err)
	}
	restarted := a098BuildGateway(t, ctx, reopened)
	if alert, mode := a124Reasons(restarted.entry); alert || !mode {
		t.Fatalf("after a restart with nothing pending: alert reason=%v mode reason=%v — want the mode restored", alert, mode)
	}
	if got := a124ModeRow(t, reopened); got != journal.ModeEntryBlocked {
		t.Fatalf("mode row after restart = %s", got)
	}
}

// (c) 허용 재잠금: 「정산 → 해제 → 세대 읽기 → 적용」 = 알림 사유 있음.
func TestThePermittedRelatchHoldsOnTheProductionGate(t *testing.T) {
	ctx := context.Background()
	_, wiring, d, _ := a124ProductionFixture(t)
	for i := 0; i < alertAttemptLimit-1; i++ {
		_ = d.cycle(ctx)
	}
	a124AtStage(d, alertStageSettled, func() { wiring.entry.Clear(execgw.ReasonAlertUndelivered) })
	_ = d.cycle(ctx)
	if alert, _ := a124Reasons(wiring.entry); !alert {
		t.Fatal("the permitted conservative relatch did not hold on the production gate")
	}
}

// (c) 강제 재잠금: 「차단 → 해제 → 승격 쓰기 실패」 = 알림 사유 있음. 실패한 쓰기는 모드 행을 만들지 않는다.
func TestTheMandatedRelatchHoldsOnTheProductionGate(t *testing.T) {
	ctx := context.Background()
	j, wiring, d, led := a124ProductionFixture(t)
	led.failEscalate = true
	for i := 0; i < alertAttemptLimit-1; i++ {
		_ = d.cycle(ctx)
	}
	a124AtStage(d, alertStageLatched, func() { wiring.entry.Clear(execgw.ReasonAlertUndelivered) })
	_ = d.cycle(ctx)
	if alert, _ := a124Reasons(wiring.entry); !alert {
		t.Fatal("a failed escalation after a clear left no latch on the production gate")
	}
	if got := a124ModeRow(t, j); got != journal.ModeNormal {
		t.Fatalf("mode row = %s after a failed escalation write", got)
	}
}
