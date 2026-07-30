package candidatesrc

// snapshot_drift_test.go is D8: the official snapshot already knew.
//
// # Why this file exists rather than a corrected parameter
//
// `docs/migration/openapi.latest.json` is the repository's copy of the official
// API description, and it has said since 2026-06-12 that TOP_GAINERS and
// TOP_LOSERS reject `duration=realtime` with a 400. The wiring that sends
// `realtime` to all three ranking types arrived forty-six days later. Nothing
// failed: the ranking that could never answer was reported as a degraded source
// on every single scan, which is the same as not being reported at all.
//
// So the defect this guard is written about is not five characters in a request.
// It is that `docs/WORKFLOW.md`'s "broker behaviour = official API fixture"
// authority was sitting in the tree and nobody read it. Fixing the wiring fixes
// today; this is what makes the next one fail.
//
// # It compares two sets rather than searching for a string
//
//	forbidden   read out of the snapshot: which ranking types the API says do not
//	            accept which duration.
//	wired       read out of candidatesrc.go: which ranking types Panel builds, and
//	            which duration the reader is handed.
//
// The assertion is that they do not intersect. Two consequences follow from the
// shape, and both are the reason for it:
//
//   - TOP_LOSERS is covered without being named. "let us add the losers list" is a
//     proposal that will be made, and it meets the same 400.
//   - A future change that moves the duration to one the API does accept for these
//     types passes without being edited, because the forbidden set is computed per
//     duration rather than per type. This guard forbids a *shape* — sending a
//     request the snapshot says will be refused — which is what the spec delta
//     forbids at requirement level (D3).
//
// # Reading two files as text is not importing them
//
// The precedent is next door: internal/candidate's panelsize_drift_test.go parses
// ../candidatesrc/candidatesrc.go, and fsguard_drift_test.go follows the ledger's
// allowlist the same way.
//
// This file lives in internal/candidatesrc, next to the wiring it reads, rather
// than in internal/candidate beside that precedent. The wiring under guard is this
// package's own: a change to Panel and a failure of this test then happen in one
// `go test ./internal/candidatesrc/...`, and the guard needs one relative path out
// of the package instead of two.
//
// # The second guard in this file, and why it is in this file
//
// Reading the wiring out of the source text buys the comparison above, and it costs
// an assumption: that the literal the reader finds is the wiring that runs. Nothing
// checked that assumption, and it fails in two directions.
//
// A ranking type with no rankingSourceID entry makes OfficialRanking return an error
// which Panel discards, so the type is named in the literal and no source is built
// from it. That one passes silently: every other guard in this package reads the
// panel it is handed and sees one element fewer, still unique, still not empty.
//
// A refactor that lifts the literal out of Panel leaves the reader looking at
// whatever other []string is in scope — and that one passes silently too. If it
// leaves the reader looking at *nothing*, it does not: wiredRankings' len(out) == 0
// Fatal already refuses that, which is why the second direction is a hole rather
// than an opening.
//
// That Fatal is therefore load-bearing and it must not be deleted as redundant with
// the test below. With it gone, an empty `declared` makes this file's two guards go
// blind together — this one passes vacuously over zero types, and the snapshot guard
// computes an empty cross product and reports no forbidden combination for any
// wiring at all. It is the only path on which both are silent at once.
//
// TestEveryRankingTypeThePanelNamesBecomesASourceItBuilds closes both by comparing
// the AST's answer against Panel's own: the types the literal names, against the
// sources Panel returns when it is called. It sits here rather than in a file of its
// own because what it pins is wiredRankings — the function directly above it — and a
// reader who has just understood how the wiring is read should not have to find out
// somewhere else what keeps that reading honest.

import (
	"encoding/json"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/candidate"
)

