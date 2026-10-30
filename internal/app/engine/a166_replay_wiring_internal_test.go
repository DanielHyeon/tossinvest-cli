//go:build linux

package engine

import (
	"context"
	"os"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/accountbudget"
	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/official"
)

func TestA166GatewayReplayBinding(t *testing.T) {
	root := t.TempDir()
	if err := os.Chmod(root, 0700); err != nil {
		t.Fatal(err)
	}
	if err := accountbudget.Provision(root, "scope", "instance"); err != nil {
		t.Fatal(err)
	}
	b, err := accountbudget.OpenContext(context.Background(), accountbudget.Config{Root: root, Scope: "scope", Instance: "instance", Clock: clock.System()})
	if err != nil {
		t.Fatal(err)
	}
	defer b.Close()
	cleanup := func(error) {}
	off := official.New(official.Credentials{APIKey: "fake", SecretKey: "fake"}, "")
	if _, err := newGatewayReplay(off, b, cleanup); err == nil {
		t.Fatal("unbound client accepted")
	}
	client, err := official.NewWithAccountBudget(official.Credentials{APIKey: "fake", SecretKey: "fake"}, "", b, cleanup)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = newGatewayReplay(client, b, cleanup); err != nil {
		t.Fatal(err)
	}
	if _, err = newGatewayReplay(off, nil, cleanup); err != nil {
		t.Fatal(err)
	}
}
