package candidate

// panelsize_drift_test.go is task 7.3: a comment that claims a panel size has to
// agree with the source that declares one.
//
// # Why a test and not a careful edit
//
// "The KR panel returns 150 rows and the US panel 100" was in this package's
// comments in five places, in the design document that produced them, and in the
// derivation of a threshold somebody nearly shipped. No panel has ever returned 150
// rows. It survived because a comment cannot be wrong in a way that fails, and it
// spread because it was the sentence that explained *why* the percentile is
// normalised — so everyone who needed that explanation quoted it.
//
// Correcting the five occurrences fixes today. This is what makes the next one fail:
// the sizes are read out of internal/candidatesrc's source, and a comment in either
// package that names a row count the sources do not declare is an error.
//
// # Reading candidatesrc as text is not importing it
//
// internal/candidate must not import internal/candidatesrc — the dependency runs one
// way and isolation_test.go's forbidden table says so. Parsing a file is not
// importing it, which is exactly the arrangement fsguard_drift_test.go already uses
// to keep this package's filesystem allowlist in step with the ledger's.

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

// declaredPanelSizes reads every row count internal/candidatesrc asks a source for.
//
// It takes them from the calls rather than from a constant, because the calls are
// what the market sees: Panel's `OfficialRanking(..., 100)` and `WTSPopular(wts, 30)`
// are the requests, and the cap inside OfficialRanking is what makes the first of
// them the truth rather than an aspiration.
func declaredPanelSizes(t *testing.T) map[int]bool {
	t.Helper()
	const path = "../candidatesrc/candidatesrc.go"
	src, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		t.Fatalf("reading %s: %v — it is the file that declares how many rows a source "+
			"asks for, and this guard is nothing without it", path, err)
	}
	file, err := parser.ParseFile(token.NewFileSet(), path, src, parser.SkipObjectResolution)
	if err != nil {
		t.Fatalf("parsing %s: %v", path, err)
	}

	out := map[int]bool{}
	ast.Inspect(file, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		id, ok := call.Fun.(*ast.Ident)
		if !ok || (id.Name != "OfficialRanking" && id.Name != "WTSPopular") {
			return true
		}
		for _, arg := range call.Args {
			lit, ok := arg.(*ast.BasicLit)
			if !ok || lit.Kind != token.INT {
				continue
			}
			if n, err := strconv.Atoi(lit.Value); err == nil && n > 1 {
				out[n] = true
			}
		}
		return true
	})
	if len(out) == 0 {
		t.Fatalf("no row count was read from %s; the guard is matching nothing and every "+
			"assertion below would pass on any comment at all", path)
	}
	return out
}

// TestTheSourcesStillDeclareTheSizesTheseCommentsClaim.
func TestTheSourcesStillDeclareTheSizesTheseCommentsClaim(t *testing.T) {
	sizes := declaredPanelSizes(t)
	for _, want := range []int{100, 30} {
		if !sizes[want] {
			t.Errorf("internal/candidatesrc no longer asks any source for %d rows; the "+
				"comments in this package that name it are now the same kind of fiction "+
				"the 150-row panel was. Declared sizes: %v", want, sizes)
		}
	}
	if sizes[150] {
		t.Error("a source now asks for 150 rows. The official ranking caps count at 100 " +
			"server-side, so a caller asking for 150 computes percentiles against a list " +
			"length that never existed — which is the defect candidatesrc.go's own comment " +
			"warns about, and the origin of the 150-row fiction")
	}
}

// rowClaim matches a comment that states a number of rows.
//
// Both spellings, because these packages comment in English and Korean and the
// fiction appeared in both. The unit words are required: this must match "returns
// 150 rows" and "KR 150행" and not "roughly 150 symbols", which is the candle
// coverage arithmetic and is about something else entirely.
var rowClaim = regexp.MustCompile(`(\d+)[ \t]?(?:-row\b|[ \t]rows?\b|행)`)

// denials are the phrases that turn a mention of a size into a denial of it.
//
// They are matched on the line rather than on the comment group, which is
// deliberately strict: a correction note has to say so on the line that repeats the
// number, so a paragraph cannot acquire a licence from a sentence three lines up.
var denials = []string{
	"never existed", "존재한 적", "has ever returned", "has ever had",
	"No panel", "no panel", "Corrected 2026", "corrected 2026", "정정 2026",
}

