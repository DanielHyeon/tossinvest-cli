// Package weeklyvaluelane provides pure, dormant KR OpenDART and US EDGAR
// weekly value evaluators. It owns no source API, persistence, activation,
// broker, or exit authority.
package weeklyvaluelane

import "time"

type Market string

const (
	MarketKR Market = "KR"
	MarketUS Market = "US"
)

type DisclosureSource string

const (
	SourceOpenDART DisclosureSource = "OPENDART"
	SourceEDGAR    DisclosureSource = "EDGAR"
)

const (
	KRWeeklyLaneID       = "kr_weekly_disclosure_value_v1"
	USWeeklyLaneID       = "us_weekly_disclosure_value_v1"
	LaneVersionV1        = "v1"
	WeeklyValueRelease   = "a069-kr-us-weekly-value-v1"
	KRDisclosureSchemaV1 = "kr-opendart-weekly-value-v1"
	USDisclosureSchemaV1 = "us-edgar-weekly-value-v1"
)

type RefusalCode string

const (
	RefusalLaneOff               RefusalCode = "LANE_OFF"
	RefusalStrictSchema          RefusalCode = "STRICT_SCHEMA_INVALID"
	RefusalSchemaInvalid         RefusalCode = "DISCLOSURE_SCHEMA_INVALID"
	RefusalMarketMismatch        RefusalCode = "MARKET_MISMATCH"
	RefusalSourceMismatch        RefusalCode = "SOURCE_MISMATCH"
	RefusalUnitInvalid           RefusalCode = "UNIT_INVALID"
	RefusalPointInTime           RefusalCode = "POINT_IN_TIME_INVALID"
	RefusalEvidenceStale         RefusalCode = "EVIDENCE_STALE"
	RefusalRevisionInvalid       RefusalCode = "REVISION_CHAIN_INVALID"
	RefusalConfigMismatch        RefusalCode = "CONFIG_MISMATCH"
	RefusalArithmeticInvalid     RefusalCode = "ARITHMETIC_INVALID"
	RefusalArithmeticMismatch    RefusalCode = "ARITHMETIC_PREIMAGE_MISMATCH"
	RefusalCalendarInvalid       RefusalCode = "OFFICIAL_CALENDAR_INVALID"
	RefusalVersionConflict       RefusalCode = "RESERVATION_VERSION_CONFLICT"
	RefusalReservationConflict   RefusalCode = "WEEK_RESERVATION_CONFLICT"
	RefusalReservationTerminal   RefusalCode = "WEEK_RESERVATION_TERMINAL"
	RefusalReservationMissing    RefusalCode = "WEEK_RESERVATION_MISSING"
	RefusalPlanInvalid           RefusalCode = "IMMUTABLE_PLAN_INVALID"
	RefusalPlanExhausted         RefusalCode = "PLAN_EXHAUSTED"
	RefusalCapInvalid            RefusalCode = "A066_CAP_INVALID"
	RefusalRiskBudgetExceeded    RefusalCode = "CAMPAIGN_RISK_BUDGET_EXCEEDED"
	RefusalRiskLatched           RefusalCode = "CAMPAIGN_RISK_LATCHED"
	RefusalLegTerminal           RefusalCode = "LEG_TERMINAL"
	RefusalRRInvalid             RefusalCode = "RR_CALCULATION_INVALID"
	RefusalRRThreshold           RefusalCode = "RR_THRESHOLD_NOT_MET"
	RefusalExecutionTermsInvalid RefusalCode = "EXECUTION_TERMS_INVALID"
	RefusalStopInvalid           RefusalCode = "STOP_INVALID"
	RefusalStructuralStopCap     RefusalCode = "STRUCTURAL_STOP_EXCEEDS_CAP"
	RefusalInvalidation          RefusalCode = "VALUE_THESIS_INVALIDATED"
)

type OutcomeKind string

const (
	OutcomeDecision     OutcomeKind = "ENTRY_ADD_DECISION"
	OutcomeRefusal      OutcomeKind = "REFUSAL"
	OutcomeInvalidation OutcomeKind = "INVALIDATION"
)

type FinancialInput struct {
	Name       string `json:"name"`
	ValueMinor string `json:"value_minor"`
	Unit       string `json:"unit"`
}

type DisclosureEvidence struct {
	SchemaVersion        string           `json:"schema_version"`
	Market               Market           `json:"market"`
	Source               DisclosureSource `json:"source"`
	Symbol               string           `json:"symbol"`
	IssuerID             string           `json:"issuer_id"`
	FilingID             string           `json:"filing_id"`
	ReportID             string           `json:"report_id"`
	RevisionID           string           `json:"revision_id"`
	SupersededRevisionID string           `json:"superseded_revision_id"`
	RevisionSequence     uint64           `json:"revision_sequence"`
	AsOf                 time.Time        `json:"as_of"`
	ObservedAt           time.Time        `json:"observed_at"`
	IngestedAt           time.Time        `json:"ingested_at"`
	CutoffAt             time.Time        `json:"cutoff_at"`
	EvaluatedAt          time.Time        `json:"evaluated_at"`
	FreshUntil           time.Time        `json:"fresh_until"`
	Currency             string           `json:"currency"`
	MonetaryUnit         string           `json:"monetary_unit"`
	MonetaryScale        uint32           `json:"monetary_scale"`
	DilutedShares        uint64           `json:"diluted_shares"`
	SharesUnit           string           `json:"shares_unit"`
	DilutionStatus       string           `json:"dilution_status"`
	DilutionFactsDigest  string           `json:"dilution_facts_digest"`
	DilutionAsOf         time.Time        `json:"dilution_as_of"`
	FinancialInputs      []FinancialInput `json:"financial_inputs"`
	ModelID              string           `json:"model_id"`
	ModelVersion         string           `json:"model_version"`
	ModelConfigDigest    string           `json:"model_config_digest"`
	ThresholdDigest      string           `json:"threshold_digest"`
	EvidenceDigest       string           `json:"evidence_digest"`
	EquityValueMinor     string           `json:"equity_value_minor"`
	FairValueMinor       string           `json:"fair_value_minor"`
	seal                 [32]byte
}

type DisclosureConfig struct {
	Market            Market
	Source            DisclosureSource
	SchemaVersion     string
	ModelVersion      string
	ModelConfigDigest string
	ThresholdDigest   string
	seal              [32]byte
}

type EvidenceResult struct {
	Accepted       bool
	Code           RefusalCode
	FairValueMinor string
	DecisionDigest string
}

type Latch string

const (
	LatchCampaignRiskOverage Latch = "CAMPAIGN_RISK_OVERAGE"
	LatchUnknownActualRisk   Latch = "UNKNOWN_ACTUAL_RISK"
)
