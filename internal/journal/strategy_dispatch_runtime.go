package journal

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/JungHoonGhae/tossinvest-cli/internal/positioncampaign"
	"github.com/JungHoonGhae/tossinvest-cli/internal/riskbucket"
)

var (
	ErrStrategyDispatchFenced = errors.New("journal: strategy dispatch owner fenced")
	// ErrStrategyDispatchLeaseConsumed is returned when a claim does not name
	// the exact current ISSUED revision. A terminal or already-claimed lease is
	// never revived for a retry.
	ErrStrategyDispatchLeaseConsumed = errors.New("journal: strategy dispatch lease already consumed or stale")
	// ErrStrategyDispatchLeaseUnavailable is returned when no durable lease can
	// be atomically consumed for the requested identity.
	ErrStrategyDispatchLeaseUnavailable = errors.New("journal: strategy dispatch lease unavailable")
	// ErrStrategyDispatchOwnerBusy prevents a replacement owner from fencing an
	// unexpired SUBMITTING lease while it is allowed to cross broker transport.
	ErrStrategyDispatchOwnerBusy = errors.New("journal: strategy dispatch owner has active transport")
	// ErrStrategyDispatchDormant is returned by authority-minting and send-capable
	// recovery APIs. This checkpoint deliberately has no production mint for
	// activation, ProtectionReady, Gateway outcome, retry, or broker transport.
	ErrStrategyDispatchDormant = errors.New("journal: strategy dispatch authority is dormant")
)

type StrategyDispatchMarket string

const (
	StrategyDispatchMarketKR StrategyDispatchMarket = "KR"
	StrategyDispatchMarketUS StrategyDispatchMarket = "US"
)

type StrategyDispatchLeaseState string

const (
	StrategyDispatchLeaseIssued     StrategyDispatchLeaseState = "ISSUED"
	StrategyDispatchLeaseClaimed    StrategyDispatchLeaseState = "CLAIMED"
	StrategyDispatchLeaseSubmitting StrategyDispatchLeaseState = "SUBMITTING"
	StrategyDispatchLeaseSubmitted  StrategyDispatchLeaseState = "SUBMITTED"
	StrategyDispatchLeaseAmbiguous  StrategyDispatchLeaseState = "AMBIGUOUS"
	StrategyDispatchLeaseRefused    StrategyDispatchLeaseState = "REFUSED"
)

type StrategyDispatchReservationDisposition string

const (
	StrategyDispatchReservationReserved    StrategyDispatchReservationDisposition = "RESERVED"
	StrategyDispatchReservationReleased    StrategyDispatchReservationDisposition = "RELEASED"
	StrategyDispatchReservationTransferred StrategyDispatchReservationDisposition = "TRANSFERRED"
	StrategyDispatchReservationHeld        StrategyDispatchReservationDisposition = "HELD"
)

type StrategyDispatchOwner struct {
	OwnerInstance string
	Epoch         uint64
	FencingToken  string
	Revision      uint64
	AcquiredAt    time.Time
}

// StrategyDispatchMarketAuthority is a persistence shape, not a capability.
// CommitStrategyDispatchMarketAuthority always refuses until an opaque,
// human-approved activation and ProtectionReady adapter exists.
type StrategyDispatchMarketAuthority struct {
	AccountRef string
	Market     StrategyDispatchMarket
	Symbol     string

	ActivationGeneration     uint64
	ActivationDigest         string
	CalendarGeneration       uint64
	ProtectionGeneration     uint64
	ProtectionSerial         string
	ProtectionDigest         string
	ReconciliationGeneration uint64
	RiskPolicyGeneration     uint64
	RiskPolicyDigest         string
	GuardianGeneration       uint64
	GuardianDigest           string
	BuildDigest              string

	ExpectedRevision uint64
	Revision         uint64
	RecordDigest     string
	UpdatedAt        time.Time
}

type StrategyDispatchLeasePlan struct {
	LeaseID, OperationID        string
	AccountRef                  string
	Market                      StrategyDispatchMarket
	Symbol                      string
	CandidateID, EvidenceDigest string
	RouterID, RouterVersion     string
	LaneID, LaneVersion         string
	CampaignID, LegID           string
	RiskReservationID           string
	GuardianDecisionID          string
	OwnerEpoch                  uint64
	FencingToken                string
	AuthorityRevision           uint64
	AuthorityDigest             string
	IssuedAt, ExpiresAt         time.Time
}

type StrategyDispatchLease struct {
	StrategyDispatchLeasePlan
	State                StrategyDispatchLeaseState
	Disposition          StrategyDispatchReservationDisposition
	Revision             uint64
	TransportStartedAt   time.Time
	RefusalCode          string
	OutcomeCode          string
	BrokerOrderID        string
	QueryDigest          string
	OutcomeObservedAt    time.Time
	LeaseDigest          string
	CreatedAt, UpdatedAt time.Time
}

type StrategyDispatchLeaseCAS struct {
	LeaseID          string
	ExpectedRevision uint64
	OwnerEpoch       uint64
	FencingToken     string
}

// strategyDispatchOutcomeCode is derived only by the composite strategy
// dispatch transaction. There is no public caller-outcome contract.
type strategyDispatchOutcomeCode string

const (
	strategyDispatchOutcomeCodeNotSent   strategyDispatchOutcomeCode = "BROKER_NOT_SENT"
	strategyDispatchOutcomeCodeConfirmed strategyDispatchOutcomeCode = "BROKER_CONFIRMED"
	strategyDispatchOutcomeCodeAmbiguous strategyDispatchOutcomeCode = "BROKER_AMBIGUOUS"
)

type strategyDispatchAttemptAuthority struct {
	attemptID, intentID, kind, state, targetOrderID, brokerOrderID string
	decisionID, attemptAccount, clientOrderID                      string
	dispatchStartedAt, settledAt, transitionAt                     string
	intentAccount, market, tradingDay, symbol, side                string
	quantity, currency                                             string
	strategyAttemptID, strategyState                               string
	strategyClientOrderID                                          string
	strategyRevision                                               int64
	campaignID                                                     string
	legSequence, qFinal                                            int64
	prospectiveToken                                               string
}

type strategyDispatchOutcomeTransition struct {
	state       StrategyDispatchLeaseState
	disposition StrategyDispatchReservationDisposition
	code        strategyDispatchOutcomeCode
	refusalCode string
}

type StrategyDispatchRecoveryAction string

const (
	// StrategyDispatchRecoveryRefuseRelease requires a future Gateway-owned
	// NOT_DISPATCHED transaction to refuse the lease and release its real holds.
	StrategyDispatchRecoveryRefuseRelease StrategyDispatchRecoveryAction = "REFUSE_RELEASE_REQUIRED"
	// StrategyDispatchRecoveryAttestedOutcome requires an opaque official exact
	// broker outcome. No constructor for that authority exists in this build.
	StrategyDispatchRecoveryAttestedOutcome StrategyDispatchRecoveryAction = "ATTESTED_OUTCOME_REQUIRED"
)

type StrategyDispatchRecoveryItem struct {
	Lease  StrategyDispatchLease
	Action StrategyDispatchRecoveryAction
}

func (j *Journal) AcquireStrategyDispatchOwner(ctx context.Context, ownerInstance string) (StrategyDispatchOwner, error) {
	if j == nil || j.db == nil || !validStrategyDispatchIdentity(ownerInstance) {
		return StrategyDispatchOwner{}, fmt.Errorf("%w: invalid strategy dispatch owner instance", ErrInvalidRequest)
	}
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return StrategyDispatchOwner{}, fmt.Errorf("journal: generating strategy dispatch fence: %w", err)
	}
	token := hex.EncodeToString(tokenBytes)
	now := j.clk.Now().UTC()
	tx, err := j.db.BeginTx(ctx, nil)
	if err != nil {
		return StrategyDispatchOwner{}, err
	}
	defer tx.Rollback()
	var activeSubmitting int
	// The official transport has a 15-second client timeout. Keep the owner
	// fence for five seconds beyond that bound after lease expiry so a call that
	// began just before expiry can settle before crash recovery advances epoch.
	takeoverCutoff := now.Add(-strategyDispatchOwnerTakeoverGrace)
	if err := tx.QueryRowContext(ctx, `SELECT count(*) FROM strategy_dispatch_leases
		WHERE state='SUBMITTING' AND disposition='RESERVED' AND expires_at>?`, formatJournalTime(takeoverCutoff)).Scan(&activeSubmitting); err != nil {
		return StrategyDispatchOwner{}, fmt.Errorf("journal: checking active strategy transport: %w", err)
	}
	if activeSubmitting != 0 {
		return StrategyDispatchOwner{}, ErrStrategyDispatchOwnerBusy
	}
	var epoch, revision uint64
	err = tx.QueryRowContext(ctx, `SELECT owner_epoch,revision FROM strategy_dispatch_owner_current WHERE owner_key='CENTRAL'`).Scan(&epoch, &revision)
	if errors.Is(err, sql.ErrNoRows) {
		epoch, revision, err = 1, 1, nil
	} else if err == nil {
		epoch++
		revision++
	}
	if err != nil {
		return StrategyDispatchOwner{}, fmt.Errorf("journal: reading current strategy dispatch owner: %w", err)
	}
	if _, err = tx.ExecContext(ctx, `INSERT INTO strategy_dispatch_owner_epochs(owner_epoch,fencing_token,owner_instance,acquired_at) VALUES(?,?,?,?)`, epoch, token, ownerInstance, formatJournalTime(now)); err != nil {
		return StrategyDispatchOwner{}, fmt.Errorf("journal: recording strategy dispatch owner epoch: %w", err)
	}
	var result sql.Result
	if epoch == 1 {
		result, err = tx.ExecContext(ctx, `INSERT INTO strategy_dispatch_owner_current(owner_key,owner_epoch,fencing_token,owner_instance,revision,acquired_at) VALUES('CENTRAL',?,?,?,?,?)`, epoch, token, ownerInstance, revision, formatJournalTime(now))
	} else {
		result, err = tx.ExecContext(ctx, `UPDATE strategy_dispatch_owner_current SET owner_epoch=?,fencing_token=?,owner_instance=?,revision=?,acquired_at=? WHERE owner_key='CENTRAL' AND revision=?`, epoch, token, ownerInstance, revision, formatJournalTime(now), revision-1)
	}
	if err != nil {
		return StrategyDispatchOwner{}, fmt.Errorf("journal: advancing strategy dispatch owner: %w", err)
	}
	if err := requireStrategyDispatchOneRow(result); err != nil {
		return StrategyDispatchOwner{}, err
	}
	if err = tx.Commit(); err != nil {
		return StrategyDispatchOwner{}, err
	}
	return StrategyDispatchOwner{OwnerInstance: ownerInstance, Epoch: epoch, FencingToken: token, Revision: revision, AcquiredAt: now}, nil
}

// CommitStrategyDispatchMarketAuthority is intentionally unissuable. Caller
// strings are not activation or ProtectionReady authority.
func (j *Journal) CommitStrategyDispatchMarketAuthority(context.Context, StrategyDispatchMarketAuthority) (StrategyDispatchMarketAuthority, error) {
	return StrategyDispatchMarketAuthority{}, ErrStrategyDispatchDormant
}

// IssueStrategyDispatchLease is intentionally unissuable. The v25 schema
// validates a future adapter's rows against the actual q_final Guardian
// decision, aggregate HELD reservation, five monetary HELD reservations, and
// exact active owner, but this build exposes no authority mint.
func (j *Journal) IssueStrategyDispatchLease(context.Context, StrategyDispatchLeasePlan) (StrategyDispatchLease, error) {
	return StrategyDispatchLease{}, ErrStrategyDispatchDormant
}

