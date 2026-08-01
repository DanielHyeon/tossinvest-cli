package console

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/JungHoonGhae/tossinvest-cli/internal/config"
	"github.com/JungHoonGhae/tossinvest-cli/internal/exitpolicy"
)

var (
	ErrProtectionActionTooEarly = errors.New("console protection: 3-second confirmation delay has not elapsed")
	ErrProtectionActionStale    = errors.New("console protection: preview capability is stale")
	ErrProtectionConfirmation   = errors.New("console protection: risk checkbox is required")
)

type ProtectionStatus struct {
	SagaID               string
	Symbol               string
	Capability           string
	Activation           string
	Desired              string
	Effective            string
	EffectiveTrigger     string
	ProtectedQuantity    string
	BrokerID             string
	UpdatedAt            string
	ReconcileReason      string
	ApplyTiming          string
	Provenance           string
	Explanation          string
	WeakeningAction      string
	WeakeningActionToken string
}

type ProtectionPreview struct {
	Symbol            string
	Before            string
	After             string
	AffectedPositions string
	AffectedQuantity  string
	CoverageGap       string
	ApplyTiming       string
	Capability        string
	Weakening         bool
}

// ProtectionCommander owns both the read model and opaque, one-shot command
// capabilities. The browser can select only an action token already attached to
// the current row; it cannot submit a symbol, trigger, quantity, or reason.
type ProtectionCommander interface {
	List(context.Context) ([]ProtectionStatus, error)
	Preview(context.Context, string) (ProtectionPreview, error)
	Apply(context.Context, string, bool) (ProtectionStatus, error)
}

// ExitPolicySettings is deliberately smaller than every operating seam.
type ExitPolicySettings interface {
	Load() (config.ExitPolicy, error)
	Save(config.ExitPolicy) error
}

type optimizationPolicyView struct {
	exitpolicy.CommonPolicy
	Selected bool
}

type optimizationPage struct {
	Nav               string
	CSRF              string
	Notice            string
	Wired             bool
	Current           config.ExitPolicy
	LoadErr           string
	EngineRunning     bool
	Policies          []optimizationPolicyView
	ProtectionWired   bool
	ProtectionLoadErr string
	Protections       []ProtectionStatus
}

func (optimizationPage) Refresh() bool { return false }

func (c *Console) handleOptimization(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		w.Header().Set("Allow", http.MethodGet+", "+http.MethodHead)
		c.refuse(w, http.StatusMethodNotAllowed, "읽기 전용 화면이다",
			"최적화 화면 조회는 GET/HEAD만 허용한다.")
		return
	}
	page := optimizationPage{
		Nav: "optimization", CSRF: c.csrf, Notice: r.URL.Query().Get("notice"),
		EngineRunning: c.engineRunning(),
	}
	if c.opts.ExitPolicies != nil {
		page.Wired = true
		value, err := c.opts.ExitPolicies.Load()
		if err != nil {
			page.LoadErr = err.Error()
		}
		page.Current = value
	}
	for _, policy := range exitpolicy.RegisteredCommonPolicies() {
		page.Policies = append(page.Policies, optimizationPolicyView{
			CommonPolicy: policy, Selected: policy.ID == page.Current.CommonPolicy,
		})
	}
	if c.opts.Protections != nil {
		page.ProtectionWired = true
		protections, err := c.opts.Protections.List(r.Context())
		if err != nil {
			page.ProtectionLoadErr = err.Error()
		} else {
			page.Protections = protections
		}
	}
	c.render(w, "optimization", page)
}

type protectionPreviewPage struct {
	Nav     string
	CSRF    string
	Preview ProtectionPreview
}

func (protectionPreviewPage) Refresh() bool { return false }

