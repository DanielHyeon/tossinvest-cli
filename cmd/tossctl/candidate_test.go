package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/candidate"
	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/runlock"
	"github.com/spf13/cobra"
)

// candidate_test.go covers the two read-only discovery commands.
//
// The theme of the assertions is what the output must not let a reader conclude.
// Under D18 two of the three vetoes have no approved threshold, so no candidate can
// ever be counted as having passed — and a screen that renders an empty reason
// column as calm reproduces, one layer up, exactly the failure the store spent four
// sections closing.

// --- fixtures --------------------------------------------------------------------

// fixtureSource is a panel source under the test's control.
type fixtureSource struct {
	id   candidate.SourceID
	rows []candidate.Row
	err  error
}

func (f *fixtureSource) ID() candidate.SourceID { return f.id }

func (f *fixtureSource) Read(context.Context, string) (candidate.Reading, error) {
	if f.err != nil {
		return candidate.Reading{}, f.err
	}
	return candidate.Reading{Rows: f.rows}, nil
}

// withCandidateFixture points both discovery seams at a temporary store and a
// panel the test supplies, and restores them afterwards.
//
// The returned function reports how many times the command under test released the
// store. The fixture owns the handle — it opened it and it closes it — so the
// release it hands out does nothing except count, which is what lets a test tell
// "the command closed what it was given" apart from "the fixture cleaned up".
func withCandidateFixture(t *testing.T, clk clock.Clock, free int64, sources ...candidate.Source) func() int {
	t.Helper()
	dir := t.TempDir()
	store, err := candidate.Open(context.Background(), candidate.Options{
		Path:  filepath.Join(dir, candidate.DBFileName),
		Clock: clk,
		FSProber: candidate.FixedFSProber(candidate.FSInfo{
			Name: "ext4", FreeBytes: free, FreeMeasured: true,
		}),
	})
	if err != nil {
		t.Fatalf("opening the fixture store: %v", err)
	}
	t.Cleanup(func() { _ = store.Close() })

	openStore, panelFor := candidateStoreFactory, candidatePanelFactory
	t.Cleanup(func() { candidateStoreFactory, candidatePanelFactory = openStore, panelFor })

	var mu sync.Mutex
	releases := 0
	candidateStoreFactory = func(context.Context, *rootOptions) (*candidate.Store, func(), error) {
		return store, func() {
			mu.Lock()
			defer mu.Unlock()
			releases++
		}, nil
	}
	candidatePanelFactory = func(*rootOptions, string) ([]candidate.Source, error) {
		return sources, nil
	}
	return func() int {
		mu.Lock()
		defer mu.Unlock()
		return releases
	}
}

// runCandidate executes a discovery subcommand and returns its stdout.
func runCandidate(t *testing.T, args ...string) (string, error) {
	t.Helper()
	out, _, err := runCandidateStreams(t, args...)
	return out, err
}

// runCandidateStreams is runCandidate with the command's stderr as well.
//
// Two of the things these commands owe an operator are written there rather than
// to stdout — a cycle that failed, and a safety gate that could not be checked —
// and a test that only reads stdout cannot tell either of them from silence.
func runCandidateStreams(t *testing.T, args ...string) (string, string, error) {
	t.Helper()
	root := newRootCmd()
	var out, errOut bytes.Buffer
	root.SetOut(&out)
	root.SetErr(&errOut)
	root.SetArgs(args)
	err := root.ExecuteContext(context.Background())
	return out.String(), errOut.String(), err
}

func discoveryRow(symbol string, rank, total int, price string) candidate.Row {
	return candidate.Row{
		Symbol: symbol, Rank: rank, RankTotal: total,
		TradingValue: "1000000", TradingVolume: "500", Price: price,
	}
}

// --- the commands are read-only --------------------------------------------------

