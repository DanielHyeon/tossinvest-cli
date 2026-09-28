package journal

// a066 5.5 완화(해제) — 사용자 결정 2026-09-28, design D8.
//
// 원칙: 자동은 조이기만 / 완화·해제는 OPERATOR + 승인 참조 + commit 전 audit / 동시 조이기는 보수 쪽 승리 /
// 진입점은 journal API + tossctl mutating 명령. 이 파일은 journal API 의 계약을 시험 저널에서만 잼.

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/riskbucket"
)

// relaxationAuditor 는 audit 줄을 모으고, fail 이면 쓰기를 실패시킴(commit 앞 audit 의 증거).
type relaxationAuditor struct {
	lines []string
	fail  bool
}

func (a *relaxationAuditor) RecordAction(action, setting, value, detail string) error {
	if a.fail {
		return errors.New("audit log unavailable")
	}
	a.lines = append(a.lines, strings.Join([]string{action, setting, value, detail}, " | "))
	return nil
}

var relaxNow = time.Date(2026, 3, 30, 2, 0, 0, 0, time.UTC)

func activateTestLock(t *testing.T, j *Journal, cause string) EntryLossLock {
	t.Helper()
	lock, _, err := j.ActivateEntryLossLock(context.Background(), EntryLossLock{AccountRef: "acct-1", Market: riskbucket.MarketKR,
		Horizon: riskbucket.HorizonShort, Cause: cause, ActivatedAt: relaxNow})
	if err != nil {
		t.Fatal(err)
	}
	return lock
}

func lockStateFor(t *testing.T, j *Journal) EntryLossLockView {
	t.Helper()
	views, err := j.ReadEntryLossLocks(context.Background(), "acct-1")
	if err != nil {
		t.Fatal(err)
	}
	for _, view := range views {
		if view.Lock.Market == riskbucket.MarketKR && view.Lock.Horizon == riskbucket.HorizonShort {
			return view
		}
	}
	t.Fatalf("no KR/SHORT lock in %+v", views)
	return EntryLossLockView{}
}

func releaseRequest(view EntryLossLockView, auditor *relaxationAuditor) EntryLossLockReleaseRequest {
	return EntryLossLockReleaseRequest{AccountRef: "acct-1", Market: riskbucket.MarketKR, Horizon: riskbucket.HorizonShort,
		LockSeq: view.Lock.Seq, ExpectedLastEvent: view.LastEvent, Actor: RelaxationActorOperator,
		Approval: "OPS-2026-0930 approved by risk owner", Reason: "loss review complete", ReleasedAt: relaxNow.Add(time.Hour), Auditor: auditor}
}

func TestA066EntryLossLockReleaseIsOperatorApprovedAuditedAndOpensEntry(t *testing.T) {
	j := openTestJournal(t)
	activateTestLock(t, j, "a066 daily loss")
	ctx := context.Background()
	if err := refuseEntryUnderLossLock(ctx, j.db, "acct-1", riskbucket.MarketKR, riskbucket.HorizonShort); !errors.Is(err, ErrRiskBucketEntryLossLocked) {
		t.Fatalf("lock not in force before release: %v", err)
	}
	view := lockStateFor(t, j)
	auditor := &relaxationAuditor{}
	record, err := j.ReleaseEntryLossLock(ctx, releaseRequest(view, auditor))
	if err != nil {
		t.Fatalf("approved release: %v", err)
	}
	if record.LockSeq != view.Lock.Seq || len(auditor.lines) != 1 || !strings.Contains(auditor.lines[0], AuditActionEntryLockRelease) ||
		!strings.Contains(auditor.lines[0], "entry_loss_lock:acct-1/KR/SHORT") || !strings.Contains(auditor.lines[0], "OPS-2026-0930") {
		t.Fatalf("release record=%+v audit=%v", record, auditor.lines)
	}
	if err := refuseEntryUnderLossLock(ctx, j.db, "acct-1", riskbucket.MarketKR, riskbucket.HorizonShort); err != nil {
		t.Fatalf("released lock still refuses entry: %v", err)
	}
	// 해제 뒤의 활성화는 새 잠금을 엶(자동은 조이기만).
	next := activateTestLock(t, j, "a066 second loss")
	if next.Seq == view.Lock.Seq {
		t.Fatalf("activation after release reused lock %d", next.Seq)
	}
	if err := refuseEntryUnderLossLock(ctx, j.db, "acct-1", riskbucket.MarketKR, riskbucket.HorizonShort); !errors.Is(err, ErrRiskBucketEntryLossLocked) {
		t.Fatalf("new lock not in force: %v", err)
	}
}