func (c *Console) handleProtectionPreview(w http.ResponseWriter, r *http.Request) {
	if c.opts.Protections == nil {
		c.refuse(w, http.StatusNotImplemented, "보호 command seam 미배선", "현재 상태는 OFF / 지원 확인 전 사용 불가다.")
		return
	}
	token := strings.TrimSpace(r.PostFormValue("action_token"))
	if token == "" {
		c.refuse(w, http.StatusForbidden, "보호 action token 거부", "현재 행에서 발급된 opaque token이 없다.")
		return
	}
	preview, err := c.opts.Protections.Preview(r.Context(), token)
	if err != nil {
		c.refuse(w, http.StatusConflict, "보호 변경 preview 거부", err.Error())
		return
	}
	if !preview.Weakening || strings.TrimSpace(preview.Capability) == "" {
		c.refuse(w, http.StatusForbidden, "보호 약화 capability 거부", "server-defined 약화 preview가 아니다.")
		return
	}
	c.render(w, "protection-preview", protectionPreviewPage{Nav: "optimization", CSRF: c.csrf, Preview: preview})
}

func (c *Console) handleProtectionApply(w http.ResponseWriter, r *http.Request) {
	if c.opts.Protections == nil {
		c.refuse(w, http.StatusNotImplemented, "보호 command seam 미배선", "아무것도 변경되지 않았다.")
		return
	}
	capability := strings.TrimSpace(r.PostFormValue("capability"))
	confirmed := r.PostFormValue("confirm") == "yes"
	if capability == "" {
		c.refuse(w, http.StatusForbidden, "보호 capability 거부", "preview에서 발급된 opaque token이 없다.")
		return
	}
	status, err := c.opts.Protections.Apply(r.Context(), capability, confirmed)
	if err != nil {
		switch {
		case errors.Is(err, ErrProtectionActionTooEarly):
			c.refuse(w, http.StatusTooEarly, "3초 확인 대기 중", "잠시 기다린 뒤 같은 preview에서 한 번만 적용하라.")
		case errors.Is(err, ErrProtectionActionStale):
			c.refuse(w, http.StatusPreconditionFailed, "보호 상태가 바뀌었다", "현재 행을 다시 불러와 preview부터 시작하라.")
		case errors.Is(err, ErrProtectionConfirmation):
			c.refuse(w, http.StatusBadRequest, "위험 변경 확인 필요", "체크박스를 선택하라. 문구 입력은 필요 없다.")
		default:
			c.refuse(w, http.StatusConflict, "보호 변경 거부", err.Error()+" 아무것도 변경되지 않았다.")
		}
		return
	}
	http.Redirect(w, r, "/optimization?category=exit-protection&notice="+url.QueryEscape(fmt.Sprintf(
		"적용됨 — %s 보호 상태 %s. engine capability에 바인딩된 현재 행만 변경됐다.", status.Symbol, status.Effective)), http.StatusSeeOther)
}

func (c *Console) handleOptimizationSave(w http.ResponseWriter, r *http.Request) {
	if c.opts.ExitPolicies == nil {
		c.refuse(w, http.StatusNotImplemented, "저장이 배선되지 않았다",
			"이 빌드에는 공통 exit 정책 저장 seam이 없다.")
		return
	}
	id := strings.TrimSpace(r.PostFormValue("common_policy"))
	if _, ok := exitpolicy.CommonPolicyByID(id); !ok {
		c.redirectOptimization(w, r, "저장 안 됨 — 등록되지 않은 정책 ID다.")
		return
	}
	if err := c.opts.ExitPolicies.Save(config.ExitPolicy{CommonPolicy: id}); err != nil {
		c.redirectOptimization(w, r, "저장 안 됨 — "+err.Error())
		return
	}
	c.redirectOptimization(w, r,
		"저장됨 — 다음 엔진 기동부터 새로 관리되는 포지션에만 적용된다. 기존 포지션은 바뀌지 않는다.")
}

func (c *Console) redirectOptimization(w http.ResponseWriter, r *http.Request, notice string) {
	http.Redirect(w, r, "/optimization?notice="+url.QueryEscape(notice), http.StatusSeeOther)
}
