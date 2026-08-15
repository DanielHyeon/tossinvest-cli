package official

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAttemptObserverSeesBodyBeforeDecode(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/oauth2/token" {
			fmt.Fprint(w, `{"access_token":"token","expires_in":3600,"token_type":"Bearer"}`)
			return
		}
		fmt.Fprint(w, `{"result":{"value":"raw"}}`)
	}))
	defer srv.Close()
	c := New(Credentials{APIKey: "k", SecretKey: "s"}, t.TempDir()+"/token", WithBaseURL(srv.URL), WithHTTPClient(srv.Client()))
	var got []AttemptTrace
	ctx := WithAttemptObserver(context.Background(), func(a AttemptTrace) { got = append(got, a) })
	var out struct {
		Value string `json:"value"`
	}
	if err := c.get(ctx, "/trace", nil, &out); err != nil {
		t.Fatalf("get: %v", err)
	}
	if len(got) != 1 || got[0].StatusCode != http.StatusOK || string(got[0].Body) != `{"result":{"value":"raw"}}` {
		t.Fatalf("traces = %+v", got)
	}
	if got[0].BodyReadComplete.Before(got[0].RequestStart) || out.Value != "raw" {
		t.Fatalf("trace/decode ordering invalid: %+v out=%+v", got[0], out)
	}
}

func TestAttemptObserverTraces401RefreshThenSuccessfulRetry(t *testing.T) {
	requests := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/oauth2/token":
			fmt.Fprint(w, `{"access_token":"token","expires_in":3600,"token_type":"Bearer"}`)
		case "/trace-retry":
			requests++
			if requests == 1 {
				w.WriteHeader(http.StatusUnauthorized)
				fmt.Fprint(w, `{"message":"unauthorized"}`)
				return
			}
			fmt.Fprint(w, `{"result":{"value":"retry"}}`)
		}
	}))
	defer srv.Close()
	c := New(Credentials{APIKey: "k", SecretKey: "s"}, t.TempDir()+"/token", WithBaseURL(srv.URL), WithHTTPClient(srv.Client()))
	var got []AttemptTrace
	ctx := WithAttemptObserver(context.Background(), func(trace AttemptTrace) { got = append(got, trace) })
	var out struct{ Value string }
	if err := c.get(ctx, "/trace-retry", nil, &out); err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[0].StatusCode != http.StatusUnauthorized || got[1].StatusCode != http.StatusOK ||
		string(got[0].Body) != `{"message":"unauthorized"}` || string(got[1].Body) != `{"result":{"value":"retry"}}` || out.Value != "retry" {
		t.Fatalf("retry traces=%+v out=%+v", got, out)
	}
}

func TestAttemptObserverReturnPrecedesDecodeAndCallerReturn(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/oauth2/token" {
			fmt.Fprint(w, `{"access_token":"token","expires_in":3600,"token_type":"Bearer"}`)
			return
		}
		fmt.Fprint(w, `{"result":{"value":"blocked"}}`)
	}))
	defer srv.Close()
	c := New(Credentials{APIKey: "k", SecretKey: "s"}, t.TempDir()+"/token", WithBaseURL(srv.URL), WithHTTPClient(srv.Client()))
	entered, release := make(chan struct{}), make(chan struct{})
	ctx := WithAttemptObserver(context.Background(), func(AttemptTrace) { close(entered); <-release })
	done := make(chan error, 1)
	var out struct{ Value string }
	go func() { done <- c.get(ctx, "/trace-block", nil, &out) }()
	<-entered
	select {
	case err := <-done:
		t.Fatalf("caller returned before observer release: %v", err)
	default:
	}
	close(release)
	if err := <-done; err != nil || out.Value != "blocked" {
		t.Fatalf("decoded result after observer = %+v err=%v", out, err)
	}
}
