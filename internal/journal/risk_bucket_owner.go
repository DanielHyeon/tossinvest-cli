package journal

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/riskbucket"
	"github.com/JungHoonGhae/tossinvest-cli/internal/riskcalc"
)

var (
	ErrRiskBucketOwnerBindBlocked    = errors.New("journal: risk bucket owner bind blocked")
	ErrRiskBucketOwnerReleaseBlocked = errors.New("journal: risk bucket owner release blocked")
)

const riskBucketBrokerZeroSource = "TOSS_OFFICIAL_OPEN_API"

// schemaV24 is additive. schemaV23 is released history and must remain byte
// identical: official broker-zero evidence and release receipts are companion
// records, while legacy reconcile rows remain observation-unknown and fail
// owner release closed.
const schemaV24 = `
CREATE TABLE risk_bucket_broker_zero_observations (
 observation_id TEXT PRIMARY KEY,
 account_ref TEXT NOT NULL,
 market TEXT NOT NULL CHECK(market IN ('KR','US')),
 symbol TEXT NOT NULL,
 actual_position_generation INTEGER NOT NULL CHECK(actual_position_generation>0),
 position_id TEXT NOT NULL REFERENCES positions(id),
 position_version INTEGER NOT NULL CHECK(position_version>0),
 reconcile_state_id TEXT NOT NULL REFERENCES reconcile_states(id),
 official_source TEXT NOT NULL CHECK(official_source='TOSS_OFFICIAL_OPEN_API'),
 broker_quantity TEXT NOT NULL CHECK(broker_quantity='0'),
 broker_as_of TEXT NOT NULL,
 capability_version TEXT NOT NULL,
 build_version TEXT NOT NULL,
 source_version TEXT NOT NULL,
 payload_digest TEXT NOT NULL,
 position_adjustment_id TEXT REFERENCES position_adjustments(id),
 position_adjustment_digest TEXT,
 record_digest TEXT NOT NULL UNIQUE,
 recorded_at TEXT NOT NULL,
 UNIQUE(account_ref,market,symbol,actual_position_generation,observation_id)
) STRICT;
CREATE INDEX idx_risk_bucket_broker_zero_scope
 ON risk_bucket_broker_zero_observations(account_ref,market,symbol,actual_position_generation,broker_as_of);
ALTER TABLE reconcile_states ADD COLUMN broker_zero_observation_id TEXT REFERENCES risk_bucket_broker_zero_observations(observation_id);
ALTER TABLE reconcile_states ADD COLUMN broker_zero_observation_digest TEXT;
ALTER TABLE reconcile_states ADD COLUMN scope_market TEXT CHECK(scope_market IN ('KR','US'));
DROP INDEX idx_reconcile_active;
CREATE UNIQUE INDEX idx_reconcile_active_legacy_symbol
 ON reconcile_states(account_ref,symbol)
 WHERE released_at IS NULL AND symbol IS NOT NULL AND scope_market IS NULL;
CREATE UNIQUE INDEX idx_reconcile_active_market_symbol
 ON reconcile_states(account_ref,symbol,scope_market)
 WHERE released_at IS NULL AND symbol IS NOT NULL AND scope_market IS NOT NULL;
CREATE TRIGGER reconcile_active_scope_no_overlap_before_insert BEFORE INSERT ON reconcile_states
WHEN (NEW.symbol IS NULL AND NEW.scope_market IS NOT NULL)
 OR (NEW.released_at IS NULL AND NEW.symbol IS NOT NULL AND EXISTS(
  SELECT 1 FROM reconcile_states r
   WHERE r.account_ref=NEW.account_ref AND r.symbol=NEW.symbol AND r.released_at IS NULL
    AND (r.scope_market IS NULL OR NEW.scope_market IS NULL OR r.scope_market=NEW.scope_market)))
BEGIN SELECT RAISE(ABORT,'active RECONCILE scope overlaps global or exact market'); END;
CREATE TRIGGER reconcile_active_scope_no_overlap_before_update BEFORE UPDATE ON reconcile_states
WHEN (NEW.symbol IS NULL AND NEW.scope_market IS NOT NULL)
 OR (NEW.released_at IS NULL AND NEW.symbol IS NOT NULL AND EXISTS(
  SELECT 1 FROM reconcile_states r
   WHERE r.id<>OLD.id AND r.account_ref=NEW.account_ref AND r.symbol=NEW.symbol AND r.released_at IS NULL
    AND (r.scope_market IS NULL OR NEW.scope_market IS NULL OR r.scope_market=NEW.scope_market)))
BEGIN SELECT RAISE(ABORT,'active RECONCILE scope overlaps global or exact market'); END;
CREATE TABLE risk_bucket_owner_release_receipts (
 account_ref TEXT NOT NULL,
 market TEXT NOT NULL CHECK(market IN ('KR','US')),
 symbol TEXT NOT NULL,
 prospective_generation TEXT NOT NULL,
 actual_generation TEXT NOT NULL,
 campaign_id TEXT NOT NULL REFERENCES position_campaigns(id),
 campaign_version INTEGER NOT NULL CHECK(campaign_version>0),
 position_id TEXT NOT NULL REFERENCES positions(id),
 position_version INTEGER NOT NULL CHECK(position_version>0),
 reconcile_state_id TEXT NOT NULL REFERENCES reconcile_states(id),
 observation_id TEXT NOT NULL REFERENCES risk_bucket_broker_zero_observations(observation_id),
 observation_digest TEXT NOT NULL,
 predecessor_event_sequence INTEGER NOT NULL CHECK(predecessor_event_sequence>0),
 predecessor_state_digest TEXT NOT NULL,
 release_event_id TEXT NOT NULL UNIQUE REFERENCES risk_bucket_events(event_id),
 release_digest TEXT NOT NULL UNIQUE,
 release_payload TEXT NOT NULL,
 released_at TEXT NOT NULL,
 PRIMARY KEY(account_ref,market,symbol,prospective_generation),
 FOREIGN KEY(account_ref,market,symbol,prospective_generation)
  REFERENCES risk_bucket_owners(account_ref,market,symbol,prospective_generation)
) STRICT;
CREATE TRIGGER risk_bucket_broker_zero_observations_no_update BEFORE UPDATE ON risk_bucket_broker_zero_observations BEGIN SELECT RAISE(ABORT,'risk bucket broker zero observations are immutable'); END;
CREATE TRIGGER risk_bucket_broker_zero_observations_no_delete BEFORE DELETE ON risk_bucket_broker_zero_observations BEGIN SELECT RAISE(ABORT,'risk bucket broker zero observations are immutable'); END;
CREATE TRIGGER risk_bucket_owner_release_receipts_no_update BEFORE UPDATE ON risk_bucket_owner_release_receipts BEGIN SELECT RAISE(ABORT,'risk bucket owner release receipts are immutable'); END;
CREATE TRIGGER risk_bucket_owner_release_receipts_no_delete BEFORE DELETE ON risk_bucket_owner_release_receipts BEGIN SELECT RAISE(ABORT,'risk bucket owner release receipts are immutable'); END;
`

// RiskBucketOwnerLifecycleError identifies the journal fact that refused an
// owner transition. Callers supply neither that fact nor a clean/dirty enum;
// both are derived from authoritative rows in the transition transaction.
type RiskBucketOwnerLifecycleError struct {
	Operation     string
	BlockingField string
}

func (e *RiskBucketOwnerLifecycleError) Error() string {
	return fmt.Sprintf("journal: risk bucket owner %s blocked by %s", e.Operation, e.BlockingField)
}

func (e *RiskBucketOwnerLifecycleError) Unwrap() error {
	if e.Operation == "bind" {
		return ErrRiskBucketOwnerBindBlocked
	}
	return ErrRiskBucketOwnerReleaseBlocked
}

type RiskBucketOwnerBindResult struct {
	Bound, AlreadyBound bool
	ActualGeneration    string
}

type RiskBucketOwnerReleaseResult struct {
	Released, AlreadyReleased bool
}

type RiskBucketBrokerZeroObservationResult struct {
	Recorded, Idempotent bool
	RecordDigest         string
}

// riskBucketOfficialZeroCapability is opaque outside this package and has no
// production constructor in this change. Its seal binds the exact official
// response identity; scalar caller claims cannot be passed to the recorder.
// A future official holdings adapter must own the sole production mint path.
type riskBucketOfficialZeroCapability struct {
	owner                                          riskbucket.OwnerKey
	brokerAsOf                                     time.Time
	capabilityVersion, buildVersion, sourceVersion string
	payloadDigest                                  string
	seal                                           string
}

func (c riskBucketOfficialZeroCapability) expectedSeal() (string, error) {
	return riskBucketRecordDigest(struct {
		Domain                                         string
		Owner                                          riskbucket.OwnerKey
		OfficialSource, BrokerQuantity                 string
		BrokerAsOf                                     string
		CapabilityVersion, BuildVersion, SourceVersion string
		PayloadDigest                                  string
	}{
		Domain: "risk-bucket-official-zero-capability-v1", Owner: c.owner,
		OfficialSource: riskBucketBrokerZeroSource, BrokerQuantity: "0",
		BrokerAsOf: canonicalRiskTime(c.brokerAsOf), CapabilityVersion: c.capabilityVersion,
		BuildVersion: c.buildVersion, SourceVersion: c.sourceVersion, PayloadDigest: c.payloadDigest,
	})
}

