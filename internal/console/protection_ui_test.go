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
)

type fakeProtectionCommander struct {
	mu        sync.Mutex
	clock     func() time.Time
	status    ProtectionStatus
	previewAt time.Time
	previews  int
	applies   int
}

func (f *fakeProtectionCommander) List(context.Context) ([]ProtectionStatus, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return []ProtectionStatus{f.status}, nil
}

func (f *fakeProtectionCommander) Preview(_ context.Context, token string) (ProtectionPreview, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if token != "opaque-row-action" {
		return ProtectionPreview{}, ErrProtectionActionStale
	}
	f.previews++
	f.previewAt = f.clock()
	return ProtectionPreview{
		Symbol: "005930", Before: "ACTIVE · 70,000원 · 1주", After: "보호 해제",
		AffectedPositions: "현재 포지션 1개", AffectedQuantity: "1주", CoverageGap: "브로커 상주 보호가 사라짐",
		ApplyTiming: "승인 즉시; 신규 entry는 계속 차단", Capability: "opaque-preview-capability", Weakening: true,
	}, nil
}

func (f *fakeProtectionCommander) Apply(_ context.Context, capability string, confirmed bool) (ProtectionStatus, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if capability != "opaque-preview-capability" || f.previewAt.IsZero() {
		return ProtectionStatus{}, ErrProtectionActionStale
	}
	if !confirmed {
		return ProtectionStatus{}, ErrProtectionConfirmation
	}
	if f.clock().Sub(f.previewAt) < 3*time.Second {
		return ProtectionStatus{}, ErrProtectionActionTooEarly
	}
	f.applies++
	f.status.Effective = "RECONCILE"
	return f.status, nil
}

func protectionStatus() ProtectionStatus {
	return ProtectionStatus{
		SagaID: "saga-1", Symbol: "005930", Capability: "SINGLE+MARKET · signed", Activation: "ON",
		Desired: "ACTIVE", Effective: "ACTIVE", EffectiveTrigger: "70,000원", ProtectedQuantity: "1주",
		BrokerID: "co-opaque-7", UpdatedAt: "2026-08-01 10:00:00 KST", ReconcileReason: "일치 · 다음 주기 재확인",
		ApplyTiming: "현재 generation", Provenance: "attestation digest abcd… · official-only",
		Explanation:     "엔진이 종료되어도 브로커가 이 손절을 감시한다.",
		WeakeningAction: "현재 보호 해제 preview", WeakeningActionToken: "opaque-row-action",
	}
}

func TestExitProtectionDefaultIsHonestInputFreeAndResponsive(t *testing.T) {
	h := newDashboardHarness(t)
	h.authenticate(t)
	resp := h.get(t, "/optimization?category=exit-protection")
	page := body(t, resp)
	for _, want := range []string{"청산/보호", "지원 확인 전 사용 불가", "Activation", "OFF", "UNWIRED", "운영자 별도 승인", `name="viewport"`, `@media (max-width: 720px)`} {
		if !strings.Contains(page, want) {
			t.Errorf("default protection page missing %q", want)
		}
	}
	section := protectionSection(t, page)
	for _, banned := range []string{`type="text"`, `type="number"`, `<textarea`, `contenteditable`, `name="symbol"`, `name="reason"`, "enable-all"} {
		if strings.Contains(strings.ToLower(section), strings.ToLower(banned)) {
			t.Errorf("default protection section contains %q", banned)
		}
	}
	if policy := resp.Header.Get("Content-Security-Policy"); !strings.Contains(policy, "default-src 'none'") || !strings.Contains(policy, "form-action 'self'") {
		t.Fatalf("CSP=%q", policy)
	}
}

