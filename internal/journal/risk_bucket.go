package journal

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/riskbucket"
)

var (
	ErrRiskBucketOwnerConflict    = errors.New("journal: risk bucket owner conflict")
	ErrRiskBucketEntryBlocked     = errors.New("journal: risk bucket entry blocked")
	ErrRiskBucketSnapshotMismatch = errors.New("journal: risk bucket snapshot mismatch")
	ErrRiskBucketReplayMismatch   = errors.New("journal: risk bucket replay mismatch")
	ErrRiskBucketStateUnknown     = errors.New("journal: risk bucket state unknown")
)

// RiskBucketSnapshotReference identifies the immutable authority record whose
// values were supplied to the pure admission calculation. The duplicated key,
// version, digest and time range are checked before any row is written.
type RiskBucketSnapshotReference struct {
	Key                                    riskbucket.BucketKey
	SnapshotID                             string
	SnapshotDigest                         string
	SnapshotVersion                        string
	PolicyDigest                           string
	ObservedAt                             time.Time
	FreshUntil                             time.Time
	PolicyObservedAt, PolicyFreshUntil     time.Time
	SnapshotObservedAt, SnapshotFreshUntil time.Time
}

func (r RiskBucketSnapshotReference) policyWindow() (time.Time, time.Time) {
	if !r.PolicyObservedAt.IsZero() || !r.PolicyFreshUntil.IsZero() {
		return r.PolicyObservedAt, r.PolicyFreshUntil
	}
	return r.ObservedAt, r.FreshUntil
}

func (r RiskBucketSnapshotReference) snapshotWindow() (time.Time, time.Time) {
	if !r.SnapshotObservedAt.IsZero() || !r.SnapshotFreshUntil.IsZero() {
		return r.SnapshotObservedAt, r.SnapshotFreshUntil
	}
	return r.ObservedAt, r.FreshUntil
}

// RiskBucketAdmissionPlan is the one transaction boundary between the existing
// Guardian reservation and the five-dimensional risk-bucket journal. Admission
// is recalculated here; callers cannot submit a forged q_final.
type RiskBucketAdmissionPlan struct {
	TransactionID         string
	DecisionID            string
	ExistingReservationID string
	Admission             riskbucket.AdmissionRequest
	Owner                 riskbucket.OwnerClaim
	Snapshots             []RiskBucketSnapshotReference
	CreatedAt             time.Time
}

type RiskBucketAdmissionReceipt struct {
	DecisionID     string
	QFinal         uint64
	OwnerReused    bool
	Idempotent     bool
	ReservationIDs []string
}

// RiskBucketState is a read-only replay projection. Legacy rows have no such
// projection and return ErrRiskBucketStateUnknown rather than fabricated zeroes.
type RiskBucketState struct {
	Owner        riskbucket.OwnerClaim
	QFinal       uint64
	Reservations map[riskbucket.BucketKey]string
	Usage        map[riskbucket.BucketKey]riskbucket.BucketUsage
	OwnerLatches map[riskbucket.Latch]bool
	Digest       string
}

