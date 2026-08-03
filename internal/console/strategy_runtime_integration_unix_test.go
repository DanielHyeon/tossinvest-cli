//go:build unix

package console

import (
	"bufio"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/httpapi"
	"github.com/JungHoonGhae/tossinvest-cli/internal/performance"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyprojection"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyprojectionrpc"
)

type integrationProjectionStore struct {
	mu       sync.RWMutex
	snapshot strategyprojection.Snapshot
}

func (s *integrationProjectionStore) Read(context.Context) (strategyprojection.Snapshot, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return strategyprojection.Clone(s.snapshot), nil
}

func (s *integrationProjectionStore) replace(snapshot strategyprojection.Snapshot) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.snapshot = strategyprojection.Clone(snapshot)
}

type integrationAPIReader struct{}

func (integrationAPIReader) Engine(context.Context) (httpapi.EngineResource, error) {
	return httpapi.EngineResource{}, nil
}
func (integrationAPIReader) Positions(context.Context) (httpapi.PositionsResource, error) {
	return httpapi.PositionsResource{}, nil
}
func (integrationAPIReader) Orders(context.Context) (httpapi.OrdersResource, error) {
	return httpapi.OrdersResource{}, nil
}
func (integrationAPIReader) Candidates(context.Context) (httpapi.CandidatesResource, error) {
	return httpapi.CandidatesResource{}, nil
}
func (integrationAPIReader) Performance(context.Context) (performance.DashboardView, error) {
	return performance.DashboardView{}, nil
}
func (integrationAPIReader) Settings(context.Context) (httpapi.SettingsResource, error) {
	return httpapi.SettingsResource{}, nil
}
func (integrationAPIReader) Optimization(context.Context) (httpapi.OptimizationRead, error) {
	return httpapi.OptimizationRead{}, nil
}

func TestStrategyRuntimeUnixConsoleAPIAndSSEConvergeWithoutCrossMarketFallback(t *testing.T) {
	now := consoleProjectionNow
	full := consoleProjectionPair(t)
	store := &integrationProjectionStore{snapshot: full}
	runtimeDir, err := os.MkdirTemp("/tmp", "sprpc-integration-")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(runtimeDir) })
	if err := os.Chmod(runtimeDir, 0o700); err != nil {
		t.Fatal(err)
	}
	server, err := strategyprojectionrpc.Start(runtimeDir, store)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = server.Close() })
	client, err := strategyprojectionrpc.Dial(context.Background(), strategyprojectionrpc.DescriptorPath(runtimeDir))
	if err != nil {
		t.Fatal(err)
	}

	stream, err := httpapi.NewStream(httpapi.StreamOptions{Epoch: "runtime-fixture"},
		httpapi.StrategyRuntimeSnapshotFunc(client, func() time.Time { return now }))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(stream.Close)
	api, err := httpapi.NewRouter(httpapi.Options{Reader: integrationAPIReader{}, StrategyRuntime: client,
		Stream: stream, Now: func() time.Time { return now }})
	if err != nil {
		t.Fatal(err)
	}
	apiServer := httptest.NewServer(api)
	t.Cleanup(apiServer.Close)
	consoleHarness := newHarness(t, func(options *Options) { options.StrategyRuntime = client })
	consoleHarness.authenticate(t)

	assertStrategyProjectionAPI(t, apiServer.URL, full)
	assertConsoleProjectionValues(t, body(t, consoleHarness.get(t, "/strategy-runtime")), full)

	partial := strategyprojection.WithMarketFailure(full, strategyprojection.MarketUS,
		strategyprojection.RefusalRuntimeUnavailable, now.Add(time.Second))
	store.replace(partial)
	assertStrategyProjectionAPI(t, apiServer.URL, partial)
	partialPage := body(t, consoleHarness.get(t, "/strategy-runtime"))
	assertConsoleProjectionValues(t, partialPage, partial)
	if !strings.Contains(partialPage, "evidence-KR") || strings.Contains(partialPage, "evidence-US") {
		t.Fatal("US failure crossed into the preserved KR console projection")
	}

	recovered := consoleProjectionPair(t)
	us := recovered.Markets[strategyprojection.MarketUS]
	reconnectedLane := "lane-US-reconnected"
	us.Lane.ID = &reconnectedLane
	recovered.Markets[strategyprojection.MarketUS] = us
	store.replace(recovered)
	streamSnapshot := readInitialStrategyRuntimeSSE(t, apiServer.URL, "stale-epoch:7")
	if !reflect.DeepEqual(streamSnapshot, recovered) {
		t.Fatalf("reconnect did not converge to full paired snapshot: got=%+v want=%+v", streamSnapshot, recovered)
	}
	assertStrategyProjectionAPI(t, apiServer.URL, recovered)
	assertConsoleProjectionValues(t, body(t, consoleHarness.get(t, "/strategy-runtime")), recovered)
}

