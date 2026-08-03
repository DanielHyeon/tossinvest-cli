package journal

// position_campaign.go owns schema v20 and the strategy-neutral persistence
// boundary for PositionCampaign. It deliberately has no broker client, order
// dispatcher, scheduler or runtime-toggle dependency. The only fill writer is
// ApplyPositionCampaignFill, which is designed to run in RecordFill's existing
// transaction after ProjectPosition and before ApplyExitFill.

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/JungHoonGhae/tossinvest-cli/internal/positioncampaign"
	"github.com/JungHoonGhae/tossinvest-cli/internal/riskcalc"
)

var (
	ErrGenerationConflict      = errors.New("journal: prospective position generation conflict")
	ErrCampaignVersionConflict = errors.New("journal: campaign version conflict")
	ErrCampaignCommandConflict = errors.New("journal: campaign command key conflict")
	ErrCampaignNotFound        = errors.New("journal: position campaign not found")
	ErrCampaignLegNotFound     = errors.New("journal: campaign leg not found")
)

const schemaV20 = `
CREATE TABLE position_campaigns (
	id                           TEXT PRIMARY KEY,
	account_ref                  TEXT NOT NULL,
	market                       TEXT NOT NULL,
	symbol                       TEXT NOT NULL,
	lane_id                      TEXT NOT NULL,
	lane_version                 TEXT NOT NULL,
	decision_id                  TEXT NOT NULL REFERENCES decisions(id),
	evidence_digest              TEXT NOT NULL,
	expected_position_generation INTEGER NOT NULL CHECK(expected_position_generation >= 0),
	expected_position_version    INTEGER NOT NULL CHECK(expected_position_version >= 0),
	prospective_token            TEXT NOT NULL UNIQUE,
	actual_position_generation   INTEGER CHECK(actual_position_generation > 0),
	state                        TEXT NOT NULL CHECK(state IN ('PLANNED','ACTIVE','EXITING','CLOSED','RECONCILE')),
	version                      INTEGER NOT NULL CHECK(version > 0),
	entry_blocked                INTEGER NOT NULL DEFAULT 0 CHECK(entry_blocked IN (0,1)),
	effective_stop               TEXT,
	stop_source                  TEXT,
	stop_policy                  TEXT,
	stop_observed_at             TEXT,
	stop_candidate               TEXT,
	stop_candidate_valid         INTEGER CHECK(stop_candidate_valid IN (0,1)),
	stop_candidate_source        TEXT,
	stop_candidate_policy        TEXT,
	stop_candidate_observed_at   TEXT,
	stop_selected_from           TEXT CHECK(stop_selected_from IN ('SAVED','CANDIDATE')),
	created_at                   TEXT NOT NULL,
	updated_at                   TEXT NOT NULL
) STRICT;

CREATE UNIQUE INDEX idx_position_campaign_active_scope
	ON position_campaigns(account_ref,market,symbol)
	WHERE state IN ('PLANNED','ACTIVE','EXITING','RECONCILE');
CREATE INDEX idx_position_campaign_scope
	ON position_campaigns(account_ref,market,symbol,state);

-- A Position row predates versioned campaign admission.  This additive
-- companion is populated only by post-v20 Position writes: legacy rows are
-- deliberately left unknown and therefore fail campaign CAS closed.
CREATE TABLE position_projection_versions (
	position_id  TEXT PRIMARY KEY REFERENCES positions(id),
	account_ref  TEXT NOT NULL,
	market       TEXT NOT NULL,
	symbol       TEXT NOT NULL,
	generation   INTEGER NOT NULL CHECK(generation > 0),
	state        TEXT NOT NULL CHECK(state IN ('FLAT','OPENING','OPEN','SCALING','CLOSING','CLOSED')),
	version      INTEGER NOT NULL CHECK(version > 0),
	updated_at   TEXT NOT NULL,
	UNIQUE(account_ref,market,symbol,generation)
) STRICT;

CREATE TRIGGER position_projection_version_after_insert
AFTER INSERT ON positions BEGIN
	INSERT OR REPLACE INTO position_projection_versions
		(position_id,account_ref,market,symbol,generation,state,version,updated_at)
	VALUES (NEW.id,NEW.account_ref,NEW.market,NEW.symbol,NEW.instance_seq,NEW.state,
	        coalesce((SELECT version+1 FROM position_projection_versions WHERE position_id=NEW.id),1),
	        strftime('%Y-%m-%dT%H:%M:%fZ','now'));
END;

CREATE TRIGGER position_projection_version_after_update
AFTER UPDATE ON positions BEGIN
	INSERT OR REPLACE INTO position_projection_versions
		(position_id,account_ref,market,symbol,generation,state,version,updated_at)
	VALUES (NEW.id,NEW.account_ref,NEW.market,NEW.symbol,NEW.instance_seq,NEW.state,
	        coalesce((SELECT version+1 FROM position_projection_versions WHERE position_id=NEW.id),1),
	        strftime('%Y-%m-%dT%H:%M:%fZ','now'));
END;

CREATE TABLE campaign_legs (
	campaign_id       TEXT NOT NULL REFERENCES position_campaigns(id),
	sequence          INTEGER NOT NULL CHECK(sequence > 0),
	plan_id           TEXT NOT NULL,
	intent_id         TEXT,
	requested_quantity TEXT NOT NULL,
	filled_quantity   TEXT NOT NULL DEFAULT '0',
	residual_quantity TEXT NOT NULL,
	state              TEXT NOT NULL CHECK(state IN ('PLANNED','SUBMITTED','PARTIAL','FILLED','CANCELLED','RECONCILE')),
	version            INTEGER NOT NULL CHECK(version > 0),
	created_at         TEXT NOT NULL,
	updated_at         TEXT NOT NULL,
	PRIMARY KEY(campaign_id,sequence),
	UNIQUE(campaign_id,plan_id)
) STRICT;

CREATE TABLE campaign_order_watermarks (
	campaign_id        TEXT NOT NULL,
	leg_sequence       INTEGER NOT NULL,
	order_id           TEXT NOT NULL,
	account_ref        TEXT NOT NULL,
	market             TEXT NOT NULL,
	trading_day        TEXT NOT NULL,
	symbol             TEXT NOT NULL,
	side               TEXT NOT NULL CHECK(side IN ('BUY','SELL')),
	decision_id        TEXT NOT NULL REFERENCES decisions(id),
	intent_id          TEXT NOT NULL,
	attempt_id         TEXT NOT NULL,
	predecessor_order_id TEXT,
	carry_baseline     TEXT NOT NULL DEFAULT '0',
	requested_cap      TEXT NOT NULL,
	cumulative_filled  TEXT NOT NULL DEFAULT '0',
	remaining_quantity TEXT NOT NULL,
	terminal           INTEGER NOT NULL DEFAULT 0 CHECK(terminal IN (0,1)),
	lineage_ambiguous  INTEGER NOT NULL DEFAULT 0 CHECK(lineage_ambiguous IN (0,1)),
	last_observation_id TEXT,
	created_at         TEXT NOT NULL,
	updated_at         TEXT NOT NULL,
	PRIMARY KEY(campaign_id,leg_sequence,order_id),
	FOREIGN KEY(campaign_id,leg_sequence) REFERENCES campaign_legs(campaign_id,sequence),
	FOREIGN KEY(intent_id) REFERENCES intents(id),
	FOREIGN KEY(attempt_id) REFERENCES mutation_attempts(id),
	FOREIGN KEY(campaign_id,leg_sequence,predecessor_order_id)
		REFERENCES campaign_order_watermarks(campaign_id,leg_sequence,order_id)
) STRICT;

CREATE INDEX idx_campaign_order_identity
	ON campaign_order_watermarks(order_id,campaign_id,leg_sequence);
CREATE UNIQUE INDEX idx_campaign_order_scope_identity
	ON campaign_order_watermarks(account_ref,market,trading_day,symbol,side,order_id);
CREATE UNIQUE INDEX idx_campaign_order_one_successor
	ON campaign_order_watermarks(campaign_id,leg_sequence,predecessor_order_id)
	WHERE predecessor_order_id IS NOT NULL;

CREATE TABLE campaign_commands (
	campaign_id    TEXT NOT NULL REFERENCES position_campaigns(id),
	command_kind  TEXT NOT NULL,
	command_key   TEXT NOT NULL,
	request_digest TEXT NOT NULL,
	result_version INTEGER NOT NULL,
	result_sequence INTEGER,
	result_error   TEXT CHECK(result_error IS NULL OR result_error IN ('INVALID_IDENTITY')),
	recorded_at   TEXT NOT NULL,
	PRIMARY KEY(campaign_id,command_kind,command_key)
) STRICT;

CREATE TRIGGER campaign_commands_no_update
BEFORE UPDATE ON campaign_commands BEGIN
	SELECT RAISE(ABORT,'campaign_commands are append-only');
END;
CREATE TRIGGER campaign_commands_no_delete
BEFORE DELETE ON campaign_commands BEGIN
	SELECT RAISE(ABORT,'campaign_commands are append-only');
END;

CREATE TABLE campaign_events (
	campaign_id        TEXT NOT NULL REFERENCES position_campaigns(id),
	sequence           INTEGER NOT NULL CHECK(sequence > 0),
	campaign_version   INTEGER NOT NULL CHECK(campaign_version > 0),
	leg_sequence       INTEGER,
	order_id           TEXT,
	event_kind         TEXT NOT NULL,
	command_kind       TEXT NOT NULL,
	command_key        TEXT NOT NULL,
	request_digest     TEXT NOT NULL,
	campaign_state     TEXT NOT NULL CHECK(campaign_state IN ('PLANNED','ACTIVE','EXITING','CLOSED','RECONCILE')),
	leg_state          TEXT CHECK(leg_state IN ('PLANNED','SUBMITTED','PARTIAL','FILLED','CANCELLED','RECONCILE')),
	leg_requested_quantity TEXT,
	leg_filled_quantity TEXT,
	leg_residual_quantity TEXT,
	position_generation INTEGER,
	delta_quantity     TEXT,
	cumulative_quantity TEXT,
	prospective_token  TEXT,
	expected_position_generation INTEGER,
	plan_id            TEXT,
	intent_id          TEXT,
	attempt_id         TEXT,
	predecessor_order_id TEXT,
	carry_baseline     TEXT,
	requested_cap      TEXT,
	order_remaining_quantity TEXT,
	order_terminal     INTEGER CHECK(order_terminal IN (0,1)),
	order_lineage_ambiguous INTEGER CHECK(order_lineage_ambiguous IN (0,1)),
	effective_stop     TEXT,
	stop_source        TEXT,
	stop_policy        TEXT,
	stop_observed_at   TEXT,
	entry_blocked      INTEGER NOT NULL CHECK(entry_blocked IN (0,1)),
	projection_digest  TEXT NOT NULL,
	recorded_at        TEXT NOT NULL,
	PRIMARY KEY(campaign_id,sequence),
	UNIQUE(campaign_id,command_kind,command_key)
) STRICT;

CREATE TRIGGER campaign_events_no_update
BEFORE UPDATE ON campaign_events BEGIN
	SELECT RAISE(ABORT,'campaign_events are append-only');
END;
CREATE TRIGGER campaign_events_no_delete
BEFORE DELETE ON campaign_events BEGIN
	SELECT RAISE(ABORT,'campaign_events are append-only');
END;

CREATE TABLE position_campaign_claims (
	account_ref       TEXT NOT NULL,
	market            TEXT NOT NULL,
	symbol            TEXT NOT NULL,
	position_generation INTEGER NOT NULL CHECK(position_generation >= 0),
	position_version  INTEGER NOT NULL CHECK(position_version >= 0),
	version           INTEGER NOT NULL CHECK(version > 0),
	prospective_token TEXT NOT NULL UNIQUE,
	campaign_id       TEXT NOT NULL UNIQUE REFERENCES position_campaigns(id),
	actual_position_generation INTEGER CHECK(actual_position_generation > 0),
	created_at        TEXT NOT NULL,
	updated_at        TEXT NOT NULL,
	PRIMARY KEY(account_ref,market,symbol)
) STRICT;
`

type CreatePositionCampaignRequest struct {
	ID                         string
	AccountRef                 string
	Market                     string
	Symbol                     string
	LaneID                     string
	LaneVersion                string
	DecisionID                 string
	EvidenceDigest             string
	ExpectedPositionGeneration int64
	ExpectedPositionVersion    int64
	ProspectiveToken           string
	CommandKey                 string
}

