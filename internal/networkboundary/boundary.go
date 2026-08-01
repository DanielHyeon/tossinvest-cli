package networkboundary

import (
	"fmt"
	"net/http"
	"net/netip"
	"strings"
)

// ServerConfig declares the private transport boundary. TLSConfigured means
// TossOS terminates TLS itself. TLSForwarded is only valid for one exact proxy
// peer and never implies trust in a subnet or forwarding chain.
type ServerConfig struct {
	Bind          string
	AllowedCIDRs  []string
	PublicOrigin  string
	TrustedProxy  string
	TLSConfigured bool
	TLSForwarded  bool
}

// Boundary is an immutable, validated network policy.
type Boundary struct {
	bind         netip.Addr
	allowed      []netip.Prefix
	publicOrigin Origin
	trustedProxy *netip.Addr
}

// RejectionWriter owns the public response contract for a request refused by
// the private transport boundary. Higher layers should inject their stable
// error writer so the boundary cannot bypass an API's JSON error envelope.
type RejectionWriter func(http.ResponseWriter, *http.Request)

func New(cfg ServerConfig) (*Boundary, error) {
	bind, err := netip.ParseAddr(strings.TrimSpace(cfg.Bind))
	if err != nil {
		return nil, fmt.Errorf("network boundary: bind must be an IP literal: %w", err)
	}
	bind = bind.Unmap()
	if bind.IsUnspecified() {
		return nil, fmt.Errorf("network boundary: refusing wildcard bind %s; use an explicit loopback/private IP", bind)
	}
	if !privateAddress(bind) {
		return nil, fmt.Errorf("network boundary: refusing public bind %s", bind)
	}
	if len(cfg.AllowedCIDRs) == 0 {
		return nil, fmt.Errorf("network boundary: at least one private allowed CIDR is required")
	}
	allowed := make([]netip.Prefix, 0, len(cfg.AllowedCIDRs))
	for _, raw := range cfg.AllowedCIDRs {
		prefix, err := netip.ParsePrefix(strings.TrimSpace(raw))
		if err != nil {
			return nil, fmt.Errorf("network boundary: invalid allowed CIDR %q: %w", raw, err)
		}
		prefix = prefix.Masked()
		if prefix.Bits() == 0 || !privatePrefix(prefix) {
			return nil, fmt.Errorf("network boundary: refusing non-private allowed CIDR %q", raw)
		}
		allowed = append(allowed, prefix)
	}
	publicOrigin, err := ParseOrigin(cfg.PublicOrigin)
	if err != nil {
		return nil, err
	}
	if publicOrigin.Scheme != "https" {
		return nil, fmt.Errorf("network boundary: public origin must use HTTPS")
	}
	if cfg.TLSConfigured == cfg.TLSForwarded {
		return nil, fmt.Errorf("network boundary: configure exactly one TLS termination mode")
	}
	var trustedProxy *netip.Addr
	if strings.TrimSpace(cfg.TrustedProxy) != "" {
		parsed, err := netip.ParseAddr(strings.TrimSpace(cfg.TrustedProxy))
		if err != nil {
			return nil, fmt.Errorf("network boundary: trusted proxy must be an exact IP literal: %w", err)
		}
		parsed = parsed.Unmap()
		if !privateAddress(parsed) {
			return nil, fmt.Errorf("network boundary: trusted proxy must be loopback/private")
		}
		trustedProxy = &parsed
	}
	if cfg.TLSForwarded && trustedProxy == nil {
		return nil, fmt.Errorf("network boundary: forwarded TLS requires one exact trusted proxy")
	}
	if cfg.TLSConfigured && trustedProxy != nil {
		return nil, fmt.Errorf("network boundary: direct TLS mode does not accept a trusted proxy")
	}
	return &Boundary{bind: bind, allowed: allowed, publicOrigin: publicOrigin, trustedProxy: trustedProxy}, nil
}

func privateAddress(addr netip.Addr) bool {
	return addr.IsLoopback() || addr.IsPrivate() || netip.MustParsePrefix("100.64.0.0/10").Contains(addr)
}

