package console

// signals_test.go covers the discovery screen (change add-candidate-discovery,
// tasks 5.5–5.7).
//
// Every assertion here is made against the RENDERED PAGE rather than against the
// view struct. The claim this screen exists to make is about what a person reads,
// and a view field named UnmeasuredCount satisfies a struct test while the
// template drops the column.

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/candidate"
)

// --- the seam a test fills ----------------------------------------------------------

// stubSignals is the reader the screen is given.
type stubSignals struct {
	reading SignalsReading
	err     error
	calls   int
}

func (s *stubSignals) Signals(context.Context) (SignalsReading, error) {
	s.calls++
	if s.err != nil {
		return SignalsReading{}, s.err
	}
	return s.reading, nil
}

var signalsNow = time.Date(2026, 7, 28, 9, 30, 0, 0, time.UTC)

// newSignalsHarness wires the screen with a stub reader and a fixed clock.
func newSignalsHarness(t *testing.T, reader SignalsReader) *harness {
	t.Helper()
	return newHarness(t, func(o *Options) {
		o.Signals = reader
		o.Now = func() time.Time { return signalsNow }
	})
}

// renderSignals fetches the rendered screen.
func renderSignals(t *testing.T, h *harness) string {
	t.Helper()
	h.authenticate(t)
	resp := h.get(t, "/signals")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET /signals = %d, want 200", resp.StatusCode)
	}
	return body(t, resp)
}

// --- fixtures -----------------------------------------------------------------------

// measuredAndClear is a candidate whose three chase vetoes were all evaluated and
// none of which fired. Under D18 this is unreachable in production — two of the
// three have no approved threshold — and it is constructed here precisely so the
// screen can be asked to tell it apart from the one below.
func measuredAndClear(symbol string) candidate.Verdict {
	return candidate.Verdict{
		Summary: candidate.Summary{Candidate: candidate.Candidate{
			Key:         candidate.Key{Market: candidate.MarketKR, Symbol: symbol},
			State:       candidate.StateActive,
			FirstSeenAt: signalsNow.Add(-20 * time.Minute),
			LastSeenAt:  signalsNow.Add(-15 * time.Second),
			Sources: []candidate.SourceID{
				candidate.SourceOfficialTradingValue, candidate.SourceWTSPopular,
			},
			SourcesAttempted: 5, SourcesResponded: 5,
		}},
		Chase: candidate.Chase{
			Key:      candidate.Key{Market: candidate.MarketKR, Symbol: symbol},
			At:       signalsNow,
			SeenLate: candidate.ClearedVeto(),
			Extended: candidate.ClearedVeto(),
			NearHigh: candidate.ClearedVeto(),
		},
	}
}

// neverChecked is the ordinary candidate: nobody spent a candle on it, so no veto
// fired and no veto was measured. D13 decision 3 makes this the majority of the
// list all day, and D10 is the rule it must not be read under.
func neverChecked(symbol string) candidate.Verdict {
	return candidate.Verdict{
		Summary: candidate.Summary{Candidate: candidate.Candidate{
			Key:         candidate.Key{Market: candidate.MarketKR, Symbol: symbol},
			State:       candidate.StateActive,
			FirstSeenAt: signalsNow.Add(-8 * time.Minute),
			LastSeenAt:  signalsNow.Add(-15 * time.Second),
			Sources:     []candidate.SourceID{candidate.SourceOfficialTradingValue},
			// Four of five sources answered: the screen has to be able to say so
			// per row as well as per panel.
			SourcesAttempted: 5, SourcesResponded: 4, Degraded: true,
		}},
		Chase: candidate.Chase{
			Key:      candidate.Key{Market: candidate.MarketKR, Symbol: symbol},
			At:       signalsNow,
			SeenLate: candidate.UnmeasuredVeto(candidate.VetoThresholdAbsent),
			Extended: candidate.UnmeasuredVeto(candidate.VetoThresholdAbsent),
			// NO_DAY_HIGH travels from level.go under its own name rather than
			// being folded into one "the input was missing" — a screen full of it
			// is the candle budget working as designed, and a screen full of
			// THRESHOLD_ABSENT is nobody having configured the veto.
			NearHigh: candidate.UnmeasuredVeto(
				candidate.VetoUnmeasured(candidate.LevelNoDayHigh)),
		},
	}
}

