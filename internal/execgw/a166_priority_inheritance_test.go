//go:build linux

package execgw_test

import (
	"context"
	"crypto/x509"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/accountbudget"
	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/official"
)

func TestA166Official401TokenInheritsCritical(t *testing.T) {
	root := t.TempDir()
	if e := os.Chmod(root, 0700); e != nil {
		t.Fatal(e)
	}
	if e := accountbudget.Provision(root, "scope", "instance"); e != nil {
		t.Fatal(e)
	}
	b, e := accountbudget.OpenContext(context.Background(), accountbudget.Config{Root: root, Scope: "scope", Instance: "instance", Clock: clock.System()})
	if e != nil {
		t.Fatal(e)
	}
	defer b.Close()
	entered := make(chan struct{})
	release := make(chan struct{})
	var count atomic.Int32
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/oauth2/token" {
			w.Write([]byte(`{"access_token":"fixture","expires_in":3600,"token_type":"Bearer"}`))
			return
		}
		if r.URL.Path == "/api/v1/accounts" {
			if count.Add(1) == 1 {
				close(entered)
				select {
				case <-release:
				case <-r.Context().Done():
					return
				}
				w.WriteHeader(401)
				return
			}
			w.Write([]byte(`{"result":[]}`))
			return
		}
		http.NotFound(w, r)
	}))
	defer srv.Close()
	roots := x509.NewCertPool()
	roots.AddCert(srv.Certificate())
	tr := srv.Client().Transport.(*http.Transport).Clone()
	tr.TLSClientConfig.RootCAs = roots
	tr.Proxy = nil
	tr.DialContext = a166FixtureDial(srv)
	c, e := official.NewWithAccountBudget(official.Credentials{APIKey: "fixture", SecretKey: "fixture"}, "", b, func(error) {}, official.WithBaseURL(srv.URL), official.WithAccountSeq(17), official.WithHTTPClient(&http.Client{Transport: tr, Timeout: 4 * time.Second}))
	if e != nil {
		t.Fatal(e)
	}
	if _, e = c.AuthHeaders(context.Background()); e != nil {
		t.Fatal(e)
	}
	ctx, cancel := context.WithTimeout(accountbudget.WithPriority(context.Background(), accountbudget.Critical), 8*time.Second)
	defer cancel()
	done := make(chan error, 1)
	go func() { _, err := c.Accounts(ctx); done <- err }()
	select {
	case <-entered:
	case <-ctx.Done():
		t.Fatal("first account call did not arrive")
	}
	leaseCh := make(chan *accountbudget.Lease, 1)
	errCh := make(chan error, 1)
	go func() {
		lease, err := b.Acquire(ctx, accountbudget.Critical)
		if err != nil {
			errCh <- err
			return
		}
		leaseCh <- lease
	}()
	waitTicket := func() bool {
		deadline := time.Now().Add(3 * time.Second)
		for time.Now().Before(deadline) {
			raw, err := os.ReadFile(filepath.Join(root, "state"))
			if err == nil {
				var state struct {
					Tickets []struct {
						Critical bool `json:"critical"`
					} `json:"tickets"`
				}
				if json.Unmarshal(raw, &state) == nil && len(state.Tickets) > 0 {
					return state.Tickets[0].Critical
				}
			}
			time.Sleep(5 * time.Millisecond)
		}
		t.Error("ticket did not appear")
		return false
	}
	if !waitTicket() {
		close(release)
		t.Fatal("holder did not queue critical")
	}
	close(release)
	var lease *accountbudget.Lease
	select {
	case lease = <-leaseCh:
	case err := <-errCh:
		t.Fatal(err)
	case <-ctx.Done():
		t.Fatal("holder did not acquire")
	}
	if !waitTicket() {
		lease.Release()
		t.Fatal("401 token did not inherit critical")
	}
	if err := lease.Release(); err != nil {
		t.Fatal(err)
	}
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-ctx.Done():
		t.Fatal("request did not finish")
	}
	if count.Load() != 2 {
		t.Fatalf("account attempts=%d", count.Load())
	}
}
