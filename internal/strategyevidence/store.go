package strategyevidence

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	_ "modernc.org/sqlite"

	marketclock "github.com/JungHoonGhae/tossinvest-cli/internal/clock"
)

const SchemaVersion = 1

var (
	ErrSchemaTooNew     = errors.New("strategy evidence: database schema is newer than this build")
	ErrRevisionConflict = errors.New("strategy evidence: source revision conflict")
)

type Options struct {
	Path        string
	BusyTimeout time.Duration
	Clock       marketclock.Clock
}

type Store struct {
	db   *sql.DB
	path string
	clk  marketclock.Clock
}

type AppendResult struct {
	Evidence   Envelope
	Idempotent bool
}

func Open(ctx context.Context, options Options) (*Store, error) {
	path := filepath.Clean(strings.TrimSpace(options.Path))
	if path == "." || path == "" {
		return nil, fmt.Errorf("strategy evidence: evidence.db path is required")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return nil, fmt.Errorf("strategy evidence: creating data directory: %w", err)
	}
	busy := options.BusyTimeout
	if busy <= 0 {
		busy = 5 * time.Second
	}
	clk := options.Clock
	if clk == nil {
		clk = marketclock.System()
	}
	db, err := sql.Open("sqlite", sqliteDSN(path, busy))
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, err
	}
	store := &Store{db: db, path: path, clk: clk}
	if err := store.migrate(ctx); err != nil {
		_ = db.Close()
		return nil, err
	}
	for _, candidate := range []string{path, path + "-wal", path + "-shm"} {
		if _, err := os.Stat(candidate); err == nil {
			_ = os.Chmod(candidate, 0o600)
		}
	}
	return store, nil
}

func sqliteDSN(path string, busy time.Duration) string {
	q := url.Values{}
	q.Add("_pragma", fmt.Sprintf("busy_timeout(%d)", busy.Milliseconds()))
	q.Add("_pragma", "journal_mode(WAL)")
	q.Add("_pragma", "synchronous(FULL)")
	q.Add("_pragma", "foreign_keys(1)")
	q.Set("_txlock", "immediate")
	uri := url.URL{Scheme: "file", Opaque: (&url.URL{Path: path}).EscapedPath()}
	return uri.String() + "?" + q.Encode()
}

