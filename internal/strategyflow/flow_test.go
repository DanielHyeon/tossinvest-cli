package strategyflow

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/candidate"
	"github.com/JungHoonGhae/tossinvest-cli/internal/continuationlane"
	"github.com/JungHoonGhae/tossinvest-cli/internal/reversallane"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategy"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyrouter"
	"github.com/JungHoonGhae/tossinvest-cli/internal/weeklyvaluelane"
)

var flowNow = time.Date(2026, 7, 31, 1, 0, 0, 0, time.UTC)

func TestPairedRegistryCoversKRUSContinuationReversalAndWeekly(t *testing.T) {
	descriptors := Descriptors()
	if err := ValidateDescriptors(descriptors); err != nil {
		t.Fatalf("paired registry: %v", err)
	}
	if len(descriptors) != 6 {
		t.Fatalf("descriptors=%d, want six paired bindings", len(descriptors))
	}
	want := map[string]bool{
		continuationlane.KRContinuationLaneID: false, continuationlane.USContinuationLaneID: false,
		reversallane.KRReversalLaneID: false, reversallane.USReversalLaneID: false,
		weeklyvaluelane.KRWeeklyLaneID: false, weeklyvaluelane.USWeeklyLaneID: false,
	}
	for _, descriptor := range descriptors {
		if _, ok := want[descriptor.LaneID]; !ok {
			t.Fatalf("unexpected descriptor: %+v", descriptor)
		}
		if descriptor.Desired != StateOff || descriptor.Effective != StateOff || descriptor.Runtime != RuntimeUnobserved {
			t.Fatalf("descriptor synthesized activation authority: %+v", descriptor)
		}
		want[descriptor.LaneID] = true
	}
	for laneID, seen := range want {
		if !seen {
			t.Fatalf("paired descriptor missing %s", laneID)
		}
	}
}

func TestApprovedCandidatesRouteAndEvaluateAllPairedBindingsWithCompleteLineage(t *testing.T) {
	for _, descriptor := range Descriptors() {
		descriptor := descriptor
		t.Run(descriptor.LaneID, func(t *testing.T) {
			approved := approvedFixture(t, descriptor.Market)
			key, err := strategyrouter.NewOwnerKey("acct", descriptor.Market, approved.Symbol(), 7)
			if err != nil {
				t.Fatal(err)
			}
			routerCalls, laneCalls := 0, 0
			registry := registryForTest(descriptor, func(LaneInput) laneEvaluation {
				laneCalls++
				return acceptedEvaluation(descriptor, key, approved)
			})
			result := evaluateWith(Request{Approved: approved, Router: strategyrouter.RouteRequest{Key: key}, Lane: inputFor(descriptor)}, func(strategyrouter.RouteRequest) strategyrouter.RouteResult {
				routerCalls++
				return strategyrouter.RouteResult{Decision: routeDecision(descriptor, key, approved)}
			}, registry)
			if result.Code != RefusalNone || result.Quantity != 1 || !result.Lineage.Valid() || !result.Lineage.Complete {
				t.Fatalf("result=%+v lineage=%+v", result, result.Lineage)
			}
			if routerCalls != 1 || laneCalls != 1 {
				t.Fatalf("calls router/lane=%d/%d", routerCalls, laneCalls)
			}
			if result.Lineage.Market != descriptor.Market || result.Lineage.CandidateLifeID != approved.CandidateLifeID() ||
				result.Lineage.RouterID != strategyrouter.RouterID || result.Lineage.RouterRelease != strategyrouter.RouterRelease || result.Lineage.LaneID != descriptor.LaneID ||
				result.Lineage.LaneVersion != descriptor.LaneVersion || result.Lineage.LaneRelease != descriptor.Release || result.Lineage.CampaignID == "" || result.Lineage.LegOrdinal == 0 ||
				result.Lineage.CandidateEvidenceDigest == "" || result.Lineage.RouterEvidenceDigest == "" || result.Lineage.LaneEvidenceDigest == "" ||
				result.Lineage.RiskBudgetDigest == "" || result.Lineage.Identity == "" {
				t.Fatalf("incomplete lineage: %+v", result.Lineage)
			}
			if result.GuardianCalls != 0 || result.BrokerCalls != 0 || result.Mutations != 0 {
				t.Fatalf("pure flow acquired authority: %+v", result)
			}
		})
	}
}