// TestTheDiscoveryCommandsDeclareThemselvesReadOnly.
//
// tasks.md 5.1 and 5.2 say `mutating: false`, and the annotation is what
// help_convention_test.go's inventory of trade actions checks against. A discovery
// command that ever grew the ability to place something would have to change this
// line, and the change would be the review's subject rather than a detail inside a
// diff.
func TestTheDiscoveryCommandsDeclareThemselvesReadOnly(t *testing.T) {
	for _, path := range []string{"tossctl candidate scan", "tossctl candidate watch"} {
		cmd := findCommandPath(t, newRootCmd(), path)
		if got := cmd.Annotations["mutating"]; got == "true" {
			t.Errorf("%s is annotated mutating=%q; discovery reads and records and does "+
				"nothing else", path, got)
		}
		if cmd.Annotations["source"] == "" {
			t.Errorf("%s has no source annotation", path)
		}
	}
}

func findCommandPath(t *testing.T, root *cobra.Command, path string) *cobra.Command {
	t.Helper()
	var found *cobra.Command
	var walk func(*cobra.Command)
	walk = func(c *cobra.Command) {
		if c.CommandPath() == path {
			found = c
		}
		for _, sub := range c.Commands() {
			walk(sub)
		}
	}
	walk(root)
	if found == nil {
		t.Fatalf("no command at %q", path)
	}
	return found
}

// --- task 5.1: what one scan has to print ----------------------------------------

// TestTheScanOutputSeparatesUnmeasuredFromPassed is spec Requirement 8 at the
// command line.
//
// 사유가 하나도 표시되지 않는다는 이유로 통과처럼 보이게 해서는 안 된다. Under D18
// the passed count is structurally zero and the unmeasured count is everything, and
// the output has to say which of those two facts a reader is looking at — a bare
// "0 vetoed" would be read as "all clear" and it means "we checked almost nothing".
func TestTheScanOutputSeparatesUnmeasuredFromPassed(t *testing.T) {
	clk := clock.NewFake(time.Date(2026, 7, 28, 9, 0, 0, 0, time.UTC))
	withCandidateFixture(t, clk, 100<<30, &fixtureSource{
		id: candidate.SourceOfficialTradingValue,
		rows: []candidate.Row{
			discoveryRow("005930", 12, 100, "70000"),
			discoveryRow("000660", 90, 100, "20000"),
		},
	})

	out, err := runCandidate(t, "candidate", "scan", "--market", "KR")
	if err != nil {
		t.Fatalf("candidate scan: %v\n%s", err, out)
	}
	for _, want := range []string{"unmeasured", "THRESHOLD_ABSENT", "passed"} {
		if !strings.Contains(out, want) {
			t.Errorf("the scan output does not mention %q:\n%s", want, out)
		}
	}
	// The words are kept apart on purpose (tasks.md 5.1): a shadow crossing is not
	// a veto that was passed.
	if !strings.Contains(out, "shadow") {
		t.Errorf("the scan output does not label the shadow record:\n%s", out)
	}
	// And it says why the passed count is zero, rather than leaving a reader to
	// conclude that everything failed.
	if !strings.Contains(out, "no approved threshold") {
		t.Errorf("the scan output reports a zero passed count without explaining that two "+
			"of the three vetoes have no threshold yet:\n%s", out)
	}
}

