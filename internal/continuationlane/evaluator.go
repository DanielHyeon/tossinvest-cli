package continuationlane

import (
	"crypto/sha256"
	"encoding/hex"
	"strconv"
	"strings"
	"time"
)

type RiskCap struct {
	Market              Market
	QFinal              uint64
	ReservationQuantity uint64
	ReservationMinor    string
	SnapshotID          string
	PolicyDigest        string
	BucketSetDigest     string
	ObservedAt          string
	FreshUntil          string
	FX                  *FrozenFX
	official            bool
	frozen              bool
	planDigest          string
	seal                [32]byte
}

type riskCapInput struct {
	QFinal              uint64
	ReservationQuantity uint64
	ReservationMinor    string
	SnapshotID          string
	PolicyDigest        string
	BucketSetDigest     string
	ObservedAt          string
	FreshUntil          string
	FX                  *FrozenFX
}

func newRiskCap(plan CampaignPlan, input riskCapInput) (RiskCap, error) {
	cap := RiskCap{Market: plan.Market, QFinal: input.QFinal, ReservationQuantity: input.ReservationQuantity,
		ReservationMinor: input.ReservationMinor, SnapshotID: strings.TrimSpace(input.SnapshotID), PolicyDigest: strings.TrimSpace(input.PolicyDigest),
		BucketSetDigest: strings.TrimSpace(input.BucketSetDigest), ObservedAt: input.ObservedAt, FreshUntil: input.FreshUntil, official: true, frozen: true}
	if input.FX != nil {
		fx := *input.FX
		cap.FX = &fx
	}
	if !validatePlan(plan) || cap.QFinal == 0 || cap.ReservationQuantity == 0 || cap.ReservationQuantity > cap.QFinal || cap.SnapshotID == "" || cap.PolicyDigest != plan.PolicyDigest || cap.BucketSetDigest == "" {
		return RiskCap{}, errInvalidRiskCap
	}
	reservation, err := parseUnsigned(cap.ReservationMinor)
	if err != nil || reservation.Sign() <= 0 {
		return RiskCap{}, errInvalidRiskCap
	}
	observed, observedErr := time.Parse(time.RFC3339Nano, cap.ObservedAt)
	fresh, freshErr := time.Parse(time.RFC3339Nano, cap.FreshUntil)
	if observedErr != nil || freshErr != nil || observed.After(fresh) {
		return RiskCap{}, errInvalidRiskCap
	}
	if plan.AccountCurrency != plan.QuoteCurrency {
		if plan.FX == nil || cap.FX == nil || *cap.FX != *plan.FX || !validFX(*cap.FX) {
			return RiskCap{}, errInvalidRiskCap
		}
	} else if cap.FX != nil {
		return RiskCap{}, errInvalidRiskCap
	}
	cap.planDigest = plan.RiskBudgetDigest
	cap.seal = riskCapSeal(cap)
	return cap, nil
}

var errInvalidRiskCap = &riskCapError{}

type riskCapError struct{}

func (*riskCapError) Error() string { return "continuation lanes: invalid frozen a066 risk cap" }

type StopCandidate struct {
	PriceMinor string
	Valid      bool
	Source     string
	Policy     string
	Version    string
	Digest     string
	ObservedAt string
	FreshUntil string
	seal       [32]byte
}

type stopCandidateInput struct {
	PriceMinor string
	Source     string
	Policy     string
	Version    string
	Digest     string
	ObservedAt string
	FreshUntil string
}

