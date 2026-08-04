package official

import (
	"context"
	"crypto/tls"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
)

func TestAuthoritativeAccountIdentityRejectsConfiguredOriginBeforeHTTP(t *testing.T) {
	hits := 0
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		hits++
		http.Error(writer, "configured origin must not be authority", http.StatusInternalServerError)
	}))
	defer server.Close()
	client := New(Credentials{APIKey: "key", SecretKey: "secret"}, filepath.Join(t.TempDir(), "token.json"),
		WithBaseURL(server.URL), WithHTTPClient(server.Client()), WithAccountSeq(7))

	if err := client.VerifyAuthoritativeAccountIdentity(context.Background(), "7"); !errors.Is(err, ErrAuthorityOrigin) {
		t.Fatalf("error = %v, want ErrAuthorityOrigin", err)
	}
	if hits != 0 {
		t.Fatalf("configured origin received %d requests", hits)
	}
}

func TestAuthoritativeAccountIdentityReverifiesExactSelectedAccount(t *testing.T) {
	client, closeServer := authoritativeAccountTestClient(t,
		`{"result":[{"accountNo":"123-45","accountSeq":7,"accountType":"BROKERAGE"}]}`,
		WithAccountSeq(7))
	defer closeServer()

	if err := client.VerifyAuthoritativeAccountIdentity(context.Background(), "7"); err != nil {
		t.Fatal(err)
	}
}

func TestAuthoritativeAccountIdentityRejectsAmbiguousScope(t *testing.T) {
	tests := []struct {
		name      string
		payload   string
		accountID string
		options   []Option
	}{
		{name: "missing", payload: `{"result":[{"accountNo":"123-45","accountSeq":8,"accountType":"BROKERAGE"}]}`, accountID: "7", options: []Option{WithAccountSeq(7)}},
		{name: "duplicate", payload: `{"result":[{"accountNo":"a","accountSeq":7,"accountType":"BROKERAGE"},{"accountNo":"b","accountSeq":7,"accountType":"BROKERAGE"}]}`, accountID: "7", options: []Option{WithAccountSeq(7)}},
		{name: "selected mismatch", payload: `{"result":[{"accountNo":"123-45","accountSeq":7,"accountType":"BROKERAGE"}]}`, accountID: "7", options: []Option{WithAccountSeq(8)}},
		{name: "noncanonical", payload: `{"result":[]}`, accountID: "07", options: []Option{WithAccountSeq(7)}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			client, closeServer := authoritativeAccountTestClient(t, test.payload, test.options...)
			defer closeServer()
			if err := client.VerifyAuthoritativeAccountIdentity(context.Background(), test.accountID); !errors.Is(err, ErrAccountIdentityAuthority) {
				t.Fatalf("error = %v, want ErrAccountIdentityAuthority", err)
			}
		})
	}
}

func authoritativeAccountTestClient(t *testing.T, accountPayload string, options ...Option) (*Client, func()) {
	t.Helper()
	server := httptest.NewTLSServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		switch request.URL.Path {
		case "/oauth2/token":
			_, _ = writer.Write([]byte(`{"access_token":"AT","expires_in":3600,"token_type":"Bearer"}`))
		case "/api/v1/accounts":
			_, _ = writer.Write([]byte(accountPayload))
		default:
			http.NotFound(writer, request)
		}
	}))
	transport := server.Client().Transport.(*http.Transport).Clone()
	transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true} // test server certificate only
	target := server.Listener.Addr().String()
	transport.DialContext = func(ctx context.Context, network, _ string) (net.Conn, error) {
		return (&net.Dialer{}).DialContext(ctx, network, target)
	}
	httpClient := &http.Client{Transport: transport}
	client := &Client{base: defaultBaseURL, hc: httpClient, authorityOrigin: true, authorityTransport: transport, configurationSealed: true}
	for _, option := range options {
		// Tests build the production-origin client directly, so apply only the
		// account selection before installing the token manager.
		client.configurationSealed = false
		option(client)
		client.configurationSealed = true
	}
	client.tm = newTokenManager(Credentials{APIKey: "key", SecretKey: "secret"}, defaultBaseURL, filepath.Join(t.TempDir(), "token.json"), httpClient)
	return client, server.Close
}
