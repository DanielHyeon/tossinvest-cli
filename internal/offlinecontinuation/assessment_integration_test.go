//go:build tossos_testseams

package offlinecontinuation

import (
	"crypto/sha256"
	"fmt"
	"strconv"
	"testing"
	"time"

	evidence "github.com/JungHoonGhae/tossinvest-cli/internal/campaignstateevidence"
	"github.com/JungHoonGhae/tossinvest-cli/internal/continuationlane"
	"github.com/JungHoonGhae/tossinvest-cli/internal/reversallane"
	"github.com/JungHoonGhae/tossinvest-cli/internal/weeklyvaluelane"
)

func TestTaggedFixturesAssessAllFamiliesMarketsWithoutMutation(t *testing.T) {
	tests := []struct {
		name string
		call func(*testing.T) (Result, string, string)
		want quantities
	}{
		{"continuation KR", continuationKRResult, quantities{"1000", "0", "0", "1000", 4, 4, "400"}},
		{"continuation US", continuationUSResult, quantities{"1000", "0", "0", "1000", 4, 4, "400"}},
		{"reversal KR", reversalKRResult, quantities{"1000", "0", "0", "1000", 4, 4, "400"}},
		{"reversal US", reversalUSResult, quantities{"1000", "0", "0", "1000", 4, 4, "400"}},
		{"weekly KR", weeklyKRResult, quantities{"1000", "20", "0", "980", 2, 2, "200"}},
		{"weekly US", weeklyUSResult, quantities{"1000", "20", "0", "980", 2, 2, "200"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, before, after := tt.call(t)
			if before != after {
				t.Fatal("native request mutated")
			}
			assertAssessed(t, got, tt.want)
		})
	}
}

type quantities struct {
	budget, filled, held, remaining string
	candidate, offline              uint64
	planned                         string
}

func assertAssessed(t *testing.T, got Result, want quantities) {
	t.Helper()
	if got.Status != statusAssessed || got.Code != "" || got.QCandidate == nil || got.QOffline == nil || got.Budget == nil || got.Filled == nil || got.Held == nil || got.Remaining == nil || got.PerUnit == nil || got.PlannedRisk == nil ||
		*got.QCandidate != want.candidate || *got.QOffline != want.offline || *got.Budget != want.budget || *got.Filled != want.filled || *got.Held != want.held || *got.Remaining != want.remaining || *got.PerUnit != "100" || *got.PlannedRisk != want.planned ||
		!got.InspectionOnly || got.LedgerCurrentness != "unverified" || got.RiskBoundSource != "caller_assumption" || got.ExecutionAuthority {
		t.Fatalf("got %+v", got)
	}
}

func TestNativeRefusalPrecedesMalformedContinuationTime(t *testing.T) {
	r, err := continuationlane.StrategyflowKRFixture("candidate", "evidence")
	if err != nil {
		t.Fatal(err)
	}
	r.Context.Leg.Ordinal = 2
	r.Context.Invalidation = continuationlane.Invalidation{Structural: true, Code: "STRUCTURAL_CHANGED"}
	r.Evidence.Envelope.EvaluatedAt = "not-a-time"
	got := AssessContinuationKR(r, continuationAssumptionForBadTime(r))
	if got.NativeKind != string(continuationlane.OutcomeInvalidation) || got.NativeCode != "" || got.InvalidationCode != "STRUCTURAL_CHANGED" || got.Code != "" || hasQuantity(got) {
		t.Fatalf("invalidation=%+v", got)
	}
	r.Context.Invalidation = continuationlane.Invalidation{}
	r.Context.Enabled = false
	got = AssessContinuationKR(r, Assumption{})
	if got.NativeKind != string(continuationlane.OutcomeRefusal) || got.NativeCode != string(continuationlane.RefusalLaneOff) || got.Code != string(continuationlane.RefusalLaneOff) || hasQuantity(got) {
		t.Fatalf("refusal=%+v", got)
	}
}

func TestFailuresDoNotExposeCandidateQuantity(t *testing.T) {
	n := nativeProposal{kind: "DECISION", quantity: 4, accepted: true}
	capture := func(time.Time) (evidence.Evidence, error) { return evidence.Evidence{}, fmt.Errorf("capture") }
	got := assessAt(n, 2, 2, 3, Assumption{Family: familyContinuation}, familyContinuation, time.Now(), capture)
	if got.Code != "CAPTURE_FAILED" || hasQuantity(got) {
		t.Fatalf("capture failure=%+v", got)
	}
	got = calculate(baseResult(n), evidence.Evidence{Plan: evidence.Plan{Budget: "1"}, Risk: evidence.Risk{Filled: "1", Held: "0"}}, "1", n.quantity)
	if got.Status != statusNoBudget || hasQuantity(got) {
		t.Fatalf("no budget=%+v", got)
	}
}