func newStopCandidate(input stopCandidateInput) (StopCandidate, error) {
	candidate := StopCandidate{PriceMinor: input.PriceMinor, Valid: true, Source: strings.TrimSpace(input.Source), Policy: strings.TrimSpace(input.Policy),
		Version: strings.TrimSpace(input.Version), Digest: strings.TrimSpace(input.Digest), ObservedAt: input.ObservedAt, FreshUntil: input.FreshUntil}
	price, priceErr := parseUnsigned(candidate.PriceMinor)
	observed, observedErr := time.Parse(time.RFC3339Nano, candidate.ObservedAt)
	fresh, freshErr := time.Parse(time.RFC3339Nano, candidate.FreshUntil)
	if priceErr != nil || price.Sign() <= 0 || candidate.Source == "" || candidate.Policy == "" || candidate.Version == "" || candidate.Digest == "" ||
		observedErr != nil || freshErr != nil || observed.After(fresh) {
		return StopCandidate{}, &stopCandidateError{}
	}
	candidate.PriceMinor = price.String()
	candidate.seal = stopCandidateSeal(candidate)
	return candidate, nil
}

type stopCandidateError struct{}

func (*stopCandidateError) Error() string {
	return "continuation lanes: invalid stop candidate provenance"
}

type Invalidation struct {
	Structural  bool
	ExitPending bool
	Code        string
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
	SavedStopProvenance     PriceProvenance
	Invalidation            Invalidation
}

type KREvaluationRequest struct {
	Context  EvaluationContext
	Evidence KREvidence
	Config   KRFlowConfig
}

type USEvaluationRequest struct {
	Context  EvaluationContext
	Evidence USEvidence
	Config   USParticipationConfig
}

type ResultLineage struct {
	AccountRef         string
	Market             Market
	Symbol             string
	PositionGeneration int64
	LaneID             string
	LaneVersion        string
	CandidateID        string
	EvidenceDigest     string
	SchemaVersion      string
	ConfigDigest       string
	CampaignID         string
	RiskBudgetDigest   string
	LegOrdinal         int
	PlannedCeiling     uint64
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
	ExecutionPolicyDigest string
	Lineage               ResultLineage
	CommonExitIndependent bool
	ExitDecisionCreated   bool
}

func EvaluateKR(request KREvaluationRequest) Outcome {
	return evaluate(request.Context, request.Evidence.Envelope, EvaluateKRFlow(request.Evidence, request.Config))
}

func EvaluateUS(request USEvaluationRequest) Outcome {
	return evaluate(request.Context, request.Evidence.Envelope, EvaluateUSParticipation(request.Evidence, request.Config))
}