type PositionCampaignRecord struct {
	ID                         string
	AccountRef                 string
	Market                     string
	Symbol                     string
	LaneID                     string
	LaneVersion                string
	DecisionID                 string
	EvidenceDigest             string
	ExpectedPositionGeneration int64
	ExpectedPositionVersion    int64
	ProspectiveToken           string
	ActualPositionGeneration   int64
	State                      positioncampaign.CampaignState
	Version                    int64
	EntryBlocked               bool
	EffectiveStop              string
	StopSource                 string
	StopPolicy                 string
	StopObservedAt             string
	StopCandidate              string
	StopCandidateValid         bool
	StopCandidateSource        string
	StopCandidatePolicy        string
	StopCandidateObservedAt    string
	StopSelectedFrom           positioncampaign.StopSelection
	CreatedAt                  string
	UpdatedAt                  string
}

type PositionCampaignLineageStatus string

const (
	PositionCampaignLineageKnown         PositionCampaignLineageStatus = "KNOWN"
	PositionCampaignLineageLegacyUnknown PositionCampaignLineageStatus = "LEGACY_UNKNOWN"
	PositionCampaignLineageNone          PositionCampaignLineageStatus = "NONE"
)

// PositionCampaignLineageRead explicitly distinguishes a pre-v20 Position
// whose campaign ancestry is unknowable from a post-v20 Position with no
// associated campaign. It never synthesizes campaign identity.
type PositionCampaignLineageRead struct {
	PositionID         string
	AccountRef         string
	Market             string
	Symbol             string
	PositionGeneration int64
	PositionVersion    int64
	CampaignID         string
	Status             PositionCampaignLineageStatus
}

type PlanCampaignLegRequest struct {
	CampaignID        string
	ExpectedVersion   int64
	CommandKey        string
	Sequence          int64
	PlanID            string
	RequestedQuantity string
}

type CampaignLegRecord struct {
	CampaignID        string
	Sequence          int64
	PlanID            string
	IntentID          string
	RequestedQuantity string
	FilledQuantity    string
	ResidualQuantity  string
	State             positioncampaign.LegState
	Version           int64
	CampaignVersion   int64
	CreatedAt         string
	UpdatedAt         string
}

type LinkCampaignOrderRequest struct {
	CampaignID         string
	LegSequence        int64
	ExpectedVersion    int64
	CommandKey         string
	OrderID            string
	IntentID           string
	AttemptID          string
	PredecessorOrderID string
	RequestedCap       string
	LineageAmbiguous   bool
}

type CampaignOrderRecord struct {
	CampaignID         string
	LegSequence        int64
	OrderID            string
	IntentID           string
	AttemptID          string
	PredecessorOrderID string
	CarryBaseline      string
	RequestedCap       string
	CumulativeFilled   string
	RemainingQuantity  string
	Terminal           bool
	LineageAmbiguous   bool
	LastObservationID  string
	CampaignVersion    int64
	CreatedAt          string
	UpdatedAt          string
}

type UpdateCampaignStopRequest struct {
	CampaignID      string
	ExpectedVersion int64
	CommandKey      string
	Candidate       positioncampaign.StopCandidate
}

type CancelCampaignRequest struct {
	CampaignID      string
	ExpectedVersion int64
	CommandKey      string
	Structural      bool
	Detail          string
}

func (j *Journal) CreatePositionCampaign(ctx context.Context, req CreatePositionCampaignRequest) (PositionCampaignRecord, error) {
	req.AccountRef = strings.TrimSpace(req.AccountRef)
	req.Market = normaliseMarket(req.Market)
	req.Symbol = normaliseSymbol(req.Symbol)
	for field, value := range map[string]string{
		"id": req.ID, "account": req.AccountRef, "market": req.Market, "symbol": req.Symbol,
		"lane id": req.LaneID, "lane version": req.LaneVersion, "decision id": req.DecisionID,
		"evidence digest": req.EvidenceDigest, "prospective token": req.ProspectiveToken,
		"command key": req.CommandKey,
	} {
		if strings.TrimSpace(value) == "" {
			return PositionCampaignRecord{}, fmt.Errorf("%w: campaign %s is empty", ErrInvalidRequest, field)
		}
	}
	if req.ExpectedPositionGeneration < 0 || req.ExpectedPositionVersion < 0 {
		return PositionCampaignRecord{}, fmt.Errorf("%w: negative expected position generation/version", ErrInvalidRequest)
	}
	if err := (positioncampaign.PositionCampaign{
		ID: req.ID, AccountRef: req.AccountRef, Market: req.Market, Symbol: req.Symbol,
		LaneID: req.LaneID, LaneVersion: req.LaneVersion, DecisionID: req.DecisionID,
		EvidenceDigest: req.EvidenceDigest, ProspectiveToken: req.ProspectiveToken,
		ExpectedPositionGeneration: req.ExpectedPositionGeneration,
	}).Validate(); err != nil {
		return PositionCampaignRecord{}, err
	}
	digest := digestParts(req.ID, req.AccountRef, req.Market, req.Symbol, req.LaneID,
		req.LaneVersion, req.DecisionID, req.EvidenceDigest,
		strconv.FormatInt(req.ExpectedPositionGeneration, 10), strconv.FormatInt(req.ExpectedPositionVersion, 10),
		req.ProspectiveToken)
	now := j.nowString()
	tx, err := j.db.BeginTx(ctx, nil)
	if err != nil {
		return PositionCampaignRecord{}, fmt.Errorf("journal: starting campaign creation: %w", err)
	}
	defer tx.Rollback()
	if existing, found, err := campaignCommandResult(ctx, tx, req.ID, "CREATE", req.CommandKey, digest); err != nil {
		return PositionCampaignRecord{}, err
	} else if found {
		if err := tx.Commit(); err != nil {
			return PositionCampaignRecord{}, err
		}
		return j.PositionCampaign(ctx, existing.campaignID)
	}

	var decisionAccount, safetyClass string
	if err := tx.QueryRowContext(ctx, `SELECT account_ref,safety_class FROM decisions WHERE id=?`,
		strings.TrimSpace(req.DecisionID)).Scan(&decisionAccount, &safetyClass); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return PositionCampaignRecord{}, fmt.Errorf("%w: campaign decision %s does not exist", positioncampaign.ErrInvalidIdentity, req.DecisionID)
		}
		return PositionCampaignRecord{}, fmt.Errorf("journal: reading campaign decision: %w", err)
	}
	if strings.TrimSpace(decisionAccount) != req.AccountRef || safetyClass != "EXPOSURE_RAISING" {
		return PositionCampaignRecord{}, fmt.Errorf("%w: decision %s is not exposure-raising authority for account %s",
			positioncampaign.ErrInvalidIdentity, req.DecisionID, req.AccountRef)
	}
	var lineageMarket, lineageSymbol, lineageLaneID, lineageLaneVersion, lineageEvidence string
	err = tx.QueryRowContext(ctx, `SELECT d.market,d.symbol,d.lane_id,d.lane_version,d.evidence_digest
		FROM strategy_attempt_lineage a
		JOIN strategy_decision_lineage d ON d.entry_decision_identity=a.entry_decision_identity
		WHERE a.risk_intent_id=?`, strings.TrimSpace(req.DecisionID)).
		Scan(&lineageMarket, &lineageSymbol, &lineageLaneID, &lineageLaneVersion, &lineageEvidence)
	if errors.Is(err, sql.ErrNoRows) {
		return PositionCampaignRecord{}, fmt.Errorf("%w: decision %s has no immutable strategy lineage",
			positioncampaign.ErrInvalidIdentity, req.DecisionID)
	}
	if err != nil {
		return PositionCampaignRecord{}, fmt.Errorf("journal: reading campaign strategy lineage: %w", err)
	}
	if normaliseMarket(lineageMarket) != req.Market || normaliseSymbol(lineageSymbol) != req.Symbol ||
		strings.TrimSpace(lineageLaneID) != strings.TrimSpace(req.LaneID) ||
		strings.TrimSpace(lineageLaneVersion) != strings.TrimSpace(req.LaneVersion) ||
		strings.TrimSpace(lineageEvidence) != strings.TrimSpace(req.EvidenceDigest) {
		return PositionCampaignRecord{}, fmt.Errorf("%w: decision %s strategy lineage does not match campaign identity",
			positioncampaign.ErrInvalidIdentity, req.DecisionID)
	}

	var currentGeneration, currentVersion int64
	var currentState string
	var positionVersion sql.NullInt64
	positionExists := true
	err = tx.QueryRowContext(ctx, `SELECT p.instance_seq,p.state,v.version
		FROM positions p LEFT JOIN position_projection_versions v ON v.position_id=p.id
		WHERE p.account_ref=? AND p.market=? AND p.symbol=?
		ORDER BY p.instance_seq DESC LIMIT 1`, req.AccountRef, req.Market, req.Symbol).
		Scan(&currentGeneration, &currentState, &positionVersion)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		positionExists = false
		currentGeneration, currentVersion, currentState = 0, 0, "FLAT"
	case err != nil:
		return PositionCampaignRecord{}, fmt.Errorf("journal: reading authoritative position generation: %w", err)
	case !positionVersion.Valid:
		return PositionCampaignRecord{}, fmt.Errorf("%w: position generation %d predates authoritative versioning",
			ErrGenerationConflict, currentGeneration)
	default:
		currentVersion = positionVersion.Int64
	}
	var activeCampaign string
	err = tx.QueryRowContext(ctx, `SELECT campaign_id FROM position_campaign_claims
		WHERE account_ref=? AND market=? AND symbol=?`, req.AccountRef, req.Market, req.Symbol).Scan(&activeCampaign)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return PositionCampaignRecord{}, fmt.Errorf("journal: reading prospective generation claim: %w", err)
	}
	activeClaim := err == nil
	if currentGeneration != req.ExpectedPositionGeneration || currentVersion != req.ExpectedPositionVersion ||
		(positionExists && currentState != "CLOSED") || activeClaim {
		return PositionCampaignRecord{}, fmt.Errorf("%w: expected generation/version %d/%d, actual %d/%d (campaign %s)",
			ErrGenerationConflict, req.ExpectedPositionGeneration, req.ExpectedPositionVersion,
			currentGeneration, currentVersion, activeCampaign)
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO position_campaigns
		(id,account_ref,market,symbol,lane_id,lane_version,decision_id,evidence_digest,
		 expected_position_generation,expected_position_version,prospective_token,state,version,
		 entry_blocked,created_at,updated_at)
		VALUES (?,?,?,?,?,?,?,?,?,?,?,'PLANNED',1,0,?,?)`,
		req.ID, req.AccountRef, req.Market, req.Symbol, strings.TrimSpace(req.LaneID),
		strings.TrimSpace(req.LaneVersion), strings.TrimSpace(req.DecisionID), strings.TrimSpace(req.EvidenceDigest),
		req.ExpectedPositionGeneration, req.ExpectedPositionVersion, strings.TrimSpace(req.ProspectiveToken), now, now); err != nil {
		if isConstraintError(err) {
			return PositionCampaignRecord{}, fmt.Errorf("%w: the campaign scope or token is already reserved: %v", ErrGenerationConflict, err)
		}
		return PositionCampaignRecord{}, fmt.Errorf("journal: inserting campaign: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO position_campaign_claims
		(account_ref,market,symbol,position_generation,position_version,version,prospective_token,campaign_id,created_at,updated_at)
		VALUES (?,?,?,?,?,1,?,?,?,?)`, req.AccountRef, req.Market, req.Symbol, currentGeneration, currentVersion,
		strings.TrimSpace(req.ProspectiveToken), req.ID, now, now); err != nil {
		return PositionCampaignRecord{}, fmt.Errorf("%w: reserving prospective generation: %v", ErrGenerationConflict, err)
	}
	if err := insertCampaignCommand(ctx, tx, req.ID, "CREATE", req.CommandKey, digest, 1, 0, now); err != nil {
		return PositionCampaignRecord{}, err
	}
	if err := insertCampaignEvent(ctx, tx, req.ID, 1, 1, 0, "", "CREATED", "CREATE", req.CommandKey,
		digest, positioncampaign.CampaignPlanned, "", 0, "", "", now); err != nil {
		return PositionCampaignRecord{}, err
	}
	if err := tx.Commit(); err != nil {
		return PositionCampaignRecord{}, fmt.Errorf("journal: committing campaign creation: %w", err)
	}
	return j.PositionCampaign(ctx, req.ID)
}

