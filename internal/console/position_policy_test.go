package console

import (
	"context"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/positionpolicy"
)

type fakePositionPolicyCommander struct {
	mu            sync.Mutex
	state         positionpolicy.State
	previews      int
	applies       int
	previewError  error
	applyError    error
	runtime       positionpolicy.ManagementRuntime
	runtimeError  error
	previewAfter  positionpolicy.State
	previewAction positionpolicy.Action
}

func (f *fakePositionPolicyCommander) Runtime(context.Context) (positionpolicy.ManagementRuntime, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.runtime, f.runtimeError
}

func (f *fakePositionPolicyCommander) List(context.Context) ([]positionpolicy.State, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return []positionpolicy.State{f.state}, nil
}

func (f *fakePositionPolicyCommander) Preview(_ context.Context, req positionpolicy.Request) (positionpolicy.Preview, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.previews++
	if f.previewError != nil {
		return positionpolicy.Preview{}, f.previewError
	}
	after := f.state
	after.Version++
	if req.Action == positionpolicy.ActionRelease {
		after.Status = positionpolicy.StatusReleased
		after.EffectivePolicyID = ""
	}
	if req.Action == positionpolicy.ActionReadopt {
		after.Status = positionpolicy.StatusManaged
		after.AdoptionGeneration++
		after.Version = 1
	}
	f.previewAfter, f.previewAction = after, req.Action
	return positionpolicy.Preview{Before: f.state, After: after, Action: req.Action,
		Reason: req.Reason, Capability: "opaque-engine-capability"}, nil
}

func (f *fakePositionPolicyCommander) Apply(_ context.Context, req positionpolicy.ApplyRequest) (positionpolicy.State, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if req.Capability != "opaque-engine-capability" {
		return positionpolicy.State{}, positionpolicy.ErrCapabilityInvalid
	}
	if (f.previewAction == positionpolicy.ActionRelease || f.previewAction == positionpolicy.ActionReadopt) && !req.Confirmed {
		return positionpolicy.State{}, positionpolicy.ErrConfirmationRequired
	}
	f.applies++
	if f.applyError != nil {
		return positionpolicy.State{}, f.applyError
	}
	f.state = f.previewAfter
	return f.state, nil
}

func managedPolicyState() positionpolicy.State {
	return positionpolicy.State{
		PositionID: "p-1", AccountRef: "acct-1", Market: "kr", Symbol: "005930",
		AdoptionGeneration: 1, Version: 4, Status: positionpolicy.StatusManaged,
		DesiredPolicyID: "", EffectivePolicyID: "COMMON_LADDER_BALANCED", PositionState: "OPEN",
		ExitEligible: true,
		Provenance:   positionpolicy.ProvenanceExternalAdoption,
		Eligibility:  positionpolicy.EligibilityExternalLifecycle,
	}
}

func policyHarness(t *testing.T, commander *fakePositionPolicyCommander) *dashboardHarness {
	t.Helper()
	return newDashboardHarness(t, func(options *Options) { options.PositionPolicies = commander })
}

func actionToken(t *testing.T, page, label string) string {
	t.Helper()
	pattern := `(?s)<form method="post" action="/position-management/preview">\s*` +
		`<input type="hidden" name="csrf"[^>]*><input type="hidden" name="action_token" value="([^"]+)">\s*` +
		`<button[^>]*aria-label="` + regexp.QuoteMeta(label) + ` preview"`
	match := regexp.MustCompile(pattern).FindStringSubmatch(page)
	if len(match) != 2 {
		t.Fatalf("action token for %q not found", label)
	}
	return match[1]
}

func applyToken(t *testing.T, page string) string {
	t.Helper()
	match := regexp.MustCompile(`name="capability" value="([^"]+)"`).FindStringSubmatch(page)
	if len(match) != 2 {
		t.Fatal("apply token not found")
	}
	return match[1]
}

func TestPositionManagementIsInputFreeResponsiveAndExplicit(t *testing.T) {
	commander := &fakePositionPolicyCommander{state: managedPolicyState()}
	h := policyHarness(t, commander)
	h.authenticate(t)
	page := h.page(t, "/position-management")

	for _, want := range []string{
		"position-management", "종목별 정책", "외부 매수 자동편입", "공통 정책 상속",
		"desired", "effective", "2~20%", "0.5% step", "exclude 우선", "1주 전량",
		`name="action_token"`, `name="csrf"`, `name="viewport"`, `@media (max-width: 720px)`,
	} {
		if !strings.Contains(page, want) {
			t.Errorf("position-management page missing %q", want)
		}
	}
	for _, banned := range []string{
		`type="text"`, `type="number"`, `<textarea`, `contenteditable`, `<select`, `<script`,
		`name="symbol"`, `name="reason"`, `name="policy_id"`,
	} {
		if strings.Contains(strings.ToLower(page), strings.ToLower(banned)) {
			t.Errorf("position-management page contains free input %q", banned)
		}
	}
	inputs := regexp.MustCompile(`<input[^>]*>`).FindAllString(page, -1)
	if len(inputs) == 0 {
		t.Fatal("no opaque action forms rendered")
	}
	for _, input := range inputs {
		if !strings.Contains(input, `type="hidden"`) ||
			(!strings.Contains(input, `name="csrf"`) && !strings.Contains(input, `name="action_token"`)) {
			t.Errorf("unexpected input: %s", input)
		}
	}
}

