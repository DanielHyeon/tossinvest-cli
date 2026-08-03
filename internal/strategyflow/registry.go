package strategyflow

import (
	"fmt"

	"github.com/JungHoonGhae/tossinvest-cli/internal/continuationlane"
	"github.com/JungHoonGhae/tossinvest-cli/internal/reversallane"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyrouter"
	"github.com/JungHoonGhae/tossinvest-cli/internal/weeklyvaluelane"
)

var pairedDescriptors = [...]Descriptor{
	{Market: strategyrouter.MarketKR, Horizon: strategyrouter.HorizonShort, LaneID: continuationlane.KRContinuationLaneID, LaneVersion: continuationlane.LaneVersionV1, Release: continuationlane.ContinuationRelease, Desired: StateOff, Effective: StateOff, Runtime: RuntimeUnobserved},
	{Market: strategyrouter.MarketUS, Horizon: strategyrouter.HorizonShort, LaneID: continuationlane.USContinuationLaneID, LaneVersion: continuationlane.LaneVersionV1, Release: continuationlane.ContinuationRelease, Desired: StateOff, Effective: StateOff, Runtime: RuntimeUnobserved},
	{Market: strategyrouter.MarketKR, Horizon: strategyrouter.HorizonShort, LaneID: reversallane.KRReversalLaneID, LaneVersion: reversallane.LaneVersionV1, Release: reversallane.ReversalRelease, Desired: StateOff, Effective: StateOff, Runtime: RuntimeUnobserved},
	{Market: strategyrouter.MarketUS, Horizon: strategyrouter.HorizonShort, LaneID: reversallane.USReversalLaneID, LaneVersion: reversallane.LaneVersionV1, Release: reversallane.ReversalRelease, Desired: StateOff, Effective: StateOff, Runtime: RuntimeUnobserved},
	{Market: strategyrouter.MarketKR, Horizon: strategyrouter.HorizonWeekly, LaneID: weeklyvaluelane.KRWeeklyLaneID, LaneVersion: weeklyvaluelane.LaneVersionV1, Release: weeklyvaluelane.WeeklyValueRelease, Desired: StateOff, Effective: StateOff, Runtime: RuntimeUnobserved},
	{Market: strategyrouter.MarketUS, Horizon: strategyrouter.HorizonWeekly, LaneID: weeklyvaluelane.USWeeklyLaneID, LaneVersion: weeklyvaluelane.LaneVersionV1, Release: weeklyvaluelane.WeeklyValueRelease, Desired: StateOff, Effective: StateOff, Runtime: RuntimeUnobserved},
}

func Descriptors() []Descriptor {
	return append([]Descriptor(nil), pairedDescriptors[:]...)
}

func ValidateDescriptors(descriptors []Descriptor) error {
	if len(descriptors) != len(pairedDescriptors) {
		return fmt.Errorf("strategyflow: descriptor count %d, want %d", len(descriptors), len(pairedDescriptors))
	}
	expected := make(map[string]Descriptor, len(pairedDescriptors))
	for _, descriptor := range pairedDescriptors {
		expected[descriptor.LaneID] = descriptor
	}
	seen := make(map[string]bool, len(descriptors))
	for _, descriptor := range descriptors {
		want, ok := expected[descriptor.LaneID]
		if !ok || seen[descriptor.LaneID] || descriptor != want {
			return fmt.Errorf("strategyflow: invalid or duplicate descriptor %q", descriptor.LaneID)
		}
		seen[descriptor.LaneID] = true
	}
	return nil
}

type registryKey struct {
	market      strategyrouter.Market
	horizon     strategyrouter.Horizon
	laneID      string
	laneVersion string
}

type laneEvaluation struct {
	accepted   bool
	nativeCode string
	quantity   uint64
	lineage    laneLineage
}

type laneLineage struct {
	AccountRef         string
	Market             strategyrouter.Market
	Symbol             string
	PositionGeneration uint64
	LaneID             string
	LaneVersion        string
	CandidateID        string
	EvidenceDigest     string
	ConfigDigest       string
	CampaignID         string
	LegOrdinal         int
	PlannedCeiling     uint64
	RiskBudgetDigest   string
}

type binding struct {
	descriptor Descriptor
	evaluate   func(LaneInput) laneEvaluation
}

type registry struct{ bindings map[registryKey]binding }

func registryForTest(descriptor Descriptor, evaluate func(LaneInput) laneEvaluation) registry {
	return newRegistry([]binding{{descriptor: descriptor, evaluate: evaluate}})
}

func newRegistry(bindings []binding) registry {
	values := make(map[registryKey]binding, len(bindings))
	for _, value := range bindings {
		values[keyFor(value.descriptor)] = value
	}
	return registry{bindings: values}
}

func (r registry) lookup(descriptor Descriptor) (binding, bool) {
	value, ok := r.bindings[keyFor(descriptor)]
	return value, ok && value.descriptor == descriptor && value.evaluate != nil
}

func keyFor(descriptor Descriptor) registryKey {
	return registryKey{market: descriptor.Market, horizon: descriptor.Horizon, laneID: descriptor.LaneID, laneVersion: descriptor.LaneVersion}
}

func descriptorFor(decision strategyrouter.RouteDecision) Descriptor {
	return Descriptor{Market: decision.Key.Market, Horizon: decision.Horizon, LaneID: decision.LaneID, LaneVersion: decision.LaneVersion,
		Desired: StateOff, Effective: StateOff, Runtime: RuntimeUnobserved}
}

func canonicalDescriptor(decision strategyrouter.RouteDecision) (Descriptor, bool) {
	probe := descriptorFor(decision)
	for _, descriptor := range pairedDescriptors {
		if keyFor(descriptor) == keyFor(probe) {
			return descriptor, true
		}
	}
	return Descriptor{}, false
}

func (input LaneInput) matches(descriptor Descriptor) bool {
	want := map[string]laneKind{
		continuationlane.KRContinuationLaneID: laneContinuationKR,
		continuationlane.USContinuationLaneID: laneContinuationUS,
		reversallane.KRReversalLaneID:         laneReversalKR,
		reversallane.USReversalLaneID:         laneReversalUS,
		weeklyvaluelane.KRWeeklyLaneID:        laneWeeklyKR,
		weeklyvaluelane.USWeeklyLaneID:        laneWeeklyUS,
	}
	return want[descriptor.LaneID] != laneUnknown && input.kind == want[descriptor.LaneID]
}
