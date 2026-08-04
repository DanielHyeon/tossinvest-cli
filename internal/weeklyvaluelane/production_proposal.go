package weeklyvaluelane

import (
	"errors"
	"strconv"
	"time"

	marketclock "github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/officialfx"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyevidence"
)

var ErrProductionProposalUnavailable = errors.New("weekly value lane: production proposal unavailable")

type ProductionStopInput struct {
	PriceMinor, Version, Source, Policy, Digest string
	ObservedAt, FreshUntil                      time.Time
}

type ProductionReservationInput struct {
	ReservationID, CampaignID, Market           string
	StableWeek, Provider, TimeZone, SessionDate string
	CalendarGeneration, CalendarDigest          string
	Status, RequestDigest, RecordDigest         string
	PlannedOrdinal                              int
	ScopeVersion                                uint64
	PositiveLegCount                            int
	ObservedAt, FreshUntil, EvaluatedAt         time.Time
}

type ProductionProposalInput struct {
	Snapshot                                    strategyevidence.Snapshot
	FX                                          officialfx.Evidence
	Market                                      Market
	AccountRef, Symbol, CandidateID, CampaignID string
	PositionGeneration                          uint64
	RiskBudgetMinor, PerShareRiskMinor          string
	PlannedQuantity                             uint64
	PolicyDigest                                string
	AccountCurrency, QuoteCurrency              string
	ModelVersion, ModelConfigDigest             string
	ThresholdDigest                             string
	Reservation                                 ProductionReservationInput
	Leg                                         LegProgress
	Stop                                        ProductionStopInput
	SavedEffectiveStopMinor                     string
	EntryPriceMinor, StagedTargetMinor          string
	EntryCostsMinor, EstimatedExitCostsMinor    string
	MinimumRRPPM                                uint64
}

type ProductionKRProposalAuthority struct {
	request                    EvaluationRequest
	snapshotID, snapshotDigest string
}

func (authority ProductionKRProposalAuthority) Request() EvaluationRequest { return authority.request }
func (authority ProductionKRProposalAuthority) SnapshotID() string         { return authority.snapshotID }
func (authority ProductionKRProposalAuthority) SnapshotDigest() string {
	return authority.snapshotDigest
}

type ProductionUSProposalAuthority struct {
	request                    EvaluationRequest
	snapshotID, snapshotDigest string
}

func (authority ProductionUSProposalAuthority) Request() EvaluationRequest { return authority.request }
func (authority ProductionUSProposalAuthority) SnapshotID() string         { return authority.snapshotID }
func (authority ProductionUSProposalAuthority) SnapshotDigest() string {
	return authority.snapshotDigest
}

type productionDisclosurePayload struct {
	SchemaVersion        string           `json:"schema_version"`
	Market               Market           `json:"market"`
	Source               DisclosureSource `json:"source"`
	Symbol               string           `json:"symbol"`
	IssuerID             string           `json:"issuer_id"`
	FilingID             string           `json:"filing_id"`
	ReportID             string           `json:"report_id"`
	RevisionID           string           `json:"revision_id"`
	SupersededRevisionID string           `json:"superseded_revision_id"`
	RevisionSequence     string           `json:"revision_sequence"`
	AsOf                 string           `json:"as_of"`
	ObservedAt           string           `json:"observed_at"`
	IngestedAt           string           `json:"ingested_at"`
	CutoffAt             string           `json:"cutoff_at"`
	EvaluatedAt          string           `json:"evaluated_at"`
	FreshUntil           string           `json:"fresh_until"`
	Currency             string           `json:"currency"`
	MonetaryUnit         string           `json:"monetary_unit"`
	MonetaryScale        string           `json:"monetary_scale"`
	DilutedShares        string           `json:"diluted_shares"`
	SharesUnit           string           `json:"shares_unit"`
	DilutionStatus       string           `json:"dilution_status"`
	DilutionFactsDigest  string           `json:"dilution_facts_digest"`
	DilutionAsOf         string           `json:"dilution_as_of"`
	FinancialInputs      []FinancialInput `json:"financial_inputs"`
	ModelID              string           `json:"model_id"`
	ModelVersion         string           `json:"model_version"`
	ModelConfigDigest    string           `json:"model_config_digest"`
	ThresholdDigest      string           `json:"threshold_digest"`
	EquityValueMinor     string           `json:"equity_value_minor"`
	FairValueMinor       string           `json:"fair_value_minor"`
}

