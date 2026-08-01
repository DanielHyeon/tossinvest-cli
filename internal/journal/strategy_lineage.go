package journal

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

const schemaV14 = `
CREATE TABLE strategy_decision_lineage(
 entry_decision_identity TEXT PRIMARY KEY,
 candidate_life_id TEXT NOT NULL,
 market TEXT NOT NULL,
 symbol TEXT NOT NULL,
 threshold_version TEXT NOT NULL,
 threshold_set_digest TEXT NOT NULL,
 evidence_digest TEXT NOT NULL,
 lane_id TEXT NOT NULL,
 lane_version TEXT NOT NULL,
 lane_source_digest TEXT NOT NULL,
 lane_constants_digest TEXT NOT NULL,
 entry_price TEXT NOT NULL,
 stop_price TEXT NOT NULL,
 target_price TEXT NOT NULL,
 quantity TEXT NOT NULL,
 policy_version TEXT NOT NULL,
 settings_digest TEXT NOT NULL,
 decision_payload TEXT NOT NULL,
 decision_payload_digest TEXT NOT NULL,
 activation_manifest_digest TEXT NOT NULL,
 created_at TEXT NOT NULL
) STRICT;
CREATE TABLE strategy_attempt_lineage(
 attempt_id TEXT PRIMARY KEY,
 account_ref TEXT NOT NULL,
 entry_decision_identity TEXT NOT NULL UNIQUE REFERENCES strategy_decision_lineage(entry_decision_identity),
 risk_intent_id TEXT NOT NULL UNIQUE REFERENCES decisions(id),
 guardian_decision_id TEXT NOT NULL,
 activation_manifest_digest TEXT NOT NULL,
 client_order_id TEXT NOT NULL UNIQUE,
 revision INTEGER NOT NULL CHECK(revision>=1),
 state TEXT NOT NULL CHECK(state IN('PLANNED','REFUSED','DISPATCHED','IN_DOUBT')),
 created_at TEXT NOT NULL
) STRICT;
CREATE INDEX idx_strategy_attempt_account_risk ON strategy_attempt_lineage(account_ref,risk_intent_id);
CREATE TABLE strategy_execution_lineage(
 id INTEGER PRIMARY KEY AUTOINCREMENT,
 account_ref TEXT NOT NULL,
 attempt_id TEXT NOT NULL REFERENCES strategy_attempt_lineage(attempt_id),
 kind TEXT NOT NULL CHECK(kind IN('DISPATCH_START','MUTATION_ATTEMPT','BROKER_ORDER','FILL','POSITION','CLOSE_OUTCOME')),
 external_ref TEXT NOT NULL,
 recorded_at TEXT NOT NULL,
 UNIQUE(attempt_id,kind,external_ref),
 UNIQUE(account_ref,kind,external_ref)
) STRICT;
CREATE UNIQUE INDEX idx_strategy_execution_reverse ON strategy_execution_lineage(account_ref,kind,external_ref,attempt_id);
CREATE UNIQUE INDEX idx_strategy_execution_singleton ON strategy_execution_lineage(attempt_id,kind) WHERE kind IN('DISPATCH_START','MUTATION_ATTEMPT','BROKER_ORDER','POSITION','CLOSE_OUTCOME');
CREATE TABLE strategy_attempt_refusals(
 id INTEGER PRIMARY KEY AUTOINCREMENT,
 attempt_id TEXT NOT NULL REFERENCES strategy_attempt_lineage(attempt_id),
 revision INTEGER NOT NULL,
 reason_code TEXT NOT NULL,
 recorded_at TEXT NOT NULL,
 UNIQUE(attempt_id,revision)
) STRICT;
CREATE TRIGGER strategy_decision_lineage_no_update BEFORE UPDATE ON strategy_decision_lineage BEGIN SELECT RAISE(ABORT,'strategy decision lineage is immutable');END;
CREATE TRIGGER strategy_decision_lineage_no_delete BEFORE DELETE ON strategy_decision_lineage BEGIN SELECT RAISE(ABORT,'strategy decision lineage is immutable');END;
CREATE TRIGGER strategy_attempt_lineage_no_delete BEFORE DELETE ON strategy_attempt_lineage BEGIN SELECT RAISE(ABORT,'strategy attempt lineage is immutable');END;
CREATE TRIGGER strategy_execution_lineage_no_update BEFORE UPDATE ON strategy_execution_lineage BEGIN SELECT RAISE(ABORT,'strategy execution lineage is append-only');END;
CREATE TRIGGER strategy_execution_lineage_no_delete BEFORE DELETE ON strategy_execution_lineage BEGIN SELECT RAISE(ABORT,'strategy execution lineage is append-only');END;
CREATE TRIGGER strategy_attempt_refusals_no_update BEFORE UPDATE ON strategy_attempt_refusals BEGIN SELECT RAISE(ABORT,'strategy refusal history is append-only');END;
CREATE TRIGGER strategy_attempt_refusals_no_delete BEFORE DELETE ON strategy_attempt_refusals BEGIN SELECT RAISE(ABORT,'strategy refusal history is append-only');END;`

var ErrStrategyTraceNotFound = errors.New("journal strategy trace: not found")

type StrategyCollisionError struct{ Stage string }

func (e *StrategyCollisionError) Error() string {
	return "journal strategy plan: " + e.Stage + " identity collision"
}