func (j *Journal) PlanCampaignLeg(ctx context.Context, req PlanCampaignLegRequest) (CampaignLegRecord, error) {
	requested, err := positiveCampaignQuantity(req.RequestedQuantity)
	if err != nil {
		return CampaignLegRecord{}, fmt.Errorf("%w: requested quantity: %v", ErrInvalidRequest, err)
	}
	if strings.TrimSpace(req.CampaignID) == "" || strings.TrimSpace(req.CommandKey) == "" ||
		strings.TrimSpace(req.PlanID) == "" || req.Sequence <= 0 {
		return CampaignLegRecord{}, fmt.Errorf("%w: campaign, command, plan and positive sequence are required", ErrInvalidRequest)
	}
	digest := digestParts(req.CampaignID, strconv.FormatInt(req.Sequence, 10), req.PlanID, requested)
	now := j.nowString()
	tx, err := j.db.BeginTx(ctx, nil)
	if err != nil {
		return CampaignLegRecord{}, err
	}
	defer tx.Rollback()
	if result, found, err := campaignCommandResult(ctx, tx, req.CampaignID, "PLAN_LEG", req.CommandKey, digest); err != nil {
		return CampaignLegRecord{}, err
	} else if found {
		leg, err := campaignLegInTx(ctx, tx, req.CampaignID, result.sequence)
		if err != nil {
			return CampaignLegRecord{}, err
		}
		if err := tx.Commit(); err != nil {
			return CampaignLegRecord{}, err
		}
		return leg, nil
	}
	state, version, blocked, err := campaignHeaderInTx(ctx, tx, req.CampaignID)
	if err != nil {
		return CampaignLegRecord{}, err
	}
	if version != req.ExpectedVersion {
		return CampaignLegRecord{}, versionConflict(req.CampaignID, req.ExpectedVersion, version)
	}
	if exposureBlocked, err := campaignExposureBlockedInTx(ctx, tx, req.CampaignID); err != nil {
		return CampaignLegRecord{}, err
	} else if exposureBlocked {
		return CampaignLegRecord{}, positioncampaign.ErrExposureBlocked
	}
	if blocked {
		return CampaignLegRecord{}, positioncampaign.ErrExposureBlocked
	}
	if _, err := positioncampaign.TransitionCampaign(state, positioncampaign.CampaignLegPlanned); err != nil {
		return CampaignLegRecord{}, err
	}
	var count int64
	if err := tx.QueryRowContext(ctx, `SELECT count(*) FROM campaign_legs WHERE campaign_id=?`, req.CampaignID).Scan(&count); err != nil {
		return CampaignLegRecord{}, err
	}
	if req.Sequence != count+1 {
		return CampaignLegRecord{}, fmt.Errorf("%w: leg sequence %d, want %d", ErrInvalidRequest, req.Sequence, count+1)
	}
	newVersion := version + 1
	if _, err := tx.ExecContext(ctx, `INSERT INTO campaign_legs
		(campaign_id,sequence,plan_id,requested_quantity,filled_quantity,residual_quantity,state,version,created_at,updated_at)
		VALUES (?,?,?,?,'0',?,'PLANNED',1,?,?)`, req.CampaignID, req.Sequence,
		strings.TrimSpace(req.PlanID), requested, requested, now, now); err != nil {
		return CampaignLegRecord{}, fmt.Errorf("journal: inserting campaign leg: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `UPDATE position_campaigns SET version=?,updated_at=? WHERE id=? AND version=?`,
		newVersion, now, req.CampaignID, version); err != nil {
		return CampaignLegRecord{}, err
	}
	if err := insertCampaignCommand(ctx, tx, req.CampaignID, "PLAN_LEG", req.CommandKey, digest, newVersion, req.Sequence, now); err != nil {
		return CampaignLegRecord{}, err
	}
	if err := insertCampaignEvent(ctx, tx, req.CampaignID, newVersion, newVersion, req.Sequence, "",
		"LEG_PLANNED", "PLAN_LEG", req.CommandKey, digest, state, positioncampaign.LegPlanned, 0, "", "", now); err != nil {
		return CampaignLegRecord{}, err
	}
	if err := tx.Commit(); err != nil {
		return CampaignLegRecord{}, err
	}
	leg, err := j.CampaignLeg(ctx, req.CampaignID, req.Sequence)
	if err == nil {
		leg.CampaignVersion = newVersion
	}
	return leg, err
}

func (j *Journal) LinkCampaignOrder(ctx context.Context, req LinkCampaignOrderRequest) (CampaignOrderRecord, error) {
	req.CampaignID = strings.TrimSpace(req.CampaignID)
	req.CommandKey = strings.TrimSpace(req.CommandKey)
	req.OrderID = strings.TrimSpace(req.OrderID)
	req.PredecessorOrderID = strings.TrimSpace(req.PredecessorOrderID)
	req.IntentID = strings.TrimSpace(req.IntentID)
	req.AttemptID = strings.TrimSpace(req.AttemptID)
	cap, err := positiveCampaignQuantity(req.RequestedCap)
	if err != nil {
		return CampaignOrderRecord{}, fmt.Errorf("%w: requested cap: %v", ErrInvalidRequest, err)
	}
	if req.CampaignID == "" || req.CommandKey == "" || req.OrderID == "" || req.LegSequence <= 0 {
		return CampaignOrderRecord{}, fmt.Errorf("%w: campaign, leg, command and order are required", ErrInvalidRequest)
	}
	if err := (positioncampaign.CampaignLeg{CampaignID: req.CampaignID, Sequence: req.LegSequence,
		PlanID: "linked", IntentID: req.IntentID, AttemptID: req.AttemptID}).Validate(); err != nil {
		return CampaignOrderRecord{}, err
	}
	digest := digestParts(req.CampaignID, strconv.FormatInt(req.LegSequence, 10), req.OrderID,
		req.PredecessorOrderID, cap, req.IntentID, req.AttemptID, strconv.FormatBool(req.LineageAmbiguous))
	now := j.nowString()
	tx, err := j.db.BeginTx(ctx, nil)
	if err != nil {
		return CampaignOrderRecord{}, err
	}
	defer tx.Rollback()
	if _, found, err := campaignCommandResult(ctx, tx, req.CampaignID, "LINK_ORDER", req.CommandKey, digest); err != nil {
		return CampaignOrderRecord{}, err
	} else if found {
		order, err := campaignOrderInTx(ctx, tx, req.CampaignID, req.LegSequence, req.OrderID)
		if err != nil {
			return CampaignOrderRecord{}, err
		}
		if err := tx.Commit(); err != nil {
			return CampaignOrderRecord{}, err
		}
		return order, nil
	}
	state, version, blocked, err := campaignHeaderInTx(ctx, tx, req.CampaignID)
	if err != nil {
		return CampaignOrderRecord{}, err
	}
	if version != req.ExpectedVersion {
		return CampaignOrderRecord{}, versionConflict(req.CampaignID, req.ExpectedVersion, version)
	}
	if exposureBlocked, err := campaignExposureBlockedInTx(ctx, tx, req.CampaignID); err != nil {
		return CampaignOrderRecord{}, err
	} else if exposureBlocked {
		return CampaignOrderRecord{}, positioncampaign.ErrExposureBlocked
	}
	if blocked {
		return CampaignOrderRecord{}, positioncampaign.ErrExposureBlocked
	}
	if req.LineageAmbiguous {
		refusal := fmt.Errorf("%w: caller reported ambiguous order lineage", positioncampaign.ErrInvalidIdentity)
		if err := latchCampaignLinkConflict(ctx, tx, req.CampaignID, version, req.CommandKey, digest, now); err != nil {
			return CampaignOrderRecord{}, err
		}
		if err := tx.Commit(); err != nil {
			return CampaignOrderRecord{}, err
		}
		return CampaignOrderRecord{}, refusal
	}
	leg, err := campaignLegInTx(ctx, tx, req.CampaignID, req.LegSequence)
	if err != nil {
		return CampaignOrderRecord{}, err
	}
	if leg.IntentID != "" && leg.IntentID != req.IntentID {
		return CampaignOrderRecord{}, fmt.Errorf("%w: leg intent %s cannot be rebound to %s",
			positioncampaign.ErrInvalidIdentity, leg.IntentID, req.IntentID)
	}
	authority, err := authoritativeCampaignOrderInTx(ctx, tx, req)
	if err != nil {
		if latchErr := latchCampaignLinkConflict(ctx, tx, req.CampaignID, version, req.CommandKey, digest, now); latchErr != nil {
			return CampaignOrderRecord{}, latchErr
		}
		if commitErr := tx.Commit(); commitErr != nil {
			return CampaignOrderRecord{}, commitErr
		}
		return CampaignOrderRecord{}, err
	}
	var duplicateOrder int
	if err := tx.QueryRowContext(ctx, `SELECT count(*) FROM campaign_order_watermarks
		WHERE account_ref=? AND market=? AND trading_day=? AND symbol=? AND side=? AND order_id=?`,
		authority.accountRef, authority.market, authority.tradingDay, authority.symbol, authority.side, req.OrderID).
		Scan(&duplicateOrder); err != nil {
		return CampaignOrderRecord{}, err
	}
	if duplicateOrder != 0 {
		conflict := fmt.Errorf("%w: broker order %s already belongs to this immutable scope", positioncampaign.ErrInvalidIdentity, req.OrderID)
		if err := latchCampaignLinkConflict(ctx, tx, req.CampaignID, version, req.CommandKey, digest, now); err != nil {
			return CampaignOrderRecord{}, err
		}
		if err := tx.Commit(); err != nil {
			return CampaignOrderRecord{}, err
		}
		return CampaignOrderRecord{}, conflict
	}
	carry := "0"
	if req.PredecessorOrderID != "" {
		var successors int
		if err := tx.QueryRowContext(ctx, `SELECT count(*) FROM campaign_order_watermarks
			WHERE campaign_id=? AND leg_sequence=? AND predecessor_order_id=?`,
			req.CampaignID, req.LegSequence, req.PredecessorOrderID).Scan(&successors); err != nil {
			return CampaignOrderRecord{}, err
		}
		if successors != 0 {
			conflict := fmt.Errorf("%w: predecessor %s already has a successor", positioncampaign.ErrInvalidIdentity, req.PredecessorOrderID)
			if err := latchCampaignLinkConflict(ctx, tx, req.CampaignID, version, req.CommandKey, digest, now); err != nil {
				return CampaignOrderRecord{}, err
			}
			if err := tx.Commit(); err != nil {
				return CampaignOrderRecord{}, err
			}
			return CampaignOrderRecord{}, conflict
		}
		predecessor, err := campaignOrderInTx(ctx, tx, req.CampaignID, req.LegSequence, req.PredecessorOrderID)
		if err != nil {
			return CampaignOrderRecord{}, fmt.Errorf("journal: replacement predecessor: %w", err)
		}
		carry = predecessor.CumulativeFilled
		if _, err := tx.ExecContext(ctx, `UPDATE campaign_order_watermarks SET terminal=1,updated_at=?
			WHERE campaign_id=? AND leg_sequence=? AND order_id=?`, now, req.CampaignID, req.LegSequence, req.PredecessorOrderID); err != nil {
			return CampaignOrderRecord{}, err
		}
	}
	legEvent := positioncampaign.LegOrderLinked
	if req.PredecessorOrderID != "" {
		legEvent = positioncampaign.LegReplacementLinked
	}
	nextLeg, transitionErr := positioncampaign.TransitionLeg(leg.State, legEvent)
	if transitionErr != nil && req.PredecessorOrderID != "" && leg.State == positioncampaign.LegCancelled {
		// A replacement identity can be confirmed after the predecessor's terminal
		// observation. Preserve the explicit lineage and derive the quantity state.
		nextLeg = positioncampaign.LegPartial
		transitionErr = nil
	}
	if transitionErr != nil {
		return CampaignOrderRecord{}, transitionErr
	}
	nextCampaign, err := positioncampaign.TransitionCampaign(state, positioncampaign.CampaignOrderLinked)
	if err != nil {
		return CampaignOrderRecord{}, err
	}
	newVersion := version + 1
	if _, err := tx.ExecContext(ctx, `INSERT INTO campaign_order_watermarks
		(campaign_id,leg_sequence,order_id,account_ref,market,trading_day,symbol,side,decision_id,
		 intent_id,attempt_id,predecessor_order_id,carry_baseline,requested_cap,
		 cumulative_filled,remaining_quantity,terminal,lineage_ambiguous,created_at,updated_at)
		VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,'0',?,0,?, ?,?)`, req.CampaignID, req.LegSequence, req.OrderID,
		authority.accountRef, authority.market, authority.tradingDay, authority.symbol, authority.side, authority.decisionID,
		req.IntentID, req.AttemptID, nullableString(req.PredecessorOrderID), carry, cap, cap,
		boolInt(authority.lineageAmbiguous), now, now); err != nil {
		return CampaignOrderRecord{}, fmt.Errorf("journal: linking campaign order: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `UPDATE campaign_legs
		SET state=?,intent_id=coalesce(intent_id,?),version=version+1,updated_at=?
		WHERE campaign_id=? AND sequence=?`, string(nextLeg), req.IntentID, now,
		req.CampaignID, req.LegSequence); err != nil {
		return CampaignOrderRecord{}, err
	}
	entryBlocked := boolInt(nextCampaign.EntryBlocked)
	nextState := nextCampaign.State
	if _, err := tx.ExecContext(ctx, `UPDATE position_campaigns SET state=?,version=?,entry_blocked=?,updated_at=?
		WHERE id=? AND version=?`, string(nextState), newVersion, entryBlocked, now, req.CampaignID, version); err != nil {
		return CampaignOrderRecord{}, err
	}
	if err := insertCampaignCommand(ctx, tx, req.CampaignID, "LINK_ORDER", req.CommandKey, digest, newVersion, req.LegSequence, now); err != nil {
		return CampaignOrderRecord{}, err
	}
	if err := insertCampaignEvent(ctx, tx, req.CampaignID, newVersion, newVersion, req.LegSequence, req.OrderID,
		"ORDER_LINKED", "LINK_ORDER", req.CommandKey, digest, nextState, nextLeg, 0, "", "", now); err != nil {
		return CampaignOrderRecord{}, err
	}
	if err := tx.Commit(); err != nil {
		return CampaignOrderRecord{}, err
	}
	order, err := j.CampaignOrderWatermark(ctx, req.CampaignID, req.LegSequence, req.OrderID)
	if err == nil {
		order.CampaignVersion = newVersion
	}
	return order, err
}

// UpdateCampaignStop records composition provenance only. It never creates,
// amends or cancels a broker protection order.
func (j *Journal) UpdateCampaignStop(ctx context.Context, req UpdateCampaignStopRequest) (PositionCampaignRecord, error) {
	req.CampaignID = strings.TrimSpace(req.CampaignID)
	req.CommandKey = strings.TrimSpace(req.CommandKey)
	if req.CampaignID == "" || req.CommandKey == "" {
		return PositionCampaignRecord{}, fmt.Errorf("%w: campaign and command key are required", ErrInvalidRequest)
	}
	digest := digestParts(req.CampaignID, req.Candidate.Price, strconv.FormatBool(req.Candidate.Valid),
		req.Candidate.Source, req.Candidate.Policy, req.Candidate.ObservedAt)
	now := j.nowString()
	tx, err := j.db.BeginTx(ctx, nil)
	if err != nil {
		return PositionCampaignRecord{}, err
	}
	defer tx.Rollback()
	if _, found, err := campaignCommandResult(ctx, tx, req.CampaignID, "UPDATE_STOP", req.CommandKey, digest); err != nil {
		return PositionCampaignRecord{}, err
	} else if found {
		if err := tx.Commit(); err != nil {
			return PositionCampaignRecord{}, err
		}
		return j.PositionCampaign(ctx, req.CampaignID)
	}
	stored, err := positionCampaignQuery(ctx, tx, req.CampaignID)
	if err != nil {
		return PositionCampaignRecord{}, err
	}
	if stored.Version != req.ExpectedVersion {
		return PositionCampaignRecord{}, versionConflict(req.CampaignID, req.ExpectedVersion, stored.Version)
	}
	var saved *positioncampaign.EffectiveStop
	if stored.EffectiveStop != "" {
		saved = &positioncampaign.EffectiveStop{
			Price: stored.EffectiveStop, Source: stored.StopSource, Policy: stored.StopPolicy,
			ObservedAt: stored.StopObservedAt, SelectedFrom: stored.StopSelectedFrom,
		}
	}
	effective, block, err := positioncampaign.ComposeLongStop(saved, req.Candidate)
	if err != nil {
		return PositionCampaignRecord{}, err
	}
	newVersion := stored.Version + 1
	blocked := stored.EntryBlocked || block
	var candidate any
	if strings.TrimSpace(req.Candidate.Price) != "" {
		candidate = strings.TrimSpace(req.Candidate.Price)
	}
	var selected any
	if effective.SelectedFrom != "" {
		selected = string(effective.SelectedFrom)
	}
	if _, err := tx.ExecContext(ctx, `UPDATE position_campaigns
		SET effective_stop=?,stop_source=?,stop_policy=?,stop_observed_at=?,stop_candidate=?,
		    stop_candidate_valid=?,stop_candidate_source=?,stop_candidate_policy=?,
		    stop_candidate_observed_at=?,stop_selected_from=?,entry_blocked=?,version=?,updated_at=?
		WHERE id=? AND version=?`, nullableString(effective.Price), nullableString(effective.Source),
		nullableString(effective.Policy), nullableString(effective.ObservedAt), candidate,
		boolInt(req.Candidate.Valid), nullableString(req.Candidate.Source), nullableString(req.Candidate.Policy),
		nullableString(req.Candidate.ObservedAt), selected, boolInt(blocked), newVersion, now,
		req.CampaignID, stored.Version); err != nil {
		return PositionCampaignRecord{}, err
	}
	if err := insertCampaignCommand(ctx, tx, req.CampaignID, "UPDATE_STOP", req.CommandKey, digest, newVersion, 0, now); err != nil {
		return PositionCampaignRecord{}, err
	}
	if err := insertCampaignEvent(ctx, tx, req.CampaignID, newVersion, newVersion, 0, "", "STOP_COMPOSED",
		"UPDATE_STOP", req.CommandKey, digest, stored.State, "", stored.ActualPositionGeneration, "", "", now); err != nil {
		return PositionCampaignRecord{}, err
	}
	if err := tx.Commit(); err != nil {
		return PositionCampaignRecord{}, err
	}
	return j.PositionCampaign(ctx, req.CampaignID)
}

// CancelProspectiveCampaign closes a PLANNED campaign before any submitted
// order exists. It terminates (but never reuses) the prospective token and does
// not call a broker cancellation endpoint.
func (j *Journal) CancelProspectiveCampaign(ctx context.Context, req CancelCampaignRequest) (PositionCampaignRecord, error) {
	req.CampaignID = strings.TrimSpace(req.CampaignID)
	req.CommandKey = strings.TrimSpace(req.CommandKey)
	if req.CampaignID == "" || req.CommandKey == "" {
		return PositionCampaignRecord{}, fmt.Errorf("%w: campaign and command key are required", ErrInvalidRequest)
	}
	digest := digestParts(req.CampaignID, strconv.FormatBool(req.Structural), req.Detail)
	now := j.nowString()
	tx, err := j.db.BeginTx(ctx, nil)
	if err != nil {
		return PositionCampaignRecord{}, err
	}
	defer tx.Rollback()
	if _, found, err := campaignCommandResult(ctx, tx, req.CampaignID, "CANCEL_PROSPECTIVE", req.CommandKey, digest); err != nil {
		return PositionCampaignRecord{}, err
	} else if found {
		if err := tx.Commit(); err != nil {
			return PositionCampaignRecord{}, err
		}
		return j.PositionCampaign(ctx, req.CampaignID)
	}
	state, version, _, err := campaignHeaderInTx(ctx, tx, req.CampaignID)
	if err != nil {
		return PositionCampaignRecord{}, err
	}
	if version != req.ExpectedVersion {
		return PositionCampaignRecord{}, versionConflict(req.CampaignID, req.ExpectedVersion, version)
	}
	event := positioncampaign.CampaignCancelledBeforeFill
	if req.Structural {
		event = positioncampaign.CampaignStructuralInvalid
	}
	next, err := positioncampaign.TransitionCampaign(state, event)
	if err != nil {
		return PositionCampaignRecord{}, err
	}
	var linked int
	if err := tx.QueryRowContext(ctx, `SELECT count(*) FROM campaign_order_watermarks WHERE campaign_id=?`, req.CampaignID).Scan(&linked); err != nil {
		return PositionCampaignRecord{}, err
	}
	if linked != 0 {
		return PositionCampaignRecord{}, fmt.Errorf("%w: campaign %s already has submitted order lineage", positioncampaign.ErrInvalidTransition, req.CampaignID)
	}
	newVersion := version + 1
	if _, err := tx.ExecContext(ctx, `UPDATE campaign_legs SET state='CANCELLED',version=version+1,updated_at=?
		WHERE campaign_id=? AND state='PLANNED'`, now, req.CampaignID); err != nil {
		return PositionCampaignRecord{}, err
	}
	if _, err := tx.ExecContext(ctx, `UPDATE position_campaigns SET state=?,entry_blocked=1,version=?,updated_at=?
		WHERE id=? AND version=?`, string(next.State), newVersion, now, req.CampaignID, version); err != nil {
		return PositionCampaignRecord{}, err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM position_campaign_claims WHERE campaign_id=?`, req.CampaignID); err != nil {
		return PositionCampaignRecord{}, err
	}
	if err := insertCampaignCommand(ctx, tx, req.CampaignID, "CANCEL_PROSPECTIVE", req.CommandKey, digest, newVersion, 0, now); err != nil {
		return PositionCampaignRecord{}, err
	}
	if err := insertCampaignEvent(ctx, tx, req.CampaignID, newVersion, newVersion, 0, "", "PROSPECTIVE_CANCELLED",
		"CANCEL_PROSPECTIVE", req.CommandKey, digest, next.State, "", 0, "", "", now); err != nil {
		return PositionCampaignRecord{}, err
	}
	if err := tx.Commit(); err != nil {
		return PositionCampaignRecord{}, err
	}
	return j.PositionCampaign(ctx, req.CampaignID)
}

// ApplyPositionCampaignFill advances campaign lineage inside the existing fill
// transaction. It never owns Position quantity or price; ProjectPosition has
// already applied those before this hook runs.
type campaignFillMatch struct {
	campaignID, orderID, cap, previous, requested, legState, campaignState string
	legSequence, version, expectedGeneration                               int64
	actualGeneration                                                       sql.NullInt64
	terminal, ambiguous                                                    bool
}

func ApplyPositionCampaignFill(ctx context.Context, tx *ApplyTx, fill AppliedFill) error {
	if err := tx.live(); err != nil {
		return err
	}
	rows, err := tx.Query(ctx, `SELECT w.campaign_id,w.leg_sequence,w.order_id,w.requested_cap,
		w.cumulative_filled,w.terminal,w.lineage_ambiguous,l.requested_quantity,l.state,
		c.state,c.version,c.expected_position_generation,c.actual_position_generation
		FROM campaign_order_watermarks w
		JOIN campaign_legs l ON l.campaign_id=w.campaign_id AND l.sequence=w.leg_sequence
		JOIN position_campaigns c ON c.id=w.campaign_id
		JOIN mutation_attempts a ON a.id=w.attempt_id AND a.intent_id=w.intent_id
		  AND a.broker_order_id=w.order_id AND a.state='CONFIRMED' AND a.kind IN ('PLACE','AMEND')
		JOIN intents i ON i.id=w.intent_id
		WHERE w.order_id=? AND w.account_ref=? AND w.market=? AND w.trading_day=?
		  AND w.symbol=? AND w.side=? AND w.decision_id=coalesce(a.decision_id,'')
		  AND i.account_ref=w.account_ref AND i.market=w.market AND i.trading_day=w.trading_day
		  AND i.symbol=w.symbol AND upper(i.side)=w.side`, fill.OrderID, strings.TrimSpace(fill.AccountRef),
		normaliseMarket(fill.Market), strings.TrimSpace(fill.TradingDay), normaliseSymbol(fill.Symbol),
		strings.ToUpper(strings.TrimSpace(fill.Side)))
	if err != nil {
		return fmt.Errorf("journal: reading campaign watermark for %s: %w", fill.OrderID, err)
	}
	defer rows.Close()
	var matches []campaignFillMatch
	for rows.Next() {
		var item campaignFillMatch
		var terminal, ambiguous int
		if err := rows.Scan(&item.campaignID, &item.legSequence, &item.orderID, &item.cap,
			&item.previous, &terminal, &ambiguous, &item.requested, &item.legState,
			&item.campaignState, &item.version, &item.expectedGeneration, &item.actualGeneration); err != nil {
			return err
		}
		item.terminal, item.ambiguous = terminal != 0, ambiguous != 0
		matches = append(matches, item)
	}
	if err := rows.Err(); err != nil {
		return err
	}
	rows.Close()
	if len(matches) == 0 {
		return nil
	}
	if len(matches) != 1 {
		return latchAmbiguousCampaignFill(ctx, tx, fill, matches)
	}
	item := matches[0]
	cumulative, err := campaignQuantity(fill.CumulativeQuantity)
	if err != nil {
		return fmt.Errorf("journal: campaign cumulative for %s: %w", fill.OrderID, err)
	}
	cmp, err := riskcalc.CompareDecimal(cumulative, item.previous)
	if err != nil {
		return err
	}
	if cmp < 0 {
		// Non-retreat: lower observations never move the watermark.
		return nil
	}
	terminalChanged := fill.Terminal && !item.terminal
	if cmp == 0 && !terminalChanged {
		return nil
	}
	delta := "0"
	if cmp > 0 {
		delta, err = riskcalc.SubDecimal(cumulative, item.previous)
		if err != nil {
			return err
		}
	}
	campaignWasClosed := item.campaignState == string(positioncampaign.CampaignClosed)
	reconcile := item.ambiguous || (item.terminal && delta != "0") || (campaignWasClosed && delta != "0")
	if canonicalFillDelta, deltaErr := campaignQuantity(fill.Delta); deltaErr != nil || canonicalFillDelta != delta {
		reconcile = true
	}
	capCmp, err := riskcalc.CompareDecimal(cumulative, item.cap)
	if err != nil {
		return err
	}
	reconcile = reconcile || capCmp > 0

	actualGeneration := item.actualGeneration.Int64
	if !item.actualGeneration.Valid && delta != "0" {
		if err := tx.tx.QueryRowContext(ctx, `SELECT coalesce(max(instance_seq),0) FROM positions
			WHERE account_ref=? AND market=? AND symbol=?`, strings.TrimSpace(fill.AccountRef),
			normaliseMarket(fill.Market), normaliseSymbol(fill.Symbol)).Scan(&actualGeneration); err != nil {
			return fmt.Errorf("journal: reading successor position generation: %w", err)
		}
		if actualGeneration == item.expectedGeneration+1 {
			if _, err := tx.Exec(ctx, `UPDATE position_campaigns SET actual_position_generation=?
				WHERE id=? AND actual_position_generation IS NULL`, actualGeneration, item.campaignID); err != nil {
				return err
			}
			if _, err := tx.Exec(ctx, `UPDATE position_campaign_claims SET actual_position_generation=?,updated_at=?
				WHERE campaign_id=? AND actual_position_generation IS NULL`, actualGeneration, fill.CommittedAt, item.campaignID); err != nil {
				return err
			}
		} else {
			reconcile = true
			actualGeneration = 0
		}
	} else if item.actualGeneration.Valid {
		var latest int64
		if err := tx.tx.QueryRowContext(ctx, `SELECT coalesce(max(instance_seq),0) FROM positions
			WHERE account_ref=? AND market=? AND symbol=?`, strings.TrimSpace(fill.AccountRef),
			normaliseMarket(fill.Market), normaliseSymbol(fill.Symbol)).Scan(&latest); err != nil {
			return fmt.Errorf("journal: checking bound position generation: %w", err)
		}
		if latest != actualGeneration {
			reconcile = true
		}
	}

	orderRemaining, err := campaignRemaining(item.cap, cumulative)
	if err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `UPDATE campaign_order_watermarks
		SET cumulative_filled=?,remaining_quantity=?,terminal=?,last_observation_id=?,updated_at=?
		WHERE campaign_id=? AND leg_sequence=? AND order_id=?`, cumulative, orderRemaining,
		boolInt(fill.Terminal || item.terminal), fillObservationID(fill), fill.CommittedAt,
		item.campaignID, item.legSequence, item.orderID); err != nil {
		return err
	}
	filled, err := campaignLegFilledInApply(ctx, tx, item.campaignID, item.legSequence)
	if err != nil {
		return err
	}
	residual, err := riskcalc.SubDecimal(item.requested, filled)
	if err != nil {
		return err
	}
	residual, err = riskcalc.MaxDecimal("0", residual)
	if err != nil {
		return err
	}
	requestedCmp, err := riskcalc.CompareDecimal(filled, item.requested)
	if err != nil {
		return err
	}
	legState := positioncampaign.LegPartial
	if requestedCmp >= 0 {
		legState = positioncampaign.LegFilled
	} else if fill.Terminal {
		hasSuccessor, err := hasCampaignSuccessor(ctx, tx, item.campaignID, item.legSequence, item.orderID)
		if err != nil {
			return err
		}
		if !hasSuccessor {
			legState = positioncampaign.LegCancelled
		}
	}
	if requestedCmp > 0 {
		reconcile = true
	}
	if reconcile {
		// Preserve terminal FILLED/CANCELLED when a predecessor receives a late
		// positive delta; the campaign latch carries the ambiguity.
		if item.terminal && (item.legState == string(positioncampaign.LegFilled) || item.legState == string(positioncampaign.LegCancelled)) {
			legState = positioncampaign.LegState(item.legState)
		}
	}
	if _, err := tx.Exec(ctx, `UPDATE campaign_legs SET filled_quantity=?,residual_quantity=?,state=?,version=version+1,updated_at=?
		WHERE campaign_id=? AND sequence=?`, filled, residual, string(legState), fill.CommittedAt,
		item.campaignID, item.legSequence); err != nil {
		return err
	}
	if err := updateCampaignSuccessorRemaining(ctx, tx, item.campaignID, item.legSequence, fill.CommittedAt); err != nil {
		return err
	}
	campaignState := positioncampaign.CampaignState(item.campaignState)
	entryBlocked := campaignState == positioncampaign.CampaignExiting || campaignState == positioncampaign.CampaignReconcile
	if campaignWasClosed {
		campaignState = positioncampaign.CampaignClosed
		entryBlocked = true
		if err := enterReconcileScopeInTx(ctx, tx, strings.TrimSpace(fill.AccountRef), "",
			ReconcileCauseIdentifierConflict, "late fill for CLOSED campaign "+item.campaignID, fill.CommittedAt); err != nil {
			return err
		}
	} else if reconcile {
		campaignState = positioncampaign.CampaignReconcile
		entryBlocked = true
	} else if campaignState == positioncampaign.CampaignPlanned {
		campaignState = positioncampaign.CampaignActive
	}
	newVersion := item.version + 1
	if _, err := tx.Exec(ctx, `UPDATE position_campaigns SET state=?,version=?,entry_blocked=?,updated_at=?
		WHERE id=? AND version=?`, string(campaignState), newVersion, boolInt(entryBlocked),
		fill.CommittedAt, item.campaignID, item.version); err != nil {
		return err
	}
	if legState == positioncampaign.LegCancelled && filled == "0" {
		var nonTerminalOrFilled int
		if err := tx.tx.QueryRowContext(ctx, `SELECT
			(SELECT count(*) FROM campaign_order_watermarks WHERE campaign_id=? AND terminal=0) +
			(SELECT count(*) FROM campaign_legs WHERE campaign_id=? AND (state<>'CANCELLED' OR filled_quantity<>'0'))`,
			item.campaignID, item.campaignID).Scan(&nonTerminalOrFilled); err != nil {
			return err
		}
		if nonTerminalOrFilled == 0 {
			campaignState = positioncampaign.CampaignClosed
			entryBlocked = true
			if _, err := tx.Exec(ctx, `UPDATE position_campaigns SET state='CLOSED',entry_blocked=1 WHERE id=?`, item.campaignID); err != nil {
				return err
			}
			if _, err := tx.Exec(ctx, `DELETE FROM position_campaign_claims WHERE campaign_id=?`, item.campaignID); err != nil {
				return err
			}
		}
	}
	commandKey := "fill:" + fillObservationID(fill)
	digest := digestParts(item.campaignID, item.orderID, cumulative, delta, fill.CommittedAt)
	if err := insertCampaignEvent(ctx, tx.tx, item.campaignID, newVersion, newVersion, item.legSequence,
		item.orderID, "ORDER_WATERMARK_ADVANCED", "APPLY_FILL", commandKey, digest,
		campaignState, legState, actualGeneration, delta, cumulative, fill.CommittedAt); err != nil {
		return err
	}
	return nil
}

func updateCampaignSuccessorRemaining(ctx context.Context, tx *ApplyTx, campaignID string, legSequence int64, now string) error {
	rows, err := tx.Query(ctx, `SELECT order_id,requested_cap,cumulative_filled
		FROM campaign_order_watermarks
		WHERE campaign_id=? AND leg_sequence=? AND predecessor_order_id IS NOT NULL AND terminal=0`,
		campaignID, legSequence)
	if err != nil {
		return err
	}
	type update struct{ orderID, remaining string }
	var updates []update
	for rows.Next() {
		var orderID, cap, cumulative string
		if err := rows.Scan(&orderID, &cap, &cumulative); err != nil {
			rows.Close()
			return err
		}
		remaining, err := campaignRemaining(cap, cumulative)
		if err != nil {
			rows.Close()
			return err
		}
		updates = append(updates, update{orderID: orderID, remaining: remaining})
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return err
	}
	rows.Close()
	for _, item := range updates {
		if _, err := tx.Exec(ctx, `UPDATE campaign_order_watermarks SET remaining_quantity=?,updated_at=?
			WHERE campaign_id=? AND leg_sequence=? AND order_id=?`, item.remaining, now,
			campaignID, legSequence, item.orderID); err != nil {
			return err
		}
	}
	return nil
}

func latchAmbiguousCampaignFill(ctx context.Context, tx *ApplyTx, fill AppliedFill, matches []campaignFillMatch) error {
	seen := make(map[string]struct{}, len(matches))
	for _, item := range matches {
		if _, exists := seen[item.campaignID]; exists {
			continue
		}
		seen[item.campaignID] = struct{}{}
		commandKey := "fill-ambiguity:" + fillObservationID(fill)
		digest := digestParts(item.campaignID, fill.OrderID, fillObservationID(fill), strconv.Itoa(len(matches)))
		if _, found, err := campaignCommandResult(ctx, tx.tx, item.campaignID,
			string(positioncampaign.CommandRecordEvidence), commandKey, digest); err != nil {
			return err
		} else if found {
			continue
		}
		var currentState string
		var currentVersion int64
		if err := tx.tx.QueryRowContext(ctx, `SELECT state,version FROM position_campaigns WHERE id=?`, item.campaignID).
			Scan(&currentState, &currentVersion); err != nil {
			return err
		}
		state := positioncampaign.CampaignReconcile
		if currentState == string(positioncampaign.CampaignClosed) {
			state = positioncampaign.CampaignClosed
		}
		newVersion := currentVersion + 1
		result, err := tx.Exec(ctx, `UPDATE position_campaigns
			SET state=?,entry_blocked=1,version=?,updated_at=? WHERE id=? AND version=?`,
			string(state), newVersion, fill.CommittedAt, item.campaignID, currentVersion)
		if err != nil {
			return err
		}
		if affected, err := result.RowsAffected(); err != nil || affected != 1 {
			return versionConflict(item.campaignID, currentVersion, newVersion)
		}
		if err := insertCampaignCommand(ctx, tx.tx, item.campaignID,
			string(positioncampaign.CommandRecordEvidence), commandKey, digest, newVersion, 0, fill.CommittedAt); err != nil {
			return err
		}
		if err := insertCampaignEvent(ctx, tx.tx, item.campaignID, newVersion, newVersion, 0, "",
			"AMBIGUOUS_ORDER_FILL", string(positioncampaign.CommandRecordEvidence), commandKey, digest,
			state, "", 0, "", "", fill.CommittedAt); err != nil {
			return err
		}
	}
	return enterReconcileScopeInTx(ctx, tx, strings.TrimSpace(fill.AccountRef), "",
		ReconcileCauseIdentifierConflict,
		fmt.Sprintf("campaign order %s has %d authoritative watermark matches", fill.OrderID, len(matches)),
		fill.CommittedAt)
}

func campaignLegFilledInApply(ctx context.Context, tx *ApplyTx, campaignID string, sequence int64) (string, error) {
	rows, err := tx.Query(ctx, `SELECT cumulative_filled FROM campaign_order_watermarks
		WHERE campaign_id=? AND leg_sequence=? ORDER BY order_id`, campaignID, sequence)
	if err != nil {
		return "", err
	}
	defer rows.Close()
	total := "0"
	for rows.Next() {
		var quantity string
		if err := rows.Scan(&quantity); err != nil {
			return "", err
		}
		total, err = riskcalc.AddDecimal(total, quantity)
		if err != nil {
			return "", err
		}
	}
	return total, rows.Err()
}

func hasCampaignSuccessor(ctx context.Context, tx *ApplyTx, campaignID string, sequence int64, orderID string) (bool, error) {
	rows, err := tx.Query(ctx, `SELECT 1 FROM campaign_order_watermarks
		WHERE campaign_id=? AND leg_sequence=? AND predecessor_order_id=? LIMIT 1`, campaignID, sequence, orderID)
	if err != nil {
		return false, fmt.Errorf("journal: reading successor of campaign order %s: %w", orderID, err)
	}
	defer rows.Close()
	found := rows.Next()
	if err := rows.Err(); err != nil {
		return false, fmt.Errorf("journal: reading successor of campaign order %s: %w", orderID, err)
	}
	return found, nil
}

func (j *Journal) PositionCampaign(ctx context.Context, id string) (PositionCampaignRecord, error) {
	return positionCampaignQuery(ctx, j.db, strings.TrimSpace(id))
}

func (j *Journal) CampaignLeg(ctx context.Context, campaignID string, sequence int64) (CampaignLegRecord, error) {
	row := j.db.QueryRowContext(ctx, `SELECT l.campaign_id,l.sequence,l.plan_id,coalesce(l.intent_id,''),l.requested_quantity,
		l.filled_quantity,l.residual_quantity,l.state,l.version,c.version,l.created_at,l.updated_at
		FROM campaign_legs l JOIN position_campaigns c ON c.id=l.campaign_id
		WHERE l.campaign_id=? AND l.sequence=?`, campaignID, sequence)
	return scanCampaignLeg(row)
}

func (j *Journal) CampaignOrderWatermark(ctx context.Context, campaignID string, sequence int64, orderID string) (CampaignOrderRecord, error) {
	row := j.db.QueryRowContext(ctx, `SELECT w.campaign_id,w.leg_sequence,w.order_id,w.intent_id,w.attempt_id,
		coalesce(w.predecessor_order_id,''),w.carry_baseline,w.requested_cap,w.cumulative_filled,
		w.remaining_quantity,w.terminal,w.lineage_ambiguous,coalesce(w.last_observation_id,''),
		c.version,w.created_at,w.updated_at
		FROM campaign_order_watermarks w JOIN position_campaigns c ON c.id=w.campaign_id
		WHERE w.campaign_id=? AND w.leg_sequence=? AND w.order_id=?`, campaignID, sequence, orderID)
	return scanCampaignOrder(row)
}

func (j *Journal) ReconstructPositionCampaign(ctx context.Context, id string) (positioncampaign.ReplayResult, error) {
	return reconstructPositionCampaign(ctx, j.db, id)
}

func (j *Journal) PositionCampaignLineage(ctx context.Context, positionID string) (PositionCampaignLineageRead, error) {
	return positionCampaignLineage(ctx, j.db, strings.TrimSpace(positionID), true)
}

// PositionCampaign exposes the campaign projection on the query-only console
// handle. The type has no campaign command or repair method.
func (r *ReadOnly) PositionCampaign(ctx context.Context, id string) (PositionCampaignRecord, error) {
	if r.version < 20 {
		return PositionCampaignRecord{}, fmt.Errorf("%w: version %d predates position campaigns", ErrSchemaTooOld, r.version)
	}
	return positionCampaignQuery(ctx, r.db, strings.TrimSpace(id))
}

// ReconstructPositionCampaign deterministically validates append-only events
// against the stored projection through the query-only connection.
func (r *ReadOnly) ReconstructPositionCampaign(ctx context.Context, id string) (positioncampaign.ReplayResult, error) {
	if r.version < 20 {
		return positioncampaign.ReplayResult{}, fmt.Errorf("%w: version %d predates position campaigns", ErrSchemaTooOld, r.version)
	}
	return reconstructPositionCampaign(ctx, r.db, id)
}

func (r *ReadOnly) PositionCampaignLineage(ctx context.Context, positionID string) (PositionCampaignLineageRead, error) {
	return positionCampaignLineage(ctx, r.db, strings.TrimSpace(positionID), r.version >= 20)
}

func positionCampaignLineage(ctx context.Context, db *sql.DB, positionID string, versioned bool) (PositionCampaignLineageRead, error) {
	var result PositionCampaignLineageRead
	if !versioned {
		err := db.QueryRowContext(ctx, `SELECT id,account_ref,market,symbol,instance_seq FROM positions WHERE id=?`, positionID).
			Scan(&result.PositionID, &result.AccountRef, &result.Market, &result.Symbol, &result.PositionGeneration)
		if err != nil {
			return PositionCampaignLineageRead{}, err
		}
		result.Status = PositionCampaignLineageLegacyUnknown
		return result, nil
	}
	var version sql.NullInt64
	err := db.QueryRowContext(ctx, `SELECT p.id,p.account_ref,p.market,p.symbol,p.instance_seq,v.version
		FROM positions p LEFT JOIN position_projection_versions v ON v.position_id=p.id WHERE p.id=?`, positionID).
		Scan(&result.PositionID, &result.AccountRef, &result.Market, &result.Symbol, &result.PositionGeneration, &version)
	if err != nil {
		return PositionCampaignLineageRead{}, err
	}
	if !version.Valid {
		result.Status = PositionCampaignLineageLegacyUnknown
		return result, nil
	}
	result.PositionVersion = version.Int64
	err = db.QueryRowContext(ctx, `SELECT id FROM position_campaigns
		WHERE account_ref=? AND market=? AND symbol=? AND actual_position_generation=?
		ORDER BY version DESC LIMIT 1`, result.AccountRef, result.Market, result.Symbol, result.PositionGeneration).
		Scan(&result.CampaignID)
	if errors.Is(err, sql.ErrNoRows) {
		result.Status = PositionCampaignLineageNone
		return result, nil
	}
	if err != nil {
		return PositionCampaignLineageRead{}, err
	}
	result.Status = PositionCampaignLineageKnown
	return result, nil
}

func reconstructPositionCampaign(ctx context.Context, db *sql.DB, id string) (positioncampaign.ReplayResult, error) {
	snapshot, err := positionCampaignQuery(ctx, db, strings.TrimSpace(id))
	if err != nil {
		return positioncampaign.ReplayResult{}, err
	}
	rows, err := db.QueryContext(ctx, `SELECT sequence,campaign_version,event_kind,command_kind,command_key,request_digest,campaign_state,
		coalesce(position_generation,0),prospective_token,expected_position_generation,
		coalesce(leg_sequence,0),coalesce(leg_state,''),coalesce(leg_requested_quantity,''),
		coalesce(leg_filled_quantity,''),coalesce(leg_residual_quantity,''),
		coalesce(plan_id,''),coalesce(intent_id,''),coalesce(attempt_id,''),
		coalesce(order_id,''),coalesce(predecessor_order_id,''),coalesce(carry_baseline,''),
		coalesce(requested_cap,''),coalesce(delta_quantity,''),coalesce(cumulative_quantity,''),
		coalesce(order_remaining_quantity,''),coalesce(order_terminal,0),coalesce(order_lineage_ambiguous,0),
		coalesce(effective_stop,''),coalesce(stop_source,''),coalesce(stop_policy,''),
		coalesce(stop_observed_at,''),entry_blocked,projection_digest
		FROM campaign_events WHERE campaign_id=? ORDER BY sequence`, id)
	if err != nil {
		return positioncampaign.ReplayResult{}, err
	}
	defer rows.Close()
	var events []positioncampaign.Event
	for rows.Next() {
		var event positioncampaign.Event
		var blocked, terminal, ambiguous int
		if err := rows.Scan(&event.Sequence, &event.CampaignVersion, &event.EventKind, &event.CommandKind,
			&event.CommandKey, &event.RequestDigest, &event.CampaignState,
			&event.PositionGeneration, &event.ProspectiveToken, &event.ExpectedPositionGeneration,
			&event.LegSequence, &event.LegState, &event.LegRequestedQuantity, &event.LegFilledQuantity,
			&event.LegResidualQuantity, &event.PlanID, &event.IntentID, &event.AttemptID, &event.OrderID,
			&event.PredecessorOrderID, &event.CarryBaseline, &event.RequestedCap,
			&event.DeltaQuantity, &event.CumulativeQuantity, &event.OrderRemainingQuantity, &terminal, &ambiguous,
			&event.EffectiveStop,
			&event.StopSource, &event.StopPolicy, &event.StopObservedAt, &blocked, &event.ProjectionDigest); err != nil {
			return positioncampaign.ReplayResult{}, err
		}
		event.EntryBlocked = blocked != 0
		event.OrderTerminal = terminal != 0
		event.OrderLineageAmbiguous = ambiguous != 0
		events = append(events, event)
	}
	if err := rows.Err(); err != nil {
		return positioncampaign.ReplayResult{}, err
	}
	projectionDigest, err := campaignProjectionDigest(ctx, db, id)
	if err != nil {
		return positioncampaign.ReplayResult{}, err
	}
	return positioncampaign.Replay(events, positioncampaign.Snapshot{
		CampaignState: snapshot.State, Version: snapshot.Version,
		PositionGeneration: snapshot.ActualPositionGeneration, ProspectiveToken: snapshot.ProspectiveToken,
		EffectiveStop: snapshot.EffectiveStop, EntryBlocked: snapshot.EntryBlocked, ProjectionDigest: projectionDigest,
	}), nil
}

func positionCampaignQuery(ctx context.Context, q interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
}, id string) (PositionCampaignRecord, error) {
	row := q.QueryRowContext(ctx, `SELECT id,account_ref,market,symbol,lane_id,lane_version,
		decision_id,evidence_digest,expected_position_generation,expected_position_version,
		prospective_token,actual_position_generation,state,version,entry_blocked,effective_stop,
		stop_source,stop_policy,stop_observed_at,stop_candidate,stop_candidate_valid,
		stop_candidate_source,stop_candidate_policy,stop_candidate_observed_at,
		stop_selected_from,created_at,updated_at
		FROM position_campaigns WHERE id=?`, id)
	var rec PositionCampaignRecord
	var actual sql.NullInt64
	var state string
	var blocked int
	var stop, source, policy, observed, candidate, candidateSource, candidatePolicy, candidateObserved, selected sql.NullString
	var candidateValid sql.NullInt64
	if err := row.Scan(&rec.ID, &rec.AccountRef, &rec.Market, &rec.Symbol, &rec.LaneID,
		&rec.LaneVersion, &rec.DecisionID, &rec.EvidenceDigest, &rec.ExpectedPositionGeneration,
		&rec.ExpectedPositionVersion, &rec.ProspectiveToken, &actual, &state, &rec.Version,
		&blocked, &stop, &source, &policy, &observed, &candidate, &candidateValid,
		&candidateSource, &candidatePolicy, &candidateObserved, &selected,
		&rec.CreatedAt, &rec.UpdatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return PositionCampaignRecord{}, ErrCampaignNotFound
		}
		return PositionCampaignRecord{}, err
	}
	rec.ActualPositionGeneration = actual.Int64
	rec.State = positioncampaign.CampaignState(state)
	rec.EntryBlocked = blocked != 0
	rec.EffectiveStop, rec.StopSource, rec.StopPolicy, rec.StopObservedAt = stop.String, source.String, policy.String, observed.String
	rec.StopCandidate = candidate.String
	rec.StopCandidateValid = candidateValid.Valid && candidateValid.Int64 != 0
	rec.StopCandidateSource, rec.StopCandidatePolicy, rec.StopCandidateObservedAt =
		candidateSource.String, candidatePolicy.String, candidateObserved.String
	rec.StopSelectedFrom = positioncampaign.StopSelection(selected.String)
	if err := (positioncampaign.PositionCampaign{
		ID: rec.ID, AccountRef: rec.AccountRef, Market: rec.Market, Symbol: rec.Symbol,
		LaneID: rec.LaneID, LaneVersion: rec.LaneVersion, DecisionID: rec.DecisionID,
		EvidenceDigest: rec.EvidenceDigest, ProspectiveToken: rec.ProspectiveToken,
		ExpectedPositionGeneration: rec.ExpectedPositionGeneration,
		ActualPositionGeneration:   rec.ActualPositionGeneration,
	}).Validate(); err != nil {
		return PositionCampaignRecord{}, err
	}
	return rec, nil
}

