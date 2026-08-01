package httpapi

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/networkboundary"
)

func TestCanonicalJSONRejectsDuplicateKeysAndNormalizesObjectOrderAndNumbers(t *testing.T) {
	t.Parallel()
	left, leftDigest, err := canonicalJSON([]byte(`{"b":1.0,"a":{"z":2e1,"x":true}}`))
	if err != nil {
		t.Fatal(err)
	}
	right, rightDigest, err := canonicalJSON([]byte(` { "a" : { "x": true, "z": 20 }, "b": 1 } `))
	if err != nil {
		t.Fatal(err)
	}
	if string(left) != string(right) || leftDigest != rightDigest {
		t.Fatalf("canonical mismatch: %s != %s", left, right)
	}
	for _, invalid := range []string{`{"a":1,"a":2}`, `[]`, `{"a":1} trailing`, `{"n":1e10001}`} {
		if _, _, err := canonicalJSON([]byte(invalid)); err == nil {
			t.Fatalf("invalid JSON accepted: %s", invalid)
		}
	}
}

func TestBrowserMutationRequiresSessionCSRFCanonicalOriginAndAuditBeforeCommand(t *testing.T) {
	t.Parallel()
	security, store, command := newMutationTestSecurity(t, MutationSecurityOptions{
		BrowserSession: func(r *http.Request) (MutationIdentity, bool) {
			cookie, err := r.Cookie("session")
			return MutationIdentity{Actor: "operator:browser", Client: "browser:safari", Mode: AuthModeBrowser}, err == nil && cookie.Value == "ok"
		},
		BrowserCSRF: func(_ *http.Request, token string) bool { return token == "csrf-ok" },
	})
	handler := security.Handler("/api/v1/optimization/previews", command)

	request := mutationRequest(`{"preset":"safe"}`)
	request.AddCookie(&http.Cookie{Name: "session", Value: "ok"})
	request.Header.Set("X-CSRF-Token", "csrf-ok")
	request.Header.Set("Origin", "https://localhost")
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusNoContent || command.calls != 1 {
		t.Fatalf("status/calls=%d/%d body=%s", recorder.Code, command.calls, recorder.Body.String())
	}
	if command.auditCountAtCall != 1 {
		t.Fatalf("audit rows at command=%d, want 1", command.auditCountAtCall)
	}
	if pending, _ := store.PendingCount(context.Background()); pending != 0 {
		t.Fatalf("pending reservations=%d", pending)
	}

	for _, mutate := range []func(*http.Request){
		func(r *http.Request) { r.Header.Del("X-CSRF-Token") },
		func(r *http.Request) { r.Header.Set("Origin", "null") },
		func(r *http.Request) { r.Header["Origin"] = []string{"https://localhost", "https://localhost"} },
		func(r *http.Request) { r.Header.Set("Origin", "https://localhost:444") },
	} {
		r := mutationRequest(`{"preset":"safe"}`)
		r.AddCookie(&http.Cookie{Name: "session", Value: "ok"})
		r.Header.Set("X-CSRF-Token", "csrf-ok")
		r.Header.Set("Origin", "https://localhost")
		mutate(r)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, r)
		if w.Code != http.StatusForbidden {
			t.Fatalf("browser guard status=%d body=%s", w.Code, w.Body.String())
		}
	}
}

func TestSignedCapabilityIsOneTimeAndIdempotencyIsBodyScoped(t *testing.T) {
	t.Parallel()
	publicKey, privateKey, _ := ed25519.GenerateKey(rand.Reader)
	now := time.Date(2026, 8, 1, 1, 2, 3, 0, time.UTC)
	verifier, _ := NewCapabilityVerifier(publicKey, func() time.Time { return now })
	security, _, command := newMutationTestSecurity(t, MutationSecurityOptions{Capability: verifier, Now: func() time.Time { return now }})
	handler := security.Handler("/api/v1/optimization/previews", command)
	r := mutationRequest(`{"preset":"safe"}`)
	canonical, digest, _ := canonicalJSON([]byte(`{"preset":"safe"}`))
	_ = canonical
	claims := CapabilityClaims{
		Version: CapabilityVersion, Nonce: "nonce-0123456789abcdef", Actor: "operator:local", Client: "ios:device-a",
		Method: http.MethodPost, Resource: r.URL.Path, BodyDigest: digest,
		IdempotencyKey: r.Header.Get("Idempotency-Key"), IfMatch: r.Header.Get("If-Match"),
		Audience: "https://localhost:443", IssuedAt: now, ExpiresAt: now.Add(CapabilityTTL),
	}
	token, _ := SignCapability(privateKey, claims)
	r.Header.Set("Authorization", "TossOS-Capability "+token)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)
	if w.Code != http.StatusNoContent || command.calls != 1 {
		t.Fatalf("capability status/calls=%d/%d body=%s", w.Code, command.calls, w.Body.String())
	}
	reuse := mutationRequest(`{"preset":"safe"}`)
	reuse.Header.Set("Authorization", "TossOS-Capability "+token)
	reused := httptest.NewRecorder()
	handler.ServeHTTP(reused, reuse)
	if reused.Code != http.StatusUnauthorized || command.calls != 1 {
		t.Fatalf("reused capability status/calls=%d/%d", reused.Code, command.calls)
	}
}

