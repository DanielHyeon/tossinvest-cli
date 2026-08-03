package strategyruntime_test

import (
	"go/parser"
	"go/token"
	"os"
	"strings"
	"testing"

	runtime "github.com/JungHoonGhae/tossinvest-cli/internal/strategyruntime"
)

func TestExternalAPICannotMintLeaseAuthorityOwnerCapabilityOrFallback(t *testing.T) {
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") || strings.HasSuffix(entry.Name(), "_test.go") {
			continue
		}
		file, err := parser.ParseFile(token.NewFileSet(), entry.Name(), nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		for _, forbidden := range []string{"NewLease", "NewAuthoritySnapshot", "NewOwnerFence", "NewRecoveryCapability", "NewSafetyFallbackManifest", "NewWorkerState"} {
			if file.Scope.Lookup(forbidden) != nil {
				t.Fatalf("production API exports authority mint %s in %s", forbidden, entry.Name())
			}
		}
	}
}

func TestExternalDefaultIsPairedOFFAndCannotForgeLease(t *testing.T) {
	state := runtime.NewCoordinatorState()
	for _, market := range []runtime.Market{runtime.MarketKR, runtime.MarketUS} {
		if state.Worker(market).Effective != runtime.EntryOff {
			t.Fatalf("market=%s default=%+v", market, state.Worker(market))
		}
	}
	forged := runtime.Lease{}
	if forged.State() != runtime.LeaseInvalid || forged.Disposition() != runtime.ReservationInvalid {
		t.Fatalf("zero forged lease appeared valid: state=%s disposition=%s", forged.State(), forged.Disposition())
	}
}
