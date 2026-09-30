package execgw

// a092 착지 단위 ④ — AC2 · C5 · C20: 운영 모드 투영의 원자 교체 · 커밋 순서 울타리 · 상태 세대 규칙.

import (
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
)

func a092ModeGate() *EntryGate {
	return NewEntryGate(clock.NewFake(time.Date(2026, 9, 30, 0, 0, 0, 0, time.UTC)), map[RequiredQuery]time.Duration{})
}

func a092Rec(seq int64, mode, cause string) journal.OperatingModeRecord {
	return journal.OperatingModeRecord{Seq: seq, Mode: mode, Cause: cause, Actor: journal.ModeActorAuto}
}

// 교체: 막는 모드에서 막는 모드로 — 설명이 새 행의 것으로 바뀜(옛 판은 Block 이 「없을 때만」 넣어 옛 설명이 남음).
func TestA092AProjectionReplacesTheModeLatch(t *testing.T) {
	g := a092ModeGate()
	g.ProjectOperatingMode(a092Rec(1, journal.ModeEntryBlocked, "EXIT_OBSERVATION_OUTAGE"))
	g.ProjectOperatingMode(a092Rec(2, journal.ModeHaltAll, "operator halt"))
	detail, blocked := g.OperatingModeBlocked()
	if !blocked || !strings.Contains(detail, journal.ModeHaltAll) || strings.Contains(detail, "OUTAGE") {
		t.Errorf("mode latch = %q (blocked=%v), want the HALT_ALL row's detail", detail, blocked)
	}
}

// C5: 커밋 순서 울타리 — 뒤바뀌어 도착한 옛 투영은 적용하지 않음. 초기값 0 이고 「보다 큰」.
func TestA092AStaleProjectionIsNotApplied(t *testing.T) {
	g := a092ModeGate()
	g.ProjectOperatingMode(a092Rec(5, journal.ModeEntryBlocked, "CREDENTIAL"))
	g.ProjectOperatingMode(a092Rec(3, journal.ModeNormal, "older relax"))
	if _, blocked := g.OperatingModeBlocked(); !blocked {
		t.Fatal("an older relaxation arriving late cleared a newer tightening — the gate went backwards")
	}
	g.ProjectOperatingMode(a092Rec(5, journal.ModeNormal, "same seq"))
	if _, blocked := g.OperatingModeBlocked(); !blocked {
		t.Error("a projection with an equal sequence was applied — the fence is strictly greater-than")
	}
	g.ProjectOperatingMode(a092Rec(6, journal.ModeNormal, "newer relax"))
	if _, blocked := g.OperatingModeBlocked(); blocked {
		t.Error("a newer relaxation was not applied")
	}
	fresh := a092ModeGate()
	fresh.ProjectOperatingMode(a092Rec(0, journal.ModeEntryBlocked, "unsequenced"))
	if _, blocked := fresh.OperatingModeBlocked(); blocked {
		t.Error("a sequence-0 projection was applied — the fence starts at 0 and admits only greater")
	}
}

// C20: 상태 세대(revision)는 모드 사유의 존재가 바뀔 때만 +1. 설명만 바뀌는 교체는 +0. 해제 세대는 안 움직임(a124).
func TestA092TheModeRevisionMovesOnlyWhenPresenceChanges(t *testing.T) {
	g := a092ModeGate()
	epoch := g.ClearEpoch(ReasonOperatingModeBlocked)
	r0 := g.revision
	g.ProjectOperatingMode(a092Rec(1, journal.ModeEntryBlocked, "A"))
	if g.revision != r0+1 {
		t.Fatalf("revision after none→blocked = %d, want %d", g.revision, r0+1)
	}
	g.ProjectOperatingMode(a092Rec(2, journal.ModeEntryBlocked, "B"))
	if g.revision != r0+1 {
		t.Errorf("revision after a detail-only replacement = %d, want %d", g.revision, r0+1)
	}
	g.ProjectOperatingMode(a092Rec(3, journal.ModeNormal, "relax"))
	if g.revision != r0+2 {
		t.Errorf("revision after blocked→none = %d, want %d", g.revision, r0+2)
	}
	g.ProjectOperatingMode(a092Rec(4, journal.ModeNormal, "again"))
	if g.revision != r0+2 {
		t.Errorf("revision after none→none = %d, want %d", g.revision, r0+2)
	}
	if got := g.ClearEpoch(ReasonOperatingModeBlocked); got != epoch {
		t.Errorf("a projection moved the clear epoch: %d → %d", epoch, got)
	}
}

// AC2: 교체 중 겹친 진입 점검이 모드 사유가 없는 순간을 보지 않음.
func TestA092NoCheckSeesAGapDuringReplacement(t *testing.T) {
	g := a092ModeGate()
	g.ProjectOperatingMode(a092Rec(1, journal.ModeEntryBlocked, "start"))
	var wg sync.WaitGroup
	stop := make(chan struct{})
	gaps := 0
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-stop:
				return
			default:
			}
			if rej := g.CheckEntry(); rej == nil {
				gaps++
			}
		}
	}()
	for i := int64(2); i < 3000; i++ {
		g.ProjectOperatingMode(a092Rec(i, journal.ModeEntryBlocked, "replace"))
	}
	close(stop)
	wg.Wait()
	if gaps != 0 {
		t.Errorf("entry checks saw no latch %d times while one blocking mode replaced another", gaps)
	}
}

// 구조 핀(AC2): 투영 본문은 g.mu 를 한 번 잡고, 다른 잠금 메서드(Block · Clear)를 부르지 않음 — 지우고 다시 넣는 사이 잠금을 놓는 형태 금지.
func TestA092ProjectionTakesTheGateLockOnce(t *testing.T) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "modegate.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	var body *ast.BlockStmt
	for _, d := range f.Decls {
		if fd, ok := d.(*ast.FuncDecl); ok && fd.Name.Name == "ProjectOperatingMode" {
			body = fd.Body
		}
	}
	if body == nil {
		t.Fatal("ProjectOperatingMode not found")
	}
	locks, unlocks := 0, 0
	ast.Inspect(body, func(n ast.Node) bool {
		c, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		s, ok := c.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		switch s.Sel.Name {
		case "Lock":
			locks++
		case "Unlock":
			unlocks++
		case "Block", "Clear", "BlockUnlessClearedSince", "ClearSymbolReason":
			t.Errorf("ProjectOperatingMode calls %s — it must replace the latch inside its own single critical section", s.Sel.Name)
		}
		return true
	})
	if locks != 1 || unlocks != 1 {
		t.Errorf("Lock=%d Unlock=%d, want exactly one critical section", locks, unlocks)
	}
}
