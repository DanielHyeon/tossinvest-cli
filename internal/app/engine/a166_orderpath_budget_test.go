//go:build linux

package engine

import (
	"context"
	"encoding/json"
	"errors"
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

func TestA166OrderPathUsesBorrowedFactoryBeforeBootstrap(t *testing.T) {
	base := t.TempDir()
	root := filepath.Join(base, "group")
	if e := os.Mkdir(root, 0700); e != nil {
		t.Fatal(e)
	}
	if e := accountbudget.Provision(root, "scope", "instance"); e != nil {
		t.Fatal(e)
	}
	eventID := func() accountbudget.FileID {
		b, err := accountbudget.Open(accountbudget.Config{Root: root, Scope: "scope", Instance: "instance", Clock: clock.System()})
		if err != nil {
			t.Fatal(err)
		}
		id, err := accountbudget.ProvisionObservation(b)
		err = errors.Join(err, b.Close())
		if err != nil {
			t.Fatal(err)
		}
		return id
	}()
	id := func(p string) accountbudget.FileID {
		st, e := os.Stat(p)
		if e != nil {
			t.Fatal(e)
		}
		v := st.Sys().(*syscall.Stat_t)
		return accountbudget.FileID{Device: uint64(v.Dev), Inode: v.Ino}
	}
	data, e := json.Marshal(map[string]any{"event_lock": eventID, "version": 2, "root": root, "scope": "scope", "instance": "instance", "identity": accountbudget.Identity{Root: id(root), Queue: id(filepath.Join(root, "queue.lock")), Wire: id(filepath.Join(root, "wire.lock"))}})
	if e != nil {
		t.Fatal(e)
	}
	manifest := filepath.Join(base, "launch.json")
	if e = os.WriteFile(manifest, data, 0600); e != nil {
		t.Fatal(e)
	}
	f, e := accountbudgetruntime.Open(context.Background(), accountbudgetruntime.Config{Mode: accountbudgetruntime.ModeRequired, Manifest: manifest, CleanupError: func(error) {}})
	if e != nil {
		t.Fatal(e)
	}
	defer f.Close()
	b, e := f.Budget()
	if e != nil {
		t.Fatal(e)
	}
	getenv := func(key string) string {
		switch key {
		case "TOSSCTL_OPENAPI_KEY":
			return "fake"
		case "TOSSCTL_OPENAPI_SECRET":
			return "fake"
		}
		return ""
	}
	path, e := NewOrderPath(Options{ConfigDir: filepath.Join(base, "profile"), Getenv: getenv, BudgetFactory: f})
	if e != nil {
		t.Fatal(e)
	}
	if !path.official.UsesAccountBudget(b) {
		t.Fatal("different budget")
	}
	lease, e := b.Acquire(context.Background(), accountbudget.Critical)
	if e != nil {
		t.Fatal(e)
	}
	defer lease.Release()
	ctx, cancel := context.WithTimeout(context.Background(), 80*time.Millisecond)
	defer cancel()
	if _, e = path.Official.Accounts(ctx); e == nil || ctx.Err() == nil {
		t.Fatalf("bootstrap escaped budget: %v", e)
	}
}

func TestA166OrderPathFixtureOverrideHasNoAuthority(t *testing.T) {
	s := httptest.NewTLSServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { t.Error("unexpected network") }))
	defer s.Close()
	getenv := func(key string) string {
		switch key {
		case "TOSSCTL_OPENAPI_KEY", "TOSSCTL_OPENAPI_SECRET":
			return "fake"
		}
		return ""
	}
	path, e := NewOrderPath(Options{ConfigDir: t.TempDir(), Getenv: getenv, OfficialOptions: []official.Option{official.WithBaseURL(s.URL), official.WithHTTPClient(s.Client())}})
	if e != nil {
		t.Fatal(e)
	}
	if _, ok := path.official.AuthorityOrigin(); ok {
		t.Fatal("fixture gained monetary authority")
	}
}
