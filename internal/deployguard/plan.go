package deployguard

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

type ActionKind string

const (
	ActionReplaceAndVerify          ActionKind = "REPLACE_AND_VERIFY"
	ActionReadRollbackCompatibility ActionKind = "READ_ROLLBACK_COMPATIBILITY"
	ActionRollbackAndVerify         ActionKind = "ROLLBACK_AND_VERIFY"
)

// Action is data, not an executable capability. A separate human-controlled
// deployment system may interpret it only after its own authorization checks.
type Action struct {
	ID          string
	Kind        ActionKind
	Service     string
	ImageDigest Digest
	Timeout     time.Duration
	IssuedAt    time.Time
	Deadline    time.Time
}

type Status string

const (
	StatusRunning          Status = "RUNNING"
	StatusSucceeded        Status = "SUCCEEDED"
	StatusFailed           Status = "FAILED"
	StatusRolledBack       Status = "ROLLED_BACK"
	StatusRecoveryRequired Status = "RECOVERY_REQUIRED"
)

type RecoveryCode string

const (
	RecoveryReplaceFailed        RecoveryCode = "REPLACE_FAILED"
	RecoverySchemaDrift          RecoveryCode = "SCHEMA_DRIFT"
	RecoveryStateDrift           RecoveryCode = "STATE_DRIFT"
	RecoveryRollbackReadFailed   RecoveryCode = "ROLLBACK_COMPATIBILITY_FAILED"
	RecoveryRollbackIncompatible RecoveryCode = "ROLLBACK_INCOMPATIBLE"
	RecoveryRollbackFailed       RecoveryCode = "ROLLBACK_FAILED"
)

type Recovery struct {
	Code                RecoveryCode
	Service             string
	EntryEffective      State
	RetainedImageDigest Digest
}

type ReplaceOutcome string

const (
	ReplaceNotApplied ReplaceOutcome = "NOT_APPLIED"
	ReplaceApplied    ReplaceOutcome = "APPLIED"
)

type Result struct {
	ActionID        string
	Service         string
	ImageDigest     Digest
	ReplaceOutcome  ReplaceOutcome
	TimedOut        bool
	Health          HealthEvidence
	SchemaVersion   uint64
	State           StateEvidence
	EnvironmentKeys []string
	Mounts          []MountIdentity
}

type successfulReplacement struct {
	serviceIndex int
}

type Execution struct {
	plan           Plan
	status         Status
	nextService    int
	sequence       uint64
	pending        Action
	succeeded      []successfulReplacement
	rollbackCursor int
	rollbackSchema uint64
	recoveries     []Recovery
}

func (e Execution) Status() Status { return e.status }

func (e Execution) Recoveries() []Recovery { return append([]Recovery(nil), e.recoveries...) }

func Start(plan Plan, issuedAt time.Time) (Execution, Action, error) {
	if err := validatePlan(plan); err != nil {
		return Execution{}, Action{}, err
	}
	if !validTrustedTime(issuedAt) || issuedAt.Before(plan.preimage.CapturedAt) {
		return Execution{}, Action{}, errors.New("deploy guard: trusted action issue time is invalid")
	}
	execution := Execution{plan: clonePlan(plan), status: StatusRunning, rollbackCursor: -1}
	action := execution.action(ActionReplaceAndVerify, 0, issuedAt)
	execution.pending = action
	return execution, action, nil
}

func Advance(execution Execution, result Result, receivedAt time.Time) (Execution, *Action, error) {
	if execution.status != StatusRunning || execution.pending.ID == "" || result.ActionID != execution.pending.ID {
		return execution, nil, errors.New("deploy guard: result does not match the single pending action")
	}
	if err := validateResultBinding(execution, result, receivedAt); err != nil {
		return execution, nil, err
	}
	next := cloneExecution(execution)
	if result.TimedOut {
		return advanceTimeout(next, result, receivedAt)
	}
	switch next.pending.Kind {
	case ActionReplaceAndVerify:
		return advanceReplacement(next, result, receivedAt)
	case ActionReadRollbackCompatibility:
		return advanceCompatibilityRead(next, result, receivedAt)
	case ActionRollbackAndVerify:
		return advanceRollback(next, result, receivedAt)
	default:
		return execution, nil, errors.New("deploy guard: invalid pending action")
	}
}