func privatePrefix(prefix netip.Prefix) bool {
	for _, allowed := range []netip.Prefix{
		netip.MustParsePrefix("127.0.0.0/8"),
		netip.MustParsePrefix("10.0.0.0/8"),
		netip.MustParsePrefix("172.16.0.0/12"),
		netip.MustParsePrefix("192.168.0.0/16"),
		netip.MustParsePrefix("100.64.0.0/10"),
		netip.MustParsePrefix("::1/128"),
		netip.MustParsePrefix("fc00::/7"),
	} {
		if prefix.Bits() >= allowed.Bits() && allowed.Contains(prefix.Addr()) {
			return true
		}
	}
	return false
}

func (b *Boundary) Bind() netip.Addr { return b.bind }

func (b *Boundary) PublicOrigin() Origin { return b.publicOrigin }

// ClientPeer returns the security peer identity. Forwarded-For is accepted only
// from the single configured TCP peer and must contain exactly one IP, never a
// comma-delimited chain.
func (b *Boundary) ClientPeer(r *http.Request) (netip.Addr, error) {
	peer, err := directPeer(r)
	if err != nil {
		return netip.Addr{}, err
	}
	if b.trustedProxy != nil && peer == *b.trustedProxy {
		if _, present := headerValues(r.Header, "Forwarded"); present {
			return netip.Addr{}, fmt.Errorf("network boundary: RFC Forwarded is unsupported; use exact X-Forwarded fields")
		}
		if _, present := headerValues(r.Header, "X-Forwarded-Port"); present {
			return netip.Addr{}, fmt.Errorf("network boundary: separate forwarded port is ambiguous")
		}
		values, present := headerValues(r.Header, "X-Forwarded-For")
		value, ok := oneHeader(values, true)
		if !present || !ok {
			return netip.Addr{}, fmt.Errorf("network boundary: trusted proxy must supply one exact forwarded peer")
		}
		client, err := netip.ParseAddr(value)
		if err != nil {
			return netip.Addr{}, fmt.Errorf("network boundary: forwarded peer must be an IP literal: %w", err)
		}
		return client.Unmap(), nil
	}
	if hasForwardingHeaders(r.Header) {
		return netip.Addr{}, fmt.Errorf("network boundary: untrusted peer supplied forwarding headers")
	}
	return peer, nil
}

func (b *Boundary) PeerAllowed(r *http.Request) bool {
	peer, err := b.ClientPeer(r)
	if err != nil {
		return false
	}
	for _, prefix := range b.allowed {
		if prefix.Contains(peer) {
			return true
		}
	}
	return false
}

func (b *Boundary) OriginMatches(r *http.Request) bool {
	if _, err := b.ClientPeer(r); err != nil {
		return false
	}
	transport, err := requestTransportOrigin(r, b.trustedProxy)
	if err != nil || transport != b.publicOrigin {
		return false
	}
	return OriginMatches(r, b.publicOrigin, b.trustedProxy)
}

// Handler enforces the private peer and canonical TLS transport boundary for
// read-only REST/SSE traffic. Authentication and mutation policy remain
// separate higher-layer concerns.
func (b *Boundary) Handler(next http.Handler) http.Handler {
	return b.HandlerWithRejection(next, defaultRejectionWriter)
}

// HandlerWithRejection composes the private peer/origin boundary in front of a
// router while leaving the externally visible error schema with the owning API.
// A nil writer fails closed with the package's stable JSON fallback.
func (b *Boundary) HandlerWithRejection(next http.Handler, reject RejectionWriter) http.Handler {
	if next == nil {
		next = http.NotFoundHandler()
	}
	if reject == nil {
		reject = defaultRejectionWriter
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !b.PeerAllowed(r) || !b.OriginMatches(r) {
			reject(w, r)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func defaultRejectionWriter(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Referrer-Policy", "no-referrer")
	w.WriteHeader(http.StatusForbidden)
	if r != nil && r.Method == http.MethodHead {
		return
	}
	_, _ = w.Write([]byte(`{"error":{"code":"BOUNDARY_REFUSED","message":"private transport boundary refused request"}}`))
}
