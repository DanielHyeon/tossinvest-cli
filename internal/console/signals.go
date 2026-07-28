package console

// signals.go is the discovery screen (change add-candidate-discovery, tasks
// 5.5–5.7): what the ranking sources have been saying, over time, and — above
// everything else on the page — what nobody has checked.
//
// # It is read-only, and there is nothing on it to press
//
// GET, inside the session gate, outside the CSRF gate. No form, no POST, no
// button, and therefore no confirmation of any kind — no typed string, no second
// click, no extra approval (user instruction 2026-07-27, console-operator-overview
// D11). Discovery cannot place an order even in principle: internal/candidate's
// dependency closure is {internal/clock}, and this screen only reads what it
// wrote.
//
// # It spends no rate budget
//
// The seam reads the discovery store and calls no source. spec Requirement 7
// ranks the account's budget engine > verification > discovery, and a browser tab
// that reloaded itself into a second discoverer would be competing with
// `tossctl candidate watch` for an undocumented limit whose 429 costs the whole
// ranking source (D14 decision 2). The page therefore shows the last scan's work
// rather than causing one.
//
// # The one thing this screen exists to prevent
//
// A row with no veto reason on it means two different things — "all three were
// measured and none fired" and "nobody ever checked" — and under the candle
// budget the second is the ordinary case for most of the list, all day (D13
// decision 3). D10 is the rule: unmeasured is never a pass. So every one of the
// three vetoes gets its own cell with its own word in it, the row carries an
// explicit verdict rather than the absence of one, and the tally puts the
// unmeasured count above the passed count.
//
// Two further traps, both of which look like defects and are not:
//
//	the passed count is 0    Two of the three vetoes have no approved threshold
//	                         anywhere in this repository (D18), so Chase.Passed()
//	                         is unreachable and the tally's Passed is permanently
//	                         zero. That is what D18, D10 and the spec produce
//	                         together. The one wrong repair is to count
//	                         THRESHOLD_ABSENT as a pass — and the second wrong one
//	                         is to hide a column because it is always zero, which
//	                         would leave the next reader with no way to find out
//	                         why. The number is shown with the sentence.
//	the denominators differ  The acceleration tally counts SERIES, because D9 makes
//	                         the series (market, symbol, source): a WTS popularity
//	                         list carries no trading value and occupies a slot every
//	                         tick as FIGURE_ABSENT. The veto tally counts
//	                         CANDIDATES. D21 makes the distinction a contract, so
//	                         each number is labelled with the thing it counts and
//	                         neither is ever rendered under the other's word.
//
// # It reads the verdict through the verdict's own type
//
// candidate.VetoState's fields are unexported and its zero value is unmeasured,
// so no map miss and no unfilled slice slot here can produce a state that renders
// as measured and safe. This file therefore carries candidate.Chase itself rather
// than translating it into a console-local triple: a translation is where the
// three states become two. `!Dangerous()` is the one-line spelling D10 forbids —
// an unmeasured veto answers false to Dangerous and to Clear — and
// cmd/tossctl/candidate_test.go's TestNoConsumerReadsAVetoThroughItsDroppable
// SecondReturn walks this package to keep it, and the droppable helper pairs, out.

import (
	"context"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/candidate"
)

// --- the seam -------------------------------------------------------------------

// SignalsReader is the discovery store's read. One method.
//
// It returns an assessment rather than a store handle so that this package never
// holds the thing that writes: the store's own API can promote, cool and prune,
// and a console that held it would be one edit from a screen that "tidied up"
// first_seen_at. cmd/tossctl opens the store, runs candidate.Assess and hands the
// answer over.
//
// Nil leaves the screen unmeasured with the reason seam 미배선 and every other
// screen working — a build without discovery wired says so rather than showing an
// empty list, because an empty list reads as "the market is quiet".
type SignalsReader interface {
	Signals(ctx context.Context) (SignalsReading, error)
}

// SignalsReading is one read of the discovery store: every enabled market's
// assessment as of one instant.
type SignalsReading struct {
	// At is when the assessment was made. It is the store's clock rather than the
	// console's, so a page rendered from a stale reading can say how stale.
	At time.Time
	// Markets is one entry per market that was asked about. A market that could
	// not be assessed is present with Why set rather than absent: a market that
	// vanishes from the page is indistinguishable from a market with nothing in it.
	Markets []SignalsMarket
}

// SignalsMarket is one market's assessment.
type SignalsMarket struct {
	// Market is "KR" or "US".
	Market string
	// Why is why this market has no assessment. Empty when it has one.
	Why string
	// Verdicts is every live candidate's complete assessment, carrying
	// candidate.Chase itself. See the file header for why it is not translated.
	Verdicts []candidate.Verdict
	// Vetoes counts the verdicts: total, vetoed, unmeasured, passed. Its
	// denominator is CANDIDATES.
	Vetoes candidate.VetoTally
	// Crossings counts the shadow acceleration record. Its denominator is SERIES
	// — (market, symbol, source) — and it is therefore larger than the candidate
	// count (D21). The two are never rendered under the same word.
	Crossings candidate.CrossingTally
	// Bands is the shadow record for the two codes D18 left without a threshold.
	// It decides nothing.
	Bands map[candidate.VetoCode]candidate.BandTally
	// Sightings is the first-sighting record reduced by the source that produced
	// it. Vetoes counts the reasons and this says whose readings they came from,
	// which is the difference between "seen_late is mostly unmeasured" and "the
	// official trading-value ranking is coming back short".
	Sightings []candidate.SourceSightings
	// Panel is the source panel's completeness for the pass this reading knows
	// about.
	Panel SignalsPanel
}

