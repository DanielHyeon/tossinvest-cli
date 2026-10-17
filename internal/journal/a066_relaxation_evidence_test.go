package journal

// a066 5.5 증거 보강 — 리뷰 R3(2026-09-29)가 "선언만 되어 있다"고 잰 D8 주장들을 행동 시험으로 못 박음.
// 각 시험 위 주석의 괄호는 그 주장을 살려 둔 변이(생존) 이름임.

import (
	"context"
	"errors"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/riskbucket"
)

// TestA066RelaxationRecordsAreAppendOnly 는 v35 세 표의 불변 트리거와 lock_seq UNIQUE 를 잼(MX01–MX07).
func TestA066RelaxationRecordsAreAppendOnly(t *testing.T) {
	j := openTestJournal(t)
	activateTestLock(t, j, "a066 daily loss")
	activateTestLock(t, j, "a066 later loss") // REAFFIRM 사건 하나
	view := lockStateFor(t, j)
	if _, err := j.ReleaseEntryLossLock(context.Background(), releaseRequest(view, &relaxationAuditor{})); err != nil {
		t.Fatal(err)
	}
	for name, stmt := range map[string]string{
		"event update":        `UPDATE risk_bucket_entry_loss_lock_events SET cause='rewritten'`,
		"event delete":        `DELETE FROM risk_bucket_entry_loss_lock_events`,
		"lock release update": `UPDATE risk_bucket_entry_loss_lock_releases SET approval='forged'`,
		"lock release delete": `DELETE FROM risk_bucket_entry_loss_lock_releases`,
		"second release of one lock": `INSERT INTO risk_bucket_entry_loss_lock_releases(lock_seq,actor,approval,reason,expected_last_event,released_at)
			SELECT lock_seq,'OPERATOR','again','again',0,'2026-03-30T05:00:00Z' FROM risk_bucket_entry_loss_lock_releases`,
	} {
		if _, err := j.db.Exec(stmt); err == nil {
			t.Errorf("%s was accepted", name)
		}
	}
	if countRiskBucketRows(t, j, "risk_bucket_entry_loss_lock_events") != 1 || countRiskBucketRows(t, j, "risk_bucket_entry_loss_lock_releases") != 1 {
		t.Fatal("append-only rows changed")
	}

	lj, key, _ := overageLatchedOwner(t)
	if _, err := lj.ReleaseRiskOverageLatch(context.Background(), latchRelease(key, ownerLatchView(t, lj, key).StateDigest, &relaxationAuditor{})); err != nil {
		t.Fatal(err)
	}
	for name, stmt := range map[string]string{
		"latch release update": `UPDATE risk_bucket_latch_releases SET approval='forged'`,
		"latch release delete": `DELETE FROM risk_bucket_latch_releases`,
	} {
		if _, err := lj.db.Exec(stmt); err == nil {
			t.Errorf("%s was accepted", name)
		}
	}
}

// TestA066LatchReleaseRefusesAnOwnerWhoseLedgerLeftItsSeal 는 봉인 뒤 원장이 바뀐(재봉인 없는) owner 의 해제가 replay
// mismatch 로 거절되고 아무것도 안 바꿈을 잼(M19 — 이전에는 출구 수 census 만 잡았음).
func TestA066LatchReleaseRefusesAnOwnerWhoseLedgerLeftItsSeal(t *testing.T) {
	j, key, _ := overageLatchedOwner(t)
	seen := ownerLatchView(t, j, key)
	// 봉인 없이 상태가 움직인 모양(R1: 체결 실패 latch 가 봉인을 다시 찍지 않는 경로).
	if _, err := j.db.Exec(`UPDATE risk_bucket_owners SET unknown_actual_latched=0 WHERE account_ref=? AND market=? AND symbol=? AND prospective_generation=?`,
		key.AccountID, string(key.Market), key.Symbol, key.ProspectiveGeneration); err != nil {
		t.Fatal(err)
	}
	_, err := j.ReleaseRiskOverageLatch(context.Background(), latchRelease(key, seen.StateDigest, &relaxationAuditor{}))
	if !errors.Is(err, ErrRiskBucketReplayMismatch) {
		t.Fatalf("err = %v, want replay mismatch", err)
	}
	if overage, _, reservations := ownerFlags(t, j, key); overage != 1 || reservations == 0 || countRiskBucketRows(t, j, "risk_bucket_latch_releases") != 0 {
		t.Fatalf("a refused release changed the owner: overage=%d reservations=%d", overage, reservations)
	}
}

