package console

// settings_tabs.go splits the settings screen by the reversibility and the
// frequency of the change, not by the name of the feature (change
// a055-console-settings-cadence).
//
// The old screen was one page with four `<section>`s in the order the changes
// that built them happened to land: adoption, Guardian limits, operating, system
// update. Nothing about that order told the operator which of them they could
// undo. The four tabs here answer that before anything else:
//
//	상시   irreversible, touched less than weekly, approved by a person
//	당일   reversible, adjusted daily — the default landing, because it is the
//	       one opened most and the fewest clicks belong to the most-used screen
//	전략   the rules themselves, as entry points into the screens that own them
//	도구   diagnostics, rarely
//
// # Entry points, not absorption
//
// 전략 and 도구 link out. /optimization?category=… is a canonical deep link that
// a050 fixed and that other requirements reference by path; moving those screens
// under /settings would mean MODIFYing every one of them on top of an unarchived
// delta stack. Each link carries the target's current one-line reading instead,
// taken from the same call the target screen makes — a tab of bare links is one
// more empty screen.

import (
	"net/http"
	"strings"

	"github.com/JungHoonGhae/tossinvest-cli/internal/positionpolicy"
)

// The settings paths. Registered as literals in console.go — registeredRoutes
// reads the route table out of the source and refuses a non-literal path — so
// these constants are the second spelling, and screen_paths_test.go is what
// keeps the two from drifting.
const (
	pathSettings         = "/settings"
	pathSettingsStanding = "/settings/standing"
	pathSettingsDaily    = "/settings/daily"
	pathSettingsStrategy = "/settings/strategy"
	pathSettingsTools    = "/settings/tools"
)

// The tab keys. They are the handler's argument and the template's switch, and
// they are spelled as string literals at the registration because opaqueHandler
// refuses a handler argument that is an identifier: `c.handleSettingsTab(tabDaily)`
// would make the whole registration unreadable to every route guard at once.
const (
	tabStanding = "standing"
	tabDaily    = "daily"
	tabStrategy = "strategy"
	tabTools    = "tools"
)

// settingsTab is one tab as the bar renders it.
type settingsTab struct {
	Key   string
	Path  string
	Label string
	// Lead is the one line under the heading: what belongs here and why it is
	// separate from the others.
	Lead string
}

// settingsTabs is the bar, in the order the classification puts them: the two
// that hold controls first, the two that link out after.
var settingsTabs = []settingsTab{
	{tabStanding, pathSettingsStanding, "상시",
		"비가역 · 주 1회 미만 · 사람이 승인한다. 한 번 켜면 되돌릴 수 없는 것들이다."},
	{tabDaily, pathSettingsDaily, "당일",
		"가역 · 일 단위로 조절한다. 좁히는 쪽은 언제든 되돌릴 수 있다."},
	{tabStrategy, pathSettingsStrategy, "전략",
		"규칙 자체. 각 화면이 자기 경로를 그대로 갖고 있고 여기는 진입점이다."},
	{tabTools, pathSettingsTools, "도구",
		"진단과 유지보수. 드물게 열고, 열었을 때 상태부터 읽는다."},
}

// settingsTabByKey is the tab record, or the default.
func settingsTabByKey(key string) settingsTab {
	for _, tab := range settingsTabs {
		if tab.Key == key {
			return tab
		}
	}
	return settingsTabs[1]
}

// entryPoint is a link out of 전략 or 도구, with the target's current reading.
type entryPoint struct {
	Label string
	Href  string
	// Summary is the target screen's own current line. It is moved, never
	// recomputed: the value comes from the same call the target makes, so the two
	// cannot disagree.
	Summary string
	// Note says what the operator will be able to do there, so a link is a
	// decision rather than a guess.
	Note string
}

// --- the tab handler -----------------------------------------------------------

// handleSettingsTab renders one tab.
//
// The four routes share every read: which seams are wired, what the file says,
// whether the engine is up. Only the markup differs, so the reads happen once
// here and the template picks what its tab owns.
func (c *Console) handleSettingsTab(tab string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		page := c.settingsView(r)
		page.Tab = settingsTabByKey(tab).Key
		switch page.Tab {
		case tabStrategy:
			page.Entries = c.strategyEntries(r)
		case tabTools:
			page.Entries = c.toolEntries()
		}
		c.render(w, "settings", page)
	}
}