// SignalsPanel is what the sources did on the pass this reading describes.
//
// Known is the absent-versus-zero flag and it is separate from Degraded on
// purpose: "no source is missing" and "nobody told us which sources are missing"
// render identically as a boolean, and only one of them means the panel was whole.
type SignalsPanel struct {
	// Known reports that a scan record reached this screen at all.
	Known bool
	// Why says what is missing when Known is false. It must not be empty then —
	// a reason-less dash is a display nobody can act on.
	Why string
	// At is when that pass ran.
	At time.Time
	// Attempted and Responded are D4's completeness.
	Attempted, Responded int
	// Degraded reports that at least one source did not answer.
	Degraded bool
	// Missing names every source that did not answer, and why. The names are the
	// point (task 5.7): an expired WTS session and a rate-limited official
	// ranking are the same boolean and opposite responses.
	Missing []candidate.SourceFailure
}

// --- the reason a discovery reading is missing ------------------------------------

// reasonDiscoveryUnreadable is one code this screen adds to the operator-console
// enumeration, and it is declared here rather than in overview.go because it
// belongs to this screen and to nothing else.
//
// None of the seven fits and two of them would be actively wrong advice.
// journal_unreadable says "start the engine so it migrates the ledger", and the
// discovery store is a different file with a different lifecycle — I-4 refused
// exactly that reuse, on the grounds that a wrong instruction is worse than no
// reason at all. broker_read_failed sends the operator to credentials, and this
// screen makes no request. seam_unwired says the build was not handed the
// capability, which is a different state this screen renders separately. What
// this one asks for is its own thing: look at the discovery store, and at whether
// anything has ever scanned into it.
//
// A reason that exists only as a free sentence cannot be counted and cannot be
// tested for absence (I-2, Manager ruling 2026-07-28), which is why this is a code
// rather than a bare sentence passed to unmeasuredBecause.
//
// Its sentence is composed by unmeasuredDiscovery below rather than added to
// overview.go's unmeasuredSentences map: that map is the seven-code enumeration
// the overview screen's own test counts, and a code belonging to one screen has no
// business changing that count.
const reasonDiscoveryUnreadable unmeasuredReason = "discovery_unreadable"

// unmeasuredDiscoverySentence is what reasonDiscoveryUnreadable tells the operator
// to do. It is never rendered without whatever specific failure came with it.
const unmeasuredDiscoverySentence = "발굴 저장소 미판독 — 이 화면은 발굴 저장소를 읽기만 한다. " +
	"저장소 파일과 권한을 확인하고, 아직 아무 스캔도 돌지 않았다면 `tossctl candidate scan`을 " +
	"한 번 돌린다."

// unmeasuredDiscovery is a discovery reading that could not be taken, with the
// specific failure beside the standing instruction.
//
// The code is what an operator greps the page for and what a test can assert the
// absence of; the detail is what says which file or which error. Neither replaces
// the other.
func unmeasuredDiscovery(detail string) reading {
	sentence := unmeasuredDiscoverySentence
	if trimmed := strings.TrimSpace(detail); trimmed != "" {
		sentence += " (" + trimmed + ")"
	}
	return unmeasuredWithDetail(reasonDiscoveryUnreadable, sentence)
}

// signalsPanelUnwired is the sentence beside a panel report this build cannot
// produce.
//
// The code is seam_unwired and that is the accurate one rather than a convenient
// one: ScanResult.Missing is the only thing that pairs a source id with the reason
// it did not answer, it lives for one cycle inside the process that ran the scan,
// and this console is not that process. It is the same shape as the overview's
// open-orders panel, which stayed seam_unwired by design until the change that
// wired it (console-operator-overview D7) — 이 칸은 0이 아니라 seam_unwired다.
const signalsPanelUnwired = "이 콘솔에 스캔 결과가 배선되어 있지 않다. 빠진 원천의 이름과 " +
	"사유는 스캔 결과에만 있으므로 `tossctl candidate scan`의 출력에서 확인한다. " +
	"\"강등 없음\"이 아니라 \"어느 원천이 빠졌는지 이 화면이 모른다\"는 뜻이다."

// signalsPanelWhy keeps an unmeasured panel from arriving without a sentence. The
// caller's own words come first when it has any, because they are the specific
// half.
func signalsPanelWhy(why string) string {
	if trimmed := strings.TrimSpace(why); trimmed != "" {
		return trimmed
	}
	return signalsPanelUnwired
}

// --- the view ---------------------------------------------------------------------

