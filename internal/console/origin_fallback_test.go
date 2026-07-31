package console

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestRemoteSameOriginEvidencePrecedence(t *testing.T) {
	rig := newRemoteTestRig(t)
	canonical := "https://" + testRemoteHost

	tests := []struct {
		name    string
		origin  []string
		referer []string
		want    bool
	}{
		{name: "canonical origin", origin: []string{canonical}, want: true},
		{name: "cross origin", origin: []string{"https://attacker.invalid"}, want: false},
		{name: "empty origin is explicit", origin: []string{""}, referer: []string{canonical + "/settings"}, want: false},
		{name: "whitespace origin is explicit", origin: []string{" \t "}, referer: []string{canonical + "/settings"}, want: false},
		{name: "multiple origins", origin: []string{canonical, canonical}, want: false},
		{name: "canonical origin ignores malformed referer", origin: []string{canonical}, referer: []string{"://broken"}, want: true},
		{name: "canonical origin ignores cross origin referer", origin: []string{canonical}, referer: []string{"https://attacker.invalid"}, want: true},
		{name: "wrong origin is not rescued", origin: []string{"https://attacker.invalid"}, referer: []string{canonical + "/restart"}, want: false},
		{name: "empty origin is not rescued", origin: []string{""}, referer: []string{canonical + "/restart"}, want: false},
		{name: "settings referer path", referer: []string{canonical + "/settings"}, want: true},
		{name: "optimization referer path", referer: []string{canonical + "/optimization?tab=exit#policy"}, want: true},
		{name: "restart referer path", referer: []string{canonical + "/restart"}, want: true},
		{name: "cross origin referer", referer: []string{"https://attacker.invalid/restart"}, want: false},
		{name: "empty referer", referer: []string{""}, want: false},
		{name: "whitespace referer", referer: []string{" \t "}, want: false},
		{name: "malformed referer", referer: []string{"://broken"}, want: false},
		{name: "relative referer", referer: []string{"/restart"}, want: false},
		{name: "multiple referers", referer: []string{canonical + "/restart", canonical + "/settings"}, want: false},
		{name: "no evidence", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := remoteRequest(http.MethodPost, "/restart", "10.8.0.12:44010", "mobile", nil)
			if tt.origin != nil {
				r.Header[http.CanonicalHeaderKey("Origin")] = tt.origin
			}
			if tt.referer != nil {
				r.Header[http.CanonicalHeaderKey("Referer")] = tt.referer
			}
			if got := rig.console.remote.sameOrigin(r); got != tt.want {
				t.Fatalf("sameOrigin = %t, want %t (Origin=%q Referer=%q)",
					got, tt.want, tt.origin, tt.referer)
			}
		})
	}
}

func TestHeaderlessRemoteLoginRemainsStrict(t *testing.T) {
	rig := newRemoteTestRig(t)
	r := remoteRequest(http.MethodPost, "/login", "10.8.0.12:44020", "mobile",
		url.Values{"token": {testRemoteToken}})
	w := serveRemote(rig.console, r)
	if w.Code != http.StatusForbidden {
		t.Fatalf("headerless login status = %d, want 403", w.Code)
	}
	if len(w.Result().Cookies()) != 0 {
		t.Fatal("headerless login issued a session cookie")
	}
	if len(*rig.audits) != 0 {
		t.Fatalf("headerless login wrote %d audit event(s)", len(*rig.audits))
	}
	rig.console.remote.mu.Lock()
	attempts := len(rig.console.remote.attempts)
	sessions := len(rig.console.remote.sessions)
	rig.console.remote.mu.Unlock()
	if attempts != 0 || sessions != 0 {
		t.Fatalf("headerless login mutated attempts=%d sessions=%d", attempts, sessions)
	}
}

