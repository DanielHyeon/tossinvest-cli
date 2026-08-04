package official

import (
	"context"
	"crypto/tls"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
)

func TestAdaptExchangeRatePreservesAuthorityFields(t *testing.T) {
	raw := apiExchangeRate{
		BaseCurrency: "USD", QuoteCurrency: "KRW", Rate: "1380.5000", MidRate: "1375.2500",
		ValidFrom: "2026-03-25T09:30:00+09:00", ValidUntil: "2026-03-25T09:31:00+09:00",
	}
	got := adaptExchangeRate(raw)
	if got.BaseCurrency != raw.BaseCurrency || got.QuoteCurrency != raw.QuoteCurrency ||
		got.RateRaw != raw.Rate || got.MidRateRaw != raw.MidRate ||
		got.ValidFromRaw != raw.ValidFrom || got.ValidUntilRaw != raw.ValidUntil {
		t.Fatalf("lossless exchange-rate fields: got %+v, raw %+v", got, raw)
	}
	if got.Base != 1375.25 || got.Close != 1380.5 {
		t.Fatalf("legacy display fields changed: base=%v close=%v", got.Base, got.Close)
	}
}

func TestExchangeRatePreservesAuthorityFieldsAcrossHTTPBoundary(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/oauth2/token":
			_, _ = w.Write([]byte(`{"access_token":"AT","expires_in":3600,"token_type":"Bearer"}`))
		case "/api/v1/exchange-rate":
			_, _ = w.Write([]byte(`{"result":{"baseCurrency":"USD","quoteCurrency":"KRW","rate":"1380.5000","midRate":"1375.2500","validFrom":"2026-03-25T09:30:00+09:00","validUntil":"2026-03-25T09:31:00+09:00"}}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()
	client := New(Credentials{APIKey: "k", SecretKey: "s"}, filepath.Join(t.TempDir(), "token.json"),
		WithBaseURL(srv.URL), WithHTTPClient(srv.Client()))

	got, err := client.ExchangeRate(context.Background(), "USD", "KRW")
	if err != nil {
		t.Fatal(err)
	}
	if got.BaseCurrency != "USD" || got.QuoteCurrency != "KRW" || got.RateRaw != "1380.5000" || got.MidRateRaw != "1375.2500" ||
		got.ValidFromRaw != "2026-03-25T09:30:00+09:00" || got.ValidUntilRaw != "2026-03-25T09:31:00+09:00" {
		t.Fatalf("lossless exchange-rate response: %+v", got)
	}
}

func TestAuthoritativeExchangeRateRejectsConfiguredClientBeforeHTTP(t *testing.T) {
	hits := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		http.Error(w, "configured transport must not be called", http.StatusInternalServerError)
	}))
	defer srv.Close()
	client := New(Credentials{APIKey: "k", SecretKey: "s"}, filepath.Join(t.TempDir(), "token.json"),
		WithBaseURL(srv.URL), WithHTTPClient(srv.Client()))

	if _, err := client.AuthoritativeExchangeRate(context.Background(), "USD", "KRW"); !errors.Is(err, ErrAuthorityOrigin) {
		t.Fatalf("error=%v, want ErrAuthorityOrigin", err)
	}
	if hits != 0 {
		t.Fatalf("configured transport called %d times", hits)
	}
}

func TestAuthoritativeExchangeRateKeepsOriginAndReadInOneBoundary(t *testing.T) {
	enteredRead := make(chan struct{})
	releaseRead := make(chan struct{})
	srv := httptest.NewTLSServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		switch request.URL.Path {
		case "/oauth2/token":
			_, _ = writer.Write([]byte(`{"access_token":"AT","expires_in":3600,"token_type":"Bearer"}`))
		case "/api/v1/exchange-rate":
			close(enteredRead)
			<-releaseRead
			_, _ = writer.Write([]byte(`{"result":{"baseCurrency":"USD","quoteCurrency":"KRW","rate":"1380.5000","midRate":"1375.2500","validFrom":"2026-03-25T09:30:00+09:00","validUntil":"2026-03-25T09:31:00+09:00"}}`))
		default:
			http.NotFound(writer, request)
		}
	}))
	defer srv.Close()
	transport := srv.Client().Transport.(*http.Transport).Clone()
	transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true} // test server certificate only
	target := srv.Listener.Addr().String()
	transport.DialContext = func(ctx context.Context, network, _ string) (net.Conn, error) {
		return (&net.Dialer{}).DialContext(ctx, network, target)
	}
	httpClient := &http.Client{Transport: transport}
	client := &Client{
		base: defaultBaseURL, hc: httpClient, authorityOrigin: true, authorityTransport: transport,
		configurationSealed: true,
	}
	client.tm = newTokenManager(Credentials{APIKey: "k", SecretKey: "s"}, defaultBaseURL, filepath.Join(t.TempDir(), "token.json"), httpClient)

	type result struct {
		rate string
		err  error
	}
	resultCh := make(chan result, 1)
	go func() {
		rate, err := client.AuthoritativeExchangeRate(context.Background(), "USD", "KRW")
		resultCh <- result{rate: rate.RateRaw, err: err}
	}()
	<-enteredRead
	replayDone := make(chan struct{})
	go func() {
		WithBaseURL("https://attacker.invalid")(client)
		WithHTTPClient(http.DefaultClient)(client)
		close(replayDone)
	}()
	close(releaseRead)
	got := <-resultCh
	<-replayDone
	if got.err != nil || got.rate != "1380.5000" {
		t.Fatalf("authoritative read: rate=%q err=%v", got.rate, got.err)
	}
	if client.BaseURL() != defaultBaseURL {
		t.Fatalf("origin changed during authoritative read: %q", client.BaseURL())
	}
}
