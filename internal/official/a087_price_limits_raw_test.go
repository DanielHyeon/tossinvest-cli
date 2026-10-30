package official

import (
	"context"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
)

func TestA087ProtectivePriceLimitsPreservesNullableRawFields(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/oauth2/token":
			_, _ = w.Write([]byte(`{"access_token":"AT","expires_in":3600,"token_type":"Bearer"}`))
		case "/api/v1/price-limits":
			if got := r.URL.Query().Get("symbol"); got != "005930" {
				t.Errorf("symbol = %q, want 005930", got)
			}
			_, _ = w.Write([]byte(`{"result":{"lowerLimitPrice":"50400","timestamp":"2026-03-30T09:30:00+09:00"}}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()
	c := New(Credentials{APIKey: "k", SecretKey: "s"}, filepath.Join(t.TempDir(), "token.json"),
		WithBaseURL(srv.URL), WithHTTPClient(srv.Client()))
	got, err := c.ProtectivePriceLimits(t.Context(), "005930")
	if err != nil {
		t.Fatalf("ProtectivePriceLimits: %v", err)
	}
	if got.LowerLimitPrice == nil || *got.LowerLimitPrice != "50400" {
		t.Fatalf("LowerLimitPrice = %#v, want raw 50400", got.LowerLimitPrice)
	}
	if got.Timestamp == nil || *got.Timestamp != "2026-03-30T09:30:00+09:00" {
		t.Fatalf("Timestamp = %#v, want raw timestamp", got.Timestamp)
	}
}

func TestA087ProtectivePriceLimitsKeepsJSONNullDistinctFromZero(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/oauth2/token":
			_, _ = w.Write([]byte(`{"access_token":"AT","expires_in":3600,"token_type":"Bearer"}`))
		case "/api/v1/price-limits":
			_, _ = w.Write([]byte(`{"result":{"lowerLimitPrice":null,"timestamp":null}}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()
	c := New(Credentials{APIKey: "k", SecretKey: "s"}, filepath.Join(t.TempDir(), "token.json"),
		WithBaseURL(srv.URL), WithHTTPClient(srv.Client()))
	got, err := c.ProtectivePriceLimits(context.Background(), "005930")
	if err != nil {
		t.Fatalf("ProtectivePriceLimits: %v", err)
	}
	if got.LowerLimitPrice != nil || got.Timestamp != nil {
		t.Fatalf("nullable result = %#v, want both nil", got)
	}
}
