package networkboundary

import (
	"crypto/tls"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCanonicalOriginUsesSchemeHostAndEffectivePort(t *testing.T) {
	t.Parallel()

	httpsDefault, err := ParseOrigin("https://Console.VPN.Test")
	if err != nil {
		t.Fatal(err)
	}
	httpsExplicit, err := ParseOrigin("https://console.vpn.test:443")
	if err != nil {
		t.Fatal(err)
	}
	if httpsDefault != httpsExplicit {
		t.Fatalf("default and explicit HTTPS ports differ: %+v != %+v", httpsDefault, httpsExplicit)
	}
	if got := httpsDefault.String(); got != "https://console.vpn.test:443" {
		t.Fatalf("canonical origin = %q", got)
	}
	if _, err := ParseOrigin("https://user@example.test:443"); err == nil {
		t.Fatal("origin with userinfo was accepted")
	}
	if _, err := ParseOrigin("null"); err == nil {
		t.Fatal("opaque origin was accepted")
	}
	if _, err := ParseOrigin("https://example.test:443/path"); err == nil {
		t.Fatal("origin with path was accepted")
	}
}

func TestOriginPrecedenceRejectsMalformedExplicitEvidence(t *testing.T) {
	t.Parallel()
	want := MustParseOrigin("https://console.vpn.test:443")

	tests := []struct {
		name   string
		origin []string
		ref    []string
	}{
		{name: "opaque origin does not fall back", origin: []string{"null"}, ref: []string{"https://console.vpn.test/ok"}},
		{name: "empty origin does not fall back", origin: []string{""}, ref: []string{"https://console.vpn.test/ok"}},
		{name: "repeated origin", origin: []string{"https://console.vpn.test", "https://console.vpn.test"}},
		{name: "repeated referer", ref: []string{"https://console.vpn.test/a", "https://console.vpn.test/b"}},
		{name: "referer userinfo", ref: []string{"https://user@console.vpn.test/a"}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodPost, "https://console.vpn.test/api/v1/settings", nil)
			r.Header["Origin"] = tc.origin
			r.Header["Referer"] = tc.ref
			if OriginMatches(r, want, nil) {
				t.Fatal("malformed higher-precedence origin evidence was accepted")
			}
		})
	}
}

func TestOriginIgnoresRefererPathAndRejectsDifferentPort(t *testing.T) {
	t.Parallel()
	want := MustParseOrigin("https://console.vpn.test:37085")
	r := httptest.NewRequest(http.MethodPost, "https://console.vpn.test:37085/api/v1/settings", nil)
	r.Header.Set("Referer", "https://console.vpn.test:37085/positions?tab=open#today")
	if !OriginMatches(r, want, nil) {
		t.Fatal("same scheme/host/port referer with another path was rejected")
	}
	r.Header.Set("Referer", "https://console.vpn.test:37086/positions")
	if OriginMatches(r, want, nil) {
		t.Fatal("referer from another port was accepted")
	}
}

func TestOnlyExactTrustedProxyMaySupplyForwardedBoundary(t *testing.T) {
	t.Parallel()
	cfg := ServerConfig{
		Bind:         "10.0.0.9",
		AllowedCIDRs: []string{"10.8.0.0/24"},
		PublicOrigin: "https://api.vpn.test:443",
		TrustedProxy: "10.0.0.7",
		TLSForwarded: true,
	}
	boundary, err := New(cfg)
	if err != nil {
		t.Fatal(err)
	}

	request := func(peer string) *http.Request {
		r := httptest.NewRequest(http.MethodPost, "http://internal:8080/api/v1/settings", nil)
		r.RemoteAddr = peer
		r.Host = "internal:8080"
		r.TLS = nil
		r.Header.Set("X-Forwarded-For", "10.8.0.42")
		r.Header.Set("X-Forwarded-Proto", "https")
		r.Header.Set("X-Forwarded-Host", "api.vpn.test:443")
		return r
	}

	trusted := request("10.0.0.7:55123")
	peer, err := boundary.ClientPeer(trusted)
	if err != nil || peer.String() != "10.8.0.42" {
		t.Fatalf("trusted proxy peer = %v, %v", peer, err)
	}
	if !boundary.OriginMatches(trusted) {
		t.Fatal("exact trusted proxy origin was rejected")
	}

	for _, peerAddr := range []string{"10.0.0.8:55123", "10.8.0.42:55123"} {
		r := request(peerAddr)
		if _, err := boundary.ClientPeer(r); err == nil {
			t.Fatalf("untrusted peer %s supplied forwarding headers", peerAddr)
		}
		if boundary.OriginMatches(r) {
			t.Fatalf("untrusted peer %s supplied forwarded origin", peerAddr)
		}
	}

	repeated := request("10.0.0.7:55123")
	repeated.Header["X-Forwarded-Host"] = []string{"api.vpn.test:443", "evil.test"}
	if boundary.OriginMatches(repeated) {
		t.Fatal("repeated forwarded host was accepted")
	}
	comma := request("10.0.0.7:55123")
	comma.Header.Set("X-Forwarded-For", "10.8.0.42, 10.0.0.2")
	if _, err := boundary.ClientPeer(comma); err == nil {
		t.Fatal("multi-hop forwarded-for was accepted")
	}
}

