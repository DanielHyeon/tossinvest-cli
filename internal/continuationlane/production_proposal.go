package continuationlane

import (
	"errors"
	"strconv"
	"time"

	marketclock "github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/officialfx"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyevidence"
)

var ErrProductionProposalUnavailable = errors.New("continuation lanes: production proposal unavailable")

type ProductionStopInput struct {
	PriceMinor, Source, Policy, Version, Digest string
	ObservedAt, FreshUntil                      time.Time
}

type ProductionProposalInput struct {
	Snapshot                                       strategyevidence.Snapshot
	FX                                             officialfx.Evidence
	Market                                         Market
	AccountRef, Symbol, CandidateID, CampaignID    string
	PositionGeneration                             int64
	RiskBudgetMinor, PerShareRiskMinor             string
	PlannedQuantity                                uint64
	PolicyDigest, ConfigDigest                     string
	AccountCurrency, QuoteCurrency                 string
	ThresholdSetID                                 string
	MinimumFlowPressurePPM                         int64
	MinimumParticipationPPM, MinimumPriceChangePPM int64
	Leg                                            LegProgress
	SavedEffectiveStopMinor                        string
	Stop                                           ProductionStopInput
	EntryPriceMinor, TargetPriceMinor              string
	FreshUntil                                     time.Time
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
	NetFlowNotionalMinor  string `json:"net_flow_notional_minor"`
	TurnoverNotionalMinor string `json:"turnover_notional_minor"`
	FlowPressurePPM       string `json:"flow_pressure_ppm"`
}

type productionUSMetrics struct {
	ParticipatingVolumeShares string `json:"participating_volume_shares"`
	BaselineVolumeShares      string `json:"baseline_volume_shares"`
	ReferencePriceMinor       string `json:"reference_price_minor"`
	LastPriceMinor            string `json:"last_price_minor"`
	ParticipationPPM          string `json:"participation_ppm"`
	PriceChangePPM            string `json:"price_change_ppm"`
}

func BuildProductionKRProposalAuthority(input ProductionProposalInput) (ProductionKRProposalAuthority, error) {
	if input.Market != MarketKR {
		return ProductionKRProposalAuthority{}, ErrProductionProposalUnavailable
	}
	header, payload, err := productionEvidence(input, strategyevidence.KindKRNetFlow, KRFlowSchemaV1)
	if err != nil || (header.Authority != strategyevidence.AuthorityKRX && header.Authority != strategyevidence.AuthorityTossOpenAPI) {
		return ProductionKRProposalAuthority{}, ErrProductionProposalUnavailable
	}
	var metrics productionKRMetrics
	if err := decodeStrict(payload, &metrics); err != nil {
		return ProductionKRProposalAuthority{}, ErrProductionProposalUnavailable
	}
	flow, ok := productionPPM(metrics.FlowPressurePPM)
	if !ok {
		return ProductionKRProposalAuthority{}, ErrProductionProposalUnavailable
	}
	config, err := NewKRFlowConfig(input.ThresholdSetID, input.ConfigDigest, input.MinimumFlowPressurePPM)
	if err != nil {
		return ProductionKRProposalAuthority{}, ErrProductionProposalUnavailable
	}
	envelope := productionEnvelope(input, header, KRFlowSchemaV1)
	evidence := KREvidence{Envelope: envelope, NotionalUnit: UnitNotionalMinor, NetFlowNotionalMinor: metrics.NetFlowNotionalMinor,
		TurnoverNotionalMinor: metrics.TurnoverNotionalMinor, FlowPressurePPM: flow}
	context, err := productionContext(input, envelope)
	if err != nil {
		return ProductionKRProposalAuthority{}, err
	}
	return ProductionKRProposalAuthority{request: KREvaluationRequest{Context: context, Evidence: evidence, Config: config},
		snapshotID: input.Snapshot.ID, snapshotDigest: input.Snapshot.Digest}, nil
}

func BuildProductionUSProposalAuthority(input ProductionProposalInput) (ProductionUSProposalAuthority, error) {
	if input.Market != MarketUS {
		return ProductionUSProposalAuthority{}, ErrProductionProposalUnavailable
	}
	header, payload, err := productionEvidence(input, strategyevidence.KindUSParticipation, USParticipationSchemaV1)
	if err != nil || header.Authority != strategyevidence.AuthorityTossOpenAPI {
		return ProductionUSProposalAuthority{}, ErrProductionProposalUnavailable
	}
	var metrics productionUSMetrics
	if err := decodeStrict(payload, &metrics); err != nil {
		return ProductionUSProposalAuthority{}, ErrProductionProposalUnavailable
	}
	participation, participationOK := productionPPM(metrics.ParticipationPPM)
	priceChange, priceOK := productionPPM(metrics.PriceChangePPM)
	if !participationOK || !priceOK {
		return ProductionUSProposalAuthority{}, ErrProductionProposalUnavailable
	}
	config, err := NewUSParticipationConfig(input.ThresholdSetID, input.ConfigDigest, input.MinimumParticipationPPM, input.MinimumPriceChangePPM)
	if err != nil {
		return ProductionUSProposalAuthority{}, ErrProductionProposalUnavailable
	}
	envelope := productionEnvelope(input, header, USParticipationSchemaV1)
	evidence := USEvidence{Envelope: envelope, VolumeUnit: UnitShares, PriceUnit: UnitQuoteMinor,
		ParticipatingVolumeShares: metrics.ParticipatingVolumeShares, BaselineVolumeShares: metrics.BaselineVolumeShares,
		ReferencePriceMinor: metrics.ReferencePriceMinor, LastPriceMinor: metrics.LastPriceMinor,
		ParticipationPPM: participation, PriceChangePPM: priceChange}
	context, err := productionContext(input, envelope)
	if err != nil {
		return ProductionUSProposalAuthority{}, err
	}
	return ProductionUSProposalAuthority{request: USEvaluationRequest{Context: context, Evidence: evidence, Config: config},
		snapshotID: input.Snapshot.ID, snapshotDigest: input.Snapshot.Digest}, nil
}