func campaignLegInTx(ctx context.Context, tx *sql.Tx, campaignID string, sequence int64) (CampaignLegRecord, error) {
	row := tx.QueryRowContext(ctx, `SELECT l.campaign_id,l.sequence,l.plan_id,coalesce(l.intent_id,''),l.requested_quantity,
		l.filled_quantity,l.residual_quantity,l.state,l.version,c.version,l.created_at,l.updated_at
		FROM campaign_legs l JOIN position_campaigns c ON c.id=l.campaign_id
		WHERE l.campaign_id=? AND l.sequence=?`, campaignID, sequence)
	return scanCampaignLeg(row)
}

func scanCampaignLeg(row rowScanner) (CampaignLegRecord, error) {
	var rec CampaignLegRecord
	var state string
	if err := row.Scan(&rec.CampaignID, &rec.Sequence, &rec.PlanID, &rec.IntentID, &rec.RequestedQuantity,
		&rec.FilledQuantity, &rec.ResidualQuantity, &state, &rec.Version, &rec.CampaignVersion,
		&rec.CreatedAt, &rec.UpdatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return CampaignLegRecord{}, ErrCampaignLegNotFound
		}
		return CampaignLegRecord{}, err
	}
	rec.State = positioncampaign.LegState(state)
	return rec, nil
}

