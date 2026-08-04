package reversallane

import "fmt"

func Descriptors() []Descriptor {
	return []Descriptor{
		{ID: KRReversalLaneID, Version: LaneVersionV1, Market: MarketKR, Release: ReversalRelease, DesiredState: StateOff, EffectiveState: StateOff},
		{ID: USReversalLaneID, Version: LaneVersionV1, Market: MarketUS, Release: ReversalRelease, DesiredState: StateOff, EffectiveState: StateOff},
	}
}

func ValidateRegistry(descriptors []Descriptor) error {
	if len(descriptors) != 2 {
		return fmt.Errorf("reversal lanes must ship as one KR/US pair")
	}
	expected := map[string]Market{KRReversalLaneID: MarketKR, USReversalLaneID: MarketUS}
	seen := make(map[string]bool, 2)
	for _, descriptor := range descriptors {
		market, ok := expected[descriptor.ID]
		if !ok || seen[descriptor.ID] || descriptor.Market != market || descriptor.Version != LaneVersionV1 || descriptor.Release != ReversalRelease || descriptor.DesiredState != StateOff || descriptor.EffectiveState != StateOff {
			return fmt.Errorf("invalid reversal descriptor: %+v", descriptor)
		}
		seen[descriptor.ID] = true
	}
	if !seen[KRReversalLaneID] || !seen[USReversalLaneID] {
		return fmt.Errorf("both reversal lane versions are required")
	}
	return nil
}