type StrategyDecisionLineage struct {
	DecisionIdentity         string
	CandidateLifeID          string
	Market                   string
	Symbol                   string
	ThresholdVersion         string
	ThresholdSetDigest       string
	EvidenceDigest           string
	LaneID                   string
	LaneVersion              string
	LaneSourceDigest         string
	LaneConstantsDigest      string
	EntryPrice               string
	StopPrice                string
	TargetPrice              string
	Quantity                 string
	PolicyVersion            string
	SettingsDigest           string
	DecisionPayload          string
	DecisionPayloadDigest    string
	ActivationManifestDigest string
	CreatedAt                time.Time
}

type StrategyAtomicPlan struct {
	RiskDecision             DecisionRequest
	Lineage                  StrategyDecisionLineage
	AttemptID                string
	GuardianDecisionID       string
	ActivationManifestDigest string
	ClientOrderID            string
	Revision                 int
	CreatedAt                time.Time
}

type StrategyPlanRequest struct {
	Lineage                  StrategyDecisionLineage
	AttemptID                string
	ActivationManifestDigest string
	Revision                 int
	CreatedAt                time.Time
}

type StrategyIssueRequest struct {
	Issue IssueRequest
	Plan  StrategyPlanRequest
}

type StrategyIssueResult struct {
	Issue   IssueResult
	Receipt StrategyPlanReceipt
}

type StrategyPlanReceipt struct {
	AttemptID        string
	AccountRef       string
	DecisionIdentity string
	RiskIntentID     string
	ClientOrderID    string
	Quantity         string
	Revision         int
	State            string
	Idempotent       bool
}

type StrategyTrace struct {
	AccountRef         string
	DecisionIdentity   string
	CandidateLifeID    string
	Market             string
	Symbol             string
	ThresholdVersion   string
	ThresholdSetDigest string
	EvidenceDigest     string
	LaneID             string
	LaneVersion        string
	RiskIntentID       string
	AttemptID          string
	ExecutionKind      string
	ExternalRef        string
}

// planStrategyEntryForTest exercises exact replay/collision behavior without
// exposing a production decision-bypass API outside package journal.
func (j *Journal) planStrategyEntryForTest(ctx context.Context, plan StrategyAtomicPlan) (StrategyPlanReceipt, error) {
	if j == nil || j.db == nil {
		return StrategyPlanReceipt{}, errors.New("journal strategy plan: journal required")
	}
	decision, err := plan.RiskDecision.build()
	if err != nil {
		return StrategyPlanReceipt{}, err
	}
	if decision.SafetyClass != SafetyClassExposureRaising || decision.PreimageKind != PreimageKindRiskIntent {
		return StrategyPlanReceipt{}, errors.New("journal strategy plan: canonical RiskIntent required")
	}
	if plan.CreatedAt.IsZero() {
		plan.CreatedAt = j.clk.Now().UTC()
	} else {
		plan.CreatedAt = plan.CreatedAt.UTC()
	}
	plan.Lineage = normalizeStrategyDecision(plan.Lineage, plan.CreatedAt)
	lineage := plan.Lineage
	if !completeStrategyLineage(lineage) ||
		lineage.ActivationManifestDigest != plan.ActivationManifestDigest ||
		strings.TrimSpace(plan.AttemptID) == "" || strings.TrimSpace(plan.GuardianDecisionID) == "" ||
		strings.TrimSpace(plan.ActivationManifestDigest) == "" || plan.Revision < 1 {
		return StrategyPlanReceipt{}, errors.New("journal strategy plan: complete exact binding required")
	}

	tx, err := j.db.BeginTx(ctx, nil)
	if err != nil {
		return StrategyPlanReceipt{}, err
	}
	defer tx.Rollback()
	inserted := 0
	if added, err := insertExactRiskDecision(ctx, tx, decision); err != nil {
		return StrategyPlanReceipt{}, err
	} else {
		inserted += added
	}
	if added, err := insertExactStrategyDecision(ctx, tx, lineage); err != nil {
		return StrategyPlanReceipt{}, err
	} else {
		inserted += added
	}
	if added, err := insertExactStrategyAttempt(ctx, tx, plan, decision.ID, decision.AccountRef); err != nil {
		return StrategyPlanReceipt{}, err
	} else {
		inserted += added
	}
	if added, err := insertExactStrategyExecution(ctx, tx, decision.AccountRef, plan.AttemptID, "DISPATCH_START", plan.ClientOrderID, plan.CreatedAt); err != nil {
		return StrategyPlanReceipt{}, err
	} else {
		inserted += added
	}
	if err := tx.Commit(); err != nil {
		return StrategyPlanReceipt{}, err
	}
	return StrategyPlanReceipt{
		AttemptID: plan.AttemptID, AccountRef: decision.AccountRef, DecisionIdentity: lineage.DecisionIdentity,
		RiskIntentID: decision.ID, ClientOrderID: decision.ClientOrderID, Quantity: lineage.Quantity,
		Revision: plan.Revision, State: "PLANNED", Idempotent: inserted == 0,
	}, nil
}

