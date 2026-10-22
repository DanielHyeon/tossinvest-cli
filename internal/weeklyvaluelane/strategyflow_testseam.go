//go:build tossos_testseams

package weeklyvaluelane

import (
	"math/big"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/officialfx"
)

// StrategyflowKRFixture returns an accepted KR weekly request only in the
// explicit repository test-seam build.
func StrategyflowKRFixture(candidateID, evidenceDigest string) (EvaluationRequest, error) {
	return strategyflowKRFixture(candidateID, evidenceDigest, 1)
}

// StrategyflowKRFollowOnFixture returns a sealed KR request with native prior
// reservation/fill transitions for one tagged follow-on ordinal.
func StrategyflowKRFollowOnFixture(candidateID, evidenceDigest string, ordinal int) (EvaluationRequest, error) {
	return strategyflowKRFixture(candidateID, evidenceDigest, ordinal)
}

func strategyflowKRFixture(candidateID, evidenceDigest string, ordinal int) (EvaluationRequest, error) {
	evidence, err := DecodeKREvidence([]byte(`{
  "schema_version":"kr-opendart-weekly-value-v1","market":"KR","source":"OPENDART","symbol":"005930","issuer_id":"00126380",
  "filing_id":"202608040001","report_id":"quarterly-2026-q2","revision_id":"rev-2","superseded_revision_id":"rev-1","revision_sequence":2,
  "as_of":"2026-06-30T00:00:00Z","observed_at":"2026-08-03T23:00:00Z","ingested_at":"2026-08-03T23:01:00Z",
  "cutoff_at":"2026-08-03T23:02:00Z","evaluated_at":"2026-08-03T23:03:00Z","fresh_until":"2026-08-10T23:03:00Z",
  "currency":"KRW","monetary_unit":"MINOR","monetary_scale":0,"diluted_shares":100,"shares_unit":"SHARES","dilution_status":"OBSERVED",
  "dilution_facts_digest":"kr-dilution-digest","dilution_as_of":"2026-06-30T00:00:00Z",
  "financial_inputs":[{"name":"equity_value","value_minor":"110000","unit":"KRW_MINOR"}],"model_id":"weekly-value",
  "model_version":"weekly-model-v1","model_config_digest":"model-config-kr","threshold_digest":"threshold-kr",
  "evidence_digest":"kr-evidence-digest","equity_value_minor":"110000","fair_value_minor":"1100"}`))
	if err != nil {
		return EvaluationRequest{}, err
	}
	evidence.EvidenceDigest = evidenceDigest
	evidence.seal = evidenceSnapshotSeal(evidence)
	return strategyflowEvaluation(candidateID, evidence, newDisclosureConfig(MarketKR, SourceOpenDART, KRDisclosureSchemaV1, "weekly-model-v1", "model-config-kr", "threshold-kr"), ordinal)
}

// StrategyflowUSFixture returns an accepted US weekly request only in the
// explicit repository test-seam build.
func StrategyflowUSFixture(candidateID, evidenceDigest string) (EvaluationRequest, error) {
	return strategyflowUSFixture(candidateID, evidenceDigest, 1)
}

// StrategyflowUSFollowOnFixture returns a sealed US request with native prior
// reservation/fill transitions for one tagged follow-on ordinal.
func StrategyflowUSFollowOnFixture(candidateID, evidenceDigest string, ordinal int) (EvaluationRequest, error) {
	return strategyflowUSFixture(candidateID, evidenceDigest, ordinal)
}