func TestA066EntryLossLockReleaseRefusals(t *testing.T) {
	for _, tc := range []struct {
		name   string
		mutate func(*EntryLossLockReleaseRequest, *relaxationAuditor)
		want   error
	}{
		{"automatic actor cannot relax", func(r *EntryLossLockReleaseRequest, _ *relaxationAuditor) { r.Actor = "AUTO" }, ErrRiskRelaxationRequiresOperator},
		{"approval reference required", func(r *EntryLossLockReleaseRequest, _ *relaxationAuditor) { r.Approval = "  " }, ErrRiskRelaxationApprovalRequired},
		{"auditor required", func(r *EntryLossLockReleaseRequest, _ *relaxationAuditor) { r.Auditor = nil }, ErrInvalidRequest},
		{"reason required", func(r *EntryLossLockReleaseRequest, _ *relaxationAuditor) { r.Reason = "" }, ErrInvalidRequest},
		{"a lock the operator did not see", func(r *EntryLossLockReleaseRequest, _ *relaxationAuditor) { r.LockSeq += 99 }, ErrRiskRelaxationStale},
		{"failing audit changes nothing", func(_ *EntryLossLockReleaseRequest, a *relaxationAuditor) { a.fail = true }, nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			j := openTestJournal(t)
			activateTestLock(t, j, "a066 daily loss")
			auditor := &relaxationAuditor{}
			req := releaseRequest(lockStateFor(t, j), auditor)
			tc.mutate(&req, auditor)
			_, err := j.ReleaseEntryLossLock(context.Background(), req)
			if err == nil || (tc.want != nil && !errors.Is(err, tc.want)) {
				t.Fatalf("err=%v, want %v", err, tc.want)
			}
			if err := refuseEntryUnderLossLock(context.Background(), j.db, "acct-1", riskbucket.MarketKR, riskbucket.HorizonShort); !errors.Is(err, ErrRiskBucketEntryLossLocked) {
				t.Fatalf("a refused release opened entry: %v", err)
			}
			if got := countRiskBucketRows(t, j, "risk_bucket_entry_loss_lock_releases"); got != 0 {
				t.Fatalf("a refused release wrote %d release rows", got)
			}
		})
	}
}

func TestA066ReaffirmAfterTheOperatorLookedMakesTheReleaseStale(t *testing.T) {
	j := openTestJournal(t)
	activateTestLock(t, j, "a066 daily loss")
	seen := lockStateFor(t, j)
	// 운영자가 본 뒤에 같은 범위로 새 조이기가 들어옴 — 잠금은 이미 열려 있으므로 REAFFIRM 으로 남음.
	again, changed, err := j.ActivateEntryLossLock(context.Background(), EntryLossLock{AccountRef: "acct-1", Market: riskbucket.MarketKR,
		Horizon: riskbucket.HorizonShort, Cause: "a066 later loss", ActivatedAt: relaxNow.Add(time.Minute)})
	if err != nil || changed || again.Seq != seen.Lock.Seq {
		t.Fatalf("reaffirm: lock=%+v changed=%v err=%v", again, changed, err)
	}
	if after := lockStateFor(t, j); after.LastEvent <= seen.LastEvent {
		t.Fatalf("reaffirm left no event: before=%d after=%d", seen.LastEvent, after.LastEvent)
	}
	_, err = j.ReleaseEntryLossLock(context.Background(), releaseRequest(seen, &relaxationAuditor{}))
	if !errors.Is(err, ErrRiskRelaxationStale) {
		t.Fatalf("release bound to the older view: err=%v, want stale", err)
	}
	if err := refuseEntryUnderLossLock(context.Background(), j.db, "acct-1", riskbucket.MarketKR, riskbucket.HorizonShort); !errors.Is(err, ErrRiskBucketEntryLossLocked) {
		t.Fatalf("stale release opened entry: %v", err)
	}
	// 새로 본 상태로는 해제됨(fail-closed 가 정상 입력을 거부하지 않음).
	if _, err := j.ReleaseEntryLossLock(context.Background(), releaseRequest(lockStateFor(t, j), &relaxationAuditor{})); err != nil {
		t.Fatalf("release bound to the fresh view: %v", err)
	}
}