// ClaimStrategyDispatchLease is the transport-free central-owner gate. It
// reloads every durable authority inside one transaction. Exact authority moves
// ISSUED to CLAIMED; any current-authority refusal irreversibly moves the lease
// and its real aggregate/five-bucket holds to REFUSED + RELEASED. This method
// cannot start transport, mint a lease, or manufacture Gateway authority.
func (j *Journal) ClaimStrategyDispatchLease(ctx context.Context, request StrategyDispatchLeaseCAS) (StrategyDispatchLease, error) {
	if j == nil || j.db == nil || !validStrategyDispatchIdentity(request.LeaseID) ||
		request.ExpectedRevision == 0 || request.OwnerEpoch == 0 || !validStrategyDispatchIdentity(request.FencingToken) {
		return StrategyDispatchLease{}, fmt.Errorf("%w: invalid strategy dispatch lease claim", ErrInvalidRequest)
	}
	tx, err := j.db.BeginTx(ctx, nil)
	if err != nil {
		return StrategyDispatchLease{}, fmt.Errorf("journal: begin strategy dispatch claim: %w", err)
	}
	defer tx.Rollback()
	if err := requireCurrentStrategyDispatchOwner(ctx, tx, request.OwnerEpoch, request.FencingToken); err != nil {
		return StrategyDispatchLease{}, err
	}
	lease, err := loadStrategyDispatchLease(ctx, tx, request.LeaseID)
	if errors.Is(err, sql.ErrNoRows) {
		return StrategyDispatchLease{}, ErrStrategyDispatchLeaseUnavailable
	}
	if err != nil {
		return StrategyDispatchLease{}, err
	}
	if lease.State != StrategyDispatchLeaseIssued {
		return StrategyDispatchLease{}, ErrStrategyDispatchLeaseConsumed
	}
	now := j.clk.Now().UTC()
	refusalCode, err := validateStrategyDispatchClaimAuthority(ctx, tx, lease, request, now)
	if err != nil {
		return StrategyDispatchLease{}, err
	}
	if refusalCode != "" {
		lease, err = refuseStrategyDispatchClaimTx(ctx, tx, lease, refusalCode, now)
	} else {
		lease, err = claimStrategyDispatchLeaseTx(ctx, tx, lease, now)
	}
	if err != nil {
		return StrategyDispatchLease{}, err
	}
	if err := tx.Commit(); err != nil {
		return StrategyDispatchLease{}, fmt.Errorf("journal: commit strategy dispatch claim: %w", err)
	}
	return lease, nil
}

func validateStrategyDispatchClaimAuthority(ctx context.Context, tx *sql.Tx, lease StrategyDispatchLease, request StrategyDispatchLeaseCAS, now time.Time) (string, error) {
	if lease.Revision != request.ExpectedRevision {
		return "LEASE_REVISION_STALE", nil
	}
	if lease.OwnerEpoch != request.OwnerEpoch || lease.FencingToken != request.FencingToken {
		return "OWNER_STALE", nil
	}
	if now.Before(lease.IssuedAt) {
		return "LEASE_NOT_YET_VALID", nil
	}
	if !now.Before(lease.ExpiresAt) {
		return "LEASE_EXPIRED", nil
	}
	marketAuthority, err := strategyDispatchClaimExists(ctx, tx, `SELECT EXISTS(
		SELECT 1 FROM strategy_dispatch_market_authorities
		WHERE account_ref=? AND market=? AND symbol=? AND revision=? AND record_digest=?)`,
		lease.AccountRef, lease.Market, lease.Symbol, lease.AuthorityRevision, lease.AuthorityDigest)
	if err != nil {
		return "", err
	}
	if !marketAuthority {
		return "MARKET_AUTHORITY_CHANGED", nil
	}
	bound, err := strategyDispatchClaimExists(ctx, tx, `SELECT EXISTS(
		SELECT 1 FROM strategy_first_leg_bindings binding
		WHERE binding.decision_id=?
		  AND binding.aggregate_reservation_id=?
		  AND binding.account_ref=? AND binding.market=? AND binding.symbol=?
		  AND binding.candidate_id=? AND binding.evidence_digest=?
		  AND binding.lane_id=? AND binding.lane_version=?
		  AND binding.router_id=? AND binding.router_version=?
		  AND binding.campaign_id=? AND binding.leg_plan_id=?)`,
		lease.GuardianDecisionID, lease.RiskReservationID, lease.AccountRef, lease.Market, lease.Symbol,
		lease.CandidateID, lease.EvidenceDigest, lease.LaneID, lease.LaneVersion, lease.RouterID, lease.RouterVersion,
		lease.CampaignID, lease.LegID)
	if err != nil {
		return "", err
	}
	if !bound {
		return "FIRST_LEG_BINDING_CHANGED", nil
	}
	current, err := strategyDispatchClaimExists(ctx, tx, `SELECT EXISTS(
		SELECT 1
		FROM strategy_first_leg_bindings binding
		JOIN risk_bucket_final_decisions q ON q.decision_id=binding.decision_id
		JOIN decisions risk_intent ON risk_intent.id=q.decision_id
		JOIN risk_reservations aggregate_hold
		  ON aggregate_hold.id=binding.aggregate_reservation_id
		 AND aggregate_hold.decision_id=q.decision_id
		JOIN strategy_attempt_lineage attempt ON attempt.attempt_id=binding.attempt_id
		JOIN strategy_decision_lineage strategy ON strategy.entry_decision_identity=binding.entry_decision_identity
		JOIN position_campaigns campaign ON campaign.id=binding.campaign_id
		JOIN position_campaign_claims claim ON claim.campaign_id=binding.campaign_id
		JOIN campaign_legs leg ON leg.campaign_id=binding.campaign_id AND leg.sequence=binding.leg_sequence
		JOIN risk_bucket_owners owner
		  ON owner.account_ref=binding.account_ref AND owner.market=binding.market
		 AND owner.symbol=binding.symbol AND owner.prospective_generation=binding.prospective_token
		WHERE binding.decision_id=?
		  AND q.existing_reservation_id=binding.aggregate_reservation_id
		  AND q.account_ref=binding.account_ref AND q.market=binding.market AND q.symbol=binding.symbol
		  AND q.owner_lane_id=binding.lane_id AND q.owner_campaign_id=binding.campaign_id
		  AND q.owner_prospective_generation=binding.prospective_token AND q.q_final=binding.q_final
		  AND risk_intent.account_ref=binding.account_ref
		  AND risk_intent.safety_class='EXPOSURE_RAISING' AND risk_intent.preimage_kind='RISK_INTENT'
		  AND aggregate_hold.account_ref=binding.account_ref AND aggregate_hold.state='HELD'
		  AND attempt.risk_intent_id=binding.decision_id AND attempt.guardian_decision_id=binding.decision_id
		  AND attempt.account_ref=binding.account_ref
		  AND attempt.entry_decision_identity=binding.entry_decision_identity
		  AND attempt.activation_manifest_digest=strategy.activation_manifest_digest
		  AND attempt.client_order_id=risk_intent.client_order_id
		  AND attempt.client_order_id=?
		  AND binding.router_id=? AND binding.router_version=?
		  AND attempt.state='PLANNED'
		  AND upper(strategy.market)=binding.market AND strategy.symbol=binding.symbol
		  AND strategy.candidate_life_id=binding.candidate_id AND strategy.evidence_digest=binding.evidence_digest
		  AND strategy.lane_id=binding.lane_id AND strategy.lane_version=binding.lane_version
		  AND strategy.quantity=CAST(binding.q_final AS TEXT)
		  AND campaign.decision_id=binding.decision_id AND campaign.account_ref=binding.account_ref
		  AND lower(campaign.market)=lower(binding.market) AND campaign.symbol=binding.symbol
		  AND campaign.lane_id=binding.lane_id AND campaign.lane_version=binding.lane_version
		  AND campaign.evidence_digest=binding.evidence_digest AND campaign.prospective_token=binding.prospective_token
		  AND campaign.state='PLANNED' AND campaign.entry_blocked=0
		  AND claim.account_ref=binding.account_ref AND lower(claim.market)=lower(binding.market)
		  AND claim.symbol=binding.symbol
		  AND claim.position_generation=campaign.expected_position_generation
		  AND claim.position_version=campaign.expected_position_version
		  AND claim.prospective_token=binding.prospective_token
		  AND leg.plan_id=binding.leg_plan_id AND leg.requested_quantity=CAST(binding.q_final AS TEXT)
		  AND leg.residual_quantity=CAST(binding.q_final AS TEXT) AND leg.filled_quantity='0' AND leg.state='PLANNED'
		  AND owner.lane_id=binding.lane_id AND owner.campaign_id=binding.campaign_id AND owner.released_at IS NULL
		  AND (SELECT count(*) FROM risk_bucket_reservations monetary
		       WHERE monetary.decision_id=binding.decision_id
		         AND monetary.existing_reservation_id=binding.aggregate_reservation_id
		         AND monetary.account_ref=binding.account_ref AND monetary.market=binding.market
		         AND monetary.symbol=binding.symbol
		         AND monetary.owner_prospective_generation=binding.prospective_token
		         AND monetary.state='HELD' AND monetary.held_minor=monetary.reserved_minor)=5
		  AND (SELECT count(DISTINCT monetary.bucket_dimension) FROM risk_bucket_reservations monetary
		       WHERE monetary.decision_id=binding.decision_id
		         AND monetary.bucket_dimension IN ('horizon','market','strategy','sector','symbol'))=5)`,
		lease.GuardianDecisionID, lease.OperationID, lease.RouterID, lease.RouterVersion)
	if err != nil {
		return "", err
	}
	if !current {
		return "FIRST_LEG_AUTHORITY_CHANGED", nil
	}
	return "", nil
}

func strategyDispatchClaimExists(ctx context.Context, tx *sql.Tx, query string, args ...any) (bool, error) {
	var exists int
	if err := tx.QueryRowContext(ctx, query, args...).Scan(&exists); err != nil {
		return false, fmt.Errorf("journal: revalidating strategy dispatch authority: %w", err)
	}
	return exists == 1, nil
}

func claimStrategyDispatchLeaseTx(ctx context.Context, tx *sql.Tx, lease StrategyDispatchLease, now time.Time) (StrategyDispatchLease, error) {
	nextRevision := lease.Revision + 1
	result, err := tx.ExecContext(ctx, `UPDATE strategy_dispatch_leases
		SET state='CLAIMED',revision=?,updated_at=?
		WHERE lease_id=? AND state='ISSUED' AND disposition='RESERVED' AND revision=?
		  AND owner_epoch=? AND fencing_token=?`,
		nextRevision, formatJournalTime(now), lease.LeaseID, lease.Revision, lease.OwnerEpoch, lease.FencingToken)
	if err != nil {
		return StrategyDispatchLease{}, fmt.Errorf("journal: claiming strategy dispatch lease: %w", err)
	}
	if err := requireStrategyDispatchOneRow(result); err != nil {
		return StrategyDispatchLease{}, err
	}
	if err := appendStrategyDispatchTransitionTx(ctx, tx, lease, StrategyDispatchLeaseClaimed,
		StrategyDispatchReservationReserved, "AUTHORITY_CURRENT", "", now); err != nil {
		return StrategyDispatchLease{}, err
	}
	return loadStrategyDispatchLease(ctx, tx, lease.LeaseID)
}

