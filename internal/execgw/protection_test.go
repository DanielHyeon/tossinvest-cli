package execgw_test

// protection_test.go is interlock clause 6 at the mutation chokepoint (change
// interlock-gates-entry-not-exit, design D2).
//
// The clause used to refuse the whole runtime at startup, which meant it also
// refused the exit observer — so a build with no broker-resident protection had
// no protection *at all*, which is the failure the clause names, applied wider.
// It now refuses one thing: a mutation that raises exposure.
//
// It is asserted here rather than in the engine because this is where
// `raisesExposure` is computed, and it is computed from the intent's own shape
// (`side == "buy"`, gateway.go:338) rather than from the Safety Class the caller
// declared. A buy wearing a RISK_REDUCING badge is refused one clause earlier by
// the class/shape check; what these tests cover is the case where the paperwork
// is entirely correct and the build still cannot protect what it would open.

import (
	"context"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/domain"
	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/orderintent"
	"github.com/JungHoonGhae/tossinvest-cli/internal/trading"
)

// newUnprotectedGateway builds the gateway the shipped binary builds.
//
// This package's test binary defaults to WIRED (export_test.go), because nearly
// every suite here drives a buy for reasons that have nothing to do with clause
// 6. These two tests are the exception, so they ask for the build's own answer.
func newUnprotectedGateway(t *testing.T, broker trading.Broker) (
	*execgw.Gateway, *journal.Journal, *clock.Fake,
) {
	t.Helper()
	clk := clock.NewFake(fixedNow)
	j := openJournal(t, clk)
	opts := execgw.Options{
		Journal: j, Trading: trading.NewService(openPolicy(), broker), Clock: clk,
		AccountRef: "acct-7", Source: "test",
	}
	opts.UseReadinessAdapterForTest()
	gw, err := execgw.New(opts)
	if err != nil {
		t.Fatalf("execgw.New: %v", err)
	}
	return gw, j, clk
}

// TestARaisingMutationIsRefusedWhileProtectionIsUnwired is the clause, moved.
//
// Everything about this place is in order — a persisted EXPOSURE_RAISING
// decision that authorises exactly this order, a held reservation, limits that
// cover it. The refusal is the one thing no configuration reaches.
func TestARaisingMutationIsRefusedWhileProtectionIsUnwired(t *testing.T) {
	broker := &fakeBroker{result: domain.MutationResult{Kind: "place", Status: "accepted", OrderID: "O-1"}}
	gw, j, clk := newUnprotectedGateway(t, broker)

	out, err := gw.Place(context.Background(), placeRequest(t, j, clk))
	if err == nil && out.State == journal.StateConfirmed {
		t.Fatal("a buy was confirmed on a build with no broker-resident protective execution")
	}
	if places, _, _ := broker.totals(); places != 0 {
		t.Errorf("broker place calls = %d, want 0 — the refusal must precede the broker", places)
	}

	// The operator has to be able to tell this refusal from a limit breach or an
	// in-flight symbol: the fix is a different change, not a different setting.
	detail := out.Detail + " "
	if err != nil {
		detail += err.Error()
	}
	for _, want := range []string{"protect"} {
		if !strings.Contains(strings.ToLower(detail), want) {
			t.Errorf("refusal %q does not name %q as the cause", detail, want)
		}
	}
}

// TestAReducingMutationIsAdmittedWhileProtectionIsUnwired is the other half, and
// it is the whole point of the change: the exits stay open.
//
// checkEntry's own comment already says it — "the whole point of blocking
// entries is to keep the exits open" — and clause 6 was the one entry block that
// did not honour it, because it closed the runtime instead of the entry.
func TestAReducingMutationIsAdmittedWhileProtectionIsUnwired(t *testing.T) {
	broker := &fakeBroker{result: domain.MutationResult{Kind: "place", Status: "accepted", OrderID: "O-2"}}
	gw, j, clk := newUnprotectedGateway(t, broker)

	sell := orderintent.PlaceIntent{
		Symbol: "005930", Market: "kr", Side: "sell", OrderType: "limit",
		Quantity: 2, Price: 70000, CurrencyMode: "KRW",
	}
	out, err := gw.Place(context.Background(), execgw.PlaceRequest{
		Intent:   sell,
		Decision: exitDecision(t, j, clk, journal.KindPlace, sell.Market, sell.Symbol, sell.Side, sell.Quantity),
	})
	if err != nil {
		t.Fatalf("a reduce-only sell must pass a gateway with no protective execution: %v", err)
	}
	if out.State != journal.StateConfirmed {
		t.Fatalf("state = %s (%s), want CONFIRMED — the exit path is what this change exists to open",
			out.State, out.Detail)
	}
	if places, _, _ := broker.totals(); places != 1 {
		t.Errorf("broker place calls = %d, want exactly 1", places)
	}
}

// TestNoShippedFileClaimsProtection locks out the retired public scalar forge.
// A production caller may supply only the sealed adapter; no bool/string/enum
// field on Options is accepted as execution authority.
func TestNoShippedFileClaimsProtection(t *testing.T) {
	t.Parallel()

	optionsType := reflect.TypeOf(execgw.Options{})
	for i := 0; i < optionsType.NumField(); i++ {
		field := optionsType.Field(i)
		if field.IsExported() && (field.Type == reflect.TypeOf(execgw.ProtectionReadiness("")) ||
			field.Type == reflect.TypeOf((*execgw.ProtectionReadiness)(nil))) {
			t.Errorf("exported Options.%s accepts scalar protection readiness", field.Name)
		}
	}

	forbidden := []string{"ProtectionOverrideForTest", "WiredProtectionForTest", "UnwiredProtectionForTest"}

	root := moduleRoot(t)
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		switch {
		case err != nil:
			return err
		case d.IsDir():
			if name := d.Name(); name == ".git" || name == "vendor" || name == "node_modules" {
				return filepath.SkipDir
			}
			return nil
		case !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go"):
			return nil
		}
		src, readErr := os.ReadFile(path)
		if readErr != nil {
			return readErr
		}
		for _, name := range forbidden {
			if !strings.Contains(string(src), name) {
				continue
			}
			rel, _ := filepath.Rel(root, path)
			t.Errorf("%s contains retired production protection forge %s", rel, name)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walking the module: %v", err)
	}
}

// moduleRoot walks up from the package directory to the directory holding go.mod.
func moduleRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	for {
		if _, statErr := os.Stat(filepath.Join(dir, "go.mod")); statErr == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("could not locate go.mod above the test directory")
		}
		dir = parent
	}
}

// TestTheBuildMarkerIsAConstant: the readiness this build reports is not a
// variable anything can reassign at run time.
func TestTheBuildMarkerIsAConstant(t *testing.T) {
	t.Parallel()

	if execgw.ProfileProtection != execgw.ProtectionUnwired {
		t.Fatalf("ProfileProtection = %q: this build wires no protective execution",
			execgw.ProfileProtection)
	}
	if reflect.ValueOf(execgw.ProfileProtection).Kind() != reflect.String {
		t.Error("ProfileProtection is not a plain string constant")
	}
}
