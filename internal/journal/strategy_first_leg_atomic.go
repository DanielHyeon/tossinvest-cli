package journal

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/JungHoonGhae/tossinvest-cli/internal/positioncampaign"
	"github.com/JungHoonGhae/tossinvest-cli/internal/riskbucket"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyrouter"
)

// FirstLegCampaignRequest carries only stable caller identities. The
// prospective token is intentionally absent and is minted inside the journal's
// BEGIN IMMEDIATE transaction.
type FirstLegCampaignRequest struct {
	CampaignID                 string
	ExpectedPositionGeneration int64
	ExpectedPositionVersion    int64
	CreateCommandKey           string
	FirstLegCommandKey         string
	FirstLegPlanID             string
}

type QFinalCampaignFirstLegRequest struct {
	Issue         QFinalIssueRequest
	Strategy      StrategyPlanRequest
	Campaign      FirstLegCampaignRequest
	RouterID      string
	RouterVersion string
	Weekly        *WeeklyFirstLegReservationBinding
}

type WeeklyFirstLegReservationBinding struct {
	ReservationID, StableWeek          string
	PlannedOrdinal                     int
	ScopeVersion                       uint64
	RequestDigest, RecordDigest        string
	CalendarGeneration, CalendarDigest string
}

type CollectQFinalCampaignFirstLeg func(context.Context, int) (QFinalCampaignFirstLegRequest, error)

// QFinalCampaignFirstLegReceipt is deliberately opaque: it does not expose the
// journal-minted token and cannot be converted into a dispatch lease.
type QFinalCampaignFirstLegReceipt struct {
	DecisionID             string
	AggregateReservationID string
	BucketReservationIDs   []string
	CampaignID             string
	AttemptID              string
	FirstLegPlanID         string
	LegSequence            int64
	Market                 string
	Symbol                 string
	RouterID               string
	RouterVersion          string
	QFinal                 uint64
	Idempotent             bool
}

type preparedQFinalFirstLeg struct {
	decision          Decision
	reserve           ReserveRequest
	reservePlan       reservePlan
	riskDecision      riskbucket.AdmissionDecision
	issuePreimage     string
	issueDigest       string
	strategyPlan      StrategyAtomicPlan
	campaign          FirstLegCampaignRequest
	campaignMarket    string
	campaignSymbol    string
	routerID          string
	routerVersion     string
	firstLegDigest    string
	bindingRecordHash string
	weekly            *WeeklyFirstLegReservationBinding
}

func (j *Journal) RecordQFinalCampaignFirstLegWithRecollection(ctx context.Context, collect CollectQFinalCampaignFirstLeg, policy RecollectPolicy) (QFinalCampaignFirstLegReceipt, error) {
	if collect == nil {
		return QFinalCampaignFirstLegReceipt{}, fmt.Errorf("%w: first-leg admission needs a snapshot collector", ErrInvalidRequest)
	}
	return recollectLoop(j.clk, policy, func(attempt int) (QFinalCampaignFirstLegReceipt, error) {
		request, err := collect(ctx, attempt)
		if err != nil {
			return QFinalCampaignFirstLegReceipt{}, fmt.Errorf("journal: collecting first-leg authority: %w", err)
		}
		return j.RecordQFinalCampaignFirstLeg(ctx, request)
	})
}

