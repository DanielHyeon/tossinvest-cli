package performance

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"
)

const (
	MaxAttributionEvidenceRows   = 1_000_000
	MaxAttributionGenerationRows = 10_000
	MaxAttributionQueryRows      = 10_000
	AttributionSemantics         = "lane-attribution/v1"
)

// AttributionEvidenceWindow is the exact journal window used to rebuild one
// account's replaceable derived attribution generation.
type AttributionEvidenceWindow struct {
	AccountRef       string
	ClosedAfter      time.Time
	ClosedAtOrBefore time.Time
}

// AttributionRebuild contains only supplied authoritative evidence. Unavailable
// rows are explicit projections for older sources that cannot supply the fill,
// cost or FX evidence required by BuildDerivedAttributionStore.
type AttributionRebuild struct {
	ID           string
	AccountRef   string
	CalculatedAt time.Time
	Positions    []PositionEvidence
	FillDeltas   []FillDelta
	Unavailable  []Attribution
}

type attributionHead struct {
	AccountRef       string
	RebuildID        string
	CalculatedAt     string
	SemanticsVersion string
	SourceDigest     string
	SourceJSON       string
	RowCount         int
}

type attributionRowEnvelope struct {
	AccountRef    string
	RebuildID     string
	Ordinal       int
	Market        string
	Ticker        string
	LaneID        string
	LaneVersion   string
	CampaignID    string
	LegID         string
	PositionID    string
	PolicyID      string
	PolicyVersion string
	LineageStatus Status
	RowJSON       string
}

type attributionQueryer interface {
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
	QueryRowContext(context.Context, string, ...any) *sql.Row
}

// NewUnavailableAttribution creates a non-numeric row for evidence gaps. It is
// deliberately incapable of accepting amounts, inferred FX, or default costs.
func NewUnavailableAttribution(key AttributionKey, missingLineage, missingMeasurements []string,
	sourceCurrency, reportingCurrency string,
) Attribution {
	return Attribution{
		Key: key, LineageStatus: StatusLinkMissing,
		MissingLineage: sortedStrings(missingLineage), MissingMeasurements: sortedStrings(missingMeasurements),
		Source:    notMeasuredBreakdown(strings.TrimSpace(sourceCurrency)),
		Reporting: notMeasuredBreakdown(strings.TrimSpace(reportingCurrency)),
	}
}

func (s *Store) migrateAttributionV2(ctx context.Context, schema string) error {
	version, err := s.SchemaVersion(ctx)
	if err != nil {
		return err
	}
	if version > SchemaVersion {
		return fmt.Errorf("%w: found %d, understand %d", ErrSchemaTooNew, version, SchemaVersion)
	}
	if version >= 2 {
		return nil
	}
	if version != 1 {
		return fmt.Errorf("performance: attribution migration requires schema v1, found %d", version)
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("performance: starting attribution migration: %w", err)
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, schema); err != nil {
		return fmt.Errorf("performance: applying schema v2: %w", err)
	}
	transactionPhaseHook("migration_v2_after_schema")
	if _, err := tx.ExecContext(ctx, "PRAGMA user_version = 2"); err != nil {
		return fmt.Errorf("performance: recording schema v2: %w", err)
	}
	transactionPhaseHook("migration_v2_after_version")
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("performance: committing schema v2: %w", err)
	}
	return nil
}

