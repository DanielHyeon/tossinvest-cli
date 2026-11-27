package main

// a206 — `tossctl engine exit-evidence-compact` 의 표면: 엔진 잠금 보유 중 거절 · --apply 없이는 쓰기 0 · 적용은 백업 + audit
// (design D3 · D4, 시험 D6 (4) · (5)).

import (
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/audit"
	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/enginelock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
)

// a206SeedJournal 은 스키마를 만든 뒤 닫힌 포지션 하나에 증거가 찬 행 하나를 직접 넣음 — 명령 표면 시험이라 행 모양은 최소임.
func a206SeedJournal(t *testing.T, dir string) string {
	t.Helper()
	path := filepath.Join(dir, journal.DBFileName)
	j, err := journal.Open(context.Background(), journal.Options{Path: path,
		FSProber: journal.FixedFSProber(journal.FSInfo{Name: "ext4", Magic: journal.MagicExt})})
	if err != nil {
		t.Fatal(err)
	}
	if err := j.Close(); err != nil {
		t.Fatal(err)
	}
	db, err := sql.Open("sqlite", "file:"+path)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	for _, stmt := range []string{
		`INSERT INTO positions(id, account_ref, market, symbol, instance_seq, state, quantity, avg_price)
		 VALUES ('pos-a206', 'acct-1', 'kr', '005930', 1, 'CLOSED', '0', '70000')`,
		`INSERT INTO exit_states(position_id, policy_kind, entry_price, initial_stop, initial_risk, baseline_price,
		   high_water, ratchet_level, completed, updated_at)
		 VALUES ('pos-a206', 'RATCHET', '70000', '68000', '2000', '68000', '70000', 'NONE', 1, '2026-03-30T00:30:00Z')`,
		`INSERT INTO exit_events(position_id, created_at, decision_id, saved_snapshot_json, recomputed_snapshot_json,
		   effective_snapshot_json, effective_source)
		 VALUES ('pos-a206', '2026-03-30T00:30:00Z', 'dec-a206', '{"s":1}', '{"r":1}', '{"r":1}', 'recomputed')`,
	} {
		if _, err := db.Exec(stmt); err != nil {
			t.Fatalf("seeding: %v", err)
		}
	}
	return path
}

// a206DirState 는 디렉터리의 파일 이름 → 내용 해시임. 잠금 파일은 명령이 만들 수 있는 유일한 비저널 파일이라 뺌.
func a206DirState(t *testing.T, dir string) map[string]string {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	out := map[string]string{}
	for _, e := range entries {
		if e.Name() == enginelock.LockFileName {
			continue
		}
		raw, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			t.Fatal(err)
		}
		sum := sha256.Sum256(raw)
		out[e.Name()] = hex.EncodeToString(sum[:])
	}
	return out
}

// a206Clock 은 백업 이름(UTC 초)을 정하는 시계임 — 같은 초의 두 번째 백업은 VACUUM INTO 가 거절하므로 실행마다 1 초 옮김.
var a206Clock = clock.NewFake(time.Date(2026, 11, 27, 20, 0, 0, 0, time.UTC))

func a206Run(t *testing.T, dir string, args ...string) (string, error) {
	t.Helper()
	a206Clock.Advance(time.Second)
	deps := productionExitEvidenceCompactDeps()
	deps.clock = a206Clock
	cmd := newEngineExitEvidenceCompactCmdWithDeps(&rootOptions{configDir: dir}, deps)
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs(args)
	err := cmd.ExecuteContext(context.Background())
	return out.String(), err
}

var a206ApplyArgs = []string{"--apply", "--operator", "op", "--approval", "ticket-1", "--note", "engine stopped in the closed window"}

func TestA206CompactCommandIsMutatingAndAttached(t *testing.T) {
	cmd := newEngineExitEvidenceCompactCmd(&rootOptions{})
	if cmd.Annotations["mutating"] != "true" {
		t.Fatalf("annotations = %v, want mutating=true", cmd.Annotations)
	}
	for _, c := range newEngineCmd(&rootOptions{}).Commands() {
		if c.Name() == "exit-evidence-compact" {
			return
		}
	}
	t.Fatal("`tossctl engine exit-evidence-compact` is not attached to `tossctl engine`")
}

