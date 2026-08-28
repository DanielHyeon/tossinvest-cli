package strategyflow

import (
	"strings"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/breakoutlane"
	"github.com/JungHoonGhae/tossinvest-cli/internal/continuationlane"
	"github.com/JungHoonGhae/tossinvest-cli/internal/reversallane"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyrouter"
	"github.com/JungHoonGhae/tossinvest-cli/internal/weeklyvaluelane"
)

// 아래 상수들은 breakoutlane 순수 코어가 요구하는 값을 그대로 쓴다.
// 하나라도 바꾸면 스냅샷이 봉인되지 않으므로, 고정 픽스처의 의미가 사라진다.
const (
	breakoutFixtureATRMinor    = 100  // 평균 실체 범위(최소 단위)
	breakoutFixtureResistance  = 1000 // 오프닝 레인지 상단 = 저항선
	breakoutFixtureRangeLow    = 900  // 오프닝 레인지 하단
	breakoutFixtureEntryMinor  = 1010 // 제안 진입가
	breakoutFixtureStopMinor   = 990  // 손절가 (진입가보다 반드시 낮다)
	breakoutFixtureTargetMinor = 1100 // 목표가 (진입가보다 반드시 높다)
	breakoutFixtureEvaluatedMS = 1_800_000_000_000
	breakoutFixtureFinalCap    = 3 // q_final 상한: q_candidate와 확실히 다르게 잡는다
)

func breakoutFixtureConfig(t *testing.T) breakoutlane.V1Config {
	t.Helper()
	input := breakoutlane.V1ConfigInput{
		Version: breakoutlane.LaneVersionV1, TickMinor: 1, OpeningRangeMinutes: 15,
		BreakoutBufferPPM: 100_000, RetestTolerancePPM: 150_000, TimeoutKR: 8, TimeoutUS: 10,
		RVOLMinPPM: 1_500_000, UpperWickRangeMaxPPM: 350_000,
		MaxQuoteAgeMS: 5_000, MaxSpreadPPM: 50_000, MaxEntryDriftPPM: 50_000,
	}
	input.Digest = breakoutlane.V1ConfigDigest(input)
	config, err := breakoutlane.NewV1Config(input)
	if err != nil {
		t.Fatalf("breakout fixture config: %v", err)
	}
	return config
}

// breakoutFixtureBars 는 ARMED 까지 도달하는 최소 봉 열이다.
// 15봉 오프닝 레인지 → 돌파봉 → 되돌림봉 → 회복봉.
func breakoutFixtureBars(t *testing.T, sessionID string) []breakoutlane.ClosedBar {
	t.Helper()
	rows := make([]breakoutlane.ClosedBarInput, 0, 18)
	for sequence := uint64(1); sequence <= 15; sequence++ {
		rows = append(rows, breakoutlane.ClosedBarInput{
			Sequence: sequence, HighMinor: breakoutFixtureResistance, LowMinor: breakoutFixtureRangeLow, CloseMinor: 950,
		})
	}
	// 돌파봉: 종가가 저항선 + 버퍼(10) 이상이고 RVOL/윗꼬리 조건을 만족한다.
	rows = append(rows, breakoutlane.ClosedBarInput{Sequence: 16, HighMinor: 1015, LowMinor: 1000, CloseMinor: 1010,
		RVOLPPM: 1_600_000, UpperWickRangePPM: 100_000})
	// 되돌림봉: 저항선과의 거리가 허용 오차 안이라 retest 로 인정된다.
	rows = append(rows, breakoutlane.ClosedBarInput{Sequence: 17, HighMinor: 1005, LowMinor: 995, CloseMinor: 1000})
	// 회복봉: 저항선 위에서 닫히므로 RECLAIMED → ARMED.
	rows = append(rows, breakoutlane.ClosedBarInput{Sequence: 18, HighMinor: 1006, LowMinor: 1000, CloseMinor: 1002})

	bars := make([]breakoutlane.ClosedBar, 0, len(rows))
	for _, row := range rows {
		row.Revision = 1
		row.IntervalMS = 60_000
		row.ID = "bar-" + strings.TrimSpace(sessionID) + "-" + itoa(row.Sequence)
		row.SessionID = sessionID
		row.RegularSession = true
		row.Closed = true
		bar, err := breakoutlane.NewClosedBar(row)
		if err != nil {
			t.Fatalf("breakout fixture bar %d: %v", row.Sequence, err)
		}
		bars = append(bars, bar)
	}
	return bars
}

func itoa(value uint64) string {
	if value == 0 {
		return "0"
	}
	digits := ""
	for value > 0 {
		digits = string(rune('0'+value%10)) + digits
		value /= 10
	}
	return digits
}