func campaignOrderInTx(ctx context.Context, tx *sql.Tx, campaignID string, sequence int64, orderID string) (CampaignOrderRecord, error) {
	row := tx.QueryRowContext(ctx, `SELECT w.campaign_id,w.leg_sequence,w.order_id,w.intent_id,w.attempt_id,
		coalesce(w.predecessor_order_id,''),w.carry_baseline,w.requested_cap,w.cumulative_filled,
		w.remaining_quantity,w.terminal,w.lineage_ambiguous,coalesce(w.last_observation_id,''),
		c.version,w.created_at,w.updated_at
		FROM campaign_order_watermarks w JOIN position_campaigns c ON c.id=w.campaign_id
		WHERE w.campaign_id=? AND w.leg_sequence=? AND w.order_id=?`, campaignID, sequence, orderID)
	return scanCampaignOrder(row)
}

func scanCampaignOrder(row rowScanner) (CampaignOrderRecord, error) {
	var rec CampaignOrderRecord
	var terminal, ambiguous int
	if err := row.Scan(&rec.CampaignID, &rec.LegSequence, &rec.OrderID, &rec.IntentID, &rec.AttemptID, &rec.PredecessorOrderID,
		&rec.CarryBaseline, &rec.RequestedCap, &rec.CumulativeFilled, &rec.RemainingQuantity,
		&terminal, &ambiguous, &rec.LastObservationID, &rec.CampaignVersion,
		&rec.CreatedAt, &rec.UpdatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return CampaignOrderRecord{}, ErrCampaignLegNotFound
		}
		return CampaignOrderRecord{}, err
	}
	rec.Terminal, rec.LineageAmbiguous = terminal != 0, ambiguous != 0
	return rec, nil
}