func refuseStrategyDispatchClaimTx(ctx context.Context, tx *sql.Tx, lease StrategyDispatchLease, refusalCode string, now time.Time) (StrategyDispatchLease, error) {
	nowText := formatJournalTime(now)
	aggregateResult, err := tx.ExecContext(ctx, `UPDATE risk_reservations
		SET state='RELEASED',released_at=?,release_reason=?
		WHERE id=? AND decision_id=? AND account_ref=? AND state='HELD'`,
		nowText, ReleaseReasonExpiredUnconsumed, lease.RiskReservationID, lease.GuardianDecisionID, lease.AccountRef)
	if err != nil {
		return StrategyDispatchLease{}, fmt.Errorf("journal: releasing refused aggregate reservation: %w", err)
	}
	if err := requireStrategyDispatchAtMostRows(aggregateResult, 1, "aggregate reservation release"); err != nil {
		return StrategyDispatchLease{}, err
	}
	var releasedAggregate int
	if err := tx.QueryRowContext(ctx, `SELECT count(*) FROM risk_reservations
		WHERE id=? AND decision_id=? AND account_ref=? AND state='RELEASED'`,
		lease.RiskReservationID, lease.GuardianDecisionID, lease.AccountRef).Scan(&releasedAggregate); err != nil {
		return StrategyDispatchLease{}, fmt.Errorf("journal: proving refused aggregate reservation: %w", err)
	}
	if releasedAggregate != 1 {
		return StrategyDispatchLease{}, fmt.Errorf("%w: aggregate release proof count=%d want=1", ErrStrategyDispatchLeaseUnavailable, releasedAggregate)
	}
	bucketResult, err := tx.ExecContext(ctx, `UPDATE risk_bucket_reservations
		SET state='RELEASED',held_minor='0',updated_at=?
		WHERE decision_id=? AND existing_reservation_id=? AND account_ref=? AND market=? AND symbol=? AND state='HELD'`,
		nowText, lease.GuardianDecisionID, lease.RiskReservationID, lease.AccountRef, lease.Market, lease.Symbol)
	if err != nil {
		return StrategyDispatchLease{}, fmt.Errorf("journal: releasing refused bucket reservations: %w", err)
	}
	if err := requireStrategyDispatchAtMostRows(bucketResult, 5, "five-bucket reservation release"); err != nil {
		return StrategyDispatchLease{}, err
	}
	var releasedBuckets, releasedDimensions int
	if err := tx.QueryRowContext(ctx, `SELECT count(*),count(DISTINCT bucket_dimension)
		FROM risk_bucket_reservations
		WHERE decision_id=? AND existing_reservation_id=? AND account_ref=? AND market=? AND symbol=?
		  AND state='RELEASED' AND held_minor='0'
		  AND bucket_dimension IN ('horizon','market','strategy','sector','symbol')`,
		lease.GuardianDecisionID, lease.RiskReservationID, lease.AccountRef, lease.Market, lease.Symbol).
		Scan(&releasedBuckets, &releasedDimensions); err != nil {
		return StrategyDispatchLease{}, fmt.Errorf("journal: proving refused bucket reservations: %w", err)
	}
	if releasedBuckets != 5 || releasedDimensions != 5 {
		return StrategyDispatchLease{}, fmt.Errorf("%w: five-bucket release proof rows=%d dimensions=%d want=5/5",
			ErrStrategyDispatchLeaseUnavailable, releasedBuckets, releasedDimensions)
	}
	nextRevision := lease.Revision + 1
	result, err := tx.ExecContext(ctx, `UPDATE strategy_dispatch_leases
		SET state='REFUSED',disposition='RELEASED',revision=?,refusal_code=?,updated_at=?
		WHERE lease_id=? AND state='ISSUED' AND disposition='RESERVED' AND revision=?
		  AND owner_epoch=? AND fencing_token=?`,
		nextRevision, refusalCode, nowText, lease.LeaseID, lease.Revision, lease.OwnerEpoch, lease.FencingToken)
	if err != nil {
		return StrategyDispatchLease{}, fmt.Errorf("journal: refusing strategy dispatch lease: %w", err)
	}
	if err := requireStrategyDispatchOneRow(result); err != nil {
		return StrategyDispatchLease{}, err
	}
	if err := appendStrategyDispatchTransitionTx(ctx, tx, lease, StrategyDispatchLeaseRefused,
		StrategyDispatchReservationReleased, refusalCode, "", now); err != nil {
		return StrategyDispatchLease{}, err
	}
	return loadStrategyDispatchLease(ctx, tx, lease.LeaseID)
}

