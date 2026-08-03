// Package reversallane provides pure KR and US short-horizon reversal
// evaluation. It owns no persistence, dispatch, activation, or exit authority.
package reversallane

import (
	"errors"
	"time"
)

type Market string

const (
	MarketKR Market = "KR"
	MarketUS Market = "US"
)

const (
	KRReversalLaneID = "kr_short_absorption_reversal_v1"
	USReversalLaneID = "us_short_dislocation_reversal_v1"
	LaneVersionV1    = "v1"
	ReversalRelease  = "a068-kr-us-reversal-v1"
)

type LaneState string

const StateOff LaneState = "OFF"

type Descriptor struct {
	ID             string
	Version        string
	Market         Market
	Release        string
	DesiredState   LaneState
	EffectiveState LaneState
}

type RefusalCode string

const (
	RefusalDisabled              RefusalCode = "LANE_DISABLED"
	RefusalStrictSchema          RefusalCode = "STRICT_SCHEMA_INVALID"
	RefusalScopeMismatch         RefusalCode = "EVIDENCE_SCOPE_MISMATCH"
	RefusalUnitInvalid           RefusalCode = "EVIDENCE_UNIT_INVALID"
	RefusalConfigMismatch        RefusalCode = "EVIDENCE_CONFIG_MISMATCH"
	RefusalTimestampInvalid      RefusalCode = "EVIDENCE_TIMESTAMP_INVALID"
	RefusalEvidenceStale         RefusalCode = "EVIDENCE_STALE"
	RefusalArithmeticInvalid     RefusalCode = "ARITHMETIC_INVALID"
	RefusalThresholdNotMet       RefusalCode = "REVERSAL_THRESHOLD_NOT_MET"
	RefusalStructuralMissing     RefusalCode = "STRUCTURAL_EVIDENCE_MISSING"
	RefusalStructuralOrder       RefusalCode = "STRUCTURAL_ORDER_INVALID"
	RefusalStructuralStale       RefusalCode = "STRUCTURAL_EVIDENCE_STALE"
	RefusalPlanInvalid           RefusalCode = "CAMPAIGN_PLAN_INVALID"
	RefusalLegTerminal           RefusalCode = "CAMPAIGN_LEG_TERMINAL"
	RefusalCapInvalid            RefusalCode = "A066_CAP_INVALID"
	RefusalRiskBudgetExceeded    RefusalCode = "CAMPAIGN_RISK_BUDGET_EXCEEDED"
	RefusalRiskLatched           RefusalCode = "CAMPAIGN_RISK_LATCHED"
	RefusalStopRetreat           RefusalCode = "EFFECTIVE_STOP_RETREAT"
	RefusalExecutionTermsInvalid RefusalCode = "EXECUTION_TERMS_INVALID"
	RefusalInvalidated           RefusalCode = "REVERSAL_INVALIDATED"
)

var ErrStrictSchema = errors.New("reversal lane: strict schema invalid")

type EvidenceUnits struct {
	Notional string `json:"notional"`
	Price    string `json:"price"`
	Volume   string `json:"volume"`
}

type CommonEnvelope struct {
	SchemaVersion      string        `json:"schema_version"`
	Market             Market        `json:"market"`
	AccountRef         string        `json:"account_ref"`
	Symbol             string        `json:"symbol"`
	PositionGeneration uint64        `json:"position_generation"`
	SourceRecordID     string        `json:"source_record_id"`
	SourceDigest       string        `json:"source_digest"`
	Units              EvidenceUnits `json:"units"`
	EffectiveAt        time.Time     `json:"effective_at"`
	ObservedAt         time.Time     `json:"observed_at"`
	IngestedAt         time.Time     `json:"ingested_at"`
	EvaluatedAt        time.Time     `json:"evaluated_at"`
	FreshUntil         time.Time     `json:"fresh_until"`
	ThresholdSet       string        `json:"threshold_set"`
	StructuralWindowNS int64         `json:"structural_window_ns"`
	ConfigDigest       string        `json:"config_digest"`
}

type KREvidence struct {
	CommonEnvelope
	AbsorbedNotionalMinor       uint64 `json:"absorbed_notional_minor"`
	AggressiveSellNotionalMinor uint64 `json:"aggressive_sell_notional_minor"`
	AbsorptionPPM               uint64 `json:"absorption_ppm"`
}

type USEvidence struct {
	CommonEnvelope
	ReferencePriceMinor      uint64 `json:"reference_price_minor"`
	DislocationLowPriceMinor uint64 `json:"dislocation_low_price_minor"`
	DislocationVolumeShares  uint64 `json:"dislocation_volume_shares"`
	BaselineVolumeShares     uint64 `json:"baseline_volume_shares"`
	DrawdownPPM              uint64 `json:"drawdown_ppm"`
	RelativeVolumePPM        uint64 `json:"relative_volume_ppm"`
}

type KRConfig struct {
	Version              string
	SchemaVersion        string
	ConfigDigest         string
	ThresholdSet         string
	MinimumAbsorptionPPM uint64
	StructuralWindow     time.Duration
}

type USConfig struct {
	Version                  string
	SchemaVersion            string
	ConfigDigest             string
	ThresholdSet             string
	MinimumDrawdownPPM       uint64
	MinimumRelativeVolumePPM uint64
	StructuralWindow         time.Duration
}

type MetricResult struct {
	Accepted          bool
	Refusal           RefusalCode
	AbsorptionPPM     uint64
	DrawdownPPM       uint64
	RelativeVolumePPM uint64
}