func campaignHeaderInTx(ctx context.Context, tx *sql.Tx, id string) (positioncampaign.CampaignState, int64, bool, error) {
	var state string
	var version int64
	var blocked int
	if err := tx.QueryRowContext(ctx, `SELECT state,version,entry_blocked FROM position_campaigns WHERE id=?`, id).
		Scan(&state, &version, &blocked); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", 0, false, ErrCampaignNotFound
		}
		return "", 0, false, err
	}
	return positioncampaign.CampaignState(state), version, blocked != 0, nil
}

// campaignExposureBlockedInTx is the journal admission port for EXIT FIRST.
// It is intentionally a bounded local read: it never waits for or calls the
// broker, and therefore cannot delay fill detection or risk-reducing paths.
func campaignExposureBlockedInTx(ctx context.Context, tx *sql.Tx, campaignID string) (bool, error) {
	var account, market, symbol string
	if err := tx.QueryRowContext(ctx, `SELECT account_ref,market,symbol FROM position_campaigns WHERE id=?`, campaignID).
		Scan(&account, &market, &symbol); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, ErrCampaignNotFound
		}
		return false, fmt.Errorf("journal: reading campaign exposure scope: %w", err)
	}
	var positionClosing int
	if err := tx.QueryRowContext(ctx, `SELECT EXISTS(
		SELECT 1 FROM positions p
		WHERE p.account_ref=? AND p.market=? AND p.symbol=? AND p.state='CLOSING'
		  AND p.instance_seq=(SELECT max(p2.instance_seq) FROM positions p2
			WHERE p2.account_ref=p.account_ref AND p2.market=p.market AND p2.symbol=p.symbol))`,
		account, market, symbol).Scan(&positionClosing); err != nil {
		return false, fmt.Errorf("journal: checking CLOSING position admission: %w", err)
	}
	if positionClosing != 0 {
		return true, nil
	}
	var riskReducingPending int
	if err := tx.QueryRowContext(ctx, `SELECT EXISTS(
		SELECT 1 FROM intents i
		WHERE i.account_ref=? AND i.market=? AND i.symbol=? AND upper(i.side)='SELL'
		AND (
			NOT EXISTS (SELECT 1 FROM mutation_attempts a WHERE a.intent_id=i.id)
			OR EXISTS (SELECT 1 FROM mutation_attempts a WHERE a.intent_id=i.id
				AND a.state IN ('RECORDED','DISPATCHED','IN_DOUBT'))
			OR EXISTS (SELECT 1 FROM mutation_attempts a WHERE a.intent_id=i.id
				AND a.state='CONFIRMED' AND a.kind IN ('PLACE','AMEND') AND a.broker_order_id<>''
				AND NOT EXISTS (SELECT 1 FROM scoped_fill_snapshots f
					WHERE f.account_ref=i.account_ref AND f.market=i.market AND f.trading_day=i.trading_day
					  AND f.symbol=i.symbol AND f.side=i.side AND f.order_id=a.broker_order_id AND f.terminal=1))
		))`, account, market, symbol).Scan(&riskReducingPending); err != nil {
		return false, fmt.Errorf("journal: checking unresolved risk-reducing admission: %w", err)
	}
	return riskReducingPending != 0, nil
}

