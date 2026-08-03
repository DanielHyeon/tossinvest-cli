package strategyruntime

import "errors"

type Descriptor struct {
	Market    Market
	Release   string
	Desired   EntryState
	Effective EntryState
	Runtime   RuntimeState
}

func Descriptors() []Descriptor {
	return []Descriptor{
		{Market: MarketKR, Release: RuntimeRelease, Desired: EntryOff, Effective: EntryOff, Runtime: RuntimeUnobserved},
		{Market: MarketUS, Release: RuntimeRelease, Desired: EntryOff, Effective: EntryOff, Runtime: RuntimeUnobserved},
	}
}

func ValidateDescriptors(descriptors []Descriptor) error {
	if len(descriptors) != 2 {
		return errors.New("strategyruntime: KR and US descriptors must ship together")
	}
	seen := make(map[Market]bool, 2)
	for _, descriptor := range descriptors {
		if !validMarket(descriptor.Market) || seen[descriptor.Market] || descriptor.Release != RuntimeRelease || descriptor.Desired != EntryOff || descriptor.Effective != EntryOff || descriptor.Runtime != RuntimeUnobserved {
			return errors.New("strategyruntime: invalid dormant descriptor")
		}
		seen[descriptor.Market] = true
	}
	if !seen[MarketKR] || !seen[MarketUS] {
		return errors.New("strategyruntime: paired market descriptor missing")
	}
	return nil
}
