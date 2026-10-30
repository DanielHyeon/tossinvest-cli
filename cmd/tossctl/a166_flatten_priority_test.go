//go:build linux

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/accountbudget"
	"github.com/JungHoonGhae/tossinvest-cli/internal/official"
	"github.com/spf13/cobra"
)

func TestA166FlattenBootstrapPriority(t *testing.T) {
	t.Setenv("TOSSCTL_OPENAPI_KEY", "fake-key")
	t.Setenv("TOSSCTL_OPENAPI_SECRET", "fake-secret")
	manifest := a166RootManifest(t)
	raw, err := os.ReadFile(manifest)
	if err != nil {
		t.Fatal(err)
	}
	var m struct {
		Root string `json:"root"`
	}
	if err = json.Unmarshal(raw, &m); err != nil {
		t.Fatal(err)
	}
	owner := a166Owner(manifest)
	defer owner.Close()
	factory, err := owner.get(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	budget, err := factory.Budget()
	if err != nil {
		t.Fatal(err)
	}
	lease, err := budget.Acquire(context.Background(), accountbudget.Critical)
	if err != nil {
		t.Fatal(err)
	}
	defer lease.Release()
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { t.Error("network reached while lease held") }))
	defer server.Close()
	template := server.Client().Transport.(*http.Transport).Clone()
	template.Proxy = nil
	want := server.Listener.Addr().String()
	host, _, err := net.SplitHostPort(want)
	if err != nil || net.ParseIP(host) == nil || !net.ParseIP(host).IsLoopback() {
		t.Fatal("fixture listener must use loopback")
	}
	template.DialContext = func(ctx context.Context, network, address string) (net.Conn, error) {
		if address != want {
			return nil, fmt.Errorf("fixture refused nonlocal destination")
		}
		return (&net.Dialer{}).DialContext(ctx, network, want)
	}
	root := &rootOptions{outputFormat: "table", budgetOwner: owner, configDir: t.TempDir(), officialOptions: []official.Option{official.WithBaseURL(server.URL), official.WithHTTPClient(&http.Client{Transport: template})}}
	ctx, cancel := context.WithCancel(context.Background())
	cmd := &cobra.Command{}
	cmd.SetContext(ctx)
	done := make(chan error, 1)
	go func() { done <- runFlatten(cmd, root, &flattenOptions{}) }()
	defer func() { cancel(); <-done }()
	deadline := time.Now().Add(4 * time.Second)
	for time.Now().Before(deadline) {
		data, e := os.ReadFile(filepath.Join(m.Root, "state"))
		if e == nil {
			var state struct {
				Tickets []struct {
					Critical bool `json:"critical"`
				} `json:"tickets"`
			}
			if json.Unmarshal(data, &state) == nil && len(state.Tickets) > 0 {
				if !state.Tickets[0].Critical {
					t.Fatal("bootstrap queued as normal")
				}
				return
			}
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("bootstrap did not queue")
}