const (
	// snapshotPath is the official API description as this repository keeps it.
	snapshotPath = "../../docs/migration/openapi.latest.json"
	// wiringPath is the file whose Panel and Read decide what gets requested.
	wiringPath = "candidatesrc.go"
	// rankingsEndpoint is the operation both sets are about.
	rankingsEndpoint = "/api/v1/rankings"
)

// unsupportedDuration matches one enum bullet of the `type` parameter that
// records a duration the API refuses for that ranking type.
//
// # Which of the snapshot's three statements this reads, and why
//
// The same fact is written three times in the snapshot and only one of them has
// the shape this guard needs.
//
//	the endpoint description   "`TOP_GAINERS` / `TOP_LOSERS` 는 `duration=realtime`
//	                           을 지원하지 않습니다" — prose, two types joined by a
//	                           slash inside one sentence, sitting in a bullet list
//	                           about five unrelated things. Enumerating types out
//	                           of it means parsing the join, and a third type added
//	                           to that sentence in any other form is missed.
//	the error example          names the combination only in a human `summary`
//	                           string. Its machine-readable half is the list of
//	                           durations that ARE allowed, and it does not say for
//	                           which type — which is the half this guard needs.
//	the `type` parameter       one bullet per enum value, adjacent to the enum that
//	                           defines the legal values. THIS ONE.
//
// The last is the only per-type statement, which is the shape of the set being
// built, and it is the one place a newly documented ranking type has to appear in
// order to be documented at all — so a future type that refuses a duration arrives
// carrying its own bullet.
//
// The em dash between the label and the constraint is not required by the pattern.
// Requiring punctuation would make a guard that stops matching when somebody
// reflows a sentence, and a guard that stops matching passes.
var unsupportedDuration = regexp.MustCompile(
	"(?m)^-[ \t]*`([A-Z0-9_]+)`[ \t]*:[^\n]*`([a-z0-9]+)`[ \t]*미지원")

// forbiddenCombinations returns, per duration, the ranking types the snapshot says
// will not accept it.
func forbiddenCombinations(t *testing.T) map[string]map[string]bool {
	t.Helper()
	return parseForbidden(t, rankingsTypeDescription(t))
}

// parseForbidden is the matcher on its own, so the positive control below can run
// the real one over a fixture.
func parseForbidden(t *testing.T, description string) map[string]map[string]bool {
	t.Helper()
	out := map[string]map[string]bool{}
	for _, m := range unsupportedDuration.FindAllStringSubmatch(description, -1) {
		typ, duration := m[1], m[2]
		if out[duration] == nil {
			out[duration] = map[string]bool{}
		}
		out[duration][typ] = true
	}
	return out
}

// rankingsTypeDescription pulls the `type` query parameter's description out of the
// snapshot.
func rankingsTypeDescription(t *testing.T) string {
	t.Helper()
	raw, err := os.ReadFile(filepath.Clean(snapshotPath))
	if err != nil {
		t.Fatalf("reading %s: %v — it is the official description of what the ranking "+
			"endpoint accepts, and this guard is nothing without it", snapshotPath, err)
	}
	var doc struct {
		Paths map[string]map[string]struct {
			Parameters []struct {
				Name        string `json:"name"`
				In          string `json:"in"`
				Description string `json:"description"`
			} `json:"parameters"`
		} `json:"paths"`
	}
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatalf("parsing %s: %v", snapshotPath, err)
	}
	for _, param := range doc.Paths[rankingsEndpoint]["get"].Parameters {
		if param.Name == "type" && param.In == "query" {
			return param.Description
		}
	}
	t.Fatalf("%s has no `type` query parameter on GET %s; the constraint this guard "+
		"reads lives in that parameter's description, and a guard that reads an empty "+
		"string passes against every wiring there is", snapshotPath, rankingsEndpoint)
	return ""
}

