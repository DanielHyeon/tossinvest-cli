package continuationlane

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestRegistryShipsKRAndUSTogetherDefaultOFF(t *testing.T) {
	descriptors := Descriptors()
	if err := ValidateRegistry(descriptors); err != nil {
		t.Fatalf("registry: %v", err)
	}
	if len(descriptors) != 2 {
		t.Fatalf("descriptor count=%d, want 2", len(descriptors))
	}
	seen := map[string]Descriptor{}
	for _, descriptor := range descriptors {
		seen[descriptor.ID] = descriptor
		if descriptor.Release != ContinuationRelease || descriptor.DesiredState != StateOff || descriptor.EffectiveState != StateOff {
			t.Fatalf("descriptor not same-release default OFF: %+v", descriptor)
		}
	}
	if seen[KRContinuationLaneID].Market != MarketKR || seen[USContinuationLaneID].Market != MarketUS {
		t.Fatalf("market-scoped descriptors=%+v", seen)
	}
	if err := ValidateRegistry(descriptors[:1]); err == nil {
		t.Fatal("one-market build passed same-release conformance")
	}
}

func TestStrictDecodersRejectDuplicateJSONKeys(t *testing.T) {
	kr, _ := validKRFixture()
	krData, err := json.Marshal(kr)
	if err != nil {
		t.Fatal(err)
	}
	krDuplicate := strings.Replace(string(krData), `"notional_unit":"NOTIONAL_MINOR"`, `"notional_unit":"NOTIONAL_MINOR","notional_unit":"NOTIONAL_MINOR"`, 1)
	if _, err := DecodeKREvidence([]byte(krDuplicate)); err == nil {
		t.Fatal("KR duplicate JSON key accepted")
	}

	us, _ := validUSFixture()
	usData, err := json.Marshal(us)
	if err != nil {
		t.Fatal(err)
	}
	usDuplicate := strings.Replace(string(usData), `"source_digest":"source-digest"`, `"source_digest":"source-digest","source_digest":"other"`, 1)
	if usDuplicate == string(usData) {
		t.Fatal("US duplicate-key fixture did not target source_digest")
	}
	if _, err := DecodeUSEvidence([]byte(usDuplicate)); err == nil {
		t.Fatal("US nested duplicate JSON key accepted")
	}
}

func TestKRAndUSStrictSchemasAcceptExactInclusiveBoundariesTogether(t *testing.T) {
	kr, krConfig := validKRFixture()
	us, usConfig := validUSFixture()

	krResult := EvaluateKRFlow(kr, krConfig)
	usResult := EvaluateUSParticipation(us, usConfig)
	if !krResult.Accepted || krResult.FlowPressurePPM != krConfig.MinimumFlowPressurePPM {
		t.Fatalf("KR boundary=%+v", krResult)
	}
	if !usResult.Accepted || usResult.ParticipationPPM != usConfig.MinimumParticipationPPM || usResult.PriceChangePPM != usConfig.MinimumPriceChangePPM {
		t.Fatalf("US boundary=%+v", usResult)
	}
}

func TestKRFlowUsesSignedTruncationTowardZero(t *testing.T) {
	metric, err := signedPPM("-1", "3")
	if err != nil || metric != -333333 {
		t.Fatalf("signed truncation=%d err=%v", metric, err)
	}
}

func TestThresholdConfigsAreSealedAndRejectUnsafeRanges(t *testing.T) {
	if _, err := NewKRFlowConfig("kr-th-v1", "kr-config", -1); err == nil {
		t.Fatal("negative KR continuation threshold accepted")
	}
	if _, err := NewUSParticipationConfig("us-th-v1", "us-config", -1, 0); err == nil {
		t.Fatal("negative US participation threshold accepted")
	}
	if _, err := NewUSParticipationConfig("us-th-v1", "us-config", 0, -1); err == nil {
		t.Fatal("negative US price continuation threshold accepted")
	}
	evidence, config := validKRFixture()
	config.MinimumFlowPressurePPM++
	if got := EvaluateKRFlow(evidence, config); got.Code != RefusalConfigInvalid {
		t.Fatalf("mutated sealed config accepted=%+v", got)
	}
}

