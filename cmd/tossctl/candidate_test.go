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
func withCandidateFixture(t *testing.T, clk clock.Clock, free int64, sources ...candidate.Source) {
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
	candidateStoreFactory = func(*rootOptions) (*candidate.Store, error) { return store, nil }
	candidatePanelFactory = func(*rootOptions, string) ([]candidate.Source, error) {
		return sources, nil
	}
}

// runCandidate executes a discovery subcommand and returns its stdout.
func runCandidate(t *testing.T, args ...string) (string, error) {
	t.Helper()
	root := newRootCmd()
	var out, errOut bytes.Buffer
	root.SetOut(&out)
	root.SetErr(&errOut)
	root.SetArgs(args)
	err := root.ExecuteContext(context.Background())
	return out.String(), err
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
		ShadowBands map[string]struct {
			Total   int            `json:"total"`
			Crossed map[string]int `json:"crossed"`
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
func TestNoConsumerReadsAVetoThroughItsDroppableSecondReturn(t *testing.T) {
	// The pairs whose second return is the measurement. Inside internal/candidate
	// they have exactly one reader each and that is deliberate (see vetoFrom); out
	// here there must be none.
	pairs := map[string]bool{"NearHigh": true, "GainExceeds": true, "PercentileExceeds": true}

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
			ast.Inspect(file, func(n ast.Node) bool {
				switch e := n.(type) {
				case *ast.CallExpr:
					sel, ok := e.Fun.(*ast.SelectorExpr)
					if ok && pairs[sel.Sel.Name] {
						t.Errorf("%s:%d calls %s; its second return is the measurement and "+
							"dropping it turns every unmeasured candidate into a clear one. "+
							"Read the verdict through Chase/VetoState instead",
							mustRel(root, path), fset.Position(e.Pos()).Line, sel.Sel.Name)
					}
				case *ast.UnaryExpr:
					if e.Op != token.NOT {
						return true
					}
					call, ok := e.X.(*ast.CallExpr)
					if !ok {
						return true
					}
					sel, ok := call.Fun.(*ast.SelectorExpr)
					if ok && sel.Sel.Name == "Dangerous" {
						t.Errorf("%s:%d negates Dangerous(); an unmeasured veto answers false "+
							"to it and to Clear(), so !Dangerous() reads 'we never looked' as "+
							"'this one is fine'. Clear() is the predicate that requires the "+
							"measurement", mustRel(root, path), fset.Position(e.Pos()).Line)
					}
				}
				return true
			})
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