// PersistAttributionRebuild atomically replaces one account's rebuildable head.
// A replay of the current ID must have the exact same canonical evidence bytes.
func (s *Store) PersistAttributionRebuild(ctx context.Context, rebuild AttributionRebuild) error {
	rows, sourceJSON, sourceDigest, err := prepareAttributionRebuild(rebuild)
	if err != nil {
		return err
	}
	var canonical AttributionRebuild
	if err := json.Unmarshal(sourceJSON, &canonical); err != nil {
		return fmt.Errorf("performance: decoding canonical attribution rebuild: %w", err)
	}
	rebuild = canonical
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("performance: starting attribution rebuild: %w", err)
	}
	defer tx.Rollback()
	var currentID, currentDigest string
	err = tx.QueryRowContext(ctx, `SELECT rebuild_id,source_digest FROM attribution_rebuilds WHERE account_ref=?`, rebuild.AccountRef).
		Scan(&currentID, &currentDigest)
	if err == nil && currentID == rebuild.ID {
		if currentDigest != sourceDigest {
			return fmt.Errorf("%w: attribution rebuild %s", ErrImmutableConflict, rebuild.ID)
		}
		if _, err := validateAttributionGeneration(ctx, tx, rebuild.AccountRef); err != nil {
			return err
		}
		return nil
	}
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("performance: reading attribution rebuild head: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM attribution_rows WHERE account_ref=?`, rebuild.AccountRef); err != nil {
		return fmt.Errorf("performance: replacing attribution rows: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO attribution_rebuilds
		(account_ref,rebuild_id,calculated_at,semantics_version,source_digest,source_json,row_count)
		VALUES(?,?,?,?,?,?,?) ON CONFLICT(account_ref) DO UPDATE SET
		rebuild_id=excluded.rebuild_id,calculated_at=excluded.calculated_at,
		semantics_version=excluded.semantics_version,source_digest=excluded.source_digest,
		source_json=excluded.source_json,row_count=excluded.row_count`,
		rebuild.AccountRef, rebuild.ID, timestamp(rebuild.CalculatedAt), AttributionSemantics,
		sourceDigest, string(sourceJSON), len(rows)); err != nil {
		return fmt.Errorf("performance: writing attribution rebuild head: %w", err)
	}
	for index, row := range rows {
		rowJSON, err := json.Marshal(row)
		if err != nil {
			return fmt.Errorf("performance: encoding attribution row %d: %w", index, err)
		}
		envelope := attributionRowEnvelope{
			AccountRef: rebuild.AccountRef, RebuildID: rebuild.ID, Ordinal: index,
			Market: row.Key.Market, Ticker: row.Key.Ticker, LaneID: row.Key.LaneID,
			LaneVersion: row.Key.LaneVersion, CampaignID: row.Key.CampaignID, LegID: row.Key.LegID,
			PositionID: row.Key.PositionID, PolicyID: row.Key.PolicyID, PolicyVersion: row.Key.PolicyVersion,
			LineageStatus: row.LineageStatus, RowJSON: string(rowJSON),
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO attribution_rows
			(account_ref,rebuild_id,ordinal,market,ticker,lane_id,lane_version,campaign_id,leg_id,
			position_id,policy_id,policy_version,lineage_status,row_digest,row_json)
			VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`, rebuild.AccountRef, rebuild.ID, index,
			row.Key.Market, row.Key.Ticker, nullable(row.Key.LaneID), nullable(row.Key.LaneVersion),
			nullable(row.Key.CampaignID), nullable(row.Key.LegID), row.Key.PositionID,
			nullable(row.Key.PolicyID), nullable(row.Key.PolicyVersion), row.LineageStatus,
			attributionRowDigest(envelope), string(rowJSON)); err != nil {
			return fmt.Errorf("performance: writing attribution row %d: %w", index, err)
		}
	}
	transactionPhaseHook("attribution_after_rows")
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("performance: committing attribution rebuild: %w", err)
	}
	return nil
}

func prepareAttributionRebuild(rebuild AttributionRebuild) ([]Attribution, []byte, string, error) {
	var err error
	rebuild, err = canonicalAttributionRebuild(rebuild)
	if err != nil {
		return nil, nil, "", err
	}
	if rebuild.ID == "" || rebuild.AccountRef == "" || rebuild.CalculatedAt.IsZero() {
		return nil, nil, "", errors.New("performance: attribution rebuild identity, account and calculation time are required")
	}
	if len(rebuild.Positions)+len(rebuild.FillDeltas)+len(rebuild.Unavailable) > MaxAttributionEvidenceRows {
		return nil, nil, "", errors.New("performance: attribution rebuild evidence exceeds bounded maximum")
	}
	derived, err := BuildDerivedAttributionStore(rebuild.Positions, rebuild.FillDeltas)
	if err != nil {
		return nil, nil, "", err
	}
	rows, err := derived.QueryAttribution(AttributionQuery{IncludeLinkMissing: true})
	if err != nil {
		return nil, nil, "", err
	}
	for _, unavailable := range rebuild.Unavailable {
		rows = append(rows, unavailable)
	}
	if len(rows) > MaxAttributionGenerationRows {
		return nil, nil, "", errors.New("performance: attribution generation exceeds bounded maximum")
	}
	sort.SliceStable(rows, func(i, j int) bool { return attributionSortKey(rows[i]) < attributionSortKey(rows[j]) })
	for index := 1; index < len(rows); index++ {
		if attributionSortKey(rows[index-1]) == attributionSortKey(rows[index]) {
			return nil, nil, "", fmt.Errorf("%w: %s", ErrDuplicateAttribution, attributionSortKey(rows[index]))
		}
	}
	sourceJSON, err := json.Marshal(rebuild)
	if err != nil {
		return nil, nil, "", fmt.Errorf("performance: encoding attribution evidence: %w", err)
	}
	return rows, sourceJSON, digestBytes(sourceJSON), nil
}

