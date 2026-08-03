package weeklyvaluelane

import (
	"crypto/sha256"
	"math/big"
	"strconv"
	"strings"
	"time"
	"unicode"
)

type StopCandidate struct {
	PriceMinor, Version    string
	Valid                  bool
	Source, Policy, Digest string
	ObservedAt, FreshUntil time.Time
	seal                   [32]byte
}

type stopCandidateInput struct {
	PriceMinor, Version, Source, Policy, Digest string
	ObservedAt, FreshUntil                      time.Time
}

func mintStopCandidate(input stopCandidateInput) StopCandidate {
	candidate := StopCandidate{PriceMinor: input.PriceMinor, Version: input.Version, Valid: true, Source: input.Source, Policy: input.Policy,
		Digest: input.Digest, ObservedAt: input.ObservedAt, FreshUntil: input.FreshUntil}
	if !validBoundedIdentity(input.Version) || !validBoundedIdentity(input.Source) || !validBoundedIdentity(input.Policy) || !validBoundedIdentity(input.Digest) ||
		input.ObservedAt.IsZero() || input.FreshUntil.IsZero() || input.ObservedAt.After(input.FreshUntil) {
		return StopCandidate{}
	}
	candidate.seal = stopCandidateSeal(candidate)
	return candidate
}

type Invalidation struct {
	Structural, ExitPending bool
	Code                    string
}

type EvaluationRequest struct {
	CandidateID                                    string
	Plan                                           CampaignPlan
	Evidence                                       DisclosureEvidence
	Config                                         DisclosureConfig
	MarketWeek                                     MarketWeekEvidence
	Reservations                                   ReservationState
	ReservationID                                  string
	Leg                                            LegProgress
	Cap                                            RiskCap
	Risk                                           RiskState
	StopCandidate                                  StopCandidate
	SavedEffectiveStopMinor                        string
	Invalidation                                   Invalidation
	EntryPriceMinor, StagedTargetMinor             string
	EntryCostsMinor, EstimatedExitCostsLeviesMinor string
	MinimumRRPPM                                   uint64
	executionTerms                                 ExecutionTermsPreimage
	savedStopAuthority                             savedStopAuthority
	authorization                                  evaluationAuthorization
}

type ResultLineage struct {
	Market                             Market
	Source                             DisclosureSource
	LaneID, LaneVersion                string
	CampaignID, CandidateID, Symbol    string
	FilingID, RevisionID               string
	SupersededRevisionID               string
	RevisionSequence                   uint64
	ModelID, ModelVersion              string
	ConfigDigest, EvidenceDigest       string
	CalendarGeneration                 string
	CalendarDigest, SessionDate        string
	StableWeek                         string
	ReservationID, PlanDigest          string
	PositionGeneration                 uint64
	RiskBudgetMinor                    string
	PlannedLegOrdinal                  int
	PlannedLegCeiling                  uint64
	CapSnapshotID, CapReservationMinor string
	CapReservationQuantity             uint64
	FXQuoteID, FXDigest                string
	DecisionDigest                     string
}

type Outcome struct {
	Kind                  OutcomeKind
	Code                  RefusalCode
	InvalidationCode      string
	Quantity              uint64
	EntryPriceMinor       string
	EffectiveStopMinor    string
	TargetPriceMinor      string
	EntryProvenance       PriceProvenance
	StopProvenance        PriceProvenance
	TargetProvenance      PriceProvenance
	ExecutionPolicy       RRExecutionPolicy
	Lineage               ResultLineage
	CommonExitIndependent bool
	ExitDecisionCreated   bool
}

func EvaluateKR(request EvaluationRequest) Outcome {
	return evaluate(request, MarketKR, SourceOpenDART, EvaluateKREvidence)
}

func EvaluateUS(request EvaluationRequest) Outcome {
	return evaluate(request, MarketUS, SourceEDGAR, EvaluateUSEvidence)
}

