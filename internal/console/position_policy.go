package console

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/config"
	"github.com/JungHoonGhae/tossinvest-cli/internal/exitpolicy"
	"github.com/JungHoonGhae/tossinvest-cli/internal/positionpolicy"
)

const positionPolicyTokenTTL = 5 * time.Minute

// PositionPolicyCommander is the console's complete position-policy authority.
// The engine implementation owns serialization and the writable journal.
type PositionPolicyCommander interface {
	List(context.Context) ([]positionpolicy.State, error)
	Runtime(context.Context) (positionpolicy.ManagementRuntime, error)
	Preview(context.Context, positionpolicy.Request) (positionpolicy.Preview, error)
	Apply(context.Context, positionpolicy.ApplyRequest) (positionpolicy.State, error)
}

type policyActionView struct {
	Label  string
	Help   string
	Token  string
	Danger bool
}

type policyRowView struct {
	State      positionpolicy.State
	Management positionpolicy.ManagementProjection
	Block      reconcileBlockView
	Actions    []policyActionView
}

type positionPolicyPage struct {
	Nav            string
	CSRF           string
	Notice         string
	Wired          bool
	LoadErr        string
	Descriptor     positionpolicy.ManagementDescriptor
	Desired        adoptionSettingsView
	Effective      adoptionSettingsView
	DesiredVerdict string
	DesiredErr     string
	RuntimeErr     string
	BlockSource    string
	Blocks         []reconcileBlockView
	Rows           []policyRowView
}

func (positionPolicyPage) Refresh() bool { return false }

type adoptionSettingsView struct {
	Known          bool
	Enabled        bool
	DefaultStopPct float64
	IncludeSymbols []string
	ExcludeSymbols []string
}

func (v adoptionSettingsView) EnabledText() string {
	if !v.Known {
		return "알 수 없음"
	}
	if v.Enabled {
		return "ON"
	}
	return "OFF"
}

func (v adoptionSettingsView) StopText() string {
	if !v.Known {
		return "알 수 없음"
	}
	if v.DefaultStopPct == 0 {
		return "미설정"
	}
	return fractionPercentText(v.DefaultStopPct) + "%"
}

func (v adoptionSettingsView) IncludesText() string {
	if !v.Known {
		return "알 수 없음"
	}
	if len(v.IncludeSymbols) == 0 {
		return "없음"
	}
	return strings.Join(v.IncludeSymbols, ", ")
}

func (v adoptionSettingsView) ExcludesText() string {
	if !v.Known {
		return "알 수 없음"
	}
	if len(v.ExcludeSymbols) == 0 {
		return "없음"
	}
	return strings.Join(v.ExcludeSymbols, ", ")
}

func adoptionSettingsDisplay(settings positionpolicy.AdoptionSettings, known bool) adoptionSettingsView {
	return adoptionSettingsView{
		Known: known, Enabled: settings.Enabled, DefaultStopPct: settings.DefaultStopPct,
		IncludeSymbols: append([]string(nil), settings.IncludeSymbols...),
		ExcludeSymbols: append([]string(nil), settings.ExcludeSymbols...),
	}
}

func desiredAdoptionSettingsDisplay(settings config.Adoption) adoptionSettingsView {
	return adoptionSettingsDisplay(positionpolicy.NewAdoptionSettings(settings.Enabled,
		settings.DefaultStopPct, settings.IncludeSymbols, settings.ExcludeSymbols, ""), true)
}

type reconcileBlockView struct {
	Present   bool
	Scope     string
	Target    string
	Reason    string
	Detail    string
	StartedAt string
	Age       string
	Permanent bool
}

func newReconcileBlockView(block positionpolicy.ReconcileBlock, now time.Time) reconcileBlockView {
	target := "계좌 전체"
	switch block.Scope {
	case positionpolicy.ScopeMarket:
		target = strings.ToUpper(strings.TrimSpace(block.Market)) + " 시장"
	case positionpolicy.ScopeSymbol:
		target = strings.ToUpper(strings.TrimSpace(block.Market)) + " · " + strings.ToUpper(strings.TrimSpace(block.Symbol))
	}
	started := "알 수 없음"
	age := "경과 시간 알 수 없음"
	if !block.StartedAt.IsZero() {
		started = block.StartedAt.UTC().Format("2006-01-02 15:04:05Z")
		d := now.Sub(block.StartedAt)
		if d < 0 {
			d = 0
		}
		switch {
		case d < time.Minute:
			age = "1분 미만"
		case d < 24*time.Hour:
			age = fmt.Sprintf("%d시간 %d분", int(d/time.Hour), int(d%time.Hour/time.Minute))
		default:
			age = fmt.Sprintf("%d일 %d시간", int(d/(24*time.Hour)), int(d%(24*time.Hour)/time.Hour))
		}
	}
	return reconcileBlockView{Present: true, Scope: string(block.Scope), Target: target,
		Reason: strings.TrimSpace(block.Reason), Detail: strings.TrimSpace(block.Detail),
		StartedAt: started, Age: age, Permanent: block.Permanent}
}