func breakoutFixtureEvidence(t *testing.T, market breakoutlane.Market) breakoutlane.EvidenceInput {
	t.Helper()
	laneID := breakoutlane.KRLaneID
	currency := "KRW"
	symbol := "005930"
	if market == breakoutlane.MarketUS {
		laneID, currency, symbol = breakoutlane.USLaneID, "USD", "AAPL"
	}
	sessionID := "session-" + strings.ToLower(string(market))
	quoteInput := breakoutlane.QuoteSealInput{BidMinor: 1009, AskMinor: 1011, LastMinor: 1010,
		SourceObservedAtMS: breakoutFixtureEvaluatedMS - 1_000, ReceivedAtMS: breakoutFixtureEvaluatedMS - 500, Currency: currency}
	quoteInput.Digest = breakoutlane.QuoteSealDigest(quoteInput)
	quote, err := breakoutlane.NewQuoteSeal(quoteInput)
	if err != nil {
		t.Fatalf("breakout fixture quote: %v", err)
	}
	fxInput := breakoutlane.FXSealInput{AccountCurrency: currency, InstrumentCurrency: currency,
		Direction: breakoutlane.FXAccountToInstrument, RateNum: 1, RateDen: 1, Scale: 1,
		AsOfMS: breakoutFixtureEvaluatedMS - 10_000, FreshUntilMS: breakoutFixtureEvaluatedMS + 10_000}
	fxInput.Digest = breakoutlane.FXSealDigest(fxInput)
	fx, err := breakoutlane.NewFXSeal(fxInput)
	if err != nil {
		t.Fatalf("breakout fixture fx: %v", err)
	}
	return breakoutlane.EvidenceInput{
		Market: market, Symbol: symbol, SessionID: sessionID, CalendarVersion: "calendar-v1",
		LaneID: laneID, LaneVersion: breakoutlane.LaneVersionV1, Config: breakoutFixtureConfig(t),
		Bars: breakoutFixtureBars(t, sessionID), ATRMinor: breakoutFixtureATRMinor,
		EvaluatedAtMS: breakoutFixtureEvaluatedMS, Quote: quote, FX: fx,
		Sizing: breakoutlane.SizingInput{
			ProposedEntryMinor: breakoutFixtureEntryMinor, StopMinor: breakoutFixtureStopMinor, TargetMinor: breakoutFixtureTargetMinor,
			RiskBudgetAccountMinor: 100_000, NotionalCapAccountMinor: 10_000_000, FinalCap: breakoutFixtureFinalCap,
		},
	}
}

func breakoutFixtureRequest(t *testing.T, market breakoutlane.Market) BreakoutRequest {
	t.Helper()
	evidence := breakoutFixtureEvidence(t, market)
	currency := "KRW"
	if market == breakoutlane.MarketUS {
		currency = "USD"
	}
	return BreakoutRequest{
		Evidence: evidence, AccountRef: "acct", CandidateID: "candidate-1",
		CampaignID: "campaign-" + strings.ToLower(string(market)), PositionGeneration: 7,
		LegOrdinal: 1, PlannedCeiling: 1, RiskBudgetDigest: "risk-" + strings.ToLower(string(market)),
		Price: BreakoutPriceEnvelope{Source: "official", Version: "v1", Digest: "price-digest",
			AsOf: "2026-08-28T00:00:00Z", Currency: currency, MinorScale: 0, UnitVersion: "minor-v1"},
		Policy: BreakoutPolicyInput{StagedTargetMinor: "1100", FairValueMinor: "1050",
			EntryCostsMinor: "0", ExitCostsMinor: "0", MinimumRRPPM: 0,
			DecisionDigest: "decision-digest", CalendarDigest: "calendar-digest", CapSnapshotID: "cap-1"},
	}
}