func advanceTimeout(execution Execution, result Result, receivedAt time.Time) (Execution, *Action, error) {
	serviceIndex := execution.serviceIndexForPending()
	service := execution.plan.preimage.Services[serviceIndex]
	switch execution.pending.Kind {
	case ActionReplaceAndVerify:
		if result.ReplaceOutcome == ReplaceApplied {
			execution.succeeded = append(execution.succeeded, successfulReplacement{serviceIndex: serviceIndex})
		}
		execution.recoveries = append(execution.recoveries, Recovery{Code: RecoveryReplaceFailed, Service: service.Name,
			EntryEffective: observedEntry(result)})
		return startRollback(execution, receivedAt)
	case ActionReadRollbackCompatibility:
		execution.recoveries = append(execution.recoveries, Recovery{Code: RecoveryRollbackReadFailed, Service: service.Name,
			EntryEffective: observedEntry(result), RetainedImageDigest: service.TargetImageDigest})
		return continueRollback(execution, true, receivedAt)
	case ActionRollbackAndVerify:
		execution.recoveries = append(execution.recoveries, Recovery{Code: RecoveryRollbackFailed, Service: service.Name,
			EntryEffective: observedEntry(result), RetainedImageDigest: service.TargetImageDigest})
		execution.pending = Action{}
		execution.status = StatusRecoveryRequired
		return execution, nil, nil
	default:
		return execution, nil, errors.New("deploy guard: timeout refers to an invalid action")
	}
}

func advanceReplacement(execution Execution, result Result, receivedAt time.Time) (Execution, *Action, error) {
	serviceIndex := execution.nextService
	service := execution.plan.preimage.Services[serviceIndex]
	code := replacementFailure(service, execution.plan.preimage.State, result)
	if code != "" {
		if result.ReplaceOutcome == ReplaceApplied {
			execution.succeeded = append(execution.succeeded, successfulReplacement{serviceIndex: serviceIndex})
		}
		execution.recoveries = append(execution.recoveries, Recovery{Code: code, Service: service.Name,
			EntryEffective: observedEntry(result)})
		return startRollback(execution, receivedAt)
	}
	execution.succeeded = append(execution.succeeded, successfulReplacement{serviceIndex: serviceIndex})
	execution.nextService++
	if execution.nextService == len(execution.plan.preimage.Services) {
		execution.pending = Action{}
		execution.status = StatusSucceeded
		return execution, nil, nil
	}
	action := execution.action(ActionReplaceAndVerify, execution.nextService, receivedAt)
	execution.pending = action
	return execution, &action, nil
}

func replacementFailure(service ServicePreimage, state StateEvidence, result Result) RecoveryCode {
	if !preservationMatches(service, state, result) {
		return RecoveryStateDrift
	}
	if result.SchemaVersion != service.PostReplaceSchemaVersion {
		return RecoverySchemaDrift
	}
	if !result.Health.Healthy {
		return RecoveryReplaceFailed
	}
	return ""
}

func startRollback(execution Execution, issuedAt time.Time) (Execution, *Action, error) {
	if len(execution.succeeded) == 0 {
		execution.pending = Action{}
		execution.status = StatusFailed
		return execution, nil, nil
	}
	execution.rollbackCursor = len(execution.succeeded) - 1
	serviceIndex := execution.succeeded[execution.rollbackCursor].serviceIndex
	action := execution.action(ActionReadRollbackCompatibility, serviceIndex, issuedAt)
	execution.pending = action
	return execution, &action, nil
}

func advanceCompatibilityRead(execution Execution, result Result, receivedAt time.Time) (Execution, *Action, error) {
	serviceIndex := execution.succeeded[execution.rollbackCursor].serviceIndex
	service := execution.plan.preimage.Services[serviceIndex]
	if !result.Health.Healthy || !preservationMatches(service, execution.plan.preimage.State, result) {
		execution.recoveries = append(execution.recoveries, Recovery{Code: RecoveryRollbackReadFailed, Service: service.Name,
			EntryEffective: observedEntry(result), RetainedImageDigest: service.TargetImageDigest})
		return continueRollback(execution, true, receivedAt)
	}
	if service.RollbackSchema.Readable.Contains(result.SchemaVersion) && service.RollbackSchema.Writable.Contains(result.SchemaVersion) {
		execution.rollbackSchema = result.SchemaVersion
		action := execution.action(ActionRollbackAndVerify, serviceIndex, receivedAt)
		execution.pending = action
		return execution, &action, nil
	}
	execution.recoveries = append(execution.recoveries, Recovery{Code: RecoveryRollbackIncompatible, Service: service.Name,
		EntryEffective: observedEntry(result), RetainedImageDigest: service.TargetImageDigest})
	return continueRollback(execution, true, receivedAt)
}

