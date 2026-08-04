package strategyevidence

import (
	"context"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"

	marketclock "github.com/JungHoonGhae/tossinvest-cli/internal/clock"
)

var ErrSnapshotUnavailable = errors.New("strategy evidence: immutable snapshot unavailable")

type SnapshotReference struct {
	ID     string
	Digest string
}

// DormantSnapshotReadPort replays an already sealed evidence.db snapshot. It
// cannot ingest, enable a lane, change a toggle, call Guardian or dispatch.
type DormantSnapshotReadPort struct{ store *Store }

func NewDormantSnapshotReadPort(store *Store) *DormantSnapshotReadPort {
	return &DormantSnapshotReadPort{store: store}
}

func (port *DormantSnapshotReadPort) Replay(ctx context.Context, market marketclock.Market, reference SnapshotReference) (Snapshot, error) {
	if port == nil || port.store == nil || port.store.db == nil || (market != marketclock.MarketKR && market != marketclock.MarketUS) || !validSnapshotReference(reference) {
		return Snapshot{}, ErrSnapshotUnavailable
	}
	var snapshot Snapshot
	var evaluationAt, ingestionCutoff string
	err := port.store.db.QueryRowContext(ctx, `SELECT snapshot_id,snapshot_digest,market,symbol,issuer_identity,issuer_mapping_version,evaluation_at,ingestion_cutoff FROM snapshots WHERE snapshot_id=? AND snapshot_digest=? AND market=?`, reference.ID, reference.Digest, market).
		Scan(&snapshot.ID, &snapshot.Digest, &snapshot.Market, &snapshot.Symbol, &snapshot.IssuerIdentity, &snapshot.IssuerMappingVersion, &evaluationAt, &ingestionCutoff)
	if errors.Is(err, sql.ErrNoRows) {
		return Snapshot{}, ErrSnapshotUnavailable
	}
	if err != nil {
		return Snapshot{}, fmt.Errorf("strategy evidence: replaying snapshot header: %w", err)
	}
	if snapshot.EvaluationAt, err = parseStamp(evaluationAt); err != nil {
		return Snapshot{}, ErrSnapshotUnavailable
	}
	if snapshot.IngestionCutoff, err = parseStamp(ingestionCutoff); err != nil {
		return Snapshot{}, ErrSnapshotUnavailable
	}
	rows, err := port.store.db.QueryContext(ctx, `SELECT `+envelopeColumns("e")+` FROM snapshot_items i JOIN evidence_records e ON e.evidence_id=i.evidence_id WHERE i.snapshot_id=? AND i.payload_digest=e.payload_digest ORDER BY i.ordinal`, reference.ID)
	if err != nil {
		return Snapshot{}, fmt.Errorf("strategy evidence: replaying snapshot items: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		item, scanErr := scanEnvelope(rows)
		if scanErr != nil {
			return Snapshot{}, ErrSnapshotUnavailable
		}
		if !snapshotItemMatchesQuery(snapshot, item) {
			return Snapshot{}, ErrSnapshotUnavailable
		}
		snapshot.Items = append(snapshot.Items, item)
	}
	if err := rows.Err(); err != nil {
		return Snapshot{}, fmt.Errorf("strategy evidence: replaying snapshot items: %w", err)
	}
	query := SnapshotQuery{Market: snapshot.Market, Symbol: snapshot.Symbol, IssuerIdentity: snapshot.IssuerIdentity, IssuerMappingVersion: snapshot.IssuerMappingVersion, EvaluationAt: snapshot.EvaluationAt, IngestionCutoff: snapshot.IngestionCutoff}
	if digest := snapshotDigest(query, snapshot.Items); digest != snapshot.Digest || snapshot.ID != "snapshot-"+digest {
		return Snapshot{}, ErrSnapshotUnavailable
	}
	return Snapshot{ID: snapshot.ID, Digest: snapshot.Digest, Market: snapshot.Market, Symbol: snapshot.Symbol, IssuerIdentity: snapshot.IssuerIdentity, IssuerMappingVersion: snapshot.IssuerMappingVersion, EvaluationAt: snapshot.EvaluationAt, IngestionCutoff: snapshot.IngestionCutoff, Items: cloneEnvelopes(snapshot.Items)}, nil
}

func snapshotItemMatchesQuery(snapshot Snapshot, item Envelope) bool {
	header := item.Header()
	if header.Market != snapshot.Market || header.Symbol != snapshot.Symbol ||
		header.IssuerIdentity != snapshot.IssuerIdentity || header.IssuerMappingVersion != snapshot.IssuerMappingVersion ||
		header.SourceEventAt.After(snapshot.EvaluationAt) || header.SourceAvailableAt.After(snapshot.EvaluationAt) ||
		header.IngestedAt.After(snapshot.IngestionCutoff) {
		return false
	}
	localDate, err := snapshot.Market.TradingDay(snapshot.EvaluationAt)
	return err == nil && header.MarketEffectiveDate <= localDate
}

func validSnapshotReference(reference SnapshotReference) bool {
	if reference.ID != strings.TrimSpace(reference.ID) || reference.Digest != strings.TrimSpace(reference.Digest) || len(reference.Digest) != 64 || reference.ID != "snapshot-"+reference.Digest {
		return false
	}
	decoded, err := hex.DecodeString(reference.Digest)
	return err == nil && len(decoded) == 32 && strings.ToLower(reference.Digest) == reference.Digest
}
