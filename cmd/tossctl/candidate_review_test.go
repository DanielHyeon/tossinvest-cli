package main

// candidate_review_test.go is the §5 review's remaining findings at the operator
// surface: the note that contradicts its own figure, the turn on which nothing was
// due, the store two read-only commands never closed, and the safety gate that
// skipped itself without saying so.

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"go/ast"
	"go/parser"
	"go/token"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/candidate"
	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/output"
)

var reportAt = time.Date(2026, 7, 28, 9, 0, 0, 0, time.UTC)

// renderReport is one CycleResult through the command's own rendering.
func renderReport(t *testing.T, format output.Format, res candidate.CycleResult) string {
	t.Helper()
	var buf bytes.Buffer
	if err := writeCandidateReport(&buf, format, res); err != nil {
		t.Fatalf("writeCandidateReport: %v", err)
	}
	return buf.String()
}

// --- the note that contradicted its own figure -----------------------------------

// TestANonZeroPassedCountIsNotShownBesideTheSentenceThatSaysItIsZero.
//
// passedNote asserts a zero — "structurally 0: seen_late and extended have no
// approved threshold" — and it was printed unconditionally, so a non-zero count
// would have rendered as `passed 7 (structurally 0: …)`. The console already
// solved this with a swapped sentence and its comment says why: a note
// contradicted by the figure next to it is a note the next reader stops checking
// against.
//
// The slot must not simply go empty either. The way this count becomes non-zero
// without a human approving a threshold is somebody counting THRESHOLD_ABSENT as
// a pass — the single repair D18 forbids — and that repair must not be able to
// make the output tidier than it was.
func TestANonZeroPassedCountIsNotShownBesideTheSentenceThatSaysItIsZero(t *testing.T) {
	res := candidate.CycleResult{
		Market: candidate.MarketKR,
		At:     reportAt,
		Vetoes: candidate.VetoTally{Total: 7, Passed: 7},
	}

	table := renderReport(t, output.FormatTable, res)
	if strings.Contains(table, "structurally 0") {
		t.Errorf("a passed count of 7 is printed beside the sentence claiming the count is "+
			"structurally zero; the two contradict each other on one line:\n%s", table)
	}
	if !strings.Contains(table, "wrong repair") {
		t.Errorf("a non-zero passed count is printed with no note at all; the way it becomes "+
			"non-zero without an approved threshold is the one repair D18 forbids:\n%s", table)
	}

	var report struct {
		Veto struct {
			Passed     int    `json:"passed"`
			PassedNote string `json:"passed_note"`
		} `json:"veto"`
	}
	if err := json.Unmarshal([]byte(renderReport(t, output.FormatJSON, res)), &report); err != nil {
		t.Fatalf("the JSON report does not parse: %v", err)
	}
	if strings.Contains(report.Veto.PassedNote, "structurally 0") {
		t.Errorf("the JSON carries the structurally-zero note beside passed = %d",
			report.Veto.Passed)
	}
	if report.Veto.PassedNote == "" {
		t.Error("the JSON drops the note entirely once the count is non-zero; a consumer then " +
			"has nothing telling it which of the two things happened")
	}

	// And the ordinary case still carries the sentence that keeps a zero from
	// reading as a failing system.
	zero := renderReport(t, output.FormatTable, candidate.CycleResult{
		Market: candidate.MarketKR, At: reportAt,
		Vetoes: candidate.VetoTally{Total: 7},
	})
	if !strings.Contains(zero, "structurally 0") {
		t.Errorf("a zero passed count lost the sentence that explains it:\n%s", zero)
	}
}

// --- the turn on which nothing was due --------------------------------------------

// TestTheReportSaysNothingWasDueRatherThanReportingAnEmptyPanel.
//
// A quiet turn renders `sources 0 attempted, 0 responded` otherwise, which is the
// shape an operator reads as an outage — and it is the shape the console renders
// as "every source answered" one level up. The number is honest and the sentence
// is what makes it readable: nobody was asked, nothing was lost.
func TestTheReportSaysNothingWasDueRatherThanReportingAnEmptyPanel(t *testing.T) {
	res := candidate.CycleResult{
		Market: candidate.MarketKR,
		At:     reportAt,
		Quiet:  true,
		NotDue: []candidate.SourceID{
			candidate.SourceOfficialTradingValue, candidate.SourceOfficialGainers,
		},
		Scan: candidate.ScanResult{Market: candidate.MarketKR, At: reportAt},
	}

	table := renderReport(t, output.FormatTable, res)
	if !strings.Contains(table, "nothing was due") {
		t.Errorf("a turn on which the schedule had nothing due renders as a panel that "+
			"attempted nothing, with no word for why:\n%s", table)
	}
	for _, id := range res.NotDue {
		if !strings.Contains(table, string(id)) {
			t.Errorf("the quiet turn does not name %s among the sources it passed over", id)
		}
	}

	var report struct {
		Quiet   bool `json:"quiet"`
		Sources struct {
			Attempted int      `json:"attempted"`
			NotDue    []string `json:"not_due"`
		} `json:"sources"`
	}
	if err := json.Unmarshal([]byte(renderReport(t, output.FormatJSON, res)), &report); err != nil {
		t.Fatalf("the JSON report does not parse: %v", err)
	}
	if !report.Quiet {
		t.Error("the JSON does not carry the fact that nothing was due; a consumer reading " +
			"attempted = 0 cannot tell a schedule from an outage")
	}
	if len(report.Sources.NotDue) != 2 {
		t.Errorf("not_due = %v, want both sources named", report.Sources.NotDue)
	}
}

