package official

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestA112MBUSCandlePreservesOneAttemptRawEvidenceAndWireHeaders(t *testing.T) {
	var calls atomic.Int32
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		if r.Method != http.MethodGet || r.URL.Path != "/api/v1/candles" || r.URL.RawQuery != "adjusted=false&count=200&interval=1m&symbol=AAPL" {
			t.Fatalf("request = %s %s?%s", r.Method, r.URL.Path, r.URL.RawQuery)
		}
		if got := r.Header.Values("Authorization"); len(got) != 1 || got[0] != "Bearer cached" {
			t.Fatalf("authorization = %#v", got)
		}
		if got := r.Header.Values("Accept-Encoding"); len(got) != 1 || got[0] != "identity" {
			t.Fatalf("accept-encoding = %#v", got)
		}
		if got := r.Header.Values("User-Agent"); len(got) != 0 {
			t.Fatalf("user-agent leaked = %#v", got)
		}
		for name := range r.Header {
			if name != "Authorization" && name != "Accept" && name != "Accept-Encoding" {
				t.Fatalf("unexpected application header %q", name)
			}
		}
		w.Header().Set("X-RateLimit-Limit", "10")
		w.Header().Set("X-RateLimit-Remaining", "9")
		w.Header().Set("X-RateLimit-Reset", "1")
		_, _ = w.Write([]byte(`{"result":{"candles":[{"timestamp":"2026-08-14T13:30:00Z","closePrice":"100.0100","currency":"USD"}],"nextBefore":"\u0000 space \uD83C\uDF4F"}}`))
	}))
	defer server.Close()

	client := a112MBUSTestClient(t, server)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	got, err := A112MBUSCandle(ctx, client, nil)
	if err != nil {
		t.Fatal(err)
	}
	if calls.Load() != 1 {
		t.Fatalf("data calls = %d, want 1", calls.Load())
	}
	if got.Method() != http.MethodGet || got.Path() != "/api/v1/candles" || got.CanonicalQuery() != "adjusted=false&count=200&interval=1m&symbol=AAPL" {
		t.Fatalf("descriptor = %s %s?%s", got.Method(), got.Path(), got.CanonicalQuery())
	}
	if !strings.Contains(string(got.Body()), `"100.0100"`) || !strings.Contains(string(got.Body()), `"USD"`) {
		t.Fatalf("raw body lost decimal/currency: %s", got.Body())
	}
	if string(got.CursorJSON()) != `"\u0000 space \uD83C\uDF4F"` || string(got.CursorValue()) != "\x00 space 🍏" {
		t.Fatalf("cursor raw=%q decoded=%q", got.CursorJSON(), got.CursorValue())
	}
	if got.RateHeader("X-RateLimit-Remaining")[0] != "9" {
		t.Fatalf("rate headers = %#v", got.RateHeaders())
	}
	copy := got.Body()
	copy[0] = '!'
	if got.Body()[0] == '!' {
		t.Fatal("caller mutated opaque result body")
	}
}

func TestA112MBUSRejectsDeadlineAndConfiguredClientBeforeDataGET(t *testing.T) {
	var calls atomic.Int32
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { calls.Add(1) }))
	defer server.Close()
	client := a112MBUSTestClient(t, server)
	if _, err := A112MBUSCandle(context.Background(), client, nil); err == nil {
		t.Fatal("missing deadline accepted")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 16*time.Second)
	defer cancel()
	if _, err := A112MBUSCandle(ctx, client, nil); err == nil {
		t.Fatal("over-limit deadline accepted")
	}
	configured := New(Credentials{}, filepath.Join(t.TempDir(), "configured.json"), WithBaseURL(server.URL), WithHTTPClient(server.Client()))
	short, cancelShort := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelShort()
	if _, err := A112MBUSCandle(short, configured, nil); err == nil {
		t.Fatal("configured client accepted")
	}
	if calls.Load() != 0 {
		t.Fatalf("data calls = %d, want 0", calls.Load())
	}
}

