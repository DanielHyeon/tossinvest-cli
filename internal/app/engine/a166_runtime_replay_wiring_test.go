//go:build linux

package engine_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"syscall"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/accountbudget"
	"github.com/JungHoonGhae/tossinvest-cli/internal/accountbudgetruntime"
	"github.com/JungHoonGhae/tossinvest-cli/internal/app/engine"
	"github.com/JungHoonGhae/tossinvest-cli/internal/official"
)

func TestA166RequiredEngineBuildsWithBorrowedReplayBudget(t *testing.T) {
	dir := isolate(t)
	writeEngineConfig(t, dir)
	writeCredentials(t, dir, "fake-key", "fake-secret")
	base := t.TempDir()
	root := filepath.Join(base, "group")
	if err := os.Mkdir(root, 0700); err != nil {
		t.Fatal(err)
	}
	if err := accountbudget.Provision(root, "scope", "instance"); err != nil {
		t.Fatal(err)
	}
	id := func(path string) accountbudget.FileID {
		st, err := os.Stat(path)
		if err != nil {
			t.Fatal(err)
		}
		s := st.Sys().(*syscall.Stat_t)
		return accountbudget.FileID{Device: uint64(s.Dev), Inode: s.Ino}
	}
	data, err := json.Marshal(map[string]any{"version": 1, "root": root, "scope": "scope", "instance": "instance", "identity": accountbudget.Identity{Root: id(root), Queue: id(filepath.Join(root, "queue.lock")), Wire: id(filepath.Join(root, "wire.lock"))}})
	if err != nil {
		t.Fatal(err)
	}
	manifest := filepath.Join(base, "manifest.json")
	if err = os.WriteFile(manifest, data, 0600); err != nil {
		t.Fatal(err)
	}
	factory, err := accountbudgetruntime.Open(context.Background(), accountbudgetruntime.Config{Mode: accountbudgetruntime.ModeRequired, Manifest: manifest, CleanupError: func(error) {}})
	if err != nil {
		t.Fatal(err)
	}
	defer factory.Close()
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/oauth2/token":
			w.Write([]byte(`{"access_token":"fixture","expires_in":3600,"token_type":"Bearer"}`))
		case "/api/v1/accounts":
			w.Write([]byte(`{"result":[{"accountNo":"fixture","accountSeq":7,"accountType":"BROKERAGE"}]}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()
	tr := srv.Client().Transport.(*http.Transport).Clone()
	tr.Proxy = nil
	want := srv.Listener.Addr().String()
	host, _, err := net.SplitHostPort(want)
	if err != nil || net.ParseIP(host) == nil || !net.ParseIP(host).IsLoopback() {
		t.Fatal("fixture listener not loopback")
	}
	tr.DialContext = func(ctx context.Context, network, address string) (net.Conn, error) {
		if address != want {
			return nil, fmt.Errorf("fixture refused nonlocal destination")
		}
		return (&net.Dialer{}).DialContext(ctx, network, want)
	}
	opts := engineOptions(dir, nil)
	opts.BudgetFactory = factory
	opts.OfficialOptions = []official.Option{official.WithBaseURL(srv.URL), official.WithHTTPClient(&http.Client{Transport: tr})}
	eng, err := engine.NewContext(context.Background(), opts)
	if err != nil {
		t.Fatal(err)
	}
	defer eng.Close()
	b, err := factory.Budget()
	if err != nil {
		t.Fatal(err)
	}
	replay := reflect.ValueOf(eng.Gateway).Elem().FieldByName("replay")
	if !replay.IsValid() || replay.IsNil() || replay.Elem().Type().String() != "*execgw.budgetedHTTPReplay" {
		t.Fatal("required engine lacks private budgeted replay")
	}
	client, ok := eng.Official.(*official.Client)
	if !ok || !client.UsesAccountBudget(b) {
		t.Fatal("official budget not shared")
	}
}
