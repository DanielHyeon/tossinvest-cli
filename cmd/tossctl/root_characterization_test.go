package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/config"
	"github.com/JungHoonGhae/tossinvest-cli/internal/orderintent"
	"github.com/JungHoonGhae/tossinvest-cli/internal/output"
	"github.com/JungHoonGhae/tossinvest-cli/internal/session"
)

// ---------------------------------------------------------------------------
// newAppContext characterization tests (harden-execution-base task 1.2)
//
// newAppContext is the single place the CLI decides which backend a run may
// mutate through, which paths it reads secrets from, and whether lineage is
// recorded. Before it is promoted to internal/app so the trading engine can
// share it, these tests pin the behaviour it has today — they must pass both
// before and after the move. Anything they do not pin is free to change; that
// is the point of writing them first.
//
// Every test is hermetic: XDG roots and the credential env vars are redirected
// to a temp dir, so no test can read a developer's real session, credentials or
// config, and no live endpoint is reachable (no HTTP is performed at all —
// newAppContext only constructs clients).
// ---------------------------------------------------------------------------

// isolate points config.DefaultPaths() at temp directories and clears the
// official-credential env vars, then returns the config dir used for
// --config-dir overrides.
func isolate(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(root, "xdg-config"))
	t.Setenv("XDG_CACHE_HOME", filepath.Join(root, "xdg-cache"))
	// LoadCredentials prefers env credentials over the file; clear both so the
	// truth table below is driven only by the file we write.
	t.Setenv("TOSSCTL_OPENAPI_KEY", "")
	t.Setenv("TOSSCTL_OPENAPI_SECRET", "")
	dir := filepath.Join(root, "cfg")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatalf("mkdir config dir: %v", err)
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

// TestNewAppContextConfigDirOverridesAllPaths pins the four path overrides
// --config-dir performs plus the two official-API paths derived from it. Tests
// and CI rely on a single directory containing everything, so a path that
// escapes the override would leak into (or read from) the developer's real
// config directory.
func TestNewAppContextConfigDirOverridesAllPaths(t *testing.T) {
	dir := isolate(t)

	app, err := newAppContext(&rootOptions{outputFormat: "table", configDir: dir})
	if err != nil {
		t.Fatalf("newAppContext: %v", err)
	}

	checks := []struct {
		name, got, want string
	}{
		{"ConfigDir", app.paths.ConfigDir, dir},
		{"ConfigFile", app.paths.ConfigFile, filepath.Join(dir, "config.json")},
		{"SessionFile", app.paths.SessionFile, filepath.Join(dir, "session.json")},
		{"LineageFile", app.paths.LineageFile, filepath.Join(dir, "trading-lineage.json")},
		{"lineageService path", app.lineageService.Path(), filepath.Join(dir, "trading-lineage.json")},
		{"tokenFile", app.tokenFile, filepath.Join(dir, "openapi-token.json")},
	}
	for _, c := range checks {
		if c.got != c.want {
			t.Errorf("%s: got %q, want %q", c.name, c.got, c.want)
		}
	}
}

// TestNewAppContextSessionFileFlagWins pins precedence: --session-file overrides
// the path --config-dir would have produced, and only that path.
func TestNewAppContextSessionFileFlagWins(t *testing.T) {
	dir := isolate(t)
	custom := filepath.Join(dir, "elsewhere", "sess.json")

	app, err := newAppContext(&rootOptions{outputFormat: "table", configDir: dir, sessionFile: custom})
	if err != nil {
		t.Fatalf("newAppContext: %v", err)
	}
	if app.paths.SessionFile != custom {
		t.Errorf("SessionFile: got %q, want %q", app.paths.SessionFile, custom)
	}
	if want := filepath.Join(dir, "config.json"); app.paths.ConfigFile != want {
		t.Errorf("ConfigFile should still follow --config-dir: got %q, want %q", app.paths.ConfigFile, want)
	}
}

// TestNewAppContextTolerantOfMissingSession pins that a run without a stored
// WTS session still builds a context. Every official-API-only command depends
// on this, and so will the engine profile, which never logs into WTS.
func TestNewAppContextTolerantOfMissingSession(t *testing.T) {
	dir := isolate(t)

	app, err := newAppContext(&rootOptions{outputFormat: "table", configDir: dir})
	if err != nil {
		t.Fatalf("missing session must not be an error, got: %v", err)
	}
	if app.session != nil {
		t.Errorf("session: got %+v, want nil", app.session)
	}
	if app.client == nil {
		t.Error("client must be constructed even without a session")
	}
	if app.tradingService == nil {
		t.Error("tradingService must be constructed even without a session")
	}
}