func appendStrategyDispatchTransitionTx(ctx context.Context, tx *sql.Tx, lease StrategyDispatchLease,
	toState StrategyDispatchLeaseState, toDisposition StrategyDispatchReservationDisposition,
	transitionCode, queryDigest string, observedAt time.Time,
) error {
	nextRevision := lease.Revision + 1
	payload := strings.Join([]string{
		lease.LeaseID, string(lease.State), string(toState), string(lease.Disposition), string(toDisposition),
		fmt.Sprintf("%d", lease.Revision), fmt.Sprintf("%d", nextRevision), transitionCode,
		lease.OperationID, queryDigest, formatJournalTime(observedAt),
	}, "\x00")
	digest := sha256.Sum256([]byte(payload))
	_, err := tx.ExecContext(ctx, `INSERT INTO strategy_dispatch_outcomes(
		outcome_id,lease_id,from_state,to_state,from_disposition,to_disposition,expected_revision,next_revision,
		transition_code,operation_identity,broker_order_id,query_digest,observed_at,record_digest)
		VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		fmt.Sprintf("dispatch-transition:%s:%d", lease.LeaseID, nextRevision), lease.LeaseID,
		lease.State, toState, lease.Disposition, toDisposition, lease.Revision, nextRevision,
		transitionCode, lease.OperationID, "", queryDigest, formatJournalTime(observedAt), hex.EncodeToString(digest[:]))
	if err != nil {
		return fmt.Errorf("journal: recording strategy dispatch transition: %w", err)
	}
	return nil
}

// BeginStrategyDispatchSubmitting is the last journal-only fence before a
// Gateway may become send-capable. It repeats the exact current authority used
// by claim in one transaction. Current authority records a durable transport
// start marker; drift is consumed before transport as REFUSED + RELEASED. This
// method never invokes a Gateway or broker transport.
func (j *Journal) BeginStrategyDispatchSubmitting(ctx context.Context, request StrategyDispatchLeaseCAS) (StrategyDispatchLease, error) {
	if j == nil || j.db == nil || !validStrategyDispatchIdentity(request.LeaseID) ||
		request.ExpectedRevision == 0 || request.OwnerEpoch == 0 || !validStrategyDispatchIdentity(request.FencingToken) {
		return StrategyDispatchLease{}, fmt.Errorf("%w: invalid strategy dispatch submitting CAS", ErrInvalidRequest)
	}
	tx, err := j.db.BeginTx(ctx, nil)
	if err != nil {
		return StrategyDispatchLease{}, fmt.Errorf("journal: begin strategy dispatch submitting: %w", err)
	}
	defer tx.Rollback()
	if err := requireCurrentStrategyDispatchOwner(ctx, tx, request.OwnerEpoch, request.FencingToken); err != nil {
		return StrategyDispatchLease{}, err
	}
	lease, err := loadStrategyDispatchLease(ctx, tx, request.LeaseID)
	if errors.Is(err, sql.ErrNoRows) {
		return StrategyDispatchLease{}, ErrStrategyDispatchLeaseUnavailable
	}
	if err != nil {
		return StrategyDispatchLease{}, err
	}
	if lease.State != StrategyDispatchLeaseClaimed {
		return StrategyDispatchLease{}, ErrStrategyDispatchLeaseConsumed
	}
	now := j.clk.Now().UTC()
	refusalCode, err := validateStrategyDispatchClaimAuthority(ctx, tx, lease, request, now)
	if err != nil {
		return StrategyDispatchLease{}, err
	}
	if refusalCode != "" {
		lease, err = refuseClaimedStrategyDispatchSubmittingTx(ctx, tx, lease, refusalCode, now)
	} else {
		lease, err = beginStrategyDispatchSubmittingTx(ctx, tx, lease, now)
	}
	if err != nil {
		return StrategyDispatchLease{}, err
	}
	if err := tx.Commit(); err != nil {
		return StrategyDispatchLease{}, fmt.Errorf("journal: commit strategy dispatch submitting: %w", err)
	}
	return lease, nil
}

func beginStrategyDispatchSubmittingTx(ctx context.Context, tx *sql.Tx, lease StrategyDispatchLease, now time.Time) (StrategyDispatchLease, error) {
	nextRevision := lease.Revision + 1
	nowText := formatJournalTime(now)
	result, err := tx.ExecContext(ctx, `UPDATE strategy_dispatch_leases
		SET state='SUBMITTING',revision=?,transport_started_at=?,updated_at=?
		WHERE lease_id=? AND state='CLAIMED' AND disposition='RESERVED' AND revision=?
		  AND owner_epoch=? AND fencing_token=? AND transport_started_at IS NULL`,
		nextRevision, nowText, nowText, lease.LeaseID, lease.Revision, lease.OwnerEpoch, lease.FencingToken)
	if err != nil {
		return StrategyDispatchLease{}, fmt.Errorf("journal: marking strategy dispatch transport start: %w", err)
	}
	if err := requireStrategyDispatchOneRow(result); err != nil {
		return StrategyDispatchLease{}, err
	}
	if err := appendStrategyDispatchTransitionTx(ctx, tx, lease, StrategyDispatchLeaseSubmitting,
		StrategyDispatchReservationReserved, "TRANSPORT_START_CURRENT", "", now); err != nil {
		return StrategyDispatchLease{}, err
	}
	return loadStrategyDispatchLease(ctx, tx, lease.LeaseID)
}

func refuseClaimedStrategyDispatchSubmittingTx(ctx context.Context, tx *sql.Tx, lease StrategyDispatchLease, refusalCode string, now time.Time) (StrategyDispatchLease, error) {
	nowText := formatJournalTime(now)
	aggregateResult, err := tx.ExecContext(ctx, `UPDATE risk_reservations
		SET state='RELEASED',released_at=?,release_reason=?
		WHERE id=? AND decision_id=? AND account_ref=? AND state='HELD'`,
		nowText, ReleaseReasonExpiredUnconsumed, lease.RiskReservationID, lease.GuardianDecisionID, lease.AccountRef)
	if err != nil {
		return StrategyDispatchLease{}, fmt.Errorf("journal: releasing submitting-refused aggregate reservation: %w", err)
	}
	if err := requireStrategyDispatchAtMostRows(aggregateResult, 1, "submitting-refused aggregate reservation release"); err != nil {
		return StrategyDispatchLease{}, err
	}
	var releasedAggregate int
	if err := tx.QueryRowContext(ctx, `SELECT count(*) FROM risk_reservations
		WHERE id=? AND decision_id=? AND account_ref=? AND state='RELEASED'`,
		lease.RiskReservationID, lease.GuardianDecisionID, lease.AccountRef).Scan(&releasedAggregate); err != nil {
		return StrategyDispatchLease{}, fmt.Errorf("journal: proving submitting-refused aggregate reservation: %w", err)
	}
	if releasedAggregate != 1 {
		return StrategyDispatchLease{}, fmt.Errorf("%w: submitting aggregate release proof count=%d want=1", ErrStrategyDispatchLeaseUnavailable, releasedAggregate)
	}
	bucketResult, err := tx.ExecContext(ctx, `UPDATE risk_bucket_reservations
		SET state='RELEASED',held_minor='0',updated_at=?
		WHERE decision_id=? AND existing_reservation_id=? AND account_ref=? AND market=? AND symbol=? AND state='HELD'`,
		nowText, lease.GuardianDecisionID, lease.RiskReservationID, lease.AccountRef, lease.Market, lease.Symbol)
	if err != nil {
		return StrategyDispatchLease{}, fmt.Errorf("journal: releasing submitting-refused bucket reservations: %w", err)
	}
	if err := requireStrategyDispatchAtMostRows(bucketResult, 5, "submitting-refused five-bucket reservation release"); err != nil {
		return StrategyDispatchLease{}, err
	}
	var releasedBuckets, releasedDimensions int
	if err := tx.QueryRowContext(ctx, `SELECT count(*),count(DISTINCT bucket_dimension)
		FROM risk_bucket_reservations
		WHERE decision_id=? AND existing_reservation_id=? AND account_ref=? AND market=? AND symbol=?
		  AND state='RELEASED' AND held_minor='0'
		  AND bucket_dimension IN ('horizon','market','strategy','sector','symbol')`,
		lease.GuardianDecisionID, lease.RiskReservationID, lease.AccountRef, lease.Market, lease.Symbol).
		Scan(&releasedBuckets, &releasedDimensions); err != nil {
		return StrategyDispatchLease{}, fmt.Errorf("journal: proving submitting-refused bucket reservations: %w", err)
	}
	if releasedBuckets != 5 || releasedDimensions != 5 {
		return StrategyDispatchLease{}, fmt.Errorf("%w: submitting five-bucket release proof rows=%d dimensions=%d want=5/5",
			ErrStrategyDispatchLeaseUnavailable, releasedBuckets, releasedDimensions)
	}
	nextRevision := lease.Revision + 1
	result, err := tx.ExecContext(ctx, `UPDATE strategy_dispatch_leases
		SET state='REFUSED',disposition='RELEASED',revision=?,refusal_code=?,updated_at=?
		WHERE lease_id=? AND state='CLAIMED' AND disposition='RESERVED' AND revision=?
		  AND owner_epoch=? AND fencing_token=? AND transport_started_at IS NULL`,
		nextRevision, refusalCode, nowText, lease.LeaseID, lease.Revision, lease.OwnerEpoch, lease.FencingToken)
	if err != nil {
		return StrategyDispatchLease{}, fmt.Errorf("journal: refusing strategy dispatch before submitting: %w", err)
	}
	if err := requireStrategyDispatchOneRow(result); err != nil {
		return StrategyDispatchLease{}, err
	}
	if err := appendStrategyDispatchTransitionTx(ctx, tx, lease, StrategyDispatchLeaseRefused,
		StrategyDispatchReservationReleased, refusalCode, "", now); err != nil {
		return StrategyDispatchLease{}, err
	}
	return loadStrategyDispatchLease(ctx, tx, lease.LeaseID)
}

const strategyDispatchSettlementTimeout = 15 * time.Second
const strategyDispatchOwnerTakeoverGrace = 20 * time.Second

// DispatchStrategyVerified is the additive strategy-only dispatch path. Unlike
// the released DispatchVerified path, its last pre-send fence binds the core
// attempt and strategy lease, and its after-send composite transaction owns the
// core outcome, real order mappings, missed-fill repair and lease disposition.
func (a *Attempt) DispatchStrategyVerified(ctx context.Context, leaseCAS StrategyDispatchLeaseCAS,
	dispatch DispatchFunc, verify ExistenceCheck,
) (Result, error) {
	if a == nil || a.j == nil || dispatch == nil || verify == nil {
		return Result{}, fmt.Errorf("%w: strategy dispatch and official verifier are required", ErrInvalidRequest)
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	res := Result{AttemptID: a.id, IntentID: a.intentID}
	lease, err := a.beginStrategyDispatchTx(ctx, leaseCAS)
	if err != nil {
		return res, err
	}
	a.state = StateDispatchStarted

	outcome := dispatch(ctx, a)
	res.Class, res.BrokerOrderID, res.Detail, res.Err = outcome.Class, outcome.BrokerOrderID, outcome.Detail, outcome.Err
	settlementCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), strategyDispatchSettlementTimeout)
	defer cancel()

	var target AttemptState
	switch outcome.Class {
	case DispatchNotSent:
		target = StateNotDispatched
		res.ReasonCode = firstNonEmpty(outcome.ReasonCode, ReasonDispatchNotSent)
	case DispatchRejected:
		target = StateFailedConfirmed
		res.ReasonCode = firstNonEmpty(outcome.ReasonCode, ReasonDispatchRejected)
	case DispatchAcked:
		if strings.TrimSpace(outcome.BrokerOrderID) == "" {
			target, res.ReasonCode = StateInDoubt, ReasonAckWithoutOrderID
			res.Detail = "broker acknowledged the strategy mutation without an order id"
			break
		}
		if err := a.strategyTransitionOnly(settlementCtx, StateAcked, outcome.BrokerOrderID, "", "", false); err != nil {
			return res, err
		}
		a.state = StateAcked
		if verifyErr := verify(settlementCtx, outcome.BrokerOrderID); verifyErr != nil {
			target, res.ReasonCode = StateInDoubt, ReasonAckRoundTripUnconfirmed
			res.Detail = fmt.Sprintf("the broker acknowledged order %q but reading it back did not confirm it: %v",
				outcome.BrokerOrderID, verifyErr)
			res.Err = verifyErr
			break
		}
		target = StateConfirmed
		res.ReasonCode = firstNonEmpty(outcome.ReasonCode, ReasonBrokerAcknowledged)
	case DispatchAmbiguous:
		target = StateInDoubt
		res.ReasonCode = firstNonEmpty(outcome.ReasonCode, ReasonDispatchAmbiguous)
	default:
		target, res.ReasonCode = StateInDoubt, ReasonUnknownDispatchClass
		res.Detail = fmt.Sprintf("dispatch reported class %q", outcome.Class)
	}

	terminal, err := a.settleStrategyDispatchTx(settlementCtx, lease, target, res.ReasonCode, res.Detail)
	if err != nil {
		return res, err
	}
	a.state = target
	res.Final = target
	res.BrokerOrderID = terminal.BrokerOrderID
	return res, nil
}

func (a *Attempt) beginStrategyDispatchTx(ctx context.Context, request StrategyDispatchLeaseCAS) (StrategyDispatchLease, error) {
	if !validStrategyDispatchIdentity(request.LeaseID) || request.ExpectedRevision == 0 || request.OwnerEpoch == 0 ||
		!validStrategyDispatchIdentity(request.FencingToken) {
		return StrategyDispatchLease{}, fmt.Errorf("%w: invalid strategy pre-send CAS", ErrInvalidRequest)
	}
	now := a.j.clk.Now().UTC()
	tx, err := a.j.db.BeginTx(ctx, nil)
	if err != nil {
		return StrategyDispatchLease{}, err
	}
	defer tx.Rollback()
	if err := requireCurrentStrategyDispatchOwner(ctx, tx, request.OwnerEpoch, request.FencingToken); err != nil {
		return StrategyDispatchLease{}, err
	}
	lease, err := loadStrategyDispatchLease(ctx, tx, request.LeaseID)
	if err != nil {
		return StrategyDispatchLease{}, err
	}
	if lease.State != StrategyDispatchLeaseClaimed || lease.Disposition != StrategyDispatchReservationReserved ||
		lease.Revision != request.ExpectedRevision {
		return StrategyDispatchLease{}, ErrStrategyDispatchLeaseConsumed
	}
	if lease.OwnerEpoch != request.OwnerEpoch || lease.FencingToken != request.FencingToken {
		return StrategyDispatchLease{}, ErrStrategyDispatchFenced
	}
	if _, err := loadStrategyDispatchAttemptAuthority(ctx, tx, lease, a.id); err != nil {
		return StrategyDispatchLease{}, err
	}
	refusalCode, err := validateStrategyDispatchClaimAuthority(ctx, tx, lease, request, now)
	if err != nil {
		return StrategyDispatchLease{}, err
	}
	if refusalCode != "" {
		// Attempt-owned dispatch never terminalizes only the lease here. The
		// Gateway closes this prepared core plus its exact claim and six holds in
		// RefuseClaimedStrategyPreTransport's single transaction.
		return StrategyDispatchLease{}, fmt.Errorf("%w: strategy pre-send authority %s", ErrStrategyDispatchLeaseUnavailable, refusalCode)
	}
	var current AttemptState
	if err := tx.QueryRowContext(ctx, `SELECT state FROM mutation_attempts WHERE id=?`, a.id).Scan(&current); err != nil {
		return StrategyDispatchLease{}, err
	}
	if err := checkTransitionAllowed(a.id, current, StateDispatchStarted, []AttemptState{StateRecorded}); err != nil {
		return StrategyDispatchLease{}, err
	}
	nowText := formatJournalTime(now)
	if err := consumeDecisionNonce(ctx, tx, a.id, nowText); err != nil {
		return StrategyDispatchLease{}, err
	}
	result, err := tx.ExecContext(ctx, `UPDATE mutation_attempts SET state='DISPATCH_STARTED',dispatch_started_at=?
		WHERE id=? AND state='RECORDED'`, nowText, a.id)
	if err != nil {
		return StrategyDispatchLease{}, err
	}
	if err := requireStrategyDispatchExactRows(result, 1, "strategy core dispatch start"); err != nil {
		return StrategyDispatchLease{}, err
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO attempt_transitions(attempt_id,from_state,to_state,at,reason_code,detail)
		VALUES(?,?,?,?,?,?)`, a.id, StateRecorded, StateDispatchStarted, nowText, "", ""); err != nil {
		return StrategyDispatchLease{}, err
	}
	submitting, err := beginStrategyDispatchSubmittingTx(ctx, tx, lease, now)
	if err != nil {
		return StrategyDispatchLease{}, err
	}
	if err := tx.Commit(); err != nil {
		return StrategyDispatchLease{}, err
	}
	return submitting, nil
}

func (a *Attempt) strategyTransitionOnly(ctx context.Context, to AttemptState, brokerOrderID, reasonCode, detail string,
	settled bool,
) error {
	tx, err := a.j.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err := transitionStrategyAttemptTx(ctx, tx, a, to, brokerOrderID, reasonCode, detail, settled,
		a.j.clk.Now().UTC()); err != nil {
		return err
	}
	return tx.Commit()
}

func (a *Attempt) settleStrategyDispatchTx(ctx context.Context, lease StrategyDispatchLease, to AttemptState,
	reasonCode, detail string,
) (StrategyDispatchLease, error) {
	tx, err := a.j.db.BeginTx(ctx, nil)
	if err != nil {
		return StrategyDispatchLease{}, err
	}
	defer tx.Rollback()
	if err := requireCurrentStrategyDispatchOwner(ctx, tx, lease.OwnerEpoch, lease.FencingToken); err != nil {
		return StrategyDispatchLease{}, err
	}
	loaded, err := loadStrategyDispatchLease(ctx, tx, lease.LeaseID)
	if err != nil {
		return StrategyDispatchLease{}, err
	}
	if loaded.State != StrategyDispatchLeaseSubmitting || loaded.Disposition != StrategyDispatchReservationReserved ||
		loaded.Revision != lease.Revision || loaded.OwnerEpoch != lease.OwnerEpoch || loaded.FencingToken != lease.FencingToken {
		return StrategyDispatchLease{}, ErrStrategyDispatchLeaseConsumed
	}
	now := a.j.clk.Now().UTC()
	if _, err := transitionStrategyAttemptTx(ctx, tx, a, to, "", reasonCode, detail, to.IsTerminal(), now); err != nil {
		return StrategyDispatchLease{}, err
	}
	authority, err := loadStrategyDispatchAttemptAuthority(ctx, tx, loaded, a.id)
	if err != nil {
		return StrategyDispatchLease{}, err
	}
	transition, observedAt, queryDigest, err := deriveStrategyDispatchOutcome(loaded, authority, now)
	if err != nil {
		return StrategyDispatchLease{}, err
	}
	if err := proveStrategyDispatchOutcomeHolds(ctx, tx, loaded, authority, transition); err != nil {
		return StrategyDispatchLease{}, err
	}
	if transition.disposition == StrategyDispatchReservationReleased {
		if err := releaseStrategyDispatchOutcomeHolds(ctx, tx, loaded, authority, observedAt); err != nil {
			return StrategyDispatchLease{}, err
		}
	}
	if transition.disposition == StrategyDispatchReservationTransferred {
		if err := a.j.linkConfirmedStrategyDispatchTx(ctx, tx, loaded, authority, observedAt); err != nil {
			return StrategyDispatchLease{}, err
		}
	}
	if err := updateStrategyDispatchOutcomeLeaseTx(ctx, tx, loaded, authority.brokerOrderID, queryDigest,
		transition, observedAt, now); err != nil {
		return StrategyDispatchLease{}, err
	}
	if err := appendStrategyDispatchBrokerOutcomeTx(ctx, tx, loaded, authority.brokerOrderID, queryDigest,
		transition, observedAt); err != nil {
		return StrategyDispatchLease{}, err
	}
	terminal, err := loadStrategyDispatchLease(ctx, tx, loaded.LeaseID)
	if err != nil {
		return StrategyDispatchLease{}, err
	}
	if err := tx.Commit(); err != nil {
		return StrategyDispatchLease{}, err
	}
	return terminal, nil
}

