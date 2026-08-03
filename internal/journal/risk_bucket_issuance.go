package journal

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/JungHoonGhae/tossinvest-cli/internal/riskbucket"
)

const qFinalPolicyMarker = "|a066-q-final:v1|"

// QFinalIssueRequest is the only journal request that can commit a Guardian
// decision and the five monetary reservations together. Admission is
// recalculated inside the transaction boundary; QFinal is never caller data.
type QFinalIssueRequest struct {
	Issue     IssueRequest
	Admission RiskBucketAdmissionPlan
}

type QFinalIssueResult struct {
	Issue     IssueResult
	Admission RiskBucketAdmissionReceipt
}

type CollectQFinalIssue func(context.Context, int) (QFinalIssueRequest, error)

func (j *Journal) RecordQFinalDecisionAndReserveWithRecollection(ctx context.Context, collect CollectQFinalIssue, policy RecollectPolicy) (QFinalIssueResult, error) {
	if collect == nil {
		return QFinalIssueResult{}, fmt.Errorf("%w: q_final issuing needs a snapshot collector", ErrInvalidRequest)
	}
	return recollectLoop(j.clk, policy, func(attempt int) (QFinalIssueResult, error) {
		request, err := collect(ctx, attempt)
		if err != nil {
			return QFinalIssueResult{}, fmt.Errorf("journal: collecting q_final broker snapshot: %w", err)
		}
		return j.RecordQFinalDecisionAndReserve(ctx, request)
	})
}

func QFinalPolicyVersion(base, transactionID string) (string, error) {
	rawBase, rawTransactionID := base, transactionID
	base = strings.TrimSpace(base)
	transactionID = strings.TrimSpace(transactionID)
	if base == "" || transactionID == "" || rawBase != base || rawTransactionID != transactionID ||
		len(base) > 256 || len(transactionID) > 256 || strings.Contains(base, qFinalPolicyMarker) || strings.ContainsAny(transactionID, "\r\n|") {
		return "", fmt.Errorf("%w: invalid q_final policy binding", ErrInvalidRequest)
	}
	return base + qFinalPolicyMarker + transactionID, nil
}

func splitQFinalPolicyVersion(version string) (base, transactionID string, required bool) {
	parts := strings.Split(version, qFinalPolicyMarker)
	if len(parts) != 2 || strings.TrimSpace(parts[0]) == "" || strings.TrimSpace(parts[1]) == "" {
		return "", "", false
	}
	return parts[0], parts[1], true
}

// RecordQFinalDecisionAndReserve commits the q_final GuardianDecision, the
// existing aggregate reservation, all five monetary reservations and owner in
// one transaction. Any owner/snapshot/bucket conflict rolls every row back.
func (j *Journal) RecordQFinalDecisionAndReserve(ctx context.Context, request QFinalIssueRequest) (QFinalIssueResult, error) {
	if j == nil || j.db == nil {
		return QFinalIssueResult{}, errors.New("journal q_final issuance: journal required")
	}
	decision, reserve, reservePlan, err := request.Issue.build()
	if err != nil {
		return QFinalIssueResult{}, err
	}
	if decision.SafetyClass != SafetyClassExposureRaising || decision.PreimageKind != PreimageKindRiskIntent {
		return QFinalIssueResult{}, errors.New("journal q_final issuance: canonical RiskIntent required")
	}
	preimage, err := ParsePreimage(decision.PreimageKind, decision.RiskPreimage)
	if err != nil {
		return QFinalIssueResult{}, err
	}
	intent, ok := preimage.(RiskIntent)
	if !ok {
		return QFinalIssueResult{}, errors.New("journal q_final issuance: RiskIntent required")
	}
	_, transactionID, marked := splitQFinalPolicyVersion(intent.PolicyVersion)
	if !marked || transactionID != request.Admission.TransactionID {
		return QFinalIssueResult{}, fmt.Errorf("%w: q_final policy marker", ErrRiskBucketSnapshotMismatch)
	}
	riskDecision := riskbucket.CalculateAdmission(request.Admission.Admission)
	if riskDecision.Refusal != nil {
		return QFinalIssueResult{}, riskDecision.Refusal
	}
	if request.Admission.DecisionID != decision.ID || request.Admission.ExistingReservationID == "" ||
		request.Admission.ExistingReservationID != soleReservationID(reserve.Reservations) {
		return QFinalIssueResult{}, fmt.Errorf("%w: q_final decision/reservation binding", ErrRiskBucketSnapshotMismatch)
	}
	quantity, ok := NormalizeDecimal(intent.Quantity)
	if !ok || quantity != strconv.FormatUint(riskDecision.QFinal, 10) {
		return QFinalIssueResult{}, fmt.Errorf("%w: decision quantity is not q_final", ErrRiskBucketSnapshotMismatch)
	}
	if err := validateRiskBucketAdmission(request.Admission, riskDecision); err != nil {
		return QFinalIssueResult{}, err
	}
	admissionPreimage, admissionDigest, err := riskBucketAdmissionDigest(request.Admission, riskDecision)
	if err != nil {
		return QFinalIssueResult{}, err
	}
	issuePreimage, issueDigest, err := qFinalIssueDigest(request.Admission, decision, reserve, reservePlan, admissionPreimage, admissionDigest)
	if err != nil {
		return QFinalIssueResult{}, err
	}

	tx, err := j.db.BeginTx(ctx, nil)
	if err != nil {
		return QFinalIssueResult{}, fmt.Errorf("journal: begin q_final issuance: %w", err)
	}
	defer tx.Rollback()
	if replayed, replay, err := recoverQFinalIssueReplayTx(ctx, tx, request.Admission, decision, reserve, reservePlan, riskDecision, issueDigest); err != nil {
		return QFinalIssueResult{}, err
	} else if replayed {
		if err := tx.Commit(); err != nil {
			return QFinalIssueResult{}, fmt.Errorf("journal: commit q_final replay %s: %w", decision.ID, err)
		}
		return replay, nil
	}
	if err := reservePrecheck(ctx, tx, reserve, reservePlan, j.clk.Now().UTC()); err != nil {
		return QFinalIssueResult{}, err
	}
	if err := insertDecisionRow(ctx, tx, decision); err != nil {
		return QFinalIssueResult{}, err
	}
	reserved, err := reserveRows(ctx, tx, reserve, reservePlan, j.clk.Now().UTC())
	if err != nil {
		return QFinalIssueResult{}, err
	}
	admissionReceipt, err := commitFreshRiskBucketAdmissionTx(ctx, tx, request.Admission, riskDecision, issuePreimage, issueDigest, reserved.Version)
	if err != nil {
		return QFinalIssueResult{}, err
	}
	if err := tx.Commit(); err != nil {
		return QFinalIssueResult{}, fmt.Errorf("journal: commit q_final issuance %s: %w", decision.ID, err)
	}
	return QFinalIssueResult{
		Issue:     IssueResult{Decision: decision, Version: reserved.Version, Reservations: reserved.Reservations},
		Admission: admissionReceipt,
	}, nil
}