func canonicalAttributionRebuild(rebuild AttributionRebuild) (AttributionRebuild, error) {
	rebuild.ID = strings.TrimSpace(rebuild.ID)
	rebuild.AccountRef = strings.TrimSpace(rebuild.AccountRef)
	rebuild.CalculatedAt = rebuild.CalculatedAt.UTC()
	for index := range rebuild.Positions {
		rebuild.Positions[index] = normalizePosition(rebuild.Positions[index])
	}
	sort.SliceStable(rebuild.Positions, func(i, j int) bool {
		left, right := rebuild.Positions[i], rebuild.Positions[j]
		return positionKey(left.Market, left.PositionID, left.CampaignID, left.LegID) <
			positionKey(right.Market, right.PositionID, right.CampaignID, right.LegID)
	})
	normalizedFillDeltas := make([]FillDelta, len(rebuild.FillDeltas))
	for index := range rebuild.FillDeltas {
		normalizedFillDeltas[index] = normalizeFillDelta(rebuild.FillDeltas[index])
	}
	sort.SliceStable(normalizedFillDeltas, func(i, j int) bool {
		left, _ := json.Marshal(normalizedFillDeltas[i])
		right, _ := json.Marshal(normalizedFillDeltas[j])
		return string(left) < string(right)
	})
	fillDeltas, err := deduplicateFillDeltas(normalizedFillDeltas)
	if err != nil {
		return AttributionRebuild{}, err
	}
	sort.SliceStable(fillDeltas, func(i, j int) bool {
		left, _ := json.Marshal(fillDeltas[i])
		right, _ := json.Marshal(fillDeltas[j])
		return string(left) < string(right)
	})
	rebuild.FillDeltas = fillDeltas
	for index := range rebuild.Unavailable {
		row := cloneAttribution(rebuild.Unavailable[index])
		normalizeUnavailableAttribution(&row)
		if err := validateUnavailableAttribution(row); err != nil {
			return AttributionRebuild{}, err
		}
		rebuild.Unavailable[index] = row
	}
	sort.SliceStable(rebuild.Unavailable, func(i, j int) bool {
		return attributionSortKey(rebuild.Unavailable[i]) < attributionSortKey(rebuild.Unavailable[j])
	})
	return rebuild, nil
}

func normalizeUnavailableAttribution(row *Attribution) {
	key := &row.Key
	key.Market = strings.ToUpper(strings.TrimSpace(key.Market))
	for _, value := range []*string{&key.Ticker, &key.LaneID, &key.LaneVersion, &key.CampaignID,
		&key.LegID, &key.PositionID, &key.PolicyID, &key.PolicyVersion} {
		*value = strings.TrimSpace(*value)
	}
	row.MissingLineage = sortedStrings(row.MissingLineage)
	row.MissingMeasurements = sortedStrings(row.MissingMeasurements)
	for index := range row.ObservedLineage {
		normalizeLineage(&row.ObservedLineage[index])
		row.ObservedLineage[index].Market = strings.ToUpper(row.ObservedLineage[index].Market)
	}
	sort.SliceStable(row.ObservedLineage, func(i, j int) bool {
		left, _ := json.Marshal(row.ObservedLineage[i])
		right, _ := json.Marshal(row.ObservedLineage[j])
		return string(left) < string(right)
	})
}

