package journal

// 이 파일은 a066 task 5.6.1 F2 의 수리임: 예약 하나가 **자기가 크기를 잰 그 예약 가격 정책 record** 를 가리키게 함.
//
// v33 까지 risk_bucket_policies 는 (dimension, value, policy_version) 하나에 record 하나만 받았고, 그 record 는 예약
// 가격(최악 가격·FX·fee·평가 시각)을 담음. 그래서 여러 진입이 함께 쓰는 bucket(strategy·horizon·sector·market)은 첫
// 진입의 가격만 받았고, 다른 가격의 두 번째 진입은 모두 "immutable policy collision" 으로 거절됐음(2026-09-27 실측).
//
// record 신원은 두 writer(CommitRiskBucketAdmission, insertFreshRiskBucketReservations)가 이미 계산하던 digest 그대로임 —
// riskBucketRecordDigest({Key, PolicyEvidence, ReservePolicy}). 새로 짓지 않고 저장·대조하는 쪽 코드에서 읽었음.

import (
	"context"
	"database/sql"
	_ "embed"
	"fmt"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/riskbucket"
)

// schemaV34 는 예약 가격 정책 record 표와 예약→record 결속 열임(추가 전용).
//
//go:embed risk_bucket_policy_records_v34.sql
var schemaV34 string

// storeRiskBucketPolicyRecord 는 bucket 하나의 예약 가격 정책 record 를 영속하고 그 digest 를 돌려줌.
//
//   - risk_bucket_policy_records: (key, digest) 로 불변 저장. 같은 digest 는 같은 내용이므로 재삽입은 무시됨.
//   - risk_bucket_policies: risk_bucket_snapshots 의 FK 부모라 key 마다 첫 record 를 계속 둠(INSERT OR IGNORE). 다른 record 가
//     먼저 있어도 더 이상 충돌이 아님 — 예약은 자기 record 를 policy_record_digest 로 가리킴.
//
// 저장 뒤 (key, digest) 행이 실제로 있는지 다시 읽어 확인함 — 없으면 결속할 대상이 없으므로 거절.
func storeRiskBucketPolicyRecord(ctx context.Context, tx *sql.Tx, key riskbucket.BucketKey, policyDigest string,
	bound riskbucket.BucketEvidenceBinding, p riskbucket.ReservePolicy, createdAt time.Time) (string, error) {
	recordDigest, err := riskBucketRecordDigest(struct {
		Key      riskbucket.BucketKey
		Evidence riskbucket.Evidence
		Policy   riskbucket.ReservePolicy
	}{bound.Key, bound.PolicyEvidence, p})
	if err != nil {
		return "", err
	}
	values := []any{string(key.Dimension), key.Value, key.PolicyVersion, policyDigest, bound.PolicyEvidence.Source,
		canonicalRiskTime(bound.PolicyEvidence.ObservedAt), canonicalRiskTime(bound.PolicyEvidence.FreshUntil), recordDigest,
		p.AccountCurrency, p.QuoteCurrency, canonicalRiskTime(p.EvaluatedAt), p.Price.WorstExecutableQuote, p.Price.Source,
		p.Price.Version, p.Price.Digest, canonicalRiskTime(p.Price.ObservedAt), canonicalRiskTime(p.Price.FreshUntil),
		p.Fee.FixedBaseMinor, p.Fee.PerUnitBaseMinor, p.Fee.MinimumBaseMinor, p.Fee.Version, p.Fee.Digest,
		p.FX.RateQuoteToBase, p.FX.Haircut, p.FX.Source, p.FX.Version, p.FX.Digest, canonicalRiskTime(p.FX.ObservedAt),
		canonicalRiskTime(p.FX.FreshUntil), canonicalRiskTime(createdAt)}
	const columns = `bucket_dimension,bucket_value,policy_version,policy_digest,policy_source,policy_observed_at,policy_fresh_until,record_digest,account_currency,quote_currency,evaluated_at,worst_price_quote,price_source,price_version,price_digest,price_observed_at,price_fresh_until,fee_fixed_base_minor,fee_per_unit_base_minor,fee_minimum_base_minor,fee_version,fee_digest,fx_rate_quote_to_base,fx_haircut,fx_source,fx_version,fx_digest,fx_observed_at,fx_fresh_until,created_at`
	const placeholders = `?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?`
	if _, err := tx.ExecContext(ctx, `INSERT OR IGNORE INTO risk_bucket_policies(`+columns+`) VALUES(`+placeholders+`)`, values...); err != nil {
		return "", err
	}
	if _, err := tx.ExecContext(ctx, `INSERT OR IGNORE INTO risk_bucket_policy_records(`+columns+`) VALUES(`+placeholders+`)`, values...); err != nil {
		return "", err
	}
	var stored int
	if err := tx.QueryRowContext(ctx, `SELECT count(*) FROM risk_bucket_policy_records WHERE bucket_dimension=? AND bucket_value=?
		AND policy_version=? AND record_digest=?`, string(key.Dimension), key.Value, key.PolicyVersion, recordDigest).Scan(&stored); err != nil || stored != 1 {
		return "", fmt.Errorf("%w: policy record %s not stored (count=%d err=%v)", ErrRiskBucketSnapshotMismatch, recordDigest, stored, err)
	}
	return recordDigest, nil
}
