package reconcile_test

// a198 보강 2(D9) — 기동은 ACKED 로 남은 취소의 쓰기를 마침: 기록 번호로 CONFIRMED, 읽기 0, 알림 0, 복구 보고에 포함.
// 확정 쓰기가 실패하면 ACKED 그대로 · 명명 알림 · 복구 성공(발주의 B6 과 같은 처리).

import (
	"context"
	"database/sql"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/reconcile"
)

func TestA198AnAckedCancelIsCompletedAtBoot(t *testing.T) {
	path := filepath.Join(t.TempDir(), "journal.db")
	a094CrashAfterAck(t, path, journal.KindCancel, "C-O-target")
	orders := &a094CountingOrders{}
	alerts := &a094Alerts{}
	j, report := a094Recover(t, path, orders, alerts)
	stored, err := j.LookupAttempt(context.Background(), "attempt-acked")
	if err != nil {
		t.Fatal(err)
	}
	if stored.State != journal.StateConfirmed || stored.BrokerOrderID != "C-O-target" || stored.TargetOrderID != "O-target" ||
		stored.ReasonCode != reconcile.ReasonBootAckedCancelCompleted {
		t.Fatalf("stored %+v, want CONFIRMED with the recorded number and reason %s", stored, reconcile.ReasonBootAckedCancelCompleted)
	}
	if orders.reads != 0 || orders.lists != 0 || len(alerts.events) != 0 {
		t.Fatalf("reads %d lists %d alerts %v — completing an acknowledged cancel reads nothing and alerts nobody",
			orders.reads, orders.lists, alerts.events)
	}
	if !slices.Contains(report.ConfirmedAcked, "attempt-acked") {
		t.Fatalf("report.ConfirmedAcked = %v, want the cancel", report.ConfirmedAcked)
	}
}

func TestA198AFailedAckedCancelWriteLeavesTheAckAndNamesIt(t *testing.T) {
	path := filepath.Join(t.TempDir(), "journal.db")
	a094CrashAfterAck(t, path, journal.KindCancel, "C-O-target")
	db, err := sql.Open("sqlite", "file:"+path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`CREATE TRIGGER a198_refuse_confirm BEFORE UPDATE OF state ON mutation_attempts
		WHEN NEW.state = 'CONFIRMED' BEGIN SELECT RAISE(ABORT, 'a198 fixture: the ledger refuses the write'); END`); err != nil {
		t.Fatalf("installing the fixture trigger: %v", err)
	}
	_ = db.Close()
	alerts := &a094Alerts{}
	j, report := a094Recover(t, path, &a094CountingOrders{}, alerts)
	stored, _ := j.LookupAttempt(context.Background(), "attempt-acked")
	if stored.State != journal.StateAcked || len(report.ConfirmedAcked) != 0 {
		t.Fatalf("state %s confirmed %v, want ACKED and nothing confirmed", stored.State, report.ConfirmedAcked)
	}
	a094AssertNamed(t, alerts, "attempt-acked")
	if !strings.Contains(alerts.events[0].Body, "ledger write failed") {
		t.Errorf("the alert does not say the write failed: %q", alerts.events[0].Body)
	}
}
