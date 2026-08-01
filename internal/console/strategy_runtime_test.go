package console

import (
	"context"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyengine"
)

type strategyRuntimeStub struct {
	reading StrategyRuntimeReading
	err     error
}

func (s strategyRuntimeStub) Read(context.Context) (StrategyRuntimeReading, error) {
	return s.reading, s.err
}

func validStrategyRuntimeFixture() StrategyRuntimeReading {
	reading := dormantStrategyRuntimeReading(time.Date(2026, 8, 1, 1, 2, 3, 0, time.UTC))
	reading.ObservedAt = time.Date(2026, 8, 1, 1, 2, 0, 0, time.UTC)
	reading.Freshness = strategyengine.RuntimeStateVerified
	return reading
}

func TestStrategyRuntimeStatusRendersServerDescriptorAndCompleteBlockers(t *testing.T) {
	reading := validStrategyRuntimeFixture()
	reading.Descriptor.Fields[0].Help = "server-owned custom help"
	reading.Descriptor.Fields[0].Desired = "server-owned desired"
	h := newHarness(t, func(o *Options) { o.StrategyRuntime = strategyRuntimeStub{reading: reading} })
	h.authenticate(t)
	page := body(t, h.get(t, "/strategy-runtime"))
	for _, want := range []string{
		"전략 파라미터", "lane 상태", "자동 기동", "LIVE 주문 승인",
		"server-owned custom help", "server-owned desired", "symbol_state_stale_seconds", "30",
		strategyengine.SourceCommit, strategyengine.FrozenSourceSetDigest,
		"StockOS source manifest", "a046 후보 provenance", "a048 scheduler/calendar",
		"a045 ProtectionReady", "Guardian 승인", "Reconciliation health", "Operating mode",
		"Kill switch", "Activation manifest", "Desired", "Effective", "Freshness", "Reason",
		"2026-08-01T01:02:03Z", "2026-08-01T01:02:00Z",
		"신규 entry가 OFF여도 exit·reconcile·보호 감독은 계속된다",
	} {
		if !strings.Contains(page, want) {
			t.Errorf("page missing %q", want)
		}
	}
}

func TestStrategyRuntimeProjectsAuthorityEntryCapabilityWithoutRecalculation(t *testing.T) {
	reading := validStrategyRuntimeFixture()
	reading.Lane = StrategyRuntimeControl{
		Default: strategyengine.RuntimeStateOff, Desired: strategyengine.RuntimeStateOn,
		Effective: strategyengine.RuntimeStateOn, Reason: strategyengine.RuntimeRefusalNone,
	}
	reading.AutoStart = reading.Lane
	reading.GateApproval = StrategyRuntimeControl{
		Default: strategyengine.RuntimeStateUnapproved, Desired: strategyengine.RuntimeStateVerified,
		Effective: strategyengine.RuntimeStateVerified, Reason: strategyengine.RuntimeRefusalNone,
	}
	reading.LiveApproval = reading.GateApproval
	// Even though all four control snapshots are effective, the authority's
	// ordered blocker result remains OFF. The console must not replace it with ON.
	reading.EntryCapability = StrategyRuntimeEntryCapability{
		Default: strategyengine.RuntimeStateOff, Desired: strategyengine.RuntimeStateOn,
		Effective:    strategyengine.RuntimeStateOff,
		FirstRefusal: strategyengine.RuntimeRefusalReconciliationUnhealthy,
	}
	reading.Blockers[5].Effective = strategyengine.RuntimeStateRefused
	reading.Blockers[5].Freshness = strategyengine.RuntimeStateStale

	h := newHarness(t, func(o *Options) { o.StrategyRuntime = strategyRuntimeStub{reading: reading} })
	h.authenticate(t)
	page := body(t, h.get(t, "/strategy-runtime"))
	for _, want := range []string{
		`data-testid="lane-effective">ON</span>`,
		`data-testid="autostart-effective">ON</span>`,
		`data-testid="gate-effective">VERIFIED</span>`,
		`data-testid="live-effective">VERIFIED</span>`,
		`data-testid="entry-effective">OFF</span>`,
		"reconciliation_unhealthy",
	} {
		if !strings.Contains(page, want) {
			t.Errorf("authority projection missing %q", want)
		}
	}
}

