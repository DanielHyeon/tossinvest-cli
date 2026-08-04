package reversallane

import (
	"errors"
	"strconv"
	"time"

	marketclock "github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/officialfx"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyevidence"
)

var ErrProductionProposalUnavailable = errors.New("reversal lane: production proposal unavailable")

type ProductionStopInput struct {
	PriceMinor, Source, Policy, Version, Digest string
	ObservedAt, FreshUntil                      time.Time
}

type ProductionProposalInput struct {
	Snapshot                                     strategyevidence.Snapshot
	FX                                           officialfx.Evidence
	Market                                       Market
	AccountRef, Symbol, CandidateID, CampaignID  string
	PositionGeneration                           uint64
	RiskBudgetMinor, PerShareRiskMinor           string
	PlannedQuantity                              uint64
	PolicyDigest, ConfigDigest                   string
	AccountCurrency, QuoteCurrency               string
	ConfigVersion, ThresholdSet                  string
	MinimumAbsorptionPPM                         uint64
	MinimumDrawdownPPM, MinimumRelativeVolumePPM uint64
	StructuralWindow                             time.Duration
	Leg                                          LegProgress
	SavedEffectiveStopMinor                      string
	Stop                                         ProductionStopInput
	Structure                                    StructuralConfirmation
	EntryPriceMinor, TargetPriceMinor            string
	FreshUntil                                   time.Time
}

type ProductionKRProposalAuthority struct {
	request                    KREvaluationRequest
	snapshotID, snapshotDigest string
}

func (authority ProductionKRProposalAuthority) Request() KREvaluationRequest {
	return authority.request
}
func (authority ProductionKRProposalAuthority) SnapshotID() string { return authority.snapshotID }
func (authority ProductionKRProposalAuthority) SnapshotDigest() string {
	return authority.snapshotDigest
}

type ProductionUSProposalAuthority struct {
	request                    USEvaluationRequest
	snapshotID, snapshotDigest string
}

func (authority ProductionUSProposalAuthority) Request() USEvaluationRequest {
	return authority.request
}
func (authority ProductionUSProposalAuthority) SnapshotID() string { return authority.snapshotID }
func (authority ProductionUSProposalAuthority) SnapshotDigest() string {
	return authority.snapshotDigest
}

type productionKRMetrics struct {
	AbsorbedNotionalMinor       string `json:"absorbed_notional_minor"`
	AggressiveSellNotionalMinor string `json:"aggressive_sell_notional_minor"`
	AbsorptionPPM               string `json:"absorption_ppm"`
}

type productionUSMetrics struct {
	ReferencePriceMinor      string `json:"reference_price_minor"`
	DislocationLowPriceMinor string `json:"dislocation_low_price_minor"`
	DislocationVolumeShares  string `json:"dislocation_volume_shares"`
	BaselineVolumeShares     string `json:"baseline_volume_shares"`
	DrawdownPPM              string `json:"drawdown_ppm"`
	RelativeVolumePPM        string `json:"relative_volume_ppm"`
}

func BuildProductionKRProposalAuthority(input ProductionProposalInput) (ProductionKRProposalAuthority, error) {
	if input.Market != MarketKR {
		return ProductionKRProposalAuthority{}, ErrProductionProposalUnavailable
	}
	header, payload, err := productionEvidence(input, strategyevidence.KindKRNetFlow, "kr-absorption-v1")
	if err != nil || (header.Authority != strategyevidence.AuthorityKRX && header.Authority != strategyevidence.AuthorityTossOpenAPI) {
		return ProductionKRProposalAuthority{}, ErrProductionProposalUnavailable
	}
	var metrics productionKRMetrics
	if err := strictDecode(payload, &metrics); err != nil {
		return ProductionKRProposalAuthority{}, ErrProductionProposalUnavailable
	}
	absorption, ok := productionUint(metrics.AbsorptionPPM)
	absorbed, absorbedOK := productionUint(metrics.AbsorbedNotionalMinor)
	aggressive, aggressiveOK := productionUint(metrics.AggressiveSellNotionalMinor)
	if !ok || !absorbedOK || !aggressiveOK {
		return ProductionKRProposalAuthority{}, ErrProductionProposalUnavailable
	}
	envelope := productionEnvelope(input, header, "kr-absorption-v1")
	evidence := KREvidence{CommonEnvelope: envelope, AbsorbedNotionalMinor: absorbed, AggressiveSellNotionalMinor: aggressive, AbsorptionPPM: absorption}
	config := KRConfig{Version: input.ConfigVersion, SchemaVersion: envelope.SchemaVersion, ConfigDigest: input.ConfigDigest,
		ThresholdSet: input.ThresholdSet, MinimumAbsorptionPPM: input.MinimumAbsorptionPPM, StructuralWindow: input.StructuralWindow}
	context, err := productionContext(input, envelope)
	if err != nil {
		return ProductionKRProposalAuthority{}, err
	}
	return ProductionKRProposalAuthority{request: KREvaluationRequest{Context: context, Evidence: evidence, Config: config, Structure: input.Structure},
		snapshotID: input.Snapshot.ID, snapshotDigest: input.Snapshot.Digest}, nil
}

