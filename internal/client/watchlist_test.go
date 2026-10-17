package client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/domain"
	"github.com/JungHoonGhae/tossinvest-cli/internal/session"
)

func TestListWatchlistFromFixtures(t *testing.T) {
	t.Parallel()
	var method, path string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		method, path = r.Method, r.URL.RequestURI()
		_, _ = w.Write([]byte(`{"result":{"watchlists":[
			{"id":2,"name":"later","ordering":1,"items":[{"code":"CODE2","symbol":"TICK2","name":"two","ordering":2,"prices":{"currency":"USD","base":2,"close":3}}]},
			{"id":3,"name":"third","ordering":2,"items":[{"code":"CODE3","name":"three","ordering":1}]},
			{"id":1,"name":"first","ordering":1,"items":[{"code":"CODE1A","symbol":"TICK1A","name":"one-a","ordering":3,"prices":{"currency":"KRW","base":1,"close":2}},{"code":"CODE1B","name":"one-b","ordering":1,"prices":null},{"code":"CODE1C","name":"one-c","ordering":1,"prices":{"currency":"USD","base":0,"close":0}}]}
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
	want := []domain.WatchlistItem{
		{Group: "later", Symbol: "CODE2", Name: "two", Currency: "USD", Base: 2, Last: 3},
		{Group: "first", Symbol: "CODE1B", Name: "one-b"},
		{Group: "first", Symbol: "CODE1C", Name: "one-c", Currency: "USD"},
		{Group: "first", Symbol: "CODE1A", Name: "one-a", Currency: "KRW", Base: 1, Last: 2},
		{Group: "third", Symbol: "CODE3", Name: "three"},
	}
	if !reflect.DeepEqual(items, want) {
		t.Fatalf("items=%+v want=%+v", items, want)
	}
}

func TestListWatchlistRejectsInvalidCollections(t *testing.T) {
	t.Parallel()
	for name, body := range map[string]string{"missing": `{}`, "null": `{"result":{"watchlists":null}}`, "watchlists object": `{"result":{"watchlists":{}}}`, "items missing": `{"result":{"watchlists":[{"id":1}]}}`, "items null": `{"result":{"watchlists":[{"id":1,"items":null}]}}`, "items object": `{"result":{"watchlists":[{"id":1,"items":{}}]}}`, "later items missing": `{"result":{"watchlists":[{"id":1,"items":[{"code":"OK"}]},{"id":2}]}}`, "empty": `{"result":{"watchlists":[]}}`, "empty items": `{"result":{"watchlists":[{"id":1,"items":[]}]}}`} {
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
