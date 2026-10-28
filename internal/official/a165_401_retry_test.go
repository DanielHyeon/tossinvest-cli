package official

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
)

func TestA165Post401RetryPreservesEncodedBody(t *testing.T) {
	var bodies [][]byte
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/oauth2/token":
			_, _ = w.Write([]byte(`{"access_token":"AT","expires_in":3600,"token_type":"Bearer"}`))
		case "/a165/orders":
			body, err := io.ReadAll(r.Body)
			if err != nil {
				t.Fatalf("read request: %v", err)
			}
			bodies = append(bodies, body)
			if len(bodies) == 1 {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			_, _ = w.Write([]byte(`{"result":{"ok":true}}`))
		}
	}))
	defer srv.Close()
	c := New(Credentials{APIKey: "k", SecretKey: "s"}, filepath.Join(t.TempDir(), "token.json"),
		WithBaseURL(srv.URL), WithHTTPClient(srv.Client()))
	var out struct {
		OK bool `json:"ok"`
	}
	if err := c.post(context.Background(), "/a165/orders", map[string]string{"clientOrderId": "a165-fixed"}, &out); err != nil {
		t.Fatalf("post: %v", err)
	}
	if len(bodies) != 2 || string(bodies[0]) != string(bodies[1]) || !out.OK {
		t.Fatalf("bodies=%q out=%+v", bodies, out)
	}
}

func TestA165PostDoesNotRetryNon401AndFinal401IsAuth(t *testing.T) {
	for _, status := range []int{http.StatusBadRequest, http.StatusTooManyRequests, http.StatusInternalServerError} {
		t.Run(http.StatusText(status), func(t *testing.T) {
			calls := 0
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch r.URL.Path {
				case "/oauth2/token":
					_, _ = w.Write([]byte(`{"access_token":"AT","expires_in":3600,"token_type":"Bearer"}`))
				case "/a165/fail":
					calls++
					w.WriteHeader(status)
				}
			}))
			defer srv.Close()
			c := New(Credentials{APIKey: "k", SecretKey: "s"}, filepath.Join(t.TempDir(), "token.json"),
				WithBaseURL(srv.URL), WithHTTPClient(srv.Client()))
			if err := c.post(context.Background(), "/a165/fail", struct{}{}, nil); err == nil {
				t.Fatal("post succeeded")
			}
			if calls != 1 {
				t.Fatalf("calls=%d, want 1", calls)
			}
		})
	}

	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/oauth2/token":
			_, _ = w.Write([]byte(`{"access_token":"AT","expires_in":3600,"token_type":"Bearer"}`))
		case "/a165/auth":
			calls++
			w.WriteHeader(http.StatusUnauthorized)
		}
	}))
	defer srv.Close()
	c := New(Credentials{APIKey: "k", SecretKey: "s"}, filepath.Join(t.TempDir(), "token.json"),
		WithBaseURL(srv.URL), WithHTTPClient(srv.Client()))
	err := c.post(context.Background(), "/a165/auth", struct{}{}, nil)
	if !errors.Is(err, ErrAuth) || calls != 2 {
		t.Fatalf("err=%v calls=%d", err, calls)
	}
}
