package weeklyvaluelane

import (
	"errors"
	"sort"
)

type LaneState string

const StateOff LaneState = "OFF"

type Descriptor struct {
	ID, Version, Release         string
	Market                       Market
	Source                       DisclosureSource
	DesiredState, EffectiveState LaneState
}

func Descriptors() []Descriptor {
	return []Descriptor{
		{ID: KRWeeklyLaneID, Version: LaneVersionV1, Release: WeeklyValueRelease, Market: MarketKR, Source: SourceOpenDART, DesiredState: StateOff, EffectiveState: StateOff},
		{ID: USWeeklyLaneID, Version: LaneVersionV1, Release: WeeklyValueRelease, Market: MarketUS, Source: SourceEDGAR, DesiredState: StateOff, EffectiveState: StateOff},
	}
}

func ValidateRegistry(descriptors []Descriptor) error {
	if len(descriptors) != 2 {
		return errors.New("weekly value registry must ship the KR/US pair")
	}
	copy := append([]Descriptor(nil), descriptors...)
	sort.Slice(copy, func(i, j int) bool { return copy[i].ID < copy[j].ID })
	want := map[string]struct {
		market Market
		source DisclosureSource
	}{KRWeeklyLaneID: {MarketKR, SourceOpenDART}, USWeeklyLaneID: {MarketUS, SourceEDGAR}}
	seen := map[string]bool{}
	for _, descriptor := range copy {
		expected, ok := want[descriptor.ID]
		if !ok || seen[descriptor.ID] || descriptor.Version != LaneVersionV1 || descriptor.Release != WeeklyValueRelease || descriptor.Market != expected.market || descriptor.Source != expected.source ||
			descriptor.DesiredState != StateOff || descriptor.EffectiveState != StateOff {
			return errors.New("invalid weekly value registry descriptor")
		}
		seen[descriptor.ID] = true
	}
	return nil
}
