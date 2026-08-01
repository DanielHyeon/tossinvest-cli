package httpapi

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestServerOwnedRequestLimitsAreNotWiderThanContract(t *testing.T) {
	if MaxRequestBodyBytes != 256<<10 {
		t.Fatalf("MaxRequestBodyBytes=%d want=%d", MaxRequestBodyBytes, 256<<10)
	}
	if DefaultReadHeaderTimeout != 5*time.Second {
		t.Fatalf("DefaultReadHeaderTimeout=%s want=5s", DefaultReadHeaderTimeout)
	}
	if DefaultReadTimeout != 5*time.Second {
		t.Fatalf("DefaultReadTimeout=%s want=5s", DefaultReadTimeout)
	}
	if DefaultMaxHeaderBytes != 32<<10 {
		t.Fatalf("DefaultMaxHeaderBytes=%d want=%d", DefaultMaxHeaderBytes, 32<<10)
	}
}

func TestLimitRequestBodyAcceptsBoundaryAndRejectsOneByteOver(t *testing.T) {
	handler := LimitRequestBody(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		_, err := io.ReadAll(request.Body)
		if err != nil {
			var tooLarge *http.MaxBytesError
			if errors.As(err, &tooLarge) {
				http.Error(w, "request body too large", http.StatusRequestEntityTooLarge)
				return
			}
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}))

	for _, test := range []struct {
		name string
		size int
		want int
	}{
		{"exact boundary", int(MaxRequestBodyBytes), http.StatusNoContent},
		{"one byte over", int(MaxRequestBodyBytes + 1), http.StatusRequestEntityTooLarge},
	} {
		t.Run(test.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodPost, "/api/v1/optimization/previews", strings.NewReader(strings.Repeat("x", test.size)))
			handler.ServeHTTP(recorder, request)
			if recorder.Code != test.want {
				t.Fatalf("body bytes=%d status=%d want=%d", test.size, recorder.Code, test.want)
			}
		})
	}
}

func TestNewServerPinsTimeoutsAndBodyLimit(t *testing.T) {
	readAll := http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		if _, err := io.ReadAll(request.Body); err != nil {
			http.Error(w, "read failed", http.StatusRequestEntityTooLarge)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})
	server, err := NewServer("127.0.0.1:0", readAll)
	if err != nil {
		t.Fatal(err)
	}
	if server.Addr != "127.0.0.1:0" || server.ReadHeaderTimeout != DefaultReadHeaderTimeout ||
		server.ReadTimeout != DefaultReadTimeout || server.MaxHeaderBytes != DefaultMaxHeaderBytes {
		t.Fatalf("server limits addr/header/read/max=%q/%s/%s/%d", server.Addr, server.ReadHeaderTimeout, server.ReadTimeout, server.MaxHeaderBytes)
	}
	if server.Handler == nil {
		t.Fatal("NewServer returned a nil handler")
	}

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/v1/optimization/previews", strings.NewReader(strings.Repeat("x", int(MaxRequestBodyBytes+1))))
	server.Handler.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("NewServer handler oversized status=%d", recorder.Code)
	}
}

func TestNewServerPreservesBodylessReadForRouter(t *testing.T) {
	router, err := NewRouter(Options{Reader: contractReader{}, Now: func() time.Time {
		return time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	}})
	if err != nil {
		t.Fatal(err)
	}
	server, err := NewServer("127.0.0.1:0", router)
	if err != nil {
		t.Fatal(err)
	}

	request := httptest.NewRequest(http.MethodGet, "/api/v1/engine", nil)
	if request.Body != http.NoBody {
		t.Fatal("test request did not start with the standard no-body sentinel")
	}
	recorder := httptest.NewRecorder()
	server.Handler.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusOK {
		t.Fatalf("bodyless GET through NewServer status=%d body=%s", recorder.Code, recorder.Body.String())
	}
}

func TestStaticMutationRouteAllowlistContainsOnlyNarrowOptimizationCommands(t *testing.T) {
	safe := map[string]struct{}{
		"/api/v1/optimization/previews":     {},
		"/api/v1/optimization/applications": {},
	}
	for route := range allowedMutationRoutes {
		if _, ok := safe[route]; !ok {
			t.Errorf("mutation route allowlist contains forbidden or broad route %q", route)
		}
	}
}

func TestForbiddenRemoteCapabilitiesHaveNoReachableRoute(t *testing.T) {
	router, err := NewRouter(Options{Reader: contractReader{}})
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{
		"/api/v1/engine/live",
		"/api/v1/gate",
		"/api/v1/kill-switch",
		"/api/v1/protection",
		"/api/v1/activation-manifest",
		"/api/v1/optimization/rollback-previews",
	} {
		for _, method := range []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete} {
			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, httptest.NewRequest(method, path, strings.NewReader(`{}`)))
			if recorder.Code != http.StatusNotFound && recorder.Code != http.StatusMethodNotAllowed {
				t.Errorf("%s %s status=%d, want 404/405", method, path, recorder.Code)
			}
		}
	}
}