func (s *Store) migrate(ctx context.Context) error {
	var version int
	if err := s.db.QueryRowContext(ctx, `PRAGMA user_version`).Scan(&version); err != nil {
		return err
	}
	if version > SchemaVersion {
		return ErrSchemaTooNew
	}
	if version == SchemaVersion {
		return nil
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	if _, err := tx.ExecContext(ctx, schemaV1); err != nil {
		return fmt.Errorf("strategy evidence: creating schema: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `PRAGMA user_version = 1`); err != nil {
		return err
	}
	return tx.Commit()
}

const schemaV1 = `
CREATE TABLE evidence_records (
  evidence_id TEXT PRIMARY KEY,
  market TEXT NOT NULL,
  symbol TEXT NOT NULL,
  issuer_identity TEXT NOT NULL,
  issuer_mapping_version TEXT NOT NULL,
  evidence_kind TEXT NOT NULL,
  schema_version TEXT NOT NULL,
  authority TEXT NOT NULL,
  source_record_id TEXT NOT NULL,
  revision_identity TEXT NOT NULL,
  supersedes_revision_identity TEXT,
  market_effective_date TEXT NOT NULL,
  source_event_at TEXT NOT NULL,
  source_available_at TEXT NOT NULL,
  observed_at TEXT NOT NULL,
  ingested_at TEXT NOT NULL,
  currency TEXT NOT NULL,
  unit TEXT NOT NULL,
  availability TEXT NOT NULL,
  confidence TEXT NOT NULL,
  canonical_payload BLOB NOT NULL,
  payload_digest TEXT NOT NULL,
  UNIQUE(authority, source_record_id, revision_identity)
) STRICT;
CREATE INDEX evidence_asof_idx ON evidence_records(market, symbol, issuer_identity, issuer_mapping_version, source_available_at, ingested_at, market_effective_date);
CREATE INDEX evidence_supersedes_idx ON evidence_records(authority, source_record_id, supersedes_revision_identity);
CREATE TABLE revision_conflicts (
  conflict_id INTEGER PRIMARY KEY,
  authority TEXT NOT NULL,
  source_record_id TEXT NOT NULL,
  revision_identity TEXT NOT NULL,
  accepted_evidence_id TEXT NOT NULL,
  accepted_digest TEXT NOT NULL,
  attempted_evidence_id TEXT NOT NULL,
  attempted_digest TEXT NOT NULL,
  attempted_payload BLOB NOT NULL,
  quarantined_at TEXT NOT NULL,
  UNIQUE(authority, source_record_id, revision_identity, attempted_digest)
) STRICT;
CREATE TABLE snapshots (
  snapshot_id TEXT PRIMARY KEY,
  market TEXT NOT NULL,
  symbol TEXT NOT NULL,
  issuer_identity TEXT NOT NULL,
  issuer_mapping_version TEXT NOT NULL,
  evaluation_at TEXT NOT NULL,
  ingestion_cutoff TEXT NOT NULL,
  snapshot_digest TEXT NOT NULL UNIQUE
) STRICT;
CREATE TABLE snapshot_items (
  snapshot_id TEXT NOT NULL REFERENCES snapshots(snapshot_id),
  ordinal INTEGER NOT NULL,
  evidence_id TEXT NOT NULL REFERENCES evidence_records(evidence_id),
  payload_digest TEXT NOT NULL,
  PRIMARY KEY(snapshot_id, ordinal),
  UNIQUE(snapshot_id, evidence_id)
) STRICT;
CREATE TRIGGER evidence_no_update BEFORE UPDATE ON evidence_records BEGIN SELECT RAISE(ABORT, 'append-only evidence'); END;
CREATE TRIGGER evidence_no_delete BEFORE DELETE ON evidence_records BEGIN SELECT RAISE(ABORT, 'append-only evidence'); END;
CREATE TRIGGER conflict_no_update BEFORE UPDATE ON revision_conflicts BEGIN SELECT RAISE(ABORT, 'append-only conflict'); END;
CREATE TRIGGER conflict_no_delete BEFORE DELETE ON revision_conflicts BEGIN SELECT RAISE(ABORT, 'append-only conflict'); END;
CREATE TRIGGER snapshot_no_update BEFORE UPDATE ON snapshots BEGIN SELECT RAISE(ABORT, 'immutable snapshot'); END;
CREATE TRIGGER snapshot_no_delete BEFORE DELETE ON snapshots BEGIN SELECT RAISE(ABORT, 'immutable snapshot'); END;
CREATE TRIGGER snapshot_item_no_update BEFORE UPDATE ON snapshot_items BEGIN SELECT RAISE(ABORT, 'immutable snapshot item'); END;
CREATE TRIGGER snapshot_item_no_delete BEFORE DELETE ON snapshot_items BEGIN SELECT RAISE(ABORT, 'immutable snapshot item'); END;
`

func (s *Store) Close() error { return s.db.Close() }
func (s *Store) Path() string { return s.path }

func (s *Store) Append(ctx context.Context, evidence Envelope) (AppendResult, error) {
	if evidence.PayloadDigest() == "" {
		return AppendResult{}, invalid(RefusalPayloadInvalid, "envelope", "use NewEnvelope")
	}
	trustedIngestedAt := s.clk.Now().UTC()
	h := evidence.Header()
	if trustedIngestedAt.Before(h.ObservedAt) {
		return AppendResult{}, invalid(RefusalTimestampInvalid, "ingested_at", "trusted store clock is before observed_at")
	}
	h.IngestedAt = trustedIngestedAt
	stamped, err := NewEnvelope(h, evidence.CanonicalPayload())
	if err != nil {
		return AppendResult{}, err
	}
	evidence = stamped
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return AppendResult{}, err
	}
	defer func() { _ = tx.Rollback() }()
	h = evidence.Header()
	existing, found, err := readByIdentity(ctx, tx, h.Authority, h.SourceRecordID, h.RevisionIdentity)
	if err != nil {
		return AppendResult{}, err
	}
	if found {
		if existing.PayloadDigest() == evidence.PayloadDigest() && sameImmutableProvenance(existing.Header(), h) {
			return AppendResult{Evidence: existing, Idempotent: true}, tx.Commit()
		}
		_, insertErr := tx.ExecContext(ctx, `INSERT OR IGNORE INTO revision_conflicts
          (authority, source_record_id, revision_identity, accepted_evidence_id, accepted_digest,
           attempted_evidence_id, attempted_digest, attempted_payload, quarantined_at)
          VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			h.Authority, h.SourceRecordID, h.RevisionIdentity, existing.Header().EvidenceID,
			existing.PayloadDigest(), h.EvidenceID, evidence.PayloadDigest(), []byte(`{"redacted":true}`), stamp(trustedIngestedAt))
		if insertErr != nil {
			return AppendResult{}, insertErr
		}
		if err := tx.Commit(); err != nil {
			return AppendResult{}, err
		}
		return AppendResult{Evidence: existing}, fmt.Errorf("%w: %s/%s/%s", ErrRevisionConflict, h.Authority, h.SourceRecordID, h.RevisionIdentity)
	}
	if h.SupersedesRevisionIdentity != "" {
		if previous, found, err := readByIdentity(ctx, tx, h.Authority, h.SourceRecordID, h.SupersedesRevisionIdentity); err != nil {
			return AppendResult{}, err
		} else if !found {
			return AppendResult{}, invalid(RefusalIdentityMismatch, "supersedes", "referenced revision is absent")
		} else if previousHeader := previous.Header(); h.Market != previousHeader.Market || h.Symbol != previousHeader.Symbol || h.IssuerIdentity != previousHeader.IssuerIdentity || h.IssuerMappingVersion != previousHeader.IssuerMappingVersion || h.Kind != previousHeader.Kind {
			return AppendResult{}, invalid(RefusalIdentityMismatch, "supersedes", "revision cannot change evidence scope identity")
		}
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO evidence_records
	      (evidence_id, market, symbol, issuer_identity, issuer_mapping_version, evidence_kind, schema_version, authority,
       source_record_id, revision_identity, supersedes_revision_identity, market_effective_date,
       source_event_at, source_available_at, observed_at, ingested_at, currency, unit, availability,
       confidence, canonical_payload, payload_digest)
	      VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NULLIF(?, ''), ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		h.EvidenceID, h.Market, h.Symbol, h.IssuerIdentity, h.IssuerMappingVersion, h.Kind, h.SchemaVersion, h.Authority,
		h.SourceRecordID, h.RevisionIdentity, h.SupersedesRevisionIdentity, h.MarketEffectiveDate,
		stamp(h.SourceEventAt), stamp(h.SourceAvailableAt), stamp(h.ObservedAt), stamp(h.IngestedAt),
		h.Currency, h.Unit, h.Availability, h.Confidence, evidence.CanonicalPayload(), evidence.PayloadDigest())
	if err != nil {
		return AppendResult{}, fmt.Errorf("strategy evidence: appending: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return AppendResult{}, err
	}
	return AppendResult{Evidence: evidence}, nil
}

func sameImmutableProvenance(left, right Header) bool {
	left.EvidenceID, right.EvidenceID = "", ""
	left.IngestedAt, right.IngestedAt = time.Time{}, time.Time{}
	return left == right
}

type SnapshotQuery struct {
	Market               marketclock.Market
	Symbol               string
	IssuerIdentity       string
	IssuerMappingVersion string
	EvaluationAt         time.Time
	IngestionCutoff      time.Time
}

type Snapshot struct {
	ID                   string
	Digest               string
	Market               marketclock.Market
	Symbol               string
	IssuerIdentity       string
	IssuerMappingVersion string
	EvaluationAt         time.Time
	IngestionCutoff      time.Time
	Items                []Envelope
}

func (s *Store) SealSnapshot(ctx context.Context, query SnapshotQuery) (Snapshot, error) {
	query.Symbol = strings.ToUpper(strings.TrimSpace(query.Symbol))
	query.IssuerIdentity = strings.TrimSpace(query.IssuerIdentity)
	query.IssuerMappingVersion = strings.TrimSpace(query.IssuerMappingVersion)
	if (query.Market != marketclock.MarketKR && query.Market != marketclock.MarketUS) || query.Symbol == "" || query.IssuerIdentity == "" || query.IssuerMappingVersion == "" || query.EvaluationAt.IsZero() || query.IngestionCutoff.IsZero() {
		return Snapshot{}, invalid(RefusalIdentityMismatch, "snapshot_query", "market, symbol, issuer, mapping version and both cutoffs are required")
	}
	query.EvaluationAt = query.EvaluationAt.UTC()
	query.IngestionCutoff = query.IngestionCutoff.UTC()
	localDate, err := query.Market.TradingDay(query.EvaluationAt)
	if err != nil {
		return Snapshot{}, err
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Snapshot{}, err
	}
	defer func() { _ = tx.Rollback() }()
	rows, err := tx.QueryContext(ctx, `SELECT `+envelopeColumns("e")+`
      FROM evidence_records e
	     WHERE e.market=? AND e.symbol=? AND e.issuer_identity=? AND e.issuer_mapping_version=? AND e.market_effective_date<=?
       AND e.source_event_at<=? AND e.source_available_at<=? AND e.ingested_at<=?
       AND NOT EXISTS (
          SELECT 1 FROM evidence_records n
           WHERE n.authority=e.authority AND n.source_record_id=e.source_record_id
             AND n.supersedes_revision_identity=e.revision_identity
             AND n.market_effective_date<=? AND n.source_event_at<=?
             AND n.source_available_at<=? AND n.ingested_at<=?
       )
     ORDER BY e.authority, e.source_record_id, e.revision_identity, e.evidence_id`,
		query.Market, query.Symbol, query.IssuerIdentity, query.IssuerMappingVersion, localDate, stamp(query.EvaluationAt), stamp(query.EvaluationAt), stamp(query.IngestionCutoff),
		localDate, stamp(query.EvaluationAt), stamp(query.EvaluationAt), stamp(query.IngestionCutoff))
	if err != nil {
		return Snapshot{}, err
	}
	var items []Envelope
	for rows.Next() {
		evidence, err := scanEnvelope(rows)
		if err != nil {
			_ = rows.Close()
			return Snapshot{}, err
		}
		items = append(items, evidence)
	}
	if err := rows.Close(); err != nil {
		return Snapshot{}, err
	}
	if err := rows.Err(); err != nil {
		return Snapshot{}, err
	}
	digest := snapshotDigest(query, items)
	id := "snapshot-" + digest
	_, err = tx.ExecContext(ctx, `INSERT OR IGNORE INTO snapshots(snapshot_id, market, symbol, issuer_identity, issuer_mapping_version, evaluation_at, ingestion_cutoff, snapshot_digest) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		id, query.Market, query.Symbol, query.IssuerIdentity, query.IssuerMappingVersion, stamp(query.EvaluationAt), stamp(query.IngestionCutoff), digest)
	if err != nil {
		return Snapshot{}, err
	}
	for index, item := range items {
		_, err = tx.ExecContext(ctx, `INSERT OR IGNORE INTO snapshot_items(snapshot_id, ordinal, evidence_id, payload_digest) VALUES (?, ?, ?, ?)`, id, index, item.Header().EvidenceID, item.PayloadDigest())
		if err != nil {
			return Snapshot{}, err
		}
	}
	if err := tx.Commit(); err != nil {
		return Snapshot{}, err
	}
	return Snapshot{ID: id, Digest: digest, Market: query.Market, Symbol: query.Symbol, IssuerIdentity: query.IssuerIdentity, IssuerMappingVersion: query.IssuerMappingVersion, EvaluationAt: query.EvaluationAt, IngestionCutoff: query.IngestionCutoff, Items: cloneEnvelopes(items)}, nil
}

func snapshotDigest(query SnapshotQuery, items []Envelope) string {
	hash := sha256.New()
	for _, value := range []string{string(query.Market), query.Symbol, query.IssuerIdentity, query.IssuerMappingVersion, stamp(query.EvaluationAt), stamp(query.IngestionCutoff)} {
		writeSnapshotDigestField(hash, value)
	}
	ordered := cloneEnvelopes(items)
	sort.Slice(ordered, func(i, j int) bool { return ordered[i].Header().EvidenceID < ordered[j].Header().EvidenceID })
	for _, item := range ordered {
		header := item.Header()
		for _, value := range []string{
			header.EvidenceID, string(header.Market), header.Symbol, header.IssuerIdentity,
			header.IssuerMappingVersion, string(header.Kind), header.SchemaVersion, string(header.Authority),
			header.SourceRecordID, header.RevisionIdentity, header.SupersedesRevisionIdentity,
			header.MarketEffectiveDate, stamp(header.SourceEventAt), stamp(header.SourceAvailableAt),
			stamp(header.ObservedAt), stamp(header.IngestedAt), header.Currency, header.Unit,
			string(header.Availability), string(header.Confidence), item.PayloadDigest(),
		} {
			writeSnapshotDigestField(hash, value)
		}
	}
	return hex.EncodeToString(hash.Sum(nil))
}

func writeSnapshotDigestField(destination interface{ Write([]byte) (int, error) }, value string) {
	var length [8]byte
	binary.BigEndian.PutUint64(length[:], uint64(len(value)))
	_, _ = destination.Write(length[:])
	_, _ = destination.Write([]byte(value))
}

func cloneEnvelopes(items []Envelope) []Envelope {
	out := make([]Envelope, len(items))
	for index, item := range items {
		out[index] = Envelope{header: item.Header(), payload: item.CanonicalPayload(), digest: item.PayloadDigest()}
	}
	return out
}

func (s *Store) EvidenceCount(ctx context.Context) (int, error) {
	var count int
	err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM evidence_records`).Scan(&count)
	return count, err
}

func (s *Store) QuarantineCount(ctx context.Context) (int, error) {
	var count int
	err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM revision_conflicts`).Scan(&count)
	return count, err
}

type scanner interface{ Scan(...any) error }

func envelopeColumns(alias string) string {
	prefix := ""
	if alias != "" {
		prefix = alias + "."
	}
	columns := []string{"evidence_id", "market", "symbol", "issuer_identity", "issuer_mapping_version", "evidence_kind", "schema_version", "authority", "source_record_id", "revision_identity", "COALESCE(supersedes_revision_identity,'')", "market_effective_date", "source_event_at", "source_available_at", "observed_at", "ingested_at", "currency", "unit", "availability", "confidence", "canonical_payload", "payload_digest"}
	for index := range columns {
		if !strings.HasPrefix(columns[index], "COALESCE") {
			columns[index] = prefix + columns[index]
		} else if prefix != "" {
			columns[index] = strings.Replace(columns[index], "supersedes_revision_identity", prefix+"supersedes_revision_identity", 1)
		}
	}
	return strings.Join(columns, ", ")
}

func readByIdentity(ctx context.Context, queryer interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
}, authority SourceAuthority, recordID, revision string) (Envelope, bool, error) {
	row := queryer.QueryRowContext(ctx, `SELECT `+envelopeColumns("")+` FROM evidence_records WHERE authority=? AND source_record_id=? AND revision_identity=?`, authority, recordID, revision)
	evidence, err := scanEnvelope(row)
	if errors.Is(err, sql.ErrNoRows) {
		return Envelope{}, false, nil
	}
	return evidence, err == nil, err
}

