# Function Logic Map: `TestMigrationV11ToV12IsAdditiveAndPreservesExistingRows`

- Source: `internal/journal/migration_v12_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| v11 journal | seeded audit rows plus a043 index | released v11 plan | preserve all rows/index |
| v12 target | lifecycle columns/tables/indexes only | `schemaV12` | exact version/artifact assertions |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1-B3 | seed/close v11, open v12, preserve all rows | durable migration | fatal on mismatch | test body |
| B4-B6 | iterate and verify v11 index plus both v12 tables | read only | fatal on missing artifact | test body |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `openJournalAtSchema(...,11)` | construct real predecessor | exact v11 | AST |
| `openTestJournalAt` | run production default to v12 | no alternate path | AST |

## State mutations and fallbacks

- v12 appends policy lifecycle state without editing a043's v11 step or historical rows.

## Safety conclusion

- Safe edit boundary: preserve v11 artifact and all prior row counts.
- High-risk impact: yes
