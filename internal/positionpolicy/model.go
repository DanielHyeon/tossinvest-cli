// Package positionpolicy defines the capability-neutral contract shared by the
// engine-owned command service and the operator console. It contains no journal,
// broker, config, or HTTP capability.
package positionpolicy

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/exitpolicy"
)

type Action string

const (
	ActionOverride Action = "OVERRIDE"
	ActionInherit  Action = "INHERIT"
	ActionRelease  Action = "RELEASE"
	ActionReadopt  Action = "READOPT"
)

type Status string

const (
	StatusManaged  Status = "MANAGED"
	StatusReleased Status = "RELEASED"
)

type Reason string

const (
	ReasonPolicyOverride Reason = "OPERATOR_POLICY_OVERRIDE"
	ReasonPolicyInherit  Reason = "OPERATOR_POLICY_INHERIT"
	ReasonRelease        Reason = "OPERATOR_RELEASE"
	ReasonReadopt        Reason = "OPERATOR_READOPT"
)

const ActorLocalOperator = "LOCAL_OPERATOR"

type Provenance string

const (
	ProvenanceUnknown          Provenance = "UNKNOWN"
	ProvenanceEngineEntry      Provenance = "ENGINE_ENTRY"
	ProvenanceExternalAdoption Provenance = "EXTERNAL_ADOPTION"
	ProvenanceAmbiguous        Provenance = "AMBIGUOUS"
)

type Eligibility string

const (
	EligibilityNone              Eligibility = "INELIGIBLE"
	EligibilityExitOnly          Eligibility = "EXIT_ONLY"
	EligibilityExternalLifecycle Eligibility = "EXTERNAL_LIFECYCLE"
)

var (
	ErrInvalidRequest       = errors.New("position policy: invalid request")
	ErrPositionNotFound     = errors.New("position policy: position not found")
	ErrVersionMismatch      = errors.New("position policy: version mismatch")
	ErrExitConflict         = errors.New("position policy: active exit conflict")
	ErrUnknownState         = errors.New("position policy: unknown lifecycle state")
	ErrIneligible           = errors.New("position policy: position is not exit eligible")
	ErrCapabilityInvalid    = errors.New("position policy: mutation capability is invalid or consumed")
	ErrCapabilityTooEarly   = errors.New("position policy: mutation capability is not active yet")
	ErrCapabilityExpired    = errors.New("position policy: mutation capability expired")
	ErrCapabilityStale      = errors.New("position policy: approved observation is stale or clock moved backward")
	ErrConfirmationRequired = errors.New("position policy: dangerous lifecycle action requires confirmation")
)

type State struct {
	PositionID         string
	AccountRef         string
	Market             string
	Symbol             string
	AdoptionGeneration int64
	Version            int64
	Status             Status
	DesiredPolicyID    string
	EffectivePolicyID  string
	ObservedAt         string
	UpdatedAt          string
	HasPendingExit     bool
	PositionState      string
	ExitEligible       bool
	Provenance         Provenance
	Eligibility        Eligibility
}

func ClassifyProvenance(entryDecisionID, adoptionID string) (Provenance, Eligibility) {
	entry := strings.TrimSpace(entryDecisionID) != ""
	adoption := strings.TrimSpace(adoptionID) != ""
	switch {
	case entry && adoption:
		return ProvenanceAmbiguous, EligibilityExitOnly
	case entry:
		return ProvenanceEngineEntry, EligibilityExitOnly
	case adoption:
		return ProvenanceExternalAdoption, EligibilityExternalLifecycle
	default:
		return ProvenanceUnknown, EligibilityNone
	}
}

func (s State) ExternalLifecycleEligible() bool {
	return s.Provenance == ProvenanceExternalAdoption && s.Eligibility == EligibilityExternalLifecycle
}

func (s State) DesiredLabel() string {
	if s.DesiredPolicyID == "" {
		return "공통 정책 상속"
	}
	return s.DesiredPolicyID
}

func (s State) EffectiveLabel() string {
	if s.EffectivePolicyID == "" {
		return "공통 정책 상속"
	}
	return s.EffectivePolicyID
}

type Request struct {
	PositionID         string
	ExpectedGeneration int64
	ExpectedVersion    int64
	Action             Action
	PolicyID           string
	Actor              string
	Reason             Reason
	At                 time.Time
	// ReAdoption is populated only by the engine-owned command service from an
	// authoritative price read. Console and HTTP clients cannot choose it.
	ReAdoption *ReAdoptionObservation `json:"re_adoption,omitempty"`
}

type ReAdoptionObservation struct {
	ObservedPrice string `json:"observed_price"`
	SyntheticStop string `json:"synthetic_stop"`
	ObservedAt    string `json:"observed_at"`
	PolicyID      string `json:"policy_id,omitempty"`
}

