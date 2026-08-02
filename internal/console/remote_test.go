package console

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"io"
	"math/big"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/verifylive"
)

const (
	testRemoteHost  = "console.vpn.test:37085"
	testRemoteToken = "0123456789abcdef0123456789abcdef"
)

func noopRemoteStarter(
	context.Context,
	verifylive.BatchConfirmer,
	io.Writer,
	string,
	[]verifylive.StepID,
) (verifylive.Summary, []verifylive.Entry, error) {
	return verifylive.Summary{}, nil, nil
}

type remoteTestRig struct {
	console *Console
	now     *time.Time
	audits  *[]RemoteAccessEvent
}

type acceptingRemoteHandoff struct{}

func (acceptingRemoteHandoff) Mint(time.Time) (string, error) { return "handoff-ok", nil }
func (acceptingRemoteHandoff) Consume(token string, _ time.Time) error {
	if token != "handoff-ok" {
		return errors.New("wrong handoff")
	}
	return nil
}

// These tests fetch a screen, not the root. The root became a 303 to the
// overview in change a054, and a request that always redirects can no longer
// tell an authorised session (200) apart from an unauthorised one (303 to
// /login) — which is the entire distinction every assertion below is making.
func newRemoteTestRig(t *testing.T, mutate ...func(*RemoteAccess)) remoteTestRig {
	t.Helper()
	cert, key := writeRemoteTestCertificate(t, "console.vpn.test")
	now := time.Date(2026, 7, 31, 1, 2, 3, 0, time.UTC)
	audits := []RemoteAccessEvent{}
	remote := RemoteAccess{
		Bind:         "0.0.0.0",
		AllowedCIDRs: []string{"10.8.0.0/24"},
		PublicURL:    "https://" + testRemoteHost,
		TLSCertFile:  cert,
		TLSKeyFile:   key,
		AccessToken:  testRemoteToken,
		RecordAccess: func(event RemoteAccessEvent) error {
			audits = append(audits, event)
			return nil
		},
	}
	for _, change := range mutate {
		change(&remote)
	}
	c, err := New(Options{
		StartVerify: noopRemoteStarter,
		Now:         func() time.Time { return now },
		Remote:      remote,
	})
	if err != nil {
		t.Fatalf("New remote console: %v", err)
	}
	return remoteTestRig{console: c, now: &now, audits: &audits}
}

func writeRemoteTestCertificate(t *testing.T, host string) (string, string) {
	t.Helper()
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	serial, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 120))
	if err != nil {
		t.Fatal(err)
	}
	template := x509.Certificate{
		SerialNumber: serial,
		Subject:      pkix.Name{CommonName: host},
		DNSNames:     []string{host},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}
	der, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		t.Fatal(err)
	}
	keyDER, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	certPath := filepath.Join(dir, "tls.crt")
	keyPath := filepath.Join(dir, "tls.key")
	if err := os.WriteFile(certPath, pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der}), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(keyPath, pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: keyDER}), 0o600); err != nil {
		t.Fatal(err)
	}
	return certPath, keyPath
}

func remoteRequest(method, target, peer, userAgent string, form url.Values) *http.Request {
	var body io.Reader
	if form != nil {
		body = strings.NewReader(form.Encode())
	}
	r := httptest.NewRequest(method, "https://"+testRemoteHost+target, body)
	r.Host = testRemoteHost
	r.RemoteAddr = peer
	r.Header.Set("User-Agent", userAgent)
	if form != nil {
		r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}
	return r
}

func serveRemote(c *Console, r *http.Request) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	c.Handler().ServeHTTP(w, r)
	return w
}

func remoteLogin(t *testing.T, c *Console, peer, userAgent string) *http.Cookie {
	t.Helper()
	r := remoteRequest(http.MethodPost, "/login", peer, userAgent, url.Values{"token": {testRemoteToken}})
	r.Header.Set("Origin", "https://"+testRemoteHost)
	w := serveRemote(c, r)
	if w.Code != http.StatusSeeOther {
		t.Fatalf("login status = %d, body=%s", w.Code, w.Body.String())
	}
	for _, cookie := range w.Result().Cookies() {
		if cookie.Name == sessionCookie {
			return cookie
		}
	}
	t.Fatal("login did not set a console session cookie")
	return nil
}