func (c riskBucketOfficialZeroCapability) valid() bool {
	if c.seal == "" || c.brokerAsOf.IsZero() || c.capabilityVersion == "" || c.buildVersion == "" ||
		c.sourceVersion == "" || c.payloadDigest == "" || strings.TrimSpace(c.capabilityVersion) != c.capabilityVersion ||
		strings.TrimSpace(c.buildVersion) != c.buildVersion || strings.TrimSpace(c.sourceVersion) != c.sourceVersion ||
		strings.TrimSpace(c.payloadDigest) != c.payloadDigest {
		return false
	}
	expected, err := c.expectedSeal()
	return err == nil && c.seal == expected
}

type riskBucketBrokerZeroRecord struct {
	ObservationID, AccountRef, Market, Symbol      string
	ActualPositionGeneration                       int64
	PositionID                                     string
	PositionVersion                                int64
	ReconcileStateID                               string
	OfficialSource, BrokerQuantity                 string
	BrokerAsOf                                     string
	CapabilityVersion, BuildVersion                string
	SourceVersion, PayloadDigest                   string
	PositionAdjustmentID, PositionAdjustmentDigest string
	RecordedAt                                     string
}

type riskBucketBrokerZeroAuthority struct {
	Record              riskBucketBrokerZeroRecord
	RecordDigest        string
	ReconcileCause      string
	ReconcileRelease    string
	ReconcileReleasedAt string
}

type riskBucketOwnerReleaseSeal struct {
	AccountRef               string `json:"account_ref"`
	Market                   string `json:"market"`
	Symbol                   string `json:"symbol"`
	ProspectiveGeneration    string `json:"prospective_generation"`
	ActualGeneration         string `json:"actual_generation"`
	LaneID                   string `json:"lane_id"`
	CampaignID               string `json:"campaign_id"`
	CampaignVersion          int64  `json:"campaign_version"`
	PositionID               string `json:"position_id"`
	PositionVersion          int64  `json:"position_version"`
	ReconcileStateID         string `json:"reconcile_state_id"`
	ObservationID            string `json:"observation_id"`
	ObservationDigest        string `json:"observation_digest"`
	PredecessorEventSequence int64  `json:"predecessor_event_sequence"`
	PredecessorStateDigest   string `json:"predecessor_state_digest"`
	ReleasedAt               string `json:"released_at"`
}

func ownerLifecycleBlocked(operation, field string) error {
	return &RiskBucketOwnerLifecycleError{Operation: operation, BlockingField: field}
}