func TestHeaderlessCanonicalTLSPostReachesCSRFAndHandlerGates(t *testing.T) {
	rig := newRemoteTestRig(t)
	called := 0
	handler := rig.console.mutating(func(w http.ResponseWriter, _ *http.Request) {
		called++
		w.WriteHeader(http.StatusNoContent)
	})

	valid := remoteRequest(http.MethodPost, "/restart", "10.8.0.12:44030", "mobile",
		url.Values{"csrf": {rig.console.csrf}})
	validResponse := httptest.NewRecorder()
	handler(validResponse, valid)
	if validResponse.Code != http.StatusNoContent {
		t.Fatalf("headerless canonical POST = %d, want 204; body=%s",
			validResponse.Code, validResponse.Body.String())
	}
	if called != 1 {
		t.Fatalf("handler calls = %d, want 1", called)
	}

	wrongCSRF := remoteRequest(http.MethodPost, "/restart", "10.8.0.12:44031", "mobile",
		url.Values{"csrf": {"wrong"}})
	wrongResponse := httptest.NewRecorder()
	handler(wrongResponse, wrongCSRF)
	if wrongResponse.Code != http.StatusForbidden ||
		!strings.Contains(wrongResponse.Body.String(), "CSRF 토큰이 없거나 일치하지 않는다") {
		t.Fatalf("wrong-CSRF response = %d %q", wrongResponse.Code, wrongResponse.Body.String())
	}
	if called != 1 {
		t.Fatalf("wrong CSRF reached handler; calls=%d", called)
	}
}

func TestExplicitOpaqueOriginCannotReachMutationHandler(t *testing.T) {
	rig := newRemoteTestRig(t)
	called := 0
	handler := rig.console.mutating(func(w http.ResponseWriter, _ *http.Request) {
		called++
		w.WriteHeader(http.StatusNoContent)
	})

	r := remoteRequest(http.MethodPost, "/restart", "10.8.0.12:44035", "mobile",
		url.Values{"csrf": {rig.console.csrf}})
	r.Header.Set("Origin", "null")
	w := httptest.NewRecorder()
	handler(w, r)

	if w.Code != http.StatusForbidden ||
		!strings.Contains(w.Body.String(), "요청 출처가 일치하지 않는다") {
		t.Fatalf("opaque-origin response = %d %q, want origin refusal", w.Code, w.Body.String())
	}
	if called != 0 {
		t.Fatalf("opaque origin reached mutation handler %d time(s)", called)
	}
}

func TestRemoteMutationOriginFallbackRejectsIndirectEvidence(t *testing.T) {
	rig := newRemoteTestRig(t)
	called := 0
	handler := rig.console.mutating(func(w http.ResponseWriter, _ *http.Request) {
		called++
		w.WriteHeader(http.StatusNoContent)
	})

	tests := []struct {
		name   string
		mutate func(*http.Request)
	}{
		{
			name: "forwarded protocol cannot replace direct TLS",
			mutate: func(r *http.Request) {
				r.TLS = nil
				r.Header.Set("X-Forwarded-Proto", "https")
				r.Header.Set("X-Forwarded-Host", testRemoteHost)
			},
		},
		{
			name: "forwarded host cannot replace request host",
			mutate: func(r *http.Request) {
				r.Host = "attacker.invalid:37085"
				r.Header.Set("X-Forwarded-Host", testRemoteHost)
				r.Header.Set("X-Forwarded-Proto", "https")
			},
		},
		{
			name: "wrong port",
			mutate: func(r *http.Request) {
				r.Host = "console.vpn.test:443"
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := remoteRequest(http.MethodPost, "/restart", "10.8.0.12:44040", "mobile",
				url.Values{"csrf": {rig.console.csrf}})
			tt.mutate(r)
			w := httptest.NewRecorder()
			handler(w, r)
			if w.Code != http.StatusForbidden ||
				!strings.Contains(w.Body.String(), "요청 출처가 일치하지 않는다") {
				t.Fatalf("response = %d %q, want origin refusal", w.Code, w.Body.String())
			}
		})
	}
	if called != 0 {
		t.Fatalf("indirect origin evidence reached handler %d time(s)", called)
	}
}