func BuildProductionUSProposalAuthority(input ProductionProposalInput) (ProductionUSProposalAuthority, error) {
	if input.Market != MarketUS {
		return ProductionUSProposalAuthority{}, ErrProductionProposalUnavailable
	}
	header, payload, err := productionEvidence(input, strategyevidence.KindUSParticipation, "us-dislocation-v1")
	if err != nil || header.Authority != strategyevidence.AuthorityTossOpenAPI {
		return ProductionUSProposalAuthority{}, ErrProductionProposalUnavailable
	}
	var metrics productionUSMetrics
	if err := strictDecode(payload, &metrics); err != nil {
		return ProductionUSProposalAuthority{}, ErrProductionProposalUnavailable
	}
	reference, referenceOK := productionUint(metrics.ReferencePriceMinor)
	low, lowOK := productionUint(metrics.DislocationLowPriceMinor)
	volume, volumeOK := productionUint(metrics.DislocationVolumeShares)
	baseline, baselineOK := productionUint(metrics.BaselineVolumeShares)
	drawdown, drawdownOK := productionUint(metrics.DrawdownPPM)
	relative, relativeOK := productionUint(metrics.RelativeVolumePPM)
	if !referenceOK || !lowOK || !volumeOK || !baselineOK || !drawdownOK || !relativeOK {
		return ProductionUSProposalAuthority{}, ErrProductionProposalUnavailable
	}
	envelope := productionEnvelope(input, header, "us-dislocation-v1")
	evidence := USEvidence{CommonEnvelope: envelope, ReferencePriceMinor: reference, DislocationLowPriceMinor: low,
		DislocationVolumeShares: volume, BaselineVolumeShares: baseline, DrawdownPPM: drawdown, RelativeVolumePPM: relative}
	config := USConfig{Version: input.ConfigVersion, SchemaVersion: envelope.SchemaVersion, ConfigDigest: input.ConfigDigest,
		ThresholdSet: input.ThresholdSet, MinimumDrawdownPPM: input.MinimumDrawdownPPM,
		MinimumRelativeVolumePPM: input.MinimumRelativeVolumePPM, StructuralWindow: input.StructuralWindow}
	context, err := productionContext(input, envelope)
	if err != nil {
		return ProductionUSProposalAuthority{}, err
	}
	return ProductionUSProposalAuthority{request: USEvaluationRequest{Context: context, Evidence: evidence, Config: config, Structure: input.Structure},
		snapshotID: input.Snapshot.ID, snapshotDigest: input.Snapshot.Digest}, nil
}

func productionEvidence(input ProductionProposalInput, kind strategyevidence.EvidenceKind, schema string) (strategyevidence.Header, []byte, error) {
	clockMarket := map[Market]marketclock.Market{MarketKR: marketclock.MarketKR, MarketUS: marketclock.MarketUS}[input.Market]
	if !input.Snapshot.Valid() || clockMarket == "" || input.Snapshot.Market != clockMarket || input.Snapshot.Symbol != input.Symbol ||
		input.Snapshot.EvaluationAt.IsZero() || input.FreshUntil.Before(input.Snapshot.EvaluationAt) ||
		input.FreshUntil.Sub(input.Snapshot.EvaluationAt) > 5*time.Minute {
		return strategyevidence.Header{}, nil, ErrProductionProposalUnavailable
	}
	var selected strategyevidence.Envelope
	found := false
	for _, item := range input.Snapshot.Items {
		if item.Header().Kind != kind {
			continue
		}
		if found {
			return strategyevidence.Header{}, nil, ErrProductionProposalUnavailable
		}
		selected, found = item, true
	}
	if !found {
		return strategyevidence.Header{}, nil, ErrProductionProposalUnavailable
	}
	header := selected.Header()
	if header.SchemaVersion != schema || header.Market != input.Snapshot.Market || header.Symbol != input.Symbol ||
		header.Availability != strategyevidence.AvailabilityAvailable || header.Confidence != strategyevidence.ConfidenceVerified ||
		header.Currency != input.QuoteCurrency {
		return strategyevidence.Header{}, nil, ErrProductionProposalUnavailable
	}
	return header, selected.CanonicalPayload(), nil
}