func TestStrategyRuntimeRejectsSemanticallyImpossibleSnapshots(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*StrategyRuntimeReading)
	}{
		{name: "lane cannot be live mode", mutate: func(r *StrategyRuntimeReading) { r.Lane.Effective = strategyengine.RuntimeStateLive }},
		{name: "approval cannot be toggle on", mutate: func(r *StrategyRuntimeReading) { r.GateApproval.Desired = strategyengine.RuntimeStateOn }},
		{name: "entry cannot be verified", mutate: func(r *StrategyRuntimeReading) { r.EntryCapability.Effective = strategyengine.RuntimeStateVerified }},
		{name: "entry default cannot be approval", mutate: func(r *StrategyRuntimeReading) { r.EntryCapability.Default = strategyengine.RuntimeStateUnapproved }},
		{name: "freshness cannot be on", mutate: func(r *StrategyRuntimeReading) { r.Freshness = strategyengine.RuntimeStateOn }},
		{name: "blocker freshness cannot be healthy", mutate: func(r *StrategyRuntimeReading) { r.Blockers[0].Freshness = strategyengine.RuntimeStateHealthy }},
		{name: "generated time required", mutate: func(r *StrategyRuntimeReading) { r.GeneratedAt = time.Time{} }},
		{name: "observed time required when verified", mutate: func(r *StrategyRuntimeReading) { r.ObservedAt = time.Time{} }},
		{name: "generated cannot precede observed", mutate: func(r *StrategyRuntimeReading) { r.GeneratedAt = r.ObservedAt.Add(-time.Second) }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reading := validStrategyRuntimeFixture()
			tt.mutate(&reading)
			h := newHarness(t, func(o *Options) { o.StrategyRuntime = strategyRuntimeStub{reading: reading} })
			h.authenticate(t)
			page := body(t, h.get(t, "/strategy-runtime"))
			for _, want := range []string{
				"상태를 읽지 못해 신규 진입 OFF를 표시한다",
				`data-testid="entry-effective">OFF</span>`,
				"runtime_read_failed",
			} {
				if !strings.Contains(page, want) {
					t.Errorf("fail-closed page missing %q", want)
				}
			}
		})
	}
}

func TestStrategyRuntimeRejectsMalformedDescriptorOrderingWithoutPanic(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*StrategyRuntimeReading)
	}{
		{name: "reordered sections", mutate: func(r *StrategyRuntimeReading) {
			r.Descriptor.Sections[0], r.Descriptor.Sections[1] = r.Descriptor.Sections[1], r.Descriptor.Sections[0]
		}},
		{name: "duplicate sections", mutate: func(r *StrategyRuntimeReading) {
			r.Descriptor.Sections[1] = r.Descriptor.Sections[0]
		}},
		{name: "reordered fields", mutate: func(r *StrategyRuntimeReading) {
			r.Descriptor.Fields[0], r.Descriptor.Fields[1] = r.Descriptor.Fields[1], r.Descriptor.Fields[0]
		}},
		{name: "duplicate blockers", mutate: func(r *StrategyRuntimeReading) {
			r.Blockers[1] = r.Blockers[0]
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reading := validStrategyRuntimeFixture()
			tt.mutate(&reading)
			h := newHarness(t, func(o *Options) { o.StrategyRuntime = strategyRuntimeStub{reading: reading} })
			h.authenticate(t)
			resp := h.get(t, "/strategy-runtime")
			if resp.StatusCode != http.StatusOK {
				t.Fatalf("GET status=%d", resp.StatusCode)
			}
			page := body(t, resp)
			for _, want := range []string{"runtime_read_failed", `data-testid="entry-effective">OFF</span>`} {
				if !strings.Contains(page, want) {
					t.Errorf("malformed descriptor fallback missing %q", want)
				}
			}
		})
	}
}

