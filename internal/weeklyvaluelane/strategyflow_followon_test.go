//go:build tossos_testseams

package weeklyvaluelane

import (
	"strconv"
	"testing"
	"time"

	evidence "github.com/JungHoonGhae/tossinvest-cli/internal/campaignstateevidence"
)

func TestStrategyflowFollowOnFixtureBuildsNativePrefixThroughOrdinalSeven(t *testing.T) {
	for _, market := range []Market{MarketKR, MarketUS} {
		for ordinal := 3; ordinal <= 7; ordinal++ {
			t.Run(string(market)+"-ordinal-"+string(rune('0'+ordinal)), func(t *testing.T) {
				var (
					request EvaluationRequest
					err     error
				)
				if market == MarketKR {
					request, err = StrategyflowKRFollowOnFixture("candidate", "evidence", ordinal)
				} else {
					request, err = StrategyflowUSFollowOnFixture("candidate", "evidence", ordinal)
				}
				if err != nil {
					t.Fatalf("fixture ordinal=%d: %v", ordinal, err)
				}
				scope := request.Reservations.scopes[reservationScopeKey(request.Plan.CampaignID(), market)]
				if scope.PositiveLegCount != ordinal-1 {
					t.Fatalf("positive legs=%d want=%d", scope.PositiveLegCount, ordinal-1)
				}
				for prior := 1; prior < ordinal; prior++ {
					if !scope.ConsumedOrdinals[prior-1] {
						t.Fatalf("ordinal %d was not consumed", prior)
					}
				}
				key := CanonicalReservationKey(request.Plan.CampaignID(), request.MarketWeek)
				entry, ok := request.Reservations.entries[key]
				if !ok || entry.Status != ReservationActive || entry.PlannedOrdinal != ordinal {
					t.Fatalf("active reservation=%+v found=%t", entry, ok)
				}
				wantFilled := strconv.Itoa(20 * (ordinal - 1))
				if request.Risk.FilledMinor() != wantFilled || request.Risk.HeldMinor() != "0" {
					t.Fatalf("risk filled=%s held=%s", request.Risk.FilledMinor(), request.Risk.HeldMinor())
				}
			})
		}
	}
}

func TestStrategyflowUSCrossCurrencyFixtureCapturesFXAndRejectsTamper(t *testing.T) {
	request, err := StrategyflowUSCrossCurrencyFollowOnFixture("candidate", "evidence", 2)
	if err != nil {
		t.Fatal(err)
	}
	if request.Plan.FX() == nil {
		t.Fatal("cross-currency fixture has no FX")
	}
	if got := ProposeUS(request); got.Kind != OutcomeDecision || got.Quantity == 0 {
		t.Fatalf("cross-currency proposal=%+v", got)
	}
	before, err := CaptureOfflineEvidence(request.Plan, request.Risk, request.Evidence.EvaluatedAt)
	if err != nil {
		t.Fatal(err)
	}
	beforeBytes, err := evidence.Encode(before)
	if err != nil {
		t.Fatal(err)
	}
	request.Plan.fx.rateQuoteToAccount = "999"
	after, err := CaptureOfflineEvidence(request.Plan, request.Risk, request.Evidence.EvaluatedAt)
	if err != nil {
		t.Fatal(err)
	}
	afterBytes, err := evidence.Encode(after)
	if err != nil {
		t.Fatal(err)
	}
	if string(beforeBytes) == string(afterBytes) {
		t.Fatal("FX mutation was not observable")
	}
	if got := ProposeUS(request); got.Kind == OutcomeDecision || got.Code != RefusalRRInvalid {
		t.Fatalf("tampered FX=%+v", got)
	}
}

func TestStrategyflowUSCrossCurrencyFixtureRefusesStaleEvidence(t *testing.T) {
	request, err := StrategyflowUSCrossCurrencyFollowOnFixture("candidate", "evidence", 2)
	if err != nil {
		t.Fatal(err)
	}
	if got := ProposeUS(request); got.Kind != OutcomeDecision {
		t.Fatalf("fresh control=%+v", got)
	}
	request.Evidence.FreshUntil = request.Evidence.EvaluatedAt.Add(-time.Nanosecond)
	request.Evidence.seal = evidenceSnapshotSeal(request.Evidence)
	request.authorization = mintDormantEvaluationAuthorization(request.Plan, request.Evidence)
	request.executionTerms = mintExecutionTermsPreimage(request.Plan, request.Evidence, request.EntryPriceMinor, request.StagedTargetMinor, request.EntryCostsMinor, request.EstimatedExitCostsLeviesMinor, request.MinimumRRPPM)
	if got := ProposeUS(request); got.Kind != OutcomeRefusal || got.Code != RefusalEvidenceStale || got.Quantity != 0 {
		t.Fatalf("stale=%+v", got)
	}
}
