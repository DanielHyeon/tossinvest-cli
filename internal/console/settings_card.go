package console

// settings_card.go is the one shape every settings form takes (change
// a055-console-settings-cadence §4).
//
// Four things, in the same order, on every card:
//
//	① 현재값 — and the value being proposed, wherever the server knows one
//	② 적용 후 — what changes and when it takes effect
//	③ 사유    — a NAMED reason wherever the save surface is missing or risky
//	④ 결과    — beside the form that produced it, not at the top of the page
//
// # The reasons are not new judgements
//
// Every one of them already existed on the old screen, written as a paragraph:
// `not .Wired`, `.LoadErr`, `.Verdict`, `.LimitsUnset`, `.NeedsStopPctCorrection`,
// `.GateBlockers`, `.EngineRunning`. This file turns those paragraphs into a
// structure with a name on it. Nothing here decides anything the screen did not
// already decide, and nothing here refuses a save — the server's existing
// validation owns that, and a screen that re-implemented it would be a second
// rule to keep in step with the first.
//
// # Why "the save surface is missing" needs a reason and not just an absence
//
// This console does not disable a form whose seam is unwired; it declines to
// render it. A check that looked for `disabled` would find nothing and pass,
// which makes the requirement unreachable. So the contract is stated the other
// way round: a card either renders a submit control, or renders a named reason
// where one would be. settings_card_test.go walks every card on every tab and
// fails on a card that has neither.

import (
	"strings"

	"github.com/JungHoonGhae/tossinvest-cli/internal/config"
)

// saveBlock is one named reason a save is impossible or risky.
type saveBlock struct {
	// Name is the short label the operator reads in place of, or beside, the
	// button. It is a name and not a sentence: the sentence is Detail.
	Name   string
	Detail string
}

// cardGuard is a card's reasons, split by whether they stop the save.
type cardGuard struct {
	// Blocks stop the save. When there is one, the card renders the reason where
	// the submit control would be.
	Blocks []saveBlock
	// Cautions do not stop it. The save surface renders, with the reason beside
	// it — the engine, not this screen, has the last word.
	Cautions []saveBlock
}

// Blocked reports that this card renders a reason instead of a submit control.
func (g cardGuard) Blocked() bool { return len(g.Blocks) > 0 }

// engineRunningCaution is reason ⑦, and it is the same sentence on every card
// because it is the same fact.
func (p settingsPage) engineRunningCaution() []saveBlock {
	if !p.EngineRunning {
		return nil
	}
	return []saveBlock{{
		Name:   "엔진 실행 중",
		Detail: "저장은 즉시 파일에 기록되지만 가동 중인 엔진은 기동 시점 설정으로 계속 동작한다 — 반영하려면 재시작한다.",
	}}
}

// AdoptionGuard is the adoption card's reasons.
func (p settingsPage) AdoptionGuard() cardGuard {
	var g cardGuard
	if !p.Wired {
		g.Blocks = append(g.Blocks, saveBlock{"저장 미배선",
			"저장이 배선되지 않았다 — 이 빌드의 콘솔에는 편입 설정 저장 seam이 주입되지 않았다. " +
				"표시는 되지만 저장할 수 없다. config.json을 직접 편집한다."})
	}
	if p.LoadErr != "" {
		g.Blocks = append(g.Blocks, saveBlock{"설정 파일 읽기 실패", p.LoadErr})
	}
	// A caution and not a block, which is the opposite of what design §3's table
	// reads like at a glance. That column says whether the CURRENT value can be
	// saved, and the server does refuse it — but withholding the form would leave
	// the operator holding an off-grid number with nothing on the screen able to
	// fix it. The one correction path runs through this form.
	if p.NeedsStopPctCorrection() {
		g.Cautions = append(g.Cautions, saveBlock{"손절폭 단위 불일치",
			"현재 파일의 값은 새 입력 단위와 맞지 않는다. 저장하려면 2%~20% 범위의 0.5% 단위로 다시 선택한다 — " +
				"그대로 제출하면 침묵 반올림 없이 거부된다."})
	}
	if p.Verdict != "" {
		g.Cautions = append(g.Cautions, saveBlock{"엔진이 거부할 블록", p.Verdict})
	}
	g.Cautions = append(g.Cautions, p.engineRunningCaution()...)
	return g
}