// wiredRankings is what Panel actually builds, by value rather than by identifier.
//
// The constant values are the API strings — `RankingTopGainers = "TOP_GAINERS"` —
// so no mapping table stands between the two sets, and a table is one more thing
// that can drift.
func wiredRankings(t *testing.T, file *ast.File) map[string]bool {
	t.Helper()
	consts := stringConstants(file)
	out := map[string]bool{}
	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Name.Name != "Panel" || fn.Recv != nil {
			continue
		}
		ast.Inspect(fn.Body, func(n ast.Node) bool {
			lit, ok := n.(*ast.CompositeLit)
			if !ok {
				return true
			}
			arr, ok := lit.Type.(*ast.ArrayType)
			if !ok {
				return true
			}
			if id, ok := arr.Elt.(*ast.Ident); !ok || id.Name != "string" {
				return true
			}
			for _, el := range lit.Elts {
				id, ok := el.(*ast.Ident)
				if !ok {
					t.Fatalf("%s: Panel builds a ranking from %T rather than a named "+
						"constant. This guard resolves identifiers to their values; an "+
						"element it cannot resolve is a request it cannot check",
						wiringPath, el)
				}
				value, known := consts[id.Name]
				if !known {
					t.Fatalf("%s: Panel names %s and this file declares no string constant "+
						"by that name", wiringPath, id.Name)
				}
				out[value] = true
			}
			return true
		})
	}
	if len(out) == 0 {
		t.Fatalf("no ranking type was read out of %s's Panel; the guard is matching "+
			"nothing and would pass against a panel wired to every forbidden "+
			"combination there is", wiringPath)
	}
	return out
}

// panelRankingTypes is what Panel builds, read by running it: every source it
// returns whose id belongs to a ranking type, named back as that type.
//
// The union over both markets. Panel's membership rule is per market (its doc), so a
// ranking that only one market builds is still a ranking the literal names, and
// asking one market would report the other's as missing.
//
// The mapping is rankingSourceID read backwards. A collision in it is fatal rather
// than resolved: two types under one id is the section-2 review's P0 — the id is
// what the scan keys "did every source that raised this candidate answer" on — and
// an inverse that quietly kept one of the two would make this guard report the
// other as a type Panel never built.
func panelRankingTypes(t *testing.T) map[string]bool {
	t.Helper()
	typeOfID := map[candidate.SourceID]string{}
	for typ, id := range rankingSourceID {
		if other, clash := typeOfID[id]; clash {
			t.Fatalf("rankingSourceID gives %s and %s the same source id %q. The scan may "+
				"only cool a candidate when every source that raised it answered and that "+
				"check is keyed by id, so one of the two would vouch for the other",
				other, typ, id)
		}
		typeOfID[id] = typ
	}

	out := map[string]bool{}
	for _, market := range []string{candidate.MarketKR, candidate.MarketUS} {
		for _, src := range Panel(market, &fakeRankings{}, nil, &fakePopular{}, nil) {
			if typ, ok := typeOfID[src.ID()]; ok {
				out[typ] = true
			}
		}
	}
	return out
}

