package console

import (
	"crypto/sha256"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"net/netip"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/networkboundary"
)

const (
	remoteSessionIdleTTL     = 30 * time.Minute
	remoteSessionAbsoluteTTL = 8 * time.Hour
	remoteLoginWindow        = 5 * time.Minute
	remoteLoginLimit         = 5
	remoteSessionLimit       = 128
	remoteAttemptPeerLimit   = 4096
	remoteMaxRequestBody     = 4096
	remoteHSTSMaxAgeSeconds  = 31536000
)

const (
	RemoteAccessLogin       = "console.remote.login"
	RemoteAccessLoginFailed = "console.remote.login_failed"
	RemoteAccessRateLimited = "console.remote.rate_limited"
	RemoteAccessLogout      = "console.remote.logout"
)

// RemoteAccessEvent is the bounded, credential-free event written to the
// operator audit trail for remote authentication.
type RemoteAccessEvent struct {
	Action string
	Peer   string
	Detail string
}

// RemoteAccess is intentionally all-or-nothing. Its zero value keeps the
// original loopback-only console. Supplying any field opts into the remote
// threat model and requires the transport/audit fields plus exactly one access
// mode: TrustedNetwork or AccessToken.
type RemoteAccess struct {
	Bind         string
	AllowedCIDRs []string
	PublicURL    string
	TLSCertFile  string
	TLSKeyFile   string
	AccessToken  string
	// TrustedNetwork makes the already-authenticated host loopback or VPN
	// membership the application access boundary. It never relaxes TLS, peer
	// CIDR, Host, Origin or CSRF checks.
	TrustedNetwork bool
	RecordAccess   func(RemoteAccessEvent) error
}

type remoteSession struct {
	peer      netip.Addr
	userAgent [32]byte
	created   time.Time
	lastSeen  time.Time
}

type remoteAttempts struct {
	at          []time.Time
	rateAudited bool
}

type remoteRuntime struct {
	bind           netip.Addr
	allowed        []netip.Prefix
	publicURL      *url.URL
	origin         string
	certificate    tls.Certificate
	token          string
	trustedNetwork bool
	recordAccess   func(RemoteAccessEvent) error
	now            func() time.Time

	mu       sync.Mutex
	sessions map[[32]byte]remoteSession
	attempts map[netip.Addr]remoteAttempts
}

func newRemoteRuntime(cfg RemoteAccess, now func() time.Time) (*remoteRuntime, error) {
	if remoteAccessEmpty(cfg) {
		return nil, nil
	}
	if err := validateRemoteFields(cfg); err != nil {
		return nil, err
	}
	bind, err := parseRemoteBind(cfg.Bind)
	if err != nil {
		return nil, err
	}
	allowed, err := parseAllowedCIDRs(cfg.AllowedCIDRs)
	if err != nil {
		return nil, err
	}
	publicURL, err := parseRemotePublicURL(cfg.PublicURL)
	if err != nil {
		return nil, err
	}
	certificate, err := loadRemoteCertificate(cfg.TLSCertFile, cfg.TLSKeyFile, publicURL.Hostname())
	if err != nil {
		return nil, err
	}
	token := strings.TrimSpace(cfg.AccessToken)
	if !cfg.TrustedNetwork && len(token) < 32 {
		return nil, fmt.Errorf("console: remote access token must contain at least 32 bytes")
	}
	if now == nil {
		now = func() time.Time { return time.Now().UTC() }
	}
	return &remoteRuntime{
		bind:           bind,
		allowed:        allowed,
		publicURL:      publicURL,
		origin:         publicURL.Scheme + "://" + publicURL.Host,
		certificate:    certificate,
		token:          token,
		trustedNetwork: cfg.TrustedNetwork,
		recordAccess:   cfg.RecordAccess,
		now:            now,
		sessions:       make(map[[32]byte]remoteSession),
		attempts:       make(map[netip.Addr]remoteAttempts),
	}, nil
}

func validateRemoteFields(cfg RemoteAccess) error {
	if strings.TrimSpace(cfg.Bind) == "" || len(cfg.AllowedCIDRs) == 0 ||
		strings.TrimSpace(cfg.PublicURL) == "" || strings.TrimSpace(cfg.TLSCertFile) == "" ||
		strings.TrimSpace(cfg.TLSKeyFile) == "" || cfg.RecordAccess == nil {
		return fmt.Errorf("console: remote access requires bind, allowed CIDR, public URL, TLS certificate/key, and audit recorder")
	}
	token := strings.TrimSpace(cfg.AccessToken)
	if cfg.TrustedNetwork && token != "" {
		return fmt.Errorf("console: trusted-network and token authentication cannot be combined")
	}
	if !cfg.TrustedNetwork && token == "" {
		return fmt.Errorf("console: remote access requires either trusted-network or an access token")
	}
	return nil
}

