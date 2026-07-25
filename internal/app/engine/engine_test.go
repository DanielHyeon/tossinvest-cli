package engine_test

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/app/engine"
	"github.com/JungHoonGhae/tossinvest-cli/internal/config"
	"github.com/JungHoonGhae/tossinvest-cli/internal/orderintent"
)

// Engine-profile wiring tests (harden-execution-base task 1.3).
//
// The CLI profile (internal/app) is permissive by design: no credentials means
// the hybrid router quietly serves everything from the web session. The engine
// profile must be the opposite — an automated trader that silently switched to a
// scraped web session would be placing live orders through a path TossOS
// declares read-only (WORKFLOW 불변 규칙). These tests pin that difference.

// isolate redirects the XDG roots and clears credential env vars, returning a
// config dir for the engine's --config-dir equivalent.
func isolate(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(root, "xdg-config"))
	t.Setenv("XDG_CACHE_HOME", filepath.Join(root, "xdg-cache"))
	// The data directory has to be isolated too, or the audit log the startup
	// interlock writes (task 4.2) lands in the developer's real
	// ~/.local/share/tossos. A test that writes outside its temp directory is a
	// test that can corrupt live state.
	t.Setenv("XDG_DATA_HOME", filepath.Join(root, "xdg-data"))
	t.Setenv("TOSSOS_DATA_DIR", filepath.Join(root, "tossos-data"))
	t.Setenv("TOSSCTL_OPENAPI_KEY", "")
	t.Setenv("TOSSCTL_OPENAPI_SECRET", "")
	dir := filepath.Join(root, "cfg")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	return dir
}

