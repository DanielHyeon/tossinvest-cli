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
	decision := snapshot.Dispatch(DispatchScope{AccountID: "acct", ProfileID: "production", Market: MarketUS}, readinessNow)
	if !decision.Allowed || decision.Generation == 0 || decision.SnapshotID == "" {
		t.Fatalf("valid US dispatch=%+v", decision)
	}
	if decision.Provenance.AccountID != "acct" || decision.Provenance.ProfileID != "production" || decision.Provenance.SupervisorDigest == "" {
		t.Fatalf("scope provenance missing=%+v", decision.Provenance)
	}
	if got := snapshot.Dispatch(DispatchScope{AccountID: "other", ProfileID: "production", Market: MarketUS}, readinessNow); got.Allowed || got.Code != RefusalScopeMismatch {
		t.Fatalf("account mismatch=%+v", got)
	}
	if got := snapshot.Dispatch(DispatchScope{AccountID: "acct", ProfileID: "production", Market: MarketUS}, decision.Provenance.ExpiresAt); got.Allowed || got.Code != RefusalExpired {
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
	snapshot.kr.Provenance.ExpiresAt = snapshot.kr.Provenance.ExpiresAt.Add(time.Hour)
	if got := snapshot.Dispatch(DispatchScope{AccountID: "acct", ProfileID: "production", Market: MarketKR}, readinessNow); got.Code != RefusalStateCorrupt || got.Allowed {
		t.Fatalf("corrupt snapshot=%+v", got)
	}
}

func TestCorruptKRMarketSnapshotDoesNotInvalidateSealedUSMarket(t *testing.T) {
	fixture := newReadinessFixture(t)
	snapshot := Assess(fixture.input(map[Market]marketAssessmentInput{
		MarketKR: fixture.validMarketInput(t, MarketKR, fixture.krPrivate),
		MarketUS: fixture.validMarketInput(t, MarketUS, fixture.usPrivate),
	})).Snapshot
	snapshot.kr.Provenance.BuildDigest = digestOf("corrupt-kr")
	if got := snapshot.Dispatch(DispatchScope{AccountID: "acct", ProfileID: "production", Market: MarketKR}, readinessNow); got.Allowed || got.Code != RefusalStateCorrupt {
		t.Fatalf("corrupt KR accepted=%+v", got)
	}
	if got := snapshot.Dispatch(DispatchScope{AccountID: "acct", ProfileID: "production", Market: MarketUS}, readinessNow); !got.Allowed || got.Code != RefusalNone {
		t.Fatalf("corrupt KR contaminated US=%+v", got)
	}
}