func TestRouterRefusalSkipsLaneAndUnsupportedBindingIsTyped(t *testing.T) {
	descriptor := Descriptors()[0]
	approved := approvedFixture(t, descriptor.Market)
	key, _ := strategyrouter.NewOwnerKey("acct", descriptor.Market, approved.Symbol(), 1)
	laneCalls := 0
	registry := registryForTest(descriptor, func(LaneInput) laneEvaluation { laneCalls++; return laneEvaluation{} })

	refused := evaluateWith(Request{Approved: approved, Router: strategyrouter.RouteRequest{Key: key}, Lane: inputFor(descriptor)}, func(strategyrouter.RouteRequest) strategyrouter.RouteResult {
		return strategyrouter.RouteResult{Code: strategyrouter.RefusalAmbiguous}
	}, registry)
	if refused.Code != RefusalRouter || refused.NativeCode != string(strategyrouter.RefusalAmbiguous) || laneCalls != 0 {
		t.Fatalf("router refusal=%+v calls=%d", refused, laneCalls)
	}

	unsupported := evaluateWith(Request{Approved: approved, Router: strategyrouter.RouteRequest{Key: key}, Lane: inputFor(descriptor)}, func(strategyrouter.RouteRequest) strategyrouter.RouteResult {
		decision := routeDecision(descriptor, key, approved)
		decision.LaneID = "us_unknown_lane_v1"
		return strategyrouter.RouteResult{Decision: decision}
	}, registry)
	if unsupported.Code != RefusalUnsupportedBinding || laneCalls != 0 {
		t.Fatalf("unsupported=%+v calls=%d", unsupported, laneCalls)
	}
}

func TestLaneRefusalPreservesFirstTypedCodeAndRouterLaneEvidence(t *testing.T) {
	descriptor := descriptorByID(t, reversallane.USReversalLaneID)
	approved := approvedFixture(t, strategyrouter.MarketUS)
	key, _ := strategyrouter.NewOwnerKey("acct", strategyrouter.MarketUS, approved.Symbol(), 4)
	registry := registryForTest(descriptor, func(LaneInput) laneEvaluation {
		return laneEvaluation{nativeCode: string(reversallane.RefusalStructuralMissing)}
	})
	result := evaluateWith(Request{Approved: approved, Router: strategyrouter.RouteRequest{Key: key}, Lane: inputFor(descriptor)}, func(strategyrouter.RouteRequest) strategyrouter.RouteResult {
		return strategyrouter.RouteResult{Decision: routeDecision(descriptor, key, approved)}
	}, registry)
	if result.Code != RefusalLane || result.NativeCode != string(reversallane.RefusalStructuralMissing) {
		t.Fatalf("lane refusal=%+v", result)
	}
	if result.Lineage.CandidateEvidenceDigest != approved.EvidenceDigest() || result.Lineage.LaneID != descriptor.LaneID ||
		result.Lineage.LaneVersion != descriptor.LaneVersion || result.Lineage.LaneRelease != descriptor.Release {
		t.Fatalf("refusal lost evidence/version lineage: %+v", result.Lineage)
	}
	if result.GuardianCalls != 0 || result.BrokerCalls != 0 || result.Mutations != 0 {
		t.Fatalf("refusal reached authority: %+v", result)
	}
}

