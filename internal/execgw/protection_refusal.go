package execgw

func protectionNotWired(plan mutationPlan, cause string) *RejectedError {
	return reject(ReasonProtectionNotWired,
		"broker-resident protection readiness refused %s of %s in %s (%s); reductions, reconciliation and fills are unaffected",
		plan.kind, plan.symbol, plan.market, cause)
}