func deniesTheSize(line string) bool {
	for _, d := range denials {
		if strings.Contains(line, d) {
			return true
		}
	}
	return false
}

// TestNoCommentClaimsAPanelSizeTheSourcesDoNotDeclare is the drift guard.
func TestNoCommentClaimsAPanelSizeTheSourcesDoNotDeclare(t *testing.T) {
	sizes := declaredPanelSizes(t)
	// A row count may also be a quantity that is not a panel: a range endpoint in a
	// volume estimate, or a quotation of the warning that named 150 as the number
	// nobody may ask for. The first is allowed by listing the sizes; the second is
	// allowed only on a line that also says it is a warning.
	fset := token.NewFileSet()
	scanned := 0
	for _, dir := range []string{".", "../candidatesrc"} {
		entries, err := os.ReadDir(dir)
		if err != nil {
			t.Fatalf("reading %s: %v", dir, err)
		}
		for _, e := range entries {
			name := e.Name()
			if e.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
				continue
			}
			path := filepath.Join(dir, name)
			file, err := parser.ParseFile(fset, path, nil, parser.ParseComments|parser.SkipObjectResolution)
			if err != nil {
				t.Fatalf("parsing %s: %v", path, err)
			}
			scanned++
			for _, group := range file.Comments {
				for _, line := range group.List {
					text := line.Text
					for _, m := range rowClaim.FindAllStringSubmatch(text, -1) {
						n, convErr := strconv.Atoi(m[1])
						if convErr != nil || sizes[n] {
							continue
						}
						// The exemption is narrow and it is not "this line is old". A
						// line may name a size no source declares only when the line is
						// *denying* it — the warning that says a request of that size
						// must not be made, or a dated correction recording what the
						// comment used to claim. Both of those are the record of the
						// fiction rather than the fiction, and deleting them would
						// remove the only evidence that it was ever there.
						if deniesTheSize(text) {
							continue
						}
						t.Errorf("%s:%d claims %q, and no source declares %d rows "+
							"(declared: %v).\n\n%s\n\n"+
							"The 150-row panel was in five comments and a design "+
							"document and it never existed anywhere else. A comment "+
							"that states a list length is a claim about the sources, "+
							"and it goes stale silently — this is the test that makes "+
							"it go stale loudly.",
							path, fset.Position(line.Pos()).Line, m[0], n, sizes,
							strings.TrimSpace(text))
					}
				}
			}
		}
	}
	if scanned < 5 {
		t.Fatalf("scanned %d production sources; the guard is reading the wrong tree", scanned)
	}
}

// TestThePanelSizeGuardCatchesTheClaimItWasWrittenFor is the positive control.
//
// The exact sentence that was in metrics.go, veto.go and the design document, run
// through the same matcher, and required to be rejected. Without it a matcher that
// had stopped matching would report success for a file full of fiction.
func TestThePanelSizeGuardCatchesTheClaimItWasWrittenFor(t *testing.T) {
	sizes := declaredPanelSizes(t)
	for _, fiction := range []string{
		"// The KR panel returns 150 rows and the US panel 100 (D8)",
		"// 원천마다 리스트 길이가 다르므로(KR 150행, US 100행)",
		"// a 150-row list and a 100-row one",
	} {
		found := false
		for _, m := range rowClaim.FindAllStringSubmatch(fiction, -1) {
			if n, err := strconv.Atoi(m[1]); err == nil && !sizes[n] {
				found = true
			}
		}
		if !found {
			t.Errorf("the matcher does not reject %q, which is the sentence this guard "+
				"exists for. Every assertion in this file is passing because it is "+
				"matching nothing", fiction)
		}
		// And the exemption does not swallow it. A denial list that grew until it
		// covered the plain claim would be the guard turning itself off.
		if deniesTheSize(fiction) {
			t.Errorf("the exemption covers %q, which is the plain claim rather than a "+
				"denial of it", fiction)
		}
	}
	// And it does not fire on the sentences that are about something else.
	for _, fine := range []string{
		"// what is left is two ticks and roughly 150 symbols — four times smaller",
		"// five sources at 30-100 rows each puts the raw table in the millions",
	} {
		for _, m := range rowClaim.FindAllStringSubmatch(fine, -1) {
			if n, err := strconv.Atoi(m[1]); err == nil && !sizes[n] {
				t.Errorf("the matcher rejects %q on the strength of %q; a guard that fires "+
					"on unrelated numbers is one somebody turns off", fine, m[0])
			}
		}
	}
}
