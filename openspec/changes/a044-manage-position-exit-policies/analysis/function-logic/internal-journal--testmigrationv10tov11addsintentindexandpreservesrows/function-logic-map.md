# Function Logic Map: `TestMigrationV10ToV11AddsIntentIndexAndPreservesRows`

- Source: `internal/journal/migration_v11_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| v10 journal | one position, intent, and exit-event lineage row | released v10 migration plan | fail test without altering fixture |
| v11 target | index-only a043 schema | `schemaV11` | exact version/index/row assertions |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1-B2 | seed and close v10 | durable fixture row | fatal | test body |
| B3-B6 | open exactly v11 and verify version/index/lineage | read only | fatal on mismatch | test body |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `openJournalAtSchema(...,11)` | isolate a043 step from later head migrations | never advances to v12 | AST |

## State mutations and fallbacks

- v11 remains owned by a043 and independently testable after v12 is appended.

## Safety conclusion

- Safe edit boundary: keep the target pinned to 11; no current-head open in this test.
- High-risk impact: yes