// handleSettings is the old path. It is a redirect now, and not a 404: the
// navigation pointed here for two changes, the terminal has printed it, and the
// spec requires the existing link to keep working.
//
// The `#adoption` fragment never reaches a server, so the redirect cannot read
// it. The browser carries it across the 303 by itself, which lands it on
// /settings/daily#adoption — a tab that does not have that section. Two things
// answer that: the daily tab renders an `#adoption` anchor that points at the
// standing tab, and `?section=adoption` (which a server CAN read) redirects
// straight there.
func (c *Console) handleSettings(w http.ResponseWriter, r *http.Request) {
	target := pathSettingsDaily
	if section := strings.TrimSpace(r.URL.Query().Get("section")); section != "" {
		if tab, anchor, ok := settingsSection(section); ok {
			target = tab + "#" + anchor
		}
	}
	if notice := r.URL.Query().Get("notice"); notice != "" {
		// A notice that arrived on the old path is still an answer somebody is
		// waiting to read.
		target = pathSettingsDaily + "?notice=" + urlQueryEscape(notice)
	}
	http.Redirect(w, r, target, http.StatusSeeOther)
}

// settingsSection maps an old in-page anchor to the tab that owns it now.
func settingsSection(section string) (tab, anchor string, ok bool) {
	switch strings.ToLower(strings.TrimSpace(section)) {
	case "adoption":
		return pathSettingsStanding, "adoption", true
	case "operating", "gate":
		return pathSettingsStanding, "gate", true
	case "guardian-limits", "limits":
		return pathSettingsDaily, "limits", true
	case "system-update":
		return pathSettingsTools, "system-update", true
	}
	return "", "", false
}

// --- where a save's answer goes back to ------------------------------------------

// settingsOrigin names the tab and the card that owns one POST route.
type settingsOrigin struct {
	Tab  string
	Form string
}

// settingsOrigins is every settings POST and the card its answer belongs beside.
//
// The route knows which form it is; the notice text does not. Deriving the card
// from the message would go quietly wrong the first time a sentence was reworded,
// and threading an argument through 45 call sites would let a new handler forget
// it and still compile. settings_cadence_test.go reads the real route table and
// fails if a settings POST is missing from here.
var settingsOrigins = map[string]settingsOrigin{
	"/settings/save":                   {pathSettingsStanding, "adoption"},
	"/settings/include":                {pathSettingsStanding, "adoption"},
	"/settings/exclude":                {pathSettingsStanding, "adoption"},
	"/settings/gate":                   {pathSettingsStanding, "gate"},
	"/settings/autostart":              {pathSettingsStanding, "autostart"},
	"/settings/limits":                 {pathSettingsDaily, "limits"},
	"/settings/limits/preset":          {pathSettingsDaily, "limits"},
	"/settings/trading":                {pathSettingsDaily, "trading"},
	"/settings/system-update/download": {pathSettingsTools, "system-update"},
	"/settings/system-update/install":  {pathSettingsTools, "system-update"},
	"/settings/notifications/on":       {pathSettingsTools, "notifications"},
	"/settings/notifications/test":     {pathSettingsTools, "notifications"},
	"/settings/notifications/off":      {pathSettingsTools, "notifications"},
}

// settingsOriginFor is the tab and card a POST returns to.
//
// The fallback is the default tab with no card, which is what the whole screen
// did before this change: an answer displayed at the top of a settings page is
// worse than one displayed beside its form, and better than one that is lost.
func settingsOriginFor(path string) settingsOrigin {
	if origin, ok := settingsOrigins[path]; ok {
		return origin
	}
	return settingsOrigin{Tab: pathSettingsDaily}
}

// --- the entry points ------------------------------------------------------------

// strategyEntries are the three rule screens, each with the line its own screen
// renders.
func (c *Console) strategyEntries(r *http.Request) []entryPoint {
	out := []entryPoint{{
		Label: "최적화", Href: "/optimization",
		Summary: c.commonPolicySummary(r),
		Note:    "공통 익절 정책과 카테고리별 파라미터. 변경은 preview → apply 두 단계다.",
	}, {
		Label: "종목 정책", Href: "/position-management",
		Summary: c.positionPolicySummary(r),
		Note:    "포지션 하나에만 다른 정책을 걸거나 자동관리를 해제한다.",
	}, {
		Label: "전략 lane", Href: "/strategy-runtime",
		Summary: c.strategyRuntimeSummary(r),
		Note:    "lane·자동 시작·게이트·LIVE 승인의 desired와 effective를 따로 읽는다.",
	}}
	return out
}