func (j *Journal) CommitRiskBucketAdmission(ctx context.Context, plan RiskBucketAdmissionPlan) (RiskBucketAdmissionReceipt, error) {
	decision := riskbucket.CalculateAdmission(plan.Admission)
	if decision.Refusal != nil {
		return RiskBucketAdmissionReceipt{}, decision.Refusal
	}
	if err := validateRiskBucketAdmission(plan, decision); err != nil {
		return RiskBucketAdmissionReceipt{}, err
	}
	preimage, digest, err := riskBucketAdmissionDigest(plan, decision)
	if err != nil {
		return RiskBucketAdmissionReceipt{}, err
	}
	tx, err := j.db.BeginTx(ctx, nil)
	if err != nil {
		return RiskBucketAdmissionReceipt{}, fmt.Errorf("journal: begin risk bucket admission: %w", err)
	}
	defer tx.Rollback()

	var priorDigest string
	var priorQ uint64
	err = tx.QueryRowContext(ctx, `SELECT request_digest,q_final FROM risk_bucket_final_decisions WHERE transaction_id=?`, plan.TransactionID).Scan(&priorDigest, &priorQ)
	if err == nil {
		if priorDigest != digest || priorQ != decision.QFinal {
			return RiskBucketAdmissionReceipt{}, fmt.Errorf("%w: transaction %s", ErrRiskBucketReplayMismatch, plan.TransactionID)
		}
		receipt := riskBucketReceipt(plan, decision, true, true)
		if err := tx.Commit(); err != nil {
			return RiskBucketAdmissionReceipt{}, err
		}
		return receipt, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return RiskBucketAdmissionReceipt{}, fmt.Errorf("journal: read risk bucket admission: %w", err)
	}

	var reservationAccount, reservationState string
	if err := tx.QueryRowContext(ctx, `SELECT account_ref,state FROM risk_reservations WHERE id=?`, plan.ExistingReservationID).Scan(&reservationAccount, &reservationState); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return RiskBucketAdmissionReceipt{}, ErrReservationNotFound
		}
		return RiskBucketAdmissionReceipt{}, err
	}
	if reservationAccount != plan.Owner.Key.AccountID || reservationState != ReservationHeld {
		return RiskBucketAdmissionReceipt{}, fmt.Errorf("%w: existing reservation binding", ErrRiskBucketSnapshotMismatch)
	}
	if err := ensureRiskBucketEntryScopeClean(ctx, tx, plan.Owner.Key); err != nil {
		return RiskBucketAdmissionReceipt{}, err
	}

	ownerReused := false
	var prospective, lane, campaign string
	err = tx.QueryRowContext(ctx, `SELECT prospective_generation,lane_id,campaign_id FROM risk_bucket_owners WHERE account_ref=? AND market=? AND symbol=? AND released_at IS NULL`, plan.Owner.Key.AccountID, string(plan.Owner.Key.Market), plan.Owner.Key.Symbol).Scan(&prospective, &lane, &campaign)
	switch {
	case err == nil:
		if prospective != plan.Owner.Key.ProspectiveGeneration || lane != plan.Owner.LaneID || campaign != plan.Owner.CampaignID {
			return RiskBucketAdmissionReceipt{}, ErrRiskBucketOwnerConflict
		}
		ownerReused = true
	case errors.Is(err, sql.ErrNoRows):
		_, err = tx.ExecContext(ctx, `INSERT INTO risk_bucket_owners(account_ref,market,symbol,prospective_generation,lane_id,campaign_id,acquired_at) VALUES(?,?,?,?,?,?,?)`, plan.Owner.Key.AccountID, string(plan.Owner.Key.Market), plan.Owner.Key.Symbol, plan.Owner.Key.ProspectiveGeneration, plan.Owner.LaneID, plan.Owner.CampaignID, canonicalRiskTime(plan.CreatedAt))
		if err != nil {
			if strings.Contains(err.Error(), "UNIQUE constraint failed") {
				return RiskBucketAdmissionReceipt{}, ErrRiskBucketOwnerConflict
			}
			return RiskBucketAdmissionReceipt{}, err
		}
	default:
		return RiskBucketAdmissionReceipt{}, err
	}
	if ownerReused {
		if err := verifyRiskBucketStateDigest(ctx, tx, plan.Owner.Key); err != nil {
			return RiskBucketAdmissionReceipt{}, err
		}
		rows, err := tx.QueryContext(ctx, `SELECT DISTINCT bucket_dimension,bucket_value,policy_version FROM risk_bucket_reservations WHERE account_ref=? AND market=? AND symbol=? AND owner_prospective_generation=?`, plan.Owner.Key.AccountID, string(plan.Owner.Key.Market), plan.Owner.Key.Symbol, plan.Owner.Key.ProspectiveGeneration)
		if err != nil {
			return RiskBucketAdmissionReceipt{}, err
		}
		existing := make(map[riskbucket.BucketKey]bool)
		for rows.Next() {
			var d, v, pv string
			if err := rows.Scan(&d, &v, &pv); err != nil {
				rows.Close()
				return RiskBucketAdmissionReceipt{}, err
			}
			existing[riskbucket.BucketKey{Dimension: riskbucket.Dimension(d), Value: v, PolicyVersion: pv}] = true
		}
		if err := rows.Close(); err != nil {
			return RiskBucketAdmissionReceipt{}, err
		}
		if len(existing) != len(decision.Caps) {
			return RiskBucketAdmissionReceipt{}, fmt.Errorf("%w: scale-in bucket identity", ErrRiskBucketSnapshotMismatch)
		}
		for _, cap := range decision.Caps {
			if !existing[cap.Key] {
				return RiskBucketAdmissionReceipt{}, fmt.Errorf("%w: scale-in bucket identity", ErrRiskBucketSnapshotMismatch)
			}
		}
	}

	var ownerSequence int64
	if err := tx.QueryRowContext(ctx, `SELECT COALESCE(MAX(owner_sequence),0)+1 FROM risk_bucket_final_decisions WHERE account_ref=? AND market=? AND symbol=? AND owner_prospective_generation=?`, plan.Owner.Key.AccountID, string(plan.Owner.Key.Market), plan.Owner.Key.Symbol, plan.Owner.Key.ProspectiveGeneration).Scan(&ownerSequence); err != nil {
		return RiskBucketAdmissionReceipt{}, err
	}
	snapshotSetDigest, err := riskBucketSnapshotSetDigest(plan.Snapshots)
	if err != nil {
		return RiskBucketAdmissionReceipt{}, err
	}
	p := plan.Admission.Policy
	_, err = tx.ExecContext(ctx, `INSERT INTO risk_bucket_final_decisions(decision_id,transaction_id,account_ref,market,symbol,q_candidate,q_existing_guardian,q_final,existing_reservation_id,request_digest,request_preimage,snapshot_set_digest,owner_prospective_generation,owner_lane_id,owner_campaign_id,owner_sequence,created_at) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`, plan.DecisionID, plan.TransactionID, plan.Owner.Key.AccountID, string(plan.Owner.Key.Market), plan.Owner.Key.Symbol, decision.QCandidate, decision.QExistingGuardian, decision.QFinal, plan.ExistingReservationID, digest, preimage, snapshotSetDigest, plan.Owner.Key.ProspectiveGeneration, plan.Owner.LaneID, plan.Owner.CampaignID, ownerSequence, canonicalRiskTime(plan.CreatedAt))
	if err != nil {
		return RiskBucketAdmissionReceipt{}, fmt.Errorf("journal: insert risk bucket decision: %w", err)
	}

	for i, cap := range decision.Caps {
		bucket := plan.Admission.Buckets[i]
		ref := plan.Snapshots[i]
		bound := bucket.BoundEvidence()
		policyRecordDigest, err := riskBucketRecordDigest(struct {
			Key      riskbucket.BucketKey
			Evidence riskbucket.Evidence
			Policy   riskbucket.ReservePolicy
		}{bound.Key, bound.PolicyEvidence, p})
		if err != nil {
			return RiskBucketAdmissionReceipt{}, err
		}
		_, err = tx.ExecContext(ctx, `INSERT OR IGNORE INTO risk_bucket_policies(bucket_dimension,bucket_value,policy_version,policy_digest,policy_source,policy_observed_at,policy_fresh_until,record_digest,account_currency,quote_currency,evaluated_at,worst_price_quote,price_source,price_version,price_digest,price_observed_at,price_fresh_until,fee_fixed_base_minor,fee_per_unit_base_minor,fee_minimum_base_minor,fee_version,fee_digest,fx_rate_quote_to_base,fx_haircut,fx_source,fx_version,fx_digest,fx_observed_at,fx_fresh_until,created_at) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`, string(cap.Key.Dimension), cap.Key.Value, cap.Key.PolicyVersion, ref.PolicyDigest, bound.PolicyEvidence.Source, canonicalRiskTime(bound.PolicyEvidence.ObservedAt), canonicalRiskTime(bound.PolicyEvidence.FreshUntil), policyRecordDigest, p.AccountCurrency, p.QuoteCurrency, canonicalRiskTime(p.EvaluatedAt), p.Price.WorstExecutableQuote, p.Price.Source, p.Price.Version, p.Price.Digest, canonicalRiskTime(p.Price.ObservedAt), canonicalRiskTime(p.Price.FreshUntil), p.Fee.FixedBaseMinor, p.Fee.PerUnitBaseMinor, p.Fee.MinimumBaseMinor, p.Fee.Version, p.Fee.Digest, p.FX.RateQuoteToBase, p.FX.Haircut, p.FX.Source, p.FX.Version, p.FX.Digest, canonicalRiskTime(p.FX.ObservedAt), canonicalRiskTime(p.FX.FreshUntil), canonicalRiskTime(plan.CreatedAt))
		if err != nil {
			return RiskBucketAdmissionReceipt{}, err
		}
		var storedPolicyDigest string
		if err := tx.QueryRowContext(ctx, `SELECT record_digest FROM risk_bucket_policies WHERE bucket_dimension=? AND bucket_value=? AND policy_version=?`, string(cap.Key.Dimension), cap.Key.Value, cap.Key.PolicyVersion).Scan(&storedPolicyDigest); err != nil || storedPolicyDigest != policyRecordDigest {
			return RiskBucketAdmissionReceipt{}, fmt.Errorf("%w: immutable policy collision", ErrRiskBucketSnapshotMismatch)
		}
		snapshotRecordDigest, err := riskBucketRecordDigest(struct {
			Reference           RiskBucketSnapshotReference
			Limit, Filled, Held string
		}{ref, bucket.LimitMinor, bucket.FilledMinor, bucket.HeldMinor})
		if err != nil {
			return RiskBucketAdmissionReceipt{}, err
		}
		snapshotObserved, snapshotFresh := ref.snapshotWindow()
		_, err = tx.ExecContext(ctx, `INSERT OR IGNORE INTO risk_bucket_snapshots(snapshot_id,snapshot_digest,snapshot_source,record_digest,bucket_dimension,bucket_value,policy_version,limit_minor,filled_minor,held_minor,snapshot_version,policy_digest,observed_at,fresh_until,created_at) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`, ref.SnapshotID, ref.SnapshotDigest, bound.SnapshotEvidence.Source, snapshotRecordDigest, string(cap.Key.Dimension), cap.Key.Value, cap.Key.PolicyVersion, bucket.LimitMinor, bucket.FilledMinor, bucket.HeldMinor, ref.SnapshotVersion, ref.PolicyDigest, canonicalRiskTime(snapshotObserved), canonicalRiskTime(snapshotFresh), canonicalRiskTime(plan.CreatedAt))
		if err != nil {
			return RiskBucketAdmissionReceipt{}, err
		}
		var storedSnapshotDigest string
		if err := tx.QueryRowContext(ctx, `SELECT record_digest FROM risk_bucket_snapshots WHERE snapshot_id=?`, ref.SnapshotID).Scan(&storedSnapshotDigest); err != nil || storedSnapshotDigest != snapshotRecordDigest {
			return RiskBucketAdmissionReceipt{}, fmt.Errorf("%w: immutable snapshot collision", ErrRiskBucketSnapshotMismatch)
		}
		reservationID := riskBucketReservationID(plan.TransactionID, cap.Key.Dimension)
		_, err = tx.ExecContext(ctx, `INSERT INTO risk_bucket_reservations(reservation_id,decision_id,existing_reservation_id,account_ref,market,symbol,owner_prospective_generation,bucket_dimension,bucket_value,policy_version,snapshot_id,reserved_minor,held_minor,filled_minor,overage_minor,state,created_at,updated_at) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,'0','0','HELD',?,?)`, reservationID, plan.DecisionID, plan.ExistingReservationID, plan.Owner.Key.AccountID, string(plan.Owner.Key.Market), plan.Owner.Key.Symbol, plan.Owner.Key.ProspectiveGeneration, string(cap.Key.Dimension), cap.Key.Value, cap.Key.PolicyVersion, ref.SnapshotID, cap.ReservationAtFinal, cap.ReservationAtFinal, canonicalRiskTime(plan.CreatedAt), canonicalRiskTime(plan.CreatedAt))
		if err != nil {
			return RiskBucketAdmissionReceipt{}, fmt.Errorf("journal: insert %s risk bucket reservation: %w", cap.Key.Dimension, err)
		}
	}
	_, stateDigest, err := loadRiskBucketState(ctx, tx, plan.Owner.Key)
	if err != nil {
		return RiskBucketAdmissionReceipt{}, err
	}
	var stateSequence int64
	if err := tx.QueryRowContext(ctx, `SELECT COALESCE(MAX(event_sequence),0)+1 FROM risk_bucket_state_snapshots WHERE account_ref=? AND market=? AND symbol=? AND prospective_generation=?`, plan.Owner.Key.AccountID, string(plan.Owner.Key.Market), plan.Owner.Key.Symbol, plan.Owner.Key.ProspectiveGeneration).Scan(&stateSequence); err != nil {
		return RiskBucketAdmissionReceipt{}, err
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO risk_bucket_state_snapshots(snapshot_id,account_ref,market,symbol,prospective_generation,state_digest,event_sequence,created_at) VALUES(?,?,?,?,?,?,?,?)`, plan.TransactionID+":state:"+strconv.FormatInt(stateSequence, 10), plan.Owner.Key.AccountID, string(plan.Owner.Key.Market), plan.Owner.Key.Symbol, plan.Owner.Key.ProspectiveGeneration, stateDigest, stateSequence, canonicalRiskTime(plan.CreatedAt))
	if err != nil {
		return RiskBucketAdmissionReceipt{}, err
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO risk_bucket_events(event_id,account_ref,market,symbol,prospective_generation,event_type,event_digest,payload,created_at) VALUES(?,?,?,?,?,'ADMISSION_COMMITTED',?,?,?)`, plan.TransactionID+":event:1", plan.Owner.Key.AccountID, string(plan.Owner.Key.Market), plan.Owner.Key.Symbol, plan.Owner.Key.ProspectiveGeneration, digest, preimage, canonicalRiskTime(plan.CreatedAt))
	if err != nil {
		return RiskBucketAdmissionReceipt{}, err
	}
	if err := tx.Commit(); err != nil {
		return RiskBucketAdmissionReceipt{}, fmt.Errorf("journal: commit risk bucket admission: %w", err)
	}
	return riskBucketReceipt(plan, decision, ownerReused, false), nil
}