func evaluate(context EvaluationContext, envelope EvidenceEnvelope, signal SignalResult) Outcome {
	outcome := Outcome{Kind: OutcomeRefusal, CommonExitIndependent: true, Lineage: lineageFor(context, envelope)}
	if !validatePlan(context.Plan) {
		outcome.Code = RefusalPlanInvalid
		return outcome
	}
	if context.Invalidation.Structural || context.Invalidation.ExitPending {
		if strings.TrimSpace(context.Invalidation.Code) == "" {
			outcome.Code = RefusalInvalidationInvalid
			return outcome
		}
		outcome.Kind = OutcomeInvalidation
		outcome.InvalidationCode = context.Invalidation.Code
		return outcome
	}
	if !context.Enabled {
		outcome.Code = RefusalLaneOff
		return outcome
	}
	if envelope.Market != context.Plan.Market || envelope.Symbol != context.Plan.Symbol || context.Cap.Market != context.Plan.Market {
		outcome.Code = RefusalMarketMismatch
		return outcome
	}
	if !signal.Accepted {
		outcome.Code = signal.Code
		return outcome
	}
	effectiveStop, ok := composeStop(context.SavedEffectiveStopMinor, context.StopCandidate, envelope.EvaluatedAt)
	if !ok {
		outcome.Code = RefusalStopInvalid
		return outcome
	}
	if context.Leg.Ordinal < 1 || context.Leg.Ordinal > 3 || context.Leg.FilledQuantity > context.Plan.LegCeilings[context.Leg.Ordinal-1] {
		outcome.Code = RefusalPlanInvalid
		return outcome
	}
	if context.Leg.Cancelled || context.Leg.Expired {
		outcome.Code = RefusalLegTerminal
		return outcome
	}
	if context.Risk.Latches[LatchCampaignRiskOverage] || context.Risk.Latches[LatchUnknownActualRisk] {
		outcome.Code = RefusalRiskLatched
		return outcome
	}
	if context.Risk.RiskBudgetDigest != context.Plan.RiskBudgetDigest || context.Risk.BudgetMinor != context.Plan.RiskBudgetMinor {
		outcome.Code = RefusalPlanInvalid
		return outcome
	}
	quantity := PlannedLegQuantity(context.Plan, context.Leg, context.Cap.QFinal)
	if quantity == 0 {
		outcome.Code = RefusalQuantityUnavailable
		return outcome
	}
	if !validCap(context.Cap, envelope.EvaluatedAt, context.Plan, quantity) {
		outcome.Code = RefusalCapInvalid
		return outcome
	}
	filled, err := parseUnsigned(context.Risk.FilledMinor)
	if err != nil {
		outcome.Code = codeForArithmetic(err)
		return outcome
	}
	held, err := parseUnsigned(context.Risk.HeldMinor)
	if err != nil {
		outcome.Code = codeForArithmetic(err)
		return outcome
	}
	proposed, err := parseUnsigned(context.Cap.ReservationMinor)
	if err != nil {
		outcome.Code = codeForArithmetic(err)
		return outcome
	}
	used, err := checkedAdd(filled, held, proposed)
	if err != nil {
		outcome.Code = codeForArithmetic(err)
		return outcome
	}
	budget, err := parseUnsigned(context.Plan.RiskBudgetMinor)
	if err != nil {
		outcome.Code = codeForArithmetic(err)
		return outcome
	}
	if used.Cmp(budget) > 0 {
		outcome.Code = RefusalRiskBudgetExceeded
		return outcome
	}
	stopAuthority, stopAuthorityOK := stopProvenance(context.Plan, envelope, context.StopCandidate, effectiveStop, context.SavedStopProvenance)
	entryAuthority, stopAuthority, targetAuthority, policyDigest, termsOK := validatedExecutionTerms(context.Plan, envelope, context.ExecutionTerms, effectiveStop, stopAuthority)
	if !stopAuthorityOK || !termsOK {
		outcome.Code = RefusalExecutionTermsInvalid
		return outcome
	}
	outcome.Kind = OutcomeDecision
	outcome.Code = RefusalNone
	outcome.Quantity = quantity
	outcome.EntryPriceMinor = entryAuthority.PriceMinor
	outcome.EffectiveStopMinor = stopAuthority.PriceMinor
	outcome.TargetPriceMinor = targetAuthority.PriceMinor
	outcome.EntryProvenance, outcome.StopProvenance, outcome.TargetProvenance, outcome.ExecutionPolicyDigest = entryAuthority, stopAuthority, targetAuthority, policyDigest
	return outcome
}

func lineageFor(context EvaluationContext, envelope EvidenceEnvelope) ResultLineage {
	ceiling := uint64(0)
	if context.Leg.Ordinal >= 1 && context.Leg.Ordinal <= 3 {
		ceiling = context.Plan.LegCeilings[context.Leg.Ordinal-1]
	}
	return ResultLineage{AccountRef: context.Plan.AccountRef, Market: context.Plan.Market, Symbol: context.Plan.Symbol,
		PositionGeneration: context.Plan.PositionGeneration, LaneID: context.Plan.LaneID, LaneVersion: context.Plan.LaneVersion,
		CandidateID: context.CandidateID, EvidenceDigest: envelope.SourceDigest, SchemaVersion: envelope.SchemaVersion,
		ConfigDigest: envelope.ConfigDigest, CampaignID: context.Plan.CampaignID, RiskBudgetDigest: context.Plan.RiskBudgetDigest,
		LegOrdinal: context.Leg.Ordinal, PlannedCeiling: ceiling}
}