func TestPositionPolicyTokenTamperAndCSRFFailBeforeCommander(t *testing.T) {
	commander := &fakePositionPolicyCommander{state: managedPolicyState()}
	h := policyHarness(t, commander)
	h.authenticate(t)
	token := actionToken(t, h.page(t, "/position-management"), "자동관리 해제")

	resp := h.post(t, "/position-management/preview", url.Values{"action_token": {token}})
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("missing CSRF = %d", resp.StatusCode)
	}
	resp = h.post(t, "/position-management/preview", url.Values{
		"csrf": {h.csrf}, "action_token": {token + "tampered"},
	})
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("tampered token = %d", resp.StatusCode)
	}
	if commander.previews != 0 || commander.applies != 0 {
		t.Fatalf("refusal reached commander: preview=%d apply=%d", commander.previews, commander.applies)
	}
}

func TestPositionPolicyReleaseCarriesEngineCapabilityAndCheckbox(t *testing.T) {
	commander := &fakePositionPolicyCommander{state: managedPolicyState()}
	h := policyHarness(t, commander)
	h.authenticate(t)
	selection := actionToken(t, h.page(t, "/position-management"), "자동관리 해제")
	previewPage := body(t, h.post(t, "/position-management/preview", url.Values{
		"csrf": {h.csrf}, "action_token": {selection},
	}))
	for _, want := range []string{"보호 공백 위험", "3초", `type="checkbox"`, "문구 입력은 요구하지 않는다"} {
		if !strings.Contains(previewPage, want) {
			t.Errorf("release preview missing %q", want)
		}
	}
	token := applyToken(t, previewPage)
	resp := h.post(t, "/position-management/apply", url.Values{"csrf": {h.csrf}})
	if resp.StatusCode != http.StatusForbidden || commander.applies != 0 {
		t.Fatalf("missing capability = %d, calls=%d", resp.StatusCode, commander.applies)
	}
	resp = h.post(t, "/position-management/apply", url.Values{
		"csrf": {h.csrf}, "capability": {token},
	})
	if resp.StatusCode != http.StatusBadRequest || commander.applies != 0 {
		t.Fatalf("unchecked danger confirmation = %d, calls=%d", resp.StatusCode, commander.applies)
	}
	_ = h.post(t, "/position-management/apply", url.Values{
		"csrf": {h.csrf}, "capability": {token}, "confirm": {"yes"},
	})
	if commander.applies != 1 {
		t.Fatalf("approved apply calls = %d", commander.applies)
	}
}

func TestPositionPolicyPreviewExplainsReleaseAndFreshReadoptSemantics(t *testing.T) {
	t.Run("release removes the exact generation from the observer", func(t *testing.T) {
		commander := &fakePositionPolicyCommander{state: managedPolicyState()}
		h := policyHarness(t, commander)
		h.authenticate(t)
		selection := actionToken(t, h.page(t, "/position-management"), "자동관리 해제")
		preview := body(t, h.post(t, "/position-management/preview", url.Values{
			"csrf": {h.csrf}, "action_token": {selection},
		}))
		if !strings.Contains(preview, "현재 generation을 exit observer 대상에서 제거") {
			t.Fatalf("release preview hides observer effect: %s", preview)
		}
	})

	t.Run("readopt states that a fresh t0 and exit snapshot are created", func(t *testing.T) {
		state := managedPolicyState()
		state.Status = positionpolicy.StatusReleased
		commander := &fakePositionPolicyCommander{state: state}
		h := policyHarness(t, commander)
		h.authenticate(t)
		selection := actionToken(t, h.page(t, "/position-management"), "새 generation 재편입")
		preview := body(t, h.post(t, "/position-management/preview", url.Values{
			"csrf": {h.csrf}, "action_token": {selection},
		}))
		if !strings.Contains(preview, "엔진 관측가로 fresh t0") ||
			!strings.Contains(preview, "high-water·rung·pending을 초기화") {
			t.Fatalf("readopt preview hides reset semantics: %s", preview)
		}

		h.clock.advance(3 * time.Second)
		resp := h.post(t, "/position-management/apply", url.Values{
			"csrf": {h.csrf}, "capability": {applyToken(t, preview)}, "confirm": {"yes"},
		})
		if notice := body(t, resp); !strings.Contains(notice, "engine preview에 바인딩") {
			t.Fatalf("readopt notice does not identify bound mutation: %s", notice)
		}
	})
}

