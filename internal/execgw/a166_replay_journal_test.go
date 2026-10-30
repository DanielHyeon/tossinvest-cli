package execgw_test

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"net/http"
	"net/http/httptest"
	"os"
	"sync/atomic"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/accountbudget"
	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/official"
	"github.com/JungHoonGhae/tossinvest-cli/internal/trading"
)

func TestA166ReplayAdmissionRefusalPreservesInDoubtJournal(t *testing.T) {
	f := newReplayFixture(t, execgw.ReplayConfig{})
	// 실제 engine entry 출처로 기록해 admission 실패 경계에 도달하게 함.
	entryGateway, err := execgw.New(execgw.Options{Journal: f.j, Trading: trading.NewService(openPolicy(), &fakeBroker{err: official.ErrServer}), Clock: f.clk, AccountRef: "acct-7", Source: "engine", Entry: f.gate})
	if err != nil {
		t.Fatal(err)
	}
	f.gw = entryGateway
	rec := f.inDoubtPlace(t)
	f.clk.Advance(execgw.DefaultReplayMinInterval)
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
	var posts atomic.Int32
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/oauth2/token" {
			w.Write([]byte(`{"access_token":"fixture","expires_in":3600,"token_type":"Bearer"}`))
			return
		}
		posts.Add(1)
		w.WriteHeader(500)
	}))
	defer srv.Close()
	roots := x509.NewCertPool()
	roots.AddCert(srv.Certificate())
	hc := &http.Client{Transport: &http.Transport{TLSClientConfig: &tls.Config{RootCAs: roots}, DialContext: a166FixtureDial(srv)}, Timeout: 3 * time.Second}
	auth, e := official.NewWithAccountBudget(official.Credentials{APIKey: "key", SecretKey: "secret"}, "", b, func(error) {}, official.WithBaseURL(srv.URL), official.WithHTTPClient(hc), official.WithAccountSeq(17))
	if e != nil {
		t.Fatal(e)
	}
	if _, e = auth.AuthHeaders(context.Background()); e != nil {
		t.Fatal(e)
	}
	replay, e := execgw.NewBudgetedHTTPReplay(srv.URL, hc, auth, b, func(error) {})
	if e != nil {
		t.Fatal(e)
	}
	gw, e := execgw.New(execgw.Options{Journal: f.j, Trading: trading.NewService(openPolicy(), &fakeBroker{err: official.ErrServer}), Clock: f.clk, AccountRef: "acct-7", Source: "test", Entry: f.gate, Replay: replay, Attested: func(context.Context) bool { return true }})
	if e != nil {
		t.Fatal(e)
	}
	held, e := b.Acquire(context.Background(), accountbudget.Critical)
	if e != nil {
		t.Fatal(e)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Millisecond)
	defer cancel()
	_, _ = gw.ReplayInDoubt(ctx, rec.ID)
	if e = held.Release(); e != nil {
		t.Fatal(e)
	}
	after, e := f.j.LookupAttempt(context.Background(), rec.ID)
	if e != nil {
		t.Fatal(e)
	}
	if posts.Load() != 0 || after.State != journal.StateInDoubt || after.ReplayCount != 1 {
		t.Fatalf("posts=%d state=%s count=%d", posts.Load(), after.State, after.ReplayCount)
	}
}