// writeEngineConfig writes a config whose OpenAPI section is hostile to the
// official path (disabled, pinned to wts) and whose trading toggles are all
// open. The engine must ignore the former and honour the latter.
func writeEngineConfig(t *testing.T, dir string) {
	t.Helper()
	cfg := config.DefaultFile()
	cfg.OpenAPI = config.OpenAPI{Enabled: false, Prefer: "wts", Fallback: true}
	cfg.Trading = config.Trading{
		Place: true, Sell: true, Fractional: true, Cancel: true, Amend: true,
		Conditional: true, AllowLiveOrderActions: true,
	}
	data, err := json.Marshal(cfg)
	if err != nil {
		t.Fatalf("marshal config: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "config.json"), data, 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}
}

func writeCredentials(t *testing.T, dir, key, secret string) {
	t.Helper()
	body, err := json.Marshal(map[string]string{"apiKey": key, "secretKey": secret})
	if err != nil {
		t.Fatalf("marshal credentials: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "openapi-credentials.json"), body, 0o600); err != nil {
		t.Fatalf("write credentials: %v", err)
	}
}

// TestNewRefusesWithoutCredentials is the headline requirement: no credentials
// means no engine, with an error that says so. The CLI would have degraded to
// the web session here.
func TestNewRefusesWithoutCredentials(t *testing.T) {
	dir := isolate(t)
	writeEngineConfig(t, dir)

	ctx, err := engine.New(engine.Options{ConfigDir: dir})
	if !errors.Is(err, engine.ErrOfficialCredentialsRequired) {
		t.Fatalf("want ErrOfficialCredentialsRequired, got %v", err)
	}
	if ctx != nil {
		t.Error("no context may be returned when startup is refused")
	}
}

// TestNewRefusesIncompleteCredentials covers a credentials file that exists but
// is half-filled. "Present but unusable" must fail closed too, not produce a
// client that 401s on its first live order.
func TestNewRefusesIncompleteCredentials(t *testing.T) {
	for _, tc := range []struct{ name, key, secret string }{
		{"no secret", "key-only", ""},
		{"no key", "", "secret-only"},
		{"both blank", "", ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dir := isolate(t)
			writeEngineConfig(t, dir)
			writeCredentials(t, dir, tc.key, tc.secret)

			if _, err := engine.New(engine.Options{ConfigDir: dir}); !errors.Is(err, engine.ErrOfficialCredentialsRequired) {
				t.Fatalf("want ErrOfficialCredentialsRequired, got %v", err)
			}
		})
	}
}

// TestNewIgnoresOpenAPIConfigToggles pins that openapi.enabled=false and
// openapi.prefer=wts — which would leave the CLI with no official client at all
// — do not affect the engine. Those toggles are a user routing preference for
// interactive use; the engine's broker is not negotiable.
func TestNewIgnoresOpenAPIConfigToggles(t *testing.T) {
	dir := isolate(t)
	writeEngineConfig(t, dir)
	writeCredentials(t, dir, "test-api-key-000000", "test-secret")

	ctx, err := engine.New(engine.Options{ConfigDir: dir})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if ctx.Official == nil {
		t.Fatal("engine must construct an official client regardless of openapi.enabled/prefer")
	}
	if ctx.BrokerForTest() == nil || ctx.TradingService == nil {
		t.Fatal("engine must wire a broker and trading service")
	}
	if ctx.Config.OpenAPI.Enabled || ctx.Config.OpenAPI.Prefer != "wts" {
		t.Errorf("config must be reported as loaded, unmodified: %+v", ctx.Config.OpenAPI)
	}
}

// TestNewResolvesPathsFromConfigDir keeps the engine reading the same files the
// CLI does — one credential file, one token cache. A divergence here is how you
// end up trading with a stale key.
func TestNewResolvesPathsFromConfigDir(t *testing.T) {
	dir := isolate(t)
	writeEngineConfig(t, dir)
	writeCredentials(t, dir, "test-api-key-000000", "test-secret")

	ctx, err := engine.New(engine.Options{ConfigDir: dir})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if want := filepath.Join(dir, "config.json"); ctx.Paths.ConfigFile != want {
		t.Errorf("ConfigFile: got %q, want %q", ctx.Paths.ConfigFile, want)
	}
	if want := filepath.Join(dir, "openapi-token.json"); ctx.TokenFile != want {
		t.Errorf("TokenFile: got %q, want %q", ctx.TokenFile, want)
	}
}

// TestNewCarriesTradingPolicy pins that the config's trading toggles reach the
// engine's trading service — the engine must not be able to place an order the
// user's config forbids.
func TestNewCarriesTradingPolicy(t *testing.T) {
	intent := orderintent.PlaceIntent{
		Symbol: "005930", Market: "kr", Side: "buy", OrderType: "limit",
		Quantity: 1, Price: 70000, CurrencyMode: "KRW",
	}

	t.Run("open config", func(t *testing.T) {
		dir := isolate(t)
		writeEngineConfig(t, dir)
		writeCredentials(t, dir, "test-api-key-000000", "test-secret")

		ctx, err := engine.New(engine.Options{ConfigDir: dir})
		if err != nil {
			t.Fatalf("New: %v", err)
		}
		if !ctx.TradingService.PreviewPlace(intent).MutationReady {
			t.Error("engine trading service must report MutationReady for an open config")
		}
	})

	t.Run("default config refuses", func(t *testing.T) {
		dir := isolate(t)
		// No config file: defaults have every trading toggle off.
		writeCredentials(t, dir, "test-api-key-000000", "test-secret")

		ctx, err := engine.New(engine.Options{ConfigDir: dir})
		if err != nil {
			t.Fatalf("New: %v", err)
		}
		if ctx.TradingService.PreviewPlace(intent).MutationReady {
			t.Error("engine must inherit the config's closed trading toggles")
		}
	})
}

// TestNewRejectsMalformedConfig — fail closed rather than run on zero-value
// policy the user never wrote.
func TestNewRejectsMalformedConfig(t *testing.T) {
	dir := isolate(t)
	writeCredentials(t, dir, "test-api-key-000000", "test-secret")
	if err := os.WriteFile(filepath.Join(dir, "config.json"), []byte("{nope"), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	if _, err := engine.New(engine.Options{ConfigDir: dir}); err == nil {
		t.Error("malformed config.json must refuse engine startup")
	}
}

// TestNewAcceptsEnvCredentials mirrors the CLI: env credentials outrank the
// file, so an operator can run the engine without a credentials file on disk.
func TestNewAcceptsEnvCredentials(t *testing.T) {
	dir := isolate(t)
	writeEngineConfig(t, dir)
	t.Setenv("TOSSCTL_OPENAPI_KEY", "env-key-0000000000")
	t.Setenv("TOSSCTL_OPENAPI_SECRET", "env-secret")

	ctx, err := engine.New(engine.Options{ConfigDir: dir})
	if err != nil {
		t.Fatalf("New with env credentials: %v", err)
	}
	if ctx.Official == nil {
		t.Error("env credentials must produce an official client")
	}
}
