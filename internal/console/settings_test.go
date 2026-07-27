package console

// settings_test.go covers the adoption-settings screen and the per-symbol
// designation (console-adoption-controls tasks 3.1-3.3). The seam is a counting
// fake, because "a refused request wrote nothing" is the claim that matters.

import (
	"errors"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/config"
)

type fakeSettings struct {
	mu      sync.Mutex
	block   config.Adoption
	verdict string
	loadErr error
	saves   int
}

func (f *fakeSettings) Load() (config.Adoption, string, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.block, f.verdict, f.loadErr
}

// Save mirrors config.Service.SaveEngineAdoption's contract: a block the engine
// would zero is refused, never written.
func (f *fakeSettings) Save(a config.Adoption) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	probe := a
	probe.Rejected = ""
	if !probe.Enabled && len(probe.IncludeSymbols) == 0 && probe.DefaultStopPct == 0 {
		// the absent block — fine
	} else if probe.DefaultStopPct < 0.02 || probe.DefaultStopPct >= 1 {
		return errors.New("stop fraction 0.02 ≤ pct < 1 refused in test fake")
	}
	f.block = a
	f.verdict = ""
	f.saves++
	return nil
}

func (f *fakeSettings) saved() (config.Adoption, int) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.block, f.saves
}

func settingsHarness(t *testing.T, seam *fakeSettings) *dashboardHarness {
	t.Helper()
	return newDashboardHarness(t, func(o *Options) {
		if seam != nil {
			o.Settings = seam
		}
	})
}

// TestTheSettingsScreenShowsTheRawBlockAndTheVerdict: a refused block renders
// the file's own lists next to the reason — the read side of P1-1.
func TestTheSettingsScreenShowsTheRawBlockAndTheVerdict(t *testing.T) {
	seam := &fakeSettings{
		block: config.Adoption{Enabled: true, DefaultStopPct: 0.005,
			ExcludeSymbols: []string{"005930"}, IncludeSymbols: []string{"000660"}},
		verdict: "stop fraction out of band",
	}
	h := settingsHarness(t, seam)
	h.authenticate(t)

	page := h.page(t, "/settings")
	for _, want := range []string{
		"엔진이 거부한다", "stop fraction out of band", "005930", "000660",
		"다음 엔진 기동부터 반영", "상시 규칙", "편입 해제 기능은 존재하지 않는다",
		"automation gate", "콘솔에서 편집할 수 없다",
	} {
		if !strings.Contains(page, want) {
			t.Errorf("the settings screen does not say %q", want)
		}
	}
}

func TestAnUnwiredSettingsSeamIsExplained(t *testing.T) {
	h := settingsHarness(t, nil)
	h.authenticate(t)
	if page := h.page(t, "/settings"); !strings.Contains(page, "배선되지 않았다") {
		t.Error("an unwired seam must be explained, not rendered as an empty form")
	}
}

// TestSavingTheFormWritesTheBlock is the happy path.
func TestSavingTheFormWritesTheBlock(t *testing.T) {
	seam := &fakeSettings{}
	h := settingsHarness(t, seam)
	h.authenticate(t)

	resp := h.post(t, "/settings/save", url.Values{
		"csrf":             {h.csrf},
		"default_stop_pct": {"0.05"},
		"include_symbols":  {"005930, 000660"},
	})
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("save = %d", resp.StatusCode)
	}
	block, saves := seam.saved()
	if saves != 1 || block.DefaultStopPct != 0.05 || len(block.IncludeSymbols) != 2 {
		t.Errorf("saved block = %+v after %d saves", block, saves)
	}
	if block.Enabled {
		t.Error("enabled turned itself on")
	}
}

// TestAnInvalidSaveWritesNothing: the seam refuses, the screen explains, config
// is unchanged.
func TestAnInvalidSaveWritesNothing(t *testing.T) {
	seam := &fakeSettings{}
	h := settingsHarness(t, seam)
	h.authenticate(t)

	h.post(t, "/settings/save", url.Values{
		"csrf":             {h.csrf},
		"default_stop_pct": {"0.001"},
		"include_symbols":  {"005930"},
	})
	if _, saves := seam.saved(); saves != 0 {
		t.Error("an out-of-band fraction was written; the engine would zero it")
	}
}