func advanceRollback(execution Execution, result Result, receivedAt time.Time) (Execution, *Action, error) {
	serviceIndex := execution.succeeded[execution.rollbackCursor].serviceIndex
	service := execution.plan.preimage.Services[serviceIndex]
	if !result.Health.Healthy || result.SchemaVersion != execution.rollbackSchema ||
		!preservationMatches(service, execution.plan.preimage.State, result) {
		execution.recoveries = append(execution.recoveries, Recovery{Code: RecoveryRollbackFailed, Service: service.Name,
			EntryEffective: observedEntry(result), RetainedImageDigest: service.TargetImageDigest})
		execution.pending = Action{}
		execution.status = StatusRecoveryRequired
		return execution, nil, nil
	}
	return continueRollback(execution, false, receivedAt)
}

func validateResultBinding(execution Execution, result Result, receivedAt time.Time) error {
	serviceIndex := execution.serviceIndexForPending()
	if serviceIndex < 0 {
		return errors.New("deploy guard: pending service is outside the frozen manifest")
	}
	service := execution.plan.preimage.Services[serviceIndex]
	if !validActionWindow(execution.pending) || result.TimedOut != receivedAt.After(execution.pending.Deadline) ||
		result.TimedOut && result.Health.Healthy {
		return errors.New("deploy guard: result does not match the sealed action window")
	}
	if result.Service != service.Name || !validHealthObservation(result.Health, execution.pending, receivedAt) {
		return errors.New("deploy guard: observed service or health evidence does not match the pending action")
	}
	wantImage := Digest("")
	switch execution.pending.Kind {
	case ActionReplaceAndVerify:
		switch result.ReplaceOutcome {
		case ReplaceApplied:
			wantImage = service.TargetImageDigest
		case ReplaceNotApplied:
			wantImage = service.CurrentImageDigest
		default:
			return errors.New("deploy guard: replacement outcome is required")
		}
		if result.Health.Healthy && result.ReplaceOutcome != ReplaceApplied {
			return errors.New("deploy guard: healthy replacement was not applied")
		}
	case ActionReadRollbackCompatibility:
		if result.ReplaceOutcome != "" {
			return errors.New("deploy guard: read-only compatibility result carries replacement outcome")
		}
		wantImage = service.TargetImageDigest
	case ActionRollbackAndVerify:
		if result.ReplaceOutcome != "" {
			return errors.New("deploy guard: rollback result carries replacement outcome")
		}
		wantImage = service.CurrentImageDigest
	default:
		return errors.New("deploy guard: invalid pending action")
	}
	if result.ImageDigest != wantImage {
		return errors.New("deploy guard: observed running image differs from the sealed action")
	}
	wantEvidence, err := ObservationDigest(execution.pending, result)
	if err != nil || result.Health.EvidenceDigest != wantEvidence {
		return errors.New("deploy guard: canonical observation evidence does not match the pending action")
	}
	return nil
}

func completePreservationEvidence(result Result) bool {
	state := result.State
	for _, digest := range []Digest{state.RenderedComposeDigest, state.ConfigDigest, state.ActivationDigest,
		state.LaneDigest, state.AutostartDigest, state.AutomationDigest, state.LiveApprovalDigest,
		state.ProtectionDigest, state.JournalDigest} {
		if !validDigest(digest) {
			return false
		}
	}
	if len(state.Markets) != 2 {
		return false
	}
	for _, market := range []Market{MarketKR, MarketUS} {
		if _, ok := state.Markets[market]; !ok {
			return false
		}
	}
	return canonicalEnvironmentKeys(result.EnvironmentKeys) && canonicalMounts(result.Mounts)
}

func observedEntry(result Result) State {
	if !completePreservationEvidence(result) {
		return StateUnknown
	}
	kr := result.State.Markets[MarketKR].EntryEffective
	us := result.State.Markets[MarketUS].EntryEffective
	if kr != us || kr != StateOff && kr != StateOn {
		return StateUnknown
	}
	return kr
}

// ObservationDigest binds a health observation to the exact sealed action,
// running image, schema and preservation evidence. The digest field itself is
// deliberately omitted from the canonical body.
func ObservationDigest(action Action, result Result) (Digest, error) {
	if !validActionWindow(action) {
		return "", errors.New("deploy guard: invalid action window")
	}
	body := struct {
		ActionID        string
		Kind            ActionKind
		Service         string
		ImageDigest     Digest
		Timeout         time.Duration
		IssuedAt        time.Time
		Deadline        time.Time
		ReplaceOutcome  ReplaceOutcome
		TimedOut        bool
		Healthy         bool
		ObservedAt      time.Time
		SchemaVersion   uint64
		State           StateEvidence
		EnvironmentKeys []string
		Mounts          []MountIdentity
	}{
		ActionID: action.ID, Kind: action.Kind, Service: result.Service, ImageDigest: result.ImageDigest,
		Timeout: action.Timeout, IssuedAt: action.IssuedAt, Deadline: action.Deadline,
		ReplaceOutcome: result.ReplaceOutcome, TimedOut: result.TimedOut, Healthy: result.Health.Healthy,
		ObservedAt: result.Health.ObservedAt, SchemaVersion: result.SchemaVersion, State: result.State,
		EnvironmentKeys: result.EnvironmentKeys, Mounts: result.Mounts,
	}
	encoded, err := json.Marshal(body)
	if err != nil {
		return "", fmt.Errorf("deploy guard: encode observation: %w", err)
	}
	return DigestBytes(encoded), nil
}