func (j *Journal) RecordQFinalCampaignFirstLeg(ctx context.Context, request QFinalCampaignFirstLegRequest) (QFinalCampaignFirstLegReceipt, error) {
	if j == nil || j.db == nil {
		return QFinalCampaignFirstLegReceipt{}, errors.New("journal first-leg admission: journal required")
	}
	if strings.TrimSpace(request.Issue.Admission.Owner.Key.ProspectiveGeneration) != "" {
		return QFinalCampaignFirstLegReceipt{}, fmt.Errorf("%w: first-leg prospective token is journal-owned", ErrInvalidRequest)
	}
	if err := validateFirstLegCampaignRequest(request.Campaign); err != nil {
		return QFinalCampaignFirstLegReceipt{}, err
	}
	if request.RouterID != strategyrouter.RouterID || request.RouterVersion != strategyrouter.RouterRelease {
		return QFinalCampaignFirstLegReceipt{}, fmt.Errorf("%w: first-leg router identity is not the sealed production router", ErrInvalidRequest)
	}

	tx, err := j.db.BeginTx(ctx, nil)
	if err != nil {
		return QFinalCampaignFirstLegReceipt{}, fmt.Errorf("journal: begin first-leg admission: %w", err)
	}
	defer tx.Rollback()

	token, replayCandidate, err := firstLegReplayTokenTx(ctx, tx, request)
	if err != nil {
		return QFinalCampaignFirstLegReceipt{}, err
	}
	if !replayCandidate {
		token, err = j.mintFirstLegToken()
		if err != nil {
			return QFinalCampaignFirstLegReceipt{}, err
		}
	}
	request.Issue.Admission.Owner.Key.ProspectiveGeneration = token
	prepared, err := j.prepareQFinalCampaignFirstLeg(request)
	if err != nil {
		return QFinalCampaignFirstLegReceipt{}, err
	}
	if err := validateWeeklyFirstLegBindingTx(ctx, tx, prepared); err != nil {
		return QFinalCampaignFirstLegReceipt{}, err
	}

	if replayCandidate {
		replayed, qFinal, err := recoverQFinalIssueReplayTx(ctx, tx, request.Issue.Admission, prepared.decision,
			prepared.reserve, prepared.reservePlan, prepared.riskDecision, prepared.issueDigest)
		if err != nil {
			return QFinalCampaignFirstLegReceipt{}, err
		}
		if !replayed {
			return QFinalCampaignFirstLegReceipt{}, fmt.Errorf("%w: first-leg binding without q_final authority", ErrRiskBucketReplayMismatch)
		}
		if err := verifyFirstLegReplayTx(ctx, tx, prepared, token); err != nil {
			return QFinalCampaignFirstLegReceipt{}, err
		}
		if err := tx.Commit(); err != nil {
			return QFinalCampaignFirstLegReceipt{}, fmt.Errorf("journal: commit first-leg replay: %w", err)
		}
		return firstLegReceipt(prepared, qFinal.Admission, true), nil
	}

	if replayed, _, err := recoverQFinalIssueReplayTx(ctx, tx, request.Issue.Admission, prepared.decision,
		prepared.reserve, prepared.reservePlan, prepared.riskDecision, prepared.issueDigest); err != nil {
		return QFinalCampaignFirstLegReceipt{}, err
	} else if replayed {
		return QFinalCampaignFirstLegReceipt{}, fmt.Errorf("%w: q_final authority exists without first-leg binding", ErrRiskBucketReplayMismatch)
	}
	if err := reservePrecheck(ctx, tx, prepared.reserve, prepared.reservePlan, j.clk.Now().UTC()); err != nil {
		return QFinalCampaignFirstLegReceipt{}, err
	}
	if err := insertDecisionRow(ctx, tx, prepared.decision); err != nil {
		return QFinalCampaignFirstLegReceipt{}, err
	}
	reserved, err := reserveRows(ctx, tx, prepared.reserve, prepared.reservePlan, j.clk.Now().UTC())
	if err != nil {
		return QFinalCampaignFirstLegReceipt{}, err
	}
	qFinalReceipt, err := commitFreshRiskBucketAdmissionTx(ctx, tx, request.Issue.Admission, prepared.riskDecision,
		prepared.issuePreimage, prepared.issueDigest, reserved.Version)
	if err != nil {
		return QFinalCampaignFirstLegReceipt{}, err
	}
	if _, err := insertExactStrategyDecision(ctx, tx, prepared.strategyPlan.Lineage); err != nil {
		return QFinalCampaignFirstLegReceipt{}, err
	}
	if _, err := insertExactStrategyAttempt(ctx, tx, prepared.strategyPlan, prepared.decision.ID, prepared.decision.AccountRef); err != nil {
		return QFinalCampaignFirstLegReceipt{}, err
	}
	if err := insertFirstLegCampaignTx(ctx, tx, prepared, token, j.nowString()); err != nil {
		return QFinalCampaignFirstLegReceipt{}, err
	}
	if err := insertFirstLegBindingTx(ctx, tx, prepared, token, j.nowString()); err != nil {
		return QFinalCampaignFirstLegReceipt{}, err
	}
	if err := tx.Commit(); err != nil {
		return QFinalCampaignFirstLegReceipt{}, fmt.Errorf("journal: commit first-leg admission: %w", err)
	}
	return firstLegReceipt(prepared, qFinalReceipt, false), nil
}

