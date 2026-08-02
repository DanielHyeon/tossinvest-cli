package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/operatorview"
	"github.com/JungHoonGhae/tossinvest-cli/internal/optimization"
	"github.com/JungHoonGhae/tossinvest-cli/internal/performance"
)

type contractReader struct {
	err error
}

func (r contractReader) Engine(context.Context) (EngineResource, error) {
	return EngineResource{Status: "stopped", Running: false, Source: "engine-marker"}, r.err
}

func (r contractReader) Positions(context.Context) (PositionsResource, error) {
	return PositionsResource{Items: []Position{{
		Market: "KR", Symbol: "005930", Quantity: "1", ManagementStatus: "managed",
		ExitLine: ExitLineFrom(operatorview.ExitLineView{Status: "fresh", StatusText: "평가 완료", CurrentProtection: "68000"}),
	}}}, r.err
}

func (r contractReader) Orders(context.Context) (OrdersResource, error) {
	return OrdersResource{Items: []Order{{ID: "order-1", Market: "KR", Symbol: "005930", Status: "OPEN"}}}, r.err
}

func (r contractReader) Candidates(context.Context) (CandidatesResource, error) {
	return CandidatesResource{Items: []Candidate{{Market: "KR", Symbol: "005930", Verdict: "eligible"}}}, r.err
}

func (r contractReader) Performance(context.Context) (performance.DashboardView, error) {
	return performance.DashboardView{Query: performance.Query{AsOf: time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC), PeriodDays: 30, Market: performance.AllMarkets, Lane: performance.AllLanes, CompleteOnly: true, MinimumSample: 20}}, r.err
}

func (r contractReader) Settings(context.Context) (SettingsResource, error) {
	return SettingsResource{Version: 7, Items: []Setting{{Key: "engine.autostart", Desired: State{Kind: "value", Value: "OFF"}, Effective: State{Kind: "value", Value: "OFF"}}}}, r.err
}

func (r contractReader) Optimization(ctx context.Context) (OptimizationRead, error) {
	registry, err := optimization.CoreRegistry(ctx)
	if err != nil {
		return OptimizationRead{}, err
	}
	return OptimizationRead{Core: optimization.View{Registry: registry,
		Snapshot: optimization.Snapshot{Version: 3, EffectiveVersion: 3}}}, r.err
}

func TestVersionedReadResourcesUseOneStableEnvelope(t *testing.T) {
	handler, err := NewRouter(Options{Reader: contractReader{}, Now: func() time.Time {
		return time.Date(2026, 8, 1, 1, 2, 3, 0, time.UTC)
	}})
	if err != nil {
		t.Fatal(err)
	}

	for _, resource := range ReadResourceNames() {
		t.Run(resource, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/v1/"+resource, nil))
			if recorder.Code != http.StatusOK {
				t.Fatalf("status = %d, body=%s", recorder.Code, recorder.Body.String())
			}
			if got := recorder.Header().Get("Content-Type"); got != "application/json; charset=utf-8" {
				t.Fatalf("content type = %q", got)
			}
			if got := recorder.Header().Get("Cache-Control"); got != "no-store" {
				t.Fatalf("cache control = %q", got)
			}
			var envelope struct {
				SchemaVersion string          `json:"schemaVersion"`
				Resource      string          `json:"resource"`
				GeneratedAt   time.Time       `json:"generatedAt"`
				Data          json.RawMessage `json:"data"`
			}
			if err := json.Unmarshal(recorder.Body.Bytes(), &envelope); err != nil {
				t.Fatal(err)
			}
			if envelope.SchemaVersion != SchemaVersion || envelope.Resource != resource || len(envelope.Data) == 0 {
				t.Fatalf("bad envelope: %+v", envelope)
			}
		})
	}
}

func TestPositionsExposeA043ExitLineWithCamelCaseFields(t *testing.T) {
	handler, _ := NewRouter(Options{Reader: contractReader{}})
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/v1/positions", nil))
	body := recorder.Body.String()
	for _, want := range []string{`"exitLine"`, `"currentProtection":"68000"`, `"statusText":"평가 완료"`} {
		if !strings.Contains(body, want) {
			t.Errorf("response lacks %s: %s", want, body)
		}
	}
	for _, forbidden := range []string{`"ExitLine"`, `"CurrentProtection"`, `"StatusText"`} {
		if strings.Contains(body, forbidden) {
			t.Errorf("response leaked Go field %s: %s", forbidden, body)
		}
	}
}

