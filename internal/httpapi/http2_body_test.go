package httpapi

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

type unknownLengthReader struct {
	io.Reader
}

func TestHTTP2BodylessReadsAndStreamRespectWireBody(t *testing.T) {
	var streamCalls atomic.Int64
	router, err := NewRouter(Options{
		Reader: contractReader{},
		Stream: http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			streamCalls.Add(1)
			w.WriteHeader(http.StatusNoContent)
		}),
		Now: func() time.Time { return time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC) },
	})
	if err != nil {
		t.Fatal(err)
	}
	server, err := NewServer("127.0.0.1:0", router)
	if err != nil {
		t.Fatal(err)
	}
	tlsServer := httptest.NewUnstartedServer(server.Handler)
	tlsServer.EnableHTTP2 = true
	tlsServer.StartTLS()
	t.Cleanup(tlsServer.Close)

	client := tlsServer.Client()
	for _, path := range []string{
		"/api/v1/engine",
		"/api/v1/positions",
		"/api/v1/orders",
		"/api/v1/candidates",
		"/api/v1/performance",
		"/api/v1/settings",
		"/api/v1/optimization",
	} {
		t.Run("bodyless GET "+path, func(t *testing.T) {
			response, err := client.Get(tlsServer.URL + path)
			if err != nil {
				t.Fatal(err)
			}
			defer response.Body.Close()
			if response.ProtoMajor != 2 {
				t.Fatalf("protocol=%s want HTTP/2", response.Proto)
			}
			if response.StatusCode != http.StatusOK {
				body, _ := io.ReadAll(response.Body)
				t.Fatalf("status=%d want=200 body=%s", response.StatusCode, body)
			}
		})
	}

	t.Run("bodyless HEAD", func(t *testing.T) {
		request, err := http.NewRequest(http.MethodHead, tlsServer.URL+"/api/v1/engine", nil)
		if err != nil {
			t.Fatal(err)
		}
		response, err := client.Do(request)
		if err != nil {
			t.Fatal(err)
		}
		defer response.Body.Close()
		body, err := io.ReadAll(response.Body)
		if err != nil {
			t.Fatal(err)
		}
		if response.ProtoMajor != 2 || response.StatusCode != http.StatusOK || len(body) != 0 {
			t.Fatalf("protocol/status/body=%s/%d/%q", response.Proto, response.StatusCode, body)
		}
	})

	t.Run("bodyless stream", func(t *testing.T) {
		response, err := client.Get(tlsServer.URL + "/api/v1/stream")
		if err != nil {
			t.Fatal(err)
		}
		defer response.Body.Close()
		if response.ProtoMajor != 2 || response.StatusCode != http.StatusNoContent {
			t.Fatalf("protocol/status=%s/%d want HTTP/2/204", response.Proto, response.StatusCode)
		}
		if streamCalls.Load() != 1 {
			t.Fatalf("stream calls=%d want=1", streamCalls.Load())
		}
	})

	for _, test := range []struct {
		name string
		path string
		body io.Reader
	}{
		{name: "known-length read body", path: "/api/v1/engine", body: strings.NewReader(`{"x":1}`)},
		{name: "unknown-length read body", path: "/api/v1/engine", body: unknownLengthReader{strings.NewReader(`{"x":1}`)}},
		{name: "known-length stream body", path: "/api/v1/stream", body: strings.NewReader(`{"x":1}`)},
		{name: "unknown-length stream body", path: "/api/v1/stream", body: unknownLengthReader{strings.NewReader(`{"x":1}`)}},
	} {
		t.Run(test.name, func(t *testing.T) {
			request, err := http.NewRequest(http.MethodGet, tlsServer.URL+test.path, test.body)
			if err != nil {
				t.Fatal(err)
			}
			response, err := client.Do(request)
			if err != nil {
				t.Fatal(err)
			}
			defer response.Body.Close()
			body, err := io.ReadAll(response.Body)
			if err != nil {
				t.Fatal(err)
			}
			if response.ProtoMajor != 2 || response.StatusCode != http.StatusBadRequest {
				t.Fatalf("protocol/status=%s/%d want HTTP/2/400 body=%s", response.Proto, response.StatusCode, body)
			}
			if !strings.Contains(string(body), `"code":"BODY_NOT_SUPPORTED"`) {
				t.Fatalf("stable body error missing: %s", body)
			}
			if test.path == "/api/v1/stream" && streamCalls.Load() != 1 {
				t.Fatalf("rejected body reached stream handler; calls=%d", streamCalls.Load())
			}
		})
	}
}

func TestRequestHasBodyUsesPreservedHTTP2ContentLength(t *testing.T) {
	for _, test := range []struct {
		name        string
		request     *http.Request
		wantHasBody bool
	}{
		{name: "nil request", request: nil, wantHasBody: false},
		{name: "known empty", request: &http.Request{Header: make(http.Header), ContentLength: 0}, wantHasBody: false},
		{name: "explicit zero header", request: &http.Request{Header: http.Header{"Content-Length": {"0"}}, ContentLength: 0}, wantHasBody: false},
		{name: "comma joined zero header", request: &http.Request{Header: http.Header{"Content-Length": {"0, 0"}}, ContentLength: 0}, wantHasBody: false},
		{name: "preserved positive header", request: &http.Request{Header: http.Header{"Content-Length": {"1"}}, ContentLength: 0}, wantHasBody: true},
		{name: "preserved malformed header", request: &http.Request{Header: http.Header{"Content-Length": {"invalid"}}, ContentLength: 0}, wantHasBody: true},
		{name: "preserved signed positive zero", request: &http.Request{Header: http.Header{"Content-Length": {"+0"}}, ContentLength: 0}, wantHasBody: true},
		{name: "preserved signed negative zero", request: &http.Request{Header: http.Header{"Content-Length": {"-0"}}, ContentLength: 0}, wantHasBody: true},
		{name: "preserved overflowing length", request: &http.Request{Header: http.Header{"Content-Length": {"18446744073709551616"}}, ContentLength: 0}, wantHasBody: true},
		{name: "unknown length", request: &http.Request{Header: make(http.Header), ContentLength: -1}, wantHasBody: true},
		{name: "declared length", request: &http.Request{Header: make(http.Header), ContentLength: 1}, wantHasBody: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			if got := requestHasBody(test.request); got != test.wantHasBody {
				t.Fatalf("requestHasBody()=%v want=%v", got, test.wantHasBody)
			}
		})
	}
}
