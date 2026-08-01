package console

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/JungHoonGhae/tossinvest-cli/internal/exitpolicy"
	strategyopt "github.com/JungHoonGhae/tossinvest-cli/internal/optimization"
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

func (optimizationPage) Refresh() bool { return false }

func (c *Console) handleOptimization(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		w.Header().Set("Allow", http.MethodGet+", "+http.MethodHead)
		c.refuse(w, http.StatusMethodNotAllowed, "읽기 전용 조회다",
			"최적화 조회는 GET/HEAD만 허용하고 변경은 CSRF가 적용된 preview/apply 경로로 분리한다.")
		return
	}
	selected, known := strategyopt.ParseCategory(strings.TrimSpace(r.URL.Query().Get("category")))
	if strings.TrimSpace(r.URL.Query().Get("category")) == "" {
		known = true
	}
	page := optimizationPage{
		Nav: "optimization", CSRF: c.csrf, Notice: r.URL.Query().Get("notice"),
		Selected: selected, EngineRunning: c.engineRunning(), LifecycleWired: c.opts.Optimization != nil,
	}
	if !known {
		page.Warning = "알 수 없는 카테고리다. 개요로 안전하게 이동했으며 어떤 동작도 실행하지 않았다."
	}
	page.Categories = optimizationCategoryViews(selected, c.opts.Optimization != nil)

	if c.opts.ExitPolicies != nil {
		value, err := c.opts.ExitPolicies.Load()
		if err != nil {
			page.LegacyLoadErr = err.Error()
		} else {
			page.LegacyCurrent = value
		}
	}
	if c.opts.Optimization != nil {
		view, err := c.opts.Optimization.Read(r.Context())
		if err != nil {
			page.LifecycleErr = err.Error()
		} else {
			page.LifecycleReady = true
			page.Snapshot = view.Snapshot
			page.Evidence = view.Evidence
			page.History = view.History
			page.Audit = view.Audit
			page.Fields = optimizationFieldViews(view.Registry.Fields(selected), view.Snapshot)
		}
	}
	for _, policy := range exitpolicy.RegisteredCommonPolicies() {
		selectedPolicy := page.Snapshot.Desired["exit.common-policy"]
		if selectedPolicy == "" {
			selectedPolicy = page.LegacyCurrent.CommonPolicy
		}
		page.Policies = append(page.Policies, optimizationPolicyView{
			CommonPolicy: policy, Selected: policy.ID == selectedPolicy,
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
	if c.opts.Optimization == nil {
		c.refuse(w, http.StatusNotImplemented, "최적화 command seam 미배선",
			"canonical control service가 없어 조회만 가능하다. legacy config seam으로 우회하지 않았다.")
		return
	}
	switch r.PostFormValue("action") {
	case "apply":
		result, err := c.opts.Optimization.Apply(r.Context(), strategyopt.ApplyRequest{
			Capability: strings.TrimSpace(r.PostFormValue("capability")), Confirmed: r.PostFormValue("confirm") == "yes",
		})
		if err != nil {
			c.writeOptimizationApplyError(w, r, strings.TrimSpace(r.PostFormValue("capability")), err)
			return
		}
		notice := fmt.Sprintf("적용됨 — desired v%d / effective v%d · audit %s. LIVE·lane·gate·기존 포지션은 그대로다.",
			result.Snapshot.Version, result.Snapshot.EffectiveVersion, result.Snapshot.AuditID)
		c.redirectOptimization(w, r, strategyopt.CategoryExitProtection, notice)
	case "rollback-preview":
		base, baseErr := strconv.ParseUint(r.PostFormValue("base_version"), 10, 64)
		target, targetErr := strconv.ParseUint(r.PostFormValue("target_version"), 10, 64)
		category, known := strategyopt.ParseCategory(r.PostFormValue("category"))
		if baseErr != nil || targetErr != nil || !known {
			c.writeOptimizationError(w, strategyopt.ErrInvalidCandidate)
			return
		}
		preview, err := c.opts.Optimization.PreviewRollback(r.Context(), strategyopt.RollbackPreviewRequest{
			BaseVersion: base, TargetVersion: target, Category: category,
		})
		if err != nil {
			c.writeOptimizationError(w, err)
			return
		}
		c.renderOptimizationPreview(w, preview)
	default:
		base, err := strconv.ParseUint(r.PostFormValue("base_version"), 10, 64)
		category, known := strategyopt.ParseCategory(r.PostFormValue("category"))
		key := strings.TrimSpace(r.PostFormValue("setting_key"))
		optionID := strings.TrimSpace(r.PostFormValue("option_id"))
		if err != nil || !known || key == "" || optionID == "" {
			c.writeOptimizationError(w, strategyopt.ErrInvalidCandidate)
			return
		}
		preview, err := c.opts.Optimization.Preview(r.Context(), strategyopt.PreviewRequest{
			BaseVersion: base, Category: category, Changes: map[string]string{key: optionID},
			Source: strategyopt.SourceServerPreset, Reason: strategyopt.ReasonServerPreset,
		})
		if err != nil {
			c.writeOptimizationError(w, err)
			return
		}
		c.renderOptimizationPreview(w, preview)
	}
}

func (c *Console) redirectOptimization(w http.ResponseWriter, r *http.Request, category strategyopt.Category, notice string) {
	destination := "/optimization?category=" + url.QueryEscape(string(category)) + "&notice=" + url.QueryEscape(notice)
	http.Redirect(w, r, destination, http.StatusSeeOther)
}