func transitionStrategyAttemptTx(ctx context.Context, tx *sql.Tx, a *Attempt, to AttemptState,
	brokerOrderID, reasonCode, detail string, settled bool, now time.Time,
) (string, error) {
	var current AttemptState
	if err := tx.QueryRowContext(ctx, `SELECT state FROM mutation_attempts WHERE id=?`, a.id).Scan(&current); err != nil {
		return "", err
	}
	if err := ValidateTransition(current, to); err != nil {
		return "", err
	}
	nowText := formatJournalTime(now)
	if to == StateConfirmed {
		var latest sql.NullString
		if err := tx.QueryRowContext(ctx, `SELECT max(f.committed_at)
			FROM scoped_fill_snapshots f JOIN intents i ON i.id=?
			WHERE f.order_id=(SELECT broker_order_id FROM mutation_attempts WHERE id=?)
			  AND f.account_ref=trim(i.account_ref) AND f.market=lower(trim(i.market))
			  AND f.trading_day=trim(i.trading_day) AND f.symbol=upper(trim(i.symbol))
			  AND f.side=upper(trim(i.side))`, a.intentID, a.id).Scan(&latest); err != nil {
			return "", err
		}
		var orderErr error
		nowText, orderErr = journalTimeStrictlyAfter(nowText, latest.String)
		if orderErr != nil {
			return "", orderErr
		}
	}
	set := []string{"state=?"}
	args := []any{to}
	if brokerOrderID != "" {
		set, args = append(set, "broker_order_id=?"), append(args, brokerOrderID)
	}
	if settled {
		set, args = append(set, "settled_at=?"), append(args, nowText)
	}
	if reasonCode != "" {
		set, args = append(set, "reason_code=?"), append(args, reasonCode)
	}
	if detail != "" {
		set, args = append(set, "detail=?"), append(args, detail)
	}
	args = append(args, a.id, current)
	result, err := tx.ExecContext(ctx, `UPDATE mutation_attempts SET `+strings.Join(set, ",")+` WHERE id=? AND state=?`, args...)
	if err != nil {
		return "", err
	}
	if err := requireStrategyDispatchExactRows(result, 1, "strategy core outcome transition"); err != nil {
		return "", err
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO attempt_transitions(attempt_id,from_state,to_state,at,reason_code,detail)
		VALUES(?,?,?,?,?,?)`, a.id, current, to, nowText, reasonCode, detail); err != nil {
		return "", err
	}
	if releasesReservations(to) {
		if _, err := releaseReservationsForAttempt(ctx, tx, a.id, ReleaseReasonBrokerTerminal,
			fmt.Sprintf("attempt %s reached %s", a.id, to), nowText); err != nil {
			return "", err
		}
	}
	return nowText, nil
}

func loadStrategyDispatchAttemptAuthority(ctx context.Context, tx *sql.Tx, lease StrategyDispatchLease, attemptID string) (strategyDispatchAttemptAuthority, error) {
	var a strategyDispatchAttemptAuthority
	err := tx.QueryRowContext(ctx, `SELECT a.id,a.intent_id,a.kind,a.state,a.target_order_id,a.broker_order_id,
		COALESCE(a.decision_id,''),COALESCE(a.account_ref,''),COALESCE(a.client_order_id,''),
		COALESCE(a.dispatch_started_at,''),COALESCE(a.settled_at,''),
		COALESCE((SELECT max(t.at) FROM attempt_transitions t WHERE t.attempt_id=a.id),''),
		i.account_ref,upper(i.market),i.trading_day,upper(i.symbol),upper(i.side),i.quantity,upper(i.currency),
		s.attempt_id,s.state,s.client_order_id,s.revision,b.campaign_id,b.leg_sequence,b.q_final,b.prospective_token
		FROM mutation_attempts a
		JOIN intents i ON i.id=a.intent_id
		JOIN strategy_first_leg_bindings b ON b.decision_id=a.decision_id
		JOIN strategy_attempt_lineage s ON s.attempt_id=b.attempt_id
		WHERE a.id=?`, attemptID).Scan(
		&a.attemptID, &a.intentID, &a.kind, &a.state, &a.targetOrderID, &a.brokerOrderID,
		&a.decisionID, &a.attemptAccount, &a.clientOrderID, &a.dispatchStartedAt, &a.settledAt, &a.transitionAt,
		&a.intentAccount, &a.market, &a.tradingDay, &a.symbol, &a.side, &a.quantity, &a.currency,
		&a.strategyAttemptID, &a.strategyState, &a.strategyClientOrderID, &a.strategyRevision,
		&a.campaignID, &a.legSequence, &a.qFinal, &a.prospectiveToken)
	if errors.Is(err, sql.ErrNoRows) {
		return strategyDispatchAttemptAuthority{}, fmt.Errorf("%w: strategy dispatch attempt binding", ErrStrategyDispatchLeaseUnavailable)
	}
	if err != nil {
		return strategyDispatchAttemptAuthority{}, fmt.Errorf("journal: loading strategy dispatch attempt authority: %w", err)
	}
	quantity, quantityErr := strconv.ParseInt(a.quantity, 10, 64)
	if a.attemptID != attemptID || a.kind != string(KindPlace) || a.targetOrderID != "" ||
		a.decisionID != lease.GuardianDecisionID || a.attemptAccount != lease.AccountRef ||
		a.intentAccount != lease.AccountRef || a.market != string(lease.Market) || a.symbol != lease.Symbol ||
		a.side != "BUY" || a.clientOrderID != lease.OperationID || a.strategyClientOrderID != lease.OperationID ||
		a.campaignID != lease.CampaignID || a.legSequence != 1 || quantityErr != nil || quantity != a.qFinal {
		return strategyDispatchAttemptAuthority{}, fmt.Errorf("%w: core attempt is not the exact strategy lease authority", ErrStrategyDispatchLeaseUnavailable)
	}
	return a, nil
}

func deriveStrategyDispatchOutcome(lease StrategyDispatchLease, a strategyDispatchAttemptAuthority, now time.Time) (strategyDispatchOutcomeTransition, time.Time, string, error) {
	observedText := a.settledAt
	if observedText == "" {
		observedText = a.transitionAt
	}
	observedAt, err := parseJournalTime(observedText)
	if err != nil || observedAt.Before(lease.TransportStartedAt) || observedAt.After(now.Add(time.Second)) {
		return strategyDispatchOutcomeTransition{}, time.Time{}, "", fmt.Errorf("%w: durable attempt outcome time is invalid", ErrInvalidRequest)
	}
	transition := strategyDispatchOutcomeTransition{state: StrategyDispatchLeaseAmbiguous,
		disposition: StrategyDispatchReservationHeld, code: strategyDispatchOutcomeCodeAmbiguous}
	switch AttemptState(a.state) {
	case StateNotDispatched, StateFailedConfirmed:
		if a.brokerOrderID == "" {
			transition = strategyDispatchOutcomeTransition{state: StrategyDispatchLeaseRefused,
				disposition: StrategyDispatchReservationReleased, code: strategyDispatchOutcomeCodeNotSent,
				refusalCode: string(strategyDispatchOutcomeCodeNotSent)}
		}
	case StateConfirmed:
		if strings.TrimSpace(a.brokerOrderID) != "" {
			transition = strategyDispatchOutcomeTransition{state: StrategyDispatchLeaseSubmitted,
				disposition: StrategyDispatchReservationTransferred, code: strategyDispatchOutcomeCodeConfirmed}
		}
	}
	payload := strings.Join([]string{lease.LeaseID, a.attemptID, a.intentID, a.decisionID, a.state,
		a.brokerOrderID, a.intentAccount, a.market, a.tradingDay, a.symbol, a.side, a.quantity,
		formatJournalTime(observedAt)}, "\x00")
	digest := sha256.Sum256([]byte(payload))
	return transition, observedAt, hex.EncodeToString(digest[:]), nil
}

func proveStrategyDispatchOutcomeHolds(ctx context.Context, tx *sql.Tx, lease StrategyDispatchLease,
	a strategyDispatchAttemptAuthority, transition strategyDispatchOutcomeTransition,
) error {
	var aggregateState, aggregateAttempt, aggregateReason string
	if err := tx.QueryRowContext(ctx, `SELECT state,COALESCE(attempt_id,''),COALESCE(release_reason,'') FROM risk_reservations
		WHERE id=? AND decision_id=? AND account_ref=?`, lease.RiskReservationID, lease.GuardianDecisionID, lease.AccountRef).
		Scan(&aggregateState, &aggregateAttempt, &aggregateReason); err != nil {
		return fmt.Errorf("journal: proving strategy dispatch aggregate outcome hold: %w", err)
	}
	if aggregateAttempt != a.attemptID {
		return fmt.Errorf("%w: strategy dispatch aggregate attempt binding", ErrStrategyDispatchLeaseUnavailable)
	}
	if transition.disposition == StrategyDispatchReservationReleased {
		if aggregateState != "HELD" && !(aggregateState == "RELEASED" && aggregateReason == ReleaseReasonBrokerTerminal) {
			return fmt.Errorf("%w: strategy dispatch aggregate release preimage", ErrStrategyDispatchLeaseUnavailable)
		}
	} else if aggregateState != "HELD" {
		return fmt.Errorf("%w: strategy dispatch aggregate outcome hold", ErrStrategyDispatchLeaseUnavailable)
	}
	var total, exact, dimensions int
	if err := tx.QueryRowContext(ctx, `SELECT count(*),
		COALESCE(sum(CASE WHEN state='HELD' AND held_minor=reserved_minor
		  AND bucket_dimension IN ('horizon','market','strategy','sector','symbol') THEN 1 ELSE 0 END),0),
		count(DISTINCT CASE WHEN state='HELD' AND held_minor=reserved_minor
		  AND bucket_dimension IN ('horizon','market','strategy','sector','symbol') THEN bucket_dimension END)
		FROM risk_bucket_reservations
		WHERE decision_id=? AND existing_reservation_id=? AND account_ref=? AND market=? AND symbol=?`,
		lease.GuardianDecisionID, lease.RiskReservationID, lease.AccountRef, lease.Market, lease.Symbol).
		Scan(&total, &exact, &dimensions); err != nil {
		return fmt.Errorf("journal: proving strategy dispatch monetary outcome holds: %w", err)
	}
	if total != 5 || exact != 5 || dimensions != 5 {
		return fmt.Errorf("%w: strategy dispatch monetary outcome holds total=%d exact=%d dimensions=%d want=5/5/5",
			ErrStrategyDispatchLeaseUnavailable, total, exact, dimensions)
	}
	return nil
}

func releaseStrategyDispatchOutcomeHolds(ctx context.Context, tx *sql.Tx, lease StrategyDispatchLease,
	a strategyDispatchAttemptAuthority, observedAt time.Time,
) error {
	observedText := formatJournalTime(observedAt)
	aggregate, err := tx.ExecContext(ctx, `UPDATE risk_reservations
		SET state='RELEASED',released_at=?,release_reason=?
		WHERE id=? AND decision_id=? AND account_ref=? AND attempt_id=? AND state='HELD'`,
		observedText, ReleaseReasonBrokerTerminal, lease.RiskReservationID, lease.GuardianDecisionID, lease.AccountRef, a.attemptID)
	if err != nil {
		return fmt.Errorf("journal: releasing strategy dispatch aggregate outcome hold: %w", err)
	}
	if rows, rowsErr := aggregate.RowsAffected(); rowsErr != nil {
		return rowsErr
	} else if rows > 1 {
		return fmt.Errorf("%w: aggregate outcome hold release affected=%d", ErrStrategyDispatchLeaseUnavailable, rows)
	}
	monetary, err := tx.ExecContext(ctx, `UPDATE risk_bucket_reservations
		SET state='RELEASED',held_minor='0',updated_at=?
		WHERE decision_id=? AND existing_reservation_id=? AND account_ref=? AND market=? AND symbol=?
		  AND state='HELD' AND held_minor=reserved_minor
		  AND bucket_dimension IN ('horizon','market','strategy','sector','symbol')`,
		observedText, lease.GuardianDecisionID, lease.RiskReservationID, lease.AccountRef, lease.Market, lease.Symbol)
	if err != nil {
		return fmt.Errorf("journal: releasing strategy dispatch monetary outcome holds: %w", err)
	}
	if err := requireStrategyDispatchExactRows(monetary, 5, "monetary outcome hold release"); err != nil {
		return err
	}
	return nil
}

func (j *Journal) linkConfirmedStrategyDispatchTx(ctx context.Context, tx *sql.Tx, lease StrategyDispatchLease,
	a strategyDispatchAttemptAuthority, observedAt time.Time,
) error {
	if err := j.registerConfirmedStrategyRiskOrderTx(ctx, tx, lease, a, observedAt); err != nil {
		return fmt.Errorf("journal: linking confirmed strategy risk order: %w", err)
	}
	if err := linkConfirmedStrategyCampaignTx(ctx, tx, lease, a, observedAt); err != nil {
		return fmt.Errorf("journal: linking confirmed strategy campaign order: %w", err)
	}
	if err := linkConfirmedStrategyExecutionTx(ctx, tx, lease, a, observedAt); err != nil {
		return fmt.Errorf("journal: linking confirmed strategy execution: %w", err)
	}
	if err := j.backfillConfirmedStrategyFillTx(ctx, tx, lease, a); err != nil {
		return fmt.Errorf("journal: backfilling confirmed strategy fill: %w", err)
	}
	return nil
}

func (j *Journal) registerConfirmedStrategyRiskOrderTx(ctx context.Context, tx *sql.Tx, lease StrategyDispatchLease,
	a strategyDispatchAttemptAuthority, observedAt time.Time,
) error {
	authority, err := loadRiskBucketOrderAuthority(ctx, tx, lease.GuardianDecisionID)
	if err != nil {
		return err
	}
	owner := riskbucket.OwnerKey{AccountID: lease.AccountRef, Market: riskbucket.Market(lease.Market),
		Symbol: lease.Symbol, ProspectiveGeneration: a.prospectiveToken}
	var active int
	if err := tx.QueryRowContext(ctx, `SELECT count(*) FROM risk_bucket_owners
		WHERE account_ref=? AND market=? AND symbol=? AND prospective_generation=? AND released_at IS NULL`,
		owner.AccountID, owner.Market, owner.Symbol, owner.ProspectiveGeneration).Scan(&active); err != nil || active != 1 {
		return fmt.Errorf("%w: active strategy risk owner", ErrRiskBucketStateUnknown)
	}
	if err := verifyRiskBucketStateDigest(ctx, tx, owner); err != nil {
		return err
	}
	reserved := make(map[riskbucket.BucketKey]string, len(authority.bindings))
	reservationIDs := make(map[riskbucket.BucketKey]string, len(authority.bindings))
	for _, binding := range authority.bindings {
		key := riskbucket.BucketKey{Dimension: riskbucket.Dimension(binding.Dimension), Value: binding.Value,
			PolicyVersion: binding.PolicyVersion}
		reserved[key] = binding.ReservedMinor
		reservationIDs[key] = binding.ReservationID
	}
	plan := RiskBucketOrderPlan{OrderID: a.brokerOrderID, DecisionID: lease.GuardianDecisionID,
		OrderQuantity: uint64(a.qFinal), ReservedMinor: reserved, ReservationPolicyDigest: authority.digest,
		QuoteCurrency: authority.quoteCurrency, BaseCurrency: authority.baseCurrency, CreatedAt: observedAt}
	digest, err := riskBucketOrderPlanDigest(plan, authority.bindings)
	if err != nil {
		return err
	}
	orderKey := riskBucketOrderKey(plan.DecisionID, plan.OrderID)
	var priorDigest string
	if err := tx.QueryRowContext(ctx, `SELECT request_digest FROM risk_bucket_orders WHERE order_key=?`, orderKey).Scan(&priorDigest); err == nil {
		if priorDigest != digest {
			return fmt.Errorf("%w: strategy risk order replay", ErrRiskBucketReplayMismatch)
		}
		return nil
	} else if !errors.Is(err, sql.ErrNoRows) {
		return err
	}
	var collision int
	if err := tx.QueryRowContext(ctx, `SELECT count(*) FROM risk_bucket_orders o
		JOIN risk_bucket_final_decisions d ON d.decision_id=o.decision_id
		WHERE o.order_id=? AND o.order_key<>? AND d.account_ref=? AND d.market=? AND d.symbol=?
		  AND d.owner_prospective_generation=?`, plan.OrderID, orderKey, owner.AccountID, owner.Market,
		owner.Symbol, owner.ProspectiveGeneration).Scan(&collision); err != nil {
		return err
	}
	if collision != 0 {
		return fmt.Errorf("%w: opaque broker order already belongs to risk owner", ErrRiskBucketReplayMismatch)
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO risk_bucket_orders(
		order_key,order_id,decision_id,predecessor_order_key,order_quantity,cumulative_fill,
		quote_currency,base_currency,reservation_policy_digest,request_digest,state,created_at,updated_at)
		VALUES(?,?,?,?,?,0,?,?,?,?,'ACTIVE',?,?)`, orderKey, plan.OrderID, plan.DecisionID, nil,
		plan.OrderQuantity, plan.QuoteCurrency, plan.BaseCurrency, plan.ReservationPolicyDigest, digest,
		canonicalRiskTime(observedAt), canonicalRiskTime(observedAt)); err != nil {
		return err
	}
	for _, binding := range authority.bindings {
		key := riskbucket.BucketKey{Dimension: riskbucket.Dimension(binding.Dimension), Value: binding.Value,
			PolicyVersion: binding.PolicyVersion}
		if _, err := tx.ExecContext(ctx, `INSERT INTO risk_bucket_order_reservations(order_key,reservation_id,reserved_minor)
			VALUES(?,?,?)`, orderKey, reservationIDs[key], reserved[key]); err != nil {
			return err
		}
	}
	return j.recordRiskBucketStateTx(ctx, tx, owner, "ORDER_REGISTERED", orderKey, digest, canonicalRiskTime(observedAt))
}

