// Package continuationlane implements pure, dormant KR and US continuation
// evaluators. It has no persistence, broker, exit, or runtime-toggle authority.
package continuationlane

type Market string

const (
	MarketKR Market = "KR"
	MarketUS Market = "US"
)

const (
	KRContinuationLaneID    = "kr_short_flow_continuation_v1"
	USContinuationLaneID    = "us_short_participation_continuation_v1"
	LaneVersionV1           = "v1"
	ContinuationRelease     = "a067-kr-us-continuation-v1"
	KRFlowSchemaV1          = "kr-flow-v1"
	USParticipationSchemaV1 = "us-participation-v1"
)

type RefusalCode string

const (
	RefusalNone                RefusalCode = ""
	RefusalLaneOff             RefusalCode = "LANE_OFF"
	RefusalMarketMismatch      RefusalCode = "MARKET_MISMATCH"
	RefusalSchemaMismatch      RefusalCode = "SCHEMA_MISMATCH"
	RefusalEvidenceInvalid     RefusalCode = "EVIDENCE_INVALID"
	RefusalInvalidTime         RefusalCode = "INVALID_TIME_ORDER"
	RefusalStaleEvidence       RefusalCode = "STALE_EVIDENCE"
	RefusalUnitMismatch        RefusalCode = "UNIT_MISMATCH"
	RefusalDigestMismatch      RefusalCode = "DIGEST_MISMATCH"
	RefusalConfigInvalid       RefusalCode = "CONFIG_INVALID"
	RefusalArithmeticMismatch  RefusalCode = "ARITHMETIC_PREIMAGE_MISMATCH"
	RefusalArithmeticOverflow  RefusalCode = "ARITHMETIC_OVERFLOW"
	RefusalArithmeticInvalid   RefusalCode = "ARITHMETIC_INVALID"
	RefusalThresholdNotMet     RefusalCode = "THRESHOLD_NOT_MET"
	RefusalPlanInvalid         RefusalCode = "IMMUTABLE_PLAN_INVALID"
	RefusalCapInvalid          RefusalCode = "A066_CAP_INVALID"
	RefusalRiskBudgetExceeded  RefusalCode = "CAMPAIGN_RISK_BUDGET_EXCEEDED"
	RefusalRiskLatched         RefusalCode = "CAMPAIGN_RISK_LATCHED"
	RefusalLegTerminal         RefusalCode = "LEG_TERMINAL"
	RefusalQuantityUnavailable RefusalCode = "QUANTITY_UNAVAILABLE"
	RefusalStopInvalid         RefusalCode = "STOP_INVALID"
	RefusalInvalidationInvalid RefusalCode = "INVALIDATION_INVALID"
)

type OutcomeKind string

const (
	OutcomeDecision     OutcomeKind = "ENTRY_ADD_DECISION"
	OutcomeRefusal      OutcomeKind = "REFUSAL"
	OutcomeInvalidation OutcomeKind = "INVALIDATION"
)

type Latch string

const (
	LatchCampaignRiskOverage Latch = "CAMPAIGN_RISK_OVERAGE"
	LatchUnknownActualRisk   Latch = "UNKNOWN_ACTUAL_RISK"
)
