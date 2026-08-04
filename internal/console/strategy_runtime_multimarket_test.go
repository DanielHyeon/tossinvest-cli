package console

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyprojection"
)

type multiMarketRuntimeStub struct {
	snapshot strategyprojection.Snapshot
	err      error
}

func (s multiMarketRuntimeStub) Read(context.Context) (strategyprojection.Snapshot, error) {
	return s.snapshot, s.err
}

func TestStrategyRuntimeDormantPairIsHonest(t *testing.T) {
	h := newHarness(t)
	h.authenticate(t)
	page := body(t, h.get(t, "/strategy-runtime"))
	for _, want := range []string{"KR strategy runtime", "US strategy runtime", "OFF", "UNOBSERVED", "UNWIRED", "NOT_CONFIGURED", "관측 없음"} {
		if !strings.Contains(page, want) {
			t.Errorf("dormant page missing %q", want)
		}
	}
	if strings.Contains(page, "한국 주식 전략 lane") || strings.Contains(page, "entry-ready") {
		t.Fatal("single-market/ready copy survived")
	}
	if mutations := h.broker.mutationCount(); mutations != 0 {
		t.Fatalf("dormant strategy health made %d broker mutations", mutations)
	}
}

func TestStrategyRuntimeMarketsRenderIndependently(t *testing.T) {
	snapshot := consoleProjectionPair(t)
	snapshot = strategyprojection.WithMarketFailure(snapshot, strategyprojection.MarketUS,
		strategyprojection.RefusalRuntimeUnavailable, consoleProjectionNow.Add(time.Second))
	h := newHarness(t, func(options *Options) { options.StrategyRuntime = multiMarketRuntimeStub{snapshot: snapshot} })
	h.authenticate(t)
	page := body(t, h.get(t, "/strategy-runtime"))
	for _, want := range []string{
		`data-market="KR"`, `data-market-status="CURRENT"`, "evidence-KR", "campaign-KR", "leg-1", "SHORT", "HEALTHY", "WIRED",
		`data-market="US"`, `data-market-status="UNKNOWN"`, "RUNTIME_UNAVAILABLE", "UNWIRED",
	} {
		if !strings.Contains(page, want) {
			t.Errorf("partial page missing %q", want)
		}
	}
	if strings.Count(page, "evidence-KR") != 1 || strings.Contains(page, "evidence-US") {
		t.Fatal("peer evidence crossed into unavailable market")
	}
}

func TestStrategyRuntimeStatusIsAuthenticatedGETOnlyAndHasNoMutationInput(t *testing.T) {
	h := newHarness(t, func(options *Options) {
		options.StrategyRuntime = multiMarketRuntimeStub{snapshot: consoleProjectionPair(t)}
	})
	if h.get(t, "/strategy-runtime").StatusCode == http.StatusOK {
		t.Fatal("unauthenticated request reached page")
	}
	h.authenticate(t)
	for _, method := range []string{http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete} {
		req, _ := http.NewRequest(method, h.srv.URL+"/strategy-runtime", nil)
		response, err := h.client.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		_ = response.Body.Close()
		if response.StatusCode != http.StatusMethodNotAllowed || response.Header.Get("Allow") != "GET, HEAD" {
			t.Errorf("%s status=%d allow=%q", method, response.StatusCode, response.Header.Get("Allow"))
		}
	}
	response := h.get(t, "/strategy-runtime")
	page := strings.ToLower(body(t, response))
	for _, forbidden := range []string{"<form", "<input", "<textarea", "<select", "<button", "contenteditable", "place order"} {
		if strings.Contains(page, forbidden) {
			t.Errorf("forbidden surface %q", forbidden)
		}
	}
	for _, want := range []string{`name="viewport"`, "@media (max-width: 720px)", "detail-grid", "status-pill",
		`href="/optimization"`, `href="/strategy-runtime/market-schedule"`, `href="/strategy-runtime">상태 새로고침`,
		"a:focus-visible", `aria-readonly="true"`} {
		if !strings.Contains(page, want) {
			t.Errorf("responsive/read-only surface missing %q", want)
		}
	}
	if strings.Contains(page, "<table") {
		t.Error("wide runtime table returned; cards/dl are required for 360px")
	}
	assertMarketScheduleDOM(t, page)
	if got := response.Header.Get("Content-Security-Policy"); got != "default-src 'none'; style-src 'unsafe-inline'; form-action 'self'; frame-ancestors 'none'; base-uri 'none'" {
		t.Errorf("CSP=%q", got)
	}
}

