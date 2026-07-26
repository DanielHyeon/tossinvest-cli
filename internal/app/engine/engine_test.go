package engine_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/app/engine"
	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/config"
	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
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
// config directory, an official client pointed at the stub, and a filesystem
// probe that answers "ext4".
//
// The probe is stubbed for the same reason internal/journal's own tests stub it
// (task 7.2): since the engine profile opens the journal, the allowlist is a
// startup condition, and TMPDIR is not necessarily on an allowlisted filesystem.
// The seam is an unexported Options field with a test-only setter, so it does not
// exist in the built binary.
func engineOptions(dir string, srv *httptest.Server) engine.Options {
	opts := engine.Options{ConfigDir: dir}
	if srv != nil {
		opts.OfficialOptions = []official.Option{
			official.WithBaseURL(srv.URL),
			official.WithHTTPClient(srv.Client()),
		}
	}
	opts.SetJournalProberForTest(journal.FixedFSProber(journal.FSInfo{
		Name: "ext4", Magic: journal.MagicExt,
	}))
	return opts
}

// startEngine builds an engine and closes it when the test ends. The journal is
// a file handle now, so a test that leaks one leaks it for the whole package run.
func startEngine(t *testing.T, dir string, srv *httptest.Server) (*engine.Context, error) {
	t.Helper()
	eng, err := engine.New(engineOptions(dir, srv))
	if eng != nil {
		t.Cleanup(func() { _ = eng.Close() })
	}
	return eng, err
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

	ctx, err := startEngine(t, dir, srv)
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

	ctx, err := startEngine(t, dir, srv)
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

		ctx, err := startEngine(t, dir, srv)
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

		ctx, err := startEngine(t, dir, srv)
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

	ctx, err := startEngine(t, dir, srv)
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

	ctx, err := startEngine(t, dir, srv)
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

			ctx, err := startEngine(t, dir, srv)
			if !errors.Is(err, engine.ErrAccountUnresolved) {
				t.Fatalf("want ErrAccountUnresolved, got %v", err)
			}
			if ctx != nil {
				t.Error("a refused startup must return no engine at all")
			}
		})
	}
}

// --- task 7.2: the journal is part of the engine profile --------------------

// TestStartupOpensTheJournalInTheConfigDir pins where the journal lands.
//
// The path convention is flatten's, kept deliberately: an explicit --config-dir
// means an isolated run, and the journal follows it so a test or a second
// profile cannot touch the real one.
func TestStartupOpensTheJournalInTheConfigDir(t *testing.T) {
	dir := isolate(t)
	writeEngineConfig(t, dir)
	writeCredentials(t, dir, "test-api-key-000000", "test-secret")
	srv, _ := engineStub(t, "123-45")

	eng, err := startEngine(t, dir, srv)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if eng.Journal == nil {
		t.Fatal("the engine profile must open a journal")
	}
	wantPath := filepath.Join(dir, journal.DBFileName)
	if eng.Journal.Path() != wantPath {
		t.Errorf("journal path = %q, want %q", eng.Journal.Path(), wantPath)
	}
	if _, err := os.Stat(wantPath); err != nil {
		t.Errorf("the journal file was not created: %v", err)
	}
}

// TestStartupRefusesOnADisallowedFilesystem states the inherited condition.
//
// The journal's filesystem allowlist (ext4/xfs/btrfs) and its integrity check
// were startup conditions for `tossctl flatten-all` only. Now that the engine
// profile opens the journal, they are startup conditions for the engine — the
// intended inheritance of the P1 journal contract (design D8 step 2), and the
// reason it is stated in a test rather than left to be discovered on a tmpfs.
func TestStartupRefusesOnADisallowedFilesystem(t *testing.T) {
	dir := isolate(t)
	writeEngineConfig(t, dir)
	writeCredentials(t, dir, "test-api-key-000000", "test-secret")
	srv, _ := engineStub(t, "123-45")

	opts := engineOptions(dir, srv)
	opts.SetJournalProberForTest(journal.FixedFSProber(journal.FSInfo{
		Name: "tmpfs", Magic: journal.MagicTmpfs,
	}))

	eng, err := engine.New(opts)
	if !errors.Is(err, journal.ErrFilesystemNotAllowed) {
		t.Fatalf("want ErrFilesystemNotAllowed, got %v", err)
	}
	if eng != nil {
		t.Error("a refused startup must return no engine at all")
	}
}

