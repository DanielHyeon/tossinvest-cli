//go:build tossos_testseams

package riskbucket

// a126 1.5 R3 (codex 3) — tasks 1.1 의 replay 결정성 문언 중 **snapshot digest** 를 생산 snapshot 생성기(LoadProductionRiskSnapshotAuthority)로
// 실측함. 이 riskbucket 내부 시험은 journal 을 import 할 수 없어(journal → riskbucket) 실 원장을 만들 수 없으므로 — a127 뒤 생성기는 주입된
// 현재 스키마와 같은 원장을 읽고, 실 원장 수락은 engine 의 TestTheRiskLoaderReadsTheRealJournal 이 잰다 — 수명주기 사실(영수증 · owner
// released_at · scope latch · 공유 owner 체결)을 축소 원장(주입 값과 같은 user_version)에 단계별로 쌓고, 매 단계 생성기를
// 두 번 부름(부를 때마다 읽기 전용 연결을 새로 엶 = 재시작 replay). 사용량 · RowDigest · latch 의 replay 는 journal 쪽
// TestA126UsageIsReplayDeterministicAcrossReleaseRevertAndSharedFills 가 생산 작성자 경로로 잼.

import (
	"context"
	"database/sql"
	"strings"
	"testing"
	"time"
)

// a126SnapshotFingerprint 는 bundle digest 와 dimension 마다 filled · held · snapshot digest.
func a126SnapshotFingerprint(t *testing.T, fixture productionRiskFixture) (string, map[Dimension]RiskSnapshotAuthorityEntry) {
	t.Helper()
	bundle, err := LoadProductionRiskSnapshotAuthority(context.Background(), fixture.config, fixture.input)
	if err != nil {
		t.Fatal(err)
	}
	parts := []string{bundle.Digest()}
	entries := map[Dimension]RiskSnapshotAuthorityEntry{}
	for _, entry := range bundle.Entries() {
		entries[entry.Bucket.Key.Dimension] = entry
		parts = append(parts, string(entry.Bucket.Key.Dimension), entry.Bucket.FilledMinor, entry.Bucket.HeldMinor, entry.Reference.SnapshotDigest)
	}
	return strings.Join(parts, "|"), entries
}