func (j *Journal) RecordStrategyDecisionAndReserve(ctx context.Context, request StrategyIssueRequest) (StrategyIssueResult, error) {
	if j == nil || j.db == nil {
		return StrategyIssueResult{}, errors.New("journal strategy issuance: journal required")
	}
	decision, reserve, reservePlan, err := request.Issue.build()
	if err != nil {
		return StrategyIssueResult{}, err
	}
	if decision.SafetyClass != SafetyClassExposureRaising || decision.PreimageKind != PreimageKindRiskIntent {
		return StrategyIssueResult{}, errors.New("journal strategy issuance: canonical RiskIntent required")
	}
	plan := StrategyAtomicPlan{
		RiskDecision: request.Issue.Decision, Lineage: request.Plan.Lineage,
		AttemptID: request.Plan.AttemptID, GuardianDecisionID: decision.ID,
		ActivationManifestDigest: request.Plan.ActivationManifestDigest,
		ClientOrderID:            decision.ClientOrderID, Revision: request.Plan.Revision, CreatedAt: request.Plan.CreatedAt,
	}
	if plan.CreatedAt.IsZero() {
		plan.CreatedAt = j.clk.Now().UTC()
	} else {
		plan.CreatedAt = plan.CreatedAt.UTC()
	}
	plan.Lineage = normalizeStrategyDecision(plan.Lineage, plan.CreatedAt)
	if !completeStrategyLineage(plan.Lineage) ||
		plan.Lineage.ActivationManifestDigest != plan.ActivationManifestDigest ||
		strings.TrimSpace(plan.AttemptID) == "" || strings.TrimSpace(plan.ActivationManifestDigest) == "" || plan.Revision < 1 {
		return StrategyIssueResult{}, errors.New("journal strategy issuance: complete exact binding required")
	}
	if err := verifyStrategyRiskBinding(decision, plan.Lineage); err != nil {
		return StrategyIssueResult{}, err
	}

	tx, err := j.db.BeginTx(ctx, nil)
	if err != nil {
		return StrategyIssueResult{}, err
	}
	defer tx.Rollback()
	if err := reservePrecheck(ctx, tx, reserve, reservePlan, j.clk.Now().UTC()); err != nil {
		return StrategyIssueResult{}, err
	}
	if err := insertDecisionRow(ctx, tx, decision); err != nil {
		return StrategyIssueResult{}, err
	}
	reserved, err := reserveRows(ctx, tx, reserve, reservePlan, j.clk.Now().UTC())
	if err != nil {
		return StrategyIssueResult{}, err
	}
	if _, err := insertExactStrategyDecision(ctx, tx, plan.Lineage); err != nil {
		return StrategyIssueResult{}, err
	}
	if _, err := insertExactStrategyAttempt(ctx, tx, plan, decision.ID, decision.AccountRef); err != nil {
		return StrategyIssueResult{}, err
	}
	if _, err := insertExactStrategyExecution(ctx, tx, decision.AccountRef, plan.AttemptID, "DISPATCH_START", decision.ClientOrderID, plan.CreatedAt); err != nil {
		return StrategyIssueResult{}, err
	}
	if err := tx.Commit(); err != nil {
		return StrategyIssueResult{}, fmt.Errorf("journal: committing strategy issuance %s: %w", decision.ID, err)
	}
	return StrategyIssueResult{
		Issue: IssueResult{Decision: decision, Version: reserved.Version, Reservations: reserved.Reservations},
		Receipt: StrategyPlanReceipt{
			AttemptID: plan.AttemptID, AccountRef: decision.AccountRef, DecisionIdentity: plan.Lineage.DecisionIdentity,
			RiskIntentID: decision.ID, ClientOrderID: decision.ClientOrderID, Quantity: plan.Lineage.Quantity,
			Revision: plan.Revision, State: "PLANNED",
		},
	}, nil
}

type CollectStrategyIssue func(context.Context, int) (StrategyIssueRequest, error)

func (j *Journal) RecordStrategyDecisionAndReserveWithRecollection(ctx context.Context, collect CollectStrategyIssue, policy RecollectPolicy) (StrategyIssueResult, error) {
	if collect == nil {
		return StrategyIssueResult{}, fmt.Errorf("%w: strategy issuing needs a snapshot collector", ErrInvalidRequest)
	}
	return recollectLoop(j.clk, policy, func(attempt int) (StrategyIssueResult, error) {
		request, err := collect(ctx, attempt)
		if err != nil {
			return StrategyIssueResult{}, fmt.Errorf("journal: collecting strategy broker snapshot: %w", err)
		}
		return j.RecordStrategyDecisionAndReserve(ctx, request)
	})
}

