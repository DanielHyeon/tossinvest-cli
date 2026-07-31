# Function Logic Map: `TestMigrationV8ToV9PreservesRowsAndAddsNullableSnapshots`

- Source: `internal/journal/migration_v9_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| v8 fixture | real schema-v8 database with seeded rows | migration helper | test fails on open/count/metadata mismatch |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1-B7 | setup, migrate explicitly to v9, verify rows and nullable columns | isolated temp database only | fatal assertion | migration_v9 test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| openJournalAtSchema/countRows | exercise the historical v8→v9 contract without advancing to v10 | test helper errors fail immediately | AST |

## State mutations and fallbacks

- Test-only fixture mutation; the a042 edit pins the historical target to v9 after head became v10.

## Safety conclusion

- Safe edit boundary: test-only historical migration harness.
- High-risk impact: no; production code is not called with LIVE authority.