// recordRiskBucketBrokerZeroObservation consumes only an opaque sealed
// capability. There is deliberately no production constructor/caller until an
// official holdings adapter can mint it from an immutable Open API response.
func (j *Journal) recordRiskBucketBrokerZeroObservation(ctx context.Context, capability riskBucketOfficialZeroCapability) (RiskBucketBrokerZeroObservationResult, error) {
	if !capability.valid() {
		return RiskBucketBrokerZeroObservationResult{}, ownerLifecycleBlocked("broker_zero", "official_capability")
	}
	if err := validateRiskBucketOwnerKey(capability.owner, "broker_zero"); err != nil {
		return RiskBucketBrokerZeroObservationResult{}, err
	}
	tx, err := j.db.BeginTx(ctx, nil)
	if err != nil {
		return RiskBucketBrokerZeroObservationResult{}, err
	}
	defer tx.Rollback()
	key := capability.owner
	observationID := "official-zero-" + capability.seal[:32]
	var actualText string
	if err := tx.QueryRowContext(ctx, `SELECT actual_generation FROM risk_bucket_owners WHERE account_ref=? AND market=? AND symbol=? AND prospective_generation=? AND released_at IS NULL AND actual_generation IS NOT NULL`, key.AccountID, string(key.Market), key.Symbol, key.ProspectiveGeneration).Scan(&actualText); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return RiskBucketBrokerZeroObservationResult{}, ownerLifecycleBlocked("broker_zero", "active_owner")
		}
		return RiskBucketBrokerZeroObservationResult{}, err
	}
	actual, err := strconv.ParseInt(actualText, 10, 64)
	if err != nil || actual <= 0 {
		return RiskBucketBrokerZeroObservationResult{}, ownerLifecycleBlocked("broker_zero", "actual_generation")
	}
	var positionID, positionState, positionQuantity, closedAt string
	var positionVersion int64
	if err := tx.QueryRowContext(ctx, `SELECT p.id,p.state,p.quantity,COALESCE(p.closed_at,''),v.version FROM positions p JOIN position_projection_versions v ON v.position_id=p.id AND v.account_ref=p.account_ref AND v.market=p.market AND v.symbol=p.symbol AND v.generation=p.instance_seq AND v.state=p.state WHERE p.account_ref=? AND lower(p.market)=? AND p.symbol=? AND p.instance_seq=?`, key.AccountID, normaliseMarket(string(key.Market)), key.Symbol, actual).
		Scan(&positionID, &positionState, &positionQuantity, &closedAt, &positionVersion); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return RiskBucketBrokerZeroObservationResult{}, ownerLifecycleBlocked("broker_zero", "position_generation")
		}
		return RiskBucketBrokerZeroObservationResult{}, err
	}
	zero, zeroErr := riskcalc.CompareDecimal(positionQuantity, "0")
	if positionState != "CLOSED" || closedAt == "" || zeroErr != nil || zero != 0 {
		return RiskBucketBrokerZeroObservationResult{}, ownerLifecycleBlocked("broker_zero", "position_closed_zero")
	}
	var reconcileID, reconcileCause, reconcileReleaseCause, reconcileReleasedAt string
	var linkedID, linkedDigest sql.NullString
	if err := tx.QueryRowContext(ctx, `SELECT id,cause,COALESCE(release_cause,''),COALESCE(released_at,''),broker_zero_observation_id,broker_zero_observation_digest
		FROM reconcile_states WHERE account_ref=? AND symbol=? AND released_at IS NOT NULL
		AND (scope_market IS NULL OR scope_market=?)
		AND ((broker_zero_observation_id IS NULL AND broker_zero_observation_digest IS NULL) OR broker_zero_observation_id=?)
		ORDER BY released_at DESC,id DESC LIMIT 1`, key.AccountID, key.Symbol, string(key.Market), observationID).
		Scan(&reconcileID, &reconcileCause, &reconcileReleaseCause, &reconcileReleasedAt, &linkedID, &linkedDigest); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return RiskBucketBrokerZeroObservationResult{}, ownerLifecycleBlocked("broker_zero", "released_reconciliation")
		}
		return RiskBucketBrokerZeroObservationResult{}, err
	}
	if reconcileCause != ReconcileCauseQuantityMismatch ||
		(reconcileReleaseCause != ReconcileReleaseRecheckMatched && reconcileReleaseCause != ReconcileReleaseAdjustmentApplied) {
		return RiskBucketBrokerZeroObservationResult{}, ownerLifecycleBlocked("broker_zero", "released_reconciliation")
	}
	brokerAsOf := canonicalRiskTime(capability.brokerAsOf)
	if fresh, err := timestampStrictlyAfter(brokerAsOf, closedAt); err != nil || !fresh {
		return RiskBucketBrokerZeroObservationResult{}, ownerLifecycleBlocked("broker_zero", "observation_stale")
	}
	if releaseAfter, err := timestampStrictlyAfter(reconcileReleasedAt, brokerAsOf); err != nil || !releaseAfter {
		return RiskBucketBrokerZeroObservationResult{}, ownerLifecycleBlocked("broker_zero", "observation_after_reconcile_release")
	}
	record := riskBucketBrokerZeroRecord{
		ObservationID: observationID, AccountRef: key.AccountID, Market: string(key.Market), Symbol: key.Symbol,
		ActualPositionGeneration: actual, PositionID: positionID, PositionVersion: positionVersion,
		ReconcileStateID: reconcileID, OfficialSource: riskBucketBrokerZeroSource, BrokerQuantity: "0",
		BrokerAsOf: brokerAsOf, CapabilityVersion: capability.capabilityVersion, BuildVersion: capability.buildVersion,
		SourceVersion: capability.sourceVersion, PayloadDigest: capability.payloadDigest, RecordedAt: reconcileReleasedAt,
	}
	if reconcileReleaseCause == ReconcileReleaseAdjustmentApplied {
		var adjustment Adjustment
		if err := tx.QueryRowContext(ctx, adjustmentSelect+` WHERE position_id=? ORDER BY broker_as_of DESC,created_at DESC,id DESC LIMIT 1`, positionID).
			Scan(&adjustment.ID, &adjustment.PositionID, &adjustment.Kind, &adjustment.ExpectedPrevQuantity,
				&adjustment.PrevQuantity, &adjustment.NewQuantity, &adjustment.PrevAvgPrice, &adjustment.NewAvgPrice,
				&adjustment.BrokerAsOf, &adjustment.Evidence, &adjustment.CreatedAt); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return RiskBucketBrokerZeroObservationResult{}, ownerLifecycleBlocked("broker_zero", "zero_adjustment")
			}
			return RiskBucketBrokerZeroObservationResult{}, err
		}
		adjustmentZero, err := riskcalc.CompareDecimal(adjustment.NewQuantity, "0")
		if err != nil || adjustmentZero != 0 || strings.TrimSpace(adjustment.Evidence) == "" {
			return RiskBucketBrokerZeroObservationResult{}, ownerLifecycleBlocked("broker_zero", "zero_adjustment")
		}
		if later, err := timestampStrictlyAfter(brokerAsOf, adjustment.BrokerAsOf); err != nil || !later {
			return RiskBucketBrokerZeroObservationResult{}, ownerLifecycleBlocked("broker_zero", "post_adjustment_recheck")
		}
		record.PositionAdjustmentID = adjustment.ID
		record.PositionAdjustmentDigest, err = riskBucketRecordDigest(adjustment)
		if err != nil {
			return RiskBucketBrokerZeroObservationResult{}, err
		}
	}
	recordDigest, err := riskBucketRecordDigest(record)
	if err != nil {
		return RiskBucketBrokerZeroObservationResult{}, err
	}
	var storedDigest string
	if err := tx.QueryRowContext(ctx, `SELECT record_digest FROM risk_bucket_broker_zero_observations WHERE observation_id=?`, observationID).Scan(&storedDigest); err == nil {
		if storedDigest != recordDigest || !linkedID.Valid || linkedID.String != observationID || !linkedDigest.Valid || linkedDigest.String != recordDigest {
			return RiskBucketBrokerZeroObservationResult{}, ownerLifecycleBlocked("broker_zero", "observation_replay")
		}
		if err := tx.Commit(); err != nil {
			return RiskBucketBrokerZeroObservationResult{}, err
		}
		return RiskBucketBrokerZeroObservationResult{Idempotent: true, RecordDigest: recordDigest}, nil
	} else if !errors.Is(err, sql.ErrNoRows) {
		return RiskBucketBrokerZeroObservationResult{}, err
	}
	if linkedID.Valid || linkedDigest.Valid {
		return RiskBucketBrokerZeroObservationResult{}, ownerLifecycleBlocked("broker_zero", "observation_replay")
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO risk_bucket_broker_zero_observations(observation_id,account_ref,market,symbol,actual_position_generation,position_id,position_version,reconcile_state_id,official_source,broker_quantity,broker_as_of,capability_version,build_version,source_version,payload_digest,position_adjustment_id,position_adjustment_digest,record_digest,recorded_at) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`, record.ObservationID, record.AccountRef, record.Market, record.Symbol, record.ActualPositionGeneration, record.PositionID, record.PositionVersion, record.ReconcileStateID, record.OfficialSource, record.BrokerQuantity, record.BrokerAsOf, record.CapabilityVersion, record.BuildVersion, record.SourceVersion, record.PayloadDigest, nullableString(record.PositionAdjustmentID), nullableString(record.PositionAdjustmentDigest), recordDigest, record.RecordedAt)
	if err != nil {
		return RiskBucketBrokerZeroObservationResult{}, err
	}
	result, err := tx.ExecContext(ctx, `UPDATE reconcile_states SET scope_market=?,broker_zero_observation_id=?,broker_zero_observation_digest=?
		WHERE id=? AND (scope_market IS NULL OR scope_market=?) AND broker_zero_observation_id IS NULL
		AND broker_zero_observation_digest IS NULL`, string(key.Market), record.ObservationID, recordDigest, reconcileID, string(key.Market))
	if err != nil {
		return RiskBucketBrokerZeroObservationResult{}, err
	}
	if affected, err := result.RowsAffected(); err != nil || affected != 1 {
		return RiskBucketBrokerZeroObservationResult{}, ownerLifecycleBlocked("broker_zero", "observation_replay")
	}
	if err := tx.Commit(); err != nil {
		return RiskBucketBrokerZeroObservationResult{}, err
	}
	return RiskBucketBrokerZeroObservationResult{Recorded: true, RecordDigest: recordDigest}, nil
}

// bindRiskBucketOwnerActual derives the Position generation from journal
// lineage. It is intentionally package-private: no runtime caller can provide
// an actual generation or bypass the authoritative campaign/Position checks.
func (j *Journal) bindRiskBucketOwnerActual(ctx context.Context, key riskbucket.OwnerKey) (RiskBucketOwnerBindResult, error) {
	if err := validateRiskBucketOwnerKey(key, "bind"); err != nil {
		return RiskBucketOwnerBindResult{}, err
	}
	tx, err := j.db.BeginTx(ctx, nil)
	if err != nil {
		return RiskBucketOwnerBindResult{}, err
	}
	defer tx.Rollback()
	result, err := j.bindRiskBucketOwnerActualInTx(ctx, tx, key, j.nowString())
	if err != nil {
		return RiskBucketOwnerBindResult{}, err
	}
	if err := tx.Commit(); err != nil {
		return RiskBucketOwnerBindResult{}, err
	}
	return result, nil
}

func (j *Journal) bindRiskBucketOwnerActualInTx(ctx context.Context, tx *sql.Tx, key riskbucket.OwnerKey, at string) (RiskBucketOwnerBindResult, error) {
	var lane, campaign string
	var current sql.NullString
	err := tx.QueryRowContext(ctx, `SELECT lane_id,campaign_id,actual_generation FROM risk_bucket_owners
		WHERE account_ref=? AND market=? AND symbol=? AND prospective_generation=? AND released_at IS NULL`,
		key.AccountID, string(key.Market), key.Symbol, key.ProspectiveGeneration).Scan(&lane, &campaign, &current)
	if errors.Is(err, sql.ErrNoRows) {
		return RiskBucketOwnerBindResult{}, ownerLifecycleBlocked("bind", "active_owner")
	}
	if err != nil {
		return RiskBucketOwnerBindResult{}, err
	}
	if err := verifyRiskBucketStateDigest(ctx, tx, key); err != nil {
		return RiskBucketOwnerBindResult{}, err
	}

	var campaignAccount, campaignMarket, campaignSymbol, campaignLane, decisionID, token, campaignState string
	var actual sql.NullInt64
	err = tx.QueryRowContext(ctx, `SELECT account_ref,market,symbol,lane_id,decision_id,prospective_token,
		actual_position_generation,state FROM position_campaigns WHERE id=?`, campaign).
		Scan(&campaignAccount, &campaignMarket, &campaignSymbol, &campaignLane, &decisionID, &token, &actual, &campaignState)
	if errors.Is(err, sql.ErrNoRows) {
		return RiskBucketOwnerBindResult{}, ownerLifecycleBlocked("bind", "campaign")
	}
	if err != nil {
		return RiskBucketOwnerBindResult{}, err
	}
	if campaignAccount != key.AccountID || normaliseMarket(campaignMarket) != normaliseMarket(string(key.Market)) ||
		normaliseSymbol(campaignSymbol) != key.Symbol || campaignLane != lane || token != key.ProspectiveGeneration {
		return RiskBucketOwnerBindResult{}, ownerLifecycleBlocked("bind", "campaign_identity")
	}
	if !actual.Valid || actual.Int64 <= 0 {
		return RiskBucketOwnerBindResult{}, ownerLifecycleBlocked("bind", "actual_generation")
	}
	if campaignState != "PLANNED" && campaignState != "ACTIVE" {
		return RiskBucketOwnerBindResult{}, ownerLifecycleBlocked("bind", "campaign_state")
	}
	var claimCount int
	if err := tx.QueryRowContext(ctx, `SELECT count(*) FROM position_campaign_claims WHERE campaign_id=?
		AND account_ref=? AND market=? AND symbol=? AND prospective_token=? AND actual_position_generation=?`,
		campaign, key.AccountID, campaignMarket, campaignSymbol, key.ProspectiveGeneration, actual.Int64).Scan(&claimCount); err != nil {
		return RiskBucketOwnerBindResult{}, err
	}
	if claimCount != 1 {
		return RiskBucketOwnerBindResult{}, ownerLifecycleBlocked("bind", "campaign_claim")
	}

	var positionState string
	var entryDecision sql.NullString
	var positionCount, latestGeneration int
	if err := tx.QueryRowContext(ctx, `SELECT count(*),COALESCE(MIN(p.state),''),MIN(p.entry_decision_id)
		FROM positions p JOIN position_projection_versions v ON v.position_id=p.id
		 AND v.account_ref=p.account_ref AND v.market=p.market AND v.symbol=p.symbol
		 AND v.generation=p.instance_seq AND v.state=p.state
		WHERE p.account_ref=? AND p.market=? AND p.symbol=? AND p.instance_seq=?`,
		key.AccountID, campaignMarket, campaignSymbol, actual.Int64).Scan(&positionCount, &positionState, &entryDecision); err != nil {
		return RiskBucketOwnerBindResult{}, err
	}
	if positionCount != 1 || positionState == "CLOSED" || !entryDecision.Valid {
		return RiskBucketOwnerBindResult{}, ownerLifecycleBlocked("bind", "position_generation")
	}
	if err := tx.QueryRowContext(ctx, `SELECT COALESCE(MAX(instance_seq),0) FROM positions WHERE account_ref=? AND market=? AND symbol=?`,
		key.AccountID, campaignMarket, campaignSymbol).Scan(&latestGeneration); err != nil {
		return RiskBucketOwnerBindResult{}, err
	}
	if int64(latestGeneration) != actual.Int64 {
		return RiskBucketOwnerBindResult{}, ownerLifecycleBlocked("bind", "latest_position_generation")
	}
	var lineageCount int
	if err := tx.QueryRowContext(ctx, `SELECT count(*) FROM risk_bucket_final_decisions d
		JOIN risk_reservations r ON r.id=d.existing_reservation_id
		WHERE d.account_ref=? AND d.market=? AND d.symbol=? AND d.owner_prospective_generation=?
		 AND d.owner_lane_id=? AND d.owner_campaign_id=? AND r.decision_id=?`, key.AccountID,
		string(key.Market), key.Symbol, key.ProspectiveGeneration, lane, campaign, entryDecision.String).Scan(&lineageCount); err != nil {
		return RiskBucketOwnerBindResult{}, err
	}
	if lineageCount == 0 || decisionID != entryDecision.String {
		return RiskBucketOwnerBindResult{}, ownerLifecycleBlocked("bind", "entry_decision_lineage")
	}
	derived := strconv.FormatInt(actual.Int64, 10)
	if current.Valid {
		if current.String != derived {
			return RiskBucketOwnerBindResult{}, ownerLifecycleBlocked("bind", "set_once_generation")
		}
		return RiskBucketOwnerBindResult{AlreadyBound: true, ActualGeneration: derived}, nil
	}
	result, err := tx.ExecContext(ctx, `UPDATE risk_bucket_owners SET actual_generation=? WHERE account_ref=? AND market=?
		AND symbol=? AND prospective_generation=? AND released_at IS NULL AND actual_generation IS NULL`, derived,
		key.AccountID, string(key.Market), key.Symbol, key.ProspectiveGeneration)
	if err != nil {
		return RiskBucketOwnerBindResult{}, err
	}
	if affected, err := result.RowsAffected(); err != nil || affected != 1 {
		return RiskBucketOwnerBindResult{}, ownerLifecycleBlocked("bind", "set_once_generation")
	}
	payload, _ := json.Marshal(map[string]string{
		"account_ref": key.AccountID, "market": string(key.Market), "symbol": key.Symbol,
		"prospective_generation": key.ProspectiveGeneration, "actual_generation": derived, "campaign_id": campaign,
	})
	digest := sha256.Sum256(payload)
	digestText := hex.EncodeToString(digest[:])
	if err := j.recordRiskBucketStateTx(ctx, tx, key, "OWNER_BOUND", "owner-bind-"+digestText[:24], digestText, at); err != nil {
		return RiskBucketOwnerBindResult{}, err
	}
	return RiskBucketOwnerBindResult{Bound: true, ActualGeneration: derived}, nil
}

// applyRiskBucketOwnerBindingInTx joins owner binding to the authoritative fill
// transaction after Position and campaign projection. Semantic gaps latch entry
// for this owner, but do not roll back or delay the broker fill.
func (j *Journal) applyRiskBucketOwnerBindingInTx(ctx context.Context, tx *sql.Tx, fill AppliedFill) error {
	if strings.ToUpper(strings.TrimSpace(fill.Side)) != "BUY" {
		return nil
	}
	delta, err := campaignQuantity(fill.Delta)
	if err != nil {
		return nil
	}
	positive, err := riskcalc.CompareDecimal(delta, "0")
	if err != nil || positive <= 0 {
		return nil
	}
	rows, err := tx.QueryContext(ctx, `SELECT DISTINCT d.account_ref,d.market,d.symbol,d.owner_prospective_generation
		FROM risk_bucket_orders o JOIN risk_bucket_final_decisions d ON d.decision_id=o.decision_id
		JOIN risk_bucket_owners ow ON ow.account_ref=d.account_ref AND ow.market=d.market AND ow.symbol=d.symbol
		 AND ow.prospective_generation=d.owner_prospective_generation AND ow.released_at IS NULL
		WHERE o.order_id=? AND d.account_ref=? AND d.market=? AND d.symbol=?`, fill.OrderID,
		strings.TrimSpace(fill.AccountRef), strings.ToUpper(strings.TrimSpace(fill.Market)), normaliseSymbol(fill.Symbol))
	if err != nil {
		return err
	}
	var keys []riskbucket.OwnerKey
	for rows.Next() {
		var key riskbucket.OwnerKey
		if err := rows.Scan(&key.AccountID, &key.Market, &key.Symbol, &key.ProspectiveGeneration); err != nil {
			rows.Close()
			return err
		}
		keys = append(keys, key)
	}
	if err := rows.Close(); err != nil {
		return err
	}
	if len(keys) == 0 {
		return j.latchReleasedOwnerLateFillInTx(ctx, tx, fill)
	}
	if len(keys) != 1 {
		for _, key := range keys {
			if err := latchRiskBucketFillFailure(ctx, tx, key, fill, "REPLAY_MISMATCH", "owner bind matched multiple active scopes"); err != nil {
				return err
			}
			if err := j.recordRiskBucketStateTx(ctx, tx, key, "OWNER_BIND_REFUSED", "owner-bind-refused-"+fillObservationID(fill), "multiple-active-scopes", fill.CommittedAt); err != nil {
				return err
			}
		}
		return nil
	}
	if _, err := j.bindRiskBucketOwnerActualInTx(ctx, tx, keys[0], fill.CommittedAt); err != nil {
		var lifecycle *RiskBucketOwnerLifecycleError
		if !errors.As(err, &lifecycle) && !isRiskBucketSemanticError(err) && !errors.Is(err, sql.ErrNoRows) {
			return err
		}
		if lifecycle == nil {
			// Snapshot/replay drift is itself the durable blocker. Do not seal the
			// drift by appending a fresh state snapshot; only latch the scope while
			// preserving the authoritative fill transaction.
			return latchRiskBucketFillFailure(ctx, tx, keys[0], fill, "REPLAY_MISMATCH", err.Error())
		}
		detail := lifecycle.BlockingField + " prevented authoritative owner bind"
		if err := latchRiskBucketFillFailure(ctx, tx, keys[0], fill, "REPLAY_MISMATCH", detail); err != nil {
			return err
		}
		return j.recordRiskBucketStateTx(ctx, tx, keys[0], "OWNER_BIND_REFUSED", "owner-bind-refused-"+fillObservationID(fill), lifecycle.BlockingField, fill.CommittedAt)
	}
	return nil
}

func (j *Journal) latchReleasedOwnerLateFillInTx(ctx context.Context, tx *sql.Tx, fill AppliedFill) error {
	rows, err := tx.QueryContext(ctx, `SELECT DISTINCT d.account_ref,d.market,d.symbol,d.owner_prospective_generation
		FROM risk_bucket_orders o JOIN risk_bucket_final_decisions d ON d.decision_id=o.decision_id
		JOIN risk_bucket_owners released ON released.account_ref=d.account_ref AND released.market=d.market
		AND released.symbol=d.symbol AND released.prospective_generation=d.owner_prospective_generation
		WHERE o.order_id=? AND d.account_ref=? AND d.market=? AND d.symbol=? AND released.released_at IS NOT NULL`,
		fill.OrderID, strings.TrimSpace(fill.AccountRef), strings.ToUpper(strings.TrimSpace(fill.Market)), normaliseSymbol(fill.Symbol))
	if err != nil {
		return err
	}
	var released []riskbucket.OwnerKey
	for rows.Next() {
		var key riskbucket.OwnerKey
		if err := rows.Scan(&key.AccountID, &key.Market, &key.Symbol, &key.ProspectiveGeneration); err != nil {
			rows.Close()
			return err
		}
		released = append(released, key)
	}
	if err := rows.Close(); err != nil {
		return err
	}
	if len(released) == 0 {
		return nil
	}
	detail := "late fill observed for a released risk owner: " + fill.OrderID
	for _, old := range released {
		if err := latchRiskBucketScope(ctx, tx, old, "ORPHAN_FILL", detail, fill.CommittedAt); err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO reconcile_states(id,account_ref,symbol,cause,evidence,entered_at,released_at,release_cause,scope_market)
			SELECT ?,?,?,?, ?,?,NULL,NULL,? WHERE NOT EXISTS(SELECT 1 FROM reconcile_states
			WHERE account_ref=? AND symbol=? AND released_at IS NULL AND (scope_market IS NULL OR scope_market=?))`,
			reconcileStateID(old.AccountID, old.Symbol, ReconcileCauseQuantityMismatch, fill.CommittedAt, string(old.Market)),
			old.AccountID, old.Symbol, ReconcileCauseQuantityMismatch, detail, fill.CommittedAt, string(old.Market),
			old.AccountID, old.Symbol, string(old.Market)); err != nil {
			return err
		}
		activeRows, err := tx.QueryContext(ctx, `SELECT account_ref,market,symbol,prospective_generation FROM risk_bucket_owners
			WHERE account_ref=? AND market=? AND symbol=? AND released_at IS NULL AND prospective_generation<>?`,
			old.AccountID, string(old.Market), old.Symbol, old.ProspectiveGeneration)
		if err != nil {
			return err
		}
		var active []riskbucket.OwnerKey
		for activeRows.Next() {
			var key riskbucket.OwnerKey
			if err := activeRows.Scan(&key.AccountID, &key.Market, &key.Symbol, &key.ProspectiveGeneration); err != nil {
				activeRows.Close()
				return err
			}
			active = append(active, key)
		}
		if err := activeRows.Close(); err != nil {
			return err
		}
		for _, key := range active {
			if err := latchRiskBucketFillFailure(ctx, tx, key, fill, "ORPHAN_FILL", detail); err != nil {
				return err
			}
			if err := j.recordRiskBucketStateTx(ctx, tx, key, "LATE_RELEASED_OWNER_FILL", "late-released-owner-fill-"+fillObservationID(fill), detail, fill.CommittedAt); err != nil {
				return err
			}
		}
	}
	return nil
}