func insertExactRiskDecision(ctx context.Context, tx *sql.Tx, decision Decision) (int, error) {
	result, err := tx.ExecContext(ctx, `INSERT OR IGNORE INTO decisions(id,account_ref,generation,safety_class,preimage_kind,risk_preimage,risk_hash,client_order_id,limits_json,nonce,issued_at,expires_at) VALUES(?,?,?,?,?,?,?,?,?,?,?,?)`,
		decision.ID, decision.AccountRef, decision.Generation, decision.SafetyClass, decision.PreimageKind,
		decision.RiskPreimage, decision.RiskHash, nullableString(decision.ClientOrderID), nullableString(decision.LimitsJSON),
		decision.Nonce, formatJournalTime(decision.IssuedAt), formatJournalTime(decision.ExpiresAt))
	if err != nil {
		return 0, err
	}
	var account, class, kind, preimage, hash, client, limits, nonce, issued, expires string
	var generation int
	err = tx.QueryRowContext(ctx, `SELECT account_ref,generation,safety_class,preimage_kind,risk_preimage,risk_hash,COALESCE(client_order_id,''),COALESCE(limits_json,''),nonce,issued_at,expires_at FROM decisions WHERE id=?`, decision.ID).
		Scan(&account, &generation, &class, &kind, &preimage, &hash, &client, &limits, &nonce, &issued, &expires)
	if err != nil || account != decision.AccountRef || generation != decision.Generation || class != decision.SafetyClass ||
		kind != decision.PreimageKind || preimage != decision.RiskPreimage || hash != decision.RiskHash ||
		client != decision.ClientOrderID || limits != decision.LimitsJSON || nonce != decision.Nonce ||
		issued != formatJournalTime(decision.IssuedAt) || expires != formatJournalTime(decision.ExpiresAt) {
		return 0, &StrategyCollisionError{Stage: "RiskIntent"}
	}
	rows, _ := result.RowsAffected()
	return int(rows), nil
}

