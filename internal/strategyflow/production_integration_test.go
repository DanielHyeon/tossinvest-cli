//go:build tossos_testseams

package strategyflow

import (
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/continuationlane"
	"github.com/JungHoonGhae/tossinvest-cli/internal/reversallane"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyrouter"
	"github.com/JungHoonGhae/tossinvest-cli/internal/weeklyvaluelane"
)

func TestProductionEvaluateUsesRealRouterAndAllSixConcreteEvaluators(t *testing.T) {
	evaluatedAt := time.Date(2026, 8, 4, 0, 0, 3, 0, time.UTC)
	for _, test := range []struct {
		name, laneID, laneVersion, laneRelease, config string
		entry, stop, target                            string
		market                                         strategyrouter.Market
		horizon                                        strategyrouter.Horizon
		input                                          func(string, string) (LaneInput, error)
	}{
		{name: "KR-continuation", entry: "110", stop: "95", target: "130", market: strategyrouter.MarketKR, horizon: strategyrouter.HorizonShort, laneID: continuationlane.KRContinuationLaneID,
			laneVersion: continuationlane.LaneVersionV1, laneRelease: continuationlane.ContinuationRelease, config: "kr-config", input: func(candidateID, evidence string) (LaneInput, error) {
				request, err := continuationlane.StrategyflowKRFixture(candidateID, evidence)
				return ContinuationKR(request), err
			}},
		{name: "US-continuation", entry: "110", stop: "95", target: "130", market: strategyrouter.MarketUS, horizon: strategyrouter.HorizonShort, laneID: continuationlane.USContinuationLaneID,
			laneVersion: continuationlane.LaneVersionV1, laneRelease: continuationlane.ContinuationRelease, config: "us-config", input: func(candidateID, evidence string) (LaneInput, error) {
				request, err := continuationlane.StrategyflowUSFixture(candidateID, evidence)
				return ContinuationUS(request), err
			}},
		{name: "KR-reversal", entry: "110", stop: "95", target: "130", market: strategyrouter.MarketKR, horizon: strategyrouter.HorizonShort, laneID: reversallane.KRReversalLaneID,
			laneVersion: reversallane.LaneVersionV1, laneRelease: reversallane.ReversalRelease, config: "kr-config-digest", input: func(candidateID, evidence string) (LaneInput, error) {
				request, err := reversallane.StrategyflowKRFixture(candidateID, evidence)
				return ReversalKR(request), err
			}},
		{name: "US-reversal", entry: "110", stop: "95", target: "130", market: strategyrouter.MarketUS, horizon: strategyrouter.HorizonShort, laneID: reversallane.USReversalLaneID,
			laneVersion: reversallane.LaneVersionV1, laneRelease: reversallane.ReversalRelease, config: "us-config-digest", input: func(candidateID, evidence string) (LaneInput, error) {
				request, err := reversallane.StrategyflowUSFixture(candidateID, evidence)
				return ReversalUS(request), err
			}},
		{name: "KR-weekly", entry: "100", stop: "90", target: "1100", market: strategyrouter.MarketKR, horizon: strategyrouter.HorizonWeekly, laneID: weeklyvaluelane.KRWeeklyLaneID,
			laneVersion: weeklyvaluelane.LaneVersionV1, laneRelease: weeklyvaluelane.WeeklyValueRelease, config: "model-config-kr", input: func(candidateID, evidence string) (LaneInput, error) {
				request, err := weeklyvaluelane.StrategyflowKRFixture(candidateID, evidence)
				return WeeklyKR(request), err
			}},
		{name: "US-weekly", entry: "100", stop: "90", target: "1200", market: strategyrouter.MarketUS, horizon: strategyrouter.HorizonWeekly, laneID: weeklyvaluelane.USWeeklyLaneID,
			laneVersion: weeklyvaluelane.LaneVersionV1, laneRelease: weeklyvaluelane.WeeklyValueRelease, config: "model-config-us", input: func(candidateID, evidence string) (LaneInput, error) {
				request, err := weeklyvaluelane.StrategyflowUSFixture(candidateID, evidence)
				return WeeklyUS(request), err
			}},
	} {
		t.Run(test.name, func(t *testing.T) {
			approved := approvedFixture(t, test.market)
			key, err := strategyrouter.NewOwnerKey("acct", test.market, approved.Symbol(), 1)
			if err != nil {
				t.Fatal(err)
			}
			laneEvidence := "real-" + test.name + "-evidence"
			routerRequest, err := strategyrouter.StrategyflowRouteFixture(key, test.horizon, test.laneID,
				test.laneVersion, laneEvidence, test.config, evaluatedAt)
			if err != nil {
				t.Fatal(err)
			}
			laneInput, err := test.input(approved.CandidateLifeID(), laneEvidence)
			if err != nil {
				t.Fatal(err)
			}
			result := Evaluate(Request{Approved: approved, Router: routerRequest, Lane: laneInput})
			if result.Code != RefusalNone || result.Quantity == 0 || !result.Lineage.Complete || !result.Lineage.Valid() {
				t.Fatalf("production flow refused: %+v", result)
			}
			if !result.ExecutionTerms.Valid() || result.ExecutionTerms.Entry().PriceMinor() != test.entry ||
				result.ExecutionTerms.EffectiveStop().PriceMinor() != test.stop || result.ExecutionTerms.Target().PriceMinor() != test.target ||
				result.ExecutionTerms.Quantity() != result.Quantity || result.ExecutionTerms.LineageIdentity() != result.Lineage.Identity {
				t.Fatalf("production execution terms mismatch: %+v", result.ExecutionTerms)
			}
			currency, scale := map[strategyrouter.Market]string{strategyrouter.MarketKR: "KRW", strategyrouter.MarketUS: "USD"}[test.market], map[strategyrouter.Market]int{strategyrouter.MarketKR: 0, strategyrouter.MarketUS: 2}[test.market]
			for _, provenance := range []PriceProvenance{result.ExecutionTerms.Entry(), result.ExecutionTerms.EffectiveStop(), result.ExecutionTerms.Target()} {
				if provenance.Source() == "" || provenance.Version() == "" || provenance.Digest() == "" || provenance.AsOf() == "" || provenance.Currency() != currency || provenance.MinorScale() != scale || provenance.UnitVersion() != "minor-v1" {
					t.Fatalf("incomplete price provenance: %+v", provenance)
				}
			}
			if result.ExecutionTerms.Policy().Identity() == "" {
				t.Fatal("missing execution policy identity")
			}
			if test.horizon == strategyrouter.HorizonWeekly {
				policy := result.ExecutionTerms.Policy()
				if policy.StagedTargetMinor() != "1300" || policy.FairValueMinor() != test.target || policy.EntryCostsMinor() != "1" || policy.ExitCostsMinor() != "1" || policy.MinimumRRPPM() != 1 || policy.DecisionDigest() == "" || policy.CalendarDigest() == "" || policy.CapSnapshotID() == "" || result.Lineage.ExecutionPolicyDigest != policy.Identity() {
					t.Fatalf("weekly full RR preimage lost: %+v", policy)
				}
			}
			if result.Lineage.Market != test.market || result.Lineage.RouterRelease != strategyrouter.RouterRelease ||
				result.Lineage.Horizon != test.horizon || result.Lineage.LaneID != test.laneID || result.Lineage.LaneVersion != test.laneVersion ||
				result.Lineage.LaneRelease != test.laneRelease ||
				result.Lineage.RouterEvidenceDigest != laneEvidence || result.Lineage.LaneEvidenceDigest != laneEvidence ||
				result.Lineage.CandidateEvidenceDigest != approved.EvidenceDigest() || result.Lineage.ConfigDigest != test.config ||
				result.Lineage.CampaignID == "" || result.Lineage.LegOrdinal != 1 || result.Lineage.RiskBudgetDigest == "" {
				t.Fatalf("production lineage mismatch: %+v", result.Lineage)
			}
			if result.GuardianCalls != 0 || result.BrokerCalls != 0 || result.Mutations != 0 {
				t.Fatalf("production pure flow acquired authority: %+v", result)
			}
			descriptor := descriptorByID(t, test.laneID)
			if descriptor.Release != test.laneRelease {
				t.Fatalf("lane release mismatch: %+v", descriptor)
			}
		})
	}
}