func TestStrategyRuntimeDormantUnixHealthDoesNotActivateEitherMarket(t *testing.T) {
	dormant := strategyprojection.DormantSnapshot(consoleProjectionNow)
	store := &integrationProjectionStore{snapshot: dormant}
	runtimeDir, err := os.MkdirTemp("/tmp", "sprpc-health-")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(runtimeDir) })
	if err := os.Chmod(runtimeDir, 0o700); err != nil {
		t.Fatal(err)
	}
	server, err := strategyprojectionrpc.Start(runtimeDir, store)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = server.Close() })
	client, err := strategyprojectionrpc.Dial(context.Background(), strategyprojectionrpc.DescriptorPath(runtimeDir))
	if err != nil {
		t.Fatal(err)
	}
	read, err := client.Read(context.Background())
	if err != nil || !reflect.DeepEqual(read, dormant) {
		t.Fatalf("dormant Unix health read=%+v err=%v", read, err)
	}
	for _, market := range []strategyprojection.Market{strategyprojection.MarketKR, strategyprojection.MarketUS} {
		item := read.Markets[market]
		if item.Lane.Desired != strategyprojection.StateOff || item.Lane.Effective != strategyprojection.StateOff ||
			item.Activation.Desired != strategyprojection.StateOff || item.Activation.Effective != strategyprojection.StateOff ||
			item.Protection.Readiness != strategyprojection.ProtectionUnwired || item.FirstRefusal != strategyprojection.RefusalNotConfigured {
			t.Fatalf("dormant health activated %s: %+v", market, item)
		}
	}
}

func assertStrategyProjectionAPI(t *testing.T, baseURL string, want strategyprojection.Snapshot) {
	t.Helper()
	response, err := http.Get(baseURL + "/api/v1/strategy-runtime")
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	var envelope struct {
		Resource string                      `json:"resource"`
		Data     strategyprojection.Snapshot `json:"data"`
	}
	if err := json.NewDecoder(response.Body).Decode(&envelope); err != nil {
		t.Fatal(err)
	}
	if response.StatusCode != http.StatusOK || envelope.Resource != "strategy-runtime" || !reflect.DeepEqual(envelope.Data, want) {
		t.Fatalf("API projection status=%d envelope=%+v want=%+v", response.StatusCode, envelope, want)
	}
}

func assertConsoleProjectionValues(t *testing.T, page string, snapshot strategyprojection.Snapshot) {
	t.Helper()
	for _, item := range strategyprojection.OrderedMarkets(snapshot) {
		for _, want := range []string{string(item.Market), string(item.Status), string(item.Lane.Desired),
			string(item.Lane.Effective), string(item.Evidence.Freshness), string(item.Protection.Readiness),
			string(item.Protection.Refusal), string(item.Reconciliation.Status), string(item.FirstRefusal)} {
			if !strings.Contains(page, want) {
				t.Errorf("console projection for %s lacks %q", item.Market, want)
			}
		}
		for _, value := range []*string{item.Lane.ID, item.Lane.Version, item.Evidence.ID, item.Evidence.Digest,
			item.Campaign.ID, item.Campaign.LegID, item.HorizonRisk.Bucket, item.HorizonRisk.PolicyVersion,
			item.Scheduler.CalendarSource, item.Scheduler.CalendarVersion, item.Activation.Digest} {
			if value != nil && !strings.Contains(page, *value) {
				t.Errorf("console projection for %s lacks value %q", item.Market, *value)
			}
		}
	}
}

func readInitialStrategyRuntimeSSE(t *testing.T, baseURL, lastEventID string) strategyprojection.Snapshot {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+"/api/v1/stream", nil)
	if err != nil {
		t.Fatal(err)
	}
	request.Header.Set("Last-Event-ID", lastEventID)
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	scanner := bufio.NewScanner(response.Body)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		var envelope struct {
			Resource string                      `json:"resource"`
			Data     strategyprojection.Snapshot `json:"data"`
		}
		if err := json.Unmarshal([]byte(strings.TrimPrefix(line, "data: ")), &envelope); err != nil {
			t.Fatal(err)
		}
		if envelope.Resource != "strategy-runtime" {
			t.Fatalf("unexpected SSE resource %q", envelope.Resource)
		}
		return envelope.Data
	}
	if err := scanner.Err(); err != nil {
		t.Fatal(err)
	}
	t.Fatal("SSE ended without a full strategy-runtime snapshot")
	return strategyprojection.Snapshot{}
}
