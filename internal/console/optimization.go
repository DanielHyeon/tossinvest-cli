package console

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/JungHoonGhae/tossinvest-cli/internal/config"
	"github.com/JungHoonGhae/tossinvest-cli/internal/exitpolicy"
)

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
	Nav           string
	CSRF          string
	Notice        string
	Wired         bool
	Current       config.ExitPolicy
	LoadErr       string
	EngineRunning bool
	Policies      []optimizationPolicyView
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
	c.render(w, "optimization", page)
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
