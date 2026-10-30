//go:build linux

package accountbudget

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"testing"
	"time"
)

func TestTransportTemplateDoesNotMutateOfficialAndAcceptsA112Clone(t *testing.T) {
	b, _ := fixture(t)
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("ok")) }))
	defer server.Close()
	pool := x509.NewCertPool()
	pool.AddCert(server.Certificate())
	base := &http.Transport{Proxy: http.ProxyFromEnvironment, DialContext: (&net.Dialer{Timeout: 30 * time.Second, KeepAlive: 30 * time.Second}).DialContext, ForceAttemptHTTP2: true, MaxIdleConns: 100, IdleConnTimeout: 90 * time.Second, TLSHandshakeTimeout: 10 * time.Second, ExpectContinueTimeout: time.Second, TLSClientConfig: &tls.Config{RootCAs: pool}}
	if _, e := NewTransport(b, base, func(error) {}); e != nil {
		t.Fatal(e)
	}
	if base.TLSNextProto != nil || len(base.TLSClientConfig.NextProtos) != 0 {
		t.Fatalf("input changed: next=%v protos=%v", base.TLSNextProto, base.TLSClientConfig.NextProtos)
	}
	clone := base.Clone()
	before := append([]string(nil), clone.TLSClientConfig.NextProtos...)
	tr, e := NewTransport(b, clone, func(error) {})
	if e != nil {
		t.Fatalf("official clone rejected: %v; alpn=%v", e, before)
	}
	if !reflect.DeepEqual(before, clone.TLSClientConfig.NextProtos) || clone.TLSNextProto != nil {
		t.Fatal("clone mutated")
	}
	req, _ := http.NewRequest("GET", server.URL, nil)
	resp, e := tr.RoundTrip(req)
	if e != nil {
		t.Fatal(e)
	}
	if resp.ProtoMajor != 1 || resp.TLS == nil || len(resp.TLS.VerifiedChains) == 0 {
		t.Fatal("not verified HTTP/1 TLS")
	}
	if _, e = io.Copy(io.Discard, resp.Body); e != nil {
		t.Fatal(e)
	}
	if e = resp.Body.Close(); e != nil {
		t.Fatal(e)
	}
}
func TestTransportTemplateRejectsUnsupportedHooks(t *testing.T) {
	for _, base := range []*http.Transport{
		{TLSNextProto: map[string]func(string, *tls.Conn) http.RoundTripper{}},
		{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}},
		{TLSClientConfig: &tls.Config{NextProtos: []string{"arbitrary"}}},
		{TLSClientConfig: &tls.Config{VerifyConnection: func(tls.ConnectionState) error { return nil }}},
	} {
		b, _ := fixture(t)
		if _, e := NewTransport(b, base, func(error) {}); e == nil {
			t.Fatal("unsupported transport accepted")
		}
	}
}

func TestTransportTemplatePreservesHTTP1Fields(t *testing.T) {
	b, _ := fixture(t)
	base := &http.Transport{Proxy: http.ProxyFromEnvironment, DialContext: (&net.Dialer{Timeout: 19 * time.Second}).DialContext, TLSHandshakeTimeout: 7 * time.Second, ResponseHeaderTimeout: 8 * time.Second, ExpectContinueTimeout: 9 * time.Second, IdleConnTimeout: 11 * time.Second, MaxIdleConns: 12, MaxIdleConnsPerHost: 13, MaxConnsPerHost: 14, MaxResponseHeaderBytes: 15, WriteBufferSize: 16, ReadBufferSize: 17, DisableCompression: true, ForceAttemptHTTP2: true}
	tr, e := NewTransport(b, base, func(error) {})
	if e != nil {
		t.Fatal(e)
	}
	x := tr.base
	if x.Proxy == nil || x.DialContext == nil || x.TLSHandshakeTimeout != base.TLSHandshakeTimeout || x.ResponseHeaderTimeout != base.ResponseHeaderTimeout || x.ExpectContinueTimeout != base.ExpectContinueTimeout || x.IdleConnTimeout != base.IdleConnTimeout || x.MaxIdleConns != base.MaxIdleConns || x.MaxIdleConnsPerHost != base.MaxIdleConnsPerHost || x.MaxConnsPerHost != base.MaxConnsPerHost || x.MaxResponseHeaderBytes != base.MaxResponseHeaderBytes || x.WriteBufferSize != base.WriteBufferSize || x.ReadBufferSize != base.ReadBufferSize || x.DisableCompression != base.DisableCompression {
		t.Fatal("HTTP/1 setting lost")
	}
	if !x.DisableKeepAlives || x.ForceAttemptHTTP2 || x.TLSNextProto == nil || base.TLSNextProto != nil || base.TLSClientConfig != nil {
		t.Fatal("template policy or original changed")
	}
}
func TestTransportTemplateRejectsProxyAndProtocolHooks(t *testing.T) {
	b, _ := fixture(t)
	for _, base := range []*http.Transport{
		{HTTP2: &http.HTTP2Config{}},
		{Protocols: &http.Protocols{}},
		{ProxyConnectHeader: http.Header{"X-Test": []string{"x"}}},
		{GetProxyConnectHeader: func(context.Context, *url.URL, string) (http.Header, error) { return nil, nil }},
		{TLSClientConfig: &tls.Config{NextProtos: []string{"http/1.1", "h2"}}},
	} {
		if _, e := NewTransport(b, base, func(error) {}); e == nil {
			t.Fatal("unsupported configuration accepted")
		}
	}
}

func TestTransportRejectsTLSCallbacks(t *testing.T) {
	b, _ := fixture(t)
	for name, cfg := range map[string]*tls.Config{
		"clock":         {Time: func() time.Time { return time.Now() }},
		"client-cert":   {GetClientCertificate: func(*tls.CertificateRequestInfo) (*tls.Certificate, error) { return nil, nil }},
		"server-cert":   {GetCertificate: func(*tls.ClientHelloInfo) (*tls.Certificate, error) { return nil, nil }},
		"server-config": {GetConfigForClient: func(*tls.ClientHelloInfo) (*tls.Config, error) { return nil, nil }},
	} {
		t.Run(name, func(t *testing.T) {
			if _, e := NewTransport(b, &http.Transport{TLSClientConfig: cfg}, func(error) {}); e == nil {
				t.Fatal("TLS callback accepted")
			}
		})
	}
}
