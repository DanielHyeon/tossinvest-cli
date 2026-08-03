package continuationlane

import "fmt"

type ActivationState string

const StateOff ActivationState = "OFF"

type Descriptor struct {
	ID             string
	Version        string
	Market         Market
	Release        string
	DesiredState   ActivationState
	EffectiveState ActivationState
}

func Descriptors() []Descriptor {
	return []Descriptor{
		{ID: KRContinuationLaneID, Version: LaneVersionV1, Market: MarketKR, Release: ContinuationRelease, DesiredState: StateOff, EffectiveState: StateOff},
		{ID: USContinuationLaneID, Version: LaneVersionV1, Market: MarketUS, Release: ContinuationRelease, DesiredState: StateOff, EffectiveState: StateOff},
	}
}

func ValidateRegistry(descriptors []Descriptor) error {
	if len(descriptors) != 2 {
		return fmt.Errorf("continuation lanes: release must contain exactly KR and US descriptors")
	}
	want := map[string]Market{KRContinuationLaneID: MarketKR, USContinuationLaneID: MarketUS}
	seen := make(map[string]bool, 2)
	for _, descriptor := range descriptors {
		market, ok := want[descriptor.ID]
		if !ok || seen[descriptor.ID] || descriptor.Market != market || descriptor.Version != LaneVersionV1 ||
			descriptor.Release != ContinuationRelease || descriptor.DesiredState != StateOff || descriptor.EffectiveState != StateOff {
			return fmt.Errorf("continuation lanes: invalid same-release descriptor %q", descriptor.ID)
		}
		seen[descriptor.ID] = true
	}
	if !seen[KRContinuationLaneID] || !seen[USContinuationLaneID] {
		return fmt.Errorf("continuation lanes: KR and US must ship together")
	}
	return nil
}

func descriptorFor(market Market) (Descriptor, bool) {
	for _, descriptor := range Descriptors() {
		if descriptor.Market == market {
			return descriptor, true
		}
	}
	return Descriptor{}, false
}