// TestTheScanJSONReportsTheCountsAnOperatorActsOn.
//
// The same facts in the machine-readable shape, because the numbers this change
// produces are the input to T3.2's threshold decision and nobody is going to derive
// a distribution by reading a table.
func TestTheScanJSONReportsTheCountsAnOperatorActsOn(t *testing.T) {
	clk := clock.NewFake(time.Date(2026, 7, 28, 9, 0, 0, 0, time.UTC))
	withCandidateFixture(t, clk, 100<<30, &fixtureSource{
		id:   candidate.SourceOfficialTradingValue,
		rows: []candidate.Row{discoveryRow("005930", 12, 100, "70000")},
	})

	out, err := runCandidate(t, "candidate", "scan", "--market", "KR", "--output", "json")
	if err != nil {
		t.Fatalf("candidate scan --output json: %v\n%s", err, out)
	}
	var report struct {
		Market string `json:"market"`
		Veto   struct {
			Total       int            `json:"total"`
			Passed      int            `json:"passed"`
			Unmeasured  int            `json:"unmeasured"`
			Reasons     map[string]int `json:"reasons"`
			PassedNote  string         `json:"passed_note"`
			NotMeasured map[string]int `json:"not_measured"`
		} `json:"veto"`
		ShadowAcceleration struct {
			Crossed map[string]int `json:"crossed"`
		} `json:"shadow_acceleration"`
		// crossed is a list and not a map: encoding/json sorts map keys as strings,
		// and the extended scale is numeric with ten edges, so the map form emitted
		// `"0","10","100","2",…`. See bandCount.
		ShadowBands map[string]struct {
			Total   int `json:"total"`
			Crossed []struct {
				Band  string `json:"band"`
				Count int    `json:"count"`
			} `json:"crossed"`
		} `json:"shadow_bands"`
	}
	if err := json.Unmarshal([]byte(out), &report); err != nil {
		t.Fatalf("the scan output is not JSON: %v\n%s", err, out)
	}
	if report.Market != "KR" {
		t.Errorf("market = %q, want KR", report.Market)
	}
	if report.Veto.Total != 1 || report.Veto.Unmeasured != 1 {
		t.Errorf("veto tally = %+v, want one candidate, unmeasured", report.Veto)
	}
	if report.Veto.Passed != 0 {
		t.Errorf("passed = %d, want 0 — an absent threshold is not a pass", report.Veto.Passed)
	}
	if report.Veto.PassedNote == "" {
		t.Error("the JSON reports a zero passed count with no note saying why it is " +
			"structurally zero; a consumer will read it as a failing system")
	}
	if report.Veto.Reasons["THRESHOLD_ABSENT"] == 0 {
		t.Errorf("reasons = %v, want THRESHOLD_ABSENT counted", report.Veto.Reasons)
	}
	if len(report.ShadowAcceleration.Crossed) != len(candidate.ShadowThresholds) {
		t.Errorf("shadow acceleration keys = %v, want all of %v",
			report.ShadowAcceleration.Crossed, candidate.ShadowThresholds)
	}
	for _, code := range []string{"seen_late", "extended"} {
		band, ok := report.ShadowBands[code]
		if !ok {
			t.Fatalf("no shadow band record for %s in %s", code, out)
		}
		if band.Total != 1 {
			t.Errorf("%s band total = %d, want 1", code, band.Total)
		}
		want := candidate.BandsFor(candidate.VetoCode(code))
		if len(band.Crossed) != len(want) {
			t.Fatalf("%s crossed has %d entries, want %d", code, len(band.Crossed), len(want))
		}
		for i, entry := range band.Crossed {
			if entry.Band != want[i] {
				t.Errorf("%s crossed[%d] = %q, want %q — the scale's order is the order a "+
					"distribution is read down, and it is the one thing a JSON object of "+
					"numeric keys cannot carry", code, i, entry.Band, want[i])
			}
		}
	}
}

// TestTheScanOutputNamesTheMissingSourcesAndTheRetreat is spec Requirement 8's last
// clause and task 5.4 together: 사유 없는 불리언 하나는 대응할 수 없는 표시다, and
// a retreat nobody wrote down is indistinguishable from a market that went quiet.
func TestTheScanOutputNamesTheMissingSourcesAndTheRetreat(t *testing.T) {
	clk := clock.NewFake(time.Date(2026, 7, 28, 9, 0, 0, 0, time.UTC))
	withCandidateFixture(t, clk, 100<<30,
		&fixtureSource{
			id:   candidate.SourceOfficialTradingValue,
			rows: []candidate.Row{discoveryRow("005930", 12, 100, "70000")},
		},
		&fixtureSource{
			id:  candidate.SourceOfficialGainers,
			err: fmt.Errorf("%w: the gainers ranking", candidate.ErrRateLimited),
		},
	)

	out, err := runCandidate(t, "candidate", "scan", "--market", "KR")
	if err != nil {
		t.Fatalf("candidate scan: %v\n%s", err, out)
	}
	if !strings.Contains(out, string(candidate.SourceOfficialGainers)) {
		t.Errorf("the degraded scan does not name the source that went missing:\n%s", out)
	}
	if !strings.Contains(out, "backoff") && !strings.Contains(out, "backing off") {
		t.Errorf("a rate-limited source produced no recorded retreat:\n%s", out)
	}
	if !strings.Contains(out, "30s") {
		t.Errorf("the retreat does not say how long it is:\n%s", out)
	}
}