func soleReservationID(reservations []ReservationRequest) string {
	if len(reservations) != 1 {
		return ""
	}
	return reservations[0].ID
}

// qFinalIssueDigest binds the whole atomic request, not only the five-bucket
// calculation. Snapshot concurrency evidence, aggregate usage/limits, the
// exact decision and CreatedAt all participate, so a stable transaction id is
// an idempotency key rather than permission to substitute an equivalent-size
// authority.
func qFinalIssueDigest(plan RiskBucketAdmissionPlan, decision Decision, reserve ReserveRequest, reservePlan reservePlan, admissionPreimage, admissionDigest string) (string, string, error) {
	type canonicalReserve struct {
		DecisionID, AccountRef string
		SnapshotAsOf           string
		ObservedVersion        int64
		StalenessNanoseconds   int64
		SnapshotUsage          []AggregateAmount
		Limits                 []AggregateAmount
		Reservations           []ReservationRequest
	}
	type canonicalDecision struct {
		ID, AccountRef, SafetyClass, PreimageKind, RiskPreimage, RiskHash string
		ClientOrderID, LimitsJSON, Nonce                                  string
		Generation                                                        int
		IssuedAt, ExpiresAt                                               string
	}
	usage := make([]AggregateAmount, 0, len(reservePlan.usage))
	limits := make([]AggregateAmount, 0, len(reservePlan.limits))
	for _, value := range reservePlan.usage {
		usage = append(usage, value)
	}
	for _, value := range reservePlan.limits {
		limits = append(limits, value)
	}
	sort.Slice(usage, func(i, k int) bool {
		if usage[i].Kind == usage[k].Kind {
			return usage[i].Currency < usage[k].Currency
		}
		return usage[i].Kind < usage[k].Kind
	})
	sort.Slice(limits, func(i, k int) bool {
		if limits[i].Kind == limits[k].Kind {
			return limits[i].Currency < limits[k].Currency
		}
		return limits[i].Kind < limits[k].Kind
	})
	rows := append([]ReservationRequest(nil), reservePlan.rows...)
	sort.Slice(rows, func(i, k int) bool { return rows[i].ID < rows[k].ID })
	preimage := struct {
		Version                            string
		Decision                           canonicalDecision
		Reserve                            canonicalReserve
		AdmissionPreimage, AdmissionDigest string
		AdmissionCreatedAt                 string
	}{
		Version: "a066-q-final-issuance:v1",
		Decision: canonicalDecision{
			ID: decision.ID, AccountRef: decision.AccountRef, Generation: decision.Generation,
			SafetyClass: decision.SafetyClass, PreimageKind: decision.PreimageKind,
			RiskPreimage: decision.RiskPreimage, RiskHash: decision.RiskHash,
			ClientOrderID: decision.ClientOrderID, LimitsJSON: decision.LimitsJSON,
			Nonce: decision.Nonce, IssuedAt: formatJournalTime(decision.IssuedAt), ExpiresAt: formatJournalTime(decision.ExpiresAt),
		},
		Reserve: canonicalReserve{
			DecisionID: reservePlan.decisionID, AccountRef: reservePlan.accountRef,
			SnapshotAsOf: formatJournalTime(reserve.SnapshotAsOf), ObservedVersion: reserve.ObservedVersion,
			StalenessNanoseconds: int64(reserve.Staleness), SnapshotUsage: usage, Limits: limits, Reservations: rows,
		},
		AdmissionPreimage: admissionPreimage, AdmissionDigest: admissionDigest,
		AdmissionCreatedAt: canonicalRiskTime(plan.CreatedAt),
	}
	raw, err := json.Marshal(preimage)
	if err != nil {
		return "", "", fmt.Errorf("journal: encode q_final issuance: %w", err)
	}
	sum := sha256.Sum256(raw)
	return string(raw), hex.EncodeToString(sum[:]), nil
}

