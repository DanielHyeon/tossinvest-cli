# Function Logic Map: `TestOpenReadOnlyMissingDatabaseCreatesNothing`

- Source: `internal/performance/readonly_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| missing nested DB path | parent data directory does not exist | temporary filesystem fixture | opener returns nil plus `ErrDatabaseMissing`; directory remains absent |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | result is not typed missing error or reader is non-nil | test failure only | fatal | this test |
| B2 | parent directory exists after attempt | test failure only | fatal | this test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `OpenReadOnly`, `os.Stat` | exercise and verify non-creating behavior | one call each, no retry | assertions |

## State mutations and fallbacks

- Only creates the outer test temp directory; the product opener must not create its requested data directory or DB.

## Safety conclusion

- Safe edit boundary: negative filesystem-side-effect test.
- High-risk impact: no direct production mutation.
