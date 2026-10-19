package console

import (
	"context"
	"database/sql"
	"path/filepath"
	"reflect"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	strategyopt "github.com/JungHoonGhae/tossinvest-cli/internal/optimization"
	"net/http"
	"net/url"
	"strings"
	"testing"
)

func TestA154ComputeOnlyHTTP(t *testing.T) {
	h := newDashboardHarness(t)
	path := "/optimization/hypothetical"
	unauth := h.post(t, path, url.Values{})
	if unauth.StatusCode != http.StatusForbidden || unauth.Header.Get("Cache-Control") != "no-store" {
		t.Fatalf("unauth %d headers %v", unauth.StatusCode, unauth.Header)
	}
	h.authenticate(t)
	f := url.Values{"csrf": {h.csrf}, "market": {"KR"}, "entry": {"70000"}, "low": {"69000"}, "atr": {"500"}, "tick": {"100"}, "min": {"60000"}, "max": {"80000"}, "budget": {"10000"}, "max_shares": {"100"}, "max_notional": {"7000000"}}
	r := h.post(t, path, f)
	page := body(t, r)
	if r.StatusCode != 200 || r.Header.Get("Cache-Control") != "no-store" || !strings.Contains(page, "69500") || !strings.Contains(page, "1400000") {
		t.Fatalf("result %d %s", r.StatusCode, page)
	}
	if strings.Count(page, "<main>") != 1 || strings.Count(page, "</main>") != 1 {
		t.Fatal("result must have one main landmark")
	}
	cases := []struct {
		path   string
		form   url.Values
		status int
	}{
		{path + "?entry=1", f, 400}, {path, url.Values{"csrf": {h.csrf, h.csrf}}, 400}, {path, url.Values{"csrf": {h.csrf}, "extra": {"1"}}, 400}, {path, url.Values{"csrf": {"wrong"}}, 403}, {path, url.Values{"csrf": {h.csrf}, "entry": {strings.Repeat("1", 9000)}}, 413},
	}
	for _, tc := range cases {
		r := h.post(t, tc.path, tc.form)
		if r.StatusCode != tc.status || r.Header.Get("Cache-Control") != "no-store" {
			t.Errorf("%s status %d want %d", tc.path, r.StatusCode, tc.status)
		}
	}
	r = h.get(t, path)
	if r.StatusCode != 405 || r.Header.Get("Cache-Control") != "no-store" {
		t.Errorf("GET %d", r.StatusCode)
	}
}

func TestA154NoAuthorityAndForgedResearchKey(t *testing.T) {
	commander := newFakeOptimizationCommander(t)
	legacy := &fakeExitPolicySettings{}
	h := newDashboardHarness(t, func(o *Options) { o.Optimization = commander; o.ExitPolicies = legacy })
	h.authenticate(t)
	f := url.Values{"csrf": {h.csrf}, "market": {"US"}, "entry": {"100"}, "tick": {"0.01"}, "min": {"1"}, "max": {"200"}, "budget": {"100"}, "max_shares": {"1000"}, "max_notional": {"100000"}}
	before := h.holdings.count()
	r := h.post(t, "/optimization/hypothetical", f)
	if r.StatusCode != 200 {
		t.Fatalf("calculator status %d", r.StatusCode)
	}
	if h.holdings.count() != before || commander.previews != 0 || commander.applies != 0 || legacy.saves != 0 {
		t.Fatal("calculator reached operational authority")
	}
	r = h.post(t, "/optimization/exit-policy", url.Values{"csrf": {h.csrf}, "base_version": {"12"}, "category": {"exit-protection"}, "setting_key": {"fixed_007"}, "option_id": {"fixed_007"}})
	if r.StatusCode != 400 || commander.previews != 0 || commander.applies != 0 || legacy.saves != 0 {
		t.Fatalf("forged key status=%d, authority reached", r.StatusCode)
	}
}

func TestA154MarketErrorIsLinkedToControl(t *testing.T) {
	h := newDashboardHarness(t)
	h.authenticate(t)
	r := h.post(t, "/optimization/hypothetical", url.Values{"csrf": {h.csrf}, "market": {"OTHER"}})
	page := body(t, r)
	for _, v := range []string{`aria-invalid="true"`, `aria-describedby="hyp-error-market"`, `id="hyp-error-market"`, `시장을 선택하세요.`} {
		if !strings.Contains(page, v) {
			t.Errorf("missing %q", v)
		}
	}
}

func TestA154ActualOptimizationDatabaseUnchanged(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), strategyopt.DatabaseFileName)
	registry, err := strategyopt.CoreRegistry(ctx)
	if err != nil {
		t.Fatal(err)
	}
	clk := clock.NewFake(time.Date(2026, 8, 1, 9, 0, 0, 0, time.UTC))
	store, err := strategyopt.Open(ctx, strategyopt.Options{Path: path, Registry: registry, Clock: clk, Actor: "a154:test"})
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	db, err := sql.Open("sqlite", "file:"+path+"?mode=ro")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	counts := func() []int {
		out := []int{}
		for _, table := range []string{"optimization_control", "optimization_snapshots", "optimization_candidates", "optimization_applications", "optimization_audit"} {
			var n int
			if err := db.QueryRow("SELECT COUNT(*) FROM " + table).Scan(&n); err != nil {
				t.Fatal(err)
			}
			out = append(out, n)
		}
		return out
	}
	before, err := store.Read(ctx)
	if err != nil {
		t.Fatal(err)
	}
	c0 := counts()
	h := newDashboardHarness(t, func(o *Options) { o.Optimization = store; o.Now = clk.Now })
	h.authenticate(t)
	response := h.post(t, "/optimization/hypothetical", url.Values{"csrf": {h.csrf}, "market": {"US"}, "entry": {"100"}, "tick": {"0.01"}, "min": {"1"}, "max": {"200"}})
	if response.StatusCode != 200 {
		t.Fatalf("compute %d", response.StatusCode)
	}
	response = h.post(t, "/optimization/exit-policy", url.Values{"csrf": {h.csrf}, "base_version": {"1"}, "category": {"exit-protection"}, "setting_key": {"fixed_007"}, "option_id": {"fixed_007"}})
	if response.StatusCode != 400 {
		t.Fatalf("forged %d", response.StatusCode)
	}
	after, err := store.Read(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(before.Snapshot, after.Snapshot) || !reflect.DeepEqual(before.Audit, after.Audit) || !reflect.DeepEqual(c0, counts()) {
		t.Fatal("real optimization store changed")
	}
}
