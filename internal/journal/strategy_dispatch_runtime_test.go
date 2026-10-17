package journal

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/riskbucket"
)

func TestStrategyDispatchOwnerEpochFencesPredecessorAndDormantAPIsMintNothing(t *testing.T) {
	j := openTestJournal(t)
	ctx := context.Background()
	first, err := j.AcquireStrategyDispatchOwner(ctx, "dispatch-a")
	if err != nil || first.Epoch != 1 || first.FencingToken == "" {
		t.Fatalf("first owner=%+v err=%v", first, err)
	}
	second, err := j.AcquireStrategyDispatchOwner(ctx, "dispatch-b")
	if err != nil || second.Epoch != 2 || second.FencingToken == first.FencingToken {
		t.Fatalf("second owner=%+v first=%+v err=%v", second, first, err)
	}
	if _, err := j.DiscoverStrategyDispatchRecovery(ctx, first); !errors.Is(err, ErrStrategyDispatchFenced) {
		t.Fatalf("stale recovery owner error=%v", err)
	}
	if items, err := j.DiscoverStrategyDispatchRecovery(ctx, second); err != nil || len(items) != 0 {
		t.Fatalf("current recovery items=%+v err=%v", items, err)
	}

	authority := StrategyDispatchMarketAuthority{AccountRef: "acct-1", Market: StrategyDispatchMarketKR, Symbol: "005930"}
	if _, err := j.CommitStrategyDispatchMarketAuthority(ctx, authority); !errors.Is(err, ErrStrategyDispatchDormant) {
		t.Fatalf("authority mint error=%v", err)
	}
	plan := StrategyDispatchLeasePlan{LeaseID: "lease-dormant", OwnerEpoch: second.Epoch, FencingToken: second.FencingToken}
	for name, err := range map[string]error{
		"issue": func() error { _, issueErr := j.IssueStrategyDispatchLease(ctx, plan); return issueErr }(),
		"recover": func() error {
			_, recoverErr := j.RecoverClaimedStrategyDispatchLease(ctx, StrategyDispatchLeaseCAS{})
			return recoverErr
		}(),
	} {
		want := ErrStrategyDispatchDormant
		if name == "recover" {
			want = ErrInvalidRequest
		}
		if !errors.Is(err, want) {
			t.Errorf("%s error=%v want=%v", name, err, want)
		}
	}
	if _, err := j.BeginStrategyDispatchSubmitting(ctx, StrategyDispatchLeaseCAS{}); !errors.Is(err, ErrInvalidRequest) {
		t.Fatalf("invalid submitting CAS error=%v", err)
	}
	for _, table := range []string{"strategy_dispatch_market_authorities", "strategy_dispatch_leases", "strategy_dispatch_outcomes"} {
		var count int
		if err := j.db.QueryRow("SELECT count(*) FROM " + table).Scan(&count); err != nil || count != 0 {
			t.Fatalf("%s rows=%d err=%v", table, count, err)
		}
	}
}

func TestStrategyDispatchLeaseSchemaRequiresExactQFinalAuthorityAndHolds(t *testing.T) {
	t.Run("valid-sealed-row", func(t *testing.T) {
		j := openStrategyDispatchV25Journal(t)
		owner, _ := j.AcquireStrategyDispatchOwner(context.Background(), "schema-valid")
		plan := prepareStrategyDispatchLease(t, j, "schema-valid", owner, StrategyDispatchMarketKR, "005930")
		if err := insertStrategyDispatchLease(j, plan, StrategyDispatchLeaseIssued, ""); err != nil {
			t.Fatalf("exact q_final row: %v", err)
		}
	})

	t.Run("fabricated-guardian", func(t *testing.T) {
		j := openStrategyDispatchV25Journal(t)
		owner, _ := j.AcquireStrategyDispatchOwner(context.Background(), "schema-forged")
		plan := prepareStrategyDispatchLease(t, j, "schema-forged", owner, StrategyDispatchMarketKR, "005930")
		plan.GuardianDecisionID = "fabricated-guardian-decision"
		if err := insertStrategyDispatchLease(j, plan, StrategyDispatchLeaseIssued, ""); err == nil {
			t.Fatal("fabricated Guardian decision was accepted")
		}
	})

	t.Run("one-monetary-hold-released", func(t *testing.T) {
		j := openStrategyDispatchV25Journal(t)
		owner, _ := j.AcquireStrategyDispatchOwner(context.Background(), "schema-released")
		plan := prepareStrategyDispatchLease(t, j, "schema-released", owner, StrategyDispatchMarketKR, "005930")
		if _, err := j.db.Exec(`UPDATE risk_bucket_reservations SET state='RELEASED',held_minor='0',updated_at='2026-03-30T00:30:02Z' WHERE decision_id=? AND bucket_dimension='symbol'`, plan.GuardianDecisionID); err != nil {
			t.Fatal(err)
		}
		if err := insertStrategyDispatchLease(j, plan, StrategyDispatchLeaseIssued, ""); err == nil || !strings.Contains(err.Error(), "exact q_final authority") {
			t.Fatalf("released monetary hold error=%v", err)
		}
	})

	t.Run("authority-digest-substitution", func(t *testing.T) {
		j := openStrategyDispatchV25Journal(t)
		owner, _ := j.AcquireStrategyDispatchOwner(context.Background(), "schema-authority")
		plan := prepareStrategyDispatchLease(t, j, "schema-authority", owner, StrategyDispatchMarketKR, "005930")
		plan.AuthorityDigest = "substituted-authority-digest"
		if err := insertStrategyDispatchLease(j, plan, StrategyDispatchLeaseIssued, ""); err == nil || !strings.Contains(err.Error(), "exact q_final authority") {
			t.Fatalf("substituted authority error=%v", err)
		}
	})
}

