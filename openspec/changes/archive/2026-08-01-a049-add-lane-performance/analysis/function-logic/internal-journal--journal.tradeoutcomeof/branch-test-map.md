# Branch Test Map: `Journal.TradeOutcomeOf`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `if` at line 684: `if err != nil {`; invariant: missing/corrupt/alternate path is explicit | `TestTheOutcomeIsFrozenInTheClosingTransaction`, `TestMigrationV14ToV15IsAdditiveNullableAndIdempotent` | not separately captured for this evidence refresh | package-targeted regression PASS before integration; rerun by gate |
| B2 | `if` at line 688: `if !rows.Next() {`; invariant: missing/corrupt/alternate path is explicit | `TestTheOutcomeIsFrozenInTheClosingTransaction`, `TestMigrationV14ToV15IsAdditiveNullableAndIdempotent` | not separately captured for this evidence refresh | package-targeted regression PASS before integration; rerun by gate |
