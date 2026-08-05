package httpapi

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/performance"
)

func TestPerformanceResourceProjectsExactMarketCampaignAttributionWithoutInventedNumbers(t *testing.T) {
	row := performance.NewUnavailableAttribution(performance.AttributionKey{
		Market: "US", Ticker: "AAPL", LaneID: "US_ST_FLOW_CONTINUATION_V1", LaneVersion: "lane/v1",
		CampaignID: "campaign-1", LegID: "leg-1", PositionID: "position-1", PolicyID: "risk", PolicyVersion: "risk/v1",
	}, []string{"fill_id"}, []string{"fees", "fx"}, "USD", "KRW")
	resource := PerformanceFrom(performance.DashboardView{Attributions: []performance.Attribution{row}})
	if len(resource.Attributions) != 1 {
		t.Fatalf("attributions=%+v", resource.Attributions)
	}
	got := resource.Attributions[0]
	if got.Market != "US" || got.CampaignID != "campaign-1" || got.LegID != "leg-1" ||
		got.LineageStatus != performance.StatusLinkMissing || got.Reporting.NetPnL.Status != performance.StatusNotMeasured ||
		got.Reporting.NetPnL.Value != "" {
		t.Fatalf("projected attribution=%+v", got)
	}
	body, err := json.Marshal(resource)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{`"attributions"`, `"campaignId":"campaign-1"`, `"missingMeasurements":["fees","fx"]`, `"netPnl":{"status":"not_measured","value":""`} {
		if !strings.Contains(string(body), want) {
			t.Fatalf("JSON missing %s: %s", want, body)
		}
	}
}