func ensureRiskBucketEntryScopeClean(ctx context.Context, tx *sql.Tx, key riskbucket.OwnerKey) error {
	var ownerLatches, scopeLatches, activeReconciles int
	if err := tx.QueryRowContext(ctx, `SELECT count(*) FROM risk_bucket_owners WHERE account_ref=? AND market=?
		AND symbol=? AND released_at IS NULL AND (risk_overage_latched=1 OR unknown_actual_latched=1)`,
		key.AccountID, string(key.Market), key.Symbol).Scan(&ownerLatches); err != nil {
		return err
	}
	if err := tx.QueryRowContext(ctx, `SELECT count(*) FROM risk_bucket_scope_latches WHERE account_ref=?
		AND market=? AND symbol=?`, key.AccountID, string(key.Market), key.Symbol).Scan(&scopeLatches); err != nil {
		return err
	}
	if err := tx.QueryRowContext(ctx, `SELECT count(*) FROM reconcile_states WHERE account_ref=? AND released_at IS NULL
		AND (symbol IS NULL OR symbol=?) AND (scope_market IS NULL OR scope_market=?)`,
		key.AccountID, key.Symbol, string(key.Market)).Scan(&activeReconciles); err != nil {
		return err
	}
	if ownerLatches != 0 || scopeLatches != 0 || activeReconciles != 0 {
		return ErrRiskBucketEntryBlocked
	}
	return nil
}

func validateRiskBucketAdmission(plan RiskBucketAdmissionPlan, decision riskbucket.AdmissionDecision) error {
	if plan.TransactionID == "" || plan.DecisionID == "" || plan.ExistingReservationID == "" || plan.CreatedAt.IsZero() || len(decision.Caps) != len(riskbucket.RequiredDimensionOrder()) || len(plan.Snapshots) != len(decision.Caps) {
		return fmt.Errorf("%w: incomplete admission identity", ErrRiskBucketSnapshotMismatch)
	}
	if plan.Owner.Key.AccountID == "" || plan.Owner.Key.Symbol == "" || plan.Owner.Key.ProspectiveGeneration == "" || plan.Owner.LaneID == "" || plan.Owner.CampaignID == "" || (plan.Owner.Key.Market != riskbucket.MarketKR && plan.Owner.Key.Market != riskbucket.MarketUS) {
		return ErrRiskBucketOwnerConflict
	}
	order := riskbucket.RequiredDimensionOrder()
	marketBound, symbolBound := false, false
	for i, cap := range decision.Caps {
		ref := plan.Snapshots[i]
		bucket := plan.Admission.Buckets[i]
		bound := bucket.BoundEvidence()
		policyEvidence, snapshotEvidence := bound.PolicyEvidence, bound.SnapshotEvidence
		policyObserved, policyFresh := ref.policyWindow()
		snapshotObserved, snapshotFresh := ref.snapshotWindow()
		if cap.Key.Dimension != order[i] || ref.Key != cap.Key || bucket.Key != cap.Key || bound.Key != cap.Key || bound.Snapshot.Key != cap.Key || ref.SnapshotID == "" ||
			ref.SnapshotVersion != cap.SnapshotVersion || ref.SnapshotVersion != bucket.SnapshotVersion || ref.SnapshotVersion != bound.Snapshot.SnapshotVersion ||
			policyEvidence.Source != riskbucket.RiskPolicyAuthoritySource || policyEvidence.Version != cap.Key.PolicyVersion || policyEvidence.Digest != ref.PolicyDigest || !policyEvidence.Official || !policyEvidence.Frozen || !policyEvidence.ObservedAt.Equal(policyObserved) || !policyEvidence.FreshUntil.Equal(policyFresh) ||
			snapshotEvidence.Source != riskbucket.RiskSnapshotAuthoritySource || snapshotEvidence.Version != ref.SnapshotVersion || snapshotEvidence.Digest != ref.SnapshotDigest || !snapshotEvidence.Official || !snapshotEvidence.Frozen || !snapshotEvidence.ObservedAt.Equal(snapshotObserved) || !snapshotEvidence.FreshUntil.Equal(snapshotFresh) ||
			bound.Snapshot.LimitMinor != bucket.LimitMinor || bound.Snapshot.FilledMinor != bucket.FilledMinor || bound.Snapshot.HeldMinor != bucket.HeldMinor {
			return fmt.Errorf("%w: %s", ErrRiskBucketSnapshotMismatch, cap.Key.Dimension)
		}
		switch cap.Key.Dimension {
		case riskbucket.DimensionMarket:
			marketBound = cap.Key.Value == string(plan.Owner.Key.Market)
		case riskbucket.DimensionSymbol:
			symbolBound = cap.Key.Value == plan.Owner.Key.Symbol
		}
	}
	if !marketBound || !symbolBound {
		return fmt.Errorf("%w: owner bucket binding", ErrRiskBucketSnapshotMismatch)
	}
	return nil
}

