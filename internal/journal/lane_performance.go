package journal

// lane_performance.go exposes the exact, SELECT-only journal half of a049's
// derived performance model. The outcome is the anchor: a missing strategy link
// never makes a closed trade disappear, and no symbol/time proximity is used to
// manufacture one. Only the identifier chain persisted by a047 can populate
// Lineage; every other row keeps Lineage nil (link_missing upstream).

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"
)

// ClosedStrategyLineage is the complete persisted strategy identifier chain.
// The pointer on ClosedStrategyTradeSource is nil unless every field is exact
// and the attempt owns one broker-order link, exactly one fill link, and the
// matching position and close-outcome links.
type ClosedStrategyLineage struct {
	CandidateLifeID          string
	ThresholdVersion         string
	ThresholdSetDigest       string
	EvidenceDigest           string
	LaneID                   string
	LaneVersion              string
	StrategyDecisionIdentity string
	RiskIntentID             string
	StrategyAttemptID        string
	MutationAttemptID        string
	BrokerOrderID            string
	FillID                   string
	PositionID               string
	CloseOutcomeID           string
}

// ClosedStrategyTradeSource is an authoritative closed outcome plus optional
// exact a047 strategy lineage. CostTotal is nil for pre-v15 rows; it is never
// replaced with zero. Policy identity is carried separately because historical
// policy rows may also be incomplete and therefore make the upstream lineage
// status link_missing without changing the frozen outcome facts.
type ClosedStrategyTradeSource struct {
	TradeID    string
	PositionID string
	CloseID    string
	Market     string
	Side       string

	DecisionAt    time.Time
	DecisionPrice string
	EntryAt       time.Time
	EntryPrice    string
	Quantity      string
	CostTotal     *string

	RealizedPnLAfterCosts string
	RealizedR             string
	ClosedAt              time.Time

	PolicyID      string
	PolicyVersion string
	Lineage       *ClosedStrategyLineage
}

// ClosedStrategyTradeSources returns one account's non-adopted frozen outcomes
// in (closedAfter, closedAtOrBefore], ordered deterministically by close then
// position ID. Both bounds and the account are mandatory. All values are bound
// SQL parameters; malformed persisted timestamps or risk bytes fail the read
// rather than being coerced into an apparently valid measurement.
func (r *ReadOnly) ClosedStrategyTradeSources(
	ctx context.Context,
	accountRef string,
	closedAfter time.Time,
	closedAtOrBefore time.Time,
) ([]ClosedStrategyTradeSource, error) {
	account := strings.TrimSpace(accountRef)
	if r == nil || r.db == nil {
		return nil, errors.New("journal: closed strategy trade source requires a read-only journal")
	}
	if account == "" {
		return nil, errors.New("journal: closed strategy trade source requires an account")
	}
	if closedAfter.IsZero() || closedAtOrBefore.IsZero() || !closedAfter.Before(closedAtOrBefore) {
		return nil, errors.New("journal: closed strategy trade source window must have ordered non-zero bounds")
	}

	rows, err := r.db.QueryContext(ctx, closedStrategyTradeSourcesSQL,
		account, account, formatJournalTime(closedAfter), formatJournalTime(closedAtOrBefore))
	if err != nil {
		return nil, fmt.Errorf("journal: reading closed strategy trade sources for %s: %w", account, err)
	}
	defer rows.Close()

	var out []ClosedStrategyTradeSource
	seen := make(map[string]struct{})
	for rows.Next() {
		row, err := scanClosedStrategyTradeSource(rows, account)
		if err != nil {
			return nil, err
		}
		if _, duplicate := seen[row.PositionID]; duplicate {
			return nil, fmt.Errorf("journal: closed strategy trade source %s has ambiguous strategy attempts", row.PositionID)
		}
		seen[row.PositionID] = struct{}{}
		out = append(out, row)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("journal: reading closed strategy trade sources for %s: %w", account, err)
	}
	return out, nil
}