// TestEveryRankingTypeThePanelNamesBecomesASourceItBuilds is what makes the AST
// reading above answerable to the code it claims to read.
//
// Panel discards the error from OfficialRanking. The comment at that line used to
// say TestEveryPanelSourceHasItsOwnID catches a type that slips through, and it does
// not: that test walks the panel it is handed and checks it for duplicate ids and
// emptiness, and a source dropped by the discarded error produces a panel that is
// one element shorter, still unique, still not empty. Every other guard in this
// package reads the same panel and sees the same nothing. So a ranking type added to
// the literal without a rankingSourceID entry was, before this test, a change where
// somebody believes the panel reads three lists and it reads two, and the whole
// suite is green — the failure shape this change exists to remove, inside the
// function this change edits, one line under the sentence that claimed to prevent it.
//
// Two sets, and the assertion is equality:
//
//	declared  the ranking types named in Panel's literal, read from the source text
//	          by wiredRankings.
//	built     the ranking types Panel actually produced sources for, read by calling
//	          it.
//
// A type in `declared` and not in `built` is a source that was silently dropped. A
// type in `built` and not in `declared` means wiredRankings is no longer looking at
// the wiring — the literal moved somewhere the AST reader does not follow, and
// TestThePanelAsksForNothingTheSnapshotSaysWillBeRefused went blind while staying
// green.
//
// Neither direction can pass vacuously. An empty `declared` is already fatal inside
// wiredRankings, and an empty `built` fails the first loop against every type the
// literal names.
func TestEveryRankingTypeThePanelNamesBecomesASourceItBuilds(t *testing.T) {
	declared := wiredRankings(t, parseWiring(t))
	built := panelRankingTypes(t)

	for _, typ := range sorted(declared) {
		if built[typ] {
			continue
		}
		t.Errorf("%s names %s in Panel's ranking literal and Panel builds no source for "+
			"it.\n\n"+
			"OfficialRanking refuses a type that rankingSourceID does not map, and Panel "+
			"discards that error, so the type is requested by nobody and the panel comes "+
			"back one element shorter. Nothing else in this package reports it: the id "+
			"checks see a panel that is still unique and still not empty, and this file's "+
			"snapshot guard sees a type the API never refused. Either give %s an entry in "+
			"rankingSourceID or take it out of the literal.\n\n"+
			"Named in the literal: %v\nBuilt by Panel: %v",
			wiringPath, typ, typ, sorted(declared), sorted(built))
	}

	for _, typ := range sorted(built) {
		if declared[typ] {
			continue
		}
		t.Errorf("Panel builds a source for %s and no ranking literal in Panel names "+
			"it.\n\n"+
			"wiredRankings reads Panel's []string literals out of %s, and this says it is "+
			"reading something other than the list that runs — the types moved to a "+
			"package-level var, or behind a helper, or into a literal it resolves "+
			"differently. Every assertion in this file is computed from that reading, so "+
			"TestThePanelAsksForNothingTheSnapshotSaysWillBeRefused would go on passing "+
			"against wiring it can no longer see.\n\n"+
			"Named in the literal: %v\nBuilt by Panel: %v",
			typ, wiringPath, sorted(declared), sorted(built))
	}
}

// wiredDurations is every duration literal handed to RankingReader.Rankings.
//
// Position three, from the interface's own signature: Rankings(ctx, typ,
// marketCountry, duration, excludeCaution, count). Taking "every string literal in
// the call" would collect the market as well, and this guard's whole content is
// which duration goes with which type.
//
// A duration that is not a literal is fatal rather than skipped. A guard that
// cannot see what is being sent has nothing to say about it, and saying nothing
// here reads as approval.
func wiredDurations(t *testing.T, file *ast.File) map[string]bool {
	t.Helper()
	out := map[string]bool{}
	ast.Inspect(file, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok || sel.Sel.Name != "Rankings" {
			return true
		}
		const durationArg = 3
		if len(call.Args) <= durationArg {
			t.Fatalf("%s: a Rankings call takes %d arguments; the duration is the fourth "+
				"and this guard reads it by position", wiringPath, len(call.Args))
		}
		lit, ok := call.Args[durationArg].(*ast.BasicLit)
		if !ok || lit.Kind != token.STRING {
			t.Fatalf("%s: the duration handed to Rankings is %T rather than a string "+
				"literal. That may well be an improvement, but this guard can no longer "+
				"see which duration is sent — and a guard that sees nothing reports no "+
				"forbidden combination for any wiring at all", wiringPath, call.Args[durationArg])
		}
		value, err := strconv.Unquote(lit.Value)
		if err != nil {
			t.Fatalf("%s: unquoting the duration %s: %v", wiringPath, lit.Value, err)
		}
		out[value] = true
		return true
	})
	if len(out) == 0 {
		t.Fatalf("no duration was read out of %s; nothing calls Rankings with a literal "+
			"and this guard has no combination to check", wiringPath)
	}
	return out
}

