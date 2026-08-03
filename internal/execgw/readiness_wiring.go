package execgw

// HasProtectionReadinessProvider reports dependency wiring only. It does not
// report WIRED and cannot authorize an entry; the current sealed snapshot is
// still checked twice for every exposure-raising dispatch.
func (g *Gateway) HasProtectionReadinessProvider() bool {
	return g != nil && g.protectionReadiness != nil
}
