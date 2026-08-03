package strategyrouter_test

import (
	"go/parser"
	"go/token"
	"os"
	"strings"
	"testing"
	"time"

	strategyrouter "github.com/JungHoonGhae/tossinvest-cli/internal/strategyrouter"
)

func TestExternalAPIExposesNoAuthorityMintingConstructor(t *testing.T) {
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
		for _, forbidden := range []string{"NewOwnerSnapshot", "NewQuotaSnapshot", "NewMarketRecord"} {
			if file.Scope.Lookup(forbidden) != nil {
				t.Fatalf("production API exports authority-minting constructor %s in %s", forbidden, entry.Name())
			}
		}
	}
}

func TestExternalCallersCannotForgeOwnerMarketOrQuotaAttestation(t *testing.T) {
	now := time.Now().UTC()
	key, err := strategyrouter.NewOwnerKey("acct", strategyrouter.MarketKR, "005930", 1)
	if err != nil {
		t.Fatal(err)
	}
	forgedSnapshot := strategyrouter.OwnerSnapshot{Key: key, Revision: 1, Digest: "forged", ObservedAt: now.Add(-time.Second), FreshUntil: now.Add(time.Minute)}
	result := strategyrouter.Route(strategyrouter.RouteRequest{
		Key: key, ExpectedOwnerRevision: 1, ExpectedMarketRevision: 1, EvaluatedAt: now,
		Snapshot: forgedSnapshot, MarketRecord: strategyrouter.DefaultMarketRecord(strategyrouter.MarketKR),
	})
	if result.Code == strategyrouter.RefusalNone {
		t.Fatalf("forged owner snapshot accepted: %+v", result)
	}

	state := strategyrouter.NewSchedulerState()
	forgedRecord := strategyrouter.MarketRecord{Market: strategyrouter.MarketKR, Desired: strategyrouter.StateOn, Effective: strategyrouter.StateOn, Revision: 2, ActivationDigest: "forged"}
	if _, cas := strategyrouter.CASMarketRecord(state, strategyrouter.MarketKR, 1, "", forgedRecord); cas.Code != strategyrouter.RefusalInvalid {
		t.Fatalf("forged ON market record accepted: %+v", cas)
	}

	authority := strategyrouter.NewQuotaAuthority()
	forgedQuota := strategyrouter.QuotaSnapshot{Key: strategyrouter.PhysicalQuotaKey{Endpoint: "quotes", ResetGeneration: "reset"}, ReportedRemaining: 100, ObservedAt: now, FreshUntil: now.Add(time.Hour), Digest: "forged"}
	if err := authority.Install(forgedQuota); err == nil {
		t.Fatal("forged physical quota snapshot installed")
	}
}
