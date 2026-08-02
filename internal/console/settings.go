package console

// settings.go is the adoption-settings surface (change console-adoption-controls
// tasks 3.1/3.2): one GET screen, two CSRF-gated POSTs, and a seam.
//
// The console does not adopt and cannot: both actions edit the engine.adoption
// config block through the injected seam, and the engine's reconcile loop — on
// its next start, behind its own interlock — is the only thing that acts on it.
// The seam's Load returns the block as the FILE spells it plus a verdict,
// because the merge's zeroing of a refused block must never round-trip through
// this screen and erase somebody's exclude list (review round 1, P1-1).

import (
	"fmt"
	"math"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/JungHoonGhae/tossinvest-cli/internal/config"
	"github.com/JungHoonGhae/tossinvest-cli/internal/localupdate"
	"github.com/JungHoonGhae/tossinvest-cli/internal/runlock"
)

// AdoptionSettings is the console's entire write surface for config: the
// engine.adoption block, read raw and written surgically. Exactly these two
// methods — static_test.go fails if the interface grows.
type AdoptionSettings interface {
	// Load returns the block as written in the file, the validation verdict
	// ("" when usable), and an error for an unreadable file.
	Load() (config.Adoption, string, error)
	// Save validates and writes the block, touching nothing else in the file.
	Save(config.Adoption) error
}

// defaultStopPct is the value the console fills in when the operator has never
// chosen one — 5%, written explicitly into the block it saves (and audited), so
// the engine still only ever runs on an explicit number. Shown on the slider as
// 기본값 (사용자 UX 결정 2026-07-27: 지정 버튼이 입력을 요구하며 거부하지 않는다).
const defaultStopPct = 0.05

type settingsPage struct {
	chrome
	// Tab is which of the four sub-screens this render is, and Tabs is the bar.
	// The classification is by reversibility and frequency, not by feature name
	// (change a055) — see settings_tabs.go.
	Tab  string
	Tabs []settingsTab
	// Entries are the links out of 전략 and 도구, each carrying the target's own
	// current line.
	Entries []entryPoint

	CSRF   string
	Notice string
	// NoticeForm names the card the notice belongs beside. A save answer at the
	// top of a screen with eight forms on it does not say which one saved.
	NoticeForm string
	// Wired reports the seam was injected; without it the screen explains and
	// the forms are not rendered.
	Wired bool
	// Block is the file's own spelling; Verdict is why the engine would refuse
	// it ("" when it would not). LoadErr is an unreadable file.
	Block   config.Adoption
	Verdict string
	LoadErr string
	// EngineRunning is the advisory marker's answer, for the honesty line: a
	// running engine keeps its startup snapshot until restarted.
	EngineRunning bool

	// --- the Guardian limits (change console-sets-guardian-limits) ---

	// LimitsWired reports the limit seam was injected. It is separate from Wired
	// because the two seams are separate: a build with one and not the other
	// renders the section it can serve and explains the other.
	LimitsWired bool
	// Gate is the automation gate block as the file spells it — `enabled`
	// included, for display. Nothing on this page posts it back.
	Gate config.AutomationGate
	// LimitsLoadErr is an unreadable config on the limit side. It must not take
	// the adoption section down with it.
	LimitsLoadErr string

	// --- the operating toggles (change console-owns-the-operating-toggles) ---

	// TradingWired and GateWired report their seams separately, for the reason
	// LimitsWired gives: a build with one and not the other renders what it can
	// serve and explains the rest.
	TradingWired bool
	GateWired    bool
	// Trading is the trading block as the file spells it. The screen edits four
	// of its fields; the rest are displayed nowhere and posted back never.
	Trading config.Trading
	// TradingLoadErr is an unreadable config on the policy side, isolated from
	// the other two sections for the same reason.
	TradingLoadErr string

	// --- engine lifecycle approval (change enable-engine-autostart-menu) ---
	AutostartWired   bool
	Autostart        bool
	AutostartLoadErr string
	AutostartNote    string

	// --- staged system update (change console-system-update) ---
	UpdateWired          bool
	ReleaseDownloadWired bool
	Update               localupdate.Inspection
	ReleaseVerified      bool
	ReleaseReceipt       signedReleaseReceipt
}

func (settingsPage) Refresh() bool { return false }

// Excludes and Includes render the lists as editable text.
func (p settingsPage) Excludes() string { return strings.Join(p.Block.ExcludeSymbols, ", ") }
func (p settingsPage) Includes() string { return strings.Join(p.Block.IncludeSymbols, ", ") }

// StopPct renders the fraction, empty when unset.
func (p settingsPage) StopPct() string {
	if p.Block.DefaultStopPct == 0 {
		return ""
	}
	return strconv.FormatFloat(p.Block.DefaultStopPct, 'f', -1, 64)
}