// stringConstants is every `Name = "value"` in the file.
func stringConstants(file *ast.File) map[string]string {
	out := map[string]string{}
	for _, decl := range file.Decls {
		gen, ok := decl.(*ast.GenDecl)
		if !ok || gen.Tok != token.CONST {
			continue
		}
		for _, spec := range gen.Specs {
			value, ok := spec.(*ast.ValueSpec)
			if !ok {
				continue
			}
			for i, name := range value.Names {
				if i >= len(value.Values) {
					continue
				}
				lit, ok := value.Values[i].(*ast.BasicLit)
				if !ok || lit.Kind != token.STRING {
					continue
				}
				if s, err := strconv.Unquote(lit.Value); err == nil {
					out[name.Name] = s
				}
			}
		}
	}
	return out
}

// parseWiring reads candidatesrc.go.
func parseWiring(t *testing.T) *ast.File {
	t.Helper()
	src, err := os.ReadFile(filepath.Clean(wiringPath))
	if err != nil {
		t.Fatalf("reading %s: %v", wiringPath, err)
	}
	file, err := parser.ParseFile(token.NewFileSet(), wiringPath, src, parser.SkipObjectResolution)
	if err != nil {
		t.Fatalf("parsing %s: %v", wiringPath, err)
	}
	return file
}

// TestThePanelAsksForNothingTheSnapshotSaysWillBeRefused is the guard.
func TestThePanelAsksForNothingTheSnapshotSaysWillBeRefused(t *testing.T) {
	forbidden := forbiddenCombinations(t)
	if len(forbidden) == 0 {
		t.Fatalf("%s yielded no forbidden type/duration combination at all. The "+
			"constraint that produced this guard — TOP_GAINERS and TOP_LOSERS refusing "+
			"`realtime` — is written in that file, so an empty set means the matcher "+
			"stopped matching rather than that the API stopped refusing. Every "+
			"assertion below would then pass against every wiring there is, which is "+
			"the state this whole change is about", snapshotPath)
	}

	// And it has to be the whole set, not merely a non-empty one.
	//
	// The emptiness check above catches a matcher that stopped matching altogether. It
	// does not catch one that lost a line, and losing a line is the cheaper accident:
	// drop the `realtime` clause from the TOP_GAINERS bullet and leave TOP_LOSERS
	// alone, and the set is {TOP_LOSERS} — non-empty, no Fatal — while a panel wired
	// to TOP_GAINERS at `realtime` passes. The guard would be silent about precisely
	// the combination it was written for, which is this change's own failure shape
	// reproduced one notch narrower inside the test written to end it.
	//
	// So the two types the snapshot refuses `realtime` for are pinned against the real
	// file, the way internal/candidate's TestTheSourcesStillDeclareTheSizesTheseComments
	// Claim pins {100, 30} against this package's source rather than against a copy.
	// The fixture control below does a different job and neither replaces the other:
	// it runs the matcher over a sentence it owns, so it also proves the matcher does
	// not over-match. This proves the file still says what the matcher is reading.
	//
	// Fatal rather than Errorf, for the same reason as the emptiness check: every
	// assertion after this point is computed from `forbidden`, so an incomplete set
	// does not make one answer wrong, it makes all of them unevidenced.
	//
	// The subset is asserted rather than the exact set. A newly documented type that
	// refuses `realtime` is information this guard should act on, not fail over — the
	// intersection below already does that. What must not happen is one going missing.
	//
	// If the API genuinely starts accepting `realtime` for one of these, this is the
	// line to edit, and editing it is a deliberate act performed with the snapshot's
	// own diff next to it. That is the difference between a fact being read and a fact
	// being remembered, and forty-six days of it being remembered is why this file
	// exists.
	for _, typ := range []string{"TOP_GAINERS", "TOP_LOSERS"} {
		if forbidden["realtime"][typ] {
			continue
		}
		t.Fatalf("%s no longer says %s refuses `realtime`, and this guard is built on the "+
			"fact that it does.\n\n"+
			"Either the snapshot was refreshed and the API changed its mind — in which "+
			"case this line is the one to edit, deliberately, with that diff in hand — or "+
			"the bullet was reworded into a shape the matcher no longer recognises. The "+
			"second is the dangerous one: the set stays non-empty, nothing is fatal, and "+
			"the combination this whole change was raised about goes back to passing.\n\n"+
			"Read for \"realtime\" out of the file: %v",
			snapshotPath, typ, sorted(forbidden["realtime"]))
	}

	file := parseWiring(t)
	wired := wiredRankings(t, file)
	durations := wiredDurations(t, file)

	for _, duration := range sorted(durations) {
		for _, typ := range sorted(wired) {
			if !forbidden[duration][typ] {
				continue
			}
			t.Errorf("Panel builds %s and the reader asks for duration %q, which %s says "+
				"that type does not accept (400 unsupported-ranking-duration).\n\n"+
				"This is not a degraded source. A request the API refuses by "+
				"construction cannot succeed on any pass, so reporting it as a "+
				"degradation makes the degradation flag true on every scan and an "+
				"operator can no longer tell a source that just went away from one that "+
				"was never there. The two repairs the spec allows are to change the "+
				"request or to take the source off the panel.\n\n"+
				"Forbidden with %q: %v", typ, duration, snapshotPath, duration,
				sorted(forbidden[duration]))
		}
	}
}

