package httpapi

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
)

type flushingRecorder struct {
	mu      sync.Mutex
	header  http.Header
	status  int
	body    bytes.Buffer
	flushes chan struct{}
}

type responseWriterWithoutFlusher struct {
	header http.Header
	status int
}

func (w *responseWriterWithoutFlusher) Header() http.Header { return w.header }
func (w *responseWriterWithoutFlusher) Write(data []byte) (int, error) {
	if w.status == 0 {
		w.status = http.StatusOK
	}
	return len(data), nil
}
func (w *responseWriterWithoutFlusher) WriteHeader(status int) { w.status = status }

func newFlushingRecorder() *flushingRecorder {
	return &flushingRecorder{header: make(http.Header), flushes: make(chan struct{}, 8)}
}

func (r *flushingRecorder) Header() http.Header { return r.header }

func (r *flushingRecorder) WriteHeader(status int) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.status == 0 {
		r.status = status
	}
}

func (r *flushingRecorder) Write(data []byte) (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.status == 0 {
		r.status = http.StatusOK
	}
	return r.body.Write(data)
}

func (r *flushingRecorder) Flush() {
	r.flushes <- struct{}{}
}

func (r *flushingRecorder) snapshot() (int, string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.status, r.body.String()
}

func waitForFlush(t *testing.T, recorder *flushingRecorder) {
	t.Helper()
	select {
	case <-recorder.flushes:
	case <-time.After(time.Second):
		t.Fatal("SSE handler did not flush")
	}
}

func TestStreamHandlerWritesSnapshotUpdatesAndHeartbeat(t *testing.T) {
	fake := clock.NewFake(time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC))
	stream, _ := newTestStream(t, StreamOptions{Epoch: "process-a", Clock: fake})
	ctx, cancel := context.WithCancel(context.Background())
	request := httptest.NewRequest(http.MethodGet, "/api/v1/stream", nil).WithContext(ctx)
	recorder := newFlushingRecorder()
	handlerDone := make(chan struct{})
	go func() {
		defer close(handlerDone)
		stream.ServeHTTP(recorder, request)
	}()

	waitForFlush(t, recorder)
	status, body := recorder.snapshot()
	if status != http.StatusOK || recorder.Header().Get("Content-Type") != "text/event-stream" {
		t.Fatalf("status/content-type=%d/%q", status, recorder.Header().Get("Content-Type"))
	}
	for _, want := range []string{"id: process-a:0\n", "event: snapshot\n", "data: {\"schema_version\":\"v1\",\"full\":true}\n\n"} {
		if !strings.Contains(body, want) {
			t.Fatalf("initial SSE body lacks %q: %q", want, body)
		}
	}

	if _, err := stream.Publish("positions", []byte(`{"count":1}`)); err != nil {
		t.Fatal(err)
	}
	waitForFlush(t, recorder)
	_, body = recorder.snapshot()
	for _, want := range []string{"id: process-a:1\n", "event: positions\n", "data: {\"count\":1}\n\n"} {
		if !strings.Contains(body, want) {
			t.Fatalf("update SSE body lacks %q: %q", want, body)
		}
	}

	if !fake.WaitForSleepers(1, time.Second) {
		t.Fatal("handler heartbeat did not use the injected clock")
	}
	fake.Advance(DefaultStreamHeartbeat)
	waitForFlush(t, recorder)
	_, body = recorder.snapshot()
	if !strings.Contains(body, ": heartbeat\n\n") {
		t.Fatalf("SSE body lacks heartbeat: %q", body)
	}

	cancel()
	select {
	case <-handlerDone:
	case <-time.After(time.Second):
		t.Fatal("SSE handler did not stop after request cancellation")
	}
	if got := stream.ClientCount(); got != 0 {
		t.Fatalf("client count after cancellation=%d", got)
	}
}

func TestRouterExposesOnlyExactStreamPath(t *testing.T) {
	var calls atomic.Int32
	streamHandler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusNoContent)
	})
	router, err := NewRouter(Options{Reader: contractReader{}, Stream: streamHandler})
	if err != nil {
		t.Fatal(err)
	}

	for _, test := range []struct {
		path string
		want int
	}{
		{"/api/v1/stream", http.StatusNoContent},
		{"/api/v1/stream/", http.StatusNotFound},
		{"/api/v1/streams", http.StatusNotFound},
	} {
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, test.path, nil))
		if recorder.Code != test.want {
			t.Errorf("GET %s status=%d want=%d", test.path, recorder.Code, test.want)
		}
	}
	if got := calls.Load(); got != 1 {
		t.Fatalf("stream handler calls=%d", got)
	}
}

func TestStreamOperationalFailureUsesStableErrorSchema(t *testing.T) {
	stream, _ := newTestStream(t, StreamOptions{Epoch: "process-error"})
	request := httptest.NewRequest(http.MethodPost, "/api/v1/stream", nil)
	recorder := newFlushingRecorder()
	stream.ServeHTTP(recorder, request)
	status, body := recorder.snapshot()
	if status != http.StatusMethodNotAllowed {
		t.Fatalf("stream error status=%d body=%s", status, body)
	}
	assertStableErrorBody(t, []byte(body), "METHOD_NOT_ALLOWED")
}