func newTrustedNetworkTestRig(t *testing.T) remoteTestRig {
	t.Helper()
	return newRemoteTestRig(t, func(remote *RemoteAccess) {
		remote.TrustedNetwork = true
		remote.AccessToken = ""
	})
}

func TestTrustedNetworkConsoleNeedsNoApplicationSession(t *testing.T) {
	rig := newTrustedNetworkTestRig(t)
	request := remoteRequest(http.MethodGet, pathOverview, "10.8.0.42:4321", "mobile", nil)
	response := serveRemote(rig.console, request)
	if response.Code != http.StatusOK {
		t.Fatalf("trusted dashboard status = %d, body=%s", response.Code, response.Body.String())
	}
	if location := response.Header().Get("Location"); location != "" {
		t.Fatalf("trusted dashboard redirected to %q", location)
	}
	for _, cookie := range response.Result().Cookies() {
		if cookie.Name == sessionCookie {
			t.Fatal("trusted dashboard issued an application session cookie")
		}
	}
	if got := rig.console.URL(); got != "https://"+testRemoteHost+"/" {
		t.Fatalf("trusted console URL = %q", got)
	}

	login := remoteRequest(http.MethodGet, "/login", "10.8.0.42:4321", "mobile", nil)
	loginResponse := serveRemote(rig.console, login)
	if strings.Contains(loginResponse.Body.String(), "원격 로그인") {
		t.Fatal("trusted-network mode exposed the application login page")
	}
}

func TestTrustedNetworkStillRejectsWrongPeerOriginAndCSRF(t *testing.T) {
	rig := newTrustedNetworkTestRig(t)

	outside := remoteRequest(http.MethodGet, pathOverview, "192.168.1.20:4321", "mobile", nil)
	if got := serveRemote(rig.console, outside).Code; got != http.StatusForbidden {
		t.Fatalf("outside peer status = %d, want 403", got)
	}

	crossOrigin := remoteRequest(
		http.MethodPost,
		"/verify/start",
		"10.8.0.42:4321",
		"mobile",
		url.Values{"csrf": {rig.console.csrf}},
	)
	crossOrigin.Header.Set("Origin", "https://evil.example")
	if got := serveRemote(rig.console, crossOrigin).Code; got != http.StatusForbidden {
		t.Fatalf("cross-origin status = %d, want 403", got)
	}

	noCSRF := remoteRequest(
		http.MethodPost,
		"/verify/start",
		"10.8.0.42:4321",
		"mobile",
		url.Values{},
	)
	noCSRF.Header.Set("Origin", "https://"+testRemoteHost)
	if got := serveRemote(rig.console, noCSRF).Code; got != http.StatusForbidden {
		t.Fatalf("missing-CSRF status = %d, want 403", got)
	}
}

func TestTrustedNetworkAndTokenAuthenticationCannotBeCombined(t *testing.T) {
	_, err := New(Options{
		StartVerify: noopRemoteStarter,
		Remote: RemoteAccess{
			Bind:           "0.0.0.0",
			AllowedCIDRs:   []string{"10.8.0.0/24"},
			PublicURL:      "https://" + testRemoteHost,
			TLSCertFile:    "unused.crt",
			TLSKeyFile:     "unused.key",
			AccessToken:    testRemoteToken,
			TrustedNetwork: true,
			RecordAccess:   func(RemoteAccessEvent) error { return nil },
		},
	})
	if err == nil {
		t.Fatal("trusted-network and token authentication were accepted together")
	}
}

func TestRemoteConfigurationIsAllOrNothing(t *testing.T) {
	_, err := New(Options{
		StartVerify: noopRemoteStarter,
		Remote:      RemoteAccess{Bind: "10.8.0.1"},
	})
	if err == nil {
		t.Fatal("a bind-only remote configuration was accepted")
	}
}