// TestA066LatchReleaseKeepsOverageAmountsAndOtherGenerations 는 해제가 overage 금액을 남기고(MX18) 같은 종목의 다른
// generation 행을 건드리지 않음을(MX17) 잼.
func TestA066LatchReleaseKeepsOverageAmountsAndOtherGenerations(t *testing.T) {
	j, key, decisionID := overageLatchedOwner(t)
	var before string
	if err := j.db.QueryRow(`SELECT group_concat(overage_minor,',') FROM (SELECT overage_minor FROM risk_bucket_reservations WHERE decision_id=? ORDER BY bucket_dimension)`, decisionID).Scan(&before); err != nil {
		t.Fatal(err)
	}
	// 같은 종목의 다른 owner generation 이 latch 를 가진 행 하나를 복제로 세움(그 generation 의 owner 봉인은 이 owner 의
	// digest 에 들어가지 않음).
	for _, stmt := range []string{
		`CREATE TEMP TABLE other_generation AS SELECT * FROM risk_bucket_reservations WHERE decision_id='` + decisionID + `' AND bucket_dimension='sector'`,
		`UPDATE other_generation SET reservation_id='other-reservation', decision_id='other-decision', owner_prospective_generation='other-generation', risk_overage_latched=1`,
		// 다른 generation 의 결정·owner 를 통째로 세우는 대신 행 하나만 위조함 — 외래 키를 이 삽입 동안만 끔(시험 픽스처).
		`PRAGMA foreign_keys=OFF`,
		`INSERT INTO risk_bucket_reservations SELECT * FROM other_generation`,
		`PRAGMA foreign_keys=ON`,
	} {
		if _, err := j.db.Exec(stmt); err != nil {
			t.Fatalf("%s: %v", stmt, err)
		}
	}
	if _, err := j.ReleaseRiskOverageLatch(context.Background(), latchRelease(key, ownerLatchView(t, j, key).StateDigest, &relaxationAuditor{})); err != nil {
		t.Fatal(err)
	}
	var otherLatched int
	if err := j.db.QueryRow(`SELECT risk_overage_latched FROM risk_bucket_reservations WHERE reservation_id='other-reservation'`).Scan(&otherLatched); err != nil {
		t.Fatal(err)
	}
	if otherLatched != 1 {
		t.Fatal("the release cleared another generation's RISK_OVERAGE")
	}
	var after string
	if err := j.db.QueryRow(`SELECT group_concat(overage_minor,',') FROM (SELECT overage_minor FROM risk_bucket_reservations WHERE decision_id=? ORDER BY bucket_dimension)`, decisionID).Scan(&after); err != nil {
		t.Fatal(err)
	}
	if after != before {
		t.Fatalf("overage amounts changed: %s → %s", before, after)
	}
}

