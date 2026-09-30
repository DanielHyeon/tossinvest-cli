package engine

// a092 24.3 M13(게이트 준비 gstack 리뷰가 찾은 거짓 완료의 정정): 기동 복원 실패 래치(모드 순번 0 에서 직접 Block)는 원장이
// 고쳐진 뒤 이어진 **성공 투영**(사람 완화 · 자동 강화)에 교체됨 — 울타리(순번 0 → 첫 순번)와 사유 존재(had) 규칙이 이
// 경로에서 서로를 막지 않음. 복원 실패의 원문 오류 로그에 계좌가 없음을 핀으로 둠(현 원장의 복원 오류 문구는 계좌를
// 담지 않음 — 담게 되면 이 핀이 깨짐, 불변식 8 (a)).

import (
	"bytes"
	"context"
	"database/sql"
	"strings"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/obs"
)

func TestA092ARestoreFailureLatchIsReplacedByTheNextProjection(t *testing.T) {
	ctx := context.Background()
	j := openTestJournal(t)
	if err := bindApplyHooks(j); err != nil {
		t.Fatal(err)
	}
	if _, _, err := j.EscalateOperatingMode(ctx, a092Account, journal.ModeTriggerCredentialRejected, nil); err != nil {
		t.Fatal(err)
	}
	db, err := sql.Open("sqlite", j.Path())
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	var created string
	if err := db.QueryRow(`SELECT created_at FROM operating_modes`).Scan(&created); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`UPDATE operating_modes SET created_at = 'not-a-time'`); err != nil {
		t.Fatal(err)
	}

	gate := execgw.NewEntryGate(clock.System(), nil)
	logs := &bytes.Buffer{}
	if err := bindOperatingModeProjection(ctx, j, gate, a092Account, obs.NewLogger(obs.LogOptions{Writer: logs, JSON: true})); err != nil {
		t.Fatal(err)
	}
	if detail := gate.Blocks()[execgw.ReasonOperatingModeBlocked]; detail != operatingModeRestoreFailed {
		t.Fatalf("mode latch after a failed restore = %q, want the restore-failure latch", detail)
	}
	if strings.Contains(logs.String(), a092Account) {
		t.Errorf("the restore-failure log carries the account number:\n%s", logs.String())
	}

	// 원장 수리 뒤 사람의 완화 — 성공 투영이 복원 실패 래치를 지움(NORMAL).
	if _, err := db.Exec(`UPDATE operating_modes SET created_at = ?`, created); err != nil {
		t.Fatal(err)
	}
	if _, _, err := j.TransitionOperatingMode(ctx, journal.TransitionModeRequest{
		AccountRef: a092Account, Mode: journal.ModeNormal, Cause: "journal repaired", Actor: journal.ModeActorOperator,
		Approval: "OPS-M13", Auditor: a092NopAuditor{},
	}); err != nil {
		t.Fatal(err)
	}
	if detail, ok := gate.Blocks()[execgw.ReasonOperatingModeBlocked]; ok {
		t.Fatalf("a successful projection left the restore-failure latch in place: %q", detail)
	}

	// 자동 강화 — 설명이 모드 행의 것(복원 실패 문구가 아님).
	if _, _, err := j.EscalateOperatingMode(ctx, a092Account, journal.ModeTriggerExitObservationOutage, nil); err != nil {
		t.Fatal(err)
	}
	detail, ok := gate.Blocks()[execgw.ReasonOperatingModeBlocked]
	if !ok || detail == operatingModeRestoreFailed || !strings.Contains(detail, journal.ModeEntryBlocked) {
		t.Errorf("mode latch after a later tightening = %q (present=%v), want the tightening's own detail", detail, ok)
	}
}
