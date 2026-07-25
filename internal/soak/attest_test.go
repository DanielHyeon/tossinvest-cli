package soak_test

// attest_test.go covers the judgement: when the soak has proven enough to write
// a capability attestation, and — more importantly — when it has not.
//
// engine-safety interlocks live order placement on this file. Every test here is
// about refusing to write one, because the failure that costs money is the
// attestation that claims something nobody measured.

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/attest"
	"github.com/JungHoonGhae/tossinvest-cli/internal/soak"
)

func criteria() soak.Criteria { return soak.DefaultCriteria() }

// evaluate judges a summary from an instant just after its last cycle, so the
// staleness rule only fires in the test that is about it.
func evaluate(s soak.Summary) (bool, []string) {
	return s.Evaluate(s.LastAt.Add(time.Hour), criteria())
}

func TestDefaultCriteriaRequiresThreeDays(t *testing.T) {
	c := criteria()
	if c.MinConsecutiveDays != 3 {
		t.Errorf("MinConsecutiveDays = %d, want 3 (execution-verification spec)", c.MinConsecutiveDays)
	}
	if c.Validity <= 0 {
		t.Error("Validity must be positive; an attestation with no expiry is trusted forever")
	}
	if c.MaxRecordAge <= 0 {
		t.Error("MaxRecordAge must be positive; otherwise a soak from last spring can be attested today")
	}
	if len(c.RequiredEndpoints) == 0 {
		t.Error("RequiredEndpoints is empty")
	}
}

func TestEvaluateAcceptsACompletedSoak(t *testing.T) {
	s := soak.Summarize(threeCleanDays())
	ok, reasons := evaluate(s)
	if !ok {
		t.Fatalf("a clean three-day soak was refused: %v", reasons)
	}
	if len(reasons) != 0 {
		t.Errorf("reasons = %v on an accepted soak", reasons)
	}
}

func TestEvaluateRefusesAShortStreak(t *testing.T) {
	s := soak.Summarize(threeCleanDays()[:2]) // one day
	ok, reasons := evaluate(s)
	if ok {
		t.Fatal("a one-day soak was accepted")
	}
	if !mentions(reasons, "consecutive") {
		t.Errorf("reasons = %v, want one naming the day count", reasons)
	}
}

func TestEvaluateRefusesAnEmptyRecord(t *testing.T) {
	ok, reasons := evaluate(soak.Summarize(nil))
	if ok {
		t.Fatal("an empty record was accepted")
	}
	if len(reasons) == 0 {
		t.Fatal("a refusal with no reason tells an operator nothing")
	}
}

// TestEvaluateRefusesWhenARequiredEndpointNeverSucceeded. The attestation's
// endpoint set is what the engine's interlock checks against; an endpoint that
// never worked must never appear in it, and a soak that never proved one has not
// finished.
func TestEvaluateRefusesWhenARequiredEndpointNeverSucceeded(t *testing.T) {
	cycles := threeCleanDays()
	for i := range cycles {
		setEndpoint(&cycles[i], soak.EndpointOrderByID, func(e *soak.EndpointResult) {
			e.OK = false
			e.Skipped = true
			e.SkipReason = "the account has no orders"
		})
	}

	ok, reasons := evaluate(soak.Summarize(cycles))
	if ok {
		t.Fatal("a soak that never read an order by id was accepted")
	}
	if !mentions(reasons, soak.EndpointOrderByID) {
		t.Errorf("reasons = %v, want one naming %s", reasons, soak.EndpointOrderByID)
	}
}

// TestEvaluateRefusesWhenTheRecordCoversTwoAccounts. An attestation names one
// account; a record spanning two describes neither.
func TestEvaluateRefusesWhenTheRecordCoversTwoAccounts(t *testing.T) {
	cycles := threeCleanDays()
	cycles[3].AccountRef = "987-65-432109"

	ok, reasons := evaluate(soak.Summarize(cycles))
	if ok {
		t.Fatal("a record covering two accounts was accepted")
	}
	if !mentions(reasons, "account") {
		t.Errorf("reasons = %v, want one naming the account conflict", reasons)
	}
}

