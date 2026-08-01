package console

// settings_exclude_test.go covers the one-click exclusion (change
// console-excludes-in-one-click, tasks 1.4-1.19).
//
// The claims that matter are not "a symbol landed in a list". They are:
//   - the click writes ONE field and leaves the rest of the block as it read it,
//   - it never invents a stop fraction the operator did not choose,
//   - it cannot leave the file saying two things the engine resolves silently,
//   - and every row's verdict tells the truth about which of the two it is.

import (
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/config"
)

// TestExcludingASymbolThroughTheSettingsEndpoint: the guarded endpoint remains
// idempotent while the positions trading view contains no mutation control.
//
// "Touches nothing else" is the whole reason this is not the full form: the
// save the operator used to have to make was a read-modify-write through a
// textarea, and a dropped symbol there is a holding that quietly becomes
// adoptable again.
func TestExcludingASymbolFromThePositionsScreen(t *testing.T) {
	seam := &fakeSettings{block: config.Adoption{
		Enabled: true, DefaultStopPct: 0.07, IncludeSymbols: []string{"005930"},
	}}
	h := settingsHarness(t, seam)
	seedJournal(t, h.journal)
	h.authenticate(t)

	if page := h.page(t, "/positions"); strings.Contains(page, `action="/settings/exclude"`) ||
		!strings.Contains(page, `href="/optimization?category=position-management"`) {
		t.Fatal("the positions screen must route management to the canonical surface without a form")
	}

	for i := 0; i < 2; i++ {
		h.post(t, "/settings/exclude", url.Values{"csrf": {h.csrf}, "symbol": {"000660"}})
	}

	block, saves := seam.saved()
	switch {
	case saves != 2:
		t.Fatalf("saves = %d, want 2 — both clicks must reach the seam", saves)
	case len(block.ExcludeSymbols) != 1 || !block.Excludes("000660"):
		t.Errorf("exclusion not idempotent: %+v", block.ExcludeSymbols)
	case !block.Enabled:
		t.Error("the exclusion turned adoption off; it must write one field only")
	case block.DefaultStopPct != 0.07:
		t.Errorf("default_stop_pct moved to %v; the exclusion must not rewrite it", block.DefaultStopPct)
	case len(block.IncludeSymbols) != 1 || !block.Included("005930"):
		t.Errorf("the include list changed to %+v; the exclusion must leave it alone",
			block.IncludeSymbols)
	}
}

// TestExcludingASymbolIsCaseAndSpaceInsensitive: the row posts what the broker
// spelled; the list is upper-case.
func TestExcludingASymbolIsCaseAndSpaceInsensitive(t *testing.T) {
	seam := &fakeSettings{}
	h := settingsHarness(t, seam)
	h.authenticate(t)

	h.post(t, "/settings/exclude", url.Values{"csrf": {h.csrf}, "symbol": {"  tsla "}})

	if block, _ := seam.saved(); !block.Excludes("TSLA") || len(block.ExcludeSymbols) != 1 {
		t.Errorf("exclude list = %+v, want exactly [TSLA]", block.ExcludeSymbols)
	}
}

// TestReleasingAnExclusion: the checked control releases, and only that symbol.
func TestReleasingAnExclusion(t *testing.T) {
	seam := &fakeSettings{block: config.Adoption{
		DefaultStopPct: 0.05, ExcludeSymbols: []string{"000660", "035420"},
	}}
	h := settingsHarness(t, seam)
	h.authenticate(t)

	resp := h.post(t, "/settings/exclude", url.Values{
		"csrf": {h.csrf}, "symbol": {"000660"}, "remove": {"1"},
	})

	block, saves := seam.saved()
	switch {
	case saves != 1:
		t.Fatalf("saves = %d, want 1", saves)
	case block.Excludes("000660"):
		t.Error("the release did not drop the symbol")
	case !block.Excludes("035420"):
		t.Error("the release dropped a symbol nobody asked about — this is the list-loss " +
			"the one-click path exists to prevent")
	}
	if b := body(t, resp); !strings.Contains(b, "제외 해제") {
		t.Errorf("the response does not say the exclusion was released: %q", b)
	}
}

// TestExclusionNeverInventsAStopFraction (design D4).
//
// The include path writes a 5% default because Adoption.validate() refuses a
// block that carries an include list with no fraction. ExcludeSymbols is not in
// that condition, so the exclusion has no such excuse: a stop fraction the
// operator never chose must not appear in the file as a side effect of saying
// "leave this holding alone".
func TestExclusionNeverInventsAStopFraction(t *testing.T) {
	seam := &fakeSettings{}
	h := settingsHarness(t, seam)
	h.authenticate(t)

	h.post(t, "/settings/exclude", url.Values{"csrf": {h.csrf}, "symbol": {"000660"}})

	block, saves := seam.saved()
	if saves != 1 || !block.Excludes("000660") {
		t.Fatalf("the exclusion did not save: %+v saves=%d", block, saves)
	}
	if block.DefaultStopPct != 0 {
		t.Errorf("default_stop_pct = %v after an exclusion; the operator chose no fraction and "+
			"the exclude list does not require one", block.DefaultStopPct)
	}
}