func TestEnrolledMTLSIdentityRequiresVerifiedChainAndSupportsSafeReplay(t *testing.T) {
	t.Parallel()
	cert := &x509.Certificate{Raw: []byte("enrolled-client-certificate")}
	fingerprint := sha256.Sum256(cert.Raw)
	identity := MutationIdentity{Actor: "operator:mobile", Client: "ios:device-a", Mode: AuthModeMTLS}
	security, _, command := newMutationTestSecurity(t, MutationSecurityOptions{MTLSIdentities: map[string]MutationIdentity{
		hex.EncodeToString(fingerprint[:]): identity,
	}})
	handler := security.Handler("/api/v1/optimization/previews", command)
	r := mutationRequest(`{"preset":"safe"}`)
	r.TLS.VerifiedChains = [][]*x509.Certificate{{cert}}
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)
	if w.Code != http.StatusNoContent || command.calls != 1 {
		t.Fatalf("mTLS status/calls=%d/%d", w.Code, command.calls)
	}
	replay := mutationRequest(`{"preset":"safe"}`)
	replay.TLS.VerifiedChains = [][]*x509.Certificate{{cert}}
	replayed := httptest.NewRecorder()
	handler.ServeHTTP(replayed, replay)
	if replayed.Code != http.StatusNoContent || command.calls != 1 {
		t.Fatalf("idempotent mTLS replay status/calls=%d/%d", replayed.Code, command.calls)
	}
	unverified := mutationRequest(`{"preset":"safe"}`)
	unverified.TLS.PeerCertificates = []*x509.Certificate{cert}
	refused := httptest.NewRecorder()
	handler.ServeHTTP(refused, unverified)
	if refused.Code != http.StatusUnauthorized {
		t.Fatalf("unverified client cert status=%d", refused.Code)
	}
}

func TestMutationPreconditionAndIdempotencyConflictNeverAutoRetry(t *testing.T) {
	t.Parallel()
	security, _, command := newMutationTestSecurity(t, MutationSecurityOptions{
		BrowserSession: func(*http.Request) (MutationIdentity, bool) {
			return MutationIdentity{Actor: "operator:browser", Client: "browser:safari", Mode: AuthModeBrowser}, true
		},
		BrowserCSRF: func(*http.Request, string) bool { return true },
	})
	command.err = &PreconditionError{CurrentVersion: `"9"`}
	handler := security.Handler("/api/v1/optimization/previews", command)
	r := mutationRequest(`{"preset":"safe"}`)
	r.Header.Set("Origin", "https://localhost")
	r.Header.Set("X-CSRF-Token", "csrf-ok")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)
	if w.Code != http.StatusPreconditionFailed || w.Header().Get("ETag") != `"9"` || command.calls != 1 {
		t.Fatalf("precondition status/etag/calls=%d/%q/%d", w.Code, w.Header().Get("ETag"), command.calls)
	}
	conflict := mutationRequest(`{"preset":"other"}`)
	conflict.Header.Set("Origin", "https://localhost")
	conflict.Header.Set("X-CSRF-Token", "csrf-ok")
	conflicted := httptest.NewRecorder()
	handler.ServeHTTP(conflicted, conflict)
	if conflicted.Code != http.StatusConflict || command.calls != 1 {
		t.Fatalf("conflict status/calls=%d/%d", conflicted.Code, command.calls)
	}
}

