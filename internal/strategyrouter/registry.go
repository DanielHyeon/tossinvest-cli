package strategyrouter

import "errors"

const (
	RouterID      = "kr-us-horizon-router"
	RouterRelease = "a070-kr-us-horizon-router-v1"
)

type Descriptor struct {
	Market    Market
	RouterID  string
	Release   string
	Desired   DesiredState
	Effective DesiredState
	Runtime   RuntimeState
}

func Descriptors() []Descriptor {
	return []Descriptor{
		{Market: MarketKR, RouterID: RouterID, Release: RouterRelease, Desired: StateOff, Effective: StateOff, Runtime: RuntimeUnobserved},
		{Market: MarketUS, RouterID: RouterID, Release: RouterRelease, Desired: StateOff, Effective: StateOff, Runtime: RuntimeUnobserved},
	}
}

func ValidateDescriptors(descriptors []Descriptor) error {
	if len(descriptors) != 2 {
		return errors.New("strategyrouter: KR and US descriptors must ship together")
	}
	seen := make(map[Market]bool, 2)
	for _, descriptor := range descriptors {
		if !validMarket(descriptor.Market) || seen[descriptor.Market] || descriptor.RouterID != RouterID || descriptor.Release != RouterRelease || descriptor.Desired != StateOff || descriptor.Effective != StateOff || descriptor.Runtime != RuntimeUnobserved {
			return errors.New("strategyrouter: invalid dormant descriptor")
		}
		seen[descriptor.Market] = true
	}
	if !seen[MarketKR] || !seen[MarketUS] {
		return errors.New("strategyrouter: paired market descriptor missing")
	}
	return nil
}
