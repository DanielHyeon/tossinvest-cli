package strategyflow

import (
	"strings"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyrouter"
)

func acceptedProjectionFixture(t *testing.T, descriptor Descriptor) Result {
	t.Helper()
	approved := approvedFixture(t, descriptor.Market)
	key, err := strategyrouter.NewOwnerKey("acct-"+strings.ToLower(string(descriptor.Market)), descriptor.Market, approved.Symbol(), 7)
	if err != nil {
		t.Fatal(err)
	}
	evaluation := acceptedEvaluation(descriptor, key, approved)
	return evaluateWith(Request{Approved: approved, Router: strategyrouter.RouteRequest{Key: key}, Lane: inputFor(descriptor)}, func(strategyrouter.RouteRequest) strategyrouter.RouteSetResult {
		return strategyrouter.RouteSetResult{Decisions: []strategyrouter.RouteDecision{routeDecision(descriptor, key, approved)}}
	}, registryForTest(descriptor, func(LaneInput) laneEvaluation { return evaluation }))
}

func TestAcceptedProjectionCoversAllFourFamiliesInBothMarketsTogether(t *testing.T) {
	descriptors := Descriptors()
	if len(descriptors) != 8 {
		t.Fatalf("descriptors=%d, want paired eight", len(descriptors))
	}
	markets := map[strategyrouter.Market]int{}
	for _, descriptor := range descriptors {
		result := acceptedProjectionFixture(t, descriptor)
		projection, err := ProjectAccepted(result)
		if err != nil {
			t.Fatalf("%s project: %v", descriptor.LaneID, err)
		}
		replay, err := ProjectAccepted(result)
		if err != nil || replay.Payload() != projection.Payload() || replay.PayloadDigest() != projection.PayloadDigest() {
			t.Fatalf("%s non-deterministic replay: first=%+v replay=%+v err=%v", descriptor.LaneID, projection, replay, err)
		}
		verified, err := VerifyAcceptedProjection(projection.Payload())
		if err != nil {
			t.Fatalf("%s verify: %v", descriptor.LaneID, err)
		}
		lineage, terms := verified.Lineage(), verified.ExecutionTerms()
		if !lineage.Valid() || !terms.Valid() || lineage.LaneID != descriptor.LaneID || lineage.LaneVersion != descriptor.LaneVersion ||
			lineage.LaneRelease != descriptor.Release || lineage.RouterID != strategyrouter.RouterID || lineage.RouterRelease != strategyrouter.RouterRelease ||
			terms.LineageIdentity() != lineage.Identity || terms.Quantity() != result.Quantity {
			t.Fatalf("%s projection lost sealed binding: lineage=%+v terms=%+v", descriptor.LaneID, lineage, terms)
		}
		markets[descriptor.Market]++
	}
	if markets[strategyrouter.MarketKR] != 4 || markets[strategyrouter.MarketUS] != 4 {
		t.Fatalf("unpaired projection matrix: %+v", markets)
	}
}

func TestAcceptedProjectionRejectsRefusedImpureAndMutatedResults(t *testing.T) {
	base := acceptedProjectionFixture(t, Descriptors()[0])
	tests := []struct {
		name   string
		mutate func(*Result)
	}{
		{"refused", func(v *Result) { v.Code = RefusalLane }},
		{"native-code", func(v *Result) { v.NativeCode = "FORGED" }},
		{"zero-quantity", func(v *Result) { v.Quantity = 0 }},
		{"guardian-call", func(v *Result) { v.GuardianCalls = 1 }},
		{"broker-call", func(v *Result) { v.BrokerCalls = 1 }},
		{"mutation", func(v *Result) { v.Mutations = 1 }},
		{"not-common-safety", func(v *Result) { v.CommonSafetyIndependent = false }},
		{"lineage", func(v *Result) { v.Lineage.Market = strategyrouter.MarketUS }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := base
			test.mutate(&result)
			if projection, err := ProjectAccepted(result); err == nil || projection.Payload() != "" {
				t.Fatalf("invalid result projected: %+v", projection)
			}
		})
	}
}

func TestVerifyAcceptedProjectionRejectsCanonicalAndSealedTamper(t *testing.T) {
	result := acceptedProjectionFixture(t, Descriptors()[0])
	projection, err := ProjectAccepted(result)
	if err != nil {
		t.Fatal(err)
	}
	payload := projection.Payload()
	tests := map[string]string{
		"unknown-field": strings.TrimSuffix(payload, "}") + `,"unknown":true}`,
		"trailing":      payload + `{}`,
		"whitespace":    " " + payload,
		"schema":        strings.Replace(payload, AcceptedProjectionSchemaVersion, "strategyflow-accepted:v9", 1),
		"router":        strings.Replace(payload, strategyrouter.RouterID, "forged-router", 1),
		"lane":          strings.Replace(payload, result.Lineage.LaneID, "forged-lane", 1),
		"lineage-seal":  strings.Replace(payload, result.Lineage.Identity, "strategy-lineage:v1:sha256:"+strings.Repeat("0", 64), 1),
		"terms-seal":    strings.Replace(payload, result.ExecutionTerms.Identity(), "strategy-execution-terms:v1:sha256:"+strings.Repeat("0", 64), 1),
		"provenance":    strings.Replace(payload, `"source":"entry"`, `"source":"forged"`, 1),
	}
	for name, tampered := range tests {
		t.Run(name, func(t *testing.T) {
			if _, err := VerifyAcceptedProjection(tampered); err == nil {
				t.Fatal("tampered projection accepted")
			}
		})
	}
}