func riskBucketAdmissionDigest(plan RiskBucketAdmissionPlan, decision riskbucket.AdmissionDecision) (string, string, error) {
	type admissionPreimage struct {
		TransactionID, DecisionID, ExistingReservationID string
		QCandidate, QExistingGuardian, QFinal            uint64
		Owner                                            riskbucket.OwnerClaim
		Snapshots                                        []RiskBucketSnapshotReference
		ConsumedBuckets                                  []riskbucket.BucketEvidenceBinding
		Policy                                           riskbucket.ReservePolicy
		Caps                                             []riskbucket.BucketCap
	}
	consumed := make([]riskbucket.BucketEvidenceBinding, 0, len(plan.Admission.Buckets))
	for _, bucket := range plan.Admission.Buckets {
		consumed = append(consumed, bucket.BoundEvidence())
	}
	raw, err := json.Marshal(admissionPreimage{
		TransactionID: plan.TransactionID, DecisionID: plan.DecisionID, ExistingReservationID: plan.ExistingReservationID,
		QCandidate: decision.QCandidate, QExistingGuardian: decision.QExistingGuardian, QFinal: decision.QFinal,
		Owner: plan.Owner, Snapshots: plan.Snapshots, ConsumedBuckets: consumed, Policy: plan.Admission.Policy, Caps: decision.Caps,
	})
	if err != nil {
		return "", "", fmt.Errorf("journal: encode risk bucket admission: %w", err)
	}
	sum := sha256.Sum256(raw)
	return string(raw), hex.EncodeToString(sum[:]), nil
}

func riskBucketSnapshotSetDigest(refs []RiskBucketSnapshotReference) (string, error) {
	raw, err := json.Marshal(refs)
	if err != nil {
		return "", fmt.Errorf("journal: encode risk bucket snapshot set: %w", err)
	}
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:]), nil
}

func riskBucketRecordDigest(value any) (string, error) {
	raw, err := json.Marshal(value)
	if err != nil {
		return "", fmt.Errorf("journal: encode risk bucket record: %w", err)
	}
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:]), nil
}

func riskBucketReceipt(plan RiskBucketAdmissionPlan, decision riskbucket.AdmissionDecision, reused, idempotent bool) RiskBucketAdmissionReceipt {
	ids := make([]string, 0, len(decision.Caps))
	for _, cap := range decision.Caps {
		ids = append(ids, riskBucketReservationID(plan.TransactionID, cap.Key.Dimension))
	}
	return RiskBucketAdmissionReceipt{DecisionID: plan.DecisionID, QFinal: decision.QFinal, OwnerReused: reused, Idempotent: idempotent, ReservationIDs: ids}
}

func riskBucketReservationID(transactionID string, dimension riskbucket.Dimension) string {
	return transactionID + ":" + string(dimension)
}
func canonicalRiskTime(value time.Time) string { return value.UTC().Format(time.RFC3339Nano) }

func (j *Journal) ReadRiskBucketState(ctx context.Context, key riskbucket.OwnerKey) (RiskBucketState, error) {
	state, digest, err := loadRiskBucketState(ctx, j.db, key)
	if err != nil {
		return RiskBucketState{}, err
	}
	var persisted string
	if err := j.db.QueryRowContext(ctx, `SELECT state_digest FROM risk_bucket_state_snapshots WHERE account_ref=? AND market=? AND symbol=? AND prospective_generation=? ORDER BY event_sequence DESC LIMIT 1`, key.AccountID, string(key.Market), key.Symbol, key.ProspectiveGeneration).Scan(&persisted); err != nil || persisted != digest {
		return RiskBucketState{}, fmt.Errorf("%w: state digest", ErrRiskBucketReplayMismatch)
	}
	state.Digest = digest
	return state, nil
}

type riskBucketQueryer interface {
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
	QueryRowContext(context.Context, string, ...any) *sql.Row
}

func loadRiskBucketState(ctx context.Context, q riskBucketQueryer, key riskbucket.OwnerKey) (RiskBucketState, string, error) {
	return loadRiskBucketStateLifecycle(ctx, q, key, false)
}

func loadReleasedRiskBucketState(ctx context.Context, q riskBucketQueryer, key riskbucket.OwnerKey) (RiskBucketState, string, error) {
	return loadRiskBucketStateLifecycle(ctx, q, key, true)
}

