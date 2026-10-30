package execgw_test

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
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

func TestA166RequiredReplayUnknownOriginFallsBack(t *testing.T) {
	f := newReplayFixture(t, execgw.ReplayConfig{})
	rec := f.inDoubtPlace(t)
	if _, err := f.j.MarkReplayStarted(context.Background(), rec.ID, 5); err != nil {
		t.Fatal(err)
	}
	var err error
	rec, err = f.j.LookupAttempt(context.Background(), rec.ID)
	if err != nil {
		t.Fatal(err)
	}
	if rec.ReplayCount != 1 {
		t.Fatalf("fixture count = %d", rec.ReplayCount)
	}
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
		w.Write([]byte(`{"result":{"orderId":"fixture-order"}}`))
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
	out, e := gw.ReplayInDoubt(context.Background(), rec.ID)
	if e != nil {
		t.Fatal(e)
	}
	if !out.QueryFallback || out.Reason != execgw.ReasonReplayIneligible || out.ReplayCount != rec.ReplayCount {
		t.Fatalf("unexpected fallback: %+v", out)
	}
	after, e := f.j.LookupAttempt(context.Background(), rec.ID)
	if e != nil {
		t.Fatal(e)
	}
	if posts.Load() != 0 || after.State != journal.StateInDoubt || after.ReplayCount != rec.ReplayCount {
		t.Fatalf("posts=%d state=%s count=%d", posts.Load(), after.State, after.ReplayCount)
	}
}

func TestA166FlattenReplayQueuesCritical(t *testing.T) {
	f := newReplayFixture(t, execgw.ReplayConfig{})
	ctx := context.Background()
	now := f.clk.Now()
	dec, err := f.j.RecordDecision(ctx, journal.DecisionRequest{ID: "fixture-reduction", AccountRef: "acct-7", SafetyClass: journal.SafetyClassRiskReducing, Kind: journal.KindPlace, Preimage: journal.ReductionIntent{AccountRef: "acct-7", Market: "kr", Symbol: "005930", Side: "SELL", MaxQuantity: "1", Reason: "fixture"}, Nonce: "fixture", IssuedAt: now, ExpiresAt: now.Add(time.Hour)})
	if err != nil {
		t.Fatal(err)
	}
	intent := journal.Intent{ID: "fixture-intent", AccountRef: "acct-7", Market: "kr", TradingDay: "2026-03-30", Symbol: "005930", Side: "SELL", OrderType: "LIMIT", Quantity: "1", Price: "1", Currency: "KRW", Source: "flatten-all", Fingerprint: "fixture"}
	attempt, err := f.j.Prepare(ctx, journal.PrepareRequest{Intent: intent, Kind: journal.KindPlace, AttemptID: "fixture-attempt", DecisionID: dec.ID, AccountRef: "acct-7", SafetyClass: dec.SafetyClass, Generation: dec.Generation, ClientOrderID: dec.ClientOrderID, WireBody: `{"fixture":true}`, SerializerVersion: "fixture"})
	if err != nil {
		t.Fatal(err)
	}
	if err = attempt.MarkDispatchStarted(ctx); err != nil {
		t.Fatal(err)
	}
	if err = attempt.MarkInDoubt(ctx, "timeout", "fixture"); err != nil {
		t.Fatal(err)
	}
	root := t.TempDir()
	if err = os.Chmod(root, 0700); err != nil {
		t.Fatal(err)
	}
	if err = accountbudget.Provision(root, "scope", "instance"); err != nil {
		t.Fatal(err)
	}
	b, err := accountbudget.OpenContext(ctx, accountbudget.Config{Root: root, Scope: "scope", Instance: "instance", Clock: clock.System()})
	if err != nil {
		t.Fatal(err)
	}
	defer b.Close()
	var posts atomic.Int32
	bodies := make(chan string, 1)
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/oauth2/token" {
			w.Write([]byte(`{"access_token":"fixture","expires_in":3600,"token_type":"Bearer"}`))
			return
		}
		posts.Add(1)
		body, _ := io.ReadAll(r.Body)
		bodies <- string(body)
		w.Write([]byte(`{"result":{"orderId":"fixture-order"}}`))
	}))
	defer srv.Close()
	roots := x509.NewCertPool()
	roots.AddCert(srv.Certificate())
	hc := &http.Client{Transport: &http.Transport{TLSClientConfig: &tls.Config{RootCAs: roots}, DialContext: a166FixtureDial(srv)}, Timeout: 4 * time.Second}
	auth, err := official.NewWithAccountBudget(official.Credentials{APIKey: "key", SecretKey: "secret"}, "", b, func(error) {}, official.WithBaseURL(srv.URL), official.WithHTTPClient(hc), official.WithAccountSeq(17))
	if err != nil {
		t.Fatal(err)
	}
	if _, err = auth.AuthHeaders(ctx); err != nil {
		t.Fatal(err)
	}
	replay, err := execgw.NewBudgetedHTTPReplay(srv.URL, hc, auth, b, func(error) {})
	if err != nil {
		t.Fatal(err)
	}
	gw, err := execgw.New(execgw.Options{Journal: f.j, Trading: trading.NewService(openPolicy(), &fakeBroker{err: official.ErrServer}), Clock: f.clk, AccountRef: "acct-7", Source: "engine", Entry: f.gate, Replay: replay, Attested: func(context.Context) bool { return true }})
	if err != nil {
		t.Fatal(err)
	}
	lease, err := b.Acquire(ctx, accountbudget.Critical)
	if err != nil {
		t.Fatal(err)
	}
	runCtx, cancel := context.WithTimeout(ctx, 6*time.Second)
	defer cancel()
	done := make(chan error, 1)
	go func() { _, e := gw.ReplayInDoubt(runCtx, "fixture-attempt"); done <- e }()
	deadline := time.Now().Add(3 * time.Second)
	found := false
	for time.Now().Before(deadline) {
		raw, e := os.ReadFile(filepath.Join(root, "state"))
		if e == nil {
			var state struct {
				Tickets []struct {
					Critical bool `json:"critical"`
				} `json:"tickets"`
			}
			if json.Unmarshal(raw, &state) == nil && len(state.Tickets) > 0 {
				found = state.Tickets[0].Critical
				break
			}
		}
		time.Sleep(5 * time.Millisecond)
	}
	if err = lease.Release(); err != nil {
		t.Fatal(err)
	}
	if err = <-done; err != nil {
		t.Fatal(err)
	}
	if !found || posts.Load() != 1 {
		t.Fatalf("critical=%v posts=%d", found, posts.Load())
	}
	if body := <-bodies; body != `{"fixture":true}` {
		t.Fatalf("stored bytes changed: %q", body)
	}
}
