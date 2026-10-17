package engine

// a092 착지 단위 ④ — 델타 ADDED 「운영 모드는 산 프로세스의 진입 게이트에 닿는다」: 생산 조립(buildGateway)이 원장의 모드
// 투영기를 게이트에 묶고, 첫 진입 점검보다 먼저 현재 모드를 복원함. 복원 실패는 기동 거부가 아니라 모드 사유 래치(C15 (ㄴ) — 손절이 산다).

import (
	"context"
	"database/sql"
	"strings"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
)

const a092Account = "123-45"

// 기동 복원이 조립 반환 시점(루프 기동 전)에 이미 모드 사유를 세움.
func TestA092AssemblyRestoresTheOperatingModeBeforeAnyLoop(t *testing.T) {
	ctx := context.Background()
	j := openTestJournal(t)
	if err := bindApplyHooks(j); err != nil {
		t.Fatal(err)
	}
	if _, _, err := j.EscalateOperatingMode(ctx, a092Account, journal.ModeTriggerCredentialRejected, nil); err != nil {
		t.Fatal(err)
	}
	wiring := a098BuildGateway(t, ctx, j)
	if detail, ok := wiring.entry.Blocks()[execgw.ReasonOperatingModeBlocked]; !ok || !strings.Contains(detail, journal.ModeEntryBlocked) {
		t.Fatalf("mode latch after assembly = %q (present=%v), want ENTRY_BLOCKED restored", detail, ok)
	}
}

// 조립 뒤의 전이가 산 게이트에 투영됨(투영기 배선) — 강화와 사람 완화 둘 다.
func TestA092TransitionsAfterAssemblyReachTheLiveGate(t *testing.T) {
	ctx := context.Background()
	j := openTestJournal(t)
	if err := bindApplyHooks(j); err != nil {
		t.Fatal(err)
	}
	wiring := a098BuildGateway(t, ctx, j)
	if _, ok := wiring.entry.Blocks()[execgw.ReasonOperatingModeBlocked]; ok {
		t.Fatal("a NORMAL account started with a mode latch")
	}
	if _, _, err := j.EscalateOperatingMode(ctx, a092Account, journal.ModeTriggerExitObservationOutage, nil); err != nil {
		t.Fatal(err)
	}
	if _, ok := wiring.entry.Blocks()[execgw.ReasonOperatingModeBlocked]; !ok {
		t.Fatal("an automatic tightening after assembly did not reach the live gate")
	}
	if _, _, err := j.TransitionOperatingMode(ctx, journal.TransitionModeRequest{
		AccountRef: a092Account, Mode: journal.ModeNormal, Cause: "checked", Actor: journal.ModeActorOperator,
		Approval: "OPS-1", Auditor: a092NopAuditor{},
	}); err != nil {
		t.Fatal(err)
	}
	if _, ok := wiring.entry.Blocks()[execgw.ReasonOperatingModeBlocked]; ok {
		t.Fatal("a human relaxation did not clear the live mode latch")
	}
}

// C15 (ㄴ) · K8: 모드 행을 못 읽으면 기동은 계속하고 진입은 모드 사유로 막힘. 설명에 원장 오류는 없음.
func TestA092AnUnreadableModeLatchesInsteadOfRefusingToStart(t *testing.T) {
	ctx := context.Background()
	j := openTestJournal(t)
	if err := bindApplyHooks(j); err != nil {
		t.Fatal(err)
	}
	if _, _, err := j.EscalateOperatingMode(ctx, a092Account, journal.ModeTriggerCredentialRejected, nil); err != nil {
		t.Fatal(err)
	}
	// 시각 칸을 못 읽는 값으로 — 모드 행 고유 오류(currentModeFromRow 의 파싱 실패).
	db, err := sql.Open("sqlite", j.Path())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`UPDATE operating_modes SET created_at = 'not-a-time'`); err != nil {
		t.Fatal(err)
	}
	_ = db.Close()

	wiring := a098BuildGateway(t, ctx, j) // 기동 거부면 여기서 Fatal
	detail, ok := wiring.entry.Blocks()[execgw.ReasonOperatingModeBlocked]
	if !ok {
		t.Fatal("an unreadable operating mode left entries open")
	}
	if strings.Contains(detail, "not-a-time") || strings.Contains(detail, a092Account) {
		t.Errorf("the latch detail carries the ledger error or the account: %q", detail)
	}
}

type a092NopAuditor struct{}

func (a092NopAuditor) RecordAction(string, string, string, string) error { return nil }