func TestStrategyRuntimeReaderFailureAndInvalidProjectionFailClosedWithoutLeakingError(t *testing.T) {
	invalid := consoleProjectionPair(t)
	delete(invalid.Markets, strategyprojection.MarketUS)
	for _, test := range []struct {
		name   string
		reader MultiMarketStrategyRuntimeReader
	}{
		{name: "reader error", reader: multiMarketRuntimeStub{err: context.Canceled}},
		{name: "invalid paired projection", reader: multiMarketRuntimeStub{snapshot: invalid}},
	} {
		t.Run(test.name, func(t *testing.T) {
			h := newHarness(t, func(options *Options) { options.StrategyRuntime = test.reader })
			h.authenticate(t)
			page := body(t, h.get(t, "/strategy-runtime"))
			for _, want := range []string{"runtime projection을 읽지 못했다", `data-market="KR" data-market-status="UNKNOWN"`,
				`data-market="US" data-market-status="UNKNOWN"`, "RUNTIME_UNAVAILABLE", "OFF", "UNWIRED"} {
				if !strings.Contains(page, want) {
					t.Errorf("fail-closed page missing %q", want)
				}
			}
			if strings.Contains(page, context.Canceled.Error()) {
				t.Fatal("raw reader error reached operator page")
			}
		})
	}
}

func TestStrategyRuntimeProjectsAuthorityFactsWithoutRecalculation(t *testing.T) {
	snapshot := consoleProjectionPair(t)
	kr := snapshot.Markets[strategyprojection.MarketKR]
	kr.Reconciliation = strategyprojection.ReconciliationProjection{Status: strategyprojection.ReconciliationUnknown,
		Refusal: strategyprojection.RefusalReconciliationUnavailable}
	kr.FirstRefusal = strategyprojection.RefusalReconciliationUnavailable
	snapshot.Markets[strategyprojection.MarketKR] = kr
	if err := strategyprojection.Validate(snapshot); err != nil {
		t.Fatal(err)
	}
	h := newHarness(t, func(options *Options) { options.StrategyRuntime = multiMarketRuntimeStub{snapshot: snapshot} })
	h.authenticate(t)
	page := body(t, h.get(t, "/strategy-runtime"))
	for _, want := range []string{"lane-KR", "ON → ON", "RECONCILIATION_UNAVAILABLE", "UNKNOWN"} {
		if !strings.Contains(page, want) {
			t.Errorf("authority projection missing %q", want)
		}
	}
}

func TestOptimizationNavigationLinksPairedRuntimeAndSchedule(t *testing.T) {
	h := newHarness(t)
	h.authenticate(t)
	page := body(t, h.get(t, "/optimization"))
	for _, want := range []string{`href="/strategy-runtime"`, `href="/strategy-runtime/market-schedule"`} {
		if !strings.Contains(page, want) {
			t.Errorf("optimization navigation missing %q", want)
		}
	}
}

func TestStrategyRuntimeSummaryUsesPairedDormantTruth(t *testing.T) {
	h := newHarness(t)
	got := h.strategyRuntimeSummary(httptest.NewRequest(http.MethodGet, "/settings", nil))
	if got != "KR OFF/UNKNOWN · US OFF/UNKNOWN — dormant 미배선" {
		t.Fatalf("summary=%q", got)
	}
}