func TestRemoteConfigurationRejectsGlobalCIDRAndCertificateNameMismatch(t *testing.T) {
	cert, key := writeRemoteTestCertificate(t, "somewhere-else.vpn.test")
	base := RemoteAccess{
		Bind:         "0.0.0.0",
		AllowedCIDRs: []string{"0.0.0.0/0"},
		PublicURL:    "https://" + testRemoteHost,
		TLSCertFile:  cert,
		TLSKeyFile:   key,
		AccessToken:  testRemoteToken,
		RecordAccess: func(RemoteAccessEvent) error { return nil },
	}
	if _, err := New(Options{StartVerify: noopRemoteStarter, Remote: base}); err == nil {
		t.Fatal("a global allowed CIDR was accepted")
	}
	base.AllowedCIDRs = []string{"10.8.0.0/24"}
	if _, err := New(Options{StartVerify: noopRemoteStarter, Remote: base}); err == nil {
		t.Fatal("a certificate for another host was accepted")
	}
}

func TestRemoteListenerMustMatchTheValidatedBind(t *testing.T) {
	rig := newRemoteTestRig(t)
	matching, err := ListenOn("0.0.0.0", 0)
	if err != nil {
		t.Fatal(err)
	}
	defer matching.Close()
	if err := rig.console.listenerAllowed(matching); err != nil {
		t.Fatalf("matching remote listener refused: %v", err)
	}

	different, err := ListenOn("127.0.0.1", 0)
	if err != nil {
		t.Fatal(err)
	}
	defer different.Close()
	if err := rig.console.listenerAllowed(different); err == nil {
		t.Fatal("remote listener on a different bind address was accepted")
	}
}

func TestRemoteURLNeverCarriesAConsoleCredential(t *testing.T) {
	rig := newRemoteTestRig(t)
	got := rig.console.URL()
	if got != "https://"+testRemoteHost+"/login" {
		t.Fatalf("remote URL = %q", got)
	}
	if strings.Contains(got, testRemoteToken) || strings.Contains(got, rig.console.SessionToken()) {
		t.Fatal("remote URL contains a credential")
	}
}

func TestRemoteLoginIssuesADistinctBoundSecureSession(t *testing.T) {
	rig := newRemoteTestRig(t)
	cookie := remoteLogin(t, rig.console, "10.8.0.9:44000", "mobile-a")
	if cookie.Value == testRemoteToken {
		t.Fatal("the long-lived login credential became the session cookie")
	}
	if !cookie.Secure || !cookie.HttpOnly || cookie.SameSite != http.SameSiteStrictMode {
		t.Fatalf("cookie is not hardened: %+v", cookie)
	}
	if len(*rig.audits) != 1 || (*rig.audits)[0].Action != RemoteAccessLogin {
		t.Fatalf("login audit = %+v", *rig.audits)
	}

	ok := remoteRequest(http.MethodGet, pathOverview, "10.8.0.9:44001", "mobile-a", nil)
	ok.AddCookie(cookie)
	if status := serveRemote(rig.console, ok).Code; status != http.StatusOK {
		t.Fatalf("bound session status = %d, want 200", status)
	}
	replay := remoteRequest(http.MethodGet, pathOverview, "10.8.0.9:44002", "mobile-b", nil)
	replay.AddCookie(cookie)
	if status := serveRemote(rig.console, replay).Code; status != http.StatusSeeOther {
		t.Fatalf("different User-Agent status = %d, want login redirect", status)
	}
}

func TestRemoteSessionExpiresAndLogoutRevokesIt(t *testing.T) {
	rig := newRemoteTestRig(t)
	cookie := remoteLogin(t, rig.console, "10.8.0.10:44000", "mobile")
	*rig.now = rig.now.Add(remoteSessionIdleTTL + time.Second)
	expired := remoteRequest(http.MethodGet, pathOverview, "10.8.0.10:44001", "mobile", nil)
	expired.AddCookie(cookie)
	if status := serveRemote(rig.console, expired).Code; status != http.StatusSeeOther {
		t.Fatalf("expired session status = %d, want login redirect", status)
	}

	cookie = remoteLogin(t, rig.console, "10.8.0.10:44002", "mobile")
	logout := remoteRequest(http.MethodPost, "/logout", "10.8.0.10:44003", "mobile",
		url.Values{"csrf": {rig.console.csrf}})
	logout.Header.Set("Origin", "https://"+testRemoteHost)
	logout.AddCookie(cookie)
	if status := serveRemote(rig.console, logout).Code; status != http.StatusSeeOther {
		t.Fatalf("logout status = %d", status)
	}
	after := remoteRequest(http.MethodGet, pathOverview, "10.8.0.10:44004", "mobile", nil)
	after.AddCookie(cookie)
	if status := serveRemote(rig.console, after).Code; status != http.StatusSeeOther {
		t.Fatalf("revoked session status = %d", status)
	}
}