// The execution CTE counts every link before selecting identifiers. MIN is only
// projected when the corresponding count is exactly one; no arbitrary fill can
// escape a 0-or-many cardinality as if it were exact. POSITION and CLOSE_OUTCOME
// are singleton kinds at schema level and are still checked against the outcome
// position here so the read remains fail-closed if constraints were bypassed.
const closedStrategyTradeSourcesSQL = `
WITH execution AS (
 SELECT account_ref, attempt_id,
        SUM(CASE WHEN kind='BROKER_ORDER' THEN 1 ELSE 0 END) AS order_count,
        MIN(CASE WHEN kind='BROKER_ORDER' THEN external_ref END) AS order_id,
		SUM(CASE WHEN kind='MUTATION_ATTEMPT' THEN 1 ELSE 0 END) AS mutation_count,
		MIN(CASE WHEN kind='MUTATION_ATTEMPT' THEN external_ref END) AS mutation_id,
        SUM(CASE WHEN kind='FILL' THEN 1 ELSE 0 END) AS fill_count,
        MIN(CASE WHEN kind='FILL' THEN external_ref END) AS fill_id,
        SUM(CASE WHEN kind='POSITION' THEN 1 ELSE 0 END) AS position_count,
        MIN(CASE WHEN kind='POSITION' THEN external_ref END) AS position_id,
        SUM(CASE WHEN kind='CLOSE_OUTCOME' THEN 1 ELSE 0 END) AS close_count,
        MIN(CASE WHEN kind='CLOSE_OUTCOME' THEN external_ref END) AS close_id
   FROM strategy_execution_lineage
	 WHERE account_ref=?
  GROUP BY account_ref, attempt_id
), exact_lineage AS (
	SELECT e.account_ref, e.attempt_id, e.mutation_id, e.order_id, e.fill_id, e.position_id, e.close_id,
        a.entry_decision_identity, a.risk_intent_id,
        d.candidate_life_id, d.market, d.symbol, d.threshold_version,
        d.threshold_set_digest, d.evidence_digest, d.lane_id, d.lane_version,
		d.entry_price, d.stop_price, d.target_price, d.quantity, d.policy_version,
		d.created_at
   FROM execution e
   JOIN strategy_attempt_lineage a
     ON a.attempt_id=e.attempt_id AND a.account_ref=e.account_ref
   JOIN strategy_decision_lineage d
     ON d.entry_decision_identity=a.entry_decision_identity
	WHERE e.mutation_count=1 AND e.order_count=1 AND e.fill_count=1
    AND e.position_count=1 AND e.close_count=1
)
SELECT o.position_id, p.market, p.symbol,
       o.realized_pnl_after_costs, o.realized_r, o.initial_quantity,
       o.closed_at, o.cost_total,
       p.opened_at, es.entry_price, es.policy_id, es.policy_version,
       risk.id, risk.account_ref, risk.preimage_kind, risk.risk_preimage,
       risk.risk_hash, risk.issued_at,
       x.candidate_life_id, x.threshold_version, x.threshold_set_digest,
       x.evidence_digest, x.lane_id, x.lane_version,
	   x.entry_decision_identity, x.attempt_id, x.mutation_id, x.order_id, x.fill_id,
	   x.market, x.symbol, x.entry_price, x.stop_price, x.target_price,
	   x.quantity, x.policy_version, x.risk_intent_id, x.created_at
  FROM trade_outcomes o
  JOIN positions p ON p.id=o.position_id
	LEFT JOIN decisions risk ON risk.id=p.entry_decision_id AND risk.account_ref=p.account_ref
	LEFT JOIN exit_states es ON es.position_id=p.id
  LEFT JOIN exact_lineage x
    ON x.account_ref=p.account_ref
   AND x.position_id=o.position_id
   AND x.close_id=o.position_id
   AND x.risk_intent_id=p.entry_decision_id
 WHERE p.account_ref=?
   AND p.adoption_id IS NULL
   AND o.closed_at>?
   AND o.closed_at<=?
 ORDER BY o.closed_at, o.position_id`