// oneMarket wraps verdicts in the shape the seam returns, with a panel that
// answered completely.
func oneMarket(verdicts ...candidate.Verdict) SignalsReading {
	chases := make([]candidate.Chase, 0, len(verdicts))
	for _, v := range verdicts {
		chases = append(chases, v.Chase)
	}
	return SignalsReading{
		At: signalsNow,
		Markets: []SignalsMarket{{
			Market:   candidate.MarketKR,
			Verdicts: verdicts,
			Vetoes:   candidate.TallyVetoes(chases),
			Panel: SignalsPanel{
				Known: true, At: signalsNow, Attempted: 5, Responded: 5,
			},
		}},
	}
}

// --- task 5.5: the list ------------------------------------------------------------

// TestTheSignalsScreenListsWhatDiscoveryFoundAndWhenItFirstSawIt.
//
// The screen's first job is the record D3 requires to stay countable: which
// symbols discovery is holding, when each was first seen, and which sources are
// vouching for it. first_seen_at is the whole claim this package makes, so it is
// on the page rather than derivable from it.
func TestTheSignalsScreenListsWhatDiscoveryFoundAndWhenItFirstSawIt(t *testing.T) {
	reader := &stubSignals{reading: oneMarket(neverChecked("005930"))}
	page := renderSignals(t, newSignalsHarness(t, reader))

	for _, want := range []string{
		"005930",
		"2026-07-28 09:22:00Z", // first_seen_at, eight minutes back
		"official_rankings_trading_value",
	} {
		if !strings.Contains(page, want) {
			t.Errorf("the screen does not carry %q; discovery's record is not on the page", want)
		}
	}
	if reader.calls == 0 {
		t.Error("the screen never asked the seam for anything")
	}
}

// TestTheSignalsScreenHasNothingToSubmitAndAsksForNoConfirmation.
//
// 읽기 전용 화면이다. The standing instruction (2026-07-27) is that this repository
// puts no typed confirmation, no second click and no extra approval on anything;
// the way to keep that true here is for the page to contain nothing that could
// carry one.
func TestTheSignalsScreenHasNothingToSubmitAndAsksForNoConfirmation(t *testing.T) {
	page := renderSignals(t, newSignalsHarness(t, &stubSignals{
		reading: oneMarket(neverChecked("005930")),
	}))
	for _, banned := range []string{"<form", "<input", "<button", "<textarea", "method=\"post\"", "csrf"} {
		if strings.Contains(strings.ToLower(page), banned) {
			t.Errorf("the signals screen contains %q; it is a 관측창 with nothing on it to press", banned)
		}
	}
	for _, banned := range []string{"입력하", "타이핑", "다시 한 번", "확인 문구"} {
		if strings.Contains(page, banned) {
			t.Errorf("the signals screen asks for %q; read-only surfaces do not manufacture friction", banned)
		}
	}
}

// signalsRowFor returns the <tr> that carries a symbol, so an assertion can be made
// about ONE row rather than about the whole page. A page-wide Contains is
// satisfied by the other row, which is precisely the confusion these tests exist
// to detect.
func signalsRowFor(t *testing.T, page, symbol string) string {
	t.Helper()
	for _, row := range strings.Split(page, "<tr>") {
		if !strings.Contains(row, ">"+symbol+"<") {
			continue
		}
		// Cut at the row's own end. Without this the LAST row runs on into the
		// paragraph below the table, and that paragraph explains the very words
		// these tests count — so an assertion about one row would be satisfied by
		// the prose beneath it.
		if end := strings.Index(row, "</tr>"); end >= 0 {
			return row[:end]
		}
		return row
	}
	t.Fatalf("no table row on the page carries %q", symbol)
	return ""
}

// --- task 5.6: unmeasured must not look like a pass ---------------------------------