// signalsView is the whole screen.
type signalsView struct {
	// Now is when this page was rendered, and TakenAt is when the reading behind
	// it was made. They are separate because the second can be much older than the
	// first and the difference is the only thing that says so.
	Now time.Time
	// Read is the whole-page reading: measured when the seam answered, unmeasured
	// with a code when it did not.
	Read reading
	// Markets is one panel per market.
	Markets []signalsMarketView
	// TakenAt and AgeSeconds describe the reading's own instant. Meaningful only
	// when Read.Known.
	TakenAt    string
	AgeSeconds int
}

// NowText is the render instant, in the same UTC spelling every other screen uses.
func (v signalsView) NowText() string { return v.Now.UTC().Format("2006-01-02 15:04:05Z") }

// signalsMarketView is one market's block.
type signalsMarketView struct {
	Market string
	// Read is measured when this market was assessed. A market that failed keeps
	// its heading and says why, because a heading that disappears is not an
	// unmeasured marker.
	Read  reading
	Panel signalsPanelView
	Veto  signalsVetoTally
	Accel signalsAccelTally
	Bands []signalsBandTally
	// Sightings says which source's readings the first-sighting refusals came from.
	// The veto census says how many and why; this says whose, which is the
	// difference between a number to watch and a source to go and look at.
	Sightings []signalsSightingSource
	Rows      []signalsRow
}

// signalsSightingSource is one ranking source's first-sighting record.
type signalsSightingSource struct {
	Source   string
	Total    int
	Measured int
	// NotMeasured is the census of refusals for this source's readings, biggest
	// first — the same ordering the veto reasons use so the two blocks read alike.
	NotMeasured []signalsReasonCount
}

// --- one candidate ------------------------------------------------------------------

// signalsRow is one candidate as the table shows it.
type signalsRow struct {
	Symbol string
	// StateText is 활성 or 냉각. Expired candidates are not assessed at all.
	StateText string
	// FirstSeen is when this life began and AgeText how long ago. The instant is
	// the claim; the age is what makes it readable at a glance.
	FirstSeen string
	AgeText   string
	// Verdict is the row's own three-state summary. It is never empty — see
	// signalsVerdict.
	Verdict signalsVerdict
	// Vetoes is D3's three, in D3's order, each with its own word.
	Vetoes []signalsVeto
	// Sighting, Expansion and Distance are the measurements the vetoes were
	// computed from, each three-state so an unmeasured one keeps its own reason.
	Sighting  signalsMetric
	Expansion signalsMetric
	Distance  signalsMetric
	// Series and SeriesMeasured count this candidate's acceleration series and how
	// many of them produced a number. The label is 계열 wherever it is rendered.
	Series, SeriesMeasured int
	// Crossed lists the shadow thresholds at least one of this row's series
	// crossed. It records and decides nothing.
	Crossed string
	// SourcesText is who raised this candidate, and Completeness is D4's answered
	// over attempted. Degraded is the panel's own verdict for the pass that last
	// touched this row.
	SourcesText  string
	Completeness string
	Degraded     bool
}

// signalsVerdict is the row's summary, and it has exactly three shapes.
//
// It is a small struct rather than a string because the template has to be able to
// style the three differently, and because a row whose verdict is computed as "no
// veto fired" would be indistinguishable from a passed one in a string.
type signalsVerdict struct {
	// Vetoed, Passed and Unchecked are mutually exclusive and exactly one is true.
	Vetoed, Passed, Unchecked bool
	// Label is the word the table shows. Never empty.
	Label string
	// Detail says how many of the three were measured, which is what makes
	// "unchecked" actionable.
	Detail string
}

// signalsMetric is one measurement the veto reads, three-state.
//
// It is not the console's `reading` and that is deliberate. `reading` carries the
// operator-console enumeration — verify_suspended, seam_unwired, never_fetched and
// the rest — which names what the CONSOLE could not do. These cells name what the
// MEASUREMENT could not have: NO_DAY_HIGH is the candle budget working as designed,
// NO_BASELINE is a candidate older than its own raw rows, THRESHOLD_ABSENT is a
// veto nobody configured. Folding the two vocabularies into one type would produce
// a page where "seam 미배선" and "NO_DAY_HIGH" look like members of the same list,
// and an operator would go looking for the wiring.
//
// Reason is candidate's own code and is never empty when Measured is false: each
// of the three measurements answers Reason() rather than the raw field, and those
// answers are documented never to come back empty.
type signalsMetric struct {
	Text     string
	Measured bool
	Reason   string
}

// Value renders the number, or the em dash the reason accompanies.
func (m signalsMetric) Value() string {
	if !m.Measured {
		return "—"
	}
	return m.Text
}

// measuredMetric is a measurement that was taken.
func measuredMetric(text string) signalsMetric {
	return signalsMetric{Text: text, Measured: true}
}

// unmeasuredMetric is a measurement that was not, under the name of the input
// that was missing. An empty reason becomes a named one rather than a blank cell.
func unmeasuredMetric(reason string) signalsMetric {
	if strings.TrimSpace(reason) == "" {
		reason = "REASON_MISSING"
	}
	return signalsMetric{Reason: reason}
}