func insertExactStrategyDecision(ctx context.Context, tx *sql.Tx, lineage StrategyDecisionLineage) (int, error) {
	created := lineage.CreatedAt.Format(time.RFC3339Nano)
	result, err := tx.ExecContext(ctx, `INSERT OR IGNORE INTO strategy_decision_lineage(entry_decision_identity,candidate_life_id,market,symbol,threshold_version,threshold_set_digest,evidence_digest,lane_id,lane_version,lane_source_digest,lane_constants_digest,entry_price,stop_price,target_price,quantity,policy_version,settings_digest,decision_payload,decision_payload_digest,activation_manifest_digest,created_at) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		lineage.DecisionIdentity, lineage.CandidateLifeID, lineage.Market, lineage.Symbol, lineage.ThresholdVersion,
		lineage.ThresholdSetDigest, lineage.EvidenceDigest, lineage.LaneID, lineage.LaneVersion, lineage.LaneSourceDigest,
		lineage.LaneConstantsDigest, lineage.EntryPrice, lineage.StopPrice, lineage.TargetPrice, lineage.Quantity, lineage.PolicyVersion, lineage.SettingsDigest, lineage.DecisionPayload, lineage.DecisionPayloadDigest,
		lineage.ActivationManifestDigest, created)
	if err != nil {
		return 0, err
	}
	var got StrategyDecisionLineage
	var gotCreated string
	err = tx.QueryRowContext(ctx, `SELECT candidate_life_id,market,symbol,threshold_version,threshold_set_digest,evidence_digest,lane_id,lane_version,lane_source_digest,lane_constants_digest,entry_price,stop_price,target_price,quantity,policy_version,settings_digest,decision_payload,decision_payload_digest,activation_manifest_digest,created_at FROM strategy_decision_lineage WHERE entry_decision_identity=?`, lineage.DecisionIdentity).
		Scan(&got.CandidateLifeID, &got.Market, &got.Symbol, &got.ThresholdVersion, &got.ThresholdSetDigest,
			&got.EvidenceDigest, &got.LaneID, &got.LaneVersion, &got.LaneSourceDigest, &got.LaneConstantsDigest,
			&got.EntryPrice, &got.StopPrice, &got.TargetPrice, &got.Quantity, &got.PolicyVersion, &got.SettingsDigest, &got.DecisionPayload, &got.DecisionPayloadDigest, &got.ActivationManifestDigest, &gotCreated)
	if err != nil || got.CandidateLifeID != lineage.CandidateLifeID || got.Market != lineage.Market ||
		got.Symbol != lineage.Symbol || got.ThresholdVersion != lineage.ThresholdVersion ||
		got.ThresholdSetDigest != lineage.ThresholdSetDigest || got.EvidenceDigest != lineage.EvidenceDigest ||
		got.LaneID != lineage.LaneID || got.LaneVersion != lineage.LaneVersion ||
		got.LaneSourceDigest != lineage.LaneSourceDigest || got.LaneConstantsDigest != lineage.LaneConstantsDigest ||
		got.EntryPrice != lineage.EntryPrice || got.StopPrice != lineage.StopPrice || got.TargetPrice != lineage.TargetPrice ||
		got.Quantity != lineage.Quantity || got.PolicyVersion != lineage.PolicyVersion || got.SettingsDigest != lineage.SettingsDigest ||
		got.DecisionPayload != lineage.DecisionPayload || got.DecisionPayloadDigest != lineage.DecisionPayloadDigest || got.ActivationManifestDigest != lineage.ActivationManifestDigest ||
		gotCreated != created {
		return 0, &StrategyCollisionError{Stage: "decision lineage"}
	}
	rows, _ := result.RowsAffected()
	return int(rows), nil
}

func insertExactStrategyAttempt(ctx context.Context, tx *sql.Tx, plan StrategyAtomicPlan, riskIntentID, accountRef string) (int, error) {
	created := plan.CreatedAt.Format(time.RFC3339Nano)
	result, err := tx.ExecContext(ctx, `INSERT OR IGNORE INTO strategy_attempt_lineage(attempt_id,account_ref,entry_decision_identity,risk_intent_id,guardian_decision_id,activation_manifest_digest,client_order_id,revision,state,created_at) VALUES(?,?,?,?,?,?,?,?,?,?)`,
		plan.AttemptID, accountRef, plan.Lineage.DecisionIdentity, riskIntentID, plan.GuardianDecisionID,
		plan.ActivationManifestDigest, plan.ClientOrderID, plan.Revision, "PLANNED", created)
	if err != nil {
		return 0, err
	}
	var gotAccount, gotDecision, gotRisk, gotGuardian, gotManifest, gotClient, gotState, gotCreated string
	var gotRevision int
	err = tx.QueryRowContext(ctx, `SELECT account_ref,entry_decision_identity,risk_intent_id,guardian_decision_id,activation_manifest_digest,client_order_id,revision,state,created_at FROM strategy_attempt_lineage WHERE attempt_id=?`, plan.AttemptID).
		Scan(&gotAccount, &gotDecision, &gotRisk, &gotGuardian, &gotManifest, &gotClient, &gotRevision, &gotState, &gotCreated)
	if err != nil || gotAccount != accountRef || gotDecision != plan.Lineage.DecisionIdentity || gotRisk != riskIntentID ||
		gotGuardian != plan.GuardianDecisionID || gotManifest != plan.ActivationManifestDigest ||
		gotClient != plan.ClientOrderID || gotRevision != plan.Revision || gotState != "PLANNED" || gotCreated != created {
		return 0, &StrategyCollisionError{Stage: "attempt"}
	}
	rows, _ := result.RowsAffected()
	return int(rows), nil
}

func insertExactStrategyExecution(ctx context.Context, tx *sql.Tx, accountRef, attemptID, kind, externalRef string, recordedAt time.Time) (int, error) {
	recorded := recordedAt.UTC().Format(time.RFC3339Nano)
	result, err := tx.ExecContext(ctx, `INSERT OR IGNORE INTO strategy_execution_lineage(account_ref,attempt_id,kind,external_ref,recorded_at) VALUES(?,?,?,?,?)`, accountRef, attemptID, kind, externalRef, recorded)
	if err != nil {
		return 0, err
	}
	var gotAttempt, gotRecorded string
	err = tx.QueryRowContext(ctx, `SELECT attempt_id,recorded_at FROM strategy_execution_lineage WHERE account_ref=? AND kind=? AND external_ref=?`, accountRef, kind, externalRef).Scan(&gotAttempt, &gotRecorded)
	if err != nil || gotAttempt != attemptID || gotRecorded != recorded {
		return 0, &StrategyCollisionError{Stage: "execution"}
	}
	rows, _ := result.RowsAffected()
	return int(rows), nil
}

func (j *Journal) RecordStrategyRefusal(ctx context.Context, receipt StrategyPlanReceipt, reason string) error {
	return j.recordStrategyTerminal(ctx, receipt, "REFUSED", reason)
}

func (j *Journal) RecordStrategyInDoubt(ctx context.Context, receipt StrategyPlanReceipt, reason string) error {
	return j.recordStrategyTerminal(ctx, receipt, "IN_DOUBT", reason)
}

func (j *Journal) RecordStrategyDispatched(ctx context.Context, receipt StrategyPlanReceipt, accountRef, mutationAttemptID, brokerOrderID string) error {
	accountRef = strings.TrimSpace(accountRef)
	mutationAttemptID = strings.TrimSpace(mutationAttemptID)
	brokerOrderID = strings.TrimSpace(brokerOrderID)
	if j == nil || j.db == nil || receipt.AttemptID == "" || receipt.AccountRef != accountRef || receipt.DecisionIdentity == "" || receipt.RiskIntentID == "" || receipt.ClientOrderID == "" || receipt.Quantity == "" || receipt.Revision < 1 ||
		(receipt.State != "PLANNED" && receipt.State != "IN_DOUBT") || accountRef == "" || mutationAttemptID == "" || brokerOrderID == "" {
		return errors.New("journal strategy dispatch: invalid")
	}
	tx, err := j.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	storedState, storedRevision, err := verifyStrategyReceiptBindingTx(ctx, tx, receipt)
	if err != nil {
		return err
	}
	if storedState != "DISPATCHED" || storedRevision != receipt.Revision+1 {
		if storedState != receipt.State || storedRevision != receipt.Revision {
			return errors.New("journal strategy dispatch: stale revision")
		}
		result, err := tx.ExecContext(ctx, `UPDATE strategy_attempt_lineage SET state='DISPATCHED',revision=revision+1 WHERE attempt_id=? AND account_ref=? AND revision=? AND state=?`, receipt.AttemptID, accountRef, receipt.Revision, receipt.State)
		if err != nil {
			return err
		}
		rows, _ := result.RowsAffected()
		if rows != 1 {
			return errors.New("journal strategy dispatch: stale revision")
		}
	}
	now := j.clk.Now().UTC()
	for _, link := range [][2]string{{"MUTATION_ATTEMPT", mutationAttemptID}, {"BROKER_ORDER", brokerOrderID}} {
		result, err := tx.ExecContext(ctx, `INSERT OR IGNORE INTO strategy_execution_lineage(account_ref,attempt_id,kind,external_ref,recorded_at) VALUES(?,?,?,?,?)`, accountRef, receipt.AttemptID, link[0], link[1], now.Format(time.RFC3339Nano))
		if err != nil {
			return err
		}
		var gotAttempt string
		if err := tx.QueryRowContext(ctx, `SELECT attempt_id FROM strategy_execution_lineage WHERE account_ref=? AND kind=? AND external_ref=?`, accountRef, link[0], link[1]).Scan(&gotAttempt); err != nil || gotAttempt != receipt.AttemptID {
			return &StrategyCollisionError{Stage: "dispatch link"}
		}
		_ = result
	}
	return tx.Commit()
}

type StrategyRecoveryError struct{ Detail string }

func (e *StrategyRecoveryError) Error() string { return "journal strategy recovery: " + e.Detail }

// PendingStrategyPlans is the account-scoped startup seam. Runtime wiring is
// intentionally absent in a047; a later approved coordinator can enumerate
// only plans that still need an exact 0/1/>1 mutation-attempt resolution.
func (j *Journal) PendingStrategyPlans(ctx context.Context, accountRef string) ([]StrategyPlanReceipt, error) {
	accountRef = strings.TrimSpace(accountRef)
	if j == nil || j.db == nil || accountRef == "" {
		return nil, errors.New("journal strategy recovery: account required")
	}
	rows, err := j.db.QueryContext(ctx, `SELECT a.attempt_id,a.account_ref,a.entry_decision_identity,a.risk_intent_id,a.client_order_id,d.quantity,a.revision,a.state FROM strategy_attempt_lineage a JOIN strategy_decision_lineage d ON d.entry_decision_identity=a.entry_decision_identity WHERE a.account_ref=? AND a.state IN('PLANNED','IN_DOUBT') ORDER BY a.created_at,a.attempt_id`, accountRef)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var pending []StrategyPlanReceipt
	for rows.Next() {
		var receipt StrategyPlanReceipt
		if err := rows.Scan(&receipt.AttemptID, &receipt.AccountRef, &receipt.DecisionIdentity, &receipt.RiskIntentID, &receipt.ClientOrderID, &receipt.Quantity, &receipt.Revision, &receipt.State); err != nil {
			return nil, err
		}
		pending = append(pending, receipt)
	}
	return pending, rows.Err()
}

func (j *Journal) RecoverStrategyDispatch(ctx context.Context, receipt StrategyPlanReceipt, accountRef string) error {
	accountRef = strings.TrimSpace(accountRef)
	if receipt.AccountRef != accountRef || (receipt.State != "PLANNED" && receipt.State != "IN_DOUBT") {
		return &StrategyRecoveryError{Detail: "receipt/account/state mismatch"}
	}
	rows, err := j.db.QueryContext(ctx, `SELECT a.id,a.state,COALESCE(a.broker_order_id,'') FROM mutation_attempts a JOIN intents i ON i.id=a.intent_id WHERE a.intent_id=? AND i.account_ref=? ORDER BY a.attempt_no DESC LIMIT 2`, receipt.AttemptID, accountRef)
	if err != nil {
		return err
	}
	defer rows.Close()
	type mutation struct{ id, state, broker string }
	var attempts []mutation
	for rows.Next() {
		var value mutation
		if err := rows.Scan(&value.id, &value.state, &value.broker); err != nil {
			return err
		}
		attempts = append(attempts, value)
	}
	if err := rows.Err(); err != nil {
		return err
	}
	if len(attempts) == 0 {
		missing := &StrategyRecoveryError{Detail: "no mutation attempt"}
		if stateErr := j.RecordStrategyRefusal(ctx, receipt, "no_mutation_attempt"); stateErr != nil {
			return errors.Join(missing, stateErr)
		}
		return missing
	}
	if len(attempts) > 1 {
		ambiguous := &StrategyRecoveryError{Detail: "multiple mutation attempts for one strategy decision"}
		if receipt.State == "PLANNED" {
			if stateErr := j.RecordStrategyInDoubt(ctx, receipt, "ambiguous_mutation_attempts"); stateErr != nil {
				return errors.Join(ambiguous, stateErr)
			}
		}
		return ambiguous
	}
	confirmed := attempts[0].state == string(StateConfirmed) && strings.TrimSpace(attempts[0].broker) != ""
	if confirmed {
		return j.RecordStrategyDispatched(ctx, receipt, accountRef, attempts[0].id, attempts[0].broker)
	}
	if attempts[0].state == string(StateNotDispatched) || attempts[0].state == string(StateFailedConfirmed) {
		terminal := &StrategyRecoveryError{Detail: "mutation attempt definitively refused"}
		if stateErr := j.RecordStrategyRefusal(ctx, receipt, "mutation_attempt_refused"); stateErr != nil {
			return errors.Join(terminal, stateErr)
		}
		return terminal
	}
	if receipt.State == "PLANNED" {
		if stateErr := j.RecordStrategyInDoubt(ctx, receipt, "mutation_attempt_requires_recovery"); stateErr != nil {
			return errors.Join(&StrategyRecoveryError{Detail: "mutation attempt is not confirmed"}, stateErr)
		}
	}
	return &StrategyRecoveryError{Detail: "mutation attempt is not confirmed"}
}

func (j *Journal) recordStrategyTerminal(ctx context.Context, receipt StrategyPlanReceipt, state, reason string) error {
	reason = strings.TrimSpace(reason)
	if j == nil || j.db == nil || receipt.AttemptID == "" || receipt.AccountRef == "" || receipt.DecisionIdentity == "" || receipt.RiskIntentID == "" || receipt.ClientOrderID == "" || receipt.Quantity == "" || receipt.Revision < 1 || reason == "" ||
		(state != "REFUSED" && state != "IN_DOUBT") {
		return errors.New("journal strategy terminal: invalid")
	}
	allowedSource := receipt.State == "PLANNED" || (state == "REFUSED" && receipt.State == "IN_DOUBT")
	if !allowedSource {
		return errors.New("journal strategy terminal: invalid source state")
	}
	tx, err := j.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	storedState, storedRevision, err := verifyStrategyReceiptBindingTx(ctx, tx, receipt)
	if err != nil {
		return err
	}
	if storedState != receipt.State || storedRevision != receipt.Revision {
		return errors.New("journal strategy terminal: stale revision")
	}
	result, err := tx.ExecContext(ctx, `UPDATE strategy_attempt_lineage SET state=?,revision=revision+1 WHERE attempt_id=? AND account_ref=? AND revision=? AND state=?`, state, receipt.AttemptID, receipt.AccountRef, receipt.Revision, receipt.State)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows != 1 {
		return errors.New("journal strategy terminal: stale revision")
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO strategy_attempt_refusals(attempt_id,revision,reason_code,recorded_at) VALUES(?,?,?,?)`, receipt.AttemptID, receipt.Revision+1, reason, j.clk.Now().UTC().Format(time.RFC3339Nano))
	if err != nil {
		return err
	}
	return tx.Commit()
}

