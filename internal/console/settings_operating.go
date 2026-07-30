package console

// settings_operating.go is the two toggles that stood between a configured
// machine and a running engine (change console-owns-the-operating-toggles).
//
// # What this screen is answering
//
// Until now the console could edit the adoption block and the Guardian ceilings,
// and the last two things — the trading policy and the automation gate — were
// `config.json` hand-edits. The stated reason was §0.7, "운영 토글 flip과 live
// 검증은 사람이 직접 승인한다". That sentence names the *approver*, not the
// place, and a person pressing a button in their own loopback console is that
// person directly approving.
//
// What the exclusion actually bought was the opposite of safety. The hand-edit
// path checks nothing: not whether the five ceilings are set, not whether the
// JSON still parses, not whether the operator knows that turning the gate on
// starts an irreversible adoption pass — and it writes no audit line. The most
// consequential toggle in the system was the one with the least protection
// around it.
//
// # The two things this screen must not become
//
//   - A second interlock. The pre-flight below judges only what the *file*
//     answers, says so, and never claims a start is guaranteed. The engine's
//     interlock stays the single judge; a screen that reproduced it would be a
//     second implementation of one rule.
//   - A friction ritual. No typed confirmation, no two-step arm-and-fire (user
//     instruction, 2026-07-27). What the operator gets before pressing is
//     *sentences*: what starts, what cannot be undone, what is still refused.

import (
	"net/http"
	"strings"

	"github.com/JungHoonGhae/tossinvest-cli/internal/config"
)

// TradingPolicySettings is the console's write surface for the trading block.
//
// Exactly two methods, and Save takes config.TradingPolicy — four fields, no
// amend, no conditional, no fractional. Those three are unreachable from this
// screen for the same reason `enabled` was unreachable from the limit screen:
// the type carries no bytes for them.
type TradingPolicySettings interface {
	// Load returns the block as written in the file.
	Load() (config.Trading, error)
	// Save writes the four toggles and nothing else.
	Save(config.TradingPolicy) error
}

// GateSwitch is the console's write surface for engine.automation_gate.enabled,
// and for that one key.
//
// It is a separate seam from LimitSettings on purpose (design D1). Keeping the
// switch off the limit seam is what makes "editing a ceiling cannot turn the
// gate on" a property of the types rather than a promise in a comment — the
// limit path still emits no bytes for `enabled`, exactly as before.
type GateSwitch interface {
	// Save writes the switch and nothing else. Reading is LimitSettings.Load's
	// job: it already returns the whole block including `enabled`, and two
	// readers of one key is how a screen ends up disagreeing with itself.
	Save(on bool) error
}

// preflightClause is one interlock clause as the screen can judge it.
type preflightClause struct {
	// Name is the clause as the engine's own enumeration spells it, so the
	// screen and the refusal read the same.
	Name string
	// OK is the verdict. Meaningless when Judgeable is false.
	OK bool
	// Detail says what is missing, in words the operator can act on.
	Detail string
	// Judgeable is false for the clauses only a constructed process can answer.
	Judgeable bool
}

// GatePreflight is what the screen can tell the operator before they press.
//
// It is deliberately partial, and the partiality is rendered rather than hidden:
// a screen that showed three green ticks and stayed silent about the three it
// cannot see would be read as "this will start".
func (p settingsPage) GatePreflight() []preflightClause {
	out := []preflightClause{}

	limits := p.Gate.Limits()
	if err := limits.Validate(); err != nil {
		out = append(out, preflightClause{
			Name: "1. 한도 다섯 개와 통화", Judgeable: true, Detail: err.Error(),
		})
	} else {
		out = append(out, preflightClause{
			Name: "1. 한도 다섯 개와 통화", OK: true, Judgeable: true,
			Detail: limitSummary(limits),
		})
	}

	policy := config.TradingPolicyOf(p.Trading)
	switch {
	case policy.Complete():
		out = append(out, preflightClause{
			Name: "3. 거래 정책", OK: true, Judgeable: true,
			Detail: "청산에 필요한 네 토글이 모두 켜져 있다",
		})
	default:
		out = append(out, preflightClause{
			Name: "3. 거래 정책", Judgeable: true,
			Detail: strings.Join(policy.Missing(), ", ") + " 가 꺼져 있다",
		})
	}

	out = append(out, preflightClause{
		Name: "2·6. 능력 증명서", Judgeable: false,
		Detail: "파일 경로·만료·계좌 일치는 엔진이 기동 시점에 판정한다",
	})
	out = append(out, preflightClause{
		Name: "4·5·7·8. Guardian 주입 · 게이트웨이 · 전송의 키 운반", Judgeable: false,
		Detail: "구성된 프로세스에서만 알 수 있다 — 이 화면은 판정하지 못한다",
	})
	return out
}

