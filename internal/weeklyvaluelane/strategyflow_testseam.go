//go:build tossos_testseams

package weeklyvaluelane

import "time"

// StrategyflowKRFixture returns an accepted KR weekly request only in the
// explicit repository test-seam build.
func StrategyflowKRFixture(candidateID, evidenceDigest string) (EvaluationRequest, error) {
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
	return strategyflowEvaluation(candidateID, evidence, newDisclosureConfig(MarketKR, SourceOpenDART, KRDisclosureSchemaV1, "weekly-model-v1", "model-config-kr", "threshold-kr"))
}

// StrategyflowUSFixture returns an accepted US weekly request only in the
// explicit repository test-seam build.
func StrategyflowUSFixture(candidateID, evidenceDigest string) (EvaluationRequest, error) {
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
		return EvaluationRequest{}, err
	}
	evidence.EvidenceDigest = evidenceDigest
	evidence.seal = evidenceSnapshotSeal(evidence)
	return strategyflowEvaluation(candidateID, evidence, newDisclosureConfig(MarketUS, SourceEDGAR, USDisclosureSchemaV1, "weekly-model-v1", "model-config-us", "threshold-us"))
}

func strategyflowEvaluation(candidateID string, evidence DisclosureEvidence, config DisclosureConfig) (EvaluationRequest, error) {
	currency := map[Market]string{MarketKR: "KRW", MarketUS: "USD"}[evidence.Market]
	laneID := map[Market]string{MarketKR: KRWeeklyLaneID, MarketUS: USWeeklyLaneID}[evidence.Market]
	plan, err := BuildCampaignPlan(PlanRequest{LaneID: laneID, LaneVersion: LaneVersionV1, Market: evidence.Market, AccountRef: "acct", Symbol: evidence.Symbol,
		CampaignID: "campaign-" + string(evidence.Market), PositionGeneration: 1, RiskBudgetMinor: "1000", PerShareRiskMinor: "10", PlannedQuantity: 14,
		PolicyDigest: "risk-policy", ConfigDigest: evidence.ModelConfigDigest, AccountCurrency: currency, QuoteCurrency: currency})
	if err != nil {
		return EvaluationRequest{}, err
	}
	week := strategyflowWeek(evidence.Market, evidence.EvaluatedAt)
	state, reservation := ApplyReservation(NewReservationState(), authorizeReservationCommand(ReservationCommand{Action: ReservationReserve, ExpectedVersion: 0,
		CampaignID: plan.CampaignID(), MarketWeek: week, ReservationID: "reservation-1", IdempotencyKey: "reserve-1", PlannedOrdinal: 1, EvaluatedAt: evidence.EvaluatedAt}))
	if !reservation.Applied {
		return EvaluationRequest{}, &strategyflowFixtureError{code: reservation.Code}
	}
	quantity := PlannedLegQuantity(plan, LegProgress{Ordinal: 1}, 20)
	cap, err := mintRiskCap(plan, riskCapInput{Authority: a066RiskCapAuthority, Version: "a066-cap-v1", QFinal: 20, ReservationQuantity: quantity,
		ReservationMinor: "20", MaxStopDistanceMinor: "15", SnapshotID: "cap-1", PolicyDigest: "risk-policy", BucketSetDigest: "buckets",
		ObservedAt: evidence.EvaluatedAt.Add(-time.Minute).Format(time.RFC3339Nano), FreshUntil: evidence.FreshUntil.Format(time.RFC3339Nano)})
	if err != nil {
		return EvaluationRequest{}, err
	}
	stop := mintStopCandidate(stopCandidateInput{PriceMinor: "90", Version: "stop-v1", Source: "structure", Policy: "stop-policy-v1", Digest: "stop-digest",
		ObservedAt: evidence.EvaluatedAt, FreshUntil: evidence.EvaluatedAt.Add(time.Hour)})
	request := EvaluationRequest{CandidateID: candidateID, Plan: plan, Evidence: evidence, Config: config, MarketWeek: week, Reservations: state,
		ReservationID: "reservation-1", Leg: LegProgress{Ordinal: 1}, Cap: cap, Risk: NewRiskState(plan), StopCandidate: stop,
		EntryPriceMinor: "100", StagedTargetMinor: "1300", EntryCostsMinor: "1", EstimatedExitCostsLeviesMinor: "1", MinimumRRPPM: 1}
	request.authorization = mintDormantEvaluationAuthorization(plan, evidence)
	request.executionTerms = mintExecutionTermsPreimage(plan, evidence, request.EntryPriceMinor, request.StagedTargetMinor, request.EntryCostsMinor, request.EstimatedExitCostsLeviesMinor, request.MinimumRRPPM)
	return request, nil
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