func TestPositionsExposeReconcileAwareAdoptionFactsWithoutAccountIdentity(t *testing.T) {
	startedAt := time.Date(2026, 8, 1, 1, 2, 3, 0, time.UTC)
	value := Position{
		Market: "US", Symbol: "AAPL", Eligible: false,
		AdoptionStatus: "RECONCILE_BLOCKED", StatusKnown: true,
		AdoptionLabel: "조정 확인 대기", AdoptionReason: "RECONCILE_BLOCK_ACTIVE",
		Included: true, Excluded: false, Candidate: true, DesignationKnown: true,
		CoveringBlock: &ReconcileBlock{Scope: "ACCOUNT", Market: "", Symbol: "",
			Reason: "QUANTITY_MISMATCH", StartedAt: &startedAt},
		StoredExitEvidence: &StoredExitEvidence{EntryPrice: "200", InitialStop: "190",
			Baseline: "190", HighWater: "205", EffectiveKnown: false, Label: "원장 기록 · 실효 미확인"},
	}
	body, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	text := string(body)
	for _, want := range []string{
		`"adoptionStatus":"RECONCILE_BLOCKED"`, `"statusKnown":true`,
		`"adoptionLabel":"조정 확인 대기"`, `"adoptionReason":"RECONCILE_BLOCK_ACTIVE"`,
		`"included":true`, `"excluded":false`, `"candidate":true`, `"designationKnown":true`,
		`"coveringBlock":{"scope":"ACCOUNT"`, `"reason":"QUANTITY_MISMATCH"`,
		`"storedExitEvidence":{"entryPrice":"200"`, `"effectiveKnown":false`,
	} {
		if !strings.Contains(text, want) {
			t.Errorf("position JSON lacks %s: %s", want, text)
		}
	}
	for _, forbidden := range []string{"accountRef", "capability", "token", "command"} {
		if strings.Contains(strings.ToLower(text), strings.ToLower(forbidden)) {
			t.Errorf("position JSON leaked forbidden field %q: %s", forbidden, text)
		}
	}
}

