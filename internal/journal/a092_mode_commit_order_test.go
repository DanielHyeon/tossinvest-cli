package journal

// a092 착지 단위 ④ — 22.3 C4 · C5 · 23.3 K14: 「현재 모드」 · 방향 판정 · 기동 복원 · 이력은 **커밋 순서(rowid)** 하나를 씀.
// 벽시계로 정하면 시각을 먼저 얻고 늦게 커밋된 전이가 「과거」가 되어 재시작 복원이 산 게이트와 다른 모드를 세울 수 있음(fail-open).

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
)

const a092ModeAccount = "acct-a092"

type a092ModeAuditor struct{}

func (a092ModeAuditor) RecordAction(string, string, string, string) error { return nil }

func a092ModeJournal(t *testing.T, path string) (*Journal, *clock.Fake, *recordingProjector) {
	t.Helper()
	clk := clock.NewFake(time.Date(2026, 9, 29, 3, 0, 0, 0, time.UTC))
	j, err := Open(context.Background(), Options{
		Path: path, Clock: clk, FSProber: FixedFSProber(FSInfo{Name: "ext4", Magic: MagicExt}),
	})
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = j.Close() })
	p := &recordingProjector{}
	if err := j.SetModeProjector(p); err != nil {
		t.Fatalf("SetModeProjector: %v", err)
	}
	return j, clk, p
}

func a092Tighten(t *testing.T, j *Journal, trigger string) OperatingModeRecord {
	t.Helper()
	rec, changed, err := j.EscalateOperatingMode(context.Background(), a092ModeAccount, trigger, nil)
	if err != nil || !changed {
		t.Fatalf("tighten: changed=%v err=%v", changed, err)
	}
	return rec
}

func a092Relax(t *testing.T, j *Journal) OperatingModeRecord {
	t.Helper()
	rec, changed, err := j.TransitionOperatingMode(context.Background(), TransitionModeRequest{
		AccountRef: a092ModeAccount, Mode: ModeNormal, Cause: "checked", Actor: ModeActorOperator,
		Approval: "OPS-1", Auditor: a092ModeAuditor{},
	})
	if err != nil || !changed {
		t.Fatalf("relax: changed=%v err=%v", changed, err)
	}
	return rec
}

// C4: 벽시계가 뒤로 간 뒤 커밋된 강화가 「현재」다 — 시각으로 정하면 앞선 완화가 현재로 읽혀 진입이 열린다.
func TestA092CurrentModeIsTheLatestCommitNotTheLatestClock(t *testing.T) {
	j, clk, _ := a092ModeJournal(t, filepath.Join(t.TempDir(), "journal.db"))
	a092Tighten(t, j, ModeTriggerExitObservationOutage)
	clk.Advance(10 * time.Second)
	a092Relax(t, j)
	clk.Advance(-20 * time.Second) // 벽시계 되감김
	a092Tighten(t, j, ModeTriggerCredentialRejected)

	if cur := currentMode(t, j, a092ModeAccount); cur.Mode != ModeEntryBlocked {
		t.Fatalf("current mode = %s, want ENTRY_BLOCKED — the latest *committed* transition is a tightening", cur.Mode)
	}
	// 방향 판정도 같은 순서: 같은 목적 모드로의 AUTO 재강화는 변화 없음이어야 함(현재가 이미 ENTRY_BLOCKED).
	if _, changed, err := j.EscalateOperatingMode(context.Background(), a092ModeAccount,
		ModeTriggerExitObservationOutage, nil); err != nil || changed {
		t.Errorf("re-tighten after the rewound one: changed=%v err=%v, want no change", changed, err)
	}
}

// C5 · 울타리: 전이 레코드와 복원 레코드는 커밋 순번(rowid)을 싣고, 순번은 커밋 순서대로 증가함.
func TestA092TransitionsCarryTheirCommitSequence(t *testing.T) {
	j, clk, p := a092ModeJournal(t, filepath.Join(t.TempDir(), "journal.db"))
	first := a092Tighten(t, j, ModeTriggerExitObservationOutage)
	clk.Advance(-time.Hour)
	second := a092Relax(t, j)
	if first.Seq <= 0 || second.Seq <= first.Seq {
		t.Fatalf("seq first=%d second=%d, want positive and increasing in commit order", first.Seq, second.Seq)
	}
	seen := p.records()
	if len(seen) != 2 || seen[0].Seq != first.Seq || seen[1].Seq != second.Seq {
		t.Fatalf("projected seqs = %+v, want %d then %d", seen, first.Seq, second.Seq)
	}
	snap, err := j.RestoreOperatingModeProjection(context.Background(), a092ModeAccount)
	if err != nil {
		t.Fatalf("restore: %v", err)
	}
	seen = p.records()
	if snap.Seq != second.Seq || seen[len(seen)-1].Seq != second.Seq || seen[len(seen)-1].Mode != ModeNormal {
		t.Errorf("restore seq snap=%d projected=%+v, want %d NORMAL", snap.Seq, seen[len(seen)-1], second.Seq)
	}
}

// K14: 이력도 같은 순서(커밋 순) — 벽시계 되감김과 무관.
func TestA092HistoryIsInCommitOrder(t *testing.T) {
	j, clk, _ := a092ModeJournal(t, filepath.Join(t.TempDir(), "journal.db"))
	a := a092Tighten(t, j, ModeTriggerExitObservationOutage)
	clk.Advance(-time.Hour)
	b := a092Relax(t, j)
	clk.Advance(-time.Hour)
	c := a092Tighten(t, j, ModeTriggerCredentialRejected)
	hist, err := j.OperatingModeHistory(context.Background(), a092ModeAccount)
	if err != nil {
		t.Fatal(err)
	}
	if len(hist) != 3 || hist[0].ID != a.ID || hist[1].ID != b.ID || hist[2].ID != c.ID {
		ids := []string{}
		for _, h := range hist {
			ids = append(ids, h.ID)
		}
		t.Fatalf("history = %v, want commit order %s %s %s", ids, a.ID, b.ID, c.ID)
	}
}

// K14: `VACUUM INTO` 복사본에서 연 원장도 현재 모드와 이력 순서가 그대로(울타리 초기값 0 과 함께 복원이 최신을 세움).
func TestA092CommitOrderSurvivesAVacuumCopy(t *testing.T) {
	dir := t.TempDir()
	j, clk, _ := a092ModeJournal(t, filepath.Join(dir, "journal.db"))
	a := a092Tighten(t, j, ModeTriggerExitObservationOutage)
	clk.Advance(-time.Hour)
	b := a092Relax(t, j)
	clk.Advance(-time.Hour)
	c := a092Tighten(t, j, ModeTriggerCredentialRejected)
	copyPath := filepath.Join(dir, "copy.db")
	if _, err := j.db.ExecContext(context.Background(), "VACUUM INTO ?", copyPath); err != nil {
		t.Fatalf("VACUUM INTO: %v", err)
	}
	db, err := sql.Open("sqlite", copyPath)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	rows, err := db.Query(`SELECT id FROM operating_modes WHERE account_ref = ? ORDER BY rowid`, a092ModeAccount)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	var got []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			t.Fatal(err)
		}
		got = append(got, id)
	}
	if len(got) != 3 || got[0] != a.ID || got[1] != b.ID || got[2] != c.ID {
		t.Errorf("rowid order in the VACUUM copy = %v, want %s %s %s", got, a.ID, b.ID, c.ID)
	}
}