// StrategyflowUSCrossCurrencyFollowOnFixture creates synthetic tagged-only
// USD-to-KRW input through existing official-FX test authority and native FX
// creation. It confers no runtime authority.
func StrategyflowUSCrossCurrencyFollowOnFixture(candidateID, evidenceDigest string, ordinal int) (EvaluationRequest, error) {
	evidence, err := strategyflowUSEvidence(evidenceDigest)
	if err != nil {
		return EvaluationRequest{}, err
	}
	fxEvidence, err := officialfx.EvidenceForAuthorityTest("USD", "KRW", "1300", "1.1", evidence.EvaluatedAt.Add(-time.Minute), evidence.FreshUntil.AddDate(0, 0, 56))
	if err != nil {
		return EvaluationRequest{}, err
	}
	reserve, err := fxEvidence.EvidenceAt(evidence.EvaluatedAt, "USD", "KRW")
	if err != nil {
		return EvaluationRequest{}, err
	}
	fx, err := mintFrozenFX(frozenFXInput{Authority: a066FXAuthority, Version: reserve.Version(), QuoteID: reserve.Digest(),
		AsOf: reserve.ObservedAt().UTC().Format(time.RFC3339Nano), FreshUntil: reserve.FreshUntil().UTC().Format(time.RFC3339Nano),
		Direction: FXQuoteToAccount, RateQuoteToAccount: reserve.RateQuoteToBase(), Haircut: reserve.Haircut(), Digest: reserve.Digest()})
	if err != nil {
		return EvaluationRequest{}, err
	}
	return strategyflowEvaluationWithPlan(candidateID, evidence, newDisclosureConfig(MarketUS, SourceEDGAR, USDisclosureSchemaV1, "weekly-model-v1", "model-config-us", "threshold-us"), ordinal, "KRW", "USD", &fx)
}

// StrategyflowUSCrossCurrencyStaleFollowOnFixture changes only the synthetic
// evidence freshness boundary after creating a valid tagged request. Seals are
// recomputed only for the changed evidence envelope; FX authority is untouched.
func StrategyflowUSCrossCurrencyStaleFollowOnFixture(candidateID, evidenceDigest string, ordinal int) (EvaluationRequest, error) {
	request, err := StrategyflowUSCrossCurrencyFollowOnFixture(candidateID, evidenceDigest, ordinal)
	if err != nil {
		return EvaluationRequest{}, err
	}
	request.Evidence.FreshUntil = request.Evidence.EvaluatedAt.Add(-time.Nanosecond)
	request.Evidence.seal = evidenceSnapshotSeal(request.Evidence)
	request.authorization = mintDormantEvaluationAuthorization(request.Plan, request.Evidence)
	request.executionTerms = mintExecutionTermsPreimage(request.Plan, request.Evidence, request.EntryPriceMinor, request.StagedTargetMinor, request.EntryCostsMinor, request.EstimatedExitCostsLeviesMinor, request.MinimumRRPPM)
	return request, nil
}

// InvalidateStrategyflowFXForTest changes a tagged fixture's existing FX
// payload without issuing a replacement seal. It is only for native refusal
// regression tests.
func InvalidateStrategyflowFXForTest(request *EvaluationRequest) error {
	if request == nil || request.Plan.fx == nil {
		return &strategyflowFixtureError{code: RefusalRRInvalid}
	}
	request.Plan.fx.rateQuoteToAccount = "999"
	return nil
}

// RemoveStrategyflowActiveReservationForTest corrupts the current tagged
// reservation state without re-sealing it, for native integrity-refusal tests.
func RemoveStrategyflowActiveReservationForTest(request *EvaluationRequest) error {
	if request == nil {
		return &strategyflowFixtureError{code: RefusalReservationMissing}
	}
	key := CanonicalReservationKey(request.Plan.CampaignID(), request.MarketWeek)
	if _, ok := request.Reservations.entries[key]; !ok {
		return &strategyflowFixtureError{code: RefusalReservationMissing}
	}
	delete(request.Reservations.entries, key)
	return nil
}

func strategyflowUSFixture(candidateID, evidenceDigest string, ordinal int) (EvaluationRequest, error) {
	evidence, err := strategyflowUSEvidence(evidenceDigest)
	if err != nil {
		return EvaluationRequest{}, err
	}
	return strategyflowEvaluation(candidateID, evidence, newDisclosureConfig(MarketUS, SourceEDGAR, USDisclosureSchemaV1, "weekly-model-v1", "model-config-us", "threshold-us"), ordinal)
}