// 태스크 2.1: 정확히 네 가족 × 두 시장 = 8개, 그리고 기본값은 언제나 OFF/OFF/UNOBSERVED 다.
func TestPairedRegistryCoversAllFourFamiliesInBothMarkets(t *testing.T) {
	descriptors := Descriptors()
	if err := ValidateDescriptors(descriptors); err != nil {
		t.Fatalf("paired registry: %v", err)
	}
	if len(descriptors) != 8 {
		t.Fatalf("descriptors=%d, want eight paired bindings", len(descriptors))
	}
	// 골든 analysis/goldens/four-family-runtime-v1.json 의 lane_id 를 그대로 옮겨 적었다.
	want := map[string]strategyrouter.Horizon{
		"kr_short_flow_continuation_v1":          strategyrouter.HorizonShort,
		"us_short_participation_continuation_v1": strategyrouter.HorizonShort,
		"kr_short_absorption_reversal_v1":        strategyrouter.HorizonShort,
		"us_short_dislocation_reversal_v1":       strategyrouter.HorizonShort,
		"kr_weekly_disclosure_value_v1":          strategyrouter.HorizonWeekly,
		"us_weekly_disclosure_value_v1":          strategyrouter.HorizonWeekly,
		"kr_short_breakout_retest_v1":            strategyrouter.HorizonShort,
		"us_short_breakout_retest_v1":            strategyrouter.HorizonShort,
	}
	seen := make(map[string]bool, len(want))
	for _, descriptor := range descriptors {
		horizon, ok := want[descriptor.LaneID]
		if !ok {
			t.Fatalf("descriptor outside the frozen golden: %+v", descriptor)
		}
		if seen[descriptor.LaneID] {
			t.Fatalf("duplicate descriptor %s", descriptor.LaneID)
		}
		if descriptor.Horizon != horizon {
			t.Fatalf("%s horizon=%s, want %s", descriptor.LaneID, descriptor.Horizon, horizon)
		}
		if descriptor.LaneVersion != "v1" || descriptor.Release == "" {
			t.Fatalf("%s is missing a lane version or release: %+v", descriptor.LaneID, descriptor)
		}
		if descriptor.Desired != StateOff || descriptor.Effective != StateOff || descriptor.Runtime != RuntimeUnobserved {
			t.Fatalf("descriptor synthesized activation authority: %+v", descriptor)
		}
		seen[descriptor.LaneID] = true
	}
	if len(seen) != len(want) {
		t.Fatalf("covered %d lanes, want %d", len(seen), len(want))
	}
}

// 태스크 2.1: 부분/중복/미지/불일치 바인딩은 전부 거절돼야 한다.
func TestValidateDescriptorsRejectsPartialDuplicateUnknownAndMismatched(t *testing.T) {
	full := Descriptors()
	t.Run("partial", func(t *testing.T) {
		if err := ValidateDescriptors(full[:len(full)-1]); err == nil {
			t.Fatal("a seven-lane matrix was accepted")
		}
	})
	t.Run("duplicate", func(t *testing.T) {
		values := append([]Descriptor(nil), full...)
		values[len(values)-1] = values[0]
		if err := ValidateDescriptors(values); err == nil {
			t.Fatal("a duplicated descriptor was accepted")
		}
	})
	t.Run("unknown", func(t *testing.T) {
		values := append([]Descriptor(nil), full...)
		values[0].LaneID = "kr_short_unknown_v1"
		if err := ValidateDescriptors(values); err == nil {
			t.Fatal("an unknown lane id was accepted")
		}
	})
	t.Run("mismatched-field", func(t *testing.T) {
		values := append([]Descriptor(nil), full...)
		values[0].Desired = State("ON")
		if err := ValidateDescriptors(values); err == nil {
			t.Fatal("a descriptor with a flipped desired state was accepted")
		}
	})
}

// 태스크 4.2: 두 레지스트리 모두 정확히 8개를 바인딩한다.
func TestBothRegistriesBindEveryPairedDescriptor(t *testing.T) {
	for name, values := range map[string]registry{"evaluate": defaultRegistry(), "propose": proposalRegistry()} {
		values := values
		t.Run(name, func(t *testing.T) {
			if len(values.bindings) != 8 {
				t.Fatalf("%s registry has %d bindings, want 8", name, len(values.bindings))
			}
			for _, descriptor := range Descriptors() {
				if _, ok := values.lookup(descriptor); !ok {
					t.Fatalf("%s registry does not bind %s", name, descriptor.LaneID)
				}
			}
		})
	}
}

// 태스크 4.2: 태그가 다른 LaneInput 은 어떤 descriptor 에도 붙지 않는다(폴백 금지).
func TestBreakoutLaneInputMatchesOnlyItsOwnDescriptor(t *testing.T) {
	inputs := map[string]LaneInput{
		breakoutlane.KRLaneID: BreakoutKR(breakoutFixtureRequest(t, breakoutlane.MarketKR)),
		breakoutlane.USLaneID: BreakoutUS(breakoutFixtureRequest(t, breakoutlane.MarketUS)),
	}
	for laneID, input := range inputs {
		for _, descriptor := range Descriptors() {
			got := input.matches(descriptor)
			want := descriptor.LaneID == laneID
			if got != want {
				t.Fatalf("%s input matched %s = %v, want %v", laneID, descriptor.LaneID, got, want)
			}
		}
	}
	// 다른 가족의 입력이 breakout descriptor 에 붙어서도 안 된다.
	others := []LaneInput{
		ContinuationKR(continuationlane.KREvaluationRequest{}),
		ReversalUS(reversallane.USEvaluationRequest{}),
		WeeklyKR(weeklyvaluelane.EvaluationRequest{}),
	}
	for _, descriptor := range Descriptors() {
		if descriptor.LaneID != breakoutlane.KRLaneID && descriptor.LaneID != breakoutlane.USLaneID {
			continue
		}
		for _, input := range others {
			if input.matches(descriptor) {
				t.Fatalf("a non-breakout input matched %s", descriptor.LaneID)
			}
		}
	}
}