func linkConfirmedStrategyCampaignTx(ctx context.Context, tx *sql.Tx, lease StrategyDispatchLease,
	a strategyDispatchAttemptAuthority, observedAt time.Time,
) error {
	state, version, blocked, err := campaignHeaderInTx(ctx, tx, lease.CampaignID)
	if err != nil {
		return err
	}
	if blocked {
		return positioncampaign.ErrExposureBlocked
	}
	if exposureBlocked, err := campaignExposureBlockedInTx(ctx, tx, lease.CampaignID); err != nil {
		return err
	} else if exposureBlocked {
		return positioncampaign.ErrExposureBlocked
	}
	leg, err := campaignLegInTx(ctx, tx, lease.CampaignID, a.legSequence)
	if err != nil {
		return err
	}
	if leg.IntentID != "" && leg.IntentID != a.intentID {
		return fmt.Errorf("%w: first leg intent already bound", positioncampaign.ErrInvalidIdentity)
	}
	req := LinkCampaignOrderRequest{CampaignID: lease.CampaignID, LegSequence: a.legSequence,
		ExpectedVersion: version, CommandKey: "strategy-dispatch:" + lease.LeaseID + ":" + a.attemptID,
		OrderID: a.brokerOrderID, IntentID: a.intentID, AttemptID: a.attemptID,
		RequestedCap: strconv.FormatInt(a.qFinal, 10)}
	digest := digestParts(req.CampaignID, strconv.FormatInt(req.LegSequence, 10), req.OrderID, "",
		req.RequestedCap, req.IntentID, req.AttemptID, "false")
	if _, found, err := campaignCommandResult(ctx, tx, req.CampaignID, "LINK_ORDER", req.CommandKey, digest); err != nil {
		return err
	} else if found {
		return nil
	}
	authority, err := authoritativeCampaignOrderInTx(ctx, tx, req)
	if err != nil {
		return err
	}
	var duplicate int
	if err := tx.QueryRowContext(ctx, `SELECT count(*) FROM campaign_order_watermarks
		WHERE account_ref=? AND market=? AND trading_day=? AND symbol=? AND side=? AND order_id=?`,
		authority.accountRef, authority.market, authority.tradingDay, authority.symbol, authority.side,
		req.OrderID).Scan(&duplicate); err != nil {
		return err
	}
	if duplicate != 0 {
		return fmt.Errorf("%w: opaque broker order already belongs to campaign scope", positioncampaign.ErrInvalidIdentity)
	}
	nextLeg, err := positioncampaign.TransitionLeg(leg.State, positioncampaign.LegOrderLinked)
	if err != nil {
		return err
	}
	nextCampaign, err := positioncampaign.TransitionCampaign(state, positioncampaign.CampaignOrderLinked)
	if err != nil {
		return err
	}
	now := formatJournalTime(observedAt)
	newVersion := version + 1
	if _, err := tx.ExecContext(ctx, `INSERT INTO campaign_order_watermarks(
		campaign_id,leg_sequence,order_id,account_ref,market,trading_day,symbol,side,decision_id,
		intent_id,attempt_id,predecessor_order_id,carry_baseline,requested_cap,cumulative_filled,
		remaining_quantity,terminal,lineage_ambiguous,created_at,updated_at)
		VALUES(?,?,?,?,?,?,?,?,?,?,?,NULL,'0',?,'0',?,0,0,?,?)`, req.CampaignID, req.LegSequence,
		req.OrderID, authority.accountRef, authority.market, authority.tradingDay, authority.symbol,
		authority.side, authority.decisionID, req.IntentID, req.AttemptID, req.RequestedCap,
		req.RequestedCap, now, now); err != nil {
		return err
	}
	legResult, err := tx.ExecContext(ctx, `UPDATE campaign_legs SET state=?,intent_id=coalesce(intent_id,?),
		version=version+1,updated_at=? WHERE campaign_id=? AND sequence=?`, string(nextLeg), req.IntentID,
		now, req.CampaignID, req.LegSequence)
	if err != nil {
		return err
	}
	if err := requireStrategyDispatchExactRows(legResult, 1, "strategy campaign leg link"); err != nil {
		return err
	}
	campaignResult, err := tx.ExecContext(ctx, `UPDATE position_campaigns SET state=?,version=?,entry_blocked=?,updated_at=?
		WHERE id=? AND version=?`, string(nextCampaign.State), newVersion, boolInt(nextCampaign.EntryBlocked), now,
		req.CampaignID, version)
	if err != nil {
		return err
	}
	if err := requireStrategyDispatchExactRows(campaignResult, 1, "strategy campaign link"); err != nil {
		return err
	}
	if err := insertCampaignCommand(ctx, tx, req.CampaignID, "LINK_ORDER", req.CommandKey, digest,
		newVersion, req.LegSequence, now); err != nil {
		return err
	}
	return insertCampaignEvent(ctx, tx, req.CampaignID, newVersion, newVersion, req.LegSequence,
		req.OrderID, "ORDER_LINKED", "LINK_ORDER", req.CommandKey, digest, nextCampaign.State,
		nextLeg, 0, "", "", now)
}