func TestA112MBUSHoldsMalformedCursorRateAndStatusWithoutRetry(t *testing.T) {
	tests := []struct {
		name   string
		status int
		body   string
		head   http.Header
	}{
		{"cursor absent", http.StatusOK, `{"result":{}}`, a112MBUSGoodRateHeaders()},
		{"cursor empty", http.StatusOK, `{"result":{"nextBefore":""}}`, a112MBUSGoodRateHeaders()},
		{"cursor number", http.StatusOK, `{"result":{"nextBefore":1}}`, a112MBUSGoodRateHeaders()},
		{"cursor object", http.StatusOK, `{"result":{"nextBefore":{}}}`, a112MBUSGoodRateHeaders()},
		{"cursor array", http.StatusOK, `{"result":{"nextBefore":[]}}`, a112MBUSGoodRateHeaders()},
		{"duplicate rate", http.StatusOK, `{"result":{"nextBefore":null}}`, http.Header{"X-RateLimit-Limit": {"10", "9"}, "X-RateLimit-Remaining": {"9"}, "X-RateLimit-Reset": {"1"}}},
		{"unauthorized", http.StatusUnauthorized, `no`, http.Header{}},
		{"forbidden", http.StatusForbidden, `no`, http.Header{}},
		{"429", http.StatusTooManyRequests, `rate limited`, http.Header{"Retry-After": {"1"}}},
		{"server error", http.StatusInternalServerError, `no`, http.Header{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var calls atomic.Int32
			server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls.Add(1)
				for key, values := range tt.head {
					for _, value := range values {
						w.Header().Add(key, value)
					}
				}
				w.WriteHeader(tt.status)
				_, _ = w.Write([]byte(tt.body))
			}))
			defer server.Close()
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if _, err := A112MBUSCandle(ctx, a112MBUSTestClient(t, server), nil); err == nil {
				t.Fatal("malformed response minted evidence")
			}
			if calls.Load() != 1 {
				t.Fatalf("calls = %d, want one attempt", calls.Load())
			}
		})
	}
}

func TestA112MBUSRejectsDuplicateResultDuplicateCursorAndInvalidUTF8(t *testing.T) {
	tests := []struct {
		name string
		body []byte
	}{
		{"duplicate top level result", []byte(`{"result":{"nextBefore":null},"result":{"nextBefore":null}}`)},
		{"duplicate payload nextBefore", []byte(`{"result":{"nextBefore":null,"nextBefore":"cursor"}}`)},
		{"invalid UTF-8 source", append([]byte(`{"result":{"nextBefore":"`), append([]byte{0xff}, []byte(`"}}`)...)...)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var calls atomic.Int32
			server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				calls.Add(1)
				for key, values := range a112MBUSGoodRateHeaders() {
					for _, value := range values {
						w.Header().Add(key, value)
					}
				}
				_, _ = w.Write(tt.body)
			}))
			defer server.Close()
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if _, err := A112MBUSCandle(ctx, a112MBUSTestClient(t, server), nil); err == nil {
				t.Fatal("ambiguous source body minted evidence")
			}
			if calls.Load() != 1 {
				t.Fatalf("calls = %d, want one", calls.Load())
			}
		})
	}
}

func TestA112MBUSRejectsUnpairedAndMisorderedCursorSurrogates(t *testing.T) {
	tests := []struct {
		name string
		body string
		want string
	}{
		{"unpaired high surrogate", `{"result":{"nextBefore":"\uD800"}}`, ""},
		{"unpaired low surrogate", `{"result":{"nextBefore":"\uDC00"}}`, ""},
		{"misordered low then high surrogate", `{"result":{"nextBefore":"\uDC00\uD800"}}`, ""},
		{"valid surrogate pair", `{"result":{"nextBefore":"\uD83C\uDF4F"}}`, "🍏"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				for key, values := range a112MBUSGoodRateHeaders() {
					for _, value := range values {
						w.Header().Add(key, value)
					}
				}
				_, _ = w.Write([]byte(tt.body))
			}))
			defer server.Close()
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			got, err := A112MBUSCandle(ctx, a112MBUSTestClient(t, server), nil)
			if tt.want == "" {
				if err == nil {
					t.Fatalf("surrogate-invalid cursor minted decoded=%q raw=%q", got.CursorValue(), got.CursorJSON())
				}
				return
			}
			if err != nil || string(got.CursorValue()) != tt.want || string(got.CursorJSON()) != `"\uD83C\uDF4F"` {
				t.Fatalf("valid pair raw=%q decoded=%q err=%v", got.CursorJSON(), got.CursorValue(), err)
			}
		})
	}
}