// 결정 48: Propose 는 상한 없는 q_candidate, Evaluate 는 FinalCap 이 걸린 값을 읽는다.
func TestBreakoutProposeIsCapFreeAndEvaluateAppliesTheFinalCap(t *testing.T) {
	for _, market := range []breakoutlane.Market{breakoutlane.MarketKR, breakoutlane.MarketUS} {
		market := market
		t.Run(string(market), func(t *testing.T) {
			request := breakoutFixtureRequest(t, market)
			input := BreakoutKR(request)
			evaluate, propose := evaluateBreakoutKR, proposeBreakoutKR
			if market == breakoutlane.MarketUS {
				input = BreakoutUS(request)
				evaluate, propose = evaluateBreakoutUS, proposeBreakoutUS
			}
			evaluated := evaluate(input)
			proposed := propose(input)
			if !evaluated.accepted || !proposed.accepted {
				t.Fatalf("fixture did not reach PROPOSED: evaluate=%+v propose=%+v", evaluated, proposed)
			}
			if evaluated.quantity != breakoutFixtureFinalCap {
				t.Fatalf("evaluate quantity=%d, want the capped %d", evaluated.quantity, breakoutFixtureFinalCap)
			}
			if proposed.quantity <= evaluated.quantity {
				t.Fatalf("propose quantity=%d must exceed the capped %d for this fixture", proposed.quantity, evaluated.quantity)
			}
		})
	}
}

// 순수 코어가 만든 다이제스트가 계보로 그대로 넘어가야 한다(재작성 금지).
func TestBreakoutAdapterCarriesPureCoreDigestsIntoLineage(t *testing.T) {
	request := breakoutFixtureRequest(t, breakoutlane.MarketKR)
	snapshot, err := breakoutlane.NewEvidenceSnapshot(request.Evidence)
	if err != nil {
		t.Fatalf("fixture snapshot: %v", err)
	}
	decision := breakoutlane.Evaluate(snapshot, nil)
	got := proposeBreakoutKR(BreakoutKR(request))
	if got.lineage.EvidenceDigest != decision.SnapshotDigest() {
		t.Fatalf("evidence digest=%q, want the pure core's %q", got.lineage.EvidenceDigest, decision.SnapshotDigest())
	}
	if got.lineage.ConfigDigest != decision.ConfigDigest() {
		t.Fatalf("config digest=%q, want the pure core's %q", got.lineage.ConfigDigest, decision.ConfigDigest())
	}
	if got.lineage.AccountRef != request.AccountRef || got.lineage.Symbol != request.Evidence.Symbol ||
		got.lineage.CampaignID != request.CampaignID || got.lineage.CandidateID != request.CandidateID ||
		got.lineage.LegOrdinal != request.LegOrdinal || got.lineage.PlannedCeiling != request.PlannedCeiling ||
		got.lineage.RiskBudgetDigest != request.RiskBudgetDigest {
		t.Fatalf("lineage envelope was not carried through: %+v", got.lineage)
	}
	if got.policy.identity == "" {
		t.Fatal("execution policy identity is empty, so the terms can never seal")
	}
}

// 봉인되지 않는 스냅샷은 거절이지 빈 수용이 아니다.
func TestBreakoutAdapterRefusesAnUnsealableSnapshot(t *testing.T) {
	request := breakoutFixtureRequest(t, breakoutlane.MarketKR)
	request.Evidence.ATRMinor = 0 // 스냅샷 구조 검증을 깨뜨린다.
	got := proposeBreakoutKR(BreakoutKR(request))
	if got.accepted {
		t.Fatal("an unsealable snapshot was accepted")
	}
	if got.nativeCode == "" {
		t.Fatal("a refusal must carry a native code")
	}
	if got.quantity != 0 {
		t.Fatalf("a refusal carried quantity %d", got.quantity)
	}
}

// 시장이 어긋난 증거는 다른 시장의 어댑터에서 거절돼야 한다.
func TestBreakoutAdapterRefusesAMarketMismatch(t *testing.T) {
	krRequest := breakoutFixtureRequest(t, breakoutlane.MarketKR)
	if got := proposeBreakoutUS(BreakoutUS(krRequest)); got.accepted {
		t.Fatal("KR evidence was accepted by the US adapter")
	}
	usRequest := breakoutFixtureRequest(t, breakoutlane.MarketUS)
	if got := proposeBreakoutKR(BreakoutKR(usRequest)); got.accepted {
		t.Fatal("US evidence was accepted by the KR adapter")
	}
}
