package protectionreadiness_test

import (
	"go/parser"
	"go/token"
	"os"
	"strings"
	"testing"

	readiness "github.com/JungHoonGhae/tossinvest-cli/internal/protectionreadiness"
)

func TestExternalAPICannotMintTrustEvidenceSupervisorOrState(t *testing.T) {
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
		for _, forbidden := range []string{"NewPinnedTrustPolicy", "NewObservedFile", "NewSupervisorBinding", "NewTrustedTime", "NewDurableState"} {
			if file.Scope.Lookup(forbidden) != nil {
				t.Fatalf("production API exports authority-minting constructor %s in %s", forbidden, entry.Name())
			}
		}
	}
}

func TestOnlyPublicSnapshotConstructorIsPairedUnwiredDefault(t *testing.T) {
	snapshot := readiness.DefaultSnapshot()
	if snapshot.Release() != readiness.ReadinessRelease {
		t.Fatalf("release=%q", snapshot.Release())
	}
	for _, market := range []readiness.Market{readiness.MarketKR, readiness.MarketUS} {
		verdict := snapshot.Verdict(market)
		if verdict.State != readiness.Unwired || verdict.Code != readiness.RefusalMissingEvidence {
			t.Fatalf("market=%s verdict=%+v", market, verdict)
		}
	}
}