func TestA066AtMostOneOpenLockPerScopeIsEnforcedByTheSchema(t *testing.T) {
	j := openTestJournal(t)
	activateTestLock(t, j, "a066 daily loss")
	insert := `INSERT INTO risk_bucket_entry_loss_locks(account_ref,market,horizon,cause,activated_at) VALUES('acct-1','KR','SHORT','raw writer','2026-03-30T02:00:00Z')`
	if _, err := j.db.Exec(insert); err == nil || !strings.Contains(err.Error(), "open") {
		t.Fatalf("a second open lock was accepted: %v", err)
	}
	if _, err := j.ReleaseEntryLossLock(context.Background(), releaseRequest(lockStateFor(t, j), &relaxationAuditor{})); err != nil {
		t.Fatal(err)
	}
	if _, err := j.db.Exec(insert); err != nil {
		t.Fatalf("after the release a new lock must be insertable: %v", err)
	}
	if _, err := j.db.Exec(`DELETE FROM risk_bucket_entry_loss_lock_releases`); err == nil {
		t.Fatal("a release record was deleted")
	}
}

// overageLatchedOwner 는 한 owner 에 RISK_OVERAGE(실제 가격 40 체결: filled 160 + HELD 30 > 한도 100)를 세우고, 뒤이은
// 실제 증거 없는 체결로 UNKNOWN_ACTUAL_RISK 도 세움.
func overageLatchedOwner(t *testing.T) (*Journal, riskbucket.OwnerKey, string) {
	t.Helper()
	ctx := context.Background()
	j, key, decisionID, reserved := riskBucketFillFixture(t, "relax-overage", "risk-relax-overage")
	if err := j.RegisterRiskBucketOrder(ctx, RiskBucketOrderPlan{OrderID: "risk-relax-overage", DecisionID: decisionID, OrderQuantity: 10, ReservedMinor: reserved, ReservationPolicyDigest: "reservation-policy-v1", QuoteCurrency: "KRW", BaseCurrency: "KRW", CreatedAt: riskFillNow}); err != nil {
		t.Fatal(err)
	}
	if _, err := j.RecordFill(ctx, observation("risk-relax-overage", "4")); err != nil {
		t.Fatal(err)
	}
	if _, err := j.completeRiskBucketFillActual(ctx, RiskBucketActualFillPlan{Owner: key, DecisionID: decisionID, OrderID: "risk-relax-overage", CumulativeFill: 4, Actual: riskBucketActual("40", "1", "0"), ObservedAt: riskFillNow}); err != nil {
		t.Fatal(err)
	}
	if _, err := j.RecordFill(ctx, observation("risk-relax-overage", "5")); err != nil {
		t.Fatal(err)
	}
	return j, key, decisionID
}

func ownerLatchView(t *testing.T, j *Journal, key riskbucket.OwnerKey) RiskOwnerLatchView {
	t.Helper()
	views, err := j.ReadRiskOwnerLatches(context.Background(), key.AccountID, key.Market)
	if err != nil {
		t.Fatal(err)
	}
	for _, view := range views {
		if view.Owner == key {
			return view
		}
	}
	t.Fatalf("owner %+v not listed in %+v", key, views)
	return RiskOwnerLatchView{}
}

func latchRelease(key riskbucket.OwnerKey, digest string, auditor *relaxationAuditor) RiskOverageLatchReleaseRequest {
	return RiskOverageLatchReleaseRequest{Owner: key, ExpectedStateDigest: digest, Actor: RelaxationActorOperator,
		Approval: "OPS-2026-0931 approved by risk owner", Reason: "overage reviewed; exposure reduced", ReleasedAt: relaxNow.Add(time.Hour), Auditor: auditor}
}

