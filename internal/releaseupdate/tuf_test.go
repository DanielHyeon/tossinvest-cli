package releaseupdate

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
)

func TestProductionTUFURLPolicyIsExact(t *testing.T) {
	for _, raw := range []string{
		"http://tuf-repo-cdn.sigstore.dev/root.json",
		"https://evil.tuf-repo-cdn.sigstore.dev/root.json",
		"https://tuf-repo-cdn.sigstore.dev.evil.invalid/root.json",
		"https://user@tuf-repo-cdn.sigstore.dev/root.json",
		"https://tuf-repo-cdn.sigstore.dev:444/root.json",
		"https://127.0.0.1/root.json",
	} {
		target, err := url.Parse(raw)
		if err != nil {
			t.Fatal(err)
		}
		if productionTUFURLAllowed(target) {
			t.Errorf("allowed TUF URL %s", raw)
		}
	}
	target, _ := url.Parse("https://tuf-repo-cdn.sigstore.dev:443/root.json")
	if !productionTUFURLAllowed(target) {
		t.Fatal("exact production TUF URL was refused")
	}
}

func TestTUFFetcherForbidsRedirectAndBoundsChunkedResponse(t *testing.T) {
	var server *httptest.Server
	server = httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/redirect":
			http.Redirect(w, r, server.URL+"/metadata", http.StatusFound)
		case "/oversized":
			w.Header().Set("Transfer-Encoding", "chunked")
			_, _ = io.WriteString(w, strings.Repeat("x", 9))
		default:
			_, _ = io.WriteString(w, "metadata")
		}
	}))
	t.Cleanup(server.Close)
	base, err := url.Parse(server.URL)
	if err != nil {
		t.Fatal(err)
	}
	allow := func(target *url.URL) bool {
		return target.Scheme == "https" && target.Host == base.Host
	}
	fetcher, err := newTUFFetcher(context.Background(), server.Client(), allow)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := fetcher.DownloadFile(server.URL+"/redirect", 64, time.Hour); err == nil {
		t.Fatal("TUF redirect succeeded")
	}
	if _, err := fetcher.DownloadFile(server.URL+"/oversized", 8, time.Hour); err == nil {
		t.Fatal("oversized chunked TUF response succeeded")
	}
}

func TestTUFFetcherUsesCallerCancellation(t *testing.T) {
	entered := make(chan struct{}, 1)
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		entered <- struct{}{}
		<-r.Context().Done()
	}))
	t.Cleanup(server.Close)
	base, err := url.Parse(server.URL)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	fetcher, err := newTUFFetcher(ctx, server.Client(), func(target *url.URL) bool {
		return target.Scheme == "https" && target.Host == base.Host
	})
	if err != nil {
		t.Fatal(err)
	}
	done := make(chan error, 1)
	go func() {
		_, err := fetcher.DownloadFile(server.URL+"/metadata", 64, time.Hour)
		done <- err
	}()
	select {
	case <-entered:
	case <-time.After(5 * time.Second):
		t.Fatal("TUF request did not start")
	}
	cancel()
	select {
	case err := <-done:
		if err == nil {
			t.Fatal("cancelled TUF request succeeded")
		}
	case <-time.After(5 * time.Second):
		t.Fatal("caller cancellation did not stop TUF request")
	}
}

func TestTUFFetcherHonorsExplicitHTTPTimeout(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(
		w http.ResponseWriter,
		r *http.Request,
	) {
		<-r.Context().Done()
	}))
	t.Cleanup(server.Close)
	base, err := url.Parse(server.URL)
	if err != nil {
		t.Fatal(err)
	}
	client := server.Client()
	client.Timeout = 25 * time.Millisecond
	fetcher, err := newTUFFetcher(context.Background(), client, func(target *url.URL) bool {
		return target.Scheme == "https" && target.Host == base.Host
	})
	if err != nil {
		t.Fatal(err)
	}
	started := time.Now()
	if _, err := fetcher.DownloadFile(server.URL+"/metadata", 64, time.Hour); err == nil {
		t.Fatal("timed-out TUF request succeeded")
	}
	if elapsed := time.Since(started); elapsed > time.Second {
		t.Fatalf("TUF HTTP timeout took %s", elapsed)
	}
}
