package strategyevidence

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	marketclock "github.com/JungHoonGhae/tossinvest-cli/internal/clock"
)

func TestOpenReadOnlyReplaysButCannotCreateSnapshotOrAppend(t *testing.T) {
	ctx := context.Background()
	header := validHeader(marketclock.MarketUS, KindUSParticipation, "readonly-evidence", "rev-1")
	now := header.IngestedAt
	path := filepath.Join(t.TempDir(), "evidence.db")
	store, err := Open(ctx, Options{Path: path, Clock: marketclock.NewFake(now)})
	if err != nil {
		t.Fatal(err)
	}
	evidence := mustEnvelope(t, header, `{"flow":1}`)
	if _, err := store.Append(ctx, evidence); err != nil {
		t.Fatal(err)
	}
	snapshot, err := store.SealSnapshot(ctx, SnapshotQuery{Market: marketclock.MarketUS, Symbol: "AAPL", IssuerIdentity: "issuer-aapl",
		IssuerMappingVersion: "issuer-map-v1", EvaluationAt: now, IngestionCutoff: now})
	if err != nil {
		t.Fatal(err)
	}
	if err := store.Close(); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(path, 0o600); err != nil {
		t.Fatal(err)
	}

	readOnly, err := OpenReadOnly(ctx, Options{Path: path, Clock: marketclock.NewFake(now)})
	if err != nil {
		t.Fatal(err)
	}
	defer readOnly.Close()
	replayed, err := NewDormantSnapshotReadPort(readOnly).Replay(ctx, marketclock.MarketUS, SnapshotReference{ID: snapshot.ID, Digest: snapshot.Digest})
	if err != nil || replayed.ID != snapshot.ID || len(replayed.Items) != 1 {
		t.Fatalf("replay=%+v err=%v", replayed, err)
	}
	newHeader := header
	newHeader.EvidenceID = "readonly-new-evidence"
	newHeader.SourceRecordID = "record-2"
	newHeader.RevisionIdentity = "rev-2"
	newEvidence := mustEnvelope(t, newHeader, `{"flow":2}`)
	if _, err := readOnly.Append(ctx, newEvidence); err == nil {
		t.Fatal("read-only evidence store appended a row")
	}
	if _, err := readOnly.SealSnapshot(ctx, SnapshotQuery{Market: marketclock.MarketUS, Symbol: "AAPL", IssuerIdentity: "issuer-aapl",
		IssuerMappingVersion: "issuer-map-v1", EvaluationAt: now, IngestionCutoff: now}); err == nil {
		t.Fatal("read-only evidence store wrote a snapshot")
	}
}

func TestOpenReadOnlyRefusesWrongModeAndSymlink(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "evidence.db")
	store, err := Open(ctx, Options{Path: path})
	if err != nil {
		t.Fatal(err)
	}
	if err := store.Close(); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(path, 0o640); err != nil {
		t.Fatal(err)
	}
	if _, err := OpenReadOnly(ctx, Options{Path: path}); !errors.Is(err, ErrSnapshotUnavailable) {
		t.Fatalf("wrong mode error=%v", err)
	}
	if err := os.Chmod(path, 0o600); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(filepath.Dir(path), "evidence-link.db")
	if err := os.Symlink(path, link); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}
	if _, err := OpenReadOnly(ctx, Options{Path: link}); !errors.Is(err, ErrSnapshotUnavailable) {
		t.Fatalf("symlink error=%v", err)
	}
}
