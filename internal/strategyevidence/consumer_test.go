package strategyevidence

import (
	"context"
	"errors"
	"testing"
	"time"

	marketclock "github.com/JungHoonGhae/tossinvest-cli/internal/clock"
)

func TestDormantSnapshotReadReplaysKRAndUSIndependentlyByIDAndDigest(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	clk := marketclock.NewFake(time.Date(2026, 8, 3, 13, 2, 0, 0, time.UTC))
	store := openTestStoreWithClock(t, clk)
	usHeader := validHeader(marketclock.MarketUS, KindUSParticipation, "dormant-us", "us-rev")
	krHeader := validHeader(marketclock.MarketKR, KindKRNetFlow, "dormant-kr", "kr-rev")
	krHeader.Authority = AuthorityKRX
	krHeader.SourceRecordID = "dormant-kr-record"
	krHeader.Currency = "KRW"
	for _, envelope := range []Envelope{mustEnvelope(t, usHeader, `{"market":"us"}`), mustEnvelope(t, krHeader, `{"market":"kr"}`)} {
		if _, err := store.Append(ctx, envelope); err != nil {
			t.Fatal(err)
		}
	}
	us, err := store.SealSnapshot(ctx, SnapshotQuery{Market: marketclock.MarketUS, Symbol: usHeader.Symbol, IssuerIdentity: usHeader.IssuerIdentity, IssuerMappingVersion: usHeader.IssuerMappingVersion, EvaluationAt: usHeader.IngestedAt.Add(time.Hour), IngestionCutoff: usHeader.IngestedAt.Add(time.Hour)})
	if err != nil {
		t.Fatal(err)
	}
	kr, err := store.SealSnapshot(ctx, SnapshotQuery{Market: marketclock.MarketKR, Symbol: krHeader.Symbol, IssuerIdentity: krHeader.IssuerIdentity, IssuerMappingVersion: krHeader.IssuerMappingVersion, EvaluationAt: krHeader.IngestedAt.Add(time.Hour), IngestionCutoff: krHeader.IngestedAt.Add(time.Hour)})
	if err != nil {
		t.Fatal(err)
	}
	port := NewDormantSnapshotReadPort(store)
	gotKR, err := port.Replay(ctx, marketclock.MarketKR, SnapshotReference{ID: kr.ID, Digest: kr.Digest})
	if err != nil || gotKR.ID != kr.ID || gotKR.Market != marketclock.MarketKR {
		t.Fatalf("KR replay=%+v err=%v", gotKR, err)
	}
	futureUSHeader := usHeader
	futureUSHeader.EvidenceID = "dormant-us-future"
	futureUSHeader.RevisionIdentity = "us-rev-2"
	futureUSHeader.SupersedesRevisionIdentity = usHeader.RevisionIdentity
	futureUSHeader.SourceAvailableAt = usHeader.SourceAvailableAt.Add(2 * time.Hour)
	futureUSHeader.ObservedAt = futureUSHeader.SourceAvailableAt.Add(time.Minute)
	futureUSHeader.SourceEventAt = futureUSHeader.SourceAvailableAt.Add(-time.Minute)
	futureUSHeader.IngestedAt = futureUSHeader.SourceAvailableAt.Add(time.Hour)
	clk.Set(futureUSHeader.SourceAvailableAt.Add(time.Hour))
	if _, err := store.Append(ctx, mustEnvelope(t, futureUSHeader, `{"market":"us-future"}`)); err != nil {
		t.Fatal(err)
	}
	historicalUS, err := port.Replay(ctx, marketclock.MarketUS, SnapshotReference{ID: us.ID, Digest: us.Digest})
	if err != nil || len(historicalUS.Items) != 1 || historicalUS.Items[0].Header().EvidenceID != usHeader.EvidenceID {
		t.Fatalf("future revision changed historical US replay=%+v err=%v", historicalUS, err)
	}
	badUS := SnapshotReference{ID: us.ID, Digest: kr.Digest}
	if _, err := port.Replay(ctx, marketclock.MarketUS, badUS); !errors.Is(err, ErrSnapshotUnavailable) {
		t.Fatalf("US mismatched digest error=%v", err)
	}
	againKR, err := port.Replay(ctx, marketclock.MarketKR, SnapshotReference{ID: kr.ID, Digest: kr.Digest})
	if err != nil || againKR.ID != kr.ID {
		t.Fatalf("US failure gated KR replay=%+v err=%v", againKR, err)
	}
	if _, err := port.Replay(ctx, marketclock.MarketUS, SnapshotReference{ID: kr.ID, Digest: kr.Digest}); !errors.Is(err, ErrSnapshotUnavailable) {
		t.Fatalf("cross-market snapshot accepted: %v", err)
	}
}