func BuildProductionKRProposalAuthority(input ProductionProposalInput) (ProductionKRProposalAuthority, error) {
	if input.Market != MarketKR {
		return ProductionKRProposalAuthority{}, ErrProductionProposalUnavailable
	}
	request, err := buildProductionProposal(input, SourceOpenDART, strategyevidence.AuthorityOpenDART, KRDisclosureSchemaV1)
	if err != nil {
		return ProductionKRProposalAuthority{}, err
	}
	return ProductionKRProposalAuthority{request: request, snapshotID: input.Snapshot.ID, snapshotDigest: input.Snapshot.Digest}, nil
}

func BuildProductionUSProposalAuthority(input ProductionProposalInput) (ProductionUSProposalAuthority, error) {
	if input.Market != MarketUS {
		return ProductionUSProposalAuthority{}, ErrProductionProposalUnavailable
	}
	request, err := buildProductionProposal(input, SourceEDGAR, strategyevidence.AuthoritySEC, USDisclosureSchemaV1)
	if err != nil {
		return ProductionUSProposalAuthority{}, err
	}
	return ProductionUSProposalAuthority{request: request, snapshotID: input.Snapshot.ID, snapshotDigest: input.Snapshot.Digest}, nil
}

func buildProductionProposal(input ProductionProposalInput, source DisclosureSource, authority strategyevidence.SourceAuthority, schema string) (EvaluationRequest, error) {
	clockMarket := map[Market]marketclock.Market{MarketKR: marketclock.MarketKR, MarketUS: marketclock.MarketUS}[input.Market]
	if !input.Snapshot.Valid() || input.Snapshot.Market != clockMarket || input.Snapshot.Symbol != input.Symbol || input.Snapshot.EvaluationAt.IsZero() {
		return EvaluationRequest{}, ErrProductionProposalUnavailable
	}
	var selected strategyevidence.Envelope
	found := false
	for _, item := range input.Snapshot.Items {
		if item.Header().Kind != strategyevidence.KindDisclosure {
			continue
		}
		if found {
			return EvaluationRequest{}, ErrProductionProposalUnavailable
		}
		selected, found = item, true
	}
	if !found {
		return EvaluationRequest{}, ErrProductionProposalUnavailable
	}
	header := selected.Header()
	if header.Authority != authority || header.SchemaVersion != schema || header.Market != clockMarket || header.Symbol != input.Symbol ||
		header.Availability != strategyevidence.AvailabilityAvailable || header.Confidence != strategyevidence.ConfidenceVerified || header.Currency != input.QuoteCurrency {
		return EvaluationRequest{}, ErrProductionProposalUnavailable
	}
	var payload productionDisclosurePayload
	if err := strictDecode(selected.CanonicalPayload(), &payload); err != nil {
		return EvaluationRequest{}, ErrProductionProposalUnavailable
	}
	evidence, err := productionDisclosureEvidence(payload, input.Snapshot.Digest)
	if err != nil || evidence.Market != input.Market || evidence.Source != source || evidence.SchemaVersion != schema || evidence.Symbol != input.Symbol ||
		evidence.Currency != input.QuoteCurrency || evidence.FilingID != header.SourceRecordID || evidence.RevisionID != header.RevisionIdentity ||
		!evidence.AsOf.Equal(header.SourceEventAt) || !evidence.ObservedAt.Equal(header.ObservedAt) || evidence.IngestedAt.After(header.IngestedAt) ||
		evidence.CutoffAt.After(input.Snapshot.IngestionCutoff) || !evidence.EvaluatedAt.Equal(input.Snapshot.EvaluationAt) {
		return EvaluationRequest{}, ErrProductionProposalUnavailable
	}
	evidence.IngestedAt = header.IngestedAt
	evidence.CutoffAt = input.Snapshot.IngestionCutoff
	evidence.seal = evidenceSnapshotSeal(evidence)
	config := newDisclosureConfig(input.Market, source, schema, input.ModelVersion, input.ModelConfigDigest, input.ThresholdDigest)
	plan, err := productionPlan(input)
	if err != nil {
		return EvaluationRequest{}, err
	}
	week, reservations, err := productionReservationState(input)
	if err != nil {
		return EvaluationRequest{}, err
	}
	stop := mintStopCandidate(stopCandidateInput{PriceMinor: input.Stop.PriceMinor, Version: input.Stop.Version, Source: input.Stop.Source,
		Policy: input.Stop.Policy, Digest: input.Stop.Digest, ObservedAt: input.Stop.ObservedAt.UTC(), FreshUntil: input.Stop.FreshUntil.UTC()})
	request := EvaluationRequest{CandidateID: input.CandidateID, Plan: plan, Evidence: evidence, Config: config, MarketWeek: week,
		Reservations: reservations, ReservationID: input.Reservation.ReservationID, Leg: input.Leg, Risk: NewRiskState(plan), StopCandidate: stop,
		SavedEffectiveStopMinor: input.SavedEffectiveStopMinor, EntryPriceMinor: input.EntryPriceMinor, StagedTargetMinor: input.StagedTargetMinor,
		EntryCostsMinor: input.EntryCostsMinor, EstimatedExitCostsLeviesMinor: input.EstimatedExitCostsMinor, MinimumRRPPM: input.MinimumRRPPM}
	request.authorization = mintDormantEvaluationAuthorization(plan, evidence)
	request.executionTerms = mintExecutionTermsPreimage(plan, evidence, request.EntryPriceMinor, request.StagedTargetMinor, request.EntryCostsMinor, request.EstimatedExitCostsLeviesMinor, request.MinimumRRPPM)
	request.savedStopAuthority = mintSavedStopAuthority(plan, evidence, input.SavedEffectiveStopMinor)
	return request, nil
}