func TestStrictSchemasRejectUnknownFieldsUnitsDigestsTimeAndOverflow(t *testing.T) {
	krJSON := `{"envelope":{"schema_version":"kr-flow-v1","market":"KR","symbol":"005930","source_record_id":"r","source_digest":"d","effective_at":"2026-08-04T00:00:00Z","observed_at":"2026-08-04T00:00:01Z","ingested_at":"2026-08-04T00:00:02Z","evaluated_at":"2026-08-04T00:00:03Z","fresh_until":"2026-08-04T00:01:00Z","threshold_set_id":"kr-th-v1","config_digest":"kr-config"},"notional_unit":"NOTIONAL_MINOR","net_flow_notional_minor":"10","turnover_notional_minor":"100","flow_pressure_ppm":100000,"unknown":true}`
	krJSON = strings.ReplaceAll(krJSON, `\"`, `"`)
	if _, err := DecodeKREvidence([]byte(krJSON)); err == nil {
		t.Fatal("KR unknown field accepted")
	}
	usJSON := `{"envelope":{"schema_version":"us-participation-v1","market":"US","symbol":"AAPL","source_record_id":"r","source_digest":"d","effective_at":"2026-08-04T00:00:00Z","observed_at":"2026-08-04T00:00:01Z","ingested_at":"2026-08-04T00:00:02Z","evaluated_at":"2026-08-04T00:00:03Z","fresh_until":"2026-08-04T00:01:00Z","threshold_set_id":"us-th-v1","config_digest":"us-config"},"volume_unit":"SHARES","price_unit":"QUOTE_MINOR","participating_volume_shares":"2","baseline_volume_shares":"10","reference_price_minor":"100","last_price_minor":"101","participation_ppm":200000,"price_change_ppm":10000,"unknown":true}`
	usJSON = strings.ReplaceAll(usJSON, `\"`, `"`)
	if _, err := DecodeUSEvidence([]byte(usJSON)); err == nil {
		t.Fatal("US unknown field accepted")
	}

	tests := []struct {
		name string
		edit func(*KREvidence, *KRFlowConfig)
		code RefusalCode
	}{
		{"unit", func(e *KREvidence, _ *KRFlowConfig) { e.NotionalUnit = "WON" }, RefusalUnitMismatch},
		{"digest", func(e *KREvidence, _ *KRFlowConfig) { e.Envelope.ConfigDigest = "other" }, RefusalDigestMismatch},
		{"time order", func(e *KREvidence, _ *KRFlowConfig) { e.Envelope.IngestedAt = "2026-08-03T23:59:00Z" }, RefusalInvalidTime},
		{"stale", func(e *KREvidence, _ *KRFlowConfig) { e.Envelope.FreshUntil = "2026-08-04T00:00:02Z" }, RefusalStaleEvidence},
		{"preimage mismatch", func(e *KREvidence, _ *KRFlowConfig) { e.FlowPressurePPM++ }, RefusalArithmeticMismatch},
		{"overflow", func(e *KREvidence, _ *KRFlowConfig) { e.NetFlowNotionalMinor = "1" + strings.Repeat("0", 100) }, RefusalArithmeticOverflow},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			evidence, config := validKRFixture()
			tt.edit(&evidence, &config)
			got := EvaluateKRFlow(evidence, config)
			if got.Accepted || got.Code != tt.code {
				t.Fatalf("result=%+v, want %s", got, tt.code)
			}
		})
	}
}

func TestUSStrictSchemaRejectsItsOwnUnitsTimeDigestAndOverflow(t *testing.T) {
	tests := []struct {
		name string
		edit func(*USEvidence, *USParticipationConfig)
		code RefusalCode
	}{
		{"volume unit", func(e *USEvidence, _ *USParticipationConfig) { e.VolumeUnit = "LOTS" }, RefusalUnitMismatch},
		{"price unit", func(e *USEvidence, _ *USParticipationConfig) { e.PriceUnit = "DOLLARS" }, RefusalUnitMismatch},
		{"config digest", func(e *USEvidence, _ *USParticipationConfig) { e.Envelope.ConfigDigest = "kr-config" }, RefusalDigestMismatch},
		{"time order", func(e *USEvidence, _ *USParticipationConfig) { e.Envelope.ObservedAt = "2026-08-03T23:59:00Z" }, RefusalInvalidTime},
		{"overflow", func(e *USEvidence, _ *USParticipationConfig) {
			e.ParticipatingVolumeShares = "1" + strings.Repeat("0", 100)
		}, RefusalArithmeticOverflow},
		{"metric mismatch", func(e *USEvidence, _ *USParticipationConfig) { e.PriceChangePPM-- }, RefusalArithmeticMismatch},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			evidence, config := validUSFixture()
			tt.edit(&evidence, &config)
			got := EvaluateUSParticipation(evidence, config)
			if got.Accepted || got.Code != tt.code {
				t.Fatalf("result=%+v want=%s", got, tt.code)
			}
		})
	}
}