func strategyflowUSEvidence(evidenceDigest string) (DisclosureEvidence, error) {
	evidence, err := DecodeUSEvidence([]byte(`{
  "schema_version":"us-edgar-weekly-value-v1","market":"US","source":"EDGAR","symbol":"AAPL","issuer_id":"0000320193",
  "filing_id":"0000320193-26-000077","report_id":"10-Q-2026-Q2","revision_id":"rev-1","superseded_revision_id":"NONE","revision_sequence":1,
  "as_of":"2026-06-30T00:00:00Z","observed_at":"2026-08-03T20:00:00Z","ingested_at":"2026-08-03T20:01:00Z",
  "cutoff_at":"2026-08-03T20:02:00Z","evaluated_at":"2026-08-03T20:03:00Z","fresh_until":"2026-08-10T20:03:00Z",
  "currency":"USD","monetary_unit":"MINOR","monetary_scale":2,"diluted_shares":100,"shares_unit":"SHARES","dilution_status":"NONE",
  "dilution_facts_digest":"us-dilution-none-digest","dilution_as_of":"2026-06-30T00:00:00Z",
  "financial_inputs":[{"name":"equity_value","value_minor":"120000","unit":"USD_MINOR"}],"model_id":"weekly-value",
  "model_version":"weekly-model-v1","model_config_digest":"model-config-us","threshold_digest":"threshold-us",
  "evidence_digest":"us-evidence-digest","equity_value_minor":"120000","fair_value_minor":"1200"}`))
	if err != nil {
		return DisclosureEvidence{}, err
	}
	evidence.EvidenceDigest = evidenceDigest
	evidence.seal = evidenceSnapshotSeal(evidence)
	return evidence, nil
}

func strategyflowEvaluation(candidateID string, evidence DisclosureEvidence, config DisclosureConfig, ordinal int) (EvaluationRequest, error) {
	currency := map[Market]string{MarketKR: "KRW", MarketUS: "USD"}[evidence.Market]
	return strategyflowEvaluationWithPlan(candidateID, evidence, config, ordinal, currency, currency, nil)
}