// TestARowNobodyMeasuredDoesNotRenderLikeARowThatCleared.
//
// This is the screen the whole change was built for. A candidate with no veto
// reason against it means two things — "we measured all three and none fired" and
// "we never checked" — and under the candle budget (five calls a second, one per
// symbol) the second is the ordinary case for most of the list, all day (D13
// decision 3).
//
// The assertion is on the two rendered rows rather than on the view, because the
// view can hold a perfectly correct three-state verdict while the template prints
// a veto only when one fired — which produces two identical rows.
func TestARowNobodyMeasuredDoesNotRenderLikeARowThatCleared(t *testing.T) {
	page := renderSignals(t, newSignalsHarness(t, &stubSignals{
		reading: oneMarket(measuredAndClear("000660"), neverChecked("005930")),
	}))
	clear, unchecked := signalsRowFor(t, page, "000660"), signalsRowFor(t, page, "005930")

	if clear == unchecked {
		t.Fatal("the measured-and-clear row and the never-checked row render identically")
	}
	// The measured row says what was measured.
	if strings.Count(clear, "측정·안전") != len(candidate.VetoCodes) {
		t.Errorf("the cleared row carries %d 측정·안전 cells, want %d — one per veto",
			strings.Count(clear, "측정·안전"), len(candidate.VetoCodes))
	}
	if !strings.Contains(clear, "통과") {
		t.Error("the row whose three vetoes were all measured and all clear does not say 통과")
	}
	// The unmeasured row says what was NOT measured, by name, and says nothing
	// that could be read as a pass.
	for _, want := range []string{"미측정", "미확인", "THRESHOLD_ABSENT", "NO_DAY_HIGH"} {
		if !strings.Contains(unchecked, want) {
			t.Errorf("the never-checked row does not carry %q; an operator cannot tell it apart "+
				"from a candidate that was checked and cleared", want)
		}
	}
	for _, banned := range []string{"측정·안전", ">통과<"} {
		if strings.Contains(unchecked, banned) {
			t.Errorf("the never-checked row renders %q; unmeasured is not a pass (D10)", banned)
		}
	}
	// And the row says how much of it was measured, which is what makes 미확인
	// actionable rather than an apology.
	if !strings.Contains(unchecked, "0 / 3 사유 측정") {
		t.Error("the never-checked row does not say how many of the three vetoes were measured")
	}
	if !strings.Contains(clear, "3 / 3 사유 측정") {
		t.Error("the cleared row does not say how many of the three vetoes were measured")
	}
}

// TestTheStructurallyZeroPassCountIsShownWithTheSentenceThatExplainsIt.
//
// Under D18 two of the three vetoes have no approved threshold anywhere in this
// repository, so Chase.Passed() is unreachable and the tally's Passed is
// permanently zero. That is an honest reading and not an outage.
//
// There are two wrong repairs and this pins both out. Counting THRESHOLD_ABSENT
// as a pass is the one D18 names. Hiding a column because it is always zero is
// the other: it leaves the next reader with the number gone and no way to find
// out why, and the first thing they will do is put it back the wrong way.
func TestTheStructurallyZeroPassCountIsShownWithTheSentenceThatExplainsIt(t *testing.T) {
	page := renderSignals(t, newSignalsHarness(t, &stubSignals{
		reading: oneMarket(neverChecked("005930"), neverChecked("000660")),
	}))
	if !strings.Contains(page, "통과한 후보") {
		t.Error("the passed count is not on the page at all; a column removed because it is " +
			"always zero leaves the next reader no way to learn why it was zero")
	}
	for _, want := range []string{"구조적으로 0", "seen_late", "extended", "D18", "고장이 아니"} {
		if !strings.Contains(page, want) {
			t.Errorf("the passed count is shown without %q; a bare \"통과 0\" sends somebody "+
				"looking for a bug, and the most natural bug to fix is the one D18 forbids", want)
		}
	}
	// The unmeasured count leads. spec Requirement 8: 화면의 기본 독해는
	// "대부분 안전"이 아니라 "대부분 미확인"이어야 한다.
	unmeasured := strings.Index(page, "미측정 후보")
	passed := strings.Index(page, "통과한 후보")
	if unmeasured < 0 || passed < 0 || unmeasured > passed {
		t.Errorf("the unmeasured count is not above the passed count (at %d and %d)",
			unmeasured, passed)
	}
}