func recoverQFinalIssueReplayTx(ctx context.Context, tx *sql.Tx, plan RiskBucketAdmissionPlan, decision Decision, reserve ReserveRequest, reservePlan reservePlan, riskDecision riskbucket.AdmissionDecision, issueDigest string) (bool, QFinalIssueResult, error) {
	rows, err := tx.QueryContext(ctx, `SELECT decision_id,transaction_id,account_ref,market,symbol,q_final,existing_reservation_id,request_digest,request_preimage,owner_prospective_generation,owner_lane_id,owner_campaign_id,owner_sequence FROM risk_bucket_final_decisions WHERE transaction_id=? OR decision_id=? ORDER BY decision_id`, plan.TransactionID, plan.DecisionID)
	if err != nil {
		return false, QFinalIssueResult{}, fmt.Errorf("journal: inspect q_final replay identity: %w", err)
	}
	type finalRecord struct {
		decisionID, transactionID, account, market, symbol, existingID, digest string
		storedPreimage, prospective, lane, campaign                            string
		qFinal                                                                 uint64
		ownerSequence                                                          int64
	}
	var records []finalRecord
	for rows.Next() {
		var record finalRecord
		if err := rows.Scan(&record.decisionID, &record.transactionID, &record.account, &record.market, &record.symbol, &record.qFinal, &record.existingID, &record.digest, &record.storedPreimage, &record.prospective, &record.lane, &record.campaign, &record.ownerSequence); err != nil {
			rows.Close()
			return false, QFinalIssueResult{}, fmt.Errorf("journal: scan q_final replay identity: %w", err)
		}
		records = append(records, record)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return false, QFinalIssueResult{}, fmt.Errorf("journal: iterate q_final replay identity: %w", err)
	}
	rows.Close()
	if len(records) == 0 {
		var collisions int
		if err := tx.QueryRowContext(ctx, `SELECT (SELECT count(*) FROM decisions WHERE id=?)+(SELECT count(*) FROM risk_reservations WHERE id=?)`, decision.ID, soleReservationID(reservePlan.rows)).Scan(&collisions); err != nil {
			return false, QFinalIssueResult{}, err
		}
		if collisions != 0 {
			return false, QFinalIssueResult{}, fmt.Errorf("%w: partial q_final issuance identity", ErrRiskBucketReplayMismatch)
		}
		return false, QFinalIssueResult{}, nil
	}
	if len(records) != 1 {
		return false, QFinalIssueResult{}, fmt.Errorf("%w: ambiguous q_final issuance identity", ErrRiskBucketReplayMismatch)
	}
	record := records[0]
	owner := plan.Owner
	if record.decisionID != decision.ID || record.transactionID != plan.TransactionID || record.account != owner.Key.AccountID ||
		record.market != string(owner.Key.Market) || record.symbol != owner.Key.Symbol || record.qFinal != riskDecision.QFinal ||
		record.existingID != plan.ExistingReservationID || record.digest != issueDigest || record.prospective != owner.Key.ProspectiveGeneration ||
		record.lane != owner.LaneID || record.campaign != owner.CampaignID || record.ownerSequence <= 0 {
		return false, QFinalIssueResult{}, fmt.Errorf("%w: divergent q_final issuance replay", ErrRiskBucketReplayMismatch)
	}
	metadata, err := parseQFinalStoredIssuance(record.storedPreimage, record.digest)
	if err != nil {
		return false, QFinalIssueResult{}, err
	}
	storedDecision, err := scanDecision(tx.QueryRowContext(ctx, decisionSelect+" WHERE id = ?", decision.ID))
	if err != nil || !sameQFinalDecision(storedDecision, decision) {
		return false, QFinalIssueResult{}, fmt.Errorf("%w: q_final replay decision: %v", ErrRiskBucketReplayMismatch, err)
	}
	storedReservation, err := scanReservation(tx.QueryRowContext(ctx, reservationSelect+" WHERE id = ?", record.existingID))
	if err != nil || !sameQFinalReservation(storedReservation, reserve, reservePlan, decision.ID) {
		return false, QFinalIssueResult{}, fmt.Errorf("%w: q_final replay aggregate reservation: %v", ErrRiskBucketReplayMismatch, err)
	}
	key := riskbucket.OwnerKey{AccountID: record.account, Market: riskbucket.Market(record.market), Symbol: record.symbol, ProspectiveGeneration: record.prospective}
	if err := verifyQFinalAdmissionRows(ctx, tx, decision.ID, record.existingID, key, record.lane, record.campaign); err != nil {
		return false, QFinalIssueResult{}, err
	}
	currentVersion, err := reservationVersion(ctx, tx, record.account)
	if err != nil {
		return false, QFinalIssueResult{}, err
	}
	if metadata.ReservationVersion <= 0 || currentVersion < metadata.ReservationVersion {
		return false, QFinalIssueResult{}, fmt.Errorf("%w: q_final replay reservation version", ErrRiskBucketReplayMismatch)
	}
	receipt := riskBucketReceipt(plan, riskDecision, metadata.OwnerReused, true)
	return true, QFinalIssueResult{
		Issue:     IssueResult{Decision: storedDecision, Version: metadata.ReservationVersion, Reservations: []Reservation{storedReservation}},
		Admission: receipt,
	}, nil
}

