// Package breakoutlane evaluates sealed, fixture-only breakout snapshots.
// It has no official-source, broker, journal, toggle, network, or clock authority.
package breakoutlane

import "errors"

type Market string

const (
	MarketKR      Market = "KR"
	MarketUS      Market = "US"
	KRLaneID             = "kr_short_breakout_retest_v1"
	USLaneID             = "us_short_breakout_retest_v1"
	LaneVersionV1        = "v1"
)

const (
	v1PPMScale              = 1_000_000
	v1OpeningRangeBars      = 15
	v1OpeningRangeMinutes   = 15
	v1BreakoutBufferPPM     = 100_000
	v1RetestToleranceMinPPM = 100_000
	v1RetestToleranceMaxPPM = 250_000
	v1TimeoutKRClosedBars   = 8
	v1TimeoutUSClosedBars   = 10
	v1RVOLMinPPM            = 1_500_000
	v1UpperWickRangeMaxPPM  = 350_000
)

type phase string

const (
	phaseDiscovered     phase = "DISCOVERED"
	phaseRangeLocked    phase = "RANGE_LOCKED"
	phaseBreakoutClosed phase = "BREAKOUT_CLOSED"
	phaseRetestWait     phase = "RETEST_WAIT"
	phaseReclaimed      phase = "RECLAIMED"
	phaseArmed          phase = "ARMED"
	phaseProposed       phase = "PROPOSED"
	phaseInvalidated    phase = "INVALIDATED"
	phaseTimedOut       phase = "TIMED_OUT"
	phaseConsumed       phase = "CONSUMED"
)

type Descriptor struct {
	Market                                                Market
	LaneID, Version, Horizon, Desired, Effective, Runtime string
}

func Descriptors() [2]Descriptor {
	return [2]Descriptor{
		{MarketKR, KRLaneID, LaneVersionV1, "SHORT", "OFF", "OFF", "UNOBSERVED"},
		{MarketUS, USLaneID, LaneVersionV1, "SHORT", "OFF", "OFF", "UNOBSERVED"},
	}
}

type RefusalCode string

const (
	RefusalNone                RefusalCode = ""
	RefusalEvidenceInvalid     RefusalCode = "EVIDENCE_INVALID"
	RefusalLineageSealMismatch RefusalCode = "LINEAGE_SEAL_MISMATCH"
	RefusalFirstTouch          RefusalCode = "FIRST_TOUCH_FORBIDDEN"
	RefusalQuoteStale          RefusalCode = "QUOTE_STALE"
	RefusalSpreadTooWide       RefusalCode = "SPREAD_TOO_WIDE"
	RefusalEntryDriftExceeded  RefusalCode = "ENTRY_DRIFT_EXCEEDED"
	RefusalFXMissing           RefusalCode = "FX_MISSING"
	RefusalFXStale             RefusalCode = "FX_STALE"
	RefusalFXCurrencyMismatch  RefusalCode = "FX_CURRENCY_MISMATCH"
	RefusalFXInvalidRate       RefusalCode = "FX_INVALID_RATE"
	RefusalSizingOverflow      RefusalCode = "SIZING_OVERFLOW"
	RefusalNonProtectiveStop   RefusalCode = "NON_PROTECTIVE_STOP"
	RefusalNonProtectiveTarget RefusalCode = "NON_PROTECTIVE_TARGET"
	RefusalZeroQuantity        RefusalCode = "ZERO_QUANTITY"
)

type Diagnostic string

const DiagnosticCorrectionAfterProposal Diagnostic = "CORRECTION_AFTER_PROPOSAL"

type V1ConfigInput struct {
	Version, Digest                                                                                                                                                              string
	TickMinor, OpeningRangeMinutes, BreakoutBufferPPM, RetestTolerancePPM, TimeoutKR, TimeoutUS, RVOLMinPPM, UpperWickRangeMaxPPM, MaxQuoteAgeMS, MaxSpreadPPM, MaxEntryDriftPPM uint64
}
type V1Config struct{ value V1ConfigInput }