// TestTheSeriesCountIsNeverPresentedAsACandidateCount is D21.
//
// The acceleration tally's unit is the (market, symbol, source) series, because
// differencing two sources' cumulative figures against each other measures the
// sources rather than the market (D9). So its total is LARGER than the candidate
// count — a WTS popularity list carries no trading value and still occupies a slot
// every tick. "후보 300개 중 12개가 임계를 넘었다" and "계열 900개 중 12개" are
// different sentences and only one of them is true.
func TestTheSeriesCountIsNeverPresentedAsACandidateCount(t *testing.T) {
	reading := oneMarket(neverChecked("005930"), neverChecked("000660"))
	// Two candidates, five series: the shape D21 describes.
	reading.Markets[0].Crossings = candidate.CrossingTally{
		Total: 5, Computed: 3,
		Crossed:     map[string]int{"1.3": 2, "1.5": 1},
		NotComputed: map[candidate.NotComputed]int{candidate.NotComputedFigureAbsent: 2},
	}
	page := renderSignals(t, newSignalsHarness(t, &stubSignals{reading: reading}))

	if !strings.Contains(page, "계열 5개") {
		t.Error("the acceleration tally does not label its 5 as 계열")
	}
	if !strings.Contains(page, "후보 2개") {
		t.Error("the veto tally does not label its 2 as 후보")
	}
	if strings.Contains(page, "후보 5개") {
		t.Error("the series count is rendered under the word 후보; a series count presented as a " +
			"candidate count is a different and much more alarming sentence")
	}
	if strings.Contains(page, "계열 2개") {
		t.Error("the candidate count is rendered under the word 계열")
	}
	// And the rule itself is on the page, in the words the code holds, so a reader
	// who has not read D21 learns why the two numbers differ.
	if !strings.Contains(page, signalsSeriesNote) {
		t.Error("the acceleration tally does not carry the sentence saying its denominator is " +
			"series and not candidates")
	}
	if !strings.Contains(page, signalsShadowNote) {
		t.Error("the shadow record is not labelled as a record; a crossing count beside a veto " +
			"tally reads as a second verdict")
	}
}

// TestNoUnmeasuredCellOnTheSignalsScreenIsAReasonlessDash.
//
// An em dash with no reason beside it tells nobody whether to wait, to fix a
// credential, to wire a seam or to configure a threshold. The zero value of every
// measurement here answers a NAMED reason rather than an empty one — Reason() on
// Sighting, Expansion and RangePosition all guarantee it — and this is the
// assertion that the template renders that name instead of the raw field.
func TestNoUnmeasuredCellOnTheSignalsScreenIsAReasonlessDash(t *testing.T) {
	// A verdict nobody assigned anything to: every measurement is its zero value.
	bare := candidate.Verdict{Summary: candidate.Summary{Candidate: candidate.Candidate{
		Key:         candidate.Key{Market: candidate.MarketKR, Symbol: "005930"},
		State:       candidate.StateActive,
		FirstSeenAt: signalsNow.Add(-time.Minute),
	}}}
	page := renderSignals(t, newSignalsHarness(t, &stubSignals{reading: oneMarket(bare)}))
	row := signalsRowFor(t, page, "005930")

	// NOT_EVALUATED is what an unassigned struct names itself, and there are SIX of
	// them on this row: the three vetoes and the three measurements they read.
	//
	// The count is six rather than three because three was satisfied by the veto
	// cells alone — the first draft of this assertion passed with the three
	// measurement cells reading the raw Why field, which is empty on a zero value.
	// Both halves have to be counted or the half being tested is the one nobody
	// looked at.
	if n := strings.Count(row, "NOT_EVALUATED"); n < 6 {
		t.Errorf("the row carries %d NOT_EVALUATED reasons, want at least 6 (three vetoes and "+
			"three measurements); a measurement nobody took renders without its name", n)
	}
	// And the reasons came from each type's Reason(), which guarantees a name,
	// rather than from the raw field with this file's own fallback behind it.
	if strings.Contains(row, "REASON_MISSING") {
		t.Error("a measurement cell fell back to REASON_MISSING; the three measurements each " +
			"answer Reason() precisely so that the fallback is unreachable, and reaching it " +
			"means the raw Why field was rendered instead")
	}
	if strings.Contains(row, "—</td>") {
		t.Error("the row closes a cell on a bare em dash; a reason-less dash tells nobody " +
			"whether to wait, to fix something or to configure a threshold")
	}
}

// TestAnUnwiredDiscoverySeamIsNotAnEmptyMarket.
//
// A build that was not handed the seam has to say so. An empty list would read as
// "discovery is running and the market is quiet", which is the same confusion as
// an unmeasured veto reading as a pass, one level up.
func TestAnUnwiredDiscoverySeamIsNotAnEmptyMarket(t *testing.T) {
	page := renderSignals(t, newSignalsHarness(t, nil))
	if !strings.Contains(page, "seam_unwired") {
		t.Error("a console without the discovery seam does not name seam_unwired")
	}
	if !strings.Contains(page, "0건이 아니다") {
		t.Error("the unwired screen does not say that this is not a count of zero")
	}
}