func ownerFlags(t *testing.T, j *Journal, key riskbucket.OwnerKey) (overage, unknown, reservationOverage int) {
	t.Helper()
	if err := j.db.QueryRow(`SELECT risk_overage_latched,unknown_actual_latched FROM risk_bucket_owners WHERE account_ref=? AND market=? AND symbol=? AND prospective_generation=?`,
		key.AccountID, string(key.Market), key.Symbol, key.ProspectiveGeneration).Scan(&overage, &unknown); err != nil {
		t.Fatal(err)
	}
	if err := j.db.QueryRow(`SELECT count(*) FROM risk_bucket_reservations WHERE account_ref=? AND market=? AND symbol=? AND owner_prospective_generation=? AND risk_overage_latched=1`,
		key.AccountID, string(key.Market), key.Symbol, key.ProspectiveGeneration).Scan(&reservationOverage); err != nil {
		t.Fatal(err)
	}
	return overage, unknown, reservationOverage
}

func TestA066OverageLatchReleaseClearsOnlyRiskOverageAndIsAudited(t *testing.T) {
	j, key, _ := overageLatchedOwner(t)
	view := ownerLatchView(t, j, key)
	if !view.OverageLatched || !view.UnknownLatched || view.StateDigest == "" {
		t.Fatalf("fixture view=%+v, want both latches and a digest", view)
	}
	auditor := &relaxationAuditor{}
	if _, err := j.ReleaseRiskOverageLatch(context.Background(), latchRelease(key, view.StateDigest, auditor)); err != nil {
		t.Fatalf("approved latch release: %v", err)
	}
	overage, unknown, reservationOverage := ownerFlags(t, j, key)
	if overage != 0 || reservationOverage != 0 || unknown != 1 {
		t.Fatalf("after release: owner overage=%d reservations with overage=%d unknown=%d, want 0/0/1", overage, reservationOverage, unknown)
	}
	if len(auditor.lines) != 1 || !strings.Contains(auditor.lines[0], AuditActionOverageLatchRelease) ||
		!strings.Contains(auditor.lines[0], "risk_owner:"+key.AccountID+"/"+string(key.Market)+"/"+key.Symbol+"/"+key.ProspectiveGeneration) {
		t.Fatalf("audit=%v", auditor.lines)
	}
	// 해제 뒤 상태는 다시 봉인돼 있어야 함 — 체결 계상·admission 의 digest 대조가 깨지지 않음.
	if err := verifyRiskBucketStateDigest(context.Background(), j.db, key); err != nil {
		t.Fatalf("state not resealed after the release: %v", err)
	}
	if got := countRiskBucketRows(t, j, "risk_bucket_latch_releases"); got != 1 {
		t.Fatalf("release records=%d", got)
	}
}

func TestA066OverageLatchReleaseIsBoundToTheStateTheOperatorRead(t *testing.T) {
	j, key, _ := overageLatchedOwner(t)
	seen := ownerLatchView(t, j, key)
	// ABA: 운영자가 digest 를 읽은 뒤 체결이 끼어 봉인이 다시 찍힘.
	if _, err := j.RecordFill(context.Background(), observation("risk-relax-overage", "6")); err != nil {
		t.Fatal(err)
	}
	_, err := j.ReleaseRiskOverageLatch(context.Background(), latchRelease(key, seen.StateDigest, &relaxationAuditor{}))
	if !errors.Is(err, ErrRiskRelaxationStale) {
		t.Fatalf("release bound to a digest older than the interleaved fill: err=%v, want stale", err)
	}
	if overage, _, _ := ownerFlags(t, j, key); overage != 1 {
		t.Fatal("stale release cleared the latch")
	}
}

func TestA066OverageLatchReleaseRefusals(t *testing.T) {
	for _, tc := range []struct {
		name   string
		mutate func(*RiskOverageLatchReleaseRequest, *relaxationAuditor)
		want   error
	}{
		{"automatic actor cannot relax", func(r *RiskOverageLatchReleaseRequest, _ *relaxationAuditor) { r.Actor = "AUTO" }, ErrRiskRelaxationRequiresOperator},
		{"approval reference required", func(r *RiskOverageLatchReleaseRequest, _ *relaxationAuditor) { r.Approval = "" }, ErrRiskRelaxationApprovalRequired},
		{"auditor required", func(r *RiskOverageLatchReleaseRequest, _ *relaxationAuditor) { r.Auditor = nil }, ErrInvalidRequest},
		{"failing audit changes nothing", func(_ *RiskOverageLatchReleaseRequest, a *relaxationAuditor) { a.fail = true }, nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			j, key, _ := overageLatchedOwner(t)
			auditor := &relaxationAuditor{}
			req := latchRelease(key, ownerLatchView(t, j, key).StateDigest, auditor)
			tc.mutate(&req, auditor)
			_, err := j.ReleaseRiskOverageLatch(context.Background(), req)
			if err == nil || (tc.want != nil && !errors.Is(err, tc.want)) {
				t.Fatalf("err=%v, want %v", err, tc.want)
			}
			if overage, _, reservations := ownerFlags(t, j, key); overage != 1 || reservations == 0 {
				t.Fatalf("a refused release cleared the latch: owner=%d reservations=%d", overage, reservations)
			}
			if got := countRiskBucketRows(t, j, "risk_bucket_latch_releases"); got != 0 {
				t.Fatalf("a refused release wrote %d records", got)
			}
		})
	}
}