func strategyflowEvaluationWithPlan(candidateID string, evidence DisclosureEvidence, config DisclosureConfig, ordinal int, accountCurrency, quoteCurrency string, fx *FrozenFX) (EvaluationRequest, error) {
	if ordinal < 1 || ordinal > 7 {
		return EvaluationRequest{}, &strategyflowFixtureError{code: RefusalPlanExhausted}
	}
	laneID := map[Market]string{MarketKR: KRWeeklyLaneID, MarketUS: USWeeklyLaneID}[evidence.Market]
	budget := "1000"
	if fx != nil {
		budget = "100000"
	}
	plan, err := BuildCampaignPlan(PlanRequest{LaneID: laneID, LaneVersion: LaneVersionV1, Market: evidence.Market, AccountRef: "acct", Symbol: evidence.Symbol,
		CampaignID: "campaign-" + string(evidence.Market), PositionGeneration: 1, RiskBudgetMinor: budget, PerShareRiskMinor: "10", PlannedQuantity: 14,
		PolicyDigest: "risk-policy", ConfigDigest: evidence.ModelConfigDigest, AccountCurrency: accountCurrency, QuoteCurrency: quoteCurrency, FX: fx})
	if err != nil {
		return EvaluationRequest{}, err
	}
	week := strategyflowWeek(evidence.Market, evidence.EvaluatedAt)
	state, risk := NewReservationState(), NewRiskState(plan)
	if ordinal > 1 {
		transfer, transferErr := strategyflowTransferMinor(plan, evidence.EvaluatedAt)
		if transferErr != nil {
			return EvaluationRequest{}, transferErr
		}
		held := new(big.Int).Mul(new(big.Int).Set(transfer), big.NewInt(int64(ordinal-1)))
		risk, err = mintRiskState(plan, "0", held.String())
		if err != nil {
			return EvaluationRequest{}, err
		}
		for prior := 1; prior < ordinal; prior++ {
			state, risk, err = strategyflowAdvanceLeg(plan, state, risk, week, evidence.EvaluatedAt, prior, transfer.String())
			if err != nil {
				return EvaluationRequest{}, err
			}
			evidence = strategyflowNextEvidence(evidence)
			week = strategyflowNextWeek(week, evidence.EvaluatedAt)
		}
		if code := admitExistingRisk(plan, risk); code != "" {
			return EvaluationRequest{}, &strategyflowFixtureError{code: code}
		}
	}
	state, reservation := ApplyReservation(state, authorizeReservationCommand(ReservationCommand{Action: ReservationReserve, ExpectedVersion: state.ScopeVersion(plan.CampaignID(), evidence.Market),
		CampaignID: plan.CampaignID(), MarketWeek: week, ReservationID: strategyflowReservationID(ordinal), IdempotencyKey: "reserve-" + strategyflowOrdinalKey(ordinal), PlannedOrdinal: ordinal, EvaluatedAt: evidence.EvaluatedAt}))
	if !reservation.Applied {
		return EvaluationRequest{}, &strategyflowFixtureError{code: reservation.Code}
	}
	quantity := PlannedLegQuantity(plan, LegProgress{Ordinal: ordinal}, 20)
	reservationMinor, err := strategyflowTransferMinor(plan, evidence.EvaluatedAt)
	if err != nil {
		return EvaluationRequest{}, err
	}
	cap, err := mintRiskCap(plan, riskCapInput{Authority: a066RiskCapAuthority, Version: "a066-cap-v1", QFinal: 20, ReservationQuantity: quantity,
		ReservationMinor: reservationMinor.String(), MaxStopDistanceMinor: "15", SnapshotID: "cap-1", PolicyDigest: "risk-policy", BucketSetDigest: "buckets",
		ObservedAt: evidence.EvaluatedAt.Add(-time.Minute).Format(time.RFC3339Nano), FreshUntil: evidence.FreshUntil.Format(time.RFC3339Nano), FX: plan.FX()})
	if err != nil {
		return EvaluationRequest{}, err
	}
	stop := mintStopCandidate(stopCandidateInput{PriceMinor: "90", Version: "stop-v1", Source: "structure", Policy: "stop-policy-v1", Digest: "stop-digest",
		ObservedAt: evidence.EvaluatedAt, FreshUntil: evidence.EvaluatedAt.Add(time.Hour)})
	request := EvaluationRequest{CandidateID: candidateID, Plan: plan, Evidence: evidence, Config: config, MarketWeek: week, Reservations: state,
		ReservationID: strategyflowReservationID(ordinal), Leg: LegProgress{Ordinal: ordinal}, Cap: cap, Risk: risk, StopCandidate: stop,
		EntryPriceMinor: "100", StagedTargetMinor: "1300", EntryCostsMinor: "1", EstimatedExitCostsLeviesMinor: "1", MinimumRRPPM: 1}
	request.authorization = mintDormantEvaluationAuthorization(plan, evidence)
	request.executionTerms = mintExecutionTermsPreimage(plan, evidence, request.EntryPriceMinor, request.StagedTargetMinor, request.EntryCostsMinor, request.EstimatedExitCostsLeviesMinor, request.MinimumRRPPM)
	return request, nil
}

func strategyflowAdvanceLeg(plan CampaignPlan, state ReservationState, risk RiskState, week MarketWeekEvidence, evaluatedAt time.Time, ordinal int, transfer string) (ReservationState, RiskState, error) {
	reservationID := strategyflowReservationID(ordinal)
	reserve := authorizeReservationCommand(ReservationCommand{Action: ReservationReserve, ExpectedVersion: state.ScopeVersion(plan.CampaignID(), plan.Market()), CampaignID: plan.CampaignID(), MarketWeek: week,
		ReservationID: reservationID, IdempotencyKey: "reserve-" + strategyflowOrdinalKey(ordinal), PlannedOrdinal: ordinal, EvaluatedAt: evaluatedAt})
	state, reserved := ApplyReservation(state, reserve)
	if !reserved.Applied {
		return state, risk, &strategyflowFixtureError{code: reserved.Code}
	}
	fill := authorizeReservationCommand(ReservationCommand{Action: ReservationPositiveFill, ExpectedVersion: state.ScopeVersion(plan.CampaignID(), plan.Market()), CampaignID: plan.CampaignID(), MarketWeek: week,
		ReservationID: reservationID, IdempotencyKey: "fill-" + strategyflowOrdinalKey(ordinal), PlannedOrdinal: ordinal, PositiveFillQuantity: 1, EvaluatedAt: evaluatedAt})
	combined, filled := ApplyPositiveFillAtomic(PositiveFillState{Reservations: state, Risk: risk}, plan, fill, strategyflowFillEvent(plan, ordinal, transfer, evaluatedAt))
	if !filled.Applied || filled.Code != "" {
		return state, risk, &strategyflowFixtureError{code: filled.Code}
	}
	return combined.Reservations, combined.Risk, nil
}

