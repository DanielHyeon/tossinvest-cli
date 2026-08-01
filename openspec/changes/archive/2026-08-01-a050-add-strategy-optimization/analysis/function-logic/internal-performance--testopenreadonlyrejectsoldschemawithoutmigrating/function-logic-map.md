# Function Logic Map: `TestOpenReadOnlyRejectsOldSchemaWithoutMigrating`

- Source: `internal/performance/readonly_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| old-schema fixture | SQLite `user_version=0`, zero application tables | temporary DB | read-only opener returns nil plus `ErrSchemaTooOld` |
| post-refusal state | version remains zero and no tables are added | raw mode-ro verification | any migration side effect fails test |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1-B3 | fixture DB open/version write/close fails | test fixture only | fatal | this test |
| B4 | old schema is accepted or reader is non-nil | test failure only | fatal | this test |
| B5 | raw mode-ro verifier cannot open | none | fatal | this test |
| B6-B7 | version/table verification queries fail | none | fatal | this test |
| B8 | version or table count changed | test failure only | fatal | this test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| raw SQLite setup/verifier | create and inspect old schema without product migration | temporary fixture only | assertions |
| `OpenReadOnly` | exercise version refusal | one call, no retry | typed error assertion |

## State mutations and fallbacks

- Only test setup writes `user_version=0`. Product read-only open must perform no migration or table creation.

## Safety conclusion

- Safe edit boundary: backward-incompatible schema refusal test.
- High-risk impact: no direct trading effect; prevents hidden migration from a console reader.