type positionPolicyPreviewPage struct {
	Nav      string
	CSRF     string
	Preview  positionpolicy.Preview
	Token    string
	Danger   bool
	WaitSecs int
}

func (positionPolicyPreviewPage) Refresh() bool { return false }
func (p positionPolicyPreviewPage) Release() bool {
	return p.Preview.Action == positionpolicy.ActionRelease
}
func (p positionPolicyPreviewPage) Readopt() bool {
	return p.Preview.Action == positionpolicy.ActionReadopt
}

type policyTokenPayload struct {
	Purpose    string                `json:"p"`
	PositionID string                `json:"i"`
	Generation int64                 `json:"g"`
	Version    int64                 `json:"v"`
	Action     positionpolicy.Action `json:"a"`
	PolicyID   string                `json:"d,omitempty"`
	Reason     positionpolicy.Reason `json:"r"`
	NotBefore  int64                 `json:"n"`
	ExpiresAt  int64                 `json:"e"`
}

func (c *Console) handlePositionManagement(w http.ResponseWriter, r *http.Request) {
	page := positionPolicyPage{
		Nav: "position-management", CSRF: c.csrf, Wired: c.opts.PositionPolicies != nil,
		Notice: r.URL.Query().Get("notice"), Descriptor: positionpolicy.Descriptor(),
	}
	if c.opts.Settings == nil {
		page.DesiredErr = "설정 read seam 미배선"
	} else if desired, verdict, err := c.opts.Settings.Load(); err != nil {
		page.DesiredErr = err.Error()
	} else {
		page.Desired = desiredAdoptionSettingsDisplay(desired)
		page.DesiredVerdict = verdict
	}
	if c.opts.PositionPolicies == nil {
		c.render(w, "position-policy", page)
		return
	}
	runtime, runtimeErr := c.opts.PositionPolicies.Runtime(r.Context())
	if runtimeErr != nil {
		page.RuntimeErr = runtimeErr.Error()
	} else {
		page.Effective = adoptionSettingsDisplay(runtime.Effective, runtime.EffectiveKnown)
		page.BlockSource = runtime.BlockSource
		for _, block := range runtime.Blocks {
			page.Blocks = append(page.Blocks, newReconcileBlockView(block, c.now()))
		}
	}
	states, err := c.opts.PositionPolicies.List(r.Context())
	if err != nil {
		page.LoadErr = err.Error()
		c.render(w, "position-policy", page)
		return
	}
	for _, state := range states {
		management := positionpolicy.ProjectManagement(positionpolicy.ManagementInput{
			Market: state.Market, Symbol: state.Symbol, JournalKnown: true,
			Managed:  state.Status == positionpolicy.StatusManaged && state.ExitEligible,
			Released: state.Status == positionpolicy.StatusReleased && state.Version > 0,
			Runtime:  runtime,
		})
		row := policyRowView{State: state, Management: management}
		if management.Block != nil {
			row.Block = newReconcileBlockView(*management.Block, c.now())
		}
		if state.Status == positionpolicy.StatusManaged {
			row.Actions = append(row.Actions, c.policyAction(state, positionpolicy.ActionInherit, "", "공통 정책 상속", false))
			for _, policy := range exitpolicy.RegisteredCommonPolicies() {
				row.Actions = append(row.Actions, c.policyAction(state, positionpolicy.ActionOverride,
					policy.ID, policy.Label, false))
			}
			if state.ExternalLifecycleEligible() {
				row.Actions = append(row.Actions, c.policyAction(state, positionpolicy.ActionRelease, "", "자동관리 해제", true))
			}
		} else if state.Status == positionpolicy.StatusReleased && state.ExternalLifecycleEligible() {
			row.Actions = append(row.Actions, c.policyAction(state, positionpolicy.ActionReadopt, "", "새 generation 재편입", true))
		}
		page.Rows = append(page.Rows, row)
	}
	c.render(w, "position-policy", page)
}