func productionEnvelope(input ProductionProposalInput, header strategyevidence.Header, schema string) CommonEnvelope {
	units := EvidenceUnits{}
	if input.Market == MarketKR {
		units.Notional = "KRW_MINOR"
	} else {
		units.Price, units.Volume = "USD_MINOR", "SHARES"
	}
	return CommonEnvelope{SchemaVersion: schema, Market: input.Market, AccountRef: input.AccountRef, Symbol: input.Symbol,
		PositionGeneration: input.PositionGeneration, SourceRecordID: header.SourceRecordID, SourceDigest: input.Snapshot.Digest,
		Units: units, EffectiveAt: header.SourceEventAt.UTC(), ObservedAt: header.ObservedAt.UTC(), IngestedAt: header.IngestedAt.UTC(),
		EvaluatedAt: input.Snapshot.EvaluationAt.UTC(), FreshUntil: input.FreshUntil.UTC(), ThresholdSet: input.ThresholdSet,
		StructuralWindowNS: int64(input.StructuralWindow), ConfigDigest: input.ConfigDigest}
}

func productionContext(input ProductionProposalInput, envelope CommonEnvelope) (EvaluationContext, error) {
	if input.Leg.Ordinal < 1 || input.Leg.Ordinal > 3 || input.Leg.Cancelled || input.Leg.Expired ||
		input.PlannedQuantity == 0 || input.PlannedQuantity > 1_000_000_000 || input.PositionGeneration == 0 ||
		input.StructuralWindow <= 0 || input.StructuralWindow > 24*time.Hour {
		return EvaluationContext{}, ErrProductionProposalUnavailable
	}
	reserve, err := input.FX.EvidenceAt(input.Snapshot.EvaluationAt, input.QuoteCurrency, input.AccountCurrency)
	if err != nil {
		return EvaluationContext{}, ErrProductionProposalUnavailable
	}
	var frozen *FrozenFX
	if input.AccountCurrency != input.QuoteCurrency {
		value, err := mintFrozenFX(FrozenFX{Authority: a066FXAuthority, Version: reserve.Version(), QuoteID: reserve.Digest(),
			AsOf: reserve.ObservedAt().UTC().Format(time.RFC3339Nano), FreshUntil: reserve.FreshUntil().UTC().Format(time.RFC3339Nano),
			Direction: FXQuoteToAccount, RateQuoteToAccount: reserve.RateQuoteToBase(), Haircut: reserve.Haircut(), Digest: reserve.Digest(),
			Official: true, Frozen: true})
		if err != nil {
			return EvaluationContext{}, ErrProductionProposalUnavailable
		}
		frozen = &value
	}
	plan, err := BuildCampaignPlan(PlanRequest{LaneID: map[Market]string{MarketKR: KRReversalLaneID, MarketUS: USReversalLaneID}[input.Market],
		LaneVersion: LaneVersionV1, Market: input.Market, AccountRef: input.AccountRef, Symbol: input.Symbol, CampaignID: input.CampaignID,
		PositionGeneration: input.PositionGeneration, RiskBudgetMinor: input.RiskBudgetMinor, PerShareRiskMinor: input.PerShareRiskMinor,
		PlannedQuantity: input.PlannedQuantity, PolicyDigest: input.PolicyDigest, ConfigDigest: input.ConfigDigest,
		AccountCurrency: input.AccountCurrency, QuoteCurrency: input.QuoteCurrency, FX: frozen})
	if err != nil {
		return EvaluationContext{}, ErrProductionProposalUnavailable
	}
	stop, err := mintStopCandidate(plan, envelope, stopCandidateInput{PriceMinor: input.Stop.PriceMinor, Source: input.Stop.Source,
		Policy: input.Stop.Policy, Version: input.Stop.Version, Digest: input.Stop.Digest,
		ObservedAt: input.Stop.ObservedAt.UTC(), FreshUntil: input.Stop.FreshUntil.UTC()})
	if err != nil {
		return EvaluationContext{}, ErrProductionProposalUnavailable
	}
	terms, err := mintExecutionTermsPreimage(plan, envelope, input.EntryPriceMinor, input.TargetPriceMinor)
	if err != nil {
		return EvaluationContext{}, ErrProductionProposalUnavailable
	}
	return EvaluationContext{Enabled: true, CandidateID: input.CandidateID, Plan: plan, Leg: input.Leg, Risk: NewRiskState(plan),
		SavedEffectiveStopMinor: input.SavedEffectiveStopMinor, StopCandidate: stop, ExecutionTerms: terms}, nil
}

func productionUint(raw string) (uint64, bool) {
	value, err := strconv.ParseUint(raw, 10, 64)
	return value, err == nil && value > 0 && strconv.FormatUint(value, 10) == raw
}