func TestRemoteSessionHasAnAbsoluteExpiryDespiteActivity(t *testing.T) {
	rig := newRemoteTestRig(t)
	cookie := remoteLogin(t, rig.console, "10.8.0.10:44000", "mobile")
	for i := 0; i < 16; i++ {
		*rig.now = rig.now.Add(29 * time.Minute)
		active := remoteRequest(http.MethodGet, pathOverview, "10.8.0.10:44001", "mobile", nil)
		active.AddCookie(cookie)
		if status := serveRemote(rig.console, active).Code; status != http.StatusOK {
			t.Fatalf("active session expired at step %d with status %d", i+1, status)
		}
	}
	*rig.now = rig.now.Add(17 * time.Minute)
	expired := remoteRequest(http.MethodGet, pathOverview, "10.8.0.10:44002", "mobile", nil)
	expired.AddCookie(cookie)
	if status := serveRemote(rig.console, expired).Code; status != http.StatusSeeOther {
		t.Fatalf("absolute-expired session status = %d, want login redirect", status)
	}
}

func TestRemoteHandoffIssuesANewAuditedRemoteSession(t *testing.T) {
	rig := newRemoteTestRig(t)
	rig.console.opts.Handoff = acceptingRemoteHandoff{}
	r := remoteRequest(http.MethodGet, "/?handoff=handoff-ok", "10.8.0.16:44000", "mobile", nil)
	w := serveRemote(rig.console, r)
	if w.Code != http.StatusSeeOther {
		t.Fatalf("handoff status = %d", w.Code)
	}
	cookies := w.Result().Cookies()
	if len(cookies) != 1 || !cookies[0].Secure || cookies[0].Value == rig.console.SessionToken() {
		t.Fatalf("handoff cookie = %+v", cookies)
	}
	if got := (*rig.audits)[0]; got.Action != RemoteAccessLogin || got.Detail != "restart handoff" {
		t.Fatalf("handoff audit = %+v", got)
	}
}

func TestRemoteLoginAuditFailureDoesNotIssueASession(t *testing.T) {
	rig := newRemoteTestRig(t, func(remote *RemoteAccess) {
		remote.RecordAccess = func(RemoteAccessEvent) error { return errors.New("disk full") }
	})
	r := remoteRequest(http.MethodPost, "/login", "10.8.0.11:44000", "mobile",
		url.Values{"token": {testRemoteToken}})
	r.Header.Set("Origin", "https://"+testRemoteHost)
	w := serveRemote(rig.console, r)
	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("audit failure status = %d, want 503", w.Code)
	}
	if len(w.Result().Cookies()) != 0 {
		t.Fatal("audit failure issued a cookie")
	}
}

func TestRemotePeerHostOriginAndCSRFAreIndependentGates(t *testing.T) {
	rig := newRemoteTestRig(t)
	cookie := remoteLogin(t, rig.console, "10.8.0.12:44000", "mobile")

	outside := remoteRequest(http.MethodGet, pathOverview, "203.0.113.9:44000", "mobile", nil)
	outside.Header.Set("X-Forwarded-For", "10.8.0.12")
	if status := serveRemote(rig.console, outside).Code; status != http.StatusForbidden {
		t.Fatalf("outside peer status = %d", status)
	}

	badHost := remoteRequest(http.MethodGet, pathOverview, "10.8.0.12:44001", "mobile", nil)
	badHost.Host = "attacker.invalid"
	badHost.AddCookie(cookie)
	if status := serveRemote(rig.console, badHost).Code; status != http.StatusForbidden {
		t.Fatalf("bad Host status = %d", status)
	}

	badOrigin := remoteRequest(http.MethodPost, "/optimization/exit-policy",
		"10.8.0.12:44002", "mobile", url.Values{"csrf": {rig.console.csrf}})
	badOrigin.Header.Set("Origin", "https://attacker.invalid")
	badOrigin.AddCookie(cookie)
	if status := serveRemote(rig.console, badOrigin).Code; status != http.StatusForbidden {
		t.Fatalf("bad Origin status = %d", status)
	}

	refererOnly := remoteRequest(http.MethodPost, "/optimization/exit-policy",
		"10.8.0.12:44003", "mobile", url.Values{"csrf": {rig.console.csrf}})
	refererOnly.Header.Set("Referer", "https://"+testRemoteHost+"/optimization")
	refererOnly.AddCookie(cookie)
	if status := serveRemote(rig.console, refererOnly).Code; status == http.StatusForbidden {
		t.Fatalf("same-origin Referer fallback was refused")
	}
}