func loadRiskBucketStateLifecycle(ctx context.Context, q riskBucketQueryer, key riskbucket.OwnerKey, includeReleased bool) (RiskBucketState, string, error) {
	var lane, campaign string
	var actual sql.NullString
	var ownerOverage, ownerUnknown int
	err := q.QueryRowContext(ctx, `SELECT lane_id,campaign_id,actual_generation,risk_overage_latched,unknown_actual_latched
		FROM risk_bucket_owners WHERE account_ref=? AND market=? AND symbol=? AND prospective_generation=?
		AND (?=1 OR released_at IS NULL)`, key.AccountID, string(key.Market), key.Symbol, key.ProspectiveGeneration, boolInt(includeReleased)).
		Scan(&lane, &campaign, &actual, &ownerOverage, &ownerUnknown)
	if errors.Is(err, sql.ErrNoRows) {
		return RiskBucketState{}, "", ErrRiskBucketStateUnknown
	}
	if err != nil {
		return RiskBucketState{}, "", err
	}
	parts := []string{"owner", key.AccountID, string(key.Market), key.Symbol, key.ProspectiveGeneration, lane, campaign, actual.String, strconv.Itoa(ownerOverage), strconv.Itoa(ownerUnknown)}
	decisionRows, err := q.QueryContext(ctx, `SELECT decision_id,transaction_id,q_candidate,q_existing_guardian,q_final,existing_reservation_id,request_digest,snapshot_set_digest,owner_sequence FROM risk_bucket_final_decisions WHERE account_ref=? AND market=? AND symbol=? AND owner_prospective_generation=? ORDER BY owner_sequence`, key.AccountID, string(key.Market), key.Symbol, key.ProspectiveGeneration)
	if err != nil {
		return RiskBucketState{}, "", err
	}
	var qTotal uint64
	decisionCount := 0
	for decisionRows.Next() {
		var id, transactionID, existingID, requestDigest, snapshotDigest string
		var candidate, guardian, final uint64
		var sequence int64
		if err := decisionRows.Scan(&id, &transactionID, &candidate, &guardian, &final, &existingID, &requestDigest, &snapshotDigest, &sequence); err != nil {
			decisionRows.Close()
			return RiskBucketState{}, "", err
		}
		if ^uint64(0)-qTotal < final {
			decisionRows.Close()
			return RiskBucketState{}, "", fmt.Errorf("%w: quantity overflow", ErrRiskBucketReplayMismatch)
		}
		qTotal += final
		decisionCount++
		parts = append(parts, "decision", strconv.FormatInt(sequence, 10), id, transactionID, strconv.FormatUint(candidate, 10), strconv.FormatUint(guardian, 10), strconv.FormatUint(final, 10), existingID, requestDigest, snapshotDigest)
	}
	if err := decisionRows.Close(); err != nil {
		return RiskBucketState{}, "", err
	}
	if decisionCount == 0 {
		return RiskBucketState{}, "", fmt.Errorf("%w: final decision", ErrRiskBucketReplayMismatch)
	}
	rows, err := q.QueryContext(ctx, `SELECT r.reservation_id,r.bucket_dimension,r.bucket_value,r.policy_version,r.snapshot_id,r.reserved_minor,r.held_minor,r.filled_minor,r.overage_minor,r.state,r.risk_overage_latched,r.unknown_actual_latched,d.owner_sequence FROM risk_bucket_reservations r JOIN risk_bucket_final_decisions d ON d.decision_id=r.decision_id WHERE r.account_ref=? AND r.market=? AND r.symbol=? AND r.owner_prospective_generation=? ORDER BY d.owner_sequence,CASE r.bucket_dimension WHEN 'horizon' THEN 1 WHEN 'market' THEN 2 WHEN 'strategy' THEN 3 WHEN 'sector' THEN 4 WHEN 'symbol' THEN 5 ELSE 99 END,r.reservation_id`, key.AccountID, string(key.Market), key.Symbol, key.ProspectiveGeneration)
	if err != nil {
		return RiskBucketState{}, "", err
	}
	defer rows.Close()
	reservations := make(map[riskbucket.BucketKey]string)
	usage := make(map[riskbucket.BucketKey]riskbucket.BucketUsage)
	dimensionCounts := make(map[int64]map[riskbucket.Dimension]bool)
	for rows.Next() {
		var id, d, v, pv, snapshotID, amount, held, filled, overage, state string
		var overageLatch, unknownLatch int
		var sequence int64
		if err := rows.Scan(&id, &d, &v, &pv, &snapshotID, &amount, &held, &filled, &overage, &state, &overageLatch, &unknownLatch, &sequence); err != nil {
			return RiskBucketState{}, "", err
		}
		dimension := riskbucket.Dimension(d)
		if !isRiskBucketDimension(dimension) {
			return RiskBucketState{}, "", fmt.Errorf("%w: invalid dimension", ErrRiskBucketReplayMismatch)
		}
		if dimensionCounts[sequence] == nil {
			dimensionCounts[sequence] = map[riskbucket.Dimension]bool{}
		}
		if dimensionCounts[sequence][dimension] {
			return RiskBucketState{}, "", fmt.Errorf("%w: duplicate dimension", ErrRiskBucketReplayMismatch)
		}
		dimensionCounts[sequence][dimension] = true
		bucketKey := riskbucket.BucketKey{Dimension: riskbucket.Dimension(d), Value: v, PolicyVersion: pv}
		var addErr error
		reservations[bucketKey], addErr = addRiskMinor(reservations[bucketKey], amount)
		if addErr != nil {
			return RiskBucketState{}, "", addErr
		}
		current := usage[bucketKey]
		current.HeldMinor, addErr = addRiskMinor(current.HeldMinor, held)
		if addErr != nil {
			return RiskBucketState{}, "", addErr
		}
		current.FilledMinor, addErr = addRiskMinor(current.FilledMinor, filled)
		if addErr != nil {
			return RiskBucketState{}, "", addErr
		}
		current.OverageMinor, addErr = addRiskMinor(current.OverageMinor, overage)
		if addErr != nil {
			return RiskBucketState{}, "", addErr
		}
		if current.Latches == nil {
			current.Latches = map[riskbucket.Latch]bool{}
		}
		current.Latches[riskbucket.LatchRiskOverage] = current.Latches[riskbucket.LatchRiskOverage] || overageLatch != 0
		current.Latches[riskbucket.LatchUnknownActualRisk] = current.Latches[riskbucket.LatchUnknownActualRisk] || unknownLatch != 0
		usage[bucketKey] = current
		parts = append(parts, "reservation", strconv.FormatInt(sequence, 10), id, d, v, pv, snapshotID, amount, held, filled, overage, state, strconv.Itoa(overageLatch), strconv.Itoa(unknownLatch))
	}
	if err := rows.Err(); err != nil {
		return RiskBucketState{}, "", err
	}
	for _, seen := range dimensionCounts {
		if len(seen) != len(riskbucket.RequiredDimensionOrder()) {
			return RiskBucketState{}, "", fmt.Errorf("%w: bucket count", ErrRiskBucketReplayMismatch)
		}
	}
	if len(dimensionCounts) != decisionCount || len(reservations) != len(riskbucket.RequiredDimensionOrder()) {
		return RiskBucketState{}, "", fmt.Errorf("%w: bucket count", ErrRiskBucketReplayMismatch)
	}
	orderRows, err := q.QueryContext(ctx, `SELECT o.order_key,o.order_id,o.decision_id,COALESCE(o.predecessor_order_key,''),o.order_quantity,o.cumulative_fill,o.quote_currency,o.base_currency,o.reservation_policy_digest,o.request_digest,o.state,o.release_reason,COALESCE(o.released_at,''),d.owner_sequence FROM risk_bucket_orders o JOIN risk_bucket_final_decisions d ON d.decision_id=o.decision_id WHERE d.account_ref=? AND d.market=? AND d.symbol=? AND d.owner_prospective_generation=? ORDER BY d.owner_sequence,o.order_key`, key.AccountID, string(key.Market), key.Symbol, key.ProspectiveGeneration)
	if err != nil {
		return RiskBucketState{}, "", err
	}
	for orderRows.Next() {
		var orderKey, orderID, decisionID, predecessor, quote, base, policyDigest, requestDigest, stateName, reason, released string
		var quantity, cumulative uint64
		var sequence int64
		if err := orderRows.Scan(&orderKey, &orderID, &decisionID, &predecessor, &quantity, &cumulative, &quote, &base, &policyDigest, &requestDigest, &stateName, &reason, &released, &sequence); err != nil {
			orderRows.Close()
			return RiskBucketState{}, "", err
		}
		parts = append(parts, "order", strconv.FormatInt(sequence, 10), orderKey, orderID, decisionID, predecessor, strconv.FormatUint(quantity, 10), strconv.FormatUint(cumulative, 10), quote, base, policyDigest, requestDigest, stateName, reason, released)
	}
	if err := orderRows.Close(); err != nil {
		return RiskBucketState{}, "", err
	}
	orderReservationRows, err := q.QueryContext(ctx, `SELECT m.order_key,m.reservation_id,m.reserved_minor FROM risk_bucket_order_reservations m JOIN risk_bucket_orders o ON o.order_key=m.order_key JOIN risk_bucket_final_decisions d ON d.decision_id=o.decision_id WHERE d.account_ref=? AND d.market=? AND d.symbol=? AND d.owner_prospective_generation=? ORDER BY m.order_key,m.reservation_id`, key.AccountID, string(key.Market), key.Symbol, key.ProspectiveGeneration)
	if err != nil {
		return RiskBucketState{}, "", err
	}
	for orderReservationRows.Next() {
		var orderKey, reservationID, amount string
		if err := orderReservationRows.Scan(&orderKey, &reservationID, &amount); err != nil {
			orderReservationRows.Close()
			return RiskBucketState{}, "", err
		}
		parts = append(parts, "order_reservation", orderKey, reservationID, amount)
	}
	if err := orderReservationRows.Close(); err != nil {
		return RiskBucketState{}, "", err
	}
	fillRows, err := q.QueryContext(ctx, `SELECT f.fill_id,f.order_key,f.order_id,f.cumulative_fill,f.delta_quantity,f.actual_known,f.fill_digest,f.observed_at FROM risk_bucket_fills f JOIN risk_bucket_orders o ON o.order_key=f.order_key JOIN risk_bucket_final_decisions d ON d.decision_id=o.decision_id WHERE d.account_ref=? AND d.market=? AND d.symbol=? AND d.owner_prospective_generation=? ORDER BY f.order_key,f.cumulative_fill,f.fill_id`, key.AccountID, string(key.Market), key.Symbol, key.ProspectiveGeneration)
	if err != nil {
		return RiskBucketState{}, "", err
	}
	for fillRows.Next() {
		var fillID, orderKey, orderID, fillDigest, observed string
		var cumulative, delta uint64
		var actualKnown int
		if err := fillRows.Scan(&fillID, &orderKey, &orderID, &cumulative, &delta, &actualKnown, &fillDigest, &observed); err != nil {
			fillRows.Close()
			return RiskBucketState{}, "", err
		}
		parts = append(parts, "fill", fillID, orderKey, orderID, strconv.FormatUint(cumulative, 10), strconv.FormatUint(delta, 10), strconv.Itoa(actualKnown), fillDigest, observed)
	}
	if err := fillRows.Close(); err != nil {
		return RiskBucketState{}, "", err
	}
	allocationRows, err := q.QueryContext(ctx, `SELECT a.fill_id,a.reservation_id,a.transfer_minor,a.filled_minor FROM risk_bucket_fill_allocations a JOIN risk_bucket_fills f ON f.fill_id=a.fill_id JOIN risk_bucket_orders o ON o.order_key=f.order_key JOIN risk_bucket_final_decisions d ON d.decision_id=o.decision_id WHERE d.account_ref=? AND d.market=? AND d.symbol=? AND d.owner_prospective_generation=? ORDER BY a.fill_id,a.reservation_id`, key.AccountID, string(key.Market), key.Symbol, key.ProspectiveGeneration)
	if err != nil {
		return RiskBucketState{}, "", err
	}
	for allocationRows.Next() {
		var fillID, reservationID, transfer, filled string
		if err := allocationRows.Scan(&fillID, &reservationID, &transfer, &filled); err != nil {
			allocationRows.Close()
			return RiskBucketState{}, "", err
		}
		parts = append(parts, "allocation", fillID, reservationID, transfer, filled)
	}
	if err := allocationRows.Close(); err != nil {
		return RiskBucketState{}, "", err
	}
	actualRows, err := q.QueryContext(ctx, `SELECT a.fill_id,a.evidence_digest,a.quote_currency,a.base_currency,a.price_quote,a.fx_rate_quote_to_base,a.allocated_fee_base_minor,a.price_source,a.price_version,a.price_digest,a.price_observed_at,a.price_fresh_until,a.fx_source,a.fx_version,a.fx_digest,a.fx_observed_at,a.fx_fresh_until,a.evaluated_at,a.observed_at FROM risk_bucket_fill_actual_evidence a JOIN risk_bucket_fills f ON f.fill_id=a.fill_id JOIN risk_bucket_orders o ON o.order_key=f.order_key JOIN risk_bucket_final_decisions d ON d.decision_id=o.decision_id WHERE d.account_ref=? AND d.market=? AND d.symbol=? AND d.owner_prospective_generation=? ORDER BY a.fill_id`, key.AccountID, string(key.Market), key.Symbol, key.ProspectiveGeneration)
	if err != nil {
		return RiskBucketState{}, "", err
	}
	for actualRows.Next() {
		values := make([]string, 19)
		targets := make([]any, 19)
		for i := range values {
			targets[i] = &values[i]
		}
		if err := actualRows.Scan(targets...); err != nil {
			actualRows.Close()
			return RiskBucketState{}, "", err
		}
		parts = append(parts, "actual")
		parts = append(parts, values...)
	}
	if err := actualRows.Close(); err != nil {
		return RiskBucketState{}, "", err
	}
	latchRows, err := q.QueryContext(ctx, `SELECT latch,detail,first_seen_at,last_seen_at FROM risk_bucket_scope_latches WHERE account_ref=? AND market=? AND symbol=? AND prospective_generation=? ORDER BY latch`, key.AccountID, string(key.Market), key.Symbol, key.ProspectiveGeneration)
	if err != nil {
		return RiskBucketState{}, "", err
	}
	for latchRows.Next() {
		var latch, detail, first, last string
		if err := latchRows.Scan(&latch, &detail, &first, &last); err != nil {
			latchRows.Close()
			return RiskBucketState{}, "", err
		}
		parts = append(parts, "scope_latch", latch, detail, first, last)
	}
	if err := latchRows.Close(); err != nil {
		return RiskBucketState{}, "", err
	}
	sum := sha256.Sum256([]byte(strings.Join(parts, "\x1f")))
	digest := hex.EncodeToString(sum[:])
	return RiskBucketState{Owner: riskbucket.OwnerClaim{Key: key, LaneID: lane, CampaignID: campaign}, QFinal: qTotal, Reservations: reservations, Usage: usage, OwnerLatches: map[riskbucket.Latch]bool{riskbucket.LatchRiskOverage: ownerOverage != 0, riskbucket.LatchUnknownActualRisk: ownerUnknown != 0}}, digest, nil
}