func productionDisclosureEvidence(payload productionDisclosurePayload, snapshotDigest string) (DisclosureEvidence, error) {
	revision, revisionOK := productionUint(payload.RevisionSequence, 64)
	scale, scaleOK := productionUint(payload.MonetaryScale, 32)
	shares, sharesOK := productionUint(payload.DilutedShares, 64)
	parseTime := func(raw string) (time.Time, error) { return time.Parse(time.RFC3339Nano, raw) }
	asOf, first := parseTime(payload.AsOf)
	observed, second := parseTime(payload.ObservedAt)
	ingested, third := parseTime(payload.IngestedAt)
	cutoff, fourth := parseTime(payload.CutoffAt)
	evaluated, fifth := parseTime(payload.EvaluatedAt)
	fresh, sixth := parseTime(payload.FreshUntil)
	dilution, seventh := parseTime(payload.DilutionAsOf)
	if !revisionOK || revision == 0 || !scaleOK || !sharesOK || shares == 0 || first != nil || second != nil || third != nil || fourth != nil || fifth != nil || sixth != nil || seventh != nil || snapshotDigest == "" {
		return DisclosureEvidence{}, ErrProductionProposalUnavailable
	}
	evidence := DisclosureEvidence{SchemaVersion: payload.SchemaVersion, Market: payload.Market, Source: payload.Source, Symbol: payload.Symbol,
		IssuerID: payload.IssuerID, FilingID: payload.FilingID, ReportID: payload.ReportID, RevisionID: payload.RevisionID,
		SupersededRevisionID: payload.SupersededRevisionID, RevisionSequence: revision, AsOf: asOf.UTC(), ObservedAt: observed.UTC(),
		IngestedAt: ingested.UTC(), CutoffAt: cutoff.UTC(), EvaluatedAt: evaluated.UTC(), FreshUntil: fresh.UTC(), Currency: payload.Currency,
		MonetaryUnit: payload.MonetaryUnit, MonetaryScale: uint32(scale), DilutedShares: shares, SharesUnit: payload.SharesUnit,
		DilutionStatus: payload.DilutionStatus, DilutionFactsDigest: payload.DilutionFactsDigest, DilutionAsOf: dilution.UTC(),
		FinancialInputs: append([]FinancialInput(nil), payload.FinancialInputs...), ModelID: payload.ModelID, ModelVersion: payload.ModelVersion,
		ModelConfigDigest: payload.ModelConfigDigest, ThresholdDigest: payload.ThresholdDigest, EvidenceDigest: snapshotDigest,
		EquityValueMinor: payload.EquityValueMinor, FairValueMinor: payload.FairValueMinor}
	evidence.seal = evidenceSnapshotSeal(evidence)
	return evidence, nil
}

