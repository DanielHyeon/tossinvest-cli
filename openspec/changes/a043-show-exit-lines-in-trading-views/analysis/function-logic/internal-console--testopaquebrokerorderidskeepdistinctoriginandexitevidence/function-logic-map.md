# Function Logic Map: `TestOpaqueBrokerOrderIDsKeepDistinctOriginAndExitEvidence`

- Source: `internal/console/orders_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| journal id `" O-1 "` and broker rows `" O-1 "`/`"O-1"` | byte-exact, same account/market/day | broker and journal persisted identity | two visible rows; only spaced id receives origin/evidence |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | rows/count are not distinct | test failure only | expose erroneous dedupe | this test |
| B2 | collect rows by exact rendered id | local map only | none | this test |
| B3 | spaced id lacks exact engine/evidence link | test failure only | expose canonicalization | this test |
| B4 | plain id inherits spaced evidence | test failure only | expose cross-link | this test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `seedEngineJournal`, `Console.orders` | seed exact persisted identity and exercise read model | local SQLite; no retry | current AST |

## State mutations and fallbacks

- Test-only journal and broker seam. It asserts both origin and exit evidence, not merely row count.

## Safety conclusion

- Safe edit boundary: byte-exact identity regression coverage.
- High-risk impact: no.
