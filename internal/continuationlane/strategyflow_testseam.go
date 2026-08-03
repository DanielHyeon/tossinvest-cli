//go:build tossos_testseams

package continuationlane

// StrategyflowKRFixture is an accepted concrete request compiled only for
// explicit cross-package integration tests.
func StrategyflowKRFixture(candidateID, evidenceDigest string) (KREvaluationRequest, error) {
	config, err := NewKRFlowConfig("kr-th-v1", "kr-config", 100000)
	if err != nil {
		return KREvaluationRequest{}, err
	}
	evidence := KREvidence{Envelope: strategyflowEnvelope(KRFlowSchemaV1, MarketKR, "005930", config.ThresholdSetID, config.Digest, evidenceDigest),
		NotionalUnit: UnitNotionalMinor, NetFlowNotionalMinor: "1", TurnoverNotionalMinor: "10", FlowPressurePPM: 100000}
	context, err := strategyflowContext(MarketKR, KRContinuationLaneID, "005930", candidateID, config.Digest)
	if err != nil {
		return KREvaluationRequest{}, err
	}
	return KREvaluationRequest{Context: context, Evidence: evidence, Config: config}, nil
}

// StrategyflowUSFixture is an accepted concrete request compiled only for
// explicit cross-package integration tests.
func StrategyflowUSFixture(candidateID, evidenceDigest string) (USEvaluationRequest, error) {
	config, err := NewUSParticipationConfig("us-th-v1", "us-config", 200000, 10000)
	if err != nil {
		return USEvaluationRequest{}, err
	}
	evidence := USEvidence{Envelope: strategyflowEnvelope(USParticipationSchemaV1, MarketUS, "AAPL", config.ThresholdSetID, config.Digest, evidenceDigest),
		VolumeUnit: UnitShares, PriceUnit: UnitQuoteMinor, ParticipatingVolumeShares: "2", BaselineVolumeShares: "10",
		ReferencePriceMinor: "100", LastPriceMinor: "101", ParticipationPPM: 200000, PriceChangePPM: 10000}
	context, err := strategyflowContext(MarketUS, USContinuationLaneID, "AAPL", candidateID, config.Digest)
	if err != nil {
		return USEvaluationRequest{}, err
	}
	return USEvaluationRequest{Context: context, Evidence: evidence, Config: config}, nil
}

func strategyflowEnvelope(schema string, market Market, symbol, thresholdSetID, configDigest, evidenceDigest string) EvidenceEnvelope {
	return EvidenceEnvelope{SchemaVersion: schema, Market: market, Symbol: symbol, SourceRecordID: "strategyflow-record", SourceDigest: evidenceDigest,
		EffectiveAt: "2026-08-04T00:00:00Z", ObservedAt: "2026-08-04T00:00:01Z", IngestedAt: "2026-08-04T00:00:02Z",
		EvaluatedAt: "2026-08-04T00:00:03Z", FreshUntil: "2026-08-04T00:01:00Z", ThresholdSetID: thresholdSetID, ConfigDigest: configDigest}
}

func strategyflowContext(market Market, laneID, symbol, candidateID, configDigest string) (EvaluationContext, error) {
	plan, err := BuildCampaignPlan(PlanRequest{LaneID: laneID, LaneVersion: LaneVersionV1, Market: market, AccountRef: "acct", Symbol: symbol,
		CampaignID: "campaign-" + string(market), PositionGeneration: 1, RiskBudgetMinor: "1000", PerShareRiskMinor: "10", PlannedQuantity: 14,
		PolicyDigest: "risk-policy", ConfigDigest: configDigest, AccountCurrency: map[Market]string{MarketKR: "KRW", MarketUS: "USD"}[market],
		QuoteCurrency: map[Market]string{MarketKR: "KRW", MarketUS: "USD"}[market]})
	if err != nil {
		return EvaluationContext{}, err
	}
	quantity := plan.LegCeilings[0]
	cap, err := newRiskCap(plan, riskCapInput{QFinal: 20, ReservationQuantity: quantity, ReservationMinor: "20", SnapshotID: "a066-snapshot",
		PolicyDigest: plan.PolicyDigest, BucketSetDigest: "buckets", ObservedAt: "2026-08-04T00:00:00Z", FreshUntil: "2026-08-04T00:01:00Z"})
	if err != nil {
		return EvaluationContext{}, err
	}
	stop, err := newStopCandidate(stopCandidateInput{PriceMinor: "95", Source: "risk", Policy: "stop-v1", Version: "v1", Digest: "stop-digest",
		ObservedAt: "2026-08-04T00:00:02Z", FreshUntil: "2026-08-04T00:01:00Z"})
	if err != nil {
		return EvaluationContext{}, err
	}
	return EvaluationContext{Enabled: true, CandidateID: candidateID, Plan: plan, Leg: LegProgress{Ordinal: 1}, Cap: cap,
		Risk: NewRiskState(plan), SavedEffectiveStopMinor: "90", StopCandidate: stop}, nil
}
