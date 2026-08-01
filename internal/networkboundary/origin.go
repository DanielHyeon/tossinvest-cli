// Package networkboundary centralizes the transport, peer and origin identity
// rules shared by the browser console and the private HTTP API.
package networkboundary

import (
	"fmt"
	"net"
	"net/http"
	"net/netip"
	"net/url"
	"strconv"
	"strings"
)

// Origin is the canonical browser origin identity. Path, query and fragment are
// deliberately absent: they are resources below an origin, not part of it.
type Origin struct {
	Scheme string
	Host   string
	Port   uint16
}

func (o Origin) String() string {
	return o.Scheme + "://" + net.JoinHostPort(o.Host, strconv.Itoa(int(o.Port)))
}

// ParseOrigin accepts an HTTP(S) origin and makes its effective port explicit.
// It rejects ambiguous authority syntax, credentials and resource components.
func ParseOrigin(raw string) (Origin, error) {
	return parseURLOrigin(raw, false)
}

// MustParseOrigin is for static/test configuration only.
func MustParseOrigin(raw string) Origin {
	origin, err := ParseOrigin(raw)
	if err != nil {
		panic(err)
	}
	return origin
}

func parseURLOrigin(raw string, allowResource bool) (Origin, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" || strings.EqualFold(raw, "null") {
		return Origin{}, fmt.Errorf("network boundary: opaque or empty origin")
	}
	u, err := url.Parse(raw)
	if err != nil {
		return Origin{}, fmt.Errorf("network boundary: invalid origin: %w", err)
	}
	if u.Opaque != "" || u.User != nil || u.Scheme == "" || u.Host == "" {
		return Origin{}, fmt.Errorf("network boundary: origin must have an unambiguous authority and no userinfo")
	}
	if !allowResource && u.Path != "" && u.Path != "/" {
		return Origin{}, fmt.Errorf("network boundary: origin must not contain a path")
	}
	if !allowResource && (u.RawQuery != "" || u.ForceQuery || u.Fragment != "") {
		return Origin{}, fmt.Errorf("network boundary: origin must not contain query or fragment")
	}
	scheme := strings.ToLower(u.Scheme)
	if scheme != "https" && scheme != "http" {
		return Origin{}, fmt.Errorf("network boundary: origin scheme must be http or https")
	}
	host := strings.ToLower(u.Hostname())
	if err := validateHost(host); err != nil {
		return Origin{}, err
	}
	portText := u.Port()
	if portText == "" {
		if scheme == "https" {
			portText = "443"
		} else {
			portText = "80"
		}
	}
	port, err := strconv.ParseUint(portText, 10, 16)
	if err != nil || port == 0 {
		return Origin{}, fmt.Errorf("network boundary: invalid effective port %q", portText)
	}
	return Origin{Scheme: scheme, Host: host, Port: uint16(port)}, nil
}

func validateHost(host string) error {
	if host == "" || strings.HasSuffix(host, ".") || strings.ContainsAny(host, "\x00\r\n\t /\\@%") {
		return fmt.Errorf("network boundary: invalid canonical host %q", host)
	}
	if _, err := netip.ParseAddr(host); err == nil {
		return nil
	}
	for _, label := range strings.Split(host, ".") {
		if label == "" || len(label) > 63 || label[0] == '-' || label[len(label)-1] == '-' {
			return fmt.Errorf("network boundary: invalid canonical host %q", host)
		}
		for _, r := range label {
			if r > 127 || !(r == '-' || r >= 'a' && r <= 'z' || r >= '0' && r <= '9') {
				return fmt.Errorf("network boundary: invalid canonical host %q", host)
			}
		}
	}
	if len(host) > 253 {
		return fmt.Errorf("network boundary: invalid canonical host %q", host)
	}
	return nil
}

// OriginMatches applies strict Origin -> Referer -> transport precedence. A
// malformed or repeated higher-precedence header is a final refusal.
func OriginMatches(r *http.Request, expected Origin, trustedProxy *netip.Addr) bool {
	if r == nil {
		return false
	}
	if explicit, present := ExplicitOriginMatches(r, expected); present {
		return explicit
	}
	actual, err := requestTransportOrigin(r, trustedProxy)
	return err == nil && actual == expected
}

// ExplicitOriginMatches evaluates browser-supplied Origin/Referer evidence. The
// second result distinguishes total header absence from explicit invalid input,
// so callers may allow a transport fallback only in the former case.
func ExplicitOriginMatches(r *http.Request, expected Origin) (matches bool, present bool) {
	if r == nil {
		return false, false
	}
	if values, present := headerValues(r.Header, "Origin"); present {
		value, ok := oneHeader(values, false)
		if !ok {
			return false, true
		}
		actual, err := ParseOrigin(value)
		return err == nil && actual == expected, true
	}
	if values, present := headerValues(r.Header, "Referer"); present {
		value, ok := oneHeader(values, false)
		if !ok {
			return false, true
		}
		actual, err := parseURLOrigin(value, true)
		return err == nil && actual == expected, true
	}
	return false, false
}

func requestTransportOrigin(r *http.Request, trustedProxy *netip.Addr) (Origin, error) {
	peer, err := directPeer(r)
	if err != nil {
		return Origin{}, err
	}
	if trustedProxy != nil && peer == trustedProxy.Unmap() {
		protoValues, protoPresent := headerValues(r.Header, "X-Forwarded-Proto")
		hostValues, hostPresent := headerValues(r.Header, "X-Forwarded-Host")
		if !protoPresent || !hostPresent {
			return Origin{}, fmt.Errorf("network boundary: trusted proxy omitted forwarded origin")
		}
		proto, ok := oneHeader(protoValues, true)
		if !ok {
			return Origin{}, fmt.Errorf("network boundary: ambiguous forwarded proto")
		}
		host, ok := oneHeader(hostValues, true)
		if !ok {
			return Origin{}, fmt.Errorf("network boundary: ambiguous forwarded host")
		}
		return ParseOrigin(proto + "://" + host)
	}
	if hasForwardingHeaders(r.Header) {
		return Origin{}, fmt.Errorf("network boundary: forwarding headers from an untrusted peer")
	}
	if r.TLS == nil {
		return Origin{}, fmt.Errorf("network boundary: direct request is not TLS")
	}
	return ParseOrigin("https://" + r.Host)
}

func directPeer(r *http.Request) (netip.Addr, error) {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return netip.Addr{}, fmt.Errorf("network boundary: remote peer is not host:port: %w", err)
	}
	peer, err := netip.ParseAddr(host)
	if err != nil {
		return netip.Addr{}, fmt.Errorf("network boundary: remote peer is not an IP literal: %w", err)
	}
	return peer.Unmap(), nil
}

func headerValues(header http.Header, name string) ([]string, bool) {
	values, ok := header[http.CanonicalHeaderKey(name)]
	return values, ok
}

func oneHeader(values []string, rejectComma bool) (string, bool) {
	if len(values) != 1 {
		return "", false
	}
	value := strings.TrimSpace(values[0])
	if value == "" || strings.ContainsAny(value, "\r\n") || rejectComma && strings.Contains(value, ",") {
		return "", false
	}
	return value, true
}

func hasForwardingHeaders(header http.Header) bool {
	for _, name := range []string{"Forwarded", "X-Forwarded-For", "X-Forwarded-Host", "X-Forwarded-Proto", "X-Forwarded-Port"} {
		if _, ok := headerValues(header, name); ok {
			return true
		}
	}
	return false
}
