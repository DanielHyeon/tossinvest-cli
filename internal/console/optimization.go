package console

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/JungHoonGhae/tossinvest-cli/internal/exitpolicy"
	strategyopt "github.com/JungHoonGhae/tossinvest-cli/internal/optimization"
)

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
	c.render(w, "optimization", page)
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
