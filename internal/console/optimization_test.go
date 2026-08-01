package console

import (
	"bytes"
	"io"
	"mime/multipart"
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
	commander := newFakeOptimizationCommander(t)
	h := newDashboardHarness(t, func(options *Options) {
		options.ExitPolicies = seam
		options.Optimization = commander
	})
	h.authenticate(t)

	page := h.page(t, "/optimization?category=exit-protection")
	for _, want := range []string{
		exitpolicy.CommonLadderBalanced, exitpolicy.CommonLadderRunner,
		exitpolicy.CommonLadderHybrid50, "추천 · 자동 저장 안 함", "신규 관리 포지션", "신규 포지션만",
		"기존 포지션", "1.8", "3.0", "4.8", "6.5", "약 50%",
	} {
		if !strings.Contains(page, want) {
			t.Errorf("optimization page does not contain %q", want)
		}
	}
	if got := strings.Count(page, `name="option_id"`); got != 3 {
		t.Fatalf("policy choices = %d, want exactly 3", got)
	}
}

func TestOptimizationSaveRequiresSessionAndCSRF(t *testing.T) {
	seam := &fakeExitPolicySettings{}
	commander := newFakeOptimizationCommander(t)
	h := newDashboardHarness(t, func(options *Options) {
		options.ExitPolicies = seam
		options.Optimization = commander
	})
	h.authenticate(t)
	values := url.Values{
		"base_version": {"12"}, "category": {"exit-protection"},
		"setting_key": {"exit.common-policy"}, "option_id": {exitpolicy.CommonLadderHybrid50},
	}

	resp := h.post(t, "/optimization/exit-policy", values)
	if resp.StatusCode == http.StatusOK && resp.Request.URL.Path != "/refused" {
		t.Fatalf("POST without CSRF unexpectedly returned %d", resp.StatusCode)
	}
	if seam.saves != 0 || commander.previews != 0 {
		t.Fatal("POST without CSRF reached a write seam")
	}

	values.Set("csrf", h.csrf)
	h.post(t, "/optimization/exit-policy", values)
	if seam.saves != 0 || commander.previews != 1 {
		t.Fatalf("legacy saves=%d previews=%d", seam.saves, commander.previews)
	}
}

func TestOptimizationRejectsAClientInventedPolicy(t *testing.T) {
	seam := &fakeExitPolicySettings{}
	commander := newFakeOptimizationCommander(t)
	h := newDashboardHarness(t, func(options *Options) {
		options.ExitPolicies = seam
		options.Optimization = commander
	})
	h.authenticate(t)
	resp := h.post(t, "/optimization/exit-policy", url.Values{
		"csrf": {h.csrf}, "base_version": {"12"}, "category": {"exit-protection"},
		"setting_key": {"exit.common-policy"}, "option_id": {"INVENTED"},
	})
	if resp.StatusCode != http.StatusBadRequest || seam.saves != 0 || commander.previews != 0 {
		t.Fatalf("response=%d legacy saves=%d previews=%d", resp.StatusCode, seam.saves, commander.previews)
	}
}

func TestOptimizationRequestBodyIsBounded(t *testing.T) {
	commander := newFakeOptimizationCommander(t)
	h := newDashboardHarness(t, func(options *Options) { options.Optimization = commander })
	h.authenticate(t)
	body := "csrf=" + url.QueryEscape(h.csrf) + "&base_version=12&category=exit-protection" +
		"&setting_key=exit.common-policy&option_id=" + strings.Repeat("x", 4096)
	req, err := http.NewRequest(http.MethodPost, h.srv.URL+"/optimization/exit-policy", strings.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := h.client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body)
	if resp.StatusCode != http.StatusRequestEntityTooLarge || commander.previews != 0 || commander.applies != 0 {
		t.Fatalf("status=%d previews=%d applies=%d, want 413 before commander", resp.StatusCode, commander.previews, commander.applies)
	}
}

func TestOptimizationRejectsUnexpectedAndDuplicateFields(t *testing.T) {
	for _, tc := range []struct {
		name string
		form url.Values
	}{
		{name: "unexpected", form: url.Values{
			"csrf": {"placeholder"}, "base_version": {"12"}, "category": {"exit-protection"},
			"setting_key": {"exit.common-policy"}, "option_id": {exitpolicy.CommonLadderBalanced},
			"invented": {"ignored-before-a050"},
		}},
		{name: "duplicate", form: url.Values{
			"csrf": {"placeholder"}, "base_version": {"12", "13"}, "category": {"exit-protection"},
			"setting_key": {"exit.common-policy"}, "option_id": {exitpolicy.CommonLadderBalanced},
		}},
		{name: "missing required", form: url.Values{
			"csrf": {"placeholder"}, "base_version": {"12"}, "category": {"exit-protection"},
			"setting_key": {"exit.common-policy"},
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			commander := newFakeOptimizationCommander(t)
			h := newDashboardHarness(t, func(options *Options) { options.Optimization = commander })
			h.authenticate(t)
			tc.form.Set("csrf", h.csrf)
			resp := h.post(t, "/optimization/exit-policy", tc.form)
			if resp.StatusCode != http.StatusBadRequest || commander.previews != 0 || commander.applies != 0 {
				t.Fatalf("status=%d previews=%d applies=%d, want 400 before commander", resp.StatusCode, commander.previews, commander.applies)
			}
		})
	}
}

func TestOptimizationRejectsMultipartBeforeReadingPollutedValues(t *testing.T) {
	commander := newFakeOptimizationCommander(t)
	h := newDashboardHarness(t, func(options *Options) { options.Optimization = commander })
	h.authenticate(t)
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	for key, value := range map[string]string{
		"csrf": h.csrf, "base_version": "12", "category": "exit-protection",
		"setting_key": "exit.common-policy", "option_id": exitpolicy.CommonLadderBalanced,
		"invented": "must-not-bypass-validation",
	} {
		if err := writer.WriteField(key, value); err != nil {
			t.Fatal(err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	req, err := http.NewRequest(http.MethodPost, h.srv.URL+"/optimization/exit-policy", &body)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	resp, err := h.client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body)
	if resp.StatusCode != http.StatusBadRequest || commander.previews != 0 || commander.applies != 0 {
		t.Fatalf("status=%d previews=%d applies=%d, want 400 before commander", resp.StatusCode, commander.previews, commander.applies)
	}
}
