package httpapi

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWriteBoundaryRejectionUsesStableErrorContract(t *testing.T) {
	t.Parallel()
	for _, method := range []string{http.MethodGet, http.MethodHead} {
		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(method, "https://localhost/api/v1/engine", nil)
		WriteBoundaryRejection(recorder, request)
		if recorder.Code != http.StatusForbidden || recorder.Header().Get("Cache-Control") != "no-store" ||
			recorder.Header().Get("Content-Type") != "application/json; charset=utf-8" {
			t.Fatalf("method=%s status=%d headers=%v", method, recorder.Code, recorder.Header())
		}
		if method == http.MethodHead {
			if recorder.Body.Len() != 0 {
				t.Fatalf("HEAD body=%s", recorder.Body.String())
			}
			continue
		}
		assertStableErrorBody(t, recorder.Body.Bytes(), "BOUNDARY_REFUSED")
	}
}