// TestEvaluateRefusesWhenNoRefreshWasObserved. Three days of a token that never
// expired proves the token was long-lived, not that anything renewed it.
func TestEvaluateRefusesWhenNoRefreshWasObserved(t *testing.T) {
	cycles := threeCleanDays()
	for i := range cycles {
		cycles[i].Credential.Refreshed = false
	}

	ok, reasons := evaluate(soak.Summarize(cycles))
	if ok {
		t.Fatal("a soak that never saw a token renew was accepted as proof of unattended refresh")
	}
	if !mentions(reasons, "refresh") {
		t.Errorf("reasons = %v, want one naming the missing refresh", reasons)
	}
}

// TestEvaluateRefusesWhenTheTokenExpiryWasNeverObservable. If the tool cannot
// see the expiry at all it cannot make the claim; the right answer is to say so,
// not to assume the best.
func TestEvaluateRefusesWhenTheTokenExpiryWasNeverObservable(t *testing.T) {
	cycles := threeCleanDays()
	for i := range cycles {
		cycles[i].Credential.Observed = false
		cycles[i].Credential.Refreshed = false
	}

	ok, reasons := evaluate(soak.Summarize(cycles))
	if ok {
		t.Fatal("a soak that never read the token expiry was accepted")
	}
	if !mentions(reasons, "expiry") {
		t.Errorf("reasons = %v, want one naming the unreadable expiry", reasons)
	}
}

// TestEvaluateRefusesACompletenessFailureInsideTheWindow. Duplicated or
// unreachable order pages mean the engine cannot see its own exposure.
func TestEvaluateRefusesACompletenessFailureInsideTheWindow(t *testing.T) {
	cycles := threeCleanDays()
	cycles[4].Completeness.OK = false
	cycles[4].Completeness.Detail = "order o7 was open but absent from the list"

	ok, reasons := evaluate(soak.Summarize(cycles))
	if ok {
		t.Fatal("a completeness failure inside the window was accepted")
	}
	if !mentions(reasons, "completeness") {
		t.Errorf("reasons = %v, want one naming the completeness failure", reasons)
	}
}

// TestEvaluateRefusesAStaleRecord. The streak is a property of the record and
// the expiry is a property of the moment the file is written, so without this a
// soak that finished months ago would mint a freshly-dated attestation today.
func TestEvaluateRefusesAStaleRecord(t *testing.T) {
	s := soak.Summarize(threeCleanDays())
	ok, reasons := s.Evaluate(s.LastAt.Add(30*24*time.Hour), criteria())
	if ok {
		t.Fatal("a month-old record was accepted")
	}
	if !mentions(reasons, "old") {
		t.Errorf("reasons = %v, want one naming the record's age", reasons)
	}
}

// --- building the file ------------------------------------------------------

func TestBuildAttestationPassesTheEngineInterlock(t *testing.T) {
	s := soak.Summarize(threeCleanDays())
	now := soakStart.AddDate(0, 0, 3)

	a, err := soak.BuildAttestation(s, criteria(), now, "tester", "")
	if err != nil {
		t.Fatalf("BuildAttestation: %v", err)
	}

	if a.FormatVersion != attest.FormatVersion {
		t.Errorf("FormatVersion = %d, want %d", a.FormatVersion, attest.FormatVersion)
	}
	if a.SoakDays != 3 {
		t.Errorf("SoakDays = %d, want 3", a.SoakDays)
	}
	if !a.IssuedAt.Equal(now) {
		t.Errorf("IssuedAt = %s, want %s", a.IssuedAt, now)
	}
	if !a.ExpiresAt.After(now) {
		t.Errorf("ExpiresAt = %s, want it in the future", a.ExpiresAt)
	}
	// The engine's own check, run against exactly what the soak proved.
	if err := a.Verify(now, "123-45-678901", soak.RequiredEndpoints()); err != nil {
		t.Fatalf("the attestation the soak produced fails its own endpoint set: %v", err)
	}
}