func TestStreamRejectsRepeatedOrOversizedLastEventID(t *testing.T) {
	stream, _ := newTestStream(t, StreamOptions{Epoch: "process-header"})
	for _, values := range [][]string{{"process-header:0", "process-header:0"}, {strings.Repeat("x", MaxLastEventIDBytes+1)}} {
		request := httptest.NewRequest(http.MethodGet, "/api/v1/stream", nil)
		request.Header["Last-Event-Id"] = values
		recorder := newFlushingRecorder()
		stream.ServeHTTP(recorder, request)
		status, body := recorder.snapshot()
		if status != http.StatusBadRequest {
			t.Fatalf("Last-Event-ID %q status=%d body=%s", values, status, body)
		}
		assertStableErrorBody(t, []byte(body), "INVALID_LAST_EVENT_ID")
	}
}

func TestStreamHeadReturnsHeadersWithoutConsumingClientSlot(t *testing.T) {
	stream, _ := newTestStream(t, StreamOptions{Epoch: "process-a", MaxClients: 1})
	recorder := httptest.NewRecorder()
	stream.ServeHTTP(recorder, httptest.NewRequest(http.MethodHead, "/api/v1/stream", nil))
	if recorder.Code != http.StatusOK || recorder.Header().Get("Content-Type") != "text/event-stream" {
		t.Fatalf("HEAD status/content-type=%d/%q", recorder.Code, recorder.Header().Get("Content-Type"))
	}
	if recorder.Body.Len() != 0 || stream.ClientCount() != 0 {
		t.Fatalf("HEAD body/client count=%d/%d", recorder.Body.Len(), stream.ClientCount())
	}
}

func TestStreamHandlerRejectsInvalidMethodAndNonStreamingWriter(t *testing.T) {
	stream, _ := newTestStream(t, StreamOptions{Epoch: "process-a"})

	methodRecorder := httptest.NewRecorder()
	stream.ServeHTTP(methodRecorder, httptest.NewRequest(http.MethodPost, "/api/v1/stream", nil))
	if methodRecorder.Code != http.StatusMethodNotAllowed || methodRecorder.Header().Get("Allow") != "GET, HEAD" {
		t.Fatalf("POST status/allow=%d/%q", methodRecorder.Code, methodRecorder.Header().Get("Allow"))
	}

	plainWriter := &responseWriterWithoutFlusher{header: make(http.Header)}
	stream.ServeHTTP(plainWriter, httptest.NewRequest(http.MethodGet, "/api/v1/stream", nil))
	if plainWriter.status != http.StatusInternalServerError {
		t.Fatalf("non-streaming writer status=%d", plainWriter.status)
	}
}

func TestStreamHandlerReportsSnapshotFailureAndClientLimit(t *testing.T) {
	stream, fixture := newTestStream(t, StreamOptions{Epoch: "process-a", MaxClients: 1})
	fixture.err = context.Canceled
	snapshotRecorder := httptest.NewRecorder()
	stream.ServeHTTP(snapshotRecorder, httptest.NewRequest(http.MethodGet, "/api/v1/stream", nil))
	if snapshotRecorder.Code != http.StatusServiceUnavailable || snapshotRecorder.Header().Get("Content-Type") == "text/event-stream" {
		t.Fatalf("snapshot failure status/content-type=%d/%q", snapshotRecorder.Code, snapshotRecorder.Header().Get("Content-Type"))
	}

	fixture.err = nil
	occupied, err := stream.subscribe(context.Background(), "process-a:0")
	if err != nil {
		t.Fatal(err)
	}
	defer occupied.close()
	limitRecorder := httptest.NewRecorder()
	stream.ServeHTTP(limitRecorder, httptest.NewRequest(http.MethodGet, "/api/v1/stream", nil))
	if limitRecorder.Code != http.StatusServiceUnavailable || limitRecorder.Header().Get("Retry-After") != "15" {
		t.Fatalf("client limit status/retry=%d/%q", limitRecorder.Code, limitRecorder.Header().Get("Retry-After"))
	}
}

func TestStreamHandlerHonorsLastEventIDOnTheWire(t *testing.T) {
	stream, fixture := newTestStream(t, StreamOptions{Epoch: "process-a"})
	for _, data := range []string{`{"n":1}`, `{"n":2}`} {
		if _, err := stream.Publish("update", []byte(data)); err != nil {
			t.Fatal(err)
		}
	}

	ctx, cancel := context.WithCancel(context.Background())
	request := httptest.NewRequest(http.MethodGet, "/api/v1/stream", nil).WithContext(ctx)
	request.Header.Set("Last-Event-ID", "process-a:1")
	recorder := newFlushingRecorder()
	done := make(chan struct{})
	go func() {
		defer close(done)
		stream.ServeHTTP(recorder, request)
	}()
	waitForFlush(t, recorder)
	_, body := recorder.snapshot()
	if fixture.calls != 0 || strings.Contains(body, "event: snapshot") || !strings.Contains(body, "id: process-a:2") {
		t.Fatalf("wire resume calls/body=%d/%q", fixture.calls, body)
	}
	cancel()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("resumed stream did not stop")
	}
}