type StructuralEventKind string

const (
	EventSweep   StructuralEventKind = "sweep"
	EventBreak   StructuralEventKind = "break"
	EventReclaim StructuralEventKind = "reclaim"
)

type StructuralEvent struct {
	Kind               StructuralEventKind
	AccountRef         string
	Market             Market
	Symbol             string
	PositionGeneration uint64
	EvidenceVersion    string
	RecordID           string
	Digest             string
	At                 time.Time
	FreshUntil         time.Time
}

type StructuralConfirmation struct {
	Sweep   StructuralEvent
	Break   StructuralEvent
	Reclaim StructuralEvent
}

type Invalidation struct {
	Structural bool
	Code       string
}

type StopCandidate struct {
	priceMinor, source, policy, version, digest, planDigest, evidenceDigest string
	observedAt, freshUntil                                                  time.Time
	seal                                                                    [32]byte
}

type stopCandidateInput struct {
	PriceMinor, Source, Policy, Version, Digest string
	ObservedAt, FreshUntil                      time.Time
}

type LegProgress struct {
	Ordinal        int
	FilledQuantity uint64
	Cancelled      bool
	Expired        bool
}

type RiskCap struct {
	Authority           string
	Version             string
	PlanDigest          string
	Market              Market
	QFinal              uint64
	ReservationQuantity uint64
	ReservationMinor    string
	SnapshotID          string
	PolicyDigest        string
	BucketSetDigest     string
	Official            bool
	Frozen              bool
	ObservedAt          time.Time
	FreshUntil          time.Time
	seal                string
}

type EvaluationContext struct {
	Enabled                 bool
	CandidateID             string
	Plan                    CampaignPlan
	Leg                     LegProgress
	Cap                     RiskCap
	Risk                    RiskState
	SavedEffectiveStopMinor string
	StopCandidate           StopCandidate
	ExecutionTerms          ExecutionTermsPreimage
	Invalidation            Invalidation
	PriceDeclined           bool
}

type KREvaluationRequest struct {
	Context   EvaluationContext
	Evidence  KREvidence
	Config    KRConfig
	Structure StructuralConfirmation
}

type USEvaluationRequest struct {
	Context   EvaluationContext
	Evidence  USEvidence
	Config    USConfig
	Structure StructuralConfirmation
}

type OutcomeKind string

const (
	OutcomeDecision     OutcomeKind = "DECISION"
	OutcomeRefusal      OutcomeKind = "REFUSAL"
	OutcomeInvalidation OutcomeKind = "INVALIDATION"
)

type DecisionLineage struct {
	Market               Market
	LaneID               string
	LaneVersion          string
	CandidateID          string
	SchemaVersion        string
	ConfigDigest         string
	MetricEvidenceDigest string
	StructuralDigest     string
	CampaignID           string
	PositionGeneration   uint64
	RiskBudgetDigest     string
	LegOrdinal           int
	PlannedCeiling       uint64
	CapSnapshotID        string
	CapPolicyDigest      string
}

type EvaluationResult struct {
	Kind                  OutcomeKind
	Code                  RefusalCode
	Action                string
	Quantity              uint64
	EntryPriceMinor       string
	EffectiveStopMinor    string
	TargetPriceMinor      string
	EntryProvenance       PriceProvenance
	StopProvenance        PriceProvenance
	TargetProvenance      PriceProvenance
	ExecutionPolicyDigest string
	Lineage               DecisionLineage
	CommonExitIndependent bool
	ExitDecisionCreated   bool
}

type CommonExitProbe struct{ StopTriggered bool }

func (p CommonExitProbe) CanProceed(_ EvaluationResult) bool { return p.StopTriggered }

type FXLaneDirection string

const FXQuoteToAccount FXLaneDirection = "QUOTE_TO_ACCOUNT"

type FrozenFX struct {
	Authority          string
	Version            string
	QuoteID            string
	AsOf               string
	Direction          FXLaneDirection
	RateQuoteToAccount string
	Haircut            string
	Digest             string
	Official           bool
	Frozen             bool
	FreshUntil         string
	seal               string
}

type Latch string

const (
	LatchCampaignRiskOverage Latch = "CAMPAIGN_RISK_OVERAGE"
	LatchUnknownActualRisk   Latch = "UNKNOWN_ACTUAL_RISK"
)

type AppliedFill struct {
	Applied     bool
	RiskMinor   string
	Fingerprint string
}

type RiskState struct {
	PlanDigest  string
	FilledMinor string
	HeldMinor   string
	Fills       map[string]AppliedFill
	Cancels     map[string]string
	Latches     map[Latch]bool
}

type FillRiskEvent struct {
	FillID                       string
	CampaignID                   string
	LegOrdinal                   int
	OrderRef                     string
	Quantity                     uint64
	TransferredReservationMinor  string
	EntryPriceMinor              string
	EffectiveStopMinor           string
	EntryFeesMinor               string
	EstimatedExitFeesLeviesMinor string
	ObservedAt                   time.Time
	SourceDigest                 string
	FX                           *FrozenFX
}

type CancelRiskEvent struct {
	CancelID         string
	CampaignID       string
	LegOrdinal       int
	OrderRef         string
	ReleaseHeldMinor string
	ObservedAt       time.Time
	SourceDigest     string
}

type RiskApplyResult struct {
	Applied   bool
	Duplicate bool
}