func verifyStrategyReceiptBindingTx(ctx context.Context, tx *sql.Tx, receipt StrategyPlanReceipt) (string, int, error) {
	var accountRef, decisionIdentity, riskIntentID, clientOrderID, quantity, state string
	var revision int
	err := tx.QueryRowContext(ctx, `SELECT a.account_ref,a.entry_decision_identity,a.risk_intent_id,a.client_order_id,d.quantity,a.state,a.revision FROM strategy_attempt_lineage a JOIN strategy_decision_lineage d ON d.entry_decision_identity=a.entry_decision_identity WHERE a.attempt_id=?`, receipt.AttemptID).
		Scan(&accountRef, &decisionIdentity, &riskIntentID, &clientOrderID, &quantity, &state, &revision)
	if err != nil || accountRef != receipt.AccountRef || decisionIdentity != receipt.DecisionIdentity ||
		riskIntentID != receipt.RiskIntentID || clientOrderID != receipt.ClientOrderID || quantity != receipt.Quantity {
		return "", 0, errors.New("journal strategy receipt: exact binding mismatch")
	}
	return state, revision, nil
}

func (j *Journal) AppendStrategyExecutionLink(ctx context.Context, accountRef, attemptID, kind, externalRef string) error {
	accountRef = strings.TrimSpace(accountRef)
	attemptID = strings.TrimSpace(attemptID)
	externalRef = strings.TrimSpace(externalRef)
	if j == nil || j.db == nil || accountRef == "" || attemptID == "" || externalRef == "" ||
		(kind != "MUTATION_ATTEMPT" && kind != "BROKER_ORDER" && kind != "FILL" && kind != "POSITION" && kind != "CLOSE_OUTCOME") {
		return errors.New("journal strategy execution: invalid")
	}
	tx, err := j.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var storedAccount string
	if err := tx.QueryRowContext(ctx, `SELECT account_ref FROM strategy_attempt_lineage WHERE attempt_id=?`, attemptID).Scan(&storedAccount); err != nil || storedAccount != accountRef {
		return errors.New("journal strategy execution: account/attempt mismatch")
	}
	if _, err := insertExactStrategyExecution(ctx, tx, accountRef, attemptID, kind, externalRef, j.clk.Now().UTC()); err != nil {
		return err
	}
	return tx.Commit()
}

