package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHTTPAPICommandIsRegisteredReadOnlyAndNonMutating(t *testing.T) {
	cmd := findCommand(newRootCmd(), "httpapi")
	if cmd == nil {
		t.Fatal("httpapi command is not registered")
	}
	if cmd.Annotations["source"] != "both" {
		t.Fatalf("source annotation = %q, want both", cmd.Annotations["source"])
	}
	if cmd.Annotations["mutating"] == "true" {
		t.Fatal("httpapi must not claim LIVE/account mutation authority")
	}
	for _, forbidden := range []string{"live", "gate", "kill", "order", "flatten", "engine-start", "engine-stop"} {
		if child := findCommand(cmd, forbidden); child != nil {
			t.Fatalf("forbidden remote authority registered as subcommand %q", forbidden)
		}
	}
}

func TestHTTPAPIDefaultsStayLoopbackNoTokenReadOnlyAndRequireTLS(t *testing.T) {
	cmd := newHTTPAPICmd(&rootOptions{})
	for _, name := range []string{"bind", "allowed-cidr", "public-url", "tls-cert", "tls-key", "trusted-proxy", "tls-forwarded"} {
		if cmd.Flag(name) == nil {
			t.Fatalf("missing flag --%s", name)
		}
	}
	if got := cmd.Flag("bind").DefValue; got != defaultHTTPAPIBind {
		t.Fatalf("bind default = %q", got)
	}
	if got := cmd.Flag("public-url").DefValue; got != defaultHTTPAPIPublicURL {
		t.Fatalf("public URL default = %q", got)
	}
	if cmd.Flag("publish-interval") != nil {
		t.Fatal("server-owned stream polling interval became client configurable")
	}
	if _, _, err := httpAPIBoundary(&httpAPIOptions{port: defaultHTTPAPIPort, bind: defaultHTTPAPIBind,
		allowedCIDRs: []string{"127.0.0.0/8"}, publicURL: defaultHTTPAPIPublicURL}); err == nil {
		t.Fatal("daemon accepted a listener without direct or forwarded TLS")
	}
}

func TestHTTPAPIBoundaryRejectsPublicAndAmbiguousTLSConfiguration(t *testing.T) {
	valid := httpAPIOptions{port: defaultHTTPAPIPort, bind: defaultHTTPAPIBind,
		allowedCIDRs: []string{"127.0.0.0/8"}, publicURL: defaultHTTPAPIPublicURL,
		tlsCert: "cert.pem", tlsKey: "key.pem"}
	if boundary, direct, err := httpAPIBoundary(&valid); err != nil || boundary == nil || !direct {
		t.Fatalf("direct TLS boundary = %v direct=%v err=%v", boundary, direct, err)
	}
	for name, mutate := range map[string]func(*httpAPIOptions){
		"public bind":      func(o *httpAPIOptions) { o.bind = "8.8.8.8" },
		"public CIDR":      func(o *httpAPIOptions) { o.allowedCIDRs = []string{"0.0.0.0/0"} },
		"certificate only": func(o *httpAPIOptions) { o.tlsKey = "" },
		"both TLS modes":   func(o *httpAPIOptions) { o.trustedProxy, o.tlsForwarded = "127.0.0.2", true },
	} {
		t.Run(name, func(t *testing.T) {
			candidate := valid
			candidate.allowedCIDRs = append([]string(nil), valid.allowedCIDRs...)
			mutate(&candidate)
			if _, _, err := httpAPIBoundary(&candidate); err == nil {
				t.Fatal("unsafe boundary accepted")
			}
		})
	}
}

func TestPrivateBoundaryHidesRouterFromUnapprovedPeers(t *testing.T) {
	boundary, _, err := httpAPIBoundary(&httpAPIOptions{port: defaultHTTPAPIPort, bind: defaultHTTPAPIBind,
		allowedCIDRs: []string{"127.0.0.0/8"}, publicURL: defaultHTTPAPIPublicURL,
		tlsCert: "cert.pem", tlsKey: "key.pem"})
	if err != nil {
		t.Fatal(err)
	}
	called := false
	handler := boundary.Handler(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { called = true }))
	request := httptest.NewRequest(http.MethodGet, defaultHTTPAPIPublicURL+"/api/v1/engine", nil)
	request.RemoteAddr = "203.0.113.7:43210"
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusForbidden || called {
		t.Fatalf("status=%d downstream=%v", recorder.Code, called)
	}
}

func TestHTTPAPICommandConstructionHasNoRuntimeSideEffects(t *testing.T) {
	cmd := newHTTPAPICmd(&rootOptions{})
	if cmd.RunE == nil {
		t.Fatal("httpapi command has no runtime handler")
	}
	// Cobra construction is the only action under test: a listener, journal or
	// broker may be opened only after RunE. Help/command discovery must remain a
	// pure operation because every root test constructs the full command tree.
	if err := cmd.Help(); err != nil {
		t.Fatalf("Help: %v", err)
	}
}