func validateFirstLegCampaignRequest(request FirstLegCampaignRequest) error {
	for field, value := range map[string]string{
		"campaign": request.CampaignID, "create command": request.CreateCommandKey,
		"first-leg command": request.FirstLegCommandKey, "first-leg plan": request.FirstLegPlanID,
	} {
		if !validStrategyDispatchIdentity(value) {
			return fmt.Errorf("%w: first-leg %s is empty or noncanonical", ErrInvalidRequest, field)
		}
	}
	if request.ExpectedPositionGeneration < 0 || request.ExpectedPositionVersion < 0 {
		return fmt.Errorf("%w: negative first-leg position generation/version", ErrInvalidRequest)
	}
	return nil
}

func (j *Journal) mintFirstLegToken() (string, error) {
	source := j.firstLegEntropy
	if source == nil {
		source = rand.Reader
	}
	raw := make([]byte, 32)
	if _, err := io.ReadFull(source, raw); err != nil {
		return "", fmt.Errorf("journal: mint first-leg prospective token: %w", err)
	}
	return hex.EncodeToString(raw), nil
}

func firstLegReplayTokenTx(ctx context.Context, tx *sql.Tx, request QFinalCampaignFirstLegRequest) (string, bool, error) {
	rows, err := tx.QueryContext(ctx, `SELECT prospective_token,decision_id,campaign_id,attempt_id,entry_decision_identity
		FROM strategy_first_leg_bindings WHERE decision_id=? OR campaign_id=? OR attempt_id=? OR entry_decision_identity=?`,
		strings.TrimSpace(request.Issue.Admission.DecisionID), request.Campaign.CampaignID,
		strings.TrimSpace(request.Strategy.AttemptID), strings.TrimSpace(request.Strategy.Lineage.DecisionIdentity))
	if err != nil {
		return "", false, err
	}
	defer rows.Close()
	var token, decisionID, campaignID, attemptID, entryDecision string
	count := 0
	for rows.Next() {
		count++
		if err := rows.Scan(&token, &decisionID, &campaignID, &attemptID, &entryDecision); err != nil {
			return "", false, err
		}
	}
	if err := rows.Err(); err != nil {
		return "", false, err
	}
	if count == 0 {
		return "", false, nil
	}
	if count != 1 || decisionID != strings.TrimSpace(request.Issue.Admission.DecisionID) ||
		campaignID != request.Campaign.CampaignID || attemptID != strings.TrimSpace(request.Strategy.AttemptID) ||
		entryDecision != strings.TrimSpace(request.Strategy.Lineage.DecisionIdentity) || len(token) != 64 {
		return "", false, fmt.Errorf("%w: divergent first-leg replay identity", ErrRiskBucketReplayMismatch)
	}
	return token, true, nil
}