func TestWrongMarketLaneInputAndForgedAcceptedLineageFailClosed(t *testing.T) {
	descriptor := descriptorByID(t, continuationlane.KRContinuationLaneID)
	approved := approvedFixture(t, strategyrouter.MarketKR)
	key, _ := strategyrouter.NewOwnerKey("acct", strategyrouter.MarketKR, approved.Symbol(), 2)
	laneCalls := 0
	registry := registryForTest(descriptor, func(LaneInput) laneEvaluation {
		laneCalls++
		forged := acceptedEvaluation(descriptor, key, approved)
		forged.lineage.CandidateID = "forged-candidate"
		return forged
	})
	route := func(strategyrouter.RouteRequest) strategyrouter.RouteResult {
		return strategyrouter.RouteResult{Decision: routeDecision(descriptor, key, approved)}
	}

	wrongInput := evaluateWith(Request{Approved: approved, Router: strategyrouter.RouteRequest{Key: key}, Lane: ContinuationUS(continuationlane.USEvaluationRequest{})}, route, registry)
	if wrongInput.Code != RefusalUnsupportedBinding || wrongInput.Lineage.LaneRelease != descriptor.Release || laneCalls != 0 {
		t.Fatalf("wrong market input=%+v calls=%d", wrongInput, laneCalls)
	}

	forged := evaluateWith(Request{Approved: approved, Router: strategyrouter.RouteRequest{Key: key}, Lane: inputFor(descriptor)}, route, registry)
	if forged.Code != RefusalLineageMismatch || laneCalls != 1 || forged.Lineage.Complete {
		t.Fatalf("forged lineage=%+v calls=%d", forged, laneCalls)
	}
}

func TestCompleteLineageSealDetectsMutation(t *testing.T) {
	descriptor := descriptorByID(t, weeklyvaluelane.USWeeklyLaneID)
	approved := approvedFixture(t, strategyrouter.MarketUS)
	key, _ := strategyrouter.NewOwnerKey("acct", strategyrouter.MarketUS, approved.Symbol(), 3)
	result := evaluateWith(Request{Approved: approved, Router: strategyrouter.RouteRequest{Key: key}, Lane: inputFor(descriptor)}, func(strategyrouter.RouteRequest) strategyrouter.RouteResult {
		return strategyrouter.RouteResult{Decision: routeDecision(descriptor, key, approved)}
	}, registryForTest(descriptor, func(LaneInput) laneEvaluation { return acceptedEvaluation(descriptor, key, approved) }))
	if !result.Lineage.Valid() {
		t.Fatalf("valid lineage rejected: %+v", result.Lineage)
	}
	result.Lineage.LegOrdinal++
	if result.Lineage.Valid() {
		t.Fatal("mutated lineage retained valid identity")
	}
}

func TestExistingOwnerRoutePinsCampaignAndAcceptsLaneConfigLineage(t *testing.T) {
	descriptor := descriptorByID(t, continuationlane.USContinuationLaneID)
	approved := approvedFixture(t, strategyrouter.MarketUS)
	key, _ := strategyrouter.NewOwnerKey("acct", strategyrouter.MarketUS, approved.Symbol(), 6)
	evaluation := acceptedEvaluation(descriptor, key, approved)
	evaluation.lineage.CampaignID = "existing-campaign"
	result := evaluateWith(Request{Approved: approved, Router: strategyrouter.RouteRequest{Key: key}, Lane: inputFor(descriptor)}, func(strategyrouter.RouteRequest) strategyrouter.RouteResult {
		return strategyrouter.RouteResult{Decision: strategyrouter.RouteDecision{Key: key, Horizon: descriptor.Horizon, LaneID: descriptor.LaneID,
			LaneVersion: descriptor.LaneVersion, CampaignID: "existing-campaign", ExistingOwner: true}}
	}, registryForTest(descriptor, func(LaneInput) laneEvaluation { return evaluation }))
	if result.Code != RefusalNone || !result.Lineage.Complete || result.Lineage.CampaignID != "existing-campaign" ||
		result.Lineage.ConfigDigest != evaluation.lineage.ConfigDigest {
		t.Fatalf("existing owner result=%+v", result)
	}

	evaluation.lineage.CampaignID = "different-campaign"
	refused := evaluateWith(Request{Approved: approved, Router: strategyrouter.RouteRequest{Key: key}, Lane: inputFor(descriptor)}, func(strategyrouter.RouteRequest) strategyrouter.RouteResult {
		return strategyrouter.RouteResult{Decision: strategyrouter.RouteDecision{Key: key, Horizon: descriptor.Horizon, LaneID: descriptor.LaneID,
			LaneVersion: descriptor.LaneVersion, CampaignID: "existing-campaign", ExistingOwner: true}}
	}, registryForTest(descriptor, func(LaneInput) laneEvaluation { return evaluation }))
	if refused.Code != RefusalLineageMismatch || refused.Lineage.Complete {
		t.Fatalf("campaign substitution was not refused: %+v", refused)
	}
}