// signalsVeto is one of the three codes for one candidate.
type signalsVeto struct {
	Code string
	// Danger and Clear both require the measurement, which is the whole of D10.
	// Both false is unmeasured, and then Reason names the input that was missing.
	Danger, Clear bool
	Reason        string
}

// Label is the word this cell shows. There is no fourth state and no empty one.
func (v signalsVeto) Label() string {
	switch {
	case v.Danger:
		return "위험"
	case v.Clear:
		return "측정·안전"
	default:
		return "미측정"
	}
}

// Class is the cell's styling hook.
func (v signalsVeto) Class() string {
	switch {
	case v.Danger:
		return "bad"
	case v.Clear:
		return "ok"
	default:
		return "muted"
	}
}

// --- the tallies ----------------------------------------------------------------------

// signalsVetoTally is the veto counts, whose denominator is CANDIDATES.
type signalsVetoTally struct {
	Candidates int
	Unmeasured int
	Vetoed     int
	Passed     int
	// PassedNote is why Passed is structurally zero. It is rendered beside the
	// number and never instead of it.
	PassedNote string
	// Alarms is the tally's own arithmetic, failed. Empty is the ordinary state and
	// it is not a claim that the numbers are right, only that they are not provably
	// wrong — which is why these are rendered beside the counts rather than instead
	// of them.
	Alarms []string
	// Codes is the per-code raised/unmeasured breakdown, in D3's order.
	Codes []signalsCodeCount
	// Reasons is the missing-input census across every code, biggest first. It is
	// what tells an operator whether they are looking at the candle budget, a
	// stale hot queue or an unconfigured threshold.
	Reasons []signalsReasonCount
}

type signalsCodeCount struct {
	Code       string
	Raised     int
	Unmeasured int
}

type signalsReasonCount struct {
	Reason string
	Count  int
}

// signalsAccelTally is the shadow acceleration record, whose denominator is
// SERIES.
type signalsAccelTally struct {
	// Series is candidate.CrossingTally.Total, and it is not a candidate count.
	Series   int
	Measured int
	// SeriesNote and ShadowNote are D21 and the record-not-a-verdict label,
	// rendered beside the numbers rather than left to the template's own prose so
	// that the two screens carrying them cannot word the same rule differently.
	SeriesNote, ShadowNote string
	// Crossed is one entry per contract threshold, in the contract's order, so a
	// threshold nobody crossed still occupies its column.
	Crossed []signalsReasonCount
	// NotComputed is the census of why a series produced no number.
	NotComputed []signalsReasonCount
}

// signalsBandTally is one shadow band's record. Bands are histogram buckets, not
// verdicts (task 4.10), and the screen says so.
type signalsBandTally struct {
	Code     string
	Total    int
	Measured int
	// ShadowNote labels the record as a record. A band is a histogram bucket and
	// never a verdict (task 4.10), and the page has to say so where the numbers are.
	ShadowNote  string
	Crossed     []signalsReasonCount
	NotMeasured []signalsReasonCount
}

// signalsPanelView is the source panel's completeness, rendered.
type signalsPanelView struct {
	// Read is measured when a scan record reached the screen.
	Read reading
	// TakenAt is when that pass ran.
	TakenAt string
	// AnsweredText is "4 / 5" — answered over attempted.
	AnsweredText string
	Degraded     bool
	// Missing names the sources that did not answer, with the reason each gave.
	Missing []signalsMissingSource
	// Unnamed reports the state that must not pass silently: the pass says it was
	// degraded and named nothing. A degraded boolean with no names is the display
	// task 5.7 exists to remove, so the screen calls it out rather than rendering
	// an empty list.
	Unnamed bool
	// NothingAttempted reports the third state of a KNOWN panel: a pass that
	// attempted no source at all.
	//
	// Known was separated from Degraded because "no source is missing" and "nobody
	// told us which sources are missing" render identically as a boolean. This is
	// the same failure one step further in: a known panel with 0 / 0 on it is not
	// a panel where every source answered, and rendering it in the ok class puts
	// the page's strongest reassurance above numbers that were computed from
	// nothing (§5 review P1-2). It is produced by a scan cycle on which the
	// schedule had nothing due — the engine yield, a tick below every source's
	// interval, a clock that stepped backwards.
	//
	// It is deliberately not Degraded: nothing was lost, and an operator sent to
	// look for a failing source finds one that is working.
	NothingAttempted bool
}

type signalsMissingSource struct {
	Source string
	Reason string
	// RateLimited separates a 429 from everything else. For the official ranking
	// — which has no WTS fallback (D14) — one 429 is the whole source, and the
	// rate at which it happens is the only measurement of an undocumented limit.
	RateLimited bool
}

// --- assembly ---------------------------------------------------------------------

