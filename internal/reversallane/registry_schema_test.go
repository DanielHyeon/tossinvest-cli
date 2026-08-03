package reversallane

import (
	"bytes"
	"errors"
	"os"
	"testing"
	"time"
)

func TestRegistryShipsKRAndUSReversalTogetherDefaultOFF(t *testing.T) {
	descriptors := Descriptors()
	if err := ValidateRegistry(descriptors); err != nil {
		t.Fatal(err)
	}
	if len(descriptors) != 2 {
		t.Fatalf("descriptors=%+v", descriptors)
	}
	seen := map[string]Descriptor{}
	for _, descriptor := range descriptors {
		seen[descriptor.ID] = descriptor
		if descriptor.Release != ReversalRelease || descriptor.DesiredState != StateOff || descriptor.EffectiveState != StateOff {
			t.Fatalf("descriptor not same-release OFF: %+v", descriptor)
		}
	}
	if seen[KRReversalLaneID].Market != MarketKR || seen[USReversalLaneID].Market != MarketUS {
		t.Fatalf("market descriptors=%+v", seen)
	}
	if err := ValidateRegistry(descriptors[:1]); err == nil {
		t.Fatal("one-market reversal build passed conformance")
	}
}

func TestKRAndUSStrictSchemasAcceptExactInclusiveThresholdsTogether(t *testing.T) {
	kr := mustKREvidence(t)
	us := mustUSEvidence(t)
	krResult := EvaluateKRMetric(kr, validKRConfig())
	usResult := EvaluateUSMetric(us, validUSConfig())
	if !krResult.Accepted || krResult.AbsorptionPPM != 250000 {
		t.Fatalf("KR=%+v", krResult)
	}
	if !usResult.Accepted || usResult.DrawdownPPM != 100000 || usResult.RelativeVolumePPM != 1500000 {
		t.Fatalf("US=%+v", usResult)
	}
}

func TestStrictSchemasRejectUnknownFieldsUnitsDigestTimeDenominatorAndOverflow(t *testing.T) {
	krBody := mustFixture(t, "kr_absorption_v1.json")
	usBody := mustFixture(t, "us_dislocation_v1.json")
	if _, err := DecodeKREvidence(withUnknownField(t, krBody)); !errors.Is(err, ErrStrictSchema) {
		t.Fatalf("KR unknown field=%v", err)
	}
	if _, err := DecodeUSEvidence(withUnknownField(t, usBody)); !errors.Is(err, ErrStrictSchema) {
		t.Fatalf("US unknown field=%v", err)
	}
	duplicate := bytes.Replace(krBody, []byte(`"market": "KR"`), []byte(`"market": "KR", "market": "KR"`), 1)
	if _, err := DecodeKREvidence(duplicate); !errors.Is(err, ErrStrictSchema) {
		t.Fatalf("duplicate key accepted=%v", err)
	}

	kr := mustKREvidence(t)
	kr.Units.Notional = "WON"
	assertMetricRefusal(t, EvaluateKRMetric(kr, validKRConfig()), RefusalUnitInvalid)
	kr = mustKREvidence(t)
	kr.ConfigDigest = "other"
	assertMetricRefusal(t, EvaluateKRMetric(kr, validKRConfig()), RefusalConfigMismatch)
	kr = mustKREvidence(t)
	kr.AggressiveSellNotionalMinor = 0
	assertMetricRefusal(t, EvaluateKRMetric(kr, validKRConfig()), RefusalArithmeticInvalid)
	kr = mustKREvidence(t)
	kr.AbsorbedNotionalMinor = ^uint64(0)
	kr.AggressiveSellNotionalMinor = 1
	assertMetricRefusal(t, EvaluateKRMetric(kr, validKRConfig()), RefusalArithmeticInvalid)

	us := mustUSEvidence(t)
	us.ObservedAt = us.IngestedAt.Add(time.Nanosecond)
	assertMetricRefusal(t, EvaluateUSMetric(us, validUSConfig()), RefusalTimestampInvalid)
	us = mustUSEvidence(t)
	us.EvaluatedAt = us.FreshUntil.Add(time.Nanosecond)
	assertMetricRefusal(t, EvaluateUSMetric(us, validUSConfig()), RefusalEvidenceStale)
	us = mustUSEvidence(t)
	us.DislocationLowPriceMinor = us.ReferencePriceMinor + 1
	assertMetricRefusal(t, EvaluateUSMetric(us, validUSConfig()), RefusalArithmeticInvalid)
	us = mustUSEvidence(t)
	us.BaselineVolumeShares = 0
	assertMetricRefusal(t, EvaluateUSMetric(us, validUSConfig()), RefusalArithmeticInvalid)
	us = mustUSEvidence(t)
	us.DislocationVolumeShares = ^uint64(0)
	us.BaselineVolumeShares = 1
	assertMetricRefusal(t, EvaluateUSMetric(us, validUSConfig()), RefusalArithmeticInvalid)
}

