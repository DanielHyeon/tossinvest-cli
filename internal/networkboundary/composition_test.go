package networkboundary_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/networkboundary"
)

func TestBoundaryComposesInjectedStableRejectionBeforeRouter(t *testing.T) {
	t.Parallel()
	boundary, err := networkboundary.New(networkboundary.ServerConfig{
		Bind: "127.0.0.1", AllowedCIDRs: []string{"127.0.0.0/8"},
		PublicOrigin: "https://localhost:443", TLSConfigured: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	routerCalls := 0
	router := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		routerCalls++
		w.WriteHeader(http.StatusNoContent)
	})
	rejectionCalls := 0
	handler := boundary.HandlerWithRejection(router, func(w http.ResponseWriter, _ *http.Request) {
		rejectionCalls++
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"schemaVersion":"tossos.operator-api.error/v1","error":{"code":"BOUNDARY_REFUSED"}}`))
	})

	request := httptest.NewRequest(http.MethodGet, "https://localhost/api/v1/engine", nil)
	request.RemoteAddr = "203.0.113.10:50000"
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusForbidden || rejectionCalls != 1 || routerCalls != 0 {
		t.Fatalf("status/rejection/router calls=%d/%d/%d", recorder.Code, rejectionCalls, routerCalls)
	}
	if recorder.Header().Get("Content-Type") != "application/json; charset=utf-8" ||
		recorder.Header().Get("Cache-Control") != "no-store" ||
		recorder.Header().Get("X-Content-Type-Options") != "nosniff" {
		t.Fatalf("rejection headers=%v", recorder.Header())
	}
	var body struct {
		SchemaVersion string `json:"schemaVersion"`
		Error         struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("rejection is not JSON: %v body=%s", err, recorder.Body.String())
	}
	if body.SchemaVersion != "tossos.operator-api.error/v1" || body.Error.Code != "BOUNDARY_REFUSED" {
		t.Fatalf("rejection body=%+v", body)
	}
}