func TestPositionPolicyStaleReturns412AndNeverRetries(t *testing.T) {
	commander := &fakePositionPolicyCommander{
		state: managedPolicyState(), applyError: positionpolicy.ErrVersionMismatch,
	}
	h := policyHarness(t, commander)
	h.authenticate(t)
	selection := actionToken(t, h.page(t, "/position-management"), "균형형")
	previewPage := body(t, h.post(t, "/position-management/preview", url.Values{
		"csrf": {h.csrf}, "action_token": {selection},
	}))
	resp := h.post(t, "/position-management/apply", url.Values{
		"csrf": {h.csrf}, "capability": {applyToken(t, previewPage)},
	})
	if resp.StatusCode != http.StatusPreconditionFailed {
		t.Fatalf("stale = %d, want 412", resp.StatusCode)
	}
	if commander.applies != 1 {
		t.Fatalf("stale command retried %d times", commander.applies)
	}
	if text := body(t, resp); !strings.Contains(text, "자동 재시도하지 않았") {
		t.Fatalf("stale response lacks recovery guidance: %s", text)
	}
}

func TestPositionPolicyActiveExitConflictIs409(t *testing.T) {
	commander := &fakePositionPolicyCommander{
		state: managedPolicyState(), previewError: positionpolicy.ErrExitConflict,
	}
	h := policyHarness(t, commander)
	h.authenticate(t)
	token := actionToken(t, h.page(t, "/position-management"), "자동관리 해제")
	resp := h.post(t, "/position-management/preview", url.Values{
		"csrf": {h.csrf}, "action_token": {token},
	})
	if resp.StatusCode != http.StatusConflict {
		t.Fatalf("active exit conflict = %d", resp.StatusCode)
	}
}

func TestPositionManagementDoesNotOfferReadoptForIneligibleHolding(t *testing.T) {
	state := managedPolicyState()
	state.Status = positionpolicy.StatusReleased
	state.ExitEligible = false
	state.Provenance = positionpolicy.ProvenanceUnknown
	state.Eligibility = positionpolicy.EligibilityNone
	commander := &fakePositionPolicyCommander{state: state}
	h := policyHarness(t, commander)
	h.authenticate(t)
	page := h.page(t, "/position-management")
	if strings.Contains(page, "새 generation 재편입") {
		t.Fatal("ineligible holding was offered re-adopt")
	}
	if !strings.Contains(page, "관리 외(운영자 해제)") || !strings.Contains(page, "OPERATOR_RELEASED") {
		t.Fatal("ineligible holding lacks clear status")
	}
}

func TestPositionManagementOffersReleaseOnlyForExternalAdoption(t *testing.T) {
	t.Run("engine entry has policy choices but no release", func(t *testing.T) {
		state := managedPolicyState()
		state.Provenance = positionpolicy.ProvenanceEngineEntry
		state.Eligibility = positionpolicy.EligibilityExitOnly
		commander := &fakePositionPolicyCommander{state: state}
		h := policyHarness(t, commander)
		h.authenticate(t)
		page := h.page(t, "/position-management")
		if strings.Contains(page, "자동관리 해제") {
			t.Fatal("engine-entry position was offered lifecycle release")
		}
		for _, want := range []string{"균형형", string(positionpolicy.ProvenanceEngineEntry), string(positionpolicy.EligibilityExitOnly)} {
			if !strings.Contains(page, want) {
				t.Errorf("engine-entry row missing %q", want)
			}
		}
	})

	t.Run("external adoption has explicit release", func(t *testing.T) {
		commander := &fakePositionPolicyCommander{state: managedPolicyState()}
		h := policyHarness(t, commander)
		h.authenticate(t)
		page := h.page(t, "/position-management")
		for _, want := range []string{"자동관리 해제", string(positionpolicy.ProvenanceExternalAdoption), string(positionpolicy.EligibilityExternalLifecycle)} {
			if !strings.Contains(page, want) {
				t.Errorf("external-adoption row missing %q", want)
			}
		}
	})
}

func TestPositionPolicyBrowserCarriesOnlyOpaqueEngineCapabilityAtApply(t *testing.T) {
	commander := &fakePositionPolicyCommander{state: managedPolicyState()}
	h := policyHarness(t, commander)
	h.authenticate(t)
	selection := actionToken(t, h.page(t, "/position-management"), "균형형")
	preview := body(t, h.post(t, "/position-management/preview", url.Values{
		"csrf": {h.csrf}, "action_token": {selection},
	}))
	if !strings.Contains(preview, `name="capability" value="opaque-engine-capability"`) {
		t.Fatalf("engine capability absent: %s", preview)
	}
	if strings.Contains(preview, "3초") {
		t.Fatalf("safe policy selection unexpectedly requires danger delay: %s", preview)
	}
	for _, forbidden := range []string{
		`name="position_id"`, `name="generation"`, `name="version"`, `name="action"`,
		`name="policy_id"`, `name="reason"`, `name="action_token"`,
	} {
		if strings.Contains(preview, forbidden) {
			t.Errorf("apply form exposes mutation scope %q", forbidden)
		}
	}
}