func isRiskBucketDimension(d riskbucket.Dimension) bool {
	for _, required := range riskbucket.RequiredDimensionOrder() {
		if d == required {
			return true
		}
	}
	return false
}
func addRiskMinor(left, right string) (string, error) {
	if left == "" {
		left = "0"
	}
	a, ok := new(big.Int).SetString(left, 10)
	if !ok || a.Sign() < 0 {
		return "", fmt.Errorf("%w: invalid minor", ErrRiskBucketReplayMismatch)
	}
	b, ok := new(big.Int).SetString(right, 10)
	if !ok || b.Sign() < 0 {
		return "", fmt.Errorf("%w: invalid minor", ErrRiskBucketReplayMismatch)
	}
	a.Add(a, b)
	if a.BitLen() > 256 {
		return "", fmt.Errorf("%w: minor overflow", ErrRiskBucketReplayMismatch)
	}
	return a.String(), nil
}

// schemaV22 is immutable released history (commit 4aee6853). Never edit it;
// later order/fill shapes belong to a new forward migration.
const schemaV22 = `
CREATE TABLE risk_bucket_policies (
 bucket_dimension TEXT NOT NULL, bucket_value TEXT NOT NULL, policy_version TEXT NOT NULL, policy_digest TEXT NOT NULL,
 policy_source TEXT NOT NULL, policy_observed_at TEXT NOT NULL, policy_fresh_until TEXT NOT NULL, record_digest TEXT NOT NULL,
 account_currency TEXT NOT NULL, quote_currency TEXT NOT NULL, evaluated_at TEXT NOT NULL,
 worst_price_quote TEXT NOT NULL, price_source TEXT NOT NULL, price_version TEXT NOT NULL, price_digest TEXT NOT NULL, price_observed_at TEXT NOT NULL, price_fresh_until TEXT NOT NULL,
 fee_fixed_base_minor TEXT NOT NULL, fee_per_unit_base_minor TEXT NOT NULL, fee_minimum_base_minor TEXT NOT NULL, fee_version TEXT NOT NULL, fee_digest TEXT NOT NULL,
 fx_rate_quote_to_base TEXT NOT NULL, fx_haircut TEXT NOT NULL, fx_source TEXT NOT NULL, fx_version TEXT NOT NULL, fx_digest TEXT NOT NULL, fx_observed_at TEXT NOT NULL, fx_fresh_until TEXT NOT NULL,
 created_at TEXT NOT NULL, PRIMARY KEY(bucket_dimension,bucket_value,policy_version)
) STRICT;
CREATE TABLE risk_bucket_snapshots (
 snapshot_id TEXT PRIMARY KEY, snapshot_digest TEXT NOT NULL UNIQUE, snapshot_source TEXT NOT NULL, record_digest TEXT NOT NULL, bucket_dimension TEXT NOT NULL, bucket_value TEXT NOT NULL, policy_version TEXT NOT NULL,
 limit_minor TEXT NOT NULL, filled_minor TEXT NOT NULL, held_minor TEXT NOT NULL, snapshot_version TEXT NOT NULL, policy_digest TEXT NOT NULL,
 observed_at TEXT NOT NULL, fresh_until TEXT NOT NULL, created_at TEXT NOT NULL,
 FOREIGN KEY(bucket_dimension,bucket_value,policy_version) REFERENCES risk_bucket_policies(bucket_dimension,bucket_value,policy_version)
) STRICT;
CREATE TABLE risk_bucket_final_decisions (
 decision_id TEXT PRIMARY KEY, transaction_id TEXT NOT NULL UNIQUE, account_ref TEXT NOT NULL, market TEXT NOT NULL CHECK(market IN ('KR','US')), symbol TEXT NOT NULL,
 q_candidate INTEGER NOT NULL CHECK(q_candidate>0), q_existing_guardian INTEGER NOT NULL CHECK(q_existing_guardian>0), q_final INTEGER NOT NULL CHECK(q_final>0 AND q_final<=q_candidate AND q_final<=q_existing_guardian),
 existing_reservation_id TEXT NOT NULL REFERENCES risk_reservations(id), request_digest TEXT NOT NULL, request_preimage TEXT NOT NULL, snapshot_set_digest TEXT NOT NULL,
 owner_prospective_generation TEXT NOT NULL, owner_lane_id TEXT NOT NULL, owner_campaign_id TEXT NOT NULL, owner_sequence INTEGER NOT NULL CHECK(owner_sequence>0), created_at TEXT NOT NULL,
 UNIQUE(account_ref,market,symbol,owner_prospective_generation,owner_sequence)
) STRICT;
CREATE TABLE risk_bucket_owners (
 account_ref TEXT NOT NULL, market TEXT NOT NULL CHECK(market IN ('KR','US')), symbol TEXT NOT NULL, prospective_generation TEXT NOT NULL,
 lane_id TEXT NOT NULL, campaign_id TEXT NOT NULL, actual_generation TEXT, acquired_at TEXT NOT NULL, released_at TEXT,
 risk_overage_latched INTEGER NOT NULL DEFAULT 0 CHECK(risk_overage_latched IN (0,1)), unknown_actual_latched INTEGER NOT NULL DEFAULT 0 CHECK(unknown_actual_latched IN (0,1)),
 PRIMARY KEY(account_ref,market,symbol,prospective_generation)
) STRICT;
CREATE UNIQUE INDEX uq_risk_bucket_active_owner ON risk_bucket_owners(account_ref,market,symbol) WHERE released_at IS NULL;
CREATE TABLE risk_bucket_reservations (
 reservation_id TEXT PRIMARY KEY, decision_id TEXT NOT NULL REFERENCES risk_bucket_final_decisions(decision_id), existing_reservation_id TEXT NOT NULL REFERENCES risk_reservations(id),
 account_ref TEXT NOT NULL, market TEXT NOT NULL, symbol TEXT NOT NULL, owner_prospective_generation TEXT NOT NULL,
 bucket_dimension TEXT NOT NULL CHECK(bucket_dimension IN ('horizon','market','strategy','sector','symbol')), bucket_value TEXT NOT NULL, policy_version TEXT NOT NULL,
 snapshot_id TEXT NOT NULL REFERENCES risk_bucket_snapshots(snapshot_id), reserved_minor TEXT NOT NULL, held_minor TEXT NOT NULL, filled_minor TEXT NOT NULL, overage_minor TEXT NOT NULL,
 state TEXT NOT NULL CHECK(state IN ('HELD','FILLED','RELEASED')), risk_overage_latched INTEGER NOT NULL DEFAULT 0 CHECK(risk_overage_latched IN (0,1)), unknown_actual_latched INTEGER NOT NULL DEFAULT 0 CHECK(unknown_actual_latched IN (0,1)),
 created_at TEXT NOT NULL, updated_at TEXT NOT NULL, UNIQUE(decision_id,bucket_dimension)
) STRICT;
CREATE TABLE risk_bucket_orders (
 order_id TEXT PRIMARY KEY, decision_id TEXT NOT NULL REFERENCES risk_bucket_final_decisions(decision_id), predecessor_order_id TEXT,
 order_quantity INTEGER NOT NULL, cumulative_fill INTEGER NOT NULL DEFAULT 0, quote_currency TEXT NOT NULL, base_currency TEXT NOT NULL,
 reservation_policy_digest TEXT NOT NULL, created_at TEXT NOT NULL, updated_at TEXT NOT NULL
) STRICT;
CREATE TABLE risk_bucket_fills (
 fill_id TEXT PRIMARY KEY, order_id TEXT NOT NULL, cumulative_fill INTEGER NOT NULL, delta_quantity INTEGER NOT NULL,
 actual_known INTEGER NOT NULL CHECK(actual_known IN (0,1)), price_quote TEXT, price_source TEXT, price_version TEXT, price_digest TEXT,
 fx_rate_quote_to_base TEXT, fx_source TEXT, fx_version TEXT, fx_digest TEXT, allocated_fee_base_minor TEXT,
 fill_digest TEXT NOT NULL, observed_at TEXT NOT NULL
) STRICT;
CREATE TABLE risk_bucket_fill_allocations (
 fill_id TEXT NOT NULL REFERENCES risk_bucket_fills(fill_id), reservation_id TEXT NOT NULL REFERENCES risk_bucket_reservations(reservation_id),
 transfer_minor TEXT NOT NULL, filled_minor TEXT NOT NULL, PRIMARY KEY(fill_id,reservation_id)
) STRICT;
CREATE TABLE risk_bucket_events (
 event_id TEXT PRIMARY KEY, account_ref TEXT NOT NULL, market TEXT NOT NULL, symbol TEXT NOT NULL, prospective_generation TEXT NOT NULL,
 event_type TEXT NOT NULL, event_digest TEXT NOT NULL, payload TEXT NOT NULL, created_at TEXT NOT NULL
) STRICT;
CREATE INDEX idx_risk_bucket_events_scope ON risk_bucket_events(account_ref,market,symbol,prospective_generation,created_at,event_id);
CREATE TABLE risk_bucket_state_snapshots (
 snapshot_id TEXT PRIMARY KEY, account_ref TEXT NOT NULL, market TEXT NOT NULL, symbol TEXT NOT NULL, prospective_generation TEXT NOT NULL,
 state_digest TEXT NOT NULL, event_sequence INTEGER NOT NULL, created_at TEXT NOT NULL,
 UNIQUE(account_ref,market,symbol,prospective_generation,event_sequence)
) STRICT;
CREATE TABLE risk_bucket_scope_latches (
 account_ref TEXT NOT NULL, market TEXT NOT NULL, symbol TEXT NOT NULL, prospective_generation TEXT NOT NULL,
 latch TEXT NOT NULL CHECK(latch IN ('RISK_OVERAGE','UNKNOWN_ACTUAL_RISK','REPLAY_MISMATCH','ORPHAN_FILL')), detail TEXT NOT NULL, first_seen_at TEXT NOT NULL, last_seen_at TEXT NOT NULL,
 PRIMARY KEY(account_ref,market,symbol,prospective_generation,latch)
) STRICT;
CREATE TRIGGER risk_bucket_policies_no_update BEFORE UPDATE ON risk_bucket_policies BEGIN SELECT RAISE(ABORT,'risk bucket policies are immutable'); END;
CREATE TRIGGER risk_bucket_policies_no_delete BEFORE DELETE ON risk_bucket_policies BEGIN SELECT RAISE(ABORT,'risk bucket policies are immutable'); END;
CREATE TRIGGER risk_bucket_snapshots_no_update BEFORE UPDATE ON risk_bucket_snapshots BEGIN SELECT RAISE(ABORT,'risk bucket snapshots are immutable'); END;
CREATE TRIGGER risk_bucket_snapshots_no_delete BEFORE DELETE ON risk_bucket_snapshots BEGIN SELECT RAISE(ABORT,'risk bucket snapshots are immutable'); END;
CREATE TRIGGER risk_bucket_decisions_no_update BEFORE UPDATE ON risk_bucket_final_decisions BEGIN SELECT RAISE(ABORT,'risk bucket decisions are immutable'); END;
CREATE TRIGGER risk_bucket_decisions_no_delete BEFORE DELETE ON risk_bucket_final_decisions BEGIN SELECT RAISE(ABORT,'risk bucket decisions are immutable'); END;
CREATE TRIGGER risk_bucket_events_no_update BEFORE UPDATE ON risk_bucket_events BEGIN SELECT RAISE(ABORT,'risk bucket events are append only'); END;
CREATE TRIGGER risk_bucket_events_no_delete BEFORE DELETE ON risk_bucket_events BEGIN SELECT RAISE(ABORT,'risk bucket events are append only'); END;
CREATE TRIGGER risk_bucket_fills_no_update BEFORE UPDATE ON risk_bucket_fills BEGIN SELECT RAISE(ABORT,'risk bucket fills are append only'); END;
CREATE TRIGGER risk_bucket_fills_no_delete BEFORE DELETE ON risk_bucket_fills BEGIN SELECT RAISE(ABORT,'risk bucket fills are append only'); END;
`

