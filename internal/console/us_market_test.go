package console

// us_market_test.go covers verify-us-market's console half: the screen is scoped
// to a market, and one market's verdicts never speak for another's.

import (
	"net/url"
	"strings"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/verifylive"
)

// TestTheVerifyScreenReadsTheMarketFromTheQuery.
func TestTheVerifyScreenReadsTheMarketFromTheQuery(t *testing.T) {
	h := newHarness(t)
	seedVerdicts(t, usRecordPath(h.record), verifylive.VerdictPass, map[verifylive.StepID]verifylive.Verdict{
		verifylive.StepOrderCancel: verifylive.VerdictFail,
	})
	h.authenticate(t)

	us := body(t, h.get(t, "/verify?market=us"))
	if !strings.Contains(us, "재측정 1단계") {
		t.Errorf("the US screen does not read the US record:\n%s", truncateForLog(us))
	}

	// The KR screen is untouched by everything above: its record is still empty.
	kr := body(t, h.get(t, "/verify"))
	if strings.Contains(kr, "재측정 1단계") {
		t.Errorf("the KR screen is showing the US record's verdicts:\n%s", truncateForLog(kr))
	}
}

// TestOneMarketsVerdictsDoNotSettleAnother.
//
// The failure this prevents is the worst one this screen can produce: a market
// reported as measured because a different market was.
func TestOneMarketsVerdictsDoNotSettleAnother(t *testing.T) {
	h := newHarness(t)
	seedVerdicts(t, h.record, verifylive.VerdictPass, nil)
	h.authenticate(t)

	us := body(t, h.get(t, "/verify?market=us"))
	if strings.Contains(us, "이어할 단계가 없다") {
		t.Errorf("the US screen thinks the KR record settled its steps:\n%s", truncateForLog(us))
	}
	if !strings.Contains(us, "검증 시작") {
		t.Errorf("the US screen does not offer a first run:\n%s", truncateForLog(us))
	}
}

// TestTheStartFormCarriesTheMarket — the run is fixed to the market it was
// started for, and the screen keeps showing that market afterwards.
func TestTheStartFormCarriesTheMarket(t *testing.T) {
	h := newHarness(t)
	h.authenticate(t)

	h.post(t, "/verify/start", url.Values{"csrf": {h.csrf}, "market": {"us"}})
	view := h.waitForBatch(t)

	if view.Market != verifylive.MarketUS {
		t.Fatalf("the run's market is %q, want %q", view.Market, verifylive.MarketUS)
	}
}

// TestAnUnknownMarketIsKR — the screen's default is the record every existing
// verdict already lives in.
func TestAnUnknownMarketIsKR(t *testing.T) {
	h := newHarness(t)
	seedVerdicts(t, h.record, verifylive.VerdictPass, nil)
	h.authenticate(t)

	page := body(t, h.get(t, "/verify?market=mars"))
	if !strings.Contains(page, "이어할 단계가 없다") {
		t.Errorf("an unrecognised market did not fall back to the KR record:\n%s", truncateForLog(page))
	}
}

// TestTheUSScreenCarriesTheUSAdvisory — and does not present KR's measured
// closed-market code as if it were the US market's.
func TestTheUSScreenCarriesTheUSAdvisory(t *testing.T) {
	h := newHarness(t)
	h.authenticate(t)

	page := body(t, h.get(t, "/verify?market=us"))
	if !strings.Contains(page, "US 정규장") {
		t.Errorf("the US screen does not report the US session:\n%s", truncateForLog(page))
	}
	if strings.Contains(page, "order-hours-closed") {
		t.Errorf("the US screen quotes the KR measurement:\n%s", truncateForLog(page))
	}
}

// TestTheStartFormsOnTheUSScreenCarryTheUSMarket.
//
// The screen and the button have to agree: a [재측정] pressed on the US screen
// that started a KR run would place KR orders for a US measurement, and the
// operator would have approved a plan for the market they were not looking at.
func TestTheStartFormsOnTheUSScreenCarryTheUSMarket(t *testing.T) {
	h := newHarness(t)
	seedVerdicts(t, usRecordPath(h.record), verifylive.VerdictPass, map[verifylive.StepID]verifylive.Verdict{
		verifylive.StepOrderCancel: verifylive.VerdictFail,
	})
	h.authenticate(t)

	page := body(t, h.get(t, "/verify?market=us"))
	for _, form := range strings.Split(page, "<form") {
		if !strings.Contains(form, `action="/verify/start"`) {
			continue
		}
		if !strings.Contains(form, `name="market" value="US"`) {
			t.Errorf("a start form on the US screen does not carry the market:\n%s", form)
		}
	}
	// And the KR screen carries KR, so the same form cannot drift into a default.
	kr := body(t, h.get(t, "/verify"))
	for _, form := range strings.Split(kr, "<form") {
		if !strings.Contains(form, `action="/verify/start"`) {
			continue
		}
		if !strings.Contains(form, `name="market" value="KR"`) {
			t.Errorf("a start form on the KR screen does not carry the market:\n%s", form)
		}
	}
}

// TestTheVerifyScreenOffersTheOtherMarket — the switch itself.
func TestTheVerifyScreenOffersTheOtherMarket(t *testing.T) {
	h := newHarness(t)
	h.authenticate(t)

	kr := body(t, h.get(t, "/verify"))
	if !strings.Contains(kr, "/verify?market=US") {
		t.Errorf("the KR screen offers no way to reach the US one:\n%s", truncateForLog(kr))
	}
	us := body(t, h.get(t, "/verify?market=us"))
	if !strings.Contains(us, "/verify?market=KR") {
		t.Errorf("the US screen offers no way back:\n%s", truncateForLog(us))
	}
}
