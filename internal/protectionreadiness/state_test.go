package protectionreadiness

import (
	"testing"
	"time"
)

func TestSerialAndTrustedTimeStateAreMonotonicAndPure(t *testing.T) {
	fixture := newReadinessFixture(t)
	firstInput := fixture.input(map[Market]marketAssessmentInput{MarketKR: fixture.validMarketInput(t, MarketKR, fixture.krPrivate)})
	first := Assess(firstInput)
	if first.Snapshot.Verdict(MarketKR).State != Wired || !validDurableState(first.NextState) {
		t.Fatalf("first acceptance=%+v", first)
	}

	retryInput := firstInput
	retryInput.State = first.NextState
	retry := Assess(retryInput)
	if got := retry.Snapshot.Verdict(MarketKR); got.State != Unwired || got.Code != RefusalSerialRollback {
		t.Fatalf("same serial replay accepted=%+v", got)
	}
	if firstInput.State.Serials[serialScope{AccountID: "acct", ProfileID: "production", Market: MarketKR}] != 0 {
		t.Fatal("pure assessment mutated input durable state")
	}

	rollbackInput := firstInput
	rollbackInput.State = first.NextState
	rollbackInput.Time = newTrustedTime(readinessNow.Add(-time.Second), "fixture-clock")
	rollback := Assess(rollbackInput)
	if got := rollback.Snapshot.Verdict(MarketKR); got.State != Unwired || got.Code != RefusalTrustedTimeRollback {
		t.Fatalf("clock rollback accepted=%+v", got)
	}

	unavailable := firstInput
	unavailable.Time = trustedTime{}
	if got := Assess(unavailable).Snapshot.Verdict(MarketKR); got.Code != RefusalTrustedTimeUnavailable {
		t.Fatalf("unavailable trusted time accepted=%+v", got)
	}
}

func TestLifetimeFutureExpiryAndBuildDriftFailClosed(t *testing.T) {
	fixture := newReadinessFixture(t)
	tests := []struct {
		name string
		edit func(*attestationBody)
		want RefusalCode
	}{
		{"maximum lifetime", func(body *attestationBody) { body.ExpiresAt = readinessNow.Add(3 * time.Hour).Format(time.RFC3339Nano) }, RefusalMaximumLifetime},
		{"future", func(body *attestationBody) { body.IssuedAt = readinessNow.Add(time.Second).Format(time.RFC3339Nano) }, RefusalIssuedInFuture},
		{"expired", func(body *attestationBody) { body.ExpiresAt = readinessNow.Format(time.RFC3339Nano) }, RefusalExpired},
		{"build drift", func(body *attestationBody) { body.BuildDigest = digestOf("old-build") }, RefusalScopeMismatch},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			body := fixture.body(MarketKR)
			test.edit(&body)
			input := fixture.marketInputForBody(t, body, fixture.krPrivate)
			if test.name == "build drift" {
				input.Scope.BuildDigest = digestOf("build")
				input.Supervisor, _ = newSupervisorBinding(supervisorBindingInput{AccountID: "acct", ProfileID: "production", Market: MarketKR, BuildDigest: input.Scope.BuildDigest, ComponentDigest: digestOf("supervisor-KR"), Wired: true})
			}
			got := Assess(fixture.input(map[Market]marketAssessmentInput{MarketKR: input})).Snapshot.Verdict(MarketKR)
			if got.State != Unwired || got.Code != test.want {
				t.Fatalf("got=%+v want=%s", got, test.want)
			}
		})
	}
}

func TestRevokedKRDoesNotLowerValidUSAndMissingEvidenceKeepsBothUnwired(t *testing.T) {
	fixture := newReadinessFixture(t)
	empty := Assess(fixture.input(nil))
	for _, market := range []Market{MarketKR, MarketUS} {
		if got := empty.Snapshot.Verdict(market); got.State != Unwired || got.Code != RefusalMissingEvidence {
			t.Fatalf("market=%s missing=%+v", market, got)
		}
	}

	krBody := fixture.body(MarketKR)
	krBody.KeyID = "revoked"
	result := Assess(fixture.input(map[Market]marketAssessmentInput{
		MarketKR: fixture.marketInputForBody(t, krBody, fixture.revokedPrivate),
		MarketUS: fixture.validMarketInput(t, MarketUS, fixture.usPrivate),
	}))
	if got := result.Snapshot.Verdict(MarketKR); got.Code != RefusalRevokedKey || got.State != Unwired {
		t.Fatalf("revoked KR=%+v", got)
	}
	if got := result.Snapshot.Verdict(MarketUS); got.Code != RefusalNone || got.State != Wired {
		t.Fatalf("valid US contaminated=%+v", got)
	}
}