func TestA066NextOverLimitFillLatchesAgainAfterTheRelease(t *testing.T) {
	j, key, _ := overageLatchedOwner(t)
	if _, err := j.ReleaseRiskOverageLatch(context.Background(), latchRelease(key, ownerLatchView(t, j, key).StateDigest, &relaxationAuditor{})); err != nil {
		t.Fatal(err)
	}
	if overage, _, _ := ownerFlags(t, j, key); overage != 0 {
		t.Fatal("release did not clear")
	}
	// 사용량은 여전히 한도 위(filled 160+ > 100) — 다음 체결이 오면 다시 latch(보수).
	if _, err := j.RecordFill(context.Background(), observation("risk-relax-overage", "7")); err != nil {
		t.Fatal(err)
	}
	if overage, _, reservations := ownerFlags(t, j, key); overage != 1 || reservations == 0 {
		t.Fatalf("over-limit fill after the release did not latch again: owner=%d reservations=%d", overage, reservations)
	}
}

func TestA066OwnerReleaseOpensOnlyAfterTheOverageLatchRelease(t *testing.T) {
	j, key := seedRiskBucketOwnerLifecycle(t, riskbucket.MarketUS, "relax-order")
	if _, err := j.bindRiskBucketOwnerActual(context.Background(), key); err != nil {
		t.Fatal(err)
	}
	closeRiskBucketOwnerLifecycle(t, j, key, false)
	if _, err := j.db.Exec(`INSERT INTO reconcile_states(id,account_ref,symbol,cause,evidence,entered_at,released_at,release_cause) VALUES(?,?,?,'QUANTITY_MISMATCH','fresh broker quantity zero','2026-03-30T00:40:00Z','2026-03-30T00:42:00Z','RECHECK_MATCHED')`, "owner-reconcile-relax-"+key.ProspectiveGeneration, key.AccountID, key.Symbol); err != nil {
		t.Fatal(err)
	}
	recordOfficialRiskBucketBrokerZeroAt(t, j, key, "official-zero-relax-"+key.ProspectiveGeneration, time.Date(2026, 3, 30, 0, 41, 0, 0, time.UTC))
	// 그 외 모든 해제 조건이 깨끗한 owner 에 RISK_OVERAGE 가 남아 있음(체결 경로가 봉인을 다시 찍은 모양).
	if _, err := j.db.Exec(`UPDATE risk_bucket_owners SET risk_overage_latched=1 WHERE account_ref=? AND market=? AND symbol=? AND prospective_generation=?`,
		key.AccountID, string(key.Market), key.Symbol, key.ProspectiveGeneration); err != nil {
		t.Fatal(err)
	}
	refreshRiskBucketOwnerSnapshot(t, j, key, "overage-latched")
	var lifecycle *RiskBucketOwnerLifecycleError
	if _, err := j.releaseRiskBucketOwner(context.Background(), key); !errors.As(err, &lifecycle) || lifecycle.BlockingField != "owner_latch" {
		t.Fatalf("owner release before the latch release: err=%v", err)
	}
	if _, err := j.ReleaseRiskOverageLatch(context.Background(), latchRelease(key, ownerLatchView(t, j, key).StateDigest, &relaxationAuditor{})); err != nil {
		t.Fatalf("latch release: %v", err)
	}
	if result, err := j.releaseRiskBucketOwner(context.Background(), key); err != nil || !result.Released {
		t.Fatalf("owner release after the latch release: result=%+v err=%v", result, err)
	}
}