func evaluate(request EvaluationRequest, market Market, source DisclosureSource, evidenceEvaluator func(DisclosureEvidence, DisclosureConfig) EvidenceResult) Outcome {
	lineage := ResultLineage{Market: market, Source: source, LaneID: request.Plan.laneID, LaneVersion: request.Plan.laneVersion, CampaignID: request.Plan.campaignID,
		CandidateID: request.CandidateID, Symbol: request.Plan.symbol, FilingID: request.Evidence.FilingID, RevisionID: request.Evidence.RevisionID,
		SupersededRevisionID: request.Evidence.SupersededRevisionID, RevisionSequence: request.Evidence.RevisionSequence,
		ModelID: request.Evidence.ModelID, ModelVersion: request.Evidence.ModelVersion, ConfigDigest: request.Evidence.ModelConfigDigest,
		EvidenceDigest: request.Evidence.EvidenceDigest, CalendarGeneration: request.MarketWeek.CalendarGeneration, CalendarDigest: request.MarketWeek.CalendarDigest,
		SessionDate: request.MarketWeek.SessionDate, StableWeek: request.MarketWeek.StableIdentity, ReservationID: request.ReservationID, PlanDigest: request.Plan.digest,
		PositionGeneration: request.Plan.positionGeneration, RiskBudgetMinor: request.Plan.riskBudgetMinor, PlannedLegOrdinal: request.Leg.Ordinal,
		CapSnapshotID: request.Cap.snapshotID, CapReservationMinor: request.Cap.reservationMinor, CapReservationQuantity: request.Cap.reservationQuantity}
	if request.Leg.Ordinal >= 1 && request.Leg.Ordinal <= len(request.Plan.legCeilings) {
		lineage.PlannedLegCeiling = request.Plan.legCeilings[request.Leg.Ordinal-1]
	}
	if request.Plan.fx != nil {
		lineage.FXQuoteID, lineage.FXDigest = request.Plan.fx.quoteID, request.Plan.fx.digest
	}
	refuse := func(code RefusalCode) Outcome {
		return Outcome{Kind: OutcomeRefusal, Code: code, Lineage: lineage, CommonExitIndependent: true}
	}
	if !request.authorization.valid(request.Plan, request.Evidence, market) {
		return refuse(RefusalLaneOff)
	}
	if !validBoundedIdentity(request.CandidateID) {
		return refuse(RefusalSchemaInvalid)
	}
	if !request.Plan.valid() || request.Plan.market != market || request.Plan.laneID != map[Market]string{MarketKR: KRWeeklyLaneID, MarketUS: USWeeklyLaneID}[market] {
		return refuse(RefusalPlanInvalid)
	}
	if request.Evidence.Market != market {
		return refuse(RefusalMarketMismatch)
	}
	if request.Evidence.Source != source {
		return refuse(RefusalSourceMismatch)
	}
	if request.Invalidation.Structural || request.Invalidation.ExitPending {
		if request.Invalidation.Code == "" {
			return refuse(RefusalInvalidation)
		}
		return Outcome{Kind: OutcomeInvalidation, Code: RefusalInvalidation, InvalidationCode: request.Invalidation.Code, Lineage: lineage, CommonExitIndependent: true}
	}
	evidenceResult := evidenceEvaluator(request.Evidence, request.Config)
	lineage.DecisionDigest = evidenceResult.DecisionDigest
	if !evidenceResult.Accepted {
		outcome := refuse(evidenceResult.Code)
		outcome.Lineage = lineage
		return outcome
	}
	if !request.executionTerms.valid(request.Plan, request.Evidence, request) {
		return refuse(RefusalExecutionTermsInvalid)
	}
	if request.Evidence.Symbol != request.Plan.symbol || request.Evidence.ModelConfigDigest != request.Plan.configDigest {
		return refuse(RefusalConfigMismatch)
	}
	if request.MarketWeek.Market != market {
		return refuse(RefusalMarketMismatch)
	}
	if code := ValidateMarketWeek(request.MarketWeek, request.Evidence.EvaluatedAt); code != "" {
		return refuse(code)
	}
	if request.Leg.Ordinal < 1 || request.Leg.Ordinal > 7 || request.Leg.Cancelled || request.Leg.Expired {
		return refuse(RefusalLegTerminal)
	}
	if !validReservationState(request.Reservations) {
		return refuse(RefusalReservationMissing)
	}
	key := CanonicalReservationKey(request.Plan.campaignID, request.MarketWeek)
	entry, exists := request.Reservations.Entry(key)
	if !exists || entry.ReservationID != request.ReservationID || entry.CampaignID != request.Plan.campaignID || entry.PlannedOrdinal != request.Leg.Ordinal {
		return refuse(RefusalReservationMissing)
	}
	if entry.Status != ReservationActive {
		return refuse(RefusalReservationTerminal)
	}
	if request.Reservations.PositiveLegCount(request.Plan.campaignID, market) >= 7 {
		return refuse(RefusalPlanExhausted)
	}
	quantity := PlannedLegQuantity(request.Plan, request.Leg, request.Cap.qFinal)
	if quantity == 0 {
		return refuse(RefusalPlanExhausted)
	}
	if !request.Cap.validAt(request.Plan, request.Evidence.EvaluatedAt, quantity) {
		return refuse(RefusalCapInvalid)
	}
	effectiveStop, stopCode := effectiveStop(request.SavedEffectiveStopMinor, request.StopCandidate, request.Evidence.EvaluatedAt)
	if stopCode != "" {
		return refuse(stopCode)
	}
	useSavedStop := effectiveStop != request.StopCandidate.PriceMinor
	if useSavedStop && !request.savedStopAuthority.valid(request.Plan, request.Evidence, effectiveStop) {
		return refuse(RefusalExecutionTermsInvalid)
	}
	if _, entryOK := canonicalPositiveMinor(request.EntryPriceMinor); !entryOK {
		return refuse(RefusalExecutionTermsInvalid)
	}
	if _, targetOK := canonicalPositiveMinor(request.StagedTargetMinor); !targetOK {
		return refuse(RefusalExecutionTermsInvalid)
	}
	entryPrice, entryOK := parseUnsigned(request.EntryPriceMinor)
	stopPrice, stopOK := parseUnsigned(effectiveStop)
	maxDistance, maxOK := parseUnsigned(request.Cap.maxStopDistanceMinor)
	if !entryOK || !stopOK || !maxOK || entryPrice.Cmp(stopPrice) <= 0 {
		return refuse(RefusalStopInvalid)
	}
	distance := new(big.Int).Sub(entryPrice, stopPrice)
	if distance.Cmp(maxDistance) > 0 {
		return refuse(RefusalStructuralStopCap)
	}
	rr := CalculateRR(RRInput{EntryPriceMinor: request.EntryPriceMinor, StagedTargetMinor: request.StagedTargetMinor, FairValueMinor: evidenceResult.FairValueMinor,
		EffectiveStopMinor: effectiveStop, Quantity: quantity, EntryCostsMinor: request.EntryCostsMinor, EstimatedExitCostsLeviesMinor: request.EstimatedExitCostsLeviesMinor,
		MinimumRRPPM: request.MinimumRRPPM, AccountCurrency: request.Plan.accountCurrency, QuoteCurrency: request.Plan.quoteCurrency, FX: request.Plan.FX(),
		EvaluatedAt: request.Evidence.EvaluatedAt})
	if !rr.Accepted {
		return refuse(rr.Code)
	}
	entryPrice, entryTermsOK := canonicalPositiveMinor(request.EntryPriceMinor)
	stopPrice, stopTermsOK := canonicalPositiveMinor(effectiveStop)
	targetPrice, targetTermsOK := canonicalPositiveMinor(rr.TargetMinor)
	if !entryTermsOK || !stopTermsOK || !targetTermsOK || stopPrice.Cmp(entryPrice) >= 0 || entryPrice.Cmp(targetPrice) >= 0 {
		return refuse(RefusalExecutionTermsInvalid)
	}
	if code := AdmitRisk(request.Plan, request.Risk, request.Cap); code != "" {
		return refuse(code)
	}
	policy := executionPolicy(request.executionTerms, lineage)
	scale, scaleOK := weeklyScale(request.Plan.quoteCurrency)
	if !scaleOK {
		return refuse(RefusalExecutionTermsInvalid)
	}
	asOf := request.Evidence.AsOf.UTC().Format(time.RFC3339Nano)
	entryProvenance := PriceProvenance{entryPrice.String(), string(request.Evidence.Source) + ":" + request.Evidence.FilingID, request.Evidence.SchemaVersion, request.Evidence.EvidenceDigest, asOf, request.Plan.quoteCurrency, "minor-v1", scale}
	stopProvenance := PriceProvenance{stopPrice.String(), request.StopCandidate.Source, request.StopCandidate.Version, request.StopCandidate.Digest, request.StopCandidate.ObservedAt.UTC().Format(time.RFC3339Nano), request.Plan.quoteCurrency, "minor-v1", scale}
	if useSavedStop {
		stopProvenance = request.savedStopAuthority.provenance
	}
	targetProvenance := PriceProvenance{targetPrice.String(), "weekly-rr-capped-target", "rr-policy-v1", policy.Identity, request.Evidence.EvaluatedAt.UTC().Format(time.RFC3339Nano), request.Plan.quoteCurrency, "minor-v1", scale}
	return Outcome{Kind: OutcomeDecision, Quantity: quantity, EntryPriceMinor: entryPrice.String(), EffectiveStopMinor: stopPrice.String(), TargetPriceMinor: targetPrice.String(),
		EntryProvenance: entryProvenance, StopProvenance: stopProvenance, TargetProvenance: targetProvenance, ExecutionPolicy: policy, Lineage: lineage, CommonExitIndependent: true}
}

