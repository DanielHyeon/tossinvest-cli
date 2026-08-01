package optimization_test

import (
	"context"
	"database/sql"
	"errors"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/optimization"
	"github.com/JungHoonGhae/tossinvest-cli/internal/settingmeta"
)

type evidenceProvider struct{ evidence optimization.Evidence }

func (p evidenceProvider) ReadEvidence(context.Context) (optimization.Evidence, error) {
	return p.evidence, nil
}

func testRegistry(t *testing.T) optimization.Registry {
	t.Helper()
	registry, err := optimization.BuildRegistry(context.Background(), optimization.ProviderBinding{
		Category: optimization.CategoryExitProtection,
		Provider: optimization.StaticProvider{Owner: "a041", Fields: []settingmeta.FieldDescriptor{
			descriptor("a041", "exit.common-policy"),
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	return registry
}

func openStore(t *testing.T, path string, clk *clock.Fake, evidence optimization.Evidence) *optimization.Store {
	t.Helper()
	store, err := optimization.Open(context.Background(), optimization.Options{
		Path: path, Registry: testRegistry(t), Clock: clk, Actor: "operator:test",
		Evidence: evidenceProvider{evidence: evidence},
	})
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = store.Close() })
	return store
}

func TestInsufficientEvidenceBlocksRecommendationButAllowsFiniteServerPreset(t *testing.T) {
	clk := clock.NewFake(time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC))
	store := openStore(t, filepath.Join(t.TempDir(), "optimization.db"), clk, optimization.Evidence{
		Status: optimization.EvidenceInsufficient, Missing: []string{"complete-lineage", "min-sample"},
	})
	view, err := store.Read(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	request := optimization.PreviewRequest{BaseVersion: view.Snapshot.Version,
		Category: optimization.CategoryExitProtection,
		Changes:  map[string]string{"exit.common-policy": "SAFE"},
		Source:   optimization.SourceEvidence, Reason: optimization.ReasonServerPreset}
	if _, err := store.Preview(context.Background(), request); !errors.Is(err, optimization.ErrInsufficientEvidence) {
		t.Fatalf("evidence-backed preview error = %v", err)
	}
	request.Source = optimization.SourceServerPreset
	preview, err := store.Preview(context.Background(), request)
	if err != nil {
		t.Fatalf("finite manual preset preview: %v", err)
	}
	if preview.Evidence.Status != optimization.EvidenceInsufficient || !preview.LiveStateUnchanged ||
		!preview.ExistingPositionsUnchanged || !preview.RiskConfirmationRequired {
		t.Fatalf("preview safety/evidence = %+v", preview)
	}
}

func TestApplyIsVersionBoundOneShotCASAndAppendOnly(t *testing.T) {
	clk := clock.NewFake(time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC))
	path := filepath.Join(t.TempDir(), "optimization.db")
	store := openStore(t, path, clk, optimization.Evidence{Status: optimization.EvidenceComplete, Digest: "evidence-v1"})
	view, _ := store.Read(context.Background())
	preview, err := store.Preview(context.Background(), optimization.PreviewRequest{
		BaseVersion: view.Snapshot.Version, Category: optimization.CategoryExitProtection,
		Changes: map[string]string{"exit.common-policy": "SAFE"}, Source: optimization.SourceServerPreset,
		Reason: optimization.ReasonServerPreset,
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.Apply(context.Background(), optimization.ApplyRequest{Capability: preview.Capability, Confirmed: true}); !errors.Is(err, optimization.ErrCapabilityTooEarly) {
		t.Fatalf("early apply error = %v", err)
	}
	clk.Advance(3 * time.Second)
	if _, err := store.Apply(context.Background(), optimization.ApplyRequest{Capability: preview.Capability}); !errors.Is(err, optimization.ErrConfirmationRequired) {
		t.Fatalf("unconfirmed apply error = %v", err)
	}
	result, err := store.Apply(context.Background(), optimization.ApplyRequest{Capability: preview.Capability, Confirmed: true})
	if err != nil {
		t.Fatal(err)
	}
	if result.Snapshot.Version != view.Snapshot.Version+1 || result.Snapshot.Desired["exit.common-policy"] != "SAFE" {
		t.Fatalf("applied snapshot = %+v", result.Snapshot)
	}
	if result.Snapshot.EffectiveEntry || result.Snapshot.ActivationManifestDigest != "" {
		t.Fatalf("risk apply restored effective authority: %+v", result.Snapshot)
	}
	replayed, err := store.Apply(context.Background(), optimization.ApplyRequest{Capability: preview.Capability, Confirmed: true})
	if err != nil || !replayed.Replayed || replayed.Snapshot.Version != result.Snapshot.Version {
		t.Fatalf("idempotent replay = %+v, %v", replayed, err)
	}
	latest, err := store.Read(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(latest.History) != 2 || len(latest.Audit) != 1 || latest.Audit[0].BeforeOptionID != "" ||
		latest.Audit[0].AfterOptionID != "SAFE" {
		t.Fatalf("history/audit = %+v / %+v", latest.History, latest.Audit)
	}
	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	rows, err := db.Query(`PRAGMA table_info(optimization_candidates)`)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		var cid, notNull, primaryKey int
		var name, kind string
		var defaultValue any
		if err := rows.Scan(&cid, &name, &kind, &notNull, &defaultValue, &primaryKey); err != nil {
			t.Fatal(err)
		}
		if name == "result_version" {
			t.Fatal("candidate schema contains mutable result_version")
		}
	}
	var applications int
	if err := db.QueryRow(`SELECT COUNT(*) FROM optimization_applications`).Scan(&applications); err != nil || applications != 1 {
		t.Fatalf("application records=%d err=%v", applications, err)
	}
}

func TestConcurrentCandidatesAcrossConnectionsHaveOneCASWinner(t *testing.T) {
	clk := clock.NewFake(time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC))
	path := filepath.Join(t.TempDir(), "optimization.db")
	one := openStore(t, path, clk, optimization.Evidence{Status: optimization.EvidenceComplete, Digest: "evidence-v1"})
	two := openStore(t, path, clk, optimization.Evidence{Status: optimization.EvidenceComplete, Digest: "evidence-v1"})
	view, _ := one.Read(context.Background())
	makePreview := func(store *optimization.Store, option string) optimization.Preview {
		preview, err := store.Preview(context.Background(), optimization.PreviewRequest{
			BaseVersion: view.Snapshot.Version, Category: optimization.CategoryExitProtection,
			Changes: map[string]string{"exit.common-policy": option}, Source: optimization.SourceServerPreset,
			Reason: optimization.ReasonServerPreset,
		})
		if err != nil {
			t.Fatal(err)
		}
		return preview
	}
	p1 := makePreview(one, "SAFE")
	p2 := makePreview(two, "BALANCED")
	clk.Advance(3 * time.Second)
	var wg sync.WaitGroup
	errs := make(chan error, 2)
	for _, pair := range []struct {
		store *optimization.Store
		token string
	}{{one, p1.Capability}, {two, p2.Capability}} {
		wg.Add(1)
		go func(store *optimization.Store, token string) {
			defer wg.Done()
			_, err := store.Apply(context.Background(), optimization.ApplyRequest{Capability: token, Confirmed: true})
			errs <- err
		}(pair.store, pair.token)
	}
	wg.Wait()
	close(errs)
	wins, conflicts := 0, 0
	for err := range errs {
		switch {
		case err == nil:
			wins++
		case errors.Is(err, optimization.ErrVersionConflict):
			conflicts++
		default:
			t.Errorf("unexpected concurrent error: %v", err)
		}
	}
	if wins != 1 || conflicts != 1 {
		t.Fatalf("wins=%d conflicts=%d", wins, conflicts)
	}
}

func TestRollbackCreatesANewVersionAndNeverRewritesHistory(t *testing.T) {
	clk := clock.NewFake(time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC))
	store := openStore(t, filepath.Join(t.TempDir(), "optimization.db"), clk,
		optimization.Evidence{Status: optimization.EvidenceComplete, Digest: "evidence-v1"})
	initial, _ := store.Read(context.Background())
	first, _ := store.Preview(context.Background(), optimization.PreviewRequest{
		BaseVersion: initial.Snapshot.Version, Category: optimization.CategoryExitProtection,
		Changes: map[string]string{"exit.common-policy": "SAFE"}, Source: optimization.SourceServerPreset,
		Reason: optimization.ReasonServerPreset,
	})
	clk.Advance(3 * time.Second)
	applied, err := store.Apply(context.Background(), optimization.ApplyRequest{Capability: first.Capability, Confirmed: true})
	if err != nil {
		t.Fatal(err)
	}
	second, err := store.Preview(context.Background(), optimization.PreviewRequest{
		BaseVersion: applied.Snapshot.Version, Category: optimization.CategoryExitProtection,
		Changes: map[string]string{"exit.common-policy": "BALANCED"}, Source: optimization.SourceServerPreset,
		Reason: optimization.ReasonServerPreset,
	})
	if err != nil {
		t.Fatal(err)
	}
	clk.Advance(3 * time.Second)
	secondApplied, err := store.Apply(context.Background(), optimization.ApplyRequest{Capability: second.Capability, Confirmed: true})
	if err != nil {
		t.Fatal(err)
	}
	rollback, err := store.PreviewRollback(context.Background(), optimization.RollbackPreviewRequest{
		BaseVersion: secondApplied.Snapshot.Version, TargetVersion: applied.Snapshot.Version,
		Category: optimization.CategoryExitProtection,
	})
	if err != nil {
		t.Fatal(err)
	}
	clk.Advance(3 * time.Second)
	rolledBack, err := store.Apply(context.Background(), optimization.ApplyRequest{Capability: rollback.Capability, Confirmed: true})
	if err != nil {
		t.Fatal(err)
	}
	if rolledBack.Snapshot.Version != secondApplied.Snapshot.Version+1 ||
		rolledBack.Snapshot.Desired["exit.common-policy"] != "SAFE" ||
		rolledBack.Snapshot.Reason != optimization.ReasonRollback {
		t.Fatalf("rollback result = %+v", rolledBack.Snapshot)
	}
	view, _ := store.Read(context.Background())
	if len(view.History) != 4 || view.History[3].Version != initial.Snapshot.Version || len(view.Audit) != 3 {
		t.Fatalf("history was rewritten: %+v", view.History)
	}
	toUnapproved, err := store.PreviewRollback(context.Background(), optimization.RollbackPreviewRequest{
		BaseVersion: rolledBack.Snapshot.Version, TargetVersion: initial.Snapshot.Version,
		Category: optimization.CategoryExitProtection,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(toUnapproved.Changes) != 1 || toUnapproved.Changes[0].AfterOptionID != "" {
		t.Fatalf("rollback-to-unapproved preview = %+v", toUnapproved)
	}
	clk.Advance(3 * time.Second)
	unapproved, err := store.Apply(context.Background(), optimization.ApplyRequest{
		Capability: toUnapproved.Capability, Confirmed: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, present := unapproved.Snapshot.Desired["exit.common-policy"]; present ||
		unapproved.Snapshot.Version != rolledBack.Snapshot.Version+1 {
		t.Fatalf("rollback-to-unapproved snapshot = %+v", unapproved.Snapshot)
	}
}

func TestAppliedSnapshotAndConsumedCapabilitySurviveProcessReopen(t *testing.T) {
	clk := clock.NewFake(time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC))
	path := filepath.Join(t.TempDir(), "optimization.db")
	store := openStore(t, path, clk, optimization.Evidence{Status: optimization.EvidenceComplete, Digest: "evidence-v1"})
	view, _ := store.Read(context.Background())
	preview, err := store.Preview(context.Background(), optimization.PreviewRequest{
		BaseVersion: view.Snapshot.Version, Category: optimization.CategoryExitProtection,
		Changes: map[string]string{"exit.common-policy": "SAFE"}, Source: optimization.SourceServerPreset,
		Reason: optimization.ReasonServerPreset,
	})
	if err != nil {
		t.Fatal(err)
	}
	clk.Advance(3 * time.Second)
	result, err := store.Apply(context.Background(), optimization.ApplyRequest{Capability: preview.Capability, Confirmed: true})
	if err != nil {
		t.Fatal(err)
	}
	if err := store.Close(); err != nil {
		t.Fatal(err)
	}

	reopened, err := optimization.Open(context.Background(), optimization.Options{
		Path: path, Registry: testRegistry(t), Clock: clk, Actor: "operator:test",
		Evidence: evidenceProvider{evidence: optimization.Evidence{Status: optimization.EvidenceComplete, Digest: "evidence-v1"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	defer reopened.Close()
	afterCrash, err := reopened.Read(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if afterCrash.Snapshot.Version != result.Snapshot.Version || len(afterCrash.Audit) != 1 {
		t.Fatalf("reopened state = %+v audit=%+v", afterCrash.Snapshot, afterCrash.Audit)
	}
	replay, err := reopened.Apply(context.Background(), optimization.ApplyRequest{Capability: preview.Capability, Confirmed: true})
	if err != nil || !replay.Replayed || replay.Snapshot.Version != result.Snapshot.Version {
		t.Fatalf("durable replay = %+v, %v", replay, err)
	}
}