func validateUnavailableAttribution(row Attribution) error {
	if row.Key.Market != "KR" && row.Key.Market != "US" || row.Key.PositionID == "" || row.LineageStatus != StatusLinkMissing {
		return errors.New("performance: unavailable attribution requires market, position and link_missing status")
	}
	if len(row.MissingLineage) == 0 || len(row.MissingMeasurements) == 0 {
		return errors.New("performance: unavailable attribution must name missing lineage and measurements")
	}
	if len(row.ObservedLineage) == 0 {
		return errors.New("performance: unavailable attribution requires observed lineage")
	}
	for _, value := range append(append([]string(nil), row.MissingLineage...), row.MissingMeasurements...) {
		if strings.TrimSpace(value) == "" {
			return errors.New("performance: unavailable attribution missing-field names cannot be empty")
		}
	}
	for _, lineage := range row.ObservedLineage {
		if lineage.Market != row.Key.Market || lineage.Ticker != row.Key.Ticker ||
			lineage.LaneID != row.Key.LaneID || lineage.LaneVersion != row.Key.LaneVersion ||
			lineage.CampaignID != row.Key.CampaignID || lineage.LegID != row.Key.LegID ||
			lineage.PositionID != row.Key.PositionID || lineage.PolicyID != row.Key.PolicyID ||
			lineage.PolicyVersion != row.Key.PolicyVersion {
			return ErrLineageConflict
		}
		if !equalStrings(row.MissingLineage, missingObservedLineage(lineage)) {
			return errors.New("performance: unavailable attribution missing lineage does not match observed evidence")
		}
	}
	if row.AcquiredQuantity != "" || row.ClosedQuantity != "" || row.ResidualQuantity != "" ||
		row.TotalEntryBasis != "" || row.AllocatedEntryBasis != "" || row.ResidualEntryBasis != "" ||
		row.FullyClosed || len(row.EntryFills) != 0 || len(row.CloseLegs) != 0 {
		return errors.New("performance: unavailable attribution cannot carry inferred quantities, basis or fill projections")
	}
	for _, breakdown := range []PnLBreakdown{row.Source, row.Reporting} {
		if breakdown.FXSource != "" || breakdown.FXSourceVersion != "" || !breakdown.FXAsOf.IsZero() || breakdown.RoundingVersion != "" {
			return errors.New("performance: unavailable attribution cannot carry FX or rounding evidence")
		}
	}
	for _, metric := range []AmountMetric{
		row.Source.GrossPnL, row.Source.EntryFees, row.Source.ExitFees, row.Source.Taxes, row.Source.FXCost, row.Source.NetPnL, row.Source.RoundingResidual,
		row.Reporting.GrossPnL, row.Reporting.EntryFees, row.Reporting.ExitFees, row.Reporting.Taxes, row.Reporting.FXCost, row.Reporting.NetPnL, row.Reporting.RoundingResidual,
	} {
		if metric.Status != StatusNotMeasured || metric.Value != "" {
			return errors.New("performance: unavailable attribution cannot carry measured amounts")
		}
	}
	return nil
}

func missingObservedLineage(lineage CompositeLineage) []string {
	fields := []struct{ name, value string }{
		{"candidate_id", lineage.CandidateID}, {"lane_id", lineage.LaneID}, {"lane_version", lineage.LaneVersion},
		{"campaign_id", lineage.CampaignID}, {"leg_id", lineage.LegID}, {"decision_id", lineage.DecisionID},
		{"attempt_id", lineage.AttemptID}, {"order_id", lineage.OrderID}, {"fill_id", lineage.FillID},
		{"position_id", lineage.PositionID}, {"close_id", lineage.CloseID}, {"close_leg_id", lineage.CloseLegID},
		{"policy_id", lineage.PolicyID}, {"policy_version", lineage.PolicyVersion},
	}
	missing := make([]string, 0, len(fields))
	for _, field := range fields {
		if field.value == "" {
			missing = append(missing, field.name)
		}
	}
	sort.Strings(missing)
	return missing
}

func equalStrings(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}