func TestDormantSnapshotReadHasNoFallbackForMissingOrMalformedReference(t *testing.T) {
	t.Parallel()
	port := NewDormantSnapshotReadPort(openTestStore(t))
	for _, reference := range []SnapshotReference{{}, {ID: "snapshot-missing", Digest: "missing"}} {
		if _, err := port.Replay(context.Background(), marketclock.MarketKR, reference); !errors.Is(err, ErrSnapshotUnavailable) {
			t.Fatalf("reference=%+v error=%v", reference, err)
		}
	}
}

func TestDormantSnapshotReadRejectsTamperedHeaderScopeAndCutoffs(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		name       string
		assignment string
		arguments  []any
	}{
		{name: "symbol", assignment: "symbol=?", arguments: []any{"MSFT"}},
		{name: "issuer", assignment: "issuer_identity=?", arguments: []any{"issuer-other"}},
		{name: "mapping", assignment: "issuer_mapping_version=?", arguments: []any{"issuer-map-v2"}},
		{name: "cross-market", assignment: "market=?, evidence_kind=?, authority=?, currency=?", arguments: []any{marketclock.MarketKR, KindKRNetFlow, AuthorityKRX, "KRW"}},
		{name: "future-effective-date", assignment: "market_effective_date=?", arguments: []any{"2026-08-04"}},
		{name: "future-dual-cutoff", assignment: "source_event_at=?, source_available_at=?, observed_at=?, ingested_at=?", arguments: []any{
			stamp(time.Date(2026, 8, 3, 15, 0, 0, 0, time.UTC)),
			stamp(time.Date(2026, 8, 3, 15, 1, 0, 0, time.UTC)),
			stamp(time.Date(2026, 8, 3, 15, 2, 0, 0, time.UTC)),
			stamp(time.Date(2026, 8, 3, 15, 3, 0, 0, time.UTC)),
		}},
	} {
		t.Run(test.name, func(t *testing.T) {
			store, snapshot := sealedSnapshotForTamper(t)
			if _, err := store.db.Exec(`DROP TRIGGER evidence_no_update`); err != nil {
				t.Fatal(err)
			}
			arguments := append([]any{}, test.arguments...)
			arguments = append(arguments, snapshot.Items[0].Header().EvidenceID)
			if _, err := store.db.Exec(`UPDATE evidence_records SET `+test.assignment+` WHERE evidence_id=?`, arguments...); err != nil {
				t.Fatal(err)
			}
			port := NewDormantSnapshotReadPort(store)
			if _, err := port.Replay(context.Background(), marketclock.MarketUS, SnapshotReference{ID: snapshot.ID, Digest: snapshot.Digest}); !errors.Is(err, ErrSnapshotUnavailable) {
				t.Fatalf("tampered %s header replay error=%v", test.name, err)
			}
		})
	}
}

func sealedSnapshotForTamper(t *testing.T) (*Store, Snapshot) {
	t.Helper()
	ctx := context.Background()
	store := openTestStore(t)
	header := validHeader(marketclock.MarketUS, KindUSParticipation, "tamper-us", "tamper-rev")
	if _, err := store.Append(ctx, mustEnvelope(t, header, `{"market":"us"}`)); err != nil {
		t.Fatal(err)
	}
	snapshot, err := store.SealSnapshot(ctx, SnapshotQuery{
		Market: marketclock.MarketUS, Symbol: header.Symbol,
		IssuerIdentity: header.IssuerIdentity, IssuerMappingVersion: header.IssuerMappingVersion,
		EvaluationAt: header.IngestedAt.Add(time.Hour), IngestionCutoff: header.IngestedAt.Add(time.Hour),
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(snapshot.Items) != 1 {
		t.Fatalf("sealed snapshot items=%d", len(snapshot.Items))
	}
	return store, snapshot
}
