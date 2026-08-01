package networkboundary

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"strings"
	"time"
)

// LoadServerCertificate validates the exact certificate identity used by a
// private listener before the socket opens. Private VPN CAs need not be in the
// host trust store, so this validates the matching key, leaf lifetime, server
// EKU and canonical DNS/IP SAN without building a public trust chain.
func LoadServerCertificate(certFile, keyFile, host string, now time.Time) (tls.Certificate, error) {
	if strings.TrimSpace(certFile) == "" || strings.TrimSpace(keyFile) == "" || strings.TrimSpace(host) == "" {
		return tls.Certificate{}, errors.New("network boundary: certificate, key and public host are required")
	}
	certificate, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("network boundary: loading TLS certificate: %w", err)
	}
	if len(certificate.Certificate) == 0 {
		return tls.Certificate{}, errors.New("network boundary: TLS certificate chain is empty")
	}
	leaf, err := x509.ParseCertificate(certificate.Certificate[0])
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("network boundary: parsing TLS certificate: %w", err)
	}
	now = now.UTC()
	if now.Before(leaf.NotBefore) || now.After(leaf.NotAfter) {
		return tls.Certificate{}, fmt.Errorf("network boundary: TLS certificate is outside its validity window")
	}
	if !serverCertificateUsage(leaf.ExtKeyUsage) {
		return tls.Certificate{}, errors.New("network boundary: TLS certificate is not valid for server authentication")
	}
	if err := leaf.VerifyHostname(host); err != nil {
		return tls.Certificate{}, fmt.Errorf("network boundary: TLS certificate does not cover public host %q: %w", host, err)
	}
	certificate.Leaf = leaf
	return certificate, nil
}

func serverCertificateUsage(usages []x509.ExtKeyUsage) bool {
	if len(usages) == 0 {
		return true
	}
	for _, usage := range usages {
		if usage == x509.ExtKeyUsageAny || usage == x509.ExtKeyUsageServerAuth {
			return true
		}
	}
	return false
}