// TestExcludingADesignatedSymbolDropsTheDesignation (design D3).
//
// The engine resolves the contradiction by ignoring the designation. A file
// that carries both is a file whose include list lies, so the console does not
// write one — and it says which half it dropped.
func TestExcludingADesignatedSymbolDropsTheDesignation(t *testing.T) {
	seam := &fakeSettings{block: config.Adoption{
		DefaultStopPct: 0.05, IncludeSymbols: []string{"000660", "005930"},
	}}
	h := settingsHarness(t, seam)
	h.authenticate(t)

	resp := h.post(t, "/settings/exclude", url.Values{"csrf": {h.csrf}, "symbol": {"000660"}})

	block, _ := seam.saved()
	switch {
	case !block.Excludes("000660"):
		t.Fatal("the symbol was not excluded")
	case block.Included("000660"):
		t.Error("the symbol is on BOTH lists; the engine would ignore the designation silently")
	case !block.Included("005930"):
		t.Error("an unrelated designation was dropped")
	}
	if b := body(t, resp); !strings.Contains(b, "편입 지정") {
		t.Errorf("the response does not say the designation was released too: %q", b)
	}
}

// TestDesignatingAnExcludedSymbolSaysTheExclusionWins (design D6).
//
// Hiding the checkbox on an excluded row is a UI decision, not enforcement. A
// direct POST still reaches the handler, and the one thing it must not do is
// answer "편입 예약됨 — 상시 규칙이다" about a symbol the engine will refuse.
func TestDesignatingAnExcludedSymbolSaysTheExclusionWins(t *testing.T) {
	seam := &fakeSettings{block: config.Adoption{
		DefaultStopPct: 0.05, ExcludeSymbols: []string{"000660"},
	}}
	h := settingsHarness(t, seam)
	h.authenticate(t)

	resp := h.post(t, "/settings/include", url.Values{"csrf": {h.csrf}, "symbol": {"000660"}})

	b := body(t, resp)
	if !strings.Contains(b, "제외가 편입보다 우선") {
		t.Errorf("the response does not say the exclusion overrides this designation: %q", b)
	}
	if strings.Contains(b, "편입 예약됨") {
		t.Errorf("the response reports a reservation the engine will not honour: %q", b)
	}
	if block, _ := seam.saved(); !block.Excludes("000660") {
		t.Error("the designation lifted the exclusion; that is the expanding direction and it " +
			"must stay an explicit act")
	}
}

// TestTheExclusionAnswerDefersToTheEngineRestart (adversarial review A1).
//
// A running engine keeps the snapshot it started with, so it can adopt this very
// symbol on its next cycle — and once adopted, the exclusion is inert for that
// position forever. The answer must therefore record a setting and say when it
// bites, never promise a present-tense "the engine will not adopt this".
func TestTheExclusionAnswerDefersToTheEngineRestart(t *testing.T) {
	seam := &fakeSettings{}
	h := settingsHarness(t, seam)
	h.authenticate(t)

	b := body(t, h.post(t, "/settings/exclude", url.Values{
		"csrf": {h.csrf}, "symbol": {"000660"},
	}))
	if !strings.Contains(b, "반영") {
		t.Errorf("the exclusion answer does not say when it takes effect: %q", b)
	}
	if !strings.Contains(b, "기록됨") {
		t.Errorf("the exclusion answer must report a recorded setting rather than an effect "+
			"the running engine has not read yet: %q", b)
	}
}

// The read-only trading view must not retain a direction to a removed row
// control.
func TestTheReleaseHintOnlyAppearsWhereTheControlIs(t *testing.T) {
	seam := &fakeSettings{block: config.Adoption{
		ExcludeSymbols: []string{"000660", "035420", "005930", "005380"},
	}}
	h := settingsHarness(t, seam)
	seedJournal(t, h.journal)
	h.authenticate(t)

	page := h.page(t, "/positions")
	if strings.Contains(page, "편입하려면") || strings.Contains(page, `action="/settings/exclude"`) {
		t.Error("the input-free view retains an obsolete release instruction or control")
	}
}

// TestAnExcludedRowIsLabelledWithoutEmbeddingAReleaseControl.
func TestAnExcludedRowIsLabelledAndOffersRelease(t *testing.T) {
	seam := &fakeSettings{block: config.Adoption{ExcludeSymbols: []string{"000660"}}}
	h := settingsHarness(t, seam)
	seedJournal(t, h.journal)
	h.authenticate(t)

	row := rowFor(t, h.page(t, "/positions"), "000660")
	switch {
	case !strings.Contains(row, "관리 제외"):
		t.Errorf("an excluded row is not labelled 관리 제외:\n%s", row)
	case strings.Contains(row, "관리 외(미편입)"):
		t.Errorf("an excluded row still reads as merely undesignated:\n%s", row)
	case !strings.Contains(row, "자동관리 제외 정책 적용 중"):
		t.Errorf("an excluded row does not explain the applied policy:\n%s", row)
	case strings.Contains(row, `action="/settings/exclude"`):
		t.Errorf("an excluded row embeds a settings mutation control:\n%s", row)
	case strings.Contains(row, `action="/settings/include"`):
		t.Errorf("an excluded row still offers designation; the engine would ignore it:\n%s", row)
	}
}

