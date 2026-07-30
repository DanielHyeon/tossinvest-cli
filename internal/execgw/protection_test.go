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
	gw, err := execgw.New(execgw.Options{
		Journal:                   j,
		Trading:                   trading.NewService(openPolicy(), broker),
		Clock:                     clk,
		AccountRef:                "acct-7",
		Source:                    "test",
		ProtectionOverrideForTest: execgw.UnwiredProtectionForTest,
	})
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

// TestNoShippedFileClaimsProtection is the guarantee that replaces "the field is
// unexported".
//
// 5.2 made the marker a constant with no config key and no Options knob, on the
// grounds that any of those is a way to claim broker-resident protective
// execution a build does not have. That property survived as long as the seam
// could live in an export_test.go file — but the suites that need to place a buy
// are in other packages (internal/reconcile's gateway tests,
// internal/app/engine's tracer end-to-end), and an export_test.go declaration is
// visible only inside its own package. So the override is exported, and the
// property is restored the way this repository already restores it for the WTS
// order mutators (internal/app/engine's deps_test.go): by proving over the AST
// that no shipped file spells it.
//
// A production assignment to ProtectionOverrideForTest is a build telling the
// gateway it can leave a stop at the broker. If this test goes red, that is what
// happened.
func TestNoShippedFileClaimsProtection(t *testing.T) {
	t.Parallel()

	// What is forbidden is *originating* a wired value, not carrying one.
	//
	// Forwarding matters: internal/app/engine has its own override and hands it
	// to the gateway it builds, so that a test which satisfies the clause on the
	// engine gets a gateway that agrees. That line is nil in every shipped binary
	// and says nothing about this build. What no shipped file may do is produce
	// the WIRED value in the first place — and there are exactly two spellings
	// for that.
	forbidden := []string{"WiredProtectionForTest", "ProtectionWired", "defaultProtection"}

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
		for i, line := range strings.Split(string(src), "\n") {
			code, _, _ := strings.Cut(line, "//")
			for _, name := range forbidden {
				if !isClaim(code, name) {
					continue
				}
				rel, _ := filepath.Rel(root, path)
				t.Errorf("%s:%d assigns %s in shipped code: a build that cannot leave a "+
					"protective order at the broker must not be able to say that it can\n\t%s",
					rel, i+1, name, strings.TrimSpace(line))
			}
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

// isClaim reports that the line originates a WIRED readiness.
//
//	ProtectionWired ProtectionReadiness = "WIRED"          the const — fine
//	var WiredProtectionForTest = &wiredForTest             its declaration — fine
//	var wiredForTest = ProtectionWired                     its declaration — fine
//	if g.protection() == ProtectionWired                   a comparison — fine
//	ProtectionOverrideForTest: opts.protectionOverride     forwarding — fine
//	opts.ProtectionOverrideForTest = WiredProtectionForTest  a claim — refused
//	ready := ProtectionWired                               a claim — refused
func isClaim(code, name string) bool {
	trimmed := strings.TrimSpace(code)
	switch {
	case strings.HasPrefix(trimmed, name+" "):
		return false // the const or field declaration
	case strings.HasPrefix(trimmed, "var "+name+" "), strings.HasPrefix(trimmed, "var "+name+" ="):
		return false // the value's own declaration
	case strings.HasPrefix(trimmed, "var wiredForTest ="), strings.HasPrefix(trimmed, "var WiredProtectionForTest ="):
		return false
	}
	at := strings.Index(trimmed, name)
	if at < 0 {
		return false
	}
	// A comparison reads the value; an assignment produces one. The distinction
	// is the two characters in front of it.
	before := strings.TrimRight(trimmed[:at], " \t")
	switch {
	case strings.HasSuffix(before, "=="), strings.HasSuffix(before, "!="):
		return false
	case strings.HasSuffix(before, "case"), strings.HasSuffix(before, "return"):
		return false
	case strings.HasSuffix(before, "="), strings.HasSuffix(before, ":="), strings.HasSuffix(before, ":"):
		return true
	}
	return false
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
