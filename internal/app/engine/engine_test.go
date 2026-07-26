package engine_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/app/engine"
	"github.com/JungHoonGhae/tossinvest-cli/internal/config"
	"github.com/JungHoonGhae/tossinvest-cli/internal/official"
	"github.com/JungHoonGhae/tossinvest-cli/internal/orderintent"
	"github.com/JungHoonGhae/tossinvest-cli/internal/testenv"
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
	// One definition of "isolated", in internal/testenv (task 4.6). The data
	// directory matters as much as the config one: the audit log the startup
	// interlock writes (task 4.2) resolves from $TOSSOS_DATA_DIR, and a test that
	// misses it writes into the developer's real ~/.local/share/tossos.
	return testenv.Isolate(t)
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

// engineStub answers the reads engine startup performs and counts the account
// list among them.
//
// Since task 7.1 the account is resolved on every start, gate or no gate, so a
// test that constructs an engine needs a broker that can answer
// GET /api/v1/accounts. Before 7.1 the gate-off path made no call at all and
// these tests passed with no server; the count is exposed so the new behaviour
// is asserted rather than assumed.
func engineStub(t *testing.T, accountNo string) (*httptest.Server, func() int) {
	t.Helper()
	var mu sync.Mutex
	calls := 0

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/oauth2/token":
			_, _ = w.Write([]byte(`{"access_token":"AT","expires_in":3600,"token_type":"Bearer"}`))
		case "/api/v1/accounts":
			mu.Lock()
			calls++
			mu.Unlock()
			_, _ = w.Write([]byte(`{"result":[{"accountNo":"` + accountNo + `","accountSeq":7,"accountType":"BROKERAGE"}]}`))
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)

	return srv, func() int {
		mu.Lock()
		defer mu.Unlock()
		return calls
	}
}

// engineOptions is the standard startup input for a test engine: an isolated
// config directory and an official client pointed at the stub.
func engineOptions(dir string, srv *httptest.Server) engine.Options {
	opts := engine.Options{ConfigDir: dir}
	if srv != nil {
		opts.OfficialOptions = []official.Option{
			official.WithBaseURL(srv.URL),
			official.WithHTTPClient(srv.Client()),
		}
	}
	return opts
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
	srv, _ := engineStub(t, "123-45")

	ctx, err := engine.New(engineOptions(dir, srv))
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
	srv, _ := engineStub(t, "123-45")

	ctx, err := engine.New(engineOptions(dir, srv))
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
		srv, _ := engineStub(t, "123-45")

		ctx, err := engine.New(engineOptions(dir, srv))
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
		srv, _ := engineStub(t, "123-45")

		ctx, err := engine.New(engineOptions(dir, srv))
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
	srv, _ := engineStub(t, "123-45")

	ctx, err := engine.New(engineOptions(dir, srv))
	if err != nil {
		t.Fatalf("New with env credentials: %v", err)
	}
	if ctx.Official == nil {
		t.Error("env credentials must produce an official client")
	}
}

// --- task 7.1: the account is resolved on every start -----------------------
//
// Before 7.1 the account read lived inside the gate's attestation check, so an
// engine with the gate off — every engine today — never learned which account
// its credentials belong to. Everything downstream needs that string: the
// journal records intents against it, the gateway refuses to be built without
// it, and the reconcile projection is scoped by it. Resolving it only when the
// gate happens to be on made the identity of the account a function of a
// feature toggle (design D8 step 1).

// TestStartupResolvesTheAccountWithTheGateOff is the requirement itself.
func TestStartupResolvesTheAccountWithTheGateOff(t *testing.T) {
	dir := isolate(t)
	writeEngineConfig(t, dir)
	writeCredentials(t, dir, "test-api-key-000000", "test-secret")
	srv, accountCalls := engineStub(t, "123-45")

	ctx, err := engine.New(engineOptions(dir, srv))
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if ctx.AccountRef != "123-45" {
		t.Errorf("AccountRef = %q, want the account the credentials resolve to", ctx.AccountRef)
	}
	if ctx.Automation.Enabled || ctx.Automation.Verified {
		t.Error("the gate must still be off; only the account read moved")
	}
	if ctx.Automation.AccountRef != "123-45" {
		t.Errorf("Automation.AccountRef = %q; the status must report the resolved account too",
			ctx.Automation.AccountRef)
	}
	if got := accountCalls(); got != 1 {
		t.Errorf("account reads = %d, want exactly 1 — one read at startup, not one per consumer", got)
	}
}

// TestStartupRefusesWhenTheAccountCannotBeResolved: an engine that cannot name
// its own account cannot journal an intent against it, so there is nothing safe
// to return. Fail closed rather than hand back a half-built engine.
func TestStartupRefusesWhenTheAccountCannotBeResolved(t *testing.T) {
	for _, tc := range []struct {
		name string
		body string
		code int
	}{
		{"broker error", `{"error":{"code":"internal","message":"boom"}}`, http.StatusInternalServerError},
		{"no accounts at all", `{"result":[]}`, http.StatusOK},
		{"account with no number", `{"result":[{"accountNo":"  ","accountSeq":7}]}`, http.StatusOK},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dir := isolate(t)
			writeEngineConfig(t, dir)
			writeCredentials(t, dir, "test-api-key-000000", "test-secret")

			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				if r.URL.Path == "/oauth2/token" {
					_, _ = w.Write([]byte(`{"access_token":"AT","expires_in":3600,"token_type":"Bearer"}`))
					return
				}
				w.WriteHeader(tc.code)
				_, _ = w.Write([]byte(tc.body))
			}))
			t.Cleanup(srv.Close)

			ctx, err := engine.New(engineOptions(dir, srv))
			if !errors.Is(err, engine.ErrAccountUnresolved) {
				t.Fatalf("want ErrAccountUnresolved, got %v", err)
			}
			if ctx != nil {
				t.Error("a refused startup must return no engine at all")
			}
		})
	}
}
