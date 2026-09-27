package journal

// a066 5.6.1: v33 저널이 v34 로 넘어와도 기존 예약 행은 그대로(policy_record_digest NULL)이고, 그 행의 주문 권위는
// 여전히 key 당 하나뿐이던 risk_bucket_policies 에서 읽힘. 새 예약은 자기 record 를 반드시 가리켜야 함.

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/riskbucket"
)

func TestMigrationV33ToV34KeepsLegacyReservationsReadable(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "journal.db")
	old := openJournalAtSchema(t, path, 33)
	seedExistingRiskReservation(t, old, "existing-legacy", "acct-1")
	// v33 writer 가 남긴 모양 그대로(v33 열만) 다섯 dimension 의 policy·snapshot·예약을 raw 로 씀.
	values := map[riskbucket.Dimension]string{riskbucket.DimensionHorizon: "SHORT", riskbucket.DimensionMarket: "KR",
		riskbucket.DimensionStrategy: "strategy-alpha", riskbucket.DimensionSector: "sector-tech", riskbucket.DimensionSymbol: "005930"}
	statements := []string{
		`INSERT INTO risk_bucket_final_decisions(decision_id,transaction_id,account_ref,market,symbol,q_candidate,q_existing_guardian,q_final,existing_reservation_id,request_digest,request_preimage,snapshot_set_digest,owner_prospective_generation,owner_lane_id,owner_campaign_id,owner_sequence,created_at) VALUES('legacy-decision','legacy-tx','acct-1','KR','005930',10,10,10,'existing-legacy','rd','rp','ss','legacy-generation','lane-a','campaign-legacy',1,'2026-03-30T00:30:00Z')`,
	}
	for _, dimension := range riskbucket.RequiredDimensionOrder() {
		d, v := string(dimension), values[dimension]
		statements = append(statements,
			fmt.Sprintf(`INSERT INTO risk_bucket_policies(bucket_dimension,bucket_value,policy_version,policy_digest,policy_source,policy_observed_at,policy_fresh_until,record_digest,account_currency,quote_currency,evaluated_at,worst_price_quote,price_source,price_version,price_digest,price_observed_at,price_fresh_until,fee_fixed_base_minor,fee_per_unit_base_minor,fee_minimum_base_minor,fee_version,fee_digest,fx_rate_quote_to_base,fx_haircut,fx_source,fx_version,fx_digest,fx_observed_at,fx_fresh_until,created_at) VALUES('%s','%s','policy-v1','pd-%s','authority','2026-03-30T00:00:00Z','2026-03-30T01:00:00Z','legacy-record-%s','KRW','KRW','2026-03-30T00:30:00Z','5','price','v1','price-d','2026-03-30T00:00:00Z','2026-03-30T01:00:00Z','0','0','0','v1','fee-d','1','1','fx','v1','fx-d','2026-03-30T00:00:00Z','2026-03-30T01:00:00Z','2026-03-30T00:30:00Z')`, d, v, d, d),
			fmt.Sprintf(`INSERT INTO risk_bucket_snapshots(snapshot_id,snapshot_digest,snapshot_source,record_digest,bucket_dimension,bucket_value,policy_version,limit_minor,filled_minor,held_minor,snapshot_version,policy_digest,observed_at,fresh_until,created_at) VALUES('legacy-snapshot-%s','sd-%s','authority','sr-%s','%s','%s','policy-v1','100','0','0','sv','pd-%s','2026-03-30T00:00:00Z','2026-03-30T01:00:00Z','2026-03-30T00:30:00Z')`, d, d, d, d, v, d),
			fmt.Sprintf(`INSERT INTO risk_bucket_reservations(reservation_id,decision_id,existing_reservation_id,account_ref,market,symbol,owner_prospective_generation,bucket_dimension,bucket_value,policy_version,snapshot_id,reserved_minor,held_minor,filled_minor,overage_minor,state,created_at,updated_at) VALUES('legacy-reservation-%s','legacy-decision','existing-legacy','acct-1','KR','005930','legacy-generation','%s','%s','policy-v1','legacy-snapshot-%s','50','50','0','0','HELD','2026-03-30T00:30:00Z','2026-03-30T00:30:00Z')`, d, d, v, d))
	}
	for _, statement := range statements {
		if _, err := old.db.Exec(statement); err != nil {
			t.Fatalf("seed v33 row: %v\n%s", err, statement)
		}
	}
	if err := old.Close(); err != nil {
		t.Fatal(err)
	}

	j := openJournalAtSchema(t, path, 34)
	defer j.Close()
	if version, err := j.SchemaVersion(ctx); err != nil || version != 34 {
		t.Fatalf("version=%d err=%v", version, err)
	}
	var records, nullBound int
	if err := j.db.QueryRow(`SELECT count(*) FROM risk_bucket_policy_records`).Scan(&records); err != nil || records != 0 {
		t.Fatalf("migration wrote %d policy records (err=%v); it must not rewrite or copy history", records, err)
	}
	if err := j.db.QueryRow(`SELECT count(*) FROM risk_bucket_reservations WHERE policy_record_digest IS NULL`).Scan(&nullBound); err != nil || nullBound != 5 {
		t.Fatalf("legacy reservations with NULL binding=%d err=%v, want 5", nullBound, err)
	}
	tx, err := j.db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	authority, err := loadRiskBucketOrderAuthority(ctx, tx, "legacy-decision")
	_ = tx.Rollback()
	if err != nil || len(authority.bindings) != 5 || authority.bindings[0].PolicyRecordDigest != "legacy-record-horizon" {
		t.Fatalf("legacy order authority after v34: %+v err=%v", authority, err)
	}
	// 새 예약은 **자기 key 의** record 를 가리켜야만 들어감(적대 리뷰 5.6.1: 이전 판본은 decision_id·dimension UNIQUE
	// 충돌로 거절돼 트리거를 재지 못했음 — 새 decision/예약 id 를 써서 트리거만 판정하게 함).
	var decisionCount int
	if err := j.db.QueryRow(`SELECT count(*) FROM risk_bucket_final_decisions WHERE decision_id='legacy-decision'`).Scan(&decisionCount); err != nil || decisionCount != 1 {
		t.Fatalf("fixture decision missing: %d err=%v", decisionCount, err)
	}
	if _, err := j.db.Exec(`INSERT INTO risk_bucket_final_decisions(decision_id,transaction_id,account_ref,market,symbol,q_candidate,q_existing_guardian,q_final,existing_reservation_id,request_digest,request_preimage,snapshot_set_digest,owner_prospective_generation,owner_lane_id,owner_campaign_id,owner_sequence,created_at) VALUES('v34-decision','v34-tx','acct-1','KR','005930',10,10,10,'existing-legacy','rd','rp','ss','legacy-generation','lane-a','campaign-legacy',2,'2026-03-30T00:30:00Z')`); err != nil {
		t.Fatal(err)
	}
	if _, err := j.db.Exec(`INSERT INTO risk_bucket_policy_records(bucket_dimension,bucket_value,policy_version,record_digest,policy_digest,policy_source,policy_observed_at,policy_fresh_until,account_currency,quote_currency,evaluated_at,worst_price_quote,price_source,price_version,price_digest,price_observed_at,price_fresh_until,fee_fixed_base_minor,fee_per_unit_base_minor,fee_minimum_base_minor,fee_version,fee_digest,fx_rate_quote_to_base,fx_haircut,fx_source,fx_version,fx_digest,fx_observed_at,fx_fresh_until,created_at) SELECT bucket_dimension,bucket_value,policy_version,'v34-record-'||bucket_dimension,policy_digest,policy_source,policy_observed_at,policy_fresh_until,account_currency,quote_currency,evaluated_at,worst_price_quote,price_source,price_version,price_digest,price_observed_at,price_fresh_until,fee_fixed_base_minor,fee_per_unit_base_minor,fee_minimum_base_minor,fee_version,fee_digest,fx_rate_quote_to_base,fx_haircut,fx_source,fx_version,fx_digest,fx_observed_at,fx_fresh_until,created_at FROM risk_bucket_policies`); err != nil {
		t.Fatal(err)
	}
	insert := func(id, dimension, value, policyVersion, digest string) error {
		var digestArg any = digest
		if digest == "" {
			digestArg = nil
		}
		_, err := j.db.Exec(`INSERT INTO risk_bucket_reservations(reservation_id,decision_id,existing_reservation_id,account_ref,market,symbol,owner_prospective_generation,bucket_dimension,bucket_value,policy_version,snapshot_id,reserved_minor,held_minor,filled_minor,overage_minor,state,created_at,updated_at,policy_record_digest) VALUES(?,'v34-decision','existing-legacy','acct-1','KR','005930','legacy-generation',?,?,?,'legacy-snapshot-'||?,'1','1','0','0','HELD','2026-03-30T00:30:00Z','2026-03-30T00:30:00Z',?)`,
			id, dimension, value, policyVersion, dimension, digestArg)
		return err
	}
	for _, tc := range []struct {
		name, dimension, value, policyVersion, digest string
		accepted                                      bool
	}{
		{"no record", "symbol", "005930", "policy-v1", "", false},
		{"unknown digest", "symbol", "005930", "policy-v1", "no-such-record", false},
		{"record of another dimension", "symbol", "005930", "policy-v1", "v34-record-horizon", false},
		{"record under another value", "horizon", "MEDIUM", "policy-v1", "v34-record-horizon", false},
		{"record under another policy version", "horizon", "SHORT", "policy-v2", "v34-record-horizon", false},
		{"own record", "symbol", "005930", "policy-v1", "v34-record-symbol", true},
	} {
		err := insert("v34-"+tc.name, tc.dimension, tc.value, tc.policyVersion, tc.digest)
		if tc.accepted != (err == nil) {
			t.Fatalf("%s: accepted=%v err=%v", tc.name, err == nil, err)
		}
		if err != nil && !strings.Contains(err.Error(), "must name its exact policy record") {
			t.Fatalf("%s: refused by something other than the v34 trigger: %v", tc.name, err)
		}
	}
	// 한 번 정한 record 는 바뀌지 않음 — v34 행(값→다른 값)과 옛 행(NULL→값) 둘 다.
	for _, id := range []string{"v34-own record", "legacy-reservation-symbol"} {
		if _, err := j.db.Exec(`UPDATE risk_bucket_reservations SET policy_record_digest='v34-record-horizon' WHERE reservation_id=?`, id); err == nil || !strings.Contains(err.Error(), "policy record is immutable") {
			t.Fatalf("binding of %s was changed or refused for another reason: %v", id, err)
		}
	}
	// fill 계상이 쓰는 갱신(held/filled/state)은 결속 불변 트리거에 걸리지 않음.
	if _, err := j.db.Exec(`UPDATE risk_bucket_reservations SET held_minor='0',filled_minor='1',state='FILLED' WHERE reservation_id='v34-own record'`); err != nil {
		t.Fatalf("fill-accounting update refused: %v", err)
	}
}