// TestTheScanSaysSoWhenItStoppedForSpace is task 5.3c reaching the operator.
func TestTheScanSaysSoWhenItStoppedForSpace(t *testing.T) {
	clk := clock.NewFake(time.Date(2026, 7, 28, 9, 0, 0, 0, time.UTC))
	withCandidateFixture(t, clk, candidate.DefaultFreeSpaceFloor-1, &fixtureSource{
		id:   candidate.SourceOfficialTradingValue,
		rows: []candidate.Row{discoveryRow("005930", 12, 100, "70000")},
	})

	out, err := runCandidate(t, "candidate", "scan", "--market", "KR")
	if err != nil {
		t.Fatalf("candidate scan: %v\n%s", err, out)
	}
	if !strings.Contains(out, "not writing") {
		t.Errorf("discovery halted for space and the output does not say so:\n%s", out)
	}
	if !strings.Contains(out, "ledger") {
		t.Errorf("the halt does not explain whose write it is protecting:\n%s", out)
	}
}

// --- task 5.3: the live verification outranks discovery --------------------------

// TestWatchRefusesToStartWhileALiveVerificationHoldsTheRunLock.
//
// spec Requirement 7: 실계좌 검증이 진행 중이면 반복 스캔은 시작하지 않아야 한다 —
// 사람이 지켜보는 쪽이 우선이다. The verification places real orders with somebody
// watching and a step lost to a 429 costs another one; a discovery cycle costs
// nothing and there will be another in fifteen seconds.
func TestWatchRefusesToStartWhileALiveVerificationHoldsTheRunLock(t *testing.T) {
	dir := t.TempDir()
	clk := clock.NewFake(time.Date(2026, 7, 28, 9, 0, 0, 0, time.UTC))
	withCandidateFixture(t, clk, 100<<30, &fixtureSource{
		id:   candidate.SourceOfficialTradingValue,
		rows: []candidate.Row{discoveryRow("005930", 12, 100, "70000")},
	})

	lock := filepath.Join(dir, runlock.FileName)
	if _, err := runlock.Acquire(lock, time.Now()); err != nil {
		t.Fatalf("acquiring the verification run lock: %v", err)
	}

	out, err := runCandidate(t, "--config-dir", dir,
		"candidate", "watch", "--market", "KR", "--cycles", "1")
	if err == nil {
		t.Fatalf("watch started while a live verification held %s:\n%s", lock, out)
	}
	combined := out + err.Error()
	if !strings.Contains(combined, lock) {
		t.Errorf("the refusal does not name the lock file:\n%s", combined)
	}
	if !strings.Contains(combined, "verification") {
		t.Errorf("the refusal does not say what is holding the budget:\n%s", combined)
	}
}