func (c *Console) policyAction(state positionpolicy.State, action positionpolicy.Action,
	policyID, label string, danger bool) policyActionView {
	reason := reasonForPolicyAction(action)
	payload := policyTokenPayload{
		Purpose: "select", PositionID: state.PositionID, Generation: state.AdoptionGeneration,
		Version: state.Version, Action: action, PolicyID: policyID, Reason: reason,
		NotBefore: c.now().UnixNano(), ExpiresAt: c.now().Add(positionPolicyTokenTTL).UnixNano(),
	}
	token, _ := c.signPositionPolicyToken(payload)
	return policyActionView{Label: label, Help: string(reason), Token: token, Danger: danger}
}

func reasonForPolicyAction(action positionpolicy.Action) positionpolicy.Reason {
	return map[positionpolicy.Action]positionpolicy.Reason{
		positionpolicy.ActionOverride: positionpolicy.ReasonPolicyOverride,
		positionpolicy.ActionInherit:  positionpolicy.ReasonPolicyInherit,
		positionpolicy.ActionRelease:  positionpolicy.ReasonRelease,
		positionpolicy.ActionReadopt:  positionpolicy.ReasonReadopt,
	}[action]
}

func (c *Console) handlePositionPolicyPreview(w http.ResponseWriter, r *http.Request) {
	if c.opts.PositionPolicies == nil {
		c.refuse(w, http.StatusNotImplemented, "정책 command seam 미배선", "엔진 소유 command service가 없다.")
		return
	}
	payload, err := c.verifyPositionPolicyToken(r.PostFormValue("action_token"), "select")
	if err != nil {
		c.refuse(w, http.StatusForbidden, "정책 action token 거부", err.Error())
		return
	}
	req := payload.request(c.now())
	preview, err := c.opts.PositionPolicies.Preview(r.Context(), req)
	if err != nil {
		c.writePositionPolicyError(w, err)
		return
	}
	if strings.TrimSpace(preview.Capability) == "" {
		c.refuse(w, http.StatusInternalServerError, "preview token 생성 실패", "아무것도 변경되지 않았다.")
		return
	}
	danger := payload.Action == positionpolicy.ActionRelease || payload.Action == positionpolicy.ActionReadopt
	wait := 0
	if danger {
		wait = 3
	}
	c.render(w, "position-policy-preview", positionPolicyPreviewPage{
		Nav: "position-management", CSRF: c.csrf, Preview: preview, Token: preview.Capability,
		Danger: danger, WaitSecs: wait,
	})
}

func (c *Console) handlePositionPolicyApply(w http.ResponseWriter, r *http.Request) {
	if c.opts.PositionPolicies == nil {
		c.refuse(w, http.StatusNotImplemented, "정책 command seam 미배선", "엔진 소유 command service가 없다.")
		return
	}
	capability := strings.TrimSpace(r.PostFormValue("capability"))
	if capability == "" {
		c.refuse(w, http.StatusForbidden, "정책 승인 token 거부", "engine capability가 없다. 아무것도 변경되지 않았다.")
		return
	}
	state, err := c.opts.PositionPolicies.Apply(r.Context(), positionpolicy.ApplyRequest{
		Capability: capability, Confirmed: r.PostFormValue("confirm") == "yes",
	})
	if err != nil {
		c.writePositionPolicyError(w, err)
		return
	}
	notice := fmt.Sprintf("적용됨 — %s · generation %d · version %d. engine preview에 바인딩된 변경만 1회 반영됐다.",
		state.Symbol, state.AdoptionGeneration, state.Version)
	http.Redirect(w, r, "/position-management?notice="+urlQueryEscape(notice), http.StatusSeeOther)
}