// signalsPassedNote is the sentence that stops a structurally zero count from
// reading as a broken system.
//
// D18 wrote the zero down as a consequence to be recorded rather than repaired:
// 둘이 THRESHOLD_ABSENT인 동안 Chase.Passed()는 도달 불가이고 집계의 Passed는 항상
// 0이다. 이것은 고장이 아니다. A screen showing "통과 0" on its own sends the next
// reader looking for a bug, and the most natural bug to "fix" is the one D18 names
// as the only wrong repair.
const signalsPassedNote = "구조적으로 0이다 — seen_late와 extended에는 승인된 임계가 없어" +
	"(design.md D18) 세 사유가 모두 측정·안전인 후보가 나올 수 없다. 고장이 아니며, " +
	"THRESHOLD_ABSENT를 통과로 세는 것이 유일하게 틀린 수리다."

// signalsPassedUnexpected is what the same slot says when the count is NOT zero.
//
// The sentence above asserts a zero, so it cannot be the one shown beside a
// non-zero number — a note contradicted by the figure next to it is a note the
// next reader stops checking against. But leaving the slot empty would be worse
// than either: the way the count becomes non-zero without a threshold being
// approved is somebody counting THRESHOLD_ABSENT as a pass, which is the single
// repair D18 forbids, and that repair must not be able to make the screen look
// tidier than before. So the slot always says something, and when the number moves
// it says which of the two things happened.
const signalsPassedUnexpected = "0이 아니다 — D18 이후 seen_late·extended의 임계가 사람 승인을 " +
	"받았거나, 아니면 THRESHOLD_ABSENT를 통과로 센 것이다. 후자가 D18이 지목한 유일하게 " +
	"틀린 수리이므로 먼저 그쪽을 확인한다."

// signalsTallyAlarm is one arithmetic contradiction as this screen words it.
//
// The judgement lives in candidate.VetoTally.Anomalies and is not repeated here.
// That split is the lesson from the fourth shadow band: a rule implemented twice
// eventually disagrees with itself, and the disagreement shows up as a page that
// looks calm. What each surface owns is the sentence, because the readers differ —
// this one is a browser page in Korean, the scan output is a terminal in English,
// and the numbers under both are the same numbers.
//
// It is rendered beside the counts rather than instead of them: the counts are what
// the alarm is about, and a screen that replaced them would hide the evidence.
func signalsTallyAlarm(a candidate.TallyAnomaly) string {
	switch a.Kind {
	case candidate.TallyPassedWithoutThreshold:
		return "경보 — " + a.Arithmetic() +
			". 임계가 없는 veto는 clear가 될 수 없으므로, 그 옆의 통과는 THRESHOLD_ABSENT를 " +
			"통과로 센 것이다. D18이 지목한 유일하게 틀린 수리다."
	default:
		return "경보 — " + a.Arithmetic() +
			". 세 버킷은 구조적으로 서로소이므로 합이 전체를 넘었다는 것은 " +
			"미측정 veto를 가진 후보가 통과로 집계됐다는 뜻이다. 미측정은 통과가 아니다(D10)."
	}
}

// signalsSeriesNote is D21 as a sentence on the page.
const signalsSeriesNote = "분모는 후보가 아니라 계열이다 — 계열은 (시장·종목·원천)이고, " +
	"WTS 인기 순위처럼 거래대금을 싣지 않는 원천도 매 tick 한 칸을 차지한다. " +
	"계열 수를 후보 수로 읽으면 안 된다."

// signalsShadowNote labels a record as a record.
const signalsShadowNote = "기록만 하고 아무것도 판정하지 않는다. 임계는 여기서 도출된다."

// signals builds the screen.
func (c *Console) signals(ctx context.Context) signalsView {
	v := signalsView{Now: c.now()}
	if c.opts.Signals == nil {
		v.Read = unmeasuredFor(reasonSeamUnwired)
		return v
	}
	got, err := c.opts.Signals.Signals(ctx)
	if err != nil {
		v.Read = unmeasuredDiscovery(err.Error())
		return v
	}
	v.Read = measured("")
	if !got.At.IsZero() {
		v.TakenAt = got.At.UTC().Format("2006-01-02 15:04:05Z")
		v.AgeSeconds = int(v.Now.Sub(got.At).Round(time.Second) / time.Second)
	}
	for _, m := range got.Markets {
		v.Markets = append(v.Markets, signalsMarketFrom(m, v.Now))
	}
	return v
}

// signalsMarketFrom assembles one market's block.
func signalsMarketFrom(m SignalsMarket, now time.Time) signalsMarketView {
	out := signalsMarketView{Market: m.Market, Read: measured("")}
	if why := strings.TrimSpace(m.Why); why != "" {
		out.Read = unmeasuredDiscovery(why)
		return out
	}
	out.Panel = signalsPanelFrom(m.Panel)
	out.Veto = signalsVetoTallyFrom(m.Vetoes)
	out.Accel = signalsAccelTallyFrom(m.Crossings)
	out.Bands = signalsBandTalliesFrom(m.Bands)
	out.Sightings = signalsSightingSourcesFrom(m.Sightings)
	for _, verdict := range m.Verdicts {
		out.Rows = append(out.Rows, signalsRowFrom(verdict, now))
	}
	sortSignalsRows(out.Rows)
	return out
}

