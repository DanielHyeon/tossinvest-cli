package protectionreadiness

import "errors"

type Descriptor struct {
	Market          Market
	Release         string
	State           Readiness
	SupervisorWired bool
}

func Descriptors() []Descriptor {
	return []Descriptor{
		{Market: MarketKR, Release: ReadinessRelease, State: Unwired},
		{Market: MarketUS, Release: ReadinessRelease, State: Unwired},
	}
}

func ValidateDescriptors(descriptors []Descriptor) error {
	if len(descriptors) != 2 {
		return errors.New("protectionreadiness: KR and US descriptors must ship together")
	}
	seen := make(map[Market]bool, 2)
	for _, descriptor := range descriptors {
		if !validMarket(descriptor.Market) || seen[descriptor.Market] || descriptor.Release != ReadinessRelease || descriptor.State != Unwired || descriptor.SupervisorWired {
			return errors.New("protectionreadiness: invalid dormant descriptor")
		}
		seen[descriptor.Market] = true
	}
	if !seen[MarketKR] || !seen[MarketUS] {
		return errors.New("protectionreadiness: paired market descriptor missing")
	}
	return nil
}
