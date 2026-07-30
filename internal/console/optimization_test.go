package console

import (
	"net/http"
	"net/url"
	"strings"
	"sync"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/config"
	"github.com/JungHoonGhae/tossinvest-cli/internal/exitpolicy"
)

type fakeExitPolicySettings struct {
	mu    sync.Mutex
	value config.ExitPolicy
	saves int
}

func (f *fakeExitPolicySettings) Load() (config.ExitPolicy, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.value, nil
}

func (f *fakeExitPolicySettings) Save(value config.ExitPolicy) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.value = value
	f.saves++
	return nil
}

func TestOptimizationShowsExactlyThreePoliciesAndExternalBehavior(t *testing.T) {
	seam := &fakeExitPolicySettings{}
	h := newDashboardHarness(t, func(options *Options) { options.ExitPolicies = seam })
	h.authenticate(t)

	page := h.page(t, "/optimization")
	for _, want := range []string{
		exitpolicy.CommonLadderBalanced, exitpolicy.CommonLadderRunner,
		exitpolicy.CommonLadderHybrid50, "권장", "외부 매수", "다음 엔진 기동",
		"기존 포지션", "1.8", "3.0", "4.8", "6.5", "약 50%",
	} {
		if !strings.Contains(page, want) {
			t.Errorf("optimization page does not contain %q", want)
		}
	}
	if got := strings.Count(page, `name="common_policy"`); got != 3 {
		t.Fatalf("policy choices = %d, want exactly 3", got)
	}
}

func TestOptimizationSaveRequiresSessionAndCSRF(t *testing.T) {
	seam := &fakeExitPolicySettings{}
	h := newDashboardHarness(t, func(options *Options) { options.ExitPolicies = seam })
	h.authenticate(t)

	resp := h.post(t, "/optimization/exit-policy", url.Values{
		"common_policy": {exitpolicy.CommonLadderHybrid50},
	})
	if resp.StatusCode == http.StatusOK && resp.Request.URL.Path != "/refused" {
		t.Fatalf("POST without CSRF unexpectedly returned %d", resp.StatusCode)
	}
	if seam.saves != 0 {
		t.Fatal("POST without CSRF reached the save seam")
	}

	h.post(t, "/optimization/exit-policy", url.Values{
		"csrf": {h.csrf}, "common_policy": {exitpolicy.CommonLadderHybrid50},
	})
	if seam.saves != 1 || seam.value.CommonPolicy != exitpolicy.CommonLadderHybrid50 {
		t.Fatalf("save = count:%d value:%+v", seam.saves, seam.value)
	}
}

func TestOptimizationRejectsAClientInventedPolicy(t *testing.T) {
	seam := &fakeExitPolicySettings{}
	h := newDashboardHarness(t, func(options *Options) { options.ExitPolicies = seam })
	h.authenticate(t)
	h.post(t, "/optimization/exit-policy", url.Values{
		"csrf": {h.csrf}, "common_policy": {"INVENTED"},
	})
	if seam.saves != 0 {
		t.Fatal("invented policy reached the save seam")
	}
}