// signalsRowFrom turns one assessment into one line.
func signalsRowFrom(v candidate.Verdict, now time.Time) signalsRow {
	row := signalsRow{
		Symbol:       v.Summary.Symbol,
		StateText:    signalsStateText(v.Summary.State),
		SourcesText:  signalsSourcesText(v.Summary.Sources),
		Completeness: strconv.Itoa(v.Summary.SourcesResponded) + " / " + strconv.Itoa(v.Summary.SourcesAttempted),
		Degraded:     v.Summary.Degraded,
	}
	if !v.Summary.FirstSeenAt.IsZero() {
		row.FirstSeen = v.Summary.FirstSeenAt.UTC().Format("2006-01-02 15:04:05Z")
		row.AgeText = signalsAgeText(now.Sub(v.Summary.FirstSeenAt))
	}
	row.Verdict = signalsVerdictFrom(v.Chase)
	row.Vetoes = signalsVetoesFrom(v.Chase)
	row.Sighting = signalsSightingMetric(v.Sighting)
	row.Expansion = signalsExpansionMetric(v.Expansion)
	row.Distance = signalsDistanceMetric(v.Range)

	crossed := map[string]bool{}
	for _, a := range v.Accelerations {
		row.Series++
		if a.Computed() {
			row.SeriesMeasured++
		}
		for _, threshold := range candidate.ShadowThresholds {
			if a.Crossed(threshold) {
				crossed[threshold] = true
			}
		}
	}
	var names []string
	for _, threshold := range candidate.ShadowThresholds {
		if crossed[threshold] {
			names = append(names, threshold)
		}
	}
	row.Crossed = strings.Join(names, " · ")
	return row
}

// signalsVerdictFrom is the row's three-state summary.
//
// The order of the three questions is the whole point and it is candidate.Chase's
// own: vetoed first, then whether every one of the three was measured, and only
// then a pass. The natural spelling — "nothing fired, therefore it passed" — is
// true of a verdict nobody computed, and under the candle budget that is most of
// the list.
func signalsVerdictFrom(c candidate.Chase) signalsVerdict {
	measuredCount := len(candidate.VetoCodes) - len(c.NotMeasured())
	detail := strconv.Itoa(measuredCount) + " / " + strconv.Itoa(len(candidate.VetoCodes)) + " 사유 측정"
	switch {
	case c.Vetoed():
		return signalsVerdict{Vetoed: true, Label: "거부", Detail: detail}
	case c.Passed():
		return signalsVerdict{Passed: true, Label: "통과", Detail: detail}
	default:
		return signalsVerdict{Unchecked: true, Label: "미확인", Detail: detail}
	}
}

// signalsVetoesFrom renders D3's three, in D3's order.
//
// Both predicates require the measurement, so an unmeasured veto is neither and
// falls to the third branch with its reason attached. The absent fourth branch is
// deliberate: candidate.Chase.State answers an unrecognised code as unmeasured, so
// a code added to VetoCodes without a field lands here as 미측정 rather than as a
// blank cell.
func signalsVetoesFrom(c candidate.Chase) []signalsVeto {
	out := make([]signalsVeto, 0, len(candidate.VetoCodes))
	for _, code := range candidate.VetoCodes {
		state := c.State(code)
		cell := signalsVeto{Code: string(code)}
		switch {
		case state.Dangerous():
			cell.Danger = true
		case state.Clear():
			cell.Clear = true
		default:
			cell.Reason = string(state.Reason())
		}
		out = append(out, cell)
	}
	return out
}

// signalsSightingMetric renders the first sighting's list position.
//
// The new-entrant marker is three-state now and this is the first build in which
// any of the three can appear. The field was declared on five types and assigned by
// no production source, so `if s.NewlyListed` was structurally false and the marker
// could not be drawn — the screen said nothing about a fact the record claimed to
// keep, and nothing failed. Rendering the negative and the unknown as well as the
// positive is what makes the difference visible: a cell that says nothing is what
// the broken version looked like.
func signalsSightingMetric(s candidate.Sighting) signalsMetric {
	if !s.Measured {
		return unmeasuredMetric(string(s.Reason()))
	}
	text := strconv.Itoa(s.Rank) + " / " + strconv.Itoa(s.RankTotal)
	if s.PercentilePct != "" {
		text += " (" + s.PercentilePct + "%p)"
	}
	return measuredMetric(text + " · " + signalsNewlyListedText(s.NewlyListed))
}

// signalsNewlyListedText spells the three states of the new-entrant fact.
//
// The unknown arm is not dead code by construction — MeasureFirstSighting refuses
// an unknown fact, so a Sighting reaching here measured has a known one today — but
// it is what any other producer of a Sighting would land on, and a blank cell there
// would be indistinguishable from the shipped defect.
func signalsNewlyListedText(n candidate.NewlyListed) string {
	switch {
	case n.Yes():
		return "신규 진입"
	case n.No():
		return "직전 읽기에 있었음"
	default:
		return "신규 진입 미상"
	}
}