func linkConfirmedStrategyExecutionTx(ctx context.Context, tx *sql.Tx, lease StrategyDispatchLease,
	a strategyDispatchAttemptAuthority, observedAt time.Time,
) error {
	if a.strategyState != "DISPATCHED" {
		if a.strategyState != "PLANNED" && a.strategyState != "IN_DOUBT" {
			return fmt.Errorf("journal strategy dispatch: stale strategy attempt state %s", a.strategyState)
		}
		result, err := tx.ExecContext(ctx, `UPDATE strategy_attempt_lineage SET state='DISPATCHED',revision=revision+1
			WHERE attempt_id=? AND account_ref=? AND revision=? AND state=?`, a.strategyAttemptID,
			lease.AccountRef, a.strategyRevision, a.strategyState)
		if err != nil {
			return err
		}
		if err := requireStrategyDispatchExactRows(result, 1, "strategy execution state"); err != nil {
			return err
		}
	}
	for _, link := range [][2]string{{"MUTATION_ATTEMPT", a.attemptID}, {"BROKER_ORDER", a.brokerOrderID}} {
		if _, err := insertExactStrategyExecution(ctx, tx, lease.AccountRef, a.strategyAttemptID,
			link[0], link[1], observedAt); err != nil {
			return err
		}
	}
	return nil
}

func (j *Journal) backfillConfirmedStrategyFillTx(ctx context.Context, tx *sql.Tx, lease StrategyDispatchLease,
	a strategyDispatchAttemptAuthority,
) error {
	var state, quantity, cumulative, average, amount, visible, committed string
	var terminal, failClosed int
	err := tx.QueryRowContext(ctx, `SELECT state,terminal,fail_closed,quantity,filled_quantity,
		average_price,filled_amount,broker_visible_at,committed_at FROM scoped_fill_snapshots
		WHERE order_id=? AND account_ref=? AND market=? AND trading_day=? AND symbol=? AND side='BUY'`,
		a.brokerOrderID, lease.AccountRef, strings.ToLower(string(lease.Market)), a.tradingDay, lease.Symbol).
		Scan(&state, &terminal, &failClosed, &quantity, &cumulative, &average, &amount, &visible, &committed)
	if errors.Is(err, sql.ErrNoRows) {
		return nil
	}
	if err != nil {
		return err
	}
	if failClosed != 0 {
		return fmt.Errorf("%w: fail-closed fill cannot be backfilled", ErrStrategyDispatchLeaseUnavailable)
	}
	fill := AppliedFill{OrderID: a.brokerOrderID, AccountRef: lease.AccountRef, Symbol: lease.Symbol,
		Market: strings.ToLower(string(lease.Market)), TradingDay: a.tradingDay, Side: "BUY", State: state,
		Terminal: terminal != 0, Delta: cumulative, CumulativeQuantity: cumulative, AveragePrice: average,
		FilledAmount: amount, OrderedQuantity: quantity, BrokerVisibleAt: visible, CommittedAt: committed}
	if fill.Terminal {
		if _, err := releaseReservationsForOrder(ctx, tx, a.brokerOrderID, a.intentID,
			ReleaseReasonBrokerTerminal,
			fmt.Sprintf("order %s derived terminal as %s", a.brokerOrderID, state), committed); err != nil {
			return err
		}
	}
	if err := j.applyRiskBucketFillInTx(ctx, tx, fill); err != nil {
		return err
	}
	if err := j.releaseTerminalRiskBucketOrderInTx(ctx, tx, fill); err != nil {
		return err
	}
	// A snapshot committed after core CONFIRMED already ran Project/Exit. Only
	// mappings were absent. A snapshot from the ACKED window ran no hooks.
	committedAt, committedErr := parseJournalTime(committed)
	settledAt, settledErr := parseJournalTime(a.settledAt)
	allHooksMissing := a.settledAt == "" || (committedErr == nil && settledErr == nil && !committedAt.After(settledAt))
	if allHooksMissing {
		// Projection ownership deliberately requires core settled_at to precede
		// the apply instant. The broker snapshot remains immutable at its original
		// committed_at; this synthetic apply instant orders the repair after the
		// confirmation committed in this same composite transaction.
		applyAt, orderErr := journalTimeStrictlyAfter(a.settledAt, a.settledAt, committed)
		if orderErr != nil {
			return orderErr
		}
		fill.CommittedAt = applyAt
		if err := j.runApplyHooks(ctx, tx, fill); err != nil {
			return err
		}
		// Rebase the mutable latest snapshot after ownership so the next broker
		// cumulative observation measures from the quantity just backfilled. The
		// append-only fill event retains the original evidence commit time.
		rebased, err := tx.ExecContext(ctx, `UPDATE scoped_fill_snapshots SET committed_at=?
			WHERE order_id=? AND account_ref=? AND market=? AND trading_day=? AND symbol=? AND side='BUY'
			  AND committed_at=?`, applyAt, a.brokerOrderID, lease.AccountRef,
			strings.ToLower(string(lease.Market)), a.tradingDay, lease.Symbol, committed)
		if err != nil {
			return err
		}
		if err := requireStrategyDispatchExactRows(rebased, 1, "strategy fill ownership rebase"); err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `UPDATE fill_snapshots SET committed_at=? WHERE order_id=?
			AND trim(account_ref)=? AND lower(trim(market))=? AND trim(trading_day)=?
			AND upper(trim(symbol))=? AND upper(trim(side))='BUY' AND committed_at=?`, applyAt,
			a.brokerOrderID, lease.AccountRef, strings.ToLower(string(lease.Market)), a.tradingDay,
			lease.Symbol, committed); err != nil {
			return err
		}
		j.applyMu.RLock()
		campaignBound := j.applyHooks.Campaign != nil
		j.applyMu.RUnlock()
		if campaignBound {
			return nil
		}
	}
	handle := &ApplyTx{tx: tx, now: committed}
	if err := ApplyPositionCampaignFill(ctx, handle, fill); err != nil {
		handle.invalidate()
		return err
	}
	handle.invalidate()
	return j.applyRiskBucketOwnerBindingInTx(ctx, tx, fill)
}

func updateStrategyDispatchOutcomeLeaseTx(ctx context.Context, tx *sql.Tx, lease StrategyDispatchLease,
	brokerOrderID, queryDigest string, transition strategyDispatchOutcomeTransition, observedAt, now time.Time,
) error {
	nextRevision := lease.Revision + 1
	result, err := tx.ExecContext(ctx, `UPDATE strategy_dispatch_leases
		SET state=?,disposition=?,revision=?,refusal_code=?,outcome_code=?,broker_order_id=?,query_digest=?,outcome_observed_at=?,updated_at=?
		WHERE lease_id=? AND state='SUBMITTING' AND disposition='RESERVED' AND revision=?
		  AND owner_epoch=? AND fencing_token=? AND transport_started_at IS NOT NULL
		  AND refusal_code='' AND outcome_code='' AND query_digest='' AND outcome_observed_at IS NULL`,
		transition.state, transition.disposition, nextRevision, transition.refusalCode, transition.code,
		brokerOrderID, queryDigest, formatJournalTime(observedAt), formatJournalTime(now),
		lease.LeaseID, lease.Revision, lease.OwnerEpoch, lease.FencingToken)
	if err != nil {
		return fmt.Errorf("journal: terminalizing strategy dispatch outcome: %w", err)
	}
	if err := requireStrategyDispatchOneRow(result); err != nil {
		return err
	}
	return nil
}

