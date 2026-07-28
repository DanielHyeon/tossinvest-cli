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
// 150 rows" and "KR 150행" and "150개 행" and not "roughly 150 symbols", which is the
// candle coverage arithmetic and is about something else entirely.
var rowClaim = regexp.MustCompile(`(\d+)[ \t]?(?:개)?[ \t]?(?:-row\b|[ \t]rows?\b|행)`)

// denials are the phrases that turn a mention of a size into a denial of it.
//
// # Corrected 2026-07-28: what "no panel" was doing here
//
// The list used to contain the bare tokens "No panel" and "no panel", and they
// exempted any line containing them — including
//
//	// The KR panel returns 150 rows and the US panel 100 (no panel note)
//
// which is the plain claim with four characters appended. They were there to cover
// the real sentences in metrics.go and veto.go, and those say "no panel has ever
// returned 150 rows", which "has ever returned" already covers. So they are gone: a
// denial has to be a phrase that cannot be part of an assertion.
var denials = []string{
	"never existed", "존재한 적", "has ever returned", "has ever had",
}

// datedCorrection is the other kind of exemption: a note recording what a comment
// used to claim.
//
// The date is required in full rather than by year, for the reason above — "corrected
// 2026" is a token somebody can append. A dated correction is a sentence with a
// specific claim in it, and the precedent is add-candidate-discovery's own in-place
// near_high sign correction.
var datedCorrection = regexp.MustCompile(`(?i)(?:corrected|정정)[ \t]+20\d\d-\d\d-\d\d`)

// sentenceBreak is where one claim ends and the next begins.
var sentenceBreak = regexp.MustCompile(`\. |\.$|。`)

// sentenceAround returns the fragment of `text` that contains byte offset `at`.
//
// This is the adjacency rule, and it is a rule rather than a distance: the denial
// has to be in the same sentence as the number, so a paragraph cannot acquire a
// licence from a clause elsewhere on the line. There is no character window here
// because a character window would be a number nobody derived — the same objection
// this whole change is about.
func sentenceAround(text string, at int) string {
	start, end := 0, len(text)
	for _, loc := range sentenceBreak.FindAllStringIndex(text, -1) {
		switch {
		case loc[1] <= at:
			start = loc[1]
		case loc[0] > at:
			return text[start:loc[0]]
		}
	}
	return text[start:end]
}