// TestBuildAttestationNeverClaimsAnEndpointItDidNotExercise is the safety
// property this whole change exists for: the read-only tool cannot put a
// mutation into the set the live-trading gate reads.
func TestBuildAttestationNeverClaimsAnEndpointItDidNotExercise(t *testing.T) {
	cycles := threeCleanDays()
	// Somebody hand-edits the record — or a future refactor widens the survey —
	// and a POST turns up among the results.
	for i := range cycles {
		cycles[i].Endpoints = append(cycles[i].Endpoints, soak.EndpointResult{
			Endpoint: "POST /api/v1/orders", OK: true, Requests: 1,
		})
	}

	_, err := soak.BuildAttestation(soak.Summarize(cycles), criteria(), soakStart.AddDate(0, 0, 3), "tester", "")
	if err == nil {
		t.Fatal("BuildAttestation put a mutation endpoint into a read-only attestation")
	}
	if !strings.Contains(err.Error(), "POST") {
		t.Errorf("err = %v, want it to name the mutation", err)
	}
}

// TestBuildAttestationRefusesAnIncompleteSoak, with the reasons attached so the
// caller can print them.
func TestBuildAttestationRefusesAnIncompleteSoak(t *testing.T) {
	s := soak.Summarize(threeCleanDays()[:2])
	_, err := soak.BuildAttestation(s, criteria(), soakStart.AddDate(0, 0, 1), "tester", "")
	if err == nil {
		t.Fatal("BuildAttestation wrote an attestation for a one-day soak")
	}
	if !errors.Is(err, soak.ErrIncomplete) {
		t.Errorf("err = %v, want it to wrap ErrIncomplete", err)
	}
}

// TestBuildAttestationCarriesTheMeasuredRate. Task 1.3 turns this number into
// the retry matrix; an attestation that dropped it would lose the measurement.
func TestBuildAttestationCarriesTheMeasuredRate(t *testing.T) {
	cycles := threeCleanDays()
	for i := range cycles {
		// Six requests inside a two-second cycle, none of them throttled.
		cycles[i].FinishedAt = cycles[i].StartedAt.Add(2 * time.Second)
	}
	a, err := soak.BuildAttestation(soak.Summarize(cycles), criteria(), soakStart.AddDate(0, 0, 3), "tester", "note")
	if err != nil {
		t.Fatalf("BuildAttestation: %v", err)
	}
	if a.RateLimitPerSecond <= 0 {
		t.Errorf("RateLimitPerSecond = %v, want the observed sustained rate", a.RateLimitPerSecond)
	}
	if !strings.Contains(a.Notes, "note") {
		t.Errorf("Notes = %q, want the operator's note preserved", a.Notes)
	}
	if a.VerifiedBy != "tester" {
		t.Errorf("VerifiedBy = %q", a.VerifiedBy)
	}
}

// TestLiveOnlyEndpointsAreAllMutations. They are reported to the operator as
// "still to do"; if a read leaked into the list it would look like the soak was
// letting itself off something it could have proven.
func TestLiveOnlyEndpointsAreAllMutations(t *testing.T) {
	for _, e := range soak.LiveOnlyEndpoints() {
		if strings.HasPrefix(e, "GET ") {
			t.Errorf("%q is a read; the soak has no excuse not to prove it", e)
		}
	}
	if len(soak.LiveOnlyEndpoints()) == 0 {
		t.Error("LiveOnlyEndpoints is empty; the operator would never be told what remains")
	}
}

// TestRequiredEndpointsAreAllReads is the same guard from the other side.
func TestRequiredEndpointsAreAllReads(t *testing.T) {
	for _, e := range soak.RequiredEndpoints() {
		if !strings.HasPrefix(e, "GET ") {
			t.Errorf("%q is not a read; a read-only soak cannot require it", e)
		}
	}
}

func mentions(reasons []string, needle string) bool {
	for _, r := range reasons {
		if strings.Contains(strings.ToLower(r), strings.ToLower(needle)) {
			return true
		}
	}
	return false
}