type qFinalStoredIssuance struct {
	Request            json.RawMessage `json:"request"`
	OwnerReused        bool            `json:"owner_reused"`
	ReservationVersion int64           `json:"reservation_version"`
}

func encodeQFinalStoredIssuance(requestPreimage string, ownerReused bool, reservationVersion int64) (string, error) {
	if !json.Valid([]byte(requestPreimage)) || reservationVersion <= 0 {
		return "", fmt.Errorf("%w: q_final stored issuance metadata", ErrRiskBucketReplayMismatch)
	}
	raw, err := json.Marshal(qFinalStoredIssuance{Request: json.RawMessage(requestPreimage), OwnerReused: ownerReused, ReservationVersion: reservationVersion})
	if err != nil {
		return "", fmt.Errorf("journal: encode q_final stored issuance: %w", err)
	}
	return string(raw), nil
}

func parseQFinalStoredIssuance(stored, expectedDigest string) (qFinalStoredIssuance, error) {
	var metadata qFinalStoredIssuance
	decoder := json.NewDecoder(strings.NewReader(stored))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&metadata); err != nil || !json.Valid(metadata.Request) {
		return qFinalStoredIssuance{}, fmt.Errorf("%w: q_final stored issuance metadata: %v", ErrRiskBucketReplayMismatch, err)
	}
	sum := sha256.Sum256(metadata.Request)
	if hex.EncodeToString(sum[:]) != expectedDigest {
		return qFinalStoredIssuance{}, fmt.Errorf("%w: q_final stored issuance digest", ErrRiskBucketReplayMismatch)
	}
	return metadata, nil
}

func sameQFinalDecision(got, want Decision) bool {
	return got.ID == want.ID && got.AccountRef == want.AccountRef && got.Generation == want.Generation &&
		got.SafetyClass == want.SafetyClass && got.PreimageKind == want.PreimageKind && got.RiskPreimage == want.RiskPreimage &&
		got.RiskHash == want.RiskHash && got.ClientOrderID == want.ClientOrderID && got.LimitsJSON == want.LimitsJSON &&
		got.Nonce == want.Nonce && got.IssuedAt.Equal(want.IssuedAt) && got.ExpiresAt.Equal(want.ExpiresAt)
}

func sameQFinalReservation(got Reservation, reserve ReserveRequest, plan reservePlan, decisionID string) bool {
	if len(plan.rows) != 1 {
		return false
	}
	want := plan.rows[0]
	return got.ID == want.ID && got.DecisionID == decisionID && got.AttemptID == "" && got.AccountRef == plan.accountRef &&
		got.Kind == want.Kind && got.Amount == want.Amount && got.Currency == want.Currency && got.TradingDay == want.TradingDay &&
		got.SnapshotAsOf.Equal(reserve.SnapshotAsOf.UTC()) && got.State == ReservationHeld && got.ReleasedAt.IsZero() && got.ReleaseReason == ""
}

