package testenv_test

// testenv_test.go proves the guards actually guard (task 4.6).
//
// A protection mechanism nobody tested is a protection mechanism that works
// until the day it matters, so each of these drives the guard the way a real
// mistake would: an official client pointed at production, a bare http.Post, a
// forgotten --config-dir.

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/official"
	"github.com/JungHoonGhae/tossinvest-cli/internal/orderintent"
	"github.com/JungHoonGhae/tossinvest-cli/internal/testenv"
)

// --- host classification ----------------------------------------------------

func TestIsRealHost(t *testing.T) {
	real := []string{
		"tossinvest.com",
		"www.tossinvest.com",
		"openapi.tossinvest.com",
		"wts-api.tossinvest.com",
		"wts-info-api.tossinvest.com",
		"wts-cert-api.tossinvest.com",
		"WTS-API.TOSSINVEST.COM",
		"openapi.tossinvest.com:443",
		"openapi.tossinvest.com.",
		"toss.im",
		"api.toss.im",
	}
	for _, host := range real {
		if !testenv.IsRealHost(host) {
			t.Errorf("IsRealHost(%q) = false; that host reaches Toss", host)
		}
	}

	fake := []string{
		"127.0.0.1:8080",
		"localhost",
		"example.com",
		// A lookalike that is not Toss must not be blocked, or every test that
		// uses a plausible fake hostname starts failing for the wrong reason.
		"tossinvest.com.evil.test",
		"nottossinvest.com",
		"",
	}
	for _, host := range fake {
		if testenv.IsRealHost(host) {
			t.Errorf("IsRealHost(%q) = true; that host is not Toss", host)
		}
	}
}

// --- the transport guard ----------------------------------------------------

// TestGuardHardFailsAMutationToARealHost is the spec's scenario: "테스트 중 실 Toss
// hostname으로 mutation 요청이 구성되면 transport 가드가 즉시 실패시킨다".
func TestGuardHardFailsAMutationToARealHost(t *testing.T) {
	var blocked *testenv.ErrRealHost
	guard := &testenv.Guard{OnBlock: func(err *testenv.ErrRealHost) { blocked = err }}
	client := &http.Client{Transport: guard}

	resp, err := client.Post("https://openapi.tossinvest.com/api/v1/orders",
		"application/json", strings.NewReader(`{"symbol":"005930","quantity":"1"}`))
	if err == nil {
		_ = resp.Body.Close()
		t.Fatal("a POST to the live order endpoint was allowed")
	}

	var target *testenv.ErrRealHost
	if !errors.As(err, &target) {
		t.Fatalf("err = %v, want ErrRealHost", err)
	}
	if !target.Mutation {
		t.Error("a POST must be reported as a mutation")
	}
	if blocked == nil {
		t.Error("OnBlock was not called, so a package-wide guard could not fail the test")
	}
	if !strings.Contains(target.Error(), "BLOCKED") {
		t.Errorf("the message must be loud about what nearly happened: %q", target.Error())
	}
}

// TestGuardBlocksEveryMutatingMethod.
func TestGuardBlocksEveryMutatingMethod(t *testing.T) {
	guard := &testenv.Guard{}
	client := &http.Client{Transport: guard}

	for _, method := range []string{http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete} {
		req, err := http.NewRequest(method, "https://wts-api.tossinvest.com/api/v1/trading/orders", nil)
		if err != nil {
			t.Fatalf("building the %s request: %v", method, err)
		}
		resp, err := client.Do(req)
		if err == nil {
			_ = resp.Body.Close()
			t.Errorf("%s to a real host was allowed", method)
			continue
		}
		var target *testenv.ErrRealHost
		if !errors.As(err, &target) || !target.Mutation {
			t.Errorf("%s: err = %v, want a mutation block", method, err)
		}
	}
}

// TestGuardBlocksReadsToo: a test whose result depends on somebody's real
// portfolio is broken in its own way, even though it damages nothing.
func TestGuardBlocksReadsToo(t *testing.T) {
	client := &http.Client{Transport: &testenv.Guard{}}

	resp, err := client.Get("https://openapi.tossinvest.com/api/v1/holdings")
	if err == nil {
		_ = resp.Body.Close()
		t.Fatal("a GET to the live API was allowed")
	}
	var target *testenv.ErrRealHost
	if !errors.As(err, &target) {
		t.Fatalf("err = %v, want ErrRealHost", err)
	}
	if target.Mutation {
		t.Error("a GET must not be reported as a mutation")
	}
}