func productionPlan(input ProductionProposalInput) (CampaignPlan, error) {
	reserve, err := input.FX.EvidenceAt(input.Snapshot.EvaluationAt, input.QuoteCurrency, input.AccountCurrency)
	if err != nil {
		return CampaignPlan{}, ErrProductionProposalUnavailable
	}
	var frozen *FrozenFX
	if input.AccountCurrency != input.QuoteCurrency {
		value, err := mintFrozenFX(frozenFXInput{Authority: a066FXAuthority, Version: reserve.Version(), QuoteID: reserve.Digest(),
			AsOf: reserve.ObservedAt().UTC().Format(time.RFC3339Nano), FreshUntil: reserve.FreshUntil().UTC().Format(time.RFC3339Nano),
			Direction: FXQuoteToAccount, RateQuoteToAccount: reserve.RateQuoteToBase(), Haircut: reserve.Haircut(), Digest: reserve.Digest()})
		if err != nil {
			return CampaignPlan{}, ErrProductionProposalUnavailable
		}
		frozen = &value
	}
	plan, err := BuildCampaignPlan(PlanRequest{LaneID: map[Market]string{MarketKR: KRWeeklyLaneID, MarketUS: USWeeklyLaneID}[input.Market],
		LaneVersion: LaneVersionV1, Market: input.Market, AccountRef: input.AccountRef, Symbol: input.Symbol, CampaignID: input.CampaignID,
		PositionGeneration: input.PositionGeneration, RiskBudgetMinor: input.RiskBudgetMinor, PerShareRiskMinor: input.PerShareRiskMinor,
		PlannedQuantity: input.PlannedQuantity, PolicyDigest: input.PolicyDigest, ConfigDigest: input.ModelConfigDigest,
		AccountCurrency: input.AccountCurrency, QuoteCurrency: input.QuoteCurrency, FX: frozen})
	if err != nil {
		return CampaignPlan{}, ErrProductionProposalUnavailable
	}
	return plan, nil
}

func productionReservationState(input ProductionProposalInput) (MarketWeekEvidence, ReservationState, error) {
	r := input.Reservation
	if r.ReservationID == "" || r.CampaignID != input.CampaignID || r.Market != string(input.Market) || r.Status != "ACTIVE" ||
		r.PlannedOrdinal != input.Leg.Ordinal || r.ScopeVersion == 0 || r.PositiveLegCount != r.PlannedOrdinal-1 || r.RequestDigest == "" || r.RecordDigest == "" ||
		!r.EvaluatedAt.Equal(input.Snapshot.EvaluationAt) {
		return MarketWeekEvidence{}, ReservationState{}, ErrProductionProposalUnavailable
	}
	week := MarketWeekEvidence{Market: input.Market, Provider: r.Provider, TimeZone: r.TimeZone, SessionDate: r.SessionDate, StableIdentity: r.StableWeek,
		Official: true, CalendarGeneration: r.CalendarGeneration, CalendarDigest: r.CalendarDigest, ObservedAt: r.ObservedAt.UTC(), FreshUntil: r.FreshUntil.UTC()}
	if code := ValidateMarketWeek(week, input.Snapshot.EvaluationAt); code != "" {
		return MarketWeekEvidence{}, ReservationState{}, ErrProductionProposalUnavailable
	}
	key := CanonicalReservationKey(input.CampaignID, week)
	state := ReservationState{entries: map[string]ReservationEntry{key: {ReservationID: r.ReservationID, CampaignID: input.CampaignID, Key: key,
		MarketWeek: week, PlannedOrdinal: r.PlannedOrdinal, Status: ReservationActive}}, receipts: map[string]reservationReceipt{},
		scopes: map[string]reservationScopeState{reservationScopeKey(input.CampaignID, input.Market): {Version: r.ScopeVersion, PositiveLegCount: r.PositiveLegCount}}}
	state.seal = reservationStateSeal(state)
	return week, state, nil
}

func productionUint(raw string, bits int) (uint64, bool) {
	value, err := strconv.ParseUint(raw, 10, bits)
	return value, err == nil && strconv.FormatUint(value, 10) == raw
}
