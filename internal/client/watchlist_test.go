package client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/session"
)

func TestListWatchlistFromFixtures(t *testing.T) {
	t.Parallel()
	var method, path string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		method, path = r.Method, r.URL.RequestURI()
		_, _ = w.Write([]byte(`{"result":{"watchlists":[
			{"id":2,"name":"later","ordering":2,"items":[{"code":"CODE2","symbol":"TICK2","name":"two","ordering":0,"prices":{"currency":"USD","base":2,"close":3}}]},
			{"id":1,"name":"first","ordering":1,"items":[{"code":"CODE1A","symbol":"TICK1A","name":"one-a","ordering":0,"prices":{"currency":"KRW","base":1,"close":2}},{"code":"CODE1B","name":"one-b","ordering":0,"prices":null}]}
		]}}`))
	}))
	defer srv.Close()
	c := New(Config{HTTPClient: srv.Client(), CertBaseURL: srv.URL, Session: &session.Session{Cookies: map[string]string{"SESSION": "x"}}})
	items, err := c.ListWatchlist(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if method != http.MethodGet || path != "/api/v1/new-watchlists?includePrice=true&lazyLoad=false" {
		t.Fatalf("route=%s %s", method, path)
	}
	if len(items) != 3 || items[0].Symbol != "CODE1A" || items[1].Symbol != "CODE1B" || items[2].Symbol != "CODE2" {
		t.Fatalf("items=%+v", items)
	}
	if items[0].Currency != "KRW" || items[1].Currency != "" || items[1].Base != 0 || items[1].Last != 0 {
		t.Fatalf("optional mapping=%+v", items[1])
	}
}

func TestListWatchlistRejectsInvalidCollections(t *testing.T) {
	t.Parallel()
	for name, body := range map[string]string{"missing": `{}`, "null": `{"result":{"watchlists":null}}`, "watchlists object": `{"result":{"watchlists":{}}}`, "items missing": `{"result":{"watchlists":[{"id":1}]}}`, "items object": `{"result":{"watchlists":[{"id":1,"items":{}}]}}`, "empty": `{"result":{"watchlists":[]}}`, "empty items": `{"result":{"watchlists":[{"id":1,"items":[]}]}}`} {
		t.Run(name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write([]byte(body)) }))
			defer srv.Close()
			c := New(Config{HTTPClient: srv.Client(), CertBaseURL: srv.URL, Session: &session.Session{Cookies: map[string]string{"SESSION": "x"}}})
			items, err := c.ListWatchlist(context.Background())
			if name == "empty" || name == "empty items" {
				if err != nil || items == nil || len(items) != 0 {
					t.Fatalf("items=%v err=%v", items, err)
				}
				return
			}
			if err == nil || items != nil {
				t.Fatalf("items=%v err=%v, want schema error", items, err)
			}
		})
	}
}

func watchlistFixturePath(t *testing.T) string {
	t.Helper()
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("failed to resolve test path")
	}
	return filepath.Join(filepath.Dir(filename), "..", "..", "fixtures", "responses", "auth-sanitized", "asset-sections-v2.json")
}