func canonicalPositiveMinor(raw string) (*big.Int, bool) {
	if raw == "" || strings.TrimSpace(raw) != raw {
		return nil, false
	}
	value, ok := parseUnsigned(raw)
	return value, ok && value.Sign() > 0 && value.String() == raw
}

func effectiveStop(saved string, candidate StopCandidate, evaluatedAt time.Time) (string, RefusalCode) {
	if !candidate.Valid || candidate.seal == ([32]byte{}) || candidate.seal != stopCandidateSeal(candidate) || candidate.ObservedAt.IsZero() ||
		candidate.FreshUntil.IsZero() || evaluatedAt.Before(candidate.ObservedAt) || evaluatedAt.After(candidate.FreshUntil) {
		return "", RefusalStopInvalid
	}
	candidatePrice, candidateOK := parseUnsigned(candidate.PriceMinor)
	if !candidateOK || candidatePrice.Sign() <= 0 {
		return "", RefusalStopInvalid
	}
	if saved == "" {
		return candidatePrice.String(), ""
	}
	savedPrice, savedOK := parseUnsigned(saved)
	if !savedOK || savedPrice.Sign() <= 0 {
		return "", RefusalStopInvalid
	}
	if savedPrice.Cmp(candidatePrice) > 0 {
		return savedPrice.String(), ""
	}
	return candidatePrice.String(), ""
}