func TestStrategyRuntimeStatusIsAuthenticatedGETOnlyAndHasNoInputSurface(t *testing.T) {
	h := newHarness(t, func(o *Options) { o.StrategyRuntime = strategyRuntimeStub{reading: validStrategyRuntimeFixture()} })
	if h.get(t, "/strategy-runtime").StatusCode == http.StatusOK {
		t.Fatal("unauthenticated request reached page")
	}
	h.authenticate(t)
	for _, method := range []string{http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete} {
		req, _ := http.NewRequest(method, h.srv.URL+"/strategy-runtime", nil)
		resp, err := h.client.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		_ = resp.Body.Close()
		if resp.StatusCode != http.StatusMethodNotAllowed {
			t.Errorf("%s=%d", method, resp.StatusCode)
		}
		if got := resp.Header.Get("Allow"); got != "GET, HEAD" {
			t.Errorf("%s Allow=%q", method, got)
		}
	}
	resp := h.get(t, "/strategy-runtime")
	page := strings.ToLower(body(t, resp))
	for _, forbidden := range []string{
		"<form", "<input", "<textarea", "<select", "<button", "contenteditable",
		"type=\"text\"", "type=\"number\"", "type=\"range\"", "enable all", "typed confirmation",
	} {
		if strings.Contains(page, forbidden) {
			t.Errorf("forbidden surface %q", forbidden)
		}
	}
	for _, want := range []string{
		`name="viewport"`, "@media (max-width: 720px)", "detail-grid", "status-pill",
		`href="/optimization"`, `href="/strategy-runtime/market-schedule"`,
		`href="/strategy-runtime">상태 새로고침`, "a:focus-visible", "aria-readonly=\"true\"",
	} {
		if !strings.Contains(page, want) {
			t.Errorf("responsive/read-only surface missing %q", want)
		}
	}
	if strings.Contains(page, "<table") {
		t.Error("wide runtime table returned; cards/dl are required for 360px")
	}
	assertMarketScheduleDOM(t, page)
	if got := resp.Header.Get("Content-Security-Policy"); got != "default-src 'none'; style-src 'unsafe-inline'; form-action 'self'; frame-ancestors 'none'; base-uri 'none'" {
		t.Errorf("CSP=%q", got)
	}
}

func TestStrategyRuntimeStatusFailsClosedOnUnwiredErrorAndInvalidSnapshot(t *testing.T) {
	tests := []struct {
		name      string
		configure func(*Options)
		want      string
	}{
		{name: "unwired", configure: func(*Options) {}, want: "source_manifest_unavailable"},
		{name: "reader error", configure: func(o *Options) { o.StrategyRuntime = strategyRuntimeStub{err: context.Canceled} }, want: "runtime_read_failed"},
		{name: "invalid snapshot", configure: func(o *Options) { o.StrategyRuntime = strategyRuntimeStub{} }, want: "runtime_read_failed"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHarness(t, tt.configure)
			h.authenticate(t)
			page := body(t, h.get(t, "/strategy-runtime"))
			for _, want := range []string{"OFF", "NOT_CONFIGURED", "UNAPPROVED", "UNOBSERVED", tt.want} {
				if !strings.Contains(page, want) {
					t.Errorf("fallback missing %q", want)
				}
			}
			if strings.Contains(page, context.Canceled.Error()) {
				t.Error("raw reader error reached operator page")
			}
		})
	}
}

func TestOptimizationNavigationLinksRuntimeAndSchedule(t *testing.T) {
	h := newHarness(t)
	h.authenticate(t)
	page := body(t, h.get(t, "/optimization"))
	for _, want := range []string{`href="/strategy-runtime"`, `href="/strategy-runtime/market-schedule"`} {
		if !strings.Contains(page, want) {
			t.Errorf("optimization navigation missing %q", want)
		}
	}
}