// TestGuardLetsHttptestThrough: the guard must not be so broad that it makes the
// sanctioned way of testing impossible.
func TestGuardLetsHttptestThrough(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"result":{"ok":true}}`))
	}))
	defer srv.Close()

	client := &http.Client{Transport: &testenv.Guard{Base: srv.Client().Transport}}
	resp, err := client.Post(srv.URL+"/api/v1/orders", "application/json", strings.NewReader("{}"))
	if err != nil {
		t.Fatalf("an httptest POST must be allowed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d", resp.StatusCode)
	}
}

// TestGuardCatchesAnOfficialClientPointedAtProduction is the realistic mistake:
// a test that builds an official.Client and forgets WithBaseURL. The client's
// default base URL is the live API, and without this guard the request would go
// out.
func TestGuardCatchesAnOfficialClientPointedAtProduction(t *testing.T) {
	var blocked *testenv.ErrRealHost
	client := official.New(
		official.Credentials{APIKey: "k", SecretKey: "s"},
		filepath.Join(t.TempDir(), "token.json"),
		official.WithHTTPClient(&http.Client{Transport: &testenv.Guard{
			OnBlock: func(err *testenv.ErrRealHost) { blocked = err },
		}}),
	)

	_, err := client.PlaceOrder(context.Background(), orderintent.PlaceIntent{
		Symbol: "005930", Market: "kr", Side: "buy", OrderType: "limit",
		Quantity: 1, Price: 70000, CurrencyMode: "KRW",
	})
	if err == nil {
		t.Fatal("an order against the default (live) base URL was allowed")
	}
	if blocked == nil {
		t.Fatalf("the guard never saw the request; err = %v", err)
	}
	if !strings.Contains(blocked.URL, "tossinvest.com") {
		t.Errorf("blocked URL = %q", blocked.URL)
	}
}

// TestInstallGuardCoversTheDefaultTransport: a code path that builds its own
// client, or calls http.Post directly, never sees an injected transport.
func TestInstallGuardCoversTheDefaultTransport(t *testing.T) {
	original := http.DefaultTransport

	t.Run("guarded", func(t *testing.T) {
		testenv.InstallGuard(t)
		guard, ok := http.DefaultTransport.(*testenv.Guard)
		if !ok {
			t.Fatalf("http.DefaultTransport = %T, want *testenv.Guard", http.DefaultTransport)
		}
		if guard.Base != original {
			t.Error("the guard did not keep the previous transport as its base")
		}
		// The blocking behaviour itself is covered above; driving it here would
		// fire the guard's OnBlock against the very test asserting the block.
	})

	if http.DefaultTransport != original {
		t.Error("InstallGuard did not restore http.DefaultTransport after the test")
	}
}

// --- isolation --------------------------------------------------------------

// TestIsolateRedirectsEveryPath. A test that isolates three of four variables
// still writes the fourth into the developer's home directory.
func TestIsolateRedirectsEveryPath(t *testing.T) {
	configDir := testenv.Isolate(t)

	if _, err := os.Stat(configDir); err != nil {
		t.Fatalf("the isolated config dir does not exist: %v", err)
	}
	home, _ := os.UserHomeDir()

	for _, name := range []string{"XDG_CONFIG_HOME", "XDG_CACHE_HOME", "XDG_DATA_HOME", "TOSSOS_DATA_DIR"} {
		value := os.Getenv(name)
		if value == "" {
			t.Errorf("%s is empty after Isolate", name)
			continue
		}
		if home != "" && !strings.HasPrefix(value, os.TempDir()) && strings.HasPrefix(value, home) {
			t.Errorf("%s = %q, which is inside the real home directory", name, value)
		}
	}
	if os.Getenv("TOSSCTL_OPENAPI_KEY") != "" || os.Getenv("TOSSCTL_OPENAPI_SECRET") != "" {
		t.Error("Isolate must clear the credential environment, or a developer's real key reaches the test")
	}
}

// TestIsolatedJournalLandsInTheTempDirectory is the end-to-end version: the
// journal resolves its own path from the environment, and after Isolate that
// path is inside the test's temp directory.
func TestIsolatedJournalLandsInTheTempDirectory(t *testing.T) {
	testenv.Isolate(t)

	path, err := journal.DefaultPath()
	if err != nil {
		t.Fatalf("DefaultPath: %v", err)
	}
	if !strings.HasPrefix(path, testenv.DataDir(t)) {
		t.Errorf("journal path = %q, want it inside the isolated data dir %q", path, testenv.DataDir(t))
	}
	home, _ := os.UserHomeDir()
	if home != "" && strings.HasPrefix(path, filepath.Join(home, ".local")) {
		t.Errorf("journal path = %q, which is the developer's real data directory", path)
	}
}

func mustRequest(t *testing.T, method, url string) *http.Request {
	t.Helper()
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		t.Fatalf("building the request: %v", err)
	}
	return req
}