func (c *Console) writePositionPolicyError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, positionpolicy.ErrVersionMismatch):
		c.refuse(w, http.StatusPreconditionFailed, "정책 version 충돌",
			"현재 generation/version을 다시 불러와 preview부터 다시 진행하라. 자동 재시도하지 않았고 아무것도 변경되지 않았다.")
	case errors.Is(err, positionpolicy.ErrExitConflict):
		c.refuse(w, http.StatusConflict, "active exit 충돌",
			"미체결 exit 또는 CLOSING 상태가 끝난 뒤 새 preview를 만들어라. 아무것도 변경되지 않았다.")
	case errors.Is(err, positionpolicy.ErrIneligible):
		c.refuse(w, http.StatusConflict, "재편입 불가",
			"entry/adoption 근거가 없는 보유분은 자동관리 대상으로 표시하거나 재편입할 수 없다. 아무것도 변경되지 않았다.")
	case errors.Is(err, positionpolicy.ErrCapabilityTooEarly):
		c.refuse(w, http.StatusTooEarly, "3초 확인 대기 중",
			"engine preview 후 최소 3초가 지나야 한다. 같은 화면에서 잠시 기다린 뒤 한 번만 다시 적용하라.")
	case errors.Is(err, positionpolicy.ErrCapabilityExpired):
		c.refuse(w, http.StatusGone, "정책 capability 만료",
			"현재값을 다시 불러와 preview부터 다시 진행하라. 아무것도 변경되지 않았다.")
	case errors.Is(err, positionpolicy.ErrCapabilityStale):
		c.refuse(w, http.StatusGone, "재편입 관측 만료",
			"승인한 t0 관측이 오래됐거나 시계가 뒤로 이동했다. 현재값에서 새 preview를 만들어라. 아무것도 변경되지 않았다.")
	case errors.Is(err, positionpolicy.ErrCapabilityInvalid):
		c.refuse(w, http.StatusForbidden, "정책 capability 거부",
			"preview 없이 적용했거나 이미 사용된 capability다. 현재값에서 새 preview를 만들어라.")
	case errors.Is(err, positionpolicy.ErrConfirmationRequired):
		c.refuse(w, http.StatusBadRequest, "위험 변경 확인 필요",
			"해제·재편입의 위험 안내를 확인한 뒤 체크박스를 선택하라. 문구 입력은 필요 없고 아무것도 변경되지 않았다.")
	default:
		c.refuse(w, http.StatusBadRequest, "정책 요청 거부", err.Error()+" 아무것도 변경되지 않았다.")
	}
}

var errPolicyTokenTooEarly = errors.New("position policy selection token is not active yet")

func (c *Console) signPositionPolicyToken(payload policyTokenPayload) (string, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	encoded := base64.RawURLEncoding.EncodeToString(body)
	mac := hmac.New(sha256.New, c.policySeal)
	_, _ = mac.Write([]byte(encoded))
	return encoded + "." + base64.RawURLEncoding.EncodeToString(mac.Sum(nil)), nil
}

func (c *Console) verifyPositionPolicyToken(token, purpose string) (policyTokenPayload, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		return policyTokenPayload{}, errors.New("opaque token 형식이 올바르지 않다")
	}
	mac := hmac.New(sha256.New, c.policySeal)
	_, _ = mac.Write([]byte(parts[0]))
	signature, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil || !hmac.Equal(signature, mac.Sum(nil)) {
		return policyTokenPayload{}, errors.New("opaque token 서명이 일치하지 않는다")
	}
	body, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return policyTokenPayload{}, errors.New("opaque token payload를 읽을 수 없다")
	}
	var payload policyTokenPayload
	if err := json.Unmarshal(body, &payload); err != nil || payload.Purpose != purpose {
		return policyTokenPayload{}, errors.New("opaque token 용도가 일치하지 않는다")
	}
	now := c.now().UnixNano()
	if now >= payload.ExpiresAt {
		return policyTokenPayload{}, errors.New("opaque token이 만료됐다")
	}
	if now < payload.NotBefore {
		return policyTokenPayload{}, errPolicyTokenTooEarly
	}
	if payload.Reason != reasonForPolicyAction(payload.Action) {
		return policyTokenPayload{}, errors.New("opaque token의 action/reason이 일치하지 않는다")
	}
	return payload, nil
}

func (p policyTokenPayload) request(now time.Time) positionpolicy.Request {
	return positionpolicy.Request{
		PositionID: p.PositionID, ExpectedGeneration: p.Generation, ExpectedVersion: p.Version,
		Action: p.Action, PolicyID: p.PolicyID, Actor: positionpolicy.ActorLocalOperator,
		Reason: p.Reason, At: now.UTC(),
	}
}