func TestSemanticCommandErrorsAreStoredAndReplayedWithoutCommandRetry(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		name, code, message string
		status              int
	}{
		{name: "bad request", status: http.StatusBadRequest, code: "INVALID_SELECTION", message: "Choose a current server-owned preset."},
		{name: "forbidden", status: http.StatusForbidden, code: "REMOTE_TRANSITION_REFUSED", message: "Use the local human approval channel."},
	} {
		t.Run(test.name, func(t *testing.T) {
			security, _, command := newMutationTestSecurity(t, MutationSecurityOptions{
				BrowserSession: func(*http.Request) (MutationIdentity, bool) {
					return MutationIdentity{Actor: "operator:browser", Client: "browser:safari", Mode: AuthModeBrowser}, true
				},
				BrowserCSRF: func(*http.Request, string) bool { return true },
			})
			command.err = NewCommandError(test.status, test.code, test.message)
			handler := security.Handler("/api/v1/optimization/previews", command)
			for attempt := range 2 {
				r := mutationRequest(`{"preset":"safe"}`)
				r.Header.Set("Origin", "https://localhost")
				r.Header.Set("X-CSRF-Token", "csrf-ok")
				w := httptest.NewRecorder()
				handler.ServeHTTP(w, r)
				if w.Code != test.status || command.calls != 1 {
					t.Fatalf("attempt=%d status/calls=%d/%d body=%s", attempt, w.Code, command.calls, w.Body.String())
				}
				assertStableErrorBody(t, w.Body.Bytes(), test.code)
			}
		})
	}
}

type recordingCommander struct {
	store            *SecurityStore
	calls            int
	auditCountAtCall int
	err              error
}

func (c *recordingCommander) Execute(_ context.Context, _ AuthorizedMutation) (MutationResult, error) {
	c.calls++
	_ = c.store.db.QueryRow(`SELECT count(*) FROM mutation_audit WHERE stage='authorized'`).Scan(&c.auditCountAtCall)
	return MutationResult{Status: http.StatusNoContent, Version: `"8"`}, c.err
}

func newMutationTestSecurity(t *testing.T, options MutationSecurityOptions) (*MutationSecurity, *SecurityStore, *recordingCommander) {
	t.Helper()
	boundary, err := networkboundary.New(networkboundary.ServerConfig{
		Bind: "127.0.0.1", AllowedCIDRs: []string{"127.0.0.0/8"}, PublicOrigin: "https://localhost", TLSConfigured: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	store, err := OpenSecurityStore(securityStoreTestPath(t))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = store.Close() })
	options.Boundary = boundary
	options.Ledger = store
	if options.Now == nil {
		options.Now = func() time.Time { return time.Date(2026, 8, 1, 1, 2, 3, 0, time.UTC) }
	}
	security, err := NewMutationSecurity(options)
	if err != nil {
		t.Fatal(err)
	}
	command := &recordingCommander{store: store}
	return security, store, command
}

func mutationRequest(body string) *http.Request {
	r := httptest.NewRequest(http.MethodPost, "https://localhost/api/v1/optimization/previews", strings.NewReader(body))
	r.RemoteAddr = "127.0.0.1:50000"
	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("Idempotency-Key", "idem-0123456789abcdef")
	r.Header.Set("If-Match", `"7"`)
	return r
}

func TestLedgerFailurePreventsCommand(t *testing.T) {
	t.Parallel()
	boundary, _ := networkboundary.New(networkboundary.ServerConfig{
		Bind: "127.0.0.1", AllowedCIDRs: []string{"127.0.0.0/8"}, PublicOrigin: "https://localhost", TLSConfigured: true,
	})
	security, err := NewMutationSecurity(MutationSecurityOptions{
		Boundary: boundary, Ledger: failingLedger{},
		BrowserSession: func(*http.Request) (MutationIdentity, bool) {
			return MutationIdentity{Actor: "operator:browser", Client: "browser:safari", Mode: AuthModeBrowser}, true
		},
		BrowserCSRF: func(*http.Request, string) bool { return true },
	})
	if err != nil {
		t.Fatal(err)
	}
	command := &recordingCommander{}
	r := mutationRequest(`{"preset":"safe"}`)
	r.Header.Set("Origin", "https://localhost")
	r.Header.Set("X-CSRF-Token", "csrf-ok")
	w := httptest.NewRecorder()
	security.Handler(r.URL.Path, command).ServeHTTP(w, r)
	if w.Code != http.StatusServiceUnavailable || command.calls != 0 {
		t.Fatalf("ledger failure status/calls=%d/%d", w.Code, command.calls)
	}
	assertStableErrorBody(t, w.Body.Bytes(), "AUDIT_UNAVAILABLE")
}

