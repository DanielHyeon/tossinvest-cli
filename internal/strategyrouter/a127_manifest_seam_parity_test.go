//go:build tossos_testseams

package strategyrouter

import (
	"reflect"
	"testing"
)

// a127 D4 — 외부 시험용 매니페스트 seam 의 본문이 내부 픽스처와 같음(둘째 철자 감시).
func TestA127RouteManifestSeamMatchesTheInternalFixture(t *testing.T) {
	fixture := newProductionRouteFixture(t)
	for _, market := range []Market{MarketKR, MarketUS} {
		if got, want := productionRouteBodyForTest(market, fixture.now), fixture.body(market); !reflect.DeepEqual(got, want) {
			t.Fatalf("%s seam body diverged from the internal fixture:\nseam=%+v\nfixture=%+v", market, got, want)
		}
	}
}