func TestStrategyDispatchBrokerOrderIDCannotCrossKRUSWithinAccount(t *testing.T) {
	j := openStrategyDispatchV25Journal(t)
	ctx := context.Background()
	owner, err := j.AcquireStrategyDispatchOwner(ctx, "cross-market-owner")
	if err != nil {
		t.Fatal(err)
	}
	kr := prepareStrategyDispatchLease(t, j, "broker-kr", owner, StrategyDispatchMarketKR, "005930")
	us := prepareStrategyDispatchLease(t, j, "broker-us", owner, StrategyDispatchMarketUS, "AAPL")
	const brokerOrderID = "broker-order-reused-across-markets"
	if err := insertStrategyDispatchLease(j, kr, StrategyDispatchLeaseSubmitted, brokerOrderID); err != nil {
		t.Fatalf("first broker binding: %v", err)
	}
	if err := insertStrategyDispatchLease(j, us, StrategyDispatchLeaseSubmitted, brokerOrderID); err == nil || !strings.Contains(err.Error(), "strategy_dispatch_leases.account_ref, strategy_dispatch_leases.broker_order_id") {
		t.Fatalf("cross-market broker reuse error=%v", err)
	}
}

func TestStrategyDispatchColdRestartDiscoversOldIssuedClaimedAndSubmitting(t *testing.T) {
	for _, tc := range []struct {
		state StrategyDispatchLeaseState
		want  StrategyDispatchRecoveryAction
	}{
		{StrategyDispatchLeaseIssued, StrategyDispatchRecoveryRefuseRelease},
		{StrategyDispatchLeaseClaimed, StrategyDispatchRecoveryRefuseRelease},
		{StrategyDispatchLeaseSubmitting, StrategyDispatchRecoveryAttestedOutcome},
	} {
		t.Run(string(tc.state), func(t *testing.T) {
			ctx := context.Background()
			path := filepath.Join(t.TempDir(), "journal.db")
			before := openStrategyDispatchTestJournal(t, path)
			oldOwner, err := before.AcquireStrategyDispatchOwner(ctx, "before-crash-"+strings.ToLower(string(tc.state)))
			if err != nil {
				t.Fatal(err)
			}
			plan := prepareStrategyDispatchLease(t, before, "restart-"+strings.ToLower(string(tc.state)), oldOwner, StrategyDispatchMarketKR, "005930")
			if err := insertStrategyDispatchLease(before, plan, tc.state, ""); err != nil {
				t.Fatal(err)
			}
			if err := before.Close(); err != nil {
				t.Fatal(err)
			}

			after := openStrategyDispatchTestJournal(t, path)
			defer after.Close()
			if tc.state == StrategyDispatchLeaseSubmitting {
				after.clk.(*clock.Fake).Set(plan.ExpiresAt.Add(strategyDispatchOwnerTakeoverGrace))
			}
			newOwner, err := after.AcquireStrategyDispatchOwner(ctx, "after-crash-"+strings.ToLower(string(tc.state)))
			if err != nil {
				t.Fatal(err)
			}
			if _, err := after.DiscoverStrategyDispatchRecovery(ctx, oldOwner); !errors.Is(err, ErrStrategyDispatchFenced) {
				t.Fatalf("old owner discovery error=%v", err)
			}
			items, err := after.DiscoverStrategyDispatchRecovery(ctx, newOwner)
			if err != nil || len(items) != 1 || items[0].Lease.LeaseID != plan.LeaseID || items[0].Lease.OwnerEpoch != oldOwner.Epoch || items[0].Action != tc.want {
				t.Fatalf("recovery items=%+v err=%v", items, err)
			}
			var state, disposition string
			if err := after.db.QueryRow(`SELECT state,disposition FROM strategy_dispatch_leases WHERE lease_id=?`, plan.LeaseID).Scan(&state, &disposition); err != nil {
				t.Fatal(err)
			}
			if state != string(tc.state) || disposition != string(StrategyDispatchReservationReserved) {
				t.Fatalf("read-only discovery mutated state=%s disposition=%s", state, disposition)
			}
		})
	}
}