func TestStrategyRuntimeSummaryReportsReadFailure(t *testing.T) {
	h := newHarness(t, func(options *Options) { options.StrategyRuntime = multiMarketRuntimeStub{err: context.Canceled} })
	got := h.strategyRuntimeSummary(httptest.NewRequest(http.MethodGet, "/settings", nil))
	if got != "읽지 못함 — 전략 lane 판독이 유효하지 않다" || strings.Contains(got, context.Canceled.Error()) {
		t.Fatalf("summary=%q", got)
	}
}

func TestStrategyRuntimeSummaryKeepsMarketsIndependent(t *testing.T) {
	snapshot := strategyprojection.WithMarketFailure(consoleProjectionPair(t), strategyprojection.MarketUS,
		strategyprojection.RefusalRuntimeUnavailable, consoleProjectionNow.Add(time.Second))
	h := newHarness(t, func(options *Options) { options.StrategyRuntime = multiMarketRuntimeStub{snapshot: snapshot} })
	got := h.strategyRuntimeSummary(httptest.NewRequest(http.MethodGet, "/settings", nil))
	if got != "KR ON→ON (NONE) · US OFF→OFF (RUNTIME_UNAVAILABLE)" {
		t.Fatalf("summary=%q", got)
	}
}

var consoleProjectionNow = time.Date(2026, 8, 4, 12, 0, 0, 0, time.UTC)

func consoleProjectionPair(t *testing.T) strategyprojection.Snapshot {
	t.Helper()
	snapshot := strategyprojection.DormantSnapshot(consoleProjectionNow)
	for _, market := range []strategyprojection.Market{strategyprojection.MarketKR, strategyprojection.MarketUS} {
		observed := consoleProjectionNow.Add(-time.Second)
		evidenceID, evidenceDigest := "evidence-"+string(market), consoleDigest("evidence-"+string(market))
		laneID, laneVersion := "lane-"+string(market), "v1"
		campaignID, legID := "campaign-"+string(market), "leg-1"
		bucket, riskVersion := "SHORT", "risk-v1"
		calendarSource, calendarVersion := "official-"+string(market), "2026.08"
		activationDigest := consoleDigest("activation-" + string(market))
		snapshot.Markets[market] = strategyprojection.MarketProjection{Market: market, Status: strategyprojection.StatusCurrent,
			Lane:        strategyprojection.LaneProjection{ID: &laneID, Version: &laneVersion, Desired: strategyprojection.StateOn, Effective: strategyprojection.StateOn},
			Evidence:    strategyprojection.EvidenceProjection{ID: &evidenceID, Digest: &evidenceDigest, Freshness: strategyprojection.FreshnessCurrent},
			Campaign:    strategyprojection.CampaignProjection{ID: &campaignID, LegID: &legID},
			HorizonRisk: strategyprojection.HorizonRiskProjection{Bucket: &bucket, PolicyVersion: &riskVersion, Status: strategyprojection.ComponentCurrent},
			Scheduler: strategyprojection.SchedulerProjection{Desired: strategyprojection.StateOn, Effective: strategyprojection.StateOn,
				CalendarSource: &calendarSource, CalendarVersion: &calendarVersion, CalendarFreshness: strategyprojection.FreshnessCurrent},
			Activation: strategyprojection.ActivationProjection{Desired: strategyprojection.StateOn, Effective: strategyprojection.StateOn,
				Digest: &activationDigest, Status: strategyprojection.ActivationConfigured},
			Protection:     strategyprojection.ProtectionProjection{Readiness: strategyprojection.ProtectionWired, Refusal: strategyprojection.RefusalNone},
			Reconciliation: strategyprojection.ReconciliationProjection{Status: strategyprojection.ReconciliationHealthy, Refusal: strategyprojection.RefusalNone},
			FirstRefusal:   strategyprojection.RefusalNone, ObservedAt: &observed}
	}
	if err := strategyprojection.Validate(snapshot); err != nil {
		t.Fatal(err)
	}
	return snapshot
}

func consoleDigest(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}