func NewV1Config(input V1ConfigInput) (V1Config, error) {
	c := V1Config{input}
	if !c.Valid() {
		return V1Config{}, errors.New("invalid v1 config")
	}
	return c, nil
}
func V1ConfigDigest(input V1ConfigInput) string {
	return "sha256:" + hashFields("tossos.breakout.config.v1", input.Version, u(input.TickMinor), u(input.OpeningRangeMinutes), u(input.BreakoutBufferPPM), u(input.RetestTolerancePPM), u(input.TimeoutKR), u(input.TimeoutUS), u(input.RVOLMinPPM), u(input.UpperWickRangeMaxPPM), u(input.MaxQuoteAgeMS), u(input.MaxSpreadPPM), u(input.MaxEntryDriftPPM))
}
func (c V1Config) Valid() bool {
	v := c.value
	return v.Version == LaneVersionV1 && canonical(v.Digest) && v.Digest == V1ConfigDigest(v) && v.TickMinor > 0 && v.OpeningRangeMinutes == v1OpeningRangeMinutes && v.BreakoutBufferPPM == v1BreakoutBufferPPM && v.RetestTolerancePPM >= v1RetestToleranceMinPPM && v.RetestTolerancePPM <= v1RetestToleranceMaxPPM && v.TimeoutKR == v1TimeoutKRClosedBars && v.TimeoutUS == v1TimeoutUSClosedBars && v.RVOLMinPPM == v1RVOLMinPPM && v.UpperWickRangeMaxPPM == v1UpperWickRangeMaxPPM && v.MaxQuoteAgeMS > 0 && v.MaxSpreadPPM > 0 && v.MaxEntryDriftPPM > 0
}
func (c V1Config) Digest() string { return c.value.Digest }

type ClosedBarInput struct {
	Sequence, Revision, IntervalMS, HighMinor, LowMinor, CloseMinor, RVOLPPM, UpperWickRangePPM uint64
	ID, SessionID                                                                               string
	RegularSession, Closed, VolumeExpanded                                                      bool
}
type ClosedBar struct{ value ClosedBarInput }

func NewClosedBar(input ClosedBarInput) (ClosedBar, error) {
	if input.Sequence == 0 || input.Revision == 0 || input.IntervalMS != 60_000 || !canonical(input.ID) || !canonical(input.SessionID) || !input.RegularSession || !input.Closed || input.HighMinor == 0 || input.LowMinor == 0 || input.CloseMinor == 0 || input.LowMinor > input.HighMinor || input.CloseMinor < input.LowMinor || input.CloseMinor > input.HighMinor {
		return ClosedBar{}, errors.New("invalid closed bar")
	}
	return ClosedBar{input}, nil
}

type QuoteSealInput struct {
	BidMinor, AskMinor, LastMinor, SourceObservedAtMS, ReceivedAtMS uint64
	Currency, Digest                                                string
}
type QuoteSeal struct{ value QuoteSealInput }

func QuoteSealDigest(input QuoteSealInput) string {
	return hashFields("quote.v1", u(input.BidMinor), u(input.AskMinor), u(input.LastMinor), input.Currency, u(input.SourceObservedAtMS), u(input.ReceivedAtMS))
}
func NewQuoteSeal(input QuoteSealInput) (QuoteSeal, error) {
	if input.BidMinor == 0 || input.AskMinor == 0 || input.LastMinor == 0 || input.BidMinor > input.AskMinor || !canonical(input.Currency) || input.Digest != QuoteSealDigest(input) {
		return QuoteSeal{}, errors.New("invalid quote seal")
	}
	return QuoteSeal{input}, nil
}

type FXDirection string

const (
	FXAccountToInstrument FXDirection = "ACCOUNT_MINOR_TO_INSTRUMENT_MINOR"
	FXInstrumentToAccount FXDirection = "INSTRUMENT_MINOR_TO_ACCOUNT_MINOR"
)