// TestA066LatchReleaseOfAnOwnerThatIsGoneIsStale 는 해제된 owner 와 없는 owner 에 대한 해제가 stale 로 거절됨을 잼
// (MX20, M22b).
func TestA066LatchReleaseOfAnOwnerThatIsGoneIsStale(t *testing.T) {
	t.Run("unknown generation", func(t *testing.T) {
		j, key, _ := overageLatchedOwner(t)
		view := ownerLatchView(t, j, key)
		key.ProspectiveGeneration = "no-such-generation"
		if _, err := j.ReleaseRiskOverageLatch(context.Background(), latchRelease(key, view.StateDigest, &relaxationAuditor{})); !errors.Is(err, ErrRiskRelaxationStale) {
			t.Fatalf("err = %v, want stale", err)
		}
	})
	t.Run("released owner", func(t *testing.T) {
		j, key, _ := overageLatchedOwner(t)
		view := ownerLatchView(t, j, key)
		if _, err := j.db.Exec(`UPDATE risk_bucket_owners SET released_at='2026-03-30T04:00:00Z' WHERE account_ref=? AND market=? AND symbol=? AND prospective_generation=?`,
			key.AccountID, string(key.Market), key.Symbol, key.ProspectiveGeneration); err != nil {
			t.Fatal(err)
		}
		if _, err := j.ReleaseRiskOverageLatch(context.Background(), latchRelease(key, view.StateDigest, &relaxationAuditor{})); !errors.Is(err, ErrRiskRelaxationStale) {
			t.Fatalf("err = %v, want stale", err)
		}
		if countRiskBucketRows(t, j, "risk_bucket_latch_releases") != 0 {
			t.Fatal("a release was recorded for a released owner")
		}
	})
}

// TestA066EntryReleaseBoundToAnotherScopesLockIsStale 는 다른 범위의 **실제로 열린** 잠금 번호로 한 해제가 stale 임을
// 잼(M03 — 이전에는 없는 번호의 외래 키 실패로만 잡혔음).
func TestA066EntryReleaseBoundToAnotherScopesLockIsStale(t *testing.T) {
	j := openTestJournal(t)
	activateTestLock(t, j, "a066 short loss")
	medium, _, err := j.ActivateEntryLossLock(context.Background(), EntryLossLock{AccountRef: "acct-1", Market: riskbucket.MarketKR,
		Horizon: riskbucket.HorizonMedium, Cause: "a066 medium loss", ActivatedAt: relaxNow})
	if err != nil {
		t.Fatal(err)
	}
	req := releaseRequest(lockStateFor(t, j), &relaxationAuditor{})
	req.LockSeq = medium.Seq // SHORT 범위 요청에 MEDIUM 의 열린 잠금 번호
	if _, err := j.ReleaseEntryLossLock(context.Background(), req); !errors.Is(err, ErrRiskRelaxationStale) {
		t.Fatalf("err = %v, want stale", err)
	}
	if countRiskBucketRows(t, j, "risk_bucket_entry_loss_lock_releases") != 0 {
		t.Fatal("a release was recorded against another scope's lock")
	}
}

// TestA066ReadOnlyViewsShowWhatTheWriterSees 는 risk-latch-show 가 쓰는 읽기 전용 경로가 원장 쪽 값(마지막 사건 번호,
// 봉인 digest)을 그대로 보여 줌을 잼(MX11, MX12).
func TestA066ReadOnlyViewsShowWhatTheWriterSees(t *testing.T) {
	ctx := context.Background()
	j, key, _ := overageLatchedOwner(t)
	activateTestLock(t, j, "a066 daily loss")
	activateTestLock(t, j, "a066 later loss")
	ro, err := OpenReadOnly(ctx, ReadOnlyOptions{Path: j.path})
	if err != nil {
		t.Fatal(err)
	}
	defer ro.Close()
	locks, err := ro.ReadEntryLossLocks(ctx, "acct-1")
	if err != nil || len(locks) != 1 {
		t.Fatalf("read-only locks = %+v (%v)", locks, err)
	}
	if want := lockStateFor(t, j); locks[0].LastEvent == 0 || locks[0].LastEvent != want.LastEvent || locks[0].Lock.Seq != want.Lock.Seq {
		t.Fatalf("read-only lock = %+v, writer sees %+v", locks[0], want)
	}
	owners, err := ro.ReadRiskOwnerLatches(ctx, key.AccountID, key.Market)
	if err != nil || len(owners) != 1 {
		t.Fatalf("read-only owners = %+v (%v)", owners, err)
	}
	if want := ownerLatchView(t, j, key); owners[0] != want || owners[0].StateDigest == "" {
		t.Fatalf("read-only owner = %+v, writer sees %+v", owners[0], want)
	}
}
