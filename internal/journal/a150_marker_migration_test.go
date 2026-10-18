package journal

import (
	"context"
	"errors"
	"path/filepath"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
)

func TestA150MarkerMigrationPreservesLegacyAndRejectsDowngrade(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "journal.db")
	old := openJournalAtSchema(t, path, 35)
	before, err := old.RecordDecision(ctx, riskRequest(t))
	if err != nil {
		t.Fatal(err)
	}
	if err = old.Close(); err != nil {
		t.Fatal(err)
	}
	current := openJournalAtSchema(t, path, 36)
	var marker int
	if err = current.db.QueryRowContext(ctx, `SELECT strategy_binding_version FROM decisions WHERE id=?`, before.ID).Scan(&marker); err != nil || marker != 0 {
		t.Fatalf("legacy marker=%d err=%v", marker, err)
	}
	after, err := current.LookupDecision(ctx, before.ID)
	if err != nil || after.RiskHash != before.RiskHash || after.RiskPreimage != before.RiskPreimage {
		t.Fatalf("legacy changed: %+v %v", after, err)
	}
	if _, err = current.db.ExecContext(ctx, `UPDATE decisions SET strategy_binding_version=4 WHERE id=?`, before.ID); err == nil {
		t.Fatal("marker update succeeded")
	}
	if err = current.Close(); err != nil {
		t.Fatal(err)
	}
	_, err = Open(ctx, Options{Path: path, Clock: clock.NewFake(migrationTestInstant), FSProber: FixedFSProber(FSInfo{Name: "ext4", Magic: MagicExt}), migrationOverride: &migrationPlan{steps: migrationsThrough(35), target: 35}})
	if !errors.Is(err, ErrSchemaTooNew) {
		t.Fatalf("old reader: %v", err)
	}
}