func (e Execution) serviceIndexForPending() int {
	for index, service := range e.plan.preimage.Services {
		if service.Name == e.pending.Service {
			return index
		}
	}
	return -1
}

func validHealthObservation(evidence HealthEvidence, action Action, receivedAt time.Time) bool {
	if !validTrustedTime(receivedAt) || !validTrustedTime(evidence.ObservedAt) ||
		!evidence.ObservedAt.After(action.IssuedAt) || evidence.ObservedAt.After(action.Deadline) ||
		receivedAt.Before(evidence.ObservedAt) {
		return false
	}
	return validDigest(evidence.EvidenceDigest)
}

func continueRollback(execution Execution, incompatible bool, issuedAt time.Time) (Execution, *Action, error) {
	execution.rollbackCursor--
	if execution.rollbackCursor < 0 {
		execution.pending = Action{}
		if incompatible || hasRecovery(execution.recoveries, RecoveryRollbackIncompatible) || hasRecovery(execution.recoveries, RecoveryRollbackFailed) {
			execution.status = StatusRecoveryRequired
		} else {
			execution.status = StatusRolledBack
		}
		return execution, nil, nil
	}
	serviceIndex := execution.succeeded[execution.rollbackCursor].serviceIndex
	action := execution.action(ActionReadRollbackCompatibility, serviceIndex, issuedAt)
	execution.pending = action
	return execution, &action, nil
}

func hasRecovery(recoveries []Recovery, code RecoveryCode) bool {
	for _, recovery := range recoveries {
		if recovery.Code == code {
			return true
		}
	}
	return false
}

func (e *Execution) action(kind ActionKind, serviceIndex int, issuedAt time.Time) Action {
	e.sequence++
	service := e.plan.preimage.Services[serviceIndex]
	image, timeout := Digest(""), time.Duration(0)
	switch kind {
	case ActionReplaceAndVerify:
		image, timeout = service.TargetImageDigest, service.Timeout
	case ActionReadRollbackCompatibility:
		image, timeout = service.TargetImageDigest, service.Timeout
	case ActionRollbackAndVerify:
		image, timeout = service.CurrentImageDigest, service.Timeout
	}
	deadline := issuedAt.Add(timeout)
	id := DigestBytes([]byte(fmt.Sprintf("%s\x00%d\x00%s\x00%s\x00%s\x00%s", e.plan.digest, e.sequence, kind,
		service.Name, issuedAt.Format(time.RFC3339Nano), deadline.Format(time.RFC3339Nano))))
	return Action{ID: string(id), Kind: kind, Service: service.Name, ImageDigest: image, Timeout: timeout,
		IssuedAt: issuedAt, Deadline: deadline}
}

func validTrustedTime(value time.Time) bool {
	return !value.IsZero() && value.Location() == time.UTC
}

func validActionWindow(action Action) bool {
	return action.ID != "" && action.Kind != "" && action.Service != "" && validDigest(action.ImageDigest) &&
		action.Timeout > 0 && action.Timeout <= MaxServiceTimeout && validTrustedTime(action.IssuedAt) &&
		validTrustedTime(action.Deadline) && action.Deadline.Equal(action.IssuedAt.Add(action.Timeout))
}

func validatePlan(plan Plan) error {
	if plan.digest == "" {
		return errors.New("deploy guard: empty plan")
	}
	if err := validatePreimage(plan.preimage); err != nil {
		return err
	}
	body, err := json.Marshal(plan.preimage)
	if err != nil || DigestBytes(body) != plan.digest {
		return errors.New("deploy guard: plan seal mismatch")
	}
	return nil
}

func clonePlan(plan Plan) Plan {
	return Plan{preimage: clonePreimage(plan.preimage), digest: plan.digest}
}

func cloneExecution(input Execution) Execution {
	out := input
	out.plan = clonePlan(input.plan)
	out.succeeded = append([]successfulReplacement(nil), input.succeeded...)
	out.recoveries = append([]Recovery(nil), input.recoveries...)
	return out
}