func (j *Journal) prepareQFinalCampaignFirstLeg(request QFinalCampaignFirstLegRequest) (preparedQFinalFirstLeg, error) {
	decision, reserve, reservePlan, err := request.Issue.Issue.build()
	if err != nil {
		return preparedQFinalFirstLeg{}, err
	}
	if decision.SafetyClass != SafetyClassExposureRaising || decision.PreimageKind != PreimageKindRiskIntent {
		return preparedQFinalFirstLeg{}, errors.New("journal first-leg admission: canonical RiskIntent required")
	}
	preimage, err := ParsePreimage(decision.PreimageKind, decision.RiskPreimage)
	if err != nil {
		return preparedQFinalFirstLeg{}, err
	}
	intent, ok := preimage.(RiskIntent)
	if !ok {
		return preparedQFinalFirstLeg{}, errors.New("journal first-leg admission: RiskIntent required")
	}
	_, transactionID, marked := splitQFinalPolicyVersion(intent.PolicyVersion)
	if !marked || transactionID != request.Issue.Admission.TransactionID {
		return preparedQFinalFirstLeg{}, fmt.Errorf("%w: q_final policy marker", ErrRiskBucketSnapshotMismatch)
	}
	riskDecision := riskbucket.CalculateAdmission(request.Issue.Admission.Admission)
	if riskDecision.Refusal != nil {
		return preparedQFinalFirstLeg{}, riskDecision.Refusal
	}
	if request.Issue.Admission.DecisionID != decision.ID || request.Issue.Admission.ExistingReservationID == "" ||
		request.Issue.Admission.ExistingReservationID != soleReservationID(reserve.Reservations) {
		return preparedQFinalFirstLeg{}, fmt.Errorf("%w: q_final decision/reservation binding", ErrRiskBucketSnapshotMismatch)
	}
	quantity, ok := NormalizeDecimal(intent.Quantity)
	if !ok || quantity != strconv.FormatUint(riskDecision.QFinal, 10) {
		return preparedQFinalFirstLeg{}, fmt.Errorf("%w: decision quantity is not q_final", ErrRiskBucketSnapshotMismatch)
	}
	if err := validateRiskBucketAdmission(request.Issue.Admission, riskDecision); err != nil {
		return preparedQFinalFirstLeg{}, err
	}
	admissionPreimage, admissionDigest, err := riskBucketAdmissionDigest(request.Issue.Admission, riskDecision)
	if err != nil {
		return preparedQFinalFirstLeg{}, err
	}
	issuePreimage, issueDigest, err := qFinalIssueDigest(request.Issue.Admission, decision, reserve, reservePlan, admissionPreimage, admissionDigest)
	if err != nil {
		return preparedQFinalFirstLeg{}, err
	}

	created := request.Strategy.CreatedAt.UTC()
	admissionCreated := request.Issue.Admission.CreatedAt.UTC()
	if request.Strategy.CreatedAt.IsZero() || request.Issue.Admission.CreatedAt.IsZero() || !created.Equal(admissionCreated) ||
		created.Before(decision.IssuedAt) || created.After(decision.ExpiresAt) {
		return preparedQFinalFirstLeg{}, fmt.Errorf("%w: first-leg strategy/q_final time binding", ErrRiskBucketSnapshotMismatch)
	}
	lineage := normalizeStrategyDecision(request.Strategy.Lineage, created)
	plan := StrategyAtomicPlan{RiskDecision: request.Issue.Issue.Decision, Lineage: lineage,
		AttemptID: request.Strategy.AttemptID, GuardianDecisionID: decision.ID,
		ActivationManifestDigest: request.Strategy.ActivationManifestDigest,
		ClientOrderID:            decision.ClientOrderID, Revision: request.Strategy.Revision, CreatedAt: created}
	if !completeStrategyLineage(lineage) || lineage.ActivationManifestDigest != plan.ActivationManifestDigest ||
		!validStrategyDispatchIdentity(plan.AttemptID) ||
		strings.TrimSpace(plan.ActivationManifestDigest) == "" || plan.Revision < 1 {
		return preparedQFinalFirstLeg{}, errors.New("journal first-leg admission: complete exact strategy binding required")
	}
	if err := verifyStrategyRiskBinding(decision, lineage); err != nil {
		return preparedQFinalFirstLeg{}, err
	}
	owner := request.Issue.Admission.Owner
	market := strings.ToUpper(strings.TrimSpace(lineage.Market))
	symbol := normaliseSymbol(lineage.Symbol)
	if market != string(owner.Key.Market) || symbol != owner.Key.Symbol || decision.AccountRef != owner.Key.AccountID ||
		lineage.LaneID != owner.LaneID || request.Campaign.CampaignID != owner.CampaignID ||
		lineage.Quantity != quantity || (market != "KR" && market != "US") {
		return preparedQFinalFirstLeg{}, fmt.Errorf("%w: first-leg cross-family scope", ErrRiskBucketSnapshotMismatch)
	}
	if err := (positioncampaign.PositionCampaign{ID: request.Campaign.CampaignID, AccountRef: decision.AccountRef,
		Market: normaliseMarket(market), Symbol: symbol, LaneID: lineage.LaneID, LaneVersion: lineage.LaneVersion,
		DecisionID: decision.ID, EvidenceDigest: lineage.EvidenceDigest,
		ProspectiveToken:           owner.Key.ProspectiveGeneration,
		ExpectedPositionGeneration: request.Campaign.ExpectedPositionGeneration}).Validate(); err != nil {
		return preparedQFinalFirstLeg{}, err
	}
	if err := validateWeeklyFirstLegBindingShape(lineage.LaneID, market, request.Campaign.CampaignID, request.Weekly); err != nil {
		return preparedQFinalFirstLeg{}, err
	}
	requestDigest, err := firstLegRequestDigest(issueDigest, plan, request.Campaign, request.RouterID, request.RouterVersion, request.Weekly)
	if err != nil {
		return preparedQFinalFirstLeg{}, err
	}
	recordDigest := digestParts("strategy-first-leg-binding:v1", decision.ID, request.Issue.Admission.ExistingReservationID,
		lineage.DecisionIdentity, plan.AttemptID, request.Campaign.CampaignID, request.Campaign.FirstLegPlanID,
		decision.AccountRef, market, symbol, lineage.CandidateLifeID, lineage.EvidenceDigest, lineage.LaneID,
		lineage.LaneVersion, request.RouterID, request.RouterVersion, owner.Key.ProspectiveGeneration, quantity, requestDigest)
	return preparedQFinalFirstLeg{decision: decision, reserve: reserve, reservePlan: reservePlan,
		riskDecision: riskDecision, issuePreimage: issuePreimage, issueDigest: issueDigest,
		strategyPlan: plan, campaign: request.Campaign, campaignMarket: normaliseMarket(market),
		campaignSymbol: symbol, routerID: request.RouterID, routerVersion: request.RouterVersion,
		firstLegDigest: requestDigest, bindingRecordHash: recordDigest, weekly: cloneWeeklyFirstLegBinding(request.Weekly)}, nil
}