// StopPctSlider is the percentage input text: the saved fraction converted to a
// human percentage, or the default.
func (p settingsPage) StopPctSlider() string {
	pct := p.Block.DefaultStopPct
	if math.IsNaN(pct) || math.IsInf(pct, 0) || pct < 0.02 || pct >= 1 {
		pct = defaultStopPct
	}
	return fractionPercentText(pct)
}

// StopPctPercent renders the percentage input with its unit.
func (p settingsPage) StopPctPercent() string {
	return p.StopPctSlider() + "%"
}

// engineRunning reads the enginelock marker's freshness — advisory, exactly as
// the dashboard draws it.
func (c *Console) engineRunning() bool {
	path := strings.TrimSpace(c.opts.EngineMarker)
	if path == "" {
		return false
	}
	fresh, _ := runlock.Fresh(path, c.now(), runlock.StaleAfter)
	return fresh
}

// settingsView reads everything the settings screen shows, for whichever tab is
// about to render it.
//
// All four tabs read the same things. Splitting the reads per tab would mean
// four places to keep in step with a new seam, and the reads are file stats and
// a config parse — the expensive screens in this console are the ones that call
// the broker, and this is not one of them (a054 pinned that at zero).
func (c *Console) settingsView(r *http.Request) settingsPage {
	page := settingsPage{
		chrome:        c.chromeOnRequest("settings"),
		Tabs:          settingsTabs,
		CSRF:          c.csrf,
		Notice:        r.URL.Query().Get("notice"),
		NoticeForm:    strings.TrimSpace(r.URL.Query().Get("form")),
		EngineRunning: c.engineRunning(),
	}
	if c.opts.Settings != nil {
		page.Wired = true
		block, verdict, err := c.opts.Settings.Load()
		if err != nil {
			page.LoadErr = err.Error()
		}
		page.Block, page.Verdict = block, verdict
	}
	if c.opts.Limits != nil {
		page.LimitsWired = true
		gate, err := c.opts.Limits.Load()
		if err != nil {
			page.LimitsLoadErr = err.Error()
		}
		page.Gate = gate
	}
	if c.opts.TradingPolicy != nil {
		page.TradingWired = true
		trading, err := c.opts.TradingPolicy.Load()
		if err != nil {
			page.TradingLoadErr = err.Error()
		}
		page.Trading = trading
	}
	// The gate seam is write-only: LimitSettings.Load already returned the whole
	// block including `enabled`, and a second reader of one key is how a screen
	// ends up disagreeing with itself.
	page.GateWired = c.opts.Gate != nil
	if c.opts.EngineBoot != nil {
		page.AutostartWired = true
		on, err := c.opts.EngineBoot.Load()
		if err != nil {
			page.AutostartLoadErr = err.Error()
		}
		page.Autostart = on
	}
	page.AutostartNote, _ = c.engineNoteNow()
	if c.opts.SystemUpdater != nil {
		page.UpdateWired = true
		page.Update = c.opts.SystemUpdater.Inspect()
		page.ReleaseReceipt, page.ReleaseVerified =
			c.signedReleaseReceipt(page.Update.Candidate.SHA256)
	}
	page.ReleaseDownloadWired =
		c.opts.ReleaseDownloader != nil && c.opts.ReleaseCandidateStager != nil
	return page
}

// handleSettingsSave is the whole form: enabled, fraction, both lists.
func (c *Console) handleSettingsSave(w http.ResponseWriter, r *http.Request) {
	if c.opts.Settings == nil {
		c.refuse(w, http.StatusNotImplemented, "저장이 배선되지 않았다",
			"이 빌드의 콘솔에는 편입 설정 저장 seam이 주입되지 않았다.")
		return
	}
	_, _, err := c.opts.Settings.Load()
	if err != nil {
		c.redirectSettings(w, r, "저장 안 됨 — 설정 파일을 읽을 수 없다: "+err.Error())
		return
	}

	next := config.Adoption{
		Enabled:        r.PostFormValue("enabled") == "on",
		ExcludeSymbols: splitSymbols(r.PostFormValue("exclude_symbols")),
		IncludeSymbols: splitSymbols(r.PostFormValue("include_symbols")),
	}
	stopFraction, err := parseStopPercent(r.PostFormValue("default_stop_percent"))
	if err != nil {
		c.redirectSettings(w, r, fmt.Sprintf(
			"저장 안 됨 — 합성 손절폭은 2%%~20%% 범위에서 0.5%% 단위로 입력해야 한다: %v", err))
		return
	}
	next.DefaultStopPct = stopFraction

	// Turning enabled on is deliberately NOT behind a typed phrase: the user
	// rejected that friction (사용자 결정 2026-07-27 — review.md). §0.7 is still
	// a human act (loopback session + CSRF + this click), both audit records
	// still fire, and the flip has no live effect before an engine start behind
	// the gate interlock.
	if err := c.opts.Settings.Save(next); err != nil {
		c.redirectSettings(w, r, "저장 안 됨 — "+err.Error())
		return
	}
	c.redirectSettings(w, r, "저장됨. "+effectNotice(c.engineRunning()))
}

