package journal

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

var (
	ErrStrategyDispatchFenced = errors.New("journal: strategy dispatch owner fenced")
	// ErrStrategyDispatchDormant is returned by every authority-bearing v25 API.
	// This checkpoint deliberately has no production mint for activation,
	// ProtectionReady, Gateway outcome, or reservation release authority.
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

// ClaimStrategyDispatchLease remains closed until claim and real reservation
// release/transfer share the Gateway-owned transaction.
func (j *Journal) ClaimStrategyDispatchLease(context.Context, StrategyDispatchLeaseCAS) (StrategyDispatchLease, error) {
	return StrategyDispatchLease{}, ErrStrategyDispatchDormant
}

// BeginStrategyDispatchSubmitting remains closed for the same reason.
func (j *Journal) BeginStrategyDispatchSubmitting(context.Context, StrategyDispatchLeaseCAS) (StrategyDispatchLease, error) {
	return StrategyDispatchLease{}, ErrStrategyDispatchDormant
}

// RecoverClaimedStrategyDispatchLease remains closed: using a shadow
// disposition or mislabelling a pre-transport release as BROKER_TERMINAL would
// fabricate authority. DiscoverStrategyDispatchRecovery is the read-only cold
// restart handoff until the official adapter lands.
func (j *Journal) RecoverClaimedStrategyDispatchLease(context.Context, StrategyDispatchLeaseCAS) (StrategyDispatchLease, error) {
	return StrategyDispatchLease{}, ErrStrategyDispatchDormant
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
	for value, target := range map[string]*time.Time{issued: &lease.IssuedAt, expires: &lease.ExpiresAt, created: &lease.CreatedAt, updated: &lease.UpdatedAt} {
		parsed, parseErr := parseJournalTime(value)
		if parseErr != nil {
			return StrategyDispatchLease{}, parseErr
		}
		*target = parsed
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