func prepareStrategyDispatchLease(t *testing.T, j *Journal, suffix string, owner StrategyDispatchOwner, market StrategyDispatchMarket, symbol string) StrategyDispatchLeasePlan {
	t.Helper()
	request := qFinalIssueFixture(t, qFinalScratchJournal(t, j), suffix)
	if market == StrategyDispatchMarketUS {
		request.Admission.Owner.Key.Market = riskbucket.MarketUS
		request.Admission.Owner.Key.Symbol = symbol
		for i, bucket := range request.Admission.Admission.Buckets {
			switch bucket.Key.Dimension {
			case riskbucket.DimensionMarket:
				rebindRiskBucket(t, &request.Admission, i, riskbucket.BucketKey{Dimension: bucket.Key.Dimension, Value: string(riskbucket.MarketUS), PolicyVersion: bucket.Key.PolicyVersion})
			case riskbucket.DimensionSymbol:
				rebindRiskBucket(t, &request.Admission, i, riskbucket.BucketKey{Dimension: bucket.Key.Dimension, Value: symbol, PolicyVersion: bucket.Key.PolicyVersion})
			}
		}
		intent := request.Issue.Decision.Preimage.(RiskIntent)
		intent.Market = "us"
		intent.Symbol = symbol
		request.Issue.Decision.Preimage = intent
	}
	// a066 5.6.1: 같은 저널에 쌓이는 진입은 원장과 맞는 snapshot 을 들고 와야 함(stale 거절) — 생산처럼 원장에서 모음.
	refreshSnapshotUsageFromLedger(t, qFinalScratchJournal(t, j), &request.Admission)
	recordQFinalIntoOlderSchema(t, j, request, suffix)
	recordDigest := "sealed-authority-record-" + suffix
	if _, err := j.db.Exec(`INSERT INTO strategy_dispatch_market_authorities(
		authority_id,account_ref,market,symbol,activation_generation,activation_digest,calendar_generation,
		protection_generation,protection_serial,protection_digest,reconciliation_generation,risk_policy_generation,
		risk_policy_digest,guardian_generation,guardian_digest,build_digest,revision,record_digest,updated_at)
		VALUES(?,?,?,?,1,?,1,1,?,?,1,1,?,1,?,?,1,?,?)`,
		"sealed-authority-"+suffix, request.Admission.Owner.Key.AccountID, market, symbol,
		"sealed-activation-"+suffix, "sealed-protection-serial-"+suffix, "sealed-protection-digest-"+suffix,
		"sealed-risk-policy-"+suffix, "sealed-guardian-"+suffix, "build-"+suffix, recordDigest, "2026-03-30T00:30:01Z"); err != nil {
		t.Fatalf("seed sealed authority %s: %v", suffix, err)
	}
	return StrategyDispatchLeasePlan{
		LeaseID: "lease-" + suffix, OperationID: "operation-" + suffix,
		AccountRef: request.Admission.Owner.Key.AccountID, Market: market, Symbol: symbol,
		CandidateID: "candidate-" + suffix, EvidenceDigest: "evidence-" + suffix,
		RouterID: "router", RouterVersion: "router-v1", LaneID: request.Admission.Owner.LaneID, LaneVersion: "lane-v1",
		CampaignID: request.Admission.Owner.CampaignID, LegID: "leg-" + suffix,
		RiskReservationID: request.Admission.ExistingReservationID, GuardianDecisionID: request.Issue.Decision.ID,
		OwnerEpoch: owner.Epoch, FencingToken: owner.FencingToken,
		AuthorityRevision: 1, AuthorityDigest: recordDigest,
		IssuedAt: time.Date(2026, 3, 30, 0, 30, 5, 0, time.UTC), ExpiresAt: time.Date(2026, 3, 30, 0, 30, 50, 0, time.UTC),
	}
}