func parseRemoteBind(raw string) (netip.Addr, error) {
	bind, err := netip.ParseAddr(strings.TrimSpace(raw))
	if err != nil {
		return netip.Addr{}, fmt.Errorf("console: remote bind must be an IP literal: %w", err)
	}
	return bind.Unmap(), nil
}

func parseAllowedCIDRs(rawCIDRs []string) ([]netip.Prefix, error) {
	allowed := make([]netip.Prefix, 0, len(rawCIDRs))
	for _, raw := range rawCIDRs {
		prefix, err := netip.ParsePrefix(strings.TrimSpace(raw))
		if err != nil {
			return nil, fmt.Errorf("console: invalid allowed CIDR %q: %w", raw, err)
		}
		prefix = prefix.Masked()
		if prefix.Bits() == 0 {
			return nil, fmt.Errorf("console: refusing global allowed CIDR %q", raw)
		}
		allowed = append(allowed, prefix)
	}
	return allowed, nil
}

func parseRemotePublicURL(raw string) (*url.URL, error) {
	publicURL, err := url.Parse(strings.TrimSpace(raw))
	if err != nil {
		return nil, fmt.Errorf("console: invalid public URL: %w", err)
	}
	if publicURL.Scheme != "https" || publicURL.Host == "" || publicURL.User != nil ||
		(publicURL.Path != "" && publicURL.Path != "/") ||
		publicURL.RawQuery != "" || publicURL.Fragment != "" {
		return nil, fmt.Errorf("console: public URL must be an HTTPS origin with no credentials, path, query, or fragment")
	}
	if _, err := strconv.Atoi(publicURL.Port()); err != nil {
		return nil, fmt.Errorf("console: public URL must include an explicit TCP port")
	}
	publicURL.Path = ""
	return publicURL, nil
}

func loadRemoteCertificate(certFile, keyFile, host string) (tls.Certificate, error) {
	certificate, err := networkboundary.LoadServerCertificate(certFile, keyFile, host, time.Now())
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("console: %w", err)
	}
	return certificate, nil
}

func remoteAccessEmpty(cfg RemoteAccess) bool {
	return strings.TrimSpace(cfg.Bind) == "" && len(cfg.AllowedCIDRs) == 0 &&
		strings.TrimSpace(cfg.PublicURL) == "" && strings.TrimSpace(cfg.TLSCertFile) == "" &&
		strings.TrimSpace(cfg.TLSKeyFile) == "" && strings.TrimSpace(cfg.AccessToken) == "" &&
		!cfg.TrustedNetwork && cfg.RecordAccess == nil
}

func (rr *remoteRuntime) peer(r *http.Request) (netip.Addr, bool) {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return netip.Addr{}, false
	}
	addr, err := netip.ParseAddr(host)
	if err != nil {
		return netip.Addr{}, false
	}
	return addr.Unmap(), true
}

func (rr *remoteRuntime) peerAllowed(peer netip.Addr) bool {
	for _, prefix := range rr.allowed {
		if prefix.Contains(peer) {
			return true
		}
	}
	return false
}

func (rr *remoteRuntime) sameOrigin(r *http.Request) bool {
	expected, err := networkboundary.ParseOrigin(rr.origin)
	if err != nil {
		return false
	}
	matches, present := networkboundary.ExplicitOriginMatches(r, expected)
	return present && matches
}

func (rr *remoteRuntime) sameOriginForMutation(r *http.Request) bool {
	_, originPresent := r.Header[http.CanonicalHeaderKey("Origin")]
	_, refererPresent := r.Header[http.CanonicalHeaderKey("Referer")]
	if originPresent || refererPresent {
		return rr.sameOrigin(r)
	}
	expected, err := networkboundary.ParseOrigin(rr.origin)
	return err == nil && networkboundary.OriginMatches(r, expected, nil)
}