// a126InsertOwnerRows 는 owner 하나(종목 · generation)의 공유 bucket 넷 FILLED 행과 결정 사본 · owner 행을 씀.
func a126InsertOwnerRows(t *testing.T, db *sql.DB, account, symbol, generation, filled string, values map[Dimension]string) {
	t.Helper()
	decision := "decision-" + generation
	for _, dimension := range []Dimension{DimensionHorizon, DimensionMarket, DimensionStrategy, DimensionSector} {
		policy, snapshot := "a126-v1", "a126-snapshot-"+generation+"-"+string(dimension)
		for _, statement := range []struct {
			sql  string
			args []any
		}{
			{`INSERT OR IGNORE INTO risk_bucket_policies(bucket_dimension,bucket_value,policy_version,record_digest) VALUES(?,?,?,?)`,
				[]any{string(dimension), values[dimension], policy, "a126-policy-record-" + string(dimension)}},
			{`INSERT INTO risk_bucket_snapshots(snapshot_id,bucket_dimension,bucket_value,policy_version) VALUES(?,?,?,?)`,
				[]any{snapshot, string(dimension), values[dimension], policy}},
			{`INSERT INTO risk_bucket_reservations(reservation_id,account_ref,bucket_dimension,bucket_value,policy_version,snapshot_id,held_minor,filled_minor,state,risk_overage_latched,unknown_actual_latched,decision_id,market,symbol,owner_prospective_generation) VALUES(?,?,?,?,?,?,'0',?,'FILLED',0,0,?,'US',?,?)`,
				[]any{"a126-reservation-" + generation + "-" + string(dimension), account, string(dimension), values[dimension], policy, snapshot, filled, decision, symbol, generation}},
		} {
			if _, err := db.Exec(statement.sql, statement.args...); err != nil {
				t.Fatal(err)
			}
		}
	}
	if _, err := db.Exec(`INSERT INTO risk_bucket_final_decisions(decision_id,account_ref,market,symbol,owner_prospective_generation) VALUES(?,?,'US',?,?)`, decision, account, symbol, generation); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO risk_bucket_owners(account_ref,market,symbol,prospective_generation) VALUES(?,'US',?,?)`, account, symbol, generation); err != nil {
		t.Fatal(err)
	}
}

func TestA126SnapshotDigestIsReplayDeterministicAcrossReleaseSharedFillAndRevert(t *testing.T) {
	fixture := newProductionRiskFixture(t, MarketUS, time.Date(2026, 8, 4, 2, 0, 0, 0, time.UTC))
	account := fixture.config.AccountID
	// 생성기 scope 는 AAPL — 떠나는 owner A(MSFT)와 공유 owner C(NVDA)는 horizon · market · strategy · sector 를 공유함.
	values := map[Dimension]string{DimensionHorizon: "SHORT", DimensionMarket: "US", DimensionStrategy: "continuation", DimensionSector: "technology"}
	exec := func(query string, args ...any) {
		t.Helper()
		db := openProductionRiskDB(t, fixture.config.JournalPath)
		defer db.Close()
		if _, err := db.Exec(query, args...); err != nil {
			t.Fatal(err)
		}
	}
	stage := func(name, wantShared string) (string, map[Dimension]RiskSnapshotAuthorityEntry) {
		t.Helper()
		first, entries := a126SnapshotFingerprint(t, fixture)
		if again, _ := a126SnapshotFingerprint(t, fixture); again != first {
			t.Fatalf("%s: replayed snapshot fingerprint %s != %s", name, again, first)
		}
		for dimension := range values {
			if got := entries[dimension].Bucket.FilledMinor; got != wantShared {
				t.Fatalf("%s: %s filled=%s, want %s", name, dimension, got, wantShared)
			}
		}
		return first, entries
	}

	db := openProductionRiskDB(t, fixture.config.JournalPath)
	a126InsertOwnerRows(t, db, account, "MSFT", "a126-gen-a", "50", values)
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	_, before := stage("before release", "50")

	exec(`INSERT INTO risk_bucket_owner_release_receipts(account_ref,market,symbol,prospective_generation,released_at) VALUES(?,'US','MSFT','a126-gen-a','2026-08-04T01:00:00Z')`, account)
	exec(`UPDATE risk_bucket_owners SET released_at='2026-08-04T01:00:00Z' WHERE prospective_generation='a126-gen-a'`)
	_, released := stage("after release", "0")
	for dimension := range values {
		// 떠남은 파생 filled 로 snapshot digest 에 닿음(RowDigest 형식은 불변 — D2).
		if released[dimension].Reference.SnapshotDigest == before[dimension].Reference.SnapshotDigest {
			t.Fatalf("%s snapshot digest did not move with the departure", dimension)
		}
	}

	db = openProductionRiskDB(t, fixture.config.JournalPath)
	a126InsertOwnerRows(t, db, account, "NVDA", "a126-gen-c", "20", values)
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	stage("after a shared owner's fill", "20")

	exec(`INSERT INTO risk_bucket_scope_latches(account_ref,market,symbol,prospective_generation) VALUES(?,'US','MSFT','a126-gen-a')`, account)
	reverted, _ := stage("after the revert", "70")

	// 되돌림은 "떠난 적 없는" 원장과 같은 snapshot 을 냄 — 영수증 · released_at · scope latch 를 지운 원장의 digest 와 같음.
	exec(`DELETE FROM risk_bucket_scope_latches`)
	exec(`DELETE FROM risk_bucket_owner_release_receipts`)
	exec(`UPDATE risk_bucket_owners SET released_at=NULL`)
	if never, _ := stage("never departed", "70"); never != reverted {
		t.Fatalf("reverted snapshot %s != never-departed snapshot %s", reverted, never)
	}
}