// GateBlockers is the judgeable clauses that would refuse, for the one-line
// summary above the button.
func (p settingsPage) GateBlockers() []string {
	var out []string
	for _, c := range p.GatePreflight() {
		if c.Judgeable && !c.OK {
			out = append(out, c.Name)
		}
	}
	return out
}

// TradingRows renders the four toggles the exit path uses.
//
// The three the engine never reaches are absent rather than shown-and-disabled:
// a toggle on screen is a question, and "should I turn this on?" has no useful
// answer for a capability nothing calls.
func (p settingsPage) TradingRows() []toggleRow {
	return []toggleRow{
		{"place", "주문 제출", p.Trading.Place,
			"매도도 제출이다 — 이것이 꺼져 있으면 손절 주문 자체가 나가지 않는다"},
		{"sell", "매도", p.Trading.Sell, "청산의 방향"},
		{"cancel", "취소", p.Trading.Cancel,
			"exit 관측 루프가 자기 발의를 취소한다 — 없으면 교체가 필요한 청산이 멈춘다"},
		{"allow_live_order_actions", "실주문 실행", p.Trading.AllowLiveOrderActions,
			"모든 mutation의 마스터 스위치"},
	}
}

type toggleRow struct {
	Key   string
	Label string
	On    bool
	Why   string
}

// AdoptionExcludes renders the current exclusion list for the gate section's
// warning: the operator about to start an irreversible adoption pass should see
// what is exempt from it without leaving the screen.
func (p settingsPage) AdoptionExcludes() string {
	if len(p.Block.ExcludeSymbols) == 0 {
		return "없음 — 무기록 보유 전부가 편입 대상이다"
	}
	return strings.Join(p.Block.ExcludeSymbols, ", ")
}

// AdoptionArmed reports that starting the engine will adopt on the first cycle.
func (p settingsPage) AdoptionArmed() bool {
	return p.Block.Enabled || len(p.Block.IncludeSymbols) > 0
}

// handleSettingsTrading saves the four toggles.
func (c *Console) handleSettingsTrading(w http.ResponseWriter, r *http.Request) {
	if c.opts.TradingPolicy == nil {
		c.refuse(w, http.StatusNotImplemented, "거래 정책 저장이 배선되지 않았다",
			"이 빌드의 콘솔에는 거래 정책 저장 seam이 주입되지 않았다.")
		return
	}
	next := config.TradingPolicy{
		Place:                 r.PostFormValue("place") != "",
		Sell:                  r.PostFormValue("sell") != "",
		Cancel:                r.PostFormValue("cancel") != "",
		AllowLiveOrderActions: r.PostFormValue("allow_live_order_actions") != "",
	}
	if err := c.opts.TradingPolicy.Save(next); err != nil {
		c.redirectSettings(w, r, "저장 안 됨 — "+err.Error())
		return
	}
	notice := "거래 정책 기록됨"
	if missing := next.Missing(); len(missing) > 0 {
		// Saved, and said out loud: a partial policy is a legitimate thing to
		// record on the way to a full one, but the operator should not discover
		// at engine start that it was partial.
		notice += " — 아직 꺼진 것: " + strings.Join(missing, ", ") +
			". 이 상태로는 게이트를 켜도 기동 인터록 3절이 거부한다"
	}
	c.redirectSettings(w, r, notice+". "+effectNotice(c.engineRunning()))
}

// handleSettingsGate flips the automation gate.
//
// It saves whatever the checkbox says, including ON with a pre-flight that
// would refuse. That is not laxity: the engine enumerates what is unmet and
// refuses to start, which is a better answer than a console that silently
// declines to record what the operator chose. The screen showed them the
// blockers before the button.
func (c *Console) handleSettingsGate(w http.ResponseWriter, r *http.Request) {
	if c.opts.Gate == nil {
		c.refuse(w, http.StatusNotImplemented, "게이트 저장이 배선되지 않았다",
			"이 빌드의 콘솔에는 자동화 게이트 seam이 주입되지 않았다.")
		return
	}
	on := r.PostFormValue("enabled") != ""
	if err := c.opts.Gate.Save(on); err != nil {
		c.redirectSettings(w, r, "저장 안 됨 — "+err.Error())
		return
	}
	if !on {
		c.redirectSettings(w, r, "자동화 게이트 OFF 기록됨. "+effectNotice(c.engineRunning()))
		return
	}
	c.redirectSettings(w, r, "자동화 게이트 ON 기록됨 — 다음 엔진 기동에서 대사·exit 관측·"+
		"체결 감지 루프가 시작된다. 편입은 첫 대사 주기에 일어나고 되돌릴 수 없다. "+effectNotice(c.engineRunning()))
}