// deniesTheSize reports that the claim at offset `at` is being denied rather than
// made.
func deniesTheSize(text string, at int) bool {
	sentence := sentenceAround(text, at)
	for _, d := range denials {
		if strings.Contains(sentence, d) {
			return true
		}
	}
	return datedCorrection.MatchString(sentence)
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
					for _, loc := range rowClaim.FindAllStringSubmatchIndex(text, -1) {
						claim := text[loc[0]:loc[1]]
						n, convErr := strconv.Atoi(text[loc[2]:loc[3]])
						if convErr != nil || sizes[n] {
							continue
						}
						// The exemption is narrow and it is not "this line is old". A
						// line may name a size no source declares only when that
						// sentence is *denying* it — the warning that says a request of
						// that size must not be made, or a dated correction recording
						// what the comment used to claim. Both of those are the record
						// of the fiction rather than the fiction, and deleting them
						// would remove the only evidence that it was ever there.
						if deniesTheSize(text, loc[0]) {
							continue
						}
						t.Errorf("%s:%d claims %q, and no source declares %d rows "+
							"(declared: %v).\n\n%s\n\n"+
							"The 150-row panel was in five comments and a design "+
							"document and it never existed anywhere else. A comment "+
							"that states a list length is a claim about the sources, "+
							"and it goes stale silently — this is the test that makes "+
							"it go stale loudly.",
							path, fset.Position(line.Pos()).Line, claim, n, sizes,
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
		// The four appended words that used to buy an exemption. "no panel" was in
		// the denial list as a bare token, so a line could assert the fiction and
		// then mention that a panel note exists.
		"// The KR panel returns 150 rows and the US panel 100 (no panel note)",
		// A year with no date behind it, for the same reason.
		"// The KR panel returns 150 rows (corrected 2026, probably)",
		// The Korean counter word, which the matcher used to require to be absent.
		"// KR 패널은 150개 행을 돌려준다",
		// A denial that is real but belongs to a different sentence. The line asserts
		// the fiction and then denies something else, which is how a paragraph gets a
		// licence from a clause elsewhere on it.
		"// The KR panel returns 150 rows. A 200-row panel has never existed.",
	} {
		found := false
		for _, loc := range rowClaim.FindAllStringSubmatchIndex(fiction, -1) {
			n, err := strconv.Atoi(fiction[loc[2]:loc[3]])
			if err != nil || sizes[n] {
				continue
			}
			if !deniesTheSize(fiction, loc[0]) {
				found = true
			}
		}
		if !found {
			t.Errorf("the guard accepts %q, which is the plain claim rather than a denial of "+
				"it. Either the matcher stopped matching or the exemption grew until it "+
				"covered the sentence the guard exists for — and both of those report "+
				"success on a file full of fiction", fiction)
		}
	}

	// And it does not fire on the sentences that are about something else, or on the
	// real corrections in this package.
	for _, fine := range []string{
		"// what is left is two ticks and roughly 150 symbols — four times smaller",
		"// five sources at 30-100 rows each puts the raw table in the millions",
		"// 100\". No panel has ever returned 150 rows, in either market. The normalisation is",
		"// Corrected 2026-07-28: this said \"the KR panel returns 150 rows and the US panel",
		"// system reads has ever had 150 rows. The official ranking serves at most 100 and",
	} {
		for _, loc := range rowClaim.FindAllStringSubmatchIndex(fine, -1) {
			n, err := strconv.Atoi(fine[loc[2]:loc[3]])
			if err != nil || sizes[n] || deniesTheSize(fine, loc[0]) {
				continue
			}
			t.Errorf("the guard rejects %q on the strength of %q; a guard that fires on "+
				"unrelated numbers, or on the record of the fiction it was written about, "+
				"is one somebody turns off", fine, fine[loc[0]:loc[1]])
		}
	}
}

// TestWhatThisGuardCannotCatch is the honest half, and it is a test rather than a
// comment so that the list cannot quietly stop being true.
//
// A text guard over comments has limits, and stating them is better than implying
// there are none: somebody who believes this guard is complete will not look for the
// next 150-row sentence, and the sentence that started all of this survived for
// months precisely because everybody assumed somebody else had checked.
//
// Each case below is a claim the matcher does not see. They are asserted as NOT
// caught, so the day one of them starts being caught this test says so and the list
// gets shorter.
func TestWhatThisGuardCannotCatch(t *testing.T) {
	sizes := declaredPanelSizes(t)
	catches := func(text string) bool {
		for _, loc := range rowClaim.FindAllStringSubmatchIndex(text, -1) {
			n, err := strconv.Atoi(text[loc[2]:loc[3]])
			if err == nil && !sizes[n] && !deniesTheSize(text, loc[0]) {
				return true
			}
		}
		return false
	}
	for _, tc := range []struct {
		why  string
		text string
	}{
		{
			why:  "a number spelled out in words has no digits to match",
			text: "// The KR panel returns one hundred and fifty rows",
		},
		{
			why:  "the matcher reads one comment line at a time, so a claim split across two is two halves of nothing",
			text: "// The KR panel returns 150",
		},
		{
			why: "the unit word is required, and `symbol` cannot be one of them: " +
				"veto.go's candle arithmetic really does say `roughly 150 symbols` about " +
				"something else, so admitting it would make the guard fire on a true sentence",
			text: "// a 150-symbol list and a 100-symbol one",
		},
		{
			why: "a denial phrase used dishonestly in the same sentence still exempts — " +
				"adjacency bounds where the licence comes from, it cannot judge whether it is deserved",
			text: "// The KR panel returns 150 rows, whatever has ever returned otherwise",
		},
	} {
		if catches(tc.text) {
			t.Errorf("the guard now catches %q (%s).\n\nThat is an improvement, and this "+
				"list is the record of what it could not do — remove this case so the "+
				"record stays accurate", tc.text, tc.why)
		}
	}

	// And the last limit is where it reads rather than what it matches: the walk
	// iterates file.Comments, so the same sentence in a string literal, an
	// identifier or a Markdown file is not seen at all. `add-candidate-discovery`'s
	// design document carried the fiction for months and this test would not have
	// found it there.
	const inALiteral = `package p
const note = "the KR panel returns 150 rows"`
	parsed, err := parser.ParseFile(token.NewFileSet(), "fixture.go", inALiteral,
		parser.ParseComments|parser.SkipObjectResolution)
	if err != nil {
		t.Fatalf("parsing the fixture: %v", err)
	}
	for _, group := range parsed.Comments {
		for _, line := range group.List {
			if catches(line.Text) {
				t.Errorf("the walk now reaches %q; string literals and documents are outside "+
					"what this guard reads, and that is the last thing on its list of "+
					"limits", line.Text)
			}
		}
	}
}