func TestSameSymbolTextStillProducesMarketScopedLineage(t *testing.T) {
	krEvidence, krConfig := validKRFixture()
	usEvidence, usConfig := validUSFixture()
	krEvidence.Envelope.Symbol = "SAME"
	usEvidence.Envelope.Symbol = "SAME"
	krPlan := mustPlan(t, MarketKR, KRContinuationLaneID, "KRW", "KRW", nil, 14, "1000")
	usPlan := mustPlan(t, MarketUS, USContinuationLaneID, "USD", "USD", nil, 14, "1000")
	krPlan.Symbol, krPlan.RiskBudgetDigest = "SAME", ""
	krPlan.RiskBudgetDigest = planDigest(krPlan)
	usPlan.Symbol, usPlan.RiskBudgetDigest = "SAME", ""
	usPlan.RiskBudgetDigest = planDigest(usPlan)
	kr := EvaluateKR(validKREvaluation(t, krPlan, krEvidence, krConfig))
	us := EvaluateUS(validUSEvaluation(t, usPlan, usEvidence, usConfig))
	if kr.Kind != OutcomeDecision || us.Kind != OutcomeDecision || kr.Lineage.Market == us.Lineage.Market || kr.Lineage.LaneID == us.Lineage.LaneID || kr.Lineage.CampaignID == us.Lineage.CampaignID {
		t.Fatalf("cross-market lineage contamination: KR=%+v US=%+v", kr, us)
	}
}

func TestMarketFailuresAndOFFStateDoNotGatePeerEvaluation(t *testing.T) {
	kr, krConfig := validKRFixture()
	us, usConfig := validUSFixture()
	kr.Envelope.FreshUntil = "2026-08-04T00:00:02Z"
	if got := EvaluateKRFlow(kr, krConfig); got.Code != RefusalStaleEvidence {
		t.Fatalf("KR=%+v", got)
	}
	if got := EvaluateUSParticipation(us, usConfig); !got.Accepted {
		t.Fatalf("US was gated by KR failure: %+v", got)
	}

	plan := mustPlan(t, MarketUS, USContinuationLaneID, "USD", "USD", nil, 14, "1000")
	request := validUSEvaluation(t, plan, us, usConfig)
	request.Context.Enabled = true
	if got := EvaluateUS(request); got.Kind != OutcomeDecision || got.Quantity == 0 {
		t.Fatalf("US evaluation did not proceed while peer remains registry OFF: %+v", got)
	}
}

func validEnvelope(schema string, market Market, symbol, thresholds, digest string) EvidenceEnvelope {
	return EvidenceEnvelope{
		SchemaVersion: schema, Market: market, Symbol: symbol, SourceRecordID: "record-1", SourceDigest: "source-digest",
		EffectiveAt: "2026-08-04T00:00:00Z", ObservedAt: "2026-08-04T00:00:01Z", IngestedAt: "2026-08-04T00:00:02Z",
		EvaluatedAt: "2026-08-04T00:00:03Z", FreshUntil: "2026-08-04T00:01:00Z", ThresholdSetID: thresholds, ConfigDigest: digest,
	}
}

func validKRFixture() (KREvidence, KRFlowConfig) {
	config, err := NewKRFlowConfig("kr-th-v1", "kr-config", 100000)
	if err != nil {
		panic(err)
	}
	evidence := KREvidence{Envelope: validEnvelope(KRFlowSchemaV1, MarketKR, "005930", config.ThresholdSetID, config.Digest), NotionalUnit: UnitNotionalMinor, NetFlowNotionalMinor: "1", TurnoverNotionalMinor: "10", FlowPressurePPM: 100000}
	return evidence, config
}

func validUSFixture() (USEvidence, USParticipationConfig) {
	config, err := NewUSParticipationConfig("us-th-v1", "us-config", 200000, 10000)
	if err != nil {
		panic(err)
	}
	evidence := USEvidence{Envelope: validEnvelope(USParticipationSchemaV1, MarketUS, "AAPL", config.ThresholdSetID, config.Digest), VolumeUnit: UnitShares, PriceUnit: UnitQuoteMinor, ParticipatingVolumeShares: "2", BaselineVolumeShares: "10", ReferencePriceMinor: "100", LastPriceMinor: "101", ParticipationPPM: 200000, PriceChangePPM: 10000}
	return evidence, config
}