func strategyflowTransferMinor(plan CampaignPlan, evaluatedAt time.Time) (*big.Int, error) {
	actual, ok := actualFillRisk(plan, strategyflowFillEvent(plan, 1, "20", evaluatedAt))
	if !ok {
		return nil, &strategyflowFixtureError{code: RefusalRiskLatched}
	}
	if actual.Cmp(big.NewInt(20)) < 0 {
		return big.NewInt(20), nil
	}
	return actual, nil
}

func strategyflowFillEvent(plan CampaignPlan, ordinal int, transfer string, evaluatedAt time.Time) FillRiskEvent {
	key := strategyflowOrdinalKey(ordinal)
	return FillRiskEvent{FillID: "fill-" + key, CampaignID: plan.CampaignID(), OrderRef: "order-" + key, LegOrdinal: ordinal, Quantity: 1, TransferredReservationMinor: transfer,
		EntryPriceMinor: "100", EffectiveStopMinor: "90", EntryFeesMinor: "0", EstimatedExitFeesLeviesMinor: "0", ObservedAt: evaluatedAt.UTC().Format(time.RFC3339Nano), SourceDigest: "source-" + key, FX: plan.FX()}
}

func strategyflowReservationID(ordinal int) string {
	return "reservation-" + strategyflowOrdinalKey(ordinal)
}
func strategyflowOrdinalKey(ordinal int) string { return string(rune('0' + ordinal)) }

func strategyflowNextEvidence(evidence DisclosureEvidence) DisclosureEvidence {
	shift := 7 * 24 * time.Hour
	evidence.ObservedAt = evidence.ObservedAt.Add(shift)
	evidence.IngestedAt = evidence.IngestedAt.Add(shift)
	evidence.CutoffAt = evidence.CutoffAt.Add(shift)
	evidence.EvaluatedAt = evidence.EvaluatedAt.Add(shift)
	evidence.FreshUntil = evidence.FreshUntil.Add(shift)
	evidence.seal = evidenceSnapshotSeal(evidence)
	return evidence
}

func strategyflowNextWeek(week MarketWeekEvidence, evaluatedAt time.Time) MarketWeekEvidence {
	monday, _ := time.Parse("2006-01-02", week.SessionDate)
	monday = monday.AddDate(0, 0, 7)
	year, isoWeek := monday.ISOWeek()
	week.SessionDate = monday.Format("2006-01-02")
	week.StableIdentity = stableMarketWeekIdentity(week.Market, year, isoWeek)
	week.ObservedAt = evaluatedAt.Add(-time.Minute)
	week.FreshUntil = evaluatedAt.Add(time.Hour)
	return week
}

func strategyflowWeek(market Market, evaluatedAt time.Time) MarketWeekEvidence {
	provider, zone, stable := "XKRX_OFFICIAL", "Asia/Seoul", "KR-XKRX-2026-W32"
	if market == MarketUS {
		provider, zone, stable = "XNYS_OFFICIAL", "America/New_York", "US-XNYS-2026-W32"
	}
	return MarketWeekEvidence{Market: market, Provider: provider, Official: true, TimeZone: zone, SessionDate: "2026-08-03", StableIdentity: stable,
		CalendarGeneration: "generation-A", CalendarDigest: "calendar", ObservedAt: evaluatedAt.Add(-time.Minute), FreshUntil: evaluatedAt.Add(time.Hour)}
}

type strategyflowFixtureError struct{ code RefusalCode }

func (e *strategyflowFixtureError) Error() string {
	return "weekly strategyflow fixture: " + string(e.code)
}