func TestAssumptionAndFollowOnFailuresHaveNoQuantity(t *testing.T) {
	r, err := continuationlane.StrategyflowKRFixture("candidate", "evidence")
	if err != nil {
		t.Fatal(err)
	}
	got := AssessContinuationKR(r, continuationAssumption(r))
	if got.Code != "FOLLOW_ON_ORDINAL_REQUIRED" || hasQuantity(got) {
		t.Fatalf("first leg=%+v", got)
	}
	r.Context.Leg.Ordinal = 2
	for _, mutate := range []func(*Assumption){
		func(a *Assumption) { a.Family = familyWeekly },
		func(a *Assumption) { a.BindingDigest = "other" },
		func(a *Assumption) { a.Market = "US" },
		func(a *Assumption) { a.AccountCurrency = "USD" },
	} {
		a := continuationAssumption(r)
		mutate(&a)
		got = AssessContinuationKR(r, a)
		if hasQuantity(got) || got.Code == "" {
			t.Fatalf("mismatch=%+v", got)
		}
	}
	r.Context.Risk.Latches[continuationlane.LatchUnknownActualRisk] = true
	got = AssessContinuationKR(r, continuationAssumption(r))
	if got.NativeCode != string(continuationlane.RefusalRiskLatched) || hasQuantity(got) {
		t.Fatalf("latched=%+v", got)
	}
}

func TestWeeklyNativeReservationAndInvalidationRemainNative(t *testing.T) {
	r, err := weeklyvaluelane.StrategyflowKRFollowOnFixture("candidate", "evidence", 2)
	if err != nil {
		t.Fatal(err)
	}
	r.ReservationID = "missing"
	got := AssessWeeklyKR(r, weeklyAssumption(r))
	if got.NativeCode != string(weeklyvaluelane.RefusalReservationMissing) || hasQuantity(got) {
		t.Fatalf("reservation=%+v", got)
	}
	r, err = weeklyvaluelane.StrategyflowKRFollowOnFixture("candidate", "evidence", 2)
	if err != nil {
		t.Fatal(err)
	}
	r.Invalidation = weeklyvaluelane.Invalidation{Structural: true, Code: "STRUCTURAL_CHANGED"}
	got = AssessWeeklyKR(r, weeklyAssumption(r))
	if got.NativeKind != string(weeklyvaluelane.OutcomeInvalidation) || got.InvalidationCode != "STRUCTURAL_CHANGED" || hasQuantity(got) {
		t.Fatalf("invalidation=%+v", got)
	}
}

func TestRequestDigestDetectsNestedMutation(t *testing.T) {
	r, err := continuationlane.StrategyflowKRFixture("candidate", "evidence")
	if err != nil {
		t.Fatal(err)
	}
	before := continuationSnapshot(t, r)
	r.Context.Risk.Latches[continuationlane.LatchUnknownActualRisk] = true
	if continuationSnapshot(t, r) == before {
		t.Fatal("capture snapshot missed nested mutation")
	}
}

func hasQuantity(r Result) bool {
	return r.QCandidate != nil || r.QOffline != nil || r.PlannedRisk != nil
}
func requestDigest(v any) string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(fmt.Sprintf("%#v", v))))
}