func (s *Store) AttributionRows(ctx context.Context, accountRef string, query AttributionQuery, limit int) ([]Attribution, error) {
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return nil, fmt.Errorf("performance: starting attribution read snapshot: %w", err)
	}
	defer tx.Rollback()
	rows, err := queryAttributionRows(ctx, tx, accountRef, query, limit)
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("performance: committing attribution read snapshot: %w", err)
	}
	return rows, nil
}

// AttributionEvidence returns the exact canonical source generation from which
// the currently persisted rows can be rebuilt and revalidated.
func (s *Store) AttributionEvidence(ctx context.Context, accountRef string) (AttributionRebuild, error) {
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return AttributionRebuild{}, fmt.Errorf("performance: starting attribution evidence snapshot: %w", err)
	}
	defer tx.Rollback()
	rebuild, err := queryAttributionEvidence(ctx, tx, accountRef)
	if err != nil {
		return AttributionRebuild{}, err
	}
	if err := tx.Commit(); err != nil {
		return AttributionRebuild{}, fmt.Errorf("performance: committing attribution evidence snapshot: %w", err)
	}
	return rebuild, nil
}

func queryAttributionEvidence(ctx context.Context, db attributionQueryer, accountRef string) (AttributionRebuild, error) {
	accountRef = strings.TrimSpace(accountRef)
	if accountRef == "" {
		return AttributionRebuild{}, errors.New("performance: attribution account is required")
	}
	head, err := validateAttributionGeneration(ctx, db, accountRef)
	if err != nil {
		return AttributionRebuild{}, err
	}
	var rebuild AttributionRebuild
	if err := json.Unmarshal([]byte(head.SourceJSON), &rebuild); err != nil {
		return AttributionRebuild{}, fmt.Errorf("performance: decoding attribution evidence: %w", err)
	}
	return rebuild, nil
}

func queryAttributionRows(ctx context.Context, db attributionQueryer, accountRef string, query AttributionQuery, limit int) ([]Attribution, error) {
	accountRef = strings.TrimSpace(accountRef)
	query.Market = strings.ToUpper(strings.TrimSpace(query.Market))
	query.Ticker = strings.TrimSpace(query.Ticker)
	if accountRef == "" {
		return nil, errors.New("performance: attribution account is required")
	}
	if query.Ticker != "" && query.Market == "" {
		return nil, ErrMarketRequired
	}
	if query.Market != "" && query.Market != "KR" && query.Market != "US" {
		return nil, errors.New("performance: attribution market must be KR or US")
	}
	if limit <= 0 || limit > MaxAttributionQueryRows {
		return nil, fmt.Errorf("performance: attribution query limit must be between 1 and %d", MaxAttributionQueryRows)
	}
	head, err := validateAttributionGeneration(ctx, db, accountRef)
	if err != nil {
		return nil, err
	}
	clauses := []string{"account_ref=?"}
	args := []any{accountRef}
	for column, value := range map[string]string{
		"market": query.Market, "ticker": query.Ticker, "lane_id": strings.TrimSpace(query.LaneID),
		"lane_version": strings.TrimSpace(query.LaneVersion), "campaign_id": strings.TrimSpace(query.CampaignID),
		"leg_id": strings.TrimSpace(query.LegID),
	} {
		if value != "" {
			clauses = append(clauses, column+"=?")
			args = append(args, value)
		}
	}
	if !query.IncludeLinkMissing {
		clauses = append(clauses, "lineage_status=?")
		args = append(args, StatusComplete)
	}
	clauses = append(clauses, "rebuild_id=?")
	args = append(args, head.RebuildID)
	args = append(args, limit)
	rows, err := db.QueryContext(ctx, `SELECT account_ref,rebuild_id,ordinal,market,ticker,
		COALESCE(lane_id,''),COALESCE(lane_version,''),COALESCE(campaign_id,''),COALESCE(leg_id,''),
		position_id,COALESCE(policy_id,''),COALESCE(policy_version,''),lineage_status,row_json,row_digest
		FROM attribution_rows WHERE `+
		strings.Join(clauses, " AND ")+` ORDER BY market,ticker,lane_id,lane_version,campaign_id,leg_id,position_id,ordinal LIMIT ?`, args...)
	if err != nil {
		return nil, fmt.Errorf("performance: querying attribution rows: %w", err)
	}
	defer rows.Close()
	var out []Attribution
	for rows.Next() {
		var envelope attributionRowEnvelope
		var digest string
		if err := rows.Scan(&envelope.AccountRef, &envelope.RebuildID, &envelope.Ordinal, &envelope.Market,
			&envelope.Ticker, &envelope.LaneID, &envelope.LaneVersion, &envelope.CampaignID, &envelope.LegID,
			&envelope.PositionID, &envelope.PolicyID, &envelope.PolicyVersion, &envelope.LineageStatus,
			&envelope.RowJSON, &digest); err != nil {
			return nil, fmt.Errorf("performance: scanning attribution row: %w", err)
		}
		if attributionRowDigest(envelope) != digest {
			return nil, fmt.Errorf("%w: attribution row bytes", ErrImmutableConflict)
		}
		var row Attribution
		if err := json.Unmarshal([]byte(envelope.RowJSON), &row); err != nil {
			return nil, fmt.Errorf("performance: decoding attribution row: %w", err)
		}
		if err := validateAttributionRowEnvelope(envelope, row); err != nil {
			return nil, err
		}
		out = append(out, row)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("performance: querying attribution rows: %w", err)
	}
	return out, nil
}