type FXSealInput struct {
	AccountCurrency, InstrumentCurrency string
	Direction                           FXDirection
	RateNum, RateDen                    uint64
	Scale                               uint32
	AsOfMS, FreshUntilMS                uint64
	Digest                              string
}
type FXSeal struct{ value FXSealInput }

func FXSealDigest(input FXSealInput) string {
	return hashFields("fx.v1", input.AccountCurrency, input.InstrumentCurrency, string(input.Direction), u(input.RateNum), u(input.RateDen), u(uint64(input.Scale)), u(input.AsOfMS), u(input.FreshUntilMS))
}
func NewFXSeal(input FXSealInput) (FXSeal, error) {
	sameCurrency := input.AccountCurrency == input.InstrumentCurrency
	if input.Direction == FXInstrumentToAccount && !sameCurrency {
		input.RateNum, input.RateDen = input.RateDen, input.RateNum
		input.Direction = FXAccountToInstrument
		input.Digest = FXSealDigest(input)
	}
	if input.Direction != FXAccountToInstrument || !canonical(input.AccountCurrency) || !canonical(input.InstrumentCurrency) || input.RateNum == 0 || input.RateDen == 0 || sameCurrency && (input.RateNum != 1 || input.RateDen != 1) || input.Scale == 0 || input.AsOfMS > input.FreshUntilMS || input.Digest != FXSealDigest(input) {
		return FXSeal{}, errors.New("invalid fx seal")
	}
	return FXSeal{input}, nil
}

type SizingInput struct{ ProposedEntryMinor, StopMinor, TargetMinor, EntrySlippageMinor, ExitSlippageMinor, RoundTripCostAccountMinor, RiskBudgetAccountMinor, NotionalCapAccountMinor, FinalCap, MinRiskRewardPPM uint64 }
type EvidenceInput struct {
	Market                             Market
	Symbol, SessionID, CalendarVersion string
	LaneID, LaneVersion                string
	Config                             V1Config
	Bars                               []ClosedBar
	ATRMinor, EvaluatedAtMS            uint64
	Quote                              QuoteSeal
	FX                                 FXSeal
	Sizing                             SizingInput
}
type EvidenceSnapshot struct {
	value  EvidenceInput
	digest string
}

func NewEvidenceSnapshot(input EvidenceInput) (EvidenceSnapshot, error) {
	if !validStructuralEvidenceInput(input) {
		return EvidenceSnapshot{}, errors.New("invalid snapshot")
	}
	copyBars := append([]ClosedBar(nil), input.Bars...)
	input.Bars = copyBars
	return EvidenceSnapshot{value: input, digest: snapshotDigest(input)}, nil
}

type Provenance struct {
	RVOLAdmission, RVOLAt1200000, RVOLAt2000000, RVOLAt2500000 bool
	ResistanceMinor, RangeLowMinor                             uint64
	Transitions                                                []string
}
type Decision struct {
	setupID, snapshotDigest, proposalID, configDigest string
	phase                                             phase
	refusal                                           RefusalCode
	diagnostic                                        Diagnostic
	candidate, final                                  uint64
	provenance                                        Provenance
	lineage                                           []barLineage
	seal                                              string
}
type barLineage struct {
	sequence, revision uint64
	id, sessionID      string
	contentDigest      string
}

func (d Decision) SetupID() string           { return d.setupID }
func (d Decision) SnapshotDigest() string    { return d.snapshotDigest }
func (d Decision) ConfigDigest() string      { return d.configDigest }
func (d Decision) ProposalID() string        { return d.proposalID }
func (d Decision) Phase() string             { return string(d.phase) }
func (d Decision) Refusal() RefusalCode      { return d.refusal }
func (d Decision) Diagnostic() Diagnostic    { return d.diagnostic }
func (d Decision) CandidateQuantity() uint64 { return d.candidate }
func (d Decision) FinalQuantity() uint64     { return d.final }
func (d Decision) Provenance() Provenance {
	p := d.provenance
	p.Transitions = append([]string(nil), p.Transitions...)
	return p
}
