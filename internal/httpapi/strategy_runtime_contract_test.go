package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyprojection"
)

type apiStrategyRuntimeStub struct {
	snapshot strategyprojection.Snapshot
	err      error
}

func (s apiStrategyRuntimeStub) Read(context.Context) (strategyprojection.Snapshot, error) {
	return s.snapshot, s.err
}

func TestStrategyRuntimeRESTUsesSharedProjectionAndStrictGuards(t *testing.T) {
	snapshot := strategyprojection.DormantSnapshot(time.Date(2026, 8, 4, 12, 0, 0, 0, time.UTC))
	snapshot = strategyprojection.WithMarketFailure(snapshot, strategyprojection.MarketUS,
		strategyprojection.RefusalRuntimeUnavailable, time.Date(2026, 8, 4, 12, 0, 1, 0, time.UTC))
	handler, err := NewRouter(Options{Reader: contractReader{}, StrategyRuntime: apiStrategyRuntimeStub{snapshot: snapshot}})
	if err != nil {
		t.Fatal(err)
	}

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/v1/strategy-runtime", nil))
	if recorder.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", recorder.Code, recorder.Body.String())
	}
	var envelope struct {
		SchemaVersion string                      `json:"schemaVersion"`
		Resource      string                      `json:"resource"`
		Data          strategyprojection.Snapshot `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &envelope); err != nil {
		t.Fatal(err)
	}
	if envelope.SchemaVersion != SchemaVersion || envelope.Resource != "strategy-runtime" || !reflect.DeepEqual(envelope.Data, snapshot) {
		t.Fatalf("envelope=%+v", envelope)
	}

	for _, test := range []struct {
		method, target, body string
		want                 int
	}{
		{http.MethodPost, "/api/v1/strategy-runtime", "", http.StatusMethodNotAllowed},
		{http.MethodGet, "/api/v1/strategy-runtime?market=KR", "", http.StatusBadRequest},
		{http.MethodGet, "/api/v1/strategy-runtime", `{}`, http.StatusBadRequest},
		{http.MethodGet, "/api/v1/strategy-runtime/KR", "", http.StatusNotFound},
		{http.MethodPost, "/api/v1/strategy-runtime/activation", `{}`, http.StatusNotFound},
	} {
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, httptest.NewRequest(test.method, test.target, strings.NewReader(test.body)))
		if recorder.Code != test.want {
			t.Errorf("%s %s=%d want=%d body=%s", test.method, test.target, recorder.Code, test.want, recorder.Body.String())
		}
	}
}

func TestStrategyRuntimeRESTDormantHealthAndSSEFullSnapshotParity(t *testing.T) {
	now := time.Date(2026, 8, 4, 12, 0, 0, 0, time.UTC)
	dormant := strategyprojection.DormantSnapshot(now)
	handler, err := NewRouter(Options{Reader: contractReader{}, Now: func() time.Time { return now }})
	if err != nil {
		t.Fatal(err)
	}
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/v1/strategy-runtime", nil))
	var rest Envelope
	if err := json.Unmarshal(recorder.Body.Bytes(), &rest); err != nil {
		t.Fatal(err)
	}
	restData, _ := json.Marshal(rest.Data)
	var restSnapshot strategyprojection.Snapshot
	if err := json.Unmarshal(restData, &restSnapshot); err != nil {
		t.Fatalf("decode REST dormant projection: %v", err)
	}
	if !reflect.DeepEqual(restSnapshot, dormant) {
		t.Fatalf("REST dormant drift got=%+v want=%+v", restSnapshot, dormant)
	}

	reader := apiStrategyRuntimeStub{snapshot: dormant}
	snapshotFunc := StrategyRuntimeSnapshotFunc(reader, func() time.Time { return now })
	streamBody, err := snapshotFunc(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	var streamEnvelope Envelope
	if err := json.Unmarshal(streamBody, &streamEnvelope); err != nil {
		t.Fatal(err)
	}
	streamData, _ := json.Marshal(streamEnvelope.Data)
	var streamSnapshot strategyprojection.Snapshot
	if err := json.Unmarshal(streamData, &streamSnapshot); err != nil {
		t.Fatalf("decode SSE dormant projection: %v", err)
	}
	if streamEnvelope.Resource != "strategy-runtime" || !reflect.DeepEqual(streamSnapshot, dormant) {
		t.Fatalf("SSE snapshot drift=%s", streamBody)
	}
}