func (rr *remoteRuntime) security(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Strict-Transport-Security", "max-age="+strconv.Itoa(remoteHSTSMaxAgeSeconds))
		w.Header().Set("Content-Security-Policy", "default-src 'none'; style-src 'unsafe-inline'; form-action 'self'; frame-ancestors 'none'; base-uri 'none'")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Referrer-Policy", "same-origin")
		w.Header().Set("Cache-Control", "no-store")

		peer, ok := rr.peer(r)
		if !ok {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		if r.URL.Path == "/healthz" && peer.IsLoopback() {
			next.ServeHTTP(w, r)
			return
		}
		if !rr.peerAllowed(peer) || !strings.EqualFold(r.Host, rr.publicURL.Host) {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (rr *remoteRuntime) login(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet, http.MethodHead:
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		if r.Method == http.MethodGet {
			_, _ = fmt.Fprint(w, `<!doctype html><html lang="ko"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1"><title>TossOS 원격 로그인</title></head><body><main><h1>TossOS 원격 로그인</h1><p>VPN 연결과 별도로 원격 접근 토큰이 필요합니다.</p><form method="post" action="/login"><label>접근 토큰 <input type="password" name="token" autocomplete="current-password" required></label><button type="submit">로그인</button></form></main></body></html>`)
		}
	case http.MethodPost:
		rr.loginPost(w, r)
	default:
		w.Header().Set("Allow", http.MethodGet+", "+http.MethodHead+", "+http.MethodPost)
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (rr *remoteRuntime) loginPost(w http.ResponseWriter, r *http.Request) {
	if !rr.sameOrigin(r) {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	peer, _ := rr.peer(r)
	if retry, limited, shouldAudit := rr.rateLimited(peer); limited {
		w.Header().Set("Retry-After", strconv.Itoa(int(retry.Round(time.Second)/time.Second)))
		if shouldAudit {
			if err := rr.record(RemoteAccessEvent{Action: RemoteAccessRateLimited, Peer: peer.String(), Detail: "login"}); err != nil {
				http.Error(w, "audit unavailable", http.StatusServiceUnavailable)
				return
			}
		}
		http.Error(w, "too many requests", http.StatusTooManyRequests)
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, remoteMaxRequestBody)
	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	if !tokenEqual(r.PostFormValue("token"), rr.token) {
		rr.recordFailure(peer)
		if err := rr.record(RemoteAccessEvent{Action: RemoteAccessLoginFailed, Peer: peer.String(), Detail: "credential"}); err != nil {
			http.Error(w, "audit unavailable", http.StatusServiceUnavailable)
			return
		}
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	if err := rr.record(RemoteAccessEvent{Action: RemoteAccessLogin, Peer: peer.String(), Detail: "token"}); err != nil {
		http.Error(w, "audit unavailable", http.StatusServiceUnavailable)
		return
	}
	rr.clearFailures(peer)
	rr.issueSession(w, r, peer)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (rr *remoteRuntime) rateLimited(peer netip.Addr) (time.Duration, bool, bool) {
	now := rr.now()
	cutoff := now.Add(-remoteLoginWindow)
	rr.mu.Lock()
	defer rr.mu.Unlock()
	rr.pruneAttemptsLocked(cutoff)
	entry := rr.attempts[peer]
	kept := entry.at[:0]
	for _, at := range entry.at {
		if at.After(cutoff) {
			kept = append(kept, at)
		}
	}
	entry.at = kept
	if len(kept) == 0 {
		entry.rateAudited = false
	}
	rr.attempts[peer] = entry
	if len(entry.at) < remoteLoginLimit {
		return 0, false, false
	}
	retry := entry.at[0].Add(remoteLoginWindow).Sub(now)
	if retry < time.Second {
		retry = time.Second
	}
	shouldAudit := !entry.rateAudited
	entry.rateAudited = true
	rr.attempts[peer] = entry
	return retry, true, shouldAudit
}

func (rr *remoteRuntime) recordFailure(peer netip.Addr) {
	rr.mu.Lock()
	defer rr.mu.Unlock()
	rr.pruneAttemptsLocked(rr.now().Add(-remoteLoginWindow))
	if _, exists := rr.attempts[peer]; !exists && len(rr.attempts) >= remoteAttemptPeerLimit {
		rr.evictOldestAttemptPeerLocked()
	}
	entry := rr.attempts[peer]
	entry.at = append(entry.at, rr.now())
	if len(entry.at) > remoteLoginLimit {
		entry.at = entry.at[len(entry.at)-remoteLoginLimit:]
	}
	rr.attempts[peer] = entry
}

func (rr *remoteRuntime) pruneAttemptsLocked(cutoff time.Time) {
	for peer, entry := range rr.attempts {
		kept := entry.at[:0]
		for _, at := range entry.at {
			if at.After(cutoff) {
				kept = append(kept, at)
			}
		}
		if len(kept) == 0 {
			delete(rr.attempts, peer)
			continue
		}
		entry.at = kept
		rr.attempts[peer] = entry
	}
}

func (rr *remoteRuntime) evictOldestAttemptPeerLocked() {
	var oldestPeer netip.Addr
	var oldest time.Time
	for peer, entry := range rr.attempts {
		if len(entry.at) == 0 {
			delete(rr.attempts, peer)
			return
		}
		if oldest.IsZero() || entry.at[len(entry.at)-1].Before(oldest) {
			oldestPeer = peer
			oldest = entry.at[len(entry.at)-1]
		}
	}
	if oldestPeer.IsValid() {
		delete(rr.attempts, oldestPeer)
	}
}

func (rr *remoteRuntime) clearFailures(peer netip.Addr) {
	rr.mu.Lock()
	delete(rr.attempts, peer)
	rr.mu.Unlock()
}

func (rr *remoteRuntime) issueSession(w http.ResponseWriter, r *http.Request, peer netip.Addr) {
	value := newToken(32)
	now := rr.now()
	rr.mu.Lock()
	rr.pruneSessionsLocked(now)
	if len(rr.sessions) >= remoteSessionLimit {
		rr.evictOldestSessionLocked()
	}
	rr.sessions[sha256.Sum256([]byte(value))] = remoteSession{
		peer:      peer,
		userAgent: sha256.Sum256([]byte(r.UserAgent())),
		created:   now,
		lastSeen:  now,
	}
	rr.mu.Unlock()
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookie,
		Value:    value,
		Path:     "/",
		Expires:  now.Add(remoteSessionAbsoluteTTL),
		MaxAge:   int(remoteSessionAbsoluteTTL / time.Second),
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})
}

func (rr *remoteRuntime) pruneSessionsLocked(now time.Time) {
	for key, session := range rr.sessions {
		if now.Sub(session.lastSeen) > remoteSessionIdleTTL ||
			now.Sub(session.created) > remoteSessionAbsoluteTTL {
			delete(rr.sessions, key)
		}
	}
}

func (rr *remoteRuntime) evictOldestSessionLocked() {
	var oldestKey [32]byte
	var oldest time.Time
	for key, session := range rr.sessions {
		if oldest.IsZero() || session.lastSeen.Before(oldest) {
			oldestKey = key
			oldest = session.lastSeen
		}
	}
	delete(rr.sessions, oldestKey)
}

func (rr *remoteRuntime) hasSession(r *http.Request) bool {
	cookie, err := r.Cookie(sessionCookie)
	if err != nil {
		return false
	}
	peer, ok := rr.peer(r)
	if !ok {
		return false
	}
	key := sha256.Sum256([]byte(cookie.Value))
	now := rr.now()
	rr.mu.Lock()
	defer rr.mu.Unlock()
	session, ok := rr.sessions[key]
	if !ok {
		return false
	}
	if now.Sub(session.lastSeen) > remoteSessionIdleTTL ||
		now.Sub(session.created) > remoteSessionAbsoluteTTL {
		delete(rr.sessions, key)
		return false
	}
	if session.peer != peer || session.userAgent != sha256.Sum256([]byte(r.UserAgent())) {
		return false
	}
	session.lastSeen = now
	rr.sessions[key] = session
	return true
}

func (rr *remoteRuntime) logout(w http.ResponseWriter, r *http.Request) {
	if cookie, err := r.Cookie(sessionCookie); err == nil {
		rr.mu.Lock()
		delete(rr.sessions, sha256.Sum256([]byte(cookie.Value)))
		rr.mu.Unlock()
	}
	peer, _ := rr.peer(r)
	http.SetCookie(w, &http.Cookie{
		Name: sessionCookie, Value: "", Path: "/", MaxAge: -1,
		Secure: true, HttpOnly: true, SameSite: http.SameSiteStrictMode,
	})
	if err := rr.record(RemoteAccessEvent{Action: RemoteAccessLogout, Peer: peer.String(), Detail: "session"}); err != nil {
		http.Error(w, "audit unavailable; session was revoked", http.StatusServiceUnavailable)
		return
	}
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func (rr *remoteRuntime) record(event RemoteAccessEvent) error {
	event.Peer = strings.TrimSpace(event.Peer)
	event.Detail = strings.TrimSpace(event.Detail)
	return rr.recordAccess(event)
}

func remoteHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		w.Header().Set("Allow", http.MethodGet+", "+http.MethodHead)
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	if r.Method == http.MethodGet {
		_, _ = fmt.Fprintln(w, "ok")
	}
}

func (c *Console) handleHealth(w http.ResponseWriter, r *http.Request) {
	remoteHealth(w, r)
}

func (c *Console) handleRemoteLogin(w http.ResponseWriter, r *http.Request) {
	if c.remote == nil {
		http.NotFound(w, r)
		return
	}
	c.remote.login(w, r)
}

func (c *Console) handleRemoteLogout(w http.ResponseWriter, r *http.Request) {
	if c.remote == nil {
		http.NotFound(w, r)
		return
	}
	c.remote.logout(w, r)
}