// --- the safety gate that skipped itself ------------------------------------------

// TestWatchSaysSoWhenItCannotTellWhetherAVerificationIsRunning.
//
// The run-lock check was `if lockPath, lerr := candidateVerifyLockPath(root); lerr
// == nil`, so a path that could not be resolved skipped spec Requirement 7's gate
// entirely and printed nothing. Fail-open is the right direction — a discovery
// loop that cannot start because a data directory is misconfigured is worse than
// one that starts beside a verification — but a fail-open gate that says nothing
// is indistinguishable from a gate that ran and passed.
func TestWatchSaysSoWhenItCannotTellWhetherAVerificationIsRunning(t *testing.T) {
	// A relative data directory is refused by journal.DataDir, which is the
	// resolution verify.go and this command share.
	t.Setenv(journal.EnvDataDir, "relative/not/absolute")
	clk := clock.NewFake(reportAt)
	withCandidateFixture(t, clk, 100<<30, &fixtureSource{
		id:   candidate.SourceOfficialTradingValue,
		rows: []candidate.Row{discoveryRow("005930", 12, 100, "70000")},
	})

	out, errOut, err := runCandidateStreams(t, "candidate", "watch", "--market", "KR", "--cycles", "1")
	if err != nil {
		t.Fatalf("candidate watch: %v\n%s\n%s", err, out, errOut)
	}
	if !strings.Contains(errOut, "verification") {
		t.Errorf("the run-lock gate could not be checked and the command said nothing about "+
			"it; a gate that skips itself silently reads exactly like a gate that passed:\n%s",
			errOut)
	}
}

// --- one reducer for two screens ---------------------------------------------------

// TestTheConsoleDoesNotBuildTheDiscoveryTalliesItself.
//
// consoleSignalsMarket assembled VetoTally, CrossingTally and the shadow bands by
// hand, duplicating internal/candidate's own reducer. The two agree today and the
// cost is entirely in the future: a fourth shadow band added to one appears on the
// scan output and not on /signals, and a band missing from a page renders as a
// band nobody crossed rather than as one nobody counted.
//
// The claim is structural, so what is checked is the source. It is deliberately
// narrow — this says nothing about any other caller of the Tally constructors,
// only that the console wiring is not one of them.
func TestTheConsoleDoesNotBuildTheDiscoveryTalliesItself(t *testing.T) {
	const path = "console.go"
	file, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
	if err != nil {
		t.Fatalf("parsing %s: %v", path, err)
	}
	built := map[string]bool{"TallyVetoes": true, "TallyCrossings": true, "TallyBands": true}
	found := false
	ast.Inspect(file, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok || !built[sel.Sel.Name] {
			return true
		}
		found = true
		t.Errorf("%s builds the discovery tallies itself (%s); reduce the verdicts through "+
			"candidate.TallyVerdicts so the scan output and /signals cannot disagree about "+
			"which bands exist", path, sel.Sel.Name)
		return true
	})
	if !found {
		// The guard has to be looking at something: a walk that read no calls at all
		// would pass whatever the file said.
		if len(file.Decls) == 0 {
			t.Fatalf("%s parsed to no declarations; the guard is reading nothing", path)
		}
	}
}

// --- the store two read-only commands never closed --------------------------------

// TestTheDiscoveryCommandsCloseTheStoreTheyOpened.
//
// candidateWiring returned `func() {}` while its own doc comment described a
// cleanup that closes what the production factory opened. An interrupted `watch`
// therefore left an un-checkpointed write-ahead log beside the ledger — on the
// filesystem D16 says discovery must be the first to stop writing to, and after a
// command whose whole argument is that it cleans up after itself every turn.
func TestTheDiscoveryCommandsCloseTheStoreTheyOpened(t *testing.T) {
	for _, args := range [][]string{
		{"candidate", "scan", "--market", "KR"},
		{"candidate", "watch", "--market", "KR", "--cycles", "1"},
	} {
		t.Run(strings.Join(args[:2], " "), func(t *testing.T) {
			clk := clock.NewFake(reportAt)
			closes := withCandidateFixture(t, clk, 100<<30, &fixtureSource{
				id:   candidate.SourceOfficialTradingValue,
				rows: []candidate.Row{discoveryRow("005930", 12, 100, "70000")},
			})
			out, err := runCandidate(t, args...)
			if err != nil {
				t.Fatalf("%v: %v\n%s", args, err, out)
			}
			if closes() != 1 {
				t.Errorf("the command released the discovery store %d times, want 1. An "+
					"interrupted watch leaves an un-checkpointed write-ahead log beside the "+
					"order ledger", closes())
			}
		})
	}
}