func TestCandidateAndLaneEvidenceAreDistinctAndBothPreserved(t *testing.T) {
	descriptor := descriptorByID(t, reversallane.KRReversalLaneID)
	approved := approvedFixture(t, strategyrouter.MarketKR)
	key, _ := strategyrouter.NewOwnerKey("acct", strategyrouter.MarketKR, approved.Symbol(), 9)
	evaluation := acceptedEvaluation(descriptor, key, approved)
	evaluation.lineage.EvidenceDigest = "lane-evidence-distinct-from-candidate"
	result := evaluateWith(Request{Approved: approved, Router: strategyrouter.RouteRequest{Key: key}, Lane: inputFor(descriptor)}, func(strategyrouter.RouteRequest) strategyrouter.RouteResult {
		decision := routeDecision(descriptor, key, approved)
		decision.EvidenceDigest = evaluation.lineage.EvidenceDigest
		return strategyrouter.RouteResult{Decision: decision}
	}, registryForTest(descriptor, func(LaneInput) laneEvaluation { return evaluation }))
	if result.Code != RefusalNone || result.Lineage.CandidateEvidenceDigest != approved.EvidenceDigest() ||
		result.Lineage.RouterEvidenceDigest != evaluation.lineage.EvidenceDigest || result.Lineage.LaneEvidenceDigest != evaluation.lineage.EvidenceDigest {
		t.Fatalf("distinct evidence lineage not preserved: %+v", result)
	}
}

func TestDefaultRegistryInvokesEveryConcreteLaneEvaluator(t *testing.T) {
	registry := defaultRegistry()
	for _, descriptor := range Descriptors() {
		descriptor := descriptor
		t.Run(descriptor.LaneID, func(t *testing.T) {
			binding, ok := registry.lookup(descriptor)
			if !ok {
				t.Fatalf("default binding missing: %+v", descriptor)
			}
			got := binding.evaluate(inputFor(descriptor))
			if got.nativeCode == "" {
				t.Fatalf("zero concrete lane request was not typed-refused: %+v", got)
			}
		})
	}
}

func descriptorByID(t *testing.T, laneID string) Descriptor {
	t.Helper()
	for _, descriptor := range Descriptors() {
		if descriptor.LaneID == laneID {
			return descriptor
		}
	}
	t.Fatalf("descriptor missing: %s", laneID)
	return Descriptor{}
}

func routeDecision(descriptor Descriptor, key strategyrouter.OwnerKey, approved strategy.ApprovedSnapshot) strategyrouter.RouteDecision {
	return strategyrouter.RouteDecision{Key: key, Horizon: descriptor.Horizon, LaneID: descriptor.LaneID, LaneVersion: descriptor.LaneVersion,
		EvidenceDigest: approved.EvidenceDigest(), ConfigDigest: "config-" + descriptor.LaneID}
}

func acceptedEvaluation(descriptor Descriptor, key strategyrouter.OwnerKey, approved strategy.ApprovedSnapshot) laneEvaluation {
	price := func(value, source string) PriceProvenance {
		return PriceProvenance{priceMinor: value, source: source, version: "v1", digest: "digest-" + source, asOf: "2026-08-04T00:00:00Z", currency: map[strategyrouter.Market]string{strategyrouter.MarketKR: "KRW", strategyrouter.MarketUS: "USD"}[descriptor.Market], minorScale: map[strategyrouter.Market]int{strategyrouter.MarketKR: 0, strategyrouter.MarketUS: 2}[descriptor.Market], unitVersion: "minor-v1"}
	}
	return laneEvaluation{accepted: true, quantity: 1, entry: price("100", "entry"), stop: price("95", "stop"), target: price("120", "target"), policy: ExecutionPolicy{identity: "policy"}, lineage: laneLineage{
		AccountRef: key.AccountRef, Market: descriptor.Market, Symbol: key.Symbol, PositionGeneration: key.PositionGeneration,
		LaneID: descriptor.LaneID, LaneVersion: descriptor.LaneVersion, CandidateID: approved.CandidateLifeID(),
		EvidenceDigest: approved.EvidenceDigest(), ConfigDigest: "config-" + descriptor.LaneID,
		CampaignID: "campaign-" + strings.ToLower(string(descriptor.Market)), LegOrdinal: 1, PlannedCeiling: 1,
		RiskBudgetDigest: "risk-" + strings.ToLower(string(descriptor.Market)),
	}}
}