func (j *Journal) LookupStrategyTrace(ctx context.Context, accountRef, kind, externalRef string) (StrategyTrace, error) {
	var trace StrategyTrace
	if strings.TrimSpace(accountRef) == "" || strings.TrimSpace(kind) == "" || strings.TrimSpace(externalRef) == "" {
		return trace, ErrStrategyTraceNotFound
	}
	err := j.db.QueryRowContext(ctx, `SELECT e.account_ref,d.entry_decision_identity,d.candidate_life_id,d.market,d.symbol,d.threshold_version,d.threshold_set_digest,d.evidence_digest,d.lane_id,d.lane_version,a.risk_intent_id,a.attempt_id,e.kind,e.external_ref FROM strategy_execution_lineage e INDEXED BY idx_strategy_execution_reverse JOIN strategy_attempt_lineage a ON a.attempt_id=e.attempt_id JOIN strategy_decision_lineage d ON d.entry_decision_identity=a.entry_decision_identity WHERE e.account_ref=? AND e.kind=? AND e.external_ref=?`, accountRef, kind, externalRef).
		Scan(&trace.AccountRef, &trace.DecisionIdentity, &trace.CandidateLifeID, &trace.Market, &trace.Symbol,
			&trace.ThresholdVersion, &trace.ThresholdSetDigest, &trace.EvidenceDigest, &trace.LaneID,
			&trace.LaneVersion, &trace.RiskIntentID, &trace.AttemptID, &trace.ExecutionKind, &trace.ExternalRef)
	if errors.Is(err, sql.ErrNoRows) {
		return StrategyTrace{}, ErrStrategyTraceNotFound
	}
	if err != nil {
		return StrategyTrace{}, fmt.Errorf("journal strategy trace: %w", err)
	}
	return trace, nil
}