// --- the screen's read runs under the caller's context ------------------------------

// TestTheSignalsSeamOpensTheStoreUnderTheCallersContext.
//
// The seam opened — and migrated — the discovery store with context.Background()
// on a page that reloads itself every fifteen seconds and whose own prose says it
// only reads. A request the browser has already abandoned still took the open, the
// schema check and every query behind it.
func TestTheSignalsSeamOpensTheStoreUnderTheCallersContext(t *testing.T) {
	seam := signalsFixtureStore(t, clock.NewFake(reportAt))
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if _, err := seam.Signals(ctx); err == nil {
		t.Error("a read whose context was already cancelled opened the store and migrated it " +
			"anyway; the request the work belongs to had gone")
	}
}

// TestTheSignalsSeamTalliesThroughTheSamePathAsAScan is the value half of the
// structural test above.
func TestTheSignalsSeamTalliesThroughTheSamePathAsAScan(t *testing.T) {
	seam := signalsFixtureStore(t, clock.NewFake(reportAt))
	ctx := context.Background()

	store, closeStore, err := seam.open(ctx)
	if err != nil {
		t.Fatalf("opening the fixture store: %v", err)
	}
	if err := store.RecordObservations(ctx, []candidate.Observation{{
		Market: candidate.MarketKR, Symbol: "005930",
		Source: candidate.SourceOfficialTradingValue, ObservedAt: reportAt,
		Reported: candidate.Reported{Rank: 12, RankTotal: 100, TradingValue: "1000000"},
	}}); err != nil {
		t.Fatalf("RecordObservations: %v", err)
	}
	if _, err := store.Promote(ctx, candidate.MarketKR, "005930", reportAt); err != nil {
		t.Fatalf("Promote: %v", err)
	}
	closeStore()

	got, err := seam.Signals(ctx)
	if err != nil {
		t.Fatalf("Signals: %v", err)
	}
	kr := got.Markets[0]
	vetoes, crossings, bands := candidate.TallyVerdicts(kr.Verdicts)
	if !reflect.DeepEqual(kr.Vetoes, vetoes) {
		t.Errorf("the seam's veto tally %+v differs from candidate.TallyVerdicts's %+v",
			kr.Vetoes, vetoes)
	}
	if kr.Crossings.Total != crossings.Total || kr.Crossings.Computed != crossings.Computed {
		t.Errorf("the seam's crossing tally %+v differs from %+v", kr.Crossings, crossings)
	}
	if len(kr.Bands) != len(bands) {
		t.Errorf("the seam reports %d shadow bands and the scan output reports %d; a band "+
			"missing from a page renders as one nobody crossed", len(kr.Bands), len(bands))
	}
	for code := range bands {
		if _, ok := kr.Bands[code]; !ok {
			t.Errorf("the seam has no %s band", code)
		}
	}
}

// --- a tick nobody could read is not the end of a session ---------------------------

// TestWatchKeepsGoingWhenAWholeTickCouldNotBeRead.
//
// runCandidateWatch's OnError returns false for everything, and the comment beside
// it explains why that is safe: "Collect already returns ErrNoSourceAnswered for
// the first, and Cycle only returns an error for the second". The second half was
// not true — Cycle propagated ErrNoSourceAnswered — so the one failure the comment
// singles out as survivable was the one that ended the session.
//
// It is reachable without a market outage: every source in the panel inside its own
// 429 backoff answers ErrHeldByBackoff, nobody read anything, and the loop that
// keeps first_seen_at alive stops during exactly the rate limit the ladder exists
// to ride out.
func TestWatchKeepsGoingWhenAWholeTickCouldNotBeRead(t *testing.T) {
	clk := clock.NewFake(reportAt)
	withCandidateFixture(t, clk, 100<<30, &fixtureSource{
		id:  candidate.SourceOfficialTradingValue,
		err: errors.New("the ranking endpoint is unreachable"),
	})

	out, errOut, err := runCandidateStreams(t, "candidate", "watch", "--market", "KR", "--cycles", "1")
	if err != nil {
		t.Fatalf("a tick on which no source answered ended the watch: %v\n%s\n%s. "+
			"A market that could not be read for one tick is not a reason to end a session; "+
			"a store that cannot be written is", err, out, errOut)
	}
	if !strings.Contains(errOut, "no source answered") {
		t.Errorf("the loop survived the tick and said nothing about it:\n%s", errOut)
	}
	if !strings.Contains(errOut, "still running") {
		t.Errorf("the message does not say the loop is continuing, so a reader cannot tell it "+
			"from the failure that ends one:\n%s", errOut)
	}
}