func verifyQFinalAdmissionRows(ctx context.Context, q riskBucketQueryer, decisionID, existingID string, key riskbucket.OwnerKey, lane, campaign string) error {
	var activeOwner int
	if err := q.QueryRowContext(ctx, `SELECT count(*) FROM risk_bucket_owners WHERE account_ref=? AND market=? AND symbol=? AND prospective_generation=? AND lane_id=? AND campaign_id=? AND released_at IS NULL`, key.AccountID, string(key.Market), key.Symbol, key.ProspectiveGeneration, lane, campaign).Scan(&activeOwner); err != nil || activeOwner != 1 {
		return fmt.Errorf("%w: q_final owner", ErrRiskBucketReplayMismatch)
	}
	if err := verifyRiskBucketStateDigest(ctx, q, key); err != nil {
		return err
	}
	rows, err := q.QueryContext(ctx, `SELECT bucket_dimension,state,reserved_minor,held_minor,existing_reservation_id,owner_prospective_generation FROM risk_bucket_reservations WHERE decision_id=? ORDER BY bucket_dimension`, decisionID)
	if err != nil {
		return err
	}
	defer rows.Close()
	seen := make(map[riskbucket.Dimension]bool)
	for rows.Next() {
		var dimension riskbucket.Dimension
		var state, reserved, held, linkedExisting, linkedOwner string
		if err := rows.Scan(&dimension, &state, &reserved, &held, &linkedExisting, &linkedOwner); err != nil {
			return err
		}
		if seen[dimension] || state != "HELD" || reserved == "" || held != reserved || linkedExisting != existingID || linkedOwner != key.ProspectiveGeneration {
			return fmt.Errorf("%w: q_final monetary reservation", ErrRiskBucketReplayMismatch)
		}
		seen[dimension] = true
	}
	if err := rows.Err(); err != nil {
		return err
	}
	for _, dimension := range riskbucket.RequiredDimensionOrder() {
		if !seen[dimension] {
			return fmt.Errorf("%w: missing %s q_final reservation", ErrRiskBucketReplayMismatch, dimension)
		}
	}
	if len(seen) != len(riskbucket.RequiredDimensionOrder()) {
		return fmt.Errorf("%w: q_final reservation dimension set", ErrRiskBucketReplayMismatch)
	}
	return nil
}