func validCap(cap RiskCap, evaluatedAt string, plan CampaignPlan, quantity uint64) bool {
	if cap.QFinal == 0 || strings.TrimSpace(cap.ReservationMinor) == "" || strings.TrimSpace(cap.SnapshotID) == "" ||
		strings.TrimSpace(cap.PolicyDigest) == "" || strings.TrimSpace(cap.BucketSetDigest) == "" || !cap.official || !cap.frozen || cap.seal != riskCapSeal(cap) ||
		cap.planDigest == "" || cap.planDigest != plan.RiskBudgetDigest || cap.PolicyDigest != plan.PolicyDigest || cap.ReservationQuantity != quantity || quantity > cap.QFinal {
		return false
	}
	if _, err := parseUnsigned(cap.ReservationMinor); err != nil {
		return false
	}
	evaluated, err := time.Parse(time.RFC3339Nano, evaluatedAt)
	if err != nil {
		return false
	}
	observed, err := time.Parse(time.RFC3339Nano, cap.ObservedAt)
	if err != nil {
		return false
	}
	fresh, err := time.Parse(time.RFC3339Nano, cap.FreshUntil)
	if err != nil || observed.After(evaluated) || fresh.Before(evaluated) {
		return false
	}
	if plan.AccountCurrency != plan.QuoteCurrency {
		return plan.FX != nil && cap.FX != nil && *cap.FX == *plan.FX && validFX(*cap.FX)
	}
	return cap.FX == nil || (plan.FX != nil && *cap.FX == *plan.FX)
}

func composeStop(saved string, candidate StopCandidate, evaluatedAt string) (string, bool) {
	if !candidate.Valid || candidate.seal == [32]byte{} || candidate.seal != stopCandidateSeal(candidate) || strings.TrimSpace(candidate.Source) == "" ||
		strings.TrimSpace(candidate.Policy) == "" || strings.TrimSpace(candidate.Version) == "" || strings.TrimSpace(candidate.Digest) == "" ||
		strings.TrimSpace(candidate.ObservedAt) == "" || strings.TrimSpace(candidate.FreshUntil) == "" {
		return "", false
	}
	observed, observedErr := time.Parse(time.RFC3339Nano, candidate.ObservedAt)
	evaluated, evaluatedErr := time.Parse(time.RFC3339Nano, evaluatedAt)
	fresh, freshErr := time.Parse(time.RFC3339Nano, candidate.FreshUntil)
	if observedErr != nil || evaluatedErr != nil || freshErr != nil || observed.After(evaluated) || evaluated.After(fresh) {
		return "", false
	}
	candidatePrice, err := parseUnsigned(candidate.PriceMinor)
	if err != nil || candidatePrice.Sign() <= 0 {
		return "", false
	}
	if saved == "" {
		return candidatePrice.String(), true
	}
	savedPrice, err := parseUnsigned(saved)
	if err != nil || savedPrice.Sign() <= 0 {
		return "", false
	}
	if savedPrice.Cmp(candidatePrice) > 0 {
		return savedPrice.String(), true
	}
	return candidatePrice.String(), true
}

func stopCandidateSeal(candidate StopCandidate) [32]byte {
	return sha256.Sum256([]byte(strings.Join([]string{candidate.PriceMinor, strconv.FormatBool(candidate.Valid), candidate.Source, candidate.Policy,
		candidate.Version, candidate.Digest, candidate.ObservedAt, candidate.FreshUntil}, "\x00")))
}

func riskCapSeal(cap RiskCap) [32]byte {
	parts := []string{cap.planDigest, string(cap.Market), strconv.FormatUint(cap.QFinal, 10), strconv.FormatUint(cap.ReservationQuantity, 10), cap.ReservationMinor,
		cap.SnapshotID, cap.PolicyDigest, cap.BucketSetDigest, cap.ObservedAt, cap.FreshUntil, strconv.FormatBool(cap.official), strconv.FormatBool(cap.frozen)}
	if cap.FX != nil {
		parts = append(parts, cap.FX.QuoteID, cap.FX.AsOf, cap.FX.Direction, cap.FX.RateQuoteToAccount, cap.FX.Haircut, cap.FX.Digest, hex.EncodeToString(cap.FX.seal[:]))
	}
	return sha256.Sum256([]byte(strings.Join(parts, "\x00")))
}
