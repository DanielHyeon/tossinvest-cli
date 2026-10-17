package main

// a092 착지 단위 ④ — `tossctl engine mode-release` 표면(22.3 C12 · C14 · C16) 과 엔진 소켓을 거친 한 바퀴.

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/app/engine"
	"github.com/JungHoonGhae/tossinvest-cli/internal/audit"
	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/obs"
	"github.com/JungHoonGhae/tossinvest-cli/internal/output"
	"github.com/spf13/cobra"
)

// 표면: mutating=true · source=local · 필수 플래그 넷(기본값 없음) · 확인 마찰 없음.
func TestA092ModeReleaseDeclaresWhatItIs(t *testing.T) {
	var release *cobra.Command
	for _, c := range leafCommands(newRootCmd()) {
		if c.CommandPath() == "tossctl engine mode-release" {
			release = c
		}
	}
	if release == nil {
		t.Fatal("tossctl engine mode-release is not registered — no production path can relax the mode")
	}
	if release.Annotations["mutating"] != "true" || release.Annotations["source"] != "local" {
		t.Errorf("annotations = %v, want mutating=true source=local", release.Annotations)
	}
	for _, name := range []string{"to", "operator", "approval", "reason"} {
		flag := release.Flags().Lookup(name)
		if flag == nil {
			t.Fatalf("no --%s", name)
		}
		if flag.DefValue != "" {
			t.Errorf("--%s defaults to %q", name, flag.DefValue)
		}
		if ann := flag.Annotations[cobra.BashCompOneRequiredFlag]; len(ann) != 1 || ann[0] != "true" {
			t.Errorf("--%s is not required", name)
		}
	}
	for _, friction := range []string{"confirm", "yes", "force"} {
		if release.Flags().Lookup(friction) != nil {
			t.Errorf("mode-release has --%s; 추가 승인 마찰을 넣지 않는다", friction)
		}
	}
}

// 엔진 없이 부르면 거절(원장 직접 경로 없음).
func TestA092ModeReleaseRefusesWithoutAnEngine(t *testing.T) {
	dir := t.TempDir()
	if err := os.Chmod(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	if _, err := dialModeControl(context.Background(), dir); err == nil ||
		!strings.Contains(err.Error(), "완화 표면이 없다") {
		t.Fatalf("err = %v, want the missing-surface message", err)
	}
}

// 출력: 변화 없음 · 통지 실패 · 남은 사유를 말함.
func TestA092ModeReleaseSaysWhatRemains(t *testing.T) {
	var b strings.Builder
	_ = writeModeReleaseResult(&b, output.FormatTable, engine.ModeReleaseResult{
		Changed: true, TransitionID: "t1", Mode: "NORMAL", Notified: true, NoticePending: true,
		EntryBlocks: []string{"ALERT_UNDELIVERED"},
	})
	if !strings.Contains(b.String(), "ALERT_UNDELIVERED") || !strings.Contains(b.String(), "재개된 것은 아니다") {
		t.Errorf("output = %q", b.String())
	}
	b.Reset()
	_ = writeModeReleaseResult(&b, output.FormatTable, engine.ModeReleaseResult{
		Changed: true, TransitionID: "t1", Mode: "NORMAL", NotifyError: "outbox disk full",
	})
	if !strings.Contains(b.String(), "완화됨 · 통지 기록 실패") {
		t.Errorf("output = %q", b.String())
	}
}

// 엔진 소켓을 거친 한 바퀴: 생산 클라이언트(dialModeControl)로 엔진의 모드 제어 엔드포인트에 붙어 완화 → 산 게이트가 풀리고
// audit 줄이 남음. 엔진 쪽 판정 시험은 internal/app/engine/a092_mode_release_test.go 몫.
func TestA092ModeReleaseGoesThroughTheRunningEngine(t *testing.T) {
	ctx := context.Background()
	dir, err := os.MkdirTemp("/tmp", "a092-mode-*") // sun_path 한도 — 긴 $TMPDIR 에서 bind 가 깨짐(gstack 리뷰), a108 하니스와 같은 선택
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(dir) })
	if err := os.Chmod(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	j, err := journal.Open(ctx, journal.Options{Path: filepath.Join(dir, journal.DBFileName),
		FSProber: journal.FixedFSProber(journal.FSInfo{Name: "ext4", Magic: journal.MagicExt})})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = j.Close() })
	gate := execgw.NewEntryGate(clock.System(), nil)
	if err := j.SetModeProjector(gate); err != nil {
		t.Fatal(err)
	}
	const account = "acct-a092"
	if _, _, err := j.EscalateOperatingMode(ctx, account, journal.ModeTriggerCredentialRejected, nil); err != nil {
		t.Fatal(err)
	}
	auditPath := filepath.Join(dir, "audit.log")
	log, err := audit.Open(audit.Options{Path: auditPath, Subject: "engine"})
	if err != nil {
		t.Fatal(err)
	}
	ectx := &engine.Context{Journal: j, Entry: gate, AccountRef: account, Audit: log,
		Notifier: &obs.Notifier{Journal: j, Gate: gate, AccountRef: account}}
	ops, err := ectx.ModeOperations()
	if err != nil {
		t.Fatal(err)
	}
	server, err := engine.StartModeControlServer(dir, ops)
	if err != nil {
		t.Fatalf("StartModeControlServer: %v", err)
	}
	t.Cleanup(func() { _ = server.Close() })

	client, err := dialModeControl(ctx, dir)
	if err != nil {
		t.Fatalf("dialModeControl: %v", err)
	}
	// 빈 칸은 엔진이 400 invalid 로 거절(CLI 검사를 건너뛴 직접 호출).
	// 운영자 입력 실수는 400 invalid — 엔진 고장(500)과 가름. 문구가 아니라 상태 · 코드로 봄(오류 문구에도 "invalid" 가 들어 있음).
	_, err = client.Release(ctx, engine.ModeReleaseRequest{To: "NORMAL", Operator: "ops"})
	var remote *modeControlError
	if !errors.As(err, &remote) || remote.status != 400 || remote.code != "invalid" {
		t.Fatalf("err = %#v, want HTTP 400 invalid", err)
	}
	res, err := client.Release(ctx, engine.ModeReleaseRequest{To: "NORMAL", Operator: "박지훈",
		Approval: "OPS-1", Reason: "credential rotated"})
	if err != nil {
		t.Fatalf("Release: %v", err)
	}
	if !res.Changed || res.Mode != journal.ModeNormal {
		t.Fatalf("result = %+v", res)
	}
	if _, blocked := gate.OperatingModeBlocked(); blocked {
		t.Error("the live gate still carries the mode latch after a release through the engine")
	}
	body, err := os.ReadFile(auditPath)
	if err != nil || !strings.Contains(string(body), "OPS-1") {
		t.Errorf("audit log lacks the approval: %v %q", err, body)
	}
}