func TestA112MBUSRejectsCrossClientTokenManagerBinding(t *testing.T) {
	var calls atomic.Int32
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls.Add(1)
		for key, values := range a112MBUSGoodRateHeaders() {
			for _, value := range values {
				w.Header().Add(key, value)
			}
		}
		_, _ = w.Write([]byte(`{"result":{"nextBefore":null}}`))
	}))
	defer server.Close()
	client := a112MBUSTestClient(t, server)
	crossClient := a112MBUSTestClient(t, server)
	client.tm = crossClient.tm // same-package spoof of a foreign client binding
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if _, err := A112MBUSCandle(ctx, client, nil); err == nil {
		t.Fatal("cross-client token manager minted evidence")
	}
	if calls.Load() != 0 {
		t.Fatalf("cross-client binding sent data calls=%d", calls.Load())
	}
}

func TestA112MBUSRejectsInvalidInitialCursorBeforeRequest(t *testing.T) {
	var calls atomic.Int32
	server := httptest.NewTLSServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { calls.Add(1) }))
	defer server.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if _, err := A112MBUSCandle(ctx, a112MBUSTestClient(t, server), []byte{0xff}); err == nil {
		t.Fatal("invalid UTF-8 initial cursor reached network")
	}
	if calls.Load() != 0 {
		t.Fatalf("invalid initial cursor sent data calls=%d", calls.Load())
	}
}

func TestA112MBUSOrderbookAndCalendarRejectMissingOrNullResult(t *testing.T) {
	for _, body := range []string{`{}`, `{"result":null}`} {
		t.Run(body, func(t *testing.T) {
			server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				for key, values := range a112MBUSGoodRateHeaders() {
					for _, value := range values {
						w.Header().Add(key, value)
					}
				}
				_, _ = w.Write([]byte(body))
			}))
			defer server.Close()
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			client := a112MBUSTestClient(t, server)
			if _, err := A112MBUSOrderbook(ctx, client); err == nil {
				t.Fatal("malformed orderbook envelope minted evidence")
			}
			if _, err := A112MBUSCalendar(ctx, client, "2026-08-14"); err == nil {
				t.Fatal("malformed calendar envelope minted evidence")
			}
		})
	}
}

func TestA112MBUSExactOrderbookAndCalendarDescriptors(t *testing.T) {
	var paths []string
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		paths = append(paths, r.URL.Path+"?"+r.URL.RawQuery)
		for key, values := range a112MBUSGoodRateHeaders() {
			for _, value := range values {
				w.Header().Add(key, value)
			}
		}
		_, _ = w.Write([]byte(`{"result":{"nextBefore":null}}`))
	}))
	defer server.Close()
	client := a112MBUSTestClient(t, server)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if _, err := A112MBUSOrderbook(ctx, client); err != nil {
		t.Fatal(err)
	}
	if _, err := A112MBUSCalendar(ctx, client, "2026-08-14"); err != nil {
		t.Fatal(err)
	}
	if got, want := strings.Join(paths, ","), "/api/v1/orderbook?symbol=AAPL,/api/v1/market-calendar/US?date=2026-08-14"; got != want {
		t.Fatalf("paths = %q, want %q", got, want)
	}
}

func TestA112MBUSDoesNotWidenRawMinuteCandlesUSRefusal(t *testing.T) {
	if _, err := (&Client{}).RawMinuteCandles(context.Background(), "US", "AAPL", 200, "", false); err == nil {
		t.Fatal("M-B0 widened existing KR-only raw reader")
	}
}

func a112MBUSGoodRateHeaders() http.Header {
	return http.Header{"X-RateLimit-Limit": {"10"}, "X-RateLimit-Remaining": {"9"}, "X-RateLimit-Reset": {"1"}}
}

func a112MBUSTestClient(t *testing.T, server *httptest.Server) *Client {
	t.Helper()
	transport := server.Client().Transport.(*http.Transport).Clone()
	transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true} // httptest server certificate only
	target := server.Listener.Addr().String()
	transport.DialContext = func(ctx context.Context, network, _ string) (net.Conn, error) {
		return (&net.Dialer{}).DialContext(ctx, network, target)
	}
	cache := filepath.Join(t.TempDir(), "token.json")
	data, err := json.Marshal(cachedToken{AccessToken: "cached", ExpiresAt: time.Now().Add(time.Hour)})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(cache, data, 0o600); err != nil {
		t.Fatal(err)
	}
	httpClient := &http.Client{Transport: transport}
	client := &Client{base: defaultBaseURL, hc: httpClient, tm: newTokenManager(Credentials{}, defaultBaseURL, cache, httpClient), authorityOrigin: true, authorityTransport: transport, configurationSealed: true}
	return client
}
