//go:build tossos_testseams

package reversallane

import "time"

// StrategyflowKRFixture returns an accepted, sealed KR request only in the
// explicit repository test-seam build.
func StrategyflowKRFixture(candidateID, evidenceDigest string) (KREvaluationRequest, error) {
	evidence, err := DecodeKREvidence([]byte(`{
  "schema_version":"kr-absorption-v1","market":"KR","account_ref":"acct","symbol":"005930","position_generation":1,
  "source_record_id":"kr-record-1","source_digest":"kr-source-digest","units":{"notional":"KRW_MINOR","price":"","volume":""},
  "effective_at":"2026-08-04T00:00:00Z","observed_at":"2026-08-04T00:00:01Z","ingested_at":"2026-08-04T00:00:02Z",
  "evaluated_at":"2026-08-04T00:00:03Z","fresh_until":"2026-08-04T00:00:03Z","threshold_set":"kr-reversal-thresholds-v1",
  "structural_window_ns":60000000000,"config_digest":"kr-config-digest","absorbed_notional_minor":1,
  "aggressive_sell_notional_minor":4,"absorption_ppm":250000}`))
	if err != nil {
		return KREvaluationRequest{}, err
	}
	evidence.SourceDigest = evidenceDigest
	config := KRConfig{Version: "kr-reversal-config-v1", SchemaVersion: "kr-absorption-v1", ConfigDigest: "kr-config-digest",
		ThresholdSet: "kr-reversal-thresholds-v1", MinimumAbsorptionPPM: 250000, StructuralWindow: time.Minute}
	context, err := strategyflowContext(MarketKR, KRReversalLaneID, "005930", candidateID, config.ConfigDigest, evidence.EvaluatedAt)
	if err != nil {
		return KREvaluationRequest{}, err
	}
	terms, err := mintExecutionTermsPreimage(context.Plan, evidence.CommonEnvelope, "110", "130")
	if err != nil {
		return KREvaluationRequest{}, err
	}
	stop, err := mintStopCandidate(context.Plan, evidence.CommonEnvelope, stopCandidateInput{PriceMinor: "95", Source: "risk", Policy: "stop-v1", Version: "v1", Digest: "stop-digest", ObservedAt: evidence.EvaluatedAt, FreshUntil: evidence.EvaluatedAt.Add(time.Minute)})
	if err != nil {
		return KREvaluationRequest{}, err
	}
	context.ExecutionTerms, context.StopCandidate = terms, stop
	return KREvaluationRequest{Context: context, Evidence: evidence, Config: config}, nil
}

// StrategyflowUSFixture returns an accepted, sealed US request only in the
// explicit repository test-seam build.
func StrategyflowUSFixture(candidateID, evidenceDigest string) (USEvaluationRequest, error) {
	evidence, err := DecodeUSEvidence([]byte(`{
  "schema_version":"us-dislocation-v1","market":"US","account_ref":"acct","symbol":"AAPL","position_generation":1,
  "source_record_id":"us-record-1","source_digest":"us-source-digest","units":{"notional":"","price":"USD_MINOR","volume":"SHARES"},
  "effective_at":"2026-08-04T00:00:00Z","observed_at":"2026-08-04T00:00:01Z","ingested_at":"2026-08-04T00:00:02Z",
  "evaluated_at":"2026-08-04T00:00:03Z","fresh_until":"2026-08-04T00:00:03Z","threshold_set":"us-reversal-thresholds-v1",
  "structural_window_ns":60000000000,"config_digest":"us-config-digest","reference_price_minor":100,
  "dislocation_low_price_minor":90,"dislocation_volume_shares":150,"baseline_volume_shares":100,
  "drawdown_ppm":100000,"relative_volume_ppm":1500000}`))
	if err != nil {
		return USEvaluationRequest{}, err
	}
	evidence.SourceDigest = evidenceDigest
	config := USConfig{Version: "us-reversal-config-v1", SchemaVersion: "us-dislocation-v1", ConfigDigest: "us-config-digest",
		ThresholdSet: "us-reversal-thresholds-v1", MinimumDrawdownPPM: 100000, MinimumRelativeVolumePPM: 1500000, StructuralWindow: time.Minute}
	context, err := strategyflowContext(MarketUS, USReversalLaneID, "AAPL", candidateID, config.ConfigDigest, evidence.EvaluatedAt)
	if err != nil {
		return USEvaluationRequest{}, err
	}
	terms, err := mintExecutionTermsPreimage(context.Plan, evidence.CommonEnvelope, "110", "130")
	if err != nil {
		return USEvaluationRequest{}, err
	}
	stop, err := mintStopCandidate(context.Plan, evidence.CommonEnvelope, stopCandidateInput{PriceMinor: "95", Source: "risk", Policy: "stop-v1", Version: "v1", Digest: "stop-digest", ObservedAt: evidence.EvaluatedAt, FreshUntil: evidence.EvaluatedAt.Add(time.Minute)})
	if err != nil {
		return USEvaluationRequest{}, err
	}
	context.ExecutionTerms, context.StopCandidate = terms, stop
	return USEvaluationRequest{Context: context, Evidence: evidence, Config: config}, nil
}

func strategyflowContext(market Market, laneID, symbol, candidateID, configDigest string, evaluatedAt time.Time) (EvaluationContext, error) {
	currency := map[Market]string{MarketKR: "KRW", MarketUS: "USD"}[market]
	plan, err := BuildCampaignPlan(PlanRequest{LaneID: laneID, LaneVersion: LaneVersionV1, Market: market, AccountRef: "acct", Symbol: symbol,
		CampaignID: "campaign-" + string(market), PositionGeneration: 1, RiskBudgetMinor: "1000", PerShareRiskMinor: "10", PlannedQuantity: 14,
		PolicyDigest: "a066-policy", ConfigDigest: configDigest, AccountCurrency: currency, QuoteCurrency: currency})
	if err != nil {
		return EvaluationContext{}, err
	}
	quantity := PlannedLegQuantity(plan, LegProgress{Ordinal: 1}, RiskCap{QFinal: 20})
	cap, err := mintRiskCap(plan, RiskCap{Authority: a066RiskCapAuthority, Version: "a066-cap-v1", Market: market, QFinal: 20,
		ReservationQuantity: quantity, ReservationMinor: "20", SnapshotID: "a066-snapshot", PolicyDigest: "a066-policy", BucketSetDigest: "bucket-digest",
		Official: true, Frozen: true, ObservedAt: evaluatedAt.Add(-time.Second), FreshUntil: evaluatedAt.Add(time.Minute)})
	if err != nil {
		return EvaluationContext{}, err
	}
	return EvaluationContext{Enabled: true, CandidateID: candidateID, Plan: plan, Leg: LegProgress{Ordinal: 1}, Cap: cap, Risk: NewRiskState(plan),
		SavedEffectiveStopMinor: "90"}, nil
}