func continuationKRResult(t *testing.T) (Result, string, string) {
	r, err := continuationlane.StrategyflowKRFixture("candidate", "evidence")
	if err != nil {
		panic(err)
	}
	r.Context.Leg.Ordinal = 2
	before := continuationSnapshot(t, r)
	got := AssessContinuationKR(r, continuationAssumption(r))
	return got, before, continuationSnapshot(t, r)
}
func continuationUSResult(t *testing.T) (Result, string, string) {
	r, err := continuationlane.StrategyflowUSFixture("candidate", "evidence")
	if err != nil {
		panic(err)
	}
	r.Context.Leg.Ordinal = 2
	before := continuationUSSnapshot(t, r)
	got := AssessContinuationUS(r, continuationUSAssumption(r))
	return got, before, continuationUSSnapshot(t, r)
}
func reversalKRResult(t *testing.T) (Result, string, string) {
	r, err := reversallane.StrategyflowKRFixture("candidate", "evidence")
	if err != nil {
		panic(err)
	}
	r.Context.Leg.Ordinal = 2
	before := reversalSnapshot(t, r)
	got := AssessReversalKR(r, reversalAssumption(r))
	return got, before, reversalSnapshot(t, r)
}
func reversalUSResult(t *testing.T) (Result, string, string) {
	r, err := reversallane.StrategyflowUSFixture("candidate", "evidence")
	if err != nil {
		panic(err)
	}
	r.Context.Leg.Ordinal = 2
	before := reversalUSSnapshot(t, r)
	got := AssessReversalUS(r, reversalUSAssumption(r))
	return got, before, reversalUSSnapshot(t, r)
}
func weeklyKRResult(t *testing.T) (Result, string, string) {
	r, err := weeklyvaluelane.StrategyflowKRFollowOnFixture("candidate", "evidence", 2)
	if err != nil {
		panic(err)
	}
	before := weeklySnapshot(t, r)
	got := AssessWeeklyKR(r, weeklyAssumption(r))
	return got, before, weeklySnapshot(t, r)
}
func weeklyUSResult(t *testing.T) (Result, string, string) {
	r, err := weeklyvaluelane.StrategyflowUSFollowOnFixture("candidate", "evidence", 2)
	if err != nil {
		panic(err)
	}
	before := weeklyUSSnapshot(t, r)
	got := AssessWeeklyUS(r, weeklyUSAssumption(r))
	return got, before, weeklyUSSnapshot(t, r)
}

func encodedSnapshot(t *testing.T, request any, captured evidence.Evidence) string {
	t.Helper()
	encoded, err := evidence.Encode(captured)
	if err != nil {
		t.Fatal(err)
	}
	return requestDigest(request) + ":" + fmt.Sprintf("%x", sha256.Sum256(encoded))
}

func continuationSnapshot(t *testing.T, r continuationlane.KREvaluationRequest) string {
	at := parseTime(r.Evidence.Envelope.EvaluatedAt)
	e, err := continuationlane.CaptureOfflineEvidence(r.Context.Plan, r.Context.Risk, at)
	if err != nil {
		t.Fatal(err)
	}
	return encodedSnapshot(t, r, e)
}
func continuationUSSnapshot(t *testing.T, r continuationlane.USEvaluationRequest) string {
	at := parseTime(r.Evidence.Envelope.EvaluatedAt)
	e, err := continuationlane.CaptureOfflineEvidence(r.Context.Plan, r.Context.Risk, at)
	if err != nil {
		t.Fatal(err)
	}
	return encodedSnapshot(t, r, e)
}
func reversalSnapshot(t *testing.T, r reversallane.KREvaluationRequest) string {
	e, err := reversallane.CaptureOfflineEvidence(r.Context.Plan, r.Context.Risk, r.Evidence.CommonEnvelope.EvaluatedAt)
	if err != nil {
		t.Fatal(err)
	}
	return encodedSnapshot(t, r, e)
}
func reversalUSSnapshot(t *testing.T, r reversallane.USEvaluationRequest) string {
	e, err := reversallane.CaptureOfflineEvidence(r.Context.Plan, r.Context.Risk, r.Evidence.CommonEnvelope.EvaluatedAt)
	if err != nil {
		t.Fatal(err)
	}
	return encodedSnapshot(t, r, e)
}
func weeklySnapshot(t *testing.T, r weeklyvaluelane.EvaluationRequest) string {
	e, err := weeklyvaluelane.CaptureOfflineEvidence(r.Plan, r.Risk, r.Evidence.EvaluatedAt)
	if err != nil {
		t.Fatal(err)
	}
	return encodedSnapshot(t, r, e)
}
func weeklyUSSnapshot(t *testing.T, r weeklyvaluelane.EvaluationRequest) string {
	return weeklySnapshot(t, r)
}

