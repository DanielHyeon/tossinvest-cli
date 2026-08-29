# Function Logic Map: `TestMigrationV24AddsOfficialZeroAuthorityAndReleaseReceiptsWithoutRewritingV23`

- Source: `internal/journal/risk_bucket_migration_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| v23 database | migrate exactly through released v24 | immutable migration history | test failure only |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1-B22 | seed v23, migrate to v24, inspect tables/columns/indexes/triggers and legacy rows | test database only | assertions | same test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| historical migration override | prevent later v25 from changing v24 golden | fatal on mismatch | AST |

## State mutations and fallbacks

- Test-only historical database writes.

## Safety conclusion

- Safe edit boundary: pin Open target to 24; no v24 SQL edit.
- High-risk impact: no runtime impact; preserves released migration proof.
