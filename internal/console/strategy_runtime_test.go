package console

import (
	"context"
	"net/http"
	"strings"
	"testing"
)

type strategyRuntimeStub struct {
	reading StrategyRuntimeReading
	err     error
}

func (s strategyRuntimeStub) Read(context.Context) (StrategyRuntimeReading, error) {
	return s.reading, s.err
}

func TestStrategyRuntimeStatusShowsSeparatedDormantStateAndBlockers(t *testing.T) {
	h := newHarness(t, func(o *Options) {
		o.StrategyRuntime = strategyRuntimeStub{reading: StrategyRuntimeReading{LaneDesired: true, AutoStartDesired: true, GateApproved: true, LiveApproved: true, Reason: "source_manifest_unavailable"}}
	})
	h.authenticate(t)
	page := body(t, h.get(t, "/strategy-runtime"))
	for _, want := range []string{"전략 파라미터", "lane 상태", "자동 기동", "LIVE 주문 승인", "0.08", "0.25", "1.2", "1.5", "0.35", "1.8", "0.7", "3.0", "10", "15", "0.20", "d75113d3", "09260ac…", "UNWIRED", "READ_ONLY", "NOT_CONFIGURED", "source_manifest_unavailable", "신규 entry가 OFF여도 exit·reconcile·보호 감독은 계속된다"} {
		if !strings.Contains(page, want) {
			t.Errorf("page missing %q", want)
		}
	}
}
func TestStrategyRuntimeStatusIsAuthenticatedReadOnlyAndHasNoControls(t *testing.T) {
	h := newHarness(t, func(o *Options) { o.StrategyRuntime = strategyRuntimeStub{} })
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
		resp.Body.Close()
		if resp.StatusCode != http.StatusMethodNotAllowed {
			t.Errorf("%s=%d", method, resp.StatusCode)
		}
	}
	resp := h.get(t, "/strategy-runtime")
	page := strings.ToLower(body(t, resp))
	for _, forbidden := range []string{"<form", "<input", "<textarea", "<select", "<button", "contenteditable", "type=\"text\"", "type=\"number\"", "type=\"range\"", "enable all", "typed confirmation"} {
		if strings.Contains(page, forbidden) {
			t.Errorf("forbidden surface %q", forbidden)
		}
	}
	assertMarketScheduleDOM(t, page)
	if got := resp.Header.Get("Content-Security-Policy"); got != "default-src 'none'; style-src 'unsafe-inline'; form-action 'self'; frame-ancestors 'none'; base-uri 'none'" {
		t.Errorf("CSP=%q", got)
	}
}
func TestStrategyRuntimeStatusFailsClosedOnNilOrReaderError(t *testing.T) {
	for _, configure := range []func(*Options){func(*Options) {}, func(o *Options) { o.StrategyRuntime = strategyRuntimeStub{err: context.Canceled} }} {
		h := newHarness(t, configure)
		h.authenticate(t)
		page := body(t, h.get(t, "/strategy-runtime"))
		for _, want := range []string{"OFF", "NOT_CONFIGURED", "source_manifest_unavailable"} {
			if !strings.Contains(page, want) {
				t.Errorf("fallback missing %q", want)
			}
		}
	}
}