// TestAStoreThatCouldNotBeReadIsItsOwnReasonAndNotTheJournalsOrTheBrokers.
//
// The operator's next move differs from all seven of the operator-console
// enumeration's: it is not a credential, not the order ledger, not a cold cache
// and not an unwired seam — the seam IS wired and it failed. A reason that exists
// only as a free sentence cannot be counted or tested for absence, so it has a
// code.
func TestAStoreThatCouldNotBeReadIsItsOwnReasonAndNotTheJournalsOrTheBrokers(t *testing.T) {
	page := renderSignals(t, newSignalsHarness(t, &stubSignals{
		err: errStubDiscovery,
	}))
	if !strings.Contains(page, "discovery_unreadable") {
		t.Error("a discovery store that could not be read has no code on the page")
	}
	if !strings.Contains(page, errStubDiscovery.Error()) {
		t.Error("the failure's own words are not on the page; a code with no detail sends an " +
			"operator to a file without saying what is wrong with it")
	}
	for _, banned := range []string{"journal_unreadable", "broker_read_failed", "seam_unwired"} {
		if strings.Contains(page, banned) {
			t.Errorf("the discovery failure is reported as %q, which is the wrong advice", banned)
		}
	}
}

// --- task 5.7: degradation says which sources went missing --------------------------

// TestDegradationIsSaidWithTheNamesOfTheMissingSources.
//
// "degraded: true" on its own is a display nobody can act on. An expired WTS
// session and a rate-limited official ranking are the same boolean and the
// opposite response — and the official ranking has no WTS fallback at all (D14),
// so one 429 is the whole source rather than a slower one.
func TestDegradationIsSaidWithTheNamesOfTheMissingSources(t *testing.T) {
	reading := oneMarket(neverChecked("005930"))
	reading.Markets[0].Panel = SignalsPanel{
		Known: true, At: signalsNow, Attempted: 5, Responded: 3, Degraded: true,
		Missing: []candidate.SourceFailure{
			{
				Source:      candidate.SourceOfficialTradingValue,
				Reason:      "429 too many requests",
				RateLimited: true,
			},
			{Source: candidate.SourceWTSPopular, Reason: "WTS 세션이 만료됐다"},
		},
	}
	page := renderSignals(t, newSignalsHarness(t, &stubSignals{reading: reading}))

	for _, want := range []string{
		"official_rankings_trading_value", "429 too many requests",
		"wts_popular", "WTS 세션이 만료됐다",
		"3 / 5",
	} {
		if !strings.Contains(page, want) {
			t.Errorf("the degradation notice does not carry %q; a reason-less boolean is a "+
				"display nobody can act on", want)
		}
	}
}

// TestADegradedPanelThatNamesNoSourceSaysSoRatherThanLookingClean.
//
// The state task 5.7 is really about is the one in between: the pass reports a
// degradation and names nothing. Rendering an empty list under a 강등 heading reads
// as "and then it recovered", so the screen reports the silence itself.
func TestADegradedPanelThatNamesNoSourceSaysSoRatherThanLookingClean(t *testing.T) {
	reading := oneMarket(neverChecked("005930"))
	reading.Markets[0].Panel = SignalsPanel{
		Known: true, At: signalsNow, Attempted: 5, Responded: 4, Degraded: true,
	}
	page := renderSignals(t, newSignalsHarness(t, &stubSignals{reading: reading}))
	if !strings.Contains(page, "강등인데 빠진 원천 이름이 없다") {
		t.Error("a degraded pass that named no source renders as an ordinary degradation; the " +
			"missing names are the only actionable half of the notice")
	}
}

// TestAPanelNobodyReportedIsNotAPanelWithNothingMissing.
//
// Known is separate from Degraded for the reason the whole change exists: "no
// source is missing" and "nobody told us which sources are missing" are the same
// boolean and only one of them means the panel was whole.
func TestAPanelNobodyReportedIsNotAPanelWithNothingMissing(t *testing.T) {
	reading := oneMarket(neverChecked("005930"))
	reading.Markets[0].Panel = SignalsPanel{Why: "이 콘솔 프로세스는 스캔을 돌리지 않는다"}
	page := renderSignals(t, newSignalsHarness(t, &stubSignals{reading: reading}))

	if !strings.Contains(page, "원천 상태 미측정") {
		t.Error("a panel nobody reported on renders as a measured one")
	}
	if !strings.Contains(page, "이 콘솔 프로세스는 스캔을 돌리지 않는다") {
		t.Error("the unmeasured panel does not carry the reason it is unmeasured")
	}
	if strings.Contains(page, "없음 — 전 원천 응답") {
		t.Error("a panel nobody reported on is rendered as \"no source is missing\"")
	}
}

// errStubDiscovery is what a store that will not open says.
var errStubDiscovery = errors.New("candidate: opening the discovery store: permission denied")