type campaignOrderAuthority struct {
	accountRef, market, tradingDay, symbol, side, decisionID string
	lineageAmbiguous                                         bool
}

func authoritativeCampaignOrderInTx(ctx context.Context, tx *sql.Tx, req LinkCampaignOrderRequest) (campaignOrderAuthority, error) {
	var authority campaignOrderAuthority
	var kind, state, brokerOrderID, targetOrderID, intentQuantity string
	err := tx.QueryRowContext(ctx, `SELECT i.account_ref,i.market,i.trading_day,i.symbol,upper(i.side),
		coalesce(a.decision_id,''),a.kind,a.state,a.broker_order_id,a.target_order_id,i.quantity
		FROM mutation_attempts a JOIN intents i ON i.id=a.intent_id
		WHERE a.id=? AND i.id=?`, req.AttemptID, req.IntentID).
		Scan(&authority.accountRef, &authority.market, &authority.tradingDay, &authority.symbol,
			&authority.side, &authority.decisionID, &kind, &state, &brokerOrderID, &targetOrderID, &intentQuantity)
	if errors.Is(err, sql.ErrNoRows) {
		return campaignOrderAuthority{}, fmt.Errorf("%w: intent/attempt lineage %s/%s does not exist",
			positioncampaign.ErrInvalidIdentity, req.IntentID, req.AttemptID)
	}
	if err != nil {
		return campaignOrderAuthority{}, fmt.Errorf("journal: reading campaign execution lineage: %w", err)
	}
	var campaignAccount, campaignMarket, campaignSymbol, campaignDecision string
	if err := tx.QueryRowContext(ctx, `SELECT account_ref,market,symbol,decision_id FROM position_campaigns WHERE id=?`, req.CampaignID).
		Scan(&campaignAccount, &campaignMarket, &campaignSymbol, &campaignDecision); err != nil {
		return campaignOrderAuthority{}, err
	}
	valid := state == string(StateConfirmed) && (kind == "PLACE" || kind == "AMEND") &&
		brokerOrderID == req.OrderID && authority.accountRef == campaignAccount &&
		normaliseMarket(authority.market) == normaliseMarket(campaignMarket) &&
		normaliseSymbol(authority.symbol) == normaliseSymbol(campaignSymbol) &&
		authority.side == "BUY" && authority.decisionID == campaignDecision
	canonicalCap, capErr := campaignQuantity(req.RequestedCap)
	valid = valid && capErr == nil
	if req.PredecessorOrderID == "" {
		canonicalIntentQuantity, quantityErr := campaignQuantity(intentQuantity)
		valid = valid && kind == "PLACE" && strings.TrimSpace(targetOrderID) == "" &&
			quantityErr == nil && canonicalIntentQuantity == canonicalCap
	} else {
		valid = valid && kind == "AMEND" && targetOrderID == req.PredecessorOrderID
		var edgeRequested string
		if err := tx.QueryRowContext(ctx, `SELECT requested_quantity FROM scoped_lineage_edges
			WHERE parent_order_id=? AND child_order_id=? AND relation='replaces'
			  AND account_ref=? AND market=? AND trading_day=? AND symbol=? AND side=?
			  AND intent_id=? AND attempt_id=?`, req.PredecessorOrderID, req.OrderID,
			authority.accountRef, authority.market, authority.tradingDay, authority.symbol,
			authority.side, req.IntentID, req.AttemptID).Scan(&edgeRequested); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				valid = false
			} else {
				return campaignOrderAuthority{}, err
			}
		}
		canonicalEdge, edgeErr := campaignQuantity(edgeRequested)
		valid = valid && edgeErr == nil && canonicalEdge == canonicalCap
	}
	if !valid {
		return campaignOrderAuthority{}, fmt.Errorf("%w: attempt %s is not authoritative for campaign order %s",
			positioncampaign.ErrInvalidIdentity, req.AttemptID, req.OrderID)
	}
	authority.market = normaliseMarket(authority.market)
	authority.symbol = normaliseSymbol(authority.symbol)
	return authority, nil
}

func latchCampaignLinkConflict(ctx context.Context, tx *sql.Tx, campaignID string, version int64, commandKey, digest, now string) error {
	newVersion := version + 1
	result, err := tx.ExecContext(ctx, `UPDATE position_campaigns
		SET state='RECONCILE',entry_blocked=1,version=version+1,updated_at=?
		WHERE id=? AND version=?`, now, campaignID, version)
	if err != nil {
		return err
	}
	if affected, err := result.RowsAffected(); err != nil || affected != 1 {
		return versionConflict(campaignID, version, version+1)
	}
	if err := insertCampaignRefusalCommand(ctx, tx, campaignID, string(positioncampaign.CommandLinkOrder),
		commandKey, digest, newVersion, now); err != nil {
		return err
	}
	return insertCampaignEvent(ctx, tx, campaignID, newVersion, newVersion, 0, "",
		"ORDER_LINK_REFUSED", string(positioncampaign.CommandLinkOrder), commandKey, digest,
		positioncampaign.CampaignReconcile, "", 0, "", "", now)
}

type commandResult struct {
	campaignID string
	version    int64
	sequence   int64
	errorCode  string
}

func campaignCommandResult(ctx context.Context, tx *sql.Tx, campaignID, kind, key, digest string) (commandResult, bool, error) {
	if err := positioncampaign.ValidateCommand(kind, key); err != nil {
		return commandResult{}, false, err
	}
	var stored string
	var result commandResult
	var sequence sql.NullInt64
	var resultError sql.NullString
	err := tx.QueryRowContext(ctx, `SELECT campaign_id,request_digest,result_version,result_sequence,result_error
		FROM campaign_commands WHERE campaign_id=? AND command_kind=? AND command_key=?`,
		campaignID, kind, strings.TrimSpace(key)).Scan(&result.campaignID, &stored, &result.version, &sequence, &resultError)
	if errors.Is(err, sql.ErrNoRows) {
		return commandResult{}, false, nil
	}
	if err != nil {
		return commandResult{}, false, err
	}
	if stored != digest {
		return commandResult{}, false, fmt.Errorf("%w: %s/%s reused with a different request", ErrCampaignCommandConflict, kind, key)
	}
	result.sequence = sequence.Int64
	result.errorCode = resultError.String
	if result.errorCode == "INVALID_IDENTITY" {
		return result, true, fmt.Errorf("%w: deterministic campaign command refusal", positioncampaign.ErrInvalidIdentity)
	}
	return result, true, nil
}

func insertCampaignCommand(ctx context.Context, tx *sql.Tx, campaignID, kind, key, digest string, version, sequence int64, now string) error {
	if err := positioncampaign.ValidateCommand(kind, key); err != nil {
		return err
	}
	var resultSequence any
	if sequence > 0 {
		resultSequence = sequence
	}
	_, err := tx.ExecContext(ctx, `INSERT INTO campaign_commands
		(campaign_id,command_kind,command_key,request_digest,result_version,result_sequence,recorded_at)
		VALUES (?,?,?,?,?,?,?)`, campaignID, kind, strings.TrimSpace(key), digest, version, resultSequence, now)
	if err != nil {
		return fmt.Errorf("journal: recording campaign command: %w", err)
	}
	return nil
}

func insertCampaignRefusalCommand(ctx context.Context, tx *sql.Tx, campaignID, kind, key, digest string, version int64, now string) error {
	if err := positioncampaign.ValidateCommand(kind, key); err != nil {
		return err
	}
	_, err := tx.ExecContext(ctx, `INSERT INTO campaign_commands
		(campaign_id,command_kind,command_key,request_digest,result_version,result_error,recorded_at)
		VALUES (?,?,?,?,?,'INVALID_IDENTITY',?)`, campaignID, kind, strings.TrimSpace(key), digest, version, now)
	if err != nil {
		return fmt.Errorf("journal: recording campaign refusal: %w", err)
	}
	return nil
}