// TestSettingsPostsWithoutCSRFWriteNothing: the mutating gate is the spec's
// scenario — session alone must not reach the seam.
func TestSettingsPostsWithoutCSRFWriteNothing(t *testing.T) {
	seam := &fakeSettings{}
	h := settingsHarness(t, seam)
	h.authenticate(t)

	for _, path := range []string{"/settings/save", "/settings/include"} {
		resp := h.post(t, path, url.Values{
			"symbol": {"005930"}, "default_stop_pct": {"0.05"},
		})
		if resp.StatusCode == http.StatusOK && resp.Request.URL.Path != "/refused" {
			// The console refuses with its own page; what matters is the seam.
			_ = resp
		}
	}
	if _, saves := seam.saved(); saves != 0 {
		t.Error("a POST without the CSRF token reached the seam")
	}
}

// TestDesignatingASymbolFromThePositionsScreen: the row's button adds the symbol
// to the include list — idempotently — and the response says it is a standing
// rule taking effect at the next engine start.
func TestDesignatingASymbolFromThePositionsScreen(t *testing.T) {
	seam := &fakeSettings{block: config.Adoption{DefaultStopPct: 0.05}}
	h := settingsHarness(t, seam)
	seedJournal(t, h.journal)
	h.authenticate(t)

	// The button renders on the unmanaged broker-only row, not on the managed one.
	page := h.page(t, "/positions")
	if !strings.Contains(page, `type="checkbox"`) || !strings.Contains(page, `action="/settings/include"`) {
		t.Fatal("the positions screen offers no designation checkbox for an unmanaged holding")
	}

	for i := 0; i < 2; i++ {
		h.post(t, "/settings/include", url.Values{
			"csrf": {h.csrf}, "symbol": {"000660"},
		})
	}
	block, saves := seam.saved()
	if saves != 2 || len(block.IncludeSymbols) != 1 || !block.Included("000660") {
		t.Errorf("designation not idempotent: %+v saves=%d", block, saves)
	}

	// The row now shows its designation instead of the button.
	if page := h.page(t, "/positions"); !strings.Contains(page, "편입 예약됨") {
		t.Error("a designated row must say 편입 예약됨 prominently")
	}
}

// TestDesignationAppliesTheDefaultStopFraction (사용자 UX 결정 2026-07-27): the
// one-click designation never bounces the user to a form. With no valid
// fraction chosen, the console writes its 5% default explicitly and says so —
// the engine still only ever runs on an explicit number.
func TestDesignationAppliesTheDefaultStopFraction(t *testing.T) {
	seam := &fakeSettings{}
	h := settingsHarness(t, seam)
	h.authenticate(t)

	resp := h.post(t, "/settings/include", url.Values{
		"csrf": {h.csrf}, "symbol": {"000660"},
	})
	block, saves := seam.saved()
	if saves != 1 || !block.Included("000660") || block.DefaultStopPct != 0.05 {
		t.Fatalf("designation with no fraction: %+v saves=%d, want the 5%% default written "+
			"explicitly", block, saves)
	}
	if body := body(t, resp); !strings.Contains(body, "기본값 5%") {
		t.Errorf("the response must say the default applied: %q", body)
	}
}

// TestRemovingADesignationOnlyAffectsTheFuture: the row's 해제 button drops the
// symbol from the include list; the response says adopted positions stay.
func TestRemovingADesignationOnlyAffectsTheFuture(t *testing.T) {
	seam := &fakeSettings{block: config.Adoption{DefaultStopPct: 0.05,
		IncludeSymbols: []string{"000660"}}}
	h := settingsHarness(t, seam)
	h.authenticate(t)

	resp := h.post(t, "/settings/include", url.Values{
		"csrf": {h.csrf}, "symbol": {"000660"}, "remove": {"1"},
	})
	block, saves := seam.saved()
	if saves != 1 || block.Included("000660") {
		t.Fatalf("removal did not drop the symbol: %+v saves=%d", block, saves)
	}
	if body := body(t, resp); !strings.Contains(body, "이미 편입된 포지션") {
		t.Errorf("the response must say adopted positions are unaffected: %q", body)
	}
}

// TestTheStopFractionIsASlider: no typing — a mouse-adjustable range input with
// the default labelled (사용자 UX 결정 2026-07-27).
func TestTheStopFractionIsASlider(t *testing.T) {
	h := settingsHarness(t, &fakeSettings{})
	h.authenticate(t)
	page := h.page(t, "/settings")
	for _, want := range []string{`type="range"`, "기본값 5%", `name="default_stop_pct"`} {
		if !strings.Contains(page, want) {
			t.Errorf("the settings screen does not carry %q; the fraction must be a slider "+
				"with its default shown, not a text field", want)
		}
	}
	if strings.Contains(page, `type="text" name="default_stop_pct"`) {
		t.Error("the fraction is still a text field")
	}
}