func TestTrustedTimeFloorAdvancesWithoutAcceptedEvidenceAndDetectsLaterRollback(t *testing.T) {
	fixture := newReadinessFixture(t)
	missing := Assess(fixture.input(nil))
	if !missing.StateCommitAllowed || missing.Mutations != 1 || missing.ExternalMutations != 0 || !missing.NextState.TrustedTimeFloor.Equal(readinessNow) {
		t.Fatalf("valid trusted observation without evidence was not committed: %+v", missing)
	}

	invalidInput := fixture.input(map[Market]marketAssessmentInput{MarketKR: fixture.validMarketInput(t, MarketKR, fixture.krPrivate)})
	invalidInput.State = missing.NextState
	invalidInput.Time = newTrustedTime(readinessNow.Add(time.Minute), "fixture-clock")
	marketInput := invalidInput.Markets[MarketKR]
	corruptSignature(&marketInput.File)
	invalidInput.Markets[MarketKR] = marketInput
	invalid := Assess(invalidInput)
	if got := invalid.Snapshot.Verdict(MarketKR); got.Code != RefusalSignature || got.State != Unwired {
		t.Fatalf("invalid evidence verdict=%+v", got)
	}
	if !invalid.StateCommitAllowed || invalid.Mutations != 1 || !invalid.NextState.TrustedTimeFloor.Equal(readinessNow.Add(time.Minute)) {
		t.Fatalf("invalid evidence prevented trusted floor advancement: %+v", invalid)
	}

	rollbackInput := fixture.input(map[Market]marketAssessmentInput{MarketKR: fixture.validMarketInput(t, MarketKR, fixture.krPrivate)})
	rollbackInput.State = invalid.NextState
	rollbackInput.Time = newTrustedTime(readinessNow.Add(30*time.Second), "fixture-clock")
	rollback := Assess(rollbackInput)
	if got := rollback.Snapshot.Verdict(MarketKR); got.Code != RefusalTrustedTimeRollback || got.State != Unwired {
		t.Fatalf("rollback after invalid interval accepted=%+v", got)
	}
	if rollback.StateCommitAllowed || rollback.Mutations != 0 || rollback.NextState.seal != invalid.NextState.seal {
		t.Fatalf("rollback state became committable or changed: %+v", rollback)
	}
}

func TestCorruptDurableStateIsPreservedAndNeverAutoRepaired(t *testing.T) {
	fixture := newReadinessFixture(t)
	scope := serialScope{AccountID: "acct", ProfileID: "production", Market: MarketKR}
	original := newDurableStateWith(readinessNow.Add(-time.Minute), map[serialScope]uint64{scope: 9})
	corrupt := original
	corrupt.Serials = map[serialScope]uint64{scope: 1}
	if validDurableState(corrupt) {
		t.Fatal("fixture corruption unexpectedly valid")
	}
	body := fixture.body(MarketKR)
	body.Serial = 10
	input := fixture.input(map[Market]marketAssessmentInput{MarketKR: fixture.marketInputForBody(t, body, fixture.krPrivate)})
	input.State = corrupt
	input.Time = newTrustedTime(readinessNow.Add(time.Minute), "fixture-clock")
	result := Assess(input)
	if got := result.Snapshot.Verdict(MarketKR); got.Code != RefusalStateCorrupt || got.State != Unwired {
		t.Fatalf("corrupt durable state was not refused=%+v", got)
	}
	if result.StateCommitAllowed || result.Mutations != 0 || result.ExternalMutations != 0 {
		t.Fatalf("corrupt state became committable=%+v", result)
	}
	if result.NextState.seal != corrupt.seal || result.NextState.Serials[scope] != 1 || validDurableState(result.NextState) {
		t.Fatalf("corrupt preimage was auto-repaired/rewritten: before=%+v after=%+v", corrupt, result.NextState)
	}
}