// TestNewAppContextLoadsStoredSession is the other half: when a session file
// exists it is loaded and handed to the context.
func TestNewAppContextLoadsStoredSession(t *testing.T) {
	dir := isolate(t)
	if err := session.WriteFile(filepath.Join(dir, "session.json"), &session.Session{
		Provider: "playwright",
		Cookies:  map[string]string{"x": "y"},
	}); err != nil {
		t.Fatalf("seed session: %v", err)
	}

	app, err := newAppContext(&rootOptions{outputFormat: "table", configDir: dir})
	if err != nil {
		t.Fatalf("newAppContext: %v", err)
	}
	if app.session == nil || app.session.Provider != "playwright" {
		t.Fatalf("session not loaded: %+v", app.session)
	}
}

// TestNewAppContextOfficialClientTruthTable is the safety-relevant one: it pins
// exactly when a run gets an official Open API client. Today the rule is
//
//	official client != nil  ⟺  credentials present AND openapi.enabled AND
//	                           effective backend != "wts"
//
// and when it is nil the hybrid router silently serves everything from the web
// session. The engine profile (task 1.3) must NOT inherit that silent fallback,
// which is why the current rule is written down here first.
func TestNewAppContextOfficialClientTruthTable(t *testing.T) {
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
		{name: "creds+disabled+openapi", creds: true, enabled: false, prefer: "openapi", wantOff: false},
		{name: "creds+disabled+wts", creds: true, enabled: false, prefer: "wts", wantOff: false},
		{name: "nocreds+enabled+auto", creds: false, enabled: true, prefer: "auto", wantOff: false},
		{name: "nocreds+enabled+openapi", creds: false, enabled: true, prefer: "openapi", wantOff: false},
		{name: "nocreds+enabled+wts", creds: false, enabled: true, prefer: "wts", wantOff: false},
		{name: "nocreds+disabled+auto", creds: false, enabled: false, prefer: "auto", wantOff: false},
		{name: "nocreds+disabled+openapi", creds: false, enabled: false, prefer: "openapi", wantOff: false},
		{name: "nocreds+disabled+wts", creds: false, enabled: false, prefer: "wts", wantOff: false},
		// --backend overrides config.prefer in both directions.
		{name: "flag openapi beats prefer wts", creds: true, enabled: true, prefer: "wts", backendFlag: "openapi", wantOff: true},
		{name: "flag wts beats prefer auto", creds: true, enabled: true, prefer: "auto", backendFlag: "wts", wantOff: false},
		// Deprecated alias resolves to openapi.
		{name: "flag official alias", creds: true, enabled: true, prefer: "wts", backendFlag: "official", wantOff: true},
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

			app, err := newAppContext(&rootOptions{outputFormat: "table", configDir: dir, backend: tc.backendFlag})
			if err != nil {
				t.Fatalf("newAppContext: %v", err)
			}
			if gotOff := app.client.Official() != nil; gotOff != tc.wantOff {
				t.Errorf("official client present = %v, want %v", gotOff, tc.wantOff)
			}
		})
	}
}

// TestNewAppContextHybridPolicyMirrorsConfig pins the routing policy handed to
// the hybrid client: the effective backend (config.prefer or --backend) and the
// fallback toggle. Fallback controls whether an official failure retries on the
// web session, so it must not drift during the refactor.
//
// The policy field is unexported; reflection is used deliberately rather than
// widening hybrid's API for a characterization test.
func TestNewAppContextHybridPolicyMirrorsConfig(t *testing.T) {
	for _, tc := range []struct {
		prefer, flag string
		fallback     bool
		wantPrefer   string
	}{
		{prefer: "auto", fallback: true, wantPrefer: "auto"},
		{prefer: "wts", fallback: false, wantPrefer: "wts"},
		{prefer: "openapi", fallback: false, wantPrefer: "openapi"},
		{prefer: "auto", flag: "wts", fallback: true, wantPrefer: "wts"},
	} {
		t.Run(tc.wantPrefer+"/"+tc.flag, func(t *testing.T) {
			dir := isolate(t)
			cfg := config.DefaultFile()
			cfg.OpenAPI = config.OpenAPI{Enabled: true, Prefer: tc.prefer, Fallback: tc.fallback}
			writeConfigFile(t, dir, cfg)
			writeCredentials(t, dir)

			app, err := newAppContext(&rootOptions{outputFormat: "table", configDir: dir, backend: tc.flag})
			if err != nil {
				t.Fatalf("newAppContext: %v", err)
			}
			pol := reflect.ValueOf(app.client).Elem().FieldByName("pol")
			if got := pol.FieldByName("Prefer").String(); got != tc.wantPrefer {
				t.Errorf("policy Prefer: got %q, want %q", got, tc.wantPrefer)
			}
			if got := pol.FieldByName("Fallback").Bool(); got != tc.fallback {
				t.Errorf("policy Fallback: got %v, want %v", got, tc.fallback)
			}
		})
	}
}