// TestEveryUnmeasuredStateOnTheSignalsScreenCarriesACodeAndASentence.
//
// Three levels of this screen can be unmeasured — the whole reading, one market,
// and the source panel — and each of them renders an em dash somewhere. A dash
// with no code beside it cannot be grepped for and cannot be tested for absence,
// which is the argument I-2's ruling makes for codes over free sentences; a code
// with no sentence sends an operator to a file without saying what is wrong with
// it.
func TestEveryUnmeasuredStateOnTheSignalsScreenCarriesACodeAndASentence(t *testing.T) {
	marketFailed := oneMarket(neverChecked("005930"))
	marketFailed.Markets[0].Why = "candidate: reading KR observations: disk I/O error"

	panelAbsent := oneMarket(neverChecked("005930"))
	panelAbsent.Markets[0].Panel = SignalsPanel{}

	for _, tc := range []struct {
		name     string
		reader   SignalsReader
		wantCode string
		wantWord string
	}{
		{"the seam is not wired", nil, "seam_unwired", "미측정"},
		{"the store would not open", &stubSignals{err: errStubDiscovery},
			"discovery_unreadable", "발굴 저장소 미판독"},
		{"one market would not read", &stubSignals{reading: marketFailed},
			"discovery_unreadable", "disk I/O error"},
		{"no scan record reached the screen", &stubSignals{reading: panelAbsent},
			"seam_unwired", "tossctl candidate scan"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			page := renderSignals(t, newSignalsHarness(t, tc.reader))
			if !strings.Contains(page, tc.wantCode) {
				t.Errorf("the page carries no %q code", tc.wantCode)
			}
			if !strings.Contains(page, tc.wantWord) {
				t.Errorf("the page carries no sentence containing %q", tc.wantWord)
			}
		})
	}
}

// TestTheDiscoveryCodeStaysOutOfTheOverviewsOwnEnumeration.
//
// The operator-console enumeration is seven codes and its own test counts them
// (TestTheUnmeasuredReasonsStayApart). This screen adds one code that belongs to
// it alone, and adding it to that map instead would change a count the overview
// screen asserts about itself — an enumeration one screen can grow from another
// file is one nobody can read as complete.
func TestTheDiscoveryCodeStaysOutOfTheOverviewsOwnEnumeration(t *testing.T) {
	if _, ok := unmeasuredSentences[reasonDiscoveryUnreadable]; ok {
		t.Error("reasonDiscoveryUnreadable is in the overview's sentence map; it belongs to " +
			"the signals screen and composes its own sentence")
	}
	// And it still renders both halves, which is the property that map would have
	// provided.
	got := unmeasuredDiscovery("permission denied")
	if got.Code() != string(reasonDiscoveryUnreadable) {
		t.Errorf("code = %q, want %q", got.Code(), reasonDiscoveryUnreadable)
	}
	if !strings.Contains(got.Why(), "발굴 저장소 미판독") || !strings.Contains(got.Why(), "permission denied") {
		t.Errorf("why = %q, want the standing instruction and the specific failure", got.Why())
	}
	// An empty detail still gets the standing instruction rather than a bare dash.
	if strings.TrimSpace(unmeasuredDiscovery("").Why()) == "" {
		t.Error("a discovery failure with no detail renders a reason-less dash")
	}
}

// TestAPassCountThatIsNotZeroSaysWhichOfTheTwoThingsHappened.
//
// The structural-zero sentence asserts a zero, so it must not be the one printed
// beside a number that is not zero — a note contradicted by the figure next to it
// is a note the next reader stops checking against. Leaving the slot empty would
// be worse than either: the way this count becomes non-zero without a human having
// approved a threshold is somebody counting THRESHOLD_ABSENT as a pass, and that
// is the single repair D18 forbids. It must not be able to make the screen look
// tidier than it did before.
func TestAPassCountThatIsNotZeroSaysWhichOfTheTwoThingsHappened(t *testing.T) {
	page := renderSignals(t, newSignalsHarness(t, &stubSignals{
		reading: oneMarket(measuredAndClear("000660")),
	}))
	if strings.Contains(page, signalsPassedNote) {
		t.Error("a non-zero pass count is shown beside the sentence claiming it is structurally " +
			"zero; the two contradict each other on the same line")
	}
	if !strings.Contains(page, "유일하게 틀린 수리") {
		t.Error("a non-zero pass count is shown with no note at all; the way it becomes " +
			"non-zero without an approved threshold is the one repair D18 forbids, and that " +
			"must not be able to make the page quieter")
	}
}