func validateAttributionGeneration(ctx context.Context, db attributionQueryer, accountRef string) (attributionHead, error) {
	accountRef = strings.TrimSpace(accountRef)
	var head attributionHead
	if err := db.QueryRowContext(ctx, `SELECT account_ref,rebuild_id,calculated_at,semantics_version,
		source_digest,source_json,row_count FROM attribution_rebuilds WHERE account_ref=?`, accountRef).
		Scan(&head.AccountRef, &head.RebuildID, &head.CalculatedAt, &head.SemanticsVersion,
			&head.SourceDigest, &head.SourceJSON, &head.RowCount); err != nil {
		return attributionHead{}, fmt.Errorf("performance: reading attribution rebuild head: %w", err)
	}
	if head.AccountRef != accountRef || head.SemanticsVersion != AttributionSemantics ||
		head.RowCount < 0 || head.RowCount > MaxAttributionGenerationRows ||
		digestBytes([]byte(head.SourceJSON)) != head.SourceDigest {
		return attributionHead{}, fmt.Errorf("%w: attribution head envelope", ErrImmutableConflict)
	}
	var rebuild AttributionRebuild
	if err := json.Unmarshal([]byte(head.SourceJSON), &rebuild); err != nil {
		return attributionHead{}, fmt.Errorf("performance: decoding attribution evidence: %w", err)
	}
	if rebuild.AccountRef != head.AccountRef || rebuild.ID != head.RebuildID ||
		timestamp(rebuild.CalculatedAt) != head.CalculatedAt {
		return attributionHead{}, fmt.Errorf("%w: attribution source envelope", ErrImmutableConflict)
	}
	rows, err := db.QueryContext(ctx, `SELECT account_ref,rebuild_id,ordinal,market,ticker,
		COALESCE(lane_id,''),COALESCE(lane_version,''),COALESCE(campaign_id,''),COALESCE(leg_id,''),
		position_id,COALESCE(policy_id,''),COALESCE(policy_version,''),lineage_status,row_json,row_digest
		FROM attribution_rows WHERE account_ref=? ORDER BY rebuild_id,ordinal`, accountRef)
	if err != nil {
		return attributionHead{}, fmt.Errorf("performance: reading attribution generation: %w", err)
	}
	defer rows.Close()
	count := 0
	for rows.Next() {
		var envelope attributionRowEnvelope
		var digest string
		if err := rows.Scan(&envelope.AccountRef, &envelope.RebuildID, &envelope.Ordinal, &envelope.Market,
			&envelope.Ticker, &envelope.LaneID, &envelope.LaneVersion, &envelope.CampaignID, &envelope.LegID,
			&envelope.PositionID, &envelope.PolicyID, &envelope.PolicyVersion, &envelope.LineageStatus,
			&envelope.RowJSON, &digest); err != nil {
			return attributionHead{}, fmt.Errorf("performance: scanning attribution generation: %w", err)
		}
		if envelope.RebuildID != head.RebuildID || envelope.Ordinal != count ||
			attributionRowDigest(envelope) != digest {
			return attributionHead{}, fmt.Errorf("%w: attribution generation row envelope", ErrImmutableConflict)
		}
		var row Attribution
		if err := json.Unmarshal([]byte(envelope.RowJSON), &row); err != nil {
			return attributionHead{}, fmt.Errorf("performance: decoding attribution generation row: %w", err)
		}
		if err := validateAttributionRowEnvelope(envelope, row); err != nil {
			return attributionHead{}, err
		}
		count++
	}
	if err := rows.Err(); err != nil {
		return attributionHead{}, fmt.Errorf("performance: reading attribution generation: %w", err)
	}
	if count != head.RowCount {
		return attributionHead{}, fmt.Errorf("%w: attribution generation row count", ErrImmutableConflict)
	}
	return head, nil
}