func (r Request) Validate() error {
	if strings.TrimSpace(r.PositionID) == "" || r.ExpectedGeneration < 1 || r.ExpectedVersion < 0 {
		return fmt.Errorf("%w: position, generation and non-negative version are required", ErrInvalidRequest)
	}
	if r.Actor != ActorLocalOperator {
		return fmt.Errorf("%w: actor is not the local operator enum", ErrInvalidRequest)
	}
	wantReason := map[Action]Reason{
		ActionOverride: ReasonPolicyOverride,
		ActionInherit:  ReasonPolicyInherit,
		ActionRelease:  ReasonRelease,
		ActionReadopt:  ReasonReadopt,
	}[r.Action]
	if wantReason == "" || r.Reason != wantReason {
		return fmt.Errorf("%w: action/reason pair is not server-defined", ErrInvalidRequest)
	}
	policyID := strings.TrimSpace(r.PolicyID)
	if r.Action == ActionOverride {
		if _, ok := exitpolicy.CommonPolicyByID(policyID); !ok {
			return fmt.Errorf("%w: unknown policy %q", ErrInvalidRequest, policyID)
		}
	} else if policyID != "" {
		return fmt.Errorf("%w: %s cannot carry a policy id", ErrInvalidRequest, r.Action)
	}
	if r.At.IsZero() {
		return fmt.Errorf("%w: command time is required", ErrInvalidRequest)
	}
	if r.Action == ActionReadopt {
		if r.ReAdoption == nil || strings.TrimSpace(r.ReAdoption.ObservedPrice) == "" ||
			strings.TrimSpace(r.ReAdoption.SyntheticStop) == "" || strings.TrimSpace(r.ReAdoption.ObservedAt) == "" {
			return fmt.Errorf("%w: re-adopt requires an engine observation and synthetic stop", ErrInvalidRequest)
		}
		if id := strings.TrimSpace(r.ReAdoption.PolicyID); id != "" {
			if _, ok := exitpolicy.CommonPolicyByID(id); !ok {
				return fmt.Errorf("%w: unknown re-adopt policy %q", ErrInvalidRequest, id)
			}
		}
	} else if r.ReAdoption != nil {
		return fmt.Errorf("%w: only re-adopt may carry an engine observation", ErrInvalidRequest)
	}
	return nil
}

type Preview struct {
	Before          State
	After           State
	Action          Action
	Reason          Reason
	RestartRequired bool
	ProtectionGap   bool
	// Capability is an engine-instance-local, one-time opaque mutation grant.
	// Its bytes reveal no scope and Apply accepts no replacement scope fields.
	Capability string `json:"capability,omitempty"`
}

type ApplyRequest struct {
	Capability string `json:"capability"`
	// Confirmed is meaningful only for the dangerous RELEASE and READOPT
	// actions already bound into Capability. It carries no mutation scope.
	Confirmed bool `json:"confirmed"`
}

type AuditEvent struct {
	ID                 int64
	PositionID         string
	AdoptionGeneration int64
	Action             Action
	BeforeJSON         string
	AfterJSON          string
	Actor              string
	Reason             Reason
	CreatedAt          string
}

type StopOption struct {
	ID      string
	Label   string
	Decimal string
}

type ManagementDescriptor struct {
	Category             string
	PositionSection      string
	AutoAdoptionSection  string
	AutoEnabledDefault   bool
	AutoEnabledDesired   bool
	AutoEnabledEffective bool
	StopDefault          string
	StopDesired          string
	StopEffective        string
	StopOptions          []StopOption
	IncludeDefault       []string
	ExcludeDefault       []string
	ExcludePrecedence    string
	ApplyTiming          string
	Provenance           string
	OneShareBehavior     string
}

// Descriptor is server-authoritative metadata for the a050 position-management
// category. Percent values are generated from integer basis points, avoiding a
// floating-point step that can omit or duplicate an option.
func Descriptor() ManagementDescriptor {
	options := make([]StopOption, 0, 37)
	for basisPoints := 200; basisPoints <= 2000; basisPoints += 50 {
		whole, half := basisPoints/100, basisPoints%100
		label := fmt.Sprintf("%d%%", whole)
		if half != 0 {
			label = fmt.Sprintf("%d.5%%", whole)
		}
		options = append(options, StopOption{
			ID: fmt.Sprintf("stop-%04d-bp", basisPoints), Label: label,
			Decimal: fmt.Sprintf("0.%04d", basisPoints),
		})
	}
	return ManagementDescriptor{
		Category: "position-management", PositionSection: "종목별 정책",
		AutoAdoptionSection: "외부 매수 자동편입", AutoEnabledDefault: false,
		AutoEnabledDesired: false, AutoEnabledEffective: false,
		StopDefault: "5%", StopDesired: "5%", StopEffective: "5%", StopOptions: options,
		IncludeDefault: []string{}, ExcludeDefault: []string{},
		ExcludePrecedence: "exclude 우선", ApplyTiming: "다음 adoption generation부터",
		Provenance: "server registry · a044", OneShareBehavior: "중간 익절 없음 · 보호선 승격 · 최종/손절 시 1주 전량",
	}
}