func insertCampaignEvent(ctx context.Context, tx *sql.Tx, campaignID string, sequence, version, legSequence int64,
	orderID, eventKind, commandKind, commandKey, digest string, campaignState positioncampaign.CampaignState,
	legState positioncampaign.LegState, positionGeneration int64, delta, cumulative, now string,
) error {
	if err := positioncampaign.ValidateCommand(commandKind, commandKey); err != nil {
		return err
	}
	var leg any
	if legSequence > 0 {
		leg = legSequence
	}
	var order any
	if orderID != "" {
		order = orderID
	}
	var legValue any
	if legState != "" {
		legValue = string(legState)
	}
	var generation any
	if positionGeneration > 0 {
		generation = positionGeneration
	}
	var deltaValue, cumulativeValue any
	if delta != "" {
		deltaValue = delta
	}
	if cumulative != "" {
		cumulativeValue = cumulative
	}
	var token string
	var expectedGeneration int64
	var eventEntryBlocked int
	var effectiveStop, stopSource, stopPolicy, stopObserved sql.NullString
	if err := tx.QueryRowContext(ctx, `SELECT prospective_token,expected_position_generation,
		effective_stop,stop_source,stop_policy,stop_observed_at,entry_blocked FROM position_campaigns WHERE id=?`,
		campaignID).Scan(&token, &expectedGeneration, &effectiveStop, &stopSource, &stopPolicy, &stopObserved,
		&eventEntryBlocked); err != nil {
		return fmt.Errorf("journal: reading campaign event identity: %w", err)
	}
	var planID, intentID, attemptID, predecessorID, carryBaseline, requestedCap sql.NullString
	var legRequested, legFilled, legResidual sql.NullString
	var orderRemaining sql.NullString
	var orderTerminal, orderAmbiguous sql.NullInt64
	if legSequence > 0 {
		if err := tx.QueryRowContext(ctx, `SELECT plan_id,intent_id,requested_quantity,filled_quantity,residual_quantity,state
			FROM campaign_legs WHERE campaign_id=? AND sequence=?`, campaignID, legSequence).
			Scan(&planID, &intentID, &legRequested, &legFilled, &legResidual, &legValue); err != nil {
			return fmt.Errorf("journal: reading campaign leg event identity: %w", err)
		}
	}
	if orderID != "" {
		if err := tx.QueryRowContext(ctx, `SELECT intent_id,attempt_id,predecessor_order_id,carry_baseline,requested_cap,
			remaining_quantity,terminal,lineage_ambiguous
			FROM campaign_order_watermarks WHERE campaign_id=? AND leg_sequence=? AND order_id=?`,
			campaignID, legSequence, orderID).Scan(&intentID, &attemptID, &predecessorID, &carryBaseline,
			&requestedCap, &orderRemaining, &orderTerminal, &orderAmbiguous); err != nil {
			return fmt.Errorf("journal: reading campaign order event identity: %w", err)
		}
	}
	projectionDigest, err := campaignProjectionDigest(ctx, tx, campaignID)
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO campaign_events
		(campaign_id,sequence,campaign_version,leg_sequence,order_id,event_kind,command_kind,
		 command_key,request_digest,campaign_state,leg_state,leg_requested_quantity,leg_filled_quantity,
		 leg_residual_quantity,position_generation,delta_quantity,
		 cumulative_quantity,prospective_token,expected_position_generation,plan_id,intent_id,
		 attempt_id,predecessor_order_id,carry_baseline,requested_cap,order_remaining_quantity,
		 order_terminal,order_lineage_ambiguous,effective_stop,stop_source,
		 stop_policy,stop_observed_at,entry_blocked,projection_digest,recorded_at)
		VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`, campaignID, sequence, version,
		leg, order, eventKind, commandKind, commandKey, digest, string(campaignState), legValue,
		legRequested, legFilled, legResidual, generation, deltaValue, cumulativeValue, token, expectedGeneration, planID, intentID,
		attemptID, predecessorID, carryBaseline, requestedCap, orderRemaining, orderTerminal, orderAmbiguous,
		effectiveStop, stopSource, stopPolicy,
		stopObserved, eventEntryBlocked, projectionDigest, now)
	if err != nil {
		return fmt.Errorf("journal: appending campaign event: %w", err)
	}
	return nil
}

type campaignProjectionQuerier interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
}

func campaignProjectionDigest(ctx context.Context, q campaignProjectionQuerier, campaignID string) (string, error) {
	parts := []string{"campaign", campaignID}
	var account, market, symbol, laneID, laneVersion, decisionID, evidence, token, state string
	var expectedGeneration, expectedVersion, version, blocked int64
	var actual sql.NullInt64
	var stop, source, policy, observed sql.NullString
	if err := q.QueryRowContext(ctx, `SELECT account_ref,market,symbol,lane_id,lane_version,decision_id,
		evidence_digest,expected_position_generation,expected_position_version,prospective_token,
		state,version,entry_blocked,actual_position_generation,
		effective_stop,stop_source,stop_policy,stop_observed_at FROM position_campaigns WHERE id=?`, campaignID).
		Scan(&account, &market, &symbol, &laneID, &laneVersion, &decisionID, &evidence,
			&expectedGeneration, &expectedVersion, &token, &state, &version, &blocked, &actual,
			&stop, &source, &policy, &observed); err != nil {
		return "", fmt.Errorf("journal: digesting campaign projection: %w", err)
	}
	parts = append(parts, "identity", account, market, symbol, laneID, laneVersion, decisionID, evidence,
		strconv.FormatInt(expectedGeneration, 10), strconv.FormatInt(expectedVersion, 10), token,
		state, strconv.FormatInt(version, 10), strconv.FormatInt(blocked, 10),
		strconv.FormatInt(actual.Int64, 10), stop.String, source.String, policy.String, observed.String)
	var claimAccount, claimMarket, claimSymbol, claimToken, claimCampaign, claimCreated, claimUpdated string
	var claimGeneration, claimPositionVersion, claimVersion int64
	var claimActual sql.NullInt64
	err := q.QueryRowContext(ctx, `SELECT account_ref,market,symbol,position_generation,position_version,version,
		prospective_token,campaign_id,actual_position_generation,created_at,updated_at
		FROM position_campaign_claims WHERE campaign_id=?`, campaignID).
		Scan(&claimAccount, &claimMarket, &claimSymbol, &claimGeneration, &claimPositionVersion, &claimVersion,
			&claimToken, &claimCampaign, &claimActual, &claimCreated, &claimUpdated)
	if errors.Is(err, sql.ErrNoRows) {
		parts = append(parts, "claim", "none")
	} else if err != nil {
		return "", fmt.Errorf("journal: digesting campaign claim: %w", err)
	} else {
		parts = append(parts, "claim", claimAccount, claimMarket, claimSymbol,
			strconv.FormatInt(claimGeneration, 10), strconv.FormatInt(claimPositionVersion, 10),
			strconv.FormatInt(claimVersion, 10), claimToken, claimCampaign,
			strconv.FormatInt(claimActual.Int64, 10), claimCreated, claimUpdated)
	}
	legs, err := q.QueryContext(ctx, `SELECT sequence,plan_id,coalesce(intent_id,''),requested_quantity,
		filled_quantity,residual_quantity,state,version FROM campaign_legs WHERE campaign_id=? ORDER BY sequence`, campaignID)
	if err != nil {
		return "", err
	}
	for legs.Next() {
		var sequence, legVersion int64
		var plan, intent, requested, filled, residual, legState string
		if err := legs.Scan(&sequence, &plan, &intent, &requested, &filled, &residual, &legState, &legVersion); err != nil {
			legs.Close()
			return "", err
		}
		parts = append(parts, "leg", strconv.FormatInt(sequence, 10), plan, intent, requested, filled,
			residual, legState, strconv.FormatInt(legVersion, 10))
	}
	if err := legs.Err(); err != nil {
		legs.Close()
		return "", err
	}
	legs.Close()
	orders, err := q.QueryContext(ctx, `SELECT leg_sequence,order_id,account_ref,market,trading_day,symbol,side,
		decision_id,intent_id,attempt_id,coalesce(predecessor_order_id,''),carry_baseline,requested_cap,
		cumulative_filled,remaining_quantity,terminal,lineage_ambiguous,coalesce(last_observation_id,'')
		FROM campaign_order_watermarks WHERE campaign_id=? ORDER BY leg_sequence,order_id`, campaignID)
	if err != nil {
		return "", err
	}
	for orders.Next() {
		var sequence, terminal, ambiguous int64
		var values [15]string
		if err := orders.Scan(&sequence, &values[0], &values[1], &values[2], &values[3], &values[4],
			&values[5], &values[6], &values[7], &values[8], &values[9], &values[10], &values[11],
			&values[12], &values[13], &terminal, &ambiguous, &values[14]); err != nil {
			orders.Close()
			return "", err
		}
		parts = append(parts, "order", strconv.FormatInt(sequence, 10))
		parts = append(parts, values[:]...)
		parts = append(parts, strconv.FormatInt(terminal, 10), strconv.FormatInt(ambiguous, 10))
	}
	if err := orders.Err(); err != nil {
		orders.Close()
		return "", err
	}
	orders.Close()
	commands, err := q.QueryContext(ctx, `SELECT command_kind,command_key,request_digest,result_version,
		coalesce(result_sequence,0),coalesce(result_error,''),recorded_at FROM campaign_commands WHERE campaign_id=?
		ORDER BY command_kind,command_key`, campaignID)
	if err != nil {
		return "", err
	}
	for commands.Next() {
		var kind, key, requestDigest, resultError, recorded string
		var resultVersion, resultSequence int64
		if err := commands.Scan(&kind, &key, &requestDigest, &resultVersion, &resultSequence, &resultError, &recorded); err != nil {
			commands.Close()
			return "", err
		}
		parts = append(parts, "command", kind, key, requestDigest, strconv.FormatInt(resultVersion, 10),
			strconv.FormatInt(resultSequence, 10), resultError, recorded)
	}
	if err := commands.Err(); err != nil {
		commands.Close()
		return "", err
	}
	commands.Close()
	return digestParts(parts...), nil
}

func campaignQuantity(value string) (string, error) {
	canonical, err := riskcalc.CanonicalDecimal(value)
	if err != nil {
		return "", err
	}
	negative, err := riskcalc.IsNegativeDecimal(canonical)
	if err != nil {
		return "", err
	}
	if negative {
		return "", fmt.Errorf("negative quantity %s", canonical)
	}
	return canonical, nil
}

func campaignRemaining(cap, cumulative string) (string, error) {
	cap, err := campaignQuantity(cap)
	if err != nil {
		return "", err
	}
	cumulative, err = campaignQuantity(cumulative)
	if err != nil {
		return "", err
	}
	cmp, err := riskcalc.CompareDecimal(cumulative, cap)
	if err != nil {
		return "", err
	}
	if cmp >= 0 {
		return "0", nil
	}
	return riskcalc.SubDecimal(cap, cumulative)
}

func positiveCampaignQuantity(value string) (string, error) {
	canonical, err := campaignQuantity(value)
	if err != nil {
		return "", err
	}
	cmp, err := riskcalc.CompareDecimal(canonical, "0")
	if err != nil {
		return "", err
	}
	if cmp <= 0 {
		return "", fmt.Errorf("quantity must be positive")
	}
	return canonical, nil
}

func digestParts(parts ...string) string {
	h := sha256.New()
	for _, part := range parts {
		h.Write([]byte(strconv.Itoa(len(part))))
		h.Write([]byte{':'})
		h.Write([]byte(part))
	}
	return hex.EncodeToString(h.Sum(nil))
}

func fillObservationID(fill AppliedFill) string {
	return digestParts(fill.AccountRef, normaliseMarket(fill.Market), normaliseSymbol(fill.Symbol),
		fill.TradingDay, fill.Side, fill.OrderID, fill.CumulativeQuantity, fill.AveragePrice,
		fill.FilledAmount, strconv.FormatBool(fill.Terminal))
}

func versionConflict(id string, expected, actual int64) error {
	return fmt.Errorf("%w: campaign %s expected %d, actual %d", ErrCampaignVersionConflict, id, expected, actual)
}

func isConstraintError(err error) bool {
	if err == nil {
		return false
	}
	text := strings.ToLower(err.Error())
	return strings.Contains(text, "constraint") || strings.Contains(text, "unique")
}