func TestServerBoundaryFailsClosedWithoutPrivateTLSBoundary(t *testing.T) {
	t.Parallel()
	tests := []ServerConfig{
		{Bind: "0.0.0.0", AllowedCIDRs: []string{"0.0.0.0/0"}, PublicOrigin: "https://api.test"},
		{Bind: "203.0.113.10", AllowedCIDRs: []string{"10.8.0.0/24"}, PublicOrigin: "https://api.test"},
		{Bind: "0.0.0.0", AllowedCIDRs: []string{"10.8.0.0/24"}, PublicOrigin: "http://api.test"},
		{Bind: "0.0.0.0", AllowedCIDRs: []string{"10.8.0.0/24"}, PublicOrigin: "https://api.test"},
		{Bind: "0.0.0.0", AllowedCIDRs: []string{"10.8.0.0/24"}, PublicOrigin: "https://api.test", TLSForwarded: true},
		{Bind: "::", AllowedCIDRs: []string{"10.8.0.0/24"}, PublicOrigin: "https://api.test", TLSConfigured: true},
	}
	for i, cfg := range tests {
		if _, err := New(cfg); err == nil {
			t.Fatalf("unsafe config %d was accepted: %+v", i, cfg)
		}
	}

	if _, err := New(ServerConfig{
		Bind: "127.0.0.1", AllowedCIDRs: []string{"127.0.0.0/8"},
		PublicOrigin: "https://127.0.0.1:37086", TLSConfigured: true,
	}); err != nil {
		t.Fatalf("loopback TLS boundary rejected: %v", err)
	}
	if _, err := New(ServerConfig{
		Bind: "10.0.0.9", AllowedCIDRs: []string{"10.8.0.0/24"},
		PublicOrigin: "https://api.vpn.test", TrustedProxy: "10.0.0.7", TLSForwarded: true,
	}); err != nil {
		t.Fatalf("private exact-proxy TLS boundary rejected: %v", err)
	}
}

func TestDirectRequestRequiresTLSAndCanonicalHost(t *testing.T) {
	t.Parallel()
	boundary, err := New(ServerConfig{
		Bind: "127.0.0.1", AllowedCIDRs: []string{"127.0.0.0/8"},
		PublicOrigin: "https://localhost:443", TLSConfigured: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	r := httptest.NewRequest(http.MethodPost, "https://localhost/api/v1/settings", nil)
	r.RemoteAddr = "127.0.0.1:50000"
	if !boundary.OriginMatches(r) {
		t.Fatal("direct canonical TLS request rejected")
	}
	r.TLS = nil
	if boundary.OriginMatches(r) {
		t.Fatal("direct plaintext request accepted")
	}
	r.TLS = &tls.ConnectionState{}
	r.Host = "localhost:444"
	if boundary.OriginMatches(r) {
		t.Fatal("direct wrong-port host accepted")
	}
}

func TestBoundaryHandlerRequiresCanonicalForwardedTLSOriginForReads(t *testing.T) {
	t.Parallel()
	boundary, err := New(ServerConfig{
		Bind: "10.0.0.9", AllowedCIDRs: []string{"10.8.0.0/24"},
		PublicOrigin: "https://api.vpn.test:443", TrustedProxy: "10.0.0.7", TLSForwarded: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	called := false
	handler := boundary.Handler(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		called = true
		w.WriteHeader(http.StatusNoContent)
	}))
	request := func() *http.Request {
		r := httptest.NewRequest(http.MethodGet, "http://internal:37086/api/v1/engine", nil)
		r.RemoteAddr = "10.0.0.7:55123"
		r.Header.Set("X-Forwarded-For", "10.8.0.42")
		return r
	}

	for _, mutate := range []func(*http.Request){
		func(r *http.Request) {},
		func(r *http.Request) {
			r.Header.Set("X-Forwarded-Proto", "http")
			r.Header.Set("X-Forwarded-Host", "api.vpn.test:443")
		},
		func(r *http.Request) {
			r.Header.Set("X-Forwarded-Proto", "https")
			r.Header.Set("X-Forwarded-Host", "other.vpn.test:443")
		},
	} {
		called = false
		r := request()
		mutate(r)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, r)
		if w.Code != http.StatusForbidden || called {
			t.Fatalf("non-canonical forwarded read status/called=%d/%v", w.Code, called)
		}
	}

	allowed := request()
	allowed.Header.Set("X-Forwarded-Proto", "https")
	allowed.Header.Set("X-Forwarded-Host", "api.vpn.test:443")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, allowed)
	if w.Code != http.StatusNoContent || !called {
		t.Fatalf("canonical forwarded read status/called=%d/%v", w.Code, called)
	}
}

func TestBoundaryHandlerDefaultRejectionIsJSONAndNoStore(t *testing.T) {
	t.Parallel()
	boundary, err := New(ServerConfig{
		Bind: "127.0.0.1", AllowedCIDRs: []string{"127.0.0.0/8"},
		PublicOrigin: "https://localhost:443", TLSConfigured: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	request := httptest.NewRequest(http.MethodGet, "https://localhost/api/v1/engine", nil)
	request.RemoteAddr = "203.0.113.10:50000"
	recorder := httptest.NewRecorder()
	boundary.Handler(nil).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusForbidden || recorder.Header().Get("Content-Type") != "application/json; charset=utf-8" ||
		recorder.Header().Get("Cache-Control") != "no-store" || !json.Valid(recorder.Body.Bytes()) {
		t.Fatalf("default rejection status=%d headers=%v body=%s", recorder.Code, recorder.Header(), recorder.Body.String())
	}
}