func productionPPM(raw string) (int64, bool) {
	value, err := strconv.ParseInt(raw, 10, 64)
	return value, err == nil && strconv.FormatInt(value, 10) == raw && value >= -1_000_000 && value <= 1_000_000
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
		header := item.Header()
		if header.Kind != kind {
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

func productionEnvelope(input ProductionProposalInput, header strategyevidence.Header, schema string) EvidenceEnvelope {
	return EvidenceEnvelope{SchemaVersion: schema, Market: input.Market, Symbol: input.Symbol, SourceRecordID: header.SourceRecordID,
		SourceDigest: input.Snapshot.Digest, EffectiveAt: header.SourceEventAt.UTC().Format(time.RFC3339Nano),
		ObservedAt: header.ObservedAt.UTC().Format(time.RFC3339Nano), IngestedAt: header.IngestedAt.UTC().Format(time.RFC3339Nano),
		EvaluatedAt: input.Snapshot.EvaluationAt.UTC().Format(time.RFC3339Nano), FreshUntil: input.FreshUntil.UTC().Format(time.RFC3339Nano),
		ThresholdSetID: input.ThresholdSetID, ConfigDigest: input.ConfigDigest}
}

func productionContext(input ProductionProposalInput, envelope EvidenceEnvelope) (EvaluationContext, error) {
	if input.Leg.Ordinal != 1 || input.Leg.FilledQuantity != 0 || input.Leg.Cancelled || input.Leg.Expired || input.PlannedQuantity == 0 ||
		input.PlannedQuantity > 1_000_000_000 || input.Symbol != input.Snapshot.Symbol || input.PositionGeneration <= 0 {
		return EvaluationContext{}, ErrProductionProposalUnavailable
	}
	reserve, err := input.FX.EvidenceAt(input.Snapshot.EvaluationAt, input.QuoteCurrency, input.AccountCurrency)
	if err != nil {
		return EvaluationContext{}, ErrProductionProposalUnavailable
	}
	var frozen *FrozenFX
	if input.AccountCurrency != input.QuoteCurrency {
		value, err := newFrozenFX(frozenFXInput{QuoteID: reserve.Digest(), AsOf: reserve.ObservedAt().UTC().Format(time.RFC3339Nano),
			Direction: FXQuoteToAccount, RateQuoteToAccount: reserve.RateQuoteToBase(), Haircut: reserve.Haircut(), Digest: reserve.Digest()})
		if err != nil {
			return EvaluationContext{}, ErrProductionProposalUnavailable
		}
		frozen = &value
	}
	plan, err := BuildCampaignPlan(PlanRequest{LaneID: map[Market]string{MarketKR: KRContinuationLaneID, MarketUS: USContinuationLaneID}[input.Market],
		LaneVersion: LaneVersionV1, Market: input.Market, AccountRef: input.AccountRef, Symbol: input.Symbol, CampaignID: input.CampaignID,
		PositionGeneration: input.PositionGeneration, RiskBudgetMinor: input.RiskBudgetMinor, PerShareRiskMinor: input.PerShareRiskMinor,
		PlannedQuantity: input.PlannedQuantity, PolicyDigest: input.PolicyDigest, ConfigDigest: input.ConfigDigest,
		AccountCurrency: input.AccountCurrency, QuoteCurrency: input.QuoteCurrency, FX: frozen})
	if err != nil {
		return EvaluationContext{}, ErrProductionProposalUnavailable
	}
	stop, err := newStopCandidate(stopCandidateInput{PriceMinor: input.Stop.PriceMinor, Source: input.Stop.Source, Policy: input.Stop.Policy,
		Version: input.Stop.Version, Digest: input.Stop.Digest, ObservedAt: input.Stop.ObservedAt.UTC().Format(time.RFC3339Nano),
		FreshUntil: input.Stop.FreshUntil.UTC().Format(time.RFC3339Nano)})
	if err != nil {
		return EvaluationContext{}, ErrProductionProposalUnavailable
	}
	terms, err := mintExecutionTermsPreimage(plan, envelope, input.EntryPriceMinor, input.TargetPriceMinor)
	if err != nil {
		return EvaluationContext{}, ErrProductionProposalUnavailable
	}
	context := EvaluationContext{Enabled: true, CandidateID: input.CandidateID, Plan: plan, Leg: input.Leg, Risk: NewRiskState(plan),
		SavedEffectiveStopMinor: input.SavedEffectiveStopMinor, StopCandidate: stop, ExecutionTerms: terms}
	context.savedStopAuthority = mintSavedStopProvenance(plan, envelope, input.SavedEffectiveStopMinor)
	return context, nil
}