// TestStartupPrunesSpentNoncesOnce is the retention sweep (task 7.2).
//
// Nobody called PruneSpentNonces before this: the invariant (retention ≥ the
// longest decision TTL) was enforced by the function and fixed by its own tests,
// but the sweep itself had no caller. Startup is that caller — a restart is the
// one moment the engine is not in the middle of anything — and the sweep runs
// once, not on a timer.
func TestStartupPrunesSpentNoncesOnce(t *testing.T) {
	dir := isolate(t)
	writeEngineConfig(t, dir)
	writeCredentials(t, dir, "test-api-key-000000", "test-secret")
	srv, _ := engineStub(t, "123-45")

	now := time.Now()
	stale := seedSpentNonce(t, dir, "stale", now.Add(-400*24*time.Hour))
	recent := seedSpentNonce(t, dir, "recent", now.Add(-time.Hour))

	eng, err := startEngine(t, dir, srv)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	ctx := context.Background()
	if spent, err := eng.Journal.NonceSpent(ctx, stale); err != nil || spent {
		t.Errorf("a consumption record older than the retention survived startup (spent=%v, err=%v)",
			spent, err)
	}
	if spent, err := eng.Journal.NonceSpent(ctx, recent); err != nil || !spent {
		t.Errorf("a record inside the retention was pruned (spent=%v, err=%v) — "+
			"a decision must never outlive the record of its own consumption", spent, err)
	}
}

// seedSpentNonce writes one consumed decision into the journal the engine will
// later open, stamped at the given instant.
//
// The consumption timestamp is the journal clock's, so the seed opens the file
// with a fake clock rather than trying to backdate a row: the same write path
// the engine uses is what produces the record.
func seedSpentNonce(t *testing.T, dir, id string, at time.Time) string {
	t.Helper()
	ctx := context.Background()

	j, err := journal.Open(ctx, journal.Options{
		Path:     filepath.Join(dir, journal.DBFileName),
		Clock:    clock.NewFake(at),
		FSProber: journal.FixedFSProber(journal.FSInfo{Name: "ext4", Magic: journal.MagicExt}),
	})
	if err != nil {
		t.Fatalf("seed %s: open journal: %v", id, err)
	}
	defer func() { _ = j.Close() }()

	dec, err := j.RecordDecision(ctx, journal.DecisionRequest{
		ID:          "dec-" + id,
		AccountRef:  "123-45",
		SafetyClass: journal.SafetyClassRiskReducing,
		Kind:        journal.KindCancel,
		Preimage: journal.ReductionIntent{
			AccountRef: "123-45", Market: "kr", Symbol: "005930", Side: "SELL",
			MaxQuantity: "1", Reason: "seed",
		},
		Nonce: "nonce-" + id, IssuedAt: at, ExpiresAt: at.Add(time.Minute),
	})
	if err != nil {
		t.Fatalf("seed %s: RecordDecision: %v", id, err)
	}

	attempt, err := j.Prepare(ctx, journal.PrepareRequest{
		Intent: journal.Intent{
			ID: "intent-" + id, Market: "kr", TradingDay: "2026-07-24", AccountRef: "123-45",
			Symbol: "005930", Side: "SELL", OrderType: "LIMIT", Quantity: "1", Price: "70000",
			Currency: "KRW", Source: "seed", Fingerprint: "fp-" + id,
		},
		Kind: journal.KindCancel, AttemptID: "attempt-" + id, TargetOrderID: "O-" + id,
		DecisionID: dec.ID, SafetyClass: dec.SafetyClass,
	})
	if err != nil {
		t.Fatalf("seed %s: Prepare: %v", id, err)
	}
	// The consumption record is written inside this transition, which is the
	// whole point of the invariant being about consumption records.
	if err := attempt.MarkDispatchStarted(ctx); err != nil {
		t.Fatalf("seed %s: MarkDispatchStarted: %v", id, err)
	}
	return dec.Nonce
}

// --- task 7.3: the gateway is constructed by the engine profile -------------