func commitFreshRiskBucketAdmissionTx(ctx context.Context, tx *sql.Tx, plan RiskBucketAdmissionPlan, decision riskbucket.AdmissionDecision, preimage, digest string, reservationVersion int64) (RiskBucketAdmissionReceipt, error) {
	var existing int
	if err := tx.QueryRowContext(ctx, `SELECT count(*) FROM risk_bucket_final_decisions WHERE transaction_id=? OR decision_id=?`, plan.TransactionID, plan.DecisionID).Scan(&existing); err != nil {
		return RiskBucketAdmissionReceipt{}, err
	}
	if existing != 0 {
		return RiskBucketAdmissionReceipt{}, fmt.Errorf("%w: q_final issuance identity", ErrRiskBucketReplayMismatch)
	}
	var reservationAccount, reservationDecision, reservationState string
	if err := tx.QueryRowContext(ctx, `SELECT account_ref,decision_id,state FROM risk_reservations WHERE id=?`, plan.ExistingReservationID).Scan(&reservationAccount, &reservationDecision, &reservationState); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return RiskBucketAdmissionReceipt{}, ErrReservationNotFound
		}
		return RiskBucketAdmissionReceipt{}, err
	}
	if reservationAccount != plan.Owner.Key.AccountID || reservationDecision != plan.DecisionID || reservationState != ReservationHeld {
		return RiskBucketAdmissionReceipt{}, fmt.Errorf("%w: existing reservation binding", ErrRiskBucketSnapshotMismatch)
	}
	if err := ensureRiskBucketEntryScopeClean(ctx, tx, plan.Owner.Key); err != nil {
		return RiskBucketAdmissionReceipt{}, err
	}

	ownerReused := false
	var prospective, lane, campaign string
	err := tx.QueryRowContext(ctx, `SELECT prospective_generation,lane_id,campaign_id FROM risk_bucket_owners WHERE account_ref=? AND market=? AND symbol=? AND released_at IS NULL`, plan.Owner.Key.AccountID, string(plan.Owner.Key.Market), plan.Owner.Key.Symbol).Scan(&prospective, &lane, &campaign)
	switch {
	case err == nil:
		if prospective != plan.Owner.Key.ProspectiveGeneration || lane != plan.Owner.LaneID || campaign != plan.Owner.CampaignID {
			return RiskBucketAdmissionReceipt{}, ErrRiskBucketOwnerConflict
		}
		ownerReused = true
	case errors.Is(err, sql.ErrNoRows):
		if _, err = tx.ExecContext(ctx, `INSERT INTO risk_bucket_owners(account_ref,market,symbol,prospective_generation,lane_id,campaign_id,acquired_at) VALUES(?,?,?,?,?,?,?)`, plan.Owner.Key.AccountID, string(plan.Owner.Key.Market), plan.Owner.Key.Symbol, plan.Owner.Key.ProspectiveGeneration, plan.Owner.LaneID, plan.Owner.CampaignID, canonicalRiskTime(plan.CreatedAt)); err != nil {
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
		if err := verifyRiskBucketScaleInIdentity(ctx, tx, plan.Owner.Key, decision.Caps); err != nil {
			return RiskBucketAdmissionReceipt{}, err
		}
	}
	storedPreimage, err := encodeQFinalStoredIssuance(preimage, ownerReused, reservationVersion)
	if err != nil {
		return RiskBucketAdmissionReceipt{}, err
	}

	var ownerSequence int64
	if err := tx.QueryRowContext(ctx, `SELECT COALESCE(MAX(owner_sequence),0)+1 FROM risk_bucket_final_decisions WHERE account_ref=? AND market=? AND symbol=? AND owner_prospective_generation=?`, plan.Owner.Key.AccountID, string(plan.Owner.Key.Market), plan.Owner.Key.Symbol, plan.Owner.Key.ProspectiveGeneration).Scan(&ownerSequence); err != nil {
		return RiskBucketAdmissionReceipt{}, err
	}
	snapshotSetDigest, err := riskBucketSnapshotSetDigest(plan.Snapshots)
	if err != nil {
		return RiskBucketAdmissionReceipt{}, err
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO risk_bucket_final_decisions(decision_id,transaction_id,account_ref,market,symbol,q_candidate,q_existing_guardian,q_final,existing_reservation_id,request_digest,request_preimage,snapshot_set_digest,owner_prospective_generation,owner_lane_id,owner_campaign_id,owner_sequence,created_at) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`, plan.DecisionID, plan.TransactionID, plan.Owner.Key.AccountID, string(plan.Owner.Key.Market), plan.Owner.Key.Symbol, decision.QCandidate, decision.QExistingGuardian, decision.QFinal, plan.ExistingReservationID, digest, storedPreimage, snapshotSetDigest, plan.Owner.Key.ProspectiveGeneration, plan.Owner.LaneID, plan.Owner.CampaignID, ownerSequence, canonicalRiskTime(plan.CreatedAt))
	if err != nil {
		return RiskBucketAdmissionReceipt{}, fmt.Errorf("journal: insert q_final risk bucket decision: %w", err)
	}
	if err := insertFreshRiskBucketReservations(ctx, tx, plan, decision); err != nil {
		return RiskBucketAdmissionReceipt{}, err
	}
	_, stateDigest, err := loadRiskBucketState(ctx, tx, plan.Owner.Key)
	if err != nil {
		return RiskBucketAdmissionReceipt{}, err
	}
	var stateSequence int64
	if err := tx.QueryRowContext(ctx, `SELECT COALESCE(MAX(event_sequence),0)+1 FROM risk_bucket_state_snapshots WHERE account_ref=? AND market=? AND symbol=? AND prospective_generation=?`, plan.Owner.Key.AccountID, string(plan.Owner.Key.Market), plan.Owner.Key.Symbol, plan.Owner.Key.ProspectiveGeneration).Scan(&stateSequence); err != nil {
		return RiskBucketAdmissionReceipt{}, err
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO risk_bucket_state_snapshots(snapshot_id,account_ref,market,symbol,prospective_generation,state_digest,event_sequence,created_at) VALUES(?,?,?,?,?,?,?,?)`, plan.TransactionID+":state:"+strconv.FormatInt(stateSequence, 10), plan.Owner.Key.AccountID, string(plan.Owner.Key.Market), plan.Owner.Key.Symbol, plan.Owner.Key.ProspectiveGeneration, stateDigest, stateSequence, canonicalRiskTime(plan.CreatedAt)); err != nil {
		return RiskBucketAdmissionReceipt{}, err
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO risk_bucket_events(event_id,account_ref,market,symbol,prospective_generation,event_type,event_digest,payload,created_at) VALUES(?,?,?,?,?,'ADMISSION_COMMITTED',?,?,?)`, plan.TransactionID+":event:1", plan.Owner.Key.AccountID, string(plan.Owner.Key.Market), plan.Owner.Key.Symbol, plan.Owner.Key.ProspectiveGeneration, digest, storedPreimage, canonicalRiskTime(plan.CreatedAt)); err != nil {
		return RiskBucketAdmissionReceipt{}, err
	}
	return riskBucketReceipt(plan, decision, ownerReused, false), nil
}

func verifyRiskBucketScaleInIdentity(ctx context.Context, tx *sql.Tx, key riskbucket.OwnerKey, caps []riskbucket.BucketCap) error {
	rows, err := tx.QueryContext(ctx, `SELECT DISTINCT bucket_dimension,bucket_value,policy_version FROM risk_bucket_reservations WHERE account_ref=? AND market=? AND symbol=? AND owner_prospective_generation=?`, key.AccountID, string(key.Market), key.Symbol, key.ProspectiveGeneration)
	if err != nil {
		return err
	}
	defer rows.Close()
	existing := make(map[riskbucket.BucketKey]bool)
	for rows.Next() {
		var d, v, pv string
		if err := rows.Scan(&d, &v, &pv); err != nil {
			return err
		}
		existing[riskbucket.BucketKey{Dimension: riskbucket.Dimension(d), Value: v, PolicyVersion: pv}] = true
	}
	if err := rows.Err(); err != nil {
		return err
	}
	if len(existing) != len(caps) {
		return fmt.Errorf("%w: scale-in bucket identity", ErrRiskBucketSnapshotMismatch)
	}
	for _, cap := range caps {
		if !existing[cap.Key] {
			return fmt.Errorf("%w: scale-in bucket identity", ErrRiskBucketSnapshotMismatch)
		}
	}
	return nil
}

func insertFreshRiskBucketReservations(ctx context.Context, tx *sql.Tx, plan RiskBucketAdmissionPlan, decision riskbucket.AdmissionDecision) error {
	p := plan.Admission.Policy
	for i, cap := range decision.Caps {
		bucket, ref := plan.Admission.Buckets[i], plan.Snapshots[i]
		bound := bucket.BoundEvidence()
		policyRecordDigest, err := riskBucketRecordDigest(struct {
			Key      riskbucket.BucketKey
			Evidence riskbucket.Evidence
			Policy   riskbucket.ReservePolicy
		}{bound.Key, bound.PolicyEvidence, p})
		if err != nil {
			return err
		}
		if _, err = tx.ExecContext(ctx, `INSERT OR IGNORE INTO risk_bucket_policies(bucket_dimension,bucket_value,policy_version,policy_digest,policy_source,policy_observed_at,policy_fresh_until,record_digest,account_currency,quote_currency,evaluated_at,worst_price_quote,price_source,price_version,price_digest,price_observed_at,price_fresh_until,fee_fixed_base_minor,fee_per_unit_base_minor,fee_minimum_base_minor,fee_version,fee_digest,fx_rate_quote_to_base,fx_haircut,fx_source,fx_version,fx_digest,fx_observed_at,fx_fresh_until,created_at) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`, string(cap.Key.Dimension), cap.Key.Value, cap.Key.PolicyVersion, ref.PolicyDigest, bound.PolicyEvidence.Source, canonicalRiskTime(bound.PolicyEvidence.ObservedAt), canonicalRiskTime(bound.PolicyEvidence.FreshUntil), policyRecordDigest, p.AccountCurrency, p.QuoteCurrency, canonicalRiskTime(p.EvaluatedAt), p.Price.WorstExecutableQuote, p.Price.Source, p.Price.Version, p.Price.Digest, canonicalRiskTime(p.Price.ObservedAt), canonicalRiskTime(p.Price.FreshUntil), p.Fee.FixedBaseMinor, p.Fee.PerUnitBaseMinor, p.Fee.MinimumBaseMinor, p.Fee.Version, p.Fee.Digest, p.FX.RateQuoteToBase, p.FX.Haircut, p.FX.Source, p.FX.Version, p.FX.Digest, canonicalRiskTime(p.FX.ObservedAt), canonicalRiskTime(p.FX.FreshUntil), canonicalRiskTime(plan.CreatedAt)); err != nil {
			return err
		}
		var storedPolicyDigest string
		if err := tx.QueryRowContext(ctx, `SELECT record_digest FROM risk_bucket_policies WHERE bucket_dimension=? AND bucket_value=? AND policy_version=?`, string(cap.Key.Dimension), cap.Key.Value, cap.Key.PolicyVersion).Scan(&storedPolicyDigest); err != nil || storedPolicyDigest != policyRecordDigest {
			return fmt.Errorf("%w: immutable policy collision", ErrRiskBucketSnapshotMismatch)
		}
		snapshotRecordDigest, err := riskBucketRecordDigest(struct {
			Reference           RiskBucketSnapshotReference
			Limit, Filled, Held string
		}{ref, bucket.LimitMinor, bucket.FilledMinor, bucket.HeldMinor})
		if err != nil {
			return err
		}
		if _, err = tx.ExecContext(ctx, `INSERT OR IGNORE INTO risk_bucket_snapshots(snapshot_id,snapshot_digest,snapshot_source,record_digest,bucket_dimension,bucket_value,policy_version,limit_minor,filled_minor,held_minor,snapshot_version,policy_digest,observed_at,fresh_until,created_at) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`, ref.SnapshotID, ref.SnapshotDigest, bound.SnapshotEvidence.Source, snapshotRecordDigest, string(cap.Key.Dimension), cap.Key.Value, cap.Key.PolicyVersion, bucket.LimitMinor, bucket.FilledMinor, bucket.HeldMinor, ref.SnapshotVersion, ref.PolicyDigest, canonicalRiskTime(ref.ObservedAt), canonicalRiskTime(ref.FreshUntil), canonicalRiskTime(plan.CreatedAt)); err != nil {
			return err
		}
		var storedSnapshotDigest string
		if err := tx.QueryRowContext(ctx, `SELECT record_digest FROM risk_bucket_snapshots WHERE snapshot_id=?`, ref.SnapshotID).Scan(&storedSnapshotDigest); err != nil || storedSnapshotDigest != snapshotRecordDigest {
			return fmt.Errorf("%w: immutable snapshot collision", ErrRiskBucketSnapshotMismatch)
		}
		reservationID := riskBucketReservationID(plan.TransactionID, cap.Key.Dimension)
		if _, err = tx.ExecContext(ctx, `INSERT INTO risk_bucket_reservations(reservation_id,decision_id,existing_reservation_id,account_ref,market,symbol,owner_prospective_generation,bucket_dimension,bucket_value,policy_version,snapshot_id,reserved_minor,held_minor,filled_minor,overage_minor,state,created_at,updated_at) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,'0','0','HELD',?,?)`, reservationID, plan.DecisionID, plan.ExistingReservationID, plan.Owner.Key.AccountID, string(plan.Owner.Key.Market), plan.Owner.Key.Symbol, plan.Owner.Key.ProspectiveGeneration, string(cap.Key.Dimension), cap.Key.Value, cap.Key.PolicyVersion, ref.SnapshotID, cap.ReservationAtFinal, cap.ReservationAtFinal, canonicalRiskTime(plan.CreatedAt), canonicalRiskTime(plan.CreatedAt)); err != nil {
			return fmt.Errorf("journal: insert %s q_final reservation: %w", cap.Key.Dimension, err)
		}
	}
	return nil
}

// RevalidateQFinalAdmission is read-only. A marked decision must still equal
// q_final, retain its legacy HELD reservation, all five monetary HELD
// reservations and its exact active owner. Missing/corrupt evidence never
// degrades to an ordinary entry.
func (j *Journal) RevalidateQFinalAdmission(ctx context.Context, decisionID string) (required bool, err error) {
	decision, err := j.LookupDecision(ctx, strings.TrimSpace(decisionID))
	if err != nil {
		return false, err
	}
	preimage, err := ParsePreimage(decision.PreimageKind, decision.RiskPreimage)
	if err != nil {
		return false, err
	}
	intent, ok := preimage.(RiskIntent)
	if !ok {
		return false, nil
	}
	_, transactionID, required := splitQFinalPolicyVersion(intent.PolicyVersion)
	if !required {
		return false, nil
	}
	var account, market, symbol, existingID, prospective, lane, campaign string
	var qFinal uint64
	err = j.db.QueryRowContext(ctx, `SELECT account_ref,market,symbol,q_final,existing_reservation_id,owner_prospective_generation,owner_lane_id,owner_campaign_id FROM risk_bucket_final_decisions WHERE decision_id=? AND transaction_id=?`, decision.ID, transactionID).Scan(&account, &market, &symbol, &qFinal, &existingID, &prospective, &lane, &campaign)
	if err != nil {
		return true, fmt.Errorf("%w: q_final decision evidence: %v", ErrRiskBucketReplayMismatch, err)
	}
	quantity, valid := NormalizeDecimal(intent.Quantity)
	if !valid || quantity != strconv.FormatUint(qFinal, 10) || account != decision.AccountRef || strings.EqualFold(market, intent.Market) == false || symbol != strings.ToUpper(intent.Symbol) {
		return true, fmt.Errorf("%w: q_final preimage mismatch", ErrRiskBucketReplayMismatch)
	}
	var legacyState, legacyDecision string
	if err := j.db.QueryRowContext(ctx, `SELECT state,decision_id FROM risk_reservations WHERE id=?`, existingID).Scan(&legacyState, &legacyDecision); err != nil || legacyState != ReservationHeld || legacyDecision != decision.ID {
		return true, fmt.Errorf("%w: q_final aggregate reservation", ErrRiskBucketReplayMismatch)
	}
	key := riskbucket.OwnerKey{AccountID: account, Market: riskbucket.Market(market), Symbol: symbol, ProspectiveGeneration: prospective}
	var activeOwner int
	if err := j.db.QueryRowContext(ctx, `SELECT count(*) FROM risk_bucket_owners WHERE account_ref=? AND market=? AND symbol=? AND prospective_generation=? AND lane_id=? AND campaign_id=? AND released_at IS NULL`, account, market, symbol, prospective, lane, campaign).Scan(&activeOwner); err != nil || activeOwner != 1 {
		return true, fmt.Errorf("%w: q_final owner", ErrRiskBucketReplayMismatch)
	}
	if err := verifyRiskBucketStateDigest(ctx, j.db, key); err != nil {
		return true, err
	}
	rows, err := j.db.QueryContext(ctx, `SELECT bucket_dimension,state,reserved_minor,held_minor,existing_reservation_id,owner_prospective_generation FROM risk_bucket_reservations WHERE decision_id=? ORDER BY bucket_dimension`, decision.ID)
	if err != nil {
		return true, err
	}
	defer rows.Close()
	seen := make(map[riskbucket.Dimension]bool)
	for rows.Next() {
		var dimension riskbucket.Dimension
		var state, reserved, held, linkedExisting, linkedOwner string
		if err := rows.Scan(&dimension, &state, &reserved, &held, &linkedExisting, &linkedOwner); err != nil {
			return true, err
		}
		if seen[dimension] || state != "HELD" || reserved == "" || held != reserved || linkedExisting != existingID || linkedOwner != prospective {
			return true, fmt.Errorf("%w: q_final monetary reservation", ErrRiskBucketReplayMismatch)
		}
		seen[dimension] = true
	}
	if err := rows.Err(); err != nil {
		return true, err
	}
	for _, dimension := range riskbucket.RequiredDimensionOrder() {
		if !seen[dimension] {
			return true, fmt.Errorf("%w: missing %s q_final reservation", ErrRiskBucketReplayMismatch, dimension)
		}
	}
	if len(seen) != len(riskbucket.RequiredDimensionOrder()) {
		return true, fmt.Errorf("%w: q_final reservation dimension set", ErrRiskBucketReplayMismatch)
	}
	return true, nil
}