// handleSettingsInclude designates one symbol from the positions screen.
func (c *Console) handleSettingsInclude(w http.ResponseWriter, r *http.Request) {
	if c.opts.Settings == nil {
		c.refuse(w, http.StatusNotImplemented, "지정이 배선되지 않았다",
			"이 빌드의 콘솔에는 편입 설정 저장 seam이 주입되지 않았다.")
		return
	}
	symbol := strings.ToUpper(strings.TrimSpace(r.PostFormValue("symbol")))
	if symbol == "" {
		c.refuse(w, http.StatusBadRequest, "심볼이 없다", "편입 지정에는 심볼이 필요하다.")
		return
	}
	current, _, err := c.opts.Settings.Load()
	if err != nil {
		c.redirectSettings(w, r, "지정 안 됨 — 설정 파일을 읽을 수 없다: "+err.Error())
		return
	}

	next := current
	next.Rejected = ""

	if r.PostFormValue("remove") == "1" {
		next.IncludeSymbols = withoutSymbol(next.IncludeSymbols, symbol)
		if err := c.opts.Settings.Save(next); err != nil {
			c.redirectSettings(w, r, "해제 안 됨 — "+err.Error())
			return
		}
		c.redirectSettings(w, r, symbol+" 지정 해제됨 — 장래 편입에만 영향이 있고, 이미 편입된 포지션은 그대로 관리된다.")
		return
	}

	if !next.Included(symbol) {
		next.IncludeSymbols = append(append([]string(nil), next.IncludeSymbols...), symbol)
		sort.Strings(next.IncludeSymbols)
	}
	// The engine resolves exclusion ∧ designation by ignoring the designation, so
	// the standing-rule sentence below is false for an excluded symbol. The row
	// hides this control on such a row (templates_portfolio.go), but hiding is a
	// UI decision and this handler is still reachable — refusing would be friction
	// and recording is harmless, so what changes is that the answer stops claiming
	// a reservation the engine will not honour (design D6).
	reserved := symbol + " 편입 예약됨 — 상시 규칙이다(청산 후 재매수도 다음 대사 주기에 재편입)."
	if next.Excludes(symbol) {
		reserved = symbol + " 지정은 기록됐으나 편입되지 않는다 — 이 심볼은 제외 목록에 있고 " +
			"제외가 편입보다 우선한다. 편입하려면 제외를 먼저 해제한다."
	}
	// The one-click designation must not bounce the user to a form (사용자 UX
	// 결정 2026-07-27): when no valid stop fraction was ever chosen, the console
	// fills in its default explicitly — the engine still never runs on an
	// implicit number, and the response says which fraction applied. A block the
	// engine would zero is still never written (Save validates).
	usedDefault := ""
	if next.DefaultStopPct < 0.02 || next.DefaultStopPct >= 1 {
		next.DefaultStopPct = defaultStopPct
		usedDefault = " 합성 손절폭은 기본값 5%가 적용됐다(편입 설정에서 조절 가능)."
	}
	if err := c.opts.Settings.Save(next); err != nil {
		c.redirectSettings(w, r, "지정 안 됨 — "+err.Error())
		return
	}
	c.redirectSettings(w, r, reserved+usedDefault+" "+effectNotice(c.engineRunning()))
}