// TestExclusionBeatsDesignationInTheLabel: the screen's precedence must be the
// engine's precedence, or the label mispredicts what the engine will do.
func TestExclusionBeatsDesignationInTheLabel(t *testing.T) {
	seam := &fakeSettings{block: config.Adoption{DefaultStopPct: 0.05,
		ExcludeSymbols: []string{"000660"}, IncludeSymbols: []string{"000660"}}}
	h := settingsHarness(t, seam)
	seedJournal(t, h.journal)
	h.authenticate(t)

	row := rowFor(t, h.page(t, "/positions"), "000660")
	if !strings.Contains(row, "관리 제외") || strings.Contains(row, "관리 편입") {
		t.Errorf("a row on both lists must read 관리 제외 — exclusion wins in the engine:\n%s", row)
	}
}

// TestAnUnreadableJournalStaysUnknownEvenWhenExcluded: a console that could not
// open the ledger has observed nothing to label.
func TestAnUnreadableJournalStaysUnknownEvenWhenExcluded(t *testing.T) {
	seam := &fakeSettings{block: config.Adoption{ExcludeSymbols: []string{"000660"}}}
	h := settingsHarness(t, seam)
	// No journal is seeded, so the ledger never answers.
	h.authenticate(t)

	row := rowFor(t, h.page(t, "/positions"), "000660")
	if !strings.Contains(row, "관리 여부 불명") || strings.Contains(row, "관리 제외") {
		t.Errorf("an unreadable ledger must stay 관리 여부 불명 regardless of the lists:\n%s", row)
	}
}

// The positions trading view embeds no exclusion control on any row.
func TestTheExcludeControlOnlyRendersOnUnmanagedKnownRows(t *testing.T) {
	seam := &fakeSettings{block: config.Adoption{DefaultStopPct: 0.05}}
	h := settingsHarness(t, seam)
	seedJournal(t, h.journal)
	h.authenticate(t)

	page := h.page(t, "/positions")
	if managed := rowFor(t, page, "005930"); strings.Contains(managed, `action="/settings/exclude"`) {
		t.Errorf("a managed row offers an exclusion that cannot take effect:\n%s", managed)
	}
	if unmanaged := rowFor(t, page, "000660"); strings.Contains(unmanaged, `action="/settings/exclude"`) {
		t.Errorf("an unmanaged row embeds an exclusion control:\n%s", unmanaged)
	}
}

// TestThePositionsViewAsksForNoTypingOrClickMutation.
func TestTheExcludeControlAsksForNoTyping(t *testing.T) {
	seam := &fakeSettings{block: config.Adoption{DefaultStopPct: 0.05}}
	h := settingsHarness(t, seam)
	seedJournal(t, h.journal)
	h.authenticate(t)

	page := h.page(t, "/positions")
	switch {
	case strings.Contains(page, "confirm("):
		t.Error("the controls still depend on a CSP-blocked confirmation handler")
	case strings.Contains(page, "<button") || strings.Contains(page, "<form"):
		t.Error("the trading view embeds a mutation control")
	case strings.Contains(page, "prompt("):
		t.Error("the positions screen asks the operator to type something")
	case strings.Contains(page, "<input") || strings.Contains(page, "<textarea") || strings.Contains(page, "<select"):
		t.Error("the positions screen grew an input field")
	}
}

// TestWithoutASeamNeitherControlRenders: an unwired console must not draw a
// button that cannot write.
func TestWithoutASeamNeitherControlRenders(t *testing.T) {
	h := settingsHarness(t, nil)
	seedJournal(t, h.journal)
	h.authenticate(t)

	page := h.page(t, "/positions")
	if strings.Contains(page, `action="/settings/exclude"`) ||
		strings.Contains(page, `action="/settings/include"`) {
		t.Error("an unwired console still draws the config controls")
	}
}

// TestExcludeRefusesWhatItCannotDo: no seam is 501, no symbol is 400, and
// neither writes.
func TestExcludeRefusesWhatItCannotDo(t *testing.T) {
	unwired := settingsHarness(t, nil)
	unwired.authenticate(t)
	resp := unwired.post(t, "/settings/exclude", url.Values{
		"csrf": {unwired.csrf}, "symbol": {"000660"},
	})
	if resp.StatusCode != http.StatusNotImplemented {
		t.Errorf("an unwired seam answered %d, want 501", resp.StatusCode)
	}

	seam := &fakeSettings{}
	h := settingsHarness(t, seam)
	h.authenticate(t)
	if resp := h.post(t, "/settings/exclude", url.Values{"csrf": {h.csrf}}); resp.StatusCode !=
		http.StatusBadRequest {
		t.Errorf("a symbol-less exclusion answered %d, want 400", resp.StatusCode)
	}
	if _, saves := seam.saved(); saves != 0 {
		t.Error("a refused exclusion still reached the seam")
	}
}
