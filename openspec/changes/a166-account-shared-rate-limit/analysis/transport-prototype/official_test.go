package prototype

import (
	"context"
	"fmt"
	"github.com/JungHoonGhae/tossinvest-cli/internal/official"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"sync"
	"testing"
)

func TestOfficialClient401AndToken(t *testing.T) {
	var mu sync.Mutex
	var paths []string
	tokens := 0
	reads := 0
	s := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		paths = append(paths, r.URL.Path)
		if r.URL.Path == "/oauth2/token" {
			tokens++
			fmt.Fprintf(w, `{"access_token":"fake-%d","expires_in":3600}`, tokens)
			return
		}
		reads++
		if reads == 1 {
			w.WriteHeader(401)
			return
		}
		fmt.Fprint(w, `{"result":{"today":{"date":"2026-10-31"}}}`)
	}))
	defer s.Close()
	l := &limiter{tls: s.Client().Transport.(*http.Transport).TLSClientConfig}
	c := official.New(official.Credentials{APIKey: "fixture", SecretKey: "fixture"}, filepath.Join(t.TempDir(), "token"), official.WithBaseURL(s.URL), official.WithHTTPClient(&http.Client{Transport: l}))
	if _, ok := c.AuthorityOrigin(); ok {
		t.Fatal("fixture gained production authority")
	}
	_, err := c.TypedMarketCalendar(context.Background(), "KR", "2026-10-31")
	if err != nil {
		t.Fatal(err)
	}
	mu.Lock()
	defer mu.Unlock()
	if tokens != 2 || reads != 2 || len(paths) != 4 {
		t.Fatalf("tokens=%d reads=%d paths=%v", tokens, reads, paths)
	}
	l.logMu.Lock()
	defer l.logMu.Unlock()
	if len(l.starts) != 4 {
		t.Fatalf("writes %d", len(l.starts))
	}
}
func TestWrappedTLSResponseState(t *testing.T) {
	s := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { fmt.Fprint(w, "ok") }))
	defer s.Close()
	l := &limiter{tls: s.Client().Transport.(*http.Transport).TLSClientConfig}
	resp, err := (&http.Client{Transport: l}).Get(s.URL)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.TLS == nil || len(resp.TLS.VerifiedChains) == 0 {
		t.Fatal("verified response TLS state missing")
	}
}
