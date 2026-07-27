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
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/JungHoonGhae/tossinvest-cli/internal/config"
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
	Nav    string
	CSRF   string
	Notice string
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

// StopPctSlider is the slider position: the saved fraction, or the default.
func (p settingsPage) StopPctSlider() string {
	pct := p.Block.DefaultStopPct
	if pct < 0.02 || pct >= 1 {
		pct = defaultStopPct
	}
	return strconv.FormatFloat(pct, 'f', -1, 64)
}

// StopPctPercent renders the slider position as a percentage label.
func (p settingsPage) StopPctPercent() string {
	f, _ := strconv.ParseFloat(p.StopPctSlider(), 64)
	return strconv.FormatFloat(f*100, 'f', -1, 64) + "%"
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

func (c *Console) handleSettings(w http.ResponseWriter, r *http.Request) {
	page := settingsPage{
		Nav:           "settings",
		CSRF:          c.csrf,
		Notice:        r.URL.Query().Get("notice"),
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
	c.render(w, "settings", page)
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
	if raw := strings.TrimSpace(r.PostFormValue("default_stop_pct")); raw != "" {
		pct, err := strconv.ParseFloat(raw, 64)
		if err != nil {
			c.redirectSettings(w, r, "저장 안 됨 — 손절폭이 숫자가 아니다: "+raw)
			return
		}
		next.DefaultStopPct = pct
	}

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
		kept := make([]string, 0, len(next.IncludeSymbols))
		for _, s := range next.IncludeSymbols {
			if s != symbol {
				kept = append(kept, s)
			}
		}
		next.IncludeSymbols = kept
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
	c.redirectSettings(w, r, symbol+" 편입 예약됨 — 상시 규칙이다(청산 후 재매수도 다음 대사 주기에 재편입)."+
		usedDefault+" "+effectNotice(c.engineRunning()))
}

// effectNotice is the fixed honesty line about when a save takes effect.
func effectNotice(running bool) string {
	if running {
		return "가동 중인 엔진은 기동 시점 설정으로 계속 동작한다 — 반영하려면 엔진을 재시작해야 한다."
	}
	return "다음 엔진 기동부터 반영된다."
}

func (c *Console) redirectSettings(w http.ResponseWriter, r *http.Request, notice string) {
	http.Redirect(w, r, "/settings?notice="+urlQueryEscape(notice), http.StatusSeeOther)
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
