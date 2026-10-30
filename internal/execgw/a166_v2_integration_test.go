//go:build linux

package execgw

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
	"syscall"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/accountbudget"
	"github.com/JungHoonGhae/tossinvest-cli/internal/accountbudgetruntime"
	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/official"
)

func TestA166ReplayRuntimeV2ObservesTokenAndPost(t *testing.T) {
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
		st, err := os.Stat(path)
		if err != nil {
			t.Fatal(err)
		}
		v := st.Sys().(*syscall.Stat_t)
		return accountbudget.FileID{Device: uint64(v.Dev), Inode: v.Ino}
	}
	data, err := json.Marshal(map[string]any{"version": 2, "root": root, "scope": "scope", "instance": "instance", "event_lock": event, "identity": accountbudget.Identity{Root: id(root), Queue: id(filepath.Join(root, "queue.lock")), Wire: id(filepath.Join(root, "wire.lock"))}})
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "launch.json")
	if err = os.WriteFile(path, data, 0600); err != nil {
		t.Fatal(err)
	}
	f, err := accountbudgetruntime.Open(context.Background(), accountbudgetruntime.Config{Mode: accountbudgetruntime.ModeRequired, Manifest: path, CleanupError: func(error) {}})
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	b, err = f.Budget()
	if err != nil {
		t.Fatal(err)
	}
	store, err := accountbudget.OpenObservation(context.Background(), b, event, clock.SystemBootSource())
	if err != nil {
		t.Fatal(err)
	}
	kinds := make(chan accountbudget.ObservationKind, 2)
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		view, err := store.Snapshot(r.Context(), accountbudget.ObservationCursor{})
		if err != nil || view.Pending == nil {
			t.Errorf("missing pending %v", err)
			return
		}
		kinds <- view.Pending.Kind
		switch r.URL.Path {
		case "/oauth2/token":
			fmt.Fprint(w, `{"access_token":"fake","expires_in":3600,"token_type":"Bearer"}`)
		case replayPlacePath:
			fmt.Fprint(w, `{"result":{"orderId":"id1","clientOrderId":"k1"}}`)
		default:
			t.Error("unexpected route")
		}
	}))
	defer srv.Close()
	pool := x509.NewCertPool()
	pool.AddCert(srv.Certificate())
	want := srv.Listener.Addr().String()
	tr := &http.Transport{TLSClientConfig: &tls.Config{RootCAs: pool}, DialContext: func(ctx context.Context, network, address string) (net.Conn, error) {
		if network != "tcp" || address != want {
			return nil, errors.New("unexpected destination")
		}
		return (&net.Dialer{}).DialContext(ctx, network, address)
	}}
	hc := &http.Client{Transport: tr, Timeout: 5 * time.Second}
	client, err := f.NewClient(official.Credentials{APIKey: "fake", SecretKey: "fake"}, filepath.Join(dir, "token"), official.WithBaseURL(srv.URL), official.WithHTTPClient(hc), official.WithAccountSeq(17))
	if err != nil {
		t.Fatal(err)
	}
	replay, err := NewBudgetedHTTPReplay(srv.URL, hc, client, b, func(error) {})
	if err != nil {
		t.Fatal(err)
	}
	resp, err := replay.ReplayPlace(context.Background(), ReplayBody{stored: `{"clientOrderId":"k1"}`})
	if err != nil || resp.Status != 200 {
		t.Fatalf("replay %+v %v", resp, err)
	}
	for _, want := range []accountbudget.ObservationKind{accountbudget.ObservationToken, accountbudget.ObservationOther} {
		if got := <-kinds; got != want {
			t.Fatalf("kind %v want %v", got, want)
		}
	}
}