func inputFor(descriptor Descriptor) LaneInput {
	switch descriptor.LaneID {
	case continuationlane.KRContinuationLaneID:
		return ContinuationKR(continuationlane.KREvaluationRequest{})
	case continuationlane.USContinuationLaneID:
		return ContinuationUS(continuationlane.USEvaluationRequest{})
	case reversallane.KRReversalLaneID:
		return ReversalKR(reversallane.KREvaluationRequest{})
	case reversallane.USReversalLaneID:
		return ReversalUS(reversallane.USEvaluationRequest{})
	case weeklyvaluelane.KRWeeklyLaneID:
		return WeeklyKR(weeklyvaluelane.EvaluationRequest{})
	case weeklyvaluelane.USWeeklyLaneID:
		return WeeklyUS(weeklyvaluelane.EvaluationRequest{})
	default:
		return LaneInput{}
	}
}

func approvedFixture(t *testing.T, market strategyrouter.Market) strategy.ApprovedSnapshot {
	t.Helper()
	marketString := string(market)
	symbol := map[strategyrouter.Market]string{strategyrouter.MarketKR: "005930", strategyrouter.MarketUS: "AAPL"}[market]
	evidence := []byte("a072 synthetic candidate evidence")
	evidenceDigest := candidate.DigestEvidence(evidence)
	document := fmt.Sprintf(`{"version":"candidate-veto-a072","market":%q,"session":"regular","metrics":[{"key":"seen_late","definition":"first-sighting rank percentile","value":"80"},{"key":"extended","definition":"gain from stored first price","value":"50"},{"key":"near_high","definition":"distance below intraday high","value":"2.0"}],"sample_window":{"from":"2026-07-01T00:00:00Z","to":"2026-07-31T00:00:00Z"},"sample_count":100,"missing_rate":"0.1","evidence_digest":%q}`, marketString, evidenceDigest)
	scope := candidate.ThresholdScope{Market: marketString, Session: candidate.SessionRegular}
	setDigest, err := candidate.DigestThresholdSetDocument(strings.NewReader(document), scope)
	if err != nil {
		t.Fatal(err)
	}
	activationJSON := fmt.Sprintf(`{"version":"candidate-veto-a072","market":%q,"session":"regular","set_digest":%q,"evidence_digest":%q,"approved_at":%q,"approved_by":"a072-test-review"}`,
		marketString, setDigest, evidenceDigest, flowNow.Format(time.RFC3339))
	activation, err := candidate.LoadActivationRecord(strings.NewReader(activationJSON))
	if err != nil {
		t.Fatal(err)
	}
	set, err := candidate.LoadThresholdSet(strings.NewReader(document), evidence, activation, scope, flowNow, time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	firstSeen := flowNow.Add(-2 * time.Minute)
	value := candidate.Candidate{Key: candidate.Key{Market: marketString, Symbol: symbol}, State: candidate.StateActive, FirstSeenAt: firstSeen, LastSeenAt: firstSeen}
	at := firstSeen.Add(time.Minute)
	approved, err := candidate.AssessApprovedCandidate(candidate.VetoInputs{Candidate: value,
		Sighting:  candidate.Sighting{Measured: true, Rank: 90, RankTotal: 100},
		Expansion: candidate.Expansion{Measured: true, FirstPrice: "100", LastPrice: "110", FirstAt: firstSeen, LastAt: at},
		Range:     candidate.RangePosition{Measured: true, High: "120", Price: "100", At: at}, At: at}, set)
	if err != nil {
		t.Fatal(err)
	}
	return strategy.SealApproved(approved)
}