// TestStartupConstructsTheGateway is the requirement itself. `execgw.New` used to
// be called from exactly one place — the flatten CLI — which meant the engine
// profile had a journal-less, gateway-less order path and nothing said so.
func TestStartupConstructsTheGateway(t *testing.T) {
	dir := isolate(t)
	writeEngineConfig(t, dir)
	writeCredentials(t, dir, "test-api-key-000000", "test-secret")
	srv, _ := engineStub(t, "123-45")

	eng, err := startEngine(t, dir, srv)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if eng.Gateway == nil {
		t.Fatal("the engine profile must construct an ExecutionGateway")
	}
	if eng.Entry == nil {
		t.Error("the gateway's entry gate must be reachable: an operator has to be able to see and clear a latch")
	}
	if eng.Resolver == nil {
		t.Error("an IN_DOUBT attempt with no resolver is an attempt nothing will ever settle")
	}
	if eng.Reconcile == nil {
		t.Error("the reconcile tracker is the projection's owner; without it a restart forgets every block")
	}

	wiring := eng.Gateway.Wiring()
	if !wiring.Orders {
		// issues.md 2026-07-26: the round-trip check is inert without this, and a
		// place is then confirmed on the broker's ack alone.
		t.Error("Options.Orders must be wired, or the identifier round trip never runs")
	}
	if !wiring.Entry {
		t.Error("Options.Entry must be wired")
	}
	if !wiring.Preflight {
		t.Error("Options.Preflight must be wired: a nil preflight is not a skipped check, it is no check")
	}
	if !wiring.Replay {
		t.Error("Options.Replay must be wired to the official token manager")
	}
	if wiring.Attested {
		t.Error("the replay capability is not attested in this build; the predicate must stay nil " +
			"so the entry point resends nothing [미측정 — 2b 전 비활성]")
	}
}

// TestStartupRebuildsTheReconcileProjection is the restart half of task 4.1: the
// journal is authoritative and the in-memory latches are a projection of it.
//
// Without the Restore call at startup, a restart silently clears every block a
// disagreement raised — which is the one failure mode persisting the states was
// meant to remove.
func TestStartupRebuildsTheReconcileProjection(t *testing.T) {
	dir := isolate(t)
	writeEngineConfig(t, dir)
	writeCredentials(t, dir, "test-api-key-000000", "test-secret")
	srv, _ := engineStub(t, "123-45")

	seedAccountWideReconcile(t, dir, "123-45", "local derivation and the broker disagree")

	eng, err := startEngine(t, dir, srv)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	rejected := eng.Entry.CheckEntry()
	if rejected == nil {
		t.Fatal("a restart must not clear an active account-wide RECONCILE state")
	}
	// The gate's latch reason comes from the row's cause (ReconcileReasonFor), so
	// a QUANTITY_MISMATCH row restores as a mismatch block; the tracker is the
	// side that knows an account-wide row is its permanent promotion.
	if rejected.Reason != execgw.ReasonReconcileMismatch {
		t.Errorf("entry refused for %q, want the reconcile block restored from the journal", rejected.Reason)
	}
	if !strings.Contains(rejected.Detail, "disagree") {
		t.Errorf("the restored block lost its evidence: %q", rejected.Detail)
	}
	blocks := eng.Reconcile.Blocks()
	if len(blocks) != 1 {
		t.Fatalf("tracker restored %d block(s), want 1", len(blocks))
	}
	if !blocks[0].Permanent {
		t.Error("an account-wide row is the permanent promotion; a clean pass must not release it")
	}
}

// seedAccountWideReconcile writes an active account-wide RECONCILE row into the
// journal the engine will later open.
func seedAccountWideReconcile(t *testing.T, dir, accountRef, evidence string) {
	t.Helper()
	ctx := context.Background()
	j, err := journal.Open(ctx, journal.Options{
		Path:     filepath.Join(dir, journal.DBFileName),
		FSProber: journal.FixedFSProber(journal.FSInfo{Name: "ext4", Magic: journal.MagicExt}),
	})
	if err != nil {
		t.Fatalf("seed reconcile: open journal: %v", err)
	}
	defer func() { _ = j.Close() }()

	if _, _, err := j.EnterReconcile(ctx, journal.EnterReconcileRequest{
		AccountRef: accountRef,
		Cause:      journal.ReconcileCauseQuantityMismatch,
		Evidence:   evidence,
	}); err != nil {
		t.Fatalf("seed reconcile: %v", err)
	}
}