func firstLegRequestDigest(issueDigest string, plan StrategyAtomicPlan, campaign FirstLegCampaignRequest, routerID, routerVersion string, weekly *WeeklyFirstLegReservationBinding) (string, error) {
	preimage := struct {
		Version       string
		IssueDigest   string
		Lineage       StrategyDecisionLineage
		AttemptID     string
		Manifest      string
		Revision      int
		CreatedAt     string
		Campaign      FirstLegCampaignRequest
		RouterID      string
		RouterVersion string
		Weekly        *WeeklyFirstLegReservationBinding
	}{"a072-first-leg:v3", issueDigest, plan.Lineage, plan.AttemptID, plan.ActivationManifestDigest,
		plan.Revision, formatJournalTime(plan.CreatedAt), campaign, routerID, routerVersion, cloneWeeklyFirstLegBinding(weekly)}
	raw, err := json.Marshal(preimage)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:]), nil
}

func insertFirstLegCampaignTx(ctx context.Context, tx *sql.Tx, prepared preparedQFinalFirstLeg, token, now string) error {
	campaign := prepared.campaign
	lineage := prepared.strategyPlan.Lineage
	var currentGeneration, currentVersion int64
	var currentState string
	var positionVersion sql.NullInt64
	positionExists := true
	err := tx.QueryRowContext(ctx, `SELECT p.instance_seq,p.state,v.version
		FROM positions p LEFT JOIN position_projection_versions v ON v.position_id=p.id
		WHERE p.account_ref=? AND p.market=? AND p.symbol=? ORDER BY p.instance_seq DESC LIMIT 1`,
		prepared.decision.AccountRef, prepared.campaignMarket, prepared.campaignSymbol).
		Scan(&currentGeneration, &currentState, &positionVersion)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		positionExists = false
		currentGeneration, currentVersion, currentState = 0, 0, "FLAT"
	case err != nil:
		return err
	case !positionVersion.Valid:
		return fmt.Errorf("%w: position generation predates authoritative versioning", ErrGenerationConflict)
	default:
		currentVersion = positionVersion.Int64
	}
	var activeCampaign string
	err = tx.QueryRowContext(ctx, `SELECT campaign_id FROM position_campaign_claims WHERE account_ref=? AND market=? AND symbol=?`,
		prepared.decision.AccountRef, prepared.campaignMarket, prepared.campaignSymbol).Scan(&activeCampaign)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return err
	}
	if currentGeneration != campaign.ExpectedPositionGeneration || currentVersion != campaign.ExpectedPositionVersion ||
		(positionExists && currentState != "CLOSED") || err == nil {
		return fmt.Errorf("%w: first-leg position or claim changed", ErrGenerationConflict)
	}
	createDigest := digestParts(campaign.CampaignID, prepared.decision.AccountRef, prepared.campaignMarket,
		prepared.campaignSymbol, lineage.LaneID, lineage.LaneVersion, prepared.decision.ID, lineage.EvidenceDigest,
		strconv.FormatInt(campaign.ExpectedPositionGeneration, 10), strconv.FormatInt(campaign.ExpectedPositionVersion, 10), token)
	if _, err := tx.ExecContext(ctx, `INSERT INTO position_campaigns
		(id,account_ref,market,symbol,lane_id,lane_version,decision_id,evidence_digest,
		 expected_position_generation,expected_position_version,prospective_token,state,version,entry_blocked,created_at,updated_at)
		VALUES(?,?,?,?,?,?,?,?,?,?,?,'PLANNED',1,0,?,?)`, campaign.CampaignID, prepared.decision.AccountRef,
		prepared.campaignMarket, prepared.campaignSymbol, lineage.LaneID, lineage.LaneVersion, prepared.decision.ID,
		lineage.EvidenceDigest, campaign.ExpectedPositionGeneration, campaign.ExpectedPositionVersion, token, now, now); err != nil {
		if isConstraintError(err) {
			return fmt.Errorf("%w: first-leg campaign scope: %v", ErrGenerationConflict, err)
		}
		return err
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO position_campaign_claims
		(account_ref,market,symbol,position_generation,position_version,version,prospective_token,campaign_id,created_at,updated_at)
		VALUES(?,?,?,?,?,1,?,?,?,?)`, prepared.decision.AccountRef, prepared.campaignMarket, prepared.campaignSymbol,
		currentGeneration, currentVersion, token, campaign.CampaignID, now, now); err != nil {
		return fmt.Errorf("%w: first-leg prospective claim: %v", ErrGenerationConflict, err)
	}
	if err := insertCampaignCommand(ctx, tx, campaign.CampaignID, "CREATE", campaign.CreateCommandKey, createDigest, 1, 0, now); err != nil {
		return err
	}
	if err := insertCampaignEvent(ctx, tx, campaign.CampaignID, 1, 1, 0, "", "CREATED", "CREATE",
		campaign.CreateCommandKey, createDigest, positioncampaign.CampaignPlanned, "", 0, "", "", now); err != nil {
		return err
	}
	quantity := strconv.FormatUint(prepared.riskDecision.QFinal, 10)
	legDigest := digestParts(campaign.CampaignID, "1", campaign.FirstLegPlanID, quantity)
	if _, err := tx.ExecContext(ctx, `INSERT INTO campaign_legs
		(campaign_id,sequence,plan_id,requested_quantity,filled_quantity,residual_quantity,state,version,created_at,updated_at)
		VALUES(?,1,?,?,'0',?,'PLANNED',1,?,?)`, campaign.CampaignID, campaign.FirstLegPlanID, quantity, quantity, now, now); err != nil {
		return err
	}
	if result, err := tx.ExecContext(ctx, `UPDATE position_campaigns SET version=2,updated_at=? WHERE id=? AND version=1`, now, campaign.CampaignID); err != nil {
		return err
	} else if affected, affectedErr := result.RowsAffected(); affectedErr != nil {
		return fmt.Errorf("journal: first-leg campaign affected rows: %w", affectedErr)
	} else if affected != 1 {
		return fmt.Errorf("%w: first-leg campaign version", ErrCampaignVersionConflict)
	}
	if err := insertCampaignCommand(ctx, tx, campaign.CampaignID, "PLAN_LEG", campaign.FirstLegCommandKey, legDigest, 2, 1, now); err != nil {
		return err
	}
	return insertCampaignEvent(ctx, tx, campaign.CampaignID, 2, 2, 1, "", "LEG_PLANNED", "PLAN_LEG",
		campaign.FirstLegCommandKey, legDigest, positioncampaign.CampaignPlanned, positioncampaign.LegPlanned, 0, "", "", now)
}

func insertFirstLegBindingTx(ctx context.Context, tx *sql.Tx, prepared preparedQFinalFirstLeg, token, now string) error {
	lineage := prepared.strategyPlan.Lineage
	_, err := tx.ExecContext(ctx, `INSERT INTO strategy_first_leg_bindings(
		decision_id,aggregate_reservation_id,entry_decision_identity,attempt_id,campaign_id,leg_sequence,
		leg_plan_id,account_ref,market,symbol,candidate_id,evidence_digest,lane_id,lane_version,
		router_id,router_version,prospective_token,q_final,request_digest,record_digest,created_at)
		VALUES(?,?,?,?,?,1,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`, prepared.decision.ID,
		prepared.reservePlan.rows[0].ID, lineage.DecisionIdentity, prepared.strategyPlan.AttemptID,
		prepared.campaign.CampaignID, prepared.campaign.FirstLegPlanID, prepared.decision.AccountRef,
		strings.ToUpper(lineage.Market), prepared.campaignSymbol, lineage.CandidateLifeID, lineage.EvidenceDigest,
		lineage.LaneID, lineage.LaneVersion, prepared.routerID, prepared.routerVersion,
		token, prepared.riskDecision.QFinal, prepared.firstLegDigest,
		prepared.bindingRecordHash, now)
	if err != nil {
		return fmt.Errorf("journal: insert exact first-leg binding: %w", err)
	}
	return insertWeeklyFirstLegBindingTx(ctx, tx, prepared, now)
}

func verifyFirstLegReplayTx(ctx context.Context, tx *sql.Tx, prepared preparedQFinalFirstLeg, token string) error {
	lineage := prepared.strategyPlan.Lineage
	var count int
	err := tx.QueryRowContext(ctx, `SELECT count(*) FROM strategy_first_leg_bindings binding
		JOIN strategy_attempt_lineage attempt ON attempt.attempt_id=binding.attempt_id
		JOIN strategy_decision_lineage strategy ON strategy.entry_decision_identity=binding.entry_decision_identity
		JOIN position_campaigns campaign ON campaign.id=binding.campaign_id
		JOIN position_campaign_claims claim ON claim.campaign_id=binding.campaign_id
		JOIN campaign_legs leg ON leg.campaign_id=binding.campaign_id AND leg.sequence=binding.leg_sequence
		WHERE binding.decision_id=? AND binding.aggregate_reservation_id=? AND binding.entry_decision_identity=?
		AND binding.attempt_id=? AND binding.campaign_id=? AND binding.leg_sequence=1 AND binding.leg_plan_id=?
		AND binding.account_ref=? AND binding.market=? AND binding.symbol=? AND binding.candidate_id=?
		AND binding.evidence_digest=? AND binding.lane_id=? AND binding.lane_version=?
		AND binding.router_id=? AND binding.router_version=?
		AND binding.prospective_token=? AND binding.q_final=? AND binding.request_digest=? AND binding.record_digest=?
		AND attempt.risk_intent_id=binding.decision_id AND attempt.guardian_decision_id=binding.decision_id
		AND strategy.quantity=? AND campaign.decision_id=binding.decision_id AND campaign.prospective_token=?
		AND claim.prospective_token=? AND leg.plan_id=? AND leg.requested_quantity=?`,
		prepared.decision.ID, prepared.reservePlan.rows[0].ID, lineage.DecisionIdentity,
		prepared.strategyPlan.AttemptID, prepared.campaign.CampaignID, prepared.campaign.FirstLegPlanID,
		prepared.decision.AccountRef, strings.ToUpper(lineage.Market), prepared.campaignSymbol,
		lineage.CandidateLifeID, lineage.EvidenceDigest, lineage.LaneID, lineage.LaneVersion,
		prepared.routerID, prepared.routerVersion, token,
		prepared.riskDecision.QFinal, prepared.firstLegDigest, prepared.bindingRecordHash,
		strconv.FormatUint(prepared.riskDecision.QFinal, 10), token, token, prepared.campaign.FirstLegPlanID,
		strconv.FormatUint(prepared.riskDecision.QFinal, 10)).Scan(&count)
	if err != nil || count != 1 {
		return fmt.Errorf("%w: first-leg replay authority", ErrRiskBucketReplayMismatch)
	}
	return verifyWeeklyFirstLegReplayTx(ctx, tx, prepared)
}

func firstLegReceipt(prepared preparedQFinalFirstLeg, qFinal RiskBucketAdmissionReceipt, idempotent bool) QFinalCampaignFirstLegReceipt {
	return QFinalCampaignFirstLegReceipt{DecisionID: prepared.decision.ID,
		AggregateReservationID: prepared.reservePlan.rows[0].ID,
		BucketReservationIDs:   append([]string(nil), qFinal.ReservationIDs...),
		CampaignID:             prepared.campaign.CampaignID, AttemptID: prepared.strategyPlan.AttemptID,
		FirstLegPlanID: prepared.campaign.FirstLegPlanID, LegSequence: 1,
		Market: strings.ToUpper(prepared.strategyPlan.Lineage.Market), Symbol: prepared.campaignSymbol,
		RouterID: prepared.routerID, RouterVersion: prepared.routerVersion,
		QFinal: prepared.riskDecision.QFinal, Idempotent: idempotent}
}