func completeStrategyLineage(lineage StrategyDecisionLineage) bool {
	values := []string{
		lineage.DecisionIdentity, lineage.CandidateLifeID, lineage.Market, lineage.Symbol,
		lineage.ThresholdVersion, lineage.ThresholdSetDigest, lineage.EvidenceDigest,
		lineage.LaneID, lineage.LaneVersion, lineage.LaneSourceDigest, lineage.LaneConstantsDigest,
		lineage.EntryPrice, lineage.StopPrice, lineage.TargetPrice, lineage.Quantity, lineage.PolicyVersion, lineage.SettingsDigest, lineage.DecisionPayload, lineage.DecisionPayloadDigest,
		lineage.ActivationManifestDigest,
	}
	for _, value := range values {
		if strings.TrimSpace(value) == "" {
			return false
		}
	}
	return true
}

func verifyStrategyRiskBinding(decision Decision, lineage StrategyDecisionLineage) error {
	if journalHash := HashPreimage(decision.RiskPreimage); journalHash != decision.RiskHash {
		return errors.New("journal strategy issuance: RiskIntent hash mismatch")
	}
	parsed, err := ParsePreimage(decision.PreimageKind, decision.RiskPreimage)
	if err != nil {
		return fmt.Errorf("journal strategy issuance: parsing canonical RiskIntent: %w", err)
	}
	riskRecord, ok := parsed.(RiskIntent)
	if !ok {
		return errors.New("journal strategy issuance: preimage is not RiskIntent")
	}
	var signalRecord struct {
		Identity    string
		Market      string
		Symbol      string
		EntryPrice  string
		StopPrice   string
		TargetPrice string
	}
	if err := json.Unmarshal([]byte(lineage.DecisionPayload), &signalRecord); err != nil {
		return fmt.Errorf("journal strategy issuance: decoding decision payload: %w", err)
	}
	payloadHash := sha256.Sum256([]byte(lineage.DecisionPayload))
	wantPayloadDigest := "sha256:" + hex.EncodeToString(payloadHash[:])
	if lineage.DecisionPayloadDigest != wantPayloadDigest || signalRecord.Identity != lineage.DecisionIdentity ||
		!strings.EqualFold(riskRecord.Market, lineage.Market) || !strings.EqualFold(signalRecord.Market, lineage.Market) ||
		riskRecord.Symbol != lineage.Symbol || signalRecord.Symbol != lineage.Symbol ||
		riskRecord.Quantity != lineage.Quantity || riskRecord.EntryPrice != lineage.EntryPrice || signalRecord.EntryPrice != lineage.EntryPrice ||
		riskRecord.StopPrice != lineage.StopPrice || signalRecord.StopPrice != lineage.StopPrice ||
		riskRecord.TargetPrice != lineage.TargetPrice || signalRecord.TargetPrice != lineage.TargetPrice ||
		riskRecord.PolicyVersion != lineage.PolicyVersion {
		return errors.New("journal strategy issuance: RiskIntent/decision exact binding mismatch")
	}
	return nil
}

func normalizeStrategyDecision(value StrategyDecisionLineage, now time.Time) StrategyDecisionLineage {
	value.DecisionIdentity = strings.TrimSpace(value.DecisionIdentity)
	value.CandidateLifeID = strings.TrimSpace(value.CandidateLifeID)
	value.Market = strings.TrimSpace(value.Market)
	value.Symbol = strings.TrimSpace(value.Symbol)
	value.ThresholdVersion = strings.TrimSpace(value.ThresholdVersion)
	value.ThresholdSetDigest = strings.TrimSpace(value.ThresholdSetDigest)
	value.EvidenceDigest = strings.TrimSpace(value.EvidenceDigest)
	value.LaneID = strings.TrimSpace(value.LaneID)
	value.LaneVersion = strings.TrimSpace(value.LaneVersion)
	value.LaneSourceDigest = strings.TrimSpace(value.LaneSourceDigest)
	value.LaneConstantsDigest = strings.TrimSpace(value.LaneConstantsDigest)
	value.EntryPrice = strings.TrimSpace(value.EntryPrice)
	value.StopPrice = strings.TrimSpace(value.StopPrice)
	value.TargetPrice = strings.TrimSpace(value.TargetPrice)
	value.Quantity = strings.TrimSpace(value.Quantity)
	value.PolicyVersion = strings.TrimSpace(value.PolicyVersion)
	value.SettingsDigest = strings.TrimSpace(value.SettingsDigest)
	value.DecisionPayload = strings.TrimSpace(value.DecisionPayload)
	value.DecisionPayloadDigest = strings.TrimSpace(value.DecisionPayloadDigest)
	value.ActivationManifestDigest = strings.TrimSpace(value.ActivationManifestDigest)
	if value.CreatedAt.IsZero() {
		value.CreatedAt = now
	} else {
		value.CreatedAt = value.CreatedAt.UTC()
	}
	return value
}
