package strategydispatch

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyengine"
)

type Reason string

const (
	ReasonDecisionInvalid Reason = "decision_invalid"
	ReasonDecisionStale   Reason = "decision_stale"
	ReasonLaneOff         Reason = "lane_off"
	ReasonKillSwitch      Reason = "kill_switch"
	ReasonProtection      Reason = "protection_unwired"
	ReasonReconcile       Reason = "reconcile_degraded"
	ReasonScheduler       Reason = "scheduler_invalid"
	ReasonAutoStart       Reason = "autostart_off"
	ReasonGate            Reason = "gate_closed"
	ReasonLive            Reason = "live_unapproved"
	ReasonActivation      Reason = "activation_invalid"
	ReasonGuardian        Reason = "guardian_refused"
	ReasonPlan            Reason = "durable_plan_failed"
	ReasonTOCTOU          Reason = "post_plan_gate_changed"
	ReasonGateway         Reason = "official_gateway_failed"
)

type Error struct {
	Reason Reason
	Cause  error
}

func (e *Error) Error() string { return "strategy dispatch: " + string(e.Reason) }
func (e *Error) Unwrap() error { return e.Cause }

type ManifestBinding struct {
	AccountRef              string
	Profile                 string
	BuildDigest             string
	CommitDigest            string
	LaneID                  string
	LaneVersion             string
	LaneSourceDigest        string
	LaneConstantsDigest     string
	ThresholdVersion        string
	ThresholdSetDigest      string
	EvidenceDigest          string
	SettingsDigest          string
	AttestationDigest       string
	AttestationExpiresAt    time.Time
	GuardianVersion         string
	GuardianLimitsDigest    string
	ReconciliationWatermark string
	ProtectionProfile       string
	ProtectionState         string
	OperatingPolicy         string
	SchedulerScope          string
	CalendarVersion         string
	LaneApproved            bool
	SchedulerApproved       bool
	AutoStartApproved       bool
	GateApproved            bool
	LiveApproved            bool
	Actor                   string
	AuditID                 string
	IssuedAt                time.Time
	ExpiresAt               time.Time
	Generation              uint64
}

type manifest struct {
	binding ManifestBinding
	digest  string
	revoked bool
}

type Activation struct {
	digest string
	valid  bool
}

func (a Activation) Digest() string {
	if !a.valid {
		return ""
	}
	return a.digest
}

type ManifestRepository struct {
	mu      sync.RWMutex
	current manifest
}

func NewDormantManifestRepository() *ManifestRepository { return &ManifestRepository{} }

var ErrNotConfigured = errors.New("strategy activation: not configured")

