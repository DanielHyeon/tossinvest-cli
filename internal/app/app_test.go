package app

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/config"
	"github.com/JungHoonGhae/tossinvest-cli/internal/output"
	"github.com/JungHoonGhae/tossinvest-cli/internal/session"
)

// These tests are the promoted half of the newAppContext characterization suite
// (cmd/tossctl/root_characterization_test.go, harden-execution-base task 1.2).
// The CLI keeps its own copy pointed at the delegating wrapper; this one pins
// the shared API the trading engine will build on.
//
// Hermetic by construction: XDG roots and credential env vars are redirected to
// a temp dir, and New performs no HTTP.

func isolate(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(root, "xdg-config"))
	t.Setenv("XDG_CACHE_HOME", filepath.Join(root, "xdg-cache"))
	t.Setenv("TOSSCTL_OPENAPI_KEY", "")
	t.Setenv("TOSSCTL_OPENAPI_SECRET", "")
	dir := filepath.Join(root, "cfg")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	return dir
}

func writeConfigFile(t *testing.T, dir string, cfg config.File) {
	t.Helper()
	data, err := json.Marshal(cfg)
	if err != nil {
		t.Fatalf("marshal config: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "config.json"), data, 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}
}

func writeCredentials(t *testing.T, dir string) {
	t.Helper()
	data := []byte(`{"apiKey":"test-api-key-000000","secretKey":"test-secret"}`)
	if err := os.WriteFile(filepath.Join(dir, "openapi-credentials.json"), data, 0o600); err != nil {
		t.Fatalf("write credentials: %v", err)
	}
}

func TestNewResolvesConfigDirPaths(t *testing.T) {
	dir := isolate(t)

	ctx, err := New(Options{OutputFormat: "table", ConfigDir: dir})
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	for _, c := range []struct{ name, got, want string }{
		{"ConfigDir", ctx.Paths.ConfigDir, dir},
		{"ConfigFile", ctx.Paths.ConfigFile, filepath.Join(dir, "config.json")},
		{"SessionFile", ctx.Paths.SessionFile, filepath.Join(dir, "session.json")},
		{"LineageFile", ctx.Paths.LineageFile, filepath.Join(dir, "trading-lineage.json")},
		{"LineageService path", ctx.LineageService.Path(), filepath.Join(dir, "trading-lineage.json")},
		{"TokenFile", ctx.TokenFile, filepath.Join(dir, "openapi-token.json")},
	} {
		if c.got != c.want {
			t.Errorf("%s: got %q, want %q", c.name, c.got, c.want)
		}
	}
	if ctx.Format != output.FormatTable {
		t.Errorf("Format: got %v, want table", ctx.Format)
	}
}

func TestNewSessionFileOptionWins(t *testing.T) {
	dir := isolate(t)
	custom := filepath.Join(dir, "custom-session.json")

	ctx, err := New(Options{OutputFormat: "table", ConfigDir: dir, SessionFile: custom})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if ctx.Paths.SessionFile != custom {
		t.Errorf("SessionFile: got %q, want %q", ctx.Paths.SessionFile, custom)
	}
}

func TestNewToleratesMissingSession(t *testing.T) {
	dir := isolate(t)

	ctx, err := New(Options{OutputFormat: "table", ConfigDir: dir})
	if err != nil {
		t.Fatalf("missing session must not be an error: %v", err)
	}
	if ctx.Session != nil {
		t.Errorf("Session: got %+v, want nil", ctx.Session)
	}
	if ctx.Client == nil || ctx.TradingService == nil {
		t.Error("Client and TradingService must be constructed without a session")
	}
}

func TestNewLoadsStoredSession(t *testing.T) {
	dir := isolate(t)
	if err := session.WriteFile(filepath.Join(dir, "session.json"), &session.Session{Provider: "playwright"}); err != nil {
		t.Fatalf("seed session: %v", err)
	}

	ctx, err := New(Options{OutputFormat: "table", ConfigDir: dir})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if ctx.Session == nil || ctx.Session.Provider != "playwright" {
		t.Fatalf("session not loaded: %+v", ctx.Session)
	}
}

// TestNewOfficialClientTruthTable carries the CLI rule forward verbatim:
// an official client exists only with credentials + openapi.enabled + a
// non-wts effective backend. The engine profile deliberately does NOT follow
// this rule (task 1.3) — it must never silently degrade to the web session.
func TestNewOfficialClientTruthTable(t *testing.T) {
	cases := []struct {
		name        string
		creds       bool
		enabled     bool
		prefer      string
		backendFlag string
		wantOff     bool
	}{
		{name: "creds+enabled+auto", creds: true, enabled: true, prefer: "auto", wantOff: true},
		{name: "creds+enabled+openapi", creds: true, enabled: true, prefer: "openapi", wantOff: true},
		{name: "creds+enabled+wts", creds: true, enabled: true, prefer: "wts", wantOff: false},
		{name: "creds+disabled+auto", creds: true, enabled: false, prefer: "auto", wantOff: false},
		{name: "nocreds+enabled+auto", creds: false, enabled: true, prefer: "auto", wantOff: false},
		{name: "nocreds+enabled+openapi", creds: false, enabled: true, prefer: "openapi", wantOff: false},
		{name: "backend flag beats prefer wts", creds: true, enabled: true, prefer: "wts", backendFlag: "openapi", wantOff: true},
		{name: "backend flag wts beats prefer auto", creds: true, enabled: true, prefer: "auto", backendFlag: "wts", wantOff: false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			dir := isolate(t)
			cfg := config.DefaultFile()
			cfg.OpenAPI = config.OpenAPI{Enabled: tc.enabled, Prefer: tc.prefer, Fallback: true}
			writeConfigFile(t, dir, cfg)
			if tc.creds {
				writeCredentials(t, dir)
			}

			ctx, err := New(Options{OutputFormat: "table", ConfigDir: dir, Backend: tc.backendFlag})
			if err != nil {
				t.Fatalf("New: %v", err)
			}
			if got := ctx.Client.Official() != nil; got != tc.wantOff {
				t.Errorf("official client present = %v, want %v", got, tc.wantOff)
			}
		})
	}
}

func TestNewRejectsBadInput(t *testing.T) {
	dir := isolate(t)

	if _, err := New(Options{OutputFormat: "yaml", ConfigDir: dir}); err == nil {
		t.Error("unsupported output format must be rejected")
	}
	_, err := New(Options{OutputFormat: "table", ConfigDir: dir, Backend: "nonsense"})
	if err == nil || !strings.Contains(err.Error(), "invalid --backend") {
		t.Errorf("invalid backend must be rejected: %v", err)
	}

	if err := os.WriteFile(filepath.Join(dir, "config.json"), []byte("{nope"), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}
	if _, err := New(Options{OutputFormat: "table", ConfigDir: dir}); err == nil {
		t.Error("malformed config must abort construction")
	}
}

func TestResolveBackendPrecedence(t *testing.T) {
	t.Parallel()
	cfg := config.OpenAPI{Enabled: true, Prefer: "auto", Fallback: true}

	if got, err := ResolveBackend(cfg, ""); err != nil || got != "auto" {
		t.Errorf("empty flag must use config: got %q, %v", got, err)
	}
	if got, err := ResolveBackend(cfg, "wts"); err != nil || got != "wts" {
		t.Errorf("flag must win: got %q, %v", got, err)
	}
	if got, err := ResolveBackend(cfg, "official"); err != nil || got != "openapi" {
		t.Errorf("deprecated alias must normalize: got %q, %v", got, err)
	}
	if _, err := ResolveBackend(cfg, "bogus"); err == nil || !strings.Contains(err.Error(), "invalid --backend") {
		t.Errorf("bogus flag must error: %v", err)
	}
}

func TestResolveOpenAPIPathsHonoursConfigDir(t *testing.T) {
	dir := isolate(t)

	credFile, tokenFile, err := ResolveOpenAPIPaths(dir)
	if err != nil {
		t.Fatalf("ResolveOpenAPIPaths: %v", err)
	}
	if want := filepath.Join(dir, "openapi-credentials.json"); credFile != want {
		t.Errorf("credFile: got %q, want %q", credFile, want)
	}
	if want := filepath.Join(dir, "openapi-token.json"); tokenFile != want {
		t.Errorf("tokenFile: got %q, want %q", tokenFile, want)
	}

	// No override: both come from DefaultPaths (config dir / cache dir).
	credFile, tokenFile, err = ResolveOpenAPIPaths("")
	if err != nil {
		t.Fatalf("ResolveOpenAPIPaths(\"\"): %v", err)
	}
	paths, err := config.DefaultPaths()
	if err != nil {
		t.Fatalf("DefaultPaths: %v", err)
	}
	if credFile != paths.CredentialsFile || tokenFile != paths.TokenFile {
		t.Errorf("default paths: got (%q,%q), want (%q,%q)", credFile, tokenFile, paths.CredentialsFile, paths.TokenFile)
	}
}