func TestOptimizationUsesCanonicalCategoryOrderAndOwnerDescriptors(t *testing.T) {
	view, err := contractReader{}.Optimization(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	resource := OptimizationFrom(view)
	gotIDs := make([]string, 0, len(resource.Categories))
	for _, category := range resource.Categories {
		gotIDs = append(gotIDs, category.ID)
	}
	wantIDs := []string{"overview", "exit-protection", "position-management", "candidate-filters", "strategy-runtime", "performance-history"}
	if !reflect.DeepEqual(gotIDs, wantIDs) {
		t.Fatalf("category order = %#v", gotIDs)
	}
	if resource.PositionManagement.AutoEnabledDefault || resource.PositionManagement.StopDefault != "5%" ||
		len(resource.PositionManagement.StopOptions) != 37 || resource.PositionManagement.ExcludePrecedence != "exclude 우선" {
		t.Fatalf("position owner descriptor drift: %+v", resource.PositionManagement)
	}
	if len(resource.CandidateFilters) != 2 || len(resource.CandidateFilters[0].Filters) == 0 {
		t.Fatalf("candidate owner descriptors missing: %+v", resource.CandidateFilters)
	}
	first := resource.CandidateFilters[0].Filters[0]
	if first.DefaultState != "unapproved" || first.DesiredValue != "" || first.EffectiveValue != "" {
		t.Fatalf("unapproved threshold laundered into a value: %+v", first)
	}
	if len(resource.Fields) != 1 || resource.Fields[0].Key != "exit.common-policy" || resource.Fields[0].Owner != "a041-complete-exit-line-contract" {
		t.Fatalf("a050 registry field drift: %+v", resource.Fields)
	}
}

func TestOptimizationKeepsDesiredAndUnavailableEffectiveAdoptionDistinct(t *testing.T) {
	view, err := contractReader{}.Optimization(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	view.PositionManagement = PositionManagementActual{
		Desired: AdoptionSettings{Enabled: true, DefaultStopPct: 0.03,
			IncludeSymbols: []string{"AAPL"}, ExcludeSymbols: []string{"TSLA"}},
		EffectiveKnown: false,
	}
	resource := OptimizationFrom(view)
	management := resource.PositionManagement
	if !management.Desired.Enabled || management.Desired.DefaultStopPct != 0.03 ||
		!reflect.DeepEqual(management.Desired.IncludeSymbols, []string{"AAPL"}) ||
		!reflect.DeepEqual(management.Desired.ExcludeSymbols, []string{"TSLA"}) {
		t.Fatalf("desired adoption drift: %+v", management.Desired)
	}
	if management.EffectiveKnown || management.Effective != nil {
		t.Fatalf("unavailable runtime was laundered into effective: %+v", management)
	}
	if !management.AutoEnabledDesired || management.StopDesired != "3%" ||
		management.AutoEnabledEffective || management.StopEffective != "알 수 없음" {
		t.Fatalf("legacy adoption summary contradicts actual knownness: %+v", management)
	}
	effective := AdoptionSettings{Enabled: false, DefaultStopPct: .05}
	view.PositionManagement.EffectiveKnown = true
	view.PositionManagement.Effective = &effective
	management = OptimizationFrom(view).PositionManagement
	if management.AutoEnabledEffective || management.StopEffective != "5%" || management.Effective == nil {
		t.Fatalf("legacy effective summary contradicts runtime: %+v", management)
	}
}

func TestReadRouterRejectsArbitraryQueryAndMutationByDefault(t *testing.T) {
	handler, _ := NewRouter(Options{Reader: contractReader{}})
	for _, tc := range []struct {
		method string
		path   string
		want   int
	}{
		{http.MethodGet, "/api/v1/performance?periodDays=365", http.StatusBadRequest},
		{http.MethodGet, "/api/v1/performance", http.StatusBadRequest},
		{http.MethodPost, "/api/v1/settings", http.StatusMethodNotAllowed},
		{http.MethodPost, "/api/v1/engine/live", http.StatusNotFound},
		{http.MethodPost, "/api/v1/gate", http.StatusNotFound},
		{http.MethodPost, "/api/v1/optimization/rollback-previews", http.StatusNotFound},
		{http.MethodGet, "/api/v1/unknown", http.StatusNotFound},
	} {
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, httptest.NewRequest(tc.method, tc.path, strings.NewReader(`{}`)))
		if recorder.Code != tc.want {
			t.Errorf("%s %s status=%d want=%d body=%s", tc.method, tc.path, recorder.Code, tc.want, recorder.Body.String())
		}
	}
}

func TestOnlyExactInjectedMutationPathIsReachable(t *testing.T) {
	called := 0
	mutation := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { called++; w.WriteHeader(http.StatusNoContent) })
	handler, err := NewRouter(Options{Reader: contractReader{}, MutationRoutes: map[string]http.Handler{
		"/api/v1/optimization/previews": mutation,
	}})
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{"/api/v1/optimization/previews", "/api/v1/optimization/previews/", "/api/v1/optimization/apply"} {
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, path, nil))
		if path == "/api/v1/optimization/previews" && recorder.Code != http.StatusNoContent {
			t.Fatalf("exact route status=%d", recorder.Code)
		}
		if path != "/api/v1/optimization/previews" && recorder.Code != http.StatusNotFound {
			t.Fatalf("near route %s status=%d", path, recorder.Code)
		}
	}
	if called != 1 {
		t.Fatalf("mutation calls=%d", called)
	}
}

func TestErrorsUseStableNonLeakingSchema(t *testing.T) {
	handler, _ := NewRouter(Options{Reader: contractReader{err: errors.New("secret db path /private/account.sqlite")}})
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/v1/orders", nil))
	if recorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("status=%d body=%s", recorder.Code, recorder.Body.String())
	}
	var response ErrorResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	if response.SchemaVersion != ErrorSchemaVersion || response.Error.Code != "RESOURCE_UNAVAILABLE" || response.Error.Message == "" {
		t.Fatalf("error contract=%+v", response)
	}
	if strings.Contains(recorder.Body.String(), "private/account") {
		t.Fatalf("internal error leaked: %s", recorder.Body.String())
	}
}

func TestRouterConfigurationRejectsUnsafeMutationPaths(t *testing.T) {
	for _, path := range []string{"settings", "/api/v1/engine/live", "/api/v1/gate", "/api/v1/optimization/rollback-previews", "/api/v1/optimization/", "/api/v1/optimization/*"} {
		_, err := NewRouter(Options{Reader: contractReader{}, MutationRoutes: map[string]http.Handler{path: http.NotFoundHandler()}})
		if err == nil {
			t.Errorf("accepted unsafe mutation path %q", path)
		}
	}
}
