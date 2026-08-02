# Function Logic Map: `positionRow.Reason`

- Source: `internal/console/portfolio.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| Broker/journal presence | known, unknown, broker-only, joined | holdings + journal read model | page-global states return no duplicate row reason |
| Management designation | desired include/exclude and typed runtime projection | settings + engine projection | desired-only rows must not claim effective adoption or protection |
| Exit eligibility/state | eligible/ineligible, exit present/absent | journal | absent evidence yields explanation only; never a price |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `switch` | selects the first authoritative explanation | none | all row-state cases |
| B2 | journal unknown or position not in journal | none | empty; page-global notice owns the explanation | unknown/broker-only tests |
| B3 | typed management projection is adoption-pending or reconcile-blocked | none | reconciliation is pending/blocked and protection is not yet effective | US pending/blocked tests |
| B4 | shared pending-designation predicate is true | none | stored request plus engine-reflection-unknown explanation | unavailable-commander and managed/designated collision tests |
| B5 | position is not exit-eligible | none | no entry/adoption basis explanation | unmanaged tests |
| B6 | eligible but exit state absent | none | observation-pending explanation | managed pending tests |
| B7 | no row-specific absence reason | none | empty | managed exit tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `Unknown` | preserve page-global journal failure wording | pure predicate | AST + portfolio tests |
| `HasManagementProjection` | distinguish engine-owned state from desired-only fallback | pure predicate | a052/a053 tests |
| `PendingDesignation` | exclude managed, released, unknown, excluded, broker-missing, or projected rows from desired-only copy | pure predicate | post-deploy collision regression + released truth table |
| `Basis` | explain the qualifying journal record | pure display projection | portfolio tests |

## State mutations and fallbacks

- Pure string projection; no state mutation, I/O, price arithmetic, or command capability.
- Known release and exclusion remain stronger than designation, and desired designation never becomes an effective/protected claim.

## Safety conclusion

- Safe edit boundary: row-local explanatory copy derived from existing read-only flags.
- High-risk impact: no direct trading side effect; wording must not imply protection when runtime is unknown.