func TestRemoteLoginIsRateLimitedByActualPeer(t *testing.T) {
	rig := newRemoteTestRig(t)
	for i := 0; i < remoteLoginLimit; i++ {
		r := remoteRequest(http.MethodPost, "/login", "10.8.0.13:44000", "mobile",
			url.Values{"token": {"wrong-token-value-that-is-long-enough"}})
		r.Header.Set("Origin", "https://"+testRemoteHost)
		r.Header.Set("X-Forwarded-For", "10.8.0."+string(rune('A'+i)))
		if status := serveRemote(rig.console, r).Code; status != http.StatusForbidden {
			t.Fatalf("attempt %d status = %d", i+1, status)
		}
	}
	r := remoteRequest(http.MethodPost, "/login", "10.8.0.13:44001", "mobile",
		url.Values{"token": {testRemoteToken}})
	r.Header.Set("Origin", "https://"+testRemoteHost)
	w := serveRemote(rig.console, r)
	if w.Code != http.StatusTooManyRequests {
		t.Fatalf("sixth attempt status = %d, want 429", w.Code)
	}
	if w.Header().Get("Retry-After") == "" {
		t.Fatal("rate limit response has no Retry-After")
	}
	for i := 0; i < 3; i++ {
		repeated := remoteRequest(http.MethodPost, "/login", "10.8.0.13:44002", "mobile",
			url.Values{"token": {testRemoteToken}})
		repeated.Header.Set("Origin", "https://"+testRemoteHost)
		serveRemote(rig.console, repeated)
	}
	rateAudits := 0
	for _, event := range *rig.audits {
		if event.Action == RemoteAccessRateLimited {
			rateAudits++
		}
	}
	if rateAudits != 1 {
		t.Fatalf("rate-limit audits = %d, want one per window", rateAudits)
	}
}

func TestRemoteResponsesHaveSecurityHeadersAndHealthIsMinimal(t *testing.T) {
	rig := newRemoteTestRig(t)
	login := remoteRequest(http.MethodGet, "/login", "10.8.0.14:44000", "mobile", nil)
	w := serveRemote(rig.console, login)
	for _, header := range []string{
		"Strict-Transport-Security", "Content-Security-Policy", "X-Frame-Options",
		"X-Content-Type-Options", "Referrer-Policy", "Cache-Control",
	} {
		if w.Header().Get(header) == "" {
			t.Errorf("missing %s", header)
		}
	}
	if got := w.Header().Get("Referrer-Policy"); got != "same-origin" {
		t.Errorf("Referrer-Policy = %q, want same-origin", got)
	}

	health := remoteRequest(http.MethodGet, "/healthz", "127.0.0.1:45000", "health", nil)
	health.Host = "127.0.0.1:37085"
	hw := serveRemote(rig.console, health)
	if hw.Code != http.StatusOK || hw.Body.String() != "ok\n" {
		t.Fatalf("health = %d %q", hw.Code, hw.Body.String())
	}
	post := remoteRequest(http.MethodPost, "/healthz", "127.0.0.1:45001", "health", nil)
	post.Host = "127.0.0.1:37085"
	if status := serveRemote(rig.console, post).Code; status != http.StatusMethodNotAllowed {
		t.Fatalf("POST health status = %d", status)
	}
}

func TestRemoteQuerySessionCredentialIsNotAccepted(t *testing.T) {
	rig := newRemoteTestRig(t)
	r := remoteRequest(http.MethodGet, "/?session="+rig.console.SessionToken(),
		"10.8.0.15:44000", "mobile", nil)
	w := serveRemote(rig.console, r)
	if w.Code != http.StatusSeeOther || w.Header().Get("Location") != "/login" {
		t.Fatalf("query session = %d location %q", w.Code, w.Header().Get("Location"))
	}
}
