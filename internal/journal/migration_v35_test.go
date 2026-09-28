package journal

// a066 5.5(D8): v34 저널이 v35 로 넘어와도 기존 잠금은 그대로 효력 있고(해제 기록 없음 = 열림), v33 의
// first-cause-wins 트리거는 "열린 잠금 1건" 트리거로 바뀌며, 새 표 셋은 비어서 시작함.

import (
	"context"
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/riskbucket"
)

func TestMigrationV34ToV35KeepsExistingLocksInForceAndSwapsTheTrigger(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "journal.db")
	old := openJournalAtSchema(t, path, 34)
	if _, err := old.db.Exec(`INSERT INTO risk_bucket_entry_loss_locks(account_ref,market,horizon,cause,activated_at) VALUES('acct-1','KR','SHORT','v34 loss','2026-03-30T00:00:00Z')`); err != nil {
		t.Fatal(err)
	}
	if err := old.Close(); err != nil {
		t.Fatal(err)
	}
	j := openJournalAtSchema(t, path, 35)
	defer j.Close()
	if version, err := j.SchemaVersion(ctx); err != nil || version != 35 {
		t.Fatalf("version=%d err=%v", version, err)
	}
	if err := refuseEntryUnderLossLock(ctx, j.db, "acct-1", riskbucket.MarketKR, riskbucket.HorizonShort); err == nil {
		t.Fatal("a v34 lock is not in force after the migration")
	}
	for _, table := range []string{"risk_bucket_entry_loss_lock_events", "risk_bucket_entry_loss_lock_releases", "risk_bucket_latch_releases"} {
		if got := countRiskBucketRows(t, j, table); got != 0 {
			t.Fatalf("%s starts with %d rows", table, got)
		}
	}
	var triggers []string
	rows, err := j.db.Query(`SELECT name FROM sqlite_master WHERE type='trigger' AND tbl_name='risk_bucket_entry_loss_locks' ORDER BY name`)
	if err != nil {
		t.Fatal(err)
	}
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			t.Fatal(err)
		}
		triggers = append(triggers, name)
	}
	rows.Close()
	if got := strings.Join(triggers, ","); got != "risk_bucket_entry_loss_lock_no_delete,risk_bucket_entry_loss_lock_no_update,risk_bucket_entry_loss_lock_one_open" {
		t.Fatalf("lock triggers after v35: %s", got)
	}
}

// TestReadOnlyRelaxationViewsRefuseAPreV35Journal 은 risk-latch-show 가 v35 이전 저널에서 날것의 "no such table"
// 대신 타입 있는 ErrSchemaTooOld 로 거절함을 잼(리뷰 R2 P3).
func TestReadOnlyRelaxationViewsRefuseAPreV35Journal(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "journal.db")
	old := openJournalAtSchema(t, path, 34)
	if err := old.Close(); err != nil {
		t.Fatal(err)
	}
	ro, err := OpenReadOnly(ctx, ReadOnlyOptions{Path: path})
	if err != nil {
		t.Fatalf("OpenReadOnly: %v", err)
	}
	defer ro.Close()
	if _, err := ro.ReadEntryLossLocks(ctx, "acct-1"); !errors.Is(err, ErrSchemaTooOld) {
		t.Fatalf("locks on v34: %v, want ErrSchemaTooOld", err)
	}
	if _, err := ro.ReadRiskOwnerLatches(ctx, "acct-1", riskbucket.MarketKR); !errors.Is(err, ErrSchemaTooOld) {
		t.Fatalf("latches on v34: %v, want ErrSchemaTooOld", err)
	}
}
