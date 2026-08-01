package console

import (
	"context"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/exitpolicy"
	strategyopt "github.com/JungHoonGhae/tossinvest-cli/internal/optimization"
)

type fakeOptimizationCommander struct {
	mu       sync.Mutex
	view     strategyopt.View
	previews int
	applies  int
}

func newFakeOptimizationCommander(t *testing.T) *fakeOptimizationCommander {
	t.Helper()
	registry, err := strategyopt.CoreRegistry(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	return &fakeOptimizationCommander{view: strategyopt.View{
		Registry: registry,
		Snapshot: strategyopt.Snapshot{Version: 12, EffectiveVersion: 11,
			Desired: map[string]string{}, Effective: map[string]string{}, RestartRequired: true,
			SettingsDigest: "settings-digest", Actor: "operator:test", Reason: strategyopt.ReasonServerPreset},
		Evidence: strategyopt.Evidence{Status: strategyopt.EvidenceUnavailable, Missing: []string{"a049-provider-unavailable"}},
	}}
}

func (f *fakeOptimizationCommander) Read(context.Context) (strategyopt.View, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.view, nil
}

func (f *fakeOptimizationCommander) Preview(_ context.Context, req strategyopt.PreviewRequest) (strategyopt.Preview, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if req.BaseVersion != f.view.Snapshot.Version {
		return strategyopt.Preview{}, strategyopt.ErrVersionConflict
	}
	for key, optionID := range req.Changes {
		field, ok := f.view.Registry.Field(key)
		if !ok || field.Category != req.Category || field.Descriptor.ValidateOption(optionID) != nil {
			return strategyopt.Preview{}, strategyopt.ErrInvalidCandidate
		}
	}
	f.previews++
	return strategyopt.Preview{BaseVersion: req.BaseVersion, Category: req.Category, Capability: "opaque-preview",
		Changes:                  []strategyopt.OptionChange{{Key: "exit.common-policy", AfterOptionID: req.Changes["exit.common-policy"]}},
		RiskConfirmationRequired: true, ExistingPositionsUnchanged: true, LiveStateUnchanged: true,
		NotBefore: time.Now().Add(3 * time.Second)}, nil
}

func (f *fakeOptimizationCommander) PreviewRollback(context.Context, strategyopt.RollbackPreviewRequest) (strategyopt.Preview, error) {
	return strategyopt.Preview{}, strategyopt.ErrInvalidCandidate
}

func (f *fakeOptimizationCommander) Apply(_ context.Context, req strategyopt.ApplyRequest) (strategyopt.ApplyResult, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if req.Capability != "opaque-preview" || !req.Confirmed {
		return strategyopt.ApplyResult{}, strategyopt.ErrCapabilityInvalid
	}
	f.applies++
	f.view.Snapshot.Version++
	return strategyopt.ApplyResult{Snapshot: f.view.Snapshot}, nil
}

func TestOptimizationShowsExactlySixCategoriesAndThreeOwnerPolicies(t *testing.T) {
	legacy := &fakeExitPolicySettings{}
	commander := newFakeOptimizationCommander(t)
	h := newDashboardHarness(t, func(options *Options) {
		options.ExitPolicies = legacy
		options.Optimization = commander
	})
	h.authenticate(t)

	page := h.page(t, "/optimization?category=exit-protection")
	for _, want := range []string{
		exitpolicy.CommonLadderBalanced, exitpolicy.CommonLadderRunner,
		exitpolicy.CommonLadderHybrid50, "추천 · 자동 저장 안 함", "신규 관리 포지션",
		"신규 포지션만", "기존 포지션", "1.8", "3.0", "4.8", "6.5", "약 50%",
		"Desired", "Effective", "Owner", "a041-complete-exit-line-contract",
	} {
		if !strings.Contains(page, want) {
			t.Errorf("optimization page does not contain %q", want)
		}
	}
	if got := strings.Count(page, `name="option_id"`); got != 3 {
		t.Fatalf("policy choices = %d, want exactly 3 server option IDs", got)
	}
	for _, category := range strategyopt.Categories() {
		if got := strings.Count(page, `href="/optimization?category=`+string(category.ID)+`"`); got != 1 {
			t.Errorf("category %s links = %d, want 1", category.ID, got)
		}
	}
}

func TestOptimizationPreviewRequiresSessionAndCSRFAndNeverUsesLegacySave(t *testing.T) {
	legacy := &fakeExitPolicySettings{}
	commander := newFakeOptimizationCommander(t)
	h := newDashboardHarness(t, func(options *Options) {
		options.ExitPolicies = legacy
		options.Optimization = commander
	})
	h.authenticate(t)
	values := url.Values{"base_version": {"12"}, "category": {"exit-protection"},
		"setting_key": {"exit.common-policy"}, "option_id": {exitpolicy.CommonLadderHybrid50}}

	resp := h.post(t, "/optimization/exit-policy", values)
	if resp.StatusCode == http.StatusOK && resp.Request.URL.Path != "/refused" {
		t.Fatalf("POST without CSRF unexpectedly returned %d", resp.StatusCode)
	}
	if commander.previews != 0 || legacy.saves != 0 {
		t.Fatal("POST without CSRF reached a write seam")
	}
	values.Set("csrf", h.csrf)
	h.post(t, "/optimization/exit-policy", values)
	if commander.previews != 1 || legacy.saves != 0 {
		t.Fatalf("preview=%d legacy saves=%d", commander.previews, legacy.saves)
	}
}

func TestOptimizationLifecycleRejectsAClientInventedPolicy(t *testing.T) {
	commander := newFakeOptimizationCommander(t)
	h := newDashboardHarness(t, func(options *Options) { options.Optimization = commander })
	h.authenticate(t)
	resp := h.post(t, "/optimization/exit-policy", url.Values{
		"csrf": {h.csrf}, "base_version": {"12"}, "category": {"exit-protection"},
		"setting_key": {"exit.common-policy"}, "option_id": {"INVENTED"},
	})
	if resp.StatusCode != http.StatusBadRequest || commander.previews != 0 {
		t.Fatalf("invented option response=%d previews=%d", resp.StatusCode, commander.previews)
	}
}

func TestOptimizationUnknownCategoryFallsBackToOverviewWithoutMutation(t *testing.T) {
	commander := newFakeOptimizationCommander(t)
	h := newDashboardHarness(t, func(options *Options) { options.Optimization = commander })
	h.authenticate(t)
	page := h.page(t, "/optimization?category=does-not-exist")
	for _, want := range []string{"알 수 없는 카테고리", "개요", "LIVE 권한", "변경 안 함"} {
		if !strings.Contains(page, want) {
			t.Errorf("fallback page lacks %q", want)
		}
	}
	if commander.previews != 0 || commander.applies != 0 {
		t.Fatal("category navigation performed a mutation")
	}
}

func TestOptimizationLifecycleUnwiredRefusesPOSTInsteadOfLegacyBypass(t *testing.T) {
	legacy := &fakeExitPolicySettings{}
	h := newDashboardHarness(t, func(options *Options) { options.ExitPolicies = legacy })
	h.authenticate(t)
	resp := h.post(t, "/optimization/exit-policy", url.Values{"csrf": {h.csrf},
		"base_version": {"1"}, "category": {"exit-protection"}, "setting_key": {"exit.common-policy"},
		"option_id": {exitpolicy.CommonLadderBalanced}})
	if resp.StatusCode != http.StatusNotImplemented || legacy.saves != 0 {
		t.Fatalf("unwired response=%d legacy saves=%d", resp.StatusCode, legacy.saves)
	}
}

func TestOptimizationApplyMapsCASConflictTo412WithoutRetry(t *testing.T) {
	commander := newFakeOptimizationCommander(t)
	commander.view.Snapshot.Version = 13
	h := newDashboardHarness(t, func(options *Options) { options.Optimization = commander })
	h.authenticate(t)
	resp := h.post(t, "/optimization/exit-policy", url.Values{"csrf": {h.csrf},
		"base_version": {"12"}, "category": {"exit-protection"}, "setting_key": {"exit.common-policy"},
		"option_id": {exitpolicy.CommonLadderBalanced}})
	if resp.StatusCode != http.StatusPreconditionFailed {
		t.Fatalf("CAS response = %d, want 412", resp.StatusCode)
	}
}