func insertStrategyDispatchLease(j *Journal, plan StrategyDispatchLeasePlan, state StrategyDispatchLeaseState, brokerOrderID string) error {
	disposition := StrategyDispatchReservationReserved
	revision := uint64(1)
	var transportStarted, outcomeObserved any
	outcomeCode, queryDigest := "", ""
	switch state {
	case StrategyDispatchLeaseClaimed:
		revision = 2
	case StrategyDispatchLeaseSubmitting:
		revision = 3
		transportStarted = "2026-03-30T00:30:10Z"
	case StrategyDispatchLeaseSubmitted:
		revision = 4
		disposition = StrategyDispatchReservationTransferred
		transportStarted = "2026-03-30T00:30:10Z"
		outcomeObserved = "2026-03-30T00:30:20Z"
		outcomeCode = "OFFICIAL_ACCEPTED"
		queryDigest = "official-query-" + plan.LeaseID
	}
	_, err := j.db.Exec(`INSERT INTO strategy_dispatch_leases(
		lease_id,operation_id,account_ref,market,symbol,candidate_id,evidence_digest,router_id,router_version,
		lane_id,lane_version,campaign_id,leg_id,risk_reservation_id,guardian_decision_id,owner_epoch,fencing_token,
		authority_revision,authority_digest,issued_at,expires_at,state,disposition,revision,transport_started_at,
		refusal_code,outcome_code,broker_order_id,query_digest,outcome_observed_at,lease_digest,created_at,updated_at)
		VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		plan.LeaseID, plan.OperationID, plan.AccountRef, plan.Market, plan.Symbol, plan.CandidateID, plan.EvidenceDigest,
		plan.RouterID, plan.RouterVersion, plan.LaneID, plan.LaneVersion, plan.CampaignID, plan.LegID,
		plan.RiskReservationID, plan.GuardianDecisionID, plan.OwnerEpoch, plan.FencingToken, plan.AuthorityRevision,
		plan.AuthorityDigest, formatJournalTime(plan.IssuedAt), formatJournalTime(plan.ExpiresAt), state, disposition,
		revision, transportStarted, "", outcomeCode, brokerOrderID, queryDigest, outcomeObserved,
		"lease-digest-"+plan.LeaseID, formatJournalTime(plan.IssuedAt), formatJournalTime(plan.IssuedAt))
	return err
}

func openStrategyDispatchTestJournal(t *testing.T, path string) *Journal {
	t.Helper()
	j, err := Open(context.Background(), Options{
		Path: path, Clock: clock.NewFake(migrationTestInstant),
		FSProber:          FixedFSProber(FSInfo{Name: "ext4", Magic: MagicExt}),
		migrationOverride: &migrationPlan{steps: migrationsThrough(25), target: 25},
	})
	if err != nil {
		t.Fatal(err)
	}
	return j
}

func openStrategyDispatchV25Journal(t *testing.T) *Journal {
	t.Helper()
	return openStrategyDispatchTestJournal(t, filepath.Join(t.TempDir(), "journal.db"))
}

// a066 5.5 fixture 수리(Manager 승인 2026-09-27, 교차 change 시험 편집).
//
// 이 파일의 시험들은 v25 저널(migrationOverride)에 q_final 행을 두고 lease 스키마를 잼. 예전 fixture 는
// **현재 코드의 writer 로 옛 스키마에 직접** 썼음 — v33 의 진입 손실 잠금 읽기가 그 저널에 없는 테이블을
// 읽고 fail-closed 로 막으면서 그 잠재 불일치가 드러남. 이제 q_final 행은 현재 스키마의 scratch 저널에서
// 현재 writer 로 만들고, 그 행을 v25 저널에 옮김. 옮길 열은 **옛 저널의 열 목록(pragma_table_info)** 이고,
// 새 writer 에만 있는 열·표는 v25 이후 migration 이 선언한 추가(schemaAdditionsAfterV25)뿐이어야 함 — 행 모양은
// 그 버전의 마이그레이션 정의에서 온 것이지 기억에서 온 것이 아님.
//
// 5.6.1 개정(2026-09-27): 처음엔 표마다 DDL 전체가 같음을 단언했으나, v34 가 risk_bucket_reservations 에 열을 더하고
// 표 하나를 새로 두면서 v25 저널과 DDL 이 정당하게 달라짐. 그래서 "같음" 을 "옛 열 ⊆ 새 열, 차이는 선언된 추가뿐"
// 으로 바꿈 — 옛 저널에 없는 열·표는 그 시절 행에 없던 것이므로 옮기지 않는 것이 옛 행의 모양 그대로임.

// qFinalIssuanceTables 는 RecordQFinalDecisionAndReserve 가 쓰는 테이블을 FK 순서로 적은 것임.
// 실제로 바뀐 테이블 집합과 같지 않으면 recordQFinalIntoOlderSchema 가 멈춤 — writer 가 새 테이블을
// 쓰기 시작하면 이 fixture 가 조용히 빠뜨리지 않게 함.
var qFinalIssuanceTables = []string{
	"decisions", "risk_reservations", "risk_bucket_owners", "risk_bucket_final_decisions",
	"risk_bucket_policies", "risk_bucket_policy_records", "risk_bucket_snapshots", "risk_bucket_reservations",
	"risk_bucket_state_snapshots", "risk_bucket_events",
}

// schemaAdditionsAfterV25 는 v25 이후 migration 이 q_final writer 표에 **더한** 것임(각 migration 정의에서 읽음):
// v34(risk_bucket_policy_records_v34.sql)가 표 risk_bucket_policy_records 와 열 risk_bucket_reservations.policy_record_digest
// 를 더함. 옛 저널로 옮길 때 이 목록만 빠질 수 있음 — 목록 밖의 차이는 fixture 결함이므로 멈춤.
var schemaAdditionsAfterV25 = struct {
	tables  map[string]bool
	columns map[string]map[string]bool
}{
	tables:  map[string]bool{"risk_bucket_policy_records": true},
	columns: map[string]map[string]bool{"risk_bucket_reservations": {"policy_record_digest": true}},
}

var qFinalScratchJournals sync.Map // *Journal(대상) → *Journal(현재 스키마 scratch)

// qFinalScratchJournal 은 대상 저널 하나에 대응하는 현재 스키마 저널임. 같은 대상에 여러 번 발급하면
// 같은 scratch 를 써서 예약 버전 등 앞선 행에 의존하는 값이 대상과 같게 이어짐.
func qFinalScratchJournal(t *testing.T, target *Journal) *Journal {
	t.Helper()
	if scratch, ok := qFinalScratchJournals.Load(target); ok {
		return scratch.(*Journal)
	}
	scratch := openTestJournal(t)
	qFinalScratchJournals.Store(target, scratch)
	t.Cleanup(func() { qFinalScratchJournals.Delete(target) })
	return scratch
}

func recordQFinalIntoOlderSchema(t *testing.T, target *Journal, request QFinalIssueRequest, suffix string) {
	t.Helper()
	ctx := context.Background()
	scratch := qFinalScratchJournal(t, target)
	tables := allJournalTables(t, scratch)
	before := make(map[string]int64, len(tables))
	for _, table := range tables {
		before[table] = maxRowID(t, scratch, table)
	}
	if _, err := scratch.RecordQFinalDecisionAndReserve(ctx, request); err != nil {
		t.Fatalf("record q_final %s: %v", suffix, err)
	}
	var changed []string
	for _, table := range tables {
		if maxRowID(t, scratch, table) != before[table] {
			changed = append(changed, table)
		}
	}
	if strings.Join(sortedCopy(changed), ",") != strings.Join(sortedCopy(qFinalIssuanceTables), ",") {
		t.Fatalf("q_final writer touched %v, fixture copies %v", changed, qFinalIssuanceTables)
	}
	tx, err := target.db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback()
	copied := []string{}
	for _, table := range qFinalIssuanceTables {
		scratchColumns := tableColumnNames(t, scratch.db, table)
		targetColumns := tableColumnNames(t, tx, table)
		if len(targetColumns) == 0 {
			if !schemaAdditionsAfterV25.tables[table] {
				t.Fatalf("older journal has no %s and no migration after v25 declares it", table)
			}
			continue // 옛 저널에 없는 표 — 그 시절 행에는 이 기록이 없었음.
		}
		// 옛 저널의 열은 전부 새 writer 에 있어야 하고, 새 writer 에만 있는 열은 선언된 추가뿐이어야 함.
		inScratch := map[string]bool{}
		for _, column := range scratchColumns {
			inScratch[column] = true
		}
		inTarget := map[string]bool{}
		for _, column := range targetColumns {
			if !inScratch[column] {
				t.Fatalf("%s.%s exists in the older journal but not in the current writer's table", table, column)
			}
			inTarget[column] = true
		}
		for _, column := range scratchColumns {
			if !inTarget[column] && !schemaAdditionsAfterV25.columns[table][column] {
				t.Fatalf("%s.%s is not in the older journal and no migration after v25 declares it", table, column)
			}
		}
		copied = append(copied, table)
		columns := targetColumns
		rows, err := scratch.db.Query(fmt.Sprintf(`SELECT %s FROM %s WHERE rowid > ? ORDER BY rowid`, strings.Join(columns, ","), table), before[table])
		if err != nil {
			t.Fatal(err)
		}
		insert := fmt.Sprintf(`INSERT INTO %s(%s) VALUES(%s)`, table, strings.Join(columns, ","),
			strings.TrimSuffix(strings.Repeat("?,", len(columns)), ","))
		for rows.Next() {
			values := make([]any, len(columns))
			pointers := make([]any, len(columns))
			for i := range values {
				pointers[i] = &values[i]
			}
			if err := rows.Scan(pointers...); err != nil {
				t.Fatal(err)
			}
			if _, err := tx.Exec(insert, values...); err != nil {
				t.Fatalf("copy %s row into the older journal: %v", table, err)
			}
		}
		if err := rows.Err(); err != nil {
			t.Fatal(err)
		}
		rows.Close()
	}
	if err := tx.Commit(); err != nil {
		t.Fatal(err)
	}
	// 옮긴 뒤 두 저널의 해당 테이블이 행 단위로 같아야 함 — 대상에는 q_final 발급 말고 이 테이블을 쓰는
	// 다른 setup 이 없으므로 불일치는 복사 결함임.
	for _, table := range copied {
		columns := tableColumnNames(t, target.db, table)
		if got, want := tableDump(t, target, table, columns), tableDump(t, scratch, table, columns); got != want {
			t.Fatalf("%s differs after copy:\n got %s\nwant %s", table, got, want)
		}
	}
}

func allJournalTables(t *testing.T, j *Journal) []string {
	t.Helper()
	rows, err := j.db.Query(`SELECT name FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%' ORDER BY name`)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	var names []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			t.Fatal(err)
		}
		names = append(names, name)
	}
	return names
}