// TestWatchStartsWhenTheRunLockIsStale.
//
// The lock is advisory and freshness is its whole contract: a crashed verification
// must cost one refused start, not a discovery that can never run again until
// somebody deletes a file by hand.
func TestWatchStartsWhenTheRunLockIsStale(t *testing.T) {
	dir := t.TempDir()
	clk := clock.NewFake(time.Date(2026, 7, 28, 9, 0, 0, 0, time.UTC))
	withCandidateFixture(t, clk, 100<<30, &fixtureSource{
		id:   candidate.SourceOfficialTradingValue,
		rows: []candidate.Row{discoveryRow("005930", 12, 100, "70000")},
	})

	lock := filepath.Join(dir, runlock.FileName)
	if _, err := runlock.Acquire(lock, time.Now()); err != nil {
		t.Fatalf("acquiring the verification run lock: %v", err)
	}
	old := time.Now().Add(-2 * runlock.StaleAfter)
	if err := os.Chtimes(lock, old, old); err != nil {
		t.Fatalf("ageing the lock: %v", err)
	}

	if _, err := runCandidate(t, "--config-dir", dir,
		"candidate", "watch", "--market", "KR", "--cycles", "1"); err != nil {
		t.Fatalf("watch refused to start behind a stale lock: %v", err)
	}
}

// --- the §4 review's warning to section 5 -----------------------------------------

// TestNoConsumerReadsAVetoThroughItsDroppableSecondReturn.
//
// The §4 review left this for section 5 by name: `if !chase.NearHigh.Dangerous() {
// score += 2 }` is one line and it compiles, and `!Dangerous()` is the spelling D10
// forbids — an unmeasured veto answers false to Dangerous and to Clear, so the
// negation reads "unmeasured" as "fine". The helper pairs are the other door:
// `near, _ := r.NearHigh(th)` drops the measured flag and turns every candidate
// nobody spent a candle on into a clear one.
//
// internal/candidate cannot prevent either from outside its own package, so this
// walks every consumer instead. It covers the console screen the follow-up change
// will add, deliberately — the failure is easier to write there than anywhere else,
// because a screen is where somebody is trying to make a column look tidy.
//
// # What this catches, exactly
//
// The §5 review measured the first version by mutation: it saw the two direct
// forms and six other spellings walked past it. Four of the six are cheap to see
// in a parse and are caught now; two are not, and they are written down below
// rather than left for the next reviewer to rediscover — isolation_test.go states
// its own gap the same way, and for the same reason. A guard that claims more than
// it checks is worse than one that claims less.
//
//	r.NearHigh(th)                a call to one of the pairs, anywhere
//	!s.Dangerous()                the negation D10 forbids
//	!(s.Dangerous())              …parenthesised
//	s.Dangerous() == false        …spelled as a comparison
//	s.Dangerous() != true         …spelled as the other comparison
//	f := s.Dangerous              …taken as a method value or a method
//	                              expression, to be called out of sight later
//
// # What it does not catch
//
//	bad := s.Dangerous()          the result held in a variable, negated later.
//	if !bad                       Following it needs dataflow and this is a parse.
//
//	f := r.NearHigh               one of the PAIRS taken as a method value.
//	                              Telling it apart from `chase.NearHigh` — which
//	                              is the safe VetoState field and is the ordinary
//	                              spelling — needs type information, and flagging
//	                              both would make the guard fire on the correct
//	                              code. So the pairs are caught at their calls
//	                              only.
//
// Neither gap is closed by this test and neither is closed anywhere else. What
// closes them is that both spellings are longer than the correct one.
func TestNoConsumerReadsAVetoThroughItsDroppableSecondReturn(t *testing.T) {
	root := moduleRootForCandidateTest(t)
	checked := 0
	for _, dir := range []string{"cmd", "internal"} {
		err := filepath.WalkDir(filepath.Join(root, dir), func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				switch d.Name() {
				case "testdata":
					return filepath.SkipDir
				case "candidate":
					// The definitions live here, and so does their single sanctioned
					// reader. band.go's own guards cover this package from inside.
					if filepath.Base(filepath.Dir(path)) == "internal" {
						return filepath.SkipDir
					}
				}
				return nil
			}
			if !strings.HasSuffix(path, ".go") {
				return nil
			}
			fset := token.NewFileSet()
			file, perr := parser.ParseFile(fset, path, nil, 0)
			if perr != nil {
				return nil // a file this build does not compile is not this test's business
			}
			checked++
			for _, finding := range droppableVetoReads(fset, file) {
				t.Errorf("%s:%s", mustRel(root, path), finding)
			}
			return nil
		})
		if err != nil {
			t.Fatalf("walking %s: %v", dir, err)
		}
	}
	if checked < 50 {
		t.Fatalf("only %d files were inspected; the walk is not covering the repository", checked)
	}
}