func appendStrategyDispatchBrokerOutcomeTx(ctx context.Context, tx *sql.Tx, lease StrategyDispatchLease,
	brokerOrderID, queryDigest string, transition strategyDispatchOutcomeTransition, observedAt time.Time,
) error {
	nextRevision := lease.Revision + 1
	payload := strings.Join([]string{
		lease.LeaseID, string(lease.State), string(transition.state), string(lease.Disposition), string(transition.disposition),
		fmt.Sprintf("%d", lease.Revision), fmt.Sprintf("%d", nextRevision), string(transition.code), lease.OperationID,
		brokerOrderID, queryDigest, formatJournalTime(observedAt),
	}, "\x00")
	digest := sha256.Sum256([]byte(payload))
	_, err := tx.ExecContext(ctx, `INSERT INTO strategy_dispatch_outcomes(
		outcome_id,lease_id,from_state,to_state,from_disposition,to_disposition,expected_revision,next_revision,
		transition_code,operation_identity,broker_order_id,query_digest,observed_at,record_digest)
		VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		fmt.Sprintf("dispatch-transition:%s:%d", lease.LeaseID, nextRevision), lease.LeaseID,
		lease.State, transition.state, lease.Disposition, transition.disposition, lease.Revision, nextRevision,
		transition.code, lease.OperationID, brokerOrderID, queryDigest, formatJournalTime(observedAt), hex.EncodeToString(digest[:]))
	if err != nil {
		return fmt.Errorf("journal: recording strategy dispatch broker outcome: %w", err)
	}
	return nil
}

func requireStrategyDispatchExactRows(result sql.Result, expected int64, authority string) error {
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("journal: %s affected rows: %w", authority, err)
	}
	if rows != expected {
		return fmt.Errorf("%w: %s affected=%d want=%d", ErrStrategyDispatchLeaseUnavailable, authority, rows, expected)
	}
	return nil
}

const strategyDispatchRecoveryClaimedNoTransport = "RECOVERY_CLAIMED_NO_TRANSPORT"

// RecoverClaimedStrategyDispatchLease is the no-send cold-restart terminalizer.
// A newly fenced central owner may close only an old CLAIMED lease whose durable
// transport-start marker is absent. It cannot adopt the lease, move it to
// SUBMITTING, call a Gateway, or create retry/resubmit authority.
func (j *Journal) RecoverClaimedStrategyDispatchLease(ctx context.Context, request StrategyDispatchLeaseCAS) (StrategyDispatchLease, error) {
	if j == nil || j.db == nil || !validStrategyDispatchIdentity(request.LeaseID) ||
		request.ExpectedRevision == 0 || request.OwnerEpoch == 0 || !validStrategyDispatchIdentity(request.FencingToken) {
		return StrategyDispatchLease{}, fmt.Errorf("%w: invalid claimed strategy dispatch recovery", ErrInvalidRequest)
	}
	tx, err := j.db.BeginTx(ctx, nil)
	if err != nil {
		return StrategyDispatchLease{}, fmt.Errorf("journal: begin claimed strategy dispatch recovery: %w", err)
	}
	defer tx.Rollback()
	if err := requireCurrentStrategyDispatchOwner(ctx, tx, request.OwnerEpoch, request.FencingToken); err != nil {
		return StrategyDispatchLease{}, err
	}
	lease, err := loadStrategyDispatchLease(ctx, tx, request.LeaseID)
	if errors.Is(err, sql.ErrNoRows) {
		return StrategyDispatchLease{}, ErrStrategyDispatchLeaseUnavailable
	}
	if err != nil {
		return StrategyDispatchLease{}, err
	}
	if lease.State != StrategyDispatchLeaseClaimed || lease.Disposition != StrategyDispatchReservationReserved ||
		lease.Revision != request.ExpectedRevision || !lease.TransportStartedAt.IsZero() {
		return StrategyDispatchLease{}, ErrStrategyDispatchLeaseConsumed
	}
	if request.OwnerEpoch <= lease.OwnerEpoch {
		return StrategyDispatchLease{}, fmt.Errorf("%w: claimed recovery requires a newer owner epoch", ErrStrategyDispatchFenced)
	}
	if err := proveClaimedStrategyDispatchPreTransportAuthority(ctx, tx, lease); err != nil {
		return StrategyDispatchLease{}, err
	}
	now := j.clk.Now().UTC()
	prepared, err := recoverablePreparedStrategyAttempt(ctx, tx, lease)
	if err != nil {
		return StrategyDispatchLease{}, err
	}
	if prepared != nil {
		if _, err := transitionStrategyAttemptTx(ctx, tx, prepared, StateNotDispatched, "",
			strategyDispatchRecoveryClaimedNoTransport,
			"cold restart proved the claimed strategy lease never crossed durable transport start", true, now); err != nil {
			return StrategyDispatchLease{}, err
		}
	}
	terminal, err := refuseClaimedStrategyDispatchSubmittingTx(ctx, tx, lease,
		strategyDispatchRecoveryClaimedNoTransport, now)
	if err != nil {
		return StrategyDispatchLease{}, err
	}
	if err := tx.Commit(); err != nil {
		return StrategyDispatchLease{}, fmt.Errorf("journal: commit claimed strategy dispatch recovery: %w", err)
	}
	return terminal, nil
}

// recoverablePreparedStrategyAttempt returns the optional durable core attempt
// that may have committed after lease claim and before the crash. More than one
// exact match is an integrity failure; a non-RECORDED match cannot be rewritten
// by this no-send recovery path.
func recoverablePreparedStrategyAttempt(ctx context.Context, tx *sql.Tx, lease StrategyDispatchLease) (*Attempt, error) {
	var count int
	if err := tx.QueryRowContext(ctx, `SELECT count(*) FROM mutation_attempts
		WHERE decision_id=? AND account_ref=? AND client_order_id=?`,
		lease.GuardianDecisionID, lease.AccountRef, lease.OperationID).Scan(&count); err != nil {
		return nil, fmt.Errorf("journal: counting prepared strategy recovery attempts: %w", err)
	}
	if count == 0 {
		return nil, nil
	}
	if count != 1 {
		return nil, fmt.Errorf("%w: claimed recovery core attempt count=%d want at most one",
			ErrStrategyDispatchLeaseUnavailable, count)
	}
	var attempt Attempt
	var state AttemptState
	if err := tx.QueryRowContext(ctx, `SELECT id,intent_id,kind,attempt_no,state FROM mutation_attempts
		WHERE decision_id=? AND account_ref=? AND client_order_id=?`,
		lease.GuardianDecisionID, lease.AccountRef, lease.OperationID).
		Scan(&attempt.id, &attempt.intentID, &attempt.kind, &attempt.attemptNo, &state); err != nil {
		return nil, fmt.Errorf("journal: loading prepared strategy recovery attempt: %w", err)
	}
	attempt.j, attempt.state = nil, state
	authority, err := loadStrategyDispatchAttemptAuthority(ctx, tx, lease, attempt.id)
	if err != nil {
		return nil, err
	}
	if state != StateRecorded || authority.state != string(StateRecorded) || authority.brokerOrderID != "" ||
		authority.dispatchStartedAt != "" || authority.settledAt != "" {
		return nil, ErrStrategyDispatchLeaseConsumed
	}
	return &attempt, nil
}

// DiscoverStrategyDispatchRecovery lets a newly fenced owner enumerate every
// old ISSUED/CLAIMED/SUBMITTING lease after a cold restart. It is read-only:
// pre-transport rows require an atomic Gateway refusal plus real release;
// SUBMITTING rows require an opaque official exact-outcome attestation.
func (j *Journal) DiscoverStrategyDispatchRecovery(ctx context.Context, owner StrategyDispatchOwner) ([]StrategyDispatchRecoveryItem, error) {
	if j == nil || j.db == nil || owner.Epoch == 0 || !validStrategyDispatchIdentity(owner.FencingToken) {
		return nil, fmt.Errorf("%w: invalid strategy dispatch recovery owner", ErrInvalidRequest)
	}
	if err := requireCurrentStrategyDispatchOwner(ctx, j.db, owner.Epoch, owner.FencingToken); err != nil {
		return nil, err
	}
	rows, err := j.db.QueryContext(ctx, `SELECT lease_id FROM strategy_dispatch_leases
		WHERE state IN ('ISSUED','CLAIMED','SUBMITTING') ORDER BY issued_at,lease_id`)
	if err != nil {
		return nil, fmt.Errorf("journal: discovering strategy dispatch recovery: %w", err)
	}
	defer rows.Close()
	var leaseIDs []string
	for rows.Next() {
		var leaseID string
		if err := rows.Scan(&leaseID); err != nil {
			return nil, err
		}
		leaseIDs = append(leaseIDs, leaseID)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	items := make([]StrategyDispatchRecoveryItem, 0, len(leaseIDs))
	for _, leaseID := range leaseIDs {
		lease, err := loadStrategyDispatchLease(ctx, j.db, leaseID)
		if err != nil {
			return nil, err
		}
		action := StrategyDispatchRecoveryRefuseRelease
		if lease.State == StrategyDispatchLeaseSubmitting {
			action = StrategyDispatchRecoveryAttestedOutcome
		}
		items = append(items, StrategyDispatchRecoveryItem{Lease: lease, Action: action})
	}
	return items, nil
}

// LookupStrategyDispatchLease is a read-only snapshot used by the Gateway to
// bind the claimed lease to the exact persisted decision and order before the
// final SUBMITTING CAS. It grants no owner, claim, transport or outcome power.
func (j *Journal) LookupStrategyDispatchLease(ctx context.Context, leaseID string) (StrategyDispatchLease, error) {
	if j == nil || j.db == nil || !validStrategyDispatchIdentity(leaseID) {
		return StrategyDispatchLease{}, fmt.Errorf("%w: invalid strategy dispatch lease lookup", ErrInvalidRequest)
	}
	lease, err := loadStrategyDispatchLease(ctx, j.db, leaseID)
	if errors.Is(err, sql.ErrNoRows) {
		return StrategyDispatchLease{}, ErrStrategyDispatchLeaseUnavailable
	}
	return lease, err
}

func loadStrategyDispatchLease(ctx context.Context, q rowQueryer, leaseID string) (StrategyDispatchLease, error) {
	var lease StrategyDispatchLease
	var issued, expires, created, updated string
	var transport, outcomeAt sql.NullString
	err := q.QueryRowContext(ctx, `SELECT lease_id,operation_id,account_ref,market,symbol,candidate_id,evidence_digest,router_id,router_version,
		lane_id,lane_version,campaign_id,leg_id,risk_reservation_id,guardian_decision_id,owner_epoch,fencing_token,authority_revision,
		authority_digest,issued_at,expires_at,state,disposition,revision,transport_started_at,refusal_code,outcome_code,broker_order_id,
		query_digest,outcome_observed_at,lease_digest,created_at,updated_at FROM strategy_dispatch_leases WHERE lease_id=?`, leaseID).Scan(
		&lease.LeaseID, &lease.OperationID, &lease.AccountRef, &lease.Market, &lease.Symbol, &lease.CandidateID, &lease.EvidenceDigest,
		&lease.RouterID, &lease.RouterVersion, &lease.LaneID, &lease.LaneVersion, &lease.CampaignID, &lease.LegID,
		&lease.RiskReservationID, &lease.GuardianDecisionID, &lease.OwnerEpoch, &lease.FencingToken, &lease.AuthorityRevision,
		&lease.AuthorityDigest, &issued, &expires, &lease.State, &lease.Disposition, &lease.Revision, &transport,
		&lease.RefusalCode, &lease.OutcomeCode, &lease.BrokerOrderID, &lease.QueryDigest, &outcomeAt, &lease.LeaseDigest, &created, &updated)
	if err != nil {
		return StrategyDispatchLease{}, fmt.Errorf("journal: loading strategy dispatch lease: %w", err)
	}
	// Parse in column order. A map keyed by timestamp loses one destination
	// whenever issued_at/created_at/updated_at are intentionally identical.
	lease.IssuedAt, err = parseJournalTime(issued)
	if err == nil {
		lease.ExpiresAt, err = parseJournalTime(expires)
	}
	if err == nil {
		lease.CreatedAt, err = parseJournalTime(created)
	}
	if err == nil {
		lease.UpdatedAt, err = parseJournalTime(updated)
	}
	if err != nil {
		return StrategyDispatchLease{}, err
	}
	if transport.Valid {
		lease.TransportStartedAt, err = parseJournalTime(transport.String)
	}
	if err == nil && outcomeAt.Valid {
		lease.OutcomeObservedAt, err = parseJournalTime(outcomeAt.String)
	}
	return lease, err
}

type rowQueryer interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
}

func requireCurrentStrategyDispatchOwner(ctx context.Context, q rowQueryer, epoch uint64, token string) error {
	var currentEpoch uint64
	var currentToken string
	if err := q.QueryRowContext(ctx, `SELECT owner_epoch,fencing_token FROM strategy_dispatch_owner_current WHERE owner_key='CENTRAL'`).Scan(&currentEpoch, &currentToken); err != nil {
		return fmt.Errorf("%w: current owner unavailable", ErrStrategyDispatchFenced)
	}
	if currentEpoch != epoch || currentToken != token {
		return ErrStrategyDispatchFenced
	}
	return nil
}

// RequireCurrentStrategyDispatchTransportAuthority is the final no-byte-sent
// fence. The caller supplies the CLAIMED revision it dispatched with; the
// atomic begin transaction must have advanced that exact lease once to
// SUBMITTING, and its owner and expiry must still be current.
func (j *Journal) RequireCurrentStrategyDispatchTransportAuthority(ctx context.Context, request StrategyDispatchLeaseCAS) error {
	if j == nil || j.db == nil || ctx == nil || !validStrategyDispatchIdentity(request.LeaseID) ||
		request.ExpectedRevision == 0 || request.ExpectedRevision == ^uint64(0) || request.OwnerEpoch == 0 ||
		!validStrategyDispatchIdentity(request.FencingToken) {
		return fmt.Errorf("%w: invalid strategy transport authority", ErrInvalidRequest)
	}
	tx, err := j.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if err := requireCurrentStrategyDispatchOwner(ctx, tx, request.OwnerEpoch, request.FencingToken); err != nil {
		return err
	}
	lease, err := loadStrategyDispatchLease(ctx, tx, request.LeaseID)
	if err != nil {
		return err
	}
	if lease.State != StrategyDispatchLeaseSubmitting || lease.Disposition != StrategyDispatchReservationReserved ||
		lease.Revision != request.ExpectedRevision+1 || lease.OwnerEpoch != request.OwnerEpoch ||
		lease.FencingToken != request.FencingToken || !j.clk.Now().UTC().Before(lease.ExpiresAt) {
		return ErrStrategyDispatchFenced
	}
	return nil
}

func validStrategyDispatchIdentity(value string) bool {
	if value == "" || value != strings.TrimSpace(value) || len(value) > 256 || !utf8.ValidString(value) {
		return false
	}
	for _, r := range value {
		if unicode.IsSpace(r) || unicode.IsControl(r) {
			return false
		}
	}
	return true
}

func requireStrategyDispatchOneRow(result sql.Result) error {
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows != 1 {
		return ErrStrategyDispatchFenced
	}
	return nil
}

func requireStrategyDispatchAtMostRows(result sql.Result, expected int64, authority string) error {
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("journal: %s affected rows: %w", authority, err)
	}
	if rows < 0 || rows > expected {
		return fmt.Errorf("%w: %s affected=%d max=%d", ErrStrategyDispatchLeaseUnavailable, authority, rows, expected)
	}
	return nil
}
