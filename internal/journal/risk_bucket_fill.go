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

type RiskBucketOrderPlan struct {
	OrderID, DecisionID, PredecessorOrderID              string
	OrderQuantity                                        uint64
	ReservedMinor                                        map[riskbucket.BucketKey]string
	ReservationPolicyDigest, QuoteCurrency, BaseCurrency string
	CreatedAt                                            time.Time
}

type RiskBucketActualFillPlan struct {
	Owner          riskbucket.OwnerKey
	DecisionID     string
	OrderID        string
	CumulativeFill uint64
	Actual         *riskbucket.ActualFillEvidence
	ObservedAt     time.Time
}

type RiskBucketOrderReleaseReason string

const (
	RiskBucketReleaseCancel         RiskBucketOrderReleaseReason = "CANCEL"
	RiskBucketReleaseExpiry         RiskBucketOrderReleaseReason = "EXPIRY"
	RiskBucketReleaseBrokerTerminal RiskBucketOrderReleaseReason = "BROKER_TERMINAL"
)

type RiskBucketOrderRelease struct {
	Owner      riskbucket.OwnerKey
	DecisionID string
	OrderID    string
	Reason     RiskBucketOrderReleaseReason
	ReleasedAt time.Time
}
type RiskBucketOrderReleaseResult struct{ Released, AlreadyReleased bool }