// LimitGuard is the Guardian-limit card's reasons.
func (p settingsPage) LimitGuard() cardGuard {
	var g cardGuard
	if !p.LimitsWired {
		g.Blocks = append(g.Blocks, saveBlock{"한도 저장 미배선",
			"한도 저장이 배선되지 않았다 — 이 빌드의 콘솔에는 Guardian 한도 저장 seam이 주입되지 않았다. " +
				"표시는 되지만 저장할 수 없다. config.json을 직접 편집한다."})
	}
	if p.LimitsLoadErr != "" {
		g.Blocks = append(g.Blocks, saveBlock{"한도 읽기 실패", p.LimitsLoadErr})
	}
	switch {
	case p.LimitsUnset():
		g.Cautions = append(g.Cautions, saveBlock{"한도 미설정",
			"다섯 값이 비어 있다. 이 상태로는 게이트를 켜도 엔진이 기동하지 않는다."})
	case p.LimitsPartlyConfigured():
		g.Cautions = append(g.Cautions, saveBlock{"한도 부분 설정",
			"기동 인터록은 다섯 값이 전부 양수·유한하고 통화가 있어야 통과시킨다: " + p.LimitVerdict()})
	}
	g.Cautions = append(g.Cautions, p.engineRunningCaution()...)
	return g
}

// TradingGuard is the trading-policy card's reasons.
func (p settingsPage) TradingGuard() cardGuard {
	var g cardGuard
	if !p.TradingWired {
		g.Blocks = append(g.Blocks, saveBlock{"거래 정책 저장 미배선",
			"거래 정책 저장이 배선되지 않았다 — 표시는 되지만 저장할 수 없다."})
	}
	if p.TradingLoadErr != "" {
		g.Blocks = append(g.Blocks, saveBlock{"거래 정책 읽기 실패", p.TradingLoadErr})
	}
	if missing := config.TradingPolicyOf(p.Trading).Missing(); len(missing) > 0 {
		g.Cautions = append(g.Cautions, saveBlock{"기동 인터록 3절 미충족",
			strings.Join(missing, ", ") + " 가 꺼져 있다. 이 상태로는 게이트를 켜도 기동이 거부된다."})
	}
	g.Cautions = append(g.Cautions, p.engineRunningCaution()...)
	return g
}

// GateGuard is the automation-gate card's reasons.
func (p settingsPage) GateGuard() cardGuard {
	var g cardGuard
	if !p.GateWired {
		g.Blocks = append(g.Blocks, saveBlock{"게이트 저장 미배선",
			"게이트 저장이 배선되지 않았다 — 표시는 되지만 저장할 수 없다."})
	}
	if blockers := p.GateBlockers(); len(blockers) > 0 {
		// Recorded, not refused. The engine enumerates what is unmet and declines
		// to start, which is a better answer than a console that silently declines
		// to record what the operator chose.
		g.Cautions = append(g.Cautions, saveBlock{"지금 켜면 기동이 거부된다",
			"미충족: " + strings.Join(blockers, ", ") + ". 기록은 되지만 엔진은 뜨지 않는다."})
	}
	g.Cautions = append(g.Cautions, p.engineRunningCaution()...)
	return g
}

// AutostartGuard is the engine-autostart card's reasons.
func (p settingsPage) AutostartGuard() cardGuard {
	var g cardGuard
	if !p.AutostartWired {
		g.Blocks = append(g.Blocks, saveBlock{"자동 시작 저장 미배선",
			"엔진 자동 시작 저장이 배선되지 않았다 — 표시는 OFF이지만 저장하거나 기동할 수 없다."})
	}
	if p.AutostartLoadErr != "" {
		g.Blocks = append(g.Blocks, saveBlock{"자동 시작 설정 읽기 실패", p.AutostartLoadErr})
	}
	if blockers := p.GateBlockers(); len(blockers) > 0 {
		g.Cautions = append(g.Cautions, saveBlock{"지금 켜면 기동이 거부된다",
			"미충족: " + strings.Join(blockers, ", ") + ". ON 저장 직후의 기동 시도도 같은 인터록에 걸린다."})
	}
	return g
}