// vetoPairs are the accessors whose second return is the measurement.
//
// Inside internal/candidate they have exactly one reader each and that is
// deliberate (see vetoFrom); out here there must be none.
var vetoPairs = map[string]bool{"NearHigh": true, "GainExceeds": true, "PercentileExceeds": true}

// droppableVetoReads returns every place in `file` that reads a chase verdict in
// one of the ways D10 forbids, as "line: what it did".
//
// The exact boundary — what it sees and what walks past it — is the doc comment on
// TestNoConsumerReadsAVetoThroughItsDroppableSecondReturn, and
// TestTheVetoConsumerGuardSeesTheSpellingsItClaimsTo is the table that keeps that
// comment honest.
func droppableVetoReads(fset *token.FileSet, file *ast.File) []string {
	// Pass one: every selector that is the thing being called. A `Dangerous`
	// selector that is NOT one of these has been taken as a method value or a
	// method expression and will be called somewhere this parse cannot see.
	called := map[*ast.SelectorExpr]bool{}
	ast.Inspect(file, func(n ast.Node) bool {
		if call, ok := n.(*ast.CallExpr); ok {
			if sel, ok := unparenExpr(call.Fun).(*ast.SelectorExpr); ok {
				called[sel] = true
			}
		}
		return true
	})

	var out []string
	add := func(pos token.Pos, what string) {
		out = append(out, fmt.Sprintf("%d: %s", fset.Position(pos).Line, what))
	}
	const pairWhy = "; its second return is the measurement and dropping it turns every " +
		"unmeasured candidate into a clear one. Read the verdict through Chase/VetoState instead"
	const dangerWhy = "; an unmeasured veto answers false to Dangerous() and to Clear(), so the " +
		"negation reads 'we never looked' as 'this one is fine'. Clear() is the predicate that " +
		"requires the measurement"

	ast.Inspect(file, func(n ast.Node) bool {
		switch e := n.(type) {
		case *ast.SelectorExpr:
			switch {
			case vetoPairs[e.Sel.Name] && called[e]:
				add(e.Pos(), "calls "+e.Sel.Name+pairWhy)
			case e.Sel.Name == "Dangerous" && !called[e]:
				// A method value or a method expression. The pairs deliberately get
				// no equivalent check: `chase.NearHigh` is the safe field read and
				// telling the two apart needs types.
				add(e.Pos(), "takes Dangerous as a method value"+dangerWhy)
			}
		case *ast.UnaryExpr:
			if e.Op == token.NOT && isDangerousCall(e.X) {
				add(e.Pos(), "negates Dangerous()"+dangerWhy)
			}
		case *ast.BinaryExpr:
			// `x.Dangerous() == false` and `x.Dangerous() != true` are the negation
			// with the exclamation mark spelled out.
			for _, pair := range [][2]ast.Expr{{e.X, e.Y}, {e.Y, e.X}} {
				if !isDangerousCall(pair[0]) {
					continue
				}
				lit, ok := unparenExpr(pair[1]).(*ast.Ident)
				if !ok {
					continue
				}
				if (e.Op == token.EQL && lit.Name == "false") ||
					(e.Op == token.NEQ && lit.Name == "true") {
					add(e.Pos(), "compares Dangerous() against "+lit.Name+dangerWhy)
				}
			}
		}
		return true
	})
	return out
}

// isDangerousCall reports a call whose callee is spelled `….Dangerous`.
func isDangerousCall(e ast.Expr) bool {
	call, ok := unparenExpr(e).(*ast.CallExpr)
	if !ok {
		return false
	}
	sel, ok := unparenExpr(call.Fun).(*ast.SelectorExpr)
	return ok && sel.Sel.Name == "Dangerous"
}

