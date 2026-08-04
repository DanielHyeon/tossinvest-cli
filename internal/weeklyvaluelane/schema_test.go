package weeklyvaluelane

import (
	"bytes"
	"errors"
	"os"
	"testing"
	"time"
)

func TestStrictOpenDARTAndEDGARSchemasAcceptExactPointInTimeReplay(t *testing.T) {
	kr := mustKREvidence(t)
	us := mustUSEvidence(t)
	krResult := EvaluateKREvidence(kr, validKRConfig())
	usResult := EvaluateUSEvidence(us, validUSConfig())
	if !krResult.Accepted || krResult.FairValueMinor != "1100" {
		t.Fatalf("KR=%+v", krResult)
	}
	if !usResult.Accepted || usResult.FairValueMinor != "1200" {
		t.Fatalf("US=%+v", usResult)
	}
	if replay := EvaluateKREvidence(kr, validKRConfig()); replay != krResult {
		t.Fatalf("KR replay drift: first=%+v replay=%+v", krResult, replay)
	}
}

func TestStrictSchemasRejectDuplicateUnknownFutureRevisionAndBrokenPreimage(t *testing.T) {
	krBody := mustFixture(t, "kr_opendart_v1.json")
	usBody := mustFixture(t, "us_edgar_v1.json")
	duplicate := bytes.Replace(krBody, []byte(`"market": "KR"`), []byte(`"market": "US", "market": "KR"`), 1)
	if _, err := DecodeKREvidence(duplicate); !errors.Is(err, ErrStrictSchema) {
		t.Fatalf("duplicate key accepted: %v", err)
	}
	unknown := append([]byte(nil), bytes.TrimSpace(usBody)...)
	unknown = append(unknown[:len(unknown)-1], []byte(`,"unknown":true}`)...)
	if _, err := DecodeUSEvidence(unknown); !errors.Is(err, ErrStrictSchema) {
		t.Fatalf("unknown field accepted: %v", err)
	}

	kr := mustKREvidence(t)
	kr.IngestedAt = kr.CutoffAt.Add(time.Nanosecond)
	kr.seal = evidenceSnapshotSeal(kr)
	assertEvidenceRefusal(t, EvaluateKREvidence(kr, validKRConfig()), RefusalPointInTime)
	kr = mustKREvidence(t)
	kr.RevisionID = kr.SupersededRevisionID
	kr.seal = evidenceSnapshotSeal(kr)
	assertEvidenceRefusal(t, EvaluateKREvidence(kr, validKRConfig()), RefusalRevisionInvalid)
	kr = mustKREvidence(t)
	kr.DilutedShares = 0
	kr.seal = evidenceSnapshotSeal(kr)
	assertEvidenceRefusal(t, EvaluateKREvidence(kr, validKRConfig()), RefusalSchemaInvalid)

	us := mustUSEvidence(t)
	us.Currency = "KRW"
	us.seal = evidenceSnapshotSeal(us)
	assertEvidenceRefusal(t, EvaluateUSEvidence(us, validUSConfig()), RefusalUnitInvalid)
	us = mustUSEvidence(t)
	us.EquityValueMinor = "120000000000000000000000000000000000000000000000000000000000000000000000000000000"
	us.seal = evidenceSnapshotSeal(us)
	assertEvidenceRefusal(t, EvaluateUSEvidence(us, validUSConfig()), RefusalArithmeticInvalid)
	us = mustUSEvidence(t)
	us.FairValueMinor = "1201"
	us.seal = evidenceSnapshotSeal(us)
	assertEvidenceRefusal(t, EvaluateUSEvidence(us, validUSConfig()), RefusalArithmeticMismatch)
}

func validKRConfig() DisclosureConfig {
	return newDisclosureConfig(MarketKR, SourceOpenDART, KRDisclosureSchemaV1, "weekly-model-v1", "model-config-kr", "threshold-kr")
}

func validUSConfig() DisclosureConfig {
	return newDisclosureConfig(MarketUS, SourceEDGAR, USDisclosureSchemaV1, "weekly-model-v1", "model-config-us", "threshold-us")
}

func mustKREvidence(t *testing.T) DisclosureEvidence {
	t.Helper()
	evidence, err := DecodeKREvidence(mustFixture(t, "kr_opendart_v1.json"))
	if err != nil {
		t.Fatal(err)
	}
	return evidence
}

func mustUSEvidence(t *testing.T) DisclosureEvidence {
	t.Helper()
	evidence, err := DecodeUSEvidence(mustFixture(t, "us_edgar_v1.json"))
	if err != nil {
		t.Fatal(err)
	}
	return evidence
}

func mustFixture(t *testing.T, name string) []byte {
	t.Helper()
	body, err := os.ReadFile("testdata/" + name)
	if err != nil {
		t.Fatal(err)
	}
	return body
}

func assertEvidenceRefusal(t *testing.T, result EvidenceResult, want RefusalCode) {
	t.Helper()
	if result.Accepted || result.Code != want {
		t.Fatalf("result=%+v want=%s", result, want)
	}
}