func TestExitProtectionCurrentRowUsesOnlyOpaqueActionAndCheckbox(t *testing.T) {
	commander := &fakeProtectionCommander{status: protectionStatus()}
	h := newDashboardHarness(t, func(options *Options) { options.Protections = commander })
	commander.clock = h.clock.Now
	h.authenticate(t)
	page := h.page(t, "/optimization?category=exit-protection")
	section := protectionSection(t, page)
	for _, want := range []string{"SINGLE", "MARKET", "70,000원", "1주", "co-opaque-7", "일치 · 다음 주기 재확인", "엔진이 종료", "현재 보호 해제 preview", `name="action_token"`} {
		if !strings.Contains(section, want) {
			t.Errorf("wired protection section missing %q", want)
		}
	}
	for _, banned := range []string{`type="text"`, `type="number"`, `<textarea`, `contenteditable`, `<select`, `name="symbol"`, `name="trigger"`, `name="quantity"`, `name="reason"`} {
		if strings.Contains(strings.ToLower(section), strings.ToLower(banned)) {
			t.Errorf("wired protection section contains free input %q", banned)
		}
	}
}

func TestExitProtectionWeakeningPreviewRequiresCheckboxAndThreeSeconds(t *testing.T) {
	commander := &fakeProtectionCommander{status: protectionStatus()}
	h := newDashboardHarness(t, func(options *Options) { options.Protections = commander })
	commander.clock = h.clock.Now
	h.authenticate(t)

	preview := body(t, h.post(t, "/optimization/exit-protection/preview", url.Values{
		"csrf": {h.csrf}, "action_token": {"opaque-row-action"},
	}))
	for _, want := range []string{"Before", "After", "현재 포지션 1개", "1주", "보호 공백 가능성", "최소 3초", `type="checkbox"`, `name="capability"`} {
		if !strings.Contains(preview, want) {
			t.Errorf("preview missing %q", want)
		}
	}
	for _, banned := range []string{`type="text"`, `type="number"`, `<textarea`, `contenteditable`, `name="symbol"`, `name="reason"`} {
		if strings.Contains(strings.ToLower(preview), strings.ToLower(banned)) {
			t.Errorf("preview contains free input %q", banned)
		}
	}

	resp := h.post(t, "/optimization/exit-protection/apply", url.Values{
		"csrf": {h.csrf}, "capability": {"opaque-preview-capability"},
	})
	if resp.StatusCode != http.StatusBadRequest || commander.applies != 0 {
		t.Fatalf("unchecked apply status=%d applies=%d", resp.StatusCode, commander.applies)
	}
	resp = h.post(t, "/optimization/exit-protection/apply", url.Values{
		"csrf": {h.csrf}, "capability": {"opaque-preview-capability"}, "confirm": {"yes"},
	})
	if resp.StatusCode != http.StatusTooEarly || commander.applies != 0 {
		t.Fatalf("early apply status=%d applies=%d", resp.StatusCode, commander.applies)
	}
	h.clock.advance(3 * time.Second)
	resp = h.post(t, "/optimization/exit-protection/apply", url.Values{
		"csrf": {h.csrf}, "capability": {"opaque-preview-capability"}, "confirm": {"yes"},
	})
	if resp.StatusCode != http.StatusOK || commander.applies != 1 {
		t.Fatalf("mature apply status=%d applies=%d", resp.StatusCode, commander.applies)
	}
}

func TestExitProtectionCSRFFailsBeforeCommander(t *testing.T) {
	commander := &fakeProtectionCommander{status: protectionStatus(), clock: time.Now}
	h := newDashboardHarness(t, func(options *Options) { options.Protections = commander })
	h.authenticate(t)
	resp := h.post(t, "/optimization/exit-protection/preview", url.Values{"action_token": {"opaque-row-action"}})
	if resp.StatusCode != http.StatusForbidden || commander.previews != 0 {
		t.Fatalf("csrf status=%d previews=%d", resp.StatusCode, commander.previews)
	}
}

func protectionSection(t *testing.T, page string) string {
	t.Helper()
	match := regexp.MustCompile(`(?s)<section id="exit-protection".*?</section>`).FindString(page)
	if match == "" {
		t.Fatal("exit-protection section not found")
	}
	return match
}