// TestNewAppContextWiresLineageIntoTradingService pins that the trading service
// carries the lineage recorder. Without it, an amend performed through any
// surface loses the old-id → new-id trail and a later cancel by the original id
// fails (issue #111). The recorder field is unexported, so reflection stands in
// for an accessor no production caller needs.
func TestNewAppContextWiresLineageIntoTradingService(t *testing.T) {
	dir := isolate(t)

	app, err := newAppContext(&rootOptions{outputFormat: "table", configDir: dir})
	if err != nil {
		t.Fatalf("newAppContext: %v", err)
	}
	if app.lineageService == nil {
		t.Fatal("lineageService must be constructed")
	}
	rec := reflect.ValueOf(app.tradingService).Elem().FieldByName("lineage")
	if !rec.IsValid() {
		t.Fatal("trading.Service has no lineage field — characterization needs updating")
	}
	if rec.IsNil() {
		t.Error("trading service must have the lineage recorder attached")
	}
}

// TestNewAppContextTradingPolicyComesFromConfig pins that the config's trading
// toggles reach the trading service. Observed through PreviewPlace, which is
// the service's own report of what it would allow — no reflection needed, and it
// is the behaviour users actually see.
func TestNewAppContextTradingPolicyComesFromConfig(t *testing.T) {
	intent := orderintent.PlaceIntent{
		Symbol: "005930", Market: "kr", Side: "buy", OrderType: "limit",
		Quantity: 1, Price: 70000, CurrencyMode: "KRW",
	}

	t.Run("defaults refuse mutation", func(t *testing.T) {
		dir := isolate(t)
		app, err := newAppContext(&rootOptions{outputFormat: "table", configDir: dir})
		if err != nil {
			t.Fatalf("newAppContext: %v", err)
		}
		preview := app.tradingService.PreviewPlace(intent)
		if preview.MutationReady {
			t.Error("default config must not report MutationReady")
		}
		if !strings.Contains(strings.Join(preview.Warnings, "\n"), "disables") {
			t.Errorf("expected a config-disabled warning, got %v", preview.Warnings)
		}
	})

	t.Run("enabled config permits mutation", func(t *testing.T) {
		dir := isolate(t)
		cfg := config.DefaultFile()
		cfg.Trading.Place = true
		cfg.Trading.AllowLiveOrderActions = true
		writeConfigFile(t, dir, cfg)

		app, err := newAppContext(&rootOptions{outputFormat: "table", configDir: dir})
		if err != nil {
			t.Fatalf("newAppContext: %v", err)
		}
		if !app.tradingService.PreviewPlace(intent).MutationReady {
			t.Error("place+allow_live config must report MutationReady")
		}
		if !app.config.Trading.Place {
			t.Error("app.config must carry the loaded trading policy")
		}
	})
}

// TestNewAppContextFormatAndBackendValidation pins the two input validations
// newAppContext performs before touching the filesystem.
func TestNewAppContextFormatAndBackendValidation(t *testing.T) {
	dir := isolate(t)

	if app, err := newAppContext(&rootOptions{outputFormat: "json", configDir: dir}); err != nil {
		t.Fatalf("json format: %v", err)
	} else if app.format != output.FormatJSON {
		t.Errorf("format: got %v, want %v", app.format, output.FormatJSON)
	}

	if _, err := newAppContext(&rootOptions{outputFormat: "yaml", configDir: dir}); err == nil {
		t.Error("invalid --output must be rejected")
	}

	_, err := newAppContext(&rootOptions{outputFormat: "table", configDir: dir, backend: "nonsense"})
	if err == nil || !strings.Contains(err.Error(), "invalid --backend") {
		t.Errorf("invalid --backend must be rejected with a pointed error, got %v", err)
	}
}

// TestNewAppContextRejectsMalformedConfig pins fail-closed behaviour: a config
// file that cannot be parsed aborts context construction instead of silently
// falling back to defaults (which would flip trading toggles to their zero
// values without telling the user).
func TestNewAppContextRejectsMalformedConfig(t *testing.T) {
	dir := isolate(t)
	if err := os.WriteFile(filepath.Join(dir, "config.json"), []byte("{not json"), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	if _, err := newAppContext(&rootOptions{outputFormat: "table", configDir: dir}); err == nil {
		t.Error("malformed config.json must abort context construction")
	}
}