// handleSettingsExclude designates one symbol as never-adopted, from the row
// that already names it.
//
// It is deliberately not a call into the full form. The form's save rebuilds the
// whole block from what the browser sent back, so excluding one symbol that way
// means re-sending the other three values — and a list that fails to make the
// round trip is a long-term holding that quietly becomes adoptable again. That
// is the failure LoadRawEngineAdoption was built to prevent
// (console-adoption-controls review round 1, P1-1); the textarea reopened it at
// the keyboard. Here the operator sends one symbol and the server re-reads the
// block, so a value this request never mentioned cannot be lost by it.
func (c *Console) handleSettingsExclude(w http.ResponseWriter, r *http.Request) {
	if c.opts.Settings == nil {
		c.refuse(w, http.StatusNotImplemented, "제외가 배선되지 않았다",
			"이 빌드의 콘솔에는 편입 설정 저장 seam이 주입되지 않았다.")
		return
	}
	symbol := strings.ToUpper(strings.TrimSpace(r.PostFormValue("symbol")))
	if symbol == "" {
		c.refuse(w, http.StatusBadRequest, "심볼이 없다", "제외 지정에는 심볼이 필요하다.")
		return
	}
	current, _, err := c.opts.Settings.Load()
	if err != nil {
		c.redirectSettings(w, r, "제외 안 됨 — 설정 파일을 읽을 수 없다: "+err.Error())
		return
	}

	next := current
	next.Rejected = ""

	if r.PostFormValue("remove") == "1" {
		next.ExcludeSymbols = withoutSymbol(next.ExcludeSymbols, symbol)
		if err := c.opts.Settings.Save(next); err != nil {
			c.redirectSettings(w, r, "제외 해제 안 됨 — "+err.Error())
			return
		}
		c.redirectSettings(w, r, symbol+" 제외 해제됨 — 편입 후보로 돌아갈 뿐이고, 실제 편입 여부는 "+
			"전역 설정과 편입 지정이 정한다. "+effectNotice(c.engineRunning()))
		return
	}

	if !next.Excludes(symbol) {
		next.ExcludeSymbols = append(append([]string(nil), next.ExcludeSymbols...), symbol)
		sort.Strings(next.ExcludeSymbols)
	}
	// Exclusion wins over designation in the engine, so a block carrying both is a
	// block whose include list has no effect. The console does not write one. This
	// direction is the conservative one — the engine ends up doing less — so it
	// happens without asking; the answer says which half went (design D3).
	dropped := ""
	if next.Included(symbol) {
		next.IncludeSymbols = withoutSymbol(next.IncludeSymbols, symbol)
		dropped = " 같은 심볼의 편입 지정도 함께 해제됐다 — 제외가 편입보다 우선하므로 남겨두면 " +
			"아무 효과 없는 지정이 된다."
	}
	// No default stop fraction is filled in here, and that asymmetry with the
	// designation path is deliberate. Adoption.validate() demands a fraction when
	// the INCLUDE list is non-empty; the exclude list is not in that condition. A
	// fraction written as a side effect of "leave this holding alone" would be a
	// number the operator never chose, sitting in the file for whenever adoption
	// is turned on later (design D4).
	if err := c.opts.Settings.Save(next); err != nil {
		c.redirectSettings(w, r, "제외 안 됨 — "+err.Error())
		return
	}
	// Present tense would be a lie while the engine is up: it runs on the snapshot
	// it started with, so it can still adopt this symbol on its very next cycle —
	// and once adopted, the exclusion is inert for that position forever. The
	// sentence therefore records the setting and lets effectNotice say when it
	// bites (adversarial review A1).
	c.redirectSettings(w, r, symbol+" 편입 제외 기록됨 — 이미 편입된 포지션이 있다면 그 포지션의 "+
		"손절·익절에는 영향이 없다."+dropped+" "+effectNotice(c.engineRunning()))
}

// withoutSymbol returns the list with one already-normalised symbol removed.
func withoutSymbol(list []string, symbol string) []string {
	kept := make([]string, 0, len(list))
	for _, s := range list {
		if s != symbol {
			kept = append(kept, s)
		}
	}
	return kept
}

// effectNotice is the fixed honesty line about when a save takes effect.
func effectNotice(running bool) string {
	if running {
		return "가동 중인 엔진은 기동 시점 설정으로 계속 동작한다 — 반영하려면 엔진을 재시작해야 한다."
	}
	return "다음 엔진 기동부터 반영된다."
}

// redirectSettings sends a save's answer back to the card that caused it.
//
// The tab and the card come from the route that is answering, not from the
// notice text: the route knows which form it is and the sentence does not, and a
// sentence-matching rule would go quietly wrong the first time one was reworded.
func (c *Console) redirectSettings(w http.ResponseWriter, r *http.Request, notice string) {
	origin := settingsOriginFor(r.URL.Path)
	target := origin.Tab + "?notice=" + urlQueryEscape(notice)
	if origin.Form != "" {
		target += "&form=" + urlQueryEscape(origin.Form) + "#" + origin.Form
	}
	http.Redirect(w, r, target, http.StatusSeeOther)
}

// splitSymbols parses a comma/whitespace separated list; normalisation proper
// (upper-case, dedupe) belongs to the config layer.
func splitSymbols(raw string) []string {
	fields := strings.FieldsFunc(raw, func(r rune) bool {
		return r == ',' || r == ' ' || r == '\t' || r == '\n' || r == '\r'
	})
	out := make([]string, 0, len(fields))
	for _, f := range fields {
		if f = strings.TrimSpace(f); f != "" {
			out = append(out, f)
		}
	}
	return out
}
