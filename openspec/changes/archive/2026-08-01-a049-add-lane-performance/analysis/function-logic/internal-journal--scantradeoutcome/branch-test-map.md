# Branch Test Map: `scanTradeOutcome`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `if` at line 729: `if err := rows.Scan(&o.PositionID, &o.RealizedPnLAfterCosts, &o.RealizedR,`; invariant: missing/corrupt/alternate path is explicit | `TestTheOutcomeIsFrozenInTheClosingTransaction`, `TestMigrationV14ToV15IsAdditiveNullableAndIdempotent` | not separately captured for this evidence refresh | package-targeted regression PASS before integration; rerun by gate |
| B2 | `if` at line 735: `if rung.Valid {`; invariant: missing/corrupt/alternate path is explicit | `TestTheOutcomeIsFrozenInTheClosingTransaction`, `TestMigrationV14ToV15IsAdditiveNullableAndIdempotent` | not separately captured for this evidence refresh | package-targeted regression PASS before integration; rerun by gate |
| B3 | `if` at line 738: `if cost.Valid {`; invariant: missing/corrupt/alternate path is explicit | `TestTheOutcomeIsFrozenInTheClosingTransaction`, `TestMigrationV14ToV15IsAdditiveNullableAndIdempotent` | not separately captured for this evidence refresh | package-targeted regression PASS before integration; rerun by gate |