// commonPolicySummary is the exit policy the optimization screen would show as
// selected, through the same expression that screen uses.
func (c *Console) commonPolicySummary(r *http.Request) string {
	if c.opts.ExitPolicies == nil && c.opts.Optimization == nil {
		return "미배선 — 이 빌드는 최적화 상태를 읽지 못한다"
	}
	var desired map[string]string
	if c.opts.Optimization != nil {
		if view, err := c.opts.Optimization.Read(r.Context()); err == nil {
			desired = view.Snapshot.Desired
		} else {
			return "읽지 못함 — " + err.Error()
		}
	}
	var legacy string
	if c.opts.ExitPolicies != nil {
		value, err := c.opts.ExitPolicies.Load()
		if err != nil {
			return "읽지 못함 — " + err.Error()
		}
		legacy = value.CommonPolicy
	}
	if selected := selectedCommonPolicy(desired, legacy); selected != "" {
		return "공통 익절 정책 " + selected
	}
	return "공통 익절 정책 미지정"
}

// selectedCommonPolicy is the one expression both this summary and the
// optimization screen resolve the selected policy with.
func selectedCommonPolicy(desired map[string]string, legacy string) string {
	if selected := desired["exit.common-policy"]; selected != "" {
		return selected
	}
	return legacy
}

// positionPolicySummary counts the states the position-policy screen lists.
func (c *Console) positionPolicySummary(r *http.Request) string {
	if c.opts.PositionPolicies == nil {
		return "미배선 — 이 빌드는 종목 정책을 읽지 못한다"
	}
	states, err := c.opts.PositionPolicies.List(r.Context())
	if err != nil {
		return "읽지 못함 — " + err.Error()
	}
	if len(states) == 0 {
		return "정책이 걸린 포지션이 없다"
	}
	managed := 0
	for _, state := range states {
		if state.Status == positionpolicy.StatusManaged {
			managed++
		}
	}
	return "포지션 " + decimalText(float64(len(states))) + "건 · 관리 " +
		decimalText(float64(managed)) + "건"
}

// strategyRuntimeSummary is the lane screen's own entry-capability projection,
// desired and effective, produced by the same projection that screen renders.
func (c *Console) strategyRuntimeSummary(r *http.Request) string {
	if c.opts.StrategyRuntime == nil {
		return "미배선 — 이 빌드는 전략 lane 상태를 읽지 못한다"
	}
	reading, err := c.opts.StrategyRuntime.Read(r.Context())
	if err != nil || !validStrategyRuntimeReading(reading) {
		return "읽지 못함 — 전략 lane 판독이 유효하지 않다"
	}
	var page strategyRuntimePage
	page.project(reading)
	return "진입 능력 desired " + page.EntryDesired.Value + " → effective " + page.EntryEffective.Value
}

// toolEntries are the diagnostic screens. All three read one snapshot, which is
// the same read the verification console and the report make.
func (c *Console) toolEntries() []entryPoint {
	snap := c.snapshot()
	return []entryPoint{{
		Label: "검증 콘솔", Href: pathVerifyConsole,
		Summary: engineSummaryText(snap) + " · " + soakSummaryText(snap),
		Note:    "엔진 기동·정지, 검증 run 승인, 재시작이 있는 화면이다.",
	}, {
		Label: "검증", Href: "/verify",
		Summary: verifySummaryText(snap),
		Note:    "실계좌 검증 단계를 순서대로 측정한다. 승인 창은 그 화면에서 뜬다.",
	}, {
		Label: "리포트", Href: "/report",
		Summary: attestSummaryText(snap),
		Note:    "soak·능력 증명서·검증 기록을 한 장으로 모은다.",
	}}
}

func engineSummaryText(s snapshot) string {
	switch {
	case !s.Engine.Wired:
		return "엔진 미배선"
	case s.Engine.Running:
		return "엔진 실행 중"
	default:
		return "엔진 정지"
	}
}

func soakSummaryText(s snapshot) string {
	switch {
	case !s.Soak.Present:
		return "soak 기록 없음"
	case s.Soak.Ready:
		return "soak " + decimalText(float64(s.Soak.StreakDays)) + "일 연속 · 충족"
	default:
		return "soak " + decimalText(float64(s.Soak.StreakDays)) + "일 연속 · 미충족"
	}
}

func verifySummaryText(s snapshot) string {
	if !s.Verify.Present {
		return s.Market + " 검증 기록 없음"
	}
	return s.Market + " 검증 " + decimalText(float64(s.Verify.Done)) + "/" +
		decimalText(float64(s.Verify.Total)) + " 단계"
}

func attestSummaryText(s snapshot) string {
	switch {
	case !s.Attestation.Present:
		return "능력 증명서 없음"
	case s.Attestation.Usable:
		return "능력 증명서 사용 가능"
	default:
		return "능력 증명서 사용 불가"
	}
}