func scanClosedStrategyTradeSource(rows *sql.Rows, account string) (ClosedStrategyTradeSource, error) {
	var (
		row                                         ClosedStrategyTradeSource
		symbol, closedAt, openedAt                  string
		cost, policyID, policyVersion               sql.NullString
		riskID, riskAccount, preimageKind           string
		riskPreimage, riskHash, riskIssuedAt        string
		candidateLifeID, thresholdVersion           sql.NullString
		thresholdSetDigest, evidenceDigest          sql.NullString
		laneID, laneVersion, decisionID             sql.NullString
		attemptID, mutationID, orderID, fillID      sql.NullString
		lineageMarket, lineageSymbol                sql.NullString
		lineageEntry, lineageStop, lineageTarget    sql.NullString
		lineageQuantity, lineagePolicy, lineageRisk sql.NullString
		lineageCreatedAt                            sql.NullString
	)
	if err := rows.Scan(
		&row.PositionID, &row.Market, &symbol,
		&row.RealizedPnLAfterCosts, &row.RealizedR, &row.Quantity,
		&closedAt, &cost,
		&openedAt, &row.EntryPrice, &policyID, &policyVersion,
		&riskID, &riskAccount, &preimageKind, &riskPreimage, &riskHash, &riskIssuedAt,
		&candidateLifeID, &thresholdVersion, &thresholdSetDigest, &evidenceDigest,
		&laneID, &laneVersion, &decisionID, &attemptID, &mutationID, &orderID, &fillID,
		&lineageMarket, &lineageSymbol, &lineageEntry, &lineageStop, &lineageTarget,
		&lineageQuantity, &lineagePolicy, &lineageRisk, &lineageCreatedAt,
	); err != nil {
		return ClosedStrategyTradeSource{}, fmt.Errorf("journal: scanning closed strategy trade source: %w", err)
	}

	row.TradeID, row.CloseID = row.PositionID, row.PositionID
	row.PolicyID, row.PolicyVersion = policyID.String, policyVersion.String
	if cost.Valid {
		row.CostTotal = &cost.String
	}
	var err error
	if row.ClosedAt, err = parseJournalTime(closedAt); err != nil {
		return ClosedStrategyTradeSource{}, fmt.Errorf("journal: trade outcome %s closed_at: %w", row.PositionID, err)
	}
	if row.EntryAt, err = parseJournalTime(openedAt); err != nil {
		return ClosedStrategyTradeSource{}, fmt.Errorf("journal: trade outcome %s opened_at: %w", row.PositionID, err)
	}
	if row.DecisionAt, err = parseJournalTime(riskIssuedAt); err != nil {
		return ClosedStrategyTradeSource{}, fmt.Errorf("journal: trade outcome %s decision issued_at: %w", row.PositionID, err)
	}

	if riskAccount != account || preimageKind != PreimageKindRiskIntent || HashPreimage(riskPreimage) != riskHash {
		return ClosedStrategyTradeSource{}, fmt.Errorf("journal: trade outcome %s has invalid risk binding", row.PositionID)
	}
	parsed, err := ParsePreimage(preimageKind, riskPreimage)
	if err != nil {
		return ClosedStrategyTradeSource{}, fmt.Errorf("journal: trade outcome %s risk preimage: %w", row.PositionID, err)
	}
	risk, ok := parsed.(RiskIntent)
	if !ok {
		return ClosedStrategyTradeSource{}, fmt.Errorf("journal: trade outcome %s is not bound to a RiskIntent", row.PositionID)
	}
	if risk.AccountRef != account || risk.Market != row.Market || risk.Symbol != symbol {
		return ClosedStrategyTradeSource{}, fmt.Errorf("journal: trade outcome %s risk venue/account mismatch", row.PositionID)
	}
	row.Side = risk.Side
	row.DecisionPrice = risk.EntryPrice
	for name, decimal := range map[string]string{
		"realized_pnl_after_costs": row.RealizedPnLAfterCosts,
		"realized_r":               row.RealizedR,
	} {
		if _, ok := new(big.Rat).SetString(strings.TrimSpace(decimal)); !ok {
			return ClosedStrategyTradeSource{}, fmt.Errorf("journal: trade outcome %s has invalid %s", row.PositionID, name)
		}
	}
	for name, decimal := range map[string]string{
		"decision price": row.DecisionPrice,
		"entry price":    row.EntryPrice,
		"quantity":       row.Quantity,
	} {
		value, ok := new(big.Rat).SetString(strings.TrimSpace(decimal))
		if !ok || value.Sign() <= 0 {
			return ClosedStrategyTradeSource{}, fmt.Errorf("journal: trade outcome %s has invalid %s", row.PositionID, name)
		}
	}
	if row.CostTotal != nil {
		value, ok := new(big.Rat).SetString(strings.TrimSpace(*row.CostTotal))
		if !ok || value.Sign() < 0 {
			return ClosedStrategyTradeSource{}, fmt.Errorf("journal: trade outcome %s has invalid cost_total", row.PositionID)
		}
	}

	// No candidate means a legitimate pre-a047 closed trade: preserve the
	// authoritative outcome and let the derived model label the missing link.
	if !candidateLifeID.Valid {
		return row, nil
	}
	lineageValues := []sql.NullString{
		candidateLifeID, thresholdVersion, thresholdSetDigest, evidenceDigest,
		laneID, laneVersion, decisionID, attemptID, mutationID, orderID, fillID,
		lineageMarket, lineageSymbol, lineageEntry, lineageStop, lineageTarget,
		lineageQuantity, lineagePolicy, lineageRisk, lineageCreatedAt,
	}
	for _, value := range lineageValues {
		if !value.Valid || strings.TrimSpace(value.String) == "" {
			return ClosedStrategyTradeSource{}, fmt.Errorf("journal: trade outcome %s has partial strategy lineage", row.PositionID)
		}
	}
	if lineageRisk.String != riskID || !strings.EqualFold(lineageMarket.String, row.Market) ||
		lineageSymbol.String != symbol || lineageEntry.String != risk.EntryPrice ||
		lineageStop.String != risk.StopPrice || lineageTarget.String != risk.TargetPrice ||
		lineageQuantity.String != risk.Quantity || lineagePolicy.String != risk.PolicyVersion {
		return ClosedStrategyTradeSource{}, fmt.Errorf("journal: trade outcome %s strategy/risk binding mismatch", row.PositionID)
	}
	strategyDecisionAt, err := parseJournalTime(lineageCreatedAt.String)
	if err != nil {
		return ClosedStrategyTradeSource{}, fmt.Errorf("journal: trade outcome %s strategy decision created_at: %w", row.PositionID, err)
	}
	row.DecisionAt = strategyDecisionAt
	row.Lineage = &ClosedStrategyLineage{
		CandidateLifeID: candidateLifeID.String, ThresholdVersion: thresholdVersion.String,
		ThresholdSetDigest: thresholdSetDigest.String, EvidenceDigest: evidenceDigest.String,
		LaneID: laneID.String, LaneVersion: laneVersion.String,
		StrategyDecisionIdentity: decisionID.String, RiskIntentID: lineageRisk.String,
		StrategyAttemptID: attemptID.String, MutationAttemptID: mutationID.String,
		BrokerOrderID: orderID.String, FillID: fillID.String,
		PositionID: row.PositionID, CloseOutcomeID: row.CloseID,
	}
	return row, nil
}
