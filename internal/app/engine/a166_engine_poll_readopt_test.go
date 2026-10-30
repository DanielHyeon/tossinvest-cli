//go:build linux

package engine

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"syscall"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/accountbudget"
	"github.com/JungHoonGhae/tossinvest-cli/internal/accountbudgetruntime"
	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/config"
	"github.com/JungHoonGhae/tossinvest-cli/internal/domain"
	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/exitpolicy"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/official"
	"github.com/JungHoonGhae/tossinvest-cli/internal/positionpolicy"
)

func TestA166ReadoptSharesEnginePriceHint(t *testing.T) {
	dir := t.TempDir()
	root := filepath.Join(dir, "group")
	if err := os.Mkdir(root, 0700); err != nil {
		t.Fatal(err)
	}
	if err := accountbudget.Provision(root, "scope", "instance"); err != nil {
		t.Fatal(err)
	}
	b, err := accountbudget.Open(accountbudget.Config{Root: root, Scope: "scope", Instance: "instance", Clock: clock.System()})
	if err != nil {
		t.Fatal(err)
	}
	event, err := accountbudget.ProvisionObservation(b)
	err = errors.Join(err, b.Close())
	if err != nil {
		t.Fatal(err)
	}
	id := func(path string) accountbudget.FileID {
		st, e := os.Stat(path)
		if e != nil {
			t.Fatal(e)
		}
		v := st.Sys().(*syscall.Stat_t)
		return accountbudget.FileID{Device: uint64(v.Dev), Inode: v.Ino}
	}
	data, err := json.Marshal(map[string]any{"version": 2, "root": root, "scope": "scope", "instance": "instance", "event_lock": event, "identity": accountbudget.Identity{Root: id(root), Queue: id(filepath.Join(root, "queue.lock")), Wire: id(filepath.Join(root, "wire.lock"))}})
	if err != nil {
		t.Fatal(err)
	}
	manifest := filepath.Join(dir, "manifest")
	if err = os.WriteFile(manifest, data, 0600); err != nil {
		t.Fatal(err)
	}
	f, err := accountbudgetruntime.Open(context.Background(), accountbudgetruntime.Config{Mode: accountbudgetruntime.ModeRequired, Manifest: manifest, CleanupError: func(error) {}})
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	b, err = f.Budget()
	if err != nil {
		t.Fatal(err)
	}
	var tokenCalls, priceCalls atomic.Int32
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/oauth2/token" {
			tokenCalls.Add(1)
			fmt.Fprint(w, `{"access_token":"fake","expires_in":3600,"token_type":"Bearer"}`)
			return
		}
		priceCalls.Add(1)
		w.Header().Set("Retry-After", "2")
		w.WriteHeader(429)
	}))
	defer server.Close()
	pool := x509.NewCertPool()
	pool.AddCert(server.Certificate())
	want := server.Listener.Addr().String()
	tr := &http.Transport{TLSClientConfig: &tls.Config{RootCAs: pool}, DialContext: func(ctx context.Context, network, address string) (net.Conn, error) {
		if network != "tcp" || address != want {
			return nil, errors.New("unexpected destination")
		}
		return (&net.Dialer{}).DialContext(ctx, network, want)
	}}
	client, err := f.NewClient(official.Credentials{APIKey: "fake", SecretKey: "fake"}, "", official.WithBaseURL(server.URL), official.WithHTTPClient(&http.Client{Transport: tr}), official.WithAccountSeq(1))
	if err != nil {
		t.Fatal(err)
	}
	retrier, err := execgw.NewEnginePollRetrier(&execgw.Retrier{AccountRef: "fixture"}, b, client, "fixture", 1)
	if err != nil {
		t.Fatal(err)
	}
	ectx := &Context{Retrier: retrier}
	ectx.Automation.Verified = true
	if _, err := ectx.ReconcileDriver(ReconcileDriverOptions{Prices: pollSliceReader{1}}); !errors.Is(err, ErrReconcileDriverUnavailable) || !strings.Contains(err.Error(), "reader does not match poll binding") {
		t.Fatalf("foreign uncomparable reader: %v", err)
	}
	err = retrier.Query(accountbudget.WithPriority(context.Background(), accountbudget.Critical), execgw.QueryPrice, func(ctx context.Context) error { _, e := client.Prices(ctx, []string{"005930"}); return e })
	if !errors.Is(err, official.ErrRateLimited) || tokenCalls.Load() != 1 || priceCalls.Load() != 1 {
		t.Fatalf("protective 429: %v token=%d price=%d", err, tokenCalls.Load(), priceCalls.Load())
	}
	driver := &ReconcileDriver{opts: ReconcileDriverOptions{Prices: client, Retrier: retrier}, clk: clock.System()}
	_, _, err = driver.observeCandidates(context.Background(), []candidate{{position: journal.Position{Symbol: "005930", Market: "kr"}}})
	if !errors.Is(err, accountbudget.ErrObservationDeferred) || priceCalls.Load() != 1 {
		t.Fatalf("adoption shared hint err=%v wire=%d", err, priceCalls.Load())
	}
	repository := &fakePolicyRepository{state: positionpolicy.State{PositionID: "p-1", Symbol: "005930", ExitEligible: true, Provenance: positionpolicy.ProvenanceExternalAdoption, Eligibility: positionpolicy.EligibilityExternalLifecycle, Status: positionpolicy.StatusReleased, AdoptionGeneration: 1, Version: 1, DesiredPolicyID: exitpolicy.CommonLadderRunner}}
	service := &PositionPolicyCommandService{j: repository, prices: client, retrier: retrier, adoption: config.Adoption{DefaultStopPct: 0.05}, commonPolicy: exitpolicy.CommonLadderBalanced, clk: clock.System()}
	request := positionpolicy.Request{PositionID: "p-1", ExpectedGeneration: 1, ExpectedVersion: 1, Action: positionpolicy.ActionReadopt, Actor: positionpolicy.ActorLocalOperator, Reason: positionpolicy.ReasonReadopt}
	_, err = service.Preview(context.Background(), request)
	if !errors.Is(err, accountbudget.ErrObservationDeferred) || tokenCalls.Load() != 1 || priceCalls.Load() != 1 || repository.previewed.PositionID != "" {
		t.Fatalf("err=%v token=%d price=%d", err, tokenCalls.Load(), priceCalls.Load())
	}
}

type pollSliceReader []int

func (pollSliceReader) Prices(context.Context, []string) ([]domain.Quote, error) {
	return nil, errors.New("unexpected call")
}
