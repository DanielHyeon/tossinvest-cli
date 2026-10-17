# Function Logic Map: `TestMigrationV21AddsOnlyNullableConsumedSnapshotReference`

- Source: `internal/journal/strategy_evidence_migration_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| v20 database fixture | migrate only through historical v21 | released-schema immutability | test failure |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1-B10 | seed, migrate and inspect historical nullable reference without later schema | test database only | assertions | this test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| historical migration override | pin v21 behavior | fatal on mismatch | AST |

## State mutations and fallbacks

- Test-only historical migration; v24 must not leak into v21 golden.

## Safety conclusion

- Safe edit boundary: explicit schema target only.
- High-risk impact: no production mutation; protects migration history.