// signalsExpansionMetric renders the expansion from the stored baseline.
func signalsExpansionMetric(e candidate.Expansion) signalsMetric {
	if !e.Measured {
		return unmeasuredMetric(string(e.Reason()))
	}
	return measuredMetric(e.GainPct + "%")
}

// signalsDistanceMetric renders the distance to the day's high.
//
// Smaller is more dangerous: near_high fires BELOW the floor, so this is the one
// column on the page where a small number is the alarming one.
func signalsDistanceMetric(r candidate.RangePosition) signalsMetric {
	if !r.Measured {
		return unmeasuredMetric(string(r.Reason()))
	}
	return measuredMetric(r.DistancePct + "%")
}

// signalsVetoTallyFrom counts CANDIDATES.
func signalsVetoTallyFrom(t candidate.VetoTally) signalsVetoTally {
	out := signalsVetoTally{
		Candidates: t.Total,
		Unmeasured: t.Unmeasured,
		Vetoed:     t.Vetoed,
		Passed:     t.Passed,
		PassedNote: signalsPassedNote,
	}
	if t.Passed != 0 {
		out.PassedNote = signalsPassedUnexpected
	}
	for _, a := range t.Anomalies() {
		out.Alarms = append(out.Alarms, signalsTallyAlarm(a))
	}
	for _, code := range candidate.VetoCodes {
		out.Codes = append(out.Codes, signalsCodeCount{
			Code:       string(code),
			Raised:     t.Raised[code],
			Unmeasured: t.NotMeasured[code],
		})
	}
	counts := map[string]int{}
	for why, n := range t.Reasons {
		counts[string(why)] = n
	}
	out.Reasons = sortedSignalsCounts(counts)
	return out
}

// signalsAccelTallyFrom counts SERIES, and the field name says so.
func signalsAccelTallyFrom(t candidate.CrossingTally) signalsAccelTally {
	out := signalsAccelTally{
		Series: t.Total, Measured: t.Computed,
		SeriesNote: signalsSeriesNote, ShadowNote: signalsShadowNote,
	}
	for _, threshold := range candidate.ShadowThresholds {
		out.Crossed = append(out.Crossed, signalsReasonCount{
			Reason: threshold, Count: t.Crossed[threshold],
		})
	}
	counts := map[string]int{}
	for why, n := range t.NotComputed {
		counts[string(why)] = n
	}
	out.NotComputed = sortedSignalsCounts(counts)
	return out
}

// signalsBandTalliesFrom renders the shadow bands in D3's code order, so the page
// does not reorder itself between refreshes as a map would.
func signalsBandTalliesFrom(in map[candidate.VetoCode]candidate.BandTally) []signalsBandTally {
	var out []signalsBandTally
	for _, code := range candidate.VetoCodes {
		tally, ok := in[code]
		if !ok {
			continue
		}
		block := signalsBandTally{
			Code: string(code), Total: tally.Total, Measured: tally.Measured,
			ShadowNote: signalsShadowNote,
		}
		for _, band := range candidate.BandsFor(code) {
			block.Crossed = append(block.Crossed, signalsReasonCount{
				Reason: band, Count: tally.Crossed[band],
			})
		}
		counts := map[string]int{}
		for why, n := range tally.NotMeasured {
			counts[string(why)] = n
		}
		block.NotMeasured = sortedSignalsCounts(counts)
		out = append(out, block)
	}
	return out
}

// signalsSightingSourcesFrom renders the per-source first-sighting record.
//
// The reduction is candidate.TallySightingSources and it is not repeated here, for
// the reason the veto tally is not repeated either: the scan output and this page
// have to be able to disagree about wording and never about counts.
func signalsSightingSourcesFrom(in []candidate.SourceSightings) []signalsSightingSource {
	var out []signalsSightingSource
	for _, s := range in {
		counts := map[string]int{}
		for why, n := range s.NotMeasured {
			counts[string(why)] = n
		}
		out = append(out, signalsSightingSource{
			Source: string(s.Source), Total: s.Total, Measured: s.Measured,
			NotMeasured: sortedSignalsCounts(counts),
		})
	}
	return out
}

// signalsPanelFrom renders the source panel's completeness.
func signalsPanelFrom(p SignalsPanel) signalsPanelView {
	if !p.Known {
		return signalsPanelView{Read: unmeasuredWithDetail(reasonSeamUnwired, signalsPanelWhy(p.Why))}
	}
	out := signalsPanelView{
		Read:         measured(""),
		AnsweredText: strconv.Itoa(p.Responded) + " / " + strconv.Itoa(p.Attempted),
		Degraded:     p.Degraded,
		// A pass that attempted nothing answered a question nobody asked. See
		// signalsPanelView.NothingAttempted — the branch exists so that 0 / 0 can
		// never reach the "전 원천 응답" sentence.
		NothingAttempted: p.Attempted <= 0,
	}
	if !p.At.IsZero() {
		out.TakenAt = p.At.UTC().Format("2006-01-02 15:04:05Z")
	}
	for _, m := range p.Missing {
		out.Missing = append(out.Missing, signalsMissingSource{
			Source:      string(m.Source),
			Reason:      signalsFailureReason(m.Reason),
			RateLimited: m.RateLimited,
		})
	}
	// Degraded with nothing named is the reason-less boolean task 5.7 removes. It
	// is reported as its own defect rather than rendered as an empty list, because
	// an empty list under a "강등" heading reads as "and it recovered".
	out.Unnamed = p.Degraded && len(out.Missing) == 0
	return out
}