// TestTheSnapshotMatcherStillReadsTheSentenceItWasWrittenFor is the positive
// control.
//
// declaredPanelSizes' Fatal on an empty set is the same idea and it is not enough
// on its own: an empty set is caught, but a matcher that found one bullet out of
// two would pass while missing exactly the type somebody adds next. So the real
// matcher is run over the sentence as the snapshot writes it, and both types are
// required.
func TestTheSnapshotMatcherStillReadsTheSentenceItWasWrittenFor(t *testing.T) {
	const asWritten = "랭킹 종류. 이름에 포함된 지표가 랭킹 기준값입니다\n" +
		"- `MARKET_TRADING_AMOUNT`: 시장 거래대금 상위\n" +
		"- `MARKET_TRADING_VOLUME`: 시장 거래량 상위\n" +
		"- `TOP_GAINERS`: 급상승 (등락률 상위) — `realtime` 미지원\n" +
		"- `TOP_LOSERS`: 급하락 (등락률 하위) — `realtime` 미지원\n" +
		"- `TOSS_SECURITIES_TRADING_AMOUNT`: 토스증권 거래대금 상위\n"

	got := parseForbidden(t, asWritten)
	for _, want := range []string{"TOP_GAINERS", "TOP_LOSERS"} {
		if !got["realtime"][want] {
			t.Errorf("the matcher no longer reads %s out of the bullet the snapshot "+
				"writes it in. Read for realtime: %v", want, sorted(got["realtime"]))
		}
	}
	// And it does not invent a constraint for the types that carry none: a matcher
	// that swept up every bullet would forbid the two rankings this panel depends on.
	for _, fine := range []string{"MARKET_TRADING_AMOUNT", "MARKET_TRADING_VOLUME"} {
		if got["realtime"][fine] {
			t.Errorf("the matcher reads %s as refusing realtime. It does not — those two "+
				"are the rankings that answer, measured 2026-07-28 — and a guard that "+
				"forbids the working wiring is one somebody deletes", fine)
		}
	}
	if len(got) != 1 {
		t.Errorf("the matcher read %d durations out of a fixture that names one: %v",
			len(got), got)
	}
}

// sorted is a stable rendering of a set, so a failure message is the same on two
// runs.
func sorted(set map[string]bool) []string {
	out := make([]string, 0, len(set))
	for k := range set {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