func loadRiskBucketBrokerZeroAuthority(ctx context.Context, tx *sql.Tx, key riskbucket.OwnerKey, actualGeneration int64, positionID string, positionVersion int64, observationID string) (riskBucketBrokerZeroAuthority, error) {
	query := `SELECT o.observation_id,o.account_ref,o.market,o.symbol,o.actual_position_generation,
		o.position_id,o.position_version,o.reconcile_state_id,o.official_source,o.broker_quantity,
		o.broker_as_of,o.capability_version,o.build_version,o.source_version,o.payload_digest,
		COALESCE(o.position_adjustment_id,''),COALESCE(o.position_adjustment_digest,''),o.record_digest,o.recorded_at,
		r.cause,COALESCE(r.release_cause,''),COALESCE(r.released_at,''),
		COALESCE(r.broker_zero_observation_id,''),COALESCE(r.broker_zero_observation_digest,'')
		FROM risk_bucket_broker_zero_observations o
		JOIN reconcile_states r ON r.id=o.reconcile_state_id
		WHERE o.account_ref=? AND o.market=? AND o.symbol=? AND o.actual_position_generation=?
		AND o.position_id=? AND o.position_version=? AND r.released_at IS NOT NULL AND r.scope_market=o.market`
	args := []any{key.AccountID, string(key.Market), key.Symbol, actualGeneration, positionID, positionVersion}
	if observationID != "" {
		query += ` AND o.observation_id=?`
		args = append(args, observationID)
	}
	query += ` ORDER BY o.broker_as_of DESC,o.observation_id DESC LIMIT 1`
	var authority riskBucketBrokerZeroAuthority
	var linkedID, linkedDigest string
	err := tx.QueryRowContext(ctx, query, args...).Scan(
		&authority.Record.ObservationID, &authority.Record.AccountRef, &authority.Record.Market,
		&authority.Record.Symbol, &authority.Record.ActualPositionGeneration, &authority.Record.PositionID,
		&authority.Record.PositionVersion, &authority.Record.ReconcileStateID, &authority.Record.OfficialSource,
		&authority.Record.BrokerQuantity, &authority.Record.BrokerAsOf, &authority.Record.CapabilityVersion,
		&authority.Record.BuildVersion, &authority.Record.SourceVersion, &authority.Record.PayloadDigest,
		&authority.Record.PositionAdjustmentID, &authority.Record.PositionAdjustmentDigest,
		&authority.RecordDigest, &authority.Record.RecordedAt, &authority.ReconcileCause,
		&authority.ReconcileRelease, &authority.ReconcileReleasedAt, &linkedID, &linkedDigest)
	if errors.Is(err, sql.ErrNoRows) {
		return riskBucketBrokerZeroAuthority{}, ownerLifecycleBlocked("release", "broker_zero_observation")
	}
	if err != nil {
		return riskBucketBrokerZeroAuthority{}, err
	}
	recomputed, err := riskBucketRecordDigest(authority.Record)
	if err != nil {
		return riskBucketBrokerZeroAuthority{}, err
	}
	valid := authority.Record.OfficialSource == riskBucketBrokerZeroSource && authority.Record.BrokerQuantity == "0" &&
		authority.Record.ObservationID != "" && authority.Record.CapabilityVersion != "" && authority.Record.BuildVersion != "" &&
		authority.Record.SourceVersion != "" && authority.Record.PayloadDigest != "" &&
		authority.RecordDigest == recomputed && linkedID == authority.Record.ObservationID && linkedDigest == recomputed &&
		authority.Record.RecordedAt == authority.ReconcileReleasedAt &&
		authority.ReconcileCause == ReconcileCauseQuantityMismatch &&
		(authority.ReconcileRelease == ReconcileReleaseRecheckMatched || authority.ReconcileRelease == ReconcileReleaseAdjustmentApplied)
	if !valid {
		return riskBucketBrokerZeroAuthority{}, ownerLifecycleBlocked("release", "broker_zero_observation")
	}
	if after, parseErr := timestampStrictlyAfter(authority.ReconcileReleasedAt, authority.Record.BrokerAsOf); parseErr != nil || !after {
		return riskBucketBrokerZeroAuthority{}, ownerLifecycleBlocked("release", "broker_zero_observation")
	}
	if authority.ReconcileRelease == ReconcileReleaseAdjustmentApplied {
		if authority.Record.PositionAdjustmentID == "" || authority.Record.PositionAdjustmentDigest == "" {
			return riskBucketBrokerZeroAuthority{}, ownerLifecycleBlocked("release", "post_adjustment_recheck")
		}
		adjustment, err := scanAdjustment(tx.QueryRowContext(ctx, adjustmentSelect+` WHERE id=? AND position_id=?`, authority.Record.PositionAdjustmentID, positionID))
		if err != nil {
			return riskBucketBrokerZeroAuthority{}, ownerLifecycleBlocked("release", "post_adjustment_recheck")
		}
		adjustmentDigest, err := riskBucketRecordDigest(adjustment)
		if err != nil {
			return riskBucketBrokerZeroAuthority{}, err
		}
		zero, compareErr := riskcalc.CompareDecimal(adjustment.NewQuantity, "0")
		fresh, freshErr := timestampStrictlyAfter(authority.Record.BrokerAsOf, adjustment.BrokerAsOf)
		if adjustmentDigest != authority.Record.PositionAdjustmentDigest || compareErr != nil || zero != 0 ||
			strings.TrimSpace(adjustment.Evidence) == "" || freshErr != nil || !fresh {
			return riskBucketBrokerZeroAuthority{}, ownerLifecycleBlocked("release", "post_adjustment_recheck")
		}
	} else if authority.Record.PositionAdjustmentID != "" || authority.Record.PositionAdjustmentDigest != "" {
		return riskBucketBrokerZeroAuthority{}, ownerLifecycleBlocked("release", "broker_zero_observation")
	}
	return authority, nil
}