func withUnknownField(t *testing.T, body []byte) []byte {
	t.Helper()
	trimmed := bytes.TrimSpace(body)
	if len(trimmed) == 0 || trimmed[len(trimmed)-1] != '}' {
		t.Fatal("fixture is not a JSON object")
	}
	result := append([]byte(nil), trimmed[:len(trimmed)-1]...)
	return append(result, []byte(`,"unknown":1}`)...)
}

func TestTimestampEqualityIsInclusive(t *testing.T) {
	kr := mustKREvidence(t)
	boundary := kr.EvaluatedAt
	kr.EffectiveAt = boundary
	kr.ObservedAt = boundary
	kr.IngestedAt = boundary
	kr.FreshUntil = boundary
	if got := EvaluateKRMetric(kr, validKRConfig()); !got.Accepted {
		t.Fatalf("inclusive timestamp equality refused=%+v", got)
	}
}

func TestMarketSchemaCannotSubstitutePeerMarketOrWaitForPeer(t *testing.T) {
	kr := mustKREvidence(t)
	kr.Market = MarketUS
	assertMetricRefusal(t, EvaluateKRMetric(kr, validKRConfig()), RefusalScopeMismatch)
	us := mustUSEvidence(t)
	us.Market = MarketKR
	assertMetricRefusal(t, EvaluateUSMetric(us, validUSConfig()), RefusalScopeMismatch)

	validKR := EvaluateKRMetric(mustKREvidence(t), validKRConfig())
	brokenUS := mustUSEvidence(t)
	brokenUS.SourceDigest = ""
	invalidUS := EvaluateUSMetric(brokenUS, validUSConfig())
	if !validKR.Accepted || invalidUS.Accepted || invalidUS.Refusal == "" {
		t.Fatalf("peer failure coupled: KR=%+v US=%+v", validKR, invalidUS)
	}
}

func validKRConfig() KRConfig {
	return KRConfig{Version: "kr-reversal-config-v1", SchemaVersion: "kr-absorption-v1", ConfigDigest: "kr-config-digest", ThresholdSet: "kr-reversal-thresholds-v1", MinimumAbsorptionPPM: 250000, StructuralWindow: time.Minute}
}

func validUSConfig() USConfig {
	return USConfig{Version: "us-reversal-config-v1", SchemaVersion: "us-dislocation-v1", ConfigDigest: "us-config-digest", ThresholdSet: "us-reversal-thresholds-v1", MinimumDrawdownPPM: 100000, MinimumRelativeVolumePPM: 1500000, StructuralWindow: time.Minute}
}

func mustKREvidence(t *testing.T) KREvidence {
	t.Helper()
	got, err := DecodeKREvidence(mustFixture(t, "kr_absorption_v1.json"))
	if err != nil {
		t.Fatal(err)
	}
	return got
}

func mustUSEvidence(t *testing.T) USEvidence {
	t.Helper()
	got, err := DecodeUSEvidence(mustFixture(t, "us_dislocation_v1.json"))
	if err != nil {
		t.Fatal(err)
	}
	return got
}

func mustFixture(t *testing.T, name string) []byte {
	t.Helper()
	body, err := os.ReadFile("testdata/" + name)
	if err != nil {
		t.Fatal(err)
	}
	return body
}

func assertMetricRefusal(t *testing.T, result MetricResult, want RefusalCode) {
	t.Helper()
	if result.Accepted || result.Refusal != want {
		t.Fatalf("metric=%+v want=%s", result, want)
	}
}