// UpdateGuard is the system-update card's reasons.
func (p settingsPage) UpdateGuard() cardGuard {
	var g cardGuard
	if !p.UpdateWired {
		g.Blocks = append(g.Blocks, saveBlock{"시스템 업데이트 미배선",
			"시스템 업데이트가 배선되지 않았다 — 이 빌드는 설정 화면에서 실행 파일을 교체할 수 없다."})
	} else if !p.Update.Installable {
		g.Blocks = append(g.Blocks, saveBlock{"설치할 수 없음", p.Update.Reason})
	}
	if p.UpdateWired && p.Update.Installable && !p.ReleaseVerified {
		g.Cautions = append(g.Cautions, saveBlock{"출처 확인 안 됨",
			"이 프로세스가 서명 릴리스를 검증해 만든 candidate라는 증거가 없다."})
	}
	return g
}

// DownloadGuard is the signed-release download card's reasons.
func (p settingsPage) DownloadGuard() cardGuard {
	var g cardGuard
	if !p.ReleaseDownloadWired {
		g.Blocks = append(g.Blocks, saveBlock{"서명 릴리스 다운로드 미배선",
			"서명 릴리스 다운로드가 배선되지 않았다 — 현재 실행 파일은 로컬 candidate만 검토할 수 있다."})
	}
	return g
}

// --- ④ the answer, beside the form that caused it ---------------------------------

// NoticeFor is the save answer belonging to one card, or "".
func (p settingsPage) NoticeFor(card string) string {
	if p.Notice == "" || p.NoticeForm != card {
		return ""
	}
	return p.Notice
}

// OrphanNotice is an answer whose card is not on this tab.
//
// It should not normally happen — every settings POST names its card — but a
// notice that names nothing is still an answer somebody is waiting to read, and
// dropping it silently would be the worst of the three options.
func (p settingsPage) OrphanNotice() string {
	if p.Notice == "" {
		return ""
	}
	if origin, ok := settingsCardTab[p.NoticeForm]; ok && origin == p.Tab {
		return ""
	}
	return p.Notice
}

// settingsCardTab says which tab renders each card, so a notice can tell whether
// its card is on this screen.
var settingsCardTab = map[string]string{
	"adoption":      tabStanding,
	"gate":          tabStanding,
	"autostart":     tabStanding,
	"limits":        tabDaily,
	"trading":       tabDaily,
	"system-update": tabTools,
	// Alert delivery is set up once and rarely reopened, and the first thing an
	// operator wants on reopening it is the state — which is the 도구 tab's whole
	// contract (a065 design D6).
	"notifications": tabTools,
}

// --- ② the preview, for the one card where the server knows both sides -------------

// limitChange is one ceiling's before and after.
type limitChange struct {
	Label string
	From  string
	To    string
	// Axis is 강화, 완화 or 변경 없음. It is empty when the currency changed,
	// because then the two numbers are not on the same scale at all.
	Axis string
}

// tierPreview is what applying one preset would do to the file as it stands.
//
// # Why the currency is a third axis and not a direction
//
// The gate has one limit currency and risk's chain refuses an intent priced in
// another, so KRW 500,000 → USD 3,000 is not a tightening — the number went down
// and what actually happened is that the domestic market's automatic entry
// closed and the US market's opened. Reading that as 강화 would be false in the
// one direction that matters, so a currency change suppresses the per-field
// direction entirely and says the consequence instead.
type tierPreview struct {
	Rows []limitChange
	// CurrencyChange reports the third axis. From is empty on a first setting.
	CurrencyChange bool
	CurrencyFrom   string
	CurrencyTo     string
	// Consequence is which market this currency closes, in the sentence the
	// console already uses after an apply.
	Consequence string
	// FirstTime reports that nothing was set before, so nothing is being widened
	// or narrowed.
	FirstTime bool
	// Effect is when this takes hold.
	Effect string
}