func scanEnvelope(row scanner) (Envelope, error) {
	var h Header
	var sourceEvent, sourceAvailable, observed, ingested string
	var payload []byte
	var storedDigest string
	if err := row.Scan(&h.EvidenceID, &h.Market, &h.Symbol, &h.IssuerIdentity, &h.IssuerMappingVersion, &h.Kind, &h.SchemaVersion, &h.Authority,
		&h.SourceRecordID, &h.RevisionIdentity, &h.SupersedesRevisionIdentity, &h.MarketEffectiveDate,
		&sourceEvent, &sourceAvailable, &observed, &ingested, &h.Currency, &h.Unit, &h.Availability, &h.Confidence,
		&payload, &storedDigest); err != nil {
		return Envelope{}, err
	}
	var err error
	if h.SourceEventAt, err = parseStamp(sourceEvent); err != nil {
		return Envelope{}, err
	}
	if h.SourceAvailableAt, err = parseStamp(sourceAvailable); err != nil {
		return Envelope{}, err
	}
	if h.ObservedAt, err = parseStamp(observed); err != nil {
		return Envelope{}, err
	}
	if h.IngestedAt, err = parseStamp(ingested); err != nil {
		return Envelope{}, err
	}
	evidence, err := NewEnvelope(h, payload)
	if err != nil {
		return Envelope{}, err
	}
	if evidence.PayloadDigest() != storedDigest {
		return Envelope{}, fmt.Errorf("strategy evidence: stored payload digest mismatch for %s", h.EvidenceID)
	}
	return evidence, nil
}