// D6 (4) — 엔진이 잠금을 쥐고 있으면 dry-run 이든 --apply 든 아무것도 하지 않고 거절함.
func TestA206CompactRefusesWhileTheEngineHoldsTheLock(t *testing.T) {
	dir := t.TempDir()
	a206SeedJournal(t, dir)
	held, err := enginelock.Acquire(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer held.Release()
	before := a206DirState(t, dir)
	for _, args := range [][]string{nil, a206ApplyArgs} {
		if _, err := a206Run(t, dir, args...); !errors.Is(err, enginelock.ErrAlreadyRunning) {
			t.Fatalf("args %v: err = %v, want the engine lock refusal", args, err)
		}
	}
	if after := a206DirState(t, dir); !mapsEqual(before, after) {
		t.Fatalf("a refused run changed the directory:\nbefore %v\nafter  %v", before, after)
	}
}

// D6 (5) — --apply 가 없으면 집계만 출력하고 저널 · 백업 · audit 어느 것도 쓰지 않음.
func TestA206CompactWithoutApplyWritesNothing(t *testing.T) {
	dir := t.TempDir()
	a206SeedJournal(t, dir)
	before := a206DirState(t, dir)
	out, err := a206Run(t, dir)
	if err != nil {
		t.Fatalf("dry-run: %v", err)
	}
	if !strings.Contains(out, "targets=1") || !strings.Contains(out, "dry-run") {
		t.Fatalf("dry-run output = %q, want the target count and the dry-run label", out)
	}
	// WAL 저널을 읽기 전용으로 열면 SQLite 가 wal-index(-shm)와 빈 -wal 을 만듦(콘솔과 같은 OpenReadOnly 동작) — 데이터가 아님.
	// 그래서 저널 바이트 동일 · -wal 이 비었음 · 그 밖의 새 파일(백업 · audit) 없음을 봄.
	after := a206DirState(t, dir)
	empty := sha256.Sum256(nil)
	if wal, ok := after[journal.DBFileName+"-wal"]; ok && wal != hex.EncodeToString(empty[:]) {
		t.Fatal("dry-run left frames in the WAL")
	}
	delete(after, journal.DBFileName+"-wal")
	delete(after, journal.DBFileName+"-shm")
	if !mapsEqual(before, after) {
		t.Fatalf("dry-run changed the directory:\nbefore %v\nafter  %v", before, after)
	}
}

// --apply 는 운영자 · 승인 참조 · note 가 모두 있어야 하고, 하나라도 비면 잠금 · 저널에 닿기 전에 거절함.
func TestA206CompactApplyNeedsTheApprovalTriple(t *testing.T) {
	dir := t.TempDir()
	a206SeedJournal(t, dir)
	before := a206DirState(t, dir)
	for i := 1; i < len(a206ApplyArgs); i += 2 {
		args := append([]string{}, a206ApplyArgs...)
		args[i+1] = " "
		if _, err := a206Run(t, dir, args...); err == nil || !strings.Contains(err.Error(), "is required") {
			t.Fatalf("without %s: err = %v, want a required-flag refusal", args[i], err)
		}
	}
	if after := a206DirState(t, dir); !mapsEqual(before, after) {
		t.Fatal("a refused --apply changed the directory")
	}
}

// --apply 는 백업을 저널 옆에 남기고, 결과를 audit 에 한 줄 남김(spec 「적용 전 백업」).
func TestA206CompactApplyBacksUpEmptiesAndAudits(t *testing.T) {
	dir := t.TempDir()
	path := a206SeedJournal(t, dir)
	out, err := a206Run(t, dir, a206ApplyArgs...)
	if err != nil {
		t.Fatalf("apply: %v\n%s", err, out)
	}
	if !strings.Contains(out, "emptied=1") {
		t.Fatalf("apply output = %q, want emptied=1", out)
	}
	backups, _ := filepath.Glob(path + ".pre-a206-*.bak")
	if len(backups) != 1 {
		t.Fatalf("backups = %v, want exactly one next to the journal", backups)
	}
	log, err := audit.Open(audit.Options{Path: filepath.Join(dir, audit.FileName)})
	if err != nil {
		t.Fatal(err)
	}
	entries, err := log.Entries()
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 || entries[0].Action != auditActionExitEvidenceCompact || entries[0].Subject != "op" ||
		!strings.Contains(entries[0].Detail, "ticket-1") || !strings.Contains(entries[0].Detail, backups[0]) {
		t.Fatalf("audit = %+v, want one compaction line by op naming the approval and the backup", entries)
	}
	// 두 번째 적용은 0 행.
	if out, err := a206Run(t, dir, a206ApplyArgs...); err != nil || !strings.Contains(out, "emptied=0") {
		t.Fatalf("second apply: %v\n%s", err, out)
	}
}

// --apply 는 저널이 없는 디렉터리(잘못 지정한 --config-dir)에서 새 원장을 만들지 않고 거절함(2단계 리뷰 r1 P2).
func TestA206CompactApplyRefusesAMissingJournal(t *testing.T) {
	dir := t.TempDir()
	if _, err := a206Run(t, dir, a206ApplyArgs...); err == nil || !strings.Contains(err.Error(), "no journal") {
		t.Fatalf("apply on an empty directory: err = %v, want the missing-journal refusal", err)
	}
	if state := a206DirState(t, dir); len(state) != 0 {
		t.Fatalf("a refused apply created %v", state)
	}
}

func mapsEqual(a, b map[string]string) bool {
	if len(a) != len(b) {
		return false
	}
	keys := make([]string, 0, len(a))
	for k := range a {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		if a[k] != b[k] {
			return false
		}
	}
	return true
}