// signalsFailureReason keeps a missing source from being named without a reason.
// The pair is the whole of task 5.7 — an expired session and a 429 are the same
// boolean and opposite responses.
func signalsFailureReason(why string) string {
	if trimmed := strings.TrimSpace(why); trimmed != "" {
		return trimmed
	}
	return "사유 없음 — 원천이 실패 이유를 남기지 않았다"
}

// --- small renderings ---------------------------------------------------------------

func signalsStateText(s candidate.State) string {
	switch s {
	case candidate.StateActive:
		return "활성"
	case candidate.StateCooled:
		return "냉각"
	case candidate.StateExpired:
		// Assess skips expired candidates, so this is a wiring defect rather than a
		// state. It is named instead of being rendered as a blank cell.
		return "만료 — 평가 대상이 아니다"
	default:
		return "상태 불명"
	}
}

func signalsSourcesText(ids []candidate.SourceID) string {
	if len(ids) == 0 {
		return "없음 — 이 후보를 올린 원천이 기록되지 않았다"
	}
	names := make([]string, 0, len(ids))
	for _, id := range ids {
		names = append(names, string(id))
	}
	return strings.Join(names, ", ")
}

// signalsAgeText renders an elapsed duration in whole units.
func signalsAgeText(d time.Duration) string {
	switch {
	case d < 0:
		return "미래 시각"
	case d < time.Minute:
		return strconv.Itoa(int(d/time.Second)) + "초 전"
	case d < time.Hour:
		return strconv.Itoa(int(d/time.Minute)) + "분 전"
	default:
		return strconv.Itoa(int(d/time.Hour)) + "시간 전"
	}
}

// sortedSignalsCounts renders a census deterministically, biggest first then by
// name. Zeroes are dropped: a census is what did happen.
func sortedSignalsCounts(counts map[string]int) []signalsReasonCount {
	var out []signalsReasonCount
	for name, n := range counts {
		if n == 0 {
			continue
		}
		out = append(out, signalsReasonCount{Reason: name, Count: n})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Count != out[j].Count {
			return out[i].Count > out[j].Count
		}
		return out[i].Reason < out[j].Reason
	})
	return out
}

// sortSignalsRows puts the rows in a fixed order: what fired, then what nobody
// checked, then what cleared, then by symbol.
//
// The order is not the argument this screen makes — the labels are — but it has to
// be deterministic, or a page that reloads every discovery tick shuffles under
// somebody reading it.
func sortSignalsRows(rows []signalsRow) {
	sort.SliceStable(rows, func(i, j int) bool {
		a, b := signalsVerdictRank(rows[i].Verdict), signalsVerdictRank(rows[j].Verdict)
		if a != b {
			return a < b
		}
		return rows[i].Symbol < rows[j].Symbol
	})
}

func signalsVerdictRank(v signalsVerdict) int {
	switch {
	case v.Vetoed:
		return 0
	case v.Unchecked:
		return 1
	default:
		return 2
	}
}

// --- the page -----------------------------------------------------------------------

type signalsPage struct {
	Nav  string
	Snap signalsView
}

// Refresh and RefreshSeconds: the screen reloads at discovery's own tick, derived
// from the contract constant rather than written again. Every reload costs zero
// broker calls and zero source reads — it is a read of the discovery store — so
// the period is a freshness bound rather than a budget one, and matching the tick
// keeps the page from showing the same pass twice or skipping one.
func (signalsPage) Refresh() bool { return true }
func (signalsPage) RefreshSeconds() int {
	return int(candidate.DefaultWatchInterval / time.Second)
}

// registerSignals puts the discovery screen on the mux.
//
// The path is /signals, matching StockOS's (user instruction 2026-07-28). It
// registers from this file rather than from console.go's routes() for the reason
// console-operator-overview established: the route table's static guards have to
// see a route wherever it is declared.
//
// It is not wrapped in `readOnly`. That wrapper is load-bearing exactly once — on
// /orders, the one path the account-verb ban grants an exception to — and
// console.go says in writing why putting it on every read route would make the
// place it matters invisible. This path names no account verb, carries no form and
// has no POST handler beside it, which is what the ban and the CSRF pairing test
// already assert about it.
func (c *Console) registerSignals(mux *http.ServeMux) {
	mux.HandleFunc("/signals", c.session0(c.handleSignals))
}

// handleSignals renders the screen. GET, inside the session gate and outside the
// CSRF gate — there is nothing here to submit.
func (c *Console) handleSignals(w http.ResponseWriter, r *http.Request) {
	c.render(w, "signals", signalsPage{Nav: "signals", Snap: c.signals(r.Context())})
}