const schemaV23 = `
DROP TRIGGER risk_bucket_fills_no_update;
DROP TRIGGER risk_bucket_fills_no_delete;
ALTER TABLE risk_bucket_orders RENAME TO risk_bucket_orders_v22;
ALTER TABLE risk_bucket_fills RENAME TO risk_bucket_fills_v22;
ALTER TABLE risk_bucket_fill_allocations RENAME TO risk_bucket_fill_allocations_v22;
CREATE TRIGGER risk_bucket_fills_v22_no_update BEFORE UPDATE ON risk_bucket_fills_v22 BEGIN SELECT RAISE(ABORT,'risk bucket v22 fills are immutable'); END;
CREATE TRIGGER risk_bucket_fills_v22_no_delete BEFORE DELETE ON risk_bucket_fills_v22 BEGIN SELECT RAISE(ABORT,'risk bucket v22 fills are immutable'); END;
CREATE TRIGGER risk_bucket_orders_v22_no_update BEFORE UPDATE ON risk_bucket_orders_v22 BEGIN SELECT RAISE(ABORT,'risk bucket v22 orders are immutable'); END;
CREATE TRIGGER risk_bucket_orders_v22_no_delete BEFORE DELETE ON risk_bucket_orders_v22 BEGIN SELECT RAISE(ABORT,'risk bucket v22 orders are immutable'); END;
CREATE TRIGGER risk_bucket_fill_allocations_v22_no_update BEFORE UPDATE ON risk_bucket_fill_allocations_v22 BEGIN SELECT RAISE(ABORT,'risk bucket v22 allocations are immutable'); END;
CREATE TRIGGER risk_bucket_fill_allocations_v22_no_delete BEFORE DELETE ON risk_bucket_fill_allocations_v22 BEGIN SELECT RAISE(ABORT,'risk bucket v22 allocations are immutable'); END;
CREATE TABLE risk_bucket_orders (
 order_key TEXT PRIMARY KEY, order_id TEXT NOT NULL, decision_id TEXT NOT NULL REFERENCES risk_bucket_final_decisions(decision_id), predecessor_order_key TEXT,
 order_quantity INTEGER NOT NULL, cumulative_fill INTEGER NOT NULL DEFAULT 0, quote_currency TEXT NOT NULL, base_currency TEXT NOT NULL,
 reservation_policy_digest TEXT NOT NULL, request_digest TEXT NOT NULL, state TEXT NOT NULL CHECK(state IN ('ACTIVE','REPLACED','RELEASED')),
 release_reason TEXT NOT NULL DEFAULT '', released_at TEXT, created_at TEXT NOT NULL, updated_at TEXT NOT NULL,
 UNIQUE(decision_id,order_id)
) STRICT;
CREATE TABLE risk_bucket_order_reservations (
 order_key TEXT NOT NULL REFERENCES risk_bucket_orders(order_key), reservation_id TEXT NOT NULL REFERENCES risk_bucket_reservations(reservation_id),
 reserved_minor TEXT NOT NULL, PRIMARY KEY(order_key,reservation_id)
) STRICT;
CREATE TABLE risk_bucket_fills (
 fill_id TEXT PRIMARY KEY, order_key TEXT NOT NULL REFERENCES risk_bucket_orders(order_key), order_id TEXT NOT NULL, cumulative_fill INTEGER NOT NULL, delta_quantity INTEGER NOT NULL,
 actual_known INTEGER NOT NULL CHECK(actual_known IN (0,1)), price_quote TEXT, price_source TEXT, price_version TEXT, price_digest TEXT,
 fx_rate_quote_to_base TEXT, fx_source TEXT, fx_version TEXT, fx_digest TEXT, allocated_fee_base_minor TEXT,
 fill_digest TEXT NOT NULL, observed_at TEXT NOT NULL
) STRICT;
CREATE TABLE risk_bucket_fill_actual_evidence (
 fill_id TEXT PRIMARY KEY REFERENCES risk_bucket_fills(fill_id), evidence_digest TEXT NOT NULL,
 quote_currency TEXT NOT NULL, base_currency TEXT NOT NULL, price_quote TEXT NOT NULL, fx_rate_quote_to_base TEXT NOT NULL, allocated_fee_base_minor TEXT NOT NULL,
 price_source TEXT NOT NULL, price_version TEXT NOT NULL, price_digest TEXT NOT NULL, price_observed_at TEXT NOT NULL, price_fresh_until TEXT NOT NULL,
 fx_source TEXT NOT NULL, fx_version TEXT NOT NULL, fx_digest TEXT NOT NULL, fx_observed_at TEXT NOT NULL, fx_fresh_until TEXT NOT NULL,
 evaluated_at TEXT NOT NULL, observed_at TEXT NOT NULL
) STRICT;
CREATE TABLE risk_bucket_fill_allocations (
 fill_id TEXT NOT NULL REFERENCES risk_bucket_fills(fill_id), reservation_id TEXT NOT NULL REFERENCES risk_bucket_reservations(reservation_id),
 transfer_minor TEXT NOT NULL, filled_minor TEXT NOT NULL, PRIMARY KEY(fill_id,reservation_id)
) STRICT;
CREATE TRIGGER risk_bucket_fills_no_update BEFORE UPDATE ON risk_bucket_fills BEGIN SELECT RAISE(ABORT,'risk bucket fills are append only'); END;
CREATE TRIGGER risk_bucket_fills_no_delete BEFORE DELETE ON risk_bucket_fills BEGIN SELECT RAISE(ABORT,'risk bucket fills are append only'); END;
CREATE TRIGGER risk_bucket_order_reservations_no_update BEFORE UPDATE ON risk_bucket_order_reservations BEGIN SELECT RAISE(ABORT,'risk bucket order reservations are immutable'); END;
CREATE TRIGGER risk_bucket_order_reservations_no_delete BEFORE DELETE ON risk_bucket_order_reservations BEGIN SELECT RAISE(ABORT,'risk bucket order reservations are immutable'); END;
CREATE TRIGGER risk_bucket_fill_allocations_identity_guard BEFORE UPDATE ON risk_bucket_fill_allocations WHEN NEW.fill_id<>OLD.fill_id OR NEW.reservation_id<>OLD.reservation_id OR NEW.transfer_minor<>OLD.transfer_minor BEGIN SELECT RAISE(ABORT,'risk bucket transfer allocation is immutable'); END;
CREATE TRIGGER risk_bucket_fill_allocations_no_delete BEFORE DELETE ON risk_bucket_fill_allocations BEGIN SELECT RAISE(ABORT,'risk bucket fill allocations cannot be deleted'); END;
CREATE TRIGGER risk_bucket_fill_actual_no_update BEFORE UPDATE ON risk_bucket_fill_actual_evidence BEGIN SELECT RAISE(ABORT,'risk bucket actual evidence is immutable'); END;
CREATE TRIGGER risk_bucket_fill_actual_no_delete BEFORE DELETE ON risk_bucket_fill_actual_evidence BEGIN SELECT RAISE(ABORT,'risk bucket actual evidence is immutable'); END;
`