func continuationAssumption(r continuationlane.KREvaluationRequest) Assumption {
	at := parseTime(r.Evidence.Envelope.EvaluatedAt)
	e, err := continuationlane.CaptureOfflineEvidence(r.Context.Plan, r.Context.Risk, at)
	if err != nil {
		panic(err)
	}
	return assumptionFor(familyContinuation, e)
}
func continuationAssumptionForBadTime(r continuationlane.KREvaluationRequest) Assumption {
	return Assumption{Family: familyContinuation}
}
func continuationUSAssumption(r continuationlane.USEvaluationRequest) Assumption {
	at := parseTime(r.Evidence.Envelope.EvaluatedAt)
	e, err := continuationlane.CaptureOfflineEvidence(r.Context.Plan, r.Context.Risk, at)
	if err != nil {
		panic(err)
	}
	return assumptionFor(familyContinuation, e)
}
func reversalAssumption(r reversallane.KREvaluationRequest) Assumption {
	e, err := reversallane.CaptureOfflineEvidence(r.Context.Plan, r.Context.Risk, r.Evidence.CommonEnvelope.EvaluatedAt)
	if err != nil {
		panic(err)
	}
	return assumptionFor(familyReversal, e)
}
func reversalUSAssumption(r reversallane.USEvaluationRequest) Assumption {
	e, err := reversallane.CaptureOfflineEvidence(r.Context.Plan, r.Context.Risk, r.Evidence.CommonEnvelope.EvaluatedAt)
	if err != nil {
		panic(err)
	}
	return assumptionFor(familyReversal, e)
}
func weeklyUSAssumption(r weeklyvaluelane.EvaluationRequest) Assumption { return weeklyAssumption(r) }

func weeklyAssumption(r weeklyvaluelane.EvaluationRequest) Assumption {
	e, err := weeklyvaluelane.CaptureOfflineEvidence(r.Plan, r.Risk, r.Evidence.EvaluatedAt)
	if err != nil {
		panic(err)
	}
	return assumptionFor(familyWeekly, e)
}
func assumptionFor(family string, e evidence.Evidence) Assumption {
	return Assumption{Family: family, BindingDigest: e.Risk.PlanDigest, Market: e.Plan.Scope.Market, AccountCurrency: e.Plan.AccountCurrency, PerUnitRiskUpperBoundMinor: "100"}
}
func parseTime(raw string) time.Time {
	v, err := time.Parse(time.RFC3339Nano, raw)
	if err != nil {
		panic(err)
	}
	return v
}

func TestTaggedWeeklyFollowOnOrdinalsThreeThroughSevenStayOffline(t *testing.T) {
	for _, market := range []weeklyvaluelane.Market{weeklyvaluelane.MarketKR, weeklyvaluelane.MarketUS} {
		for ordinal := 3; ordinal <= 7; ordinal++ {
			t.Run(string(market)+"-ordinal-"+strconv.Itoa(ordinal), func(t *testing.T) {
				var (
					request weeklyvaluelane.EvaluationRequest
					err     error
				)
				if market == weeklyvaluelane.MarketKR {
					request, err = weeklyvaluelane.StrategyflowKRFollowOnFixture("candidate", "evidence", ordinal)
				} else {
					request, err = weeklyvaluelane.StrategyflowUSFollowOnFixture("candidate", "evidence", ordinal)
				}
				if err != nil {
					t.Fatal(err)
				}
				before := weeklySnapshot(t, request)
				var got Result
				if market == weeklyvaluelane.MarketKR {
					got = AssessWeeklyKR(request, weeklyAssumption(request))
				} else {
					got = AssessWeeklyUS(request, weeklyUSAssumption(request))
				}
				assertAssessed(t, got, quantities{"1000", strconv.Itoa(20 * (ordinal - 1)), "0", strconv.Itoa(1000 - 20*(ordinal-1)), 2, 2, "200"})
				if after := weeklySnapshot(t, request); after != before {
					t.Fatal("native follow-on request mutated")
				}
			})
		}
	}
}