func TestTransientCompletionFailureRetriesAuditNotCommand(t *testing.T) {
	t.Parallel()
	security, store, command := newMutationTestSecurity(t, MutationSecurityOptions{
		BrowserSession: func(*http.Request) (MutationIdentity, bool) {
			return MutationIdentity{Actor: "operator:browser", Client: "browser:safari", Mode: AuthModeBrowser}, true
		},
		BrowserCSRF: func(*http.Request, string) bool { return true },
	})
	flaky := &failOnceCompletionLedger{inner: store}
	security.ledger = flaky
	r := mutationRequest(`{"preset":"safe"}`)
	r.Header.Set("Origin", "https://localhost")
	r.Header.Set("X-CSRF-Token", "csrf-ok")
	w := httptest.NewRecorder()
	security.Handler(r.URL.Path, command).ServeHTTP(w, r)
	if w.Code != http.StatusNoContent || command.calls != 1 || flaky.completionCalls != 2 {
		t.Fatalf("status/command/completion calls=%d/%d/%d body=%s", w.Code, command.calls, flaky.completionCalls, w.Body.String())
	}
	if pending, err := store.PendingCount(context.Background()); err != nil || pending != 0 {
		t.Fatalf("pending after completion recovery=%d err=%v", pending, err)
	}
}

func TestExhaustedCompletionFailureLeavesPendingWithoutCommandRetry(t *testing.T) {
	t.Parallel()
	security, store, command := newMutationTestSecurity(t, MutationSecurityOptions{
		BrowserSession: func(*http.Request) (MutationIdentity, bool) {
			return MutationIdentity{Actor: "operator:browser", Client: "browser:safari", Mode: AuthModeBrowser}, true
		},
		BrowserCSRF: func(*http.Request, string) bool { return true },
	})
	failing := &failEveryCompletionLedger{inner: store}
	security.ledger = failing
	r := mutationRequest(`{"preset":"safe"}`)
	r.Header.Set("Origin", "https://localhost")
	r.Header.Set("X-CSRF-Token", "csrf-ok")
	w := httptest.NewRecorder()
	security.Handler(r.URL.Path, command).ServeHTTP(w, r)
	if w.Code != http.StatusServiceUnavailable || command.calls != 1 || failing.completionCalls != 3 {
		t.Fatalf("status/command/completion calls=%d/%d/%d body=%s", w.Code, command.calls, failing.completionCalls, w.Body.String())
	}
	assertStableErrorBody(t, w.Body.Bytes(), "AUDIT_UNAVAILABLE")
	if pending, err := store.PendingCount(context.Background()); err != nil || pending != 1 {
		t.Fatalf("pending after exhausted completion=%d err=%v", pending, err)
	}
}

type failOnceCompletionLedger struct {
	inner           MutationLedger
	completionCalls int
}

type failEveryCompletionLedger struct {
	inner           MutationLedger
	completionCalls int
}

func (l *failEveryCompletionLedger) Reserve(ctx context.Context, request MutationLedgerRequest) (MutationReservation, error) {
	return l.inner.Reserve(ctx, request)
}

func (l *failEveryCompletionLedger) Complete(context.Context, int64, StoredMutationResponse) error {
	l.completionCalls++
	return errors.New("injected persistent completion failure")
}

func (l *failOnceCompletionLedger) Reserve(ctx context.Context, request MutationLedgerRequest) (MutationReservation, error) {
	return l.inner.Reserve(ctx, request)
}

func (l *failOnceCompletionLedger) Complete(ctx context.Context, id int64, response StoredMutationResponse) error {
	l.completionCalls++
	if l.completionCalls == 1 {
		return errors.New("injected transient completion failure")
	}
	return l.inner.Complete(ctx, id, response)
}

type failingLedger struct{}

func (failingLedger) Reserve(context.Context, MutationLedgerRequest) (MutationReservation, error) {
	return MutationReservation{}, errors.New("disk full")
}
func (failingLedger) Complete(context.Context, int64, StoredMutationResponse) error { return nil }