func maxRowID(t *testing.T, j *Journal, table string) int64 {
	t.Helper()
	var id int64
	if err := j.db.QueryRow(fmt.Sprintf(`SELECT coalesce(max(rowid),0) FROM %s`, table)).Scan(&id); err != nil {
		t.Fatal(err)
	}
	return id
}

// tableColumnNames 는 표의 열 이름을 선언 순서대로 돌려줌(표가 없으면 빈 목록).
func tableColumnNames(t *testing.T, q interface {
	Query(string, ...any) (*sql.Rows, error)
}, table string) []string {
	t.Helper()
	rows, err := q.Query(fmt.Sprintf(`SELECT name FROM pragma_table_info('%s') ORDER BY cid`, table))
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	var names []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			t.Fatal(err)
		}
		names = append(names, name)
	}
	return names
}

func tableDump(t *testing.T, j *Journal, table string, columns []string) string {
	t.Helper()
	rows, err := j.db.Query(fmt.Sprintf(`SELECT %s FROM %s ORDER BY rowid`, strings.Join(columns, ","), table))
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	var out strings.Builder
	for rows.Next() {
		values := make([]any, len(columns))
		pointers := make([]any, len(columns))
		for i := range values {
			pointers[i] = &values[i]
		}
		if err := rows.Scan(pointers...); err != nil {
			t.Fatal(err)
		}
		fmt.Fprintf(&out, "%v;", values)
	}
	return out.String()
}

func sortedCopy(values []string) []string {
	out := append([]string(nil), values...)
	for i := 1; i < len(out); i++ {
		for k := i; k > 0 && out[k] < out[k-1]; k-- {
			out[k], out[k-1] = out[k-1], out[k]
		}
	}
	return out
}