func stopCandidateSeal(candidate StopCandidate) [32]byte {
	return sha256.Sum256([]byte(strings.Join([]string{"weekly-value-stop-v1", candidate.PriceMinor, candidate.Version,
		strconv.FormatBool(candidate.Valid), candidate.Source, candidate.Policy, candidate.Digest, canonicalTime(candidate.ObservedAt), canonicalTime(candidate.FreshUntil)}, "\x00")))
}

type evaluationAuthorization struct {
	market       Market
	planDigest   string
	evidenceSeal [32]byte
	evaluatedAt  time.Time
	seal         [32]byte
}

func mintDormantEvaluationAuthorization(plan CampaignPlan, evidence DisclosureEvidence) evaluationAuthorization {
	authorization := evaluationAuthorization{market: plan.market, planDigest: plan.digest, evidenceSeal: evidence.seal, evaluatedAt: evidence.EvaluatedAt}
	if !plan.valid() || evidence.seal == ([32]byte{}) || evidence.seal != evidenceSnapshotSeal(evidence) {
		return evaluationAuthorization{}
	}
	authorization.seal = evaluationAuthorizationSeal(authorization)
	return authorization
}

func (authorization evaluationAuthorization) valid(plan CampaignPlan, evidence DisclosureEvidence, market Market) bool {
	return authorization.seal != ([32]byte{}) && authorization.seal == evaluationAuthorizationSeal(authorization) && authorization.market == market &&
		authorization.planDigest == plan.digest && authorization.evidenceSeal == evidence.seal && authorization.evaluatedAt.Equal(evidence.EvaluatedAt)
}

func evaluationAuthorizationSeal(authorization evaluationAuthorization) [32]byte {
	return sha256.Sum256([]byte(strings.Join([]string{"weekly-value-dormant-evaluation-v1", string(authorization.market), authorization.planDigest,
		canonicalTime(authorization.evaluatedAt), string(authorization.evidenceSeal[:])}, "\x00")))
}

func validBoundedIdentity(value string) bool {
	if strings.TrimSpace(value) == "" || len(value) > maxIdentityBytes {
		return false
	}
	for _, r := range value {
		if unicode.IsControl(r) {
			return false
		}
	}
	return true
}

type CommonExitProbe struct{ Emergency bool }

func (probe CommonExitProbe) CanProceed(outcome Outcome) bool {
	return probe.Emergency && outcome.CommonExitIndependent
}