// previewAgainst is what this tier would change, measured against the file.
//
// It takes the limits and the "anything set at all" answer rather than the gate
// block, and that is not fussiness: engineproc_test.go allows the spelling
// "AutomationGate" in exactly three files, because naming the config block is
// how a file starts deciding things about the gate. This file has no business
// with the switch, so it never receives it.
func (t tierCard) previewAgainst(current config.GuardianLimits, anySet, engineRunning bool) tierPreview {
	next := t.Limits
	preview := tierPreview{
		CurrencyFrom: strings.TrimSpace(current.Currency),
		CurrencyTo:   strings.TrimSpace(next.Currency),
		Consequence:  currencyConsequence(next.Currency),
		Effect:       effectNotice(engineRunning),
		FirstTime:    !anySet,
	}
	preview.CurrencyChange = !preview.FirstTime &&
		!strings.EqualFold(preview.CurrencyFrom, preview.CurrencyTo)

	comparable := !preview.CurrencyChange && !preview.FirstTime
	rows := []struct {
		label    string
		from, to float64
		ratio    bool
	}{
		{"1회 주문 수량 상한", current.MaxOrderQuantity, next.MaxOrderQuantity, false},
		{"1회 주문 금액 상한", current.MaxOrderNotional, next.MaxOrderNotional, false},
		{"계좌 개방 노출 상한", current.MaxTotalExposure, next.MaxTotalExposure, false},
		{"일일 손실 한도(금액)", current.MaxDailyLossAmount, next.MaxDailyLossAmount, false},
		{"일일 손실 한도(비율)", current.MaxDailyLossRatio, next.MaxDailyLossRatio, true},
	}
	for _, row := range rows {
		change := limitChange{Label: row.label}
		if row.ratio {
			change.From, change.To = limitRatioText(row.from), limitRatioText(row.to)
		} else {
			change.From, change.To = limitValueText(row.from), limitValueText(row.to)
		}
		if comparable {
			change.Axis = limitDirection(row.from, row.to)
		}
		preview.Rows = append(preview.Rows, change)
	}
	return preview
}

// limitDirection names which way one ceiling moved, within one currency.
//
// A narrower ceiling is 강화 and a wider one is 완화. Neither is refused here:
// the screen says the direction and the server's existing validation — the
// registered tier ceiling — decides what is allowed. The safety invariant asks
// for conservative changes to be the easy ones, not for the console to grow a
// second opinion about them.
func limitDirection(from, to float64) string {
	switch {
	case from == to:
		return "변경 없음"
	case to < from:
		return "강화"
	default:
		return "완화"
	}
}

// --- ③ the tab header's operating state -------------------------------------------

// stateChip is one fact in a tab's header.
type stateChip struct {
	Label string
	Value string
	// Tone is "", "ok", "warn". Never the only carrier: the word is always there.
	Tone string
	// Href points at the tab that owns the control, when this tab only displays
	// the value. Empty when this tab owns it.
	Href string
}

// OperatingState is the three facts a person changing a setting needs in front
// of them: is the gate on, is the engine up, are the limits set.
//
// Each chip links to the tab that owns the switch when this tab does not. That
// link is the reason the "one control, one tab" rule does not blind the operator:
// the value is readable everywhere, the control exists once.
func (p settingsPage) OperatingState() []stateChip {
	gate := stateChip{Label: "자동화 게이트", Value: "OFF", Tone: "warn"}
	if p.Gate.Enabled {
		gate.Value, gate.Tone = "ON", "ok"
	}
	if p.Tab != tabStanding {
		gate.Href = pathSettingsStanding + "#gate"
	}

	engine := stateChip{Label: "엔진", Value: "정지"}
	if p.EngineRunning {
		engine.Value, engine.Tone = "실행 중", "ok"
	}

	limits := stateChip{Label: "Guardian 한도", Value: "미설정", Tone: "warn"}
	switch {
	case p.LimitsConfigured():
		limits.Value, limits.Tone = "설정됨", "ok"
	case p.LimitsPartlyConfigured():
		limits.Value, limits.Tone = "부분 설정", "warn"
	}
	if p.Tab != tabDaily {
		limits.Href = pathSettingsDaily + "#limits"
	}

	return []stateChip{gate, engine, limits}
}

// CurrentTab is the record for this render, for the header.
func (p settingsPage) CurrentTab() settingsTab { return settingsTabByKey(p.Tab) }

// LimitCurrencyConsequence is what the currently recorded currency costs. It is
// reason ⑦ of the never-folded list, so it renders as a notice and not as muted
// prose.
func (p settingsPage) LimitCurrencyConsequence() string {
	if c := strings.TrimSpace(p.Gate.LimitCurrency); c != "" {
		return currencyConsequence(c)
	}
	return "한도 통화가 아직 없다 — 통화를 기록하면 Guardian 체인은 다른 통화의 자동 진입을 거부한다."
}

// EffectNotice is when a save on this screen takes hold.
func (p settingsPage) EffectNotice() string { return effectNotice(p.EngineRunning) }
