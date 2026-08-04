package protectionreadiness

import (
	"testing"
	"time"
)

func TestDispatchDecisionBindsExactScopeExpiryGenerationAndProvenance(t *testing.T) {
	fixture := newReadinessFixture(t)
	result := Assess(fixture.input(map[Market]marketAssessmentInput{
		MarketUS: fixture.validMarketInput(t, MarketUS, fixture.usPrivate),
	}))
	snapshot := result.Snapshot
	scope := exactDispatchScope(snapshot, MarketUS, 50)
	decision := snapshot.Dispatch(scope, readinessNow)
	if !decision.Allowed || decision.Generation == 0 || decision.SnapshotID == "" {
		t.Fatalf("valid US dispatch=%+v", decision)
	}
	if decision.Provenance.AccountID != "acct" || decision.Provenance.ProfileID != "production" || decision.Provenance.SupervisorDigest == "" {
		t.Fatalf("scope provenance missing=%+v", decision.Provenance)
	}
	wrongAccount := scope
	wrongAccount.AccountID = "other"
	if got := snapshot.Dispatch(wrongAccount, readinessNow); got.Allowed || got.Code != RefusalScopeMismatch {
		t.Fatalf("account mismatch=%+v", got)
	}
	if got := snapshot.Dispatch(scope, decision.Provenance.ExpiresAt); got.Allowed || got.Code != RefusalExpired {
		t.Fatalf("expired snapshot=%+v", got)
	}
	if got := snapshot.Dispatch(DispatchScope{AccountID: "acct", ProfileID: "production", Market: MarketKR}, readinessNow); got.Allowed || got.Code != RefusalMissingEvidence {
		t.Fatalf("missing KR contaminated by US=%+v", got)
	}
}

func TestDispatchRejectsCorruptAndFutureIssuedSnapshots(t *testing.T) {
	fixture := newReadinessFixture(t)
	snapshot := Assess(fixture.input(map[Market]marketAssessmentInput{
		MarketKR: fixture.validMarketInput(t, MarketKR, fixture.krPrivate),
	})).Snapshot
	scope := exactDispatchScope(snapshot, MarketKR, 50)
	snapshot.kr.Provenance.ExpiresAt = snapshot.kr.Provenance.ExpiresAt.Add(time.Hour)
	if got := snapshot.Dispatch(scope, readinessNow); got.Code != RefusalStateCorrupt || got.Allowed {
		t.Fatalf("corrupt snapshot=%+v", got)
	}
}

func TestCorruptKRMarketSnapshotDoesNotInvalidateSealedUSMarket(t *testing.T) {
	fixture := newReadinessFixture(t)
	snapshot := Assess(fixture.input(map[Market]marketAssessmentInput{
		MarketKR: fixture.validMarketInput(t, MarketKR, fixture.krPrivate),
		MarketUS: fixture.validMarketInput(t, MarketUS, fixture.usPrivate),
	})).Snapshot
	krScope := exactDispatchScope(snapshot, MarketKR, 50)
	usScope := exactDispatchScope(snapshot, MarketUS, 50)
	snapshot.kr.Provenance.BuildDigest = digestOf("corrupt-kr")
	if got := snapshot.Dispatch(krScope, readinessNow); got.Allowed || got.Code != RefusalStateCorrupt {
		t.Fatalf("corrupt KR accepted=%+v", got)
	}
	if got := snapshot.Dispatch(usScope, readinessNow); !got.Allowed || got.Code != RefusalNone {
		t.Fatalf("corrupt KR contaminated US=%+v", got)
	}
}

func TestDispatchRejectsQuantityOrderSessionTriggerReplaceAndCapabilitySubstitution(t *testing.T) {
	fixture := newReadinessFixture(t)
	snapshot := Assess(fixture.input(map[Market]marketAssessmentInput{
		MarketKR: fixture.validMarketInput(t, MarketKR, fixture.krPrivate),
	})).Snapshot
	base := exactDispatchScope(snapshot, MarketKR, 50)
	if got := snapshot.Dispatch(base, readinessNow); !got.Allowed {
		t.Fatalf("valid exact dispatch=%+v", got)
	}
	cases := []struct {
		name   string
		mutate func(*DispatchScope)
	}{
		{"quantity below", func(scope *DispatchScope) { scope.Quantity = 0 }},
		{"quantity above", func(scope *DispatchScope) { scope.Quantity = 101 }},
		{"order type", func(scope *DispatchScope) { scope.OrderType = "MARKET" }},
		{"session", func(scope *DispatchScope) { scope.SessionScope = "EXTENDED" }},
		{"trigger", func(scope *DispatchScope) { scope.TriggerSource = "MARK_PRICE" }},
		{"replace", func(scope *DispatchScope) { scope.ReplaceSemantics = ReplaceContinuousCoverage }},
		{"broker capability", func(scope *DispatchScope) { scope.BrokerCapabilityDigest = digestOf("other-broker") }},
		{"tool digest", func(scope *DispatchScope) { scope.ToolDigest = digestOf("other-tool") }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			scope := base
			tc.mutate(&scope)
			if got := snapshot.Dispatch(scope, readinessNow); got.Allowed || got.Code != RefusalScopeMismatch {
				t.Fatalf("substitution accepted=%+v", got)
			}
		})
	}
}

func exactDispatchScope(snapshot ReadinessSnapshot, market Market, quantity uint64) DispatchScope {
	provenance := snapshot.Verdict(market).Provenance
	return DispatchScope{
		AccountID: provenance.AccountID, ProfileID: provenance.ProfileID, Market: market,
		OrderType: provenance.OrderType, Quantity: quantity, SessionScope: provenance.SessionScope,
		TriggerSource: provenance.TriggerSource, ReplaceSemantics: provenance.ReplaceSemantics,
		BrokerCapabilityDigest: provenance.BrokerCapabilityDigest, ToolDigest: provenance.ToolDigest,
	}
}