func (r *ManifestRepository) Verify(expected ManifestBinding, now time.Time) (Activation, error) {
	if r == nil {
		return Activation{}, ErrNotConfigured
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.verifyLocked(expected, now)
}

// WithVerifiedLease holds the repository read lease through fn. A revocation
// that wins before the lease causes call zero; one arriving after it waits for
// the official call to finish instead of opening a validation/call race.
func (r *ManifestRepository) WithVerifiedLease(expected ManifestBinding, now time.Time, fn func(Activation) error) error {
	if r == nil || fn == nil {
		return ErrNotConfigured
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	activation, err := r.verifyLocked(expected, now)
	if err != nil {
		return err
	}
	return fn(activation)
}

func (r *ManifestRepository) verifyLocked(expected ManifestBinding, now time.Time) (Activation, error) {
	if now.IsZero() || r.current.digest == "" {
		return Activation{}, ErrNotConfigured
	}
	digest, _ := manifestDigest(r.current.binding)
	if r.current.revoked || digest != r.current.digest || r.current.binding != expected ||
		expected.Generation == 0 || !expected.IssuedAt.Before(expected.ExpiresAt) ||
		now.Before(expected.IssuedAt) || !now.Before(expected.ExpiresAt) ||
		!now.Before(expected.AttestationExpiresAt) || expected.ProtectionState != "WIRED" ||
		!expected.LaneApproved || !expected.SchedulerApproved || !expected.AutoStartApproved ||
		!expected.GateApproved || !expected.LiveApproved {
		return Activation{}, errors.New("strategy activation: invalid")
	}
	return Activation{digest: r.current.digest, valid: true}, nil
}

func manifestDigest(binding ManifestBinding) (string, error) {
	raw, err := json.Marshal(binding)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(raw)
	return "sha256:" + hex.EncodeToString(sum[:]), nil
}

// DecisionBinding is the complete scalar DecisionRecord at the post-decision
// gate. Defining it from DecisionRecord makes every provenance, time, market
// input, indicator, bar, live-price and accept-reason field part of Go's exact
// comparable equality. A newly added DecisionRecord field therefore cannot be
// silently omitted from this laundering guard.
type DecisionBinding strategyengine.DecisionRecord

type OrderSettings struct {
	OrderType      string
	Currency       string
	SettingsDigest string
}

type GateSnapshot struct {
	Binding          ManifestBinding
	Decision         DecisionBinding
	Order            OrderSettings
	LaneDesired      bool
	LaneEffective    bool
	ProtectionWired  bool
	ReconcileHealthy bool
	SchedulerValid   bool
	AutoStart        bool
	GateOpen         bool
	LiveApproved     bool
	KillSwitch       bool
	Revision         uint64
}

type GateReader interface {
	ReadGate(context.Context) (GateSnapshot, error)
	WithLease(context.Context, uint64, func(GateSnapshot) error) error
}

type AtomicPlan struct {
	AttemptID          string
	Decision           strategyengine.Decision
	RiskIntentID       string
	GuardianDecisionID string
	GuardianGeneration int
	ManifestDigest     string
	ClientOrderID      string
	Quantity           string
	Order              OrderSettings
	GateRevision       uint64
}

type PlanReceipt struct {
	AttemptID        string
	AccountRef       string
	DecisionIdentity string
	RiskIntentID     string
	ClientOrderID    string
	Quantity         string
	Revision         uint64
	State            string
}

type IssueRequest struct {
	Decision       strategyengine.Decision
	AttemptID      string
	ManifestDigest string
	Binding        ManifestBinding
	Order          OrderSettings
}

type StrategyIssuer interface {
	IssueAndPlan(context.Context, IssueRequest) (AtomicPlan, PlanReceipt, error)
	RecordStrategyRefusal(context.Context, PlanReceipt, Reason) error
	RecordStrategyInDoubt(context.Context, PlanReceipt, Reason) error
	RecordStrategyDispatched(context.Context, PlanReceipt, string, string) error
}

type OfficialGateway interface {
	PlaceStrategyEntry(context.Context, AtomicPlan) (execgw.Outcome, error)
}

type outcomeDisposition string

const (
	outcomeDispatched outcomeDisposition = "DISPATCHED"
	outcomeRefused    outcomeDisposition = "REFUSED"
	outcomeInDoubt    outcomeDisposition = "IN_DOUBT"
)

type Dependencies struct {
	Gates    GateReader
	Manifest *ManifestRepository
	Issuer   StrategyIssuer
	Gateway  OfficialGateway
	Now      func() time.Time
}

func Dispatch(ctx context.Context, decision strategyengine.Decision, deps Dependencies) error {
	if !decision.Valid() {
		return &Error{Reason: ReasonDecisionInvalid}
	}
	return dispatchValidated(ctx, decision, decision.Record(), deps)
}

// dispatchValidated is the package-private test seam after the opaque decision
// gate. Production reaches it only through Dispatch; tests can exercise lease,
// terminal-classification and TOCTOU branches without a decision minter.
func dispatchValidated(ctx context.Context, decision strategyengine.Decision, payload strategyengine.DecisionRecord, deps Dependencies) error {
	if deps.Gates == nil || deps.Manifest == nil || deps.Issuer == nil ||
		deps.Gateway == nil || deps.Now == nil {
		return &Error{Reason: ReasonActivation}
	}
	now := deps.Now()
	if now.IsZero() || now.UnixNano() >= payload.ExpiresAt {
		return &Error{Reason: ReasonDecisionStale}
	}
	first, err := deps.Gates.ReadGate(ctx)
	if err != nil {
		return &Error{Reason: ReasonGate, Cause: err}
	}
	if reason := checkGate(first, payload); reason != "" {
		return &Error{Reason: reason}
	}
	activation, err := deps.Manifest.Verify(first.Binding, now)
	if err != nil {
		return &Error{Reason: ReasonActivation, Cause: err}
	}
	plan, receipt, err := deps.Issuer.IssueAndPlan(ctx, IssueRequest{
		Decision: decision, AttemptID: deterministicAttemptID(first.Binding.AccountRef, payload.Identity, first.Binding.Generation), ManifestDigest: activation.Digest(),
		Binding: first.Binding, Order: first.Order,
	})
	if err != nil {
		return &Error{Reason: ReasonGuardian, Cause: err}
	}
	plan.GateRevision = first.Revision

	leaseEntered := false
	gatewayCalled := false
	var officialOutcome execgw.Outcome
	var officialErr error
	err = deps.Gates.WithLease(ctx, first.Revision, func(current GateSnapshot) error {
		leaseEntered = true
		if current.Revision != first.Revision || current.Binding != first.Binding || current.Decision != first.Decision {
			return &Error{Reason: ReasonTOCTOU}
		}
		if reason := checkGate(current, payload); reason != "" {
			return &Error{Reason: ReasonTOCTOU}
		}
		manifestErr := deps.Manifest.WithVerifiedLease(current.Binding, deps.Now(), func(currentActivation Activation) error {
			if currentActivation.Digest() != plan.ManifestDigest {
				return &Error{Reason: ReasonTOCTOU}
			}
			dispatchNow := deps.Now()
			if dispatchNow.IsZero() || dispatchNow.UnixNano() >= payload.ExpiresAt {
				return &Error{Reason: ReasonTOCTOU, Cause: errors.New("strategy decision expired after planning")}
			}
			gatewayCalled = true
			officialOutcome, officialErr = deps.Gateway.PlaceStrategyEntry(ctx, plan)
			return officialErr
		})
		if manifestErr != nil && !gatewayCalled {
			return &Error{Reason: ReasonTOCTOU, Cause: manifestErr}
		}
		return manifestErr
	})
	if err == nil {
		disposition, outcomeErr := classifyOfficialOutcome(officialOutcome, officialErr)
		switch disposition {
		case outcomeDispatched:
			if recordErr := deps.Issuer.RecordStrategyDispatched(ctx, receipt, officialOutcome.AttemptID, officialOutcome.BrokerOrderID); recordErr != nil {
				return &Error{Reason: ReasonPlan, Cause: recordErr}
			}
			return nil
		case outcomeRefused:
			if recordErr := deps.Issuer.RecordStrategyRefusal(ctx, receipt, ReasonGateway); recordErr != nil {
				return &Error{Reason: ReasonPlan, Cause: errors.Join(outcomeErr, recordErr)}
			}
			return &Error{Reason: ReasonGateway, Cause: outcomeErr}
		default:
			if recordErr := deps.Issuer.RecordStrategyInDoubt(ctx, receipt, ReasonGateway); recordErr != nil {
				return &Error{Reason: ReasonPlan, Cause: errors.Join(outcomeErr, recordErr)}
			}
			return &Error{Reason: ReasonGateway, Cause: outcomeErr}
		}
	}
	var typed *Error
	if !leaseEntered || (errors.As(err, &typed) && typed.Reason == ReasonTOCTOU) {
		if recordErr := deps.Issuer.RecordStrategyRefusal(ctx, receipt, ReasonTOCTOU); recordErr != nil {
			return &Error{Reason: ReasonPlan, Cause: errors.Join(err, recordErr)}
		}
		if typed != nil {
			return typed
		}
		return &Error{Reason: ReasonTOCTOU, Cause: err}
	}
	disposition, outcomeErr := classifyOfficialOutcome(officialOutcome, err)
	if disposition == outcomeRefused {
		if recordErr := deps.Issuer.RecordStrategyRefusal(ctx, receipt, ReasonGateway); recordErr != nil {
			return &Error{Reason: ReasonPlan, Cause: errors.Join(outcomeErr, recordErr)}
		}
		return &Error{Reason: ReasonGateway, Cause: outcomeErr}
	}
	if disposition == outcomeDispatched {
		if recordErr := deps.Issuer.RecordStrategyDispatched(ctx, receipt, officialOutcome.AttemptID, officialOutcome.BrokerOrderID); recordErr != nil {
			return &Error{Reason: ReasonPlan, Cause: recordErr}
		}
		return nil
	}
	if recordErr := deps.Issuer.RecordStrategyInDoubt(ctx, receipt, ReasonGateway); recordErr != nil {
		return &Error{Reason: ReasonPlan, Cause: errors.Join(outcomeErr, recordErr)}
	}
	return &Error{Reason: ReasonGateway, Cause: outcomeErr}
}

func classifyOfficialOutcome(outcome execgw.Outcome, callErr error) (outcomeDisposition, error) {
	if callErr == nil && outcome.State == journal.StateConfirmed && outcome.AttemptID != "" && outcome.BrokerOrderID != "" {
		return outcomeDispatched, nil
	}
	if callErr == nil {
		callErr = errors.New("strategy dispatch: official outcome is not exact confirmed")
	}
	if outcome.AttemptID == "" || outcome.State == journal.StateNotDispatched || outcome.State == journal.StateFailedConfirmed {
		return outcomeRefused, callErr
	}
	return outcomeInDoubt, callErr
}

func deterministicAttemptID(accountRef, decisionIdentity string, generation uint64) string {
	raw := accountRef + "\x00" + decisionIdentity + "\x00" + fmt.Sprintf("%d", generation)
	sum := sha256.Sum256([]byte(raw))
	return "strategy-attempt:v1:sha256:" + hex.EncodeToString(sum[:])
}

func checkGate(gate GateSnapshot, payload strategyengine.DecisionRecord) Reason {
	binding := gate.Binding
	if payload.LaneID != binding.LaneID || payload.LaneVersion != binding.LaneVersion ||
		payload.SourceDigest != binding.LaneSourceDigest || payload.ConstantsDigest != binding.LaneConstantsDigest ||
		payload.ThresholdVersion != binding.ThresholdVersion || payload.ThresholdSetDigest != binding.ThresholdSetDigest ||
		payload.EvidenceDigest != binding.EvidenceDigest {
		return ReasonActivation
	}
	wantDecision := DecisionBinding(payload)
	if gate.Decision != wantDecision || gate.Order.SettingsDigest != binding.SettingsDigest ||
		gate.Order.OrderType != "LIMIT" || gate.Order.Currency != "KRW" {
		return ReasonActivation
	}
	switch {
	case !gate.LaneDesired || !gate.LaneEffective:
		return ReasonLaneOff
	case gate.KillSwitch:
		return ReasonKillSwitch
	case !gate.ProtectionWired:
		return ReasonProtection
	case !gate.ReconcileHealthy:
		return ReasonReconcile
	case !gate.SchedulerValid:
		return ReasonScheduler
	case !gate.AutoStart:
		return ReasonAutoStart
	case !gate.GateOpen:
		return ReasonGate
	case !gate.LiveApproved:
		return ReasonLive
	}
	return ""
}