func loadRiskBucketPredecessorSeal(ctx context.Context, tx *sql.Tx, key riskbucket.OwnerKey) (int64, string, error) {
	var sequence int64
	var digest string
	err := tx.QueryRowContext(ctx, `SELECT event_sequence,state_digest FROM risk_bucket_state_snapshots
		WHERE account_ref=? AND market=? AND symbol=? AND prospective_generation=?
		ORDER BY event_sequence DESC LIMIT 1`, key.AccountID, string(key.Market), key.Symbol, key.ProspectiveGeneration).
		Scan(&sequence, &digest)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, "", ownerLifecycleBlocked("release", "state_digest")
	}
	if err != nil {
		return 0, "", err
	}
	if sequence <= 0 || digest == "" {
		return 0, "", ownerLifecycleBlocked("release", "state_digest")
	}
	return sequence, digest, nil
}

func validateRiskBucketOwnerReleaseReceipt(ctx context.Context, tx *sql.Tx, key riskbucket.OwnerKey, lane, campaign, actualText, releasedAt string) error {
	var seal riskBucketOwnerReleaseSeal
	var eventID, releaseDigest, releasePayload, receiptReleasedAt string
	err := tx.QueryRowContext(ctx, `SELECT account_ref,market,symbol,prospective_generation,actual_generation,
		campaign_id,campaign_version,position_id,position_version,reconcile_state_id,observation_id,
		observation_digest,predecessor_event_sequence,predecessor_state_digest,release_event_id,
		release_digest,release_payload,released_at FROM risk_bucket_owner_release_receipts
		WHERE account_ref=? AND market=? AND symbol=? AND prospective_generation=?`, key.AccountID, string(key.Market), key.Symbol, key.ProspectiveGeneration).
		Scan(&seal.AccountRef, &seal.Market, &seal.Symbol, &seal.ProspectiveGeneration, &seal.ActualGeneration,
			&seal.CampaignID, &seal.CampaignVersion, &seal.PositionID, &seal.PositionVersion,
			&seal.ReconcileStateID, &seal.ObservationID, &seal.ObservationDigest,
			&seal.PredecessorEventSequence, &seal.PredecessorStateDigest, &eventID,
			&releaseDigest, &releasePayload, &receiptReleasedAt)
	if err != nil {
		return ownerLifecycleBlocked("release", "release_receipt")
	}
	seal.LaneID = lane
	seal.ReleasedAt = receiptReleasedAt
	if seal.AccountRef != key.AccountID || seal.Market != string(key.Market) || seal.Symbol != key.Symbol ||
		seal.ProspectiveGeneration != key.ProspectiveGeneration || seal.ActualGeneration != actualText ||
		seal.CampaignID != campaign || receiptReleasedAt != releasedAt {
		return ownerLifecycleBlocked("release", "release_receipt")
	}
	actualGeneration, err := strconv.ParseInt(actualText, 10, 64)
	if err != nil || actualGeneration <= 0 {
		return ownerLifecycleBlocked("release", "release_receipt")
	}
	var campaignAccount, campaignMarket, campaignSymbol, campaignLane, token, campaignState string
	var campaignActual sql.NullInt64
	var campaignVersion int64
	err = tx.QueryRowContext(ctx, `SELECT account_ref,market,symbol,lane_id,prospective_token,
		actual_position_generation,state,version FROM position_campaigns WHERE id=?`, campaign).
		Scan(&campaignAccount, &campaignMarket, &campaignSymbol, &campaignLane, &token,
			&campaignActual, &campaignState, &campaignVersion)
	if err != nil || campaignAccount != key.AccountID || normaliseMarket(campaignMarket) != normaliseMarket(string(key.Market)) ||
		normaliseSymbol(campaignSymbol) != key.Symbol || campaignLane != lane || token != key.ProspectiveGeneration ||
		!campaignActual.Valid || campaignActual.Int64 != actualGeneration || campaignState != "CLOSED" || campaignVersion != seal.CampaignVersion {
		return ownerLifecycleBlocked("release", "release_receipt")
	}
	var positionAccount, positionMarket, positionSymbol, positionState, positionQuantity string
	var positionGeneration, positionVersion int64
	err = tx.QueryRowContext(ctx, `SELECT p.account_ref,p.market,p.symbol,p.instance_seq,p.state,p.quantity,v.version
		FROM positions p JOIN position_projection_versions v ON v.position_id=p.id AND v.account_ref=p.account_ref
		AND v.market=p.market AND v.symbol=p.symbol AND v.generation=p.instance_seq AND v.state=p.state WHERE p.id=?`, seal.PositionID).
		Scan(&positionAccount, &positionMarket, &positionSymbol, &positionGeneration, &positionState, &positionQuantity, &positionVersion)
	zero, zeroErr := riskcalc.CompareDecimal(positionQuantity, "0")
	if err != nil || positionAccount != key.AccountID || normaliseMarket(positionMarket) != normaliseMarket(string(key.Market)) ||
		normaliseSymbol(positionSymbol) != key.Symbol || positionGeneration != actualGeneration || positionState != "CLOSED" ||
		zeroErr != nil || zero != 0 || positionVersion != seal.PositionVersion {
		return ownerLifecycleBlocked("release", "release_receipt")
	}
	authority, err := loadRiskBucketBrokerZeroAuthority(ctx, tx, key, actualGeneration, seal.PositionID, seal.PositionVersion, seal.ObservationID)
	if err != nil || authority.RecordDigest != seal.ObservationDigest || authority.Record.ReconcileStateID != seal.ReconcileStateID {
		return ownerLifecycleBlocked("release", "release_receipt")
	}
	sequence, stateDigest, err := loadRiskBucketPredecessorSeal(ctx, tx, key)
	if err != nil || sequence != seal.PredecessorEventSequence || stateDigest != seal.PredecessorStateDigest {
		return ownerLifecycleBlocked("release", "release_receipt")
	}
	_, recomputedStateDigest, err := loadReleasedRiskBucketState(ctx, tx, key)
	if err != nil || recomputedStateDigest != seal.PredecessorStateDigest {
		return ownerLifecycleBlocked("release", "release_receipt")
	}
	canonicalPayload, err := json.Marshal(seal)
	if err != nil {
		return err
	}
	digest := sha256.Sum256(canonicalPayload)
	canonicalDigest := hex.EncodeToString(digest[:])
	if releasePayload != string(canonicalPayload) || releaseDigest != canonicalDigest || eventID != "owner-release-"+canonicalDigest[:24] {
		return ownerLifecycleBlocked("release", "release_receipt")
	}
	var eventAccount, eventMarket, eventSymbol, eventProspective, eventType, eventDigest, eventPayload, eventAt string
	err = tx.QueryRowContext(ctx, `SELECT account_ref,market,symbol,prospective_generation,event_type,event_digest,payload,created_at
		FROM risk_bucket_events WHERE event_id=?`, eventID).
		Scan(&eventAccount, &eventMarket, &eventSymbol, &eventProspective, &eventType, &eventDigest, &eventPayload, &eventAt)
	if err != nil || eventAccount != key.AccountID || eventMarket != string(key.Market) || eventSymbol != key.Symbol ||
		eventProspective != key.ProspectiveGeneration || eventType != "OWNER_RELEASED" || eventDigest != canonicalDigest ||
		eventPayload != string(canonicalPayload) || eventAt != releasedAt {
		return ownerLifecycleBlocked("release", "release_receipt")
	}
	return nil
}