// unparenExpr strips redundant parentheses. `!(s.Dangerous())` is the same
// sentence as `!s.Dangerous()` and walked past the first version of this guard.
func unparenExpr(e ast.Expr) ast.Expr {
	for {
		paren, ok := e.(*ast.ParenExpr)
		if !ok {
			return e
		}
		e = paren.X
	}
}

// TestTheVetoConsumerGuardSeesTheSpellingsItClaimsTo is the guard's own guard.
//
// Every static test in this repository can fail in one direction only, and the §5
// review measured this one by mutation: six spellings compiled and walked past it.
// Each row here is one of those spellings or one of the things the guard must not
// fire on, and the two `false` rows at the end are the boundary written as a test
// rather than only as a sentence.
func TestTheVetoConsumerGuardSeesTheSpellingsItClaimsTo(t *testing.T) {
	const prologue = `
type state struct{}
func (state) Dangerous() bool { return false }
func (state) Clear() bool { return false }
type chase struct{ NearHigh state }
type pos struct{}
func (pos) NearHigh(th string) (bool, bool) { return false, false }
`
	for _, tc := range []struct {
		name string
		body string
		want bool
	}{
		{"a direct pair call", `func f(r pos) { _, _ = r.NearHigh("2.0") }`, true},
		{"a pair call with the measurement dropped", `func f(r pos) bool { n, _ := r.NearHigh("2.0"); return n }`, true},
		{"a negated Dangerous", `func f(c chase) bool { return !c.NearHigh.Dangerous() }`, true},
		{"a parenthesised negated Dangerous", `func f(c chase) bool { return !(c.NearHigh.Dangerous()) }`, true},
		{"Dangerous compared to false", `func f(c chase) bool { return c.NearHigh.Dangerous() == false }`, true},
		{"false compared to Dangerous", `func f(c chase) bool { return false == c.NearHigh.Dangerous() }`, true},
		{"Dangerous compared to true with !=", `func f(c chase) bool { return c.NearHigh.Dangerous() != true }`, true},
		{"a method value", `func f(c chase) bool { g := c.NearHigh.Dangerous; return !g() }`, true},
		{"a method expression", `func f(s state) bool { g := state.Dangerous; return !g(s) }`, true},
		{"a negated method expression call", `func f(s state) bool { return !state.Dangerous(s) }`, true},

		// The correct spellings, which the guard must leave alone. A guard that
		// fired on these would be removed rather than obeyed.
		{"a sanctioned call in a switch", `func f(s state) string {
	switch {
	case s.Dangerous():
		return "위험"
	case s.Clear():
		return "측정·안전"
	}
	return "미측정"
}`, false},
		{"the field read the pairs share a name with", `func f(c chase) state { return c.NearHigh }`, false},

		// The boundary. Both of these are real ways to write the defect and both
		// need more than a parse to see; they are here so the claim above cannot
		// quietly grow.
		{"the result held in a variable", `func f(c chase) bool { bad := c.NearHigh.Dangerous(); return !bad }`, false},
		{"a pair taken as a method value", `func f(r pos) func(string) (bool, bool) { return r.NearHigh }`, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			fset := token.NewFileSet()
			file, err := parser.ParseFile(fset, "probe.go", "package main\n"+prologue+tc.body, 0)
			if err != nil {
				t.Fatalf("parsing the probe: %v", err)
			}
			got := droppableVetoReads(fset, file)
			if (len(got) > 0) != tc.want {
				t.Errorf("guard found %v, want a finding = %v", got, tc.want)
			}
		})
	}
}

func moduleRootForCandidateTest(t *testing.T) string {
	t.Helper()
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatalf("resolving the repository root: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "go.mod")); err != nil {
		t.Fatalf("%s does not look like the repository root: %v", root, err)
	}
	return root
}

func mustRel(root, path string) string {
	rel, err := filepath.Rel(root, path)
	if err != nil {
		return path
	}
	return rel
}
