//go:build tossos_testseams

package execgw

func StrategyProtectionAuthorityForTest(market string, generation uint64, identity string) StrategyProtectionAuthority {
	return StrategyProtectionAuthority{market: market, generation: generation, identity: identity}
}

func StrategyEntryGateAuthorityForTest(generation uint64, digest string) StrategyEntryGateAuthority {
	return StrategyEntryGateAuthority{generation: generation, digest: digest}
}