// releaseRiskBucketOwner derives every release predicate in one transaction.
// It never deletes a campaign claim, cancels a protection order, or accepts a
// caller attestation as authority.
func (j *Journal) releaseRiskBucketOwner(ctx context.Context, key riskbucket.OwnerKey) (RiskBucketOwnerReleaseResult, error) {
	if err := validateRiskBucketOwnerKey(key, "release"); err != nil {
		return RiskBucketOwnerReleaseResult{}, err
	}
	tx, err := j.db.BeginTx(ctx, nil)
	if err != nil {
		return RiskBucketOwnerReleaseResult{}, err
	}
	defer tx.Rollback()
	var lane, campaign, acquired string
	var actual, released sql.NullString
	err = tx.QueryRowContext(ctx, `SELECT lane_id,campaign_id,actual_generation,acquired_at,released_at FROM risk_bucket_owners
		WHERE account_ref=? AND market=? AND symbol=? AND prospective_generation=?`, key.AccountID,
		string(key.Market), key.Symbol, key.ProspectiveGeneration).Scan(&lane, &campaign, &actual, &acquired, &released)
	if errors.Is(err, sql.ErrNoRows) {
		return RiskBucketOwnerReleaseResult{}, ownerLifecycleBlocked("release", "owner")
	}
	if err != nil {
		return RiskBucketOwnerReleaseResult{}, err
	}
	if released.Valid {
		if !actual.Valid || actual.String == "" {
			return RiskBucketOwnerReleaseResult{}, ownerLifecycleBlocked("release", "release_receipt")
		}
		if err := validateRiskBucketOwnerReleaseReceipt(ctx, tx, key, lane, campaign, actual.String, released.String); err != nil {
			return RiskBucketOwnerReleaseResult{}, err
		}
		if err := tx.Commit(); err != nil {
			return RiskBucketOwnerReleaseResult{}, err
		}
		return RiskBucketOwnerReleaseResult{AlreadyReleased: true}, nil
	}
	if err := verifyRiskBucketStateDigest(ctx, tx, key); err != nil {
		return RiskBucketOwnerReleaseResult{}, err
	}
	actualGeneration, parseErr := strconv.ParseInt(actual.String, 10, 64)
	if !actual.Valid || parseErr != nil || actualGeneration <= 0 {
		return RiskBucketOwnerReleaseResult{}, ownerLifecycleBlocked("release", "actual_generation")
	}
	var campaignAccount, campaignMarket, campaignSymbol, campaignLane, decisionID, token, campaignState, campaignUpdated string
	var campaignActual sql.NullInt64
	var campaignVersion int64
	if err := tx.QueryRowContext(ctx, `SELECT account_ref,market,symbol,lane_id,decision_id,prospective_token,
		actual_position_generation,state,version,updated_at FROM position_campaigns WHERE id=?`, campaign).
		Scan(&campaignAccount, &campaignMarket, &campaignSymbol, &campaignLane, &decisionID, &token,
			&campaignActual, &campaignState, &campaignVersion, &campaignUpdated); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return RiskBucketOwnerReleaseResult{}, ownerLifecycleBlocked("release", "campaign")
		}
		return RiskBucketOwnerReleaseResult{}, err
	}
	if campaignAccount != key.AccountID || normaliseMarket(campaignMarket) != normaliseMarket(string(key.Market)) ||
		normaliseSymbol(campaignSymbol) != key.Symbol || campaignLane != lane || token != key.ProspectiveGeneration ||
		!campaignActual.Valid || campaignActual.Int64 != actualGeneration {
		return RiskBucketOwnerReleaseResult{}, ownerLifecycleBlocked("release", "campaign_identity")
	}
	if campaignState != "CLOSED" {
		return RiskBucketOwnerReleaseResult{}, ownerLifecycleBlocked("release", "campaign_closed")
	}
	var activeClaims int
	if err := tx.QueryRowContext(ctx, `SELECT count(*) FROM position_campaign_claims WHERE campaign_id=? OR
		(account_ref=? AND market=? AND symbol=?)`, campaign, key.AccountID, campaignMarket, campaignSymbol).Scan(&activeClaims); err != nil {
		return RiskBucketOwnerReleaseResult{}, err
	}
	if activeClaims != 0 {
		return RiskBucketOwnerReleaseResult{}, ownerLifecycleBlocked("release", "sell_claim")
	}

	var positionID, positionState, positionQuantity, closedAt string
	var entryDecision sql.NullString
	var positionVersion int64
	var latestGeneration int
	if err := tx.QueryRowContext(ctx, `SELECT p.id,p.state,p.quantity,COALESCE(p.closed_at,''),p.entry_decision_id,v.version FROM positions p
		JOIN position_projection_versions v ON v.position_id=p.id AND v.account_ref=p.account_ref
		 AND v.market=p.market AND v.symbol=p.symbol AND v.generation=p.instance_seq AND v.state=p.state
		WHERE p.account_ref=? AND p.market=? AND p.symbol=? AND p.instance_seq=?`, key.AccountID,
		campaignMarket, campaignSymbol, actualGeneration).Scan(&positionID, &positionState, &positionQuantity, &closedAt, &entryDecision, &positionVersion); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return RiskBucketOwnerReleaseResult{}, ownerLifecycleBlocked("release", "position_generation")
		}
		return RiskBucketOwnerReleaseResult{}, err
	}
	if !entryDecision.Valid || positionVersion <= 0 {
		return RiskBucketOwnerReleaseResult{}, ownerLifecycleBlocked("release", "position_generation")
	}
	if err := tx.QueryRowContext(ctx, `SELECT COALESCE(MAX(instance_seq),0) FROM positions WHERE account_ref=? AND market=? AND symbol=?`, key.AccountID, campaignMarket, campaignSymbol).Scan(&latestGeneration); err != nil {
		return RiskBucketOwnerReleaseResult{}, err
	}
	if int64(latestGeneration) != actualGeneration {
		return RiskBucketOwnerReleaseResult{}, ownerLifecycleBlocked("release", "latest_position_generation")
	}
	if positionState != "CLOSED" || closedAt == "" {
		return RiskBucketOwnerReleaseResult{}, ownerLifecycleBlocked("release", "position_closed")
	}
	zero, err := riskcalc.CompareDecimal(positionQuantity, "0")
	if err != nil || zero != 0 {
		return RiskBucketOwnerReleaseResult{}, ownerLifecycleBlocked("release", "position_quantity")
	}
	var lineageCount int
	if err := tx.QueryRowContext(ctx, `SELECT count(*) FROM risk_bucket_final_decisions d JOIN risk_reservations r
		ON r.id=d.existing_reservation_id WHERE d.account_ref=? AND d.market=? AND d.symbol=?
		AND d.owner_prospective_generation=? AND d.owner_lane_id=? AND d.owner_campaign_id=? AND r.decision_id=?`,
		key.AccountID, string(key.Market), key.Symbol, key.ProspectiveGeneration, lane, campaign, entryDecision.String).Scan(&lineageCount); err != nil {
		return RiskBucketOwnerReleaseResult{}, err
	}
	if lineageCount == 0 || decisionID != entryDecision.String {
		return RiskBucketOwnerReleaseResult{}, ownerLifecycleBlocked("release", "entry_decision_lineage")
	}

	checks := []struct {
		field, query string
		args         []any
	}{
		{"legacy_held", `SELECT count(*) FROM risk_reservations r JOIN risk_bucket_final_decisions d ON d.existing_reservation_id=r.id WHERE d.account_ref=? AND d.market=? AND d.symbol=? AND d.owner_prospective_generation=? AND r.state='HELD'`, []any{key.AccountID, string(key.Market), key.Symbol, key.ProspectiveGeneration}},
		{"bucket_held", `SELECT count(*) FROM risk_bucket_reservations WHERE account_ref=? AND market=? AND symbol=? AND owner_prospective_generation=? AND (held_minor<>'0' OR state='HELD')`, []any{key.AccountID, string(key.Market), key.Symbol, key.ProspectiveGeneration}},
		{"pending_entry", `SELECT count(*) FROM risk_bucket_orders o JOIN risk_bucket_final_decisions d ON d.decision_id=o.decision_id WHERE d.account_ref=? AND d.market=? AND d.symbol=? AND d.owner_prospective_generation=? AND o.state='ACTIVE'`, []any{key.AccountID, string(key.Market), key.Symbol, key.ProspectiveGeneration}},
		{"owner_latch", `SELECT count(*) FROM risk_bucket_owners WHERE account_ref=? AND market=? AND symbol=? AND prospective_generation=? AND (risk_overage_latched=1 OR unknown_actual_latched=1)`, []any{key.AccountID, string(key.Market), key.Symbol, key.ProspectiveGeneration}},
		{"scope_latch", `SELECT count(*) FROM risk_bucket_scope_latches WHERE account_ref=? AND market=? AND symbol=? AND prospective_generation=?`, []any{key.AccountID, string(key.Market), key.Symbol, key.ProspectiveGeneration}},
		{"unresolved_fill", `SELECT count(*) FROM risk_bucket_fills f JOIN risk_bucket_orders o ON o.order_key=f.order_key JOIN risk_bucket_final_decisions d ON d.decision_id=o.decision_id WHERE d.account_ref=? AND d.market=? AND d.symbol=? AND d.owner_prospective_generation=? AND (f.actual_known=0 OR NOT EXISTS(SELECT 1 FROM risk_bucket_fill_actual_evidence a WHERE a.fill_id=f.fill_id))`, []any{key.AccountID, string(key.Market), key.Symbol, key.ProspectiveGeneration}},
		{"unresolved_fill", `SELECT count(*) FROM risk_bucket_events WHERE account_ref=? AND market=? AND symbol=? AND prospective_generation=? AND event_type='FILL_UNACCOUNTED'`, []any{key.AccountID, string(key.Market), key.Symbol, key.ProspectiveGeneration}},
		{"protection_saga", `SELECT count(*) FROM protection_sagas WHERE account_ref=? AND market=? AND symbol=? AND generation=? AND state NOT IN ('TRIGGERED','CLOSED')`, []any{key.AccountID, string(key.Market), key.Symbol, actualGeneration}},
		{"protection_order", `SELECT count(*) FROM protection_mutation_attempts a JOIN protection_sagas s ON s.saga_id=a.saga_id WHERE s.account_ref=? AND s.market=? AND s.symbol=? AND a.generation=? AND a.state<>'CLOSED'`, []any{key.AccountID, string(key.Market), key.Symbol, actualGeneration}},
		{"sell_claim", `SELECT count(*) FROM campaign_order_watermarks w JOIN position_campaigns c ON c.id=w.campaign_id WHERE c.id=? AND w.side='SELL' AND (w.terminal=0 OR w.lineage_ambiguous=1)`, []any{campaign}},
	}
	for _, check := range checks {
		var count int
		if err := tx.QueryRowContext(ctx, check.query, check.args...).Scan(&count); err != nil {
			return RiskBucketOwnerReleaseResult{}, err
		}
		if count != 0 {
			return RiskBucketOwnerReleaseResult{}, ownerLifecycleBlocked("release", check.field)
		}
	}
	if dirty, err := unresolvedScopedMutation(ctx, tx, key.AccountID, campaignMarket, campaignSymbol, "BUY"); err != nil {
		return RiskBucketOwnerReleaseResult{}, err
	} else if dirty {
		return RiskBucketOwnerReleaseResult{}, ownerLifecycleBlocked("release", "pending_entry")
	}
	if dirty, err := unresolvedScopedMutation(ctx, tx, key.AccountID, campaignMarket, campaignSymbol, "SELL"); err != nil {
		return RiskBucketOwnerReleaseResult{}, err
	} else if dirty {
		return RiskBucketOwnerReleaseResult{}, ownerLifecycleBlocked("release", "sell_mutation")
	}

	var activeReconcile int
	if err := tx.QueryRowContext(ctx, `SELECT count(*) FROM reconcile_states WHERE account_ref=? AND released_at IS NULL
		AND (symbol IS NULL OR symbol=?) AND (scope_market IS NULL OR scope_market=?)`, key.AccountID, key.Symbol, string(key.Market)).Scan(&activeReconcile); err != nil {
		return RiskBucketOwnerReleaseResult{}, err
	}
	if activeReconcile != 0 {
		return RiskBucketOwnerReleaseResult{}, ownerLifecycleBlocked("release", "broker_reconciled")
	}
	authority, err := loadRiskBucketBrokerZeroAuthority(ctx, tx, key, actualGeneration, positionID, positionVersion, "")
	if err != nil {
		return RiskBucketOwnerReleaseResult{}, err
	}
	predecessors := []string{acquired, closedAt, campaignUpdated}
	for _, latest := range []struct {
		query string
		args  []any
	}{
		{`SELECT COALESCE(MAX(updated_at),'') FROM risk_bucket_reservations WHERE account_ref=? AND market=? AND symbol=? AND owner_prospective_generation=?`, []any{key.AccountID, string(key.Market), key.Symbol, key.ProspectiveGeneration}},
		{`SELECT COALESCE(MAX(updated_at),'') FROM risk_bucket_orders o JOIN risk_bucket_final_decisions d ON d.decision_id=o.decision_id WHERE d.account_ref=? AND d.market=? AND d.symbol=? AND d.owner_prospective_generation=?`, []any{key.AccountID, string(key.Market), key.Symbol, key.ProspectiveGeneration}},
		{`SELECT COALESCE(MAX(f.observed_at),'') FROM risk_bucket_fills f JOIN risk_bucket_orders o ON o.order_key=f.order_key JOIN risk_bucket_final_decisions d ON d.decision_id=o.decision_id WHERE d.account_ref=? AND d.market=? AND d.symbol=? AND d.owner_prospective_generation=?`, []any{key.AccountID, string(key.Market), key.Symbol, key.ProspectiveGeneration}},
		{`SELECT COALESCE(MAX(updated_at),'') FROM protection_sagas WHERE account_ref=? AND market=? AND symbol=? AND generation=?`, []any{key.AccountID, string(key.Market), key.Symbol, actualGeneration}},
		{`SELECT COALESCE(MAX(a.updated_at),'') FROM protection_mutation_attempts a JOIN protection_sagas s ON s.saga_id=a.saga_id WHERE s.account_ref=? AND s.market=? AND s.symbol=? AND a.generation=?`, []any{key.AccountID, string(key.Market), key.Symbol, actualGeneration}},
		{`SELECT COALESCE(MAX(i.created_at),'') FROM intents i WHERE i.account_ref=? AND lower(i.market)=? AND upper(i.symbol)=? AND upper(i.side) IN ('BUY','SELL')`, []any{key.AccountID, normaliseMarket(campaignMarket), key.Symbol}},
		{`SELECT COALESCE(MAX(COALESCE(a.settled_at,a.dispatch_started_at,a.recorded_at)),'') FROM mutation_attempts a JOIN intents i ON i.id=a.intent_id WHERE i.account_ref=? AND lower(i.market)=? AND upper(i.symbol)=? AND upper(i.side) IN ('BUY','SELL')`, []any{key.AccountID, normaliseMarket(campaignMarket), key.Symbol}},
		{`SELECT COALESCE(MAX(committed_at),'') FROM scoped_fill_snapshots WHERE account_ref=? AND lower(market)=? AND upper(symbol)=? AND upper(side) IN ('BUY','SELL')`, []any{key.AccountID, normaliseMarket(campaignMarket), key.Symbol}},
		{`SELECT COALESCE(MAX(updated_at),'') FROM campaign_order_watermarks WHERE campaign_id=?`, []any{campaign}},
		{`SELECT COALESCE(MAX(CASE WHEN broker_as_of>created_at THEN broker_as_of ELSE created_at END),'') FROM position_adjustments WHERE position_id IN (SELECT id FROM positions WHERE account_ref=? AND market=? AND symbol=? AND instance_seq=?)`, []any{key.AccountID, campaignMarket, campaignSymbol, actualGeneration}},
	} {
		var value string
		if err := tx.QueryRowContext(ctx, latest.query, latest.args...).Scan(&value); err != nil {
			return RiskBucketOwnerReleaseResult{}, err
		}
		predecessors = append(predecessors, value)
	}
	if fresh, err := timestampStrictlyAfter(authority.Record.BrokerAsOf, predecessors...); err != nil || !fresh {
		return RiskBucketOwnerReleaseResult{}, ownerLifecycleBlocked("release", "broker_reconciliation_stale")
	}
	releaseAt, err := journalTimeStrictlyAfter(j.nowString(), append(predecessors, authority.Record.BrokerAsOf, authority.ReconcileReleasedAt)...)
	if err != nil {
		return RiskBucketOwnerReleaseResult{}, err
	}
	predecessorSequence, predecessorDigest, err := loadRiskBucketPredecessorSeal(ctx, tx, key)
	if err != nil {
		return RiskBucketOwnerReleaseResult{}, err
	}
	seal := riskBucketOwnerReleaseSeal{
		AccountRef: key.AccountID, Market: string(key.Market), Symbol: key.Symbol,
		ProspectiveGeneration: key.ProspectiveGeneration, ActualGeneration: actual.String, LaneID: lane,
		CampaignID: campaign, CampaignVersion: campaignVersion, PositionID: positionID, PositionVersion: positionVersion,
		ReconcileStateID: authority.Record.ReconcileStateID, ObservationID: authority.Record.ObservationID,
		ObservationDigest: authority.RecordDigest, PredecessorEventSequence: predecessorSequence,
		PredecessorStateDigest: predecessorDigest, ReleasedAt: releaseAt,
	}
	payload, err := json.Marshal(seal)
	if err != nil {
		return RiskBucketOwnerReleaseResult{}, err
	}
	digest := sha256.Sum256(payload)
	digestText := hex.EncodeToString(digest[:])
	result, err := tx.ExecContext(ctx, `UPDATE risk_bucket_owners SET released_at=? WHERE account_ref=? AND market=? AND symbol=?
		AND prospective_generation=? AND released_at IS NULL`, releaseAt, key.AccountID, string(key.Market), key.Symbol, key.ProspectiveGeneration)
	if err != nil {
		return RiskBucketOwnerReleaseResult{}, err
	}
	if affected, err := result.RowsAffected(); err != nil || affected != 1 {
		return RiskBucketOwnerReleaseResult{}, ownerLifecycleBlocked("release", "owner_race")
	}
	eventID := "owner-release-" + digestText[:24]
	if _, err := tx.ExecContext(ctx, `INSERT INTO risk_bucket_events(event_id,account_ref,market,symbol,prospective_generation,event_type,event_digest,payload,created_at) VALUES(?,?,?,?,?,'OWNER_RELEASED',?,?,?)`, eventID, key.AccountID, string(key.Market), key.Symbol, key.ProspectiveGeneration, digestText, string(payload), releaseAt); err != nil {
		return RiskBucketOwnerReleaseResult{}, err
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO risk_bucket_owner_release_receipts(
		account_ref,market,symbol,prospective_generation,actual_generation,campaign_id,campaign_version,
		position_id,position_version,reconcile_state_id,observation_id,observation_digest,
		predecessor_event_sequence,predecessor_state_digest,release_event_id,release_digest,release_payload,released_at)
		VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`, key.AccountID, string(key.Market), key.Symbol,
		key.ProspectiveGeneration, actual.String, campaign, campaignVersion, positionID, positionVersion,
		authority.Record.ReconcileStateID, authority.Record.ObservationID, authority.RecordDigest,
		predecessorSequence, predecessorDigest, eventID, digestText, string(payload), releaseAt); err != nil {
		return RiskBucketOwnerReleaseResult{}, err
	}
	if err := tx.Commit(); err != nil {
		return RiskBucketOwnerReleaseResult{}, err
	}
	return RiskBucketOwnerReleaseResult{Released: true}, nil
}

func unresolvedScopedMutation(ctx context.Context, tx *sql.Tx, account, market, symbol, side string) (bool, error) {
	var dirty int
	err := tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM intents i WHERE i.account_ref=?
		AND lower(i.market)=? AND upper(i.symbol)=? AND upper(i.side)=? AND (
		NOT EXISTS(SELECT 1 FROM mutation_attempts a WHERE a.intent_id=i.id)
		OR EXISTS(SELECT 1 FROM mutation_attempts a WHERE a.intent_id=i.id AND (
			a.kind NOT IN ('PLACE','CANCEL','AMEND')
			OR a.state NOT IN ('NOT_DISPATCHED','FAILED_CONFIRMED','CONFIRMED')
			OR (a.state='CONFIRMED' AND a.kind IN ('PLACE','AMEND') AND (a.broker_order_id='' OR NOT EXISTS(
				SELECT 1 FROM scoped_fill_snapshots f WHERE f.account_ref=i.account_ref AND lower(f.market)=lower(i.market)
				AND f.trading_day=i.trading_day AND upper(f.symbol)=upper(i.symbol) AND upper(f.side)=upper(i.side)
				AND f.order_id=a.broker_order_id AND f.terminal=1 AND f.fail_closed=0)))))))`,
		account, normaliseMarket(market), normaliseSymbol(symbol), strings.ToUpper(side)).Scan(&dirty)
	return dirty != 0, err
}

func timestampStrictlyAfter(candidate string, predecessors ...string) (bool, error) {
	current, err := time.Parse(time.RFC3339Nano, candidate)
	if err != nil {
		return false, err
	}
	for _, text := range predecessors {
		if strings.TrimSpace(text) == "" {
			continue
		}
		predecessor, err := time.Parse(time.RFC3339Nano, text)
		if err != nil {
			return false, err
		}
		if !current.After(predecessor) {
			return false, nil
		}
	}
	return true, nil
}

func validateRiskBucketOwnerKey(key riskbucket.OwnerKey, operation string) error {
	if strings.TrimSpace(key.AccountID) == "" || normaliseSymbol(key.Symbol) != key.Symbol ||
		strings.TrimSpace(key.ProspectiveGeneration) == "" || (key.Market != riskbucket.MarketKR && key.Market != riskbucket.MarketUS) {
		return ownerLifecycleBlocked(operation, "owner_key")
	}
	return nil
}