func validateAttributionRowEnvelope(envelope attributionRowEnvelope, row Attribution) error {
	if envelope.Market != row.Key.Market || envelope.Ticker != row.Key.Ticker ||
		envelope.LaneID != row.Key.LaneID || envelope.LaneVersion != row.Key.LaneVersion ||
		envelope.CampaignID != row.Key.CampaignID || envelope.LegID != row.Key.LegID ||
		envelope.PositionID != row.Key.PositionID || envelope.PolicyID != row.Key.PolicyID ||
		envelope.PolicyVersion != row.Key.PolicyVersion || envelope.LineageStatus != row.LineageStatus {
		return fmt.Errorf("%w: attribution row shadow scope", ErrImmutableConflict)
	}
	return nil
}

func attributionRowDigest(envelope attributionRowEnvelope) string {
	raw, _ := json.Marshal(envelope)
	return digestBytes(raw)
}

func attributionSortKey(row Attribution) string {
	return strings.Join([]string{row.Key.Market, row.Key.Ticker, row.Key.LaneID, row.Key.LaneVersion,
		row.Key.CampaignID, row.Key.LegID, row.Key.PositionID, row.Key.PolicyID, row.Key.PolicyVersion}, "\x00")
}

func sortedStrings(values []string) []string {
	out := append([]string(nil), values...)
	for index := range out {
		out[index] = strings.TrimSpace(out[index])
	}
	sort.Strings(out)
	return out
}

func digestBytes(value []byte) string {
	sum := sha256.Sum256(value)
	return hex.EncodeToString(sum[:])
}

const schemaV2 = `
CREATE TABLE attribution_rebuilds (
	account_ref TEXT PRIMARY KEY,
	rebuild_id TEXT NOT NULL,
	calculated_at TEXT NOT NULL,
	semantics_version TEXT NOT NULL,
	source_digest TEXT NOT NULL,
	source_json TEXT NOT NULL,
	row_count INTEGER NOT NULL CHECK(row_count >= 0 AND row_count <= 10000)
) STRICT;

CREATE TABLE attribution_rows (
	account_ref TEXT NOT NULL REFERENCES attribution_rebuilds(account_ref) ON DELETE CASCADE,
	rebuild_id TEXT NOT NULL,
	ordinal INTEGER NOT NULL CHECK(ordinal >= 0 AND ordinal < 10000),
	market TEXT NOT NULL CHECK(market IN ('KR','US')),
	ticker TEXT NOT NULL,
	lane_id TEXT,
	lane_version TEXT,
	campaign_id TEXT,
	leg_id TEXT,
	position_id TEXT NOT NULL,
	policy_id TEXT,
	policy_version TEXT,
	lineage_status TEXT NOT NULL CHECK(lineage_status IN ('complete','link_missing')),
	row_digest TEXT NOT NULL,
	row_json TEXT NOT NULL,
	PRIMARY KEY(account_ref,rebuild_id,ordinal)
) STRICT;
CREATE INDEX idx_attribution_rows_scope ON attribution_rows
	(account_ref,market,ticker,lane_id,lane_version,campaign_id,leg_id,position_id);
CREATE UNIQUE INDEX idx_attribution_rows_unique_key ON attribution_rows
	(account_ref,rebuild_id,market,ticker,COALESCE(lane_id,''),COALESCE(lane_version,''),
	 COALESCE(campaign_id,''),COALESCE(leg_id,''),position_id,COALESCE(policy_id,''),COALESCE(policy_version,''));
`
