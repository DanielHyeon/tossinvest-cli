package journal

import (
	"compress/gzip"
	"context"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func TestA150NewFirstLegRefusesLegacyBindingAtomically(t *testing.T) {
	j := openTestJournal(t)
	request := legacyFirstLegAtomicFixture(t, j, "a150-legacy-new", "acct-a150", "KR", "005930")
	if _, err := j.RecordQFinalCampaignFirstLeg(context.Background(), request); err == nil {
		t.Fatal("new first leg accepted unversioned legacy binding")
	}
	for _, table := range []string{"decisions", "risk_reservations", "strategy_decision_lineage", "position_campaigns", "risk_bucket_final_decisions", "strategy_first_leg_bindings"} {
		if got := countRiskBucketRows(t, j, table); got != 0 {
			t.Errorf("%s=%d; expected no partial admission", table, got)
		}
	}
}

// The v35 database was produced by the unmodified pre-a150 base using its
// first-leg admission. It proves an old admission remains replayable after v36.
func TestA150HistoricalFirstLegReplayAfterMigration(t *testing.T) {
	input, err := os.Open("testdata/a150-first-leg-v35.db.gz")
	if err != nil {
		t.Fatal(err)
	}
	defer input.Close()
	reader, err := gzip.NewReader(input)
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()
	path := filepath.Join(t.TempDir(), "legacy.db")
	output, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := io.Copy(output, reader); err != nil {
		t.Fatal(err)
	}
	if err := output.Close(); err != nil {
		t.Fatal(err)
	}
	fixtureJournal := openTestJournal(t)
	request := legacyFirstLegAtomicFixture(t, fixtureJournal, "a150-legacy-replay", "acct-a150", "KR", "005930")
	j := openTestJournalAt(t, path)
	receipt, err := j.RecordQFinalCampaignFirstLeg(context.Background(), request)
	if err != nil {
		t.Fatal(err)
	}
	if !receipt.Idempotent {
		t.Fatalf("historical admission not replayed: %+v", receipt)
	}
	protection, err := j.StrategyProtection(context.Background(), request.Issue.Issue.Decision.ID, request.Campaign.CampaignID)
	if err != nil || !protection.Legacy {
		t.Fatalf("legacy read=%+v err=%v", protection, err)
	}
}