func (j *Journal) RegisterRiskBucketOrder(ctx context.Context, plan RiskBucketOrderPlan) error {
	plan.OrderID = strings.TrimSpace(plan.OrderID)
	plan.DecisionID = strings.TrimSpace(plan.DecisionID)
	plan.PredecessorOrderID = strings.TrimSpace(plan.PredecessorOrderID)
	plan.QuoteCurrency = strings.ToUpper(strings.TrimSpace(plan.QuoteCurrency))
	plan.BaseCurrency = strings.ToUpper(strings.TrimSpace(plan.BaseCurrency))
	if plan.OrderID == "" || plan.DecisionID == "" || plan.OrderQuantity == 0 || len(plan.ReservedMinor) != len(riskbucket.RequiredDimensionOrder()) || plan.CreatedAt.IsZero() {
		return fmt.Errorf("%w: incomplete risk bucket order", ErrInvalidRequest)
	}
	tx, err := j.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var account, market, symbol, prospective string
	if err := tx.QueryRowContext(ctx, `SELECT account_ref,market,symbol,owner_prospective_generation FROM risk_bucket_final_decisions WHERE decision_id=?`, plan.DecisionID).Scan(&account, &market, &symbol, &prospective); err != nil {
		return fmt.Errorf("%w: risk decision", ErrRiskBucketStateUnknown)
	}
	orderKey := riskBucketOrderKey(plan.DecisionID, plan.OrderID)
	var active int
	if err := tx.QueryRowContext(ctx, `SELECT count(*) FROM risk_bucket_owners WHERE account_ref=? AND market=? AND symbol=? AND prospective_generation=? AND released_at IS NULL`, account, market, symbol, prospective).Scan(&active); err != nil || active != 1 {
		return fmt.Errorf("%w: active owner", ErrRiskBucketStateUnknown)
	}
	var confirmedAuthority int
	var confirmedQuantity string
	if err := tx.QueryRowContext(ctx, `SELECT count(*),COALESCE(MIN(TRIM(i.quantity)),'') FROM mutation_attempts a JOIN intents i ON i.id=a.intent_id JOIN risk_bucket_final_decisions d ON d.decision_id=? JOIN risk_reservations legacy ON legacy.id=d.existing_reservation_id WHERE a.broker_order_id=? AND a.state='CONFIRMED' AND a.kind IN ('PLACE','AMEND') AND a.decision_id=legacy.decision_id AND TRIM(i.account_ref)=d.account_ref AND UPPER(TRIM(i.market))=d.market AND UPPER(TRIM(i.symbol))=d.symbol AND UPPER(TRIM(i.side))='BUY'`, plan.DecisionID, plan.OrderID).Scan(&confirmedAuthority, &confirmedQuantity); err != nil {
		return err
	}
	confirmedOrderQuantity, quantityErr := strconv.ParseUint(confirmedQuantity, 10, 64)
	if confirmedAuthority != 1 || quantityErr != nil || confirmedOrderQuantity != plan.OrderQuantity {
		return fmt.Errorf("%w: order lacks one exact confirmed authority", ErrRiskBucketReplayMismatch)
	}
	authority, err := loadRiskBucketOrderAuthority(ctx, tx, plan.DecisionID)
	if err != nil {
		return err
	}
	// Policy and currency identity are journal-derived authority. Caller-supplied
	// strings are deliberately ignored and can never become the replay seal.
	plan.ReservationPolicyDigest = authority.digest
	plan.QuoteCurrency = authority.quoteCurrency
	plan.BaseCurrency = authority.baseCurrency
	digest, err := riskBucketOrderPlanDigest(plan, authority.bindings)
	if err != nil {
		return err
	}
	var priorDigest string
	if err := tx.QueryRowContext(ctx, `SELECT request_digest FROM risk_bucket_orders WHERE order_key=?`, orderKey).Scan(&priorDigest); err == nil {
		if priorDigest != digest {
			return fmt.Errorf("%w: order registration", ErrRiskBucketReplayMismatch)
		}
		return tx.Commit()
	} else if !errors.Is(err, sql.ErrNoRows) {
		return err
	}
	var ownerOrderIDCollisions int
	if err := tx.QueryRowContext(ctx, `SELECT count(*) FROM risk_bucket_orders o JOIN risk_bucket_final_decisions d ON d.decision_id=o.decision_id WHERE o.order_id=? AND o.order_key<>? AND d.account_ref=? AND d.market=? AND d.symbol=? AND d.owner_prospective_generation=?`, plan.OrderID, orderKey, account, market, symbol, prospective).Scan(&ownerOrderIDCollisions); err != nil {
		return err
	}
	if ownerOrderIDCollisions != 0 {
		return fmt.Errorf("%w: broker order id already bound inside owner", ErrRiskBucketReplayMismatch)
	}
	if err := verifyRiskBucketStateDigest(ctx, tx, riskbucket.OwnerKey{AccountID: account, Market: riskbucket.Market(market), Symbol: symbol, ProspectiveGeneration: prospective}); err != nil {
		return err
	}

	reservationIDs := make(map[riskbucket.BucketKey]string)
	rows, err := tx.QueryContext(ctx, `SELECT reservation_id,bucket_dimension,bucket_value,policy_version,held_minor FROM risk_bucket_reservations WHERE decision_id=?`, plan.DecisionID)
	if err != nil {
		return err
	}
	currentHeld := make(map[riskbucket.BucketKey]string)
	for rows.Next() {
		var id, d, v, pv, held string
		if err := rows.Scan(&id, &d, &v, &pv, &held); err != nil {
			rows.Close()
			return err
		}
		key := riskbucket.BucketKey{Dimension: riskbucket.Dimension(d), Value: v, PolicyVersion: pv}
		reservationIDs[key] = id
		currentHeld[key] = held
	}
	if err := rows.Close(); err != nil {
		return err
	}
	if len(reservationIDs) != len(riskbucket.RequiredDimensionOrder()) {
		return fmt.Errorf("%w: reservation set", ErrRiskBucketReplayMismatch)
	}
	for key, amount := range plan.ReservedMinor {
		if reservationIDs[key] == "" {
			return fmt.Errorf("%w: order bucket key", ErrRiskBucketSnapshotMismatch)
		}
		if _, err := parseRiskMinor(amount); err != nil {
			return err
		}
	}

	var predecessorKey any
	if plan.PredecessorOrderID != "" {
		predKey := riskBucketOrderKey(plan.DecisionID, plan.PredecessorOrderID)
		predecessorKey = predKey
		var predecessorDecision, predecessorState string
		if err := tx.QueryRowContext(ctx, `SELECT decision_id,state FROM risk_bucket_orders WHERE order_key=?`, predKey).Scan(&predecessorDecision, &predecessorState); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return fmt.Errorf("%w: replacement predecessor missing", ErrRiskBucketReplayMismatch)
			}
			return err
		}
		if predecessorDecision != plan.DecisionID || predecessorState != "ACTIVE" {
			return fmt.Errorf("%w: replacement predecessor is not exact active order", ErrRiskBucketReplayMismatch)
		}
		remaining, err := riskBucketOrderRemaining(ctx, tx, predKey)
		if err != nil {
			return err
		}
		if !equalRiskMinorMap(remaining, plan.ReservedMinor) {
			return fmt.Errorf("%w: replacement remaining reservation", ErrRiskBucketReplayMismatch)
		}
		result, err := tx.ExecContext(ctx, `UPDATE risk_bucket_orders SET state='REPLACED',updated_at=? WHERE order_key=? AND decision_id=? AND state='ACTIVE'`, canonicalRiskTime(plan.CreatedAt), predKey, plan.DecisionID)
		if err != nil {
			return err
		}
		if affected, err := result.RowsAffected(); err != nil || affected != 1 {
			if err != nil {
				return err
			}
			return fmt.Errorf("%w: replacement predecessor transition", ErrRiskBucketReplayMismatch)
		}
	} else {
		if !equalRiskMinorMap(currentHeld, plan.ReservedMinor) {
			return fmt.Errorf("%w: initial order reservation", ErrRiskBucketReplayMismatch)
		}
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO risk_bucket_orders(order_key,order_id,decision_id,predecessor_order_key,order_quantity,cumulative_fill,quote_currency,base_currency,reservation_policy_digest,request_digest,state,created_at,updated_at) VALUES(?,?,?,?,?,0,?,?,?,?, 'ACTIVE',?,?)`, orderKey, plan.OrderID, plan.DecisionID, predecessorKey, plan.OrderQuantity, plan.QuoteCurrency, plan.BaseCurrency, plan.ReservationPolicyDigest, digest, canonicalRiskTime(plan.CreatedAt), canonicalRiskTime(plan.CreatedAt)); err != nil {
		return err
	}
	for key, amount := range plan.ReservedMinor {
		if _, err := tx.ExecContext(ctx, `INSERT INTO risk_bucket_order_reservations(order_key,reservation_id,reserved_minor) VALUES(?,?,?)`, orderKey, reservationIDs[key], amount); err != nil {
			return err
		}
	}
	if err := j.recordRiskBucketStateTx(ctx, tx, riskbucket.OwnerKey{AccountID: account, Market: riskbucket.Market(market), Symbol: symbol, ProspectiveGeneration: prospective}, "ORDER_REGISTERED", orderKey, digest, canonicalRiskTime(plan.CreatedAt)); err != nil {
		return err
	}
	return tx.Commit()
}

// applyRiskBucketFillInTx is invoked only for locally-owned BUY fills from
// RecordFill's authoritative transaction. Semantic/reconstruction failures are
// latched and return nil so they can never drop the broker fill or Position.
func (j *Journal) applyRiskBucketFillInTx(ctx context.Context, tx *sql.Tx, fill AppliedFill) error {
	if strings.ToUpper(strings.TrimSpace(fill.Side)) != "BUY" || fill.Delta == "0" {
		return nil
	}
	order, found, err := riskBucketOrderForFill(ctx, tx, fill)
	if err != nil {
		if errors.Is(err, ErrRiskBucketReplayMismatch) {
			return latchRiskBucketFillFailureForScope(ctx, tx, fill, err.Error())
		}
		return err
	}
	if !found {
		return nil
	}
	key := order.ownerKey()
	if err := verifyRiskBucketStateDigest(ctx, tx, key); err != nil {
		if isRiskBucketSemanticError(err) {
			return latchRiskBucketFillFailure(ctx, tx, key, fill, "REPLAY_MISMATCH", err.Error())
		}
		return err
	}
	state, event, err := loadRiskBucketFillTransition(ctx, tx, order, fill.CumulativeQuantity, nil)
	if err != nil {
		if isRiskBucketSemanticError(err) {
			return latchRiskBucketFillFailure(ctx, tx, key, fill, "REPLAY_MISMATCH", err.Error())
		}
		return err
	}
	next, result, err := riskbucket.ApplyFill(state, event)
	if err != nil {
		return latchRiskBucketFillFailure(ctx, tx, key, fill, "REPLAY_MISMATCH", err.Error())
	}
	if result.Duplicate {
		return nil
	}
	if err := persistRiskBucketFillTransition(ctx, tx, order, event, state, next, result, nil, fill.CommittedAt); err != nil {
		if errors.Is(err, ErrRiskBucketReplayMismatch) {
			return latchRiskBucketFillFailure(ctx, tx, key, fill, "REPLAY_MISMATCH", err.Error())
		}
		return err
	}
	eventDigest, err := riskBucketFillEventDigest(event)
	if err != nil {
		return err
	}
	return j.recordRiskBucketStateTx(ctx, tx, key, "FILL_APPLIED", event.FillID, eventDigest, fill.CommittedAt)
}

// releaseTerminalRiskBucketOrderInTx releases the unfilled portion of a
// locally-owned BUY after brokerstate derived a terminal lifecycle. It shares
// RecordFill's or strategy settlement's transaction; semantic damage latches
// the scope instead of rejecting broker evidence.
func (j *Journal) releaseTerminalRiskBucketOrderInTx(ctx context.Context, tx *sql.Tx, fill AppliedFill) error {
	if !fill.Terminal || strings.ToUpper(strings.TrimSpace(fill.Side)) != "BUY" {
		return nil
	}
	order, found, err := riskBucketOrderForFill(ctx, tx, fill)
	if err != nil {
		if errors.Is(err, ErrRiskBucketReplayMismatch) {
			return latchRiskBucketFillFailureForScope(ctx, tx, fill, err.Error())
		}
		return err
	}
	if !found {
		return nil
	}
	if err := verifyRiskBucketStateDigest(ctx, tx, order.ownerKey()); err != nil {
		if isRiskBucketSemanticError(err) {
			return latchRiskBucketFillFailure(ctx, tx, order.ownerKey(), fill, "REPLAY_MISMATCH", err.Error())
		}
		return err
	}
	if order.state == "RELEASED" {
		return nil
	}
	if order.state == "REPLACED" {
		return latchRiskBucketFillFailure(ctx, tx, order.ownerKey(), fill, "REPLAY_MISMATCH",
			"terminal predecessor reservation belongs to its successor")
	}
	_, err = j.recordReleasedRiskBucketOrderInTx(ctx, tx, order, RiskBucketReleaseBrokerTerminal, fill.CommittedAt)
	if errors.Is(err, ErrRiskBucketReplayMismatch) {
		return latchRiskBucketFillFailure(ctx, tx, order.ownerKey(), fill, "REPLAY_MISMATCH", err.Error())
	}
	return err
}

func isRiskBucketSemanticError(err error) bool {
	return errors.Is(err, ErrRiskBucketReplayMismatch) || errors.Is(err, ErrRiskBucketSnapshotMismatch) || errors.Is(err, ErrRiskBucketStateUnknown) || strings.Contains(err.Error(), "sql: Scan error")
}

// completeRiskBucketFillActual remains package-private until an official,
// capability-sealed broker evidence adapter exists. Caller-supplied Official
// booleans are not sufficient production authority.
func (j *Journal) completeRiskBucketFillActual(ctx context.Context, plan RiskBucketActualFillPlan) (riskbucket.FillResult, error) {
	if strings.TrimSpace(plan.OrderID) == "" || strings.TrimSpace(plan.DecisionID) == "" || plan.Owner.AccountID == "" || plan.Owner.Symbol == "" || plan.Owner.ProspectiveGeneration == "" || (plan.Owner.Market != riskbucket.MarketKR && plan.Owner.Market != riskbucket.MarketUS) || plan.CumulativeFill == 0 || plan.Actual == nil || plan.ObservedAt.IsZero() {
		return riskbucket.FillResult{}, fmt.Errorf("%w: incomplete actual fill evidence", ErrInvalidRequest)
	}
	tx, err := j.db.BeginTx(ctx, nil)
	if err != nil {
		return riskbucket.FillResult{}, err
	}
	defer tx.Rollback()
	order, found, err := riskBucketOrderByID(ctx, tx, plan.Owner, plan.DecisionID, plan.OrderID)
	if err != nil {
		return riskbucket.FillResult{}, err
	}
	if !found {
		return riskbucket.FillResult{}, ErrRiskBucketStateUnknown
	}
	key := order.ownerKey()
	if err := verifyRiskBucketStateDigest(ctx, tx, key); err != nil {
		return riskbucket.FillResult{}, err
	}
	fillID := riskBucketFillID(order.orderKey, plan.CumulativeFill)
	actualDigest, err := riskBucketRecordDigest(plan.Actual)
	if err != nil {
		return riskbucket.FillResult{}, err
	}
	var stored string
	if err := tx.QueryRowContext(ctx, `SELECT evidence_digest FROM risk_bucket_fill_actual_evidence WHERE fill_id=?`, fillID).Scan(&stored); err == nil {
		if stored != actualDigest {
			return riskbucket.FillResult{}, fmt.Errorf("%w: actual fill evidence", ErrRiskBucketReplayMismatch)
		}
		if err := tx.Commit(); err != nil {
			return riskbucket.FillResult{}, err
		}
		return riskbucket.FillResult{Duplicate: true}, nil
	} else if !errors.Is(err, sql.ErrNoRows) {
		return riskbucket.FillResult{}, err
	}
	state, event, err := loadRiskBucketFillTransition(ctx, tx, order, strconv.FormatUint(plan.CumulativeFill, 10), plan.Actual)
	if err != nil {
		return riskbucket.FillResult{}, err
	}
	next, result, err := riskbucket.ApplyFill(state, event)
	if err != nil {
		return riskbucket.FillResult{}, err
	}
	if !result.ActualEvidenceCompleted {
		return result, fmt.Errorf("%w: actual evidence did not complete", ErrRiskBucketReplayMismatch)
	}
	if err := persistRiskBucketFillTransition(ctx, tx, order, event, state, next, result, plan.Actual, canonicalRiskTime(plan.ObservedAt)); err != nil {
		return riskbucket.FillResult{}, err
	}
	if err := j.recordRiskBucketStateTx(ctx, tx, key, "ACTUAL_EVIDENCE_COMPLETED", fillID, actualDigest, canonicalRiskTime(plan.ObservedAt)); err != nil {
		return riskbucket.FillResult{}, err
	}
	if err := tx.Commit(); err != nil {
		return riskbucket.FillResult{}, err
	}
	return result, nil
}

// releaseRiskBucketOrder remains package-private until confirmed cancel/expiry,
// broker-zero and clean lifecycle evidence are derived inside the journal.
func (j *Journal) releaseRiskBucketOrder(ctx context.Context, req RiskBucketOrderRelease) (RiskBucketOrderReleaseResult, error) {
	if strings.TrimSpace(req.OrderID) == "" || strings.TrimSpace(req.DecisionID) == "" || req.Owner.AccountID == "" || req.Owner.Symbol == "" || req.Owner.ProspectiveGeneration == "" || (req.Owner.Market != riskbucket.MarketKR && req.Owner.Market != riskbucket.MarketUS) || (req.Reason != RiskBucketReleaseCancel && req.Reason != RiskBucketReleaseExpiry) || req.ReleasedAt.IsZero() {
		return RiskBucketOrderReleaseResult{}, fmt.Errorf("%w: invalid risk order release", ErrInvalidRequest)
	}
	tx, err := j.db.BeginTx(ctx, nil)
	if err != nil {
		return RiskBucketOrderReleaseResult{}, err
	}
	defer tx.Rollback()
	order, found, err := riskBucketOrderByID(ctx, tx, req.Owner, req.DecisionID, req.OrderID)
	if err != nil {
		return RiskBucketOrderReleaseResult{}, err
	}
	if !found {
		return RiskBucketOrderReleaseResult{}, ErrRiskBucketStateUnknown
	}
	if err := verifyRiskBucketStateDigest(ctx, tx, order.ownerKey()); err != nil {
		return RiskBucketOrderReleaseResult{}, err
	}
	result, err := j.recordReleasedRiskBucketOrderInTx(ctx, tx, order, req.Reason, canonicalRiskTime(req.ReleasedAt))
	if err != nil {
		return RiskBucketOrderReleaseResult{}, err
	}
	if err := tx.Commit(); err != nil {
		return RiskBucketOrderReleaseResult{}, err
	}
	return result, nil
}

func releaseRiskBucketOrderInTx(ctx context.Context, tx *sql.Tx, order riskBucketOrderRecord,
	reason RiskBucketOrderReleaseReason, releasedAt string,
) (RiskBucketOrderReleaseResult, error) {
	if order.state == "RELEASED" {
		if order.releaseReason != string(reason) {
			return RiskBucketOrderReleaseResult{}, fmt.Errorf("%w: release reason", ErrRiskBucketReplayMismatch)
		}
		return RiskBucketOrderReleaseResult{AlreadyReleased: true}, nil
	}
	if order.state == "REPLACED" {
		return RiskBucketOrderReleaseResult{}, fmt.Errorf("%w: replaced predecessor reservation belongs to successor", ErrRiskBucketReplayMismatch)
	}
	remaining, err := riskBucketOrderRemaining(ctx, tx, order.orderKey)
	if err != nil {
		return RiskBucketOrderReleaseResult{}, err
	}
	for key, amount := range remaining {
		reservationID := order.reservations[key]
		if reservationID == "" {
			return RiskBucketOrderReleaseResult{}, fmt.Errorf("%w: release reservation", ErrRiskBucketReplayMismatch)
		}
		var held string
		if err := tx.QueryRowContext(ctx, `SELECT held_minor FROM risk_bucket_reservations WHERE reservation_id=?`, reservationID).Scan(&held); err != nil {
			return RiskBucketOrderReleaseResult{}, err
		}
		next, err := subtractRiskMinorFloor(held, amount)
		if err != nil {
			return RiskBucketOrderReleaseResult{}, err
		}
		if _, err := tx.ExecContext(ctx, `UPDATE risk_bucket_reservations SET held_minor=?,state=CASE WHEN ?='0' THEN 'RELEASED' ELSE state END,updated_at=? WHERE reservation_id=?`, next, next, releasedAt, reservationID); err != nil {
			return RiskBucketOrderReleaseResult{}, err
		}
	}
	if _, err := tx.ExecContext(ctx, `UPDATE risk_bucket_orders SET state='RELEASED',release_reason=?,released_at=?,updated_at=? WHERE order_key=?`, string(reason), releasedAt, releasedAt, order.orderKey); err != nil {
		return RiskBucketOrderReleaseResult{}, err
	}
	return RiskBucketOrderReleaseResult{Released: true}, nil
}

func (j *Journal) recordReleasedRiskBucketOrderInTx(ctx context.Context, tx *sql.Tx,
	order riskBucketOrderRecord, reason RiskBucketOrderReleaseReason, releasedAt string,
) (RiskBucketOrderReleaseResult, error) {
	result, err := releaseRiskBucketOrderInTx(ctx, tx, order, reason, releasedAt)
	if err != nil || result.AlreadyReleased {
		return result, err
	}
	if err := j.recordRiskBucketStateTx(ctx, tx, order.ownerKey(), "ORDER_RELEASED", order.orderKey, string(reason), releasedAt); err != nil {
		return RiskBucketOrderReleaseResult{}, err
	}
	return result, nil
}

type riskBucketOrderRecord struct {
	orderKey, orderID, decisionID, account, market, symbol, prospective, state string
	releaseReason                                                              string
	quantity, cumulative                                                       uint64
	quote, base, policyDigest                                                  string
	reservations                                                               map[riskbucket.BucketKey]string
}

func (o riskBucketOrderRecord) ownerKey() riskbucket.OwnerKey {
	return riskbucket.OwnerKey{AccountID: o.account, Market: riskbucket.Market(o.market), Symbol: o.symbol, ProspectiveGeneration: o.prospective}
}

type riskBucketOrderAuthorityBinding struct {
	Dimension, Value, PolicyVersion, ReservationID, ReservedMinor string
	SnapshotID, SnapshotDigest, PolicyDigest, PolicyRecordDigest  string
}

type riskBucketOrderAuthority struct {
	digest, quoteCurrency, baseCurrency string
	bindings                            []riskBucketOrderAuthorityBinding
}

func loadRiskBucketOrderAuthority(ctx context.Context, tx *sql.Tx, decisionID string) (riskBucketOrderAuthority, error) {
	rows, err := tx.QueryContext(ctx, `SELECT r.bucket_dimension,r.bucket_value,r.policy_version,r.reservation_id,r.reserved_minor,s.snapshot_id,s.snapshot_digest,s.policy_digest,p.policy_digest,p.record_digest,p.quote_currency,p.account_currency FROM risk_bucket_reservations r JOIN risk_bucket_snapshots s ON s.snapshot_id=r.snapshot_id AND s.bucket_dimension=r.bucket_dimension AND s.bucket_value=r.bucket_value AND s.policy_version=r.policy_version JOIN risk_bucket_policies p ON p.bucket_dimension=r.bucket_dimension AND p.bucket_value=r.bucket_value AND p.policy_version=r.policy_version WHERE r.decision_id=?`, decisionID)
	if err != nil {
		return riskBucketOrderAuthority{}, err
	}
	defer rows.Close()
	byDimension := make(map[riskbucket.Dimension]riskBucketOrderAuthorityBinding)
	required := make(map[riskbucket.Dimension]bool, len(riskbucket.RequiredDimensionOrder()))
	for _, dimension := range riskbucket.RequiredDimensionOrder() {
		required[dimension] = true
	}
	var quote, base string
	for rows.Next() {
		var binding riskBucketOrderAuthorityBinding
		var policyDigest, rowQuote, rowBase string
		if err := rows.Scan(&binding.Dimension, &binding.Value, &binding.PolicyVersion, &binding.ReservationID, &binding.ReservedMinor, &binding.SnapshotID, &binding.SnapshotDigest, &binding.PolicyDigest, &policyDigest, &binding.PolicyRecordDigest, &rowQuote, &rowBase); err != nil {
			return riskBucketOrderAuthority{}, fmt.Errorf("%w: order authority scan: %v", ErrRiskBucketReplayMismatch, err)
		}
		dimension := riskbucket.Dimension(binding.Dimension)
		if _, duplicate := byDimension[dimension]; !required[dimension] || duplicate || binding.PolicyDigest != policyDigest || rowQuote == "" || rowBase == "" || (quote != "" && quote != rowQuote) || (base != "" && base != rowBase) {
			return riskBucketOrderAuthority{}, fmt.Errorf("%w: non-canonical order authority", ErrRiskBucketSnapshotMismatch)
		}
		quote, base = rowQuote, rowBase
		byDimension[dimension] = binding
	}
	if err := rows.Err(); err != nil {
		return riskBucketOrderAuthority{}, err
	}
	if len(byDimension) != len(required) {
		return riskBucketOrderAuthority{}, fmt.Errorf("%w: order authority dimension set", ErrRiskBucketSnapshotMismatch)
	}
	bindings := make([]riskBucketOrderAuthorityBinding, 0, len(riskbucket.RequiredDimensionOrder()))
	for _, dimension := range riskbucket.RequiredDimensionOrder() {
		binding, ok := byDimension[dimension]
		if !ok {
			return riskBucketOrderAuthority{}, fmt.Errorf("%w: missing %s order authority", ErrRiskBucketSnapshotMismatch, dimension)
		}
		bindings = append(bindings, binding)
	}
	digest, err := riskBucketRecordDigest(bindings)
	if err != nil {
		return riskBucketOrderAuthority{}, err
	}
	return riskBucketOrderAuthority{digest: digest, quoteCurrency: quote, baseCurrency: base, bindings: bindings}, nil
}

func riskBucketOrderPlanDigest(plan RiskBucketOrderPlan, bindings []riskBucketOrderAuthorityBinding) (string, error) {
	type reservation struct {
		Key    riskbucket.BucketKey
		Amount string
	}
	ordered := make([]reservation, 0, len(riskbucket.RequiredDimensionOrder()))
	for _, dimension := range riskbucket.RequiredDimensionOrder() {
		var found *riskBucketOrderAuthorityBinding
		for i := range bindings {
			if riskbucket.Dimension(bindings[i].Dimension) == dimension {
				found = &bindings[i]
				break
			}
		}
		if found == nil {
			return "", fmt.Errorf("%w: missing canonical order binding", ErrRiskBucketSnapshotMismatch)
		}
		key := riskbucket.BucketKey{Dimension: dimension, Value: found.Value, PolicyVersion: found.PolicyVersion}
		amount, ok := plan.ReservedMinor[key]
		if !ok {
			return "", fmt.Errorf("%w: missing canonical order reservation", ErrRiskBucketSnapshotMismatch)
		}
		ordered = append(ordered, reservation{Key: key, Amount: amount})
	}
	return riskBucketRecordDigest(struct {
		OrderID, DecisionID, PredecessorOrderID              string
		OrderQuantity                                        uint64
		Reservations                                         []reservation
		ReservationPolicyDigest, QuoteCurrency, BaseCurrency string
		CreatedAt                                            string
	}{plan.OrderID, plan.DecisionID, plan.PredecessorOrderID, plan.OrderQuantity, ordered, plan.ReservationPolicyDigest, plan.QuoteCurrency, plan.BaseCurrency, canonicalRiskTime(plan.CreatedAt)})
}

func riskBucketFillEventDigest(event riskbucket.FillEvent) (string, error) {
	type reservation struct {
		Key                riskbucket.BucketKey
		Amount, TargetHeld string
	}
	ordered := make([]reservation, 0, len(riskbucket.RequiredDimensionOrder()))
	for _, dimension := range riskbucket.RequiredDimensionOrder() {
		var matched bool
		for key, amount := range event.ReservedMinor {
			if key.Dimension == dimension {
				if matched {
					return "", fmt.Errorf("%w: duplicate fill reservation dimension", ErrRiskBucketReplayMismatch)
				}
				ordered = append(ordered, reservation{Key: key, Amount: amount, TargetHeld: event.TargetHeldMinor[key]})
				matched = true
			}
		}
		if !matched {
			return "", fmt.Errorf("%w: missing fill reservation dimension", ErrRiskBucketReplayMismatch)
		}
	}
	return riskBucketRecordDigest(struct {
		FillID, OrderKey, OrderID, ReservationPolicyDigest, QuoteCurrency, BaseCurrency string
		OrderQuantity, NewCumulativeFill                                                uint64
		Reservations                                                                    []reservation
		Actual                                                                          *riskbucket.ActualFillEvidence
	}{event.FillID, event.OrderKey, event.OrderID, event.ReservationPolicyDigest, event.QuoteCurrency, event.BaseCurrency, event.OrderQuantity, event.NewCumulativeFill, ordered, event.Actual})
}

func riskBucketOrderForFill(ctx context.Context, tx *sql.Tx, fill AppliedFill) (riskBucketOrderRecord, bool, error) {
	return queryRiskBucketOrder(ctx, tx, `WHERE o.order_id=? AND d.account_ref=? AND d.market=? AND d.symbol=? AND ow.released_at IS NULL`, fill.OrderID, strings.TrimSpace(fill.AccountRef), strings.ToUpper(strings.TrimSpace(fill.Market)), strings.ToUpper(strings.TrimSpace(fill.Symbol)))
}
func riskBucketOrderByID(ctx context.Context, tx *sql.Tx, owner riskbucket.OwnerKey, decisionID, orderID string) (riskBucketOrderRecord, bool, error) {
	return queryRiskBucketOrder(ctx, tx, `WHERE o.order_id=? AND o.decision_id=? AND d.account_ref=? AND d.market=? AND d.symbol=? AND d.owner_prospective_generation=? AND ow.released_at IS NULL`, strings.TrimSpace(orderID), strings.TrimSpace(decisionID), owner.AccountID, string(owner.Market), owner.Symbol, owner.ProspectiveGeneration)
}
func queryRiskBucketOrder(ctx context.Context, tx *sql.Tx, where string, args ...any) (riskBucketOrderRecord, bool, error) {
	query := `SELECT o.order_key,o.order_id,o.decision_id,d.account_ref,d.market,d.symbol,d.owner_prospective_generation,o.state,o.release_reason,o.order_quantity,o.cumulative_fill,o.quote_currency,o.base_currency,o.reservation_policy_digest FROM risk_bucket_orders o JOIN risk_bucket_final_decisions d ON d.decision_id=o.decision_id JOIN risk_bucket_owners ow ON ow.account_ref=d.account_ref AND ow.market=d.market AND ow.symbol=d.symbol AND ow.prospective_generation=d.owner_prospective_generation ` + where
	rows, err := tx.QueryContext(ctx, query, args...)
	if err != nil {
		return riskBucketOrderRecord{}, false, err
	}
	defer rows.Close()
	var matches []riskBucketOrderRecord
	for rows.Next() {
		var o riskBucketOrderRecord
		if err := rows.Scan(&o.orderKey, &o.orderID, &o.decisionID, &o.account, &o.market, &o.symbol, &o.prospective, &o.state, &o.releaseReason, &o.quantity, &o.cumulative, &o.quote, &o.base, &o.policyDigest); err != nil {
			return riskBucketOrderRecord{}, false, fmt.Errorf("%w: risk order scan: %v", ErrRiskBucketReplayMismatch, err)
		}
		matches = append(matches, o)
	}
	if err := rows.Err(); err != nil {
		return riskBucketOrderRecord{}, false, err
	}
	if len(matches) == 0 {
		return riskBucketOrderRecord{}, false, nil
	}
	if len(matches) != 1 {
		return riskBucketOrderRecord{}, false, fmt.Errorf("%w: ambiguous risk order", ErrRiskBucketReplayMismatch)
	}
	o := matches[0]
	o.reservations = map[riskbucket.BucketKey]string{}
	r, err := tx.QueryContext(ctx, `SELECT r.bucket_dimension,r.bucket_value,r.policy_version,r.reservation_id FROM risk_bucket_order_reservations m JOIN risk_bucket_reservations r ON r.reservation_id=m.reservation_id WHERE m.order_key=?`, o.orderKey)
	if err != nil {
		return riskBucketOrderRecord{}, false, err
	}
	for r.Next() {
		var d, v, pv, id string
		if err := r.Scan(&d, &v, &pv, &id); err != nil {
			r.Close()
			return riskBucketOrderRecord{}, false, fmt.Errorf("%w: risk order reservation scan: %v", ErrRiskBucketReplayMismatch, err)
		}
		o.reservations[riskbucket.BucketKey{Dimension: riskbucket.Dimension(d), Value: v, PolicyVersion: pv}] = id
	}
	if err := r.Close(); err != nil {
		return riskBucketOrderRecord{}, false, err
	}
	return o, true, nil
}

func riskBucketOrderKey(decisionID, orderID string) string {
	sum := sha256.Sum256([]byte(decisionID + "\x00" + orderID))
	return "rbo:" + hex.EncodeToString(sum[:16])
}
func riskBucketFillID(orderKey string, cumulative uint64) string {
	sum := sha256.Sum256([]byte(orderKey + "\x00" + strconv.FormatUint(cumulative, 10)))
	return "rbf:" + hex.EncodeToString(sum[:16])
}
func parseRiskMinor(raw string) (*big.Int, error) {
	n, ok := new(big.Int).SetString(raw, 10)
	if !ok || n.Sign() < 0 || n.BitLen() > 256 {
		return nil, fmt.Errorf("%w: invalid minor %q", ErrRiskBucketReplayMismatch, raw)
	}
	return n, nil
}
func subtractRiskMinorFloor(left, right string) (string, error) {
	a, err := parseRiskMinor(left)
	if err != nil {
		return "", err
	}
	b, err := parseRiskMinor(right)
	if err != nil {
		return "", err
	}
	a.Sub(a, b)
	if a.Sign() < 0 {
		return "0", nil
	}
	return a.String(), nil
}

func subtractRiskMinorExact(left, right string) (string, error) {
	a, err := parseRiskMinor(left)
	if err != nil {
		return "", err
	}
	b, err := parseRiskMinor(right)
	if err != nil {
		return "", err
	}
	if a.Cmp(b) < 0 {
		return "", fmt.Errorf("%w: aggregate target reservation underflow", ErrRiskBucketReplayMismatch)
	}
	return new(big.Int).Sub(a, b).String(), nil
}

func riskMinorMonotoneDelta(high, low string) (string, error) {
	a, err := parseRiskMinor(high)
	if err != nil {
		return "", err
	}
	b, err := parseRiskMinor(low)
	if err != nil {
		return "", err
	}
	if a.Cmp(b) < 0 {
		return "", fmt.Errorf("%w: non-monotone aggregate usage", ErrRiskBucketReplayMismatch)
	}
	return new(big.Int).Sub(a, b).String(), nil
}
func equalRiskMinorMap(a, b map[riskbucket.BucketKey]string) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		left, e1 := parseRiskMinor(v)
		right, e2 := parseRiskMinor(b[k])
		if e1 != nil || e2 != nil || left.Cmp(right) != 0 {
			return false
		}
	}
	return true
}

func riskBucketOrderRemaining(ctx context.Context, tx *sql.Tx, orderKey string) (map[riskbucket.BucketKey]string, error) {
	rows, err := tx.QueryContext(ctx, `SELECT r.bucket_dimension,r.bucket_value,r.policy_version,m.reservation_id,m.reserved_minor FROM risk_bucket_order_reservations m JOIN risk_bucket_reservations r ON r.reservation_id=m.reservation_id WHERE m.order_key=?`, orderKey)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[riskbucket.BucketKey]string{}
	for rows.Next() {
		var d, v, pv, reservationID, reserved string
		if err := rows.Scan(&d, &v, &pv, &reservationID, &reserved); err != nil {
			return nil, err
		}
		transferred := "0"
		allocations, err := tx.QueryContext(ctx, `SELECT a.transfer_minor FROM risk_bucket_fill_allocations a JOIN risk_bucket_fills f ON f.fill_id=a.fill_id WHERE f.order_key=? AND a.reservation_id=? ORDER BY f.cumulative_fill,f.fill_id`, orderKey, reservationID)
		if err != nil {
			return nil, err
		}
		for allocations.Next() {
			var amount string
			if err := allocations.Scan(&amount); err != nil {
				allocations.Close()
				return nil, err
			}
			transferred, err = addRiskMinor(transferred, amount)
			if err != nil {
				allocations.Close()
				return nil, err
			}
		}
		if err := allocations.Close(); err != nil {
			return nil, err
		}
		remaining, err := subtractRiskMinorFloor(reserved, transferred)
		if err != nil {
			return nil, err
		}
		out[riskbucket.BucketKey{Dimension: riskbucket.Dimension(d), Value: v, PolicyVersion: pv}] = remaining
	}
	return out, rows.Err()
}

func loadRiskBucketFillTransition(ctx context.Context, tx *sql.Tx, target riskBucketOrderRecord, cumulativeRaw string, actual *riskbucket.ActualFillEvidence) (riskbucket.FillState, riskbucket.FillEvent, error) {
	cumulative, err := strconv.ParseUint(strings.TrimSpace(cumulativeRaw), 10, 64)
	if err != nil || cumulative == 0 {
		return riskbucket.FillState{}, riskbucket.FillEvent{}, fmt.Errorf("%w: non-integral cumulative fill", ErrRiskBucketReplayMismatch)
	}
	state := riskbucket.FillState{Buckets: map[riskbucket.BucketKey]riskbucket.BucketUsage{}, Orders: map[string]riskbucket.OrderFillState{}, OwnerLatches: map[riskbucket.Latch]bool{}}
	var ownerOverage, ownerUnknown int
	if err := tx.QueryRowContext(ctx, `SELECT risk_overage_latched,unknown_actual_latched FROM risk_bucket_owners WHERE account_ref=? AND market=? AND symbol=? AND prospective_generation=?`, target.account, target.market, target.symbol, target.prospective).Scan(&ownerOverage, &ownerUnknown); err != nil {
		return state, riskbucket.FillEvent{}, err
	}
	state.OwnerLatches[riskbucket.LatchRiskOverage] = ownerOverage != 0
	state.OwnerLatches[riskbucket.LatchUnknownActualRisk] = ownerUnknown != 0
	rows, err := tx.QueryContext(ctx, `SELECT r.decision_id,r.bucket_dimension,r.bucket_value,r.policy_version,s.limit_minor,r.held_minor,r.filled_minor,r.overage_minor,r.risk_overage_latched,r.unknown_actual_latched FROM risk_bucket_reservations r JOIN risk_bucket_snapshots s ON s.snapshot_id=r.snapshot_id JOIN risk_bucket_final_decisions d ON d.decision_id=r.decision_id WHERE d.account_ref=? AND d.market=? AND d.symbol=? AND d.owner_prospective_generation=? ORDER BY d.owner_sequence,CASE r.bucket_dimension WHEN 'horizon' THEN 1 WHEN 'market' THEN 2 WHEN 'strategy' THEN 3 WHEN 'sector' THEN 4 WHEN 'symbol' THEN 5 ELSE 99 END`, target.account, target.market, target.symbol, target.prospective)
	if err != nil {
		return state, riskbucket.FillEvent{}, err
	}
	decisionBuckets := map[string]map[riskbucket.BucketKey]bool{}
	for rows.Next() {
		var decisionID, d, v, pv, limit, held, filled, overage string
		var ol, ul int
		if err := rows.Scan(&decisionID, &d, &v, &pv, &limit, &held, &filled, &overage, &ol, &ul); err != nil {
			rows.Close()
			return state, riskbucket.FillEvent{}, err
		}
		key := riskbucket.BucketKey{Dimension: riskbucket.Dimension(d), Value: v, PolicyVersion: pv}
		if !isRiskBucketDimension(key.Dimension) {
			rows.Close()
			return state, riskbucket.FillEvent{}, fmt.Errorf("%w: aggregate fill dimension", ErrRiskBucketReplayMismatch)
		}
		if decisionBuckets[decisionID] == nil {
			decisionBuckets[decisionID] = map[riskbucket.BucketKey]bool{}
		}
		if decisionBuckets[decisionID][key] {
			rows.Close()
			return state, riskbucket.FillEvent{}, fmt.Errorf("%w: duplicate aggregate fill bucket", ErrRiskBucketReplayMismatch)
		}
		decisionBuckets[decisionID][key] = true
		usage := state.Buckets[key]
		if usage.LimitMinor == "" {
			usage.LimitMinor = limit
		} else {
			currentLimit, currentErr := parseRiskMinor(usage.LimitMinor)
			candidateLimit, candidateErr := parseRiskMinor(limit)
			if currentErr != nil || candidateErr != nil {
				rows.Close()
				return state, riskbucket.FillEvent{}, fmt.Errorf("%w: aggregate fill limit", ErrRiskBucketReplayMismatch)
			}
			if candidateLimit.Cmp(currentLimit) < 0 {
				usage.LimitMinor = candidateLimit.String()
			}
		}
		var addErr error
		usage.HeldMinor, addErr = addRiskMinor(usage.HeldMinor, held)
		if addErr != nil {
			rows.Close()
			return state, riskbucket.FillEvent{}, addErr
		}
		usage.FilledMinor, addErr = addRiskMinor(usage.FilledMinor, filled)
		if addErr != nil {
			rows.Close()
			return state, riskbucket.FillEvent{}, addErr
		}
		usage.OverageMinor, addErr = addRiskMinor(usage.OverageMinor, overage)
		if addErr != nil {
			rows.Close()
			return state, riskbucket.FillEvent{}, addErr
		}
		if usage.Latches == nil {
			usage.Latches = map[riskbucket.Latch]bool{}
		}
		usage.Latches[riskbucket.LatchRiskOverage] = usage.Latches[riskbucket.LatchRiskOverage] || ol != 0
		usage.Latches[riskbucket.LatchUnknownActualRisk] = usage.Latches[riskbucket.LatchUnknownActualRisk] || ul != 0
		state.Buckets[key] = usage
	}
	if err := rows.Close(); err != nil {
		return state, riskbucket.FillEvent{}, err
	}
	for _, seen := range decisionBuckets {
		if len(seen) != len(riskbucket.RequiredDimensionOrder()) {
			return state, riskbucket.FillEvent{}, fmt.Errorf("%w: aggregate decision bucket count", ErrRiskBucketReplayMismatch)
		}
	}
	if len(decisionBuckets) == 0 || len(state.Buckets) != len(riskbucket.RequiredDimensionOrder()) {
		return state, riskbucket.FillEvent{}, fmt.Errorf("%w: fill bucket count", ErrRiskBucketReplayMismatch)
	}
	orders, err := tx.QueryContext(ctx, `SELECT o.order_key,o.order_id,o.order_quantity,o.cumulative_fill,o.quote_currency,o.base_currency,o.reservation_policy_digest FROM risk_bucket_orders o JOIN risk_bucket_final_decisions d ON d.decision_id=o.decision_id WHERE d.account_ref=? AND d.market=? AND d.symbol=? AND d.owner_prospective_generation=? ORDER BY d.owner_sequence,o.order_key`, target.account, target.market, target.symbol, target.prospective)
	if err != nil {
		return state, riskbucket.FillEvent{}, err
	}
	orderKeys := map[string]string{}
	brokerOrderIDs := map[string]string{}
	for orders.Next() {
		var orderKey, orderID, quote, base, digest string
		var quantity, watermark uint64
		if err := orders.Scan(&orderKey, &orderID, &quantity, &watermark, &quote, &base, &digest); err != nil {
			orders.Close()
			return state, riskbucket.FillEvent{}, err
		}
		reserved, err := riskBucketOrderReserved(ctx, tx, orderKey)
		if err != nil {
			orders.Close()
			return state, riskbucket.FillEvent{}, err
		}
		transferred := map[riskbucket.BucketKey]string{}
		for key := range reserved {
			transferred[key] = "0"
		}
		if previousKey := brokerOrderIDs[orderID]; previousKey != "" && previousKey != orderKey {
			orders.Close()
			return state, riskbucket.FillEvent{}, fmt.Errorf("%w: owner broker order id collision", ErrRiskBucketReplayMismatch)
		}
		brokerOrderIDs[orderID] = orderKey
		state.Orders[orderKey] = riskbucket.OrderFillState{OrderQuantity: quantity, CumulativeFill: watermark, QuoteCurrency: quote, BaseCurrency: base, ReservedMinor: reserved, TransferredMinor: transferred, ReservationPolicyDigest: digest, Fills: map[string]riskbucket.FillRecord{}}
		orderKeys[orderKey] = orderKey
	}
	if err := orders.Close(); err != nil {
		return state, riskbucket.FillEvent{}, err
	}
	fills, err := tx.QueryContext(ctx, `SELECT f.fill_id,f.order_key,f.cumulative_fill,f.delta_quantity,CASE WHEN f.actual_known=1 OR EXISTS(SELECT 1 FROM risk_bucket_fill_actual_evidence a WHERE a.fill_id=f.fill_id) THEN 1 ELSE 0 END FROM risk_bucket_fills f JOIN risk_bucket_orders o ON o.order_key=f.order_key JOIN risk_bucket_final_decisions d ON d.decision_id=o.decision_id WHERE d.account_ref=? AND d.market=? AND d.symbol=? AND d.owner_prospective_generation=? ORDER BY f.order_key,f.cumulative_fill,f.fill_id`, target.account, target.market, target.symbol, target.prospective)
	if err != nil {
		return state, riskbucket.FillEvent{}, err
	}
	for fills.Next() {
		var fillID, orderKey string
		var cum, delta uint64
		var actualKnown int
		if err := fills.Scan(&fillID, &orderKey, &cum, &delta, &actualKnown); err != nil {
			fills.Close()
			return state, riskbucket.FillEvent{}, err
		}
		orderIdentity := orderKeys[orderKey]
		if orderIdentity == "" {
			fills.Close()
			return state, riskbucket.FillEvent{}, fmt.Errorf("%w: fill order key", ErrRiskBucketReplayMismatch)
		}
		order := state.Orders[orderIdentity]
		record := riskbucket.FillRecord{CumulativeFill: cum, DeltaQuantity: delta, TransferMinor: map[riskbucket.BucketKey]string{}, FilledMinor: map[riskbucket.BucketKey]string{}, ActualKnown: actualKnown != 0}
		alloc, err := tx.QueryContext(ctx, `SELECT r.bucket_dimension,r.bucket_value,r.policy_version,a.transfer_minor,a.filled_minor FROM risk_bucket_fill_allocations a JOIN risk_bucket_reservations r ON r.reservation_id=a.reservation_id WHERE a.fill_id=?`, fillID)
		if err != nil {
			fills.Close()
			return state, riskbucket.FillEvent{}, err
		}
		for alloc.Next() {
			var d, v, pv, transfer, filled string
			if err := alloc.Scan(&d, &v, &pv, &transfer, &filled); err != nil {
				alloc.Close()
				fills.Close()
				return state, riskbucket.FillEvent{}, err
			}
			key := riskbucket.BucketKey{Dimension: riskbucket.Dimension(d), Value: v, PolicyVersion: pv}
			record.TransferMinor[key] = transfer
			record.FilledMinor[key] = filled
			updatedTransferred, err := addRiskMinor(order.TransferredMinor[key], transfer)
			if err != nil {
				alloc.Close()
				fills.Close()
				return state, riskbucket.FillEvent{}, err
			}
			order.TransferredMinor[key] = updatedTransferred
		}
		if err := alloc.Close(); err != nil {
			fills.Close()
			return state, riskbucket.FillEvent{}, err
		}
		order.Fills[fillID] = record
		state.Orders[orderIdentity] = order
	}
	if err := fills.Close(); err != nil {
		return state, riskbucket.FillEvent{}, err
	}
	reserved := state.Orders[target.orderKey].ReservedMinor
	if len(reserved) != len(riskbucket.RequiredDimensionOrder()) {
		return state, riskbucket.FillEvent{}, fmt.Errorf("%w: target order reservation", ErrRiskBucketReplayMismatch)
	}
	targetHeld, err := riskBucketOrderHeld(ctx, tx, target.orderKey)
	if err != nil {
		return state, riskbucket.FillEvent{}, err
	}
	event := riskbucket.FillEvent{FillID: riskBucketFillID(target.orderKey, cumulative), OrderKey: target.orderKey, OrderID: target.orderID, OrderQuantity: target.quantity, NewCumulativeFill: cumulative, ReservedMinor: reserved, TargetHeldMinor: targetHeld, ReservationPolicyDigest: target.policyDigest, QuoteCurrency: target.quote, BaseCurrency: target.base, Actual: actual}
	return state, event, nil
}

func riskBucketOrderReserved(ctx context.Context, tx *sql.Tx, orderKey string) (map[riskbucket.BucketKey]string, error) {
	rows, err := tx.QueryContext(ctx, `SELECT r.bucket_dimension,r.bucket_value,r.policy_version,m.reserved_minor FROM risk_bucket_order_reservations m JOIN risk_bucket_reservations r ON r.reservation_id=m.reservation_id WHERE m.order_key=?`, orderKey)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[riskbucket.BucketKey]string{}
	for rows.Next() {
		var d, v, pv, amount string
		if err := rows.Scan(&d, &v, &pv, &amount); err != nil {
			return nil, err
		}
		out[riskbucket.BucketKey{Dimension: riskbucket.Dimension(d), Value: v, PolicyVersion: pv}] = amount
	}
	if len(out) != len(riskbucket.RequiredDimensionOrder()) {
		return nil, fmt.Errorf("%w: order reservation count", ErrRiskBucketReplayMismatch)
	}
	return out, rows.Err()
}

func riskBucketOrderHeld(ctx context.Context, tx *sql.Tx, orderKey string) (map[riskbucket.BucketKey]string, error) {
	rows, err := tx.QueryContext(ctx, `SELECT r.bucket_dimension,r.bucket_value,r.policy_version,r.held_minor FROM risk_bucket_order_reservations m JOIN risk_bucket_reservations r ON r.reservation_id=m.reservation_id WHERE m.order_key=?`, orderKey)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[riskbucket.BucketKey]string{}
	for rows.Next() {
		var d, v, pv, amount string
		if err := rows.Scan(&d, &v, &pv, &amount); err != nil {
			return nil, err
		}
		out[riskbucket.BucketKey{Dimension: riskbucket.Dimension(d), Value: v, PolicyVersion: pv}] = amount
	}
	if len(out) != len(riskbucket.RequiredDimensionOrder()) {
		return nil, fmt.Errorf("%w: order held count", ErrRiskBucketReplayMismatch)
	}
	return out, rows.Err()
}

func persistRiskBucketFillTransition(ctx context.Context, tx *sql.Tx, order riskBucketOrderRecord, event riskbucket.FillEvent, previous, next riskbucket.FillState, result riskbucket.FillResult, actual *riskbucket.ActualFillEvidence, observedAt string) error {
	orderIdentity := event.OrderKey
	if orderIdentity == "" {
		orderIdentity = event.OrderID
	}
	record := next.Orders[orderIdentity].Fills[event.FillID]
	if len(order.reservations) != len(riskbucket.RequiredDimensionOrder()) {
		return fmt.Errorf("%w: fill reservation identity", ErrRiskBucketReplayMismatch)
	}
	type reservationUpdate struct {
		reservationID, held, filled, overage string
		overageLatch, unknownLatch           bool
	}
	updates := make(map[riskbucket.BucketKey]reservationUpdate, len(next.Buckets))
	for key, usage := range next.Buckets {
		reservationID := order.reservations[key]
		previousUsage, ok := previous.Buckets[key]
		if reservationID == "" || !ok {
			return fmt.Errorf("%w: fill reservation identity", ErrRiskBucketReplayMismatch)
		}
		heldDelta, err := riskMinorMonotoneDelta(previousUsage.HeldMinor, usage.HeldMinor)
		if err != nil {
			return err
		}
		filledDelta, err := riskMinorMonotoneDelta(usage.FilledMinor, previousUsage.FilledMinor)
		if err != nil {
			return err
		}
		overageDelta, err := riskMinorMonotoneDelta(usage.OverageMinor, previousUsage.OverageMinor)
		if err != nil {
			return err
		}
		var targetHeld, targetFilled, targetOverage string
		if err := tx.QueryRowContext(ctx, `SELECT held_minor,filled_minor,overage_minor FROM risk_bucket_reservations WHERE reservation_id=?`, reservationID).Scan(&targetHeld, &targetFilled, &targetOverage); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return fmt.Errorf("%w: target fill reservation disappeared", ErrRiskBucketReplayMismatch)
			}
			return err
		}
		targetHeld, err = subtractRiskMinorExact(targetHeld, heldDelta)
		if err != nil {
			return err
		}
		targetFilled, err = addRiskMinor(targetFilled, filledDelta)
		if err != nil {
			return err
		}
		targetOverage, err = addRiskMinor(targetOverage, overageDelta)
		if err != nil {
			return err
		}
		updates[key] = reservationUpdate{reservationID: reservationID, held: targetHeld, filled: targetFilled, overage: targetOverage, overageLatch: usage.Latches[riskbucket.LatchRiskOverage], unknownLatch: usage.Latches[riskbucket.LatchUnknownActualRisk]}
	}
	var evidenceDigest, fillDigest string
	var err error
	if result.ActualEvidenceCompleted {
		evidenceDigest, err = riskBucketRecordDigest(actual)
		if err != nil {
			return err
		}
	} else {
		fillDigest, err = riskBucketRecordDigest(struct {
			FillID, OrderKey string
			Cumulative       uint64
		}{event.FillID, order.orderKey, event.NewCumulativeFill})
		if err != nil {
			return err
		}
	}
	if result.ActualEvidenceCompleted {
		if _, err := tx.ExecContext(ctx, `INSERT INTO risk_bucket_fill_actual_evidence(fill_id,evidence_digest,quote_currency,base_currency,price_quote,fx_rate_quote_to_base,allocated_fee_base_minor,price_source,price_version,price_digest,price_observed_at,price_fresh_until,fx_source,fx_version,fx_digest,fx_observed_at,fx_fresh_until,evaluated_at,observed_at) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`, event.FillID, evidenceDigest, actual.QuoteCurrency, actual.BaseCurrency, actual.PriceQuote, actual.FXRateQuoteToBase, actual.AllocatedFeeBaseMinor, actual.Price.Source, actual.Price.Version, actual.Price.Digest, canonicalRiskTime(actual.Price.ObservedAt), canonicalRiskTime(actual.Price.FreshUntil), actual.FX.Source, actual.FX.Version, actual.FX.Digest, canonicalRiskTime(actual.FX.ObservedAt), canonicalRiskTime(actual.FX.FreshUntil), canonicalRiskTime(actual.EvaluatedAt), observedAt); err != nil {
			return err
		}
	} else {
		if _, err := tx.ExecContext(ctx, `INSERT INTO risk_bucket_fills(fill_id,order_key,order_id,cumulative_fill,delta_quantity,actual_known,fill_digest,observed_at) VALUES(?,?,?,?,?,0,?,?)`, event.FillID, order.orderKey, order.orderID, event.NewCumulativeFill, result.DeltaQuantity, fillDigest, observedAt); err != nil {
			return err
		}
	}
	for key, update := range updates {
		if _, err := tx.ExecContext(ctx, `UPDATE risk_bucket_reservations SET held_minor=?,filled_minor=?,overage_minor=?,state=CASE WHEN ?='0' THEN 'FILLED' ELSE state END,updated_at=? WHERE reservation_id=?`, update.held, update.filled, update.overage, update.held, observedAt, update.reservationID); err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `UPDATE risk_bucket_reservations SET risk_overage_latched=?,unknown_actual_latched=?,updated_at=? WHERE account_ref=? AND market=? AND symbol=? AND owner_prospective_generation=? AND bucket_dimension=? AND bucket_value=? AND policy_version=?`, boolInt(update.overageLatch), boolInt(update.unknownLatch), observedAt, order.account, order.market, order.symbol, order.prospective, string(key.Dimension), key.Value, key.PolicyVersion); err != nil {
			return err
		}
		if result.ActualEvidenceCompleted {
			if _, err := tx.ExecContext(ctx, `UPDATE risk_bucket_fill_allocations SET filled_minor=? WHERE fill_id=? AND reservation_id=?`, record.FilledMinor[key], event.FillID, update.reservationID); err != nil {
				return err
			}
		} else {
			if _, err := tx.ExecContext(ctx, `INSERT INTO risk_bucket_fill_allocations(fill_id,reservation_id,transfer_minor,filled_minor) VALUES(?,?,?,?)`, event.FillID, update.reservationID, record.TransferMinor[key], record.FilledMinor[key]); err != nil {
				return err
			}
		}
	}
	if _, err := tx.ExecContext(ctx, `UPDATE risk_bucket_orders SET cumulative_fill=?,updated_at=? WHERE order_key=?`, event.NewCumulativeFill, observedAt, order.orderKey); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `UPDATE risk_bucket_owners SET risk_overage_latched=?,unknown_actual_latched=? WHERE account_ref=? AND market=? AND symbol=? AND prospective_generation=?`, boolInt(next.OwnerLatches[riskbucket.LatchRiskOverage]), boolInt(next.OwnerLatches[riskbucket.LatchUnknownActualRisk]), order.account, order.market, order.symbol, order.prospective); err != nil {
		return err
	}
	return nil
}

func verifyRiskBucketStateDigest(ctx context.Context, q riskBucketQueryer, key riskbucket.OwnerKey) error {
	_, digest, err := loadRiskBucketState(ctx, q, key)
	if err != nil {
		return err
	}
	var persisted string
	if err := q.QueryRowContext(ctx, `SELECT state_digest FROM risk_bucket_state_snapshots WHERE account_ref=? AND market=? AND symbol=? AND prospective_generation=? ORDER BY event_sequence DESC LIMIT 1`, key.AccountID, string(key.Market), key.Symbol, key.ProspectiveGeneration).Scan(&persisted); err != nil {
		return err
	}
	if persisted != digest {
		return fmt.Errorf("%w: state digest", ErrRiskBucketReplayMismatch)
	}
	return nil
}

func (j *Journal) recordRiskBucketStateTx(ctx context.Context, tx *sql.Tx, key riskbucket.OwnerKey, eventType, eventID, eventDigest, at string) error {
	_, digest, err := loadRiskBucketState(ctx, tx, key)
	if err != nil {
		return err
	}
	var sequence int64
	if err := tx.QueryRowContext(ctx, `SELECT COALESCE(MAX(event_sequence),0)+1 FROM risk_bucket_state_snapshots WHERE account_ref=? AND market=? AND symbol=? AND prospective_generation=?`, key.AccountID, string(key.Market), key.Symbol, key.ProspectiveGeneration).Scan(&sequence); err != nil {
		return err
	}
	snapshotID := eventID + ":state:" + strconv.FormatInt(sequence, 10)
	if _, err := tx.ExecContext(ctx, `INSERT INTO risk_bucket_state_snapshots(snapshot_id,account_ref,market,symbol,prospective_generation,state_digest,event_sequence,created_at) VALUES(?,?,?,?,?,?,?,?)`, snapshotID, key.AccountID, string(key.Market), key.Symbol, key.ProspectiveGeneration, digest, sequence, at); err != nil {
		return err
	}
	payload, _ := json.Marshal(map[string]any{"event_id": eventID, "sequence": sequence})
	journalEventID := eventID + ":risk-event:" + strconv.FormatInt(sequence, 10)
	_, err = tx.ExecContext(ctx, `INSERT INTO risk_bucket_events(event_id,account_ref,market,symbol,prospective_generation,event_type,event_digest,payload,created_at) VALUES(?,?,?,?,?,?,?,?,?)`, journalEventID, key.AccountID, string(key.Market), key.Symbol, key.ProspectiveGeneration, eventType, eventDigest, string(payload), at)
	return err
}

func latchRiskBucketScope(ctx context.Context, tx *sql.Tx, key riskbucket.OwnerKey, latch, detail, at string) error {
	if latch != "REPLAY_MISMATCH" && latch != "ORPHAN_FILL" {
		return fmt.Errorf("invalid risk latch %s", latch)
	}
	_, err := tx.ExecContext(ctx, `INSERT INTO risk_bucket_scope_latches(account_ref,market,symbol,prospective_generation,latch,detail,first_seen_at,last_seen_at) VALUES(?,?,?,?,?,?,?,?) ON CONFLICT(account_ref,market,symbol,prospective_generation,latch) DO UPDATE SET detail=excluded.detail,last_seen_at=excluded.last_seen_at`, key.AccountID, string(key.Market), key.Symbol, key.ProspectiveGeneration, latch, detail, at, at)
	return err
}

func latchRiskBucketFillFailure(ctx context.Context, tx *sql.Tx, key riskbucket.OwnerKey, fill AppliedFill, latch, detail string) error {
	if err := latchRiskBucketScope(ctx, tx, key, latch, detail, fill.CommittedAt); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `UPDATE risk_bucket_owners SET unknown_actual_latched=1 WHERE account_ref=? AND market=? AND symbol=? AND prospective_generation=?`, key.AccountID, string(key.Market), key.Symbol, key.ProspectiveGeneration); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `UPDATE risk_bucket_reservations SET unknown_actual_latched=1,updated_at=? WHERE account_ref=? AND market=? AND symbol=? AND owner_prospective_generation=?`, fill.CommittedAt, key.AccountID, string(key.Market), key.Symbol, key.ProspectiveGeneration); err != nil {
		return err
	}
	payload, _ := json.Marshal(struct{ OrderID, Account, Market, Symbol, Cumulative, Detail string }{fill.OrderID, fill.AccountRef, fill.Market, fill.Symbol, fill.CumulativeQuantity, detail})
	digest := sha256.Sum256(payload)
	digestText := hex.EncodeToString(digest[:])
	id := "risk-unaccounted:" + digestText[:32]
	_, err := tx.ExecContext(ctx, `INSERT OR IGNORE INTO risk_bucket_events(event_id,account_ref,market,symbol,prospective_generation,event_type,event_digest,payload,created_at) VALUES(?,?,?,?,?,'FILL_UNACCOUNTED',?,?,?)`, id, key.AccountID, string(key.Market), key.Symbol, key.ProspectiveGeneration, digestText, string(payload), fill.CommittedAt)
	return err
}

func latchRiskBucketFillFailureForScope(ctx context.Context, tx *sql.Tx, fill AppliedFill, detail string) error {
	rows, err := tx.QueryContext(ctx, `SELECT DISTINCT d.account_ref,d.market,d.symbol,d.owner_prospective_generation FROM risk_bucket_orders o JOIN risk_bucket_final_decisions d ON d.decision_id=o.decision_id JOIN risk_bucket_owners ow ON ow.account_ref=d.account_ref AND ow.market=d.market AND ow.symbol=d.symbol AND ow.prospective_generation=d.owner_prospective_generation WHERE o.order_id=? AND d.account_ref=? AND d.market=? AND d.symbol=? AND ow.released_at IS NULL`, fill.OrderID, strings.TrimSpace(fill.AccountRef), strings.ToUpper(strings.TrimSpace(fill.Market)), strings.ToUpper(strings.TrimSpace(fill.Symbol)))
	if err != nil {
		return err
	}
	defer rows.Close()
	var keys []riskbucket.OwnerKey
	for rows.Next() {
		var key riskbucket.OwnerKey
		if err := rows.Scan(&key.AccountID, &key.Market, &key.Symbol, &key.ProspectiveGeneration); err != nil {
			return err
		}
		keys = append(keys, key)
	}
	if err := rows.Err(); err != nil {
		return err
	}
	for _, key := range keys {
		if err := latchRiskBucketFillFailure(ctx, tx, key, fill, "REPLAY_MISMATCH", detail); err != nil {
			return err
		}
	}
	return nil
}
