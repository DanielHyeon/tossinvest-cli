//go:build tossos_testseams

package strategyrouter_test

// a127 D4 · S2 — route 적재기는 `journal.Open` 이 연(마이그레이션된) 실제 원장을 `journal.SchemaVersion` 주입으로 읽는다. 내부 시험은 journal 을
// import 할 수 없어서(journal → strategyrouter) 외부 시험 패키지에 둔다. 동결 리터럴로 되돌리면(S2) 이 시험이 실패한다.

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyrouter"
)

func TestA127RouteLoaderReadsARealJournalOpenedByJournalOpen(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "journal.db")
	j, err := journal.Open(context.Background(), journal.Options{Path: path,
		FSProber: journal.FixedFSProber(journal.FSInfo{Name: "ext4", Magic: journal.MagicExt})})
	if err != nil {
		t.Fatalf("journal.Open: %v", err)
	}
	t.Cleanup(func() { _ = j.Close() })
	if err := os.Chmod(path, 0o600); err != nil {
		t.Fatal(err)
	}
	for _, market := range []strategyrouter.Market{strategyrouter.MarketKR, strategyrouter.MarketUS} {
		config, err := strategyrouter.SignedProductionRouteConfigForTest(t.TempDir(), path, market, time.Date(2026, 8, 4, 1, 0, 0, 0, time.UTC), journal.SchemaVersion)
		if err != nil {
			t.Fatal(err)
		}
		batch, err := strategyrouter.LoadProductionRouteAuthorityBatch(context.Background(), config, []strategyrouter.ProductionRouteTarget{{Symbol: config.Symbol}})
		if err != nil {
			t.Fatalf("%s: real journal (schema %d) refused: %v", market, journal.SchemaVersion, err)
		}
		if _, ok := batch.For(config.Symbol); !ok || batch.Len() != 1 {
			t.Fatalf("%s: batch len=%d, want the one signed scope routed from the real journal", market, batch.Len())
		}
	}
}