func TestTaggedWeeklyCrossCurrencyFXAndStaleBoundaryRemainOffline(t *testing.T) {
	request, err := weeklyvaluelane.StrategyflowUSCrossCurrencyFollowOnFixture("candidate", "evidence", 2)
	if err != nil {
		t.Fatal(err)
	}
	captured, err := weeklyvaluelane.CaptureOfflineEvidence(request.Plan, request.Risk, request.Evidence.EvaluatedAt)
	if err != nil || captured.Plan.FX == nil || captured.Plan.AccountCurrency != "KRW" || captured.Plan.QuoteCurrency != "USD" {
		t.Fatalf("cross currency capture=%+v err=%v", captured, err)
	}
	before := weeklyUSSnapshot(t, request)
	assumption := weeklyUSAssumption(request)
	assumption.PerUnitRiskUpperBoundMinor = "14300"
	got := AssessWeeklyUS(request, assumption)
	if got.Status != statusAssessed || got.QCandidate == nil || got.QOffline == nil || *got.QCandidate != 2 || *got.QOffline != 2 || got.PlannedRisk == nil || *got.PlannedRisk != "28600" {
		t.Fatalf("cross currency assessment=%+v", got)
	}
	assertOfflineBoundary(t, got)
	if after := weeklyUSSnapshot(t, request); after != before {
		t.Fatal("cross currency request mutated")
	}
	assumption.AccountCurrency = "USD"
	wrong := AssessWeeklyUS(request, assumption)
	if wrong.Code != "ASSUMPTION_CURRENCY_MISMATCH" || hasQuantity(wrong) {
		t.Fatalf("USD assumption=%+v", wrong)
	}
	assertOfflineBoundary(t, wrong)

	stale, err := weeklyvaluelane.StrategyflowUSCrossCurrencyStaleFollowOnFixture("candidate", "evidence", 2)
	if err != nil {
		t.Fatal(err)
	}
	staleResult := AssessWeeklyUS(stale, weeklyUSAssumption(stale))
	if staleResult.NativeCode != string(weeklyvaluelane.RefusalEvidenceStale) || staleResult.Code != string(weeklyvaluelane.RefusalEvidenceStale) || staleResult.Status != statusRefused || hasQuantity(staleResult) {
		t.Fatalf("stale=%+v", staleResult)
	}
	assertOfflineBoundary(t, staleResult)
}

func TestTaggedWeeklyAdapterPreservesNativeFXTamperRefusal(t *testing.T) {
	request, err := weeklyvaluelane.StrategyflowUSCrossCurrencyFollowOnFixture("candidate", "evidence", 2)
	if err != nil {
		t.Fatal(err)
	}
	assumption := weeklyUSAssumption(request)
	assumption.PerUnitRiskUpperBoundMinor = "14300"
	if err := weeklyvaluelane.InvalidateStrategyflowFXForTest(&request); err != nil {
		t.Fatal(err)
	}
	got := AssessWeeklyUS(request, assumption)
	if got.Status != statusRefused || got.NativeCode != string(weeklyvaluelane.RefusalRRInvalid) || got.Code != string(weeklyvaluelane.RefusalRRInvalid) || hasQuantity(got) {
		t.Fatalf("tampered FX assessment=%+v", got)
	}
	assertOfflineBoundary(t, got)
}

func TestTaggedWeeklyAdapterRejectsMissingCurrentReservation(t *testing.T) {
	for _, market := range []weeklyvaluelane.Market{weeklyvaluelane.MarketKR, weeklyvaluelane.MarketUS} {
		for ordinal := 3; ordinal <= 7; ordinal++ {
			t.Run(string(market)+"-ordinal-"+strconv.Itoa(ordinal), func(t *testing.T) {
				var (
					request weeklyvaluelane.EvaluationRequest
					err     error
				)
				if market == weeklyvaluelane.MarketKR {
					request, err = weeklyvaluelane.StrategyflowKRFollowOnFixture("candidate", "evidence", ordinal)
				} else {
					request, err = weeklyvaluelane.StrategyflowUSFollowOnFixture("candidate", "evidence", ordinal)
				}
				if err != nil {
					t.Fatal(err)
				}
				if err := weeklyvaluelane.RemoveStrategyflowActiveReservationForTest(&request); err != nil {
					t.Fatal(err)
				}
				var got Result
				if market == weeklyvaluelane.MarketKR {
					got = AssessWeeklyKR(request, weeklyAssumption(request))
				} else {
					got = AssessWeeklyUS(request, weeklyUSAssumption(request))
				}
				if got.Status != statusRefused || got.NativeKind != string(weeklyvaluelane.OutcomeRefusal) || got.NativeCode != string(weeklyvaluelane.RefusalReservationMissing) || got.Code != string(weeklyvaluelane.RefusalReservationMissing) || hasQuantity(got) {
					t.Fatalf("reservation-state integrity assessment=%+v", got)
				}
				assertOfflineBoundary(t, got)
			})
		}
	}
}

func assertOfflineBoundary(t *testing.T, result Result) {
	t.Helper()
	if !result.InspectionOnly || result.LedgerCurrentness != "unverified" || result.ExecutionAuthority {
		t.Fatalf("authority boundary=%+v", result)
	}
}
